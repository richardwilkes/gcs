/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.model;

import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.advantage.CMAdvantageContainerType;
import com.trollworks.gcs.model.advantage.CMAdvantageList;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.equipment.CMEquipmentList;
import com.trollworks.gcs.model.feature.CMAttributeBonusLimitation;
import com.trollworks.gcs.model.feature.CMBonus;
import com.trollworks.gcs.model.feature.CMBonusAttributeType;
import com.trollworks.gcs.model.feature.CMCostReduction;
import com.trollworks.gcs.model.feature.CMFeature;
import com.trollworks.gcs.model.feature.CMHitLocation;
import com.trollworks.gcs.model.feature.CMLeveledAmount;
import com.trollworks.gcs.model.feature.CMSkillBonus;
import com.trollworks.gcs.model.feature.CMSpellBonus;
import com.trollworks.gcs.model.feature.CMWeaponBonus;
import com.trollworks.gcs.model.modifier.CMModifier;
import com.trollworks.gcs.model.names.USCensusNames;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMSkillList;
import com.trollworks.gcs.model.skill.CMTechnique;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.model.spell.CMSpellList;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.toolkit.collections.TKEnumExtractor;
import com.trollworks.toolkit.collections.TKFilteredIterator;
import com.trollworks.toolkit.io.TKBase64;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.print.TKPageOrientation;
import com.trollworks.toolkit.print.TKPrintManager;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.utility.units.TKLengthUnits;
import com.trollworks.toolkit.utility.units.TKWeightUnits;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.outline.TKRowIterator;

import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.IOException;
import java.text.DateFormat;
import java.text.MessageFormat;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.Random;

import javax.imageio.ImageIO;

