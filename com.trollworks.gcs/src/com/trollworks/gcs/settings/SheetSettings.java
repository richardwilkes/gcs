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
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.DisplayOption;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.ChangeNotifier;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.page.PageSettings;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class SheetSettings implements ChangeNotifier {
    public static final  String KEY_ATTRIBUTES                    = "attributes";
    private static final String KEY_BLOCK_LAYOUT                  = "block_layout";
    private static final String KEY_DEFAULT_LENGTH_UNITS          = "default_length_units";
    private static final String KEY_DEFAULT_WEIGHT_UNITS          = "default_weight_units";
    public static final  String KEY_HIT_LOCATIONS                 = "hit_locations";
    private static final String KEY_MODIFIERS_DISPLAY             = "modifiers_display";
    private static final String KEY_NOTES_DISPLAY                 = "notes_display";
    private static final String KEY_PAGE                          = "page";
    private static final String KEY_SHOW_ADVANTAGE_MODIFIER_ADJ   = "show_advantage_modifier_adj";
    private static final String KEY_SHOW_COLLEGE_IN_SPELLS        = "show_college_in_sheet_spells";
    private static final String KEY_SHOW_DIFFICULTY               = "show_difficulty";
    private static final String KEY_SHOW_EQUIPMENT_MODIFIER_ADJ   = "show_equipment_modifier_adj";
    private static final String KEY_SHOW_SPELL_ADJ                = "show_spell_adj";
    private static final String KEY_USE_MODIFYING_DICE_PLUS_ADDS  = "use_modifying_dice_plus_adds";
    private static final String KEY_USE_MULTIPLICATIVE_MODIFIERS  = "use_multiplicative_modifiers";
    private static final String KEY_USE_SIMPLE_METRIC_CONVERSIONS = "use_simple_metric_conversions";
    private static final String KEY_USE_TITLE_IN_FOOTER           = "use_title_in_footer";
    private static final String KEY_USER_DESCRIPTION_DISPLAY      = "user_description_display";
    private static final String KEY_DAMAGE_PROGRESSION            = "damage_progression";

    // TODO: Eliminate these deprecated keys after a suitable waiting period; added July 7, 2021
    private static final String DEPRECATED_KEY_USE_KNOW_YOUR_OWN_STRENGTH      = "use_know_your_own_strength";
    private static final String DEPRECATED_KEY_USE_REDUCED_SWING               = "use_reduced_swing";
    private static final String DEPRECATED_KEY_USE_THRUST_EQUALS_SWING_MINUS_2 = "use_thrust_equals_swing_minus_2";

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
    private DamageProgression         mDamageProgression;
    private boolean                   mUseMultiplicativeModifiers; // P102
    private boolean                   mUseModifyingDicePlusAdds; // B269
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

    public SheetSettings(Path path) throws IOException {
        this();
        try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m = Json.asMap(Json.parse(in));
            if (!m.isEmpty()) {
                int version = m.getInt(Settings.VERSION);
                if (version >= Settings.MINIMUM_VERSION && version <= DataFile.CURRENT_VERSION) {
                    load(m.getMap(Settings.SHEET_SETTINGS));
                }
            }
        }
    }

    public void copyFrom(SheetSettings other) {
        mDefaultLengthUnits = other.mDefaultLengthUnits;
        mDefaultWeightUnits = other.mDefaultWeightUnits;
        mBlockLayout = new ArrayList<>(other.mBlockLayout);
        mUserDescriptionDisplay = other.mUserDescriptionDisplay;
        mModifiersDisplay = other.mModifiersDisplay;
        mNotesDisplay = other.mNotesDisplay;
        mAttributes = AttributeDef.cloneMap(other.mAttributes);
        mHitLocations = other.mHitLocations.clone();
        mPageSettings = new PageSettings(this, other.mPageSettings);
        mUseMultiplicativeModifiers = other.mUseMultiplicativeModifiers;
        mUseModifyingDicePlusAdds = other.mUseModifyingDicePlusAdds;
        mDamageProgression = other.mDamageProgression;
        mUseSimpleMetricConversions = other.mUseSimpleMetricConversions;
        mShowCollegeInSpells = other.mShowCollegeInSpells;
        mShowDifficulty = other.mShowDifficulty;
        mShowAdvantageModifierAdj = other.mShowAdvantageModifierAdj;
        mShowEquipmentModifierAdj = other.mShowEquipmentModifierAdj;
        mShowSpellAdj = other.mShowSpellAdj;
        mUseTitleInFooter = other.mUseTitleInFooter;
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
            mHitLocations = HitLocationTable.createHumanoidTable();
            mPageSettings = new PageSettings(this);
            mUseMultiplicativeModifiers = false;
            mUseModifyingDicePlusAdds = false;
            mDamageProgression = DamageProgression.BASIC_SET;
            mUseSimpleMetricConversions = true;
            mShowCollegeInSpells = false;
            mShowDifficulty = false;
            mShowAdvantageModifierAdj = false;
            mShowEquipmentModifierAdj = false;
            mShowSpellAdj = true;
            mUseTitleInFooter = false;
        } else {
            copyFrom(Settings.getInstance().getSheetSettings());
        }
    }

    public void load(JsonMap m) {
        reset();
        mDefaultLengthUnits = Enums.extract(m.getString(KEY_DEFAULT_LENGTH_UNITS), LengthUnits.values(), mDefaultLengthUnits);
        mDefaultWeightUnits = Enums.extract(m.getString(KEY_DEFAULT_WEIGHT_UNITS), WeightUnits.values(), mDefaultWeightUnits);
        mUserDescriptionDisplay = Enums.extract(m.getString(KEY_USER_DESCRIPTION_DISPLAY), DisplayOption.values(), mUserDescriptionDisplay);
        mModifiersDisplay = Enums.extract(m.getString(KEY_MODIFIERS_DISPLAY), DisplayOption.values(), mModifiersDisplay);
        mNotesDisplay = Enums.extract(m.getString(KEY_NOTES_DISPLAY), DisplayOption.values(), mNotesDisplay);
        mUseMultiplicativeModifiers = m.getBooleanWithDefault(KEY_USE_MULTIPLICATIVE_MODIFIERS, mUseMultiplicativeModifiers);
        mUseModifyingDicePlusAdds = m.getBooleanWithDefault(KEY_USE_MODIFYING_DICE_PLUS_ADDS, mUseModifyingDicePlusAdds);
        if (m.has(KEY_DAMAGE_PROGRESSION)) {
            mDamageProgression = Enums.extract(m.getString(KEY_DAMAGE_PROGRESSION), DamageProgression.values(), DamageProgression.BASIC_SET);
        } else if (m.getBoolean(DEPRECATED_KEY_USE_KNOW_YOUR_OWN_STRENGTH)) {
            mDamageProgression = DamageProgression.KNOWING_YOUR_OWN_STRENGTH;
        } else if (m.getBoolean(DEPRECATED_KEY_USE_REDUCED_SWING)) {
            mDamageProgression = DamageProgression.NO_SCHOOL_GROGNARD_DAMAGE;
        } else if (m.getBoolean(DEPRECATED_KEY_USE_THRUST_EQUALS_SWING_MINUS_2)) {
            mDamageProgression = DamageProgression.THRUST_EQUALS_SWING_MINUS_2;
        } else {
            mDamageProgression = DamageProgression.BASIC_SET;
        }
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

    public void save(Path path) throws IOException {
        SafeFileUpdater trans = new SafeFileUpdater();
        trans.begin();
        try {
            Files.createDirectories(path.getParent());
            File file = trans.getTransactionFile(path.toFile());
            try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(file, StandardCharsets.UTF_8)), "\t")) {
                w.startMap();
                w.keyValue(Settings.VERSION, DataFile.CURRENT_VERSION);
                w.key(Settings.SHEET_SETTINGS);
                save(w, false);
                w.endMap();
            }
        } catch (IOException ioe) {
            trans.abort();
            throw ioe;
        }
        trans.commit();
    }

    public void save(JsonWriter w, boolean full) throws IOException {
        w.startMap();
        w.keyValue(KEY_DEFAULT_LENGTH_UNITS, Enums.toId(mDefaultLengthUnits));
        w.keyValue(KEY_DEFAULT_WEIGHT_UNITS, Enums.toId(mDefaultWeightUnits));
        w.keyValue(KEY_USER_DESCRIPTION_DISPLAY, Enums.toId(mUserDescriptionDisplay));
        w.keyValue(KEY_MODIFIERS_DISPLAY, Enums.toId(mModifiersDisplay));
        w.keyValue(KEY_NOTES_DISPLAY, Enums.toId(mNotesDisplay));
        w.keyValue(KEY_USE_MULTIPLICATIVE_MODIFIERS, mUseMultiplicativeModifiers);
        w.keyValue(KEY_USE_MODIFYING_DICE_PLUS_ADDS, mUseModifyingDicePlusAdds);
        w.keyValue(KEY_DAMAGE_PROGRESSION, Enums.toId(mDamageProgression));
        w.keyValue(KEY_USE_SIMPLE_METRIC_CONVERSIONS, mUseSimpleMetricConversions);
        w.keyValue(KEY_SHOW_COLLEGE_IN_SPELLS, mShowCollegeInSpells);
        w.keyValue(KEY_SHOW_DIFFICULTY, mShowDifficulty);
        w.keyValue(KEY_SHOW_ADVANTAGE_MODIFIER_ADJ, mShowAdvantageModifierAdj);
        w.keyValue(KEY_SHOW_EQUIPMENT_MODIFIER_ADJ, mShowEquipmentModifierAdj);
        w.keyValue(KEY_SHOW_SPELL_ADJ, mShowSpellAdj);
        w.keyValue(KEY_USE_TITLE_IN_FOOTER, mUseTitleInFooter);
        w.key(KEY_PAGE);
        mPageSettings.toJSON(w);
        w.key(KEY_BLOCK_LAYOUT);
        w.startArray();
        for (String one : mBlockLayout) {
            w.value(one);
        }
        w.endArray();
        if (full) {
            w.key(KEY_ATTRIBUTES);
            AttributeDef.writeOrdered(w, mAttributes);
            w.key(KEY_HIT_LOCATIONS);
            mHitLocations.toJSON(w, mCharacter);
        }
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

    public DamageProgression getDamageProgression() {
        return mDamageProgression;
    }

    public void setDamageProgression(DamageProgression progression) {
        if (mDamageProgression != progression) {
            mDamageProgression = progression;
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
