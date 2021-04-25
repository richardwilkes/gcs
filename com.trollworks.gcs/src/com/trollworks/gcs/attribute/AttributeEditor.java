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
import com.trollworks.gcs.utility.I18n;

import java.util.Map;

public class AttributeEditor extends BandedPanel {
    private Map<String, AttributeDef> mAttributes;
    private Runnable                  mAdjustCallback;

    public AttributeEditor(Map<String, AttributeDef> attributes, Runnable adjustCallback) {
        super(I18n.Text("Attributes"));
        setLayout(new PrecisionLayout());
        mAttributes = attributes;
        mAdjustCallback = adjustCallback;
        addAttributes();
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
        setPreferredSize(getPreferredSize());
    }
}
