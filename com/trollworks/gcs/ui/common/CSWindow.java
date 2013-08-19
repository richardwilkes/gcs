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
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.common;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.ui.preferences.CSPreferencesWindow;
import com.trollworks.gcs.ui.sheet.CSSheetWindow;
import com.trollworks.gcs.ui.template.CSTemplateWindow;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.io.TKUpdateChecker;
import com.trollworks.toolkit.utility.TKBrowser;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKToolBar;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuBar;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKOutlineHeaderCM;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.search.TKSearch;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKFileDialog;
import com.trollworks.toolkit.window.TKOpenManager;
import com.trollworks.toolkit.window.TKOptionDialog;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Component;
import java.awt.KeyboardFocusManager;
import java.awt.image.BufferedImage;
import java.io.File;
import java.text.MessageFormat;
import java.util.ArrayList;

/** The common window superclass, which handles most menu commands. */
public class CSWindow extends TKWindow {
	/** The command for whether the "natural punch" weapon should be present. */
	public static final String	CMD_ADD_NATURAL_PUNCH				= "AddNaturalPunch";				//$NON-NLS-1$
	/** The command for whether the "natural kick" weapon should be present. */
	public static final String	CMD_ADD_NATURAL_KICK				= "AddNaturalKick";				//$NON-NLS-1$
	/** The command for whether the "natural kick w/boots" weapon should be present. */
	public static final String	CMD_ADD_NATURAL_KICK_WITH_BOOTS		= "AddNaturalKickBoots";			//$NON-NLS-1$
	/** The command for randomizing the character description fields. */
	public static final String	CMD_RANDOMIZE_DESCRIPTION			= "RandomizeDescription";			//$NON-NLS-1$
	/** The command for generating a random female name. */
	public static final String	CMD_RANDOMIZE_FEMALE_NAME			= "FemaleRandomName";				//$NON-NLS-1$
	/** The command for generating a random male name. */
	public static final String	CMD_RANDOM_MALE_NAME				= "MaleRandomName";				//$NON-NLS-1$
	/** The command for copying list items to the character sheet. */
	public static final String	CMD_COPY_TO_SHEET					= "CopyToSheet";					//$NON-NLS-1$
	/** The command for copying list items to the template sheet. */
	public static final String	CMD_COPY_TO_TEMPLATE				= "CopyToTemplate";				//$NON-NLS-1$
	/** The command for applying a template to the character sheet. */
	public static final String	CMD_APPLY_TEMPLATE_TO_SHEET			= "ApplyTemplateToSheet";			//$NON-NLS-1$
	/** The command for creating a new character sheet. */
	public static final String	CMD_NEW_SHEET						= "NewCharacterSheet";				//$NON-NLS-1$
	/** The command for creating a new template. */
	public static final String	CMD_NEW_TEMPLATE					= "NewTemplate";					//$NON-NLS-1$
	/** The command for creating a new list. */
	public static final String	CMD_NEW_LIST						= "NewList";						//$NON-NLS-1$
	/** The command for opening a list. */
	public static final String	CMD_OPEN_LIST						= "OpenList";						//$NON-NLS-1$
	/** The command for opening an editor. */
	public static final String	CMD_OPEN_EDITOR						= "OpenEditor";					//$NON-NLS-1$
	/** The command for creating a new (dis)advantage. */
	public static final String	CMD_NEW_ADVANTAGE					= "NewAdvantage";					//$NON-NLS-1$
	/** The command for creating a new (dis)advantage container. */
	public static final String	CMD_NEW_ADVANTAGE_CONTAINER			= "NewAdvantageContainer";			//$NON-NLS-1$
	/** The command for creating a new skill. */
	public static final String	CMD_NEW_SKILL						= "NewSkill";						//$NON-NLS-1$
	/** The command for creating a new skill container. */
	public static final String	CMD_NEW_SKILL_CONTAINER				= "NewSkillContainer";				//$NON-NLS-1$
	/** The command for creating a new technique. */
	public static final String	CMD_NEW_TECHNIQUE					= "NewTechnique";					//$NON-NLS-1$
	/** The command for creating a new spell. */
	public static final String	CMD_NEW_SPELL						= "NewSpell";						//$NON-NLS-1$
	/** The command for creating a new spell container. */
	public static final String	CMD_NEW_SPELL_CONTAINER				= "NewSpellContainer";				//$NON-NLS-1$
	/** The command for creating a new carried equipment. */
	public static final String	CMD_NEW_CARRIED_EQUIPMENT			= "NewCarriedEquipment";			//$NON-NLS-1$
	/** The command for creating a new carried equipment container. */
	public static final String	CMD_NEW_CARRIED_EQUIPMENT_CONTAINER	= "NewCarriedEquipmentContainer";	//$NON-NLS-1$
	/** The command for creating a new equipment. */
	public static final String	CMD_NEW_EQUIPMENT					= "NewEquipment";					//$NON-NLS-1$
	/** The command for creating a new equipment container. */
	public static final String	CMD_NEW_EQUIPMENT_CONTAINER			= "NewEquipmentContainer";			//$NON-NLS-1$
	/** The command for incrementing a field within an item. */
	public static final String	CMD_INCREMENT						= "Increment";						//$NON-NLS-1$
	/** The command for decrementing a field within an item. */
	public static final String	CMD_DECREMENT						= "Decrement";						//$NON-NLS-1$
	/** The command for toggling the equipped state of equipment. */
	public static final String	CMD_TOGGLE_EQUIPPED					= "ToggleEquipped";				//$NON-NLS-1$
	private static final String	CMD_UPDATE_CHECK					= "UpdateCheck";					//$NON-NLS-1$
	/** The command for showing the release notes. */
	public static final String	CMD_RELEASE_NOTES					= "ShowReleaseNotes";				//$NON-NLS-1$
	/** The command for showing the todo list. */
	public static final String	CMD_TODO_LIST						= "ShowToDoList";					//$NON-NLS-1$
	/** The command for showing the user's manual. */
	public static final String	CMD_USERS_MANUAL					= "ShowUsersManual";				//$NON-NLS-1$
	/** The command for showing the license. */
	public static final String	CMD_LICENSE							= "ShowLicense";					//$NON-NLS-1$
	/** The command for going to the web site. */
	public static final String	CMD_WEB_SITE						= "WebSite";						//$NON-NLS-1$
	/** The command for going to the mailing list web site. */
	public static final String	CMD_MAILING_LISTS					= "MailingLists";					//$NON-NLS-1$
	/** The command for jumping to the find field. */
	public static final String	CMD_JUMP_TO_FIND					= "JumpToFind";					//$NON-NLS-1$
	private BufferedImage		mWindowIcon;
	private TKMenu				mListMenu;

