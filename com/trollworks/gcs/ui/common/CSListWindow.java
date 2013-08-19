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

import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.advantage.CMAdvantageList;
import com.trollworks.gcs.model.equipment.CMEquipmentList;
import com.trollworks.gcs.model.skill.CMSkillList;
import com.trollworks.gcs.model.spell.CMSpellList;
import com.trollworks.gcs.ui.advantage.CSAdvantageListWindow;
import com.trollworks.gcs.ui.equipment.CSEquipmentListWindow;
import com.trollworks.gcs.ui.sheet.CSSheetWindow;
import com.trollworks.gcs.ui.skills.CSSkillListWindow;
import com.trollworks.gcs.ui.spell.CSSpellListWindow;
import com.trollworks.gcs.ui.template.CSTemplateWindow;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKItemRenderer;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKToolBar;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.outline.TKRowIterator;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.search.TKSearch;
import com.trollworks.toolkit.widget.search.TKSearchTarget;
import com.trollworks.toolkit.window.TKOptionDialog;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.io.File;
import java.util.ArrayList;
import java.util.Collection;

/** The common list window superclass. */
public abstract class CSListWindow extends CSWindow implements ActionListener, TKSearchTarget {
	private static final String	CONFIG					= "Config";			//$NON-NLS-1$
	private static final String	CMD_TOGGLE_EDIT_MODE	= "ToggleEditMode";	//$NON-NLS-1$
	private static final String	CMD_SIZE_COLUMNS_TO_FIT	= "SizeColumnsToFit";	//$NON-NLS-1$
	private static final String	CMD_TOGGLE_ROWS_OPEN	= "ToggleRowsOpen";	//$NON-NLS-1$
	/** The outline. */
	protected CSOutline			mOutline;
	/** The list file. */
	protected CMListFile		mListFile;
	private String				mNewItemCmd;
	private String				mNewContainerCmd;

	/**
	 * Looks for an existing list window for the specified list.
	 * 
	 * @param list The list to look for.
	 * @return The list window for the specified list, if any.
	 */
	public static CSListWindow findListWindow(CMListFile list) {
		for (CSListWindow window : TKWindow.getWindows(CSListWindow.class)) {
			if (window.getList() == list) {
				return window;
			}
		}

		return null;
	}

	/**
	 * Looks for an existing list window for the specified file.
	 * 
	 * @param file The list file to look for.
	 * @return The list window for the specified file, if any.
	 */
	public static CSListWindow findListWindow(File file) {
		String fullPath = TKPath.getFullPath(file);

		for (CSListWindow window : TKWindow.getWindows(CSListWindow.class)) {
			File wFile = window.getList().getFile();

			if (wFile != null) {
				if (TKPath.getFullPath(wFile).equals(fullPath)) {
					return window;
				}
			}
		}
		return null;
	}

	/**
	 * Displays a list for the specified list.
	 * 
	 * @param list The list to display.
	 */
	public static void displayListWindow(CMListFile list) {
		CSListWindow window = findListWindow(list);

		if (window == null) {
			if (list instanceof CMAdvantageList) {
				window = new CSAdvantageListWindow((CMAdvantageList) list);
			} else if (list instanceof CMSkillList) {
				window = new CSSkillListWindow((CMSkillList) list);
			} else if (list instanceof CMSpellList) {
				window = new CSSpellListWindow((CMSpellList) list);
			} else if (list instanceof CMEquipmentList) {
				window = new CSEquipmentListWindow((CMEquipmentList) list);
			} else {
				assert false : "Unknown list type"; //$NON-NLS-1$
				return;
			}
		}
		window.setVisible(true);
	}

	/**
	 * Creates a list window.
	 * 
	 * @param listFile The list file this window will display.
	 * @param newItemCmd The command for creating a new item.
	 * @param newContainerCmd The command for creating a new container.
	 */
	protected CSListWindow(CMListFile listFile, String newItemCmd, String newContainerCmd) {
		super(null, listFile.getFileIcon(true), listFile.getFileIcon(false));
		mListFile = listFile;
		mNewItemCmd = newItemCmd;
		mNewContainerCmd = newContainerCmd;
		mOutline = createOutline();
		if (listFile.getFile() != null) {
			mOutline.getModel().setLocked(true);
		}
		createToolBar();
		initializeOutline();
		adjustWindowTitle();
		restoreBounds();
		mListFile.setModified(false);
		mListFile.discardAllEdits();
		setUndoManager(mListFile);
	}

