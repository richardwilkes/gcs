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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageContainerType;
import com.trollworks.gcs.attribute.Attribute;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.attribute.AttributeType;
import com.trollworks.gcs.attribute.PoolThreshold;
import com.trollworks.gcs.attribute.ThresholdOps;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.feature.AttributeBonusLimitation;
import com.trollworks.gcs.feature.Bonus;
import com.trollworks.gcs.feature.CostReduction;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.feature.LeveledAmount;
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.gcs.feature.SkillPointBonus;
import com.trollworks.gcs.feature.SkillSelectionType;
import com.trollworks.gcs.feature.SpellBonus;
import com.trollworks.gcs.feature.SpellPointBonus;
import com.trollworks.gcs.feature.WeaponBonus;
import com.trollworks.gcs.feature.WeaponSelectionType;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.RitualMagicSpell;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.print.PrintManager;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.FilteredIterator;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.undo.StdUndoManager;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.utility.units.WeightValue;

import java.io.IOException;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Pattern;

/** A GURPS character. */
public class GURPSCharacter extends CollectedModels {
    private static final int    CURRENT_JSON_VERSION = 1;
    private static final String KEY_ROOT             = "character";
    private static final String KEY_ATTRIBUTES       = "attributes";
    private static final String KEY_CREATED_DATE     = "created_date";
    private static final String KEY_MODIFIED_DATE    = "modified_date";
    private static final String KEY_THIRD_PARTY_DATA = "third_party";
    private static final String KEY_TOTAL_POINTS     = "total_points";

    // TODO: Eliminate these deprecated keys after a suitable waiting period; added April 15, 2021
    private static final String KEY_DX        = "DX";
    private static final String KEY_FP_ADJ    = "FP_adj";
    private static final String KEY_FP_DAMAGE = "fp_damage";
    private static final String KEY_HP_ADJ    = "HP_adj";
    private static final String KEY_HP_DAMAGE = "hp_damage";
    private static final String KEY_HT        = "HT";
    private static final String KEY_IQ        = "IQ";
    private static final String KEY_MOVE_ADJ  = "move_adj";
    private static final String KEY_PER_ADJ   = "per_adj";
    private static final String KEY_SPEED_ADJ = "speed_adj";
    private static final String KEY_ST        = "ST";
    private static final String KEY_WILL_ADJ  = "will_adj";

    private static final Pattern UL_PATTERN = Pattern.compile("<ul>");

    private long                                mModifiedOn;
    private long                                mCreatedOn;
    private HashMap<String, ArrayList<Feature>> mFeatureMap;
    private JsonMap                             mThirdPartyData;
    private Map<String, Attribute>              mAttributes;
    private int                                 mLiftingStrengthBonus;
    private int                                 mStrikingStrengthBonus;
    private int                                 mDodgeBonus;
    private int                                 mParryBonus;
    private int                                 mBlockBonus;
    private int                                 mTotalPoints;
    private Settings                            mSettings;
    private Profile                             mProfile;
    private Armor                               mArmor;
    private WeightValue                         mCachedWeightCarried;
    private WeightValue                         mCachedWeightCarriedForSkills;
    private Fixed6                              mCachedWealthCarried;
    private Fixed6                              mCachedWealthNotCarried;
    private int                                 mCachedAttributePoints;
    private int                                 mCachedAdvantagePoints;
    private int                                 mCachedDisadvantagePoints;
    private int                                 mCachedQuirkPoints;
    private int                                 mCachedSkillPoints;
    private int                                 mCachedSpellPoints;
    private int                                 mCachedRacePoints;
    private PrintManager                        mPageSettings;
    private String                              mPageSettingsString;
    private boolean                             mSkillsUpdated;
    private boolean                             mSpellsUpdated;

    /** Creates a new character with only default values set. */
    public GURPSCharacter() {
        characterInitialize(Preferences.getInstance().autoFillProfile());
        calculateAll();
    }

    /**
     * Creates a new character from the specified file.
     *
     * @param path The path to load the data from.
     * @throws IOException if the data cannot be read or the file doesn't contain a valid character
     *                     sheet.
     */
    public GURPSCharacter(Path path) throws IOException {
        this();
        load(path);
    }

    private void characterInitialize(boolean full) {
        mSettings = new Settings(this);
        mFeatureMap = new HashMap<>();
        mTotalPoints = Preferences.getInstance().getInitialPoints();
        mAttributes = new HashMap<>();
        for (String attrID : mSettings.getAttributes().keySet()) {
            mAttributes.put(attrID, new Attribute(attrID));
        }
        mProfile = new Profile(this, full);
        mArmor = new Armor(this);
        mCachedWeightCarried = new WeightValue(Fixed6.ZERO, mSettings.defaultWeightUnits());
        mCachedWeightCarriedForSkills = new WeightValue(Fixed6.ZERO, mSettings.defaultWeightUnits());
        mPageSettings = Preferences.getInstance().getDefaultPageSettings();
        mPageSettingsString = "{}";
        if (mPageSettings != null) {
            mPageSettings = new PrintManager(mPageSettings);
            mPageSettingsString = mPageSettings.toString();
        }
        mModifiedOn = System.currentTimeMillis() / FieldFactory.TIMESTAMP_FACTOR;
        mCreatedOn = mModifiedOn;
    }

    public Map<String, Attribute> getAttributes() {
        return mAttributes;
    }

    @Override
    public void notifyOfChange() {
        setModifiedOn(System.currentTimeMillis() / FieldFactory.TIMESTAMP_FACTOR);
        super.notifyOfChange();
    }

    /** @return The page settings. May return {@code null} if no printer has been defined. */
    public PrintManager getPageSettings() {
        return mPageSettings;
    }

    public String getLastPageSettingsAsString() {
        return mPageSettingsString;
    }

    @Override
    public FileType getFileType() {
        return FileType.SHEET;
    }

    @Override
    public RetinaIcon getFileIcons() {
        return Images.GCS_FILE;
    }

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    public void recalculate() {
        updateSkills();
        processFeaturesAndPrereqs();
        calculateAll();
    }

