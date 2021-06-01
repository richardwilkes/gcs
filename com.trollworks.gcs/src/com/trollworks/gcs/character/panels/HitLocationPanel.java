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

import com.trollworks.gcs.body.HitLocation;
import com.trollworks.gcs.body.HitLocationTable;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageHeader;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.ThemeColor;
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
    private static float colorAdjustment = -0.1f;

    /**
     * Creates a new hit location panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public HitLocationPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(7).setSpacing(2, 0).setMargins(0), sheet.getCharacter().getSettings().getHitLocations().getName());

        addHorizontalBackground(createHeader(I18n.Text("Roll"), null), ThemeColor.HEADER);
        addVerticalBackground(createDivider(), ThemeColor.DIVIDER);
        createHeader(I18n.Text("Where"), null);
        addVerticalBackground(createDivider(), ThemeColor.DIVIDER);
        createHeader(I18n.Text("Penalty"), I18n.Text("The hit penalty for targeting a specific hit location"));
        addVerticalBackground(createDivider(), ThemeColor.DIVIDER);
        createHeader(I18n.Text("DR"), null);

        addTable(sheet, sheet.getCharacter().getSettings().getHitLocations(), 0, ThemeColor.BANDING, ThemeColor.CONTENT, false);
    }

    private boolean addTable(CharacterSheet sheet, HitLocationTable table, int depth, Color band1Color, Color band2Color, boolean band) {
        for (HitLocation location : table.getLocations()) {
            String    name   = location.getTableName();
            String    prefix = Text.makeFiller(depth * 3, ' ');
            PageLabel first  = createLabel(prefix + location.getRollRange(), MessageFormat.format(I18n.Text("<html><body>The random roll needed to hit the <b>{0}</b> hit location</body></html>"), name), SwingConstants.LEFT);
            addHorizontalBackground(first, band ? band1Color : band2Color);
            band = !band;
            createDivider();
            createLabel(prefix + name, Text.wrapPlainTextForToolTip(location.getDescription()), SwingConstants.LEFT);
            createDivider();
            createLabel(Integer.toString(location.getHitPenalty()), MessageFormat.format(I18n.Text("<html><body>The hit penalty for targeting the <b>{0}</b> hit location</body></html>"), name), SwingConstants.RIGHT);
            createDivider();
            StringBuilder tooltip = new StringBuilder();
            int           dr      = location.getDR(sheet.getCharacter(), tooltip);
            //noinspection DynamicRegexReplaceableByCompiledPattern
            createDRField(sheet, Integer.valueOf(dr), String.format(I18n.Text("<html><body>The DR covering the <b>%s</b> hit location%s</body></html>"), name, tooltip.toString().replaceAll("\n", "<br>")));
            if (location.getSubTable() != null) {
                band = addTable(sheet, location.getSubTable(), depth + 1, band1Color, band2Color, band);
            }
        }
        return band;
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

    private PageLabel createLabel(String title, String tooltip, int alignment) {
        PageLabel label = new PageLabel(title, null);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        label.setHorizontalAlignment(alignment);
        add(label, new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        return label;
    }

    private void createDRField(CharacterSheet sheet, Object value, String tooltip) {
        PageField field = new PageField(FieldFactory.POSINT5, value, sheet, SwingConstants.RIGHT, tooltip, ThemeColor.ON_CONTENT);
        add(field, new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
    }
}
