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
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;

/** Provides the "Add Natural Kick w/Boots" command. */
public class AddNaturalKickWithBootsCommand extends Command {
	@Localize("Include Kick w/Boots In Weapons")
	@Localize(locale = "de", value = "Führe Tritt mit Schuh als Waffe auf")
	@Localize(locale = "ru", value = "Отображать пинок (в ботинке) в оружии")
	private static String								ADD_NATURAL_KICK_WITH_BOOTS;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String							CMD_ADD_NATURAL_KICK_WITH_BOOTS	= "AddNaturalKickWithBoots";			//$NON-NLS-1$

	/** The singleton {@link AddNaturalKickWithBootsCommand}. */
	public static final AddNaturalKickWithBootsCommand	INSTANCE						= new AddNaturalKickWithBootsCommand();

	private AddNaturalKickWithBootsCommand() {
		super(ADD_NATURAL_KICK_WITH_BOOTS, CMD_ADD_NATURAL_KICK_WITH_BOOTS);
	}

	@Override
	public void adjust() {
		CharacterSheet sheet = getTarget(CharacterSheet.class);
		if (sheet != null) {
			setEnabled(true);
			setMarked(sheet.getCharacter().includeKickBoots());
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
			character.setIncludeKickBoots(!character.includeKickBoots());
		}
	}
}
