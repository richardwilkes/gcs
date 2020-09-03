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

/** The attribute affected by a {@link AttributeBonus}. */
public enum BonusAttributeType {
    /** The ST attribute. */
    ST {
        @Override
        public String toString() {
            return I18n.Text("ST");
        }
    },
    /** The DX attribute. */
    DX {
        @Override
        public String toString() {
            return I18n.Text("DX");
        }
    },
    /** The IQ attribute. */
    IQ {
        @Override
        public String toString() {
            return I18n.Text("IQ");
        }
    },
    /** The HT attribute. */
    HT {
        @Override
        public String toString() {
            return I18n.Text("HT");
        }
    },
    /** The Will attribute. */
    WILL {
        @Override
        public String toString() {
            return I18n.Text("will");
        }
    },
    /** The Fright Check attribute. */
    FRIGHT_CHECK {
        @Override
        public String toString() {
            return I18n.Text("fright checks");
        }
    },
    /** The Perception attribute. */
    PERCEPTION {
        @Override
        public String toString() {
            return I18n.Text("perception");
        }
    },
    /** The Vision attribute. */
    VISION {
        @Override
        public String toString() {
            return I18n.Text("vision");
        }
    },
    /** The Hearing attribute. */
    HEARING {
        @Override
        public String toString() {
            return I18n.Text("hearing");
        }
    },
    /** The TasteSmell attribute. */
    TASTE_SMELL {
        @Override
        public String toString() {
            return I18n.Text("taste & smell");
        }
    },
    /** The Touch attribute. */
    TOUCH {
        @Override
        public String toString() {
            return I18n.Text("touch");
        }
    },
    /** The Dodge attribute. */
    DODGE {
        @Override
        public String toString() {
            return I18n.Text("dodge");
        }
    },
    /** The Parry attribute. */
    PARRY {
        @Override
        public String toString() {
            return I18n.Text("parry");
        }
    },
    /** The Block attribute. */
    BLOCK {
        @Override
        public String toString() {
            return I18n.Text("block");
        }
    },
    /** The Speed attribute. */
    SPEED {
        @Override
        public String toString() {
            return I18n.Text("basic speed");
        }

        @Override
        public boolean isIntegerOnly() {
            return false;
        }
    },
    /** The Move attribute. */
    MOVE {
        @Override
        public String toString() {
            return I18n.Text("basic move");
        }
    },
    /** The FP attribute. */
    FP {
        @Override
        public String toString() {
            return I18n.Text("FP");
        }
    },
    /** The HP attribute. */
    HP {
        @Override
        public String toString() {
            return I18n.Text("HP");
        }
    },
    /** The size modifier attribute. */
    SM {
        @Override
        public String toString() {
            return I18n.Text("size modifier");
        }
    };

    /** @return The presentation name. */
    public String getPresentationName() {
        String name = name();
        if (name.length() > 2) {
            name = Character.toUpperCase(name.charAt(0)) + name.substring(1).toLowerCase();
        }
        return name;
    }

    /** @return {@code true} if only integer values are permitted. */
    @SuppressWarnings("static-method")
    public boolean isIntegerOnly() {
        return true;
    }
}
