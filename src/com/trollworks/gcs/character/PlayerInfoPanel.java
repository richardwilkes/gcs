/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.utility.I18n;

import javax.swing.SwingConstants;

/** The character player info panel. */
public class PlayerInfoPanel extends DropPanel {
    /**
     * Creates a new player info panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public PlayerInfoPanel(CharacterSheet sheet) {
        super(new ColumnLayout(2, 2, 0), I18n.Text("Player Information"));
        createLabelAndField(this, sheet, Profile.ID_PLAYER_NAME, I18n.Text("Player:"), null, SwingConstants.LEFT);
        createLabelAndField(this, sheet, Profile.ID_CAMPAIGN, I18n.Text("Campaign:"), null, SwingConstants.LEFT);
        createLabelAndField(this, sheet, GURPSCharacter.ID_CREATED_ON, I18n.Text("Created On:"), null, SwingConstants.LEFT);
    }
}
