/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.DescriptionRandomizer;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;

import javax.swing.JOptionPane;

/** Provides the "Randomize Description" command. */
public class RandomizeDescriptionCommand extends Command {
	@Localize("Randomize Description\u2026")
	private static String							RANDOMIZE_DESCRIPTION;
	@Localize("Description Randomizer")
	private static String							RANDOMIZER;
	@Localize("Apply")
	private static String							APPLY;
	@Localize("Cancel")
	private static String							CANCEL;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String						CMD_RANDOMIZE_DESCRIPTION	= "RandomizeDescription";				//$NON-NLS-1$

	/** The singleton {@link RandomizeDescriptionCommand}. */
	public static final RandomizeDescriptionCommand	INSTANCE					= new RandomizeDescriptionCommand();

	private RandomizeDescriptionCommand() {
		super(RANDOMIZE_DESCRIPTION, CMD_RANDOMIZE_DESCRIPTION);
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
			if (WindowUtils.showOptionDialog(null, panel, RANDOMIZER, true, JOptionPane.OK_CANCEL_OPTION, JOptionPane.INFORMATION_MESSAGE, GCSImages.getCharacterSheetDocumentIcons().getIcon(32), new String[] { APPLY, CANCEL }, APPLY) == JOptionPane.OK_OPTION) {
				panel.applyChanges();
			}
		}
	}
}
