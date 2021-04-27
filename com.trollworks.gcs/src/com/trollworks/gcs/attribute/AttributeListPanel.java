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

import java.awt.Component;
import java.awt.Dimension;
import java.awt.Rectangle;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

public class AttributeListPanel extends BandedPanel {
    private Map<String, AttributeDef> mAttributes;
    private Runnable                  mAdjustCallback;

    public AttributeListPanel(Map<String, AttributeDef> attributes, Runnable adjustCallback) {
        super("attribute-list");
        setLayout(new PrecisionLayout());
        mAttributes = attributes;
        mAdjustCallback = adjustCallback;
        fillAttributes();
    }

    @Override
    public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
        return 16;
    }

    @Override
    public Dimension getPreferredScrollableViewportSize() {
        return new Dimension(32, 32); // This needs to be small to allow the scroll pane to work
    }

    @Override
    public boolean getScrollableTracksViewportWidth() {
        return true;
    }

    public void reset(Map<String, AttributeDef> attributes) {
        removeAll();
        mAttributes.clear();
        mAttributes.putAll(attributes);
        fillAttributes();
    }

    private void fillAttributes() {
        for (AttributeDef def : AttributeDef.getOrdered(mAttributes)) {
            add(new AttributePanel(mAttributes, def, mAdjustCallback), new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        }
        adjustButtons();
    }

    public void adjustButtons() {
        Component[] children = getComponents();
        int         count    = children.length;
        for (int i = 0; i < count; i++) {
            ((AttributePanel) children[i]).adjustButtons(i == 0, i == count - 1);
        }
    }

    public void renumber() {
        Component[] children = getComponents();
        int         count    = children.length;
        for (int i = 0; i < count; i++) {
            AttributePanel attrPanel = (AttributePanel) children[i];
            attrPanel.mAttrDef.setOrder(i + 1);
            attrPanel.adjustButtons(i == 0, i == count - 1);
        }
        repaint();
    }

    public void addAttribute() {
        String id    = "id";
        int    order = 0;
        if (!mAttributes.isEmpty()) {
            List<AttributeDef> ordered = AttributeDef.getOrdered(mAttributes);
            order = ordered.get(ordered.size() - 1).getOrder() + 1;
            Set<String> ids = new HashSet<>();
            for (AttributeDef def : ordered) {
                ids.add(def.getID());
            }
            int counter = 1;
            while (ids.contains(id)) {
                id = "id" + ++counter;
            }
        }
        AttributeDef def = new AttributeDef(id, AttributeType.INTEGER, "name", "description", "10", order, 0, 0);
        mAttributes.put(def.getID(), def);
        mAdjustCallback.run();
        add(new AttributePanel(mAttributes, def, mAdjustCallback), new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        scrollRectToVisible(new Rectangle(0, getPreferredSize().height - 1, 1, 1));
        ((AttributePanel) getComponent(getComponentCount() - 1)).focusIDField();
        adjustButtons();
    }
}