    private void calculateAll() {
        calculateAttributePoints();
        calculateAdvantagePoints();
        calculateSkillPoints();
        calculateSpellPoints();
        calculateWeightAndWealthCarried(false);
        calculateWealthNotCarried(false);
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        characterInitialize(false);
        mSettings.load(m.getMap(Settings.KEY_ROOT));
        mCreatedOn = Numbers.extractDateTime(Numbers.DATE_TIME_STORED_FORMAT, m.getString(KEY_CREATED_DATE)) / FieldFactory.TIMESTAMP_FACTOR;
        mProfile.load(m.getMap(Profile.KEY_PROFILE));
        if (m.has(KEY_ATTRIBUTES)) {
            mAttributes = new HashMap<>();
            JsonArray a      = m.getArray(KEY_ATTRIBUTES);
            int       length = a.size();
            for (int i = 0; i < length; i++) {
                Attribute attr = new Attribute(a.getMap(i));
                mAttributes.put(attr.getID(), attr);
            }
        } else {
            for (String attrID : mSettings.getAttributes().keySet()) {
                Attribute attr = mAttributes.get(attrID);
                if (attr != null) {
                    switch (attrID) {
                    case "st":
                        attr.initTo(m.getInt(KEY_ST) - 10, 0);
                        break;
                    case "dx":
                        attr.initTo(m.getInt(KEY_DX) - 10, 0);
                        break;
                    case "iq":
                        attr.initTo(m.getInt(KEY_IQ) - 10, 0);
                        break;
                    case "ht":
                        attr.initTo(m.getInt(KEY_HT) - 10, 0);
                        break;
                    case "will":
                        attr.initTo(m.getInt(KEY_WILL_ADJ), 0);
                        break;
                    case "per":
                        attr.initTo(m.getInt(KEY_PER_ADJ), 0);
                        break;
                    case "basic_speed":
                        attr.initTo(m.getDouble(KEY_SPEED_ADJ), 0);
                        break;
                    case "basic_move":
                        attr.initTo(m.getInt(KEY_MOVE_ADJ), 0);
                        break;
                    case "hp":
                        attr.initTo(m.getInt(KEY_HP_ADJ), m.getInt(KEY_HP_DAMAGE));
                        break;
                    case "fp":
                        attr.initTo(m.getInt(KEY_FP_ADJ), m.getInt(KEY_FP_DAMAGE));
                        break;
                    default:
                        break;
                    }
                }
            }
        }
        mTotalPoints = m.getInt(KEY_TOTAL_POINTS);
        loadModels(m, state);
        if (mPageSettings != null && m.has(PrintManager.KEY_ROOT)) {
            mPageSettings = new PrintManager(m.getMap(PrintManager.KEY_ROOT));
            mPageSettingsString = mPageSettings.toString();
        }
        // Loop through the skills and update their levels. It is necessary to do this here and not
        // as they are loaded, since references to defaults won't work until the entire list is
        // available.
        for (Skill skill : getSkillsIterator()) {
            skill.updateLevel(false);
        }
        calculateAll();
        mThirdPartyData = m.getMap(KEY_THIRD_PARTY_DATA);
        mModifiedOn = Numbers.extractDateTime(Numbers.DATE_TIME_STORED_FORMAT, m.getString(KEY_MODIFIED_DATE)) / FieldFactory.TIMESTAMP_FACTOR; // Must be last
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        w.key(Settings.KEY_ROOT);
        mSettings.save(w);
        w.keyValue(KEY_CREATED_DATE, Numbers.formatDateTime(Numbers.DATE_TIME_STORED_FORMAT, mCreatedOn * FieldFactory.TIMESTAMP_FACTOR));
        w.keyValue(KEY_MODIFIED_DATE, Numbers.formatDateTime(Numbers.DATE_TIME_STORED_FORMAT, mModifiedOn * FieldFactory.TIMESTAMP_FACTOR));
        w.key(Profile.KEY_PROFILE);
        mProfile.save(w);
        w.key(KEY_ATTRIBUTES);
        w.startArray();
        for (AttributeDef def : AttributeDef.getOrdered(mSettings.getAttributes())) {
            Attribute attr = mAttributes.get(def.getID());
            if (attr != null) {
                attr.toJSON(this, w);
            }
        }
        w.endArray();
        w.keyValue(KEY_TOTAL_POINTS, mTotalPoints);
        saveModels(w, saveType);
        if (saveType != SaveType.HASH) {
            if (mPageSettings != null) {
                w.key(PrintManager.KEY_ROOT);
                mPageSettings.save(w, LengthUnits.IN);
                mPageSettingsString = mPageSettings.toString();
            }
            w.keyValueNotEmpty(KEY_THIRD_PARTY_DATA, mThirdPartyData);
        }
    }

    /** @return The created on date. */
    public long getCreatedOn() {
        return mCreatedOn;
    }

    /** @return The modified date, in seconds since midnight, January 1, 1970 UTC. */
    public long getModifiedOn() {
        return mModifiedOn;
    }

    /** @param whenInSeconds The modified date, in seconds since midnight, January 1, 1970 UTC. */
    public void setModifiedOn(long whenInSeconds) {
        if (mModifiedOn != whenInSeconds) {
            mModifiedOn = whenInSeconds;
            super.notifyOfChange(); // Skip local notifyOfChange, since it update the modified on value
        }
    }

    public void updateSkills() {
        for (Skill skill : getSkillsIterator()) {
            skill.updateLevel(true);
        }
        mSkillsUpdated = true;
    }

    private void updateSpells() {
        for (Spell spell : getSpellsIterator()) {
            spell.updateLevel(true);
        }
        mSpellsUpdated = true;
    }

    /** @return The current lifting strength bonus from features. */
    public int getLiftingStrengthBonus() {
        return mLiftingStrengthBonus;
    }

    /** @param bonus The new lifting strength bonus. */
    public void setLiftingStrengthBonus(int bonus) {
        if (mLiftingStrengthBonus != bonus) {
            mLiftingStrengthBonus = bonus;
            notifyOfChange();
        }
    }

    /** @return The current striking strength bonus from features. */
    public int getStrikingStrengthBonus() {
        return mStrikingStrengthBonus;
    }

    /** @param bonus The new striking strength bonus. */
    public void setStrikingStrengthBonus(int bonus) {
        if (mStrikingStrengthBonus != bonus) {
            mStrikingStrengthBonus = bonus;
            notifyOfChange();
        }
    }

    public int getAttributeCurrentIntValue(String name) {
        Attribute attr = getAttributes().get(name);
        if (attr == null) {
            return Integer.MIN_VALUE;
        }
        return attr.getCurrentIntValue(this);
    }

    public int getAttributeIntValue(String name) {
        Attribute attr = getAttributes().get(name);
        if (attr == null) {
            return Integer.MIN_VALUE;
        }
        return attr.getIntValue(this);
    }

    public double getAttributeDoubleValue(String name) {
        Attribute attr = getAttributes().get(name);
        if (attr == null) {
            return Integer.MIN_VALUE;
        }
        return attr.getDoubleValue(this);
    }

    public int getAttributeCost(String name) {
        Attribute attr = getAttributes().get(name);
        if (attr == null) {
            return 0;
        }
        return attr.getPointCost(this);
    }

