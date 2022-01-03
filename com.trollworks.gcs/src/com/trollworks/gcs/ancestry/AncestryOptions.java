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
import com.trollworks.gcs.expression.EvaluationException;
import com.trollworks.gcs.expression.Evaluator;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

/** Options that may be randomized for a character's ancestry. */
public class AncestryOptions {
    private static final String                     KEY_NAME               = "name";
    private static final String                     KEY_HEIGHT_FORMULA     = "height_formula";
    private static final String                     KEY_WEIGHT_FORMULA     = "weight_formula";
    private static final String                     KEY_AGE_FORMULA        = "age_formula";
    private static final String                     KEY_NAME_GENERATORS    = "name_generators";
    private static final String                     KEY_HAIR_OPTIONS       = "hair_options";
    private static final String                     KEY_EYE_OPTIONS        = "eye_options";
    private static final String                     KEY_SKIN_OPTIONS       = "skin_options";
    private static final String                     KEY_HANDEDNESS_OPTIONS = "handedness_options";
    public               String                     mName;
    public               String                     mHeightFormula;
    public               String                     mWeightFormula;
    public               String                     mAgeFormula;
    public               List<WeightedStringOption> mHairOptions;
    public               List<WeightedStringOption> mEyeOptions;
    public               List<WeightedStringOption> mSkinOptions;
    public               List<WeightedStringOption> mHandednessOptions;
    public               List<String>               mNameGenerators;

    public AncestryOptions(String name) {
        mName = name;
        mHeightFormula = "";
        mWeightFormula = "";
        mAgeFormula = "";
        mHairOptions = new ArrayList<>();
        mEyeOptions = new ArrayList<>();
        mSkinOptions = new ArrayList<>();
        mHandednessOptions = new ArrayList<>();
        mNameGenerators = new ArrayList<>();
    }