	private void createToolBar() {
		TKToolBar toolbar = new TKToolBar(this);
		TKButton button;
		TKSearch search;

		button = new TKButton(mOutline.getModel().isLocked() ? CSImage.getLockedIcon() : CSImage.getUnlockedIcon());
		button.setToolTipText(Msgs.TOGGLE_EDIT_MODE_TOOLTIP);
		toolbar.addControl(button, -1, CMD_TOGGLE_EDIT_MODE);
		toolbar.addSpacer();
		button = new TKButton(TKImage.getToggleOpenIcon());
		button.setToolTipText(Msgs.TOGGLE_ROWS_OPEN_TOOLTIP);
		toolbar.addControl(button, -1, CMD_TOGGLE_ROWS_OPEN);
		button = new TKButton(TKImage.getSizeToFitIcon());
		button.setToolTipText(Msgs.SIZE_COLUMNS_TO_FIT_TOOLTIP);
		toolbar.addControl(button, -1, CMD_SIZE_COLUMNS_TO_FIT);
		toolbar.addSpacer();
		search = new TKSearch(toolbar, CSFont.KEY_FIELD, CSFont.KEY_LABEL);
		toolbar.addControl(search, -1, TKSearch.CMD_SEARCH);

		setTKToolBar(toolbar);
	}

