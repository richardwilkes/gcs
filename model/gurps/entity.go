// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"hash"
	"io/fs"
	"log/slog"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/threshold"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/zeebo/xxh3"
)

var (
	_ ListProvider     = &Entity{}
	_ DataOwner        = &Entity{}
	_ Hashable         = &Entity{}
	_ PageInfoProvider = &Entity{}
)

// PointsBreakdown holds the points spent on a character.
type PointsBreakdown struct {
	Ancestry      fxp.Int
	Attributes    fxp.Int
	Advantages    fxp.Int
	Disadvantages fxp.Int
	Quirks        fxp.Int
	Skills        fxp.Int
	Spells        fxp.Int
}

// Total returns the total number of points spent on a character.
func (pb *PointsBreakdown) Total() fxp.Int {
	return pb.Ancestry + pb.Attributes + pb.Advantages + pb.Disadvantages + pb.Quirks + pb.Skills + pb.Spells
}

// EntityData holds the Entity data that is written to disk.
type EntityData struct {
	Version          int             `json:"version"`
	ID               tid.TID         `json:"id"`
	TotalPoints      fxp.Int         `json:"total_points"`
	PointsRecord     []*PointsRecord `json:"points_record,omitempty"`
	Profile          Profile         `json:"profile"`
	SheetSettings    *SheetSettings  `json:"settings,omitzero"`
	Attributes       *Attributes     `json:"attributes,omitzero"`
	Traits           []*Trait        `json:"traits,omitempty"`
	Skills           []*Skill        `json:"skills,omitempty"`
	Spells           []*Spell        `json:"spells,omitempty"`
	CarriedEquipment []*Equipment    `json:"equipment,omitempty"`
	OtherEquipment   []*Equipment    `json:"other_equipment,omitempty"`
	Notes            []*Note         `json:"notes,omitempty"`
	CreatedOn        jio.Time        `json:"created_date"`
	ModifiedOn       jio.Time        `json:"modified_date"`
	ThirdParty       map[string]any  `json:"third_party,omitempty"`
}

type features struct {
	attributeBonuses     []*AttributeBonus
	costReductions       []*CostReduction
	drBonuses            []*DRBonus
	maxUsesBonuses       []*EquipmentMaxUsesBonus
	skillBonuses         []*SkillBonus
	skillPointBonuses    []*SkillPointBonus
	spellBonuses         []*SpellBonus
	spellPointBonuses    []*SpellPointBonus
	traitBonuses         []*TraitBonus
	traitMaxLevelBonuses []*TraitMaxLevelBonus
	weaponBonuses        []*WeaponBonus
	selectorOverrides    []*SelectorOverride
}

// Entity holds the base information for various types of entities: PC, NPC, Creature, etc.
type Entity struct {
	EntityData
	LiftingStrengthBonus           fxp.Int
	StrikingStrengthBonus          fxp.Int
	ThrowingStrengthBonus          fxp.Int
	DodgeBonus                     fxp.Int
	ParryBonus                     fxp.Int
	ParryBonusTooltip              string
	BlockBonus                     fxp.Int
	BlockBonusTooltip              string
	srcMatcher                     *SrcMatcher
	features                       features
	variableResolverExclusions     map[string]bool
	skillResolverExclusions        map[string]bool
	scriptCache                    map[scriptResolveKey]scriptResolveResult // not safe for concurrent use
	scriptResolvingDepth           int
	abandonedScripts               int64
	variableCache                  map[string]string
	basicLiftCache                 fxp.Weight
	encumbranceLevelCache          encumbrance.Level
	encumbranceLevelForSkillsCache encumbrance.Level
	// unsettled records that the last recalculation found data that never settles (see Recalculate), so that it is
	// logged once rather than on every edit.
	unsettled bool
}

// NewEntityFromFile loads an Entity from a file.
func NewEntityFromFile(fileSystem fs.FS, filePath string) (*Entity, error) {
	return newEntityFromFile(fileSystem, filePath)
}

// NewEntityFromFileWithSavedCalc loads an Entity from a file without recalculating it. The derived values that the
// file's "calc" objects record for its traits and notes are kept instead (see Trait.StringWithSavedCalc and
// Note.StringWithSavedCalc), so a reader that only needs the text of the sheet -- the navigator's deep search, which
// indexes every sheet in the libraries -- is spared the cost of deriving them, which runs the scripts in the data. The
// entity it returns has not had its items attached, its skills leveled or its features processed, and is not fit for
// display, editing or saving.
func NewEntityFromFileWithSavedCalc(fileSystem fs.FS, filePath string) (*Entity, error) {
	return newEntityFromFile(fileSystem, filePath, json.WithUnmarshalers(savedCalcMarker))
}

func newEntityFromFile(fileSystem fs.FS, filePath string, opts ...json.Options) (*Entity, error) {
	var e Entity
	e.DiscardCaches()
	if err := jio.LoadVersionedFile(fileSystem, filePath, &e, &e.Version, opts...); err != nil {
		return nil, err
	}
	return &e, nil
}

// NewEntity creates a new Entity.
func NewEntity() *Entity {
	settings := GlobalSettings().GeneralSettings()
	var e Entity
	e.DiscardCaches()
	e.ID = tid.MustNewTID(kinds.Entity)
	e.TotalPoints = settings.InitialPoints
	e.PointsRecord = append(e.PointsRecord, &PointsRecord{
		When:   jio.Now(),
		Points: settings.InitialPoints,
		Reason: i18n.Text("Initial points"),
	})
	e.CreatedOn = jio.Now()
	e.SheetSettings = GlobalSettings().SheetSettings().Clone(&e)
	e.Attributes = NewAttributes(&e)
	if settings.AutoFillProfile {
		e.Profile.AutoFill(&e)
	}
	if settings.AutoAddNaturalAttacks {
		e.Traits = append(e.Traits, NewNaturalAttacks(&e, nil))
	}
	e.ModifiedOn = e.CreatedOn
	e.Recalculate()
	return &e
}

// DataOwner returns the data owner.
func (e *Entity) DataOwner() DataOwner {
	return e
}

// OwningEntity returns the Entity.
func (e *Entity) OwningEntity() *Entity {
	return e
}

// SourceMatcher returns the SourceMatcher.
func (e *Entity) SourceMatcher() *SrcMatcher {
	if e.srcMatcher == nil {
		e.srcMatcher = &SrcMatcher{}
	}
	return e.srcMatcher
}

// Entity implements EntityProvider.
func (e *Entity) Entity() *Entity {
	return e
}

// Save the Entity to a file as JSON.
func (e *Entity) Save(filePath string) error {
	e.Recalculate()
	AdjustEquipmentUsesForSave(e.CarriedEquipment)
	AdjustEquipmentUsesForSave(e.OtherEquipment)
	return jio.SaveToFile(filePath, e)
}