	/**
	 * Creates character sheet preview window.
	 * 
	 * @param title The window title. May be <code>null</code>.
	 * @param largeIcon The 32x32 window icon. OK to pass in a 16x16 icon here.
	 * @param smallIcon The 16x16 window icon.
	 */
	public CSWindow(String title, BufferedImage largeIcon, BufferedImage smallIcon) {
		super(title, largeIcon);
		mWindowIcon = smallIcon;
		createMenuBar();
		rebuildListMenu();
	}

	private void addMenuItem(TKMenu menu, String cmd) {
		menu.add(new TKMenuItem(CSMenuKeys.getTitle(cmd), CSMenuKeys.getKeyStroke(cmd), cmd));
	}

	private void createMenuBar() {
		TKMenuBar bar = new TKMenuBar(this);
		TKMenu menu = new TKMenu(Msgs.FILE_MENU);

		addMenuItem(menu, CMD_NEW_SHEET);
		addMenuItem(menu, CMD_NEW_TEMPLATE);
		addMenuItem(menu, CMD_NEW_LIST);
		addMenuItem(menu, CMD_OPEN);
		addMenuItem(menu, CMD_ATTEMPT_CLOSE);
		addMenuItem(menu, CMD_SAVE);
		addMenuItem(menu, CMD_SAVE_AS);
		menu.addSeparator();
		addMenuItem(menu, CMD_PAGE_SETUP);
		addMenuItem(menu, CMD_PRINT);
		menu.addSeparator();
		addMenuItem(menu, CMD_QUIT);
		bar.add(menu);

		menu = new TKMenu(Msgs.EDIT_MENU);
		addMenuItem(menu, CMD_UNDO);
		addMenuItem(menu, CMD_REDO);
		menu.addSeparator();
		addMenuItem(menu, CMD_CUT);
		addMenuItem(menu, CMD_COPY);
		addMenuItem(menu, CMD_PASTE);
		addMenuItem(menu, CMD_DUPLICATE);
		addMenuItem(menu, CMD_CLEAR);
		addMenuItem(menu, CMD_SELECT_ALL);
		menu.addSeparator();
		addMenuItem(menu, CMD_INCREMENT);
		addMenuItem(menu, CMD_DECREMENT);
		addMenuItem(menu, CMD_TOGGLE_EQUIPPED);
		menu.addSeparator();
		addMenuItem(menu, CMD_JUMP_TO_FIND);
		menu.addSeparator();
		addMenuItem(menu, CMD_RANDOMIZE_DESCRIPTION);
		addMenuItem(menu, CMD_RANDOMIZE_FEMALE_NAME);
		addMenuItem(menu, CMD_RANDOM_MALE_NAME);
		menu.addSeparator();
		addMenuItem(menu, CMD_ADD_NATURAL_PUNCH);
		addMenuItem(menu, CMD_ADD_NATURAL_KICK);
		addMenuItem(menu, CMD_ADD_NATURAL_KICK_WITH_BOOTS);
		menu.addSeparator();
		addMenuItem(menu, TKOutlineHeaderCM.CMD_RESET_COLUMNS);
		addMenuItem(menu, CMD_RESET_CONFIRMATION_DIALOGS);
		menu.addSeparator();
		addMenuItem(menu, CMD_PREFERENCES);
		bar.add(menu);

		menu = new TKMenu(Msgs.ITEM_MENU);
		addMenuItem(menu, CMD_OPEN_EDITOR);
		addMenuItem(menu, CMD_COPY_TO_SHEET);
		addMenuItem(menu, CMD_COPY_TO_TEMPLATE);
		addMenuItem(menu, CMD_APPLY_TEMPLATE_TO_SHEET);
		menu.addSeparator();
		addMenuItem(menu, CMD_NEW_ADVANTAGE);
		addMenuItem(menu, CMD_NEW_ADVANTAGE_CONTAINER);
		menu.addSeparator();
		addMenuItem(menu, CMD_NEW_SKILL);
		addMenuItem(menu, CMD_NEW_SKILL_CONTAINER);
		addMenuItem(menu, CMD_NEW_TECHNIQUE);
		menu.addSeparator();
		addMenuItem(menu, CMD_NEW_SPELL);
		addMenuItem(menu, CMD_NEW_SPELL_CONTAINER);
		menu.addSeparator();
		addMenuItem(menu, CMD_NEW_CARRIED_EQUIPMENT);
		addMenuItem(menu, CMD_NEW_CARRIED_EQUIPMENT_CONTAINER);
		menu.addSeparator();
		addMenuItem(menu, CMD_NEW_EQUIPMENT);
		addMenuItem(menu, CMD_NEW_EQUIPMENT_CONTAINER);
		bar.add(menu);

		menu = new TKMenu(Msgs.DATA_MENU);
		mListMenu = menu;
		bar.add(menu);

		menu = new TKMenu(Msgs.WINDOW_MENU);
		bar.add(menu);
		bar.mapMenuToKey(menu, KEY_STD_MENUBAR_WINDOW_MENU);

		menu = new TKMenu(Msgs.HELP_MENU);
		addMenuItem(menu, CMD_ABOUT);
		menu.addSeparator();
		addMenuItem(menu, CMD_UPDATE_CHECK);
		menu.addSeparator();
		addMenuItem(menu, CMD_RELEASE_NOTES);
		addMenuItem(menu, CMD_TODO_LIST);
		addMenuItem(menu, CMD_USERS_MANUAL);
		addMenuItem(menu, CMD_LICENSE);
		menu.addSeparator();
		addMenuItem(menu, CMD_WEB_SITE);
		addMenuItem(menu, CMD_MAILING_LISTS);
		bar.add(menu);

		bar.installQAMenu(true);

		setTKMenuBar(bar);
	}

