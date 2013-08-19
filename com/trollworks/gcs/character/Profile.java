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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.names.USCensusNames;
import com.trollworks.gcs.feature.BonusAttributeType;
import com.trollworks.gcs.utility.collections.EnumExtractor;
import com.trollworks.gcs.utility.io.Base64;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.utility.io.xml.XMLNodeType;
import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.utility.text.NumberUtils;
import com.trollworks.gcs.utility.text.TextUtility;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.IOException;
import java.text.MessageFormat;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.Random;

import javax.imageio.ImageIO;

/** Holds the character profile. */
public class Profile {
	private static String			MSG_HAIR_FORMAT;
	private static String			MSG_BROWN;
	private static String			MSG_BLACK;
	private static String			MSG_BLOND;
	private static String			MSG_REDHEAD;
	private static String			MSG_BALD;
	private static String			MSG_STRAIGHT;
	private static String			MSG_CURLY;
	private static String			MSG_WAVY;
	private static String			MSG_SHORT;
	private static String			MSG_MEDIUM;
	private static String			MSG_LONG;
	private static String			MSG_BLUE;
	private static String			MSG_GREEN;
	private static String			MSG_GREY;
	private static String			MSG_VIOLET;
	private static String			MSG_FRECKLED;
	private static String			MSG_TAN;
	private static String			MSG_LIGHT_TAN;
	private static String			MSG_DARK_TAN;
	private static String			MSG_LIGHT_BROWN;
	private static String			MSG_DARK_BROWN;
	private static String			MSG_PALE;
	private static String			MSG_RIGHT;
	private static String			MSG_LEFT;
	private static String			MSG_MALE;
	private static String			MSG_FEMALE;
	private static String			MSG_DEFAULT_RACE;
	private static String			MSG_NAME_UNDO;
	private static String			MSG_TITLE_UNDO;
	private static String			MSG_AGE_UNDO;
	private static String			MSG_BIRTHDAY_UNDO;
	private static String			MSG_EYE_COLOR_UNDO;
	private static String			MSG_HAIR_UNDO;
	private static String			MSG_SKIN_COLOR_UNDO;
	private static String			MSG_HANDEDNESS_UNDO;
	private static String			MSG_HEIGHT_UNDO;
	private static String			MSG_WEIGHT_UNDO;
	private static String			MSG_GENDER_UNDO;
	private static String			MSG_RACE_UNDO;
	private static String			MSG_RELIGION_UNDO;
	private static String			MSG_PLAYER_NAME_UNDO;
	private static String			MSG_CAMPAIGN_UNDO;
	private static String			MSG_SIZE_MODIFIER_UNDO;
	private static String			MSG_TECH_LEVEL_UNDO;
	private static String			MSG_PORTRAIT_UNDO;
	private static String			MSG_PORTRAIT_COMMENT;
	private static String			MSG_PORTRAIT_WRITE_ERROR;
	private static String			MSG_NOTES_UNDO;

	static {
		LocalizedMessages.initialize(Profile.class);
	}

