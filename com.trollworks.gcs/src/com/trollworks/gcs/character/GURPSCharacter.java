/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.feature.AttributeBonusLimitation;
import com.trollworks.gcs.feature.Bonus;
import com.trollworks.gcs.feature.BonusAttributeType;
import com.trollworks.gcs.feature.CostReduction;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.feature.LeveledAmount;
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.gcs.feature.SkillSelectionType;
import com.trollworks.gcs.feature.SpellBonus;
import com.trollworks.gcs.feature.WeaponBonus;
import com.trollworks.gcs.feature.WeaponSelectionType;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.RitualMagicSpell;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.print.PrintManager;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowIterator;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.FilteredIterator;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.undo.StdUndoManager;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.utility.units.WeightValue;
import com.trollworks.gcs.utility.xml.XMLNodeType;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.nio.file.Path;
import java.text.DateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.List;
import java.util.Set;

/** A GURPS character. */
public class GURPSCharacter extends DataFile {
    private static final int                                 CURRENT_JSON_VERSION                 = 1;
    private static final int                                 CURRENT_VERSION                      = 5;
    /**
     * The version where equipment was separated out into different lists based on carried/not
     * carried status.
     */
    public static final  int                                 SEPARATED_EQUIPMENT_VERSION          = 4;
    /**
     * The version where HP and FP damage tracking was introduced, rather than a free-form text
     * field.
     */
    public static final  int                                 HP_FP_DAMAGE_TRACKING                = 5;
    private static final String                              TAG_ROOT                             = "character";
    private static final String                              TAG_CREATED_DATE                     = "created_date";
    private static final String                              TAG_MODIFIED_DATE                    = "modified_date";
    private static final String                              TAG_HP_DAMAGE                        = "hp_damage";
    private static final String                              TAG_FP_DAMAGE                        = "fp_damage";
    private static final String                              TAG_UNSPENT_POINTS                   = "unspent_points";
    private static final String                              TAG_TOTAL_POINTS                     = "total_points";
    private static final String                              TAG_INCLUDE_PUNCH                    = "include_punch";
    private static final String                              TAG_INCLUDE_KICK                     = "include_kick";
    private static final String                              TAG_INCLUDE_BOOTS                    = "include_kick_with_boots";
    private static final String                              KEY_HP_ADJ                           = "HP_adj";
    private static final String                              KEY_FP_ADJ                           = "FP_adj";
    private static final String                              KEY_ST                               = "ST";
    private static final String                              KEY_DX                               = "DX";
    private static final String                              KEY_IQ                               = "IQ";
    private static final String                              KEY_HT                               = "HT";
    private static final String                              KEY_WILL_ADJ                         = "will_adj";
    private static final String                              KEY_PER_ADJ                          = "per_adj";
    private static final String                              KEY_SPEED_ADJ                        = "speed_adj";
    private static final String                              KEY_MOVE_ADJ                         = "move_adj";
    public static final  String                              KEY_ADVANTAGES                       = "advantages";
    public static final  String                              KEY_SKILLS                           = "skills";
    public static final  String                              KEY_SPELLS                           = "spells";
    public static final  String                              KEY_EQUIPMENT                        = "equipment";
    public static final  String                              KEY_OTHER_EQUIPMENT                  = "other_equipment";
    public static final  String                              KEY_NOTES                            = "notes";
    /** The prefix for all character IDs. */
    public static final  String                              CHARACTER_PREFIX                     = "gcs.";
    /** The field ID for last modified date changes. */
    public static final  String                              ID_MODIFIED                          = CHARACTER_PREFIX + "Modified";
    /** The field ID for created on date changes. */
    public static final  String                              ID_CREATED                           = CHARACTER_PREFIX + "Created";
    /**
     * The prefix used to indicate a point value is requested from {@link #getValueForID(String)}.
     */
    public static final  String                              POINTS_PREFIX                        = CHARACTER_PREFIX + "points.";
    /** The prefix used in front of all IDs for basic attributes. */
    public static final  String                              ATTRIBUTES_PREFIX                    = CHARACTER_PREFIX + "ba.";
    /** The field ID for strength (ST) changes. */
    public static final  String                              ID_STRENGTH                          = ATTRIBUTES_PREFIX + BonusAttributeType.ST.name();
    /** The field ID for lifting strength bonuses -- used by features. */
    public static final  String                              ID_LIFTING_STRENGTH                  = ID_STRENGTH + AttributeBonusLimitation.LIFTING_ONLY.name();
    /** The field ID for striking strength bonuses -- used by features. */
    public static final  String                              ID_STRIKING_STRENGTH                 = ID_STRENGTH + AttributeBonusLimitation.STRIKING_ONLY.name();
    /** The field ID for dexterity (DX) changes. */
    public static final  String                              ID_DEXTERITY                         = ATTRIBUTES_PREFIX + BonusAttributeType.DX.name();
    /** The field ID for intelligence (IQ) changes. */
    public static final  String                              ID_INTELLIGENCE                      = ATTRIBUTES_PREFIX + BonusAttributeType.IQ.name();
    /** The field ID for health (HT) changes. */
    public static final  String                              ID_HEALTH                            = ATTRIBUTES_PREFIX + BonusAttributeType.HT.name();
    /** The field ID for perception changes. */
    public static final  String                              ID_PERCEPTION                        = ATTRIBUTES_PREFIX + BonusAttributeType.PERCEPTION.name();
    /** The field ID for vision changes. */
    public static final  String                              ID_VISION                            = ATTRIBUTES_PREFIX + BonusAttributeType.VISION.name();
    /** The field ID for hearing changes. */
    public static final  String                              ID_HEARING                           = ATTRIBUTES_PREFIX + BonusAttributeType.HEARING.name();
    /** The field ID for taste changes. */
    public static final  String                              ID_TASTE_AND_SMELL                   = ATTRIBUTES_PREFIX + BonusAttributeType.TASTE_SMELL.name();
    /** The field ID for smell changes. */
    public static final  String                              ID_TOUCH                             = ATTRIBUTES_PREFIX + BonusAttributeType.TOUCH.name();
    /** The field ID for will changes. */
    public static final  String                              ID_WILL                              = ATTRIBUTES_PREFIX + BonusAttributeType.WILL.name();
    /** The field ID for fright check changes. */
    public static final  String                              ID_FRIGHT_CHECK                      = ATTRIBUTES_PREFIX + BonusAttributeType.FRIGHT_CHECK.name();
    /** The field ID for basic speed changes. */
    public static final  String                              ID_BASIC_SPEED                       = ATTRIBUTES_PREFIX + BonusAttributeType.SPEED.name();
    /** The field ID for basic move changes. */
    public static final  String                              ID_BASIC_MOVE                        = ATTRIBUTES_PREFIX + BonusAttributeType.MOVE.name();
    /** The prefix used in front of all IDs for dodge changes. */
    public static final  String                              DODGE_PREFIX                         = ATTRIBUTES_PREFIX + BonusAttributeType.DODGE.name() + "#.";
    /** The field ID for dodge bonus changes. */
    public static final  String                              ID_DODGE_BONUS                       = ATTRIBUTES_PREFIX + BonusAttributeType.DODGE.name();
    /** The field ID for parry bonus changes. */
    public static final  String                              ID_PARRY_BONUS                       = ATTRIBUTES_PREFIX + BonusAttributeType.PARRY.name();
    /** The field ID for block bonus changes. */
    public static final  String                              ID_BLOCK_BONUS                       = ATTRIBUTES_PREFIX + BonusAttributeType.BLOCK.name();
    /** The prefix used in front of all IDs for move changes. */
    public static final  String                              MOVE_PREFIX                          = ATTRIBUTES_PREFIX + BonusAttributeType.MOVE.name() + "#.";
    /** The field ID for carried weight changes. */
    public static final  String                              ID_CARRIED_WEIGHT                    = CHARACTER_PREFIX + "CarriedWeight";
    /** The field ID for carried wealth changes. */
    public static final  String                              ID_CARRIED_WEALTH                    = CHARACTER_PREFIX + "CarriedWealth";
    /** The field ID for other wealth changes. */
    public static final  String                              ID_NOT_CARRIED_WEALTH                = CHARACTER_PREFIX + "NotCarriedWealth";
    /** The prefix used in front of all IDs for encumbrance changes. */
    public static final  String                              MAXIMUM_CARRY_PREFIX                 = ATTRIBUTES_PREFIX + "MaximumCarry";
    private static final String                              LIFT_PREFIX                          = ATTRIBUTES_PREFIX + "lift.";
    /** The field ID for basic lift changes. */
    public static final  String                              ID_BASIC_LIFT                        = LIFT_PREFIX + "BasicLift";
    /** The field ID for one-handed lift changes. */
    public static final  String                              ID_ONE_HANDED_LIFT                   = LIFT_PREFIX + "OneHandedLift";
    /** The field ID for two-handed lift changes. */
    public static final  String                              ID_TWO_HANDED_LIFT                   = LIFT_PREFIX + "TwoHandedLift";
    /** The field ID for shove and knock over changes. */
    public static final  String                              ID_SHOVE_AND_KNOCK_OVER              = LIFT_PREFIX + "ShoveAndKnockOver";
    /** The field ID for running shove and knock over changes. */
    public static final  String                              ID_RUNNING_SHOVE_AND_KNOCK_OVER      = LIFT_PREFIX + "RunningShoveAndKnockOver";
    /** The field ID for carry on back changes. */
    public static final  String                              ID_CARRY_ON_BACK                     = LIFT_PREFIX + "CarryOnBack";
    /** The field ID for carry on back changes. */
    public static final  String                              ID_SHIFT_SLIGHTLY                    = LIFT_PREFIX + "ShiftSlightly";
    /** The prefix used in front of all IDs for point summaries. */
    public static final  String                              POINT_SUMMARY_PREFIX                 = CHARACTER_PREFIX + "ps.";
    /** The field ID for point total changes. */
    public static final  String                              ID_TOTAL_POINTS                      = POINT_SUMMARY_PREFIX + "TotalPoints";
    /** The field ID for attribute point summary changes. */
    public static final  String                              ID_ATTRIBUTE_POINTS                  = POINT_SUMMARY_PREFIX + "AttributePoints";
    /** The field ID for advantage point summary changes. */
    public static final  String                              ID_ADVANTAGE_POINTS                  = POINT_SUMMARY_PREFIX + "AdvantagePoints";
    /** The field ID for disadvantage point summary changes. */
    public static final  String                              ID_DISADVANTAGE_POINTS               = POINT_SUMMARY_PREFIX + "DisadvantagePoints";
    /** The field ID for quirk point summary changes. */
    public static final  String                              ID_QUIRK_POINTS                      = POINT_SUMMARY_PREFIX + "QuirkPoints";
    /** The field ID for skill point summary changes. */
    public static final  String                              ID_SKILL_POINTS                      = POINT_SUMMARY_PREFIX + "SkillPoints";
    /** The field ID for spell point summary changes. */
    public static final  String                              ID_SPELL_POINTS                      = POINT_SUMMARY_PREFIX + "SpellPoints";
    /** The field ID for racial point summary changes. */
    public static final  String                              ID_RACE_POINTS                       = POINT_SUMMARY_PREFIX + "RacePoints";
    /** The field ID for unspent point changes. */
    public static final  String                              ID_UNSPENT_POINTS                    = POINT_SUMMARY_PREFIX + "UnspentPoints";
    /** The prefix used in front of all IDs for basic damage. */
    public static final  String                              BASIC_DAMAGE_PREFIX                  = CHARACTER_PREFIX + "bd.";
    /** The field ID for basic thrust damage changes. */
    public static final  String                              ID_BASIC_THRUST                      = BASIC_DAMAGE_PREFIX + "Thrust";
    /** The field ID for basic swing damage changes. */
    public static final  String                              ID_BASIC_SWING                       = BASIC_DAMAGE_PREFIX + "Swing";
    private static final String                              HIT_POINTS_PREFIX                    = ATTRIBUTES_PREFIX + "derived_hp.";
    /** The field ID for hit point changes. */
    public static final  String                              ID_HIT_POINTS                        = ATTRIBUTES_PREFIX + BonusAttributeType.HP.name();
    /** The field ID for hit point damage changes. */
    public static final  String                              ID_HIT_POINTS_DAMAGE                 = HIT_POINTS_PREFIX + "Damage";
    /** The field ID for current hit point changes. */
    public static final  String                              ID_CURRENT_HP                        = HIT_POINTS_PREFIX + "Current";
    /** The field ID for reeling hit point changes. */
    public static final  String                              ID_REELING_HIT_POINTS                = HIT_POINTS_PREFIX + "Reeling";
    /** The field ID for unconscious check hit point changes. */
    public static final  String                              ID_UNCONSCIOUS_CHECKS_HIT_POINTS     = HIT_POINTS_PREFIX + "UnconsciousChecks";
    /** The field ID for death check #1 hit point changes. */
    public static final  String                              ID_DEATH_CHECK_1_HIT_POINTS          = HIT_POINTS_PREFIX + "DeathCheck1";
    /** The field ID for death check #2 hit point changes. */
    public static final  String                              ID_DEATH_CHECK_2_HIT_POINTS          = HIT_POINTS_PREFIX + "DeathCheck2";
    /** The field ID for death check #3 hit point changes. */
    public static final  String                              ID_DEATH_CHECK_3_HIT_POINTS          = HIT_POINTS_PREFIX + "DeathCheck3";
    /** The field ID for death check #4 hit point changes. */
    public static final  String                              ID_DEATH_CHECK_4_HIT_POINTS          = HIT_POINTS_PREFIX + "DeathCheck4";
    /** The field ID for dead hit point changes. */
    public static final  String                              ID_DEAD_HIT_POINTS                   = HIT_POINTS_PREFIX + "Dead";
    private static final String                              FATIGUE_POINTS_PREFIX                = ATTRIBUTES_PREFIX + "derived_fp.";
    /** The field ID for fatigue point changes. */
    public static final  String                              ID_FATIGUE_POINTS                    = ATTRIBUTES_PREFIX + BonusAttributeType.FP.name();
    /** The field ID for fatigue point damage changes. */
    public static final  String                              ID_FATIGUE_POINTS_DAMAGE             = FATIGUE_POINTS_PREFIX + "Damage";
    /** The field ID for current fatigue point changes. */
    public static final  String                              ID_CURRENT_FP                        = FATIGUE_POINTS_PREFIX + "Current";
    /** The field ID for tired fatigue point changes. */
    public static final  String                              ID_TIRED_FATIGUE_POINTS              = FATIGUE_POINTS_PREFIX + "Tired";
    /** The field ID for unconscious check fatigue point changes. */
    public static final  String                              ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS = FATIGUE_POINTS_PREFIX + "UnconsciousChecks";
    /** The field ID for unconscious fatigue point changes. */
    public static final  String                              ID_UNCONSCIOUS_FATIGUE_POINTS        = FATIGUE_POINTS_PREFIX + "Unconscious";
    private              long                                mModifiedOn;
    private              long                                mCreatedOn;
    private              HashMap<String, ArrayList<Feature>> mFeatureMap;
    private              int                                 mStrength;
    private              int                                 mStrengthBonus;
    private              int                                 mLiftingStrengthBonus;
    private              int                                 mStrikingStrengthBonus;
    private              int                                 mStrengthCostReduction;
    private              int                                 mDexterity;
    private              int                                 mDexterityBonus;
    private              int                                 mDexterityCostReduction;
    private              int                                 mIntelligence;
    private              int                                 mIntelligenceBonus;
    private              int                                 mIntelligenceCostReduction;
    private              int                                 mHealth;
    private              int                                 mHealthBonus;
    private              int                                 mHealthCostReduction;
    private              int                                 mWillAdj;
    private              int                                 mWillBonus;
    private              int                                 mFrightCheckBonus;
    private              int                                 mPerAdj;
    private              int                                 mPerceptionBonus;
    private              int                                 mVisionBonus;
    private              int                                 mHearingBonus;
    private              int                                 mTasteAndSmellBonus;
    private              int                                 mTouchBonus;
    private              int                                 mHitPointsDamage;
    private              int                                 mHitPointsAdj;
    private              int                                 mHitPointBonus;
    private              int                                 mFatiguePoints;
    private              int                                 mFatiguePointsDamage;
    private              int                                 mFatiguePointBonus;
    private              double                              mSpeedAdj;
    private              double                              mSpeedBonus;
    private              int                                 mMoveAdj;
    private              int                                 mMoveBonus;
    private              int                                 mDodgeBonus;
    private              int                                 mParryBonus;
    private              int                                 mBlockBonus;
    private              int                                 mTotalPoints;
    private              Settings                            mSettings;
    private              Profile                             mProfile;
    private              Armor                               mArmor;
    private              OutlineModel                        mAdvantages;
    private              OutlineModel                        mSkills;
    private              OutlineModel                        mSpells;
    private              OutlineModel                        mEquipment;
    private              OutlineModel                        mOtherEquipment;
    private              OutlineModel                        mNotes;
    private              WeightValue                         mCachedWeightCarried;
    private              Fixed6                              mCachedWealthCarried;
    private              Fixed6                              mCachedWealthNotCarried;
    private              int                                 mCachedAttributePoints;
    private              int                                 mCachedAdvantagePoints;
    private              int                                 mCachedDisadvantagePoints;
    private              int                                 mCachedQuirkPoints;
    private              int                                 mCachedSkillPoints;
    private              int                                 mCachedSpellPoints;
    private              int                                 mCachedRacePoints;
    private              PrintManager                        mPageSettings;
    private              String                              mPageSettingsString;
    private              boolean                             mSkillsUpdated;
    private              boolean                             mSpellsUpdated;
    private              boolean                             mDidModify;
    private              boolean                             mNeedAttributePointCalculation;
    private              boolean                             mNeedAdvantagesPointCalculation;
    private              boolean                             mNeedSkillPointCalculation;
    private              boolean                             mNeedSpellPointCalculation;
    private              boolean                             mNeedEquipmentCalculation;

