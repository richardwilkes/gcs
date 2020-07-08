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
import com.trollworks.gcs.page.PageHeader;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;
import javax.swing.SwingConstants;

/** The character hit location panel. */
public class HitLocationPanel extends DropPanel {

    /**
     * Creates a new hit location panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public HitLocationPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(7).setSpacing(2, 0).setMargins(0), I18n.Text("Hit Locations"));

        addHorizontalBackground(createHeader(I18n.Text("Roll"), null), Color.black);
        addVerticalBackground(createDivider(), Color.black);
        createHeader(I18n.Text("Where"), null);
        addVerticalBackground(createDivider(), Color.black);
        createHeader(I18n.Text("Penalty"), I18n.Text("The hit penalty for targeting a specific hit location"));
        addVerticalBackground(createDivider(), Color.black);
        createHeader(I18n.Text("DR"), null);

        GURPSCharacter   character = sheet.getCharacter();
        HitLocationTable table     = character.getProfile().getHitLocationTable();
        for (HitLocationTableEntry entry : table.getEntries()) {
            createLabel(entry.getRoll(), MessageFormat.format(I18n.Text("<html><body>The random roll needed to hit the <b>{0}</b> hit location</body></html>"), entry.getName()), true);
            createDivider();
            createLabel(entry.getName(), Text.wrapPlainTextForToolTip(entry.getLocation().getDescription()), true);
            createDivider();
            createLabel(Integer.toString(entry.getHitPenalty()), MessageFormat.format(I18n.Text("<html><body>The hit penalty for targeting the <b>{0}</b> hit location</body></html>"), entry.getName()), false);
            createDivider();
            createDRField(sheet, entry.getKey(), MessageFormat.format(I18n.Text("<html><body>The total DR protecting the <b>{0}</b> hit location</body></html>"), entry.getName()));
        }
    }

    @Override
    public Dimension getMaximumSize() {
        Dimension size = super.getMaximumSize();
        size.width = getPreferredSize().width;
        return size;
    }

    private PageHeader createHeader(String title, String tooltip) {
        PageHeader header = new PageHeader(title, tooltip);
        add(header, new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.MIDDLE));
        return header;
    }

    private Wrapper createDivider() {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        return panel;
    }

    private void createLabel(String title, String tooltip, boolean center) {
        PageLabel label = new PageLabel(title, null);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        label.setHorizontalAlignment(center ? SwingConstants.CENTER : SwingConstants.RIGHT);
        add(label, new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
    }

    private void createDRField(CharacterSheet sheet, String key, String tooltip) {
        PageField field = new PageField(sheet, key, SwingConstants.RIGHT, false, tooltip);
        add(field, new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
    }
}