	@Override public BufferedImage getTitleIcon() {
		return mWindowIcon;
	}

	/** Call to force a rebuild of the list menu. */
	public void rebuildListMenu() {
		mListMenu.removeAll();
		addToListMenu(CSListCollectionThread.get().getLists(), mListMenu);
	}

	private void addToListMenu(ArrayList<?> list, TKMenu menu) {
		int count = list.size();

		for (int i = 1; i < count; i++) {
			Object entry = list.get(i);

			if (entry instanceof ArrayList) {
				ArrayList<?> subList = (ArrayList<?>) entry;
				TKMenu subMenu = new TKMenu((String) subList.get(0), TKImage.getFolderIcon());

				addToListMenu(subList, subMenu);
				menu.add(subMenu);
			} else {
				File file = (File) entry;
				TKMenuItem item = new TKMenuItem(TKPath.getLeafName(file.getName(), false), TKFileDialog.getIconForFile(file), CMD_OPEN_LIST);

				item.setUserObject(file);
				menu.add(item);
			}
		}
	}

	@Override public boolean attemptClose() {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();

		if (focus instanceof TKPanel) {
			((TKPanel) focus).windowFocus(false);
		}
		try {
			if (isModified()) {
				int result = TKOptionDialog.confirm(this, MessageFormat.format(Msgs.SAVE_CHANGES, getTitle()));

				if (result == TKDialog.CANCEL) {
					return false;
				}
				if (result == TKDialog.OK) {
					save();
					if (isModified()) {
						// Something went wrong, or the user cancelled
						return false;
					}
				}
			}
		} finally {
			if (focus instanceof TKPanel && isInForeground()) {
				((TKPanel) focus).windowFocus(true);
			}
		}

		return super.attemptClose();
	}

