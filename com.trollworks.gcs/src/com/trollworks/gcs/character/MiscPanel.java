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
import com.trollworks.gcs.ui.layout.ColumnLayout;
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
        super(new ColumnLayout(2, 2, 0), I18n.Text("Miscellaneous"), true);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_CREATED, I18n.Text("Created:"), null, SwingConstants.LEFT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_MODIFIED, I18n.Text("Modified:"), null, SwingConstants.LEFT);
        createLabelAndDisabledField(this, sheet, Settings.PREFIX, I18n.Text("Options:"), Text.wrapPlainTextForToolTip(I18n.Text("Each letter represents an optional rule. A uppercase letter indicates the rule is in use while a lowercase letter indicates the rule is not in use.")), SwingConstants.LEFT);
    }
}
