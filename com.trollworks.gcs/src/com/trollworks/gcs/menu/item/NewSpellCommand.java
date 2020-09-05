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

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.spell.RitualMagicSpell;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellsDockable;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Spell" command. */
public class NewSpellCommand extends Command {
    /** The action command this command will issue. */
    public static final String          CMD_SPELL              = "NewSpell";
    /** The action command this command will issue. */
    public static final String          CMD_SPELL_CONTAINER    = "NewSpellContainer";
    /** The action command this command will issue. */
    public static final String          CMD_RITUAL_MAGIC_SPELL = "NewRitualMagicSpell";
    /** The "New Spell" command. */
    public static final NewSpellCommand INSTANCE               = new NewSpellCommand(false, I18n.Text("New Spell"), CMD_SPELL, COMMAND_MODIFIER);
    /** The "New Spell Container" command. */
    public static final NewSpellCommand CONTAINER_INSTANCE     = new NewSpellCommand(true, I18n.Text("New Spell Container"), CMD_SPELL_CONTAINER, SHIFTED_COMMAND_MODIFIER);
    /** The "New Technique" command. */
    public static final NewSpellCommand RITUAL_MAGIC_INSTANCE  = new NewSpellCommand();
    private             boolean         mContainer;
    private             boolean         mRitualMagic;

    private NewSpellCommand() {
        super(I18n.Text("New Ritual Magic Spell"), CMD_RITUAL_MAGIC_SPELL);
        mRitualMagic = true;
    }

    private NewSpellCommand(boolean container, String title, String cmd, int modifiers) {
        super(title, cmd, KeyEvent.VK_B, modifiers);
        mContainer = container;
    }

    @Override
    public void adjust() {
        SpellsDockable spells = getTarget(SpellsDockable.class);
        if (spells != null) {
            setEnabled(!spells.getOutline().getModel().isLocked());
        } else {
            if (getTarget(SheetDockable.class) != null) {
                setEnabled(true);
            } else {
                setEnabled(getTarget(TemplateDockable.class) != null);
            }
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ListOutline    outline;
        DataFile       dataFile;
        SpellsDockable spells = getTarget(SpellsDockable.class);
        if (spells != null) {
            dataFile = spells.getDataFile();
            outline = spells.getOutline();
            if (outline.getModel().isLocked()) {
                return;
            }
        } else {
            SheetDockable sheet = getTarget(SheetDockable.class);
            if (sheet != null) {
                dataFile = sheet.getDataFile();
                outline = sheet.getSheet().getSpellOutline();
            } else {
                TemplateDockable template = getTarget(TemplateDockable.class);
                if (template != null) {
                    dataFile = template.getDataFile();
                    outline = template.getTemplate().getSpellOutline();
                } else {
                    return;
                }
            }
        }
        Spell spell = mRitualMagic ? new RitualMagicSpell(dataFile) : new Spell(dataFile, mContainer);
        outline.addRow(spell, getTitle(), false);
        outline.getModel().select(spell, false);
        outline.scrollSelectionIntoView();
        outline.openDetailEditor(true);
    }
}