// MarshalJSONTo implements json.MarshalerTo.
//
// Writing an entity out does not recalculate it first. Recalculating is not a read-only operation — it updates the
// skill levels, the features and prerequisites, and the "defaulted_from" of every skill, the last of which is written
// to disk — so merely hashing an entity to ask whether it has unsaved changes would rewrite part of it, from values
// the scripts in the data produce, which are not guaranteed to come out the same twice. The callers that need the
// derived state current — Save, the exporters, and the sheet whenever anything changes — recalculate for themselves.
func (e *Entity) MarshalJSONTo(enc *jsontext.Encoder) error {
	if omitCalc(enc) {
		data := e.EntityData
		data.Version = jio.CurrentDataVersion
		return json.MarshalEncode(enc, &data)
	}
	type calc struct {
		Swing                 dice.Dice  `json:"swing"`
		Thrust                dice.Dice  `json:"thrust"`
		BasicLift             fxp.Weight `json:"basic_lift"`
		LiftingStrengthBonus  fxp.Int    `json:"lifting_st_bonus,omitzero"`
		StrikingStrengthBonus fxp.Int    `json:"striking_st_bonus,omitzero"`
		ThrowingStrengthBonus fxp.Int    `json:"throwing_st_bonus,omitzero"`
		DodgeBonus            fxp.Int    `json:"dodge_bonus,omitzero"`
		ParryBonus            fxp.Int    `json:"parry_bonus,omitzero"`
		BlockBonus            fxp.Int    `json:"block_bonus,omitzero"`
		Move                  []int      `json:"move"`
		Dodge                 []int      `json:"dodge"`
	}
	data := struct {
		EntityData
		Calc calc `json:"calc"`
	}{
		EntityData: e.EntityData,
		Calc: calc{
			Swing:                 e.Swing(),
			Thrust:                e.Thrust(),
			BasicLift:             e.BasicLift(),
			LiftingStrengthBonus:  e.LiftingStrengthBonus,
			StrikingStrengthBonus: e.StrikingStrengthBonus,
			ThrowingStrengthBonus: e.ThrowingStrengthBonus,
			DodgeBonus:            e.DodgeBonus,
			ParryBonus:            e.ParryBonus,
			BlockBonus:            e.BlockBonus,
			Move:                  make([]int, len(encumbrance.Levels)),
			Dodge:                 make([]int, len(encumbrance.Levels)),
		},
	}
	for i, one := range encumbrance.Levels {
		data.Calc.Move[i] = e.Move(one)
		data.Calc.Dodge[i] = e.Dodge(one)
	}
	data.Version = jio.CurrentDataVersion
	return json.MarshalEncode(enc, &data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom. The entity is recalculated once loaded, unless the unmarshal is
// one that keeps the saved "calc" values instead (see NewEntityFromFileWithSavedCalc).
func (e *Entity) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var content struct {
		EntityData
		OldTraits []*Trait `json:"advantages"`
	}
	if err := json.UnmarshalDecode(dec, &content); err != nil {
		return err
	}
	e.EntityData = content.EntityData
	if e.Traits == nil && content.OldTraits != nil {
		e.Traits = content.OldTraits
	}
	if !tid.IsKindAndValid(e.ID, kinds.Entity) {
		e.ID = tid.MustNewTID(kinds.Entity)
	}
	// The clone comes from the published snapshot rather than the live global settings, since unmarshaling may be
	// running on a background goroutine (the deep search content cache) while the UI thread mutates the live settings.
	if e.SheetSettings == nil {
		e.SheetSettings = globalSheetSettingsClone(e)
	}
	if e.Attributes == nil {
		e.Attributes = NewAttributes(e)
	}
	if e.Version < noNeedForRewrapVersion {
		e.SheetSettings.BodyType.Rewrap()
	}
	var total fxp.Int
	for _, rec := range e.PointsRecord {
		total += rec.Points
	}
	if total != e.TotalPoints {
		e.PointsRecord = append(e.PointsRecord, &PointsRecord{
			Points: e.TotalPoints - total,
			When:   jio.Now(),
			Reason: i18n.Text("Reconciliation"),
		})
		slices.SortFunc(e.PointsRecord, func(a, b *PointsRecord) int { return b.When.Compare(a.When) })
	}
	if !keepSavedCalc(dec) {
		e.Recalculate()
	}
	return nil
}

// DiscardCaches discards the internal caches.
func (e *Entity) DiscardCaches() {
	e.discardCaches(false)
}

// discardCaches discards the internal caches. With keepAbandonedScripts set, the results recorded for scripts that
// were stopped before they could produce an answer are kept, as a recalculation wants between its passes: such a
// script would almost certainly be stopped again, and running it on every pass would make a recalculation that is
// repeated on every edit take seconds. The kept result is still reported as abandoned to whatever reads it, and the
// next recalculation runs the script again. The cost is that a script stopped once stays stopped for the rest of that
// recalculation, so a skill or spell whose level it feeds keeps the level it had, since Skill.UpdateLevel and
// Spell.UpdateLevel keep what was there rather than record a level computed from a stand-in, and the passes settle
// with that level in place.
func (e *Entity) discardCaches(keepAbandonedScripts bool) {
	e.variableResolverExclusions = make(map[string]bool)
	e.skillResolverExclusions = make(map[string]bool)
	scriptCache := make(map[scriptResolveKey]scriptResolveResult)
	if keepAbandonedScripts {
		for key, result := range e.scriptCache {
			if result.abandoned {
				scriptCache[key] = result
			}
		}
	}
	e.scriptCache = scriptCache
	e.variableCache = make(map[string]string)
	e.basicLiftCache = -1
	e.encumbranceLevelCache = encumbrance.LastLevel + 1
	e.encumbranceLevelForSkillsCache = encumbrance.LastLevel + 1
}

// maxRecalculationPasses caps the passes one round of recalculation makes before concluding that the sheet's data will
// never settle. Legitimate data settles in a handful of passes: a pass carries each change one link along the chain of
// things that depend on it, so a chain of n dependent traits settles in n+2 passes, one to disable each link in turn,
// one for the levels to lose the features of the last, and one that finds nothing left to change, and the chains on a
// real sheet are a few links long. Data that contradicts itself is normally caught sooner, when the passes repeat a
// state, so the cap matters only for data that never repeats one, such as a prerequisite script that consults a random
// number, and for a chain of more links than the cap allows for, which is stopped short and reported as data that
// never settles; see recalculateUntilSettled for how the two are told apart. Every pass runs every script again, so
// the cost of the cap is the cost of a pass on the sheet times this many, times the rounds the recalculation makes,
// which maxRecalculationPassesInAll bounds.
const maxRecalculationPasses = 32

// maxRecalculationPassesInAll caps the passes all the rounds of one recalculation make between them; see Recalculate
// for the rounds. Each round is capped by maxRecalculationPasses on its own, but a round that marks a trait as caught
// in a contradiction is followed by another, so the rounds alone would let data that never repeats a state, such as
// prerequisite scripts that consult a random number, cost a round for every trait it flips. A legitimate contradiction
// is caught within a few passes per round and unwound in a few rounds, so this is only ever reached by such data, and
// the round it stops is treated as one stopped at its own cap.
const maxRecalculationPassesInAll = 4 * maxRecalculationPasses

// Recalculate the statistics.
//
// The derived values depend on one another in cycles: the features in effect depend on which traits are enabled, the
// attributes and the skill & spell levels depend on the features, the prerequisites depend on the levels and the
// attributes, and when the sheet enforces trait prerequisites, which traits are enabled depends on the prerequisites.
// Scripts may read any of it from anywhere. Rather than track those dependencies, the recalculation repeats passes
// over everything until a pass changes nothing, at which point every derived value agrees with every other.
//
// Which traits the sheet disables is re-derived from the sheet's data alone: every trait the user has enabled starts
// out contributing its features, whatever the previous recalculation decided. Starting from the previous decisions
// would let history leak in: a trait whose own features are what satisfy its prerequisites would stay disabled once it
// had been, while the same sheet loaded afresh would enable it. The skill and spell levels do start from those the
// previous recalculation left, which are normally already right after an edit, since legitimate data settles at the
// same levels from any start.
//
// Each pass judges the prerequisites of every trait against the same state of the sheet and applies the verdicts only
// once all have been judged, so the outcome does not depend on the order in which the traits are listed.
//
// Data that contradicts itself, such as a trait that requires the absence of a trait that requires it, never settles:
// the passes cycle through the same states. The traits whose enablement flips within that cycle are taken to be caught
// in the contradiction, marked as such, and left enabled as the user set them, with any unmet prerequisite still
// shown, since no choice among them satisfies the data. A trait whose prerequisites turn on one of those flips along
// with it and is swept up too, since the passes cannot tell a trait the contradiction runs through from one that
// merely follows it; a child of a flipping container is the exception, since it is judged only while the container is
// enabled and so is not seen to flip (see markContradictedTraits). A further round then follows, in which the marked
// traits stay enabled and every other trait is judged as usual against the traits in effect. A round that cycles again
// marks the traits that flipped in it too, and the rounds end once one settles or marks nothing further. A cycle need
// not involve any trait: two skills whose prerequisites each cap the other's level, and are each penalized while
// unmet, cycle between both penalized and neither. The sheet is then left in a state of the cycle chosen by the cycle
// alone, so that the same data comes out the same on every recalculation. Data that never repeats a state either, such
// as a prerequisite script that consults a random number, is stopped at a cap, and the traits that flipped back and
// forth over the later passes are treated the same way. The rounds such data prompts are stopped at a cap of their own
// on the passes they make between them; see maxRecalculationPassesInAll. Whether the data settled is reported by
// Unsettled.
func (e *Entity) Recalculate() {
	e.recalculate(maxRecalculationPassesInAll)
}

// recalculate is Recalculate with the cap on the passes all of its rounds make between them given rather than fixed,
// and returns how many passes they made. Recalculate gives it maxRecalculationPassesInAll; a test may give it less to
// see the rounds stopped short.
func (e *Entity) recalculate(passBudget int) int {
	if e == nil {
		return 0
	}
	e.EnsureAttachments()
	e.DiscardCaches()
	e.SourceMatcher().PrepareHashes(e)
	Traverse(func(t *Trait) bool {
		t.resetPrereqVerdict()
		return false
	}, false, false, e.Traits...)
	contradicted := false
	made := 0
	for made < passBudget {
		settled, newlyContradicted, passes := e.recalculateUntilSettled(min(passBudget-made, maxRecalculationPasses))
		made += passes
		if settled {
			break
		}
		contradicted = true
		if !newlyContradicted {
			break
		}
	}
	if contradicted && !e.unsettled {
		slog.Warn("the sheet's data never settles: no state of the sheet satisfies all of its prerequisites, "+
			"defaults, features and scripts at once", "name", e.Profile.Name)
	}
	e.unsettled = contradicted
	return made
}

// Unsettled returns true if the last recalculation found that the sheet's data never settles: no state of the sheet
// satisfies all of its prerequisites, defaults, features and scripts at once, so the sheet was left in one state of
// the cycle they run through; see Recalculate. A trait caught in a contradiction among the prerequisites is marked as
// such (see Trait.ContradictedPrereqs), but a cycle need not involve any trait, so this is the only sign of one that
// does not.
func (e *Entity) Unsettled() bool {
	return e != nil && e.unsettled
}

// recalculateUntilSettled makes recalculation passes until one leaves the derived state as it found it and returns
// settled. Otherwise the passes are found to be repeating a state, or the limit on them is reached first, and either
// way the traits whose enablement flipped back and forth are marked as caught in a contradiction, with
// newlyContradicted reporting whether any trait not already marked was, since a further round may then settle; see
// Recalculate. passes is how many passes were made. A cycle is left in its canonical state, whatever state the passes
// were in when it was found; see settleOnCanonicalCycleState. The limit is maxRecalculationPasses, or what is left of
// maxRecalculationPassesInAll when that is less; see recalculate.
//
// Whether a pass changed anything is judged by comparing the derived state before and after it rather than by what
// the pass itself saw change, since a script evaluated during the pass may recompute a level as it reads it, leaving
// the pass to find that level already in place. A prerequisite judged against the level before the script ran would
// otherwise be left standing, with nothing to prompt the pass that would judge it again.
func (e *Entity) recalculateUntilSettled(limit int) (settled, newlyContradicted bool, passes int) {
	// The state the passes begin from is kept along with those they reach, so that a pass that comes back to it is
	// seen to, whether the data has settled or has cycled back.
	states := [][]uint64{e.derivedState(0)}
	var verdicts [][]traitPrereqVerdict
	traitCount := 0
	for passes < limit {
		passes++
		e.recalculationPass()
		passVerdicts := e.traitPrereqVerdicts(traitCount)
		traitCount = len(passVerdicts)
		verdicts = append(verdicts, passVerdicts)
		state := e.derivedState(len(states[0]))
		if slices.Equal(states[len(states)-1], state) {
			return true, false, passes
		}
		if i := slices.IndexFunc(states, func(seen []uint64) bool { return slices.Equal(seen, state) }); i >= 0 {
			// The cycle is the states from index i on, and the verdicts of the passes made from them sit at the same
			// indices, since a pass's verdicts are recorded at the index of the state it began from. The sheet is moved
			// to the cycle's canonical state before any trait is marked, since marking one changes what the passes do.
			e.settleOnCanonicalCycleState(states[i:])
			return false, e.markContradictedTraits(verdicts[i:], true), passes
		}
		states = append(states, state)
	}
	// No state repeated, so the passes were stopped at the limit. Legitimate data has long since settled by the second
	// half of them, so a trait that flipped back and forth over that half is taken to be caught in a contradiction. A
	// trait that flipped only once over it is not: it is a link of a chain longer than the cap allows for, still
	// settling, and is left as its last verdict had it, so that the chain is stopped short rather than given the wrong
	// answer; see maxRecalculationPasses. A limit cut short by maxRecalculationPassesInAll is judged the same way, for
	// want of anything better, and only ever falls to data that has been found never to settle already.
	return false, e.markContradictedTraits(verdicts[len(verdicts)/2:], false), passes
}

// traitPrereqVerdicts returns the verdict the last pass reached for each trait, in traversal order. The capacity is a
// hint for the length and may be zero.
func (e *Entity) traitPrereqVerdicts(capacity int) []traitPrereqVerdict {
	verdicts := make([]traitPrereqVerdict, 0, capacity)
	Traverse(func(t *Trait) bool {
		verdicts = append(verdicts, t.prereqVerdict)
		return false
	}, false, false, e.Traits...)
	return verdicts
}

// markContradictedTraits marks as caught in a contradiction every trait that the verdicts, one set per pass in
// traversal order, flipped back and forth, that is, changed more than once between disabling it and leaving it
// enabled, and returns whether any trait not already marked was. With cyclic set, the verdicts are those of a cycle of
// passes, so the change from the last set back to the first counts too. A trait is judged only while it is enabled,
// itself and by way of the containers above it, so the verdict that it was not judged says nothing about it and is
// passed over: a child of a container that flips is not itself flipping. A marked trait is left enabled from here on,
// so that the next pass judges the others against it.
func (e *Entity) markContradictedTraits(verdicts [][]traitPrereqVerdict, cyclic bool) bool {
	marked := false
	index := 0
	Traverse(func(t *Trait) bool {
		flips := 0
		var first, last traitPrereqVerdict // prereqsNotJudged until a judged verdict is seen
		for _, pass := range verdicts {
			verdict := pass[index]
			if verdict == prereqsNotJudged {
				continue
			}
			if first == prereqsNotJudged {
				first = verdict
			} else if verdict != last {
				flips++
			}
			last = verdict
		}
		if cyclic && last != first {
			flips++
		}
		index++
		if flips > 1 && !t.prereqContradicted {
			t.prereqContradicted = true
			t.setPrereqVerdict(prereqsLeaveEnabled)
			marked = true
		}
		return false
	}, false, false, e.Traits...)
	return marked
}

// recalculationPass recomputes every derived value once, each in turn from those it is built on: the features in
// effect, along with the missing-equipment penalties that the previous pass's judgment of the skill and spell
// prerequisites earned; the prerequisites of the skills and spells, which decide the penalties for the next pass; the
// skill and spell levels; and the prerequisites of the traits and equipment, which read those levels. Whatever reads a
// value before the pass has recomputed it, such as a skill default resolved against the other skills' levels or a
// script reading a level from anywhere, sees the previous pass's and is a pass behind until the next one, which is why
// passes repeat until one changes nothing; see recalculateUntilSettled. The penalties are put in place before anything
// in the pass is judged, rather than as each skill or spell is judged, so that everything in the pass sees the same set
// of them: a script computes the level it reads from the penalties in place as it runs, so one judged early in the
// walk would otherwise see fewer than one judged late, with nothing to prompt a further pass to put that right. The
// caches are discarded first, so that nothing in the pass is computed from a variable, lift or script result the
// previous pass cached, except for the results kept for the reason discardCaches gives.
func (e *Entity) recalculationPass() {
	e.discardCaches(true)
	e.processFeatures()
	e.applyEquipmentPenalties()
	e.processSkillAndSpellPrereqs()
	e.UpdateSkills()
	e.UpdateSpells()
	e.processTraitPrereqs()
	e.processEquipmentPrereqs()
}

// settleOnCanonicalCycleState makes further passes until the sheet is in the canonical state of a cycle the passes
// have been found to be repeating, given the states of the cycle in order, the first being the one the sheet is in.
// Which state the passes are found repeating in depends on the levels they began from, which the previous
// recalculation left, so a sheet whose levels cycle would otherwise change on every edit. The canonical state is chosen
// by the parts of the derived state that differ within the cycle alone, so that it is the same whatever else is on the
// sheet, with their values sorted, so that it depends on the identities of those parts rather than the order they are
// listed in: it is the state whose sorted values come first. Two states sort alike only if fingerprints collide, in
// which case the earlier state in the cycle is taken. Which state it is has no meaning beyond that, since no state of
// the cycle satisfies the data.
func (e *Entity) settleOnCanonicalCycleState(cycle [][]uint64) {
	keys := make([][]uint64, len(cycle))
	for i, first := range cycle[0] {
		if slices.ContainsFunc(cycle[1:], func(state []uint64) bool { return state[i] != first }) {
			for j, state := range cycle {
				keys[j] = append(keys[j], state[i])
			}
		}
	}
	canonical := 0
	for j, key := range keys {
		slices.Sort(key)
		if slices.Compare(key, keys[canonical]) < 0 {
			canonical = j
		}
	}
	for range canonical {
		e.recalculationPass()
	}
}

// derivedState fingerprints the derived state that one recalculation pass hands to the next, one fingerprint per
// trait, skill and spell in traversal order: whether the sheet has disabled the trait, the level of the skill or spell
// and whether it takes the missing-equipment penalty, and the default of the skill, each along with the ID of what it
// describes. Everything else a pass reads is either the sheet's data or recomputed from these before it is read, so
// two passes that begin from the same state end in the same one, and a state seen before means the passes have
// entered a cycle they will never leave. Whether a trait was judged at all is left out, since the next pass reads only
// whether it is enabled. The capacity is a hint for the length and may be zero.
func (e *Entity) derivedState(capacity int) []uint64 {
	state := make([]uint64, 0, capacity)
	h := xxh3.New()
	fingerprint := func(id tid.TID, hashContents func()) {
		h.Reset()
		xhash.StringWithLen(h, string(id))
		hashContents()
		state = append(state, h.Sum64())
	}
	Traverse(func(t *Trait) bool {
		fingerprint(t.TID, func() { xhash.Bool(h, t.prereqVerdict == prereqsDisable) })
		return false
	}, false, false, e.Traits...)
	Traverse(func(s *Skill) bool {
		fingerprint(s.TID, func() {
			s.LevelData.Hash(h)
			xhash.Bool(h, s.takesEquipmentPenalty)
			xhash.Bool(h, s.DefaultedFrom != nil)
			if s.DefaultedFrom != nil {
				s.DefaultedFrom.Hash(h)
				xhash.Num64(h, s.DefaultedFrom.Level)
				xhash.Num64(h, s.DefaultedFrom.Points)
				xhash.Num64(h, s.DefaultedFrom.AdjLevel)
			}
		})
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		fingerprint(s.TID, func() {
			s.LevelData.Hash(h)
			xhash.Bool(h, s.takesEquipmentPenalty)
		})
		return false
	}, false, true, e.Spells...)
	return state
}

// EnsureAttachments ensures that all attachments have their owning entity set to the Entity.
func (e *Entity) EnsureAttachments() {
	e.SheetSettings.SetOwningEntity(e)
	for _, attr := range e.Attributes.Set {
		attr.Entity = e
	}
	SetDataOwnerAll(e, e.Traits)
	SetDataOwnerAll(e, e.Skills)
	SetDataOwnerAll(e, e.Spells)
	SetDataOwnerAll(e, e.CarriedEquipment)
	SetDataOwnerAll(e, e.OtherEquipment)
	SetDataOwnerAll(e, e.Notes)
}

func (e *Entity) processFeatures() {
	e.features = features{}
	e.forEachActiveFeatureList(func(owner, mod fmt.Stringer, leveled LeveledOwner, list Features) {
		for _, f := range list {
			e.processFeature(owner, mod, f, leveled)
		}
	})
	// Self-control-derived features are generated after the full walk, since resolving a trait's self-control roll
	// and adjustment through a selector override requires every override to have been collected first.
	Traverse(func(t *Trait) bool {
		for _, f := range FeaturesForSelfControlRoll(t.ResolvedSelfControl(nil), t.ResolvedSelfControlAdjustment(nil)) {
			e.processFeature(t, nil, f, t)
		}
		return false
	}, true, false, e.Traits...)
	e.LiftingStrengthBonus = e.AttributeBonusFor(StrengthID, stlimit.LiftingOnly, nil).Floor()
	e.StrikingStrengthBonus = e.AttributeBonusFor(StrengthID, stlimit.StrikingOnly, nil).Floor()
	e.ThrowingStrengthBonus = e.AttributeBonusFor(StrengthID, stlimit.ThrowingOnly, nil).Floor()
	for _, attr := range e.Attributes.Set {
		if def := attr.AttributeDef(); def != nil {
			attr.Bonus = e.AttributeBonusFor(attr.AttrID, stlimit.None, nil)
			if !def.AllowsDecimal() {
				attr.Bonus = attr.Bonus.Floor()
			}
			attr.CostReduction = e.CostReductionFor(attr.AttrID)
		} else {
			attr.Bonus = 0
			attr.CostReduction = 0
		}
	}
	e.Profile.Update(e)
	if e.ResolveAttribute(DodgeID) == nil {
		e.DodgeBonus = e.AttributeBonusFor(DodgeID, stlimit.None, nil).Floor()
	} else {
		e.DodgeBonus = 0
	}
	var tooltip xbytes.InsertBuffer
	e.ParryBonus = e.AttributeBonusFor(ParryID, stlimit.None, &tooltip).Floor()
	e.ParryBonusTooltip = tooltip.String()
	tooltip.Reset()
	e.BlockBonus = e.AttributeBonusFor(BlockID, stlimit.None, &tooltip).Floor()
	e.BlockBonusTooltip = tooltip.String()
}

// forEachActiveFeatureList calls fn with each list of features currently in effect on the entity: those of every
// enabled, non-container trait, of every non-container skill and spell, of every carried piece of equipment that is
// really equipped, and of each of their enabled modifiers. Switchable features, whether on an item or on one of its
// modifiers, only take effect while the switch of the primary item is on. The owner is the primary item and mod is the
// modifier carrying the list, or nil when the list is the item's own. The leveled owner is the node whose level drives
// a per-level amount: the modifier itself for trait modifiers, which can have levels of their own, and the equipment
// for equipment modifiers, which cannot. Everything that gathers features from the entity goes through this walk, so
// they all agree on which features are in effect.
func (e *Entity) forEachActiveFeatureList(fn func(owner, mod fmt.Stringer, leveled LeveledOwner, list Features)) {
	Traverse(func(t *Trait) bool {
		if !t.Container() {
			fn(t, nil, t, t.ActiveFeatures())
		}
		Traverse(func(mod *TraitModifier) bool {
			fn(t, mod, mod, mod.Features.Active(t.SwitchedOn))
			return false
		}, true, true, t.Modifiers...)
		return false
	}, true, false, e.Traits...)
	Traverse(func(s *Skill) bool {
		fn(s, nil, s, s.ActiveFeatures())
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		fn(s, nil, s, s.ActiveFeatures())
		return false
	}, false, true, e.Spells...)
	Traverse(func(eqp *Equipment) bool {
		if eqp.ReallyEquipped() {
			forEachActiveEquipmentFeatureList(eqp, func(mod fmt.Stringer, list Features) {
				fn(eqp, mod, eqp, list)
			})
		}
		return false
	}, false, false, e.CarriedEquipment...)
}

// forEachActiveEquipmentFeatureList calls fn with the equipment's own active features and then with those of each of
// its enabled modifiers, subject to the equipment's switch. mod is the modifier carrying the list, or nil for the
// equipment's own. Whether the equipment is equipped is not consulted.
func forEachActiveEquipmentFeatureList(eqp *Equipment, fn func(mod fmt.Stringer, list Features)) {
	fn(nil, eqp.ActiveFeatures())
	Traverse(func(mod *EquipmentModifier) bool {
		fn(mod, mod.Features.Active(eqp.SwitchedOn))
		return false
	}, true, true, eqp.Modifiers...)
}

// processFeature collects a feature into the entity's feature lists. The owner is the primary item the feature came
// from (a trait, skill, spell, or piece of equipment) and the sub-owner, when present, is the modifier of that item
// carrying the feature; together they name the source in tooltips. The leveled owner is the node whose level drives a
// per-level amount: the modifier itself for trait modifiers, which can have levels of their own, and the primary item
// for equipment modifiers, which cannot.
func (e *Entity) processFeature(owner, subOwner fmt.Stringer, f Feature, leveledOwner LeveledOwner) {
	if bonus, ok := f.(Bonus); ok {
		bonus.SetOwner(owner)
		bonus.SetSubOwner(subOwner)
		bonus.SetLeveledOwner(leveledOwner)
	}
	if override, ok := f.(Override); ok {
		override.SetOwner(owner)
		override.SetSubOwner(subOwner)
	}
	switch actual := f.(type) {
	case *AttributeBonus:
		e.features.attributeBonuses = append(e.features.attributeBonuses, actual)
	case *CostReduction:
		e.features.costReductions = append(e.features.costReductions, actual)
	case *EquipmentMaxUsesBonus:
		e.features.maxUsesBonuses = append(e.features.maxUsesBonuses, actual)
	case *DRBonus:
		if len(actual.Locations) == 0 { // "this armor"
			e.expandThisArmorDRBonus(owner, subOwner, leveledOwner, actual)
		} else {
			e.features.drBonuses = append(e.features.drBonuses, actual)
		}
	case *SkillBonus:
		e.features.skillBonuses = append(e.features.skillBonuses, actual)
	case *SkillPointBonus:
		e.features.skillPointBonuses = append(e.features.skillPointBonuses, actual)
	case *SpellBonus:
		e.features.spellBonuses = append(e.features.spellBonuses, actual)
	case *SpellPointBonus:
		e.features.spellPointBonuses = append(e.features.spellPointBonuses, actual)
	case *TraitBonus:
		e.features.traitBonuses = append(e.features.traitBonuses, actual)
	case *TraitMaxLevelBonus:
		e.features.traitMaxLevelBonuses = append(e.features.traitMaxLevelBonuses, actual)
	case *WeaponBonus:
		e.features.weaponBonuses = append(e.features.weaponBonuses, actual)
	case *SelectorOverride:
		e.features.selectorOverrides = append(e.features.selectorOverrides, actual)
	case *ConditionalModifierBonus, *ContainedWeightReduction, *ReactionBonus:
		// Not collected at this stage
	case *UnknownFeature:
		// A feature this version of GCS doesn't understand. It is preserved on save, but has no effect.
	default:
		errs.Log(errs.New("unhandled feature"), "type", f.FeatureType())
	}
}

// defenseBonus returns the entity's bonus for ParryID or BlockID, and zero for any other ID.
func (e *Entity) defenseBonus(defenseID string) fxp.Int {
	switch defenseID {
	case ParryID:
		return e.ParryBonus
	case BlockID:
		return e.BlockBonus
	default:
		return 0
	}
}

// expandThisArmorDRBonus handles a DR bonus that specifies no locations (a "this armor" bonus). Such a bonus applies to
// whatever locations the owning piece of equipment already grants DR to, so a single copy of it, carrying the
// original's specialization, is emitted covering the union of the locations named by the equipment's other DR bonuses
// and those of its enabled modifiers -- a hood that adds DR to the skull makes the skull part of what "this armor"
// covers, no matter which of the two the bonus itself came from. Exactly one copy is emitted, so a location named by
// more than one of those DR bonuses still receives the amount just once. If the owner isn't a piece of equipment, the
// bonus is dropped, since there is nothing for it to attach to.
func (e *Entity) expandThisArmorDRBonus(owner, subOwner fmt.Stringer, leveledOwner LeveledOwner, src *DRBonus) {
	eqp, ok := owner.(*Equipment)
	if !ok {
		return
	}
	// Keyed by the lowercased location, since matching is case-insensitive; the value is the first spelling seen.
	locations := make(map[string]string)
	forEachActiveEquipmentFeatureList(eqp, func(_ fmt.Stringer, list Features) {
		for _, f := range list {
			drBonus, ok2 := f.(*DRBonus)
			if !ok2 || len(drBonus.Locations) == 0 {
				continue
			}
			for _, loc := range drBonus.Locations {
				key := strings.ToLower(loc)
				if _, exists := locations[key]; !exists {
					locations[key] = loc
				}
			}
		}
	})
	if len(locations) == 0 {
		return
	}
	// The switch flag is carried over so that the copy still describes itself the way the original does. It plays no
	// part in whether the copy applies, since the original has already passed the switch gate to get here.
	bonus := &DRBonus{
		Type:           feature.DRBonus,
		FeatureSwitch:  src.FeatureSwitch,
		Locations:      slices.Sorted(maps.Values(locations)),
		Specialization: src.Specialization,
		LeveledAmount:  src.LeveledAmount,
	}
	bonus.SetOwner(owner)
	bonus.SetSubOwner(subOwner)
	bonus.SetLeveledOwner(leveledOwner)
	e.features.drBonuses = append(e.features.drBonuses, bonus)
}

// unsatisfiedReasonPrefix separates the individual reasons within an UnsatisfiedReason.
const unsatisfiedReasonPrefix = "\n- "

// applyEquipmentPenalties adds to the collected features the missing-equipment penalty of every skill and spell that
// processSkillAndSpellPrereqs last found unsatisfied on account of an equipped-equipment prerequisite. It runs before
// the prerequisites are judged again, so that every level computed during the pass counts the same penalties; see
// recalculationPass.
func (e *Entity) applyEquipmentPenalties() {
	Traverse(func(s *Skill) bool {
		if s.takesEquipmentPenalty {
			penalty := NewSkillBonus()
			penalty.NameCriteria.Qualifier = s.NameWithReplacements()
			penalty.SpecializationCriteria.Compare = criteria.IsText
			penalty.SpecializationCriteria.Qualifier = s.SpecializationWithReplacements()
			penalty.OptionalSpecializationCriteria.Compare = criteria.IsText
			penalty.OptionalSpecializationCriteria.Qualifier = s.OptionalSpecializationWithReplacements()
			penalty.Amount = missingEquipmentPenalty(s.TechLevel)
			penalty.SetOwner(s)
			e.features.skillBonuses = append(e.features.skillBonuses, penalty)
		}
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		if s.takesEquipmentPenalty {
			penalty := NewSpellBonus()
			penalty.SpellMatchType = spellmatch.Name
			penalty.NameCriteria.Qualifier = s.NameWithReplacements()
			penalty.Amount = missingEquipmentPenalty(s.TechLevel)
			penalty.SetOwner(s)
			e.features.spellBonuses = append(e.features.spellBonuses, penalty)
		}
		return false
	}, false, true, e.Spells...)
}

// processSkillAndSpellPrereqs evaluates the prerequisites of every skill and spell, recording the reason each is
// unsatisfied. A skill or spell left unsatisfied by an equipped-equipment prerequisite is noted as taking the
// missing-equipment penalty to its level, which applyEquipmentPenalties adds to the collected features in the next
// pass. A technique or ritual magic spell whose prerequisites are met must also have the skill it is based on.
func (e *Entity) processSkillAndSpellPrereqs() {
	Traverse(func(s *Skill) bool {
		if s.Container() {
			s.UnsatisfiedReason = ""
			s.takesEquipmentPenalty = false
			return false
		}
		s.UnsatisfiedReason = e.evaluatePrereqs(s.Prereq, s, &s.takesEquipmentPenalty)
		if s.UnsatisfiedReason == "" && s.IsTechnique() {
			s.UnsatisfiedReason = unsatisfiedReason(func(tooltip *xbytes.InsertBuffer) bool {
				return s.TechniqueSatisfied(tooltip, unsatisfiedReasonPrefix)
			})
		}
		return false
	}, false, false, e.Skills...)
	Traverse(func(s *Spell) bool {
		if s.Container() {
			s.UnsatisfiedReason = ""
			s.takesEquipmentPenalty = false
			return false
		}
		s.UnsatisfiedReason = e.evaluatePrereqs(s.Prereq, s, &s.takesEquipmentPenalty)
		if s.UnsatisfiedReason == "" && s.IsRitualMagic() {
			s.UnsatisfiedReason = unsatisfiedReason(func(tooltip *xbytes.InsertBuffer) bool {
				return s.RitualMagicSatisfied(tooltip, unsatisfiedReasonPrefix)
			})
		}
		return false
	}, false, false, e.Spells...)
}

// processEquipmentPrereqs evaluates the prerequisites of every piece of equipment, recording the reason each is
// unsatisfied.
func (e *Entity) processEquipmentPrereqs() {
	equipmentFunc := func(eqp *Equipment) bool {
		eqp.UnsatisfiedReason = e.evaluatePrereqs(eqp.Prereq, eqp, nil)
		return false
	}
	Traverse(equipmentFunc, false, false, e.CarriedEquipment...)
	Traverse(equipmentFunc, false, false, e.OtherEquipment...)
}

// processTraitPrereqs evaluates the prerequisites and maximum level of every trait, recording the reason each is
// unsatisfied. When the sheet enforces trait prerequisites, a trait whose prerequisites are unmet is disabled, unless
// it has been marked as caught in a contradiction, and one the sheet had disabled is re-enabled once its prerequisites
// are met. Every trait is judged against the traits as they stood when the walk began, and the verdicts are applied
// only after it, so the outcome does not depend on the order the traits are listed in; what one trait's disabling
// means for another is discovered in the next pass.
func (e *Entity) processTraitPrereqs() {
	// Traverse all traits, not just the enabled ones, so that a trait that becomes disabled has any previously
	// recorded unsatisfied reason cleared. Prerequisites are only judged for traits the user has enabled that are not
	// inside a disabled container. That the sheet disabled a trait does not exempt it: the verdict is what this pass
	// computes, and honoring the previous one would keep the trait disabled after its prerequisites are met. The
	// reason is kept for a trait disabled this way, so the sheet can show why it is disabled.
	enforce := e.SheetSettings.EnforceTraitPrereqs
	var verdicts []traitPrereqVerdict
	Traverse(func(t *Trait) bool {
		t.UnsatisfiedReason = ""
		verdict := prereqsNotJudged
		if !t.Disabled && (t.parent == nil || t.parent.Enabled()) {
			t.UnsatisfiedReason = e.evaluatePrereqs(t.Prereq, t, nil)
			if maximum := t.ResolvedMaxLevels(); maximum > 0 && t.Levels > maximum {
				reason := levelExceedsMaximumReason(maximum)
				if t.UnsatisfiedReason == "" {
					t.UnsatisfiedReason = reason
				} else {
					t.UnsatisfiedReason += unsatisfiedReasonPrefix + reason
				}
			}
			verdict = prereqsLeaveEnabled
			if enforce && t.UnsatisfiedReason != "" && !t.prereqContradicted {
				verdict = prereqsDisable
			}
		}
		verdicts = append(verdicts, verdict)
		return false
	}, false, false, e.Traits...)
	index := 0
	Traverse(func(t *Trait) bool {
		t.setPrereqVerdict(verdicts[index])
		index++
		return false
	}, false, false, e.Traits...)
}

func levelExceedsMaximumReason(maximum fxp.Int) string {
	return i18n.Text("Level exceeds the maximum of ") + maximum.String()
}

// evaluatePrereqs evaluates the prerequisites, which may be nil, of the node given as exclude and returns the reason to
// record when they are not met, or "" when they are. hasEquipmentPenalty, if non-nil, is set to whether they are unmet
// on account of an equipped-equipment prerequisite, which is what earns a skill or spell the missing-equipment penalty.
func (e *Entity) evaluatePrereqs(prereq *PrereqList, exclude any, hasEquipmentPenalty *bool) string {
	if hasEquipmentPenalty != nil {
		*hasEquipmentPenalty = false
	}
	return unsatisfiedReason(func(tooltip *xbytes.InsertBuffer) bool {
		return prereq == nil || prereq.Satisfied(e, exclude, tooltip, unsatisfiedReasonPrefix, hasEquipmentPenalty)
	})
}

// unsatisfiedReason runs check, which appends its reasons to the tooltip when it fails, and returns the reason to
// record when it fails, or "" when it passes.
func unsatisfiedReason(check func(tooltip *xbytes.InsertBuffer) bool) string {
	var tooltip xbytes.InsertBuffer
	if check(&tooltip) {
		return ""
	}
	return i18n.Text("Prerequisites have not been met:") + tooltip.String()
}

// missingEquipmentPenalty returns the penalty a skill or spell suffers when its equipment prerequisite is unmet: -10
// when it has a tech level, since it depends on that equipment, and -5 otherwise.
func missingEquipmentPenalty(techLevel *string) fxp.Int {
	if techLevel != nil && *techLevel != "" {
		return -fxp.Ten
	}
	return -fxp.Five
}

// UpdateSkills updates the levels of all skills.
func (e *Entity) UpdateSkills() {
	Traverse(func(s *Skill) bool {
		s.UpdateLevel()
		return false
	}, false, true, e.Skills...)
}

// UpdateSpells updates the levels of all spells.
func (e *Entity) UpdateSpells() {
	Traverse(func(s *Spell) bool {
		s.UpdateLevel()
		return false
	}, false, true, e.Spells...)
}

// UnspentPoints returns the number of unspent points.
func (e *Entity) UnspentPoints() fxp.Int {
	return e.TotalPoints - e.PointsBreakdown().Total()
}

// SetUnspentPoints sets the number of unspent points.
func (e *Entity) SetUnspentPoints(unspent fxp.Int) {
	if unspent != e.UnspentPoints() {
		e.TotalPoints = unspent + e.PointsBreakdown().Total()
	}
}

// PointsBreakdown returns the point breakdown for spent points.
func (e *Entity) PointsBreakdown() *PointsBreakdown {
	var pb PointsBreakdown
	for _, attr := range e.Attributes.Set {
		pb.Attributes += attr.PointCost()
	}
	for _, one := range e.Traits {
		calculateSingleTraitPoints(one, &pb)
	}
	Traverse(func(s *Skill) bool {
		pb.Skills += s.Points
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		pb.Spells += s.Points
		return false
	}, false, true, e.Spells...)
	return &pb
}

func calculateSingleTraitPoints(t *Trait, pb *PointsBreakdown) {
	if t.selfDisabled() {
		return
	}
	if t.Container() {
		switch t.ContainerType {
		case container.Group:
			for _, child := range t.Children {
				calculateSingleTraitPoints(child, pb)
			}
			return
		case container.Ancestry:
			pb.Ancestry += t.AdjustedPoints()
			return
		case container.Attributes:
			pb.Attributes += t.AdjustedPoints()
			return
		default:
		}
	}
	pts := t.AdjustedPoints()
	switch {
	case pts == -fxp.One:
		pb.Quirks += pts
	case pts > 0:
		pb.Advantages += pts
	case pts < 0:
		pb.Disadvantages += pts
	}
}

// WealthCarried returns the current wealth being carried.
func (e *Entity) WealthCarried() fxp.Int {
	var value fxp.Int
	for _, one := range e.CarriedEquipment {
		value += one.ExtendedValue()
	}
	return value
}

// WealthNotCarried returns the current wealth not being carried.
func (e *Entity) WealthNotCarried() fxp.Int {
	var value fxp.Int
	for _, one := range e.OtherEquipment {
		value += one.ExtendedValue()
	}
	return value
}

// StrikingStrength returns the adjusted ST for striking purposes.
func (e *Entity) StrikingStrength() fxp.Int {
	return e.derivedStrength(StrikingStrengthID, e.StrikingStrengthBonus)
}

// LiftingStrength returns the adjusted ST for lifting purposes.
func (e *Entity) LiftingStrength() fxp.Int {
	return e.derivedStrength(LiftingStrengthID, e.LiftingStrengthBonus)
}

// ThrowingStrength returns the adjusted ST for throwing purposes.
func (e *Entity) ThrowingStrength() fxp.Int {
	return e.derivedStrength(ThrowingStrengthID, e.ThrowingStrengthBonus)
}

// derivedStrength returns the floored sum of the bonus and the current value of the dedicated strength attribute with
// the given ID, or of ST (never less than 0) when the sheet doesn't define that attribute.
func (e *Entity) derivedStrength(attrID string, bonus fxp.Int) fxp.Int {
	var st fxp.Int
	if e.ResolveAttribute(attrID) != nil {
		st = e.ResolveAttributeCurrent(attrID)
	} else {
		st = e.ResolveAttributeCurrent(StrengthID).Max(0)
	}
	return (st + bonus).Floor()
}

// TelekineticStrength returns the total telekinetic strength.
func (e *Entity) TelekineticStrength() fxp.Int {
	levels, _ := e.TraitLevels("telekinesis")
	return levels.Floor()
}

// TraitLevels returns the sum of the current levels of every enabled, leveled trait whose name matches the given one,
// ignoring case, along with whether any such trait exists.
func (e *Entity) TraitLevels(name string) (levels fxp.Int, found bool) {
	Traverse(func(t *Trait) bool {
		if t.IsLeveled() && strings.EqualFold(t.NameWithReplacements(), name) {
			levels += t.CurrentLevel()
			found = true
		}
		return false
	}, true, false, e.Traits...)
	return levels, found
}

// Thrust returns the thrust value for the current strength.
func (e *Entity) Thrust() dice.Dice {
	return e.ThrustFor(e.StrikingStrength().AsInteger[int]())
}

// LiftingThrust returns the lifting thrust value for the current strength.
func (e *Entity) LiftingThrust() dice.Dice {
	return e.ThrustFor(e.LiftingStrength().AsInteger[int]())
}

// IQThrust returns the IQ thrust value for the current intelligence.
func (e *Entity) IQThrust() dice.Dice {
	return e.ThrustFor(e.ResolveAttributeCurrent(IntelligenceID).AsInteger[int]())
}

// TelekineticThrust returns the telekinetic thrust value for the current telekinesis level.
func (e *Entity) TelekineticThrust() dice.Dice {
	return e.ThrustFor(e.TelekineticStrength().AsInteger[int]())
}

// ThrustFor returns the thrust value for the provided strength.
func (e *Entity) ThrustFor(st int) dice.Dice {
	return e.SheetSettings.DamageProgression.Thrust(st)
}

// Swing returns the swing value for the current strength.
func (e *Entity) Swing() dice.Dice {
	return e.SwingFor(e.StrikingStrength().AsInteger[int]())
}

// LiftingSwing returns the lifting swing value for the current strength.
func (e *Entity) LiftingSwing() dice.Dice {
	return e.SwingFor(e.LiftingStrength().AsInteger[int]())
}

// IQSwing returns the IQ swing value for the current intelligence.
func (e *Entity) IQSwing() dice.Dice {
	return e.SwingFor(e.ResolveAttributeCurrent(IntelligenceID).AsInteger[int]())
}

// TelekineticSwing returns the telekinetic swing value for the current telekinesis level.
func (e *Entity) TelekineticSwing() dice.Dice {
	return e.SwingFor(e.TelekineticStrength().AsInteger[int]())
}

// SwingFor returns the swing value for the provided strength.
func (e *Entity) SwingFor(st int) dice.Dice {
	return e.SheetSettings.DamageProgression.Swing(st)
}

// AttributeBonusFor returns the bonus for the given attribute.
func (e *Entity) AttributeBonusFor(attributeID string, limitation stlimit.Option, tooltip *xbytes.InsertBuffer) fxp.Int {
	var total fxp.Int
	for _, one := range e.features.attributeBonuses {
		if one.ActualLimitation() == limitation && one.Attribute == attributeID {
			total += one.AdjustedAmount()
			one.AddToTooltip(tooltip)
		}
	}
	return total
}

// CostReductionFor returns the total cost reduction for the given ID.
func (e *Entity) CostReductionFor(attributeID string) fxp.Int {
	var total fxp.Int
	for _, one := range e.features.costReductions {
		if one.Attribute == attributeID {
			total += one.Percentage
		}
	}
	if total > fxp.Eighty {
		total = fxp.Eighty
	}
	return total.Max(0)
}

// sumBonuses returns the total adjusted amount of the bonuses in the list that match, adding each of them to the
// tooltip, which may be nil. match receives the bonus along with the nameable replacements of its owner.
func sumBonuses[T Bonus](list []T, tooltip *xbytes.InsertBuffer, match func(bonus T, replacements map[string]string) bool) fxp.Int {
	var total fxp.Int
	for _, bonus := range list {
		if match(bonus, bonusReplacements(bonus)) {
			total += bonus.AdjustedAmount()
			bonus.AddToTooltip(tooltip)
		}
	}
	return total
}

// collectBonuses returns the bonuses in the list that match, adding each of them to the tooltip, which may be nil.
// match receives the bonus along with the nameable replacements of its owner.
func collectBonuses[T Bonus](list []T, tooltip *xbytes.InsertBuffer, match func(bonus T, replacements map[string]string) bool) []T {
	var result []T
	for _, bonus := range list {
		if match(bonus, bonusReplacements(bonus)) {
			result = append(result, bonus)
			bonus.AddToTooltip(tooltip)
		}
	}
	return result
}

// TraitMaxLevelBonusesFor returns the "traits whose name" max-level bonuses that match the given name and tags.
func (e *Entity) TraitMaxLevelBonusesFor(name string, tags []string, tooltip *xbytes.InsertBuffer) []*TraitMaxLevelBonus {
	return collectBonuses(e.features.traitMaxLevelBonuses, tooltip,
		func(bonus *TraitMaxLevelBonus, replacements map[string]string) bool {
			return bonus.SelectionType == traitsel.TraitWithName &&
				bonus.NameCriteria.Matches(replacements, name) &&
				bonus.TagsCriteria.MatchesList(replacements, tags...)
		})
}

// EquipmentMaxUsesBonusesFor returns the "equipment whose name" max-uses bonuses that match the given name and tags.
func (e *Entity) EquipmentMaxUsesBonusesFor(name string, tags []string, tooltip *xbytes.InsertBuffer) []*EquipmentMaxUsesBonus {
	return collectBonuses(e.features.maxUsesBonuses, tooltip,
		func(bonus *EquipmentMaxUsesBonus, replacements map[string]string) bool {
			return bonus.SelectionType == equipmentsel.EquipmentWithName &&
				bonus.NameCriteria.Matches(replacements, name) &&
				bonus.TagsCriteria.MatchesList(replacements, tags...)
		})
}

// AddDRBonusesFor locates any active DR bonuses and adds them to the map. If 'drMap' is nil, it will be created. The
// provided map (or the newly created one) will be returned.
func (e *Entity) AddDRBonusesFor(locationID string, tooltip *xbytes.InsertBuffer, drMap map[string]int) map[string]int {
	return e.addDRBonusesFor(locationID, tooltip, drMap, false)
}

// AddArmorDRBonusesFor locates the active DR bonuses that come from worn armor and adds them to the map. If 'drMap' is
// nil, it will be created. The provided map (or the newly created one) will be returned.
//
// The falling rules (BX431) count all armor DR as flexible for the purpose of blunt trauma, while innate DR -- a hit
// location's own DR, or DR granted by a trait, skill or spell -- does not stop a fall that way at all, so the two have
// to be told apart. A DR bonus counts as armor when the item it came from is a piece of equipment.
func (e *Entity) AddArmorDRBonusesFor(locationID string, drMap map[string]int) map[string]int {
	return e.addDRBonusesFor(locationID, nil, drMap, true)
}

func (e *Entity) addDRBonusesFor(locationID string, tooltip *xbytes.InsertBuffer, drMap map[string]int, armorOnly bool) map[string]int {
	if drMap == nil {
		drMap = make(map[string]int)
	}
	isTopLevel := false
	for _, one := range e.SheetSettings.BodyType.Locations {
		if one.LocID == locationID {
			isTopLevel = true
			break
		}
	}
	for _, one := range e.features.drBonuses {
		if armorOnly {
			if _, ok := one.Owner().(*Equipment); !ok {
				continue
			}
		}
		for _, loc := range one.Locations {
			if (loc == AllID && isTopLevel) || strings.EqualFold(loc, locationID) {
				drMap[strings.ToLower(one.SpecializationWithReplacements())] += one.AdjustedAmount().AsInteger[int]()
				one.AddToTooltip(tooltip)
				break
			}
		}
	}
	return drMap
}

// SkillBonusFor returns the total bonus for the matching skill bonuses.
func (e *Entity) SkillBonusFor(name, specialization, optionalSpecialization string, tags []string, tooltip *xbytes.InsertBuffer) fxp.Int {
	return sumBonuses(e.features.skillBonuses, tooltip, func(bonus *SkillBonus, replacements map[string]string) bool {
		return bonus.SelectionType == skillsel.Name &&
			bonus.NameCriteria.Matches(replacements, name) &&
			bonus.SpecializationCriteria.Matches(replacements, specialization) &&
			bonus.OptionalSpecializationCriteria.Matches(replacements, optionalSpecialization) &&
			bonus.TagsCriteria.MatchesList(replacements, tags...)
	})
}

// SkillPointBonusFor returns the total point bonus for the matching skill point bonuses.
func (e *Entity) SkillPointBonusFor(name, specialization, optionalSpecialization string, tags []string, tooltip *xbytes.InsertBuffer) fxp.Int {
	return sumBonuses(e.features.skillPointBonuses, tooltip,
		func(bonus *SkillPointBonus, replacements map[string]string) bool {
			return bonus.NameCriteria.Matches(replacements, name) &&
				bonus.SpecializationCriteria.Matches(replacements, specialization) &&
				bonus.OptionalSpecializationCriteria.Matches(replacements, optionalSpecialization) &&
				bonus.TagsCriteria.MatchesList(replacements, tags...)
		})
}

// spellMatcher is implemented by the bonuses that select spells by name, power source, college or tag.
type spellMatcher interface {
	Bonus
	MatchesSpell(replacements map[string]string, name, powerSource string, colleges, tags []string) bool
}

// sumSpellBonuses returns the total adjusted amount of the bonuses in the list that match the given spell, adding each
// of them to the tooltip, which may be nil.
func sumSpellBonuses[T spellMatcher](list []T, name, powerSource string, colleges, tags []string, tooltip *xbytes.InsertBuffer) fxp.Int {
	return sumBonuses(list, tooltip, func(bonus T, replacements map[string]string) bool {
		return bonus.MatchesSpell(replacements, name, powerSource, colleges, tags)
	})
}

// SpellBonusFor returns the total bonus for the matching spell bonuses.
func (e *Entity) SpellBonusFor(name, powerSource string, colleges, tags []string, tooltip *xbytes.InsertBuffer) fxp.Int {
	return sumSpellBonuses(e.features.spellBonuses, name, powerSource, colleges, tags, tooltip)
}

// SpellPointBonusFor returns the total point bonus for the matching spell point bonuses.
func (e *Entity) SpellPointBonusFor(name, powerSource string, colleges, tags []string, tooltip *xbytes.InsertBuffer) fxp.Int {
	return sumSpellBonuses(e.features.spellPointBonuses, name, powerSource, colleges, tags, tooltip)
}

// TraitBonusFor returns the total bonus for the matching trait bonuses.
func (e *Entity) TraitBonusFor(name string, tags []string, tooltip *xbytes.InsertBuffer) fxp.Int {
	return sumBonuses(e.features.traitBonuses, tooltip, func(bonus *TraitBonus, replacements map[string]string) bool {
		return bonus.NameCriteria.Matches(replacements, name) &&
			bonus.TagsCriteria.MatchesList(replacements, tags...)
	})
}

// AddWeaponWithSkillBonusesFor adds the bonuses for matching weapons to the map. If 'm' is nil, it will be created.
// The provided map (or the newly created one) will be returned.
func (e *Entity) AddWeaponWithSkillBonusesFor(name, specialization, usage string, tags []string, dieCount dieCountFunc, tooltip *xbytes.InsertBuffer, m map[*WeaponBonus]bool, allowedFeatureTypes map[feature.Type]bool) map[*WeaponBonus]bool {
	if m == nil {
		m = make(map[*WeaponBonus]bool)
	}
	rsl := fxp.Min
	for _, sk := range e.SkillNamed(name, specialization, true, nil) {
		if rsl < sk.LevelData.RelativeLevel {
			rsl = sk.LevelData.RelativeLevel
		}
	}
	for _, bonus := range e.features.weaponBonuses {
		if allowedFeatureTypes[bonus.Type] &&
			bonus.SelectionType == wsel.WithRequiredSkill &&
			bonus.RelativeLevelCriteria.Matches(rsl) {
			replacements := bonusReplacements(bonus)
			if bonus.NameCriteria.Matches(replacements, name) &&
				bonus.SpecializationCriteria.Matches(replacements, specialization) &&
				bonus.UsageCriteria.Matches(replacements, usage) &&
				bonus.TagsCriteria.MatchesList(replacements, tags...) {
				addWeaponBonusToMap(bonus, dieCount, tooltip, m)
			}
		}
	}
	return m
}

// AddNamedWeaponBonusesFor adds the bonuses for matching weapons to the map. If 'm' is nil, it will be created. The
// provided map (or the newly created one) will be returned.
func (e *Entity) AddNamedWeaponBonusesFor(nameQualifier, usageQualifier string, tagsQualifier []string, dieCount dieCountFunc, tooltip *xbytes.InsertBuffer, m map[*WeaponBonus]bool, allowedFeatureTypes map[feature.Type]bool) map[*WeaponBonus]bool {
	if m == nil {
		m = make(map[*WeaponBonus]bool)
	}
	for _, bonus := range e.features.weaponBonuses {
		if allowedFeatureTypes[bonus.Type] &&
			bonus.SelectionType == wsel.WithName {
			replacements := bonusReplacements(bonus)
			if bonus.NameCriteria.Matches(replacements, nameQualifier) &&
				bonus.SpecializationCriteria.Matches(replacements, usageQualifier) &&
				bonus.TagsCriteria.MatchesList(replacements, tagsQualifier...) {
				addWeaponBonusToMap(bonus, dieCount, tooltip, m)
			}
		}
	}
	return m
}

func addWeaponBonusToMap(bonus *WeaponBonus, dieCount dieCountFunc, tooltip *xbytes.InsertBuffer, m map[*WeaponBonus]bool) {
	if m[bonus] {
		return
	}
	if tooltip != nil {
		bonus.addToTooltip(bonus.resolveDieCount(dieCount), bonus.DerivedLeveledOwner(), tooltip)
	}
	m[bonus] = true
}

// NamedWeaponSkillBonusesFor returns the bonuses for matching weapons.
func (e *Entity) NamedWeaponSkillBonusesFor(name, usage string, tags []string, tooltip *xbytes.InsertBuffer) []*SkillBonus {
	return collectBonuses(e.features.skillBonuses, tooltip, func(bonus *SkillBonus, replacements map[string]string) bool {
		return bonus.SelectionType == skillsel.WeaponsWithName &&
			bonus.NameCriteria.Matches(replacements, name) &&
			bonus.SpecializationCriteria.Matches(replacements, usage) &&
			bonus.TagsCriteria.MatchesList(replacements, tags...)
	})
}

// Move returns the current Move value for the given Encumbrance.
func (e *Entity) Move(enc encumbrance.Level) int {
	var initialMove fxp.Int
	if e.ResolveAttribute(MoveID) != nil {
		initialMove = e.ResolveAttributeCurrent(MoveID)
	} else {
		initialMove = e.ResolveAttributeCurrent(BasicMoveID).Max(0)
	}
	if divisor := 2 * min(CountThresholdOpMet(threshold.HalveMove, e.Attributes), 2); divisor > 0 {
		initialMove = initialMove.Div(fxp.FromInteger(divisor)).Ceil()
	}
	move := initialMove.Mul(fxp.Ten + fxp.Two.Mul(enc.Penalty())).Div(fxp.Ten).Floor()
	if move < fxp.One {
		if initialMove > 0 {
			return 1
		}
		return 0
	}
	return move.AsInteger[int]()
}

// BestSkillNamed returns the best skill that matches.
func (e *Entity) BestSkillNamed(name, specialization string, requirePoints bool, excludes map[string]bool) *Skill {
	var best *Skill
	level := fxp.Min
	for _, sk := range e.SkillNamed(name, specialization, requirePoints, excludes) {
		skillLevel := sk.CalculateLevel(excludes).Level
		if best == nil || level < skillLevel {
			best = sk
			level = skillLevel
		}
	}
	return best
}

// SkillNamed returns a list of skills that match by exact (case-insensitive) name and specialization.
func (e *Entity) SkillNamed(name, specialization string, requirePoints bool, excludes map[string]bool) []*Skill {
	var list []*Skill
	Traverse(func(sk *Skill) bool {
		if !excludes[sk.String()] {
			if !requirePoints || sk.IsTechnique() || sk.AdjustedPoints(nil) > 0 {
				if strings.EqualFold(sk.NameWithReplacements(), name) {
					if specialization == "" || strings.EqualFold(sk.SpecializationWithReplacements(), specialization) ||
						strings.EqualFold(sk.OptionalSpecializationWithReplacements(), specialization) {
						list = append(list, sk)
					}
				}
			}
		}
		return false
	}, false, true, e.Skills...)
	return list
}

// BestSkillIn returns the highest-level skill in the list, with the first one encountered winning a tie. The excludes
// are the ones the list was gathered with, and are passed along so that each candidate's level is calculated under the
// same terms.
func BestSkillIn(list []*Skill, excludes map[string]bool) *Skill {
	var best *Skill
	var level fxp.Int
	for _, sk := range list {
		skillLevel := sk.CalculateLevel(excludes).Level
		if best == nil || level < skillLevel {
			best = sk
			level = skillLevel
		}
	}
	return best
}

// skillSpecializationMatches reports whether a specialization criteria matches a skill that has the given required and
// optional specializations. A skill with a required specialization may be matched by either it or its optional
// specialization. A skill with no required specialization is matched by its optional specialization (which is the empty
// string for a fully unspecialized skill) - this keeps an unspecialized skill distinct from a sibling that carries an
// optional specialization, so e.g. an "is empty" criteria matches only the truly unspecialized skill.
func skillSpecializationMatches(specializationCriteria criteria.Text, replacements map[string]string, requiredSpecialization, optionalSpecialization string) bool {
	if requiredSpecialization != "" {
		return specializationCriteria.Matches(replacements, requiredSpecialization) ||
			(optionalSpecialization != "" && specializationCriteria.Matches(replacements, optionalSpecialization))
	}
	return specializationCriteria.Matches(replacements, optionalSpecialization)
}

// SkillMatching returns a list of skills whose name and specialization match the given criteria.
func (e *Entity) SkillMatching(nameCriteria, specializationCriteria criteria.Text, replacements map[string]string, requirePoints bool, excludes map[string]bool) []*Skill {
	var list []*Skill
	Traverse(func(sk *Skill) bool {
		if !excludes[sk.String()] {
			if !requirePoints || sk.IsTechnique() || sk.AdjustedPoints(nil) > 0 {
				if nameCriteria.Matches(replacements, sk.NameWithReplacements()) &&
					skillSpecializationMatches(specializationCriteria, replacements,
						sk.SpecializationWithReplacements(), sk.OptionalSpecializationWithReplacements()) {
					list = append(list, sk)
				}
			}
		}
		return false
	}, false, true, e.Skills...)
	return list
}

// Dodge returns the current Dodge value for the given Encumbrance.
func (e *Entity) Dodge(enc encumbrance.Level) int {
	var dodge fxp.Int
	if e.ResolveAttribute(DodgeID) != nil {
		dodge = e.ResolveAttributeCurrent(DodgeID)
	} else {
		dodge = e.ResolveAttributeCurrent(BasicSpeedID).Max(0) + fxp.Three
	}
	dodge += e.DodgeBonus
	divisor := 2 * min(CountThresholdOpMet(threshold.HalveDodge, e.Attributes), 2)
	if divisor > 0 {
		dodge = dodge.Div(fxp.FromInteger(divisor)).Ceil()
	}
	return (dodge + enc.Penalty()).Max(fxp.One).AsInteger[int]()
}

// EncumbranceLevel returns the current Encumbrance level.
func (e *Entity) EncumbranceLevel(forSkills bool) encumbrance.Level {
	if forSkills {
		if e.encumbranceLevelForSkillsCache != encumbrance.LastLevel+1 {
			return e.encumbranceLevelForSkillsCache
		}
	} else if e.encumbranceLevelCache != encumbrance.LastLevel+1 {
		return e.encumbranceLevelCache
	}
	carried := e.WeightCarried(forSkills)
	for _, one := range encumbrance.Levels {
		if carried <= e.MaximumCarry(one) {
			if forSkills {
				e.encumbranceLevelForSkillsCache = one
			} else {
				e.encumbranceLevelCache = one
			}
			return one
		}
	}
	if forSkills {
		e.encumbranceLevelForSkillsCache = encumbrance.ExtraHeavy
	} else {
		e.encumbranceLevelCache = encumbrance.ExtraHeavy
	}
	return encumbrance.ExtraHeavy
}

// WeightUnit returns the weight unit that should be used for display.
func (e *Entity) WeightUnit() fxp.WeightUnit {
	return e.SheetSettings.DefaultWeightUnits
}

// WeightCarried returns the carried weight.
func (e *Entity) WeightCarried(forSkills bool) fxp.Weight {
	var total fxp.Weight
	for _, one := range e.CarriedEquipment {
		total += one.ExtendedWeight(forSkills, e.SheetSettings.DefaultWeightUnits)
	}
	return total
}

// MaximumCarry returns the maximum amount the Entity can carry for the specified encumbrance level.
func (e *Entity) MaximumCarry(enc encumbrance.Level) fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(enc.WeightMultiplier()))
}