	/** The root XML tag. */
	public static final String		TAG_ROOT			= "profile";																													//$NON-NLS-1$
	/** The preferences module name. */
	public static final String		MODULE				= "GURPSCharacter";																											//$NON-NLS-1$
	/** The prefix used in front of all IDs for profile. */
	public static final String		PROFILE_PREFIX		= GURPSCharacter.CHARACTER_PREFIX + "pi.";																						//$NON-NLS-1$
	/** The field ID for portrait changes. */
	public static final String		ID_PORTRAIT			= PROFILE_PREFIX + "Portrait";																									//$NON-NLS-1$
	/** The field ID for name changes. */
	public static final String		ID_NAME				= PROFILE_PREFIX + "Name";																										//$NON-NLS-1$
	/** The field ID for notes changes. */
	public static final String		ID_NOTES			= PROFILE_PREFIX + "Notes";																									//$NON-NLS-1$
	/** The field ID for title changes. */
	public static final String		ID_TITLE			= PROFILE_PREFIX + "Title";																									//$NON-NLS-1$
	/** The field ID for age changes. */
	public static final String		ID_AGE				= PROFILE_PREFIX + "Age";																										//$NON-NLS-1$
	/** The field ID for birthday changes. */
	public static final String		ID_BIRTHDAY			= PROFILE_PREFIX + "Birthday";																									//$NON-NLS-1$
	/** The field ID for eye color changes. */
	public static final String		ID_EYE_COLOR		= PROFILE_PREFIX + "EyeColor";																									//$NON-NLS-1$
	/** The field ID for hair color changes. */
	public static final String		ID_HAIR				= PROFILE_PREFIX + "Hair";																										//$NON-NLS-1$
	/** The field ID for skin color changes. */
	public static final String		ID_SKIN_COLOR		= PROFILE_PREFIX + "SkinColor";																								//$NON-NLS-1$
	/** The field ID for handedness changes. */
	public static final String		ID_HANDEDNESS		= PROFILE_PREFIX + "Handedness";																								//$NON-NLS-1$
	/** The field ID for height changes. */
	public static final String		ID_HEIGHT			= PROFILE_PREFIX + "Height";																									//$NON-NLS-1$
	/** The field ID for weight changes. */
	public static final String		ID_WEIGHT			= PROFILE_PREFIX + "Weight";																									//$NON-NLS-1$
	/** The field ID for gender changes. */
	public static final String		ID_GENDER			= PROFILE_PREFIX + "Gender";																									//$NON-NLS-1$
	/** The field ID for race changes. */
	public static final String		ID_RACE				= PROFILE_PREFIX + "Race";																										//$NON-NLS-1$
	/** The field ID for religion changes. */
	public static final String		ID_RELIGION			= PROFILE_PREFIX + "Religion";																									//$NON-NLS-1$
	/** The field ID for player name changes. */
	public static final String		ID_PLAYER_NAME		= PROFILE_PREFIX + "PlayerName";																								//$NON-NLS-1$
	/** The field ID for campaign changes. */
	public static final String		ID_CAMPAIGN			= PROFILE_PREFIX + "Campaign";																									//$NON-NLS-1$
	/** The field ID for tech level changes. */
	public static final String		ID_TECH_LEVEL		= PROFILE_PREFIX + "TechLevel";																								//$NON-NLS-1$
	/** The field ID for size modifier changes. */
	public static final String		ID_SIZE_MODIFIER	= PROFILE_PREFIX + BonusAttributeType.SM.name();
	/** The default portrait marker. */
	public static final String		DEFAULT_PORTRAIT	= "!\000";																														//$NON-NLS-1$
	/** The default Tech Level. */
	public static final String		DEFAULT_TECH_LEVEL	= "4";																															//$NON-NLS-1$
	/** The height, in 1/72nds of an inch, of the portrait. */
	public static final int			PORTRAIT_HEIGHT		= 96;
	/** The width, in 1/72nds of an inch, of the portrait. */
	public static final int			PORTRAIT_WIDTH		= 3 * PORTRAIT_HEIGHT / 4;
	private static final String		TAG_PLAYER_NAME		= "player_name";																												//$NON-NLS-1$
	private static final String		TAG_CAMPAIGN		= "campaign";																													//$NON-NLS-1$
	private static final String		TAG_NAME			= "name";																														//$NON-NLS-1$
	private static final String		TAG_TITLE			= "title";																														//$NON-NLS-1$
	private static final String		TAG_AGE				= "age";																														//$NON-NLS-1$
	private static final String		TAG_BIRTHDAY		= "birthday";																													//$NON-NLS-1$
	private static final String		TAG_EYES			= "eyes";																														//$NON-NLS-1$
	private static final String		TAG_HAIR			= "hair";																														//$NON-NLS-1$
	private static final String		TAG_SKIN			= "skin";																														//$NON-NLS-1$
	private static final String		TAG_HANDEDNESS		= "handedness";																												//$NON-NLS-1$
	private static final String		TAG_HEIGHT			= "height";																													//$NON-NLS-1$
	private static final String		ATTRIBUTE_UNITS		= "units";																														//$NON-NLS-1$
	private static final String		TAG_WEIGHT			= "weight";																													//$NON-NLS-1$
	private static final String		TAG_GENDER			= "gender";																													//$NON-NLS-1$
	private static final String		TAG_RACE			= "race";																														//$NON-NLS-1$
	private static final String		TAG_TECH_LEVEL		= "tech_level";																												//$NON-NLS-1$
	private static final String		TAG_RELIGION		= "religion";																													//$NON-NLS-1$
	private static final String		TAG_PORTRAIT		= "portrait";																													//$NON-NLS-1$
	private static final String		TAG_NOTES			= "notes";																														//$NON-NLS-1$
	private static final String		EMPTY				= "";																															//$NON-NLS-1$
	private static final Random		RANDOM				= new Random();
	private static final String[]	EYE_OPTIONS			= new String[] { MSG_BROWN, MSG_BROWN, MSG_BLUE, MSG_BLUE, MSG_GREEN, MSG_GREY, MSG_VIOLET };
	private static final String[]	SKIN_OPTIONS		= new String[] { MSG_FRECKLED, MSG_TAN, MSG_LIGHT_TAN, MSG_DARK_TAN, MSG_BROWN, MSG_LIGHT_BROWN, MSG_DARK_BROWN, MSG_PALE };
	private static final String[]	HANDEDNESS_OPTIONS	= new String[] { MSG_RIGHT, MSG_RIGHT, MSG_RIGHT, MSG_LEFT };
	private static final String[]	GENDER_OPTIONS		= new String[] { MSG_MALE, MSG_MALE, MSG_MALE, MSG_FEMALE };
	private static final String[]	HAIR_OPTIONS;
	private GURPSCharacter			mCharacter;
	private boolean					mCustomPortrait;
	private BufferedImage			mPortrait;
	private BufferedImage			mDisplayPortrait;
	private String					mName;
	private String					mTitle;
	private int						mAge;
	private String					mBirthday;
	private String					mEyeColor;
	private String					mHair;
	private String					mSkinColor;
	private String					mHandedness;
	private int						mHeight;
	private double					mWeight;
	private int						mSizeModifier;
	private int						mSizeModifierBonus;
	private String					mGender;
	private String					mRace;
	private String					mReligion;
	private String					mPlayerName;
	private String					mCampaign;
	private String					mTechLevel;
	private String					mNotes;

