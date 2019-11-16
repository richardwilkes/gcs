/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.character.names.USCensusNames;
import com.trollworks.gcs.feature.BonusAttributeType;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.preferences.DisplayPreferences;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.RetinaIcon;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.utility.FileType;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.text.Text;
import com.trollworks.toolkit.utility.units.LengthUnits;
import com.trollworks.toolkit.utility.units.LengthValue;
import com.trollworks.toolkit.utility.units.WeightUnits;
import com.trollworks.toolkit.utility.units.WeightValue;

import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.IOException;
import java.text.MessageFormat;
import java.text.SimpleDateFormat;
import java.util.Base64;
import java.util.Date;
import java.util.Random;

import javax.imageio.ImageIO;

/** Holds the character profile. */
public class Profile {
    /** The root XML tag. */
    public static final String  TAG_ROOT           = "profile";
    /** The preferences module name. */
    public static final String  MODULE             = "GURPSCharacter";
    /** The prefix used in front of all IDs for profile. */
    public static final String  PROFILE_PREFIX     = GURPSCharacter.CHARACTER_PREFIX + "pi.";
    /** The field ID for portrait changes. */
    public static final String  ID_PORTRAIT        = PROFILE_PREFIX + "Portrait";
    /** The field ID for name changes. */
    public static final String  ID_NAME            = PROFILE_PREFIX + "Name";
    /** The field ID for title changes. */
    public static final String  ID_TITLE           = PROFILE_PREFIX + "Title";
    /** The field ID for age changes. */
    public static final String  ID_AGE             = PROFILE_PREFIX + "Age";
    /** The field ID for birthday changes. */
    public static final String  ID_BIRTHDAY        = PROFILE_PREFIX + "Birthday";
    /** The field ID for eye color changes. */
    public static final String  ID_EYE_COLOR       = PROFILE_PREFIX + "EyeColor";
    /** The field ID for hair color changes. */
    public static final String  ID_HAIR            = PROFILE_PREFIX + "Hair";
    /** The field ID for skin color changes. */
    public static final String  ID_SKIN_COLOR      = PROFILE_PREFIX + "SkinColor";
    /** The field ID for handedness changes. */
    public static final String  ID_HANDEDNESS      = PROFILE_PREFIX + "Handedness";
    /** The field ID for height changes. */
    public static final String  ID_HEIGHT          = PROFILE_PREFIX + "Height";
    /** The field ID for weight changes. */
    public static final String  ID_WEIGHT          = PROFILE_PREFIX + "Weight";
    /** The field ID for gender changes. */
    public static final String  ID_GENDER          = PROFILE_PREFIX + "Gender";
    /** The field ID for race changes. */
    public static final String  ID_RACE            = PROFILE_PREFIX + "Race";
    /** The field ID for religion changes. */
    public static final String  ID_RELIGION        = PROFILE_PREFIX + "Religion";
    /** The field ID for player name changes. */
    public static final String  ID_PLAYER_NAME     = PROFILE_PREFIX + "PlayerName";
    /** The field ID for campaign changes. */
    public static final String  ID_CAMPAIGN        = PROFILE_PREFIX + "Campaign";
    /** The field ID for tech level changes. */
    public static final String  ID_TECH_LEVEL      = PROFILE_PREFIX + "TechLevel";
    /** The field ID for size modifier changes. */
    public static final String  ID_SIZE_MODIFIER   = PROFILE_PREFIX + BonusAttributeType.SM.name();
    /** The field ID for body type changes. */
    public static final String  ID_BODY_TYPE       = PROFILE_PREFIX + "BodyType";
    /** The default portrait marker. */
    public static final String  DEFAULT_PORTRAIT   = "!\000";
    /** The default Tech Level. */
    public static final String  DEFAULT_TECH_LEVEL = "3";
    /** The height, in 1/72nds of an inch, of the portrait. */
    public static final int     PORTRAIT_HEIGHT    = 96;
    /** The width, in 1/72nds of an inch, of the portrait. */
    public static final int     PORTRAIT_WIDTH     = 3 * PORTRAIT_HEIGHT / 4;
    private static final String TAG_PLAYER_NAME    = "player_name";
    private static final String TAG_CAMPAIGN       = "campaign";
    private static final String TAG_NAME           = "name";
    private static final String TAG_TITLE          = "title";
    private static final String TAG_AGE            = "age";
    private static final String TAG_BIRTHDAY       = "birthday";
    private static final String TAG_EYES           = "eyes";
    private static final String TAG_HAIR           = "hair";
    private static final String TAG_SKIN           = "skin";
    private static final String TAG_HANDEDNESS     = "handedness";
    private static final String TAG_HEIGHT         = "height";
    private static final String TAG_WEIGHT         = "weight";
    private static final String TAG_GENDER         = "gender";
    private static final String TAG_RACE           = "race";
    private static final String TAG_TECH_LEVEL     = "tech_level";
    private static final String TAG_RELIGION       = "religion";
    private static final String TAG_PORTRAIT       = "portrait";
    private static final String TAG_OLD_NOTES      = "notes";
    private static final String TAG_BODY_TYPE      = "body_type";
    private static final Random RANDOM             = new Random();
    private GURPSCharacter      mCharacter;
    private boolean             mCustomPortrait;
    private RetinaIcon          mPortrait;
    private String              mName;
    private String              mTitle;
    private int                 mAge;
    private String              mBirthday;
    private String              mEyeColor;
    private String              mHair;
    private String              mSkinColor;
    private String              mHandedness;
    private LengthValue         mHeight;
    private WeightValue         mWeight;
    private int                 mSizeModifier;
    private int                 mSizeModifierBonus;
    private String              mGender;
    private String              mRace;
    private String              mReligion;
    private String              mPlayerName;
    private String              mCampaign;
    private String              mTechLevel;
    private HitLocationTable    mHitLocationTable;

