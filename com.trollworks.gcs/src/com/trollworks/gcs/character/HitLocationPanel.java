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
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Container;
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
        super(new ColumnLayout(7, 2, 0), I18n.Text("Hit Location"));

        GURPSCharacter   character = sheet.getCharacter();
        HitLocationTable table     = character.getProfile().getHitLocationTable();

        Wrapper wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
        addHorizontalBackground(createHeader(wrapper, I18n.Text("Roll"), null), Color.black);
        for (HitLocationTableEntry entry : table.getEntries()) {
            createLabel(wrapper, entry.getRoll(), MessageFormat.format(I18n.Text("<html><body>The random roll needed to hit the <b>{0}</b> hit location</body></html>"), entry.getName()), SwingConstants.CENTER);
        }
        wrapper.setAlignmentY(TOP_ALIGNMENT);
        add(wrapper);

        createDivider();

        wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
        createHeader(wrapper, I18n.Text("Where"), null);
        for (HitLocationTableEntry entry : table.getEntries()) {
            createLabel(wrapper, entry.getName(), Text.wrapPlainTextForToolTip(entry.getLocation().getDescription()), SwingConstants.CENTER);
        }
        wrapper.setAlignmentY(TOP_ALIGNMENT);
        add(wrapper);

        createDivider();

        wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
        createHeader(wrapper, "-", I18n.Text("The hit penalty for targeting a specific hit location"));
        for (HitLocationTableEntry entry : table.getEntries()) {
            createLabel(wrapper, Integer.toString(entry.getHitPenalty()), MessageFormat.format(I18n.Text("<html><body>The hit penalty for targeting the <b>{0}</b> hit location</body></html>"), entry.getName()), SwingConstants.RIGHT);
        }
        wrapper.setAlignmentY(TOP_ALIGNMENT);
        add(wrapper);

        createDivider();

        wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
        createHeader(wrapper, I18n.Text("DR"), null);
        for (HitLocationTableEntry entry : table.getEntries()) {
            createDisabledField(wrapper, sheet, entry.getKey(), MessageFormat.format(I18n.Text("<html><body>The total DR protecting the <b>{0}</b> hit location</body></html>"), entry.getName()), SwingConstants.RIGHT);
        }
        wrapper.setAlignmentY(TOP_ALIGNMENT);
        add(wrapper);
    }

    @Override
    public Dimension getMaximumSize() {
        Dimension size = super.getMaximumSize();
        size.width = getPreferredSize().width;
        return size;
    }

    private void createDivider() {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        addVerticalBackground(panel, Color.black);
    }

    @SuppressWarnings("static-method")
    private void createLabel(Container panel, String title, String tooltip, int alignment) {
        PageLabel label = new PageLabel(title, null);
        label.setHorizontalAlignment(alignment);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        panel.add(label);
    }
}