    /** Creates a new character with only default values set. */
    public GURPSCharacter() {
        characterInitialize(true);
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
        load(path);
    }

    private void characterInitialize(boolean full) {
        mSettings = new Settings(this);
        mFeatureMap = new HashMap<>();
        mAdvantages = new OutlineModel();
        mSkills = new OutlineModel();
        mSpells = new OutlineModel();
        mEquipment = new OutlineModel();
        mOtherEquipment = new OutlineModel();
        mOtherEquipment.setProperty(EquipmentList.TAG_OTHER_ROOT, Boolean.TRUE);
        mNotes = new OutlineModel();
        mTotalPoints = Preferences.getInstance().getInitialPoints();
        mStrength = 10;
        mDexterity = 10;
        mIntelligence = 10;
        mHealth = 10;
        mHitPointsDamage = 0;
        mFatiguePointsDamage = 0;
        mProfile = new Profile(this, full);
        mArmor = new Armor(this);
        mCachedWeightCarried = new WeightValue(Fixed6.ZERO, mSettings.defaultWeightUnits());
        mPageSettings = Preferences.getInstance().getDefaultPageSettings();
        mPageSettingsString = "{}";
        if (mPageSettings != null) {
            mPageSettings = new PrintManager(mPageSettings);
            mPageSettingsString = mPageSettings.toString();
        }
        mModifiedOn = System.currentTimeMillis();
        mCreatedOn = mModifiedOn;
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
        return TAG_ROOT;
    }

    @Override
    public int getXMLTagVersion() {
        return CURRENT_VERSION;
    }

    @Override
    public String getXMLTagName() {
        return TAG_ROOT;
    }

    @Override
    protected final void loadSelf(XMLReader reader, LoadState state) throws IOException {
        String marker        = reader.getMarker();
        int    unspentPoints = 0;
        int    currentHP     = Integer.MIN_VALUE;
        int    currentFP     = Integer.MIN_VALUE;
        characterInitialize(false);
        long modifiedOn = mModifiedOn;
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();

                if (state.mDataFileVersion == 0) {
                    if (mProfile.loadTag(reader, name)) {
                        continue;
                    }
                }

                if (state.mDataFileVersion < HP_FP_DAMAGE_TRACKING) {
                    if ("current_hp".equals(name)) {
                        currentHP = reader.readInteger(Integer.MIN_VALUE);
                        continue;
                    } else if ("current_fp".equals(name)) {
                        currentFP = reader.readInteger(Integer.MIN_VALUE);
                        continue;
                    }
                }

                if (Settings.TAG_ROOT.equals(name)) {
                    mSettings.load(reader);
                } else if (Profile.TAG_ROOT.equals(name)) {
                    mProfile.load(reader);
                } else if (TAG_CREATED_DATE.equals(name)) {
                    mCreatedOn = Numbers.extractDateTime(reader.readText());
                } else if (TAG_MODIFIED_DATE.equals(name)) {
                    modifiedOn = Numbers.extractDateTime(reader.readText());
                } else if (BonusAttributeType.HP.getXMLTag().equals(name)) {
                    mHitPointsAdj = reader.readInteger(0);
                } else if (TAG_HP_DAMAGE.equals(name)) {
                    mHitPointsDamage = reader.readInteger(0);
                } else if (BonusAttributeType.FP.getXMLTag().equals(name)) {
                    mFatiguePoints = reader.readInteger(0);
                } else if (TAG_FP_DAMAGE.equals(name)) {
                    mFatiguePointsDamage = reader.readInteger(0);
                } else if (TAG_UNSPENT_POINTS.equals(name)) {
                    unspentPoints = reader.readInteger(0);
                } else if (TAG_TOTAL_POINTS.equals(name)) {
                    mTotalPoints = reader.readInteger(0);
                } else if (BonusAttributeType.ST.getXMLTag().equals(name)) {
                    mStrength = reader.readInteger(0);
                } else if (BonusAttributeType.DX.getXMLTag().equals(name)) {
                    mDexterity = reader.readInteger(0);
                } else if (BonusAttributeType.IQ.getXMLTag().equals(name)) {
                    mIntelligence = reader.readInteger(0);
                } else if (BonusAttributeType.HT.getXMLTag().equals(name)) {
                    mHealth = reader.readInteger(0);
                } else if (BonusAttributeType.WILL.getXMLTag().equals(name)) {
                    mWillAdj = reader.readInteger(0);
                } else if (BonusAttributeType.PERCEPTION.getXMLTag().equals(name)) {
                    mPerAdj = reader.readInteger(0);
                } else if (BonusAttributeType.SPEED.getXMLTag().equals(name)) {
                    mSpeedAdj = reader.readDouble(0.0);
                } else if (BonusAttributeType.MOVE.getXMLTag().equals(name)) {
                    mMoveAdj = reader.readInteger(0);
                } else if (AdvantageList.TAG_ROOT.equals(name)) {
                    loadAdvantageList(reader, state);
                } else if (SkillList.TAG_ROOT.equals(name)) {
                    loadSkillList(reader, state);
                } else if (SpellList.TAG_ROOT.equals(name)) {
                    loadSpellList(reader, state);
                } else if (EquipmentList.TAG_CARRIED_ROOT.equals(name)) {
                    loadEquipmentList(reader, state, mEquipment);
                } else if (EquipmentList.TAG_OTHER_ROOT.equals(name)) {
                    loadEquipmentList(reader, state, mOtherEquipment);
                } else if (NoteList.TAG_ROOT.equals(name)) {
                    loadNoteList(reader, state);
                } else if (mPageSettings != null && PrintManager.TAG_ROOT.equals(name)) {
                    mPageSettings.load(reader);
                    mPageSettingsString = mPageSettings.toString();
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));