    Profile(GURPSCharacter character, boolean full) {
        mCharacter        = character;
        mCustomPortrait   = false;
        mPortrait         = null;
        mTitle            = "";
        mAge              = full ? getRandomAge() : 0;
        mBirthday         = full ? getRandomMonthAndDay() : "";
        mEyeColor         = full ? getRandomEyeColor() : "";
        mHair             = full ? getRandomHair() : "";
        mSkinColor        = full ? getRandomSkinColor() : "";
        mHandedness       = full ? getRandomHandedness() : "";
        mHeight           = full ? getRandomHeight(mCharacter.getStrength(), getSizeModifier()) : new LengthValue(0, DisplayPreferences.getLengthUnits());
        mWeight           = full ? getRandomWeight(mCharacter.getStrength(), getSizeModifier(), 1.0) : new WeightValue(0, DisplayPreferences.getWeightUnits());
        mGender           = full ? getRandomGender() : "";
        mName             = full && SheetPreferences.isNewCharacterAutoNamed() ? USCensusNames.INSTANCE.getFullName(mGender == I18n.Text("Male")) : "";
        mRace             = full ? I18n.Text("Human") : "";
        mTechLevel        = full ? getDefaultTechLevel() : "";
        mReligion         = "";
        mPlayerName       = full ? getDefaultPlayerName() : "";
        mCampaign         = full ? getDefaultCampaign() : "";
        mPortrait         = createPortrait(getPortraitFromPortraitPath(getDefaultPortraitPath()));
        mHitLocationTable = HitLocationTable.HUMANOID;
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
        } else if (TAG_OLD_NOTES.equals(tag)) {
            Note note = new Note(mCharacter, false);
            note.setDescription(Text.standardizeLineEndings(reader.readText()));
            mCharacter.getNotesRoot().addRow(note, false);
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
        } else if (BonusAttributeType.SM.getXMLTag().equals(tag) || "size_modifier".equals(tag)) {
            mSizeModifier = reader.readInteger(0);
        } else if (TAG_GENDER.equals(tag)) {
            mGender = reader.readText();
        } else if (TAG_RACE.equals(tag)) {
            mRace = reader.readText();
        } else if (TAG_BODY_TYPE.equals(tag)) {
            mHitLocationTable = HitLocationTable.MAP.get(reader.readText());
            if (mHitLocationTable == null) {
                mHitLocationTable = HitLocationTable.HUMANOID;
            }
        } else if (TAG_TECH_LEVEL.equals(tag)) {
            mTechLevel = reader.readText();
        } else if (TAG_RELIGION.equals(tag)) {
            mReligion = reader.readText();
        } else if (TAG_PORTRAIT.equals(tag)) {
            try {
                mPortrait       = createPortrait(StdImage.loadImage(Base64.getMimeDecoder().decode(reader.readText())));
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
        out.simpleTag(TAG_BODY_TYPE, mHitLocationTable.getKey());
        out.simpleTagNotEmpty(TAG_TECH_LEVEL, mTechLevel);
        out.simpleTagNotEmpty(TAG_RELIGION, mReligion);
        if (mCustomPortrait && mPortrait != null) {
            try {
                try (ByteArrayOutputStream baos = new ByteArrayOutputStream()) {
                    ImageIO.write(mPortrait.getRetina(), FileType.PNG_EXTENSION, baos);
                    out.writeComment("The portrait is a PNG file encoded as Base64");
                    out.startSimpleTagEOL(TAG_PORTRAIT);
                    out.println(Text.standardizeLineEndings(Base64.getMimeEncoder().encodeToString(baos.toByteArray())));
                    out.endTagEOL(TAG_PORTRAIT, true);
                }
            } catch (Exception ex) {
                throw new RuntimeException("Could not write portrait.");
            }
        }
        out.endTagEOL(TAG_ROOT, true);
    }

    void update() {
        setSizeModifierBonus(mCharacter.getIntegerBonusFor(GURPSCharacter.ATTRIBUTES_PREFIX + BonusAttributeType.SM.name()));
    }

    /** @return The portrait. */
    public RetinaIcon getPortrait() {
        return mPortrait;
    }

    /**
     * Sets the portrait.
     *
     * @param portrait The new portrait.
     */
    public void setPortrait(StdImage portrait) {
        if (portrait == null ? mPortrait != null : mPortrait.getRetina() != portrait) {
            mCustomPortrait = true;
            mCharacter.postUndoEdit(I18n.Text("Portrait Change"), ID_PORTRAIT, mPortrait, portrait);
            mPortrait = createPortrait(portrait);
            mCharacter.notifySingle(ID_PORTRAIT, mPortrait);
        }
    }

    public static RetinaIcon createPortrait(StdImage image) {
        if (image == null) {
            return null;
        }
        StdImage normal;
        StdImage retina;
        int      width  = image.getWidth();
        int      height = image.getHeight();
        if (width == PORTRAIT_WIDTH * 2 && height == PORTRAIT_HEIGHT * 2) {
            retina = image;
            normal = StdImage.scale(retina, PORTRAIT_WIDTH, PORTRAIT_HEIGHT);
        } else if (width == PORTRAIT_WIDTH && height == PORTRAIT_HEIGHT) {
            normal = image;
            retina = StdImage.scale(normal, PORTRAIT_WIDTH * 2, PORTRAIT_HEIGHT * 2);
        } else {
            retina = StdImage.scale(image, PORTRAIT_WIDTH * 2, PORTRAIT_HEIGHT * 2);
            normal = StdImage.scale(retina, PORTRAIT_WIDTH, PORTRAIT_HEIGHT);
        }
        return new RetinaIcon(normal, retina);
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
            mCharacter.postUndoEdit(I18n.Text("Name Change"), ID_NAME, mName, name);
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
            mCharacter.postUndoEdit(I18n.Text("Gender Change"), ID_GENDER, mGender, gender);
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
            mCharacter.postUndoEdit(I18n.Text("Race Change"), ID_RACE, mRace, race);
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
            mCharacter.postUndoEdit(I18n.Text("Religion Change"), ID_RELIGION, mReligion, religion);
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
            mCharacter.postUndoEdit(I18n.Text("Player Name Change"), ID_PLAYER_NAME, mPlayerName, player);
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
            mCharacter.postUndoEdit(I18n.Text("Campaign Change"), ID_CAMPAIGN, mCampaign, campaign);
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
            mCharacter.postUndoEdit(I18n.Text("Tech Level Change"), ID_TECH_LEVEL, mTechLevel, techLevel);
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
            mCharacter.postUndoEdit(I18n.Text("Title Change"), ID_TITLE, mTitle, title);
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
            Integer value = Integer.valueOf(age);

            mCharacter.postUndoEdit(I18n.Text("Age Change"), ID_AGE, Integer.valueOf(mAge), value);
            mAge = age;
            mCharacter.notifySingle(ID_AGE, value);
        }
    }