    /** @return The basic thrusting damage. */
    public Dice getThrust() {
        return getThrust(getAttributeIntValue("st") + mStrikingStrengthBonus);
    }

    /**
     * @param strength The strength to return basic thrusting damage for.
     * @return The basic thrusting damage.
     */
    public Dice getThrust(int strength) {
        if (mSettings.useThrustEqualsSwingMinus2()) {
            Dice dice = getSwing(strength);
            dice.add(-2);
            return dice;
        }
        if (mSettings.useReducedSwing()) {
            if (strength < 19) {
                return new Dice(1, -(6 - (strength - 1) / 2));
            }
            int dice = 1;
            int adds = (strength - 10) / 2 - 2;
            if ((strength - 10) % 2 == 1) {
                adds++;
            }
            dice += 2 * (adds / 7);
            adds %= 7;
            dice += adds / 4;
            adds %= 4;
            if (adds == 3) {
                dice++;
                adds = -1;
            }

            return new Dice(dice, adds);
        }

        if (mSettings.useKnowYourOwnStrength()) {
            if (strength < 12) {
                return new Dice(1, strength - 12);
            }
            return new Dice((strength - 7) / 4, (strength + 1) % 4 - 1);
        }

        int value = strength;

        if (value < 19) {
            return new Dice(1, -(6 - (value - 1) / 2));
        }

        value -= 11;
        if (strength > 50) {
            value--;
            if (strength > 79) {
                value -= 1 + (strength - 80) / 5;
            }
        }
        return new Dice(value / 8 + 1, value % 8 / 2 - 1);
    }

    /** @return The basic swinging damage. */
    public Dice getSwing() {
        return getSwing(getAttributeIntValue("st") + mStrikingStrengthBonus);
    }

    /**
     * @param strength The strength to return basic swinging damage for.
     * @return The basic thrusting damage.
     */
    public Dice getSwing(int strength) {
        if (mSettings.useReducedSwing()) {
            if (strength < 10) {
                return new Dice(1, -(5 - (strength - 1) / 2));
            }

            int dice = 1;
            int adds = (strength - 10) / 2;
            dice += 2 * (adds / 7);
            adds %= 7;
            dice += adds / 4;
            adds %= 4;
            if (adds == 3) {
                dice++;
                adds = -1;
            }

            return new Dice(dice, adds);
        }

        if (mSettings.useKnowYourOwnStrength()) {
            if (strength < 10) {
                return new Dice(1, strength - 10);
            }
            return new Dice((strength - 5) / 4, (strength - 1) % 4 - 1);
        }

        int value = strength;

        if (value < 10) {
            return new Dice(1, -(5 - (value - 1) / 2));
        }

        if (value < 28) {
            value -= 9;
            return new Dice(value / 4 + 1, value % 4 - 1);
        }

        if (strength > 40) {
            value -= (strength - 40) / 5;
        }

        if (strength > 59) {
            value++;
        }
        value += 9;
        return new Dice(value / 8 + 1, value % 8 / 2 - 1);
    }

    /** @return Basic lift. */
    public WeightValue getBasicLift() {
        return getBasicLift(defaultWeightUnits());
    }

    private WeightValue getBasicLift(WeightUnits desiredUnits) {
        Fixed6      ten = new Fixed6(10);
        WeightUnits units;
        Fixed6      divisor;
        Fixed6      multiplier;
        Fixed6      roundAt;
        if (useSimpleMetricConversions() && defaultWeightUnits().isMetric()) {
            units = WeightUnits.KG;
            divisor = ten;
            multiplier = Fixed6.ONE;
            roundAt = new Fixed6(5);
        } else {
            units = WeightUnits.LB;
            divisor = new Fixed6(5);
            multiplier = new Fixed6(2);
            roundAt = ten;
        }
        int strength = getAttributeIntValue("st") + mLiftingStrengthBonus;
        if (isThresholdOpMet(ThresholdOps.HALVE_ST)) {
            boolean plusOne = strength % 2 != 0;
            strength /= 2;
            if (plusOne) {
                strength++;
            }
        }
        Fixed6 value;
        if (strength < 1) {
            value = Fixed6.ZERO;
        } else {
            if (mSettings.useKnowYourOwnStrength()) {
                int diff = 0;
                if (strength > 19) {
                    diff = strength / 10 - 1;
                    strength -= diff * 10;
                }
                value = new Fixed6(Math.pow(10.0, strength / 10.0)).mul(multiplier);
                value = strength <= 6 ? value.mul(ten).round().div(ten) : value.round();
                value = value.mul(new Fixed6(Math.pow(10, diff)));
            } else {
                //noinspection UnnecessaryExplicitNumericCast
                value = new Fixed6((long) strength * (long) strength).div(divisor);
            }
            if (value.greaterThanOrEqual(roundAt)) {
                value = value.round();
            }
            value = value.mul(ten).trunc().div(ten);
        }
        return new WeightValue(desiredUnits.convert(units, value), desiredUnits);
    }

    private WeightValue getMultipleOfBasicLift(int multiple) {
        WeightValue lift = getBasicLift();
        lift.setValue(lift.getValue().mul(new Fixed6(multiple)));
        return lift;
    }

    /** @return The one-handed lift value. */
    public WeightValue getOneHandedLift() {
        return getMultipleOfBasicLift(2);
    }

    /** @return The two-handed lift value. */
    public WeightValue getTwoHandedLift() {
        return getMultipleOfBasicLift(8);
    }

    /** @return The shove and knock over value. */
    public WeightValue getShoveAndKnockOver() {
        return getMultipleOfBasicLift(12);
    }

    /** @return The running shove and knock over value. */
    public WeightValue getRunningShoveAndKnockOver() {
        return getMultipleOfBasicLift(24);
    }

    /** @return The carry on back value. */
    public WeightValue getCarryOnBack() {
        return getMultipleOfBasicLift(15);
    }

    /** @return The shift slightly value. */
    public WeightValue getShiftSlightly() {
        return getMultipleOfBasicLift(50);
    }

    /**
     * @param encumbrance The encumbrance level.
     * @return The maximum amount the character can carry for the specified encumbrance level.
     */
    public WeightValue getMaximumCarry(Encumbrance encumbrance) {
        WeightUnits desiredUnits = defaultWeightUnits();
        WeightUnits calcUnits    = useSimpleMetricConversions() && desiredUnits.isMetric() ? WeightUnits.KG : WeightUnits.LB;
        WeightValue lift         = getBasicLift(calcUnits);
        lift.setValue(lift.getValue().mul(new Fixed6(encumbrance.getWeightMultiplier())));
        return new WeightValue(desiredUnits.convert(calcUnits, lift.getValue()), desiredUnits);
    }

