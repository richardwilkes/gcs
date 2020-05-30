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

package com.trollworks.gcs.utility.json;

/** Represents a 'null' in JSON. */
public final class JsonNull {
    /** The singleton instance. */
    public static final JsonNull INSTANCE = new JsonNull();

    private JsonNull() {
        // Singleton
    }

    @Override
    public boolean equals(Object other) {
        return other == null || other == this;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    @Override
    public String toString() {
        return "null";
    }
}
