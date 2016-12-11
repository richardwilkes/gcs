/*
 * Copyright (c) 1998-2016 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageContainerType;
import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.feature.AttributeBonusLimitation;
import com.trollworks.gcs.feature.Bonus;
import com.trollworks.gcs.feature.BonusAttributeType;
import com.trollworks.gcs.feature.CostReduction;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.gcs.feature.SpellBonus;
import com.trollworks.gcs.feature.WeaponBonus;
import com.trollworks.gcs.modifier.Modifier;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.preferences.OutputPreferences;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.collections.FilteredIterator;
import com.trollworks.toolkit.io.Log;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.image.StdImageSet;
import com.trollworks.toolkit.ui.print.PrintManager;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.RowIterator;
import com.trollworks.toolkit.utility.Dice;
import com.trollworks.toolkit.utility.FileType;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.undo.StdUndoManager;
import com.trollworks.toolkit.utility.units.LengthUnits;
import com.trollworks.toolkit.utility.units.WeightUnits;
import com.trollworks.toolkit.utility.units.WeightValue;

import java.io.File;
import java.io.IOException;
import java.text.DateFormat;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.HashMap;
import java.util.HashSet;

/** A GURPS character. */
public class GURPSCharacter extends DataFile {
	@Localize("Modified at {0} on {1}")
	@Localize(locale = "de", value = "Verändert um {0} am {1}.")
	@Localize(locale = "ru", value = "Изменен в {0} {1}")
	@Localize(locale = "es", value = "Modificado el {0} a las {1}")
	private static String	LAST_MODIFIED;
	@Localize("Created On Change")
	@Localize(locale = "de", value = "Erstellt am ändern")
	@Localize(locale = "ru", value = "Изменить дату создания")
	@Localize(locale = "es", value = "Cambiar Fecha de creacción")
	private static String	CREATED_ON_UNDO;
	@Localize("Strength Change")
	@Localize(locale = "de", value = "Stärke ändern")
	@Localize(locale = "ru", value = "Изменить силу")
	@Localize(locale = "es", value = "Cambiar Fuerza")
	private static String	STRENGTH_UNDO;
	@Localize("Dexterity Change")
	@Localize(locale = "de", value = "Geschick ändern")
	@Localize(locale = "ru", value = "Изменить ловкость")
	@Localize(locale = "es", value = "Cambiar Destreza")
	private static String	DEXTERITY_UNDO;
	@Localize("Intelligence Change")
	@Localize(locale = "de", value = "Intelligenz ändern")
	@Localize(locale = "ru", value = "Смена интеллекта")
	@Localize(locale = "es", value = "Cambiar Inteligencia")
	private static String	INTELLIGENCE_UNDO;
	@Localize("Health Change")
	@Localize(locale = "de", value = "Konstitution ändern")
	@Localize(locale = "ru", value = "Смена уровня здоровья")
	@Localize(locale = "es", value = "Cambiar Salud")
	private static String	HEALTH_UNDO;
	@Localize("Basic Speed Change")
	@Localize(locale = "de", value = "Grundgeschwindigkeit ändern")
	@Localize(locale = "ru", value = "Смена Базовой Скорости")
	@Localize(locale = "es", value = "Cambiar Velocidad Básica")
	private static String	BASIC_SPEED_UNDO;
	@Localize("Basic Move Change")
	@Localize(locale = "de", value = "Grundbewegung ändern")
	@Localize(locale = "ru", value = "Смена Базового Движения")
	@Localize(locale = "es", value = "Cambiar Movimiento Básico")
	private static String	BASIC_MOVE_UNDO;
	@Localize("Perception Change")
	@Localize(locale = "de", value = "Wahrnehmung ändern")
	@Localize(locale = "ru", value = "Смена восприятия")
	@Localize(locale = "es", value = "Cambiar Percepción")
	private static String	PERCEPTION_UNDO;
	@Localize("Will Change")
	@Localize(locale = "de", value = "Wille ändern")
	@Localize(locale = "ru", value = "Изменить волю")
	@Localize(locale = "es", value = "Cambiar Voluntad")
	private static String	WILL_UNDO;
	@Localize("Earned Points Change")
	@Localize(locale = "de", value = "Verdiente Punkte ändern")
	@Localize(locale = "ru", value = "Изменить заработаные очки")
	@Localize(locale = "es", value = "Cambiar Puntos Obtenidos")
	private static String	EARNED_POINTS_UNDO;
	@Localize("Hit Points Change")
	@Localize(locale = "de", value = "Normale Trefferpunkte ändern")
	@Localize(locale = "ru", value = "Изменить единицы здоровья")
	@Localize(locale = "es", value = "Cambiar Puntos de Vida")
	private static String	HIT_POINTS_UNDO;
	@Localize("Current Hit Points Change")
	@Localize(locale = "de", value = "Aktuelle Trefferpunkte ändern")
	@Localize(locale = "ru", value = "Смена текущих очков (единиц) жизни")
	@Localize(locale = "es", value = "Cambiar Puntos de Vida Actuales")
	private static String	CURRENT_HIT_POINTS_UNDO;
	@Localize("Fatigue Points Change")
	@Localize(locale = "de", value = "Normale Erschöpfungspunkte ändern")
	@Localize(locale = "ru", value = "Изменить очки усталости")
	@Localize(locale = "es", value = "Cambiar Puntos de Fatiga")
	private static String	FATIGUE_POINTS_UNDO;
	@Localize("Current Fatigue Points Change")
	@Localize(locale = "de", value = "Aktuelle Erschöpfungspunkte ändern")
	@Localize(locale = "ru", value = "Изменить текущие единицы усталости")
	@Localize(locale = "es", value = "Cambiar Puntos de Fatiga Actuales")
	private static String	CURRENT_FATIGUE_POINTS_UNDO;
	@Localize("Include Punch In Weapons")
	@Localize(locale = "de", value = "Schlag als Waffe aufführen")
	@Localize(locale = "ru", value = "Отображать удар в оружии")
	@Localize(locale = "es", value = "Incluir Puñetazo como Arma")
	private static String	INCLUDE_PUNCH_UNDO;
	@Localize("Include Kick In Weapons")
	@Localize(locale = "de", value = "Tritt als Waffe aufführen")
	@Localize(locale = "ru", value = "Отображать пинок в оружии")
	@Localize(locale = "es", value = "Incluir Patada como Arma")
	private static String	INCLUDE_KICK_UNDO;
	@Localize("Include Kick w/Boots In Weapons")
	@Localize(locale = "de", value = "Tritt mit Schuh als Waffe aufführen")
	@Localize(locale = "ru", value = "Отображать пинок (в ботинке) в оружии")
	@Localize(locale = "es", value = "Incluir Patada con botas como Arma")
	private static String	INCLUDE_BOOTS_UNDO;
	@Localize("Unable to set a value for %s")
	@Localize(locale = "de", value = "Kann keinen Wert für %s setzen")
	@Localize(locale = "ru", value = "Невозможно установить значение для %s")
	@Localize(locale = "es", value = "No puede establecerse un valor para %s")
	private static String	UNABLE_TO_SET_VALUE;

	static {
		Localization.initialize();
	}

