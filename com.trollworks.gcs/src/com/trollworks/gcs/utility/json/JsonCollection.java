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

/** Common base class for JSON collections. */
public abstract class JsonCollection {
    @Override
    public final String toString() {
        return toString(false);
    }

    public final String toString(boolean compact) {
        return appendTo(new StringBuilder(), compact, 0).toString();
    }

    public abstract StringBuilder appendTo(StringBuilder buffer, boolean compact, int depth);

    protected static void indent(StringBuilder buffer, boolean compact, int depth) {
        if (!compact) {
            buffer.append("\t".repeat(depth));
        }
    }
}