    /**
     * @param encumbrance The encumbrance level.
     * @return The character's ground move for the specified encumbrance level.
     */
    public int getMove(Encumbrance encumbrance) {
        int initialMove = getAttributeIntValue("basic_move");
        int divisor     = 2 * Math.min(countThresholdOpMet(ThresholdOps.HALVE_MOVE), 2);
        if (divisor > 0) {
            boolean plusOne = (initialMove % divisor) != 0;
            initialMove /= divisor;
            if (plusOne) {
                initialMove++;
            }
        }
        int move = initialMove * (10 + 2 * encumbrance.getEncumbrancePenalty()) / 10;
        if (move < 1) {
            return initialMove > 0 ? 1 : 0;
        }
        return move;
    }

    /**
     * @param encumbrance The encumbrance level.
     * @return The character's dodge for the specified encumbrance level.
     */
    public int getDodge(Encumbrance encumbrance) {
        int dodge   = 3 + mDodgeBonus + getAttributeIntValue("basic_speed");
        int divisor = 2 * Math.min(countThresholdOpMet(ThresholdOps.HALVE_DODGE), 2);
        if (divisor > 0) {
            boolean plusOne = (dodge % divisor) != 0;
            dodge /= divisor;
            if (plusOne) {
                dodge++;
            }
        }
        return Math.max(dodge + encumbrance.getEncumbrancePenalty(), 1);
    }


    /** @return The dodge bonus. */
    public int getDodgeBonus() {
        return mDodgeBonus;
    }

    /** @param bonus The dodge bonus. */
    public void setDodgeBonus(int bonus) {
        if (mDodgeBonus != bonus) {
            mDodgeBonus = bonus;
            notifyOfChange();
        }
    }

    /** @return The parry bonus. */
    public int getParryBonus() {
        return mParryBonus;
    }

    /** @param bonus The parry bonus. */
    public void setParryBonus(int bonus) {
        if (mParryBonus != bonus) {
            mParryBonus = bonus;
            notifyOfChange();
        }
    }

    /** @return The block bonus. */
    public int getBlockBonus() {
        return mBlockBonus;
    }

    /** @param bonus The block bonus. */
    public void setBlockBonus(int bonus) {
        if (mBlockBonus != bonus) {
            mBlockBonus = bonus;
            notifyOfChange();
        }
    }

    /** @return The current encumbrance level. */
    public Encumbrance getEncumbranceLevel(boolean forSkills) {
        Fixed6 carried = getWeightCarried(forSkills).getNormalizedValue();
        for (Encumbrance encumbrance : Encumbrance.values()) {
            if (carried.lessThanOrEqual(getMaximumCarry(encumbrance).getNormalizedValue())) {
                return encumbrance;
            }
        }
        return Encumbrance.EXTRA_HEAVY;
    }

    /**
     * @return {@code true} if the carried weight is greater than the maximum allowed for an
     *         extra-heavy load.
     */
    public boolean isCarryingGreaterThanMaxLoad(boolean forSkills) {
        return getWeightCarried(forSkills).getNormalizedValue().greaterThan(getMaximumCarry(Encumbrance.EXTRA_HEAVY).getNormalizedValue());
    }

    /** @return The current weight being carried. */
    public WeightValue getWeightCarried(boolean forSkills) {
        return forSkills ? mCachedWeightCarriedForSkills : mCachedWeightCarried;
    }

    /** @return The current wealth being carried. */
    public Fixed6 getWealthCarried() {
        return mCachedWealthCarried;
    }

    /** @return The current wealth not being carried. */
    public Fixed6 getWealthNotCarried() {
        return mCachedWealthNotCarried;
    }

    /**
     * Convert a metric {@link WeightValue} by GURPS Metric rules into an imperial one. If an
     * imperial {@link WeightValue} is passed as an argument, it will be returned unchanged.
     *
     * @param value The {@link WeightValue} to be converted by GURPS Metric rules.
     * @return The converted imperial {@link WeightValue}.
     */
    public static WeightValue convertFromGurpsMetric(WeightValue value) {
        return switch (value.getUnits()) {
            case G -> new WeightValue(value.getValue().div(new Fixed6(30)), WeightUnits.OZ);
            case KG -> new WeightValue(value.getValue().mul(new Fixed6(2)), WeightUnits.LB);
            case T -> new WeightValue(value.getValue(), WeightUnits.LT);
            default -> value;
        };
    }

    /**
     * Convert an imperial {@link WeightValue} by GURPS Metric rules into a metric one. If a metric
     * {@link WeightValue} is passed as an argument, it will be returned unchanged.
     *
     * @param value The {@link WeightValue} to be converted by GURPS Metric rules.
     * @return The converted metric {@link WeightValue}.
     */
    public static WeightValue convertToGurpsMetric(WeightValue value) {
        return switch (value.getUnits()) {
            case LB -> new WeightValue(value.getValue().div(new Fixed6(2)), WeightUnits.KG);
            case LT, TN -> new WeightValue(value.getValue(), WeightUnits.T);
            case OZ -> new WeightValue(value.getValue().mul(new Fixed6(30)), WeightUnits.G);
            default -> value;
        };
    }

    /**
     * Calculate the total weight and wealth carried.
     *
     * @param notify Whether to send out notifications if the resulting values are different from
     *               the previous values.
     */
    public void calculateWeightAndWealthCarried(boolean notify) {
        WeightValue savedWeight          = new WeightValue(mCachedWeightCarried);
        WeightValue savedWeightForSkills = new WeightValue(mCachedWeightCarriedForSkills);
        Fixed6      savedWealth          = mCachedWealthCarried;
        mCachedWeightCarried = new WeightValue(Fixed6.ZERO, defaultWeightUnits());
        mCachedWeightCarriedForSkills = new WeightValue(Fixed6.ZERO, defaultWeightUnits());
        mCachedWealthCarried = Fixed6.ZERO;
        for (Row one : getEquipmentModel().getTopLevelRows()) {
            Equipment   equipment = (Equipment) one;
            WeightValue weight    = new WeightValue(equipment.getExtendedWeight(false));
            if (useSimpleMetricConversions()) {
                weight = defaultWeightUnits().isMetric() ? convertToGurpsMetric(weight) : convertFromGurpsMetric(weight);
            }
            mCachedWeightCarried.add(weight);
            mCachedWealthCarried = mCachedWealthCarried.add(equipment.getExtendedValue());

            weight = new WeightValue(equipment.getExtendedWeight(true));
            if (useSimpleMetricConversions()) {
                weight = defaultWeightUnits().isMetric() ? convertToGurpsMetric(weight) : convertFromGurpsMetric(weight);
            }
            mCachedWeightCarriedForSkills.add(weight);
        }
        if (notify) {
            if (!savedWeight.equals(mCachedWeightCarried) || !savedWeightForSkills.equals(mCachedWeightCarriedForSkills) || !mCachedWealthCarried.equals(savedWealth)) {
                notifyOfChange();
            }
        }
    }