	@Override public void dispose() {
		saveBounds();
		super.dispose();
	}

	/** @return <code>true</code> if the window's contents have been modified. */
	public boolean isModified() {
		return false;
	}

	/** @return The backing file object, if any. */
	public File getBackingFile() {
		return null;
	}

	/**
	 * @return The file filters to be used in the "Save As..." dialog. The first one will be used as
	 *         the default.
	 */
	public TKFileFilter[] getFileFilters() {
		return null;
	}

	/** Saves the window's data. */
	public void save() {
		File file = getBackingFile();

		if (file == null) {
			saveAs();
		} else {
			saveTo(file);
		}
	}

	/**
	 * Saves the window's data to the specified file.
	 * 
	 * @param file The file to create.
	 * @return The name of the file(s) that were actually created.
	 */
	public ArrayList<File> saveTo(@SuppressWarnings("unused") File file) {
		return new ArrayList<File>();
	}

	/** Allows the user to save the file under another name. */
	public void saveAs() {
		TKFileDialog dialog = new TKFileDialog(this, false, false);
		TKFileFilter[] filters = getFileFilters();

		if (filters != null && filters.length > 0) {
			for (TKFileFilter element : filters) {
				dialog.addFileFilter(element);
			}
			dialog.setActiveFileFilter(filters[0]);
		}
		dialog.setFileName(getSaveAsName());
		if (dialog.doModal() == TKDialog.OK) {
			saveTo(dialog.getSelectedItem());
		}
	}

	/** @return The initial "Save As..." name. */
	protected String getSaveAsName() {
		return getTitle();
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		boolean handled = super.adjustMenuItem(command, item);

		if (!handled) {
			handled = true;
			if (CMD_NEW_SHEET.equals(command) || CMD_NEW_TEMPLATE.equals(command) || CMD_NEW_LIST.equals(command) || CMD_SAVE_AS.equals(command) || CMD_PREFERENCES.equals(command) || CMD_OPEN_LIST.equals(command) || CMD_RELEASE_NOTES.equals(command) || CMD_TODO_LIST.equals(command) || CMD_USERS_MANUAL.equals(command) || CMD_LICENSE.equals(command) || CMD_WEB_SITE.equals(command) || CMD_MAILING_LISTS.equals(command)) {
				item.setEnabled(true);
			} else if (CMD_UPDATE_CHECK.equals(command)) {
				item.setTitle(TKUpdateChecker.getResult());
				item.setEnabled(TKUpdateChecker.isNewVersionAvailable());
			} else if (CMD_SAVE.equals(command)) {
				item.setEnabled(isModified());
			} else if (CMD_OPEN_EDITOR.equals(command)) {
				TKOutline outline = getFocusedOutline();
				boolean enabled = false;

				if (outline != null) {
					enabled = outline.getModel().hasSelection();
					if (!enabled) {
						enabled = outline.getEditRow() != null;
					}
				}
				item.setEnabled(enabled);
			} else if (CMD_JUMP_TO_FIND.equals(command)) {
				TKToolBar toolbar = getTKToolBar();

				if (toolbar != null) {
					TKPanel panel = toolbar.getItemForCommand(TKSearch.CMD_SEARCH);

					item.setEnabled(panel != null);
				} else {
					item.setEnabled(false);
				}
			} else if (CMD_INCREMENT.equals(command) || CMD_DECREMENT.equals(command)) {
				item.setTitle(CSMenuKeys.getTitle(command));
				item.setEnabled(false);
			} else {
				handled = false;
			}
		}
		return handled;
	}

