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
import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.character.names.USCensusNames;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.utility.I18n;

import javax.swing.SwingConstants;

/** The character identity panel. */
public class IdentityPanel extends DropPanel {
    private PageField mNameField;

    /**
     * Creates a new identity panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public IdentityPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0), I18n.text("Identity"));
        Profile profile = sheet.getCharacter().getProfile();
        mNameField = createRandomizableField(sheet, profile.getName(), I18n.text("Name"), "character name",
                (c, v) -> c.getProfile().setName((String) v), () -> {
                    mNameField.requestFocus();
                    profile.setName(USCensusNames.INSTANCE.getFullName(!profile.getGender().equalsIgnoreCase(I18n.text("Female"))));
                });
        createStringField(sheet, profile.getTitle(), I18n.text("Title"), "character title",
                (c, v) -> c.getProfile().setTitle((String) v));
        createStringField(sheet, profile.getOrganization(), I18n.text("Organization"), "organization",
                (c, v) -> c.getProfile().setOrganization((String) v));
    }

    private PageField createRandomizableField(CharacterSheet sheet, String value, String title, String tag, CharacterSetter setter, Runnable randomizer) {
        add(new FontAwesomeButton("\uf074", ThemeFont.PAGE_LABEL_PRIMARY.getFont().getSize() * 8 / 10,
                String.format(I18n.text("Randomize %s"), title), randomizer));
        add(new PageLabel(title), new PrecisionLayoutData().setEndHorizontalAlignment());
        PageField field = new PageField(FieldFactory.STRING, value, setter, sheet, tag,
                SwingConstants.LEFT, true, null);
        add(field, createFieldLayoutData());
        return field;
    }

    private void createStringField(CharacterSheet sheet, String value, String title, String tag, CharacterSetter setter) {
        add(new PageLabel(title), new PrecisionLayoutData().setEndHorizontalAlignment().setHorizontalSpan(2));
        add(new PageField(FieldFactory.STRING, value, setter, sheet, tag, SwingConstants.LEFT, true,
                null), createFieldLayoutData());
    }

    private static PrecisionLayoutData createFieldLayoutData() {
        return new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setLeftMargin(4);
    }
}
