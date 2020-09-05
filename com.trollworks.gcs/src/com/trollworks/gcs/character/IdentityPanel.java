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

/** The character identity panel. */
public class IdentityPanel extends DropPanel {
    /**
     * Creates a new identity panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public IdentityPanel(CharacterSheet sheet) {
        super(new ColumnLayout(2, 2, 0), I18n.Text("Identity"));
        createLabelAndField(this, sheet, Profile.ID_NAME, I18n.Text("Name:"), null);
        createLabelAndField(this, sheet, Profile.ID_TITLE, I18n.Text("Title:"), null);
        createLabelAndField(this, sheet, Profile.ID_PLAYER_NAME, I18n.Text("Player:"), null);
    }
}