    /** @return A random age. */
    public int getRandomAge() {
        Advantage lifespan = mCharacter.getAdvantageNamed("Unaging");
        int       base     = 16;
        int       mod      = 7;
        int       levels;

        if (lifespan != null) {
            return 18 + RANDOM.nextInt(7);
        }

        if (RANDOM.nextInt(3) == 1) {
            mod += 7;
            if (RANDOM.nextInt(4) == 1) {
                mod += 13;
            }
        }

        lifespan = mCharacter.getAdvantageNamed("Short Lifespan");
        if (lifespan != null) {
            levels = lifespan.getLevels();
            base   = base >> levels;
            mod    = mod >> levels;
        } else {
            lifespan = mCharacter.getAdvantageNamed("Extended Lifespan");
            if (lifespan != null) {
                levels = lifespan.getLevels();
                base   = base << levels;
                mod    = mod << levels;
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
            mCharacter.postUndoEdit(I18n.Text("Birthday Change"), ID_BIRTHDAY, mBirthday, birthday);
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
            mCharacter.postUndoEdit(I18n.Text("Eye Color Change"), ID_EYE_COLOR, mEyeColor, eyeColor);
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
            mCharacter.postUndoEdit(I18n.Text("Hair Change"), ID_HAIR, mHair, hair);
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
            mCharacter.postUndoEdit(I18n.Text("Skin Color Change"), ID_SKIN_COLOR, mSkinColor, skinColor);
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
            mCharacter.postUndoEdit(I18n.Text("Handedness Change"), ID_HANDEDNESS, mHandedness, handedness);
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
            mCharacter.postUndoEdit(I18n.Text("Height Change"), ID_HEIGHT, new LengthValue(mHeight), height);
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
            mCharacter.postUndoEdit(I18n.Text("Weight Change"), ID_WEIGHT, new WeightValue(mWeight), weight);
            mWeight = weight;
            mCharacter.notifySingle(ID_WEIGHT, weight);
        }
    }

    /** @return The multiplier compared to average weight for this character. */
    public double getWeightMultiplier() {
        if (mCharacter.hasAdvantageNamed("Very Fat")) {
            return 2.0;
        } else if (mCharacter.hasAdvantageNamed("Fat")) {
            return 1.5;
        } else if (mCharacter.hasAdvantageNamed("Overweight")) {
            return 1.3;
        } else if (mCharacter.hasAdvantageNamed("Skinny")) {
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
            Integer value = Integer.valueOf(size);

            mCharacter.postUndoEdit(I18n.Text("Size Modifier Change"), ID_SIZE_MODIFIER, Integer.valueOf(totalSizeModifier), value);
            mSizeModifier = size - mSizeModifierBonus;
            mCharacter.notifySingle(ID_SIZE_MODIFIER, value);
        }
    }

    /** @param bonus The new size modifier bonus. */
    public void setSizeModifierBonus(int bonus) {
        if (mSizeModifierBonus != bonus) {
            mSizeModifierBonus = bonus;
            mCharacter.notifySingle(ID_SIZE_MODIFIER, Integer.valueOf(getSizeModifier()));
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
            } else if (ID_TITLE.equals(id)) {
                return getTitle();
            } else if (ID_AGE.equals(id)) {
                return Integer.valueOf(getAge());
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
                return Integer.valueOf(getSizeModifier());
            } else if (ID_BODY_TYPE.equals(id)) {
                return mHitLocationTable;
            }
        }
        return null;
    }

