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
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.page.PagePoints;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.text.DiceFormatter;

import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

/** The character attributes panel. */
public class AttributesPanel extends DropPanel {
    /**
     * Creates a new attributes panel.
     *
     * @param sheet   The sheet to display the data for.
     * @param primary {@code true} if the primary attributes should be shown, otherwise the
     *                secondary ones will be. "primary" here is defined as those whose attribute
     *                base is a number and not a reference to another attribute.
     */
    public AttributesPanel(CharacterSheet sheet, boolean primary) {
        super(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(2, 0).setAlignment(PrecisionLayoutAlignment.FILL, PrecisionLayoutAlignment.FILL), primary ? I18n.text("Primary Attributes") : I18n.text("Secondary Attributes"));
        GURPSCharacter gch = sheet.getCharacter();
        for (AttributeDef def : AttributeDef.getOrdered(gch.getSheetSettings().getAttributes())) {
            if (def.getType() != AttributeType.POOL) {
                if (def.isPrimary() == primary) {
                    createAttributeField(sheet, gch, def);
                }
            }
        }
        if (primary) {
            addDivider();
            createDiceField(sheet, gch.getThrust(), I18n.text("Basic Thrust"));
            createDiceField(sheet, gch.getSwing(), I18n.text("Basic Swing"));
        }
    }

    private void createAttributeField(CharacterSheet sheet, GURPSCharacter gch, AttributeDef def) {
        Attribute attr = gch.getAttributes().get(def.getID());
        if (attr == null) {
            Log.error(String.format("unable to locate attribute data for '%s'", def.getID()));
            return;
        }
        PageField field;
        if (def.getType() == AttributeType.DECIMAL) {
            field = new PageField(FieldFactory.FLOAT, Double.valueOf(attr.getDoubleValue(gch)), (c, v) -> attr.setDoubleValue(c, ((Double) v).doubleValue()), sheet, Attribute.ID_ATTR_PREFIX + attr.getID(), SwingConstants.RIGHT, true, null);
        } else {
            field = new PageField(FieldFactory.POSINT5, Integer.valueOf(attr.getIntValue(gch)), (c, v) -> attr.setIntValue(c, ((Integer) v).intValue()), sheet, Attribute.ID_ATTR_PREFIX + attr.getID(), SwingConstants.RIGHT, true, null);
        }
        add(new PagePoints(attr.getPointCost(gch)), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(def.getCombinedName(), field));
    }

    private void createDiceField(CharacterSheet sheet, Dice dice, String title) {
        PageField field = new PageField(new DefaultFormatterFactory(new DiceFormatter(sheet.getCharacter())), dice, sheet, SwingConstants.RIGHT, null);
        add(field, new PrecisionLayoutData().setHorizontalSpan(2).setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, field));
    }

    private void addDivider() {
        Wrapper panel = new Wrapper();
        add(panel, new PrecisionLayoutData().setHorizontalSpan(3).setHeightHint(1).setMargins(3, 0, 2, 0));
        addHorizontalBackground(panel, ThemeColor.ON_CONTENT);
    }
}
