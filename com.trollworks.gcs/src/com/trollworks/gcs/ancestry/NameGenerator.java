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

package com.trollworks.gcs.ancestry;

import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.NamedData;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.io.BufferedReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class NameGenerator {
    private static final String                     KEY_TYPE              = "type";
    private static final String                     KEY_TRAINING_DATA     = "training_data";
    private static final Map<String, NameGenerator> REGISTERED_GENERATORS = new HashMap<>();
    private              NameGenerationType         mType;
    private              List<String>               mTrainingData;
    private              int                        mMin;
    private              int                        mMax;
    private              Map<String, Entry>         mEntries;

    public static final NameGenerator get(String name) {
        if (REGISTERED_GENERATORS.isEmpty()) {
            for (NamedData<List<NamedData<NameGenerator>>> list : NamedData.scanLibraries(FileType.NAME_GENERATOR_SETTINGS, Dirs.SETTINGS, NameGenerator::new)) {
                for (NamedData<NameGenerator> data : list.getData()) {
                    REGISTERED_GENERATORS.putIfAbsent(data.getName(), data.getData());
                }
            }
        }
        return REGISTERED_GENERATORS.get(name);
    }

    /**
     * Creates a new name generator.
     *
     * @param path The path to the json file to load from.
     */
    public NameGenerator(Path path) throws IOException {
        try (BufferedReader fileReader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m = Json.asMap(Json.parse(fileReader));
            mType = Enums.extract(m.getString(KEY_TYPE), NameGenerationType.values(), NameGenerationType.SIMPLE);
            JsonArray    a            = m.getArray(KEY_TRAINING_DATA);
            int          count        = a.size();
            List<String> trainingData = new ArrayList<>(count);
            for (int i = 0; i < count; i++) {
                trainingData.add(a.getString(i));
            }
            initialize(trainingData);
        }
    }

    /**
     * Creates a new name generator.
     *
     * @param trainingData The training data.
     */
    public NameGenerator(NameGenerationType type, List<String> trainingData) {
        mType = type;
        initialize(trainingData);
    }

    private void initialize(List<String> trainingData) {
        mTrainingData = new ArrayList<>(trainingData.size());
        for (String one : trainingData) {
            one = one.trim();
            if (one.length() >= 2) {
                mTrainingData.add(one.toLowerCase());
            }
        }
        if (mType == NameGenerationType.MARKOV_CHAIN) {
            mMin = 20;
            mMax = 2;
            Map<String, Map<Character, Integer>> builders = new HashMap<>();
            for (String one : mTrainingData) {
                int len = one.length();
                if (mMin > len) {
                    mMin = len;
                }
                if (mMax < len) {
                    mMax = len;
                }
                for (int i = 2; i < len; i++) {
                    String charGroup = one.substring(i - 2, i);
                    if (!builders.containsKey(charGroup)) {
                        builders.put(charGroup, new HashMap<>());
                    }
                    Map<Character, Integer> m              = builders.get(charGroup);
                    Character               subsequentChar = Character.valueOf(one.charAt(i));
                    m.put(subsequentChar, Integer.valueOf(m.getOrDefault(subsequentChar, Integer.valueOf(0)).intValue() + 1));
                }
            }
            mEntries = new HashMap<>();
            for (Map.Entry<String, Map<Character, Integer>> entry : builders.entrySet()) {
                mEntries.put(entry.getKey(), new Entry(entry.getValue()));
            }
        }
    }

    public void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_TYPE, Enums.toId(mType));
        w.key(KEY_TRAINING_DATA);
        w.startArray();
        for (String one : mTrainingData) {
            w.value(one);
        }
        w.endArray();
        w.endMap();
    }

    public String generate() {
        switch (mType) {
            case SIMPLE -> {
                if (mTrainingData.isEmpty()) {
                    return "";
                }
                String name = mTrainingData.get(Dice.RANDOM.nextInt(mTrainingData.size()));
                return Character.toUpperCase(name.charAt(0)) + name.substring(1).toLowerCase();
            }
            case MARKOV_CHAIN -> {
                int           targetSize = Dice.RANDOM.nextInt(mMin, mMax + 1);
                StringBuilder buffer     = new StringBuilder();
                buffer.append((String) mEntries.keySet().toArray()[Dice.RANDOM.nextInt(mEntries.size())]);
                for (int i = 2; i < targetSize; i++) {
                    String sub = buffer.substring(i - 2, i);
                    if (!mEntries.containsKey(sub)) {
                        break;
                    }
                    char next = mEntries.get(sub).chooseCharacter();
                    if (next == 0) {
                        break;
                    }
                    buffer.append(next);
                }
                buffer.setCharAt(0, Character.toUpperCase(buffer.charAt(0)));
                return buffer.toString();
            }
            default -> {
                return "";
            }
        }
    }

    static class Entry {
        private char[] mCharacters;
        private int[]  mThresholds;

        Entry(Map<Character, Integer> occurrences) {
            mCharacters = new char[occurrences.size()];
            mThresholds = new int[occurrences.size()];
            int i = 0;
            for (Map.Entry<Character, Integer> entry : occurrences.entrySet()) {
                mCharacters[i] = entry.getKey().charValue();
                mThresholds[i] = entry.getValue().intValue() + (i > 0 ? mThresholds[i - 1] : 0);
                i++;
            }
        }

        char chooseCharacter() {
            int threshold = Dice.RANDOM.nextInt(mThresholds[mThresholds.length - 1] + 1);
            int count     = mThresholds.length;
            for (int i = 0; i < count; i++) {
                if (mThresholds[i] >= threshold) {
                    return mCharacters[i];
                }
            }
            return 0;
        }
    }
}