	/** The extension for character sheets. */
	public static final String					EXTENSION								= "gcs";														//$NON-NLS-1$
	private static final int					CURRENT_VERSION							= 3;
	private static final String					EMPTY									= "";															//$NON-NLS-1$
	private static final String					TAG_ROOT								= "character";													//$NON-NLS-1$
	private static final String					TAG_CREATED_DATE						= "created_date";												//$NON-NLS-1$
	private static final String					TAG_MODIFIED_DATE						= "modified_date";												//$NON-NLS-1$
	private static final String					TAG_CURRENT_HP							= "current_hp";													//$NON-NLS-1$
	private static final String					TAG_CURRENT_FP							= "current_fp";													//$NON-NLS-1$
	private static final String					TAG_UNSPENT_POINTS						= "unspent_points";												//$NON-NLS-1$
	private static final String					TAG_TOTAL_POINTS						= "total_points";												//$NON-NLS-1$
	private static final String					TAG_INCLUDE_PUNCH						= "include_punch";												//$NON-NLS-1$
	private static final String					TAG_INCLUDE_KICK						= "include_kick";												//$NON-NLS-1$
	private static final String					TAG_INCLUDE_BOOTS						= "include_kick_with_boots";									//$NON-NLS-1$
	private static final String					ATTRIBUTE_CARRIED						= "carried";													//$NON-NLS-1$
	/** The prefix for all character IDs. */
	public static final String					CHARACTER_PREFIX						= "gcs.";														//$NON-NLS-1$
	/** The field ID for last modified date changes. */
	public static final String					ID_LAST_MODIFIED						= CHARACTER_PREFIX + "LastModifiedDate";						//$NON-NLS-1$
	/** The field ID for created on date changes. */
	public static final String					ID_CREATED_ON							= CHARACTER_PREFIX + "CreatedOn";								//$NON-NLS-1$
	/** The field ID for include punch changes. */
	public static final String					ID_INCLUDE_PUNCH						= CHARACTER_PREFIX + "IncludePunch";							//$NON-NLS-1$
	/** The field ID for include kick changes. */
	public static final String					ID_INCLUDE_KICK							= CHARACTER_PREFIX + "IncludeKickFeet";							//$NON-NLS-1$
	/** The field ID for include kick with boots changes. */
	public static final String					ID_INCLUDE_BOOTS						= CHARACTER_PREFIX + "IncludeKickBoots";						//$NON-NLS-1$
	/**
	 * The prefix used to indicate a point value is requested from {@link #getValueForID(String)}.
	 */
	public static final String					POINTS_PREFIX							= CHARACTER_PREFIX + "points.";									//$NON-NLS-1$
	/** The prefix used in front of all IDs for basic attributes. */
	public static final String					ATTRIBUTES_PREFIX						= CHARACTER_PREFIX + "ba.";										//$NON-NLS-1$
	/** The field ID for strength (ST) changes. */
	public static final String					ID_STRENGTH								= ATTRIBUTES_PREFIX + BonusAttributeType.ST.name();
	/** The field ID for lifting strength bonuses -- used by features. */
	public static final String					ID_LIFTING_STRENGTH						= ID_STRENGTH + AttributeBonusLimitation.LIFTING_ONLY.name();
	/** The field ID for striking strength bonuses -- used by features. */
	public static final String					ID_STRIKING_STRENGTH					= ID_STRENGTH + AttributeBonusLimitation.STRIKING_ONLY.name();
	/** The field ID for dexterity (DX) changes. */
	public static final String					ID_DEXTERITY							= ATTRIBUTES_PREFIX + BonusAttributeType.DX.name();
	/** The field ID for intelligence (IQ) changes. */
	public static final String					ID_INTELLIGENCE							= ATTRIBUTES_PREFIX + BonusAttributeType.IQ.name();
	/** The field ID for health (HT) changes. */
	public static final String					ID_HEALTH								= ATTRIBUTES_PREFIX + BonusAttributeType.HT.name();
	/** The field ID for perception changes. */
	public static final String					ID_PERCEPTION							= ATTRIBUTES_PREFIX + BonusAttributeType.PERCEPTION.name();
	/** The field ID for vision changes. */
	public static final String					ID_VISION								= ATTRIBUTES_PREFIX + BonusAttributeType.VISION.name();
	/** The field ID for hearing changes. */
	public static final String					ID_HEARING								= ATTRIBUTES_PREFIX + BonusAttributeType.HEARING.name();
	/** The field ID for taste changes. */
	public static final String					ID_TASTE_AND_SMELL						= ATTRIBUTES_PREFIX + BonusAttributeType.TASTE_SMELL.name();
	/** The field ID for smell changes. */
	public static final String					ID_TOUCH								= ATTRIBUTES_PREFIX + BonusAttributeType.TOUCH.name();
	/** The field ID for will changes. */
	public static final String					ID_WILL									= ATTRIBUTES_PREFIX + BonusAttributeType.WILL.name();
	/** The field ID for fright check changes. */
	public static final String					ID_FRIGHT_CHECK							= ATTRIBUTES_PREFIX + BonusAttributeType.FRIGHT_CHECK.name();
	/** The field ID for basic speed changes. */
	public static final String					ID_BASIC_SPEED							= ATTRIBUTES_PREFIX + BonusAttributeType.SPEED.name();
	/** The field ID for basic move changes. */
	public static final String					ID_BASIC_MOVE							= ATTRIBUTES_PREFIX + BonusAttributeType.MOVE.name();
	/** The prefix used in front of all IDs for dodge changes. */
	public static final String					DODGE_PREFIX							= ATTRIBUTES_PREFIX + BonusAttributeType.DODGE.name() + "#.";	//$NON-NLS-1$
	/** The field ID for dodge bonus changes. */
	public static final String					ID_DODGE_BONUS							= ATTRIBUTES_PREFIX + BonusAttributeType.DODGE.name();
	/** The field ID for parry bonus changes. */
	public static final String					ID_PARRY_BONUS							= ATTRIBUTES_PREFIX + BonusAttributeType.PARRY.name();
	/** The field ID for block bonus changes. */
	public static final String					ID_BLOCK_BONUS							= ATTRIBUTES_PREFIX + BonusAttributeType.BLOCK.name();
	/** The prefix used in front of all IDs for move changes. */
	public static final String					MOVE_PREFIX								= ATTRIBUTES_PREFIX + BonusAttributeType.MOVE.name() + "#.";	//$NON-NLS-1$
	/** The field ID for carried weight changes. */
	public static final String					ID_CARRIED_WEIGHT						= CHARACTER_PREFIX + "CarriedWeight";							//$NON-NLS-1$
	/** The field ID for carried wealth changes. */
	public static final String					ID_CARRIED_WEALTH						= CHARACTER_PREFIX + "CarriedWealth";							//$NON-NLS-1$
	/** The prefix used in front of all IDs for encumbrance changes. */
	public static final String					MAXIMUM_CARRY_PREFIX					= ATTRIBUTES_PREFIX + "MaximumCarry";							//$NON-NLS-1$
	private static final String					LIFT_PREFIX								= ATTRIBUTES_PREFIX + "lift.";									//$NON-NLS-1$
	/** The field ID for basic lift changes. */
	public static final String					ID_BASIC_LIFT							= LIFT_PREFIX + "BasicLift";									//$NON-NLS-1$
	/** The field ID for one-handed lift changes. */
	public static final String					ID_ONE_HANDED_LIFT						= LIFT_PREFIX + "OneHandedLift";								//$NON-NLS-1$
	/** The field ID for two-handed lift changes. */
	public static final String					ID_TWO_HANDED_LIFT						= LIFT_PREFIX + "TwoHandedLift";								//$NON-NLS-1$
	/** The field ID for shove and knock over changes. */
	public static final String					ID_SHOVE_AND_KNOCK_OVER					= LIFT_PREFIX + "ShoveAndKnockOver";							//$NON-NLS-1$
	/** The field ID for running shove and knock over changes. */
	public static final String					ID_RUNNING_SHOVE_AND_KNOCK_OVER			= LIFT_PREFIX + "RunningShoveAndKnockOver";						//$NON-NLS-1$
	/** The field ID for carry on back changes. */
	public static final String					ID_CARRY_ON_BACK						= LIFT_PREFIX + "CarryOnBack";									//$NON-NLS-1$
	/** The field ID for carry on back changes. */
	public static final String					ID_SHIFT_SLIGHTLY						= LIFT_PREFIX + "ShiftSlightly";								//$NON-NLS-1$
	/** The prefix used in front of all IDs for point summaries. */
	public static final String					POINT_SUMMARY_PREFIX					= CHARACTER_PREFIX + "ps.";										//$NON-NLS-1$
	/** The field ID for point total changes. */
	public static final String					ID_TOTAL_POINTS							= POINT_SUMMARY_PREFIX + "TotalPoints";							//$NON-NLS-1$
	/** The field ID for attribute point summary changes. */
	public static final String					ID_ATTRIBUTE_POINTS						= POINT_SUMMARY_PREFIX + "AttributePoints";						//$NON-NLS-1$
	/** The field ID for advantage point summary changes. */
	public static final String					ID_ADVANTAGE_POINTS						= POINT_SUMMARY_PREFIX + "AdvantagePoints";						//$NON-NLS-1$
	/** The field ID for disadvantage point summary changes. */
	public static final String					ID_DISADVANTAGE_POINTS					= POINT_SUMMARY_PREFIX + "DisadvantagePoints";					//$NON-NLS-1$
	/** The field ID for quirk point summary changes. */
	public static final String					ID_QUIRK_POINTS							= POINT_SUMMARY_PREFIX + "QuirkPoints";							//$NON-NLS-1$
	/** The field ID for skill point summary changes. */
	public static final String					ID_SKILL_POINTS							= POINT_SUMMARY_PREFIX + "SkillPoints";							//$NON-NLS-1$
	/** The field ID for spell point summary changes. */
	public static final String					ID_SPELL_POINTS							= POINT_SUMMARY_PREFIX + "SpellPoints";							//$NON-NLS-1$
	/** The field ID for racial point summary changes. */
	public static final String					ID_RACE_POINTS							= POINT_SUMMARY_PREFIX + "RacePoints";							//$NON-NLS-1$
	/** The field ID for earned point changes. */
	public static final String					ID_EARNED_POINTS						= POINT_SUMMARY_PREFIX + "EarnedPoints";						//$NON-NLS-1$
	/** The prefix used in front of all IDs for basic damage. */
	public static final String					BASIC_DAMAGE_PREFIX						= CHARACTER_PREFIX + "bd.";										//$NON-NLS-1$
	/** The field ID for basic thrust damage changes. */
	public static final String					ID_BASIC_THRUST							= BASIC_DAMAGE_PREFIX + "Thrust";								//$NON-NLS-1$
	/** The field ID for basic swing damage changes. */
	public static final String					ID_BASIC_SWING							= BASIC_DAMAGE_PREFIX + "Swing";								//$NON-NLS-1$
	private static final String					HIT_POINTS_PREFIX						= ATTRIBUTES_PREFIX + "derived_hp.";							//$NON-NLS-1$
	/** The field ID for hit point changes. */
	public static final String					ID_HIT_POINTS							= ATTRIBUTES_PREFIX + BonusAttributeType.HP.name();
	/** The field ID for current hit point changes. */
	public static final String					ID_CURRENT_HIT_POINTS					= HIT_POINTS_PREFIX + "Current";								//$NON-NLS-1$
	/** The field ID for reeling hit point changes. */
	public static final String					ID_REELING_HIT_POINTS					= HIT_POINTS_PREFIX + "Reeling";								//$NON-NLS-1$
	/** The field ID for unconscious check hit point changes. */
	public static final String					ID_UNCONSCIOUS_CHECKS_HIT_POINTS		= HIT_POINTS_PREFIX + "UnconsciousChecks";						//$NON-NLS-1$
	/** The field ID for death check #1 hit point changes. */
	public static final String					ID_DEATH_CHECK_1_HIT_POINTS				= HIT_POINTS_PREFIX + "DeathCheck1";							//$NON-NLS-1$
	/** The field ID for death check #2 hit point changes. */
	public static final String					ID_DEATH_CHECK_2_HIT_POINTS				= HIT_POINTS_PREFIX + "DeathCheck2";							//$NON-NLS-1$
	/** The field ID for death check #3 hit point changes. */
	public static final String					ID_DEATH_CHECK_3_HIT_POINTS				= HIT_POINTS_PREFIX + "DeathCheck3";							//$NON-NLS-1$
	/** The field ID for death check #4 hit point changes. */
	public static final String					ID_DEATH_CHECK_4_HIT_POINTS				= HIT_POINTS_PREFIX + "DeathCheck4";							//$NON-NLS-1$
	/** The field ID for dead hit point changes. */
	public static final String					ID_DEAD_HIT_POINTS						= HIT_POINTS_PREFIX + "Dead";									//$NON-NLS-1$
	private static final String					FATIGUE_POINTS_PREFIX					= ATTRIBUTES_PREFIX + "derived_fp.";							//$NON-NLS-1$
	/** The field ID for fatigue point changes. */
	public static final String					ID_FATIGUE_POINTS						= ATTRIBUTES_PREFIX + BonusAttributeType.FP.name();
	/** The field ID for current fatigue point changes. */
	public static final String					ID_CURRENT_FATIGUE_POINTS				= FATIGUE_POINTS_PREFIX + "Current";							//$NON-NLS-1$
	/** The field ID for tired fatigue point changes. */
	public static final String					ID_TIRED_FATIGUE_POINTS					= FATIGUE_POINTS_PREFIX + "Tired";								//$NON-NLS-1$
	/** The field ID for unconscious check fatigue point changes. */
	public static final String					ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS	= FATIGUE_POINTS_PREFIX + "UnconsciousChecks";					//$NON-NLS-1$
	/** The field ID for unconscious fatigue point changes. */
	public static final String					ID_UNCONSCIOUS_FATIGUE_POINTS			= FATIGUE_POINTS_PREFIX + "Unconscious";						//$NON-NLS-1$
	private long								mLastModified;
	private long								mCreatedOn;
	private HashMap<String, ArrayList<Feature>>	mFeatureMap;
	private int									mStrength;
	private int									mStrengthBonus;
	private int									mLiftingStrengthBonus;
	private int									mStrikingStrengthBonus;
	private int									mStrengthCostReduction;
	private int									mDexterity;
	private int									mDexterityBonus;
	private int									mDexterityCostReduction;
	private int									mIntelligence;
	private int									mIntelligenceBonus;
	private int									mIntelligenceCostReduction;
	private int									mHealth;
	private int									mHealthBonus;
	private int									mHealthCostReduction;
	private int									mWill;
	private int									mWillBonus;
	private int									mFrightCheckBonus;
	private int									mPerception;
	private int									mPerceptionBonus;
	private int									mVisionBonus;
	private int									mHearingBonus;
	private int									mTasteAndSmellBonus;
	private int									mTouchBonus;
	private String								mCurrentHitPoints;
	private int									mHitPoints;
	private int									mHitPointBonus;
	private int									mFatiguePoints;
	private String								mCurrentFatiguePoints;
	private int									mFatiguePointBonus;
	private double								mSpeed;
	private double								mSpeedBonus;
	private int									mMove;
	private int									mMoveBonus;
	private int									mDodgeBonus;
	private int									mParryBonus;
	private int									mBlockBonus;
	private int									mTotalPoints;
	private Profile								mDescription;
	private Armor								mArmor;
	private OutlineModel						mAdvantages;
	private OutlineModel						mSkills;
	private OutlineModel						mSpells;
	private OutlineModel						mEquipment;
	private OutlineModel						mNotes;
	private boolean								mDidModify;
	private boolean								mNeedAttributePointCalculation;
	private boolean								mNeedAdvantagesPointCalculation;
	private boolean								mNeedSkillPointCalculation;
	private boolean								mNeedSpellPointCalculation;
	private boolean								mNeedEquipmentCalculation;
	private WeightValue							mCachedWeightCarried;
	private double								mCachedWealthCarried;
	private int									mCachedAttributePoints;
	private int									mCachedAdvantagePoints;
	private int									mCachedDisadvantagePoints;
	private int									mCachedQuirkPoints;
	private int									mCachedSkillPoints;
	private int									mCachedSpellPoints;
	private int									mCachedRacePoints;
	private boolean								mSkillsUpdated;
	private boolean								mSpellsUpdated;
	private PrintManager						mPageSettings;
	private boolean								mIncludePunch;
	private boolean								mIncludeKick;
	private boolean								mIncludeKickBoots;

