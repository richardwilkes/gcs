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
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import javax.swing.SwingConstants;

/** The miscellaneous info panel. */
public class MiscPanel extends DropPanel {
    /**
     * Creates a new miscellaneous info panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public MiscPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(2).setMargins(0).setSpacing(4, 0), I18n.Text("Miscellaneous"));
        GURPSCharacter gch = sheet.getCharacter();
        createTimestampField(sheet, gch.getCreatedOn(), I18n.Text("Created:"));
        createTimestampField(sheet, gch.getModifiedOn(), I18n.Text("Modified:"));
        createStringField(sheet, gch.getSettings().optionsCode(), I18n.Text("Options:"), Text.wrapPlainTextForToolTip(I18n.Text("Each letter represents an optional rule. A uppercase letter indicates the rule is in use while a lowercase letter indicates the rule is not in use.")));
    }

    private void createTimestampField(CharacterSheet sheet, long timeStampseconds, String title) {
        add(new PageLabel(title, null), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(new PageField(FieldFactory.DATETIME, Long.valueOf(timeStampseconds), sheet, SwingConstants.LEFT, null, ThemeColor.ON_PAGE), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void createStringField(CharacterSheet sheet, String value, String title, String tooltip) {
        PageField field = new PageField(FieldFactory.STRING, value, sheet, SwingConstants.LEFT, tooltip, ThemeColor.ON_PAGE);
        add(new PageLabel(title, field), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }
}
