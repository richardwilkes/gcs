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

package com.trollworks.gcs.menu.item;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;


import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "Copy To Template" command. */
public class CopyToTemplateCommand extends Command {
	@Localize("Copy To Template")
	private static String COPY_TO_TEMPLATE;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String					CMD_COPY_TO_TEMPLATE	= "CopyToTemplate";			//$NON-NLS-1$
	/** The singleton {@link CopyToTemplateCommand}. */
	public static final CopyToTemplateCommand	INSTANCE				= new CopyToTemplateCommand();

	private CopyToTemplateCommand() {
		super(COPY_TO_TEMPLATE, CMD_COPY_TO_TEMPLATE, KeyEvent.VK_T, SHIFTED_COMMAND_MODIFIER);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			setEnabled(((LibraryWindow) window).getOutline().getModel().hasSelection() && TemplateWindow.getTopTemplate() != null);
		} else {
			setEnabled(false);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			OutlineModel outlineModel = ((LibraryWindow) window).getOutline().getModel();
			if (outlineModel.hasSelection()) {
				TemplateWindow template = TemplateWindow.getTopTemplate();
				if (template != null) {
					template.addRows(outlineModel.getSelectionAsList(true));
				}
			}
		}
	}
}
