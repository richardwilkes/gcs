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

import com.trollworks.gcs.ancestry.Ancestry;
import com.trollworks.gcs.attribute.Attribute;
import com.trollworks.gcs.body.HitLocationTable;
import com.trollworks.gcs.calendar.Calendar;
import com.trollworks.gcs.calendar.CalendarException;
import com.trollworks.gcs.calendar.Date;
import com.trollworks.gcs.settings.CalendarRef;
import com.trollworks.gcs.settings.GeneralSettings;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.settings.SheetSettings;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.NamedData;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.LengthValue;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.utility.units.WeightValue;
import static java.time.format.TextStyle.FULL;
import static java.time.temporal.ChronoField.DAY_OF_MONTH;
import static java.time.temporal.ChronoField.MONTH_OF_YEAR;

import java.awt.Graphics2D;
import java.awt.Transparency;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeFormatterBuilder;
import java.time.format.SignStyle;
import java.util.Base64;
import java.util.List;
import javax.imageio.ImageIO;

/** Holds the character profile. */
public class Profile {
    private static final String KEY_AGE          = "age";
    private static final String KEY_BIRTHDAY     = "birthday";
    private static final String KEY_EYES         = "eyes";
    private static final String KEY_GENDER       = "gender";
    private static final String KEY_HAIR         = "hair";
    private static final String KEY_HANDEDNESS   = "handedness";
    private static final String KEY_HEIGHT       = "height";
    private static final String KEY_NAME         = "name";
    private static final String KEY_ORGANIZATION = "organization";
    private static final String KEY_PLAYER_NAME  = "player_name";
    private static final String KEY_PORTRAIT     = "portrait";
    private static final String KEY_RELIGION     = "religion";
    private static final String KEY_SKIN         = "skin";
    private static final String KEY_SM           = "SM";
    private static final String KEY_TITLE        = "title";
    private static final String KEY_TL           = "tech_level";
    private static final String KEY_WEIGHT       = "weight";

    private static final String KEY_BODY_TYPE = "body_type"; // Deprecated May 9, 2021

    public static final  int               PORTRAIT_HEIGHT      = 96; // Height of the portrait, in 1/72nds of an inch
    public static final  int               PORTRAIT_WIDTH       = 3 * PORTRAIT_HEIGHT / 4; // Width of the portrait, in 1/72nds of an inch
    private static final DateTimeFormatter MONTH_AND_DAY_FORMAT = new DateTimeFormatterBuilder().parseCaseInsensitive().parseLenient().appendText(MONTH_OF_YEAR, FULL).appendLiteral(' ').appendValue(DAY_OF_MONTH, 1, 2, SignStyle.NOT_NEGATIVE).toFormatter();

    private GURPSCharacter mCharacter;
    private RetinaIcon     mPortrait;
    private String         mName;
    private String         mTitle;
    private String         mOrganization;
    private String         mAge;
    private String         mBirthday;
    private String         mEyeColor;
    private String         mHair;
    private String         mSkinColor;
    private String         mHandedness;
    private LengthValue    mHeight;
    private WeightValue    mWeight;
    private int            mSizeModifier;
    private int            mSizeModifierBonus;
    private String         mGender;
    private String         mReligion;
    private String         mPlayerName;
    private String         mTechLevel;

