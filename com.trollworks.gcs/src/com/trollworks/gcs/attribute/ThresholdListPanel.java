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

package com.trollworks.gcs.attribute;

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BandedPanel;

import java.awt.Color;
import java.awt.Component;
import java.util.List;

public class ThresholdListPanel extends BandedPanel {
    private AttributeDef mAttrDef;
    private Runnable     mAdjustCallback;

    public ThresholdListPanel(AttributeDef attrDef, Runnable adjustCallback) {
        super(false);
        setLayout(new PrecisionLayout().setMargins(0));
        setBorder(new LineBorder(Color.LIGHT_GRAY));
        setBackground(Colors.adjustSaturation(ThemeColor.BANDING, -0.05f));
        mAttrDef = attrDef;
        mAdjustCallback = adjustCallback;
        fillThresholds();
    }

    @Override
    protected Color getBandingColor(boolean odd) {
        return Colors.adjustSaturation(super.getBandingColor(odd), 0.05f);
    }

    private void fillThresholds() {
        List<PoolThreshold> thresholds = mAttrDef.getThresholds();
        for (PoolThreshold threshold : thresholds) {
            add(new ThresholdPanel(thresholds, threshold, mAdjustCallback), new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        }
        adjustButtons();
    }

    public void adjustButtons() {
        Component[] children = getComponents();
        int         count    = children.length;
        for (int i = 0; i < count; i++) {
            ((ThresholdPanel) children[i]).adjustButtons(i == 0, i == count - 1);
        }
    }
}
