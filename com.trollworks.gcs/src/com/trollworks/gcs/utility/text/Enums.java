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

package com.trollworks.gcs.utility.text;

/** A utility for consistent extraction of an {@link Enum} value from a text buffer. */
public class Enums {
    public static final String toId(Enum<?> value) {
        return value.name().toLowerCase();
    }

    /**
     * @param <T>          The type of {@link Enum}.
     * @param buffer       The buffer to load from.
     * @param values       The possible values.
     * @param defaultValue The default value to use in case of no match.
     * @return The {@link Enum} representing the buffer.
     */
    public static final <T extends Enum<?>> T extract(String buffer, T[] values, T defaultValue) {
        T value = extract(buffer, values);
        return value != null ? value : defaultValue;
    }

    /**
     * @param <T>    The type of {@link Enum}.
     * @param buffer The buffer to load from.
     * @param values The possible values.
     * @return The {@link Enum} representing the buffer, or {@code null} if a match could not be
     *         found.
     */
    public static final <T extends Enum<?>> T extract(String buffer, T[] values) {
        if (buffer != null) {
            for (T type : values) {
                String name = type.name();
                if (name.equalsIgnoreCase(buffer) || name.replace('_', ' ').equalsIgnoreCase(buffer) || name.replace('_', ',').equalsIgnoreCase(buffer) || name.replace('_', '-').equalsIgnoreCase(buffer) || name.replaceAll("_", "").equalsIgnoreCase(buffer)) {
                    return type;
                }
            }
        }
        return null;
    }
}
