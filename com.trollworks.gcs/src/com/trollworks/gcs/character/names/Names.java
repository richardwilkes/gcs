/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character.names;

import com.trollworks.gcs.utility.Log;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;

/** An abstract base class for name generation. */
public abstract class Names {
    /**
     * A random number generator that can be shared amongst instances of the {@link Names} class.
     */
    protected static final Random RANDOM = new Random();

    /** @return A newly generated male first name. */
    public abstract String getMaleFirstName();

    /** @return A newly generated female first name. */
    public abstract String getFemaleFirstName();

    /**
     * @param male Whether to generate a male or female name.
     * @return A newly generated first name.
     */
    public String getGivenName(boolean male) {
        return male ? getMaleFirstName() : getFemaleFirstName();
    }

    /** @return A newly generated last name. */
    public abstract String getLastName();

    /**
     * @param male Whether to generate a male or female name.
     * @return A newly generated full (first and last) name.
     */
    public String getFullName(boolean male) {
        return getGivenName(male) + " " + getLastName();
    }

    /**
     * Loads names from a resource file into an array of strings. The names in the file should be
     * listed one per line.
     *
     * @param fileName The file name to load the name data from.
     * @param fallback A single name to use in case the file couldn't be loaded.
     * @return An array of strings representing the contents of the file.
     */
    protected static final String[] loadNames(String fileName, String fallback) {
        List<String> names = new ArrayList<>();
        try {
            try (BufferedReader in = new BufferedReader(new InputStreamReader(Names.class.getModule().getResourceAsStream("/names/" + fileName), StandardCharsets.UTF_8))) {
                String line = in.readLine();
                while (line != null) {
                    line = line.trim();
                    if (!line.isBlank()) {
                        names.add(line);
                    }
                    line = in.readLine();
                }
            }
        } catch (Exception exception) {
            Log.error(exception);
        }
        if (names.isEmpty()) {
            names.add(fallback);
        }
        return names.toArray(new String[0]);
    }
}
