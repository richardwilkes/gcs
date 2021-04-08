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

package com.trollworks.gcs.character;

import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.utility.VersionException;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

public class Settings {
    private static final int            CURRENT_JSON_VERSION                   = 1;
    private static final int            CURRENT_VERSION                        = 1;
    private static final int            MINIMUM_VERSION                        = 0;
    public static final  String         TAG_ROOT                               = "settings";
    public static final  String         TAG_DEFAULT_LENGTH_UNITS               = "default_length_units";
    public static final  String         TAG_DEFAULT_WEIGHT_UNITS               = "default_weight_units";
    public static final  String         TAG_BLOCK_LAYOUT                       = "block_layout";
    public static final  String         TAG_USER_DESCRIPTION_DISPLAY           = "user_description_display";
    public static final  String         TAG_MODIFIERS_DISPLAY                  = "modifiers_display";
    public static final  String         TAG_NOTES_DISPLAY                      = "notes_display";
    public static final  String         DEPRECATED_TAG_BASE_WILL_AND_PER_ON_10 = "base_will_and_per_on_10";
    public static final  String         TAG_BASE_WILL_ON_10                    = "base_will_on_10";
    public static final  String         TAG_BASE_PER_ON_10                     = "base_per_on_10";
    public static final  String         TAG_USE_MULTIPLICATIVE_MODIFIERS       = "use_multiplicative_modifiers";
    public static final  String         TAG_USE_MODIFYING_DICE_PLUS_ADDS       = "use_modifying_dice_plus_adds";
    public static final  String         TAG_USE_KNOW_YOUR_OWN_STRENGTH         = "use_know_your_own_strength";
    public static final  String         TAG_USE_REDUCED_SWING                  = "use_reduced_swing";
    public static final  String         TAG_USE_THRUST_EQUALS_SWING_MINUS_2    = "use_thrust_equals_swing_minus_2";
    public static final  String         TAG_USE_SIMPLE_METRIC_CONVERSIONS      = "use_simple_metric_conversions";
    public static final  String         TAG_SHOW_COLLEGE_IN_SPELLS             = "show_college_in_sheet_spells";
    public static final  String         TAG_SHOW_DIFFICULTY                    = "show_difficulty";
    public static final  String         TAG_SHOW_ADVANTAGE_MODIFIER_ADJ        = "show_advantage_modifier_adj";
    public static final  String         TAG_SHOW_EQUIPMENT_MODIFIER_ADJ        = "show_equipment_modifier_adj";
    public static final  String         TAG_SHOW_SPELL_ADJ                    = "show_spell_adj";
    public static final  String         TAG_USE_TITLE_IN_FOOTER                = "use_title_in_footer";
    private static final String         TAG_EXTRA_SPACE_AROUND_ENCUMBRANCE     = "extra_space_around_encumbrance";
    public static final  String         PREFIX                                 = GURPSCharacter.CHARACTER_PREFIX + "settings.";
    public static final  String         ID_DEFAULT_LENGTH_UNITS                = PREFIX + TAG_DEFAULT_LENGTH_UNITS;
    public static final  String         ID_DEFAULT_WEIGHT_UNITS                = PREFIX + TAG_DEFAULT_WEIGHT_UNITS;
    public static final  String         ID_BLOCK_LAYOUT                        = PREFIX + TAG_BLOCK_LAYOUT;
    public static final  String         ID_USER_DESCRIPTION_DISPLAY            = PREFIX + TAG_USER_DESCRIPTION_DISPLAY;
    public static final  String         ID_MODIFIERS_DISPLAY                   = PREFIX + TAG_MODIFIERS_DISPLAY;
    public static final  String         ID_NOTES_DISPLAY                       = PREFIX + TAG_NOTES_DISPLAY;
    public static final  String         ID_BASE_WILL_ON_10                     = PREFIX + TAG_BASE_WILL_ON_10;
    public static final  String         ID_BASE_PER_ON_10                      = PREFIX + TAG_BASE_PER_ON_10;
    public static final  String         ID_USE_MULTIPLICATIVE_MODIFIERS        = PREFIX + TAG_USE_MULTIPLICATIVE_MODIFIERS;
    public static final  String         ID_USE_MODIFYING_DICE_PLUS_ADDS        = PREFIX + TAG_USE_MODIFYING_DICE_PLUS_ADDS;
    public static final  String         ID_USE_KNOW_YOUR_OWN_STRENGTH          = PREFIX + TAG_USE_KNOW_YOUR_OWN_STRENGTH;
    public static final  String         ID_USE_REDUCED_SWING                   = PREFIX + TAG_USE_REDUCED_SWING;
    public static final  String         ID_USE_THRUST_EQUALS_SWING_MINUS_2     = PREFIX + TAG_USE_THRUST_EQUALS_SWING_MINUS_2;
    public static final  String         ID_USE_SIMPLE_METRIC_CONVERSIONS       = PREFIX + TAG_USE_SIMPLE_METRIC_CONVERSIONS;
    public static final  String         ID_SHOW_COLLEGE_IN_SPELLS              = PREFIX + TAG_SHOW_COLLEGE_IN_SPELLS;
    public static final  String         ID_SHOW_DIFFICULTY                     = PREFIX + TAG_SHOW_DIFFICULTY;
    public static final  String         ID_SHOW_ADVANTAGE_MODIFIER_ADJ         = PREFIX + TAG_SHOW_ADVANTAGE_MODIFIER_ADJ;
    public static final  String         ID_SHOW_EQUIPMENT_MODIFIER_ADJ         = PREFIX + TAG_SHOW_EQUIPMENT_MODIFIER_ADJ;
    public static final  String         ID_SHOW_SPELL_ADJ                     = PREFIX + TAG_SHOW_SPELL_ADJ;
    public static final  String         ID_USE_TITLE_IN_FOOTER                 = PREFIX + TAG_USE_TITLE_IN_FOOTER;
    public static final  String         ID_EXTRA_SPACE_AROUND_ENCUMBRANCE      = PREFIX + TAG_EXTRA_SPACE_AROUND_ENCUMBRANCE;
    private              GURPSCharacter mCharacter;
    private              LengthUnits    mDefaultLengthUnits;
    private              WeightUnits    mDefaultWeightUnits;
    private              List<String>   mBlockLayout;
    private              DisplayOption  mUserDescriptionDisplay;
    private              DisplayOption  mModifiersDisplay;
    private              DisplayOption  mNotesDisplay;
    private              boolean        mBaseWillOn10; // Home brew
    private              boolean        mBasePerOn10; // Home brew
    private              boolean        mUseMultiplicativeModifiers; // P102
    private              boolean        mUseModifyingDicePlusAdds; // B269
    private              boolean        mUseKnowYourOwnStrength; // PY83
    private              boolean        mUseReducedSwing; // Adjusting Swing Damage from noschoolgrognard.blogspot.com
    private              boolean        mUseThrustEqualsSwingMinus2; // Home brew
    private              boolean        mUseSimpleMetricConversions; // B9
    private              boolean        mShowCollegeInSpells;
    private              boolean        mShowDifficulty;
    private              boolean        mShowAdvantageModifierAdj;
    private              boolean        mShowEquipmentModifierAdj;
    private              boolean        mShowSpellAdj;
    private              boolean        mUseTitleInFooter;
    private              boolean        mExtraSpaceAroundEncumbrance;

