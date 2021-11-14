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

import java.io.BufferedReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class NameGenerator {
    private static Map<String, NameGenerator> REGISTERED_GENERATORS = new HashMap<>();
    private        int                        mMin;
    private        int                        mMax;
    private        Map<String, Entry>         mEntries;

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

    public NameGenerator(Path path) throws IOException {
        mMin = 20;
        mMax = 2;
        mEntries = new HashMap<>();
        Map<String, Map<Character, Integer>> builders = new HashMap<>();
        try (BufferedReader fileReader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            String line = fileReader.readLine();
            while (line != null) {
                line = line.trim();
                if (!line.isBlank()) {
                    int len = line.length();
                    if (len < 2) {
                        continue;
                    }
                    if (mMin > len) {
                        mMin = len;
                    }
                    if (mMax < len) {
                        mMax = len;
                    }
                    line = line.toLowerCase();
                    for (int i = 2; i < len; i++) {
                        String charGroup = line.substring(i - 2, i);
                        if (!builders.containsKey(charGroup)) {
                            builders.put(charGroup, new HashMap<>());
                        }
                        Map<Character, Integer> m              = builders.get(charGroup);
                        char                    subsequentChar = line.charAt(i);
                        m.put(subsequentChar, m.getOrDefault(subsequentChar, 0) + 1);
                    }
                }
                line = fileReader.readLine();
            }
        }
        for (Map.Entry<String, Map<Character, Integer>> entry : builders.entrySet()) {
            mEntries.put(entry.getKey(), new Entry(entry.getValue()));
        }
    }

    public String generate() {
        if (mEntries.isEmpty()) {
            return "";
        }
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
            for (int i = 0; i < mThresholds.length; i++) {
                if (mThresholds[i] >= threshold) {
                    return mCharacters[i];
                }
            }
            return 0;
        }
    }
}