/** A GURPS character. */
public class CMCharacter extends CMDataFile {
	private static final String						EMPTY									= "";																																	//$NON-NLS-1$
	private static final String						TAG_ROOT								= "character";																															//$NON-NLS-1$
	private static final String						TAG_PLAYER_NAME							= "player_name";																														//$NON-NLS-1$
	private static final String						TAG_CAMPAIGN							= "campaign";																															//$NON-NLS-1$
	private static final String						TAG_NAME								= "name";																																//$NON-NLS-1$
	private static final String						TAG_TITLE								= "title";																																//$NON-NLS-1$
	private static final String						TAG_AGE									= "age";																																//$NON-NLS-1$
	private static final String						TAG_BIRTHDAY							= "birthday";																															//$NON-NLS-1$
	private static final String						TAG_EYES								= "eyes";																																//$NON-NLS-1$
	private static final String						TAG_HAIR								= "hair";																																//$NON-NLS-1$
	private static final String						TAG_SKIN								= "skin";																																//$NON-NLS-1$
	private static final String						TAG_HANDEDNESS							= "handedness";																														//$NON-NLS-1$
	private static final String						TAG_HEIGHT								= "height";																															//$NON-NLS-1$
	private static final String						ATTRIBUTE_UNITS							= "units";																																//$NON-NLS-1$
	private static final String						TAG_WEIGHT								= "weight";																															//$NON-NLS-1$
	private static final String						TAG_GENDER								= "gender";																															//$NON-NLS-1$
	private static final String						TAG_RACE								= "race";																																//$NON-NLS-1$
	private static final String						TAG_TECH_LEVEL							= "tech_level";																														//$NON-NLS-1$
	private static final String						TAG_RELIGION							= "religion";																															//$NON-NLS-1$
	private static final String						TAG_CREATED_DATE						= "created_date";																														//$NON-NLS-1$
	private static final String						TAG_MODIFIED_DATE						= "modified_date";																														//$NON-NLS-1$
	private static final String						TAG_CURRENT_HP							= "current_hp";																														//$NON-NLS-1$
	private static final String						TAG_CURRENT_FP							= "current_fp";																														//$NON-NLS-1$
	private static final String						TAG_UNSPENT_POINTS						= "unspent_points";																													//$NON-NLS-1$
	private static final String						TAG_TOTAL_POINTS						= "total_points";																														//$NON-NLS-1$
	private static final String						TAG_INCLUDE_PUNCH						= "include_punch";																														//$NON-NLS-1$
	private static final String						TAG_INCLUDE_KICK						= "include_kick";																														//$NON-NLS-1$
	private static final String						TAG_INCLUDE_BOOTS						= "include_kick_with_boots";																											//$NON-NLS-1$
	private static final String						TAG_PORTRAIT							= "portrait";																															//$NON-NLS-1$
	private static final String						TAG_NOTES								= "notes";																																//$NON-NLS-1$
	private static final String						ATTRIBUTE_CARRIED						= "carried";																															//$NON-NLS-1$
	/** The prefix for all character IDs. */
	public static final String						CHARACTER_PREFIX						= "gcs.";																																//$NON-NLS-1$
	/** The default portrait marker. */
	public static final String						DEFAULT_PORTRAIT						= "!\000";																																//$NON-NLS-1$
	/** The default Tech Level. */
	public static final String						DEFAULT_TECH_LEVEL						= "4";																																	//$NON-NLS-1$
	/** The preferences module name. */
	public static final String						MODULE									= "GURPSCharacter";																													//$NON-NLS-1$
	/** The field ID for last modified date changes. */
	public static final String						ID_LAST_MODIFIED						= CHARACTER_PREFIX + "LastModifiedDate";																								//$NON-NLS-1$
	/** The field ID for created on date changes. */
	public static final String						ID_CREATED_ON							= CHARACTER_PREFIX + "CreatedOn";																										//$NON-NLS-1$
	/** The field ID for include punch changes. */
	public static final String						ID_INCLUDE_PUNCH						= CHARACTER_PREFIX + "IncludePunch";																									//$NON-NLS-1$
	/** The field ID for include kick changes. */
	public static final String						ID_INCLUDE_KICK							= CHARACTER_PREFIX + "IncludeKickFeet";																								//$NON-NLS-1$
	/** The field ID for include kick with boots changes. */
	public static final String						ID_INCLUDE_BOOTS						= CHARACTER_PREFIX + "IncludeKickBoots";																								//$NON-NLS-1$
	/**
	 * The prefix used to indicate a point value is requested from {@link #getValueForID(String)}.
	 */
	public static final String						POINTS_PREFIX							= CHARACTER_PREFIX + "points.";																										//$NON-NLS-1$
	/** The prefix used in front of all IDs for basic attributes. */
	public static final String						ATTRIBUTES_PREFIX						= CHARACTER_PREFIX + "ba.";																											//$NON-NLS-1$
	/** The field ID for strength (ST) changes. */
	public static final String						ID_STRENGTH								= ATTRIBUTES_PREFIX + CMBonusAttributeType.ST.name();
	/** The field ID for lifting strength bonuses -- used by features. */
	public static final String						ID_LIFTING_STRENGTH						= ID_STRENGTH + CMAttributeBonusLimitation.LIFTING_ONLY.name();
	/** The field ID for striking strength bonuses -- used by features. */
	public static final String						ID_STRIKING_STRENGTH					= ID_STRENGTH + CMAttributeBonusLimitation.STRIKING_ONLY.name();
	/** The field ID for dexterity (DX) changes. */
	public static final String						ID_DEXTERITY							= ATTRIBUTES_PREFIX + CMBonusAttributeType.DX.name();
	/** The field ID for intelligence (IQ) changes. */
	public static final String						ID_INTELLIGENCE							= ATTRIBUTES_PREFIX + CMBonusAttributeType.IQ.name();
	/** The field ID for health (HT) changes. */
	public static final String						ID_HEALTH								= ATTRIBUTES_PREFIX + CMBonusAttributeType.HT.name();
	/** The field ID for perception changes. */
	public static final String						ID_PERCEPTION							= ATTRIBUTES_PREFIX + CMBonusAttributeType.PERCEPTION.name();
	/** The field ID for vision changes. */
	public static final String						ID_VISION								= ATTRIBUTES_PREFIX + CMBonusAttributeType.VISION.name();
	/** The field ID for hearing changes. */
	public static final String						ID_HEARING								= ATTRIBUTES_PREFIX + CMBonusAttributeType.HEARING.name();
	/** The field ID for taste changes. */
	public static final String						ID_TASTE_AND_SMELL						= ATTRIBUTES_PREFIX + CMBonusAttributeType.TASTE_SMELL.name();
	/** The field ID for smell changes. */
	public static final String						ID_TOUCH								= ATTRIBUTES_PREFIX + CMBonusAttributeType.TOUCH.name();
	/** The field ID for will changes. */
	public static final String						ID_WILL									= ATTRIBUTES_PREFIX + CMBonusAttributeType.WILL.name();
	/** The field ID for basic speed changes. */
	public static final String						ID_BASIC_SPEED							= ATTRIBUTES_PREFIX + CMBonusAttributeType.SPEED.name();
	/** The field ID for basic move changes. */
	public static final String						ID_BASIC_MOVE							= ATTRIBUTES_PREFIX + CMBonusAttributeType.MOVE.name();
	/** The prefix used in front of all IDs for dodge changes. */
	public static final String						DODGE_PREFIX							= ATTRIBUTES_PREFIX + CMBonusAttributeType.DODGE.name() + "#.";																		//$NON-NLS-1$
	/** The field ID for dodge bonus changes. */
	public static final String						ID_DODGE_BONUS							= ATTRIBUTES_PREFIX + CMBonusAttributeType.DODGE.name();
	/** The field ID for parry bonus changes. */
	public static final String						ID_PARRY_BONUS							= ATTRIBUTES_PREFIX + CMBonusAttributeType.PARRY.name();
	/** The field ID for block bonus changes. */
	public static final String						ID_BLOCK_BONUS							= ATTRIBUTES_PREFIX + CMBonusAttributeType.BLOCK.name();
	/** The prefix used in front of all IDs for move changes. */
	public static final String						MOVE_PREFIX								= ATTRIBUTES_PREFIX + CMBonusAttributeType.MOVE.name() + "#.";																			//$NON-NLS-1$
	/** The field ID for carried weight changes. */
	public static final String						ID_CARRIED_WEIGHT						= CHARACTER_PREFIX + "CarriedWeight";																									//$NON-NLS-1$
	/** The field ID for carried wealth changes. */
	public static final String						ID_CARRIED_WEALTH						= CHARACTER_PREFIX + "CarriedWealth";																									//$NON-NLS-1$
	/** The prefix used in front of all IDs for encumbrance changes. */
	public static final String						MAXIMUM_CARRY_PREFIX					= ATTRIBUTES_PREFIX + "MaximumCarry";																									//$NON-NLS-1$
	private static final String						LIFT_PREFIX								= ATTRIBUTES_PREFIX + "lift.";																											//$NON-NLS-1$
	/** The field ID for basic lift changes. */
	public static final String						ID_BASIC_LIFT							= LIFT_PREFIX + "BasicLift";																											//$NON-NLS-1$
	/** The field ID for one-handed lift changes. */
	public static final String						ID_ONE_HANDED_LIFT						= LIFT_PREFIX + ".OneHandedLift";																										//$NON-NLS-1$
	/** The field ID for two-handed lift changes. */
	public static final String						ID_TWO_HANDED_LIFT						= LIFT_PREFIX + ".TwoHandedLift";																										//$NON-NLS-1$
	/** The field ID for shove and knock over changes. */
	public static final String						ID_SHOVE_AND_KNOCK_OVER					= LIFT_PREFIX + ".ShoveAndKnockOver";																									//$NON-NLS-1$
	/** The field ID for running shove and knock over changes. */
	public static final String						ID_RUNNING_SHOVE_AND_KNOCK_OVER			= LIFT_PREFIX + ".RunningShoveAndKnockOver";																							//$NON-NLS-1$
	/** The field ID for carry on back changes. */
	public static final String						ID_CARRY_ON_BACK						= LIFT_PREFIX + ".CarryOnBack";																										//$NON-NLS-1$
	/** The field ID for carry on back changes. */
	public static final String						ID_SHIFT_SLIGHTLY						= LIFT_PREFIX + ".ShiftSlightly";																										//$NON-NLS-1$
	/** The prefix used in front of all IDs for profile. */
	public static final String						PROFILE_PREFIX							= CHARACTER_PREFIX + "pi.";																											//$NON-NLS-1$
	/** The field ID for portrait changes. */
	public static final String						ID_PORTRAIT								= PROFILE_PREFIX + "Portrait";																											//$NON-NLS-1$
	/** The field ID for name changes. */
	public static final String						ID_NAME									= PROFILE_PREFIX + "Name";																												//$NON-NLS-1$
	/** The field ID for notes changes. */
	public static final String						ID_NOTES								= PROFILE_PREFIX + "Notes";																											//$NON-NLS-1$
	/** The field ID for title changes. */
	public static final String						ID_TITLE								= PROFILE_PREFIX + "Title";																											//$NON-NLS-1$
	/** The field ID for age changes. */
	public static final String						ID_AGE									= PROFILE_PREFIX + "Age";																												//$NON-NLS-1$
	/** The field ID for birthday changes. */
	public static final String						ID_BIRTHDAY								= PROFILE_PREFIX + "Birthday";																											//$NON-NLS-1$
	/** The field ID for eye color changes. */
	public static final String						ID_EYE_COLOR							= PROFILE_PREFIX + "EyeColor";																											//$NON-NLS-1$
	/** The field ID for hair color changes. */
	public static final String						ID_HAIR									= PROFILE_PREFIX + "Hair";																												//$NON-NLS-1$
	/** The field ID for skin color changes. */
	public static final String						ID_SKIN_COLOR							= PROFILE_PREFIX + "SkinColor";																										//$NON-NLS-1$
	/** The field ID for handedness changes. */
	public static final String						ID_HANDEDNESS							= PROFILE_PREFIX + "Handedness";																										//$NON-NLS-1$
	/** The field ID for height changes. */
	public static final String						ID_HEIGHT								= PROFILE_PREFIX + "Height";																											//$NON-NLS-1$
	/** The field ID for weight changes. */
	public static final String						ID_WEIGHT								= PROFILE_PREFIX + "Weight";																											//$NON-NLS-1$
	/** The field ID for gender changes. */
	public static final String						ID_GENDER								= PROFILE_PREFIX + "Gender";																											//$NON-NLS-1$
	/** The field ID for race changes. */
	public static final String						ID_RACE									= PROFILE_PREFIX + "Race";																												//$NON-NLS-1$
	/** The field ID for religion changes. */
	public static final String						ID_RELIGION								= PROFILE_PREFIX + "Religion";																											//$NON-NLS-1$
	/** The field ID for player name changes. */
	public static final String						ID_PLAYER_NAME							= PROFILE_PREFIX + "PlayerName";																										//$NON-NLS-1$
	/** The field ID for campaign changes. */
	public static final String						ID_CAMPAIGN								= PROFILE_PREFIX + "Campaign";																											//$NON-NLS-1$
	/** The field ID for size modifier changes. */
	public static final String						ID_SIZE_MODIFIER						= ATTRIBUTES_PREFIX + CMBonusAttributeType.SM.name();
	/** The field ID for tech level changes. */
	public static final String						ID_TECH_LEVEL							= PROFILE_PREFIX + "TechLevel";																										//$NON-NLS-1$
	/** The prefix used in front of all IDs for point summaries. */
	public static final String						POINT_SUMMARY_PREFIX					= CHARACTER_PREFIX + "ps.";																											//$NON-NLS-1$
	/** The field ID for point total changes. */
	public static final String						ID_TOTAL_POINTS							= POINT_SUMMARY_PREFIX + "TotalPoints";																								//$NON-NLS-1$
	/** The field ID for attribute point summary changes. */
	public static final String						ID_ATTRIBUTE_POINTS						= POINT_SUMMARY_PREFIX + "AttributePoints";																							//$NON-NLS-1$
	/** The field ID for advantage point summary changes. */
	public static final String						ID_ADVANTAGE_POINTS						= POINT_SUMMARY_PREFIX + "AdvantagePoints";																							//$NON-NLS-1$
	/** The field ID for disadvantage point summary changes. */
	public static final String						ID_DISADVANTAGE_POINTS					= POINT_SUMMARY_PREFIX + "DisadvantagePoints";																							//$NON-NLS-1$
	/** The field ID for quirk point summary changes. */
	public static final String						ID_QUIRK_POINTS							= POINT_SUMMARY_PREFIX + "QuirkPoints";																								//$NON-NLS-1$
	/** The field ID for skill point summary changes. */
	public static final String						ID_SKILL_POINTS							= POINT_SUMMARY_PREFIX + "SkillPoints";																								//$NON-NLS-1$
	/** The field ID for spell point summary changes. */
	public static final String						ID_SPELL_POINTS							= POINT_SUMMARY_PREFIX + "SpellPoints";																								//$NON-NLS-1$
	/** The field ID for racial point summary changes. */
	public static final String						ID_RACE_POINTS							= POINT_SUMMARY_PREFIX + "RacePoints";																									//$NON-NLS-1$
	/** The field ID for earned point changes. */
	public static final String						ID_EARNED_POINTS						= POINT_SUMMARY_PREFIX + "EarnedPoints";																								//$NON-NLS-1$
	/** The prefix used in front of all IDs for damage resistance. */
	public static final String						DR_PREFIX								= CHARACTER_PREFIX + "dr.";																											//$NON-NLS-1$
	/** The skull hit location's DR. */
	public static final String						ID_SKULL_DR								= DR_PREFIX + CMHitLocation.SKULL.name();
	/** The eyes hit location's DR. */
	public static final String						ID_EYES_DR								= DR_PREFIX + CMHitLocation.EYES.name();
	/** The face hit location's DR. */
	public static final String						ID_FACE_DR								= DR_PREFIX + CMHitLocation.FACE.name();
	/** The neck hit location's DR. */
	public static final String						ID_NECK_DR								= DR_PREFIX + CMHitLocation.NECK.name();
	/** The torso hit location's DR. */
	public static final String						ID_TORSO_DR								= DR_PREFIX + CMHitLocation.TORSO.name();
	/** The vitals hit location's DR. */
	public static final String						ID_VITALS_DR							= DR_PREFIX + CMHitLocation.VITALS.name();
	private static final String						ID_FULL_BODY_DR							= DR_PREFIX + CMHitLocation.FULL_BODY.name();
	private static final String						ID_FULL_BODY_EXCEPT_EYES_DR				= DR_PREFIX + CMHitLocation.FULL_BODY_EXCEPT_EYES.name();
	/** The groin hit location's DR. */
	public static final String						ID_GROIN_DR								= DR_PREFIX + CMHitLocation.GROIN.name();
	/** The arm hit location's DR. */
	public static final String						ID_ARM_DR								= DR_PREFIX + CMHitLocation.ARMS.name();
	/** The hand hit location's DR. */
	public static final String						ID_HAND_DR								= DR_PREFIX + CMHitLocation.HANDS.name();
	/** The leg hit location's DR. */
	public static final String						ID_LEG_DR								= DR_PREFIX + CMHitLocation.LEGS.name();
	/** The foot hit location's DR. */
	public static final String						ID_FOOT_DR								= DR_PREFIX + CMHitLocation.FEET.name();
	/** The prefix used in front of all IDs for basic damage. */
	public static final String						BASIC_DAMAGE_PREFIX						= CHARACTER_PREFIX + "bd.";																											//$NON-NLS-1$
	/** The field ID for basic thrust damage changes. */
	public static final String						ID_BASIC_THRUST							= BASIC_DAMAGE_PREFIX + "Thrust";																										//$NON-NLS-1$
	/** The field ID for basic swing damage changes. */
	public static final String						ID_BASIC_SWING							= BASIC_DAMAGE_PREFIX + "Swing";																										//$NON-NLS-1$
	private static final String						HIT_POINTS_PREFIX						= ATTRIBUTES_PREFIX + "derived_hp.";																									//$NON-NLS-1$
	/** The field ID for hit point changes. */
	public static final String						ID_HIT_POINTS							= ATTRIBUTES_PREFIX + CMBonusAttributeType.HP.name();
	/** The field ID for current hit point changes. */
	public static final String						ID_CURRENT_HIT_POINTS					= HIT_POINTS_PREFIX + "Current";																										//$NON-NLS-1$
	/** The field ID for reeling hit point changes. */
	public static final String						ID_REELING_HIT_POINTS					= HIT_POINTS_PREFIX + "Reeling";																										//$NON-NLS-1$
	/** The field ID for unconscious check hit point changes. */
	public static final String						ID_UNCONSCIOUS_CHECKS_HIT_POINTS		= HIT_POINTS_PREFIX + "UnconsciousChecks";																								//$NON-NLS-1$
	/** The field ID for death check #1 hit point changes. */
	public static final String						ID_DEATH_CHECK_1_HIT_POINTS				= HIT_POINTS_PREFIX + "DeathCheck1";																									//$NON-NLS-1$
	/** The field ID for death check #2 hit point changes. */
	public static final String						ID_DEATH_CHECK_2_HIT_POINTS				= HIT_POINTS_PREFIX + "DeathCheck2";																									//$NON-NLS-1$
	/** The field ID for death check #3 hit point changes. */
	public static final String						ID_DEATH_CHECK_3_HIT_POINTS				= HIT_POINTS_PREFIX + "DeathCheck3";																									//$NON-NLS-1$
	/** The field ID for death check #4 hit point changes. */
	public static final String						ID_DEATH_CHECK_4_HIT_POINTS				= HIT_POINTS_PREFIX + "DeathCheck4";																									//$NON-NLS-1$
	/** The field ID for dead hit point changes. */
	public static final String						ID_DEAD_HIT_POINTS						= HIT_POINTS_PREFIX + "Dead";																											//$NON-NLS-1$
	private static final String						FATIGUE_POINTS_PREFIX					= ATTRIBUTES_PREFIX + "derived_fp.";																									//$NON-NLS-1$
	/** The field ID for fatigue point changes. */
	public static final String						ID_FATIGUE_POINTS						= ATTRIBUTES_PREFIX + CMBonusAttributeType.FP.name();
	/** The field ID for current fatigue point changes. */
	public static final String						ID_CURRENT_FATIGUE_POINTS				= FATIGUE_POINTS_PREFIX + "Current";																									//$NON-NLS-1$
	/** The field ID for tired fatigue point changes. */
	public static final String						ID_TIRED_FATIGUE_POINTS					= FATIGUE_POINTS_PREFIX + "Tired";																										//$NON-NLS-1$
	/** The field ID for unconscious check fatigue point changes. */
	public static final String						ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS	= FATIGUE_POINTS_PREFIX + "UnconsciousChecks";																							//$NON-NLS-1$
	/** The field ID for unconscious fatigue point changes. */
	public static final String						ID_UNCONSCIOUS_FATIGUE_POINTS			= FATIGUE_POINTS_PREFIX + "Unconscious";																								//$NON-NLS-1$
	/** The encumbrance level for "None". */
	public static final int							ENCUMBRANCE_NONE						= 0;
	/** The encumbrance level for "Light". */
	public static final int							ENCUMBRANCE_LIGHT						= 1;
	/** The encumbrance level for "Medium". */
	public static final int							ENCUMBRANCE_MEDIUM						= 2;
	/** The encumbrance level for "Heavy". */
	public static final int							ENCUMBRANCE_HEAVY						= 3;
	/** The encumbrance level for "Extra Heavy". */
	public static final int							ENCUMBRANCE_EXTRA_HEAVY					= 4;
	/** The number of difference encumbrance levels. */
	public static final int							ENCUMBRANCE_LEVELS						= 5;
	/** The height, in 1/72nds of an inch, of the portrait. */
	public static final int							PORTRAIT_HEIGHT							= 96;
	/** The width, in 1/72nds of an inch, of the portrait. */
	public static final int							PORTRAIT_WIDTH							= 3 * PORTRAIT_HEIGHT / 4;
	private static final double[]					ENCUMBRANCE_MULTIPLIER					= { 1.0, 2.0, 3.0, 6.0, 10.0 };
	private static final Random						RANDOM									= new Random();
	private static final String[]					EYE_OPTIONS								= new String[] { Msgs.BROWN, Msgs.BROWN, Msgs.BLUE, Msgs.BLUE, Msgs.GREEN, Msgs.GREY, Msgs.VIOLET };
	private static final String[]					SKIN_OPTIONS							= new String[] { Msgs.FRECKLED, Msgs.TAN, Msgs.LIGHT_TAN, Msgs.DARK_TAN, Msgs.BROWN, Msgs.LIGHT_BROWN, Msgs.DARK_BROWN, Msgs.PALE };
	private static final String[]					HANDEDNESS_OPTIONS						= new String[] { Msgs.RIGHT, Msgs.RIGHT, Msgs.RIGHT, Msgs.LEFT };
	private static final String[]					GENDER_OPTIONS							= new String[] { Msgs.MALE, Msgs.MALE, Msgs.MALE, Msgs.FEMALE };
	private static final String[]					HAIR_OPTIONS;
	private long									mLastModified;
	private long									mCreatedOn;
	private HashMap<String, ArrayList<CMFeature>>	mFeatureMap;
	private int										mStrength;
	private int										mStrengthBonus;
	private int										mLiftingStrengthBonus;
	private int										mStrikingStrengthBonus;
	private int										mStrengthCostReduction;
	private int										mDexterity;
	private int										mDexterityBonus;
	private int										mDexterityCostReduction;
	private int										mIntelligence;
	private int										mIntelligenceBonus;
	private int										mIntelligenceCostReduction;
	private int										mHealth;
	private int										mHealthBonus;
	private int										mHealthCostReduction;
	private int										mWill;
	private int										mWillBonus;
	private int										mPerception;
	private int										mPerceptionBonus;
	private int										mVisionBonus;
	private int										mHearingBonus;
	private int										mTasteAndSmellBonus;
	private int										mTouchBonus;
	private String									mCurrentHitPoints;
	private int										mHitPoints;
	private int										mHitPointBonus;
	private int										mFatiguePoints;
	private String									mCurrentFatiguePoints;
	private int										mFatiguePointBonus;
	private double									mSpeed;
	private double									mSpeedBonus;
	private int										mMove;
	private int										mMoveBonus;
	private int										mDodgeBonus;
	private int										mParryBonus;
	private int										mBlockBonus;
	private boolean									mCustomPortrait;
	private BufferedImage							mPortrait;
	private BufferedImage							mDisplayPortrait;
	private String									mName;
	private String									mNotes;
	private String									mTitle;
	private int										mAge;
	private String									mBirthday;
	private String									mEyeColor;
	private String									mHair;
	private String									mSkinColor;
	private String									mHandedness;
	private int										mHeight;
	private double									mWeight;
	private int										mSizeModifier;
	private int										mSizeModifierBonus;
	private String									mGender;
	private String									mRace;
	private String									mReligion;
	private String									mPlayerName;
	private String									mCampaign;
	private String									mTechLevel;
	private int										mTotalPoints;
	private int										mSkullDR;
	private int										mEyesDR;
	private int										mFaceDR;
	private int										mNeckDR;
	private int										mTorsoDR;
	private int										mVitalsDR;
	private int										mGroinDR;
	private int										mArmDR;
	private int										mHandDR;
	private int										mLegDR;
	private int										mFootDR;
	private TKOutlineModel							mAdvantages;
	private TKOutlineModel							mSkills;
	private TKOutlineModel							mSpells;
	private TKOutlineModel							mCarriedEquipment;
	private TKOutlineModel							mOtherEquipment;
	private boolean									mDidModify;
	private boolean									mNeedAttributePointCalculation;
	private boolean									mNeedAdvantagesPointCalculation;
	private boolean									mNeedSkillPointCalculation;
	private boolean									mNeedSpellPointCalculation;
	private boolean									mNeedEquipmentCalculation;
	private double									mCachedWeightCarried;
	private double									mCachedWealthCarried;
	private int										mCachedAttributePoints;
	private int										mCachedAdvantagePoints;
	private int										mCachedDisadvantagePoints;
	private int										mCachedQuirkPoints;
	private int										mCachedSkillPoints;
	private int										mCachedSpellPoints;
	private int										mCachedRacePoints;
	private boolean									mSkillsUpdated;
	private boolean									mSpellsUpdated;
	private TKPrintManager							mPageSettings;
	private boolean									mIncludePunch;
	private boolean									mIncludeKick;
	private boolean									mIncludeKickBoots;

