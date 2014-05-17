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
import com.trollworks.toolkit.utility.undo.MultipleUndo;
import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.outline.Row;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;

/** Provides the "Apply Template To Sheet" command. */
public class ApplyTemplateCommand extends Command {
	@Localize("Apply Template To Character Sheet")
	private static String APPLY_TEMPLATE_TO_SHEET;
	@Localize("Apply Template")
	private static String UNDO;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String					CMD_APPLY_TEMPLATE	= "ApplyTemplate";				//$NON-NLS-1$
	/** The singleton {@link ApplyTemplateCommand}. */
	public static final ApplyTemplateCommand	INSTANCE			= new ApplyTemplateCommand();

	private ApplyTemplateCommand() {
		super(APPLY_TEMPLATE_TO_SHEET, CMD_APPLY_TEMPLATE, KeyEvent.VK_A, SHIFTED_COMMAND_MODIFIER);
	}

	@Override
	public void adjust() {
		Window window = getActiveWindow();
		if (window instanceof TemplateWindow) {
			setEnabled(SheetWindow.getTopSheet() != null);
		} else {
			setEnabled(false);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Window window = getActiveWindow();
		if (window instanceof TemplateWindow) {
			Template template = ((TemplateWindow) window).getTemplate();
			MultipleUndo edit = new MultipleUndo(UNDO);
			ArrayList<Row> rows = new ArrayList<>();
			String notes = template.getNotes().trim();
			SheetWindow sheet = SheetWindow.getTopSheet();
			template.addEdit(edit);
			rows.addAll(template.getAdvantagesModel().getTopLevelRows());
			rows.addAll(template.getSkillsModel().getTopLevelRows());
			rows.addAll(template.getSpellsModel().getTopLevelRows());
			rows.addAll(template.getEquipmentModel().getTopLevelRows());
			sheet.addRows(rows);
			if (notes.length() > 0) {
				Profile description = sheet.getCharacter().getDescription();
				String prevNotes = description.getNotes().trim();
				if (prevNotes.length() > 0) {
					notes = prevNotes + "\n\n" + notes; //$NON-NLS-1$
				}
				description.setNotes(notes);
			}
			edit.end();
		}
	}
}