    public Settings(GURPSCharacter character) {
        Preferences prefs = Preferences.getInstance();
        mCharacter = character;
        mDefaultLengthUnits = prefs.getDefaultLengthUnits();
        mDefaultWeightUnits = prefs.getDefaultWeightUnits();
        mBlockLayout = new ArrayList<>(prefs.getBlockLayout());
        mUserDescriptionDisplay = prefs.getUserDescriptionDisplay();
        mModifiersDisplay = prefs.getModifiersDisplay();
        mNotesDisplay = prefs.getNotesDisplay();
        mBaseWillOn10 = prefs.baseWillOn10();
        mBasePerOn10 = prefs.basePerOn10();
        mUseMultiplicativeModifiers = prefs.useMultiplicativeModifiers();
        mUseModifyingDicePlusAdds = prefs.useModifyingDicePlusAdds();
        mUseKnowYourOwnStrength = prefs.useKnowYourOwnStrength();
        mUseReducedSwing = prefs.useReducedSwing();
        mUseThrustEqualsSwingMinus2 = prefs.useThrustEqualsSwingMinus2();
        mUseSimpleMetricConversions = prefs.useSimpleMetricConversions();
        mShowCollegeInSpells = prefs.showCollegeInSheetSpells();
        mShowDifficulty = prefs.showDifficulty();
        mShowAdvantageModifierAdj = prefs.showAdvantageModifierAdj();
        mShowEquipmentModifierAdj = prefs.showEquipmentModifierAdj();
        mShowSpellAdj = prefs.showSpellAdj();
        mUseTitleInFooter = prefs.useTitleInFooter();
        mExtraSpaceAroundEncumbrance = prefs.extraSpaceAroundEncumbrance();
    }