    /**
     * @param id    The field ID to set the value for.
     * @param value The value to set.
     */
    public void setValueForID(String id, Object value) {
        if (id != null && id.startsWith(PROFILE_PREFIX)) {
            if (ID_NAME.equals(id)) {
                setName((String) value);
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
            } else if (ID_BODY_TYPE.equals(id)) {
                setHitLocationTable((HitLocationTable) value);
            }
        }
    }

    /** @return A random hair color, style & length. */
    public static String getRandomHair() {
        if (RANDOM.nextInt(5) == 0) {
            return I18n.Text("Bald");
        }

        String color;
        switch (RANDOM.nextInt(9)) {
        case 0:
        case 1:
        case 2:
            color = I18n.Text("Black");
            break;
        case 3:
        case 4:
            color = I18n.Text("Blond");
            break;
        case 5:
            color = I18n.Text("Readhead");
            break;
        default:
            color = I18n.Text("Brown");
            break;
        }

        String style;
        switch (RANDOM.nextInt(3)) {
        case 0:
            style = I18n.Text("Curly");
            break;
        case 1:
            style = I18n.Text("Wavy");
            break;
        default:
            style = I18n.Text("Straight");
            break;
        }

        String length;
        switch (RANDOM.nextInt(3)) {
        case 0:
            length = I18n.Text("Short");
            break;
        case 1:
            length = I18n.Text("Long");
            break;
        default:
            length = I18n.Text("Medium");
            break;
        }

        return MessageFormat.format("{0}, {1}, {2}", color, style, length);
    }

    /** @return A random eye color. */
    public static String getRandomEyeColor() {
        switch (RANDOM.nextInt(8)) {
        case 0:
        case 1:
            return I18n.Text("Blue");
        case 2:
            return I18n.Text("Green");
        case 3:
            return I18n.Text("Grey");
        case 4:
            return I18n.Text("Violet");
        default:
            return I18n.Text("Brown");
        }
    }