// OneHandedLift returns the one-handed lift value.
func (e *Entity) OneHandedLift() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Two))
}

// TwoHandedLift returns the two-handed lift value.
func (e *Entity) TwoHandedLift() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Eight))
}

// ShoveAndKnockOver returns the shove & knock over value.
func (e *Entity) ShoveAndKnockOver() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Twelve))
}

// RunningShoveAndKnockOver returns the running shove & knock over value.
func (e *Entity) RunningShoveAndKnockOver() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.TwentyFour))
}

// CarryOnBack returns the carry on back value.
func (e *Entity) CarryOnBack() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Fifteen))
}

// ShiftSlightly returns the shift slightly value.
func (e *Entity) ShiftSlightly() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Fifty))
}

// BasicLift returns the entity's Basic Lift.
func (e *Entity) BasicLift() fxp.Weight {
	if e.basicLiftCache != -1 {
		return e.basicLiftCache
	}
	e.basicLiftCache = e.BasicLiftForST(e.LiftingStrength())
	return e.basicLiftCache
}

// BasicLiftForST returns the entity's Basic Lift as if their base ST was the given value.
func (e *Entity) BasicLiftForST(st fxp.Int) fxp.Weight {
	st = st.Floor()
	if IsThresholdOpMet(threshold.HalveST, e.Attributes) {
		st = st.Div(fxp.Two)
		if st != st.Floor() {
			st = st.Floor() + fxp.One
		}
	}
	return BasicLiftForST(st, e.SheetSettings.DamageProgression)
}