	static {
		ArrayList<String> hair = new ArrayList<String>(100);
		String[] colors = { Msgs.BROWN, Msgs.BROWN, Msgs.BROWN, Msgs.BLACK, Msgs.BLACK, Msgs.BLACK, Msgs.BLOND, Msgs.BLOND, Msgs.REDHEAD };
		String[] styles = { Msgs.STRAIGHT, Msgs.CURLY, Msgs.WAVY };
		String[] lengths = { Msgs.SHORT, Msgs.MEDIUM, Msgs.LONG };

		for (String element : colors) {
			for (String style : styles) {
				for (String length : lengths) {
					hair.add(MessageFormat.format(Msgs.HAIR_FORMAT, element, style, length));
				}
			}
		}
		hair.add(Msgs.BALD);
		HAIR_OPTIONS = hair.toArray(new String[hair.size()]);
	}

	/** @return A random hair color, style & length. */
	public static String getRandomHair() {
		return HAIR_OPTIONS[RANDOM.nextInt(HAIR_OPTIONS.length)];
	}

	/** @return A random eye color. */
	public static String getRandomEyeColor() {
		return EYE_OPTIONS[RANDOM.nextInt(EYE_OPTIONS.length)];
	}

	/** @return A random sking color. */
	public static String getRandomSkinColor() {
		return SKIN_OPTIONS[RANDOM.nextInt(SKIN_OPTIONS.length)];
	}

	/** @return A random handedness. */
	public static String getRandomHandedness() {
		return HANDEDNESS_OPTIONS[RANDOM.nextInt(HANDEDNESS_OPTIONS.length)];
	}

	/** @return A random gender. */
	public static String getRandomGender() {
		return GENDER_OPTIONS[RANDOM.nextInt(GENDER_OPTIONS.length)];
	}

	/** @return A random month and day. */
	public static String getRandomMonthAndDay() {
		SimpleDateFormat formatter = new SimpleDateFormat("MMMM d"); //$NON-NLS-1$

		return formatter.format(new Date(RANDOM.nextLong()));
	}

	/**
	 * @param strength The strength to base the height on.
	 * @return A random height, in inches.
	 */
	public static int getRandomHeight(int strength) {
		int base;

		if (strength < 7) {
			base = 52;
		} else if (strength < 10) {
			base = 55 + (strength - 7) * 3;
		} else if (strength == 10) {
			base = 63;
		} else if (strength < 14) {
			base = 65 + (strength - 11) * 3;
		} else {
			base = 74;
		}
		return base + RANDOM.nextInt(11);
	}

	/**
	 * @param strength The strength to base the weight on.
	 * @param multiplier The weight multiplier for being under- or overweight.
	 * @return A random weight, in pounds.
	 */
	public static int getRandomWeight(int strength, double multiplier) {
		int base;
		int range;

		if (strength < 7) {
			base = 60;
			range = 61;
		} else if (strength < 10) {
			base = 75 + (strength - 7) * 15;
			range = 61;
		} else if (strength == 10) {
			base = 115;
			range = 61;
		} else if (strength < 14) {
			base = 125 + (strength - 11) * 15;
			range = 71 + (strength - 11) * 10;
		} else {
			base = 170;
			range = 101;
		}
		return (int) Math.round((base + RANDOM.nextInt(range)) * multiplier);
	}

	/** @return The default player name. */
	public static String getDefaultPlayerName() {
		return TKPreferences.getInstance().getStringValue(MODULE, ID_NAME, System.getProperty("user.name")); //$NON-NLS-1$
	}

	/** @param name The default player name. */
	public static void setDefaultPlayerName(String name) {
		TKPreferences.getInstance().setValue(MODULE, ID_NAME, name);
	}

	/** @return The default campaign value. */
	public static String getDefaultCampaign() {
		return TKPreferences.getInstance().getStringValue(MODULE, ID_CAMPAIGN, EMPTY);
	}

	/** @param campaign The default campaign value. */
	public static void setDefaultCampaign(String campaign) {
		TKPreferences.getInstance().setValue(MODULE, ID_CAMPAIGN, campaign);
	}

	/** @return The default tech level. */
	public static String getDefaultTechLevel() {
		return TKPreferences.getInstance().getStringValue(MODULE, ID_TECH_LEVEL, DEFAULT_TECH_LEVEL);
	}

	/** @param techLevel The default tech level. */
	public static void setDefaultTechLevel(String techLevel) {
		TKPreferences.getInstance().setValue(MODULE, ID_TECH_LEVEL, techLevel);
	}

	/**
	 * @param path The path to load.
	 * @return The portrait.
	 */
	public static BufferedImage getPortraitFromPortraitPath(String path) {
		if (DEFAULT_PORTRAIT.equals(path)) {
			return CSImage.getDefaultPortrait();
		}
		return path != null ? TKImage.loadImage(new File(path)) : null;
	}

	/** @return The default portrait path. */
	public static String getDefaultPortraitPath() {
		return TKPreferences.getInstance().getStringValue(MODULE, ID_PORTRAIT, DEFAULT_PORTRAIT);
	}

	/** @param path The default portrait path. */
	public static void setDefaultPortraitPath(String path) {
		TKPreferences.getInstance().setValue(MODULE, ID_PORTRAIT, path);
	}

	/** Creates a new character with only default values set. */
	public CMCharacter() {
		super();
		characterInitialize(true);
		initialize();
		calculateAll();
	}

	/**
	 * Creates a new character from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @throws IOException if the data cannot be read or the file doesn't contain a valid character
	 *             sheet.
	 */
	public CMCharacter(File file) throws IOException {
		super(file);
		initialize();
	}

	private void characterInitialize(boolean full) {
		mFeatureMap = new HashMap<String, ArrayList<CMFeature>>();
		mAdvantages = new TKOutlineModel();
		mSkills = new TKOutlineModel();
		mSpells = new TKOutlineModel();
		mCarriedEquipment = new TKOutlineModel();
		mOtherEquipment = new TKOutlineModel();
		mStrength = 10;
		mDexterity = 10;
		mIntelligence = 10;
		mHealth = 10;
		mCurrentHitPoints = EMPTY;
		mCurrentFatiguePoints = EMPTY;
		mSkullDR = 2;
		mIncludePunch = true;
		mIncludeKick = true;
		mIncludeKickBoots = true;
		mCustomPortrait = false;
		mPortrait = null;
		mDisplayPortrait = null;
		mTitle = EMPTY;
		mAge = full ? getRandomAge() : 0;
		mBirthday = full ? getRandomMonthAndDay() : EMPTY;
		mEyeColor = full ? getRandomEyeColor() : EMPTY;
		mHair = full ? getRandomHair() : EMPTY;
		mSkinColor = full ? getRandomSkinColor() : EMPTY;
		mHandedness = full ? getRandomHandedness() : EMPTY;
		mHeight = full ? getRandomHeight(getStrength()) : 0;
		mWeight = full ? getRandomWeight(getStrength(), 1.0) : 0.0;
		mGender = full ? getRandomGender() : EMPTY;
		mName = full ? USCensusNames.INSTANCE.getFullName(mGender == Msgs.MALE) : EMPTY;
		mNotes = EMPTY;
		mRace = full ? Msgs.DEFAULT_RACE : EMPTY;
		mTechLevel = full ? getDefaultTechLevel() : EMPTY;
		mReligion = EMPTY;
		mPlayerName = full ? getDefaultPlayerName() : EMPTY;
		mCampaign = full ? getDefaultCampaign() : EMPTY;
		setPortraitInternal(getPortraitFromPortraitPath(getDefaultPortraitPath()));
		try {
			mPageSettings = new TKPrintManager(TKPageOrientation.PORTRAIT, 0.5, TKLengthUnits.INCHES);
		} catch (Exception exception) {
			mPageSettings = null;
		}
		mLastModified = System.currentTimeMillis();
		mCreatedOn = mLastModified;
		// This will force the long value to match the string value.
		setCreatedOn(getCreatedOn());
	}

	/** @return The page settings. May return <code>null</code> if not printer has been defined. */
	public TKPrintManager getPageSettings() {
		return mPageSettings;
	}

	@Override public BufferedImage getFileIcon(boolean large) {
		return CSImage.getCharacterSheetIcon(large);
	}

	@Override protected final void loadSelf(TKXMLReader reader, Object param) throws IOException {
		String marker = reader.getMarker();
		int unspentPoints = 0;

		characterInitialize(false);
		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (TAG_PLAYER_NAME.equals(name)) {
					mPlayerName = reader.readText();
				} else if (TAG_CAMPAIGN.equals(name)) {
					mCampaign = reader.readText();
				} else if (TAG_NAME.equals(name)) {
					mName = reader.readText();
				} else if (TAG_NOTES.equals(name)) {
					mNotes = reader.readText();
				} else if (TAG_TITLE.equals(name)) {
					mTitle = reader.readText();
				} else if (TAG_AGE.equals(name)) {
					mAge = reader.readInteger(0);
				} else if (TAG_BIRTHDAY.equals(name)) {
					mBirthday = reader.readText();
				} else if (TAG_EYES.equals(name)) {
					mEyeColor = reader.readText();
				} else if (TAG_HAIR.equals(name)) {
					mHair = reader.readText();
				} else if (TAG_SKIN.equals(name)) {
					mSkinColor = reader.readText();
				} else if (TAG_HANDEDNESS.equals(name)) {
					mHandedness = reader.readText();
				} else if (TAG_HEIGHT.equals(name)) {
					TKLengthUnits units = (TKLengthUnits) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_UNITS), TKLengthUnits.values());

