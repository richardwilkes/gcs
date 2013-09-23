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

import static com.trollworks.gcs.menu.item.NewSkillCommand_LS.*;

import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.menu.Command;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

@Localized({
				@LS(key = "SKILL", msg = "New Skill"),
				@LS(key = "SKILL_CONTAINER", msg = "New Skill Container"),
				@LS(key = "TECHNIQUE", msg = "New Technique"),
})
/** Provides the "New Skill" command. */
public class NewSkillCommand extends Command {
	/** The action command this command will issue. */
	public static final String			CMD_NEW_SKILL			= "NewSkill";																											//$NON-NLS-1$
	/** The action command this command will issue. */
	public static final String			CMD_NEW_SKILL_CONTAINER	= "NewSkillContainer";																									//$NON-NLS-1$
	/** The action command this command will issue. */
	public static final String			CMD_NEW_TECHNIQUE		= "NewTechnique";																										//$NON-NLS-1$

	/** The "New Skill" command. */
	public static final NewSkillCommand	INSTANCE				= new NewSkillCommand(false, false, SKILL, CMD_NEW_SKILL, KeyEvent.VK_K, COMMAND_MODIFIER);
	/** The "New Skill Container" command. */
	public static final NewSkillCommand	CONTAINER_INSTANCE		= new NewSkillCommand(true, false, SKILL_CONTAINER, CMD_NEW_SKILL_CONTAINER, KeyEvent.VK_K, SHIFTED_COMMAND_MODIFIER);
	/** The "New Technique" command. */
	public static final NewSkillCommand	TECHNIQUE_INSTANCE		= new NewSkillCommand(false, true, TECHNIQUE, CMD_NEW_TECHNIQUE, KeyEvent.VK_T, COMMAND_MODIFIER);
	private boolean						mContainer;
	private boolean						mTechnique;

	private NewSkillCommand(boolean container, boolean isTechnique, String title, String cmd, int keyCode, int modifiers) {
		super(title, cmd, keyCode, modifiers);
		mContainer = container;
		mTechnique = isTechnique;
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			setEnabled(!((LibraryWindow) window).getOutline().getModel().isLocked());
		} else {
			setEnabled(window instanceof SheetWindow || window instanceof TemplateWindow);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		ListOutline outline;
		DataFile dataFile;

		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			LibraryWindow libraryWindow = (LibraryWindow) window;
			libraryWindow.switchToSkills();
			dataFile = libraryWindow.getLibraryFile();
			outline = libraryWindow.getOutline();
		} else if (window instanceof SheetWindow) {
			SheetWindow sheetWindow = (SheetWindow) window;
			outline = sheetWindow.getSheet().getSkillOutline();
			dataFile = sheetWindow.getCharacter();
		} else if (window instanceof TemplateWindow) {
			TemplateWindow templateWindow = (TemplateWindow) window;
			outline = templateWindow.getSheet().getSkillOutline();
			dataFile = templateWindow.getTemplate();
		} else {
			return;
		}

		Skill skill = mTechnique ? new Technique(dataFile) : new Skill(dataFile, mContainer);
		outline.addRow(skill, getTitle(), false);
		outline.getModel().select(skill, false);
		outline.scrollSelectionIntoView();
		outline.openDetailEditor(true);
	}
}