// BasicLiftForST returns the Basic Lift for the given ST under the given damage progression (BX17), for a character
// that is not on a sheet. A sheet's character has its own BasicLiftForST, which also honors any threshold that halves
// its ST.
func BasicLiftForST(st fxp.Int, damageProgression progression.Option) fxp.Weight {
	st = st.Floor()
	if st < fxp.One {
		return 0
	}
	var v fxp.Int
	if damageProgression == progression.KnowingYourOwnStrength {
		var diff fxp.Int
		if st > fxp.Nineteen {
			diff = st.Div(fxp.Ten).Floor() - fxp.One
			st -= diff.Mul(fxp.Ten)
		}
		v = fxp.FromFloat(math.Pow(10, st.AsFloat[float64]()/10)).Mul(fxp.Two)
		if st <= fxp.Six {
			v = v.Mul(fxp.Ten).Round().Div(fxp.Ten)
		} else {
			v = v.Round()
		}
		v = v.Mul(fxp.FromFloat(math.Pow(10, diff.AsFloat[float64]())))
	} else {
		v = st.Mul(st).Div(fxp.Five)
	}
	if v >= fxp.Ten {
		v = v.Round()
	}
	return fxp.Weight(v.Mul(fxp.Ten).Floor().Div(fxp.Ten))
}

