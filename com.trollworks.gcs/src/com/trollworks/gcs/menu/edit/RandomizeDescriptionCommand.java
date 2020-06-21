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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.DescriptionRandomizer;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import javax.swing.JOptionPane;

/** Provides the "Randomize Description" command. */
public class RandomizeDescriptionCommand extends Command {
    /** The action command this command will issue. */
    public static final String                      CMD_RANDOMIZE_DESCRIPTION = "RandomizeDescription";
    /** The singleton {@link RandomizeDescriptionCommand}. */
    public static final RandomizeDescriptionCommand INSTANCE                  = new RandomizeDescriptionCommand();

    private RandomizeDescriptionCommand() {
        super(I18n.Text("Randomize Description…"), CMD_RANDOMIZE_DESCRIPTION);
    }

    @Override
    public void adjust() {
        setEnabled(getTarget(CharacterSheet.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        CharacterSheet target = getTarget(CharacterSheet.class);
        if (target != null) {
            DescriptionRandomizer panel = new DescriptionRandomizer(target.getCharacter());
            if (WindowUtils.showOptionDialog(null, panel, I18n.Text("Description Randomizer"), true, JOptionPane.OK_CANCEL_OPTION, JOptionPane.INFORMATION_MESSAGE, Images.GCS_MARKER, new String[]{I18n.Text("Apply"), I18n.Text("Cancel")}, I18n.Text("Apply")) == JOptionPane.OK_OPTION) {
                panel.applyChanges();
            }
        }
    }
}