    Profile(GURPSCharacter character, boolean full) {
        mCharacter = character;
        mPortrait = null;
        mTitle = "";
        mOrganization = "";
        mReligion = "";
        SheetSettings sheetSettings = mCharacter.getSheetSettings();
        if (full) {
            mGender = getRandomGender("");
            mAge = Numbers.format(getRandomAge(-1));
            mBirthday = getRandomBirthday("");
            mEyeColor = getRandomEyeColor("");
            mHair = getRandomHair("");
            mSkinColor = getRandomSkin("");
            mHandedness = getRandomHandedness("");
            mHeight = getRandomHeight(new LengthValue(Fixed6.ZERO, sheetSettings.defaultLengthUnits()));
            mWeight = getRandomWeight(Fixed6.ONE, new WeightValue(Fixed6.ZERO, sheetSettings.defaultWeightUnits()));
            mName = getRandomName("");
            GeneralSettings settings = Settings.getInstance().getGeneralSettings();
            mTechLevel = settings.getDefaultTechLevel();
            mPlayerName = settings.getDefaultPlayerName();
        } else {
            mGender = "";
            mAge = "";
            mBirthday = "";
            mEyeColor = "";
            mHair = "";
            mSkinColor = "";
            mHandedness = "";
            mHeight = new LengthValue(Fixed6.ZERO, sheetSettings.defaultLengthUnits());
            mWeight = new WeightValue(Fixed6.ZERO, sheetSettings.defaultWeightUnits());
            mName = "";
            mTechLevel = "";
            mPlayerName = "";
        }
    }

    void load(JsonMap m) {
        mPlayerName = m.getString(KEY_PLAYER_NAME);
        mName = m.getString(KEY_NAME);
        mTitle = m.getString(KEY_TITLE);
        mOrganization = m.getString(KEY_ORGANIZATION);
        mAge = m.getString(KEY_AGE);
        mBirthday = m.getString(KEY_BIRTHDAY);
        mEyeColor = m.getString(KEY_EYES);
        mHair = m.getString(KEY_HAIR);
        mSkinColor = m.getString(KEY_SKIN);
        mHandedness = m.getString(KEY_HANDEDNESS);
        mHeight = LengthValue.extract(m.getString(KEY_HEIGHT), false);
        mWeight = WeightValue.extract(m.getString(KEY_WEIGHT), false);
        mSizeModifier = m.getInt(KEY_SM);
        mGender = m.getString(KEY_GENDER);
        mTechLevel = m.getString(KEY_TL);
        mReligion = m.getString(KEY_RELIGION);

        // Legacy; GCS v4.29.1 or earlier
        if (m.has(KEY_BODY_TYPE)) {
            String bodyType = m.getString(KEY_BODY_TYPE);
            if (bodyType.startsWith("winged_")) {
                bodyType = bodyType.substring(7) + ".winged";
            }
            outer:
            for (NamedData<List<NamedData<HitLocationTable>>> list :
                    NamedData.scanLibraries(FileType.BODY_SETTINGS, Dirs.SETTINGS,
                            HitLocationTable::new)) {
                for (NamedData<HitLocationTable> one : list.getData()) {
                    if (bodyType.equals(one.getData().getID())) {
                        mCharacter.getSheetSettings().setHitLocations(one.getData());
                        break outer;
                    }
                }
            }
        }

        if (m.has(KEY_PORTRAIT)) {
            try {
                mPortrait = createPortrait(Img.create(new ByteArrayInputStream(
                        Base64.getDecoder().decode(m.getString(KEY_PORTRAIT)))));
            } catch (Exception imageException) {
                Log.error(imageException);
            }
        }
    }