func (e *Entity) isSkillLevelResolutionExcluded(name, specialization, optionalSpecialization string) bool {
	if e.skillResolverExclusions[e.skillLevelResolutionKey(name, specialization, optionalSpecialization)] {
		args := []any{"name", name}
		if specialization != "" {
			args = append(args, "specialization", specialization)
		}
		if optionalSpecialization != "" {
			args = append(args, "optionalSpecialization", optionalSpecialization)
		}
		slog.Error("attempt to resolve skill level via itself", args...)
		return true
	}
	return false
}

func (e *Entity) registerSkillLevelResolutionExclusion(name, specialization, optionalSpecialization string) {
	e.skillResolverExclusions[e.skillLevelResolutionKey(name, specialization, optionalSpecialization)] = true
}

func (e *Entity) unregisterSkillLevelResolutionExclusion(name, specialization, optionalSpecialization string) {
	delete(e.skillResolverExclusions, e.skillLevelResolutionKey(name, specialization, optionalSpecialization))
}

func (e *Entity) skillLevelResolutionKey(name, specialization, optionalSpecialization string) string {
	return name + "\u0000" + specialization + "\u0000" + optionalSpecialization
}

// ResolveVariable resolves a variable to a value.
func (e *Entity) ResolveVariable(variableName string) string {
	if e.variableResolverExclusions[variableName] {
		slog.Error("attempt to resolve variable via itself", "name", variableName)
		return ""
	}
	if v, ok := e.variableCache[variableName]; ok {
		return v
	}
	e.variableResolverExclusions[variableName] = true
	defer func() { delete(e.variableResolverExclusions, variableName) }()
	if SizeModifierID == variableName {
		result := strconv.Itoa(e.Profile.AdjustedSizeModifier())
		e.variableCache[variableName] = result
		return result
	}
	parts := strings.SplitN(variableName, ".", 2)
	attr := e.Attributes.Set[parts[0]]
	if attr == nil {
		slog.Error("no such variable", "name", variableName)
		return ""
	}
	def := attr.AttributeDef()
	if def == nil {
		slog.Error("no such variable definition", "name", variableName)
		return ""
	}
	if (def.Type == attribute.Pool || def.Type == attribute.PoolRef) && len(parts) > 1 && parts[1] == "current" {
		result := attr.Current().String()
		e.variableCache[variableName] = result
		return result
	}
	result := attr.Maximum().String()
	e.variableCache[variableName] = result
	return result
}

