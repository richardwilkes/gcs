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

package com.trollworks.gcs.template;

import com.trollworks.gcs.character.CollectedOutlines;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.dnd.DropTarget;
import java.awt.event.ActionEvent;

/** The template sheet. */
public class TemplateSheet extends CollectedOutlines {
    private Template mTemplate;

    /**
     * Creates a new {@link TemplateSheet}.
     *
     * @param template The template to display the data for.
     */
    public TemplateSheet(Template template) {
        mTemplate = template;
        setLayout(new ColumnLayout(1, 0, 5));
        setOpaque(true);
        setBackground(Color.WHITE);
        setBorder(new EmptyBorder(5));

        // Make sure our primary outlines exist
        createOutlines(template);
        add(new TemplateOutlinePanel(getAdvantagesOutline(), I18n.Text("Advantages, Disadvantages & Quirks")));
        add(new TemplateOutlinePanel(getSkillsOutline(), I18n.Text("Skills")));
        add(new TemplateOutlinePanel(getSpellsOutline(), I18n.Text("Spells")));
        add(new TemplateOutlinePanel(getEquipmentOutline(), I18n.Text("Equipment")));
        add(new TemplateOutlinePanel(getOtherEquipmentOutline(), I18n.Text("Other Equipment")));
        add(new TemplateOutlinePanel(getNotesOutline(), I18n.Text("Notes")));

        revalidate();
        mTemplate.addChangeListener(this);
        setDropTarget(new DropTarget(this, this));
        adjustSize();
    }

    @Override
    public void dispose() {
        mTemplate.removeChangeListener(this);
        super.dispose();
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (Outline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(command)) {
            markForRebuild();
        }
    }

    @Override
    public void rebuild() {
        syncOutline(getAdvantagesOutline());
        syncOutline(getSkillsOutline());
        SpellOutline spellOutline = getSpellsOutline();
        spellOutline.resetColumns();
        syncOutline(spellOutline);
        syncOutline(getEquipmentOutline());
        syncOutline(getOtherEquipmentOutline());
        syncOutline(getNotesOutline());
        adjustSize();
    }

    private void syncOutline(Outline outline) {
        if (outline != null) {
            outline.sizeColumnsToFit();
        }
    }

    void adjustSize() {
        Scale scale = Scale.get(this);
        updateRowHeights();
        revalidate();
        Dimension size = getLayout().preferredLayoutSize(this);
        size.width = scale.scale(8 * 72);
        UIUtilities.setOnlySize(this, size);
        setSize(size);
    }
}