        // Loop through the skills and update their levels. It is necessary to do this here and not
        // as they are loaded, since references to defaults won't work until the entire list is
        // available.
        for (Skill skill : getSkillsIterator()) {
            skill.updateLevel(false);
        }

        calculateAll();
        if (unspentPoints != 0) {
            setUnspentPoints(unspentPoints);
        }

        if (state.mDataFileVersion < HP_FP_DAMAGE_TRACKING) {
            if (currentHP != Integer.MIN_VALUE) {
                mHitPointsDamage = -Math.min(currentHP - getHitPointsAdj(), 0);
            }
            if (currentFP != Integer.MIN_VALUE) {
                mFatiguePointsDamage = -Math.min(currentFP - getFatiguePoints(), 0);
            }
        }
        mModifiedOn = modifiedOn;
    }

    private void loadAdvantageList(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Advantage.TAG_ADVANTAGE.equals(name) || Advantage.TAG_ADVANTAGE_CONTAINER.equals(name)) {
                    mAdvantages.addRow(new Advantage(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }

    private void loadSkillList(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Skill.TAG_SKILL.equals(name) || Skill.TAG_SKILL_CONTAINER.equals(name)) {
                    mSkills.addRow(new Skill(this, reader, state), true);
                } else if (Technique.TAG_TECHNIQUE.equals(name)) {
                    mSkills.addRow(new Technique(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }

    private void loadSpellList(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Spell.TAG_SPELL.equals(name) || Spell.TAG_SPELL_CONTAINER.equals(name)) {
                    mSpells.addRow(new Spell(this, reader, state), true);
                } else if (RitualMagicSpell.TAG_RITUAL_MAGIC_SPELL.equals(name)) {
                    mSpells.addRow(new RitualMagicSpell(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }

    private void loadEquipmentList(XMLReader reader, LoadState state, OutlineModel equipmentList) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Equipment.TAG_EQUIPMENT.equals(name) || Equipment.TAG_EQUIPMENT_CONTAINER.equals(name)) {
                    state.mUncarriedEquipment = new HashSet<>();
                    Equipment equipment = new Equipment(this, reader, state);
                    if (state.mDataFileVersion < SEPARATED_EQUIPMENT_VERSION && equipmentList == mEquipment && !state.mUncarriedEquipment.isEmpty()) {
                        if (addToEquipment(state.mUncarriedEquipment, equipment)) {
                            equipmentList.addRow(equipment, true);
                        }
                    } else {
                        equipmentList.addRow(equipment, true);
                    }
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }

    private boolean addToEquipment(HashSet<Equipment> uncarried, Equipment equipment) {
        if (uncarried.contains(equipment)) {
            mOtherEquipment.addRow(equipment, true);
            return false;
        }
        List<Row> children = equipment.getChildren();
        if (children != null) {
            for (Row child : new ArrayList<>(children)) {
                if (!addToEquipment(uncarried, (Equipment) child)) {
                    equipment.removeChild(child);
                }
            }
        }
        return true;
    }

    private void loadNoteList(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Note.TAG_NOTE.equals(name) || Note.TAG_NOTE_CONTAINER.equals(name)) {
                    mNotes.addRow(new Note(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
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
        mSettings.load(m.getMap(Settings.TAG_ROOT));
        mCreatedOn = Numbers.extractDateTime(m.getString(TAG_CREATED_DATE));
        mProfile.load(m.getMap(Profile.TAG_ROOT));
        mHitPointsAdj = m.getInt(KEY_HP_ADJ);
        mHitPointsDamage = m.getInt(TAG_HP_DAMAGE);
        mFatiguePoints = m.getInt(KEY_FP_ADJ);
        mFatiguePointsDamage = m.getInt(TAG_FP_DAMAGE);
        mTotalPoints = m.getInt(TAG_TOTAL_POINTS);
        mStrength = m.getInt(KEY_ST);
        mDexterity = m.getInt(KEY_DX);
        mIntelligence = m.getInt(KEY_IQ);
        mHealth = m.getInt(KEY_HT);
        mWillAdj = m.getInt(KEY_WILL_ADJ);
        mPerAdj = m.getInt(KEY_PER_ADJ);
        mSpeedAdj = m.getDouble(KEY_SPEED_ADJ);
        mMoveAdj = m.getInt(KEY_MOVE_ADJ);
        AdvantageList.loadIntoModel(this, m.getArray(KEY_ADVANTAGES), mAdvantages, state);
        SkillList.loadIntoModel(this, m.getArray(KEY_SKILLS), mSkills, state);
        SpellList.loadIntoModel(this, m.getArray(KEY_SPELLS), mSpells, state);
        EquipmentList.loadIntoModel(this, m.getArray(KEY_EQUIPMENT), mEquipment, state);
        EquipmentList.loadIntoModel(this, m.getArray(KEY_OTHER_EQUIPMENT), mOtherEquipment, state);
        NoteList.loadIntoModel(this, m.getArray(KEY_NOTES), mNotes, state);
        if (mPageSettings != null && m.has(PrintManager.TAG_ROOT)) {
            mPageSettings = new PrintManager(m.getMap(PrintManager.TAG_ROOT));
            mPageSettingsString = mPageSettings.toString();
        }
        // Loop through the skills and update their levels. It is necessary to do this here and not
        // as they are loaded, since references to defaults won't work until the entire list is
        // available.
        for (Skill skill : getSkillsIterator()) {
            skill.updateLevel(false);
        }
        calculateAll();
        mModifiedOn = Numbers.extractDateTime(m.getString(TAG_MODIFIED_DATE)); // Must be last
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        w.key(Settings.TAG_ROOT);
        mSettings.save(w);
        w.keyValue(TAG_CREATED_DATE, DateFormat.getDateTimeInstance(DateFormat.MEDIUM, DateFormat.SHORT).format(new Date(mCreatedOn)));
        w.keyValue(TAG_MODIFIED_DATE, DateFormat.getDateTimeInstance(DateFormat.MEDIUM, DateFormat.SHORT).format(new Date(mModifiedOn)));
        w.key(Profile.TAG_ROOT);
        mProfile.save(w);
        w.keyValueNot(KEY_HP_ADJ, mHitPointsAdj, 0);
        w.keyValueNot(TAG_HP_DAMAGE, mHitPointsDamage, 0);
        w.keyValueNot(KEY_FP_ADJ, mFatiguePoints, 0);
        w.keyValueNot(TAG_FP_DAMAGE, mFatiguePointsDamage, 0);
        w.keyValue(TAG_TOTAL_POINTS, mTotalPoints);
        w.keyValue(KEY_ST, mStrength);
        w.keyValue(KEY_DX, mDexterity);
        w.keyValue(KEY_IQ, mIntelligence);
        w.keyValue(KEY_HT, mHealth);
        w.keyValueNot(KEY_WILL_ADJ, mWillAdj, 0);
        w.keyValueNot(KEY_PER_ADJ, mPerAdj, 0);
        w.keyValueNot(KEY_SPEED_ADJ, mSpeedAdj, 0);
        w.keyValueNot(KEY_MOVE_ADJ, mMoveAdj, 0);
        ListRow.saveList(w, KEY_ADVANTAGES, mAdvantages.getTopLevelRows(), false);
        ListRow.saveList(w, KEY_SKILLS, mSkills.getTopLevelRows(), false);
        ListRow.saveList(w, KEY_SPELLS, mSpells.getTopLevelRows(), false);
        ListRow.saveList(w, KEY_EQUIPMENT, mEquipment.getTopLevelRows(), false);
        ListRow.saveList(w, KEY_OTHER_EQUIPMENT, mOtherEquipment.getTopLevelRows(), false);
        ListRow.saveList(w, KEY_NOTES, mNotes.getTopLevelRows(), false);
        if (mPageSettings != null) {
            w.key(PrintManager.TAG_ROOT);
            mPageSettings.save(w, LengthUnits.IN);
            mPageSettingsString = mPageSettings.toString();
        }
    }

    /**
     * @param id The field ID to retrieve the data for.
     * @return The value of the specified field ID, or {@code null} if the field ID is invalid.
     */
    public Object getValueForID(String id) {
        if (id == null) {
            return null;
        }
        if (Settings.PREFIX.equals(id)) { // Special to retrieve options code
            return mSettings.optionsCode();
        } else if (id.startsWith(POINTS_PREFIX)) {
            id = id.substring(POINTS_PREFIX.length());
            if (ID_STRENGTH.equals(id)) {
                return Integer.valueOf(getStrengthPoints());
            } else if (ID_DEXTERITY.equals(id)) {
                return Integer.valueOf(getDexterityPoints());
            } else if (ID_INTELLIGENCE.equals(id)) {
                return Integer.valueOf(getIntelligencePoints());
            } else if (ID_HEALTH.equals(id)) {
                return Integer.valueOf(getHealthPoints());
            } else if (ID_WILL.equals(id)) {
                return Integer.valueOf(getWillPoints());
            } else if (ID_PERCEPTION.equals(id)) {
                return Integer.valueOf(getPerceptionPoints());
            } else if (ID_BASIC_SPEED.equals(id)) {
                return Integer.valueOf(getBasicSpeedPoints());
            } else if (ID_BASIC_MOVE.equals(id)) {
                return Integer.valueOf(getBasicMovePoints());
            } else if (ID_FATIGUE_POINTS.equals(id)) {
                return Integer.valueOf(getFatiguePointPoints());
            } else if (ID_HIT_POINTS.equals(id)) {
                return Integer.valueOf(getHitPointPoints());
            }
            return null;
        } else if (ID_MODIFIED.equals(id)) {
            return Long.valueOf(getModifiedOn());
        } else if (ID_CREATED.equals(id)) {
            return Long.valueOf(getCreatedOn());
        } else if (ID_STRENGTH.equals(id)) {
            return Integer.valueOf(getStrength());
        } else if (ID_DEXTERITY.equals(id)) {
            return Integer.valueOf(getDexterity());
        } else if (ID_INTELLIGENCE.equals(id)) {
            return Integer.valueOf(getIntelligence());
        } else if (ID_HEALTH.equals(id)) {
            return Integer.valueOf(getHealth());
        } else if (ID_BASIC_SPEED.equals(id)) {
            return Double.valueOf(getBasicSpeed());
        } else if (ID_BASIC_MOVE.equals(id)) {
            return Integer.valueOf(getBasicMove());
        } else if (ID_BASIC_LIFT.equals(id)) {
            return getBasicLift();
        } else if (ID_PERCEPTION.equals(id)) {
            return Integer.valueOf(getPerAdj());
        } else if (ID_VISION.equals(id)) {
            return Integer.valueOf(getVision());
        } else if (ID_HEARING.equals(id)) {
            return Integer.valueOf(getHearing());
        } else if (ID_TASTE_AND_SMELL.equals(id)) {
            return Integer.valueOf(getTasteAndSmell());
        } else if (ID_TOUCH.equals(id)) {
            return Integer.valueOf(getTouch());
        } else if (ID_WILL.equals(id)) {
            return Integer.valueOf(getWillAdj());
        } else if (ID_FRIGHT_CHECK.equals(id)) {
            return Integer.valueOf(getFrightCheck());
        } else if (ID_ATTRIBUTE_POINTS.equals(id)) {
            return Integer.valueOf(getAttributePoints());
        } else if (ID_ADVANTAGE_POINTS.equals(id)) {
            return Integer.valueOf(getAdvantagePoints());
        } else if (ID_DISADVANTAGE_POINTS.equals(id)) {
            return Integer.valueOf(getDisadvantagePoints());
        } else if (ID_QUIRK_POINTS.equals(id)) {
            return Integer.valueOf(getQuirkPoints());
        } else if (ID_SKILL_POINTS.equals(id)) {
            return Integer.valueOf(getSkillPoints());
        } else if (ID_SPELL_POINTS.equals(id)) {
            return Integer.valueOf(getSpellPoints());
        } else if (ID_RACE_POINTS.equals(id)) {
            return Integer.valueOf(getRacePoints());
        } else if (ID_UNSPENT_POINTS.equals(id)) {
            return Integer.valueOf(getUnspentPoints());
        } else if (ID_ONE_HANDED_LIFT.equals(id)) {
            return getOneHandedLift();
        } else if (ID_TWO_HANDED_LIFT.equals(id)) {
            return getTwoHandedLift();
        } else if (ID_SHOVE_AND_KNOCK_OVER.equals(id)) {
            return getShoveAndKnockOver();
        } else if (ID_RUNNING_SHOVE_AND_KNOCK_OVER.equals(id)) {
            return getRunningShoveAndKnockOver();
        } else if (ID_CARRY_ON_BACK.equals(id)) {
            return getCarryOnBack();
        } else if (ID_SHIFT_SLIGHTLY.equals(id)) {
            return getShiftSlightly();
        } else if (ID_TOTAL_POINTS.equals(id)) {
            return Integer.valueOf(getTotalPoints());
        } else if (ID_BASIC_THRUST.equals(id)) {
            return getThrust();
        } else if (ID_BASIC_SWING.equals(id)) {
            return getSwing();
        } else if (ID_HIT_POINTS.equals(id)) {
            return Integer.valueOf(getHitPointsAdj());
        } else if (ID_HIT_POINTS_DAMAGE.equals(id)) {
            return Integer.valueOf(getHitPointsDamage());
        } else if (ID_CURRENT_HP.equals(id)) {
            return Integer.valueOf(getCurrentHitPoints());
        } else if (ID_REELING_HIT_POINTS.equals(id)) {
            return Integer.valueOf(getReelingHitPoints());
        } else if (ID_UNCONSCIOUS_CHECKS_HIT_POINTS.equals(id)) {
            return Integer.valueOf(getUnconsciousChecksHitPoints());
        } else if (ID_DEATH_CHECK_1_HIT_POINTS.equals(id)) {
            return Integer.valueOf(getDeathCheck1HitPoints());
        } else if (ID_DEATH_CHECK_2_HIT_POINTS.equals(id)) {
            return Integer.valueOf(getDeathCheck2HitPoints());
        } else if (ID_DEATH_CHECK_3_HIT_POINTS.equals(id)) {
            return Integer.valueOf(getDeathCheck3HitPoints());
        } else if (ID_DEATH_CHECK_4_HIT_POINTS.equals(id)) {
            return Integer.valueOf(getDeathCheck4HitPoints());
        } else if (ID_DEAD_HIT_POINTS.equals(id)) {
            return Integer.valueOf(getDeadHitPoints());
        } else if (ID_FATIGUE_POINTS.equals(id)) {
            return Integer.valueOf(getFatiguePoints());
        } else if (ID_FATIGUE_POINTS_DAMAGE.equals(id)) {
            return Integer.valueOf(getFatiguePointsDamage());
        } else if (ID_CURRENT_FP.equals(id)) {
            return Integer.valueOf(getCurrentFatiguePoints());
        } else if (ID_TIRED_FATIGUE_POINTS.equals(id)) {
            return Integer.valueOf(getTiredFatiguePoints());
        } else if (ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS.equals(id)) {
            return Integer.valueOf(getUnconsciousChecksFatiguePoints());
        } else if (ID_UNCONSCIOUS_FATIGUE_POINTS.equals(id)) {
            return Integer.valueOf(getUnconsciousFatiguePoints());
        } else if (ID_PARRY_BONUS.equals(id)) {
            return Integer.valueOf(getParryBonus());
        } else if (ID_BLOCK_BONUS.equals(id)) {
            return Integer.valueOf(getBlockBonus());
        } else if (ID_DODGE_BONUS.equals(id)) {
            return Integer.valueOf(getDodgeBonus());
        } else if (id.startsWith(Profile.PROFILE_PREFIX)) {
            return mProfile.getValueForID(id);
        } else if (id.startsWith(Armor.DR_PREFIX)) {
            return mArmor.getValueForID(id);
        } else {
            for (Encumbrance encumbrance : Encumbrance.values()) {
                int index = encumbrance.ordinal();
                if ((DODGE_PREFIX + index).equals(id)) {
                    return Integer.valueOf(getDodge(encumbrance));
                }
                if ((MOVE_PREFIX + index).equals(id)) {
                    return Integer.valueOf(getMove(encumbrance));
                }
                if ((MAXIMUM_CARRY_PREFIX + index).equals(id)) {
                    return getMaximumCarry(encumbrance);
                }
            }
            return null;
        }
    }

    /**
     * @param id    The field ID to set the value for.
     * @param value The value to set.
     */
    public void setValueForID(String id, Object value) {
        if (id != null) {
            if (ID_STRENGTH.equals(id)) {
                setStrength(((Integer) value).intValue());
            } else if (ID_DEXTERITY.equals(id)) {
                setDexterity(((Integer) value).intValue());
            } else if (ID_INTELLIGENCE.equals(id)) {
                setIntelligence(((Integer) value).intValue());
            } else if (ID_HEALTH.equals(id)) {
                setHealth(((Integer) value).intValue());
            } else if (ID_BASIC_SPEED.equals(id)) {
                setBasicSpeed(((Double) value).doubleValue());
            } else if (ID_BASIC_MOVE.equals(id)) {
                setBasicMove(((Integer) value).intValue());
            } else if (ID_PERCEPTION.equals(id)) {
                setPerAdj(((Integer) value).intValue());
            } else if (ID_WILL.equals(id)) {
                setWillAdj(((Integer) value).intValue());
            } else if (ID_UNSPENT_POINTS.equals(id)) {
                setUnspentPoints(((Integer) value).intValue());
            } else if (ID_HIT_POINTS.equals(id)) {
                setHitPointsAdj(((Integer) value).intValue());
            } else if (ID_HIT_POINTS_DAMAGE.equals(id)) {
                setHitPointsDamage(((Integer) value).intValue());
            } else if (ID_CURRENT_HP.equals(id)) {
                setHitPointsDamage(-Math.min(((Integer) value).intValue() - getHitPointsAdj(), 0));
            } else if (ID_FATIGUE_POINTS.equals(id)) {
                setFatiguePoints(((Integer) value).intValue());
            } else if (ID_FATIGUE_POINTS_DAMAGE.equals(id)) {
                setFatiguePointsDamage(((Integer) value).intValue());
            } else if (ID_CURRENT_FP.equals(id)) {
                setFatiguePointsDamage(-Math.min(((Integer) value).intValue() - getFatiguePoints(), 0));
            } else if (id.startsWith(Profile.PROFILE_PREFIX)) {
                mProfile.setValueForID(id, value);
            } else if (id.startsWith(Armor.DR_PREFIX)) {
                mArmor.setValueForID(id, value);
            } else {
                Log.error(String.format(I18n.Text("Unable to set a value for %s"), id));
            }
        }
    }

    @Override
    protected void startNotifyAtBatchLevelZero() {
        mDidModify = false;
        mNeedAttributePointCalculation = false;
        mNeedAdvantagesPointCalculation = false;
        mNeedSkillPointCalculation = false;
        mNeedSpellPointCalculation = false;
        mNeedEquipmentCalculation = false;
    }

    @Override
    public void notify(String type, Object data) {
        super.notify(type, data);
        if (Advantage.ID_POINTS.equals(type) || Advantage.ID_ROUND_COST_DOWN.equals(type) || Advantage.ID_LEVELS.equals(type) || Advantage.ID_CONTAINER_TYPE.equals(type) || Advantage.ID_LIST_CHANGED.equals(type) || Advantage.ID_CR.equals(type) || AdvantageModifier.ID_LIST_CHANGED.equals(type) || AdvantageModifier.ID_ENABLED.equals(type)) {
            mNeedAdvantagesPointCalculation = true;
        }
        if (Skill.ID_POINTS.equals(type) || Skill.ID_LIST_CHANGED.equals(type)) {
            mNeedSkillPointCalculation = true;
        }
        if (Spell.ID_POINTS.equals(type) || Spell.ID_LIST_CHANGED.equals(type)) {
            mNeedSpellPointCalculation = true;
        }
        if (Equipment.ID_QUANTITY.equals(type) || Equipment.ID_WEIGHT.equals(type) || Equipment.ID_EXTENDED_WEIGHT.equals(type) || Equipment.ID_LIST_CHANGED.equals(type) || EquipmentModifier.ID_WEIGHT_ADJ.equals(type) || EquipmentModifier.ID_COST_ADJ.equals(type) || EquipmentModifier.ID_ENABLED.equals(type)) {
            mNeedEquipmentCalculation = true;
        }
        if (Profile.ID_SIZE_MODIFIER.equals(type) || Settings.ID_USE_KNOW_YOUR_OWN_STRENGTH.equals(type)) {
            mNeedAttributePointCalculation = true;
        }
    }

    @Override
    protected void notifyOccured() {
        mDidModify = true;
    }

    @Override
    protected void endNotifyAtBatchLevelOne() {
        if (mNeedAttributePointCalculation) {
            calculateAttributePoints();
            notify(ID_ATTRIBUTE_POINTS, Integer.valueOf(getAttributePoints()));
        }
        if (mNeedAdvantagesPointCalculation) {
            calculateAdvantagePoints();
            notify(ID_ADVANTAGE_POINTS, Integer.valueOf(getAdvantagePoints()));
            notify(ID_DISADVANTAGE_POINTS, Integer.valueOf(getDisadvantagePoints()));
            notify(ID_QUIRK_POINTS, Integer.valueOf(getQuirkPoints()));
            notify(ID_RACE_POINTS, Integer.valueOf(getRacePoints()));
        }
        if (mNeedSkillPointCalculation) {
            calculateSkillPoints();
            notify(ID_SKILL_POINTS, Integer.valueOf(getSkillPoints()));
        }
        if (mNeedSpellPointCalculation) {
            calculateSpellPoints();
            notify(ID_SPELL_POINTS, Integer.valueOf(getSpellPoints()));
        }
        if (mNeedAttributePointCalculation || mNeedAdvantagesPointCalculation || mNeedSkillPointCalculation || mNeedSpellPointCalculation) {
            notify(ID_UNSPENT_POINTS, Integer.valueOf(getUnspentPoints()));
        }
        if (mNeedEquipmentCalculation) {
            calculateWeightAndWealthCarried(true);
            calculateWealthNotCarried(true);
        }
        if (mDidModify) {
            setModifiedOn(System.currentTimeMillis());
        }
    }

    /** @return The created on date. */
    public long getCreatedOn() {
        return mCreatedOn;
    }

    /** @return The modified date. */
    public long getModifiedOn() {
        return mModifiedOn;
    }

    public void setModifiedOn(long when) {
        if (mModifiedOn != when) {
            mModifiedOn = when;
            notify(ID_MODIFIED, Long.valueOf(mModifiedOn));
        }
    }

    private void updateSkills() {
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

    /** @return The strength (ST). */
    public int getStrength() {
        return mStrength + mStrengthBonus;
    }

    /**
     * Sets the strength (ST).
     *
     * @param strength The new strength.
     */
    public void setStrength(int strength) {
        int oldStrength = getStrength();
        if (oldStrength != strength) {
            postUndoEdit(I18n.Text("Strength Change"), ID_STRENGTH, Integer.valueOf(oldStrength), Integer.valueOf(strength));
            updateStrengthInfo(strength - mStrengthBonus, mStrengthBonus, mLiftingStrengthBonus, mStrikingStrengthBonus);
        }
    }

    /** @return The current strength bonus from features. */
    public int getStrengthBonus() {
        return mStrengthBonus;
    }

    /** @param bonus The new strength bonus. */
    public void setStrengthBonus(int bonus) {
        if (mStrengthBonus != bonus) {
            updateStrengthInfo(mStrength, bonus, mLiftingStrengthBonus, mStrikingStrengthBonus);
        }
    }

    /** @param reduction The cost reduction for strength. */
    public void setStrengthCostReduction(int reduction) {
        if (mStrengthCostReduction != reduction) {
            mStrengthCostReduction = reduction;
            mNeedAttributePointCalculation = true;
        }
    }

    /** @return The current lifting strength bonus from features. */
    public int getLiftingStrengthBonus() {
        return mLiftingStrengthBonus;
    }

    /** @param bonus The new lifting strength bonus. */
    public void setLiftingStrengthBonus(int bonus) {
        if (mLiftingStrengthBonus != bonus) {
            updateStrengthInfo(mStrength, mStrengthBonus, bonus, mStrikingStrengthBonus);
        }
    }

    /** @return The current striking strength bonus from features. */
    public int getStrikingStrengthBonus() {
        return mStrikingStrengthBonus;
    }

    /** @param bonus The new striking strength bonus. */
    public void setStrikingStrengthBonus(int bonus) {
        if (mStrikingStrengthBonus != bonus) {
            updateStrengthInfo(mStrength, mStrengthBonus, mLiftingStrengthBonus, bonus);
        }
    }

    private void updateStrengthInfo(int strength, int bonus, int liftingBonus, int strikingBonus) {
        Dice        thrust   = getThrust();
        Dice        swing    = getSwing();
        WeightValue lift     = getBasicLift();
        boolean     notifyST = mStrength != strength || mStrengthBonus != bonus;
        Dice        dice;

        mStrength = strength;
        mStrengthBonus = bonus;
        mLiftingStrengthBonus = liftingBonus;
        mStrikingStrengthBonus = strikingBonus;

        startNotify();
        if (notifyST) {
            notify(ID_STRENGTH, Integer.valueOf(getStrength()));
            notifyOfBaseHitPointChange();
        }
        WeightValue newLift = getBasicLift();
        if (!newLift.equals(lift)) {
            notifyBasicLift();
        }

        dice = getThrust();
        if (!dice.equals(thrust)) {
            notify(ID_BASIC_THRUST, dice);
        }
        dice = getSwing();
        if (!dice.equals(swing)) {
            notify(ID_BASIC_SWING, dice);
        }

        updateSkills();
        mNeedAttributePointCalculation = true;
        endNotify();
    }

    public void notifyBasicLift() {
        notify(ID_BASIC_LIFT, getBasicLift());
        notify(ID_ONE_HANDED_LIFT, getOneHandedLift());
        notify(ID_TWO_HANDED_LIFT, getTwoHandedLift());
        notify(ID_SHOVE_AND_KNOCK_OVER, getShoveAndKnockOver());
        notify(ID_RUNNING_SHOVE_AND_KNOCK_OVER, getRunningShoveAndKnockOver());
        notify(ID_CARRY_ON_BACK, getCarryOnBack());
        notify(ID_SHIFT_SLIGHTLY, getShiftSlightly());
        for (Encumbrance encumbrance : Encumbrance.values()) {
            notify(MAXIMUM_CARRY_PREFIX + encumbrance.ordinal(), getMaximumCarry(encumbrance));
        }
    }

    /** @return The number of points spent on strength. */
    public int getStrengthPoints() {
        int reduction = mStrengthCostReduction;
        if (!mSettings.useKnowYourOwnStrength()) {
            reduction += mProfile.getSizeModifier() * 10;
        }
        return getPointsForAttribute(mStrength - 10, 10, reduction);
    }

    private static int getPointsForAttribute(int delta, int ptsPerLevel, int reduction) {
        int amt = delta * ptsPerLevel;
        if (reduction > 0 && delta > 0) {
            if (reduction > 80) {
                reduction = 80;
            }
            amt = (99 + amt * (100 - reduction)) / 100;
        }
        return amt;
    }

    /** @return The basic thrusting damage. */
    public Dice getThrust() {
        return getThrust(getStrength() + mStrikingStrengthBonus);
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
        return getSwing(getStrength() + mStrikingStrengthBonus);
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
        int strength = getStrength() + mLiftingStrengthBonus;
        if (isTired()) {
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
     * @return The character's basic speed.
     */
    public double getBasicSpeed() {
        return mSpeedAdj + mSpeedBonus + getRawBasicSpeed();
    }

    private double getRawBasicSpeed() {
        return (getDexterity() + getHealth()) / 4.0;
    }

    /**
     * Sets the basic speed.
     *
     * @param speed The new basic speed.
     */
    public void setBasicSpeed(double speed) {
        double oldBasicSpeed = getBasicSpeed();
        if (oldBasicSpeed != speed) {
            postUndoEdit(I18n.Text("Basic Speed Change"), ID_BASIC_SPEED, Double.valueOf(oldBasicSpeed), Double.valueOf(speed));
            updateBasicSpeedInfo(speed - (mSpeedBonus + getRawBasicSpeed()), mSpeedBonus);
        }
    }

    /** @return The basic speed bonus. */
    public double getBasicSpeedBonus() {
        return mSpeedBonus;
    }

    /** @param bonus The basic speed bonus. */
    public void setBasicSpeedBonus(double bonus) {
        if (mSpeedBonus != bonus) {
            updateBasicSpeedInfo(mSpeedAdj, bonus);
        }
    }

    private void updateBasicSpeedInfo(double speed, double bonus) {
        int   move = getBasicMove();
        int[] data = preserveMoveAndDodge();
        int   tmp;

        mSpeedAdj = speed;
        mSpeedBonus = bonus;

        startNotify();
        notify(ID_BASIC_SPEED, Double.valueOf(getBasicSpeed()));
        tmp = getBasicMove();
        if (move != tmp) {
            notify(ID_BASIC_MOVE, Integer.valueOf(tmp));
        }
        notifyIfMoveOrDodgeAltered(data);
        mNeedAttributePointCalculation = true;
        endNotify();
    }

    /** @return The number of points spent on basic speed. */
    public int getBasicSpeedPoints() {
        return (int) (mSpeedAdj * 20.0);
    }

    /**
     * @return The character's basic move.
     */
    public int getBasicMove() {
        return Math.max(mMoveAdj + mMoveBonus + getRawBasicMove(), 0);
    }

    private int getRawBasicMove() {
        return (int) Math.floor(getBasicSpeed());
    }

    /**
     * Sets the basic move.
     *
     * @param move The new basic move.
     */
    public void setBasicMove(int move) {
        int oldBasicMove = getBasicMove();

        if (oldBasicMove != move) {
            postUndoEdit(I18n.Text("Basic Move Change"), ID_BASIC_MOVE, Integer.valueOf(oldBasicMove), Integer.valueOf(move));
            updateBasicMoveInfo(move - (mMoveBonus + getRawBasicMove()), mMoveBonus);
        }
    }

    /** @return The basic move bonus. */
    public int getBasicMoveBonus() {
        return mMoveBonus;
    }

    /** @param bonus The basic move bonus. */
    public void setBasicMoveBonus(int bonus) {
        if (mMoveBonus != bonus) {
            updateBasicMoveInfo(mMoveAdj, bonus);
        }
    }

    private void updateBasicMoveInfo(int move, int bonus) {
        int[] data = preserveMoveAndDodge();

        startNotify();
        mMoveAdj = move;
        mMoveBonus = bonus;
        notify(ID_BASIC_MOVE, Integer.valueOf(getBasicMove()));
        notifyIfMoveOrDodgeAltered(data);
        mNeedAttributePointCalculation = true;
        endNotify();
    }

    /** @return The number of points spent on basic move. */
    public int getBasicMovePoints() {
        return mMoveAdj * 5;
    }

    /**
     * @param encumbrance The encumbrance level.
     * @return The character's ground move for the specified encumbrance level.
     */
    public int getMove(Encumbrance encumbrance) {
        int     initialMove = getBasicMove();
        boolean reeling     = isReeling();
        boolean tired       = isTired();
        if (reeling || tired) {
            int     divisor = (reeling && tired) ? 4 : 2;
            boolean plusOne = initialMove % divisor != 0;
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
        int     dodge   = 3 + mDodgeBonus + (int) Math.floor(getBasicSpeed());
        boolean reeling = isReeling();
        boolean tired   = isTired();
        if (reeling || tired) {
            int     divisor = (reeling && tired) ? 4 : 2;
            boolean plusOne = dodge % divisor != 0;
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
            int[] data = preserveMoveAndDodge();

            mDodgeBonus = bonus;
            startNotify();
            notifySingle(ID_DODGE_BONUS, Integer.valueOf(mDodgeBonus));
            notifyIfMoveOrDodgeAltered(data);
            endNotify();
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
            notifySingle(ID_PARRY_BONUS, Integer.valueOf(mParryBonus));
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
            notifySingle(ID_BLOCK_BONUS, Integer.valueOf(mBlockBonus));
        }
    }

    /** @return The current encumbrance level. */
    public Encumbrance getEncumbranceLevel() {
        Fixed6 carried = getWeightCarried().getNormalizedValue();
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
    public boolean isCarryingGreaterThanMaxLoad() {
        return getWeightCarried().getNormalizedValue().greaterThan(getMaximumCarry(Encumbrance.EXTRA_HEAVY).getNormalizedValue());
    }

    /** @return The current weight being carried. */
    public WeightValue getWeightCarried() {
        return mCachedWeightCarried;
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
        switch (value.getUnits()) {
        case G:
            return new WeightValue(value.getValue().div(new Fixed6(30)), WeightUnits.OZ);
        case KG:
            return new WeightValue(value.getValue().mul(new Fixed6(2)), WeightUnits.LB);
        case T:
            return new WeightValue(value.getValue(), WeightUnits.LT);
        default:
            return value;
        }
    }

    /**
     * Convert an imperial {@link WeightValue} by GURPS Metric rules into a metric one. If a metric
     * {@link WeightValue} is passed as an argument, it will be returned unchanged.
     *
     * @param value The {@link WeightValue} to be converted by GURPS Metric rules.
     * @return The converted metric {@link WeightValue}.
     */
    public static WeightValue convertToGurpsMetric(WeightValue value) {
        switch (value.getUnits()) {
        case LB:
            return new WeightValue(value.getValue().div(new Fixed6(2)), WeightUnits.KG);
        case LT:
        case TN:
            return new WeightValue(value.getValue(), WeightUnits.T);
        case OZ:
            return new WeightValue(value.getValue().mul(new Fixed6(30)), WeightUnits.G);
        default:
            return value;
        }
    }

    /**
     * Calculate the total weight and wealth carried.
     *
     * @param notify Whether to send out notifications if the resulting values are different from
     *               the previous values.
     */
    public void calculateWeightAndWealthCarried(boolean notify) {
        WeightValue savedWeight = new WeightValue(mCachedWeightCarried);
        Fixed6      savedWealth = mCachedWealthCarried;
        mCachedWeightCarried = new WeightValue(Fixed6.ZERO, defaultWeightUnits());
        mCachedWealthCarried = Fixed6.ZERO;
        for (Row one : mEquipment.getTopLevelRows()) {
            Equipment   equipment = (Equipment) one;
            WeightValue weight    = new WeightValue(equipment.getExtendedWeight());
            if (useSimpleMetricConversions()) {
                weight = defaultWeightUnits().isMetric() ? convertToGurpsMetric(weight) : convertFromGurpsMetric(weight);
            }
            mCachedWeightCarried.add(weight);
            mCachedWealthCarried = mCachedWealthCarried.add(equipment.getExtendedValue());
        }
        if (notify) {
            if (!savedWeight.equals(mCachedWeightCarried)) {
                notify(ID_CARRIED_WEIGHT, mCachedWeightCarried);
            }
            if (!mCachedWealthCarried.equals(savedWealth)) {
                notify(ID_CARRIED_WEALTH, mCachedWealthCarried);
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
        for (Row one : mOtherEquipment.getTopLevelRows()) {
            mCachedWealthNotCarried = mCachedWealthNotCarried.add(((Equipment) one).getExtendedValue());
        }
        if (notify) {
            if (!mCachedWealthNotCarried.equals(savedWealth)) {
                notify(ID_NOT_CARRIED_WEALTH, mCachedWealthNotCarried);
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

    private void notifyIfMoveOrDodgeAltered(int[] data) {
        Encumbrance[] values = Encumbrance.values();
        for (Encumbrance encumbrance : values) {
            int index = encumbrance.ordinal();
            int tmp   = getDodge(encumbrance);
            if (tmp != data[values.length + index]) {
                notify(DODGE_PREFIX + index, Integer.valueOf(tmp));
            }
            tmp = getMove(encumbrance);
            if (tmp != data[index]) {
                notify(MOVE_PREFIX + index, Integer.valueOf(tmp));
            }
        }
    }

    public void notifyMoveAndDodge() {
        for (Encumbrance encumbrance : Encumbrance.values()) {
            int index = encumbrance.ordinal();
            notify(DODGE_PREFIX + index, Integer.valueOf(getDodge(encumbrance)));
            notify(MOVE_PREFIX + index, Integer.valueOf(getMove(encumbrance)));
        }
    }

    /** @return The dexterity (DX). */
    public int getDexterity() {
        return mDexterity + mDexterityBonus;
    }

    /**
     * Sets the dexterity (DX).
     *
     * @param dexterity The new dexterity.
     */
    public void setDexterity(int dexterity) {
        int oldDexterity = getDexterity();

        if (oldDexterity != dexterity) {
            postUndoEdit(I18n.Text("Dexterity Change"), ID_DEXTERITY, Integer.valueOf(oldDexterity), Integer.valueOf(dexterity));
            updateDexterityInfo(dexterity - mDexterityBonus, mDexterityBonus);
        }
    }

    /** @return The dexterity bonus. */
    public int getDexterityBonus() {
        return mDexterityBonus;
    }

    /** @param bonus The new dexterity bonus. */
    public void setDexterityBonus(int bonus) {
        if (mDexterityBonus != bonus) {
            updateDexterityInfo(mDexterity, bonus);
        }
    }

    /** @param reduction The cost reduction for dexterity. */
    public void setDexterityCostReduction(int reduction) {
        if (mDexterityCostReduction != reduction) {
            mDexterityCostReduction = reduction;
            mNeedAttributePointCalculation = true;
        }
    }

    private void updateDexterityInfo(int dexterity, int bonus) {
        double speed = getBasicSpeed();
        int    move  = getBasicMove();
        int[]  data  = preserveMoveAndDodge();
        double newSpeed;
        int    newMove;

        mDexterity = dexterity;
        mDexterityBonus = bonus;

        startNotify();
        notify(ID_DEXTERITY, Integer.valueOf(getDexterity()));
        newSpeed = getBasicSpeed();
        if (newSpeed != speed) {
            notify(ID_BASIC_SPEED, Double.valueOf(newSpeed));
        }
        newMove = getBasicMove();
        if (newMove != move) {
            notify(ID_BASIC_MOVE, Integer.valueOf(newMove));
        }
        notifyIfMoveOrDodgeAltered(data);
        updateSkills();
        mNeedAttributePointCalculation = true;
        endNotify();
    }

    /** @return The number of points spent on dexterity. */
    public int getDexterityPoints() {
        return getPointsForAttribute(mDexterity - 10, 20, mDexterityCostReduction);
    }

    /** @return The intelligence (IQ). */
    public int getIntelligence() {
        return mIntelligence + mIntelligenceBonus;
    }

    /**
     * Sets the intelligence (IQ).
     *
     * @param intelligence The new intelligence.
     */
    public void setIntelligence(int intelligence) {
        int oldIntelligence = getIntelligence();
        if (oldIntelligence != intelligence) {
            postUndoEdit(I18n.Text("Intelligence Change"), ID_INTELLIGENCE, Integer.valueOf(oldIntelligence), Integer.valueOf(intelligence));
            updateIntelligenceInfo(intelligence - mIntelligenceBonus, mIntelligenceBonus);
        }
    }

    /** @return The intelligence bonus. */
    public int getIntelligenceBonus() {
        return mIntelligenceBonus;
    }

    /** @param bonus The new intelligence bonus. */
    public void setIntelligenceBonus(int bonus) {
        if (mIntelligenceBonus != bonus) {
            updateIntelligenceInfo(mIntelligence, bonus);
        }
    }

    /** @param reduction The cost reduction for intelligence. */
    public void setIntelligenceCostReduction(int reduction) {
        if (mIntelligenceCostReduction != reduction) {
            mIntelligenceCostReduction = reduction;
            mNeedAttributePointCalculation = true;
        }
    }

    private void updateIntelligenceInfo(int intelligence, int bonus) {
        int perception = getPerAdj();
        int will       = getWillAdj();
        int newPerception;
        int newWill;

        mIntelligence = intelligence;
        mIntelligenceBonus = bonus;

        startNotify();
        notify(ID_INTELLIGENCE, Integer.valueOf(getIntelligence()));
        newPerception = getPerAdj();
        if (newPerception != perception) {
            notify(ID_PERCEPTION, Integer.valueOf(newPerception));
            notify(ID_VISION, Integer.valueOf(getVision()));
            notify(ID_HEARING, Integer.valueOf(getHearing()));
            notify(ID_TASTE_AND_SMELL, Integer.valueOf(getTasteAndSmell()));
            notify(ID_TOUCH, Integer.valueOf(getTouch()));
        }
        newWill = getWillAdj();
        if (newWill != will) {
            notify(ID_WILL, Integer.valueOf(newWill));
            notify(ID_FRIGHT_CHECK, Integer.valueOf(getFrightCheck()));
        }
        updateSkills();
        updateSpells();
        mNeedAttributePointCalculation = true;
        endNotify();
    }

    /** @return The number of points spent on intelligence. */
    public int getIntelligencePoints() {
        return getPointsForAttribute(mIntelligence - 10, 20, mIntelligenceCostReduction);
    }

    /** @return The health (HT). */
    public int getHealth() {
        return mHealth + mHealthBonus;
    }

    /**
     * Sets the health (HT).
     *
     * @param health The new health.
     */
    public void setHealth(int health) {
        int oldHealth = getHealth();

        if (oldHealth != health) {
            postUndoEdit(I18n.Text("Health Change"), ID_HEALTH, Integer.valueOf(oldHealth), Integer.valueOf(health));
            updateHealthInfo(health - mHealthBonus, mHealthBonus);
        }
    }

    /** @return The health bonus. */
    public int getHealthBonus() {
        return mHealthBonus;
    }

    /** @param bonus The new health bonus. */
    public void setHealthBonus(int bonus) {
        if (mHealthBonus != bonus) {
            updateHealthInfo(mHealth, bonus);
        }
    }

    /** @param reduction The cost reduction for health. */
    public void setHealthCostReduction(int reduction) {
        if (mHealthCostReduction != reduction) {
            mHealthCostReduction = reduction;
            mNeedAttributePointCalculation = true;
        }
    }

    private void updateHealthInfo(int health, int bonus) {
        double speed = getBasicSpeed();
        int    move  = getBasicMove();
        int[]  data  = preserveMoveAndDodge();
        double newSpeed;
        int    tmp;

        mHealth = health;
        mHealthBonus = bonus;

        startNotify();
        notify(ID_HEALTH, Integer.valueOf(getHealth()));

        newSpeed = getBasicSpeed();
        if (newSpeed != speed) {
            notify(ID_BASIC_SPEED, Double.valueOf(newSpeed));
        }

        tmp = getBasicMove();
        if (tmp != move) {
            notify(ID_BASIC_MOVE, Integer.valueOf(tmp));
        }
        notifyIfMoveOrDodgeAltered(data);
        notifyOfBaseFatiguePointChange();
        updateSkills();
        mNeedAttributePointCalculation = true;
        endNotify();
    }

    /** @return The number of points spent on health. */
    public int getHealthPoints() {
        return getPointsForAttribute(mHealth - 10, 10, mHealthCostReduction);
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
            Integer value = Integer.valueOf(unspent);

            postUndoEdit(I18n.Text("Unspent Points Change"), ID_UNSPENT_POINTS, Integer.valueOf(current), value);
            mTotalPoints = unspent + getSpentPoints();
            startNotify();
            notify(ID_UNSPENT_POINTS, value);
            notify(ID_TOTAL_POINTS, Integer.valueOf(getTotalPoints()));
            endNotify();
        }
    }

    /** @return The number of points spent on basic attributes. */
    public int getAttributePoints() {
        return mCachedAttributePoints;
    }

    private void calculateAttributePoints() {
        mCachedAttributePoints = getStrengthPoints() + getDexterityPoints() + getIntelligencePoints() + getHealthPoints() + getWillPoints() + getPerceptionPoints() + getBasicSpeedPoints() + getBasicMovePoints() + getHitPointPoints() + getFatiguePointPoints();
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
        for (Advantage advantage : new FilteredIterator<>(mAdvantages.getTopLevelRows(), Advantage.class)) {
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
            if (!skill.canHaveChildren()) {
                mCachedSkillPoints += skill.getPoints();
            }
        }
    }

    /** @return The number of points spent on spells. */
    public int getSpellPoints() {
        return mCachedSpellPoints;
    }

    private void calculateSpellPoints() {
        mCachedSpellPoints = 0;
        for (Spell spell : getSpellsIterator()) {
            if (!spell.canHaveChildren()) {
                mCachedSpellPoints += spell.getPoints();
            }
        }
    }

    public int getCurrentHitPoints() {
        return getHitPointsAdj() - getHitPointsDamage();
    }

    /** @return The hit points (HP). */
    public int getHitPointsAdj() {
        return getStrength() + mHitPointsAdj + mHitPointBonus;
    }

    /**
     * Sets the hit points (HP).
     *
     * @param hp The new hit points.
     */
    public void setHitPointsAdj(int hp) {
        int oldHP = getHitPointsAdj();
        if (oldHP != hp) {
            postUndoEdit(I18n.Text("Hit Points Change"), ID_HIT_POINTS, Integer.valueOf(oldHP), Integer.valueOf(hp));
            startNotify();
            mHitPointsAdj = hp - (getStrength() + mHitPointBonus);
            mNeedAttributePointCalculation = true;
            notifyOfBaseHitPointChange();
            endNotify();
        }
    }

    /** @return The number of points spent on hit points. */
    public int getHitPointPoints() {
        int pts = 2 * mHitPointsAdj;
        if (!mSettings.useKnowYourOwnStrength()) {
            int sizeModifier = mProfile.getSizeModifier();
            if (sizeModifier > 0) {
                int rem;
                if (sizeModifier > 8) {
                    sizeModifier = 8;
                }
                pts *= 10 - sizeModifier;
                rem = pts % 10;
                pts /= 10;
                if (rem > 4) {
                    pts++;
                } else if (rem < -5) {
                    pts--;
                }
            }
        }
        return pts;
    }

    /** @return The hit point bonus. */
    public int getHitPointBonus() {
        return mHitPointBonus;
    }

    /** @param bonus The hit point bonus. */
    public void setHitPointBonus(int bonus) {
        if (mHitPointBonus != bonus) {
            mHitPointBonus = bonus;
            notifyOfBaseHitPointChange();
        }
    }

    private void notifyOfBaseHitPointChange() {
        startNotify();
        notify(ID_HIT_POINTS, Integer.valueOf(getHitPointsAdj()));
        notify(ID_DEATH_CHECK_1_HIT_POINTS, Integer.valueOf(getDeathCheck1HitPoints()));
        notify(ID_DEATH_CHECK_2_HIT_POINTS, Integer.valueOf(getDeathCheck2HitPoints()));
        notify(ID_DEATH_CHECK_3_HIT_POINTS, Integer.valueOf(getDeathCheck3HitPoints()));
        notify(ID_DEATH_CHECK_4_HIT_POINTS, Integer.valueOf(getDeathCheck4HitPoints()));
        notify(ID_DEAD_HIT_POINTS, Integer.valueOf(getDeadHitPoints()));
        notify(ID_REELING_HIT_POINTS, Integer.valueOf(getReelingHitPoints()));
        notify(ID_CURRENT_HP, Integer.valueOf(getHitPointsAdj() - mHitPointsDamage));
        endNotify();
    }

    /** @return The hit points damage. */
    public int getHitPointsDamage() {
        return mHitPointsDamage;
    }

    /**
     * Sets the hit points damage.
     *
     * @param damage The damage amount.
     */
    public void setHitPointsDamage(int damage) {
        if (mHitPointsDamage != damage) {
            postUndoEdit(I18n.Text("Current Hit Points Change"), ID_HIT_POINTS_DAMAGE, Integer.valueOf(mHitPointsDamage), Integer.valueOf(damage));
            mHitPointsDamage = damage;
            notifySingle(ID_HIT_POINTS_DAMAGE, Integer.valueOf(mHitPointsDamage));
            notifySingle(ID_CURRENT_HP, Integer.valueOf(getHitPointsAdj() - mHitPointsDamage));
        }
    }

    /** @return The number of hit points where "reeling" effects start. */
    public int getReelingHitPoints() {
        int hp        = getHitPointsAdj();
        int threshold = hp / 3;
        if (hp % 3 != 0) {
            threshold++;
        }
        return Math.max(--threshold, 0);
    }

    public boolean isReeling() {
        return getCurrentHitPoints() <= getReelingHitPoints();
    }

    public boolean isCollapsedFromHP() {
        return getCurrentHitPoints() <= getUnconsciousChecksHitPoints();
    }

    public boolean isDeathCheck1() {
        return getCurrentHitPoints() <= getDeathCheck1HitPoints();
    }

    public boolean isDeathCheck2() {
        return getCurrentHitPoints() <= getDeathCheck2HitPoints();
    }

    public boolean isDeathCheck3() {
        return getCurrentHitPoints() <= getDeathCheck3HitPoints();
    }

    public boolean isDeathCheck4() {
        return getCurrentHitPoints() <= getDeathCheck4HitPoints();
    }

    public boolean isDead() {
        return getCurrentHitPoints() <= getDeadHitPoints();
    }

    /** @return The number of hit points where unconsciousness checks must start being made. */
    @SuppressWarnings("static-method")
    public int getUnconsciousChecksHitPoints() {
        return 0;
    }

    /** @return The number of hit points where the first death check must be made. */
    public int getDeathCheck1HitPoints() {
        return -1 * getHitPointsAdj();
    }

    /** @return The number of hit points where the second death check must be made. */
    public int getDeathCheck2HitPoints() {
        return -2 * getHitPointsAdj();
    }

    /** @return The number of hit points where the third death check must be made. */
    public int getDeathCheck3HitPoints() {
        return -3 * getHitPointsAdj();
    }

    /** @return The number of hit points where the fourth death check must be made. */
    public int getDeathCheck4HitPoints() {
        return -4 * getHitPointsAdj();
    }

    /** @return The number of hit points where the character is just dead. */
    public int getDeadHitPoints() {
        return -5 * getHitPointsAdj();
    }

    /** @return The will. */
    public int getWillAdj() {
        return mWillAdj + mWillBonus + (mSettings.baseWillAndPerOn10() ? 10 : getIntelligence());
    }

    /** @param willAdj The new will. */
    public void setWillAdj(int willAdj) {
        int oldWill = getWillAdj();
        if (oldWill != willAdj) {
            postUndoEdit(I18n.Text("Will Change"), ID_WILL, Integer.valueOf(oldWill), Integer.valueOf(willAdj));
            updateWillInfo(willAdj - (mWillBonus + (mSettings.baseWillAndPerOn10() ? 10 : getIntelligence())), mWillBonus);
        }
    }

    /** @return The will bonus. */
    public int getWillBonus() {
        return mWillBonus;
    }

    /** @param bonus The new will bonus. */
    public void setWillBonus(int bonus) {
        if (mWillBonus != bonus) {
            updateWillInfo(mWillAdj, bonus);
        }
    }

    private void updateWillInfo(int will, int bonus) {
        mWillAdj = will;
        mWillBonus = bonus;

        startNotify();
        notify(ID_WILL, Integer.valueOf(getWillAdj()));
        notify(ID_FRIGHT_CHECK, Integer.valueOf(getFrightCheck()));
        updateSkills();
        mNeedAttributePointCalculation = true;
        endNotify();
    }

    /** Called to ensure notifications are sent out when the optional IQ rule use is changed. */
    public void updateWillAndPerceptionDueToOptionalIQRuleUseChange() {
        updateWillInfo(mWillAdj, mWillBonus);
        updatePerceptionInfo(mPerAdj, mPerceptionBonus);
    }

    /** @return The number of points spent on will. */
    public int getWillPoints() {
        return mWillAdj * 5;
    }

    /** @return The fright check. */
    public int getFrightCheck() {
        return getWillAdj() + mFrightCheckBonus;
    }

    /** @return The fright check bonus. */
    public int getFrightCheckBonus() {
        return mFrightCheckBonus;
    }

    /** @param bonus The new fright check bonus. */
    public void setFrightCheckBonus(int bonus) {
        if (mFrightCheckBonus != bonus) {
            mFrightCheckBonus = bonus;
            startNotify();
            notify(ID_FRIGHT_CHECK, Integer.valueOf(getFrightCheck()));
            endNotify();
        }
    }

    /** @return The vision. */
    public int getVision() {
        return getPerAdj() + mVisionBonus;
    }

    /** @return The vision bonus. */
    public int getVisionBonus() {
        return mVisionBonus;
    }

    /** @param bonus The new vision bonus. */
    public void setVisionBonus(int bonus) {
        if (mVisionBonus != bonus) {
            mVisionBonus = bonus;
            startNotify();
            notify(ID_VISION, Integer.valueOf(getVision()));
            endNotify();
        }
    }

    /** @return The hearing. */
    public int getHearing() {
        return getPerAdj() + mHearingBonus;
    }

    /** @return The hearing bonus. */
    public int getHearingBonus() {
        return mHearingBonus;
    }

    /** @param bonus The new hearing bonus. */
    public void setHearingBonus(int bonus) {
        if (mHearingBonus != bonus) {
            mHearingBonus = bonus;
            startNotify();
            notify(ID_HEARING, Integer.valueOf(getHearing()));
            endNotify();
        }
    }

    /** @return The touch perception. */
    public int getTouch() {
        return getPerAdj() + mTouchBonus;
    }

    /** @return The touch bonus. */
    public int getTouchBonus() {
        return mTouchBonus;
    }

    /** @param bonus The new touch bonus. */
    public void setTouchBonus(int bonus) {
        if (mTouchBonus != bonus) {
            mTouchBonus = bonus;
            startNotify();
            notify(ID_TOUCH, Integer.valueOf(getTouch()));
            endNotify();
        }
    }

    /** @return The taste and smell perception. */
    public int getTasteAndSmell() {
        return getPerAdj() + mTasteAndSmellBonus;
    }

    /** @return The taste and smell bonus. */
    public int getTasteAndSmellBonus() {
        return mTasteAndSmellBonus;
    }

    /** @param bonus The new taste and smell bonus. */
    public void setTasteAndSmellBonus(int bonus) {
        if (mTasteAndSmellBonus != bonus) {
            mTasteAndSmellBonus = bonus;
            startNotify();
            notify(ID_TASTE_AND_SMELL, Integer.valueOf(getTasteAndSmell()));
            endNotify();
        }
    }

    /** @return The perception (Per). */
    public int getPerAdj() {
        return mPerAdj + mPerceptionBonus + (mSettings.baseWillAndPerOn10() ? 10 : getIntelligence());
    }

    /**
     * Sets the perception.
     *
     * @param perAdj The new perception.
     */
    public void setPerAdj(int perAdj) {
        int oldPerception = getPerAdj();
        if (oldPerception != perAdj) {
            postUndoEdit(I18n.Text("Perception Change"), ID_PERCEPTION, Integer.valueOf(oldPerception), Integer.valueOf(perAdj));
            updatePerceptionInfo(perAdj - (mPerceptionBonus + (mSettings.baseWillAndPerOn10() ? 10 : getIntelligence())), mPerceptionBonus);
        }
    }

    /** @return The perception bonus. */
    public int getPerceptionBonus() {
        return mPerceptionBonus;
    }

    /** @param bonus The new perception bonus. */
    public void setPerceptionBonus(int bonus) {
        if (mPerceptionBonus != bonus) {
            updatePerceptionInfo(mPerAdj, bonus);
        }
    }

    private void updatePerceptionInfo(int perception, int bonus) {
        mPerAdj = perception;
        mPerceptionBonus = bonus;
        startNotify();
        notify(ID_PERCEPTION, Integer.valueOf(getPerAdj()));
        notify(ID_VISION, Integer.valueOf(getVision()));
        notify(ID_HEARING, Integer.valueOf(getHearing()));
        notify(ID_TASTE_AND_SMELL, Integer.valueOf(getTasteAndSmell()));
        notify(ID_TOUCH, Integer.valueOf(getTouch()));
        updateSkills();
        mNeedAttributePointCalculation = true;
        endNotify();
    }

    /** @return The number of points spent on perception. */
    public int getPerceptionPoints() {
        return mPerAdj * 5;
    }

    public int getCurrentFatiguePoints() {
        return getFatiguePoints() - getFatiguePointsDamage();
    }

    /** @return The fatigue points (FP). */
    public int getFatiguePoints() {
        return getHealth() + mFatiguePoints + mFatiguePointBonus;
    }

    /**
     * Sets the fatigue points (FP).
     *
     * @param fp The new fatigue points.
     */
    public void setFatiguePoints(int fp) {
        int oldFP = getFatiguePoints();
        if (oldFP != fp) {
            postUndoEdit(I18n.Text("Fatigue Points Change"), ID_FATIGUE_POINTS, Integer.valueOf(oldFP), Integer.valueOf(fp));
            startNotify();
            mFatiguePoints = fp - (getHealth() + mFatiguePointBonus);
            mNeedAttributePointCalculation = true;
            notifyOfBaseFatiguePointChange();
            endNotify();
        }
    }

    /** @return The number of points spent on fatigue points. */
    public int getFatiguePointPoints() {
        return 3 * mFatiguePoints;
    }

    /** @return The fatigue point bonus. */
    public int getFatiguePointBonus() {
        return mFatiguePointBonus;
    }

    /** @param bonus The fatigue point bonus. */
    public void setFatiguePointBonus(int bonus) {
        if (mFatiguePointBonus != bonus) {
            mFatiguePointBonus = bonus;
            notifyOfBaseFatiguePointChange();
        }
    }

    private void notifyOfBaseFatiguePointChange() {
        startNotify();
        notify(ID_FATIGUE_POINTS, Integer.valueOf(getFatiguePoints()));
        notify(ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, Integer.valueOf(getUnconsciousChecksFatiguePoints()));
        notify(ID_UNCONSCIOUS_FATIGUE_POINTS, Integer.valueOf(getUnconsciousFatiguePoints()));
        notify(ID_TIRED_FATIGUE_POINTS, Integer.valueOf(getTiredFatiguePoints()));
        notify(ID_CURRENT_FP, Integer.valueOf(getFatiguePoints() - mFatiguePointsDamage));
        endNotify();
    }

    /** @return The fatigue points damage. */
    public int getFatiguePointsDamage() {
        return mFatiguePointsDamage;
    }

    /**
     * Sets the fatigue points damage.
     *
     * @param damage The damage amount.
     */
    public void setFatiguePointsDamage(int damage) {
        if (mFatiguePointsDamage != damage) {
            postUndoEdit(I18n.Text("Current Fatigue Points Change"), ID_FATIGUE_POINTS_DAMAGE, Integer.valueOf(mFatiguePointsDamage), Integer.valueOf(damage));
            mFatiguePointsDamage = damage;
            notifySingle(ID_FATIGUE_POINTS_DAMAGE, Integer.valueOf(mFatiguePointsDamage));
            notifySingle(ID_CURRENT_FP, Integer.valueOf(getFatiguePoints() - mFatiguePointsDamage));
        }
    }

    /** @return The number of fatigue points where "tired" effects start. */
    public int getTiredFatiguePoints() {
        int fp        = getFatiguePoints();
        int threshold = fp / 3;
        if (fp % 3 != 0) {
            threshold++;
        }
        return Math.max(--threshold, 0);
    }

    public boolean isTired() {
        return getCurrentFatiguePoints() <= getTiredFatiguePoints();
    }

    public boolean isCollapsedFromFP() {
        return getCurrentFatiguePoints() <= getUnconsciousChecksFatiguePoints();
    }

    public boolean isUnconscious() {
        return getCurrentFatiguePoints() <= getUnconsciousFatiguePoints();
    }

    /** @return The number of fatigue points where unconsciousness checks must start being made. */
    @SuppressWarnings("static-method")
    public int getUnconsciousChecksFatiguePoints() {
        return 0;
    }

    /** @return The number of hit points where the character falls over, unconscious. */
    public int getUnconsciousFatiguePoints() {
        return -1 * getFatiguePoints();
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

    /** @return The outline model for the character's advantages. */
    public OutlineModel getAdvantagesModel() {
        return mAdvantages;
    }

    /**
     * @param includeDisabled {@code true} if disabled entries should be included.
     * @return A recursive iterator over the character's advantages.
     */
    public RowIterator<Advantage> getAdvantagesIterator(boolean includeDisabled) {
        if (includeDisabled) {
            return new RowIterator<>(mAdvantages);
        }
        return new RowIterator<>(mAdvantages, (row) -> row.isEnabled());
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

    /** @return The outline model for the character's skills. */
    public OutlineModel getSkillsRoot() {
        return mSkills;
    }

    /** @return A recursive iterable for the character's skills. */
    public RowIterator<Skill> getSkillsIterator() {
        return new RowIterator<>(mSkills);
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

    /** @return The outline model for the character's spells. */
    public OutlineModel getSpellsRoot() {
        return mSpells;
    }

    /** @return A recursive iterator over the character's spells. */
    public RowIterator<Spell> getSpellsIterator() {
        return new RowIterator<>(mSpells);
    }

    /** @return The outline model for the character's equipment. */
    public OutlineModel getEquipmentRoot() {
        return mEquipment;
    }

    /** @return A recursive iterator over the character's equipment. */
    public RowIterator<Equipment> getEquipmentIterator() {
        return new RowIterator<>(mEquipment);
    }

    /** @return The outline model for the character's other equipment. */
    public OutlineModel getOtherEquipmentRoot() {
        return mOtherEquipment;
    }

    /** @return A recursive iterator over the character's other equipment. */
    public RowIterator<Equipment> getOtherEquipmentIterator() {
        return new RowIterator<>(mOtherEquipment);
    }

    /** @return The outline model for the character's notes. */
    public OutlineModel getNotesRoot() {
        return mNotes;
    }

    /** @return A recursive iterator over the character's notes. */
    public RowIterator<Note> getNoteIterator() {
        return new RowIterator<>(mNotes);
    }

    public boolean processFeaturesAndPrereqs() {
        boolean needRepaint = processFeatures();
        needRepaint |= processPrerequisites(getAdvantagesIterator(false));
        needRepaint |= processPrerequisites(getSkillsIterator());
        needRepaint |= processPrerequisites(getSpellsIterator());
        needRepaint |= processPrerequisites(getEquipmentIterator());
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
                row.setReasonForUnsatisfied(builder.toString().replaceAll("<ul>", "<ul style='margin-top: 0; margin-bottom: 0;'>"));
            }
        }
        return needRepaint;
    }

    /** @param map The new feature map. */
    public void setFeatureMap(HashMap<String, ArrayList<Feature>> map) {
        mFeatureMap = map;
        mSkillsUpdated = false;
        mSpellsUpdated = false;

        startNotify();
        setStrengthBonus(getIntegerBonusFor(ID_STRENGTH));
        setStrengthCostReduction(getCostReductionFor(ID_STRENGTH));
        setLiftingStrengthBonus(getIntegerBonusFor(ID_LIFTING_STRENGTH));
        setStrikingStrengthBonus(getIntegerBonusFor(ID_STRIKING_STRENGTH));
        setDexterityBonus(getIntegerBonusFor(ID_DEXTERITY));
        setDexterityCostReduction(getCostReductionFor(ID_DEXTERITY));
        setIntelligenceBonus(getIntegerBonusFor(ID_INTELLIGENCE));
        setIntelligenceCostReduction(getCostReductionFor(ID_INTELLIGENCE));
        setHealthBonus(getIntegerBonusFor(ID_HEALTH));
        setHealthCostReduction(getCostReductionFor(ID_HEALTH));
        setWillBonus(getIntegerBonusFor(ID_WILL));
        setFrightCheckBonus(getIntegerBonusFor(ID_FRIGHT_CHECK));
        setPerceptionBonus(getIntegerBonusFor(ID_PERCEPTION));
        setVisionBonus(getIntegerBonusFor(ID_VISION));
        setHearingBonus(getIntegerBonusFor(ID_HEARING));
        setTasteAndSmellBonus(getIntegerBonusFor(ID_TASTE_AND_SMELL));
        setTouchBonus(getIntegerBonusFor(ID_TOUCH));
        setHitPointBonus(getIntegerBonusFor(ID_HIT_POINTS));
        setFatiguePointBonus(getIntegerBonusFor(ID_FATIGUE_POINTS));
        mProfile.update();
        setDodgeBonus(getIntegerBonusFor(ID_DODGE_BONUS));
        setParryBonus(getIntegerBonusFor(ID_PARRY_BONUS));
        setBlockBonus(getIntegerBonusFor(ID_BLOCK_BONUS));
        setBasicSpeedBonus(getDoubleBonusFor(ID_BASIC_SPEED));
        setBasicMoveBonus(getIntegerBonusFor(ID_BASIC_MOVE));
        mArmor.update();
        if (!mSkillsUpdated) {
            updateSkills();
        }
        if (!mSpellsUpdated) {
            updateSpells();
        }
        endNotify();
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
     * @return The bonuses.
     */
    public List<WeaponBonus> getWeaponComparedBonusesFor(String id, String nameQualifier, String specializationQualifier, Set<String> categoriesQualifier, StringBuilder toolTip) {
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
                        if (bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(specializationQualifier) && bonus.getLevelCriteria().matches(rsl) && bonus.matchesCategories(categoriesQualifier)) {
                            bonuses.add(bonus);
                            bonus.addToToolTip(toolTip);
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
     * @return The bonuses.
     */
    public List<WeaponBonus> getNamedWeaponBonusesFor(String id, String nameQualifier, String usageQualifier, Set<String> categoriesQualifier, StringBuilder toolTip) {
        List<WeaponBonus> bonuses = new ArrayList<>();
        List<Feature>     list    = mFeatureMap.get(id.toLowerCase());
        if (list != null) {
            for (Feature feature : list) {
                if (feature instanceof WeaponBonus) {
                    WeaponBonus bonus = (WeaponBonus) feature;
                    if (bonus.getWeaponSelectionType() == WeaponSelectionType.WEAPONS_WITH_NAME && bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(usageQualifier) && bonus.matchesCategories(categoriesQualifier)) {
                        bonuses.add(bonus);
                        bonus.addToToolTip(toolTip);
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
     * @param id     The ID of the field being changed.
     * @param before The original value.
     * @param after  The new value.
     */
    void postUndoEdit(String name, String id, Object before, Object after) {
        StdUndoManager mgr = getUndoManager();
        if (!mgr.isInTransaction()) {
            if (before instanceof ListRow ? !((ListRow) before).isEquivalentTo(after) : !before.equals(after)) {
                addEdit(new CharacterFieldUndo(this, name, id, before, after));
            }
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
}