    /**
     * Calculate the total wealth not carried.
     *
     * @param notify Whether to send out notifications if the resulting values are different from
     *               the previous values.
     */
    public void calculateWealthNotCarried(boolean notify) {
        Fixed6 savedWealth = mCachedWealthNotCarried;
        mCachedWealthNotCarried = Fixed6.ZERO;
        for (Row one : getOtherEquipmentModel().getTopLevelRows()) {
            mCachedWealthNotCarried = mCachedWealthNotCarried.add(((Equipment) one).getExtendedValue());
        }
        if (notify) {
            if (!mCachedWealthNotCarried.equals(savedWealth)) {
                notifyOfChange();
            }
        }
    }

    private int[] preserveMoveAndDodge() {
        Encumbrance[] values = Encumbrance.values();
        int[]         data   = new int[values.length * 2];
        for (Encumbrance encumbrance : values) {
            int index = encumbrance.ordinal();
            data[index] = getMove(encumbrance);
            data[values.length + index] = getDodge(encumbrance);
        }
        return data;
    }

    /** @return The total number of points this character has. */
    public int getTotalPoints() {
        return mTotalPoints;
    }

    /** @return The total number of points spent. */
    public int getSpentPoints() {
        return getAttributePoints() + getAdvantagePoints() + getDisadvantagePoints() + getQuirkPoints() + getSkillPoints() + getSpellPoints() + getRacePoints();
    }

    /** @return The number of unspent points. */
    public int getUnspentPoints() {
        return mTotalPoints - getSpentPoints();
    }

    /**
     * Sets the unspent character points.
     *
     * @param unspent The new unspent character points.
     */
    public void setUnspentPoints(int unspent) {
        int current = getUnspentPoints();
        if (current != unspent) {
            postUndoEdit(I18n.Text("Unspent Points Change"), (c, v) -> c.setUnspentPoints(((Integer) v).intValue()), Integer.valueOf(current), Integer.valueOf(unspent));
            mTotalPoints = unspent + getSpentPoints();
            notifyOfChange();
        }
    }

    /** @return The number of points spent on basic attributes. */
    public int getAttributePoints() {
        return mCachedAttributePoints;
    }

    private void calculateAttributePoints() {
        mCachedAttributePoints = 0;
        for (Attribute attr : mAttributes.values()) {
            mCachedAttributePoints += attr.getPointCost(this);
        }
    }

    /** @return The number of points spent on a racial package. */
    public int getRacePoints() {
        return mCachedRacePoints;
    }

    /** @return The number of points spent on advantages. */
    public int getAdvantagePoints() {
        return mCachedAdvantagePoints;
    }

    /** @return The number of points spent on disadvantages. */
    public int getDisadvantagePoints() {
        return mCachedDisadvantagePoints;
    }

    /** @return The number of points spent on quirks. */
    public int getQuirkPoints() {
        return mCachedQuirkPoints;
    }

    private void calculateAdvantagePoints() {
        mCachedAdvantagePoints = 0;
        mCachedDisadvantagePoints = 0;
        mCachedRacePoints = 0;
        mCachedQuirkPoints = 0;
        for (Advantage advantage : new FilteredIterator<>(getAdvantagesModel().getTopLevelRows(), Advantage.class)) {
            calculateSingleAdvantagePoints(advantage);
        }
    }

    private void calculateSingleAdvantagePoints(Advantage advantage) {
        if (advantage.canHaveChildren()) {
            AdvantageContainerType type = advantage.getContainerType();
            if (type == AdvantageContainerType.GROUP) {
                for (Advantage child : new FilteredIterator<>(advantage.getChildren(), Advantage.class)) {
                    calculateSingleAdvantagePoints(child);
                }
                return;
            } else if (type == AdvantageContainerType.RACE) {
                mCachedRacePoints += advantage.getAdjustedPoints();
                return;
            }
        }

        int pts = advantage.getAdjustedPoints();
        if (pts > 0) {
            mCachedAdvantagePoints += pts;
        } else if (pts < -1) {
            mCachedDisadvantagePoints += pts;
        } else if (pts == -1) {
            mCachedQuirkPoints--;
        }
    }

    /** @return The number of points spent on skills. */
    public int getSkillPoints() {
        return mCachedSkillPoints;
    }

    private void calculateSkillPoints() {
        mCachedSkillPoints = 0;
        for (Skill skill : getSkillsIterator()) {
            mCachedSkillPoints += skill.getRawPoints();
        }
    }

    /** @return The number of points spent on spells. */
    public int getSpellPoints() {
        return mCachedSpellPoints;
    }

    private void calculateSpellPoints() {
        mCachedSpellPoints = 0;
        for (Spell spell : getSpellsIterator()) {
            mCachedSpellPoints += spell.getRawPoints();
        }
    }

    /** @return The {@link Profile} data. */
    public Profile getProfile() {
        return mProfile;
    }

    public Settings getSettings() {
        return mSettings;
    }

    /** @return The {@link Armor} stats. */
    public Armor getArmor() {
        return mArmor;
    }

    /**
     * Searches the character's current advantages list for the specified name.
     *
     * @param name The name to look for.
     * @return The advantage, if present, or {@code null}.
     */
    public Advantage getAdvantageNamed(String name) {
        for (Advantage advantage : getAdvantagesIterator(false)) {
            if (advantage.getName().equals(name)) {
                return advantage;
            }
        }
        return null;
    }

    /**
     * Searches the character's current advantages list for the specified name.
     *
     * @param name The name to look for.
     * @return Whether it is present or not.
     */
    public boolean hasAdvantageNamed(String name) {
        return getAdvantageNamed(name) != null;
    }

    /**
     * Searches the character's current skill list for the specified name.
     *
     * @param name           The name to look for.
     * @param specialization The specialization to look for. Pass in {@code null} or an empty string
     *                       to ignore.
     * @param requirePoints  Only look at {@link Skill}s that have points. {@link Technique}s,
     *                       however, still won't need points even if this is {@code true}.
     * @param excludes       The set of {@link Skill}s to exclude from consideration.
     * @return The skill if it is present, or {@code null} if its not.
     */
    public List<Skill> getSkillNamed(String name, String specialization, boolean requirePoints, Set<String> excludes) {
        List<Skill> skills              = new ArrayList<>();
        boolean     checkSpecialization = specialization != null && !specialization.isEmpty();
        for (Skill skill : getSkillsIterator()) {
            if (!skill.canHaveChildren()) {
                if (excludes == null || !excludes.contains(skill.toString())) {
                    if (!requirePoints || skill instanceof Technique || skill.getPoints() > 0) {
                        if (skill.getName().equalsIgnoreCase(name)) {
                            if (!checkSpecialization || skill.getSpecialization().equalsIgnoreCase(specialization)) {
                                skills.add(skill);
                            }
                        }
                    }
                }
            }
        }
        return skills;
    }