    void load(JsonMap m) throws IOException {
        int version = m.getInt(LoadState.ATTRIBUTE_VERSION);
        if (version < MINIMUM_VERSION) {
            throw VersionException.createTooOld();
        }
        if (version > CURRENT_VERSION) {
            throw VersionException.createTooNew();
        }
        mDefaultLengthUnits = Enums.extract(m.getString(TAG_DEFAULT_LENGTH_UNITS), LengthUnits.values(), Preferences.DEFAULT_DEFAULT_LENGTH_UNITS);
        mDefaultWeightUnits = Enums.extract(m.getString(TAG_DEFAULT_WEIGHT_UNITS), WeightUnits.values(), Preferences.DEFAULT_DEFAULT_WEIGHT_UNITS);
        mUserDescriptionDisplay = Enums.extract(m.getString(TAG_USER_DESCRIPTION_DISPLAY), DisplayOption.values(), Preferences.DEFAULT_USER_DESCRIPTION_DISPLAY);
        mModifiersDisplay = Enums.extract(m.getString(TAG_MODIFIERS_DISPLAY), DisplayOption.values(), Preferences.DEFAULT_MODIFIERS_DISPLAY);
        mNotesDisplay = Enums.extract(m.getString(TAG_NOTES_DISPLAY), DisplayOption.values(), Preferences.DEFAULT_NOTES_DISPLAY);
        if (m.has(DEPRECATED_TAG_BASE_WILL_AND_PER_ON_10)) {
            mBaseWillOn10 = m.getBoolean(DEPRECATED_TAG_BASE_WILL_AND_PER_ON_10);
            mBasePerOn10 = m.getBoolean(DEPRECATED_TAG_BASE_WILL_AND_PER_ON_10);
        } else {
            mBaseWillOn10 = m.getBoolean(TAG_BASE_WILL_ON_10);
            mBasePerOn10 = m.getBoolean(TAG_BASE_PER_ON_10);
        }
        mUseMultiplicativeModifiers = m.getBoolean(TAG_USE_MULTIPLICATIVE_MODIFIERS);
        mUseModifyingDicePlusAdds = m.getBoolean(TAG_USE_MODIFYING_DICE_PLUS_ADDS);
        mUseKnowYourOwnStrength = m.getBoolean(TAG_USE_KNOW_YOUR_OWN_STRENGTH);
        mUseReducedSwing = m.getBoolean(TAG_USE_REDUCED_SWING);
        mUseThrustEqualsSwingMinus2 = m.getBoolean(TAG_USE_THRUST_EQUALS_SWING_MINUS_2);
        mUseSimpleMetricConversions = m.getBoolean(TAG_USE_SIMPLE_METRIC_CONVERSIONS);
        mShowCollegeInSpells = m.getBoolean(TAG_SHOW_COLLEGE_IN_SPELLS);
        mShowDifficulty = m.getBoolean(TAG_SHOW_DIFFICULTY);
        mShowAdvantageModifierAdj = m.getBoolean(TAG_SHOW_ADVANTAGE_MODIFIER_ADJ);
        mShowEquipmentModifierAdj = m.getBoolean(TAG_SHOW_EQUIPMENT_MODIFIER_ADJ);
        if (m.has(TAG_SHOW_SPELL_ADJ)) {
            mShowSpellAdj = m.getBoolean(TAG_SHOW_SPELL_ADJ);
        } else {
            mShowSpellAdj = Preferences.DEFAULT_SHOW_SPELL_ADJ;
        }
        mUseTitleInFooter = m.getBoolean(TAG_USE_TITLE_IN_FOOTER);
        mExtraSpaceAroundEncumbrance = m.getBoolean(TAG_EXTRA_SPACE_AROUND_ENCUMBRANCE);
        mBlockLayout = new ArrayList<>();
        JsonArray a     = m.getArray(TAG_BLOCK_LAYOUT);
        int       count = a.size();
        for (int i = 0; i < count; i++) {
            mBlockLayout.add(a.getString(i));
        }
    }

