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

package com.trollworks.gcs.character;

import com.trollworks.gcs.character.names.USCensusNames;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.utility.I18n;

import javax.swing.SwingConstants;

/** The character identity panel. */
public class IdentityPanel extends DropPanel {
    /**
     * Creates a new identity panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public IdentityPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0), I18n.Text("Identity"));
        createLabelAndRandomizableField(sheet, Profile.ID_NAME, I18n.Text("Name:"));
        createLabelAndField(sheet, Profile.ID_TITLE, I18n.Text("Title:"));
        createLabelAndField(sheet, Profile.ID_PLAYER_NAME, I18n.Text("Player:"));
    }

    private void createLabelAndRandomizableField(CharacterSheet sheet, String key, String title) {
        PageField field = new PageField(sheet, key, SwingConstants.LEFT, true, null);
        IconButton button = new IconButton(Images.RANDOMIZE, null, () -> sheet.getCharacter().setValueForID(key, USCensusNames.INSTANCE.getFullName(!sheet.getCharacter().getProfile().getGender().equalsIgnoreCase(I18n.Text("Female")))));
        button.setToolTipText(I18n.Text("Create a new random name"));
        add(button);
        add(new PageLabel(title, field), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setLeftMargin(4));
    }

    private void createLabelAndField(CharacterSheet sheet, String key, String title) {
        PageField field = new PageField(sheet, key, SwingConstants.LEFT, true, null);
        add(new PageLabel(title, field), new PrecisionLayoutData().setEndHorizontalAlignment().setHorizontalSpan(2));
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setLeftMargin(4));
    }
}
