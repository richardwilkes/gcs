/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.character.names.USCensusNames;
import com.trollworks.gcs.feature.BonusAttributeType;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.text.TextUtility;
import com.trollworks.toolkit.utility.units.LengthUnits;
import com.trollworks.toolkit.utility.units.LengthValue;
import com.trollworks.toolkit.utility.units.WeightUnits;
import com.trollworks.toolkit.utility.units.WeightValue;

import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.IOException;
import java.text.MessageFormat;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Base64;
import java.util.Date;
import java.util.Random;

import javax.imageio.ImageIO;

/** Holds the character profile. */
public class Profile {
	@Localize("{0}, {1}, {2}")
	@Localize(locale = "de", value = "{0}, {1}, {2}")
	private static String			HAIR_FORMAT;
	@Localize("MMMM d")
	@Localize(locale = "de", value = "d. MMM.")
	private static String			BIRTHDAY_FORMAT;
	@Localize("Brown")
	@Localize(locale = "de", value = "Braun")
	@Localize(locale = "ru", value = "Коричнев(ые,ая)")
	private static String			BROWN;
	@Localize("Black")
	@Localize(locale = "de", value = "Schwarz")
	@Localize(locale = "ru", value = "Чёрный")
	private static String			BLACK;
	@Localize("Blond")
	@Localize(locale = "de", value = "Blond")
	@Localize(locale = "ru", value = "Белокурые")
	private static String			BLOND;
	@Localize("Redhead")
	@Localize(locale = "de", value = "Rot")
	@Localize(locale = "ru", value = "Рыжые")
	private static String			REDHEAD;
	@Localize("Bald")
	@Localize(locale = "de", value = "Glatze")
	@Localize(locale = "ru", value = "Лысый")
	private static String			BALD;
	@Localize("Straight")
	@Localize(locale = "de", value = "glatt")
	@Localize(locale = "ru", value = "Прямые")
	private static String			STRAIGHT;
	@Localize("Curly")
	@Localize(locale = "de", value = "lockig")
	@Localize(locale = "ru", value = "Вьющиеся")
	private static String			CURLY;
	@Localize("Wavy")
	@Localize(locale = "de", value = "gewellt")
	@Localize(locale = "ru", value = "Волнистые")
	private static String			WAVY;
	@Localize("Short")
	@Localize(locale = "de", value = "kurz")
	@Localize(locale = "ru", value = "Короткие")
	private static String			SHORT;
	@Localize("Medium")
	@Localize(locale = "de", value = "mittel")
	@Localize(locale = "ru", value = "Средние")
	private static String			MEDIUM;
	@Localize("Long")
	@Localize(locale = "de", value = "lang")
	@Localize(locale = "ru", value = "Длинные")
	private static String			LONG;
	@Localize("Blue")
	@Localize(locale = "de", value = "Blau")
	@Localize(locale = "ru", value = "Синие")
	private static String			BLUE;
	@Localize("Green")
	@Localize(locale = "de", value = "Grün")
	@Localize(locale = "ru", value = "Зелёные")
	private static String			GREEN;
	@Localize("Grey")
	@Localize(locale = "de", value = "Grau")
	@Localize(locale = "ru", value = "Серые")
	private static String			GREY;
	@Localize("Violet")
	@Localize(locale = "de", value = "Violett")
	@Localize(locale = "ru", value = "Фиолетовые")
	private static String			VIOLET;
	@Localize("Freckled")
	@Localize(locale = "de", value = "Sommersprossen")
	@Localize(locale = "ru", value = "Веснушчатая")
	private static String			FRECKLED;
	@Localize("Tan")
	@Localize(locale = "de", value = "Gebräunt")
	@Localize(locale = "ru", value = "Загорелая")
	private static String			TAN;
	@Localize("Light Tan")
	@Localize(locale = "de", value = "Leicht gebräunt")
	@Localize(locale = "ru", value = "Слегка загорелая")
	private static String			LIGHT_TAN;
	@Localize("Dark Tan")
	@Localize(locale = "de", value = "Stark gebräunt")
	@Localize(locale = "ru", value = "Сильно-загорелая")
	private static String			DARK_TAN;
	@Localize("Light Brown")
	@Localize(locale = "de", value = "Hellbraun")
	@Localize(locale = "ru", value = "Светло-коричневая")
	private static String			LIGHT_BROWN;
	@Localize("Dark Brown")
	@Localize(locale = "de", value = "Dunkelbraun")
	@Localize(locale = "ru", value = "Тёмно-коричневая")
	private static String			DARK_BROWN;
	@Localize("Pale")
	@Localize(locale = "de", value = "Bleich")
	@Localize(locale = "ru", value = "Бледная")
	private static String			PALE;
	@Localize("Right")
	@Localize(locale = "de", value = "Rechts")
	@Localize(locale = "ru", value = "Правая")
	private static String			RIGHT;
	@Localize("Left")
	@Localize(locale = "de", value = "Links")
	@Localize(locale = "ru", value = "Левая")
	private static String			LEFT;
	@Localize("Male")
	@Localize(locale = "de", value = "Männlich")
	@Localize(locale = "ru", value = "муж.")
	private static String			MALE;
	@Localize("Female")
	@Localize(locale = "de", value = "Weiblich")
	@Localize(locale = "ru", value = "жен.")
	private static String			FEMALE;
	@Localize("Human")
	@Localize(locale = "de", value = "Mensch")
	@Localize(locale = "ru", value = "Человек")
	private static String			DEFAULT_RACE;
	@Localize("Name Change")
	@Localize(locale = "de", value = "Namen ändern")
	@Localize(locale = "ru", value = "Изменить имя")
	private static String			NAME_UNDO;
	@Localize("Title Change")
	@Localize(locale = "de", value = "Titel ändern")
	@Localize(locale = "ru", value = "Изменить статус")
	private static String			TITLE_UNDO;
	@Localize("Age Change")
	@Localize(locale = "de", value = "Alter ändern")
	@Localize(locale = "ru", value = "Изменить возраст")
	private static String			AGE_UNDO;
	@Localize("Birthday Change")
	@Localize(locale = "de", value = "Geburtstag ändern")
	@Localize(locale = "ru", value = "Смена дня рождения")
	private static String			BIRTHDAY_UNDO;
	@Localize("Eye Color Change")
	@Localize(locale = "de", value = "Augenfarbe ändern")
	@Localize(locale = "ru", value = "Изменить цвет глаз")
	private static String			EYE_COLOR_UNDO;
	@Localize("Hair Change")
	@Localize(locale = "de", value = "Haar ändern")
	@Localize(locale = "ru", value = "Изменить причёску")
	private static String			HAIR_UNDO;
	@Localize("Skin Color Change")
	@Localize(locale = "de", value = "Hautfarbe ändern")
	@Localize(locale = "ru", value = "Изменить цвет кожи")
	private static String			SKIN_COLOR_UNDO;
	@Localize("Handedness Change")
	@Localize(locale = "de", value = "Händigkeit ändern")
	@Localize(locale = "ru", value = "Смена основной руки")
	private static String			HANDEDNESS_UNDO;
	@Localize("Height Change")
	@Localize(locale = "de", value = "Höhe ändern")
	@Localize(locale = "ru", value = "Изменить рост")
	private static String			HEIGHT_UNDO;
	@Localize("Weight Change")
	@Localize(locale = "de", value = "Gewicht ändern")
	@Localize(locale = "ru", value = "Изменить вес")
	private static String			WEIGHT_UNDO;
	@Localize("Gender Change")
	@Localize(locale = "de", value = "Geschlecht ändern")
	@Localize(locale = "ru", value = "Изменить пол")
	private static String			GENDER_UNDO;
	@Localize("Race Change")
	@Localize(locale = "de", value = "Rasse ändern")
	@Localize(locale = "ru", value = "Смена расы")
	private static String			RACE_UNDO;
	@Localize("Religion Change")
	@Localize(locale = "de", value = "Religion ändern")
	@Localize(locale = "ru", value = "Изменить религию")
	private static String			RELIGION_UNDO;
	@Localize("Player Name Change")
	@Localize(locale = "de", value = "Spielernamen ändern")
	@Localize(locale = "ru", value = "Изменить имя игрока")
	private static String			PLAYER_NAME_UNDO;
	@Localize("Campaign Change")
	@Localize(locale = "de", value = "Kampagne ändern")
	@Localize(locale = "ru", value = "Изменить компанию")
	private static String			CAMPAIGN_UNDO;
	@Localize("Size Modifier Change")
	@Localize(locale = "de", value = "Größenmodifikator ändern")
	@Localize(locale = "ru", value = "Изменить модификатор размера")
	private static String			SIZE_MODIFIER_UNDO;
	@Localize("Tech Level Change")
	@Localize(locale = "de", value = "Techlevel ändern")
	@Localize(locale = "ru", value = "Смена технологического уровня")
	private static String			TECH_LEVEL_UNDO;
	@Localize("Portrait Change")
	@Localize(locale = "de", value = "Portrait ändern")
	@Localize(locale = "ru", value = "Смена изображения")
	private static String			PORTRAIT_UNDO;
	@Localize("The portrait is a PNG file encoded as Base64.")
	@Localize(locale = "de", value = "Das Portrait ist eine Base64-kodierte PNG-Datei.")
	@Localize(locale = "ru", value = "Изображение представляет собой файл PNG (закодирован в Base64).")
	private static String			PORTRAIT_COMMENT;
	@Localize("Could not write portrait.")
	@Localize(locale = "de", value = "Kann Portrait nicht schreiben.")
	@Localize(locale = "ru", value = "Не удалось записать изображение.")
	private static String			PORTRAIT_WRITE_ERROR;
	@Localize("Notes Change")
	@Localize(locale = "de", value = "Notizen ändern")
	@Localize(locale = "ru", value = "Смена заметки")
	private static String			NOTES_UNDO;