	/** Creates a new character with only default values set. */
	public GURPSCharacter() {
		super();
		characterInitialize(true);
		calculateAll();
	}

	/**
	 * Creates a new character from the specified file.
	 *
	 * @param file The file to load the data from.
	 * @throws IOException if the data cannot be read or the file doesn't contain a valid character
	 *             sheet.
	 */
	public GURPSCharacter(File file) throws IOException {
		super();
		load(file);
	}

	private void characterInitialize(boolean full) {
		mFeatureMap = new HashMap<>();
		mAdvantages = new OutlineModel();
		mSkills = new OutlineModel();
		mSpells = new OutlineModel();
		mEquipment = new OutlineModel();
		mNotes = new OutlineModel();
		mTotalPoints = SheetPreferences.getInitialPoints();
		mStrength = 10;
		mDexterity = 10;
		mIntelligence = 10;
		mHealth = 10;
		mCurrentHitPoints = EMPTY;
		mCurrentFatiguePoints = EMPTY;
		mDescription = new Profile(this, full);
		mArmor = new Armor(this);
		mIncludePunch = true;
		mIncludeKick = true;
		mIncludeKickBoots = true;
		mCachedWeightCarried = new WeightValue(0, SheetPreferences.getWeightUnits());
		mPageSettings = OutputPreferences.getDefaultPageSettings();
		mLastModified = System.currentTimeMillis();
		mCreatedOn = mLastModified;
		// This will force the long value to match the string value.
		setCreatedOn(getCreatedOn());
	}

	/** @return The page settings. May return <code>null</code> if no printer has been defined. */
	public PrintManager getPageSettings() {
		return mPageSettings;
	}

	@Override
	public FileType getFileType() {
		return FileType.getByExtension(EXTENSION);
	}

	@Override
	public StdImageSet getFileIcons() {
		return GCSImages.getCharacterSheetDocumentIcons();
	}

	@Override
	protected final void loadSelf(XMLReader reader, LoadState state) throws IOException {
		String marker = reader.getMarker();
		int unspentPoints = 0;

		characterInitialize(false);
		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String name = reader.getName();

				if (state.mDataFileVersion == 0) {
					if (mDescription.loadTag(reader, name)) {
						continue;
					}
				}

				if (Profile.TAG_ROOT.equals(name)) {
					mDescription.load(reader);
				} else if (TAG_CREATED_DATE.equals(name)) {
					mCreatedOn = Numbers.extractDate(reader.readText());
				} else if (TAG_MODIFIED_DATE.equals(name)) {
					mLastModified = Numbers.extractDateTime(reader.readText());
				} else if (BonusAttributeType.HP.getXMLTag().equals(name)) {
					mHitPoints = reader.readInteger(0);
				} else if (TAG_CURRENT_HP.equals(name)) {
					mCurrentHitPoints = reader.readText();
				} else if (BonusAttributeType.FP.getXMLTag().equals(name)) {
					mFatiguePoints = reader.readInteger(0);
				} else if (TAG_CURRENT_FP.equals(name)) {
					mCurrentFatiguePoints = reader.readText();
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
					mWill = reader.readInteger(0);
				} else if (BonusAttributeType.PERCEPTION.getXMLTag().equals(name)) {
					mPerception = reader.readInteger(0);
				} else if (BonusAttributeType.SPEED.getXMLTag().equals(name)) {
					mSpeed = reader.readDouble(0.0);
				} else if (BonusAttributeType.MOVE.getXMLTag().equals(name)) {
					mMove = reader.readInteger(0);
				} else if (TAG_INCLUDE_PUNCH.equals(name)) {
					mIncludePunch = reader.readBoolean();
				} else if (TAG_INCLUDE_KICK.equals(name)) {
					mIncludeKick = reader.readBoolean();
				} else if (TAG_INCLUDE_BOOTS.equals(name)) {
					mIncludeKickBoots = reader.readBoolean();
				} else if (AdvantageList.TAG_ROOT.equals(name)) {
					loadAdvantageList(reader, state);
				} else if (SkillList.TAG_ROOT.equals(name)) {
					loadSkillList(reader, state);
				} else if (SpellList.TAG_ROOT.equals(name)) {
					loadSpellList(reader, state);
				} else if (EquipmentList.TAG_ROOT.equals(name)) {
					loadEquipmentList(reader, state);
				} else if (NoteList.TAG_ROOT.equals(name)) {
					loadNoteList(reader, state);
				} else if (PrintManager.TAG_ROOT.equals(name)) {
					if (mPageSettings != null) {
						mPageSettings.load(reader);
					}
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));

