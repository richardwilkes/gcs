/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.item;

import static com.trollworks.gcs.menu.item.ApplyTemplateCommand_LS.*;

import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.undo.MultipleUndo;
import com.trollworks.ttk.widgets.outline.Row;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;

import javax.swing.JMenuItem;

@Localized({
				@LS(key = "APPLY_TEMPLATE_TO_SHEET", msg = "Apply Template To Character Sheet"),
				@LS(key = "UNDO", msg = "Apply Template"),
})
/** Provides the "Apply Template To Sheet" command. */
public class ApplyTemplateCommand extends Command {
	/** The action command this command will issue. */
	public static final String					CMD_APPLY_TEMPLATE	= "ApplyTemplate";				//$NON-NLS-1$
	/** The singleton {@link ApplyTemplateCommand}. */
	public static final ApplyTemplateCommand	INSTANCE			= new ApplyTemplateCommand();

	private ApplyTemplateCommand() {
		super(APPLY_TEMPLATE_TO_SHEET, CMD_APPLY_TEMPLATE, KeyEvent.VK_A, SHIFTED_COMMAND_MODIFIER);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
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
