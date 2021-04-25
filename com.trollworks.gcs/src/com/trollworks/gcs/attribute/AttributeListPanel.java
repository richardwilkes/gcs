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

import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BandedPanel;

import java.awt.Dimension;
import java.awt.Rectangle;
import java.util.Map;

public class AttributeListPanel extends BandedPanel {
    private Map<String, AttributeDef> mAttributes;
    private Runnable                  mAdjustCallback;

    public AttributeListPanel(Map<String, AttributeDef> attributes, Runnable adjustCallback) {
        super("attribute-list");
        setLayout(new PrecisionLayout());
        mAttributes = attributes;
        mAdjustCallback = adjustCallback;
        addAttributes();
    }

    @Override
    public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
        return 16;
    }

    @Override
    public Dimension getPreferredScrollableViewportSize() {
        return new Dimension(32, 32); // This needs to be small to allow the scroll pane to work
    }

    public void reset(Map<String, AttributeDef> attributes) {
        removeAll();
        mAttributes.clear();
        mAttributes.putAll(attributes);
        addAttributes();
    }

    private void addAttributes() {
        for (AttributeDef def : AttributeDef.getOrdered(mAttributes)) {
            add(new AttributePanel(def, mAdjustCallback), new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        }
    }
}