    /**
     * Searches the character's current {@link Skill} list for the {@link Skill} with the best level
     * that matches the name.
     *
     * @param name           The {@link Skill} name to look for.
     * @param specialization An optional specialization to look for. Pass {@code null} if it is not
     *                       needed.
     * @param requirePoints  Only look at {@link Skill}s that have points. {@link Technique}s,
     *                       however, still won't need points even if this is {@code true}.
     * @param excludes       The set of {@link Skill}s to exclude from consideration.
     * @return The {@link Skill} that matches with the highest level.
     */
    public Skill getBestSkillNamed(String name, String specialization, boolean requirePoints, Set<String> excludes) {
        Skill best  = null;
        int   level = Integer.MIN_VALUE;
        for (Skill skill : getSkillNamed(name, specialization, requirePoints, excludes)) {
            int skillLevel = skill.getLevel(excludes);
            if (best == null || skillLevel > level) {
                best = skill;
                level = skillLevel;
            }
        }
        return best;
    }

    public boolean processFeaturesAndPrereqs() {
        boolean needRepaint = processFeatures();
        needRepaint |= processPrerequisites(getAdvantagesIterator(false));
        needRepaint |= processPrerequisites(getSkillsIterator());
        needRepaint |= processPrerequisites(getSpellsIterator());
        needRepaint |= processPrerequisites(getEquipmentIterator());
        needRepaint |= processPrerequisites(getOtherEquipmentIterator());
        return needRepaint;
    }

    private boolean processFeatures() {
        HashMap<String, ArrayList<Feature>> map         = new HashMap<>();
        boolean                             needRepaint = buildFeatureMap(map, getAdvantagesIterator(false));
        needRepaint |= buildFeatureMap(map, getSkillsIterator());
        needRepaint |= buildFeatureMap(map, getSpellsIterator());
        needRepaint |= buildFeatureMap(map, getEquipmentIterator());
        setFeatureMap(map);
        return needRepaint;
    }

    private boolean buildFeatureMap(HashMap<String, ArrayList<Feature>> map, Iterator<? extends ListRow> iterator) {
        boolean needRepaint = false;
        while (iterator.hasNext()) {
            ListRow row = iterator.next();
            if (row instanceof Equipment) {
                Equipment equipment = (Equipment) row;
                if (!equipment.isEquipped() || equipment.getQuantity() < 1) {
                    // Don't allow unequipped equipment to affect the character
                    continue;
                }
            }
            for (Feature feature : row.getFeatures()) {
                needRepaint |= processFeature(map, row instanceof Advantage ? ((Advantage) row).getLevels() : 0, feature);
                if (feature instanceof Bonus) {
                    ((Bonus) feature).setParent(row);
                }
            }
            if (row instanceof Advantage) {
                Advantage advantage = (Advantage) row;
                for (Bonus bonus : advantage.getCRAdj().getBonuses(advantage.getCR())) {
                    needRepaint |= processFeature(map, 0, bonus);
                    bonus.setParent(row);
                }
                for (AdvantageModifier modifier : advantage.getModifiers()) {
                    if (modifier.isEnabled()) {
                        for (Feature feature : modifier.getFeatures()) {
                            needRepaint |= processFeature(map, modifier.getLevels(), feature);
                            if (feature instanceof Bonus) {
                                ((Bonus) feature).setParent(row);
                            }
                        }
                    }
                }
            }
            if (row instanceof Equipment) {
                Equipment equipment = (Equipment) row;
                for (EquipmentModifier modifier : equipment.getModifiers()) {
                    if (modifier.isEnabled()) {
                        for (Feature feature : modifier.getFeatures()) {
                            needRepaint |= processFeature(map, 0, feature);
                            if (feature instanceof Bonus) {
                                ((Bonus) feature).setParent(row);
                            }
                        }
                    }
                }
            }
        }
        return needRepaint;
    }

    private boolean processFeature(HashMap<String, ArrayList<Feature>> map, int levels, Feature feature) {
        String             key         = feature.getKey().toLowerCase();
        ArrayList<Feature> list        = map.get(key);
        boolean            needRepaint = false;
        if (list == null) {
            list = new ArrayList<>(1);
            map.put(key, list);
        }
        if (feature instanceof Bonus) {
            LeveledAmount amount = ((Bonus) feature).getAmount();
            if (amount.getLevel() != levels) {
                amount.setLevel(levels);
                needRepaint = true;
            }
        }
        list.add(feature);
        return needRepaint;
    }

    private boolean processPrerequisites(Iterator<? extends ListRow> iterator) {
        boolean       needRepaint = false;
        StringBuilder builder     = new StringBuilder();
        while (iterator.hasNext()) {
            ListRow row = iterator.next();
            builder.setLength(0);
            boolean satisfied = row.getPrereqs().satisfied(this, row, builder, "<li>");
            if (satisfied && row instanceof Technique) {
                satisfied = ((Technique) row).satisfied(builder, "<li>");
            }
            if (satisfied && row instanceof RitualMagicSpell) {
                satisfied = ((RitualMagicSpell) row).satisfied(builder, "<li>");
            }
            if (row.isSatisfied() != satisfied) {
                row.setSatisfied(satisfied);
                needRepaint = true;
            }
            if (!satisfied) {
                builder.insert(0, "<html><body>" + I18n.Text("Reason:") + "<ul>");
                builder.append("</ul></body></html>");
                row.setReasonForUnsatisfied(UL_PATTERN.matcher(builder.toString()).replaceAll("<ul style='margin-top: 0; margin-bottom: 0;'>"));
            }
        }
        return needRepaint;
    }

