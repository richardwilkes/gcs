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
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;

/** The character description panel. */
public class DescriptionPanel extends DropPanel {
    /**
     * Creates a new description panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public DescriptionPanel(CharacterSheet sheet) {
        super(new ColumnLayout(5, 2, 0), I18n.Text("Description"));

        Wrapper wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
        createLabelAndField(wrapper, sheet, Profile.ID_GENDER, I18n.Text("Gender:"), null);
        createLabelAndField(wrapper, sheet, Profile.ID_AGE, I18n.Text("Age:"), null);
        createLabelAndField(wrapper, sheet, Profile.ID_BIRTHDAY, I18n.Text("Birthday:"), null);
        createLabelAndField(wrapper, sheet, Profile.ID_RELIGION, I18n.Text("Religion:"), null);
        add(wrapper);

        createDivider();

        wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
        createLabelAndField(wrapper, sheet, Profile.ID_HEIGHT, I18n.Text("Height:"), null);
        createLabelAndField(wrapper, sheet, Profile.ID_WEIGHT, I18n.Text("Weight:"), null);
        createLabelAndField(wrapper, sheet, Profile.ID_SIZE_MODIFIER, I18n.Text("Size:"), I18n.Text("The character's size modifier"));
        createLabelAndField(wrapper, sheet, Profile.ID_TECH_LEVEL, I18n.Text("TL:"), I18n.Text("<html><body>TL0: Stone Age<br>TL1: Bronze Age<br>TL2: Iron Age<br>TL3: Medieval<br>TL4: Age of Sail<br>TL5: Industrial Revolution<br>TL6: Mechanized Age<br>TL7: Nuclear Age<br>TL8: Digital Age<br>TL9: Microtech Age<br>TL10: Robotic Age<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>"));
        add(wrapper);

        createDivider();

        wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
        createLabelAndField(wrapper, sheet, Profile.ID_HAIR, I18n.Text("Hair:"), I18n.Text("The character's hair style and color"));
        createLabelAndField(wrapper, sheet, Profile.ID_EYE_COLOR, I18n.Text("Eyes:"), I18n.Text("The character's eye color"));
        createLabelAndField(wrapper, sheet, Profile.ID_SKIN_COLOR, I18n.Text("Skin:"), I18n.Text("The character's skin color"));
        createLabelAndField(wrapper, sheet, Profile.ID_HANDEDNESS, I18n.Text("Hand:"), I18n.Text("The character's preferred hand"));
        add(wrapper);
    }

    private void createDivider() {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        addVerticalBackground(panel, ThemeColor.ON_PAGE);
    }
}