	private TKOutline getFocusedOutline() {
		Component comp = getFocusOwner();

		if (comp != null && !(comp instanceof TKOutline)) {
			comp = comp.getParent();
		}
		if (comp instanceof TKOutline) {
			return (TKOutline) comp;
		}
		return null;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		boolean handled = super.obeyCommand(command, item);

		if (!handled) {
			handled = true;
			if (CMD_NEW_SHEET.equals(command)) {
				CSSheetWindow.displaySheetWindow(new CMCharacter());
			} else if (CMD_NEW_TEMPLATE.equals(command)) {
				CSTemplateWindow.displayTemplateWindow(new CMTemplate());
			} else if (CMD_NEW_LIST.equals(command)) {
				CSListOpener.getInstance().newList();
			} else if (CMD_OPEN_LIST.equals(command)) {
				TKOpenManager.openFiles(this, new File[] { (File) item.getUserObject() }, true);
			} else if (CMD_SAVE.equals(command)) {
				save();
			} else if (CMD_SAVE_AS.equals(command)) {
				saveAs();
			} else if (CMD_PREFERENCES.equals(command)) {
				CSPreferencesWindow.display();
			} else if (CMD_UPDATE_CHECK.equals(command)) {
				TKUpdateChecker.goToUpdate();
			} else if (CMD_RELEASE_NOTES.equals(command)) {
				showHTMLFile("release_notes.html"); //$NON-NLS-1$
			} else if (CMD_TODO_LIST.equals(command)) {
				showHTMLFile("todo.html"); //$NON-NLS-1$
			} else if (CMD_USERS_MANUAL.equals(command)) {
				showHTMLFile("guide.html"); //$NON-NLS-1$
			} else if (CMD_LICENSE.equals(command)) {
				showHTMLFile("license.html"); //$NON-NLS-1$
			} else if (CMD_WEB_SITE.equals(command)) {
				TKBrowser.getPreferredBrowser().openURL("http://gcs.trollworks.com"); //$NON-NLS-1$
			} else if (CMD_MAILING_LISTS.equals(command)) {
				TKBrowser.getPreferredBrowser().openURL("http://www.trollworks.com/mailman/listinfo"); //$NON-NLS-1$
			} else if (CMD_OPEN_EDITOR.equals(command)) {
				TKOutline outline = getFocusedOutline();

				if (outline != null) {
					TKOutlineModel model = outline.getModel();

					if (!model.hasSelection()) {
						model.select(outline.getEditRow(), false);
					}
					((CSOutline) outline.getRealOutline()).openDetailEditor(false);
				}
			} else if (CMD_JUMP_TO_FIND.equals(command)) {
				jumpToFindField();
			} else {
				handled = false;
			}
		}
		return handled;
	}

	/** Causes the focus to jump to the field field, if it is present. */
	public void jumpToFindField() {
		TKToolBar toolbar = getTKToolBar();

		if (toolbar != null) {
			TKPanel panel = toolbar.getItemForCommand(TKSearch.CMD_SEARCH);

			if (panel != null) {
				((TKSearch) panel).getFilterField().requestFocusInWindow();
			}
		}
	}

	private void showHTMLFile(String fileName) {
		File docsDir = new File(System.getProperty("app.home", "."), "docs"); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$

		TKBrowser.getPreferredBrowser().openFile(new File(docsDir, fileName));
	}
}