    /** @param map The new feature map. */
    public void setFeatureMap(HashMap<String, ArrayList<Feature>> map) {
        mFeatureMap = map;
        mSkillsUpdated = false;
        mSpellsUpdated = false;
        String strPrefix = Attribute.ID_ATTR_PREFIX + "st.";
        setLiftingStrengthBonus(getIntegerBonusFor(strPrefix + AttributeBonusLimitation.LIFTING_ONLY.name()));
        setStrikingStrengthBonus(getIntegerBonusFor(strPrefix + AttributeBonusLimitation.STRIKING_ONLY.name()));
        for (Attribute attr : mAttributes.values()) {
            String       attrID = Attribute.ID_ATTR_PREFIX + attr.getID();
            AttributeDef def    = attr.getAttrDef(this);
            if (def != null) {
                if (def.getType() == AttributeType.DECIMAL) {
                    attr.setBonus(this, getDoubleBonusFor(attrID));
                } else {
                    attr.setBonus(this, getIntegerBonusFor(attrID));
                }
                attr.setCostReduction(this, getCostReductionFor(attrID));
            }
        }
        mProfile.update();
        setDodgeBonus(getIntegerBonusFor(Attribute.ID_ATTR_PREFIX + "dodge"));
        setParryBonus(getIntegerBonusFor(Attribute.ID_ATTR_PREFIX + "parry"));
        setBlockBonus(getIntegerBonusFor(Attribute.ID_ATTR_PREFIX + "block"));
        mArmor.update();
        if (!mSkillsUpdated) {
            updateSkills();
        }
        if (!mSpellsUpdated) {
            updateSpells();
        }
    }

