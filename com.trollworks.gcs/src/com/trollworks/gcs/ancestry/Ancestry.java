/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ancestry;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.BufferedReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

/** Holds details necessary to generate ancestry-specific customizations. */
public class Ancestry {
    private static final String                        KEY_COMMON_OPTIONS   = "common_options";
    private static final String                        KEY_GENDER_OPTIONS   = "gender_options";
    public               AncestryOptions               mCommonOptions;
    public               List<WeightedAncestryOptions> mGenderOptions;

    public Ancestry() {
        mCommonOptions = new AncestryOptions("").setToDefaults();
        mCommonOptions.mNameGenerators = List.of("Human Last");
        mGenderOptions = new ArrayList<>();
        AncestryOptions male = new AncestryOptions("Male");
        male.mNameGenerators = List.of("Human First - Male", "Human Last");
        mGenderOptions.add(new WeightedAncestryOptions(1, male));
        AncestryOptions female = new AncestryOptions("Female");
        female.mNameGenerators = List.of("Human First - Female", "Human Last");
        mGenderOptions.add(new WeightedAncestryOptions(1, female));
    }

    public Ancestry(Path path) throws IOException {
        try (BufferedReader fileReader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m = Json.asMap(Json.parse(fileReader));
            if (m.has(KEY_COMMON_OPTIONS)) {
                mCommonOptions = new AncestryOptions(m.getMap(KEY_COMMON_OPTIONS));
            }
            mGenderOptions = WeightedOption.loadList(m, KEY_GENDER_OPTIONS, WeightedAncestryOptions.class);
        }
    }

    public void save(JsonWriter w) throws IOException {
        w.startMap();
        if (mCommonOptions != null) {
            w.key(KEY_COMMON_OPTIONS);
            mCommonOptions.save(w);
        }
        WeightedOption.saveList(w, KEY_GENDER_OPTIONS, mGenderOptions);
        w.endMap();
    }

    public String getRandomGender() {
        WeightedAncestryOptions choice = WeightedOption.choose(mGenderOptions);
        if (choice != null) {
            return choice.mValue.mName;
        }
        Ancestry ancestry = AncestryRef.DEFAULT.ancestry();
        if (ancestry != this) {
            return ancestry.getRandomGender();
        }
        return "Male";
    }

    private WeightedAncestryOptions getGenderedOptions(String gender) {
        gender = gender.trim();
        for (WeightedAncestryOptions options : mGenderOptions) {
            if (gender.equalsIgnoreCase(options.mValue.mName)) {
                return options;
            }
        }
        return null;
    }

    public double getRandomHeightInInches(GURPSCharacter gchar, String gender) {
        WeightedAncestryOptions options = getGenderedOptions(gender);
        if (options != null && !options.mValue.mHeightFormula.isBlank()) {
            return options.mValue.getRandomHeightInInches(gchar);
        }
        if (mCommonOptions != null && !mCommonOptions.mHeightFormula.isBlank()) {
            return mCommonOptions.getRandomHeightInInches(gchar);
        }
        Ancestry ancestry = AncestryRef.DEFAULT.ancestry();
        if (ancestry != this) {
            return ancestry.getRandomHeightInInches(gchar, gender);
        }
        return 64;
    }

    public double getRandomWeightInPounds(GURPSCharacter gchar, String gender) {
        WeightedAncestryOptions options = getGenderedOptions(gender);
        if (options != null && !options.mValue.mWeightFormula.isBlank()) {
            return options.mValue.getRandomWeightInPounds(gchar);
        }
        if (mCommonOptions != null && !mCommonOptions.mWeightFormula.isBlank()) {
            return mCommonOptions.getRandomWeightInPounds(gchar);
        }
        Ancestry ancestry = AncestryRef.DEFAULT.ancestry();
        if (ancestry != this) {
            return ancestry.getRandomWeightInPounds(gchar, gender);
        }
        return 140;
    }