    void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValueNot(KEY_PLAYER_NAME, mPlayerName, "");
        w.keyValueNot(KEY_NAME, mName, "");
        w.keyValueNot(KEY_TITLE, mTitle, "");
        w.keyValueNot(KEY_ORGANIZATION, mOrganization, "");
        w.keyValueNot(KEY_AGE, mAge, "");
        w.keyValueNot(KEY_BIRTHDAY, mBirthday, "");
        w.keyValueNot(KEY_EYES, mEyeColor, "");
        w.keyValueNot(KEY_HAIR, mHair, "");
        w.keyValueNot(KEY_SKIN, mSkinColor, "");
        w.keyValueNot(KEY_HANDEDNESS, mHandedness, "");
        if (!mHeight.getNormalizedValue().equals(Fixed6.ZERO)) {
            w.keyValue(KEY_HEIGHT, mHeight.toString(false));
        }
        if (!mWeight.getNormalizedValue().equals(Fixed6.ZERO)) {
            w.keyValue(KEY_WEIGHT, mWeight.toString(false));
        }
        w.keyValueNot(KEY_SM, mSizeModifier, 0);
        w.keyValueNot(KEY_GENDER, mGender, "");
        w.keyValueNot(KEY_TL, mTechLevel, "");
        w.keyValueNot(KEY_RELIGION, mReligion, "");
        if (mPortrait != null) {
            try (ByteArrayOutputStream baos = new ByteArrayOutputStream()) {
                ImageIO.write(mPortrait.getRetina(), FileType.PNG.getExtension(), baos);
                w.keyValue(KEY_PORTRAIT, Base64.getEncoder().encodeToString(baos.toByteArray()));
            } catch (Exception imageException) {
                Log.warn(imageException);
            }
        }
        w.endMap();
    }

    void update() {
        setSizeModifierBonus(mCharacter.getIntegerBonusFor(Attribute.ID_ATTR_PREFIX + "sm"));
    }

    /** @return The portrait. */
    public RetinaIcon getPortrait() {
        return mPortrait;
    }

    /** @return The portrait, or the default image if none is set. */
    public RetinaIcon getPortraitWithFallback() {
        return mPortrait == null ? Images.DEFAULT_PORTRAIT : mPortrait;
    }

    /**
     * Sets the portrait.
     *
     * @param portrait The new portrait.
     */
    public void setPortrait(Img portrait) {
        if (portrait == null) {
            if (mPortrait != null) {
                mCharacter.postUndoEdit(I18n.text("Portrait Change"), (c, v) -> c.getProfile().setPortrait(v != null ? ((RetinaIcon) v).getRetina() : null), mPortrait, null);
                mPortrait = null;
                mCharacter.notifyOfChange();
            }
        } else if (mPortrait == null || mPortrait.getRetina() != portrait) {
            RetinaIcon newPortrait = createPortrait(portrait);
            mCharacter.postUndoEdit(I18n.text("Portrait Change"), (c, v) -> c.getProfile().setPortrait(v != null ? ((RetinaIcon) v).getRetina() : null), mPortrait, newPortrait);
            mPortrait = newPortrait;
            mCharacter.notifyOfChange();
        }
    }

    public static RetinaIcon createPortrait(Img image) {
        if (image == null) {
            return null;
        }
        Img normal;
        Img retina;
        int width  = image.getWidth();
        int height = image.getHeight();
        if (width == PORTRAIT_WIDTH * 2 && height == PORTRAIT_HEIGHT * 2) {
            retina = image;
            normal = retina.scale(PORTRAIT_WIDTH, PORTRAIT_HEIGHT);
        } else if (width == PORTRAIT_WIDTH && height == PORTRAIT_HEIGHT) {
            normal = image;
            retina = normal.scale(PORTRAIT_WIDTH * 2, PORTRAIT_HEIGHT * 2);
        } else {
            int    dw = PORTRAIT_WIDTH * 2;
            int    dh = PORTRAIT_HEIGHT * 2;
            double r  = Math.min(dw / (double) width, dh / (double) height);
            int    w  = (int) (width * r);
            int    h  = (int) (height * r);
            retina = Img.create(dw, dh, Transparency.TRANSLUCENT);
            Graphics2D gc = retina.getGraphics();
            gc.drawImage(image.scale(w, h), (dw - w) / 2, (dh - h) / 2, null);
            gc.dispose();
            normal = retina.scale(PORTRAIT_WIDTH, PORTRAIT_HEIGHT);
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
            mCharacter.postUndoEdit(I18n.text("Name Change"), (c, v) -> c.getProfile().setName((String) v), mName, name);
            mName = name;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Gender Change"), (c, v) -> c.getProfile().setGender((String) v), mGender, gender);
            mGender = gender;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Religion Change"), (c, v) -> c.getProfile().setReligion((String) v), mReligion, religion);
            mReligion = religion;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Player Name Change"), (c, v) -> c.getProfile().setPlayerName((String) v), mPlayerName, player);
            mPlayerName = player;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Tech Level Change"), (c, v) -> c.getProfile().setTechLevel((String) v), mTechLevel, techLevel);
            mTechLevel = techLevel;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Title Change"), (c, v) -> c.getProfile().setTitle((String) v), mTitle, title);
            mTitle = title;
            mCharacter.notifyOfChange();
        }
    }

    /** @return The organization. */
    public String getOrganization() {
        return mOrganization;
    }

    /**
     * Sets the organization.
     *
     * @param organization The new organization.
     */
    public void setOrganization(String organization) {
        if (!mOrganization.equals(organization)) {
            mCharacter.postUndoEdit(I18n.text("Organization Change"), (c, v) -> c.getProfile().setOrganization((String) v), mOrganization, organization);
            mOrganization = organization;
            mCharacter.notifyOfChange();
        }
    }

    /** @return The age. */
    public String getAge() {
        return mAge;
    }

    /**
     * Sets the age.
     *
     * @param age The new age.
     */
    public void setAge(String age) {
        if (!mAge.equals(age)) {
            mCharacter.postUndoEdit(I18n.text("Age Change"), (c, v) -> c.getProfile().setAge((String) v), mAge, age);
            mAge = age;
            mCharacter.notifyOfChange();
        }
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
            mCharacter.postUndoEdit(I18n.text("Birthday Change"), (c, v) -> c.getProfile().setBirthday((String) v), mBirthday, birthday);
            mBirthday = birthday;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Eye Color Change"), (c, v) -> c.getProfile().setEyeColor((String) v), mEyeColor, eyeColor);
            mEyeColor = eyeColor;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Hair Change"), (c, v) -> c.getProfile().setHair((String) v), mHair, hair);
            mHair = hair;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Skin Color Change"), (c, v) -> c.getProfile().setSkinColor((String) v), mSkinColor, skinColor);
            mSkinColor = skinColor;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Handedness Change"), (c, v) -> c.getProfile().setHandedness((String) v), mHandedness, handedness);
            mHandedness = handedness;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Height Change"), (c, v) -> c.getProfile().setHeight((LengthValue) v), new LengthValue(mHeight), height);
            mHeight = height;
            mCharacter.notifyOfChange();
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
            mCharacter.postUndoEdit(I18n.text("Weight Change"), (c, v) -> c.getProfile().setWeight((WeightValue) v), new WeightValue(mWeight), weight);
            mWeight = weight;
            mCharacter.notifyOfChange();
        }
    }

    /** @return The multiplier compared to average weight for this character. */
    public Fixed6 getWeightMultiplier() {
        if (mCharacter.hasAdvantageNamed("Very Fat")) {
            return new Fixed6(2);
        } else if (mCharacter.hasAdvantageNamed("Fat")) {
            return new Fixed6("1.5", Fixed6.ZERO, false);
        } else if (mCharacter.hasAdvantageNamed("Overweight")) {
            return new Fixed6("1.3", Fixed6.ZERO, false);
        } else if (mCharacter.hasAdvantageNamed("Skinny")) {
            return new Fixed6("0.67", Fixed6.ZERO, false);
        }
        return Fixed6.ONE;
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

            mCharacter.postUndoEdit(I18n.text("Size Modifier Change"), (c, v) -> c.getProfile().setSizeModifier(((Integer) v).intValue()), Integer.valueOf(totalSizeModifier), value);
            mSizeModifier = size - mSizeModifierBonus;
            mCharacter.notifyOfChange();
        }
    }

    /** @param bonus The new size modifier bonus. */
    public void setSizeModifierBonus(int bonus) {
        if (mSizeModifierBonus != bonus) {
            mSizeModifierBonus = bonus;
            mCharacter.notifyOfChange();
        }
    }

    /** @return A random age. */
    @SuppressWarnings({"AutoBoxing", "AutoUnboxing"})
    public int getRandomAge(int not) {
        Ancestry ancestry = mCharacter.getAncestry();
        return getRandomized(not, () -> ancestry.getRandomAge(mCharacter, mGender));
    }

    /** @return A random hair color, style & length. */
    public String getRandomHair(String not) {
        Ancestry ancestry = mCharacter.getAncestry();
        return getRandomized(not, () -> ancestry.getRandomHair(mGender));
    }

    /** @return A random eye color. */
    public String getRandomEyeColor(String not) {
        Ancestry ancestry = mCharacter.getAncestry();
        return getRandomized(not, () -> ancestry.getRandomEyeColor(mGender));
    }

    /** @return A random skin. */
    public String getRandomSkin(String not) {
        Ancestry ancestry = mCharacter.getAncestry();
        return getRandomized(not, () -> ancestry.getRandomSkin(mGender));
    }

    /** @return A random handedness. */
    public String getRandomHandedness(String not) {
        Ancestry ancestry = mCharacter.getAncestry();
        return getRandomized(not, () -> ancestry.getRandomHandedness(mGender));
    }

    /** @return A random gender. */
    public String getRandomGender(String not) {
        Ancestry ancestry = mCharacter.getAncestry();
        return getRandomized(not, ancestry::getRandomGender);
    }

    /** @return A random birthday. */
    public String getRandomBirthday(String not) {
        int      year = 1;
        int      base = 0;
        Calendar cal  = CalendarRef.currentCalendar();
        if (cal.mLeapYear != null) {
            while (!cal.isLeapYear(year)) {
                year++;
            }
            try {
                base = new Date(cal, 1, 1, year).days();
            } catch (CalendarException e) {
                Log.error(e);
            }
        }
        int start = base;
        int range = cal.days(year);
        return getRandomized(not, () -> new Date(cal, start + Dice.RANDOM.nextInt(range)).format("%M %D"));
    }

    /**
     * @param not A height not to return, if possible.
     * @return A random height.
     */
    public LengthValue getRandomHeight(LengthValue not) {
        Ancestry    ancestry     = mCharacter.getAncestry();
        LengthUnits desiredUnits = mCharacter.getSheetSettings().defaultLengthUnits();
        return getRandomized(not, () -> {
            Fixed6 base = new Fixed6(ancestry.getRandomHeightInInches(mCharacter, mGender));
            return new LengthValue(desiredUnits.convert(LengthUnits.IN, base).round(), desiredUnits);
        });
    }

    /**
     * @param multiplier The weight multiplier for being under- or overweight.
     * @param not        A weight not to return, if possible.
     * @return A random weight.
     */
    public WeightValue getRandomWeight(Fixed6 multiplier, WeightValue not) {
        Ancestry    ancestry     = mCharacter.getAncestry();
        WeightUnits desiredUnits = mCharacter.getSheetSettings().defaultWeightUnits();
        return getRandomized(not, () -> {
            Fixed6 base = new Fixed6(ancestry.getRandomWeightInPounds(mCharacter, mGender));
            base = base.mul(multiplier).round();
            if (base.lessThan(Fixed6.ONE)) {
                base = Fixed6.ONE;
            }
            return new WeightValue(desiredUnits.convert(WeightUnits.LB, base).round(), desiredUnits);
        });
    }

    /** @return A random name. */
    public String getRandomName(String not) {
        Ancestry ancestry = mCharacter.getAncestry();
        return getRandomized(not, () -> ancestry.getRandomName(mGender));
    }

    interface Randomizer<T> {
        T getRandomResult();
    }

    private static <T> T getRandomized(T not, Randomizer<T> randomizer) {
        T   result;
        int maxAttempts = 5;
        do {
            result = randomizer.getRandomResult();
            if (--maxAttempts == 0) {
                break;
            }
        } while (result.equals(not));
        return result;
    }
}