// ResolveAttributeDef resolves the given attribute ID to its AttributeDef, or nil.
func (e *Entity) ResolveAttributeDef(attrID string) *AttributeDef {
	if e != nil {
		if a, ok := e.Attributes.Set[attrID]; ok {
			return a.AttributeDef()
		}
	}
	return nil
}

// ResolveAttributeName resolves the given attribute ID to its name, or <unknown>.
func (e *Entity) ResolveAttributeName(attrID string) string {
	if def := e.ResolveAttributeDef(attrID); def != nil {
		return def.Name
	}
	return i18n.Text("<unknown>")
}

// ResolveAttribute resolves the given attribute ID to its Attribute, or nil.
func (e *Entity) ResolveAttribute(attrID string) *Attribute {
	if e != nil {
		if a, ok := e.Attributes.Set[attrID]; ok {
			return a
		}
	}
	return nil
}

// ResolveAttributeCurrent resolves the given attribute ID to its current value, or fxp.Min.
func (e *Entity) ResolveAttributeCurrent(attrID string) fxp.Int {
	if e != nil {
		return e.Attributes.Current(attrID)
	}
	return fxp.Min
}

// PreservesUserDesc returns true if the user description widget should be preserved when written to disk. Normally,
// only character sheets should return true for this.
func (e *Entity) PreservesUserDesc() bool {
	return true
}

