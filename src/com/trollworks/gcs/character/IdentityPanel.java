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

package com.trollworks.gcs.character;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.utility.Localization;

import javax.swing.SwingConstants;

/** The character identity panel. */
public class IdentityPanel extends DropPanel {
	@Localize("Identity")
	@Localize(locale = "de", value = "Identität")
	@Localize(locale = "ru", value = "Личность")
	private static String	IDENTITY;
	@Localize("Name:")
	@Localize(locale = "de", value = "Name:")
	@Localize(locale = "ru", value = "Имя:")
	private static String	NAME;
	@Localize("Title:")
	@Localize(locale = "de", value = "Titel:")
	@Localize(locale = "ru", value = "Статус:")
	private static String	TITLE;
	@Localize("Religion:")
	@Localize(locale = "de", value = "Religion:")
	@Localize(locale = "ru", value = "Религия:")
	private static String	RELIGION;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new identity panel.
	 *
	 * @param character The character to display the data for.
	 */
	public IdentityPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0), IDENTITY);
		createLabelAndField(this, character, Profile.ID_NAME, NAME, null, SwingConstants.LEFT);
		createLabelAndField(this, character, Profile.ID_TITLE, TITLE, null, SwingConstants.LEFT);
		createLabelAndField(this, character, Profile.ID_RELIGION, RELIGION, null, SwingConstants.LEFT);
	}
}
