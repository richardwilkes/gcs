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

/** The character player info panel. */
public class PlayerInfoPanel extends DropPanel {
	@Localize("Player Information")
	@Localize(locale = "de", value = "Spielerinformation")
	@Localize(locale = "ru", value = "Информация об игроке")
	private static String	PLAYER_INFO;
	@Localize("Player:")
	@Localize(locale = "de", value = "Spieler:")
	@Localize(locale = "ru", value = "Игрок:")
	private static String	PLAYER_NAME;
	@Localize("Campaign:")
	@Localize(locale = "de", value = "Kampagne:")
	@Localize(locale = "ru", value = "Компания:")
	private static String	CAMPAIGN;
	@Localize("Created On:")
	@Localize(locale = "de", value = "Erstellt am:")
	@Localize(locale = "ru", value = "Создан:")
	private static String	CREATED_ON;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new player info panel.
	 *
	 * @param character The character to display the data for.
	 */
	public PlayerInfoPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0), PLAYER_INFO);
		createLabelAndField(this, character, Profile.ID_PLAYER_NAME, PLAYER_NAME, null, SwingConstants.LEFT);
		createLabelAndField(this, character, Profile.ID_CAMPAIGN, CAMPAIGN, null, SwingConstants.LEFT);
		createLabelAndField(this, character, GURPSCharacter.ID_CREATED_ON, CREATED_ON, null, SwingConstants.LEFT);
	}
}
