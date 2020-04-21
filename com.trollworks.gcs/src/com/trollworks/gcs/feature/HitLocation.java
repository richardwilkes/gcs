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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.utility.I18n;

import java.util.ArrayList;

/** Hit locations. */
public enum HitLocation {
    /** The skull hit location. */
    SKULL {
        @Override
        public String toString() {
            return I18n.Text("to the skull");
        }
    },
    /** The eyes hit location. */
    EYES {
        @Override
        public String toString() {
            return I18n.Text("to the eyes");
        }
    },
    /** The face hit location. */
    FACE {
        @Override
        public String toString() {
            return I18n.Text("to the face");
        }
    },
    /** The neck hit location. */
    NECK {
        @Override
        public String toString() {
            return I18n.Text("to the neck");
        }
    },
    /** The torso hit location. */
    TORSO {
        @Override
        public String toString() {
            return I18n.Text("to the torso");
        }
    },
    /** The vitals hit location. */
    VITALS {
        @Override
        public String toString() {
            return I18n.Text("to the vitals");
        }

        @Override
        public boolean isChoosable() {
            return false;
        }
    },
    /** The groin hit location. */
    GROIN {
        @Override
        public String toString() {
            return I18n.Text("to the groin");
        }
    },
    /** The arm hit location. */
    ARMS {
        @Override
        public String toString() {
            return I18n.Text("to the arms");
        }
    },
    /** The hand hit location. */
    HANDS {
        @Override
        public String toString() {
            return I18n.Text("to the hands");
        }
    },
    /** The leg hit location. */
    LEGS {
        @Override
        public String toString() {
            return I18n.Text("to the legs");
        }
    },
    /** The foot hit location. */
    FEET {
        @Override
        public String toString() {
            return I18n.Text("to the feet");
        }
    },
    /** The tail hit location. */
    TAIL {
        @Override
        public String toString() {
            return I18n.Text("to the tail");
        }
    },
    /** The wing hit location. */
    WINGS {
        @Override
        public String toString() {
            return I18n.Text("to the wings");
        }
    },
    /** The fin hit location. */
    FINS {
        @Override
        public String toString() {
            return I18n.Text("to the fins");
        }
    },
    /** The brain hit location. */
    BRAIN {
        @Override
        public String toString() {
            return I18n.Text("to the brain");
        }
    },
    /** The full body hit location. */
    FULL_BODY {
        @Override
        public String toString() {
            return I18n.Text("to the full body");
        }
    },
    /** The full body except eyes hit location. */
    FULL_BODY_EXCEPT_EYES {
        @Override
        public String toString() {
            return I18n.Text("to the full body except the eyes");
        }
    };

    /** @return The hit locations that can be chosen as an armor protection spot. */
    public static HitLocation[] getChoosableLocations() {
        ArrayList<HitLocation> list = new ArrayList<>();
        for (HitLocation one : values()) {
            if (one.isChoosable()) {
                list.add(one);
            }
        }
        return list.toArray(new HitLocation[0]);
    }

    /** @return Whether this location is choosable as an armor protection spot. */
    @SuppressWarnings("static-method")
    public boolean isChoosable() {
        return true;
    }
}
