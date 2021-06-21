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
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;

import javax.swing.SwingConstants;

/** The miscellaneous info panel. */
public class MiscPanel extends DropPanel {
    /**
     * Creates a new miscellaneous info panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public MiscPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(2).setMargins(0).setSpacing(4, 0), I18n.text("Miscellaneous"));
        GURPSCharacter gch = sheet.getCharacter();
        createTimestampField(sheet, gch.getCreatedOn(), I18n.text("Created"));
        createTimestampField(sheet, gch.getModifiedOn(), I18n.text("Modified"));
        createStringField(sheet, gch.getProfile().getPlayerName(), I18n.text("Player"), "player", (c, v) -> c.getProfile().setPlayerName((String) v));
    }

    private void createTimestampField(CharacterSheet sheet, long timeStampseconds, String title) {
        add(new PageLabel(title), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(new PageField(FieldFactory.DATETIME, Long.valueOf(timeStampseconds), sheet, SwingConstants.LEFT, null), createFieldLayoutData());
    }

    private void createStringField(CharacterSheet sheet, String value, String title, String tag, CharacterSetter setter) {
        add(new PageLabel(title), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(new PageField(FieldFactory.STRING, value, setter, sheet, tag, SwingConstants.LEFT, true, null), createFieldLayoutData());
    }

    private static PrecisionLayoutData createFieldLayoutData() {
        return new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true);
    }
}
