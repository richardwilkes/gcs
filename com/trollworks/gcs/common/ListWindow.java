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

package com.trollworks.gcs.common;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.advantage.AdvantageListWindow;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.equipment.EquipmentListWindow;
import com.trollworks.gcs.menu.file.Saveable;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.SkillListWindow;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.spell.SpellListWindow;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.IconButton;
import com.trollworks.gcs.widgets.WindowUtils;
import com.trollworks.gcs.widgets.layout.FlexRow;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.Outline;
import com.trollworks.gcs.widgets.outline.OutlineModel;
import com.trollworks.gcs.widgets.outline.Row;
import com.trollworks.gcs.widgets.outline.RowItemRenderer;
import com.trollworks.gcs.widgets.outline.RowIterator;
import com.trollworks.gcs.widgets.search.Search;
import com.trollworks.gcs.widgets.search.SearchTarget;

import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.io.File;
import java.util.ArrayList;

import javax.swing.ImageIcon;
import javax.swing.JScrollPane;
import javax.swing.JToolBar;
import javax.swing.ListCellRenderer;

/** The common list window superclass. */
public abstract class ListWindow extends AppWindow implements Saveable, ActionListener, SearchTarget {
	private static String		MSG_TOGGLE_ROWS_OPEN_TOOLTIP;
	private static String		MSG_SIZE_COLUMNS_TO_FIT_TOOLTIP;
	private static String		MSG_TOGGLE_EDIT_MODE_TOOLTIP;
	private static String		MSG_SAVE_ERROR;
	private static final String	CONFIG	= "Config";				//$NON-NLS-1$
	/** The outline. */
	protected ListOutline		mOutline;
	/** The list file. */
	protected ListFile			mListFile;
	private Search				mSearch;
	private IconButton			mToggleLockButton;
	private IconButton			mToggleRowsButton;
	private IconButton			mSizeColumnsButton;

	static {
		LocalizedMessages.initialize(ListWindow.class);
	}