					if (units == null) {
						// Old output didn't include a units attribute, as it was in a mixed
						// feet/inches format.
						mHeight = TKNumberUtils.getHeight(reader.readText());
					} else {
						mHeight = (int) TKLengthUnits.INCHES.convert(units, reader.readDouble(0));
					}
				} else if (TAG_WEIGHT.equals(name)) {
					mWeight = TKWeightUnits.POUNDS.convert((TKWeightUnits) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_UNITS), TKWeightUnits.values(), TKWeightUnits.POUNDS), reader.readDouble(0));
				} else if (CMBonusAttributeType.SM.getXMLTag().equals(name) || "size_modifier".equals(name)) { //$NON-NLS-1$
					mSizeModifier = reader.readInteger(0);
				} else if (TAG_GENDER.equals(name)) {
					mGender = reader.readText();
				} else if (TAG_RACE.equals(name)) {
					mRace = reader.readText();
				} else if (TAG_TECH_LEVEL.equals(name)) {
					mTechLevel = reader.readText();
				} else if (TAG_RELIGION.equals(name)) {
					mReligion = reader.readText();
				} else if (TAG_CREATED_DATE.equals(name)) {
					mCreatedOn = TKNumberUtils.getDate(reader.readText());
				} else if (TAG_MODIFIED_DATE.equals(name)) {
					mLastModified = TKNumberUtils.getDateTime(reader.readText());
				} else if (CMBonusAttributeType.HP.getXMLTag().equals(name)) {
					mHitPoints = reader.readInteger(0);
				} else if (TAG_CURRENT_HP.equals(name)) {
					mCurrentHitPoints = reader.readText();
				} else if (CMBonusAttributeType.FP.getXMLTag().equals(name)) {
					mFatiguePoints = reader.readInteger(0);
				} else if (TAG_CURRENT_FP.equals(name)) {
					mCurrentFatiguePoints = reader.readText();
				} else if (TAG_UNSPENT_POINTS.equals(name)) {
					unspentPoints = reader.readInteger(0);
				} else if (TAG_TOTAL_POINTS.equals(name)) {
					mTotalPoints = reader.readInteger(0);
				} else if (CMBonusAttributeType.ST.getXMLTag().equals(name)) {
					mStrength = reader.readInteger(0);
				} else if (CMBonusAttributeType.DX.getXMLTag().equals(name)) {
					mDexterity = reader.readInteger(0);
				} else if (CMBonusAttributeType.IQ.getXMLTag().equals(name)) {
					mIntelligence = reader.readInteger(0);
				} else if (CMBonusAttributeType.HT.getXMLTag().equals(name)) {
					mHealth = reader.readInteger(0);
				} else if (CMBonusAttributeType.WILL.getXMLTag().equals(name)) {
					mWill = reader.readInteger(0);
				} else if (CMBonusAttributeType.PERCEPTION.getXMLTag().equals(name)) {
					mPerception = reader.readInteger(0);
				} else if (CMBonusAttributeType.SPEED.getXMLTag().equals(name)) {
					mSpeed = reader.readDouble(0.0);
				} else if (CMBonusAttributeType.MOVE.getXMLTag().equals(name)) {
					mMove = reader.readInteger(0);
				} else if (TAG_INCLUDE_PUNCH.equals(name)) {
					mIncludePunch = reader.readBoolean();
				} else if (TAG_INCLUDE_KICK.equals(name)) {
					mIncludeKick = reader.readBoolean();
				} else if (TAG_INCLUDE_BOOTS.equals(name)) {
					mIncludeKickBoots = reader.readBoolean();
				} else if (TAG_PORTRAIT.equals(name)) {
					try {
						setPortraitInternal(TKImage.loadImage(TKBase64.decode(reader.readText())));
						mCustomPortrait = true;
					} catch (Exception imageException) {
						// Ignore
					}
				} else if (CMAdvantageList.TAG_ROOT.equals(name)) {
					loadAdvantageList(reader);
				} else if (CMSkillList.TAG_ROOT.equals(name)) {
					loadSkillList(reader);
				} else if (CMSpellList.TAG_ROOT.equals(name)) {
					loadSpellList(reader);
				} else if (CMEquipmentList.TAG_ROOT.equals(name)) {
					loadEquipmentList(reader);
				} else if (TKPrintManager.TAG_ROOT.equals(name)) {
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

	private void loadAdvantageList(TKXMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMAdvantage.TAG_ADVANTAGE.equals(name) || CMAdvantage.TAG_ADVANTAGE_CONTAINER.equals(name)) {
					mAdvantages.addRow(new CMAdvantage(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	private void loadSkillList(TKXMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMSkill.TAG_SKILL.equals(name) || CMSkill.TAG_SKILL_CONTAINER.equals(name)) {
					mSkills.addRow(new CMSkill(this, reader), true);
				} else if (CMTechnique.TAG_TECHNIQUE.equals(name)) {
					mSkills.addRow(new CMTechnique(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	private void loadSpellList(TKXMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMSpell.TAG_SPELL.equals(name) || CMSpell.TAG_SPELL_CONTAINER.equals(name)) {
					mSpells.addRow(new CMSpell(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	private void loadEquipmentList(TKXMLReader reader) throws IOException {
		String marker = reader.getMarker();
		boolean carried = reader.isAttributeSet(ATTRIBUTE_CARRIED);

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMEquipment.TAG_EQUIPMENT.equals(name) || CMEquipment.TAG_EQUIPMENT_CONTAINER.equals(name)) {
					CMEquipment equipment = new CMEquipment(this, reader);

					if (carried) {
						mCarriedEquipment.addRow(equipment, true);
					} else {
						mOtherEquipment.addRow(equipment, true);
					}
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
		calculateWeightAndWealthCarried();
	}

	@Override public String getXMLTagName() {
		return TAG_ROOT;
	}

	@Override protected void saveSelf(TKXMLWriter out) {
		Iterator<TKRow> iterator;

		out.simpleTagNotEmpty(TAG_PLAYER_NAME, mPlayerName);
		out.simpleTagNotEmpty(TAG_CAMPAIGN, mCampaign);
		out.simpleTagNotEmpty(TAG_NAME, mName);
		out.simpleTagNotEmpty(TAG_TITLE, mTitle);
		out.simpleTag(TAG_AGE, mAge);
		out.simpleTagNotEmpty(TAG_BIRTHDAY, mBirthday);
		out.simpleTagNotEmpty(TAG_EYES, mEyeColor);
		out.simpleTagNotEmpty(TAG_HAIR, mHair);
		out.simpleTagNotEmpty(TAG_SKIN, mSkinColor);
		out.simpleTagNotEmpty(TAG_HANDEDNESS, mHandedness);
		out.simpleTagWithAttribute(TAG_HEIGHT, Integer.toString(mHeight), ATTRIBUTE_UNITS, TKLengthUnits.INCHES.toString());
		out.simpleTagWithAttribute(TAG_WEIGHT, Double.toString(mWeight), ATTRIBUTE_UNITS, TKWeightUnits.POUNDS.toString());
		out.simpleTag(CMBonusAttributeType.SM.getXMLTag(), mSizeModifier);
		out.simpleTagNotEmpty(TAG_GENDER, mGender);
		out.simpleTagNotEmpty(TAG_RACE, mRace);
		out.simpleTagNotEmpty(TAG_TECH_LEVEL, mTechLevel);
		out.simpleTagNotEmpty(TAG_RELIGION, mReligion);
		out.simpleTag(TAG_CREATED_DATE, DateFormat.getDateInstance(DateFormat.MEDIUM).format(new Date(mCreatedOn)));
		out.simpleTag(TAG_MODIFIED_DATE, DateFormat.getDateTimeInstance(DateFormat.MEDIUM, DateFormat.SHORT).format(new Date(mLastModified)));
		out.simpleTag(CMBonusAttributeType.HP.getXMLTag(), mHitPoints);
		out.simpleTagNotEmpty(TAG_CURRENT_HP, mCurrentHitPoints);
		out.simpleTag(CMBonusAttributeType.FP.getXMLTag(), mFatiguePoints);
		out.simpleTagNotEmpty(TAG_CURRENT_FP, mCurrentFatiguePoints);
		out.simpleTag(TAG_TOTAL_POINTS, mTotalPoints);
		out.simpleTag(CMBonusAttributeType.ST.getXMLTag(), mStrength);
		out.simpleTag(CMBonusAttributeType.DX.getXMLTag(), mDexterity);
		out.simpleTag(CMBonusAttributeType.IQ.getXMLTag(), mIntelligence);
		out.simpleTag(CMBonusAttributeType.HT.getXMLTag(), mHealth);
		out.simpleTag(CMBonusAttributeType.WILL.getXMLTag(), mWill);
		out.simpleTag(CMBonusAttributeType.PERCEPTION.getXMLTag(), mPerception);
		out.simpleTag(CMBonusAttributeType.SPEED.getXMLTag(), mSpeed);
		out.simpleTag(CMBonusAttributeType.MOVE.getXMLTag(), mMove);
		out.simpleTag(TAG_INCLUDE_PUNCH, mIncludePunch);
		out.simpleTag(TAG_INCLUDE_KICK, mIncludeKick);
		out.simpleTag(TAG_INCLUDE_BOOTS, mIncludeKickBoots);
		out.simpleTagNotEmpty(TAG_NOTES, mNotes);
		if (mPageSettings != null) {
			mPageSettings.save(out, TKLengthUnits.INCHES);
		}

		if (mAdvantages.getRowCount() > 0) {
			out.startSimpleTagEOL(CMAdvantageList.TAG_ROOT);
			for (iterator = mAdvantages.getTopLevelRows().iterator(); iterator.hasNext();) {
				((CMAdvantage) iterator.next()).save(out, false);
			}
			out.endTagEOL(CMAdvantageList.TAG_ROOT, true);
		}

		if (mSkills.getRowCount() > 0) {
			out.startSimpleTagEOL(CMSkillList.TAG_ROOT);
			for (iterator = mSkills.getTopLevelRows().iterator(); iterator.hasNext();) {
				((CMRow) iterator.next()).save(out, false);
			}
			out.endTagEOL(CMSkillList.TAG_ROOT, true);
		}

		if (mSpells.getRowCount() > 0) {
			out.startSimpleTagEOL(CMSpellList.TAG_ROOT);
			for (iterator = mSpells.getTopLevelRows().iterator(); iterator.hasNext();) {
				((CMSpell) iterator.next()).save(out, false);
			}
			out.endTagEOL(CMSpellList.TAG_ROOT, true);
		}

		if (mCarriedEquipment.getRowCount() > 0) {
			out.startTag(CMEquipmentList.TAG_ROOT);
			out.writeAttribute(ATTRIBUTE_CARRIED, true);
			out.finishTagEOL();
			for (iterator = mCarriedEquipment.getTopLevelRows().iterator(); iterator.hasNext();) {
				((CMEquipment) iterator.next()).save(out, false);
			}
			out.endTagEOL(CMEquipmentList.TAG_ROOT, true);
		}

		if (mOtherEquipment.getRowCount() > 0) {
			out.startTag(CMEquipmentList.TAG_ROOT);
			out.writeAttribute(ATTRIBUTE_CARRIED, false);
			out.finishTagEOL();
			for (iterator = mOtherEquipment.getTopLevelRows().iterator(); iterator.hasNext();) {
				((CMEquipment) iterator.next()).save(out, false);
			}
			out.endTagEOL(CMEquipmentList.TAG_ROOT, true);
		}

		if (mCustomPortrait && mPortrait != null) {
			try {
				ByteArrayOutputStream baos = new ByteArrayOutputStream();

				ImageIO.write(mPortrait, "png", baos); //$NON-NLS-1$
				baos.close();
				out.writeComment(Msgs.PORTRAIT_COMMENT);
				out.startSimpleTagEOL(TAG_PORTRAIT);
				out.println(TKBase64.encode(baos.toByteArray()));
				out.endTagEOL(TAG_PORTRAIT, true);
			} catch (Exception ex) {
				throw new RuntimeException(Msgs.PORTRAIT_WRITE_ERROR);
			}
		}
	}

	/**
	 * @param id The field ID to retrieve the data for.
	 * @return The value of the specified field ID, or <code>null</code> if the field ID is
	 *         invalid.
	 */
	public Object getValueForID(String id) {
		if (id != null && id.startsWith(POINTS_PREFIX)) {
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
			return getCreatedOn();
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
			return new Double(getBasicLift());
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
		} else if (ID_NAME.equals(id)) {
			return getName();
		} else if (ID_NOTES.equals(id)) {
			return getNotes();
		} else if (ID_TITLE.equals(id)) {
			return getTitle();
		} else if (ID_AGE.equals(id)) {
			return new Integer(getAge());
		} else if (ID_BIRTHDAY.equals(id)) {
			return getBirthday();
		} else if (ID_EYE_COLOR.equals(id)) {
			return getEyeColor();
		} else if (ID_HAIR.equals(id)) {
			return getHair();
		} else if (ID_SKIN_COLOR.equals(id)) {
			return getSkinColor();
		} else if (ID_HANDEDNESS.equals(id)) {
			return getHandedness();
		} else if (ID_HEIGHT.equals(id)) {
			return new Integer(getHeight());
		} else if (ID_WEIGHT.equals(id)) {
			return new Double(getWeight());
		} else if (ID_GENDER.equals(id)) {
			return getGender();
		} else if (ID_RACE.equals(id)) {
			return getRace();
		} else if (ID_RELIGION.equals(id)) {
			return getReligion();
		} else if (ID_PLAYER_NAME.equals(id)) {
			return getPlayerName();
		} else if (ID_CAMPAIGN.equals(id)) {
			return getCampaign();
		} else if (ID_SIZE_MODIFIER.equals(id)) {
			return new Integer(getSizeModifier());
		} else if (ID_TECH_LEVEL.equals(id)) {
			return getTechLevel();
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
		} else if (ID_SKULL_DR.equals(id)) {
			return new Integer(getSkullDR());
		} else if (ID_EYES_DR.equals(id)) {
			return new Integer(getEyesDR());
		} else if (ID_FACE_DR.equals(id)) {
			return new Integer(getFaceDR());
		} else if (ID_NECK_DR.equals(id)) {
			return new Integer(getNeckDR());
		} else if (ID_TORSO_DR.equals(id)) {
			return new Integer(getTorsoDR());
		} else if (ID_VITALS_DR.equals(id)) {
			return new Integer(getVitalsDR());
		} else if (ID_GROIN_DR.equals(id)) {
			return new Integer(getGroinDR());
		} else if (ID_ARM_DR.equals(id)) {
			return new Integer(getArmDR());
		} else if (ID_HAND_DR.equals(id)) {
			return new Integer(getHandDR());
		} else if (ID_LEG_DR.equals(id)) {
			return new Integer(getLegDR());
		} else if (ID_FOOT_DR.equals(id)) {
			return new Integer(getFootDR());
		} else if (ID_ONE_HANDED_LIFT.equals(id)) {
			return new Double(getOneHandedLift());
		} else if (ID_TWO_HANDED_LIFT.equals(id)) {
			return new Double(getTwoHandedLift());
		} else if (ID_SHOVE_AND_KNOCK_OVER.equals(id)) {
			return new Double(getShoveAndKnockOver());
		} else if (ID_RUNNING_SHOVE_AND_KNOCK_OVER.equals(id)) {
			return new Double(getRunningShoveAndKnockOver());
		} else if (ID_CARRY_ON_BACK.equals(id)) {
			return new Double(getCarryOnBack());
		} else if (ID_SHIFT_SLIGHTLY.equals(id)) {
			return new Double(getShiftSlightly());
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
		} else {
			for (int i = 0; i < ENCUMBRANCE_LEVELS; i++) {
				if ((DODGE_PREFIX + i).equals(id)) {
					return new Integer(getDodge(i));
				}
				if ((MOVE_PREFIX + i).equals(id)) {
					return new Integer(getMove(i));
				}
				if ((MAXIMUM_CARRY_PREFIX + i).equals(id)) {
					return new Double(getMaximumCarry(i));
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
		if (ID_CREATED_ON.equals(id)) {
			if (value instanceof Long) {
				setCreatedOn(((Long) value).longValue());
			} else {
				setCreatedOn((String) value);
			}
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
		} else if (ID_NAME.equals(id)) {
			setName((String) value);
		} else if (ID_NOTES.equals(id)) {
			setNotes((String) value);
		} else if (ID_TITLE.equals(id)) {
			setTitle((String) value);
		} else if (ID_AGE.equals(id)) {
			setAge(((Integer) value).intValue());
		} else if (ID_BIRTHDAY.equals(id)) {
			setBirthday((String) value);
		} else if (ID_EYE_COLOR.equals(id)) {
			setEyeColor((String) value);
		} else if (ID_HAIR.equals(id)) {
			setHair((String) value);
		} else if (ID_SKIN_COLOR.equals(id)) {
			setSkinColor((String) value);
		} else if (ID_HANDEDNESS.equals(id)) {
			setHandedness((String) value);
		} else if (ID_HEIGHT.equals(id)) {
			setHeight(((Integer) value).intValue());
		} else if (ID_WEIGHT.equals(id)) {
			setWeight(((Double) value).doubleValue());
		} else if (ID_GENDER.equals(id)) {
			setGender((String) value);
		} else if (ID_RACE.equals(id)) {
			setRace((String) value);
		} else if (ID_RELIGION.equals(id)) {
			setReligion((String) value);
		} else if (ID_PLAYER_NAME.equals(id)) {
			setPlayerName((String) value);
		} else if (ID_CAMPAIGN.equals(id)) {
			setCampaign((String) value);
		} else if (ID_SIZE_MODIFIER.equals(id)) {
			setSizeModifier(((Integer) value).intValue());
		} else if (ID_TECH_LEVEL.equals(id)) {
			setTechLevel((String) value);
		} else if (ID_EARNED_POINTS.equals(id)) {
			setEarnedPoints(((Integer) value).intValue());
		} else if (ID_SKULL_DR.equals(id)) {
			setSkullDR(((Integer) value).intValue());
		} else if (ID_EYES_DR.equals(id)) {
			setEyesDR(((Integer) value).intValue());
		} else if (ID_FACE_DR.equals(id)) {
			setFaceDR(((Integer) value).intValue());
		} else if (ID_NECK_DR.equals(id)) {
			setNeckDR(((Integer) value).intValue());
		} else if (ID_TORSO_DR.equals(id)) {
			setTorsoDR(((Integer) value).intValue());
		} else if (ID_VITALS_DR.equals(id)) {
			setVitalsDR(((Integer) value).intValue());
		} else if (ID_GROIN_DR.equals(id)) {
			setGroinDR(((Integer) value).intValue());
		} else if (ID_ARM_DR.equals(id)) {
			setArmDR(((Integer) value).intValue());
		} else if (ID_HAND_DR.equals(id)) {
			setHandDR(((Integer) value).intValue());
		} else if (ID_LEG_DR.equals(id)) {
			setLegDR(((Integer) value).intValue());
		} else if (ID_FOOT_DR.equals(id)) {
			setFootDR(((Integer) value).intValue());
		} else if (ID_HIT_POINTS.equals(id)) {
			setHitPoints(((Integer) value).intValue());
		} else if (ID_CURRENT_HIT_POINTS.equals(id)) {
			setCurrentHitPoints((String) value);
		} else if (ID_FATIGUE_POINTS.equals(id)) {
			setFatiguePoints(((Integer) value).intValue());
		} else if (ID_CURRENT_FATIGUE_POINTS.equals(id)) {
			setCurrentFatiguePoints((String) value);
		} else if (ID_PORTRAIT.equals(id)) {
			if (value instanceof BufferedImage) {
				setPortrait((BufferedImage) value);
			}
		} else {
			assert false : "Unable to set value for: " + id; //$NON-NLS-1$
		}
	}

	@Override protected void startNotifyAtBatchLevelZero() {
		mDidModify = false;
		mNeedAttributePointCalculation = false;
		mNeedAdvantagesPointCalculation = false;
		mNeedSkillPointCalculation = false;
		mNeedSpellPointCalculation = false;
		mNeedEquipmentCalculation = false;
	}

	@Override public void notify(String type, Object data) {
		super.notify(type, data);
		if (CMAdvantage.ID_POINTS.equals(type) || CMAdvantage.ID_LEVELS.equals(type) || CMAdvantage.ID_CONTAINER_TYPE.equals(type) || CMAdvantage.ID_LIST_CHANGED.equals(type) || CMModifier.ID_LIST_CHANGED.equals(type) || CMModifier.ID_ENABLED.equals(type)) {
			mNeedAdvantagesPointCalculation = true;
		}
		if (CMSkill.ID_POINTS.equals(type) || CMSkill.ID_LIST_CHANGED.equals(type)) {
			mNeedSkillPointCalculation = true;
		}
		if (CMSpell.ID_POINTS.equals(type) || CMSpell.ID_LIST_CHANGED.equals(type)) {
			mNeedSpellPointCalculation = true;
		}
		if (CMEquipment.ID_QUANTITY.equals(type) || CMEquipment.ID_WEIGHT.equals(type) || CMEquipment.ID_EXTENDED_WEIGHT.equals(type) || CMEquipment.ID_LIST_CHANGED.equals(type)) {
			mNeedEquipmentCalculation = true;
		}
	}

	@Override protected void notifyOccured() {
		mDidModify = true;
	}

	@Override protected void endNotifyAtBatchLevelOne() {
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
			double savedWeight = mCachedWeightCarried;
			double savedWealth = mCachedWealthCarried;

			calculateWeightAndWealthCarried();
			if (savedWeight != mCachedWeightCarried) {
				notify(ID_CARRIED_WEIGHT, new Double(mCachedWeightCarried));
			}
			if (savedWealth != mCachedWealthCarried) {
				notify(ID_CARRIED_WEALTH, new Double(mCachedWealthCarried));
			}
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

		return MessageFormat.format(Msgs.LAST_MODIFIED, DateFormat.getTimeInstance(DateFormat.SHORT).format(date), DateFormat.getDateInstance(DateFormat.MEDIUM).format(date));
	}

	/** @return The created on date. */
	public String getCreatedOn() {
		Date date = new Date(mCreatedOn);

		return DateFormat.getDateInstance(DateFormat.MEDIUM).format(date);
	}

	/**
	 * Sets the created on date.
	 * 
	 * @param date The new created on date.
	 */
	public void setCreatedOn(long date) {
		if (mCreatedOn != date) {
			postUndoEdit(Msgs.CREATED_ON_UNDO, ID_CREATED_ON, new Long(mCreatedOn), new Long(date));
			mCreatedOn = date;
			notifySingle(ID_CREATED_ON, getCreatedOn());
		}
	}

	/**
	 * Sets the created on date.
	 * 
	 * @param date The new created on date.
	 */
	public void setCreatedOn(String date) {
		setCreatedOn(TKNumberUtils.getDate(date));
	}

	private void updateSkills() {
		for (CMSkill skill : getSkillsIterator()) {
			skill.updateLevel(true);
		}
		mSkillsUpdated = true;
	}

	private void updateSpells() {
		for (CMSpell spell : getSpellsIterator()) {
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
			postUndoEdit(Msgs.STRENGTH_UNDO, ID_STRENGTH, new Integer(oldStrength), new Integer(strength));
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
		CMDice thrust = getThrust();
		CMDice swing = getSwing();
		double lift = getBasicLift();
		boolean notifyST = mStrength != strength || mStrengthBonus != bonus;
		CMDice dice;
		double newLift;

		mStrength = strength;
		mStrengthBonus = bonus;
		mLiftingStrengthBonus = liftingBonus;
		mStrikingStrengthBonus = strikingBonus;

		startNotify();
		if (notifyST) {
			notify(ID_STRENGTH, new Integer(getStrength()));
			notifyOfBaseHitPointChange();
		}
		newLift = getBasicLift();
		if (newLift != lift) {
			notify(ID_BASIC_LIFT, new Double(newLift));
			notify(ID_ONE_HANDED_LIFT, new Double(getOneHandedLift()));
			notify(ID_TWO_HANDED_LIFT, new Double(getTwoHandedLift()));
			notify(ID_SHOVE_AND_KNOCK_OVER, new Double(getShoveAndKnockOver()));
			notify(ID_RUNNING_SHOVE_AND_KNOCK_OVER, new Double(getRunningShoveAndKnockOver()));
			notify(ID_CARRY_ON_BACK, new Double(getCarryOnBack()));
			notify(ID_SHIFT_SLIGHTLY, new Double(getShiftSlightly()));
			for (int i = 0; i < ENCUMBRANCE_LEVELS; i++) {
				notify(MAXIMUM_CARRY_PREFIX + i, new Double(getMaximumCarry(i)));
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
		return getPointsForAttribute(mStrength - 10, 10, mStrengthCostReduction + getSizeModifier() * 10);
	}

	private int getPointsForAttribute(int delta, int ptsPerLevel, int reduction) {
		int amt = delta * ptsPerLevel;

		if (reduction > 0) {
			int rounder = delta < 0 ? -99 : 99;

			if (reduction > 80) {
				reduction = 80;
			}
			amt = (rounder + amt * (100 - reduction)) / 100;
		}

		return amt;
	}

	/** @return The basic thrusting damage. */
	public CMDice getThrust() {
		return getThrust(getStrength() + mStrikingStrengthBonus);
	}

	/**
	 * @param strength The strength to return basic thrusting damage for.
	 * @return The basic thrusting damage.
	 */
	public static CMDice getThrust(int strength) {
		int value = strength;

		if (value < 19) {
			return new CMDice(1, -(6 - (value - 1) / 2));
		}

		value -= 11;
		if (strength > 50) {
			value--;
			if (strength > 79) {
				value -= 1 + (strength - 80) / 5;
			}
		}
		return new CMDice(value / 8 + 1, value % 8 / 2 - 1);
	}

	/** @return The basic swinging damage. */
	public CMDice getSwing() {
		return getSwing(getStrength() + mStrikingStrengthBonus);
	}

	/**
	 * @param strength The strength to return basic swinging damage for.
	 * @return The basic thrusting damage.
	 */
	public static CMDice getSwing(int strength) {
		int value = strength;

		if (value < 10) {
			return new CMDice(1, -(5 - (value - 1) / 2));
		}

		if (value < 28) {
			value -= 9;
			return new CMDice(value / 4 + 1, value % 4 - 1);
		}

		if (strength > 40) {
			value -= (strength - 40) / 5;
		}

		if (strength > 59) {
			value++;
		}
		value += 9;
		return new CMDice(value / 8 + 1, value % 8 / 2 - 1);
	}

	/**
	 * @return Basic lift.
	 */
	public double getBasicLift() {
		int strength = getStrength() + mLiftingStrengthBonus;
		double value = strength * strength / 5.0;

		if (value >= 10.0) {
			value = Math.round(value);
		}
		return value;
	}

	/** @return The one-handed lift value. */
	public double getOneHandedLift() {
		return getBasicLift() * 2.0;
	}

	/** @return The two-handed lift value. */
	public double getTwoHandedLift() {
		return getBasicLift() * 8.0;
	}

	/** @return The shove and knock over value. */
	public double getShoveAndKnockOver() {
		return getBasicLift() * 12.0;
	}

	/** @return The running shove and knock over value. */
	public double getRunningShoveAndKnockOver() {
		return getBasicLift() * 24.0;
	}

	/** @return The carry on back value. */
	public double getCarryOnBack() {
		return getBasicLift() * 15.0;
	}

	/** @return The shift slightly value. */
	public double getShiftSlightly() {
		return getBasicLift() * 50.0;
	}

	/**
	 * @param level The encumbrance level
	 * @return The maximum amount the character can carry for the specified encumbrance level.
	 */
	public double getMaximumCarry(int level) {
		assert level >= ENCUMBRANCE_NONE && level < ENCUMBRANCE_LEVELS;

		return Math.floor(getBasicLift() * ENCUMBRANCE_MULTIPLIER[level] * 10.0) / 10.0;
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
			postUndoEdit(Msgs.BASIC_SPEED_UNDO, ID_BASIC_SPEED, new Double(oldBasicSpeed), new Double(speed));
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
			postUndoEdit(Msgs.BASIC_MOVE_UNDO, ID_BASIC_MOVE, new Integer(oldBasicMove), new Integer(move));
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
	 * @param level The encumbrance level
	 * @return The character's ground move for the specified encumbrance level.
	 */
	public int getMove(int level) {
		assert level >= ENCUMBRANCE_NONE && level < ENCUMBRANCE_LEVELS;

		int basicMove = getBasicMove();
		int move = basicMove * (10 - 2 * level) / 10;

		if (move < 1) {
			return basicMove > 0 ? 1 : 0;
		}
		return move;
	}

	/**
	 * @param level The encumbrance level
	 * @return The character's dodge for the specified encumbrance level.
	 */
	public int getDodge(int level) {
		assert level >= ENCUMBRANCE_NONE && level < ENCUMBRANCE_LEVELS;

		int dodge = (int) Math.floor(getBasicSpeed()) + 3 + getEncumbrancePenalty(level) + mDodgeBonus;

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

	/**
	 * @param level The encumbrance level (0-4)
	 * @return The encumbrance penalty for the specified encumbrance level.
	 */
	public int getEncumbrancePenalty(int level) {
		return -level;
	}

	/**
	 * @return The current encumbrance level.
	 */
	public int getEncumbranceLevel() {
		double carried = getWeightCarried();

		for (int i = ENCUMBRANCE_NONE; i < ENCUMBRANCE_LEVELS; i++) {
			if (carried <= getMaximumCarry(i)) {
				return i;
			}
		}
		return ENCUMBRANCE_EXTRA_HEAVY;
	}

	/** @return The current weight being carried. */
	public double getWeightCarried() {
		return mCachedWeightCarried;
	}

	/** @return The current wealth being carried. */
	public double getWealthCarried() {
		return mCachedWealthCarried;
	}

	private void calculateWeightAndWealthCarried() {
		mCachedWeightCarried = 0.0;
		mCachedWealthCarried = 0.0;
		for (CMEquipment equipment : getCarriedEquipmentIterator()) {
			int quantity = equipment.getQuantity();

			mCachedWeightCarried += quantity * equipment.getWeight();
			mCachedWealthCarried += quantity * equipment.getValue();
		}
	}

	private int[] preserveMoveAndDodge() {
		int[] data = new int[ENCUMBRANCE_LEVELS * 2];

		for (int i = 0; i < ENCUMBRANCE_LEVELS; i++) {
			data[i] = getMove(i);
			data[ENCUMBRANCE_LEVELS + i] = getDodge(i);
		}
		return data;
	}

	private void notifyIfMoveOrDodgeAltered(int[] data) {
		for (int i = 0; i < ENCUMBRANCE_LEVELS; i++) {
			int tmp = getDodge(i);

			if (tmp != data[ENCUMBRANCE_LEVELS + i]) {
				notify(DODGE_PREFIX + i, new Integer(tmp));
			}
			tmp = getMove(i);
			if (tmp != data[i]) {
				notify(MOVE_PREFIX + i, new Integer(tmp));
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
			postUndoEdit(Msgs.DEXTERITY_UNDO, ID_DEXTERITY, new Integer(oldDexterity), new Integer(dexterity));
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
			postUndoEdit(Msgs.INTELLIGENCE_UNDO, ID_INTELLIGENCE, new Integer(oldIntelligence), new Integer(intelligence));
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
			postUndoEdit(Msgs.HEALTH_UNDO, ID_HEALTH, new Integer(oldHealth), new Integer(health));
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

			postUndoEdit(Msgs.EARNED_POINTS_UNDO, ID_EARNED_POINTS, new Integer(current), value);
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

		for (CMAdvantage advantage : new TKFilteredIterator<CMAdvantage>(mAdvantages.getTopLevelRows(), CMAdvantage.class)) {
			calculateSingleAdvantagePoints(advantage);
		}
	}

	private void calculateSingleAdvantagePoints(CMAdvantage advantage) {
		if (advantage.canHaveChildren()) {
			CMAdvantageContainerType type = advantage.getContainerType();

			if (type == CMAdvantageContainerType.GROUP) {
				for (CMAdvantage child : new TKFilteredIterator<CMAdvantage>(advantage.getChildren(), CMAdvantage.class)) {
					calculateSingleAdvantagePoints(child);
				}
				return;
			} else if (type == CMAdvantageContainerType.RACE) {
				mCachedRacePoints = advantage.getAdjustedPoints();
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
		for (CMSkill skill : getSkillsIterator()) {
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
		for (CMSpell spell : getSpellsIterator()) {
			if (!spell.canHaveChildren()) {
				mCachedSpellPoints += spell.getPoints();
			}
		}
	}

	/**
	 * @param forPrinting Pass in <code>true</code> to retrieve the double resolution portrait
	 *            (for printing), or <code>false</code> to retrieve the normal resolution portrait
	 *            (for display).
	 * @return The portrait.
	 */
	public BufferedImage getPortrait(boolean forPrinting) {
		return forPrinting ? mPortrait : mDisplayPortrait;
	}

	/**
	 * Sets the portrait.
	 * 
	 * @param portrait The new portrait.
	 */
	public void setPortrait(BufferedImage portrait) {
		if (mPortrait != portrait) {
			mCustomPortrait = true;
			postUndoEdit(Msgs.PORTRAIT_UNDO, ID_PORTRAIT, mPortrait, portrait);
			setPortraitInternal(portrait);
			notifySingle(ID_PORTRAIT, mPortrait);
		}
	}

	private void setPortraitInternal(BufferedImage portrait) {
		if (portrait == null) {
			mPortrait = null;
			mDisplayPortrait = null;
		} else {
			if (portrait.getWidth() != PORTRAIT_WIDTH * 2 || portrait.getHeight() != PORTRAIT_HEIGHT * 2) {
				portrait = TKImage.scale(portrait, PORTRAIT_WIDTH * 2, PORTRAIT_HEIGHT * 2);
			}
			mPortrait = portrait;
			mDisplayPortrait = TKImage.scale(mPortrait, PORTRAIT_WIDTH, PORTRAIT_HEIGHT);
		}
	}

	/** @return The name. */
	@Override public String getName() {
		return mName;
	}

	/**
	 * Sets the name.
	 * 
	 * @param name The new name.
	 */
	public void setName(String name) {
		if (!mName.equals(name)) {
			postUndoEdit(Msgs.NAME_UNDO, ID_NAME, mName, name);
			mName = name;
			notifySingle(ID_NAME, mName);
		}
	}

	/** @return The notes. */
	public String getNotes() {
		return mNotes;
	}

	/**
	 * Sets the notes.
	 * 
	 * @param notes The new notes.
	 */
	public void setNotes(String notes) {
		if (!mNotes.equals(notes)) {
			postUndoEdit(Msgs.NOTES_UNDO, ID_NOTES, mNotes, notes);
			mNotes = notes;
			notifySingle(ID_NOTES, mNotes);
		}
	}

	/** @return The title. */
	public String getTitle() {
		return mTitle;
	}

	/**
	 * Sets the title.
	 * 
	 * @param title The new title.
	 */
	public void setTitle(String title) {
		if (!mTitle.equals(title)) {
			postUndoEdit(Msgs.TITLE_UNDO, ID_TITLE, mTitle, title);
			mTitle = title;
			notifySingle(ID_TITLE, mTitle);
		}
	}

	/** @return The age. */
	public int getAge() {
		return mAge;
	}

	/**
	 * Sets the age.
	 * 
	 * @param age The new age.
	 */
	public void setAge(int age) {
		if (mAge != age) {
			Integer value = new Integer(age);

			postUndoEdit(Msgs.AGE_UNDO, ID_AGE, new Integer(mAge), value);
			mAge = age;
			notifySingle(ID_AGE, value);
		}
	}

	/** @return A random age. */
	public int getRandomAge() {
		CMAdvantage lifespan = getAdvantageNamed("Unaging"); //$NON-NLS-1$
		int base = 16;
		int mod = 7;
		int levels;

		if (lifespan != null) {
			return 18 + RANDOM.nextInt(7);
		}

		if (RANDOM.nextInt(3) == 1) {
			mod += 7;
			if (RANDOM.nextInt(4) == 1) {
				mod += 13;
			}
		}

		lifespan = getAdvantageNamed("Short Lifespan"); //$NON-NLS-1$
		if (lifespan != null) {
			levels = lifespan.getLevels();
			base = base >> levels;
			mod = mod >> levels;
		} else {
			lifespan = getAdvantageNamed("Extended Lifespan"); //$NON-NLS-1$
			if (lifespan != null) {
				levels = lifespan.getLevels();
				base = base << levels;
				mod = mod << levels;
			}
		}
		if (mod < 1) {
			mod = 1;
		}

		return base + RANDOM.nextInt(mod);
	}

	/** @return The date of birth. */
	public String getBirthday() {
		return mBirthday;
	}

	/**
	 * Sets the date of birth.
	 * 
	 * @param birthday The new date of birth.
	 */
	public void setBirthday(String birthday) {
		if (!mBirthday.equals(birthday)) {
			postUndoEdit(Msgs.BIRTHDAY_UNDO, ID_BIRTHDAY, mBirthday, birthday);
			mBirthday = birthday;
			notifySingle(ID_BIRTHDAY, mBirthday);
		}
	}

	/** @return The eye color. */
	public String getEyeColor() {
		return mEyeColor;
	}

	/**
	 * Sets the eye color.
	 * 
	 * @param eyeColor The new eye color.
	 */
	public void setEyeColor(String eyeColor) {
		if (!mEyeColor.equals(eyeColor)) {
			postUndoEdit(Msgs.EYE_COLOR_UNDO, ID_EYE_COLOR, mEyeColor, eyeColor);
			mEyeColor = eyeColor;
			notifySingle(ID_EYE_COLOR, mEyeColor);
		}
	}

	/** @return The hair. */
	public String getHair() {
		return mHair;
	}

	/**
	 * Sets the hair.
	 * 
	 * @param hair The new hair.
	 */
	public void setHair(String hair) {
		if (!mHair.equals(hair)) {
			postUndoEdit(Msgs.HAIR_UNDO, ID_HAIR, mHair, hair);
			mHair = hair;
			notifySingle(ID_HAIR, mHair);
		}
	}

	/** @return The skin color. */
	public String getSkinColor() {
		return mSkinColor;
	}

	/**
	 * Sets the skin color.
	 * 
	 * @param skinColor The new skin color.
	 */
	public void setSkinColor(String skinColor) {
		if (!mSkinColor.equals(skinColor)) {
			postUndoEdit(Msgs.SKIN_COLOR_UNDO, ID_SKIN_COLOR, mSkinColor, skinColor);
			mSkinColor = skinColor;
			notifySingle(ID_SKIN_COLOR, mSkinColor);
		}
	}

	/** @return The handedness. */
	public String getHandedness() {
		return mHandedness;
	}

	/**
	 * Sets the handedness.
	 * 
	 * @param handedness The new handedness.
	 */
	public void setHandedness(String handedness) {
		if (!mHandedness.equals(handedness)) {
			postUndoEdit(Msgs.HANDEDNESS_UNDO, ID_HANDEDNESS, mHandedness, handedness);
			mHandedness = handedness;
			notifySingle(ID_HANDEDNESS, mHandedness);
		}
	}

	/** @return The height, in inches. */
	public int getHeight() {
		return mHeight;
	}

	/**
	 * Sets the height.
	 * 
	 * @param height The new height.
	 */
	public void setHeight(int height) {
		if (mHeight != height) {
			Integer value = new Integer(height);

			postUndoEdit(Msgs.HEIGHT_UNDO, ID_HEIGHT, new Integer(mHeight), value);
			mHeight = height;
			notifySingle(ID_HEIGHT, value);
		}
	}

	/** @return The weight. */
	public double getWeight() {
		return mWeight;
	}

	/**
	 * Sets the weight.
	 * 
	 * @param weight The new weight.
	 */
	public void setWeight(double weight) {
		if (mWeight != weight) {
			Double value = new Double(weight);

			postUndoEdit(Msgs.WEIGHT_UNDO, ID_WEIGHT, new Double(mWeight), value);
			mWeight = weight;
			notifySingle(ID_WEIGHT, value);
		}
	}

	/** @return The multiplier compared to average weight for this character. */
	public double getWeightMultiplier() {
		if (hasAdvantageNamed("Very Fat")) { //$NON-NLS-1$
			return 2.0;
		} else if (hasAdvantageNamed("Fat")) { //$NON-NLS-1$
			return 1.5;
		} else if (hasAdvantageNamed("Overweight")) { //$NON-NLS-1$
			return 1.3;
		} else if (hasAdvantageNamed("Skinny")) { //$NON-NLS-1$
			return 0.67;
		}
		return 1.0;
	}

	/** @return The size modifier. */
	public int getSizeModifier() {
		return mSizeModifier + mSizeModifierBonus;
	}

	/** @return The size modifier bonus. */
	public int getSizeModifierBonus() {
		return mSizeModifierBonus;
	}

	/** @param size The new size modifier. */
	public void setSizeModifier(int size) {
		int totalSizeModifier = getSizeModifier();

		if (totalSizeModifier != size) {
			Integer value = new Integer(size);

			postUndoEdit(Msgs.SIZE_MODIFIER_UNDO, ID_SIZE_MODIFIER, new Integer(totalSizeModifier), value);
			mSizeModifier = size - mSizeModifierBonus;
			notifySingle(ID_SIZE_MODIFIER, value);
		}
	}

	/** @param bonus The new size modifier bonus. */
	public void setSizeModifierBonus(int bonus) {
		if (mSizeModifierBonus != bonus) {
			mSizeModifierBonus = bonus;
			notifySingle(ID_SIZE_MODIFIER, new Integer(getSizeModifier()));
		}
	}

	/** @return Whether to include the punch natural weapon or not. */
	public boolean includePunch() {
		return mIncludePunch;
	}

	/** @param include Whether to include the punch natural weapon or not. */
	public void setIncludePunch(boolean include) {
		if (mIncludePunch != include) {
			postUndoEdit(Msgs.INCLUDE_PUNCH_UNDO, ID_INCLUDE_PUNCH, new Boolean(mIncludePunch), new Boolean(include));
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
			postUndoEdit(Msgs.INCLUDE_KICK_UNDO, ID_INCLUDE_KICK, new Boolean(mIncludeKick), new Boolean(include));
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
			postUndoEdit(Msgs.INCLUDE_BOOTS_UNDO, ID_INCLUDE_BOOTS, new Boolean(mIncludeKickBoots), new Boolean(include));
			mIncludeKickBoots = include;
			notifySingle(ID_INCLUDE_BOOTS, new Boolean(mIncludeKickBoots));
		}
	}

	/** @return The gender. */
	public String getGender() {
		return mGender;
	}

	/**
	 * Sets the gender.
	 * 
	 * @param gender The new gender.
	 */
	public void setGender(String gender) {
		if (!mGender.equals(gender)) {
			postUndoEdit(Msgs.GENDER_UNDO, ID_GENDER, mGender, gender);
			mGender = gender;
			notifySingle(ID_GENDER, mGender);
		}
	}

	/** @return The race. */
	public String getRace() {
		return mRace;
	}

	/**
	 * Sets the race.
	 * 
	 * @param race The new race.
	 */
	public void setRace(String race) {
		if (!mRace.equals(race)) {
			postUndoEdit(Msgs.RACE_UNDO, ID_RACE, mRace, race);
			mRace = race;
			notifySingle(ID_RACE, mRace);
		}
	}

	/** @return The religion. */
	public String getReligion() {
		return mReligion;
	}

	/**
	 * Sets the religion.
	 * 
	 * @param religion The new religion.
	 */
	public void setReligion(String religion) {
		if (!mReligion.equals(religion)) {
			postUndoEdit(Msgs.RELIGION_UNDO, ID_RELIGION, mReligion, religion);
			mReligion = religion;
			notifySingle(ID_RELIGION, mReligion);
		}
	}

	/** @return The player's name. */
	public String getPlayerName() {
		return mPlayerName;
	}

	/**
	 * Sets the player's name.
	 * 
	 * @param player The new player's name.
	 */
	public void setPlayerName(String player) {
		if (!mPlayerName.equals(player)) {
			postUndoEdit(Msgs.PLAYER_NAME_UNDO, ID_PLAYER_NAME, mPlayerName, player);
			mPlayerName = player;
			notifySingle(ID_PLAYER_NAME, mPlayerName);
		}
	}

	/** @return The campaign. */
	public String getCampaign() {
		return mCampaign;
	}

	/**
	 * Sets the campaign.
	 * 
	 * @param campaign The new campaign.
	 */
	public void setCampaign(String campaign) {
		if (!mCampaign.equals(campaign)) {
			postUndoEdit(Msgs.CAMPAIGN_UNDO, ID_CAMPAIGN, mCampaign, campaign);
			mCampaign = campaign;
			notifySingle(ID_CAMPAIGN, mCampaign);
		}
	}

	/** @return The tech level. */
	public String getTechLevel() {
		return mTechLevel;
	}

	/**
	 * Sets the tech level.
	 * 
	 * @param techLevel The new tech level.
	 */
	public void setTechLevel(String techLevel) {
		if (!mTechLevel.equals(techLevel)) {
			postUndoEdit(Msgs.TECH_LEVEL_UNDO, ID_TECH_LEVEL, mTechLevel, techLevel);
			mTechLevel = techLevel;
			notifySingle(ID_TECH_LEVEL, mTechLevel);
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
			postUndoEdit(Msgs.HIT_POINTS_UNDO, ID_HIT_POINTS, new Integer(oldHP), new Integer(hp));
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
		int sizeModifier = getSizeModifier();

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
			postUndoEdit(Msgs.CURRENT_HIT_POINTS_UNDO, ID_CURRENT_HIT_POINTS, mCurrentHitPoints, hp);
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
		return mWill + mWillBonus + getIntelligence();
	}

	/**
	 * Sets the will.
	 * 
	 * @param will The new will.
	 */
	public void setWill(int will) {
		int oldWill = getWill();

		if (oldWill != will) {
			postUndoEdit(Msgs.WILL_UNDO, ID_WILL, new Integer(oldWill), new Integer(will));
			updateWillInfo(will - (mWillBonus + getIntelligence()), mWillBonus);
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
		updateSkills();
		mNeedAttributePointCalculation = true;
		endNotify();
	}

	/** @return The number of points spent on will. */
	public int getWillPoints() {
		return mWill * 5;
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
		return mPerception + mPerceptionBonus + getIntelligence();
	}

	/**
	 * Sets the perception.
	 * 
	 * @param perception The new perception.
	 */
	public void setPerception(int perception) {
		int oldPerception = getPerception();

		if (oldPerception != perception) {
			postUndoEdit(Msgs.PERCEPTION_UNDO, ID_PERCEPTION, new Integer(oldPerception), new Integer(perception));
			updatePerceptionInfo(perception - (mPerceptionBonus + getIntelligence()), mPerceptionBonus);
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
			postUndoEdit(Msgs.FATIGUE_POINTS_UNDO, ID_FATIGUE_POINTS, new Integer(oldFP), new Integer(fp));
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
			postUndoEdit(Msgs.CURRENT_FATIGUE_POINTS_UNDO, ID_CURRENT_FATIGUE_POINTS, mCurrentFatiguePoints, fp);
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
	public int getUnconsciousChecksFatiguePoints() {
		return 0;
	}

	/** @return The number of hit points where the character falls over, unconscious. */
	public int getUnconsciousFatiguePoints() {
		return -1 * getFatiguePoints();
	}

	/** @return The skull hit location's DR. */
	public int getSkullDR() {
		return mSkullDR;
	}

	/**
	 * Sets the skull hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setSkullDR(int dr) {
		if (mSkullDR != dr) {
			mSkullDR = dr;
			notifySingle(ID_SKULL_DR, new Integer(mSkullDR));
		}
	}

	/** @return The eyes hit location's DR. */
	public int getEyesDR() {
		return mEyesDR;
	}

	/**
	 * Sets the eyes hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setEyesDR(int dr) {
		if (mEyesDR != dr) {
			mEyesDR = dr;
			notifySingle(ID_EYES_DR, new Integer(mEyesDR));
		}
	}

	/** @return The face hit location's DR. */
	public int getFaceDR() {
		return mFaceDR;
	}

	/**
	 * Sets the face hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setFaceDR(int dr) {
		if (mFaceDR != dr) {
			mFaceDR = dr;
			notifySingle(ID_FACE_DR, new Integer(mFaceDR));
		}
	}

	/** @return The neck hit location's DR. */
	public int getNeckDR() {
		return mNeckDR;
	}

	/**
	 * Sets the neck hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setNeckDR(int dr) {
		if (mNeckDR != dr) {
			mNeckDR = dr;
			notifySingle(ID_NECK_DR, new Integer(mNeckDR));
		}
	}

	/** @return The torso hit location's DR. */
	public int getTorsoDR() {
		return mTorsoDR;
	}

	/**
	 * Sets the torso hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setTorsoDR(int dr) {
		if (mTorsoDR != dr) {
			mTorsoDR = dr;
			notifySingle(ID_TORSO_DR, new Integer(mTorsoDR));
		}
	}

	/** @return The vitals hit location's DR. */
	public int getVitalsDR() {
		return mVitalsDR;
	}

	/**
	 * Sets the vitals hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setVitalsDR(int dr) {
		if (mVitalsDR != dr) {
			mVitalsDR = dr;
			notifySingle(ID_VITALS_DR, new Integer(mVitalsDR));
		}
	}

	/** @return The groin hit location's DR. */
	public int getGroinDR() {
		return mGroinDR;
	}

	/**
	 * Sets the groin hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setGroinDR(int dr) {
		if (mGroinDR != dr) {
			mGroinDR = dr;
			notifySingle(ID_GROIN_DR, new Integer(mGroinDR));
		}
	}

	/** @return The arm hit location's DR. */
	public int getArmDR() {
		return mArmDR;
	}

	/**
	 * Sets the arm hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setArmDR(int dr) {
		if (mArmDR != dr) {
			mArmDR = dr;
			notifySingle(ID_ARM_DR, new Integer(mArmDR));
		}
	}

	/** @return The hand hit location's DR. */
	public int getHandDR() {
		return mHandDR;
	}

	/**
	 * Sets the hand hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setHandDR(int dr) {
		if (mHandDR != dr) {
			mHandDR = dr;
			notifySingle(ID_HAND_DR, new Integer(mHandDR));
		}
	}

	/** @return The leg hit location's DR. */
	public int getLegDR() {
		return mLegDR;
	}

	/**
	 * Sets the leg hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setLegDR(int dr) {
		if (mLegDR != dr) {
			mLegDR = dr;
			notifySingle(ID_LEG_DR, new Integer(mLegDR));
		}
	}

	/** @return The foot hit location's DR. */
	public int getFootDR() {
		return mFootDR;
	}

	/**
	 * Sets the foot hit location's DR.
	 * 
	 * @param dr The DR amount.
	 */
	public void setFootDR(int dr) {
		if (mFootDR != dr) {
			mFootDR = dr;
			notifySingle(ID_FOOT_DR, new Integer(mFootDR));
		}
	}

	/** @return The outline model for the character's (dis)advantages. */
	public TKOutlineModel getAdvantagesModel() {
		return mAdvantages;
	}

	/** @return A recursive iterator over the character's (dis)advantages. */
	public TKRowIterator<CMAdvantage> getAdvantagesIterator() {
		return new TKRowIterator<CMAdvantage>(mAdvantages);
	}

	/**
	 * Searches the character's current (dis)advantages list for the specified name.
	 * 
	 * @param name The name to look for.
	 * @return The (dis)advantage, if present, or <code>null</code>.
	 */
	public CMAdvantage getAdvantageNamed(String name) {
		for (CMAdvantage advantage : getAdvantagesIterator()) {
			if (advantage.getName().equals(name)) {
				return advantage;
			}
		}
		return null;
	}

	/**
	 * Searches the character's current (dis)advantages list for the specified name.
	 * 
	 * @param name The name to look for.
	 * @return Whether it is present or not.
	 */
	public boolean hasAdvantageNamed(String name) {
		return getAdvantageNamed(name) != null;
	}

	/** @return The outline model for the character's skills. */
	public TKOutlineModel getSkillsRoot() {
		return mSkills;
	}

	/** @return A recursive iterable for the character's skills. */
	public TKRowIterator<CMSkill> getSkillsIterator() {
		return new TKRowIterator<CMSkill>(mSkills);
	}

	/**
	 * Searches the character's current skill list for the specified name.
	 * 
	 * @param name The name to look for.
	 * @param specialization The specialization to look for. Pass in <code>null</code> or an empty
	 *            string to ignore.
	 * @param requirePoints Only look at {@link CMSkill}s that have points.
	 * @param excludes The set of {@link CMSkill}s to exclude from consideration.
	 * @return The skill if it is present, or <code>null</code> if its not.
	 */
	public ArrayList<CMSkill> getSkillNamed(String name, String specialization, boolean requirePoints, HashSet<CMSkill> excludes) {
		ArrayList<CMSkill> skills = new ArrayList<CMSkill>();
		boolean checkSpecialization = specialization != null && specialization.length() > 0;

		for (CMSkill skill : getSkillsIterator()) {
			if (excludes == null || !excludes.contains(skill)) {
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
	 * Searches the character's current {@link CMSkill} list for the {@link CMSkill} with the best
	 * level that matches the name.
	 * 
	 * @param name The {@link CMSkill} name to look for.
	 * @param specialization An optional specialization to look for. Pass <code>null</code> if it
	 *            is not needed.
	 * @param requirePoints Only look at {@link CMSkill}s that have points.
	 * @param excludes The set of {@link CMSkill}s to exclude from consideration.
	 * @return The {@link CMSkill} that matches with the highest level.
	 */
	public CMSkill getBestSkillNamed(String name, String specialization, boolean requirePoints, HashSet<CMSkill> excludes) {
		CMSkill best = null;
		int level = Integer.MIN_VALUE;

		for (CMSkill skill : getSkillNamed(name, specialization, requirePoints, excludes)) {
			int skillLevel = skill.getLevel(excludes);

			if (best == null || skillLevel > level) {
				best = skill;
				level = skillLevel;
			}
		}
		return best;
	}

	/** @return The outline model for the character's spells. */
	public TKOutlineModel getSpellsRoot() {
		return mSpells;
	}

	/** @return A recursive iterator over the character's spells. */
	public TKRowIterator<CMSpell> getSpellsIterator() {
		return new TKRowIterator<CMSpell>(mSpells);
	}

	/** @return The outline model for the character's carried equipment. */
	public TKOutlineModel getCarriedEquipmentRoot() {
		return mCarriedEquipment;
	}

	/** @return A recursive iterator over the character's carried equipment. */
	public TKRowIterator<CMEquipment> getCarriedEquipmentIterator() {
		return new TKRowIterator<CMEquipment>(mCarriedEquipment);
	}

	/** @return The outline model for the character's other equipment. */
	public TKOutlineModel getOtherEquipmentRoot() {
		return mOtherEquipment;
	}

	/** @return A recursive iterator over the character's other equipment. */
	public TKRowIterator<CMEquipment> getOtherEquipmentIterator() {
		return new TKRowIterator<CMEquipment>(mOtherEquipment);
	}

	/** @param map The new feature map. */
	public void setFeatureMap(HashMap<String, ArrayList<CMFeature>> map) {
		int fullBodyDR;
		int fullBodyNoEyesDR;

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
		setPerceptionBonus(getIntegerBonusFor(ID_PERCEPTION));
		setVisionBonus(getIntegerBonusFor(ID_VISION));
		setHearingBonus(getIntegerBonusFor(ID_HEARING));
		setTasteAndSmellBonus(getIntegerBonusFor(ID_TASTE_AND_SMELL));
		setTouchBonus(getIntegerBonusFor(ID_TOUCH));
		setHitPointBonus(getIntegerBonusFor(ID_HIT_POINTS));
		setFatiguePointBonus(getIntegerBonusFor(ID_FATIGUE_POINTS));
		setSizeModifierBonus(getIntegerBonusFor(ID_SIZE_MODIFIER));
		setDodgeBonus(getIntegerBonusFor(ID_DODGE_BONUS));
		setParryBonus(getIntegerBonusFor(ID_PARRY_BONUS));
		setBlockBonus(getIntegerBonusFor(ID_BLOCK_BONUS));
		setBasicSpeedBonus(getDoubleBonusFor(ID_BASIC_SPEED));
		setBasicMoveBonus(getIntegerBonusFor(ID_BASIC_MOVE));
		fullBodyDR = getIntegerBonusFor(ID_FULL_BODY_DR);
		fullBodyNoEyesDR = getIntegerBonusFor(ID_FULL_BODY_EXCEPT_EYES_DR);
		setSkullDR(2 + getIntegerBonusFor(ID_SKULL_DR) + fullBodyDR + fullBodyNoEyesDR);
		setEyesDR(getIntegerBonusFor(ID_EYES_DR) + fullBodyDR);
		setFaceDR(getIntegerBonusFor(ID_FACE_DR) + fullBodyDR + fullBodyNoEyesDR);
		setNeckDR(getIntegerBonusFor(ID_NECK_DR) + fullBodyDR + fullBodyNoEyesDR);
		setTorsoDR(getIntegerBonusFor(ID_TORSO_DR) + fullBodyDR + fullBodyNoEyesDR);
		setVitalsDR(getIntegerBonusFor(ID_VITALS_DR) + fullBodyDR + fullBodyNoEyesDR);
		setGroinDR(getIntegerBonusFor(ID_GROIN_DR) + fullBodyDR + fullBodyNoEyesDR);
		setArmDR(getIntegerBonusFor(ID_ARM_DR) + fullBodyDR + fullBodyNoEyesDR);
		setHandDR(getIntegerBonusFor(ID_HAND_DR) + fullBodyDR + fullBodyNoEyesDR);
		setLegDR(getIntegerBonusFor(ID_LEG_DR) + fullBodyDR + fullBodyNoEyesDR);
		setFootDR(getIntegerBonusFor(ID_FOOT_DR) + fullBodyDR + fullBodyNoEyesDR);
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
		ArrayList<CMFeature> list = mFeatureMap.get(id.toLowerCase());

		if (list != null) {
			for (CMFeature feature : list) {
				if (feature instanceof CMCostReduction) {
					total += ((CMCostReduction) feature).getPercentage();
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
		ArrayList<CMFeature> list = mFeatureMap.get(id.toLowerCase());
		if (list != null) {
			for (CMFeature feature : list) {
				if (feature instanceof CMBonus && !(feature instanceof CMWeaponBonus)) {
					total += ((CMBonus) feature).getAmount().getIntegerAdjustedAmount();
				}
			}
		}
		return total;
	}

	/**
	 * @param id The feature ID to search for.
	 * @param nameQualifier The name qualifier.
	 * @param specializationQualifier The specialization qualifier.
	 * @return The bonus.
	 */
	public ArrayList<CMLeveledAmount> getWeaponComparedBonusesFor(String id, String nameQualifier, String specializationQualifier) {
		ArrayList<CMLeveledAmount> bonuses = new ArrayList<CMLeveledAmount>();
		int rsl = Integer.MIN_VALUE;

		for (CMSkill skill : getSkillNamed(nameQualifier, specializationQualifier, true, null)) {
			int srsl = skill.getRelativeLevel();

			if (srsl > rsl) {
				rsl = srsl;
			}
		}

		if (rsl != Integer.MIN_VALUE) {
			ArrayList<CMFeature> list = mFeatureMap.get(id.toLowerCase());
			if (list != null) {
				for (CMFeature feature : list) {
					if (feature instanceof CMWeaponBonus) {
						CMWeaponBonus bonus = (CMWeaponBonus) feature;

						if (bonus.getNameCriteria().matches(nameQualifier) && bonus.getSpecializationCriteria().matches(specializationQualifier) && bonus.getLevelCriteria().matches(rsl)) {
							bonuses.add(bonus.getAmount());
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
		ArrayList<CMFeature> list = mFeatureMap.get(id.toLowerCase());

		if (list != null) {
			for (CMFeature feature : list) {
				if (feature instanceof CMSkillBonus) {
					CMSkillBonus bonus = (CMSkillBonus) feature;

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
		ArrayList<CMFeature> list = mFeatureMap.get(id.toLowerCase());

		if (list != null) {
			for (CMFeature feature : list) {
				if (feature instanceof CMSpellBonus) {
					CMSpellBonus bonus = (CMSpellBonus) feature;

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
		ArrayList<CMFeature> list = mFeatureMap.get(id.toLowerCase());

		if (list != null) {
			for (CMFeature feature : list) {
				if (feature instanceof CMBonus && !(feature instanceof CMWeaponBonus)) {
					total += ((CMBonus) feature).getAmount().getAdjustedAmount();
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
	private void postUndoEdit(String name, String id, Object before, Object after) {
		if (!isUndoBeingApplied() && !before.equals(after)) {
			addEdit(new CMCharacterFieldUndo(this, name, id, before, after));
		}
	}
}