    public AncestryOptions(JsonMap m) {
        mName = m.getString(KEY_NAME);
        mHeightFormula = m.getString(KEY_HEIGHT_FORMULA);
        mWeightFormula = m.getString(KEY_WEIGHT_FORMULA);
        mAgeFormula = m.getString(KEY_AGE_FORMULA);
        mHairOptions = WeightedOption.loadList(m, KEY_HAIR_OPTIONS, WeightedStringOption.class);
        mEyeOptions = WeightedOption.loadList(m, KEY_EYE_OPTIONS, WeightedStringOption.class);
        mSkinOptions = WeightedOption.loadList(m, KEY_SKIN_OPTIONS, WeightedStringOption.class);
        mHandednessOptions = WeightedOption.loadList(m, KEY_HANDEDNESS_OPTIONS, WeightedStringOption.class);
        mNameGenerators = new ArrayList<>();
        if (m.has(KEY_NAME_GENERATORS)) {
            JsonArray a     = m.getArray(KEY_NAME_GENERATORS);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                mNameGenerators.add(a.getString(i));
            }
        }
    }

    public void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValueNot(KEY_NAME, mName, "");
        w.keyValueNot(KEY_HEIGHT_FORMULA, mHeightFormula, "");
        w.keyValueNot(KEY_WEIGHT_FORMULA, mWeightFormula, "");
        w.keyValueNot(KEY_AGE_FORMULA, mAgeFormula, "");
        WeightedOption.saveList(w, KEY_HAIR_OPTIONS, mHairOptions);
        WeightedOption.saveList(w, KEY_EYE_OPTIONS, mEyeOptions);
        WeightedOption.saveList(w, KEY_SKIN_OPTIONS, mSkinOptions);
        WeightedOption.saveList(w, KEY_HANDEDNESS_OPTIONS, mHandednessOptions);
        if (!mNameGenerators.isEmpty()) {
            w.key(KEY_NAME_GENERATORS);
            w.startArray();
            for (String one : mNameGenerators) {
                w.value(one);
            }
            w.endArray();
        }
        w.endMap();
    }

    public AncestryOptions setToDefaults() {
        mHeightFormula = "roll(dice(1, 8, if($st < 7, 51, if($st < 10, 54 + ($st - 7) * 3, if($st == 10, 62, if($st < 14, 64 + ($st - 11) * 3, 73))))))";
        mWeightFormula = "roll(dice(1, if($st < 11, 61, if($st < 14, 71 + ($st - 11) * 10, 101)), if($st < 7, 60, if($st < 10, 75 + ($st - 7) * 15, if($st == 10, 115, if($st < 14, 125 + ($st - 11) * 15, 170))))))";
        mAgeFormula = "roll(1d12+14)";

        mHairOptions = new ArrayList<>();
        mHairOptions.add(new WeightedStringOption(14, "Black"));
        mHairOptions.add(new WeightedStringOption(8, "Brown"));
        mHairOptions.add(new WeightedStringOption(5, "Blond"));
        mHairOptions.add(new WeightedStringOption(3, "Redhead"));
        mHairOptions.add(new WeightedStringOption(1, "Bald"));

        mEyeOptions = new ArrayList<>();
        mEyeOptions.add(new WeightedStringOption(52, "Brown"));
        mEyeOptions.add(new WeightedStringOption(18, "Blue"));
        mEyeOptions.add(new WeightedStringOption(10, "Amber"));
        mEyeOptions.add(new WeightedStringOption(10, "Hazel"));
        mEyeOptions.add(new WeightedStringOption(6, "Gray"));
        mEyeOptions.add(new WeightedStringOption(4, "Green"));

        mSkinOptions = new ArrayList<>();
        mSkinOptions.add(new WeightedStringOption(3, "Dark Brown"));
        mSkinOptions.add(new WeightedStringOption(3, "Brown"));
        mSkinOptions.add(new WeightedStringOption(3, "Olive"));
        mSkinOptions.add(new WeightedStringOption(3, "Light Brown"));
        mSkinOptions.add(new WeightedStringOption(3, "Tan"));
        mSkinOptions.add(new WeightedStringOption(3, "Freckled"));
        mSkinOptions.add(new WeightedStringOption(1, "Pale"));

        mHandednessOptions = new ArrayList<>();
        mHandednessOptions.add(new WeightedStringOption(1, "Left"));
        mHandednessOptions.add(new WeightedStringOption(9, "Right"));

        return this;
    }

    public double getRandomHeightInInches(GURPSCharacter gchar) {
        Evaluator evaluator = new Evaluator(gchar);
        try {
            return evaluator.evaluateToNumber(mHeightFormula);
        } catch (EvaluationException e) {
            Log.error(e);
            return 64;
        }
    }

    public double getRandomWeightInPounds(GURPSCharacter gchar) {
        Evaluator evaluator = new Evaluator(gchar);
        try {
            return evaluator.evaluateToNumber(mWeightFormula);
        } catch (EvaluationException e) {
            Log.error(e);
            return 145;
        }
    }

    public int getRandomAge(GURPSCharacter gchar) {
        Evaluator evaluator = new Evaluator(gchar);
        try {
            return evaluator.evaluateToInteger(mAgeFormula);
        } catch (EvaluationException e) {
            Log.error(e);
            return 18;
        }
    }

    public String getRandomHair() {
        return pick(mHairOptions);
    }

    public String getRandomEyeColor() {
        return pick(mEyeOptions);
    }

    public String getRandomSkin() {
        return pick(mSkinOptions);
    }

    public String getRandomHandedness() {
        return pick(mHandednessOptions);
    }

    public String pick(List<WeightedStringOption> options) {
        WeightedStringOption choice = WeightedOption.choose(options);
        if (choice != null) {
            return choice.mValue;
        }
        return null;
    }

    public String getRandomName() {
        StringBuilder buffer = new StringBuilder();
        if (mNameGenerators != null) {
            for (String one : mNameGenerators) {
                NameGenerator generator = NameGenerator.get(one);
                if (generator != null) {
                    String text = generator.generate();
                    if (text != null) {
                        text = text.trim();
                        if (!text.isBlank()) {
                            if (!buffer.isEmpty()) {
                                buffer.append(" ");
                            }
                            buffer.append(text);
                        }
                    }
                }
            }
        }
        return buffer.toString();
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }

        AncestryOptions that = (AncestryOptions) other;

        if (!mName.equals(that.mName)) {
            return false;
        }
        if (!mHeightFormula.equals(that.mHeightFormula)) {
            return false;
        }
        if (!mWeightFormula.equals(that.mWeightFormula)) {
            return false;
        }
        if (!mAgeFormula.equals(that.mAgeFormula)) {
            return false;
        }
        if (!mHairOptions.equals(that.mHairOptions)) {
            return false;
        }
        if (!mEyeOptions.equals(that.mEyeOptions)) {
            return false;
        }
        if (!mSkinOptions.equals(that.mSkinOptions)) {
            return false;
        }
        if (!mHandednessOptions.equals(that.mHandednessOptions)) {
            return false;
        }
        return mNameGenerators.equals(that.mNameGenerators);
    }

    @Override
    public int hashCode() {
        int result = mName.hashCode();
        result = 31 * result + mHeightFormula.hashCode();
        result = 31 * result + mWeightFormula.hashCode();
        result = 31 * result + mAgeFormula.hashCode();
        result = 31 * result + mHairOptions.hashCode();
        result = 31 * result + mEyeOptions.hashCode();
        result = 31 * result + mSkinOptions.hashCode();
        result = 31 * result + mHandednessOptions.hashCode();
        result = 31 * result + mNameGenerators.hashCode();
        return result;
    }
}
