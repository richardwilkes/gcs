/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.outline.ColumnUtils;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.utility.I18n;

import java.awt.EventQueue;
import java.awt.Insets;
import java.awt.dnd.DropTarget;
import java.awt.event.ActionEvent;

/** The template sheet. */
public class TemplateSheet extends CollectedOutlines {
    private Template mTemplate;
    private boolean  mIgnoreContentSizeChange;

    /**
     * Creates a new TemplateSheet.
     *
     * @param template The template to display the data for.
     */
    public TemplateSheet(Template template) {
        mTemplate = template;
        setLayout(new PrecisionLayout().setVerticalSpacing(5).setHorizontalSpacing(0).setMargins(5));
        setOpaque(true);
        setBackground(Colors.PAGE);

        // Make sure our primary outlines exist
        createOutlines(template);
        PrecisionLayoutData layoutData = new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true);
        add(new TemplateOutlinePanel(getAdvantagesOutline(), I18n.text("Advantages, Disadvantages & Quirks")), layoutData.clone());
        add(new TemplateOutlinePanel(getSkillsOutline(), I18n.text("Skills")), layoutData.clone());
        add(new TemplateOutlinePanel(getSpellsOutline(), I18n.text("Spells")), layoutData.clone());
        add(new TemplateOutlinePanel(getEquipmentOutline(), I18n.text("Equipment")), layoutData.clone());
        add(new TemplateOutlinePanel(getOtherEquipmentOutline(), I18n.text("Other Equipment")), layoutData.clone());
        add(new TemplateOutlinePanel(getNotesOutline(), I18n.text("Notes")), layoutData.clone());

        mTemplate.addChangeListener(this);
        setDropTarget(new DropTarget(this, this));
        rebuild();
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
            if (!mIgnoreContentSizeChange) {
                markForRebuild();
            }
        }
    }

    @Override
    public void rebuild() {
        mIgnoreContentSizeChange = true;
        mTemplate.recalculate();
        Insets insets = getInsets();
        int    width  = Scale.get(this).scale(8 * 72) - (insets.left + insets.right);
        ColumnUtils.pack(getAdvantagesOutline(), width);
        SkillOutline skillsOutline = getSkillsOutline();
        skillsOutline.resetColumns();
        ColumnUtils.pack(skillsOutline, width);
        SpellOutline spellOutline = getSpellsOutline();
        spellOutline.resetColumns();
        ColumnUtils.pack(spellOutline, width);
        ColumnUtils.pack(getEquipmentOutline(), width);
        ColumnUtils.pack(getOtherEquipmentOutline(), width);
        ColumnUtils.pack(getNotesOutline(), width);
        setSize(getPreferredSize());
        repaint();
        // This next line must be put into the event queue to prevent an endless rebuild cycle due
        // to deferred processing.
        EventQueue.invokeLater(() -> mIgnoreContentSizeChange = false);
    }
}
