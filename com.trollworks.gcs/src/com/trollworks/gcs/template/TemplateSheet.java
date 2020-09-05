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

package com.trollworks.gcs.template;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.CollectedOutlines;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineSyncer;
import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.dnd.DropTarget;
import java.awt.event.ActionEvent;

/** The template sheet. */
public class TemplateSheet extends CollectedOutlines {
    private static final EmptyBorder NORMAL_BORDER = new EmptyBorder(5);
    /** Used to determine whether a resize action is pending. */
    protected            boolean     mSizePending;

    /**
     * Creates a new {@link TemplateSheet}.
     *
     * @param template The template to display the data for.
     */
    public TemplateSheet(Template template) {
        setLayout(new ColumnLayout(1, 0, 5));
        setOpaque(true);
        setBackground(Color.WHITE);
        setBorder(NORMAL_BORDER);

        // Make sure our primary outlines exist
        createOutlines(template);
        add(new TemplateOutlinePanel(getAdvantageOutline(), I18n.Text("Advantages, Disadvantages & Quirks")));
        add(new TemplateOutlinePanel(getSkillOutline(), I18n.Text("Skills")));
        add(new TemplateOutlinePanel(getSpellOutline(), I18n.Text("Spells")));
        add(new TemplateOutlinePanel(getEquipmentOutline(), I18n.Text("Equipment")));
        add(new TemplateOutlinePanel(getOtherEquipmentOutline(), I18n.Text("Other Equipment")));
        add(new TemplateOutlinePanel(getNoteOutline(), I18n.Text("Notes")));

        // Ensure everything is laid out and register for notification
        revalidate();
        template.addTarget(this, Template.TEMPLATE_PREFIX, GURPSCharacter.CHARACTER_PREFIX);
        Preferences.getInstance().getNotifier().add(this, Preferences.KEY_SHOW_COLLEGE_IN_SHEET_SPELLS);

        setDropTarget(new DropTarget(this, this));

        runAdjustSize();
    }

    @Override
    protected void scaleChanged() {
        revalidate();
        repaint();
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (Outline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(command)) {
            adjustSize();
        }
    }

    private void adjustSize() {
        if (!mSizePending) {
            mSizePending = true;
            EventQueue.invokeLater(this::runAdjustSize);
        }
    }

    void runAdjustSize() {
        Scale scale = Scale.get(this);
        mSizePending = false;
        updateRowHeights();
        revalidate();
        Dimension size = getLayout().preferredLayoutSize(this);
        size.width = scale.scale(8 * 72);
        UIUtilities.setOnlySize(this, size);
        setSize(size);
    }

    @Override
    public void handleNotification(Object producer, String type, Object data) {
        if (type.startsWith(Advantage.PREFIX)) {
            OutlineSyncer.add(getAdvantageOutline());
        } else if (type.startsWith(Skill.PREFIX)) {
            OutlineSyncer.add(getSkillOutline());
        } else if (type.startsWith(Spell.PREFIX)) {
            OutlineSyncer.add(getSpellOutline());
        } else if (type.startsWith(Equipment.PREFIX)) {
            OutlineSyncer.add(getEquipmentOutline());
            OutlineSyncer.add(getOtherEquipmentOutline());
        } else if (type.startsWith(Note.PREFIX)) {
            OutlineSyncer.add(getNoteOutline());
        } else if (Preferences.KEY_SHOW_COLLEGE_IN_SHEET_SPELLS.equals(type)) {
            SpellOutline spellOutline = getSpellOutline();
            spellOutline.resetColumns();
            OutlineSyncer.add(spellOutline);
        }
        if (!inBatchMode()) {
            validate();
        }
    }
}