// Ancestry returns the current Ancestry.
func (e *Entity) Ancestry() *Ancestry {
	var anc *Ancestry
	Traverse(func(t *Trait) bool {
		if t.Container() && t.ContainerType == container.Ancestry && t.Enabled() {
			if anc = LookupAncestry(t.Ancestry, GlobalSettings().Libraries); anc != nil {
				return true
			}
		}
		return false
	}, true, false, e.Traits...)
	if anc == nil {
		if anc = LookupAncestry(DefaultAncestry, GlobalSettings().Libraries); anc == nil {
			// The default ancestry couldn't be loaded (e.g. a library file with the same name is present but contains
			// invalid data). Rather than crashing, log the problem and fall back to an empty ancestry so randomization
			// still produces sane defaults.
			errs.Log(errs.New("unable to load default ancestry (Human); using built-in defaults"))
			anc = &Ancestry{Name: DefaultAncestry}
		}
	}
	return anc
}

// WeaponOwner implements WeaponListProvider. In the case of an Entity, always returns nil, as entities rely on
// sub-components for their weapons and don't allow them to be created directly.
func (e *Entity) WeaponOwner() WeaponOwner {
	return nil
}

// Weapons implements WeaponListProvider.
func (e *Entity) Weapons(melee, includeUnequipped, excludeHidden bool) []*Weapon {
	m := make(map[uint64]*Weapon)
	Traverse(func(t *Trait) bool {
		for _, w := range t.Weapons {
			if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
				m[w.HashDisplayedState()] = w
			}
		}
		return false
	}, true, true, e.Traits...)
	Traverse(func(eqp *Equipment) bool {
		if eqp.Quantity > 0 && (includeUnequipped || eqp.ReallyEquipped()) {
			for _, w := range eqp.Weapons {
				if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
					w.NotEquipped = !eqp.ReallyEquipped()
					w.NotCarried = false
					m[w.HashDisplayedState()] = w
				}
			}
		}
		return false
	}, false, false, e.CarriedEquipment...)
	if includeUnequipped {
		Traverse(func(eqp *Equipment) bool {
			if eqp.Quantity > 0 {
				for _, w := range eqp.Weapons {
					if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
						w.NotEquipped = true
						w.NotCarried = true
						m[w.HashDisplayedState()] = w
					}
				}
			}
			return false
		}, false, false, e.OtherEquipment...)
	}
	Traverse(func(s *Skill) bool {
		for _, w := range s.Weapons {
			if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
				m[w.HashDisplayedState()] = w
			}
		}
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		for _, w := range s.Weapons {
			if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
				m[w.HashDisplayedState()] = w
			}
		}
		return false
	}, false, true, e.Spells...)
	list := make([]*Weapon, 0, len(m))
	for _, v := range m {
		list = append(list, v)
	}
	slices.SortFunc(list, func(a, b *Weapon) int { return a.Compare(b) })
	return list
}