    /** @return A random sking color. */
    public static String getRandomSkinColor() {
        switch (RANDOM.nextInt(8)) {
        case 0:
            return I18n.Text("Freckled");
        case 1:
            return I18n.Text("Light Tan");
        case 2:
            return I18n.Text("Dark Tan");
        case 3:
            return I18n.Text("Brown");
        case 4:
            return I18n.Text("Light Brown");
        case 5:
            return I18n.Text("Dark Brown");
        case 6:
            return I18n.Text("Pale");
        default:
            return I18n.Text("Tan");
        }
    }

    /** @return A random handedness. */
    public static String getRandomHandedness() {
        switch (RANDOM.nextInt(4)) {
        case 0:
            return I18n.Text("Left");
        default:
            return I18n.Text("Right");
        }
    }

    /** @return A random gender. */
    public static String getRandomGender() {
        switch (RANDOM.nextInt(2)) {
        case 0:
            return I18n.Text("Female");
        default:
            return I18n.Text("Male");
        }
    }

    /** @return A random month and day. */
    public static String getRandomMonthAndDay() {
        SimpleDateFormat formatter = new SimpleDateFormat(I18n.Text("MMMM d"));
        return formatter.format(new Date(RANDOM.nextLong()));
    }

    /**
     * @param strength The strength to base the height on.
     * @param sm       The size modifier to use.
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
        boolean useMetric = DisplayPreferences.getWeightUnits().isMetric();
        if (useMetric) {
            base  = (int) Math.round(LengthUnits.CM.convert(LengthUnits.FT_IN, base));
            base += RANDOM.nextInt(16);
        } else {
            base += RANDOM.nextInt(11);
        }
        if (sm != 0) {
            base = (int) Math.max(Math.round(base * Math.pow(10.0, sm / 6.0)), 1);
        }
        LengthUnits calcUnits    = useMetric ? LengthUnits.CM : LengthUnits.FT_IN;
        LengthUnits desiredUnits = DisplayPreferences.getLengthUnits();
        return new LengthValue(desiredUnits.convert(calcUnits, base), desiredUnits);
    }

    /**
     * @param strength   The strength to base the weight on.
     * @param sm         The size modifier to use.
     * @param multiplier The weight multiplier for being under- or overweight.
     * @return A random weight.
     */
    public static WeightValue getRandomWeight(int strength, int sm, double multiplier) {
        int base;
        int range;

        if (strength < 7) {
            base  = 60;
            range = 61;
        } else if (strength < 10) {
            base  = 75 + (strength - 7) * 15;
            range = 61;
        } else if (strength == 10) {
            base  = 115;
            range = 61;
        } else if (strength < 14) {
            base  = 125 + (strength - 11) * 15;
            range = 71 + (strength - 11) * 10;
        } else {
            base  = 170;
            range = 101;
        }
        boolean useMetric = DisplayPreferences.getWeightUnits().isMetric();
        if (useMetric) {
            base  = (int) Math.round(WeightUnits.KG.convert(WeightUnits.LB, base));
            range = (int) Math.round(WeightUnits.KG.convert(WeightUnits.LB, range - 1)) + 1;
        }
        base += RANDOM.nextInt(range);
        if (sm != 0) {
            base = (int) Math.round(base * Math.pow(1000.0, sm / 6.0));
        }
        base = (int) Math.max(Math.round(base * multiplier), 1);
        WeightUnits calcUnits    = useMetric ? WeightUnits.KG : WeightUnits.LB;
        WeightUnits desiredUnits = DisplayPreferences.getWeightUnits();
        return new WeightValue(desiredUnits.convert(calcUnits, base), desiredUnits);
    }

    /** @return The default player name. */
    public static String getDefaultPlayerName() {
        return Preferences.getInstance().getStringValue(MODULE, Profile.ID_NAME, System.getProperty("user.name"));
    }

    /** @param name The default player name. */
    public static void setDefaultPlayerName(String name) {
        Preferences.getInstance().setValue(MODULE, Profile.ID_NAME, name);
    }

    /** @return The default campaign value. */
    public static String getDefaultCampaign() {
        return Preferences.getInstance().getStringValue(MODULE, Profile.ID_CAMPAIGN, "");
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
            return StdImage.get("default_portrait");
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

    /** @return The hit location table. */
    public HitLocationTable getHitLocationTable() {
        return mHitLocationTable;
    }

    /** @param table The hit location table. */
    public void setHitLocationTable(HitLocationTable table) {
        if (mHitLocationTable != table) {
            mCharacter.postUndoEdit(I18n.Text("Body Type Change"), ID_BODY_TYPE, mHitLocationTable, table);
            mHitLocationTable = table;
            mCharacter.notifySingle(ID_BODY_TYPE, mHitLocationTable);
        }
    }
}
