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
        createLabelAndDisabledField(sheet, GURPSCharacter.ID_CREATED, I18n.Text("Created:"), null);
        createLabelAndDisabledField(sheet, GURPSCharacter.ID_MODIFIED, I18n.Text("Modified:"), null);
        createLabelAndDisabledField(sheet, Settings.PREFIX, I18n.Text("Options:"), Text.wrapPlainTextForToolTip(I18n.Text("Each letter represents an optional rule. A uppercase letter indicates the rule is in use while a lowercase letter indicates the rule is not in use.")));
    }

    private void createLabelAndDisabledField(CharacterSheet sheet, String key, String title, String tooltip) {
        PageField field = new PageField(sheet, key, SwingConstants.LEFT, false, tooltip);
        add(new PageLabel(title, field), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }
}
