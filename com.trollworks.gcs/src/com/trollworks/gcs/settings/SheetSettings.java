/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.settings;

import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.body.HitLocationTable;
import com.trollworks.gcs.body.LibraryHitLocationTables;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.DisplayOption;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.ChangeNotifier;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.page.PageSettings;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class SheetSettings implements ChangeNotifier {
    public static final  String KEY_ATTRIBUTES                      = "attributes";
    private static final String KEY_BLOCK_LAYOUT                    = "block_layout";
    private static final String KEY_DEFAULT_LENGTH_UNITS            = "default_length_units";
    private static final String KEY_DEFAULT_WEIGHT_UNITS            = "default_weight_units";
    public static final String KEY_HIT_LOCATIONS                   = "hit_locations";
    private static final String KEY_MODIFIERS_DISPLAY               = "modifiers_display";
    private static final String KEY_NOTES_DISPLAY                   = "notes_display";
    private static final String KEY_PAGE                            = "page";
    private static final String KEY_SHOW_ADVANTAGE_MODIFIER_ADJ     = "show_advantage_modifier_adj";
    private static final String KEY_SHOW_COLLEGE_IN_SPELLS          = "show_college_in_sheet_spells";
    private static final String KEY_SHOW_DIFFICULTY                 = "show_difficulty";
    private static final String KEY_SHOW_EQUIPMENT_MODIFIER_ADJ     = "show_equipment_modifier_adj";
    private static final String KEY_SHOW_SPELL_ADJ                  = "show_spell_adj";
    private static final String KEY_USE_KNOW_YOUR_OWN_STRENGTH      = "use_know_your_own_strength";
    private static final String KEY_USE_MODIFYING_DICE_PLUS_ADDS    = "use_modifying_dice_plus_adds";
    private static final String KEY_USE_MULTIPLICATIVE_MODIFIERS    = "use_multiplicative_modifiers";
    private static final String KEY_USE_REDUCED_SWING               = "use_reduced_swing";
    private static final String KEY_USE_SIMPLE_METRIC_CONVERSIONS   = "use_simple_metric_conversions";
    private static final String KEY_USE_THRUST_EQUALS_SWING_MINUS_2 = "use_thrust_equals_swing_minus_2";
    private static final String KEY_USE_TITLE_IN_FOOTER             = "use_title_in_footer";
    private static final String KEY_USER_DESCRIPTION_DISPLAY        = "user_description_display";

    private GURPSCharacter            mCharacter;
    private LengthUnits               mDefaultLengthUnits;
    private WeightUnits               mDefaultWeightUnits;
    private List<String>              mBlockLayout;
    private DisplayOption             mUserDescriptionDisplay;
    private DisplayOption             mModifiersDisplay;
    private DisplayOption             mNotesDisplay;
    private Map<String, AttributeDef> mAttributes;
    private HitLocationTable          mHitLocations;
    private PageSettings              mPageSettings;
    private boolean                   mUseMultiplicativeModifiers; // P102
    private boolean                   mUseModifyingDicePlusAdds; // B269
    private boolean                   mUseKnowYourOwnStrength; // PY83
    private boolean                   mUseReducedSwing; // Adjusting Swing Damage from noschoolgrognard.blogspot.com
    private boolean                   mUseThrustEqualsSwingMinus2; // Home brew
    private boolean                   mUseSimpleMetricConversions; // B9
    private boolean                   mShowCollegeInSpells;
    private boolean                   mShowDifficulty;
    private boolean                   mShowAdvantageModifierAdj;
    private boolean                   mShowEquipmentModifierAdj;
    private boolean                   mShowSpellAdj;
    private boolean                   mUseTitleInFooter;

    /**
     * @param character The {@link GURPSCharacter} to retrieve settings for, or {@code null}.
     * @return The SheetSettings from the specified {@link GURPSCharacter} or from global
     *         preferences if the character is {@code null}.
     */
    public static SheetSettings get(GURPSCharacter character) {
        return character == null ? Settings.getInstance().getSheetSettings() : character.getSheetSettings();
    }

    /** Creates new default character sheet settings. */
    public SheetSettings() {
        reset();
    }

    /**
     * Creates new settings for a character sheet.
     *
     * @param character The {@link GURPSCharacter} these are for. Passing in {@code null} is the
     *                  same as calling {@link #SheetSettings()} instead.
     */
    public SheetSettings(GURPSCharacter character) {
        mCharacter = character;
        reset();
    }

    /** Reset these settings to their defaults. */
    public void reset() {
        if (mCharacter == null) {
            mDefaultLengthUnits = LengthUnits.FT_IN;
            mDefaultWeightUnits = WeightUnits.LB;
            mBlockLayout = new ArrayList<>(List.of(
                    CharacterSheet.REACTIONS_KEY + " " + CharacterSheet.CONDITIONAL_MODIFIERS_KEY,
                    CharacterSheet.MELEE_KEY,
                    CharacterSheet.RANGED_KEY,
                    CharacterSheet.ADVANTAGES_KEY + " " + CharacterSheet.SKILLS_KEY,
                    CharacterSheet.SPELLS_KEY,
                    CharacterSheet.EQUIPMENT_KEY,
                    CharacterSheet.OTHER_EQUIPMENT_KEY,
                    CharacterSheet.NOTES_KEY));
            mUserDescriptionDisplay = DisplayOption.TOOLTIP;
            mModifiersDisplay = DisplayOption.INLINE;
            mNotesDisplay = DisplayOption.INLINE;
            mAttributes = AttributeDef.createStandardAttributes();
            mHitLocations = LibraryHitLocationTables.getHumanoid().clone();
            mPageSettings = new PageSettings(this);
            mUseMultiplicativeModifiers = false;
            mUseModifyingDicePlusAdds = false;
            mUseKnowYourOwnStrength = false;
            mUseReducedSwing = false;
            mUseThrustEqualsSwingMinus2 = false;
            mUseSimpleMetricConversions = true;
            mShowCollegeInSpells = false;
            mShowDifficulty = false;
            mShowAdvantageModifierAdj = false;
            mShowEquipmentModifierAdj = false;
            mShowSpellAdj = true;
            mUseTitleInFooter = false;
        } else {
            SheetSettings defaults = Settings.getInstance().getSheetSettings();
            mDefaultLengthUnits = defaults.mDefaultLengthUnits;
            mDefaultWeightUnits = defaults.mDefaultWeightUnits;
            mBlockLayout = new ArrayList<>(defaults.mBlockLayout);
            mUserDescriptionDisplay = defaults.mUserDescriptionDisplay;
            mModifiersDisplay = defaults.mModifiersDisplay;
            mNotesDisplay = defaults.mNotesDisplay;
            mAttributes = AttributeDef.cloneMap(defaults.mAttributes);
            mHitLocations = defaults.mHitLocations.clone();
            mPageSettings = new PageSettings(this, defaults.mPageSettings);
            mUseMultiplicativeModifiers = defaults.mUseMultiplicativeModifiers;
            mUseModifyingDicePlusAdds = defaults.mUseModifyingDicePlusAdds;
            mUseKnowYourOwnStrength = defaults.mUseKnowYourOwnStrength;
            mUseReducedSwing = defaults.mUseReducedSwing;
            mUseThrustEqualsSwingMinus2 = defaults.mUseThrustEqualsSwingMinus2;
            mUseSimpleMetricConversions = defaults.mUseSimpleMetricConversions;
            mShowCollegeInSpells = defaults.mShowCollegeInSpells;
            mShowDifficulty = defaults.mShowDifficulty;
            mShowAdvantageModifierAdj = defaults.mShowAdvantageModifierAdj;
            mShowEquipmentModifierAdj = defaults.mShowEquipmentModifierAdj;
            mShowSpellAdj = defaults.mShowSpellAdj;
            mUseTitleInFooter = defaults.mUseTitleInFooter;
        }
    }

    public void load(JsonMap m, LoadState state) {
        reset();
        mDefaultLengthUnits = Enums.extract(m.getString(KEY_DEFAULT_LENGTH_UNITS), LengthUnits.values(), mDefaultLengthUnits);
        mDefaultWeightUnits = Enums.extract(m.getString(KEY_DEFAULT_WEIGHT_UNITS), WeightUnits.values(), mDefaultWeightUnits);
        mUserDescriptionDisplay = Enums.extract(m.getString(KEY_USER_DESCRIPTION_DISPLAY), DisplayOption.values(), mUserDescriptionDisplay);
        mModifiersDisplay = Enums.extract(m.getString(KEY_MODIFIERS_DISPLAY), DisplayOption.values(), mModifiersDisplay);
        mNotesDisplay = Enums.extract(m.getString(KEY_NOTES_DISPLAY), DisplayOption.values(), mNotesDisplay);
        mUseMultiplicativeModifiers = m.getBooleanWithDefault(KEY_USE_MULTIPLICATIVE_MODIFIERS, mUseMultiplicativeModifiers);
        mUseModifyingDicePlusAdds = m.getBooleanWithDefault(KEY_USE_MODIFYING_DICE_PLUS_ADDS, mUseModifyingDicePlusAdds);
        mUseKnowYourOwnStrength = m.getBooleanWithDefault(KEY_USE_KNOW_YOUR_OWN_STRENGTH, mUseKnowYourOwnStrength);
        mUseReducedSwing = m.getBooleanWithDefault(KEY_USE_REDUCED_SWING, mUseReducedSwing);
        mUseThrustEqualsSwingMinus2 = m.getBooleanWithDefault(KEY_USE_THRUST_EQUALS_SWING_MINUS_2, mUseThrustEqualsSwingMinus2);
        mUseSimpleMetricConversions = m.getBooleanWithDefault(KEY_USE_SIMPLE_METRIC_CONVERSIONS, mUseSimpleMetricConversions);
        mShowCollegeInSpells = m.getBooleanWithDefault(KEY_SHOW_COLLEGE_IN_SPELLS, mShowCollegeInSpells);
        mShowDifficulty = m.getBooleanWithDefault(KEY_SHOW_DIFFICULTY, mShowDifficulty);
        mShowAdvantageModifierAdj = m.getBooleanWithDefault(KEY_SHOW_ADVANTAGE_MODIFIER_ADJ, mShowAdvantageModifierAdj);
        mShowEquipmentModifierAdj = m.getBooleanWithDefault(KEY_SHOW_EQUIPMENT_MODIFIER_ADJ, mShowEquipmentModifierAdj);
        mShowSpellAdj = m.getBooleanWithDefault(KEY_SHOW_SPELL_ADJ, mShowSpellAdj);
        mUseTitleInFooter = m.getBooleanWithDefault(KEY_USE_TITLE_IN_FOOTER, mUseTitleInFooter);
        if (m.has(KEY_ATTRIBUTES)) {
            mAttributes = AttributeDef.load(m.getArray(KEY_ATTRIBUTES));
        }
        if (m.has(KEY_HIT_LOCATIONS)) {
            mHitLocations = new HitLocationTable(m.getMap(KEY_HIT_LOCATIONS));
        }
        if (m.has(KEY_PAGE)) {
            mPageSettings.load(m.getMap(KEY_PAGE));
        }
        if (m.has(KEY_BLOCK_LAYOUT)) {
            JsonArray array = m.getArray(KEY_BLOCK_LAYOUT);
            int       count = array.size();
            mBlockLayout = new ArrayList<>();
            for (int i = 0; i < count; i++) {
                mBlockLayout.add(array.getString(i));
            }
        }
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_DEFAULT_LENGTH_UNITS, Enums.toId(mDefaultLengthUnits));
        w.keyValue(KEY_DEFAULT_WEIGHT_UNITS, Enums.toId(mDefaultWeightUnits));
        w.keyValue(KEY_USER_DESCRIPTION_DISPLAY, Enums.toId(mUserDescriptionDisplay));
        w.keyValue(KEY_MODIFIERS_DISPLAY, Enums.toId(mModifiersDisplay));
        w.keyValue(KEY_NOTES_DISPLAY, Enums.toId(mNotesDisplay));
        w.keyValue(KEY_USE_MULTIPLICATIVE_MODIFIERS, mUseMultiplicativeModifiers);
        w.keyValue(KEY_USE_MODIFYING_DICE_PLUS_ADDS, mUseModifyingDicePlusAdds);
        w.keyValue(KEY_USE_KNOW_YOUR_OWN_STRENGTH, mUseKnowYourOwnStrength);
        w.keyValue(KEY_USE_REDUCED_SWING, mUseReducedSwing);
        w.keyValue(KEY_USE_THRUST_EQUALS_SWING_MINUS_2, mUseThrustEqualsSwingMinus2);
        w.keyValue(KEY_USE_SIMPLE_METRIC_CONVERSIONS, mUseSimpleMetricConversions);
        w.keyValue(KEY_SHOW_COLLEGE_IN_SPELLS, mShowCollegeInSpells);
        w.keyValue(KEY_SHOW_DIFFICULTY, mShowDifficulty);
        w.keyValue(KEY_SHOW_ADVANTAGE_MODIFIER_ADJ, mShowAdvantageModifierAdj);
        w.keyValue(KEY_SHOW_EQUIPMENT_MODIFIER_ADJ, mShowEquipmentModifierAdj);
        w.keyValue(KEY_SHOW_SPELL_ADJ, mShowSpellAdj);
        w.keyValue(KEY_USE_TITLE_IN_FOOTER, mUseTitleInFooter);
        w.key(KEY_ATTRIBUTES);
        AttributeDef.writeOrdered(w, mAttributes);
        w.key(KEY_HIT_LOCATIONS);
        mHitLocations.toJSON(w, mCharacter);
        w.key(KEY_PAGE);
        mPageSettings.toJSON(w);
        w.key(KEY_BLOCK_LAYOUT);
        w.startArray();
        for (String one : mBlockLayout) {
            w.value(one);
        }
        w.endArray();
        w.endMap();
    }

    public void notifyOfChange() {
        if (mCharacter == null) {
            Settings.getInstance().notifyOfChange();
        } else {
            mCharacter.notifyOfChange();
        }
    }

    public LengthUnits defaultLengthUnits() {
        return mDefaultLengthUnits;
    }

    public void setDefaultLengthUnits(LengthUnits defaultLengthUnits) {
        if (mDefaultLengthUnits != defaultLengthUnits) {
            mDefaultLengthUnits = defaultLengthUnits;
            notifyOfChange();
        }
    }

    public WeightUnits defaultWeightUnits() {
        return mDefaultWeightUnits;
    }

    public void setDefaultWeightUnits(WeightUnits defaultWeightUnits) {
        if (mDefaultWeightUnits != defaultWeightUnits) {
            mDefaultWeightUnits = defaultWeightUnits;
            notifyOfChange();
        }
    }

    public List<String> blockLayout() {
        return mBlockLayout;
    }

    public void setBlockLayout(List<String> blockLayout) {
        if (!mBlockLayout.equals(blockLayout)) {
            mBlockLayout = new ArrayList<>(blockLayout);
            notifyOfChange();
        }
    }

    public DisplayOption userDescriptionDisplay() {
        return mUserDescriptionDisplay;
    }

    public void setUserDescriptionDisplay(DisplayOption userDescriptionDisplay) {
        if (mUserDescriptionDisplay != userDescriptionDisplay) {
            mUserDescriptionDisplay = userDescriptionDisplay;
            notifyOfChange();
        }
    }

    public DisplayOption modifiersDisplay() {
        return mModifiersDisplay;
    }

    public void setModifiersDisplay(DisplayOption modifiersDisplay) {
        if (mModifiersDisplay != modifiersDisplay) {
            mModifiersDisplay = modifiersDisplay;
            notifyOfChange();
        }
    }

    public DisplayOption notesDisplay() {
        return mNotesDisplay;
    }

    public void setNotesDisplay(DisplayOption notesDisplay) {
        if (mNotesDisplay != notesDisplay) {
            mNotesDisplay = notesDisplay;
            notifyOfChange();
        }
    }

    public boolean useMultiplicativeModifiers() {
        return mUseMultiplicativeModifiers;
    }

    public void setUseMultiplicativeModifiers(boolean useMultiplicativeModifiers) {
        if (mUseMultiplicativeModifiers != useMultiplicativeModifiers) {
            mUseMultiplicativeModifiers = useMultiplicativeModifiers;
            notifyOfChange();
        }
    }

    public boolean useModifyingDicePlusAdds() {
        return mUseModifyingDicePlusAdds;
    }

    public void setUseModifyingDicePlusAdds(boolean useModifyingDicePlusAdds) {
        if (mUseModifyingDicePlusAdds != useModifyingDicePlusAdds) {
            mUseModifyingDicePlusAdds = useModifyingDicePlusAdds;
            notifyOfChange();
        }
    }

    public boolean useKnowYourOwnStrength() {
        return mUseKnowYourOwnStrength;
    }

    public void setUseKnowYourOwnStrength(boolean useKnowYourOwnStrength) {
        if (mUseKnowYourOwnStrength != useKnowYourOwnStrength) {
            mUseKnowYourOwnStrength = useKnowYourOwnStrength;
            notifyOfChange();
        }
    }

    public boolean useReducedSwing() {
        return mUseReducedSwing;
    }

    public void setUseReducedSwing(boolean useReducedSwing) {
        if (mUseReducedSwing != useReducedSwing) {
            mUseReducedSwing = useReducedSwing;
            notifyOfChange();
        }
    }

    public boolean useThrustEqualsSwingMinus2() {
        return mUseThrustEqualsSwingMinus2;
    }

    public void setUseThrustEqualsSwingMinus2(boolean useThrustEqualsSwingMinus2) {
        if (mUseThrustEqualsSwingMinus2 != useThrustEqualsSwingMinus2) {
            mUseThrustEqualsSwingMinus2 = useThrustEqualsSwingMinus2;
            notifyOfChange();
        }
    }

    public boolean useSimpleMetricConversions() {
        return mUseSimpleMetricConversions;
    }

    public void setUseSimpleMetricConversions(boolean useSimpleMetricConversions) {
        if (mUseSimpleMetricConversions != useSimpleMetricConversions) {
            mUseSimpleMetricConversions = useSimpleMetricConversions;
            notifyOfChange();
        }
    }

    public boolean showCollegeInSpells() {
        return mShowCollegeInSpells;
    }

    public void setShowCollegeInSpells(boolean show) {
        if (mShowCollegeInSpells != show) {
            mShowCollegeInSpells = show;
            notifyOfChange();
        }
    }

    public boolean showDifficulty() {
        return mShowDifficulty;
    }

    public void setShowDifficulty(boolean show) {
        if (mShowDifficulty != show) {
            mShowDifficulty = show;
            notifyOfChange();
        }
    }

    public boolean showAdvantageModifierAdj() {
        return mShowAdvantageModifierAdj;
    }

    public void setShowAdvantageModifierAdj(boolean show) {
        if (mShowAdvantageModifierAdj != show) {
            mShowAdvantageModifierAdj = show;
            notifyOfChange();
        }
    }

    public boolean showEquipmentModifierAdj() {
        return mShowEquipmentModifierAdj;
    }

    public void setShowEquipmentModifierAdj(boolean show) {
        if (mShowEquipmentModifierAdj != show) {
            mShowEquipmentModifierAdj = show;
            notifyOfChange();
        }
    }

    public boolean showSpellAdj() {
        return mShowSpellAdj;
    }

    public void setShowSpellAdj(boolean show) {
        if (mShowSpellAdj != show) {
            mShowSpellAdj = show;
            notifyOfChange();
        }
    }

    public boolean useTitleInFooter() {
        return mUseTitleInFooter;
    }

    public void setUseTitleInFooter(boolean show) {
        if (mUseTitleInFooter != show) {
            mUseTitleInFooter = show;
            notifyOfChange();
        }
    }

    public Map<String, AttributeDef> getAttributes() {
        return mAttributes;
    }

    public void setAttributes(Map<String, AttributeDef> attributes) {
        if (!mAttributes.equals(attributes)) {
            mAttributes = AttributeDef.cloneMap(attributes);
            notifyOfChange();
        }
    }

    public HitLocationTable getHitLocations() {
        return mHitLocations;
    }

    public void setHitLocations(HitLocationTable hitLocations) {
        if (!mHitLocations.equals(hitLocations)) {
            mHitLocations = hitLocations;
            notifyOfChange();
        }
    }

    public PageSettings getPageSettings() {
        return mPageSettings;
    }
}
