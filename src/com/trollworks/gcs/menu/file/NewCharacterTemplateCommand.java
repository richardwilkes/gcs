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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;

/** Provides the "New Character Template" command. */
public class NewCharacterTemplateCommand extends Command {
	@Localize("New Character Template")
	@Localize(locale = "de", value = "Neue Charaktervorlage")
	@Localize(locale = "ru", value = "Новый шаблон персонажа")
	private static String							NEW_CHARACTER_TEMPLATE;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String						CMD_NEW_CHARACTER_TEMPLATE	= "NewCharacterTemplate";				//$NON-NLS-1$

	/** The singletone {@link NewCharacterTemplateCommand}. */
	public static final NewCharacterTemplateCommand	INSTANCE					= new NewCharacterTemplateCommand();

	private NewCharacterTemplateCommand() {
		super(NEW_CHARACTER_TEMPLATE, CMD_NEW_CHARACTER_TEMPLATE);
	}

	@Override
	public void adjust() {
		// Do nothing. We're always enabled.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		newTemplate();
	}

	/** @return The newly created a new {@link TemplateDockable}. */
	public static TemplateDockable newTemplate() {
		LibraryExplorerDockable library = LibraryExplorerDockable.get();
		if (library != null) {
			TemplateDockable template = new TemplateDockable(new Template());
			library.dockTemplate(template);
			return template;
		}
		return null;
	}
}
