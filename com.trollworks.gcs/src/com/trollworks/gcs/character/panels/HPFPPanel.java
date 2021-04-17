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

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.CharacterSetter;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.page.PagePoints;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Label;

import java.awt.Color;
import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;

public abstract class HPFPPanel extends DropPanel {
    protected HPFPPanel(String title) {
        super(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(2, 0).setAlignment(PrecisionLayoutAlignment.FILL, PrecisionLayoutAlignment.FILL), title);
    }

    protected void createEditableField(CharacterSheet sheet, Integer points, JFormattedTextField.AbstractFormatterFactory formatter, int value, CharacterSetter setter, String tag, String title, String tooltip) {
        PageField field = new PageField(formatter, Integer.valueOf(value), setter, sheet, tag, SwingConstants.RIGHT, true, tooltip, ThemeColor.ON_PAGE);
        add(points != null ? new PagePoints(points.intValue()) : new Label(), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, field));
    }

    protected void createField(CharacterSheet sheet, int value, String title, String tooltip, boolean warn) {
        Color     color = warn ? ThemeColor.ON_WARN : ThemeColor.ON_PAGE;
        PageField field = new PageField(FieldFactory.INT7, Integer.valueOf(value), sheet, SwingConstants.RIGHT, tooltip, color);
        add(new Label(), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, color, field));
        if (warn) {
            addHorizontalBackground(field, ThemeColor.WARN);
        }
    }
}
