/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.names.USCensusNames;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.text.MessageFormat;

/** Provides the "Generate Random Name" command. */
public class RandomizeNameCommand extends Command {
	@Localize("Male")
	@Localize(locale = "de", value = "männlichen")
	@Localize(locale = "ru", value = "муж.")
	private static String						MALE;
	@Localize("Female")
	@Localize(locale = "de", value = "weiblichen")
	@Localize(locale = "ru", value = "жен.")
	private static String						FEMALE;
	@Localize("Generate Random {0} Name")
	@Localize(locale = "de", value = "Erstelle zufälligen {0} Namen")
	@Localize(locale = "ru", value = "Создать случайное {0} имя")
	private static String						TITLE;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String					CMD_GENERATE_RANDOM_MALE_NAME	= "GenerateRandomMaleName";																							//$NON-NLS-1$
	/** The action command this command will issue. */
	public static final String					CMD_GENERATE_RANDOM_FEMALE_NAME	= "GenerateRandomFemaleName";																							//$NON-NLS-1$
	/** The male {@link RandomizeNameCommand}. */
	public static final RandomizeNameCommand	MALE_INSTANCE					= new RandomizeNameCommand(MALE, CMD_GENERATE_RANDOM_MALE_NAME, KeyEvent.VK_V, Command.SHIFTED_COMMAND_MODIFIER);
	/** The female {@link RandomizeNameCommand}. */
	public static final RandomizeNameCommand	FEMALE_INSTANCE					= new RandomizeNameCommand(FEMALE, CMD_GENERATE_RANDOM_FEMALE_NAME, KeyEvent.VK_I, Command.SHIFTED_COMMAND_MODIFIER);

	private RandomizeNameCommand(String type, String cmd, int keyCode, int modifiers) {
		super(MessageFormat.format(TITLE, type), cmd, keyCode, modifiers);
	}

	@Override
	public void adjust() {
		setEnabled(getTarget(CharacterSheet.class) != null);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		CharacterSheet target = getTarget(CharacterSheet.class);
		if (target != null) {
			target.getCharacter().getDescription().setName(USCensusNames.INSTANCE.getFullName(this == MALE_INSTANCE));
		}
	}
}