    void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(LoadState.ATTRIBUTE_VERSION, CURRENT_JSON_VERSION);
        w.keyValue(TAG_DEFAULT_LENGTH_UNITS, Enums.toId(mDefaultLengthUnits));
        w.keyValue(TAG_DEFAULT_WEIGHT_UNITS, Enums.toId(mDefaultWeightUnits));
        w.keyValue(TAG_USER_DESCRIPTION_DISPLAY, Enums.toId(mUserDescriptionDisplay));
        w.keyValue(TAG_MODIFIERS_DISPLAY, Enums.toId(mModifiersDisplay));
        w.keyValue(TAG_NOTES_DISPLAY, Enums.toId(mNotesDisplay));
        w.keyValue(TAG_BASE_WILL_ON_10, mBaseWillOn10);
        w.keyValue(TAG_BASE_PER_ON_10, mBasePerOn10);
        w.keyValue(TAG_USE_MULTIPLICATIVE_MODIFIERS, mUseMultiplicativeModifiers);
        w.keyValue(TAG_USE_MODIFYING_DICE_PLUS_ADDS, mUseModifyingDicePlusAdds);
        w.keyValue(TAG_USE_KNOW_YOUR_OWN_STRENGTH, mUseKnowYourOwnStrength);
        w.keyValue(TAG_USE_REDUCED_SWING, mUseReducedSwing);
        w.keyValue(TAG_USE_THRUST_EQUALS_SWING_MINUS_2, mUseThrustEqualsSwingMinus2);
        w.keyValue(TAG_USE_SIMPLE_METRIC_CONVERSIONS, mUseSimpleMetricConversions);
        w.keyValue(TAG_SHOW_COLLEGE_IN_SPELLS, mShowCollegeInSpells);
        w.keyValue(TAG_SHOW_DIFFICULTY, mShowDifficulty);
        w.keyValue(TAG_SHOW_ADVANTAGE_MODIFIER_ADJ, mShowAdvantageModifierAdj);
        w.keyValue(TAG_SHOW_EQUIPMENT_MODIFIER_ADJ, mShowEquipmentModifierAdj);
        w.keyValue(TAG_SHOW_SPELL_ADJ, mShowSpellAdj);
        w.keyValue(TAG_USE_TITLE_IN_FOOTER, mUseTitleInFooter);
        w.keyValue(TAG_EXTRA_SPACE_AROUND_ENCUMBRANCE, mExtraSpaceAroundEncumbrance);
        w.key(TAG_BLOCK_LAYOUT);
        w.startArray();
        for (String one : mBlockLayout) {
            w.value(one);
        }
        w.endArray();
        w.endMap();
    }

    @SuppressWarnings("StringBufferReplaceableByString")
    public String optionsCode() {
        StringBuilder buffer = new StringBuilder();
        buffer.append(mBaseWillOn10 ? 'W' : 'w');
        buffer.append(mBasePerOn10 ? 'P' : 'p');
        buffer.append(mUseMultiplicativeModifiers ? 'M' : 'm');
        buffer.append(mUseModifyingDicePlusAdds ? 'D' : 'd');
        buffer.append(mUseKnowYourOwnStrength ? 'K' : 'k');
        buffer.append(mUseReducedSwing ? 'S' : 's');
        buffer.append(mUseThrustEqualsSwingMinus2 ? 'T' : 't');
        buffer.append(mUseSimpleMetricConversions ? 'C' : 'c');
        return buffer.toString();
    }

    public LengthUnits defaultLengthUnits() {
        return mDefaultLengthUnits;
    }

    public void setDefaultLengthUnits(LengthUnits defaultLengthUnits) {
        if (mDefaultLengthUnits != defaultLengthUnits) {
            mDefaultLengthUnits = defaultLengthUnits;
            mCharacter.notifySingle(ID_DEFAULT_LENGTH_UNITS, mDefaultLengthUnits);
        }
    }

    public WeightUnits defaultWeightUnits() {
        return mDefaultWeightUnits;
    }

    public void setDefaultWeightUnits(WeightUnits defaultWeightUnits) {
        if (mDefaultWeightUnits != defaultWeightUnits) {
            mDefaultWeightUnits = defaultWeightUnits;
            mCharacter.notifySingle(ID_DEFAULT_WEIGHT_UNITS, mDefaultWeightUnits);
        }
    }

    public List<String> blockLayout() {
        return mBlockLayout;
    }

    public void setBlockLayout(List<String> blockLayout) {
        if (!mBlockLayout.equals(blockLayout)) {
            mBlockLayout = new ArrayList<>(blockLayout);
            mCharacter.notifySingle(ID_BLOCK_LAYOUT, mBlockLayout);
        }
    }

    public DisplayOption userDescriptionDisplay() {
        return mUserDescriptionDisplay;
    }

    public void setUserDescriptionDisplay(DisplayOption userDescriptionDisplay) {
        if (mUserDescriptionDisplay != userDescriptionDisplay) {
            mUserDescriptionDisplay = userDescriptionDisplay;
            mCharacter.notifySingle(ID_USER_DESCRIPTION_DISPLAY, mUserDescriptionDisplay);
        }
    }

    public DisplayOption modifiersDisplay() {
        return mModifiersDisplay;
    }

    public void setModifiersDisplay(DisplayOption modifiersDisplay) {
        if (mModifiersDisplay != modifiersDisplay) {
            mModifiersDisplay = modifiersDisplay;
            mCharacter.notifySingle(ID_MODIFIERS_DISPLAY, mModifiersDisplay);
        }
    }

    public DisplayOption notesDisplay() {
        return mNotesDisplay;
    }

    public void setNotesDisplay(DisplayOption notesDisplay) {
        if (mNotesDisplay != notesDisplay) {
            mNotesDisplay = notesDisplay;
            mCharacter.notifySingle(ID_NOTES_DISPLAY, mNotesDisplay);
        }
    }

    public boolean baseWillOn10() {
        return mBaseWillOn10;
    }

    public void setBaseWillOn10(boolean baseWillOn10) {
        if (mBaseWillOn10 != baseWillOn10) {
            mBaseWillOn10 = baseWillOn10;
            mCharacter.notifySingle(ID_BASE_WILL_ON_10, Boolean.valueOf(mBaseWillOn10));
        }
    }

    public boolean basePerOn10() {
        return mBasePerOn10;
    }

    public void setBasePerOn10(boolean basePerOn10) {
        if (mBasePerOn10 != basePerOn10) {
            mBasePerOn10 = basePerOn10;
            mCharacter.notifySingle(ID_BASE_PER_ON_10, Boolean.valueOf(mBasePerOn10));
        }
    }

    public boolean useMultiplicativeModifiers() {
        return mUseMultiplicativeModifiers;
    }

    public void setUseMultiplicativeModifiers(boolean useMultiplicativeModifiers) {
        if (mUseMultiplicativeModifiers != useMultiplicativeModifiers) {
            mUseMultiplicativeModifiers = useMultiplicativeModifiers;
            mCharacter.notifySingle(ID_USE_MULTIPLICATIVE_MODIFIERS, Boolean.valueOf(mUseMultiplicativeModifiers));
        }
    }

    public boolean useModifyingDicePlusAdds() {
        return mUseModifyingDicePlusAdds;
    }

    public void setUseModifyingDicePlusAdds(boolean useModifyingDicePlusAdds) {
        if (mUseModifyingDicePlusAdds != useModifyingDicePlusAdds) {
            mUseModifyingDicePlusAdds = useModifyingDicePlusAdds;
            mCharacter.notifySingle(ID_USE_MODIFYING_DICE_PLUS_ADDS, Boolean.valueOf(mUseModifyingDicePlusAdds));
        }
    }

    public boolean useKnowYourOwnStrength() {
        return mUseKnowYourOwnStrength;
    }

    public void setUseKnowYourOwnStrength(boolean useKnowYourOwnStrength) {
        if (mUseKnowYourOwnStrength != useKnowYourOwnStrength) {
            mUseKnowYourOwnStrength = useKnowYourOwnStrength;
            mCharacter.notifySingle(ID_USE_KNOW_YOUR_OWN_STRENGTH, Boolean.valueOf(mUseKnowYourOwnStrength));
        }
    }

    public boolean useReducedSwing() {
        return mUseReducedSwing;
    }

    public void setUseReducedSwing(boolean useReducedSwing) {
        if (mUseReducedSwing != useReducedSwing) {
            mUseReducedSwing = useReducedSwing;
            mCharacter.notifySingle(ID_USE_REDUCED_SWING, Boolean.valueOf(mUseReducedSwing));
        }
    }

    public boolean useThrustEqualsSwingMinus2() {
        return mUseThrustEqualsSwingMinus2;
    }

    public void setUseThrustEqualsSwingMinus2(boolean useThrustEqualsSwingMinus2) {
        if (mUseThrustEqualsSwingMinus2 != useThrustEqualsSwingMinus2) {
            mUseThrustEqualsSwingMinus2 = useThrustEqualsSwingMinus2;
            mCharacter.notifySingle(ID_USE_THRUST_EQUALS_SWING_MINUS_2, Boolean.valueOf(mUseThrustEqualsSwingMinus2));
        }
    }

    public boolean useSimpleMetricConversions() {
        return mUseSimpleMetricConversions;
    }

    public void setUseSimpleMetricConversions(boolean useSimpleMetricConversions) {
        if (mUseSimpleMetricConversions != useSimpleMetricConversions) {
            mUseSimpleMetricConversions = useSimpleMetricConversions;
            mCharacter.notifySingle(ID_USE_SIMPLE_METRIC_CONVERSIONS, Boolean.valueOf(mUseSimpleMetricConversions));
        }
    }

    public boolean showCollegeInSpells() {
        return mShowCollegeInSpells;
    }

    public void setShowCollegeInSpells(boolean show) {
        if (mShowCollegeInSpells != show) {
            mShowCollegeInSpells = show;
            mCharacter.notifySingle(ID_SHOW_COLLEGE_IN_SPELLS, Boolean.valueOf(mShowCollegeInSpells));
        }
    }

    public boolean showDifficulty() {
        return mShowDifficulty;
    }

    public void setShowDifficulty(boolean show) {
        if (mShowDifficulty != show) {
            mShowDifficulty = show;
            mCharacter.notifySingle(ID_SHOW_DIFFICULTY, Boolean.valueOf(mShowDifficulty));
        }
    }

    public boolean showAdvantageModifierAdj() {
        return mShowAdvantageModifierAdj;
    }

    public void setShowAdvantageModifierAdj(boolean show) {
        if (mShowAdvantageModifierAdj != show) {
            mShowAdvantageModifierAdj = show;
            mCharacter.notifySingle(ID_SHOW_ADVANTAGE_MODIFIER_ADJ, Boolean.valueOf(mShowAdvantageModifierAdj));
        }
    }

    public boolean showEquipmentModifierAdj() {
        return mShowEquipmentModifierAdj;
    }

    public void setShowEquipmentModifierAdj(boolean show) {
        if (mShowEquipmentModifierAdj != show) {
            mShowEquipmentModifierAdj = show;
            mCharacter.notifySingle(ID_SHOW_EQUIPMENT_MODIFIER_ADJ, Boolean.valueOf(mShowEquipmentModifierAdj));
        }
    }

    public boolean showSpellAdj() {
        return mShowSpellAdj;
    }

    public void setShowSpellAdj(boolean show) {
        if (mShowSpellAdj != show) {
            mShowSpellAdj = show;
            mCharacter.notifySingle(ID_SHOW_SPELL_ADJ, Boolean.valueOf(mShowSpellAdj));
        }
    }

    public boolean useTitleInFooter() {
        return mUseTitleInFooter;
    }

    public void setUseTitleInFooter(boolean show) {
        if (mUseTitleInFooter != show) {
            mUseTitleInFooter = show;
            mCharacter.notifySingle(ID_USE_TITLE_IN_FOOTER, Boolean.valueOf(mUseTitleInFooter));
        }
    }

    public boolean extraSpaceAroundEncumbrance() {
        return mExtraSpaceAroundEncumbrance;
    }

    public void setExtraSpaceAroundEncumbrance(boolean extraSpaceAroundEncumbrance) {
        if (mExtraSpaceAroundEncumbrance != extraSpaceAroundEncumbrance) {
            mExtraSpaceAroundEncumbrance = extraSpaceAroundEncumbrance;
            mCharacter.notifySingle(ID_EXTRA_SPACE_AROUND_ENCUMBRANCE, Boolean.valueOf(mExtraSpaceAroundEncumbrance));
        }
    }
}