	/**
	 * Looks for an existing list window for the specified list.
	 * 
	 * @param list The list to look for.
	 * @return The list window for the specified list, if any.
	 */
	public static ListWindow findListWindow(ListFile list) {
		for (ListWindow window : AppWindow.getWindows(ListWindow.class)) {
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
	public static ListWindow findListWindow(File file) {
		String fullPath = Path.getFullPath(file);

		for (ListWindow window : AppWindow.getWindows(ListWindow.class)) {
			File wFile = window.getList().getFile();

			if (wFile != null) {
				if (Path.getFullPath(wFile).equals(fullPath)) {
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
	 * @return The diplayed window.
	 */
	public static ListWindow displayListWindow(ListFile list) {
		ListWindow window = findListWindow(list);

		if (window == null) {
			if (list instanceof AdvantageList) {
				window = new AdvantageListWindow((AdvantageList) list);
			} else if (list instanceof SkillList) {
				window = new SkillListWindow((SkillList) list);
			} else if (list instanceof SpellList) {
				window = new SpellListWindow((SpellList) list);
			} else if (list instanceof EquipmentList) {
				window = new EquipmentListWindow((EquipmentList) list);
			} else {
				assert false : "Unknown list type"; //$NON-NLS-1$
				return null;
			}
		}
		window.setVisible(true);
		return window;
	}

	/**
	 * Creates a list window.
	 * 
	 * @param listFile The list file this window will display.
	 * @param newItemCmd The command for creating a new item.
	 * @param newContainerCmd The command for creating a new container.
	 */
	protected ListWindow(ListFile listFile, String newItemCmd, String newContainerCmd) {
		super(null, listFile.getFileIcon(true), listFile.getFileIcon(false));
		mListFile = listFile;
		mOutline = createOutline();
		if (listFile.getFile() != null) {
			mOutline.getModel().setLocked(true);
		}
		createToolBar();
		initializeOutline();
		adjustWindowTitle();
		restoreBounds();
		mListFile.setModified(false);
		getUndoManager().discardAllEdits();
		mListFile.setUndoManager(getUndoManager());
	}

	@Override protected void createToolBarContents(JToolBar toolbar, FlexRow row) {
		mToggleLockButton = createToolBarButton(toolbar, row, mOutline.getModel().isLocked() ? Images.getLockedIcon() : Images.getUnlockedIcon(), MSG_TOGGLE_EDIT_MODE_TOOLTIP);
		mToggleRowsButton = createToolBarButton(toolbar, row, Images.getToggleOpenIcon(), MSG_TOGGLE_ROWS_OPEN_TOOLTIP);
		mSizeColumnsButton = createToolBarButton(toolbar, row, Images.getSizeToFitIcon(), MSG_SIZE_COLUMNS_TO_FIT_TOOLTIP);
		mSearch = new Search(this);
		toolbar.add(mSearch);
		row.add(mSearch);
	}

	private IconButton createToolBarButton(JToolBar toolbar, FlexRow row, BufferedImage image, String tooltip) {
		IconButton button = new IconButton(image, tooltip);
		button.addActionListener(this);
		toolbar.add(button);
		row.add(button);
		return button;
	}

	/** @return The underlying {@link ListOutline}. */
	public ListOutline getOutline() {
		return mOutline;
	}

	/** @return The outline to use. */
	protected abstract ListOutline createOutline();

	private void initializeOutline() {
		Preferences prefs = getWindowPreferences();
		String config = prefs.getStringValue(getWindowPrefsPrefix(), CONFIG);
		OutlineModel outlineModel = mOutline.getModel();
		String defConfig;
		JScrollPane scroller;

		mOutline.setDynamicRowHeight(true);
		defConfig = outlineModel.getSortConfig();
		outlineModel.applySortConfig(defConfig);
		scroller = new JScrollPane(mOutline);
		scroller.setColumnHeaderView(mOutline.getHeaderPanel());
		scroller.setMinimumSize(new Dimension(200, 150));
		add(scroller);
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
			title = AppWindow.getNextUntitledWindowName(getClass(), getUntitledName(), this);
		} else {
			title = Path.getLeafName(file.getName(), false);
		}
		setTitle(title);
	}

	/** @return The standard untitled name for this window. */
	protected abstract String getUntitledName();

	@Override public void dispose() {
		getWindowPreferences().setValue(getWindowPrefsPrefix(), CONFIG, mOutline.getConfig());
		mListFile.resetNotifier();
		super.dispose();
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		if (src == mToggleLockButton) {
			OutlineModel model = mOutline.getModel();
			boolean locked = !model.isLocked();
			model.setLocked(locked);
			mToggleLockButton.setIcon(new ImageIcon(locked ? Images.getLockedIcon() : Images.getUnlockedIcon()));
		} else if (src == mToggleRowsButton) {
			mOutline.getModel().toggleRowOpenState();
		} else if (src == mSizeColumnsButton) {
			mOutline.sizeColumnsToFit();
		} else {
			String cmd = event.getActionCommand();
			if (Outline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(cmd)) {
				mListFile.setModified(true);
			}
		}
	}

	/** @return The consumer group for this window. */
	public String getConsumerGroup() {
		return "CSListWindow:" + mListFile.getUniqueID(); //$NON-NLS-1$
	}

	@Override public String getWindowPrefsPrefix() {
		return getConsumerGroup() + "."; //$NON-NLS-1$
	}

	/** @return The list associated with this window. */
	public ListFile getList() {
		return mListFile;
	}

	public ListCellRenderer getSearchRenderer() {
		return new RowItemRenderer();
	}

	public void jumpToSearchField() {
		mSearch.requestFocus();
	}

	public Object[] search(String text) {
		ArrayList<Object> list = new ArrayList<Object>();

		text = text.toLowerCase();
		for (ListRow row : new RowIterator<ListRow>(mOutline.getModel())) {
			if (row.contains(text, true)) {
				list.add(row);
			}
		}
		return list.toArray();
	}

	public void searchSelect(Object[] selection) {
		ArrayList<Row> rows = new ArrayList<Row>();

		for (Object rowObj : selection) {
			Row row = (Row) rowObj;

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

	public boolean isModified() {
		return mListFile != null && mListFile.isModified();
	}

	public String getPreferredSavePath() {
		return getTitle();
	}

	public File getBackingFile() {
		return mListFile.getFile();
	}

	public File[] saveTo(File file) {
		if (mListFile.save(file)) {
			mListFile.setFile(file);
			adjustWindowTitle();
			return new File[] { file };
		}
		WindowUtils.showError(this, MSG_SAVE_ERROR);
		return new File[0];
	}
}
