/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.settings;

import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;

public enum DamageProgression {
    BASIC_SET {
        @Override
        public String toString() {
            return I18n.text("Basic Set");
        }

        @Override
        public Dice calculateThrust(int strength) {
            int value = strength;
            if (value < 19) {
                return new Dice(1, -(6 - (value - 1) / 2));
            }
            value -= 11;
            if (strength > 50) {
                value--;
                if (strength > 79) {
                    value -= 1 + (strength - 80) / 5;
                }
            }
            return new Dice(value / 8 + 1, value % 8 / 2 - 1);
        }

        @Override
        public Dice calculateSwing(int strength) {
            int value = strength;
            if (value < 10) {
                return new Dice(1, -(5 - (value - 1) / 2));
            }
            if (value < 28) {
                value -= 9;
                return new Dice(value / 4 + 1, value % 4 - 1);
            }
            if (strength > 40) {
                value -= (strength - 40) / 5;
            }
            if (strength > 59) {
                value++;
            }
            value += 9;
            return new Dice(value / 8 + 1, value % 8 / 2 - 1);
        }
    },
    KNOWING_YOUR_OWN_STRENGTH {
        @Override
        public String toString() {
            return I18n.text("Knowing Your Own Strength");
        }

        @Override
        public String getFootnote() {
            return I18n.text("Pyramid 3-83, pages 16-19");
        }

        @Override
        public Dice calculateThrust(int strength) {
            if (strength < 12) {
                return new Dice(1, strength - 12);
            }
            return new Dice((strength - 7) / 4, (strength + 1) % 4 - 1);
        }

        @Override
        public Dice calculateSwing(int strength) {
            if (strength < 10) {
                return new Dice(1, strength - 10);
            }
            return new Dice((strength - 5) / 4, (strength - 1) % 4 - 1);
        }
    },
    NO_SCHOOL_GROGNARD_DAMAGE {
        @Override
        public String toString() {
            return I18n.text("No School Grognard");
        }

        @Override
        public String getFootnote() {
            return I18n.text("https://noschoolgrognard.blogspot.com/2013/04/adjusting-swing-damage-in-dungeon.html");
        }

        @Override
        public Dice calculateThrust(int strength) {
            if (strength < 11) {
                return new Dice(1, -(14 - strength) / 2);
            }
            strength -= 11;
            return new Dice(strength / 8 + 1, (strength % 8) / 2 - 1);
        }

        @Override
        public Dice calculateSwing(int strength) {
            return calculateThrust(strength + 3);
        }
    },
    THRUST_EQUALS_SWING_MINUS_2 {
        @Override
        public String toString() {
            return I18n.text("Thrust = Swing-2");
        }

        @Override
        public String getFootnote() {
            return I18n.text("https://github.com/richardwilkes/gcs/issues/97");
        }

        @Override
        public Dice calculateThrust(int strength) {
            Dice dice = calculateSwing(strength);
            dice.add(-2);
            return dice;
        }

        @Override
        public Dice calculateSwing(int strength) {
            return BASIC_SET.calculateSwing(strength);
        }
    },
    SWING_EQUALS_THRUST_PLUS_2 {
        @Override
        public String toString() {
            return I18n.text("Swing = Thrust+2");
        }

        @Override
        public String getFootnote() {
            return I18n.text("Houserule originating with Kevin Smyth. See https://gamingballistic.com/2020/12/04/df-eastmarch-boss-fight-and-house-rules/");
        }

        @Override
        public Dice calculateThrust(int strength) {
            return BASIC_SET.calculateThrust(strength);
        }

        @Override
        public Dice calculateSwing(int strength) {
            Dice dice = calculateThrust(strength);
            dice.add(2);
            return dice;
        }
    },
    PHOENIX_FLAME_D3 {
        @Override
        public String toString() {
            return I18n.text("PhoenixFlame d3");
        }

        @Override
        public String getFootnote() {
            return I18n.text("Houserule that use d3s instead of d6s for Damage. See: https://github.com/richardwilkes/gcs/pull/393");
        }

        @Override
        public Dice calculateThrust(int strength) {
            if (strength < 7) {
                return new Dice(1, 6, ((Math.max(strength, 1) + 1) / 2) - 7, 1);
            }
            if (strength < 10) {
                return new Dice(1, 3, ((strength + 1) / 2) - 5, 1);
            }
            int base = strength - 8;
            return new Dice(base / 2, 3, base % 2, 1);
        }

        @Override
        public Dice calculateSwing(int strength) {
            return calculateThrust(strength);
        }
    };

    public String getTooltip() {
        String tooltip  = I18n.text("Determines the method used to calculate thrust and swing damage");
        String footnote = getFootnote();
        if (footnote != null && !footnote.isBlank()) {
            return tooltip + ".\n" + footnote;
        }
        return tooltip;
    }

    public String getFootnote() {
        return null;
    }

    public abstract Dice calculateThrust(int strength);

    public abstract Dice calculateSwing(int strength);
}
