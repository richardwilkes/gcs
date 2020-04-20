/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

/** Provides the "Add Natural Punch" command. */
public class AddNaturalPunchCommand extends Command {
    /** The action command this command will issue. */
    public static final String                 CMD_ADD_NATURAL_PUNCH = "AddNaturalPunch";
    /** The singleton {@link AddNaturalPunchCommand}. */
    public static final AddNaturalPunchCommand INSTANCE              = new AddNaturalPunchCommand();

    private AddNaturalPunchCommand() {
        super(I18n.Text("Include Punch In Weapons"), CMD_ADD_NATURAL_PUNCH);
    }

    @Override
    public void adjust() {
        CharacterSheet sheet = getTarget(CharacterSheet.class);
        if (sheet != null) {
            setEnabled(true);
            setMarked(sheet.getCharacter().includePunch());
        } else {
            setEnabled(false);
            setMarked(false);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        CharacterSheet sheet = getTarget(CharacterSheet.class);
        if (sheet != null) {
            GURPSCharacter character = sheet.getCharacter();
            character.setIncludePunch(!character.includePunch());
        }
    }
}