    public int getRandomAge(GURPSCharacter gchar, String gender) {
        WeightedAncestryOptions options = getGenderedOptions(gender);
        if (options != null && !options.mValue.mAgeFormula.isBlank()) {
            return options.mValue.getRandomAge(gchar);
        }
        if (mCommonOptions != null && !mCommonOptions.mAgeFormula.isBlank()) {
            return mCommonOptions.getRandomAge(gchar);
        }
        Ancestry ancestry = AncestryRef.DEFAULT.ancestry();
        if (ancestry != this) {
            return ancestry.getRandomAge(gchar, gender);
        }
        return 20;
    }

    public String getRandomHair(String gender) {
        WeightedAncestryOptions options = getGenderedOptions(gender);
        if (options != null && !options.mValue.mHairOptions.isEmpty()) {
            return options.mValue.getRandomHair();
        }
        if (mCommonOptions != null && !mCommonOptions.mHairOptions.isEmpty()) {
            return mCommonOptions.getRandomHair();
        }
        Ancestry ancestry = AncestryRef.DEFAULT.ancestry();
        if (ancestry != this) {
            return ancestry.getRandomHair(gender);
        }
        return "Brown";
    }

    public String getRandomEyeColor(String gender) {
        WeightedAncestryOptions options = getGenderedOptions(gender);
        if (options != null && !options.mValue.mEyeOptions.isEmpty()) {
            return options.mValue.getRandomEyeColor();
        }
        if (mCommonOptions != null && !mCommonOptions.mEyeOptions.isEmpty()) {
            return mCommonOptions.getRandomEyeColor();
        }
        Ancestry ancestry = AncestryRef.DEFAULT.ancestry();
        if (ancestry != this) {
            return ancestry.getRandomEyeColor(gender);
        }
        return "Brown";
    }

    public String getRandomSkin(String gender) {
        WeightedAncestryOptions options = getGenderedOptions(gender);
        if (options != null && !options.mValue.mSkinOptions.isEmpty()) {
            return options.mValue.getRandomSkin();
        }
        if (mCommonOptions != null && !mCommonOptions.mSkinOptions.isEmpty()) {
            return mCommonOptions.getRandomSkin();
        }
        Ancestry ancestry = AncestryRef.DEFAULT.ancestry();
        if (ancestry != this) {
            return ancestry.getRandomSkin(gender);
        }
        return "Brown";
    }

    public String getRandomHandedness(String gender) {
        WeightedAncestryOptions options = getGenderedOptions(gender);
        if (options != null && !options.mValue.mHandednessOptions.isEmpty()) {
            return options.mValue.getRandomHandedness();
        }
        if (mCommonOptions != null && !mCommonOptions.mHandednessOptions.isEmpty()) {
            return mCommonOptions.getRandomHandedness();
        }
        Ancestry ancestry = AncestryRef.DEFAULT.ancestry();
        if (ancestry != this) {
            return ancestry.getRandomHandedness(gender);
        }
        return "Right";
    }

    public String getRandomName(String gender) {
        WeightedAncestryOptions options = getGenderedOptions(gender);
        if (options != null && options.mValue.mNameGenerators != null && !options.mValue.mNameGenerators.isEmpty()) {
            return options.mValue.getRandomName();
        }
        if (mCommonOptions != null && mCommonOptions.mNameGenerators != null && !mCommonOptions.mNameGenerators.isEmpty()) {
            return mCommonOptions.getRandomName();
        }
        Ancestry ancestry = AncestryRef.DEFAULT.ancestry();
        if (ancestry != this) {
            return ancestry.getRandomName(gender);
        }
        return "";
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }

        Ancestry ancestry = (Ancestry) other;

        if (!mCommonOptions.equals(ancestry.mCommonOptions)) {
            return false;
        }
        return mGenderOptions.equals(ancestry.mGenderOptions);
    }

    @Override
    public int hashCode() {
        int result = mCommonOptions.hashCode();
        result = 31 * result + mGenderOptions.hashCode();
        return result;
    }
}
