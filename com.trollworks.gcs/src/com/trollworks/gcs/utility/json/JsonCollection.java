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

package com.trollworks.gcs.utility.json;

import com.trollworks.gcs.utility.Log;

import java.io.IOException;

/** Common base class for JSON collections. */
public abstract class JsonCollection {
    public abstract boolean isEmpty();

    @Override
    public final String toString() {
        return toString(false);
    }

    public final String toString(boolean compact) {
        StringBuilder buffer = new StringBuilder();
        try {
            appendTo(buffer, compact, 0);
        } catch (IOException ioe) {
            Log.error(ioe);
        }
        return buffer.toString();
    }

    public abstract void appendTo(Appendable buffer, boolean compact, int depth) throws IOException;

    protected static void indent(Appendable buffer, boolean compact, int depth) throws IOException {
        if (!compact) {
            buffer.append("\t".repeat(depth));
        }
    }
}
