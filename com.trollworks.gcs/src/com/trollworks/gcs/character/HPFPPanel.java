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

package com.trollworks.gcs.character;

import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.page.PagePoints;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.notification.NotifierTarget;

import java.awt.Color;
import javax.swing.SwingConstants;

public abstract class HPFPPanel extends DropPanel implements NotifierTarget {
    protected CharacterSheet mSheet;

    protected HPFPPanel(CharacterSheet sheet, String title) {
        super(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(2, 0).setAlignment(PrecisionLayoutAlignment.FILL, PrecisionLayoutAlignment.FILL), title);
        mSheet = sheet;
    }

    protected PageField addLabelAndField(CharacterSheet sheet, String key, String title, String tooltip, boolean enabled) {
        PagePoints points = new PagePoints(sheet, key);
        PageField  field  = new PageField(sheet, key, SwingConstants.RIGHT, enabled, tooltip);
        PageLabel  label  = new PageLabel(title, field);
        field.putClientProperty(PagePoints.class, points);
        field.putClientProperty(PageLabel.class, label);
        add(points, new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(label);
        return field;
    }

    protected abstract void adjustColors();

    protected void adjustColor(PageField field, boolean add) {
        Color textColor;
        if (add) {
            addHorizontalBackground(field, ThemeColor.WARN);
            textColor = ThemeColor.ON_WARN;
        } else {
            removeHorizontalBackground(field);
            textColor = ThemeColor.ON_PAGE;
        }
        PagePoints point = (PagePoints) field.getClientProperty(PagePoints.class);
        point.setForeground(textColor);
        field.setDisabledTextColor(textColor);
        PageLabel label = (PageLabel) field.getClientProperty(PageLabel.class);
        label.setForeground(textColor);
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        adjustColors();
    }
}
