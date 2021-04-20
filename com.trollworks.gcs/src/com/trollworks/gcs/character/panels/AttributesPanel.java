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

import com.trollworks.gcs.character.CharacterSetter;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.attribute.Attribute;
import com.trollworks.gcs.character.attribute.AttributeDef;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.page.PagePoints;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.DiceFormatter;

import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

/** The character attributes panel. */
public class AttributesPanel extends DropPanel {
    /**
     * Creates a new attributes panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public AttributesPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(2, 0).setAlignment(PrecisionLayoutAlignment.FILL, PrecisionLayoutAlignment.FILL), I18n.Text("Attributes"));
        GURPSCharacter gch = sheet.getCharacter();
        for (AttributeDef def : AttributeDef.getOrdered(gch.getSettings().getAttributes())) {
            createEditableAttributeField(sheet, gch, def);
        }
        addDivider();
        createEditableIntegerField(sheet, Integer.valueOf(gch.getWillPoints()), gch.getWillAdj(), (c, v) -> c.setWillAdj(((Integer) v).intValue()), "Will", I18n.TextWithContext(1, "Will"));
        createField(sheet, gch.getFrightCheck(), I18n.Text("Fright Check"));
        addDivider();
        createEditableFloatField(sheet, Integer.valueOf(gch.getBasicSpeedPoints()), gch.getBasicSpeed(), (c, v) -> c.setBasicSpeed(((Double) v).doubleValue()), "basic speed", I18n.Text("Basic Speed"));
        createEditableIntegerField(sheet, Integer.valueOf(gch.getBasicMovePoints()), gch.getBasicMove(), (c, v) -> c.setBasicMove(((Integer) v).intValue()), "basic move", I18n.Text("Basic Move"));
        addDivider();
        createEditableIntegerField(sheet, Integer.valueOf(gch.getPerceptionPoints()), gch.getPerAdj(), (c, v) -> c.setPerAdj(((Integer) v).intValue()), "Per", I18n.Text("Perception (Per)"));
        createField(sheet, gch.getVision(), I18n.Text("Vision"));
        createField(sheet, gch.getHearing(), I18n.Text("Hearing"));
        createField(sheet, gch.getTasteAndSmell(), I18n.Text("Taste & Smell"));
        createField(sheet, gch.getTouch(), I18n.Text("Touch"));
        addDivider();
        createDiceField(sheet, gch.getThrust(), I18n.Text("Basic Thrust"));
        createDiceField(sheet, gch.getSwing(), I18n.Text("Basic Swing"));
    }

    private void createEditableAttributeField(CharacterSheet sheet, GURPSCharacter gch, AttributeDef def) {
        Attribute attr = gch.getAttributes().get(def.getID());
        createEditableIntegerField(sheet, Integer.valueOf(attr.getPointCost(gch)), attr.getValue(gch), (c, v) -> attr.setValue(c, ((Integer) v).intValue()), attr.getAttrID(), String.format("%s (%s)", def.getDescription(), def.getName()));
    }

    private void createEditableIntegerField(CharacterSheet sheet, Integer points, int value, CharacterSetter setter, String tag, String title) {
        PageField field = new PageField(FieldFactory.POSINT5, Integer.valueOf(value), setter, sheet, tag, SwingConstants.RIGHT, true, null, ThemeColor.ON_PAGE);
        add(points != null ? new PagePoints(points.intValue()) : new Label(), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, field));
    }

    private void createEditableFloatField(CharacterSheet sheet, Integer points, double value, CharacterSetter setter, String tag, String title) {
        PageField field = new PageField(FieldFactory.FLOAT, Double.valueOf(value), setter, sheet, tag, SwingConstants.RIGHT, true, null, ThemeColor.ON_PAGE);
        add(points != null ? new PagePoints(points.intValue()) : new Label(), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, field));
    }

    private void createField(CharacterSheet sheet, int value, String title) {
        PageField field = new PageField(FieldFactory.POSINT5, Integer.valueOf(value), sheet, SwingConstants.RIGHT, null, ThemeColor.ON_PAGE);
        add(new Label(), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, field));
    }

    private void createDiceField(CharacterSheet sheet, Dice dice, String title) {
        PageField field = new PageField(new DefaultFormatterFactory(new DiceFormatter(sheet.getCharacter())), dice, sheet, SwingConstants.RIGHT, null, ThemeColor.ON_PAGE);
        add(new Label(), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, field));
    }

    private void addDivider() {
        Wrapper panel = new Wrapper();
        add(panel, new PrecisionLayoutData().setHorizontalSpan(3).setHeightHint(1));
        addHorizontalBackground(panel, ThemeColor.ON_PAGE);
    }
}
