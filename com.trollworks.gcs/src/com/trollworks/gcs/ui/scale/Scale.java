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

package com.trollworks.gcs.ui.scale;

import com.trollworks.gcs.ui.UIUtilities;

import java.awt.Component;
import java.awt.Font;
import java.awt.Insets;

/** Provides convenience for scaling. */
public class Scale {
    private static Scale  OVERRIDE;
    private        double mScale;

    public static void setOverride(Scale scale) {
        OVERRIDE = scale;
    }

    /**
     * @param comp The component to determine the scale for.
     * @return The scale.
     */
    public static Scale get(Component comp) {
        if (OVERRIDE != null) {
            return OVERRIDE;
        }
        ScaleRoot root  = UIUtilities.getSelfOrAncestorOfType(comp, ScaleRoot.class);
        Scale     scale = null;
        if (root != null) {
            scale = root.getScale();
        }
        return scale != null ? scale : Scales.ACTUAL_SIZE.getScale();
    }

    /**
     * Creates a new scale.
     *
     * @param scale The scale to use, where 1.0 is 100%, 2.0 is 200%, etc.
     */
    public Scale(double scale) {
        mScale = scale;
    }

    /** @return The scale, where 1.0 is 100%, 2.0 is 200%, etc. */
    public double getScale() {
        return mScale;
    }

    /**
     * @param font The font to scale.
     * @return The scaled font.
     */
    public Font scale(Font font) {
        if (mScale == 1) {
            return font;
        }
        return font.deriveFont((float) (font.getSize() * mScale));
    }

    /**
     * @param insets The insets to scale.
     * @return The scaled insets.
     */
    public Insets scale(Insets insets) {
        if (mScale == 1) {
            return insets;
        }
        return new Insets((int) (insets.top * mScale), (int) (insets.left * mScale), (int) (insets.bottom * mScale), (int) (insets.right * mScale));
    }

    /**
     * @param values The values to scale.
     * @return The scaled values.
     */
    public double[] scale(double[] values) {
        if (mScale == 1) {
            return values;
        }
        int      length = values.length;
        double[] scaled = new double[length];
        for (int i = 0; i < length; i++) {
            scaled[i] = values[i] * mScale;
        }
        return scaled;
    }

    /**
     * @param value The value to scale.
     * @return The scaled value.
     */
    public int scale(int value) {
        return (int) (value * mScale);
    }

    /**
     * @param value The value to scale.
     * @return The scaled value.
     */
    public float scale(float value) {
        return (float) (value * mScale);
    }

    /**
     * @param value The value to scale.
     * @return The scaled value.
     */
    public double scale(double value) {
        return value * mScale;
    }
}