// SetWeapons implements WeaponListProvider.
func (e *Entity) SetWeapons(_ bool, _ []*Weapon) {
	// Not permitted
}

// gatherConditionalModifiers walks the entity's active features, collecting conditional modifiers (or reactions) into
// the root rows of the table that shows them. namespace is that table's block key. collectFromList extracts the
// relevant bonuses from a feature list into the collector, and perTrait, if non-nil, contributes any additional
// modifiers for each enabled trait (used for self-control reaction penalties). keep, if non-nil, selects the modifiers
// to retain; it is applied before they are filed under their groups, so that a group emptied by it is dropped too.
// Reactions and ConditionalModifiers share this so their selection of nodes and merge ordering stay identical.
func (e *Entity) gatherConditionalModifiers(
	namespace string,
	collectFromList func(source string, features Features, c *condModCollector),
	perTrait func(source string, t *Trait, c *condModCollector),
	keep func(*ConditionalModifier) bool,
) []*ConditionalModifier {
	collector := newCondModCollector(e.ID, namespace)
	e.forEachActiveFeatureList(func(owner, _ fmt.Stringer, _ LeveledOwner, list Features) {
		collectFromList(conditionalModifierSource(owner), list, collector)
	})
	if perTrait != nil {
		Traverse(func(t *Trait) bool {
			perTrait(conditionalModifierSource(t), t, collector)
			return false
		}, true, false, e.Traits...)
	}
	return collector.rows(keep)
}

// conditionalModifierSource returns the text that names the owner of a conditional modifier or reaction in the list
// of sources that contributed to it.
func conditionalModifierSource(owner fmt.Stringer) string {
	switch actual := owner.(type) {
	case *Trait:
		return i18n.Text("from trait ") + actual.String()
	case *Skill:
		return i18n.Text("from skill ") + actual.String()
	case *Spell:
		return i18n.Text("from spell ") + actual.String()
	case *Equipment:
		return i18n.Text("from equipment ") + actual.NameWithReplacements()
	default:
		return i18n.Text("from ") + owner.String()
	}
}

// Reactions returns the current set of reactions. Those filed under a group are returned as the children of a container
// row named for the group.
func (e *Entity) Reactions() []*ConditionalModifier {
	return e.gatherConditionalModifiers(BlockReactionsKey, situationModifiersFromFeatureList[*ReactionBonus],
		func(source string, t *Trait, c *condModCollector) {
			resolvedSelfControl := t.ResolvedSelfControl(nil)
			if resolvedSelfControl != selfctrl.None && t.ResolvedSelfControlAdjustment(nil) == selfctrl.ReactionPenalty {
				// The self-control penalty is derived from the trait rather than from a bonus the user wrote, so there
				// is no group for it to be filed under.
				c.add(source, "", fmt.Sprintf(i18n.Text("from others when %s is triggered"), t.String()),
					fxp.FromInteger(selfctrl.ReactionPenalty.Adjustment(resolvedSelfControl)))
			}
		}, nil)
}

// ConditionalModifiers returns the current set of conditional modifiers. Those filed under a group are returned as the
// children of a container row named for the group. If the sheet settings have HideZeroValueConditionalMods enabled,
// modifiers whose amounts total to zero are omitted, and a group left with no members by that is omitted along with
// them.
func (e *Entity) ConditionalModifiers() []*ConditionalModifier {
	var keep func(*ConditionalModifier) bool
	if SheetSettingsFor(e).HideZeroValueConditionalMods {
		keep = func(c *ConditionalModifier) bool { return c.Total() != 0 }
	}
	return e.gatherConditionalModifiers(BlockConditionalModifiersKey,
		situationModifiersFromFeatureList[*ConditionalModifierBonus], nil, keep)
}

// TraitList implements ListProvider
func (e *Entity) TraitList() []*Trait {
	return e.Traits
}

// HasTraitNamed returns true if the entity has an enabled trait whose name matches the given name, ignoring case and
// surrounding whitespace. Returns false if the entity is nil or the name is empty.
func (e *Entity) HasTraitNamed(name string) bool {
	if e == nil {
		return false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	found := false
	// onlyEnabled=true also skips traits nested under a disabled container.
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(strings.TrimSpace(t.NameWithReplacements()), name) {
			found = true
			return true
		}
		return false
	}, true, false, e.Traits...)
	return found
}

// SetTraitList implements ListProvider
func (e *Entity) SetTraitList(list []*Trait) {
	SetDataOwnerAll(e, list)
	e.Traits = list
}

// CarriedEquipmentList implements ListProvider
func (e *Entity) CarriedEquipmentList() []*Equipment {
	return e.CarriedEquipment
}

// SetCarriedEquipmentList implements ListProvider
func (e *Entity) SetCarriedEquipmentList(list []*Equipment) {
	SetDataOwnerAll(e, list)
	e.CarriedEquipment = list
}

// OtherEquipmentList implements ListProvider
func (e *Entity) OtherEquipmentList() []*Equipment {
	return e.OtherEquipment
}

// SetOtherEquipmentList implements ListProvider
func (e *Entity) SetOtherEquipmentList(list []*Equipment) {
	SetDataOwnerAll(e, list)
	e.OtherEquipment = list
}

// SkillList implements ListProvider
func (e *Entity) SkillList() []*Skill {
	return e.Skills
}

// SetSkillList implements ListProvider
func (e *Entity) SetSkillList(list []*Skill) {
	SetDataOwnerAll(e, list)
	e.Skills = list
}

// SpellList implements ListProvider
func (e *Entity) SpellList() []*Spell {
	return e.Spells
}

// SetSpellList implements ListProvider
func (e *Entity) SetSpellList(list []*Spell) {
	SetDataOwnerAll(e, list)
	e.Spells = list
}

// NoteList implements ListProvider
func (e *Entity) NoteList() []*Note {
	return e.Notes
}

// SetNoteList implements ListProvider
func (e *Entity) SetNoteList(list []*Note) {
	SetDataOwnerAll(e, list)
	e.Notes = list
}

// Hash writes this object's contents into the hasher.
func (e *Entity) Hash(h hash.Hash) {
	saved := e.ModifiedOn
	e.ModifiedOn = jio.Time{}
	defer func() { e.ModifiedOn = saved }()
	HashJSON(h, e)
}

// SetPointsRecord sets a new points record list, adjusting the total points.
func (e *Entity) SetPointsRecord(record []*PointsRecord) {
	e.PointsRecord = ClonePointsRecordList(record)
	SortPointsRecordList(e.PointsRecord)
	e.TotalPoints = 0
	for _, rec := range record {
		e.TotalPoints += rec.Points
	}
}

// SyncWithLibrarySources syncs the entity with the library sources.
func (e *Entity) SyncWithLibrarySources() {
	syncWithLibrarySources(e)
}

// PageSettings implements PageInfoProvider.
func (e *Entity) PageSettings() *PageSettings {
	return e.SheetSettings.Page
}

// PageTitle implements PageInfoProvider.
func (e *Entity) PageTitle() string {
	if e.SheetSettings.UseTitleInFooter {
		return e.Profile.Title
	}
	return e.Profile.Name
}

// ModifiedOnString implements PageInfoProvider.
func (e *Entity) ModifiedOnString() string {
	return e.ModifiedOn.String()
}

// PageKeywords implements PageInfoProvider.
func (e *Entity) PageKeywords() string {
	return "GCS Character Sheet"
}
