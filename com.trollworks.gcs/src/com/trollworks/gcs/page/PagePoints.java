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

package com.trollworks.gcs.page;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.Text;

import javax.swing.UIManager;

/** A points field in a page. */
public class PagePoints extends Label implements NotifierTarget {
    private CharacterSheet mSheet;
    private String         mConsumedType;

    /**
     * Creates a new points field.
     *
     * @param sheet        The sheet to listen to.
     * @param consumedType The field to listen to.
     */
    public PagePoints(CharacterSheet sheet, String consumedType) {
        super(getFormattedValue(sheet, consumedType));
        mSheet = sheet;
        mConsumedType = consumedType;
        setFont(UIManager.getFont(Fonts.KEY_LABEL_SECONDARY));
        setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Points spent")));
        UIUtilities.setToPreferredSizeOnly(this);
        mSheet.getCharacter().addTarget(this, consumedType);
    }

    private static String getFormattedValue(CharacterSheet sheet, String consumedType) {
        Object value = sheet.getCharacter().getValueForID(GURPSCharacter.POINTS_PREFIX + consumedType);
        return value != null ? "[" + value + "]" : "";
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        setText(getFormattedValue(mSheet, mConsumedType));
        setPreferredSize(null);
        UIUtilities.setToPreferredSizeOnly(this);
        invalidate();
        repaint();
    }
}
