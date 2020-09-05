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
import com.trollworks.gcs.character.names.USCensusNames;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.text.MessageFormat;

/** Provides the "Generate Random Name" command. */
public class RandomizeNameCommand extends Command {
    /** The action command this command will issue. */
    public static final String               CMD_GENERATE_RANDOM_MALE_NAME   = "GenerateRandomMaleName";
    /** The action command this command will issue. */
    public static final String               CMD_GENERATE_RANDOM_FEMALE_NAME = "GenerateRandomFemaleName";
    /** The male {@link RandomizeNameCommand}. */
    public static final RandomizeNameCommand MALE_INSTANCE                   = new RandomizeNameCommand(I18n.Text("Male"), CMD_GENERATE_RANDOM_MALE_NAME, KeyEvent.VK_V);
    /** The female {@link RandomizeNameCommand}. */
    public static final RandomizeNameCommand FEMALE_INSTANCE                 = new RandomizeNameCommand(I18n.Text("Female"), CMD_GENERATE_RANDOM_FEMALE_NAME, KeyEvent.VK_I);

    private RandomizeNameCommand(String type, String cmd, int keyCode) {
        super(MessageFormat.format(I18n.Text("Generate Random {0} Name"), type), cmd, keyCode, Command.SHIFTED_COMMAND_MODIFIER);
    }

    @Override
    public void adjust() {
        setEnabled(getTarget(CharacterSheet.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        CharacterSheet target = getTarget(CharacterSheet.class);
        if (target != null) {
            target.getCharacter().getProfile().setName(USCensusNames.INSTANCE.getFullName(this == MALE_INSTANCE));
        }
    }
}