		calculateAll();
		if (unspentPoints != 0) {
			setEarnedPoints(unspentPoints);
		}
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
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	private void loadEquipmentList(XMLReader reader, LoadState state) throws IOException {
		String marker = reader.getMarker();
		state.mDefaultCarried = state.mDataFileVersion != 0 ? true : reader.isAttributeSet(ATTRIBUTE_CARRIED);
		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String name = reader.getName();
				if (Equipment.TAG_EQUIPMENT.equals(name) || Equipment.TAG_EQUIPMENT_CONTAINER.equals(name)) {
					mEquipment.addRow(new Equipment(this, reader, state), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
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
	protected void saveSelf(XMLWriter out) {
		out.simpleTag(TAG_CREATED_DATE, DateFormat.getDateInstance(DateFormat.MEDIUM).format(new Date(mCreatedOn)));
		out.simpleTag(TAG_MODIFIED_DATE, DateFormat.getDateTimeInstance(DateFormat.MEDIUM, DateFormat.SHORT).format(new Date(mLastModified)));
		mDescription.save(out);
		out.simpleTag(BonusAttributeType.HP.getXMLTag(), mHitPoints);
		out.simpleTagNotEmpty(TAG_CURRENT_HP, mCurrentHitPoints);
		out.simpleTag(BonusAttributeType.FP.getXMLTag(), mFatiguePoints);
		out.simpleTagNotEmpty(TAG_CURRENT_FP, mCurrentFatiguePoints);
		out.simpleTag(TAG_TOTAL_POINTS, mTotalPoints);
		out.simpleTag(BonusAttributeType.ST.getXMLTag(), mStrength);
		out.simpleTag(BonusAttributeType.DX.getXMLTag(), mDexterity);
		out.simpleTag(BonusAttributeType.IQ.getXMLTag(), mIntelligence);
		out.simpleTag(BonusAttributeType.HT.getXMLTag(), mHealth);
		out.simpleTag(BonusAttributeType.WILL.getXMLTag(), mWill);
		out.simpleTag(BonusAttributeType.PERCEPTION.getXMLTag(), mPerception);
		out.simpleTag(BonusAttributeType.SPEED.getXMLTag(), mSpeed);
		out.simpleTag(BonusAttributeType.MOVE.getXMLTag(), mMove);
		out.simpleTag(TAG_INCLUDE_PUNCH, mIncludePunch);
		out.simpleTag(TAG_INCLUDE_KICK, mIncludeKick);
		out.simpleTag(TAG_INCLUDE_BOOTS, mIncludeKickBoots);

		saveList(AdvantageList.TAG_ROOT, mAdvantages, out);
		saveList(SkillList.TAG_ROOT, mSkills, out);
		saveList(SpellList.TAG_ROOT, mSpells, out);
		saveList(EquipmentList.TAG_ROOT, mEquipment, out);
		saveList(NoteList.TAG_ROOT, mNotes, out);

		if (mPageSettings != null) {
			mPageSettings.save(out, LengthUnits.IN);
		}
	}

	private static void saveList(String tag, OutlineModel model, XMLWriter out) {
		if (model.getRowCount() > 0) {
			out.startSimpleTagEOL(tag);
			for (ListRow row : new FilteredIterator<>(model.getTopLevelRows(), ListRow.class)) {
				row.save(out, false);
			}
			out.endTagEOL(tag, true);
		}
	}

	/**
	 * @param id The field ID to retrieve the data for.
	 * @return The value of the specified field ID, or <code>null</code> if the field ID is invalid.
	 */
	public Object getValueForID(String id) {
		if (id == null) {
			return null;
		}
		if (id.startsWith(POINTS_PREFIX)) {
			id = id.substring(POINTS_PREFIX.length());
			if (ID_STRENGTH.equals(id)) {
				return new Integer(getStrengthPoints());
			} else if (ID_DEXTERITY.equals(id)) {
				return new Integer(getDexterityPoints());
			} else if (ID_INTELLIGENCE.equals(id)) {
				return new Integer(getIntelligencePoints());
			} else if (ID_HEALTH.equals(id)) {
				return new Integer(getHealthPoints());
			} else if (ID_WILL.equals(id)) {
				return new Integer(getWillPoints());
			} else if (ID_PERCEPTION.equals(id)) {
				return new Integer(getPerceptionPoints());
			} else if (ID_BASIC_SPEED.equals(id)) {
				return new Integer(getBasicSpeedPoints());
			} else if (ID_BASIC_MOVE.equals(id)) {
				return new Integer(getBasicMovePoints());
			} else if (ID_FATIGUE_POINTS.equals(id)) {
				return new Integer(getFatiguePointPoints());
			} else if (ID_HIT_POINTS.equals(id)) {
				return new Integer(getHitPointPoints());
			}
			return null;
		} else if (ID_LAST_MODIFIED.equals(id)) {
			return getLastModified();
		} else if (ID_CREATED_ON.equals(id)) {
			return new Long(getCreatedOn());
		} else if (ID_STRENGTH.equals(id)) {
			return new Integer(getStrength());
		} else if (ID_DEXTERITY.equals(id)) {
			return new Integer(getDexterity());
		} else if (ID_INTELLIGENCE.equals(id)) {
			return new Integer(getIntelligence());
		} else if (ID_HEALTH.equals(id)) {
			return new Integer(getHealth());
		} else if (ID_BASIC_SPEED.equals(id)) {
			return new Double(getBasicSpeed());
		} else if (ID_BASIC_MOVE.equals(id)) {
			return new Integer(getBasicMove());
		} else if (ID_BASIC_LIFT.equals(id)) {
			return getBasicLift();
		} else if (ID_PERCEPTION.equals(id)) {
			return new Integer(getPerception());
		} else if (ID_VISION.equals(id)) {
			return new Integer(getVision());
		} else if (ID_HEARING.equals(id)) {
			return new Integer(getHearing());
		} else if (ID_TASTE_AND_SMELL.equals(id)) {
			return new Integer(getTasteAndSmell());
		} else if (ID_TOUCH.equals(id)) {
			return new Integer(getTouch());
		} else if (ID_WILL.equals(id)) {
			return new Integer(getWill());
		} else if (ID_FRIGHT_CHECK.equals(id)) {
			return new Integer(getFrightCheck());
		} else if (ID_ATTRIBUTE_POINTS.equals(id)) {
			return new Integer(getAttributePoints());
		} else if (ID_ADVANTAGE_POINTS.equals(id)) {
			return new Integer(getAdvantagePoints());
		} else if (ID_DISADVANTAGE_POINTS.equals(id)) {
			return new Integer(getDisadvantagePoints());
		} else if (ID_QUIRK_POINTS.equals(id)) {
			return new Integer(getQuirkPoints());
		} else if (ID_SKILL_POINTS.equals(id)) {
			return new Integer(getSkillPoints());
		} else if (ID_SPELL_POINTS.equals(id)) {
			return new Integer(getSpellPoints());
		} else if (ID_RACE_POINTS.equals(id)) {
			return new Integer(getRacePoints());
		} else if (ID_EARNED_POINTS.equals(id)) {
			return new Integer(getEarnedPoints());
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
			return new Integer(getTotalPoints());
		} else if (ID_BASIC_THRUST.equals(id)) {
			return getThrust();
		} else if (ID_BASIC_SWING.equals(id)) {
			return getSwing();
		} else if (ID_HIT_POINTS.equals(id)) {
			return new Integer(getHitPoints());
		} else if (ID_CURRENT_HIT_POINTS.equals(id)) {
			return getCurrentHitPoints();
		} else if (ID_REELING_HIT_POINTS.equals(id)) {
			return new Integer(getReelingHitPoints());
		} else if (ID_UNCONSCIOUS_CHECKS_HIT_POINTS.equals(id)) {
			return new Integer(getUnconsciousChecksHitPoints());
		} else if (ID_DEATH_CHECK_1_HIT_POINTS.equals(id)) {
			return new Integer(getDeathCheck1HitPoints());
		} else if (ID_DEATH_CHECK_2_HIT_POINTS.equals(id)) {
			return new Integer(getDeathCheck2HitPoints());
		} else if (ID_DEATH_CHECK_3_HIT_POINTS.equals(id)) {
			return new Integer(getDeathCheck3HitPoints());
		} else if (ID_DEATH_CHECK_4_HIT_POINTS.equals(id)) {
			return new Integer(getDeathCheck4HitPoints());
		} else if (ID_DEAD_HIT_POINTS.equals(id)) {
			return new Integer(getDeadHitPoints());
		} else if (ID_FATIGUE_POINTS.equals(id)) {
			return new Integer(getFatiguePoints());
		} else if (ID_CURRENT_FATIGUE_POINTS.equals(id)) {
			return getCurrentFatiguePoints();
		} else if (ID_TIRED_FATIGUE_POINTS.equals(id)) {
			return new Integer(getTiredFatiguePoints());
		} else if (ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS.equals(id)) {
			return new Integer(getUnconsciousChecksFatiguePoints());
		} else if (ID_UNCONSCIOUS_FATIGUE_POINTS.equals(id)) {
			return new Integer(getUnconsciousFatiguePoints());
		} else if (ID_PARRY_BONUS.equals(id)) {
			return new Integer(getParryBonus());
		} else if (ID_BLOCK_BONUS.equals(id)) {
			return new Integer(getBlockBonus());
		} else if (ID_DODGE_BONUS.equals(id)) {
			return new Integer(getDodgeBonus());
		} else if (id.startsWith(Profile.PROFILE_PREFIX)) {
			return mDescription.getValueForID(id);
		} else if (id.startsWith(Armor.DR_PREFIX)) {
			return mArmor.getValueForID(id);
		} else {
			for (Encumbrance encumbrance : Encumbrance.values()) {
				int index = encumbrance.ordinal();
				if ((DODGE_PREFIX + index).equals(id)) {
					return new Integer(getDodge(encumbrance));
				}
				if ((MOVE_PREFIX + index).equals(id)) {
					return new Integer(getMove(encumbrance));
				}
				if ((MAXIMUM_CARRY_PREFIX + index).equals(id)) {
					return getMaximumCarry(encumbrance);
				}
			}
			return null;
		}
	}

	/**
	 * @param id The field ID to set the value for.
	 * @param value The value to set.
	 */
	public void setValueForID(String id, Object value) {
		if (id != null) {
			if (ID_CREATED_ON.equals(id)) {
				if (value instanceof Long) {
					setCreatedOn(((Long) value).longValue());
				} else {
					setCreatedOn((String) value);
				}
			} else if (ID_INCLUDE_PUNCH.equals(id)) {
				setIncludePunch(((Boolean) value).booleanValue());
			} else if (ID_INCLUDE_KICK.equals(id)) {
				setIncludeKick(((Boolean) value).booleanValue());
			} else if (ID_INCLUDE_BOOTS.equals(id)) {
				setIncludeKickBoots(((Boolean) value).booleanValue());
			} else if (ID_STRENGTH.equals(id)) {
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
				setPerception(((Integer) value).intValue());
			} else if (ID_WILL.equals(id)) {
				setWill(((Integer) value).intValue());
			} else if (ID_EARNED_POINTS.equals(id)) {
				setEarnedPoints(((Integer) value).intValue());
			} else if (ID_HIT_POINTS.equals(id)) {
				setHitPoints(((Integer) value).intValue());
			} else if (ID_CURRENT_HIT_POINTS.equals(id)) {
				setCurrentHitPoints((String) value);
			} else if (ID_FATIGUE_POINTS.equals(id)) {
				setFatiguePoints(((Integer) value).intValue());
			} else if (ID_CURRENT_FATIGUE_POINTS.equals(id)) {
				setCurrentFatiguePoints((String) value);
			} else if (id.startsWith(Profile.PROFILE_PREFIX)) {
				mDescription.setValueForID(id, value);
			} else if (id.startsWith(Armor.DR_PREFIX)) {
				mArmor.setValueForID(id, value);
			} else {
				Log.error(String.format(UNABLE_TO_SET_VALUE, id));
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
		if (Advantage.ID_POINTS.equals(type) || Advantage.ID_ROUND_COST_DOWN.equals(type) || Advantage.ID_LEVELS.equals(type) || Advantage.ID_CONTAINER_TYPE.equals(type) || Advantage.ID_LIST_CHANGED.equals(type) || Advantage.ID_CR.equals(type) || Modifier.ID_LIST_CHANGED.equals(type) || Modifier.ID_ENABLED.equals(type)) {
			mNeedAdvantagesPointCalculation = true;
		}
		if (Skill.ID_POINTS.equals(type) || Skill.ID_LIST_CHANGED.equals(type)) {
			mNeedSkillPointCalculation = true;
		}
		if (Spell.ID_POINTS.equals(type) || Spell.ID_LIST_CHANGED.equals(type)) {
			mNeedSpellPointCalculation = true;
		}
		if (Equipment.ID_QUANTITY.equals(type) || Equipment.ID_WEIGHT.equals(type) || Equipment.ID_EXTENDED_WEIGHT.equals(type) || Equipment.ID_LIST_CHANGED.equals(type)) {
			mNeedEquipmentCalculation = true;
		}
		if (Profile.ID_SIZE_MODIFIER.equals(type)) {
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
			notify(ID_ATTRIBUTE_POINTS, new Integer(getAttributePoints()));
		}
		if (mNeedAdvantagesPointCalculation) {
			calculateAdvantagePoints();
			notify(ID_ADVANTAGE_POINTS, new Integer(getAdvantagePoints()));
			notify(ID_DISADVANTAGE_POINTS, new Integer(getDisadvantagePoints()));
			notify(ID_QUIRK_POINTS, new Integer(getQuirkPoints()));
			notify(ID_RACE_POINTS, new Integer(getRacePoints()));
		}
		if (mNeedSkillPointCalculation) {
			calculateSkillPoints();
			notify(ID_SKILL_POINTS, new Integer(getSkillPoints()));
		}
		if (mNeedSpellPointCalculation) {
			calculateSpellPoints();
			notify(ID_SPELL_POINTS, new Integer(getSpellPoints()));
		}
		if (mNeedAttributePointCalculation || mNeedAdvantagesPointCalculation || mNeedSkillPointCalculation || mNeedSpellPointCalculation) {
			notify(ID_EARNED_POINTS, new Integer(getEarnedPoints()));
		}
		if (mNeedEquipmentCalculation) {
			calculateWeightAndWealthCarried(true);
		}
		if (mDidModify) {
			long now = System.currentTimeMillis();

			if (mLastModified != now) {
				mLastModified = now;
				notify(ID_LAST_MODIFIED, new Long(mLastModified));
			}
		}
	}

	/** @return The last modified date and time. */
	public String getLastModified() {
		Date date = new Date(mLastModified);
		return MessageFormat.format(LAST_MODIFIED, DateFormat.getTimeInstance(DateFormat.SHORT).format(date), DateFormat.getDateInstance(DateFormat.MEDIUM).format(date));
	}

	/** @return The created on date. */
	public long getCreatedOn() {
		return mCreatedOn;
	}

	/**
	 * Sets the created on date.
	 *
	 * @param date The new created on date.
	 */
	public void setCreatedOn(long date) {
		if (mCreatedOn != date) {
			Long value = new Long(date);
			postUndoEdit(CREATED_ON_UNDO, ID_CREATED_ON, new Long(mCreatedOn), value);
			mCreatedOn = date;
			notifySingle(ID_CREATED_ON, value);
		}
	}

	/**
	 * Sets the created on date.
	 *
	 * @param date The new created on date.
	 */
	public void setCreatedOn(String date) {
		setCreatedOn(Numbers.extractDate(date));
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
			postUndoEdit(STRENGTH_UNDO, ID_STRENGTH, new Integer(oldStrength), new Integer(strength));
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
		Dice thrust = getThrust();
		Dice swing = getSwing();
		WeightValue lift = getBasicLift();
		boolean notifyST = mStrength != strength || mStrengthBonus != bonus;
		Dice dice;

		mStrength = strength;
		mStrengthBonus = bonus;
		mLiftingStrengthBonus = liftingBonus;
		mStrikingStrengthBonus = strikingBonus;

		startNotify();
		if (notifyST) {
			notify(ID_STRENGTH, new Integer(getStrength()));
			notifyOfBaseHitPointChange();
		}
		WeightValue newLift = getBasicLift();
		if (!newLift.equals(lift)) {
			notify(ID_BASIC_LIFT, newLift);
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

	/** @return The number of points spent on strength. */
	public int getStrengthPoints() {
		return getPointsForAttribute(mStrength - 10, 10, mStrengthCostReduction + mDescription.getSizeModifier() * 10);
	}

	private static int getPointsForAttribute(int delta, int ptsPerLevel, int reduction) {
		int amt = delta * ptsPerLevel;

		if (reduction > 0 && delta > 0) {
			int rounder = delta < 0 ? -99 : 99;

			if (reduction > 80) {
				reduction = 80;
			}
			amt = (rounder + amt * (100 - reduction)) / 100;
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
	public static Dice getThrust(int strength) {
		if (SheetPreferences.areOptionalStrengthRulesUsed()) {
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
	public static Dice getSwing(int strength) {
		if (SheetPreferences.areOptionalStrengthRulesUsed()) {
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
		return getBasicLift(SheetPreferences.getWeightUnits());
	}

	private WeightValue getBasicLift(WeightUnits desiredUnits) {
		WeightUnits units;
		double divisor;
		double multiplier;
		double roundAt;
		if (SheetPreferences.areGurpsMetricRulesUsed() && SheetPreferences.getWeightUnits().isMetric()) {
			units = WeightUnits.KG;
			divisor = 10;
			multiplier = 1;
			roundAt = 5;
		} else {
			units = WeightUnits.LB;
			divisor = 5;
			multiplier = 2;
			roundAt = 10;
		}
		int strength = getStrength() + mLiftingStrengthBonus;
		double value;
		if (strength < 1) {
			value = 0;
		} else {
			if (SheetPreferences.areOptionalStrengthRulesUsed()) {
				int diff = 0;
				if (strength > 19) {
					diff = strength / 10 - 1;
					strength -= diff * 10;
				}
				value = Math.pow(10.0, strength / 10.0) * multiplier;
				if (strength <= 6) {
					value = Math.round(value * 10.0) / 10.0;
				} else {
					value = Math.round(value);
				}
				value *= Math.pow(10, diff);
			} else {
				value = strength * strength / divisor;
			}
			if (value >= roundAt) {
				value = Math.round(value);
			}
			value = Math.floor(value * 10.0) / 10.0;
		}
		return new WeightValue(desiredUnits.convert(units, value), desiredUnits);
	}

	private WeightValue getMultipleOfBasicLift(double multiple) {
		WeightValue lift = getBasicLift();
		lift.setValue(lift.getValue() * multiple);
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
		WeightUnits calcUnits = SheetPreferences.areGurpsMetricRulesUsed() && SheetPreferences.getWeightUnits().isMetric() ? WeightUnits.KG : WeightUnits.LB;
		WeightValue lift = getBasicLift(calcUnits);
		lift.setValue(lift.getValue() * encumbrance.getWeightMultiplier());
		WeightUnits desiredUnits = SheetPreferences.getWeightUnits();
		return new WeightValue(desiredUnits.convert(calcUnits, lift.getValue()), desiredUnits);
	}

	/**
	 * @return The character's basic speed.
	 */
	public double getBasicSpeed() {
		return mSpeed + mSpeedBonus + getRawBasicSpeed();
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
			postUndoEdit(BASIC_SPEED_UNDO, ID_BASIC_SPEED, new Double(oldBasicSpeed), new Double(speed));
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
			updateBasicSpeedInfo(mSpeed, bonus);
		}
	}

	private void updateBasicSpeedInfo(double speed, double bonus) {
		int move = getBasicMove();
		int[] data = preserveMoveAndDodge();
		int tmp;

		mSpeed = speed;
		mSpeedBonus = bonus;

		startNotify();
		notify(ID_BASIC_SPEED, new Double(getBasicSpeed()));
		tmp = getBasicMove();
		if (move != tmp) {
			notify(ID_BASIC_MOVE, new Integer(tmp));
		}
		notifyIfMoveOrDodgeAltered(data);
		mNeedAttributePointCalculation = true;
		endNotify();
	}

	/** @return The number of points spent on basic speed. */
	public int getBasicSpeedPoints() {
		return (int) (mSpeed * 20.0);
	}

	/**
	 * @return The character's basic move.
	 */
	public int getBasicMove() {
		int move = mMove + mMoveBonus + getRawBasicMove();

		return move < 0 ? 0 : move;
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
			postUndoEdit(BASIC_MOVE_UNDO, ID_BASIC_MOVE, new Integer(oldBasicMove), new Integer(move));
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
			updateBasicMoveInfo(mMove, bonus);
		}
	}

	private void updateBasicMoveInfo(int move, int bonus) {
		int[] data = preserveMoveAndDodge();

		startNotify();
		mMove = move;
		mMoveBonus = bonus;
		notify(ID_BASIC_MOVE, new Integer(getBasicMove()));
		notifyIfMoveOrDodgeAltered(data);
		mNeedAttributePointCalculation = true;
		endNotify();
	}

	/** @return The number of points spent on basic move. */
	public int getBasicMovePoints() {
		return mMove * 5;
	}

	/**
	 * @param encumbrance The encumbrance level.
	 * @return The character's ground move for the specified encumbrance level.
	 */
	public int getMove(Encumbrance encumbrance) {
		int basicMove = getBasicMove();
		int move = basicMove * (10 + 2 * encumbrance.getEncumbrancePenalty()) / 10;
		if (move < 1) {
			return basicMove > 0 ? 1 : 0;
		}
		return move;
	}

	/**
	 * @param encumbrance The encumbrance level.
	 * @return The character's dodge for the specified encumbrance level.
	 */
	public int getDodge(Encumbrance encumbrance) {
		int dodge = (int) Math.floor(getBasicSpeed()) + 3 + encumbrance.getEncumbrancePenalty() + mDodgeBonus;
		return dodge < 1 ? 1 : dodge;
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
			notifySingle(ID_DODGE_BONUS, new Integer(mDodgeBonus));
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
			notifySingle(ID_PARRY_BONUS, new Integer(mParryBonus));
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
			notifySingle(ID_BLOCK_BONUS, new Integer(mBlockBonus));
		}
	}

	/** @return The current encumbrance level. */
	public Encumbrance getEncumbranceLevel() {
		double carried = getWeightCarried().getNormalizedValue();
		for (Encumbrance encumbrance : Encumbrance.values()) {
			if (carried <= getMaximumCarry(encumbrance).getNormalizedValue()) {
				return encumbrance;
			}
		}
		return Encumbrance.EXTRA_HEAVY;
	}

	/**
	 * @return <code>true</code> if the carried weight is greater than the maximum allowed for an
	 *         extra-heavy load.
	 */
	public boolean isCarryingGreaterThanMaxLoad() {
		return getWeightCarried().getNormalizedValue() > getMaximumCarry(Encumbrance.EXTRA_HEAVY).getNormalizedValue();
	}

	/** @return The current weight being carried. */
	public WeightValue getWeightCarried() {
		return mCachedWeightCarried;
	}

	/** @return The current wealth being carried. */
	public double getWealthCarried() {
		return mCachedWealthCarried;
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
				return new WeightValue(value.getValue() / 30, WeightUnits.OZ);
			case KG:
				return new WeightValue(value.getValue() * 2, WeightUnits.LB);
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
				return new WeightValue(value.getValue() / 2, WeightUnits.KG);
			case LT:
				return new WeightValue(value.getValue(), WeightUnits.T);
			case OZ:
				return new WeightValue(value.getValue() * 30, WeightUnits.G);
			case TN:
				return new WeightValue(value.getValue(), WeightUnits.T);
			default:
				return value;
		}
	}

	/**
	 * Calculate the total weight and wealth carried.
	 *
	 * @param notify Whether to send out notifications if the resulting values are different from
	 *            the previous values.
	 */
	public void calculateWeightAndWealthCarried(boolean notify) {
		WeightValue savedWeight = new WeightValue(mCachedWeightCarried);
		double savedWealth = mCachedWealthCarried;
		mCachedWeightCarried = new WeightValue(0, SheetPreferences.getWeightUnits());
		mCachedWealthCarried = 0.0;
		for (Row one : mEquipment.getTopLevelRows()) {
			Equipment equipment = (Equipment) one;
			if (equipment.isCarried()) {
				WeightValue weight = new WeightValue(equipment.getExtendedWeight());
				if (SheetPreferences.areGurpsMetricRulesUsed()) {
					if (SheetPreferences.getWeightUnits().isMetric()) {
						weight = GURPSCharacter.convertToGurpsMetric(weight);
					} else {
						weight = GURPSCharacter.convertFromGurpsMetric(weight);
					}
				}
				mCachedWeightCarried.add(weight);
			}
			mCachedWealthCarried += equipment.getExtendedValue();
		}
		if (notify) {
			if (!savedWeight.equals(mCachedWeightCarried)) {
				notify(ID_CARRIED_WEIGHT, mCachedWeightCarried);
			}
			if (savedWealth != mCachedWealthCarried) {
				notify(ID_CARRIED_WEALTH, new Double(mCachedWealthCarried));
			}
		}
	}

	private int[] preserveMoveAndDodge() {
		Encumbrance[] values = Encumbrance.values();
		int[] data = new int[values.length * 2];
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
			int tmp = getDodge(encumbrance);
			if (tmp != data[values.length + index]) {
				notify(DODGE_PREFIX + index, new Integer(tmp));
			}
			tmp = getMove(encumbrance);
			if (tmp != data[index]) {
				notify(MOVE_PREFIX + index, new Integer(tmp));
			}
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
			postUndoEdit(DEXTERITY_UNDO, ID_DEXTERITY, new Integer(oldDexterity), new Integer(dexterity));
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
		int move = getBasicMove();
		int[] data = preserveMoveAndDodge();
		double newSpeed;
		int newMove;

		mDexterity = dexterity;
		mDexterityBonus = bonus;

		startNotify();
		notify(ID_DEXTERITY, new Integer(getDexterity()));
		newSpeed = getBasicSpeed();
		if (newSpeed != speed) {
			notify(ID_BASIC_SPEED, new Double(newSpeed));
		}
		newMove = getBasicMove();
		if (newMove != move) {
			notify(ID_BASIC_MOVE, new Integer(newMove));
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
			postUndoEdit(INTELLIGENCE_UNDO, ID_INTELLIGENCE, new Integer(oldIntelligence), new Integer(intelligence));
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
		int perception = getPerception();
		int will = getWill();
		int newPerception;
		int newWill;

		mIntelligence = intelligence;
		mIntelligenceBonus = bonus;

		startNotify();
		notify(ID_INTELLIGENCE, new Integer(getIntelligence()));
		newPerception = getPerception();
		if (newPerception != perception) {
			notify(ID_PERCEPTION, new Integer(newPerception));
			notify(ID_VISION, new Integer(getVision()));
			notify(ID_HEARING, new Integer(getHearing()));
			notify(ID_TASTE_AND_SMELL, new Integer(getTasteAndSmell()));
			notify(ID_TOUCH, new Integer(getTouch()));
		}
		newWill = getWill();
		if (newWill != will) {
			notify(ID_WILL, new Integer(newWill));
			notify(ID_FRIGHT_CHECK, new Integer(getFrightCheck()));
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
			postUndoEdit(HEALTH_UNDO, ID_HEALTH, new Integer(oldHealth), new Integer(health));
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
		int move = getBasicMove();
		int[] data = preserveMoveAndDodge();
		double newSpeed;
		int tmp;

		mHealth = health;
		mHealthBonus = bonus;

		startNotify();
		notify(ID_HEALTH, new Integer(getHealth()));

		newSpeed = getBasicSpeed();
		if (newSpeed != speed) {
			notify(ID_BASIC_SPEED, new Double(newSpeed));
		}

		tmp = getBasicMove();
		if (tmp != move) {
			notify(ID_BASIC_MOVE, new Integer(tmp));
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

	/** @return The number of earned points. */
	public int getEarnedPoints() {
		return mTotalPoints - getSpentPoints();
	}

	/**
	 * Sets the earned character points.
	 *
	 * @param earned The new earned character points.
	 */
	public void setEarnedPoints(int earned) {
		int current = getEarnedPoints();

		if (current != earned) {
			Integer value = new Integer(earned);

			postUndoEdit(EARNED_POINTS_UNDO, ID_EARNED_POINTS, new Integer(current), value);
			mTotalPoints = earned + getSpentPoints();
			startNotify();
			notify(ID_EARNED_POINTS, value);
			notify(ID_TOTAL_POINTS, new Integer(getTotalPoints()));
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

	/** @return Whether to include the punch natural weapon or not. */
	public boolean includePunch() {
		return mIncludePunch;
	}

	/** @param include Whether to include the punch natural weapon or not. */
	public void setIncludePunch(boolean include) {
		if (mIncludePunch != include) {
			postUndoEdit(INCLUDE_PUNCH_UNDO, ID_INCLUDE_PUNCH, new Boolean(mIncludePunch), new Boolean(include));
			mIncludePunch = include;
			notifySingle(ID_INCLUDE_PUNCH, new Boolean(mIncludePunch));
		}
	}

	/** @return Whether to include the kick natural weapon or not. */
	public boolean includeKick() {
		return mIncludeKick;
	}

	/** @param include Whether to include the kick natural weapon or not. */
	public void setIncludeKick(boolean include) {
		if (mIncludeKick != include) {
			postUndoEdit(INCLUDE_KICK_UNDO, ID_INCLUDE_KICK, new Boolean(mIncludeKick), new Boolean(include));
			mIncludeKick = include;
			notifySingle(ID_INCLUDE_KICK, new Boolean(mIncludeKick));
		}
	}

	/** @return Whether to include the kick w/boots natural weapon or not. */
	public boolean includeKickBoots() {
		return mIncludeKickBoots;
	}

	/** @param include Whether to include the kick w/boots natural weapon or not. */
	public void setIncludeKickBoots(boolean include) {
		if (mIncludeKickBoots != include) {
			postUndoEdit(INCLUDE_BOOTS_UNDO, ID_INCLUDE_BOOTS, new Boolean(mIncludeKickBoots), new Boolean(include));
			mIncludeKickBoots = include;
			notifySingle(ID_INCLUDE_BOOTS, new Boolean(mIncludeKickBoots));
		}
	}

	/** @return The hit points (HP). */
	public int getHitPoints() {
		return getStrength() + mHitPoints + mHitPointBonus;
	}

	/**
	 * Sets the hit points (HP).
	 *
	 * @param hp The new hit points.
	 */
	public void setHitPoints(int hp) {
		int oldHP = getHitPoints();

		if (oldHP != hp) {
			postUndoEdit(HIT_POINTS_UNDO, ID_HIT_POINTS, new Integer(oldHP), new Integer(hp));
			startNotify();
			mHitPoints = hp - (getStrength() + mHitPointBonus);
			mNeedAttributePointCalculation = true;
			notifyOfBaseHitPointChange();
			endNotify();
		}
	}

	/** @return The number of points spent on hit points. */
	public int getHitPointPoints() {
		int pts = 2 * mHitPoints;
		int sizeModifier = mDescription.getSizeModifier();

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
		notify(ID_HIT_POINTS, new Integer(getHitPoints()));
		notify(ID_DEATH_CHECK_1_HIT_POINTS, new Integer(getDeathCheck1HitPoints()));
		notify(ID_DEATH_CHECK_2_HIT_POINTS, new Integer(getDeathCheck2HitPoints()));
		notify(ID_DEATH_CHECK_3_HIT_POINTS, new Integer(getDeathCheck3HitPoints()));
		notify(ID_DEATH_CHECK_4_HIT_POINTS, new Integer(getDeathCheck4HitPoints()));
		notify(ID_DEAD_HIT_POINTS, new Integer(getDeadHitPoints()));
		notify(ID_REELING_HIT_POINTS, new Integer(getReelingHitPoints()));
		endNotify();
	}

	/** @return The current hit points. */
	public String getCurrentHitPoints() {
		return mCurrentHitPoints;
	}

	/**
	 * Sets the current hit points.
	 *
	 * @param hp The hit point amount.
	 */
	public void setCurrentHitPoints(String hp) {
		if (!mCurrentHitPoints.equals(hp)) {
			postUndoEdit(CURRENT_HIT_POINTS_UNDO, ID_CURRENT_HIT_POINTS, mCurrentHitPoints, hp);
			mCurrentHitPoints = hp;
			notifySingle(ID_CURRENT_HIT_POINTS, mCurrentHitPoints);
		}
	}

	/** @return The number of hit points where "reeling" effects start. */
	public int getReelingHitPoints() {
		int reeling = (getHitPoints() - 1) / 3;

		return reeling < 0 ? 0 : reeling;
	}

	/** @return The number of hit points where unconsciousness checks must start being made. */
	@SuppressWarnings("static-method")
	public int getUnconsciousChecksHitPoints() {
		return 0;
	}

	/** @return The number of hit points where the first death check must be made. */
	public int getDeathCheck1HitPoints() {
		return -1 * getHitPoints();
	}

	/** @return The number of hit points where the second death check must be made. */
	public int getDeathCheck2HitPoints() {
		return -2 * getHitPoints();
	}

	/** @return The number of hit points where the third death check must be made. */
	public int getDeathCheck3HitPoints() {
		return -3 * getHitPoints();
	}

	/** @return The number of hit points where the fourth death check must be made. */
	public int getDeathCheck4HitPoints() {
		return -4 * getHitPoints();
	}

	/** @return The number of hit points where the character is just dead. */
	public int getDeadHitPoints() {
		return -5 * getHitPoints();
	}

	/** @return The will. */
	public int getWill() {
		return mWill + mWillBonus + (SheetPreferences.areOptionalIQRulesUsed() ? 10 : getIntelligence());
	}

	/** @param will The new will. */
	public void setWill(int will) {
		int oldWill = getWill();
		if (oldWill != will) {
			postUndoEdit(WILL_UNDO, ID_WILL, new Integer(oldWill), new Integer(will));
			updateWillInfo(will - (mWillBonus + (SheetPreferences.areOptionalIQRulesUsed() ? 10 : getIntelligence())), mWillBonus);
		}
	}

	/** @return The will bonus. */
	public int getWillBonus() {
		return mWillBonus;
	}

	/** @param bonus The new will bonus. */
	public void setWillBonus(int bonus) {
		if (mWillBonus != bonus) {
			updateWillInfo(mWill, bonus);
		}
	}

	private void updateWillInfo(int will, int bonus) {
		mWill = will;
		mWillBonus = bonus;

		startNotify();
		notify(ID_WILL, new Integer(getWill()));
		notify(ID_FRIGHT_CHECK, new Integer(getFrightCheck()));
		updateSkills();
		mNeedAttributePointCalculation = true;
		endNotify();
	}

	/** Called to ensure notifications are sent out when the optional IQ rule use is changed. */
	public void updateWillAndPerceptionDueToOptionalIQRuleUseChange() {
		updateWillInfo(mWill, mWillBonus);
		updatePerceptionInfo(mPerception, mPerceptionBonus);
	}

	/** @return The number of points spent on will. */
	public int getWillPoints() {
		return mWill * 5;
	}

	/** @return The fright check. */
	public int getFrightCheck() {
		return getWill() + mFrightCheckBonus;
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
			notify(ID_FRIGHT_CHECK, new Integer(getFrightCheck()));
			endNotify();
		}
	}

	/** @return The vision. */
	public int getVision() {
		return getPerception() + mVisionBonus;
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
			notify(ID_VISION, new Integer(getVision()));
			endNotify();
		}
	}

	/** @return The hearing. */
	public int getHearing() {
		return getPerception() + mHearingBonus;
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
			notify(ID_HEARING, new Integer(getHearing()));
			endNotify();
		}
	}

	/** @return The touch perception. */
	public int getTouch() {
		return getPerception() + mTouchBonus;
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
			notify(ID_TOUCH, new Integer(getTouch()));
			endNotify();
		}
	}

	/** @return The taste and smell perception. */
	public int getTasteAndSmell() {
		return getPerception() + mTasteAndSmellBonus;
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
			notify(ID_TASTE_AND_SMELL, new Integer(getTasteAndSmell()));
			endNotify();
		}
	}

	/** @return The perception (Per). */
	public int getPerception() {
		return mPerception + mPerceptionBonus + (SheetPreferences.areOptionalIQRulesUsed() ? 10 : getIntelligence());
	}

	/**
	 * Sets the perception.
	 *
	 * @param perception The new perception.
	 */
	public void setPerception(int perception) {
		int oldPerception = getPerception();
		if (oldPerception != perception) {
			postUndoEdit(PERCEPTION_UNDO, ID_PERCEPTION, new Integer(oldPerception), new Integer(perception));
			updatePerceptionInfo(perception - (mPerceptionBonus + (SheetPreferences.areOptionalIQRulesUsed() ? 10 : getIntelligence())), mPerceptionBonus);
		}
	}

	/** @return The perception bonus. */
	public int getPerceptionBonus() {
		return mPerceptionBonus;
	}

	/** @param bonus The new perception bonus. */
	public void setPerceptionBonus(int bonus) {
		if (mPerceptionBonus != bonus) {
			updatePerceptionInfo(mPerception, bonus);
		}
	}

	private void updatePerceptionInfo(int perception, int bonus) {
		mPerception = perception;
		mPerceptionBonus = bonus;

		startNotify();
		notify(ID_PERCEPTION, new Integer(getPerception()));
		notify(ID_VISION, new Integer(getVision()));
		notify(ID_HEARING, new Integer(getHearing()));
		notify(ID_TASTE_AND_SMELL, new Integer(getTasteAndSmell()));
		notify(ID_TOUCH, new Integer(getTouch()));
		updateSkills();
		mNeedAttributePointCalculation = true;
		endNotify();
	}

	/** @return The number of points spent on perception. */
	public int getPerceptionPoints() {
		return mPerception * 5;
	}

	/** @return The fatigue points (FP). */
	public int getFatiguePoints() {
		return getHealth() + mFatiguePoints + mFatiguePointBonus;
	}

	/**
	 * Sets the fatigue points (HP).
	 *
	 * @param fp The new fatigue points.
	 */
	public void setFatiguePoints(int fp) {
		int oldFP = getFatiguePoints();

		if (oldFP != fp) {
			postUndoEdit(FATIGUE_POINTS_UNDO, ID_FATIGUE_POINTS, new Integer(oldFP), new Integer(fp));
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
		notify(ID_FATIGUE_POINTS, new Integer(getFatiguePoints()));
		notify(ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, new Integer(getUnconsciousChecksFatiguePoints()));
		notify(ID_UNCONSCIOUS_FATIGUE_POINTS, new Integer(getUnconsciousFatiguePoints()));
		notify(ID_TIRED_FATIGUE_POINTS, new Integer(getTiredFatiguePoints()));
		endNotify();
	}

	/** @return The current fatigue points. */
	public String getCurrentFatiguePoints() {
		return mCurrentFatiguePoints;
	}

	/**
	 * Sets the current fatigue points.
	 *
	 * @param fp The fatigue point amount.
	 */
	public void setCurrentFatiguePoints(String fp) {
		if (!mCurrentFatiguePoints.equals(fp)) {
			postUndoEdit(CURRENT_FATIGUE_POINTS_UNDO, ID_CURRENT_FATIGUE_POINTS, mCurrentFatiguePoints, fp);
			mCurrentFatiguePoints = fp;
			notifySingle(ID_CURRENT_FATIGUE_POINTS, mCurrentFatiguePoints);
		}
	}

	/** @return The number of fatigue points where "tired" effects start. */
	public int getTiredFatiguePoints() {
		int tired = (getFatiguePoints() - 1) / 3;

		return tired < 0 ? 0 : tired;
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
	public Profile getDescription() {
		return mDescription;
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
	 * @param includeDisabled <code>true</code> if disabled entries should be included.
	 * @return A recursive iterator over the character's advantages.
	 */
	public RowIterator<Advantage> getAdvantagesIterator(boolean includeDisabled) {
		if (includeDisabled) {
			return new RowIterator<>(mAdvantages);
		}
		return new RowIterator<>(mAdvantages, (row) -> {
			return row.isEnabled();
		});
	}

	/**
	 * Searches the character's current advantages list for the specified name.
	 *
	 * @param name The name to look for.
	 * @return The advantage, if present, or <code>null</code>.
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
	 * @param name The name to look for.
	 * @param specialization The specialization to look for. Pass in <code>null</code> or an empty
	 *            string to ignore.
	 * @param requirePoints Only look at {@link Skill}s that have points.
	 * @param excludes The set of {@link Skill}s to exclude from consideration.
	 * @return The skill if it is present, or <code>null</code> if its not.
	 */
	public ArrayList<Skill> getSkillNamed(String name, String specialization, boolean requirePoints, HashSet<String> excludes) {
		ArrayList<Skill> skills = new ArrayList<>();
		boolean checkSpecialization = specialization != null && specialization.length() > 0;
		for (Skill skill : getSkillsIterator()) {
			if (excludes == null || !excludes.contains(skill.toString())) {
				if (!requirePoints || skill.getPoints() > 0) {
					if (skill.getName().equalsIgnoreCase(name)) {
						if (!checkSpecialization || skill.getSpecialization().equalsIgnoreCase(specialization)) {
							skills.add(skill);
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
	 * @param name The {@link Skill} name to look for.
	 * @param specialization An optional specialization to look for. Pass <code>null</code> if it is
	 *            not needed.
	 * @param requirePoints Only look at {@link Skill}s that have points.
	 * @param excludes The set of {@link Skill}s to exclude from consideration.
	 * @return The {@link Skill} that matches with the highest level.
	 */
	public Skill getBestSkillNamed(String name, String specialization, boolean requirePoints, HashSet<String> excludes) {
		Skill best = null;
		int level = Integer.MIN_VALUE;

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

	/** @return The outline model for the character's notes. */
	public OutlineModel getNotesRoot() {
		return mNotes;
	}

	/** @return A recursive iterator over the character's notes. */
	public RowIterator<Note> getNoteIterator() {
		return new RowIterator<>(mNotes);
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
		mDescription.update();
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
		int total = 0;
		ArrayList<Feature> list = mFeatureMap.get(id.toLowerCase());

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
		int total = 0;
		ArrayList<Feature> list = mFeatureMap.get(id.toLowerCase());
		if (list != null) {
			for (Feature feature : list) {
				if (feature instanceof Bonus && !(feature instanceof WeaponBonus)) {
					total += ((Bonus) feature).getAmount().getIntegerAdjustedAmount();
				}
			}
		}
		return total;
	}

	/**
	 * @param id The feature ID to search for.
	 * @param nameQualifier The name qualifier.
	 * @param specializationQualifier The specialization qualifier.
	 * @return The bonuses.
	 */
	public ArrayList<WeaponBonus> getWeaponComparedBonusesFor(String id, String nameQualifier, String specializationQualifier) {
		ArrayList<WeaponBonus> bonuses = new ArrayList<>();
		int rsl = Integer.MIN_VALUE;

		for (Skill skill : getSkillNamed(nameQualifier, specializationQualifier, true, null)) {
			int srsl = skill.getRelativeLevel();

			if (srsl > rsl) {
				rsl = srsl;
			}
		}

		if (rsl != Integer.MIN_VALUE) {
			ArrayList<Feature> list = mFeatureMap.get(id.toLowerCase());
			if (list != null) {
				for (Feature feature : list) {
					if (feature instanceof WeaponBonus) {
						WeaponBonus bonus = (WeaponBonus) feature;
						if (bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(specializationQualifier) && bonus.getLevelCriteria().matches(rsl)) {
							bonuses.add(bonus);
						}
					}
				}
			}
		}
		return bonuses;
	}

	/**
	 * @param id The feature ID to search for.
	 * @param nameQualifier The name qualifier.
	 * @param specializationQualifier The specialization qualifier.
	 * @return The bonus.
	 */
	public int getSkillComparedIntegerBonusFor(String id, String nameQualifier, String specializationQualifier) {
		int total = 0;
		ArrayList<Feature> list = mFeatureMap.get(id.toLowerCase());
		if (list != null) {
			for (Feature feature : list) {
				if (feature instanceof SkillBonus) {
					SkillBonus bonus = (SkillBonus) feature;
					if (bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(specializationQualifier)) {
						total += bonus.getAmount().getIntegerAdjustedAmount();
					}
				}
			}
		}
		return total;
	}

	/**
	 * @param id The feature ID to search for.
	 * @param qualifier The qualifier.
	 * @return The bonus.
	 */
	public int getSpellComparedIntegerBonusFor(String id, String qualifier) {
		int total = 0;
		ArrayList<Feature> list = mFeatureMap.get(id.toLowerCase());
		if (list != null) {
			for (Feature feature : list) {
				if (feature instanceof SpellBonus) {
					SpellBonus bonus = (SpellBonus) feature;
					if (bonus.getNameCriteria().matches(qualifier)) {
						total += bonus.getAmount().getIntegerAdjustedAmount();
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
		double total = 0;
		ArrayList<Feature> list = mFeatureMap.get(id.toLowerCase());
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
	 * @param name The name of the undo.
	 * @param id The ID of the field being changed.
	 * @param before The original value.
	 * @param after The new value.
	 */
	void postUndoEdit(String name, String id, Object before, Object after) {
		StdUndoManager mgr = getUndoManager();
		if (!mgr.isInTransaction()) {
			boolean add;
			if (before instanceof ListRow) {
				add = !((ListRow) before).isEquivalentTo(after);
			} else {
				add = !before.equals(after);
			}
			if (add) {
				addEdit(new CharacterFieldUndo(this, name, id, before, after));
			}
		}
	}
}