	static {
		Localization.initialize();
	}

	/** The root XML tag. */
	public static final String		TAG_ROOT			= "profile";																					//$NON-NLS-1$
	/** The preferences module name. */
	public static final String		MODULE				= "GURPSCharacter";																			//$NON-NLS-1$
	/** The prefix used in front of all IDs for profile. */
	public static final String		PROFILE_PREFIX		= GURPSCharacter.CHARACTER_PREFIX + "pi.";														//$NON-NLS-1$
	/** The field ID for portrait changes. */
	public static final String		ID_PORTRAIT			= PROFILE_PREFIX + "Portrait";																	//$NON-NLS-1$
	/** The field ID for name changes. */
	public static final String		ID_NAME				= PROFILE_PREFIX + "Name";																		//$NON-NLS-1$
	/** The field ID for notes changes. */
	public static final String		ID_NOTES			= PROFILE_PREFIX + "Notes";																	//$NON-NLS-1$
	/** The field ID for title changes. */
	public static final String		ID_TITLE			= PROFILE_PREFIX + "Title";																	//$NON-NLS-1$
	/** The field ID for age changes. */
	public static final String		ID_AGE				= PROFILE_PREFIX + "Age";																		//$NON-NLS-1$
	/** The field ID for birthday changes. */
	public static final String		ID_BIRTHDAY			= PROFILE_PREFIX + "Birthday";																	//$NON-NLS-1$
	/** The field ID for eye color changes. */
	public static final String		ID_EYE_COLOR		= PROFILE_PREFIX + "EyeColor";																	//$NON-NLS-1$
	/** The field ID for hair color changes. */
	public static final String		ID_HAIR				= PROFILE_PREFIX + "Hair";																		//$NON-NLS-1$
	/** The field ID for skin color changes. */
	public static final String		ID_SKIN_COLOR		= PROFILE_PREFIX + "SkinColor";																//$NON-NLS-1$
	/** The field ID for handedness changes. */
	public static final String		ID_HANDEDNESS		= PROFILE_PREFIX + "Handedness";																//$NON-NLS-1$
	/** The field ID for height changes. */
	public static final String		ID_HEIGHT			= PROFILE_PREFIX + "Height";																	//$NON-NLS-1$
	/** The field ID for weight changes. */
	public static final String		ID_WEIGHT			= PROFILE_PREFIX + "Weight";																	//$NON-NLS-1$
	/** The field ID for gender changes. */
	public static final String		ID_GENDER			= PROFILE_PREFIX + "Gender";																	//$NON-NLS-1$
	/** The field ID for race changes. */
	public static final String		ID_RACE				= PROFILE_PREFIX + "Race";																		//$NON-NLS-1$
	/** The field ID for religion changes. */
	public static final String		ID_RELIGION			= PROFILE_PREFIX + "Religion";																	//$NON-NLS-1$
	/** The field ID for player name changes. */
	public static final String		ID_PLAYER_NAME		= PROFILE_PREFIX + "PlayerName";																//$NON-NLS-1$
	/** The field ID for campaign changes. */
	public static final String		ID_CAMPAIGN			= PROFILE_PREFIX + "Campaign";																	//$NON-NLS-1$
	/** The field ID for tech level changes. */
	public static final String		ID_TECH_LEVEL		= PROFILE_PREFIX + "TechLevel";																//$NON-NLS-1$
	/** The field ID for size modifier changes. */
	public static final String		ID_SIZE_MODIFIER	= PROFILE_PREFIX + BonusAttributeType.SM.name();
	/** The default portrait marker. */
	public static final String		DEFAULT_PORTRAIT	= "!\000";																						//$NON-NLS-1$
	/** The default Tech Level. */
	public static final String		DEFAULT_TECH_LEVEL	= "4";																							//$NON-NLS-1$
	/** The height, in 1/72nds of an inch, of the portrait. */
	public static final int			PORTRAIT_HEIGHT		= 96;
	/** The width, in 1/72nds of an inch, of the portrait. */
	public static final int			PORTRAIT_WIDTH		= 3 * PORTRAIT_HEIGHT / 4;
	private static final String		TAG_PLAYER_NAME		= "player_name";																				//$NON-NLS-1$
	private static final String		TAG_CAMPAIGN		= "campaign";																					//$NON-NLS-1$
	private static final String		TAG_NAME			= "name";																						//$NON-NLS-1$
	private static final String		TAG_TITLE			= "title";																						//$NON-NLS-1$
	private static final String		TAG_AGE				= "age";																						//$NON-NLS-1$
	private static final String		TAG_BIRTHDAY		= "birthday";																					//$NON-NLS-1$
	private static final String		TAG_EYES			= "eyes";																						//$NON-NLS-1$
	private static final String		TAG_HAIR			= "hair";																						//$NON-NLS-1$
	private static final String		TAG_SKIN			= "skin";																						//$NON-NLS-1$
	private static final String		TAG_HANDEDNESS		= "handedness";																				//$NON-NLS-1$
	private static final String		TAG_HEIGHT			= "height";																					//$NON-NLS-1$
	private static final String		TAG_WEIGHT			= "weight";																					//$NON-NLS-1$
	private static final String		TAG_GENDER			= "gender";																					//$NON-NLS-1$
	private static final String		TAG_RACE			= "race";																						//$NON-NLS-1$
	private static final String		TAG_TECH_LEVEL		= "tech_level";																				//$NON-NLS-1$
	private static final String		TAG_RELIGION		= "religion";																					//$NON-NLS-1$
	private static final String		TAG_PORTRAIT		= "portrait";																					//$NON-NLS-1$
	private static final String		TAG_NOTES			= "notes";																						//$NON-NLS-1$
	private static final String		EMPTY				= "";																							//$NON-NLS-1$
	private static final Random		RANDOM				= new Random();
	private static final String[]	EYE_OPTIONS			= new String[] { BROWN, BROWN, BLUE, BLUE, GREEN, GREY, VIOLET };
	private static final String[]	SKIN_OPTIONS		= new String[] { FRECKLED, TAN, LIGHT_TAN, DARK_TAN, BROWN, LIGHT_BROWN, DARK_BROWN, PALE };
	private static final String[]	HANDEDNESS_OPTIONS	= new String[] { RIGHT, RIGHT, RIGHT, LEFT };
	private static final String[]	GENDER_OPTIONS		= new String[] { MALE, MALE, MALE, FEMALE };
	private static final String[]	HAIR_OPTIONS;
	private GURPSCharacter			mCharacter;
	private boolean					mCustomPortrait;
	private StdImage				mPortrait;
	private StdImage				mDisplayPortrait;
	private String					mName;
	private String					mTitle;
	private int						mAge;
	private String					mBirthday;
	private String					mEyeColor;
	private String					mHair;
	private String					mSkinColor;
	private String					mHandedness;
	private LengthValue				mHeight;
	private WeightValue				mWeight;
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
		mHeight = full ? getRandomHeight(mCharacter.getStrength(), getSizeModifier()) : new LengthValue(0, SheetPreferences.getLengthUnits());
		mWeight = full ? getRandomWeight(mCharacter.getStrength(), getSizeModifier(), 1.0) : new WeightValue(0, SheetPreferences.getWeightUnits());
		mGender = full ? getRandomGender() : EMPTY;
		mName = full && SheetPreferences.isNewCharacterAutoNamed() ? USCensusNames.INSTANCE.getFullName(mGender == MALE) : EMPTY;
		mRace = full ? DEFAULT_RACE : EMPTY;
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
			mHeight = LengthValue.extract(reader.readText(), false);
		} else if (TAG_WEIGHT.equals(tag)) {
			mWeight = WeightValue.extract(reader.readText(), false);
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
				setPortraitInternal(StdImage.loadImage(Base64.getMimeDecoder().decode(reader.readText())));
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
		if (mHeight.getNormalizedValue() != 0) {
			out.simpleTag(TAG_HEIGHT, mHeight.toString(false));
		}
		if (mWeight.getNormalizedValue() != 0) {
			out.simpleTag(TAG_WEIGHT, mWeight.toString(false));
		}
		out.simpleTag(BonusAttributeType.SM.getXMLTag(), mSizeModifier);
		out.simpleTagNotEmpty(TAG_GENDER, mGender);
		out.simpleTagNotEmpty(TAG_RACE, mRace);
		out.simpleTagNotEmpty(TAG_TECH_LEVEL, mTechLevel);
		out.simpleTagNotEmpty(TAG_RELIGION, mReligion);
		out.simpleTagNotEmpty(TAG_NOTES, mNotes);
		if (mCustomPortrait && mPortrait != null) {
			try {
				try (ByteArrayOutputStream baos = new ByteArrayOutputStream()) {
					ImageIO.write(mPortrait, "png", baos); //$NON-NLS-1$
					out.writeComment(PORTRAIT_COMMENT);
					out.startSimpleTagEOL(TAG_PORTRAIT);
					out.println(Base64.getMimeEncoder().encodeToString(baos.toByteArray()));
					out.endTagEOL(TAG_PORTRAIT, true);
				}
			} catch (Exception ex) {
				throw new RuntimeException(PORTRAIT_WRITE_ERROR);
			}
		}
		out.endTagEOL(TAG_ROOT, true);
	}

	void update() {
		setSizeModifierBonus(mCharacter.getIntegerBonusFor(GURPSCharacter.ATTRIBUTES_PREFIX + BonusAttributeType.SM.name()));
	}

	/**
	 * @param forPrinting Pass in <code>true</code> to retrieve the double resolution portrait (for
	 *            printing), or <code>false</code> to retrieve the normal resolution portrait (for
	 *            display).
	 * @return The portrait.
	 */
	public StdImage getPortrait(boolean forPrinting) {
		return forPrinting ? mPortrait : mDisplayPortrait;
	}

	/**
	 * Sets the portrait.
	 *
	 * @param portrait The new portrait.
	 */
	public void setPortrait(StdImage portrait) {
		if (mPortrait != portrait) {
			mCustomPortrait = true;
			mCharacter.postUndoEdit(PORTRAIT_UNDO, ID_PORTRAIT, mPortrait, portrait);
			setPortraitInternal(portrait);
			mCharacter.notifySingle(ID_PORTRAIT, mPortrait);
		}
	}

	private void setPortraitInternal(StdImage portrait) {
		if (portrait == null) {
			mPortrait = null;
			mDisplayPortrait = null;
		} else {
			if (portrait.getWidth() != PORTRAIT_WIDTH * 2 || portrait.getHeight() != PORTRAIT_HEIGHT * 2) {
				portrait = StdImage.scale(portrait, PORTRAIT_WIDTH * 2, PORTRAIT_HEIGHT * 2);
			}
			mPortrait = portrait;
			mDisplayPortrait = StdImage.scale(mPortrait, PORTRAIT_WIDTH, PORTRAIT_HEIGHT);
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
			mCharacter.postUndoEdit(NAME_UNDO, ID_NAME, mName, name);
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
			mCharacter.postUndoEdit(GENDER_UNDO, ID_GENDER, mGender, gender);
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
			mCharacter.postUndoEdit(RACE_UNDO, ID_RACE, mRace, race);
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
			mCharacter.postUndoEdit(RELIGION_UNDO, ID_RELIGION, mReligion, religion);
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
			mCharacter.postUndoEdit(PLAYER_NAME_UNDO, ID_PLAYER_NAME, mPlayerName, player);
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
			mCharacter.postUndoEdit(CAMPAIGN_UNDO, ID_CAMPAIGN, mCampaign, campaign);
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
			mCharacter.postUndoEdit(TECH_LEVEL_UNDO, ID_TECH_LEVEL, mTechLevel, techLevel);
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
			mCharacter.postUndoEdit(TITLE_UNDO, ID_TITLE, mTitle, title);
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

			mCharacter.postUndoEdit(AGE_UNDO, ID_AGE, new Integer(mAge), value);
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
			mCharacter.postUndoEdit(BIRTHDAY_UNDO, ID_BIRTHDAY, mBirthday, birthday);
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
			mCharacter.postUndoEdit(EYE_COLOR_UNDO, ID_EYE_COLOR, mEyeColor, eyeColor);
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
			mCharacter.postUndoEdit(HAIR_UNDO, ID_HAIR, mHair, hair);
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
			mCharacter.postUndoEdit(SKIN_COLOR_UNDO, ID_SKIN_COLOR, mSkinColor, skinColor);
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
			mCharacter.postUndoEdit(HANDEDNESS_UNDO, ID_HANDEDNESS, mHandedness, handedness);
			mHandedness = handedness;
			mCharacter.notifySingle(ID_HANDEDNESS, mHandedness);
		}
	}

	/** @return The height. */
	public LengthValue getHeight() {
		return mHeight;
	}

	/**
	 * Sets the height.
	 *
	 * @param height The new height.
	 */
	public void setHeight(LengthValue height) {
		if (!mHeight.equals(height)) {
			height = new LengthValue(height);
			mCharacter.postUndoEdit(HEIGHT_UNDO, ID_HEIGHT, new LengthValue(mHeight), height);
			mHeight = height;
			mCharacter.notifySingle(ID_HEIGHT, height);
		}
	}

	/** @return The weight. */
	public WeightValue getWeight() {
		return mWeight;
	}

	/**
	 * Sets the weight.
	 *
	 * @param weight The new weight.
	 */
	public void setWeight(WeightValue weight) {
		if (!mWeight.equals(weight)) {
			weight = new WeightValue(weight);
			mCharacter.postUndoEdit(WEIGHT_UNDO, ID_WEIGHT, new WeightValue(mWeight), weight);
			mWeight = weight;
			mCharacter.notifySingle(ID_WEIGHT, weight);
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
			mCharacter.postUndoEdit(NOTES_UNDO, ID_NOTES, mNotes, notes);
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

			mCharacter.postUndoEdit(SIZE_MODIFIER_UNDO, ID_SIZE_MODIFIER, new Integer(totalSizeModifier), value);
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
	 * @return The value of the specified field ID, or <code>null</code> if the field ID is invalid.
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
				return new LengthValue(getHeight());
			} else if (ID_WEIGHT.equals(id)) {
				return new WeightValue(getWeight());
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
				setHeight((LengthValue) value);
			} else if (ID_WEIGHT.equals(id)) {
				setWeight((WeightValue) value);
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
				if (value instanceof StdImage) {
					setPortrait((StdImage) value);
				}
			} else if (ID_SIZE_MODIFIER.equals(id)) {
				setSizeModifier(((Integer) value).intValue());
			}
		}
	}

	static {
		ArrayList<String> hair = new ArrayList<>(100);
		String[] colors = { BROWN, BROWN, BROWN, BLACK, BLACK, BLACK, BLOND, BLOND, REDHEAD };
		String[] styles = { STRAIGHT, CURLY, WAVY };
		String[] lengths = { SHORT, MEDIUM, LONG };

		for (String element : colors) {
			for (String style : styles) {
				for (String length : lengths) {
					hair.add(MessageFormat.format(HAIR_FORMAT, element, style, length));
				}
			}
		}
		hair.add(BALD);
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
		SimpleDateFormat formatter = new SimpleDateFormat(BIRTHDAY_FORMAT);
		return formatter.format(new Date(RANDOM.nextLong()));
	}

	/**
	 * @param strength The strength to base the height on.
	 * @param sm The size modifier to use.
	 * @return A random height.
	 */
	public static LengthValue getRandomHeight(int strength, int sm) {
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
		if (SheetPreferences.areGurpsMetricRulesUsed()) {
			base = (int) Math.round(LengthUnits.CM.convert(LengthUnits.FT_IN, base));
			base += RANDOM.nextInt(16);
		} else {
			base += RANDOM.nextInt(11);
		}
		if (sm != 0) {
			base = (int) Math.max(Math.round(base * Math.pow(10.0, sm / 6.0)), 1);
		}
		LengthUnits calcUnits = SheetPreferences.areGurpsMetricRulesUsed() ? LengthUnits.CM : LengthUnits.FT_IN;
		LengthUnits desiredUnits = SheetPreferences.getLengthUnits();
		return new LengthValue(desiredUnits.convert(calcUnits, base), desiredUnits);
	}

	/**
	 * @param strength The strength to base the weight on.
	 * @param sm The size modifier to use.
	 * @param multiplier The weight multiplier for being under- or overweight.
	 * @return A random weight.
	 */
	public static WeightValue getRandomWeight(int strength, int sm, double multiplier) {
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
		if (SheetPreferences.areGurpsMetricRulesUsed()) {
			base = (int) Math.round(WeightUnits.KG.convert(WeightUnits.LB, base));
			range = (int) Math.round(WeightUnits.KG.convert(WeightUnits.LB, range - 1)) + 1;
		}
		base += RANDOM.nextInt(range);
		if (sm != 0) {
			base = (int) Math.round(base * Math.pow(1000.0, sm / 6.0));
		}
		base = (int) Math.max(Math.round(base * multiplier), 1);
		WeightUnits calcUnits = SheetPreferences.areGurpsMetricRulesUsed() ? WeightUnits.KG : WeightUnits.LB;
		WeightUnits desiredUnits = SheetPreferences.getWeightUnits();
		return new WeightValue(desiredUnits.convert(calcUnits, base), desiredUnits);
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
	public static StdImage getPortraitFromPortraitPath(String path) {
		if (DEFAULT_PORTRAIT.equals(path)) {
			return GCSImages.getDefaultPortrait();
		}
		return path != null ? StdImage.loadImage(new File(path)) : null;
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
