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

package com.trollworks.gcs.criteria;

import com.trollworks.gcs.utility.I18n;

/** The allowed string comparison types. */
public enum StringCompareType {
    /** The comparison for "is anything". */
    ANY {
        @Override
        public String toString() {
            return I18n.Text("is anything");
        }

        @Override
        public boolean matches(String qualifier, String data) {
            return true;
        }
    },
    /** The comparison for "is". */
    IS {
        @Override
        public String toString() {
            return I18n.Text("is");
        }

        @Override
        public boolean matches(String qualifier, String data) {
            return data.equalsIgnoreCase(qualifier);
        }
    },
    /** The comparison for "is not". */
    IS_NOT {
        @Override
        public String toString() {
            return I18n.Text("is not");
        }

        @Override
        public boolean matches(String qualifier, String data) {
            return !data.equalsIgnoreCase(qualifier);
        }
    },
    /** The comparison for "contains". */
    CONTAINS {
        @Override
        public String toString() {
            return I18n.Text("contains");
        }

        @Override
        public boolean matches(String qualifier, String data) {
            return data.toLowerCase().contains(qualifier.toLowerCase());
        }
    },
    /** The comparison for "does not contain". */
    DOES_NOT_CONTAIN {
        @Override
        public String toString() {
            return I18n.Text("does not contain");
        }

        @Override
        public boolean matches(String qualifier, String data) {
            return !data.toLowerCase().contains(qualifier.toLowerCase());
        }
    },
    /** The comparison for "starts with". */
    STARTS_WITH {
        @Override
        public String toString() {
            return I18n.Text("starts with");
        }

        @Override
        public boolean matches(String qualifier, String data) {
            return data.toLowerCase().startsWith(qualifier.toLowerCase());
        }
    },
    /** The comparison for "does not start with". */
    DOES_NOT_START_WITH {
        @Override
        public String toString() {
            return I18n.Text("does not start with");
        }

        @Override
        public boolean matches(String qualifier, String data) {
            return !data.toLowerCase().startsWith(qualifier.toLowerCase());
        }
    },
    /** The comparison for "ends with". */
    ENDS_WITH {
        @Override
        public String toString() {
            return I18n.Text("ends with");
        }

        @Override
        public boolean matches(String qualifier, String data) {
            return data.toLowerCase().endsWith(qualifier.toLowerCase());
        }
    },
    /** The comparison for "does not end with". */
    DOES_NOT_END_WITH {
        @Override
        public String toString() {
            return I18n.Text("does not end with");
        }

        @Override
        public boolean matches(String qualifier, String data) {
            return !data.toLowerCase().endsWith(qualifier.toLowerCase());
        }
    };

    /**
     * @param qualifier The qualifier.
     * @return The description of this comparison type.
     */
    public String describe(String qualifier) {
        return toString() + " \"" + qualifier + '"';
    }

    /**
     * Performs a comparison.
     *
     * @param qualifier The qualifier to use in conjunction with this {@link StringCompareType}.
     * @param data      The data to check.
     * @return Whether the data matches the criteria or not.
     */
    public abstract boolean matches(String qualifier, String data);
}
