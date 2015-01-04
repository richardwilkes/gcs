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

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantagesDockable;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Advantage" command. */
public class NewAdvantageCommand extends Command {
	@Localize("New Advantage")
	@Localize(locale = "de", value = "Neuer Vorteil")
	@Localize(locale = "ru", value = "Новое преимущество")
	private static String					ADVANTAGE;
	@Localize("New Advantage Container")
	@Localize(locale = "de", value = "Neuer Vorteils-Container")
	@Localize(locale = "ru", value = "Новый контейнер преимущества")
	private static String					ADVANTAGE_CONTAINER;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String				CMD_NEW_ADVANTAGE			= "NewAdvantage";																											//$NON-NLS-1$
	/** The action command this command will issue. */
	public static final String				CMD_NEW_ADVANTAGE_CONTAINER	= "NewAdvantageContainer";																									//$NON-NLS-1$
	/** The "New Advantage" command. */
	public static final NewAdvantageCommand	INSTANCE					= new NewAdvantageCommand(false, ADVANTAGE, CMD_NEW_ADVANTAGE, KeyEvent.VK_D, COMMAND_MODIFIER);
	/** The "New Advantage Container" command. */
	public static final NewAdvantageCommand	CONTAINER_INSTANCE			= new NewAdvantageCommand(true, ADVANTAGE_CONTAINER, CMD_NEW_ADVANTAGE_CONTAINER, KeyEvent.VK_D, SHIFTED_COMMAND_MODIFIER);
	private boolean							mContainer;

	private NewAdvantageCommand(boolean container, String title, String cmd, int keyCode, int modifiers) {
		super(title, cmd, keyCode, modifiers);
		mContainer = container;
	}

	@Override
	public void adjust() {
		AdvantagesDockable adq = getTarget(AdvantagesDockable.class);
		if (adq != null) {
			setEnabled(!adq.getOutline().getModel().isLocked());
		} else {
			SheetDockable sheet = getTarget(SheetDockable.class);
			if (sheet != null) {
				setEnabled(true);
			} else {
				setEnabled(getTarget(TemplateDockable.class) != null);
			}
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		ListOutline outline;
		DataFile dataFile;
		AdvantagesDockable adq = getTarget(AdvantagesDockable.class);
		if (adq != null) {
			dataFile = adq.getDataFile();
			outline = adq.getOutline();
			if (outline.getModel().isLocked()) {
				return;
			}
		} else {
			SheetDockable sheet = getTarget(SheetDockable.class);
			if (sheet != null) {
				dataFile = sheet.getDataFile();
				outline = sheet.getSheet().getAdvantageOutline();
			} else {
				TemplateDockable template = getTarget(TemplateDockable.class);
				if (template != null) {
					dataFile = template.getDataFile();
					outline = template.getTemplate().getAdvantageOutline();
				} else {
					return;
				}
			}
		}
		Advantage advantage = new Advantage(dataFile, mContainer);
		outline.addRow(advantage, getTitle(), false);
		outline.getModel().select(advantage, false);
		outline.scrollSelectionIntoView();
		outline.openDetailEditor(true);
	}
}
