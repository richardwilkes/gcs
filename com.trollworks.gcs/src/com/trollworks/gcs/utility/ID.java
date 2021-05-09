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

package com.trollworks.gcs.utility;

import java.util.Set;

public final class ID {
    private ID() {
    }

    public static String sanitize(String id, Set<String> reserved, boolean permitLeadingDigits) {
        StringBuilder buffer = new StringBuilder();
        int           length = id.length();
        for (int i = 0; i < length; i++) {
            char ch = id.charAt(i);
            if (ch >= 'A' && ch <= 'Z') {
                ch = Character.toLowerCase(ch);
            }
            if (ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9' && (permitLeadingDigits || !buffer.isEmpty()))) {
                buffer.append(ch);
            }
        }
        id = buffer.toString();
        if (reserved != null && reserved.contains(id)) {
            id += "_";
        }
        return id;
    }
}