	Profile(GURPSCharacter character, boolean full) {
		mCharacter = character;
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
		mHeight = full ? getRandomHeight(mCharacter.getStrength(), getSizeModifier()) : 0;
		mWeight = full ? getRandomWeight(mCharacter.getStrength(), getSizeModifier(), 1.0) : 0.0;
		mGender = full ? getRandomGender() : EMPTY;
		mName = full ? USCensusNames.INSTANCE.getFullName(mGender == MSG_MALE) : EMPTY;
		mRace = full ? MSG_DEFAULT_RACE : EMPTY;
		mTechLevel = full ? getDefaultTechLevel() : EMPTY;
		mReligion = EMPTY;
		mPlayerName = full ? getDefaultPlayerName() : EMPTY;
		mCampaign = full ? getDefaultCampaign() : EMPTY;
		mNotes = EMPTY;
		setPortraitInternal(getPortraitFromPortraitPath(getDefaultPortraitPath()));
	}

	void load(XMLReader reader) throws IOException {
		String marker = reader.getMarker();
		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String tag = reader.getName();
				if (!loadTag(reader, tag)) {
					reader.skipTag(tag);
				}
			}
		} while (reader.withinMarker(marker));
	}

	boolean loadTag(XMLReader reader, String tag) throws IOException {
		if (TAG_PLAYER_NAME.equals(tag)) {
			mPlayerName = reader.readText();
		} else if (TAG_CAMPAIGN.equals(tag)) {
			mCampaign = reader.readText();
		} else if (TAG_NAME.equals(tag)) {
			mName = reader.readText();
		} else if (TAG_NOTES.equals(tag)) {
			mNotes = TextUtility.standardizeLineEndings(reader.readText());
		} else if (TAG_TITLE.equals(tag)) {
			mTitle = reader.readText();
		} else if (TAG_AGE.equals(tag)) {
			mAge = reader.readInteger(0);
		} else if (TAG_BIRTHDAY.equals(tag)) {
			mBirthday = reader.readText();
		} else if (TAG_EYES.equals(tag)) {
			mEyeColor = reader.readText();
		} else if (TAG_HAIR.equals(tag)) {
			mHair = reader.readText();
		} else if (TAG_SKIN.equals(tag)) {
			mSkinColor = reader.readText();
		} else if (TAG_HANDEDNESS.equals(tag)) {
			mHandedness = reader.readText();
		} else if (TAG_HEIGHT.equals(tag)) {
			LengthUnits units = (LengthUnits) EnumExtractor.extract(reader.getAttribute(ATTRIBUTE_UNITS), LengthUnits.values());
			if (units == null) {
				// Old output didn't include a units attribute, as it was in a mixed
				// feet/inches format.
				mHeight = NumberUtils.getHeight(reader.readText());
			} else {
				mHeight = (int) LengthUnits.INCHES.convert(units, reader.readDouble(0));
			}
		} else if (TAG_WEIGHT.equals(tag)) {
			mWeight = WeightUnits.POUNDS.convert((WeightUnits) EnumExtractor.extract(reader.getAttribute(ATTRIBUTE_UNITS), WeightUnits.values(), WeightUnits.POUNDS), reader.readDouble(0));
		} else if (BonusAttributeType.SM.getXMLTag().equals(tag) || "size_modifier".equals(tag)) { //$NON-NLS-1$
			mSizeModifier = reader.readInteger(0);
		} else if (TAG_GENDER.equals(tag)) {
			mGender = reader.readText();
		} else if (TAG_RACE.equals(tag)) {
			mRace = reader.readText();
		} else if (TAG_TECH_LEVEL.equals(tag)) {
			mTechLevel = reader.readText();
		} else if (TAG_RELIGION.equals(tag)) {
			mReligion = reader.readText();
		} else if (TAG_PORTRAIT.equals(tag)) {
			try {
				setPortraitInternal(Images.loadImage(Base64.decode(reader.readText())));
				mCustomPortrait = true;
			} catch (Exception imageException) {
				// Ignore
			}
		} else {
			return false;
		}
		return true;
	}

	void save(XMLWriter out) {
		out.startSimpleTagEOL(TAG_ROOT);
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
		out.simpleTagWithAttribute(TAG_HEIGHT, Integer.toString(mHeight), ATTRIBUTE_UNITS, LengthUnits.INCHES.toString());
		out.simpleTagWithAttribute(TAG_WEIGHT, Double.toString(mWeight), ATTRIBUTE_UNITS, WeightUnits.POUNDS.toString());
		out.simpleTag(BonusAttributeType.SM.getXMLTag(), mSizeModifier);
		out.simpleTagNotEmpty(TAG_GENDER, mGender);
		out.simpleTagNotEmpty(TAG_RACE, mRace);
		out.simpleTagNotEmpty(TAG_TECH_LEVEL, mTechLevel);
		out.simpleTagNotEmpty(TAG_RELIGION, mReligion);
		out.simpleTagNotEmpty(TAG_NOTES, mNotes);
		if (mCustomPortrait && mPortrait != null) {
			try {
				ByteArrayOutputStream baos = new ByteArrayOutputStream();

				ImageIO.write(mPortrait, "png", baos); //$NON-NLS-1$
				baos.close();
				out.writeComment(MSG_PORTRAIT_COMMENT);
				out.startSimpleTagEOL(TAG_PORTRAIT);
				out.println(Base64.encode(baos.toByteArray()));
				out.endTagEOL(TAG_PORTRAIT, true);
			} catch (Exception ex) {
				throw new RuntimeException(MSG_PORTRAIT_WRITE_ERROR);
			}
		}
		out.endTagEOL(TAG_ROOT, true);
	}

	void update() {
		setSizeModifierBonus(mCharacter.getIntegerBonusFor(GURPSCharacter.ATTRIBUTES_PREFIX + BonusAttributeType.SM.name()));
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
			mCharacter.postUndoEdit(MSG_PORTRAIT_UNDO, ID_PORTRAIT, mPortrait, portrait);
			setPortraitInternal(portrait);
			mCharacter.notifySingle(ID_PORTRAIT, mPortrait);
		}
	}

	private void setPortraitInternal(BufferedImage portrait) {
		if (portrait == null) {
			mPortrait = null;
			mDisplayPortrait = null;
		} else {
			if (portrait.getWidth() != PORTRAIT_WIDTH * 2 || portrait.getHeight() != PORTRAIT_HEIGHT * 2) {
				portrait = Images.scale(portrait, PORTRAIT_WIDTH * 2, PORTRAIT_HEIGHT * 2);
			}
			mPortrait = portrait;
			mDisplayPortrait = Images.scale(mPortrait, PORTRAIT_WIDTH, PORTRAIT_HEIGHT);
		}
	}

	/** @return The name. */
	public String getName() {
		return mName;
	}

	/**
	 * Sets the name.
	 * 
	 * @param name The new name.
	 */
	public void setName(String name) {
		if (!mName.equals(name)) {
			mCharacter.postUndoEdit(MSG_NAME_UNDO, ID_NAME, mName, name);
			mName = name;
			mCharacter.notifySingle(ID_NAME, mName);
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
			mCharacter.postUndoEdit(MSG_GENDER_UNDO, ID_GENDER, mGender, gender);
			mGender = gender;
			mCharacter.notifySingle(ID_GENDER, mGender);
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
			mCharacter.postUndoEdit(MSG_RACE_UNDO, ID_RACE, mRace, race);
			mRace = race;
			mCharacter.notifySingle(ID_RACE, mRace);
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
			mCharacter.postUndoEdit(MSG_RELIGION_UNDO, ID_RELIGION, mReligion, religion);
			mReligion = religion;
			mCharacter.notifySingle(ID_RELIGION, mReligion);
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
			mCharacter.postUndoEdit(MSG_PLAYER_NAME_UNDO, ID_PLAYER_NAME, mPlayerName, player);
			mPlayerName = player;
			mCharacter.notifySingle(ID_PLAYER_NAME, mPlayerName);
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
			mCharacter.postUndoEdit(MSG_CAMPAIGN_UNDO, ID_CAMPAIGN, mCampaign, campaign);
			mCampaign = campaign;
			mCharacter.notifySingle(ID_CAMPAIGN, mCampaign);
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
			mCharacter.postUndoEdit(MSG_TECH_LEVEL_UNDO, ID_TECH_LEVEL, mTechLevel, techLevel);
			mTechLevel = techLevel;
			mCharacter.notifySingle(ID_TECH_LEVEL, mTechLevel);
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
			mCharacter.postUndoEdit(MSG_TITLE_UNDO, ID_TITLE, mTitle, title);
			mTitle = title;
			mCharacter.notifySingle(ID_TITLE, mTitle);
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

			mCharacter.postUndoEdit(MSG_AGE_UNDO, ID_AGE, new Integer(mAge), value);
			mAge = age;
			mCharacter.notifySingle(ID_AGE, value);
		}
	}

	/** @return A random age. */
	public int getRandomAge() {
		Advantage lifespan = mCharacter.getAdvantageNamed("Unaging"); //$NON-NLS-1$
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

		lifespan = mCharacter.getAdvantageNamed("Short Lifespan"); //$NON-NLS-1$
		if (lifespan != null) {
			levels = lifespan.getLevels();
			base = base >> levels;
			mod = mod >> levels;
		} else {
			lifespan = mCharacter.getAdvantageNamed("Extended Lifespan"); //$NON-NLS-1$
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
			mCharacter.postUndoEdit(MSG_BIRTHDAY_UNDO, ID_BIRTHDAY, mBirthday, birthday);
			mBirthday = birthday;
			mCharacter.notifySingle(ID_BIRTHDAY, mBirthday);
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
			mCharacter.postUndoEdit(MSG_EYE_COLOR_UNDO, ID_EYE_COLOR, mEyeColor, eyeColor);
			mEyeColor = eyeColor;
			mCharacter.notifySingle(ID_EYE_COLOR, mEyeColor);
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
			mCharacter.postUndoEdit(MSG_HAIR_UNDO, ID_HAIR, mHair, hair);
			mHair = hair;
			mCharacter.notifySingle(ID_HAIR, mHair);
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
			mCharacter.postUndoEdit(MSG_SKIN_COLOR_UNDO, ID_SKIN_COLOR, mSkinColor, skinColor);
			mSkinColor = skinColor;
			mCharacter.notifySingle(ID_SKIN_COLOR, mSkinColor);
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
			mCharacter.postUndoEdit(MSG_HANDEDNESS_UNDO, ID_HANDEDNESS, mHandedness, handedness);
			mHandedness = handedness;
			mCharacter.notifySingle(ID_HANDEDNESS, mHandedness);
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

			mCharacter.postUndoEdit(MSG_HEIGHT_UNDO, ID_HEIGHT, new Integer(mHeight), value);
			mHeight = height;
			mCharacter.notifySingle(ID_HEIGHT, value);
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

			mCharacter.postUndoEdit(MSG_WEIGHT_UNDO, ID_WEIGHT, new Double(mWeight), value);
			mWeight = weight;
			mCharacter.notifySingle(ID_WEIGHT, value);
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
			mCharacter.postUndoEdit(MSG_NOTES_UNDO, ID_NOTES, mNotes, notes);
			mNotes = notes;
			mCharacter.notifySingle(ID_NOTES, mNotes);
		}
	}

	/** @return The multiplier compared to average weight for this character. */
	public double getWeightMultiplier() {
		if (mCharacter.hasAdvantageNamed("Very Fat")) { //$NON-NLS-1$
			return 2.0;
		} else if (mCharacter.hasAdvantageNamed("Fat")) { //$NON-NLS-1$
			return 1.5;
		} else if (mCharacter.hasAdvantageNamed("Overweight")) { //$NON-NLS-1$
			return 1.3;
		} else if (mCharacter.hasAdvantageNamed("Skinny")) { //$NON-NLS-1$
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

			mCharacter.postUndoEdit(MSG_SIZE_MODIFIER_UNDO, ID_SIZE_MODIFIER, new Integer(totalSizeModifier), value);
			mSizeModifier = size - mSizeModifierBonus;
			mCharacter.notifySingle(ID_SIZE_MODIFIER, value);
		}
	}

	/** @param bonus The new size modifier bonus. */
	public void setSizeModifierBonus(int bonus) {
		if (mSizeModifierBonus != bonus) {
			mSizeModifierBonus = bonus;
			mCharacter.notifySingle(ID_SIZE_MODIFIER, new Integer(getSizeModifier()));
		}
	}

	/**
	 * @param id The field ID to retrieve the data for.
	 * @return The value of the specified field ID, or <code>null</code> if the field ID is
	 *         invalid.
	 */
	public Object getValueForID(String id) {
		if (id != null && id.startsWith(PROFILE_PREFIX)) {
			if (ID_NAME.equals(id)) {
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
			} else if (ID_TECH_LEVEL.equals(id)) {
				return getTechLevel();
			} else if (ID_SIZE_MODIFIER.equals(id)) {
				return new Integer(getSizeModifier());
			}
		}
		return null;
	}

	/**
	 * @param id The field ID to set the value for.
	 * @param value The value to set.
	 */
	public void setValueForID(String id, Object value) {
		if (id != null && id.startsWith(PROFILE_PREFIX)) {
			if (ID_NAME.equals(id)) {
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
			} else if (ID_TECH_LEVEL.equals(id)) {
				setTechLevel((String) value);
			} else if (ID_PORTRAIT.equals(id)) {
				if (value instanceof BufferedImage) {
					setPortrait((BufferedImage) value);
				}
			} else if (ID_SIZE_MODIFIER.equals(id)) {
				setSizeModifier(((Integer) value).intValue());
			}
		}
	}

	static {
		ArrayList<String> hair = new ArrayList<String>(100);
		String[] colors = { MSG_BROWN, MSG_BROWN, MSG_BROWN, MSG_BLACK, MSG_BLACK, MSG_BLACK, MSG_BLOND, MSG_BLOND, MSG_REDHEAD };
		String[] styles = { MSG_STRAIGHT, MSG_CURLY, MSG_WAVY };
		String[] lengths = { MSG_SHORT, MSG_MEDIUM, MSG_LONG };

		for (String element : colors) {
			for (String style : styles) {
				for (String length : lengths) {
					hair.add(MessageFormat.format(MSG_HAIR_FORMAT, element, style, length));
				}
			}
		}
		hair.add(MSG_BALD);
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
	 * @param sm The size modifier to use.
	 * @return A random height, in inches.
	 */
	public static int getRandomHeight(int strength, int sm) {
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
		base += RANDOM.nextInt(11);
		if (sm != 0) {
			base = (int) Math.max(Math.round(base * Math.pow(10.0, sm / 6.0)), 1);
		}
		return base;
	}

	/**
	 * @param strength The strength to base the weight on.
	 * @param sm The size modifier to use.
	 * @param multiplier The weight multiplier for being under- or overweight.
	 * @return A random weight, in pounds.
	 */
	public static int getRandomWeight(int strength, int sm, double multiplier) {
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
		base += RANDOM.nextInt(range);
		if (sm != 0) {
			base = (int) Math.round(base * Math.pow(1000.0, sm / 6.0));
		}
		return (int) Math.max(Math.round(base * multiplier), 1);
	}

	/** @return The default player name. */
	public static String getDefaultPlayerName() {
		return Preferences.getInstance().getStringValue(MODULE, Profile.ID_NAME, System.getProperty("user.name")); //$NON-NLS-1$
	}

	/** @param name The default player name. */
	public static void setDefaultPlayerName(String name) {
		Preferences.getInstance().setValue(MODULE, Profile.ID_NAME, name);
	}

	/** @return The default campaign value. */
	public static String getDefaultCampaign() {
		return Preferences.getInstance().getStringValue(MODULE, Profile.ID_CAMPAIGN, EMPTY);
	}

	/** @param campaign The default campaign value. */
	public static void setDefaultCampaign(String campaign) {
		Preferences.getInstance().setValue(MODULE, Profile.ID_CAMPAIGN, campaign);
	}

	/** @return The default tech level. */
	public static String getDefaultTechLevel() {
		return Preferences.getInstance().getStringValue(MODULE, Profile.ID_TECH_LEVEL, DEFAULT_TECH_LEVEL);
	}

	/** @param techLevel The default tech level. */
	public static void setDefaultTechLevel(String techLevel) {
		Preferences.getInstance().setValue(MODULE, Profile.ID_TECH_LEVEL, techLevel);
	}

	/**
	 * @param path The path to load.
	 * @return The portrait.
	 */
	public static BufferedImage getPortraitFromPortraitPath(String path) {
		if (DEFAULT_PORTRAIT.equals(path)) {
			return Images.getDefaultPortrait();
		}
		return path != null ? Images.loadImage(new File(path)) : null;
	}

	/** @return The default portrait path. */
	public static String getDefaultPortraitPath() {
		return Preferences.getInstance().getStringValue(MODULE, Profile.ID_PORTRAIT, DEFAULT_PORTRAIT);
	}

	/** @param path The default portrait path. */
	public static void setDefaultPortraitPath(String path) {
		Preferences.getInstance().setValue(MODULE, Profile.ID_PORTRAIT, path);
	}
}