    /**
     * @param id The cost reduction ID to search for.
     * @return The cost reduction, as a percentage.
     */
    public int getCostReductionFor(String id) {
        int           total = 0;
        List<Feature> list  = mFeatureMap.get(id.toLowerCase());

        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof CostReduction) {
                    total += ((CostReduction) feature).getPercentage();
                }
            }
        }
        if (total > 80) {
            total = 80;
        }
        return total;
    }

    /**
     * @param id The feature ID to search for.
     * @return The bonus.
     */
    public int getIntegerBonusFor(String id) {
        return getIntegerBonusFor(id, null);
    }

    /**
     * @param id      The feature ID to search for.
     * @param toolTip The toolTip being built.
     * @return The bonus.
     */
    public int getIntegerBonusFor(String id, StringBuilder toolTip) {
        int           total = 0;
        List<Feature> list  = mFeatureMap.get(id.toLowerCase());
        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof Bonus && !(feature instanceof WeaponBonus)) {
                    Bonus bonus = (Bonus) feature;
                    total += bonus.getAmount().getIntegerAdjustedAmount();
                    bonus.addToToolTip(toolTip);
                }
            }
        }
        return total;
    }

    /**
     * @param id                      The feature ID to search for.
     * @param nameQualifier           The name qualifier.
     * @param specializationQualifier The specialization qualifier.
     * @param categoriesQualifier     The categories qualifier.
     * @param dieCount                The number of dice for the base weapon damage.
     * @param toolTip                 A buffer to write a tooltip into. May be null.
     * @return The bonuses.
     */
    public List<WeaponBonus> getWeaponComparedBonusesFor(String id, String nameQualifier, String specializationQualifier, Set<String> categoriesQualifier, int dieCount, StringBuilder toolTip) {
        List<WeaponBonus> bonuses = new ArrayList<>();
        int               rsl     = Integer.MIN_VALUE;
        for (Skill skill : getSkillNamed(nameQualifier, specializationQualifier, true, null)) {
            int srsl = skill.getRelativeLevel();
            if (srsl > rsl) {
                rsl = srsl;
            }
        }
        if (rsl != Integer.MIN_VALUE) {
            List<Feature> list = mFeatureMap.get(id.toLowerCase());
            if (list != null) {
                for (Feature feature : list) {
                    if (feature instanceof WeaponBonus) {
                        WeaponBonus bonus = (WeaponBonus) feature;
                        if (bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(specializationQualifier) && bonus.getRelativeLevelCriteria().matches(rsl) && bonus.matchesCategories(categoriesQualifier)) {
                            bonuses.add(bonus);
                            LeveledAmount amount = bonus.getAmount();
                            int           level  = amount.getLevel();
                            amount.setLevel(dieCount);
                            bonus.addToToolTip(toolTip);
                            amount.setLevel(level);
                        }
                    }
                }
            }
        }
        return bonuses;
    }

    /**
     * @param id                  The feature ID to search for.
     * @param nameQualifier       The name qualifier.
     * @param usageQualifier      The usage qualifier.
     * @param categoriesQualifier The categories qualifier.
     * @param dieCount            The number of dice for the base weapon damage.
     * @param toolTip             A buffer to write a tooltip into. May be null.
     * @return The bonuses.
     */
    public List<WeaponBonus> getNamedWeaponBonusesFor(String id, String nameQualifier, String usageQualifier, Set<String> categoriesQualifier, int dieCount, StringBuilder toolTip) {
        List<WeaponBonus> bonuses = new ArrayList<>();
        List<Feature>     list    = mFeatureMap.get(id.toLowerCase());
        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof WeaponBonus) {
                    WeaponBonus bonus = (WeaponBonus) feature;
                    if (bonus.getWeaponSelectionType() == WeaponSelectionType.WEAPONS_WITH_NAME && bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(usageQualifier) && bonus.matchesCategories(categoriesQualifier)) {
                        bonuses.add(bonus);
                        LeveledAmount amount = bonus.getAmount();
                        int           level  = amount.getLevel();
                        amount.setLevel(dieCount);
                        bonus.addToToolTip(toolTip);
                        amount.setLevel(level);
                    }
                }
            }
        }
        return bonuses;
    }

    /**
     * @param id                  The feature ID to search for.
     * @param nameQualifier       The name qualifier.
     * @param usageQualifier      The usage qualifier.
     * @param categoriesQualifier The categories qualifier.
     * @return The bonuses.
     */
    public List<SkillBonus> getNamedWeaponSkillBonusesFor(String id, String nameQualifier, String usageQualifier, Set<String> categoriesQualifier, StringBuilder toolTip) {
        List<SkillBonus> bonuses = new ArrayList<>();
        List<Feature>    list    = mFeatureMap.get(id.toLowerCase());
        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof SkillBonus) {
                    SkillBonus bonus = (SkillBonus) feature;
                    if (bonus.getSkillSelectionType() == SkillSelectionType.WEAPONS_WITH_NAME && bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(usageQualifier) && bonus.matchesCategories(categoriesQualifier)) {
                        bonuses.add(bonus);
                        bonus.addToToolTip(toolTip);
                    }
                }
            }
        }
        return bonuses;
    }

    /**
     * @param id                      The feature ID to search for.
     * @param nameQualifier           The name qualifier.
     * @param specializationQualifier The specialization qualifier.
     * @param categoryQualifier       The categories qualifier
     * @return The bonus.
     */
    public int getSkillComparedIntegerBonusFor(String id, String nameQualifier, String specializationQualifier, Set<String> categoryQualifier) {
        return getSkillComparedIntegerBonusFor(id, nameQualifier, specializationQualifier, categoryQualifier, null);
    }

    /**
     * @param id                      The feature ID to search for.
     * @param nameQualifier           The name qualifier.
     * @param specializationQualifier The specialization qualifier.
     * @param categoryQualifier       The categories qualifier
     * @param toolTip                 The toolTip being built
     * @return The bonus.
     */
    public int getSkillComparedIntegerBonusFor(String id, String nameQualifier, String specializationQualifier, Set<String> categoryQualifier, StringBuilder toolTip) {
        int           total = 0;
        List<Feature> list  = mFeatureMap.get(id.toLowerCase());
        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof SkillBonus) {
                    SkillBonus bonus = (SkillBonus) feature;
                    if (bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(specializationQualifier) && bonus.matchesCategories(categoryQualifier)) {
                        total += bonus.getAmount().getIntegerAdjustedAmount();
                        bonus.addToToolTip(toolTip);
                    }
                }
            }
        }
        return total;
    }

    /**
     * @param id                      The feature ID to search for.
     * @param nameQualifier           The name qualifier.
     * @param specializationQualifier The specialization qualifier.
     * @param categoryQualifier       The categories qualifier
     * @return The bonus.
     */
    public int getSkillPointComparedIntegerBonusFor(String id, String nameQualifier, String specializationQualifier, Set<String> categoryQualifier) {
        return getSkillPointComparedIntegerBonusFor(id, nameQualifier, specializationQualifier, categoryQualifier, null);
    }

    /**
     * @param id                      The feature ID to search for.
     * @param nameQualifier           The name qualifier.
     * @param specializationQualifier The specialization qualifier.
     * @param categoryQualifier       The categories qualifier
     * @param toolTip                 The toolTip being built
     * @return The point bonus.
     */
    public int getSkillPointComparedIntegerBonusFor(String id, String nameQualifier, String specializationQualifier, Set<String> categoryQualifier, StringBuilder toolTip) {
        int           total = 0;
        List<Feature> list  = mFeatureMap.get(id.toLowerCase());
        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof SkillPointBonus) {
                    SkillPointBonus bonus = (SkillPointBonus) feature;
                    if (bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(specializationQualifier) && bonus.matchesCategories(categoryQualifier)) {
                        total += bonus.getAmount().getIntegerAdjustedAmount();
                        bonus.addToToolTip(toolTip);
                    }
                }
            }
        }
        return total;
    }

    /**
     * @param id         The feature ID to search for.
     * @param qualifier  The qualifier.
     * @param categories The categories qualifier
     * @return The bonus.
     */
    public int getSpellComparedIntegerBonusFor(String id, String qualifier, Set<String> categories, StringBuilder toolTip) {
        int           total = 0;
        List<Feature> list  = mFeatureMap.get(id.toLowerCase());
        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof SpellBonus) {
                    SpellBonus bonus = (SpellBonus) feature;
                    if (bonus.getNameCriteria().matches(qualifier) && bonus.matchesCategories(categories)) {
                        total += bonus.getAmount().getIntegerAdjustedAmount();
                        bonus.addToToolTip(toolTip);
                    }
                }
            }
        }
        return total;
    }

    /**
     * @param id         The feature ID to search for.
     * @param qualifier  The qualifier.
     * @param categories The categories qualifier
     * @return The bonus.
     */
    public int getSpellPointComparedIntegerBonusFor(String id, String qualifier, Set<String> categories) {
        return getSpellPointComparedIntegerBonusFor(id, qualifier, categories, null);
    }

    /**
     * @param id         The feature ID to search for.
     * @param qualifier  The qualifier.
     * @param categories The categories qualifier
     * @param toolTip    The toolTip being built
     * @return The point bonus.
     */
    public int getSpellPointComparedIntegerBonusFor(String id, String qualifier, Set<String> categories, StringBuilder toolTip) {
        int           total = 0;
        List<Feature> list  = mFeatureMap.get(id.toLowerCase());
        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof SpellPointBonus) {
                    SpellPointBonus bonus = (SpellPointBonus) feature;
                    if (bonus.getNameCriteria().matches(qualifier) && bonus.matchesCategories(categories)) {
                        total += bonus.getAmount().getIntegerAdjustedAmount();
                        bonus.addToToolTip(toolTip);
                    }
                }
            }
        }
        return total;
    }

    /**
     * @param id The feature ID to search for.
     * @return The bonus.
     */
    public double getDoubleBonusFor(String id) {
        double        total = 0;
        List<Feature> list  = mFeatureMap.get(id.toLowerCase());
        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof Bonus && !(feature instanceof WeaponBonus)) {
                    total += ((Bonus) feature).getAmount().getAdjustedAmount();
                }
            }
        }
        return total;
    }

    /**
     * Post an undo edit if we're not currently in an undo.
     *
     * @param name   The name of the undo.
     * @param setter The field setter.
     * @param before The original value.
     * @param after  The new value.
     */
    public void postUndoEdit(String name, CharacterSetter setter, Object before, Object after) {
        StdUndoManager mgr = getUndoManager();
        if (!mgr.isInTransaction() && !before.equals(after)) {
            addEdit(new CharacterUndo(this, name, setter, before, after));
        }
    }

    @Override
    public WeightUnits defaultWeightUnits() {
        return mSettings.defaultWeightUnits();
    }

    @Override
    public boolean useSimpleMetricConversions() {
        return mSettings.useSimpleMetricConversions();
    }

    @Override
    public boolean useMultiplicativeModifiers() {
        return mSettings.useMultiplicativeModifiers();
    }

    @Override
    public boolean useModifyingDicePlusAdds() {
        return mSettings.useModifyingDicePlusAdds();
    }

    @Override
    public DisplayOption userDescriptionDisplay() {
        return mSettings.userDescriptionDisplay();
    }

    @Override
    public DisplayOption modifiersDisplay() {
        return mSettings.modifiersDisplay();
    }

    @Override
    public DisplayOption notesDisplay() {
        return mSettings.notesDisplay();
    }

    public boolean isThresholdOpMet(ThresholdOps op) {
        for (Attribute attr : mAttributes.values()) {
            PoolThreshold threshold = attr.getCurrentThreshold(this);
            if (threshold != null && threshold.getOps().contains(op)) {
                return true;
            }
        }
        return false;
    }

    public int countThresholdOpMet(ThresholdOps op) {
        int total = 0;
        for (Attribute attr : mAttributes.values()) {
            PoolThreshold threshold = attr.getCurrentThreshold(this);
            if (threshold != null && threshold.getOps().contains(op)) {
                total++;
            }
        }
        return total;
    }

    @Override
    public Map<String, AttributeDef> getAttributeDefs() {
        return getSettings().getAttributes();
    }
}