	@Override public boolean adjustToolBarItem(String command, TKPanel item) {
		if (CMD_SIZE_COLUMNS_TO_FIT.equals(command) || CMD_TOGGLE_ROWS_OPEN.equals(command) || CMD_TOGGLE_EDIT_MODE.equals(command) || TKSearch.CMD_SEARCH.equals(command)) {
			item.setEnabled(true);
		} else {
			return super.adjustToolBarItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyToolBarCommand(String command, TKPanel item) {
		if (CMD_SIZE_COLUMNS_TO_FIT.equals(command)) {
			mOutline.sizeColumnsToFit();
		} else if (CMD_TOGGLE_ROWS_OPEN.equals(command)) {
			mOutline.getModel().toggleRowOpenState();
		} else if (CMD_TOGGLE_EDIT_MODE.equals(command)) {
			TKOutlineModel model = mOutline.getModel();
			boolean locked = !model.isLocked();

			model.setLocked(locked);
			((TKButton) item).setImage(locked ? CSImage.getLockedIcon() : CSImage.getUnlockedIcon());
		} else {
			return super.obeyToolBarCommand(command, item);
		}
		return true;
	}

	/** @return The outline to use. */
	protected abstract CSOutline createOutline();

	private void initializeOutline() {
		TKPreferences prefs = getWindowPreferences();
		String config = prefs.getStringValue(getWindowPrefsPrefix(), CONFIG);
		TKOutlineModel outlineModel = mOutline.getModel();
		String defConfig;
		TKScrollPanel scroller;

		mOutline.setDynamicRowHeight(true);
		defConfig = outlineModel.getSortConfig();
		outlineModel.applySortConfig(defConfig);
		scroller = new TKScrollPanel(mOutline);
		scroller.setHorizontalHeader(mOutline.getHeaderPanel());
		scroller.setMinimumSize(new Dimension(200, 150));
		scroller.getContentBorderView().setBorder(new TKLineBorder(TKColor.SCROLL_BAR_LINE, 1, TKLineBorder.TOP_EDGE, false));
		setContent(scroller);
		defConfig = mOutline.getDefaultConfig();
		if (config != null) {
			mOutline.applyConfig(config);
		}
		pack();
		mOutline.addActionListener(this);
	}

	/** Called to adjust the window title to reflect the window's contents. */
	protected void adjustWindowTitle() {
		File file = mListFile.getFile();
		String title;

		if (file == null) {
			title = TKWindow.getNextUntitledWindowName(getClass(), getUntitledName(), this);
		} else {
			title = TKPath.getLeafName(file.getName(), false);
		}
		setTitle(title);
	}

	/** @return The standard untitled name for this window. */
	protected abstract String getUntitledName();

	@Override public void dispose() {
		getWindowPreferences().setValue(getWindowPrefsPrefix(), CONFIG, mOutline.getConfig());
		mListFile.resetNotifier();
		mListFile.noLongerNeeded();
		super.dispose();
	}

	public void actionPerformed(ActionEvent event) {
		String cmd = event.getActionCommand();

		if (TKOutline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(cmd)) {
			mListFile.setModified(true);
		}
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (mNewItemCmd.equals(command) || mNewContainerCmd.equals(command)) {
			item.setEnabled(!mOutline.getModel().isLocked());
		} else if (CMD_COPY_TO_SHEET.equals(command)) {
			item.setEnabled(mOutline.getModel().hasSelection() && CSSheetWindow.getTopSheet() != null);
		} else if (CMD_COPY_TO_TEMPLATE.equals(command)) {
			item.setEnabled(mOutline.getModel().hasSelection() && CSTemplateWindow.getTopTemplate() != null);
		} else {
			return super.adjustMenuItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (mNewItemCmd.equals(command)) {
			addRow(mListFile.createNewRow(false), item.getTitle());
		} else if (mNewContainerCmd.equals(command)) {
			addRow(mListFile.createNewRow(true), item.getTitle());
		} else if (CMD_COPY_TO_SHEET.equals(command)) {
			copySelectionToSheet();
		} else if (CMD_COPY_TO_TEMPLATE.equals(command)) {
			copySelectionToTemplate();
		} else if (CMD_REDO.equals(command) || CMD_UNDO.equals(command)) {
			mOutline.setNoEditOnNextKeyboardFocus();
			mOutline.stopEditing();
			if (super.obeyCommand(command, item)) {
				repaint();
			}
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	private void copySelectionToSheet() {
		TKOutlineModel outlineModel = mOutline.getModel();

		if (outlineModel.hasSelection()) {
			CSSheetWindow sheet = CSSheetWindow.getTopSheet();

			if (sheet != null) {
				sheet.addRows(outlineModel.getSelectionAsList(true));
			}
		}
	}

	private void copySelectionToTemplate() {
		TKOutlineModel outlineModel = mOutline.getModel();

		if (outlineModel.hasSelection()) {
			CSTemplateWindow template = CSTemplateWindow.getTopTemplate();

			if (template != null) {
				template.addRows(outlineModel.getSelectionAsList(true));
			}
		}
	}

	/**
	 * @param row The row to add.
	 * @param name The name for the undo.
	 */
	protected void addRow(CMRow row, String name) {
		mOutline.addRow(row, name, false);
		mOutline.startEditing(row, true);
	}

	@Override public boolean isModified() {
		return mListFile != null && mListFile.isModified();
	}

	@Override public File getBackingFile() {
		return mListFile.getFile();
	}

	/** @return The consumer group for this window. */
	public String getConsumerGroup() {
		return "CSListWindow:" + mListFile.getUniqueID(); //$NON-NLS-1$
	}

	@Override public String getWindowPrefsPrefix() {
		return getConsumerGroup() + "."; //$NON-NLS-1$
	}

	@Override public ArrayList<File> saveTo(File file) {
		ArrayList<File> result = new ArrayList<File>();

		if (mListFile.save(file)) {
			result.add(file);
			mListFile.setFile(file);
			adjustWindowTitle();
		} else {
			TKOptionDialog.error(this, Msgs.SAVE_ERROR);
		}
		return result;
	}

	/** @return The list associated with this window. */
	public CMListFile getList() {
		return mListFile;
	}

	@Override public TKFileFilter getPreferredFileFilter(TKFileFilter[] filters) {
		return CSFileOpener.getPreferredFileFilter(filters);
	}

	public TKItemRenderer getSearchRenderer() {
		return new CSRowItemRenderer();
	}

	public Collection<Object> search(String text) {
		ArrayList<Object> list = new ArrayList<Object>();

		text = text.toLowerCase();
		for (CMRow row : new TKRowIterator<CMRow>(mOutline.getModel())) {
			if (row.contains(text, true)) {
				list.add(row);
			}
		}
		return list;
	}

	public void searchSelect(Collection<Object> selection) {
		ArrayList<TKRow> rows = new ArrayList<TKRow>();

		for (Object rowObj : selection) {
			TKRow row = (TKRow) rowObj;

			rows.add(row);
			row = row.getParent();
			while (row != null) {
				row.setOpen(true);
				row = row.getParent();
			}
		}

		mOutline.getModel().select(rows, false);
		EventQueue.invokeLater(new Runnable() {
			public void run() {
				mOutline.scrollSelectionIntoView();
			}
		});
		mOutline.requestFocus();
	}
}
