// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculator

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// HikingExtraEffortFP is what extra effort on the march adds to the FP lost when the hiker stops (BX357).
const HikingExtraEffortFP = 2

// hikingRestMinutesPerFP is how long a rest takes to recover 1 FP for someone with no help (BX427).
const hikingRestMinutesPerFP = 10

// TerrainModifier is a kind of ground a day's hike can cross, or the weather it is made in, and what it does to the
// distance (BX351).
type TerrainModifier struct {
	Name           string
	Modifier       fxp.Int // What the distance is multiplied by.
	ModifierInRain fxp.Int // What replaces Modifier for a road in the rain; zero when the rain changes nothing.
	IsRoad         bool
	IsRain         bool
	IsSnow         bool
	IsIce          bool
	Default        bool // Whether this is the choice a calculator starts on.
}

// String implements fmt.Stringer.
func (t TerrainModifier) String() string {
	return t.Name
}

// TerrainChoices returns the kinds of ground a day's hike can cross and what each does to the distance (BX351).
func TerrainChoices() []TerrainModifier {
	return []TerrainModifier{
		{Name: i18n.Text("Broken Ground"), Modifier: fxp.Half},
		{Name: i18n.Text("Deep Snow"), Modifier: fxp.Fifth, IsSnow: true},
		{Name: i18n.Text("Desert"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Desert, Hard-packed"), Modifier: fxp.OneAndAQuarter},
		{Name: i18n.Text("Forest"), Modifier: fxp.Half},
		{Name: i18n.Text("Forest, Dense"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Forest, Light"), Modifier: fxp.One},
		{Name: i18n.Text("Frozen Lake"), Modifier: fxp.Half, IsIce: true},
		{Name: i18n.Text("Frozen River"), Modifier: fxp.Half, IsIce: true},
		{Name: i18n.Text("Hills, Rolling"), Modifier: fxp.One},
		{Name: i18n.Text("Hills, Steep"), Modifier: fxp.Half},
		{Name: i18n.Text("Jungle"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Mountains"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Mud"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Plains, Level"), Modifier: fxp.OneAndAQuarter},
		{Name: i18n.Text("Road, Cobblestone"), Modifier: fxp.One, IsRoad: true},
		{Name: i18n.Text("Road, Dirt"), Modifier: fxp.One, ModifierInRain: fxp.Fifth, IsRoad: true, Default: true},
		{Name: i18n.Text("Road, Gravel"), Modifier: fxp.One, ModifierInRain: fxp.Fifth, IsRoad: true},
		{Name: i18n.Text("Road, Paved"), Modifier: fxp.OneAndAQuarter, ModifierInRain: fxp.One, IsRoad: true},
		{Name: i18n.Text("Sand"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Sand, Hard-packed"), Modifier: fxp.OneAndAQuarter},
		{Name: i18n.Text("Swamp"), Modifier: fxp.Fifth},
	}
}

// WeatherChoices returns the weather a day's hike can be made in and what each does to the distance (BX351).
func WeatherChoices() []TerrainModifier {
	return []TerrainModifier{
		{Name: i18n.Text("Normal"), Modifier: fxp.One, Default: true},
		{Name: i18n.Text("Rain"), Modifier: fxp.Half, IsRain: true},
		{Name: i18n.Text("Sleet"), Modifier: fxp.Half, IsIce: true},
		{Name: i18n.Text("Snow"), Modifier: fxp.Half, IsSnow: true},
		{Name: i18n.Text("Snow, Heavy"), Modifier: fxp.Quarter, IsSnow: true},
	}
}

// HikingIntensity is how hard a day's hike is pushed, as the hours it spends on the march.
type HikingIntensity struct {
	Name        string
	HoursHiking fxp.Int
	IsCustom    bool // Whether the hours are typed in instead.
	IsForaging  bool // Whether the rest of the day goes to foraging (BX427).
	Default     bool // Whether this is the choice a calculator starts on.
}

// String implements fmt.Stringer.
func (h HikingIntensity) String() string {
	return h.Name
}

// HikingIntensityChoices returns how hard the day's hike can be pushed, as the hours it spends on the march.
func HikingIntensityChoices() []HikingIntensity {
	return []HikingIntensity{
		{Name: i18n.Text("Forced March"), HoursHiking: fxp.Sixteen},
		{Name: i18n.Text("Long March"), HoursHiking: fxp.Twelve},
		{Name: i18n.Text("Normal"), HoursHiking: fxp.Eight, Default: true},
		{Name: i18n.Text("Foraging"), HoursHiking: fxp.Four, IsForaging: true},
		{Name: i18n.Text("Custom"), IsCustom: true},
	}
}

// HikingHeat is what the day's heat adds to each hour's FP cost (BX426).
type HikingHeat struct {
	Name string
	FP   int
}

// String implements fmt.Stringer.
func (h HikingHeat) String() string {
	return h.Name
}

// HikingHeatChoices returns what the day's heat can add to each hour's FP cost (BX426).
func HikingHeatChoices() []HikingHeat {
	return []HikingHeat{
		{Name: i18n.Text("Temperate")},
		{Name: i18n.Text("Hot (+1 FP per hour)"), FP: 1},
		{Name: i18n.Text("Hot, in plate armor, an overcoat, etc. (+2 FP per hour)"), FP: 2},
	}
}

// HikingFitness is how fit the hiker is (BX55): Fit and Very Fit recover FP at twice the usual rate, and Very Fit
// loses FP to exertion at half the usual rate.
type HikingFitness struct {
	Name    string
	Fit     bool
	VeryFit bool
}

// String implements fmt.Stringer.
func (h HikingFitness) String() string {
	return h.Name
}

// HikingFitnessChoices returns how fit the hiker can be (BX55), in the order HikerStats.FitnessIndex counts them.
func HikingFitnessChoices() []HikingFitness {
	return []HikingFitness{
		{Name: i18n.Text("Average")},
		{Name: i18n.Text("Fit (recovers FP twice as fast)"), Fit: true},
		{Name: i18n.Text("Very Fit (recovers FP twice as fast, loses FP half as fast)"), Fit: true, VeryFit: true},
	}
}

// HikingRecoverEnergy is what the Recover Energy spell (BX248) does for the hiker's rest, by the skill it is known at.
type HikingRecoverEnergy struct {
	Name         string
	MinutesPerFP int // 0 when the spell is no help.
}

// String implements fmt.Stringer.
func (h HikingRecoverEnergy) String() string {
	return h.Name
}

// HikingRecoverEnergyChoices returns what the Recover Energy spell (BX248) can do for the hiker's rest, in the order
// HikerStats.RecoverEnergyIndex counts them.
func HikingRecoverEnergyChoices() []HikingRecoverEnergy {
	return []HikingRecoverEnergy{
		{Name: i18n.Text("None")},
		{Name: i18n.Text("Skill 15+ (1 FP per 5 minutes of rest)"), MinutesPerFP: 5},
		{Name: i18n.Text("Skill 20+ (1 FP per 2 minutes of rest)"), MinutesPerFP: 2},
	}
}

// Hiker is the character making a day's hike (BX351, BX426).
type Hiker struct {
	Move          int               // Move, with encumbrance.
	EnhancedMove  fxp.Int           // Levels of Enhanced Move (Ground).
	Encumbrance   encumbrance.Level // Each level adds 1 FP to the cost of each hour.
	FP            fxp.Int           // The hiker's FP, or zero or less when unknown, in which case the day assumes full pace throughout.
	Fitness       HikingFitness
	RecoverEnergy HikingRecoverEnergy
}

// RestMinutesPerFP returns how long the hiker's rest takes to recover 1 FP: 10 minutes (BX427), or 5 for a Fit or Very
// Fit hiker (BX55), or what Recover Energy gives if that is faster (BX248).
func (h *Hiker) RestMinutesPerFP() int {
	minutes := hikingRestMinutesPerFP
	if h.Fitness.Fit {
		minutes /= 2
	}
	if spell := h.RecoverEnergy.MinutesPerFP; spell > 0 && spell < minutes {
		minutes = spell
	}
	return minutes
}

// condition returns the state the hiker is in with the given FP left (BX426).
func (h *Hiker) condition(fpLeft fxp.Int) HikerCondition {
	switch {
	case fpLeft <= -h.FP:
		return UnconsciousHiker
	case fpLeft <= 0:
		return ExhaustedHiker
	case fpLeft.Mul(fxp.Three) < h.FP:
		return VeryTiredHiker
	default:
		return FreshHiker
	}
}

// Hike is a day of travel: the hiker, the ground and weather, how hard the day is pushed, and the rest and extra
// effort that go into it (BX351, BX426).
type Hike struct {
	Hiker              Hiker
	Terrain            TerrainModifier
	Weather            TerrainModifier
	Heat               HikingHeat
	Hours              fxp.Int // Hours of hiking in the day.
	RestMinutes        int     // Minutes of rest halfway through the day; 0 for none.
	RestExtraFP        int     // FP restored by Lend Energy, potions and the like while resting.
	ExtraEffortPenalty int     // The penalty taken on the roll for extra effort, 0 or less; only a SuccessfulRoll uses it.
	RestMeal           bool    // Whether a decent meal is eaten while resting (+1 FP).
	UsingSkis          bool
	UsingSkates        bool
	RoadsAreCleared    bool // Whether a road has been cleared of its snow or ice.
	SuccessfulRoll     bool // Whether the Hiking, Skiing or Skating roll succeeded, for +20% to the distance.
	Metric             bool // Whether distances are reported in GURPS kilometers rather than miles.
}

// ExtraEffort reports whether the hiker is putting in extra effort, which takes a successful roll and a penalty on it.
func (h *Hike) ExtraEffort() bool {
	return h.SuccessfulRoll && h.ExtraEffortPenalty < 0
}

// RoadsCanBeCleared reports whether the hike is on a road under snow or ice, the only ground that can be cleared.
func (h *Hike) RoadsCanBeCleared() bool {
	return h.Terrain.IsRoad && (h.Weather.IsIce || h.Weather.IsSnow)
}

// FPPerHour returns what each hour of the march costs in FP (BX426): 1, plus 1 per level of encumbrance, plus what the
// heat adds, halved for a Very Fit hiker (BX55).
func (h *Hike) FPPerHour() fxp.Int {
	perHour := fxp.FromInteger(1 + int(h.Hiker.Encumbrance) + h.Heat.FP)
	if h.Hiker.Fitness.VeryFit {
		perHour = perHour.Div(fxp.Two)
	}
	return perHour
}

// RestAfter returns how many hours into the march the hiker rests, or 0 when there is no rest: halfway through, at the
// end of a whole hour, and only when some of the march is still to come.
func (h *Hike) RestAfter() fxp.Int {
	if h.RestMinutes <= 0 {
		return 0
	}
	after := h.Hours.Div(fxp.Two).Floor().Max(fxp.One)
	if after >= h.Hours {
		return 0
	}
	return after
}

// restRecovery returns what the rest gives back to a hiker with the given FP left: 1 FP per RestMinutesPerFP of it,
// plus 1 for a decent meal and whatever spells and the like restore, but never more than was lost (BX427).
func (h *Hike) restRecovery(fpLeft fxp.Int) fxp.Int {
	recovered := h.RestMinutes / h.Hiker.RestMinutesPerFP()
	if h.RestMeal {
		recovered++
	}
	recovered += h.RestExtraFP
	amount := fxp.FromInteger(recovered)
	if h.Hiker.FP > 0 {
		amount = amount.Min(h.Hiker.FP - fpLeft).Max(0)
	}
	return amount
}

// distanceForHours returns the distance covered in the given hours of travel at the hiker's full pace, in miles, or in
// GURPS kilometers when Metric (BX351).
func (h *Hike) distanceForHours(hours fxp.Int) fxp.Int {
	distance := fxp.FromInteger(h.Hiker.Move * 10)

	distance = distance.Mul(hours).Div(fxp.Sixteen)

	if h.Hiker.EnhancedMove > 0 {
		distance = distance.Mul(fxp.One + h.Hiker.EnhancedMove)
	}

	mod := h.Terrain.Modifier
	if h.Terrain.IsIce && h.UsingSkates {
		mod = fxp.OneAndAQuarter
	}
	if h.Terrain.IsSnow && h.UsingSkis {
		mod = fxp.One
	}

	switch {
	case h.Weather.IsRain:
		if h.Terrain.IsRoad {
			if h.Terrain.ModifierInRain != 0 {
				mod = h.Terrain.ModifierInRain
			}
		} else {
			mod = mod.Mul(h.Weather.Modifier)
		}
	case h.Weather.IsSnow:
		if h.Terrain.IsRoad {
			mod = fxp.One
		}
		if (!h.Terrain.IsRoad || !h.RoadsAreCleared) && !h.UsingSkis {
			mod = mod.Mul(h.Weather.Modifier)
		}
	case h.Weather.IsIce:
		if h.Terrain.IsRoad {
			mod = fxp.One
		}
		if (!h.Terrain.IsRoad || !h.RoadsAreCleared) && !h.UsingSkates {
			mod = mod.Mul(h.Weather.Modifier)
		}
	}
	distance = distance.Mul(mod)

	mod = fxp.One
	if h.SuccessfulRoll {
		mod = fxp.OnePointTwo
		if h.ExtraEffortPenalty < 0 {
			mod += fxp.FromInteger(-5 * h.ExtraEffortPenalty).Div(fxp.Hundred)
		}
	}
	distance = distance.Mul(mod)

	if h.Metric {
		// miles -> inches -> GURPS kilometers
		distance = fxp.Kilometer.FromInches(fxp.Mile.ToInches(distance))
	}
	return distance
}

// HikingHourKind is what a row of a day's hour-by-hour breakdown covers.
type HikingHourKind byte

// The possible HikingHourKind values.
const (
	MarchHour HikingHourKind = iota // An hour, or the part of one that ends the march, on the move.
	RestHour                        // The rest halfway through the day.
	StopHour                        // The stop at the end of the day, where extra effort takes its toll.
)

// HikerCondition is the state the FP lost so far leaves the hiker in (BX426).
type HikerCondition byte

// The possible HikerCondition values.
const (
	FreshHiker       HikerCondition = iota // A third or more of the FP are left.
	VeryTiredHiker                         // Fewer than a third of the FP are left: Move, Dodge and ST are halved.
	ExhaustedHiker                         // No FP are left: going on takes a Will roll, and each further FP lost costs 1 HP.
	UnconsciousHiker                       // The FP have reached their negative, and the hiker has fallen unconscious.
)

// Pace returns the fraction of the usual pace the hiker can keep up in this condition: half once fewer than a third
// of the FP are left, and none once unconscious (BX426).
func (c HikerCondition) Pace() fxp.Int {
	switch c {
	case VeryTiredHiker, ExhaustedHiker:
		return fxp.Half
	case UnconsciousHiker:
		return 0
	default:
		return fxp.One
	}
}

// HikingHour is one row of a day's hour-by-hour breakdown: what the hour covered, what it did to the hiker's FP, and
// the state it left the hiker in. The rows for the rest halfway through and for the stop at the end of the day cover
// no distance.
type HikingHour struct {
	Kind      HikingHourKind
	Condition HikerCondition // Meaningful only when the hiker's FP are known.
	Elapsed   fxp.Int        // Hours into the day at the end of the row.
	Distance  fxp.Int        // Covered in the hour.
	Total     fxp.Int        // Covered so far.
	FPChange  fxp.Int        // Negative for a loss.
	FPLeft    fxp.Int        // Meaningful only when the hiker's FP are known.
}

// HikingDay is what a day of travel comes to, worked out hour by hour with the effects of the FP lost applied as they
// arrive (BX426): Move is halved once fewer than a third of the hiker's FP are left, going on at 0 FP or less takes a
// Will roll, which the day assumes is made, and at -FP the hiker falls unconscious and the day ends. The hours are
// what the hiker can be expected to manage, and the cost is what they and the stop at the end come to. Distances are
// in miles, or in GURPS kilometers for a Metric hike.
type HikingDay struct {
	Hours            []HikingHour
	Distance         fxp.Int
	FPCost           fxp.Int
	RestRecovered    fxp.Int // What the rest halfway through gave back; 0 when there was no rest.
	TiredAfter       fxp.Int // Hours into the day the hiker became very tired; 0 if never.
	OutAfter         fxp.Int // Hours into the day the hiker ran out of FP; 0 if never.
	UnconsciousAfter fxp.Int // Hours into the day the hiker fell unconscious; 0 if never.
}

// Day works out the day of travel. Without FP to go on, the hiker is assumed to keep full pace all day.
func (h *Hike) Day() HikingDay {
	var day HikingDay
	hourly := h.distanceForHours(fxp.One)
	perHour := h.FPPerHour()
	known := h.Hiker.FP > 0
	fpLeft := h.Hiker.FP
	pace := fxp.One
	restAfter := h.RestAfter()
	var elapsed fxp.Int
	for remaining := h.Hours; remaining > 0; {
		step := remaining.Min(fxp.One)
		remaining -= step
		elapsed += step
		hour := HikingHour{
			Kind:     MarchHour,
			Elapsed:  elapsed,
			Distance: hourly.Mul(step).Mul(pace),
			FPChange: -perHour.Mul(step),
		}
		day.Distance += hour.Distance
		hour.Total = day.Distance
		day.FPCost -= hour.FPChange
		if known {
			// FP never fall below -FP; whatever would take them further comes off HP instead.
			fpLeft = (fpLeft + hour.FPChange).Max(-h.Hiker.FP)
			hour.FPLeft = fpLeft
			hour.Condition = h.Hiker.condition(fpLeft)
			pace = hour.Condition.Pace()
			if day.TiredAfter == 0 && pace < fxp.One {
				day.TiredAfter = elapsed
			}
			if day.OutAfter == 0 && fpLeft <= 0 {
				day.OutAfter = elapsed
			}
			if fpLeft <= -h.Hiker.FP {
				day.UnconsciousAfter = elapsed
				remaining = 0
			}
		}
		day.Hours = append(day.Hours, hour)
		if remaining > 0 && elapsed == restAfter {
			rest := HikingHour{Kind: RestHour, Elapsed: elapsed, FPChange: h.restRecovery(fpLeft)}
			day.RestRecovered = rest.FPChange
			if known {
				fpLeft += rest.FPChange
				rest.FPLeft = fpLeft
				rest.Condition = h.Hiker.condition(fpLeft)
				pace = rest.Condition.Pace()
			}
			day.Hours = append(day.Hours, rest)
		}
	}
	if h.ExtraEffort() {
		stop := HikingHour{Kind: StopHour, Elapsed: elapsed, FPChange: -fxp.FromInteger(HikingExtraEffortFP)}
		day.FPCost -= stop.FPChange
		if known {
			fpLeft = (fpLeft + stop.FPChange).Max(-h.Hiker.FP)
			stop.FPLeft = fpLeft
			stop.Condition = h.Hiker.condition(fpLeft)
			if day.UnconsciousAfter == 0 && fpLeft <= -h.Hiker.FP {
				day.UnconsciousAfter = elapsed
			}
		}
		day.Hours = append(day.Hours, stop)
	}
	return day
}

// HikingTimeInDays returns the number of days needed to cover distanceToCover while traveling distancePerDay each day,
// rounded to a tenth of a day. ok is false when distancePerDay is 0, as the travel time is then undefined (and
// fxp.Int.Div would panic).
func HikingTimeInDays(distanceToCover, distancePerDay fxp.Int) (days fxp.Int, ok bool) {
	if distancePerDay == 0 {
		return 0, false
	}
	return distanceToCover.Mul(fxp.Ten).Div(distancePerDay).Round().Div(fxp.Ten), true
}

// HikerStats is what a character sheet supplies about a hiker.
type HikerStats struct {
	Move               int               // Move at the current encumbrance.
	EnhancedMove       fxp.Int           // Levels of Enhanced Move (Ground).
	Encumbrance        encumbrance.Level // The current encumbrance.
	FP                 fxp.Int           // The maximum FP; meaningful only when HasFP.
	HasFP              bool              // Whether the sheet defines FP at all.
	FitnessIndex       int               // The index into HikingFitnessChoices of the character's fitness.
	RecoverEnergyIndex int               // The index into HikingRecoverEnergyChoices of the character's Recover Energy.
}

// HikerStatsFromEntity reads the hiker's numbers from the sheet's character. The entity must not be nil.
func HikerStatsFromEntity(entity *gurps.Entity) HikerStats {
	s := HikerStats{Encumbrance: entity.EncumbranceLevel(false)}
	s.Move = entity.Move(s.Encumbrance)
	s.EnhancedMove, _ = entity.TraitLevels("enhanced move (ground)")
	if entity.ResolveAttribute(gurps.FatiguePointsID) != nil {
		s.FP = entity.Attributes.Maximum(gurps.FatiguePointsID).Max(0)
		s.HasFP = true
	}
	switch {
	case entity.HasTraitNamed("Very Fit"):
		s.FitnessIndex = 2
	case entity.HasTraitNamed("Fit"):
		s.FitnessIndex = 1
	}
	if level := spellLevel(entity, "Recover Energy"); level >= 20 {
		s.RecoverEnergyIndex = 2
	} else if level >= 15 {
		s.RecoverEnergyIndex = 1
	}
	return s
}
