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

package com.trollworks.gcs.skill;

/** The possible skill difficulty levels. */
public enum SkillDifficulty {
    /** The "easy" difficulty. */
    E {
        @Override
        public String toString() {
            return "E";
        }
    },
    /** The "average" difficulty. */
    A {
        @Override
        public String toString() {
            return "A";
        }
    },
    /** The "hard" difficulty. */
    H {
        @Override
        public String toString() {
            return "H";
        }
    },
    /** The "very hard" difficulty. */
    VH {
        @Override
        public String toString() {
            return "VH";
        }
    },
    /** The "wildcard" difficulty. */
    W {
        @Override
        public String toString() {
            return "W";
        }

        @Override
        public int getBaseRelativeLevel() {
            return VH.getBaseRelativeLevel();
        }
    };

    /** @return The base relative skill level at 0 points. */
    public int getBaseRelativeLevel() {
        return -ordinal();
    }
}
