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

package com.trollworks.gcs.character.panels;

import com.trollworks.gcs.attribute.Attribute;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.attribute.AttributeType;
import com.trollworks.gcs.attribute.PoolThreshold;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.page.PagePoints;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;

import java.util.Map;
import java.util.Objects;
import javax.swing.SwingConstants;

public class PointPoolsPanel extends DropPanel {
    public PointPoolsPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(6).setMargins(0).setSpacing(2, 0).setFillAlignment(), I18n.text("Point Pools"));
        GURPSCharacter         gch        = sheet.getCharacter();
        Map<String, Attribute> attributes = gch.getAttributes();
        for (AttributeDef def : AttributeDef.getOrdered(gch.getSheetSettings().getAttributes())) {
            if (def.getType() == AttributeType.POOL) {
                Attribute attr = attributes.get(def.getID());
                if (attr != null) {
                    addPool(sheet, gch, def, attr);
                }
            }
        }
    }

    private void addPool(CharacterSheet sheet, GURPSCharacter gch, AttributeDef def, Attribute attr) {
        String id   = attr.getID();
        String name = def.getName();
        add(new PagePoints(attr.getPointCost(gch)), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(new PageField(FieldFactory.INT7, Integer.valueOf(attr.getCurrentIntValue(gch)), (c, v) -> attr.setDamage(gch, -Math.min(((Integer) v).intValue() - attr.getIntValue(gch), 0)), sheet, Attribute.ID_ATTR_PREFIX + id + ".current", SwingConstants.RIGHT, true, String.format(I18n.text("Current %s"), name)), new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(I18n.text("of")));
        add(new PageField(FieldFactory.POSINT6, Integer.valueOf(attr.getIntValue(gch)), (c, v) -> attr.setIntValue(gch, ((Integer) v).intValue()), sheet, Attribute.ID_ATTR_PREFIX + id, SwingConstants.RIGHT, true, String.format(I18n.text("Maximum %s"), name)), new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        PageLabel label    = new PageLabel(name);
        String    fullName = def.getFullName();
        label.setToolTipText(fullName.isBlank() ? null : fullName);
        add(label);
        PageLabel state = new PageLabel("");
        state.setThemeFont(ThemeFont.PAGE_LABEL_SECONDARY);
        PoolThreshold threshold = attr.getCurrentThreshold(gch);
        if (threshold != null) {
            state.setText(String.format("[%s]", threshold.getState()));
            String explanation = threshold.getExplanation();
            if (explanation.isEmpty()) {
                explanation = null;
            }
            if (!Objects.equals(explanation, state.getToolTipText())) {
                state.setToolTipText(explanation);
            }
        }
        add(state, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.END));
    }
}
