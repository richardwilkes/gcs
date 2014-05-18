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

package com.trollworks.gcs.library;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.menu.edit.JumpToSearchTarget;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.widgets.IconButton;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.WindowSizeEnforcer;
import com.trollworks.toolkit.ui.image.ToolkitIcon;
import com.trollworks.toolkit.ui.image.ToolkitImage;
import com.trollworks.toolkit.ui.layout.FlexColumn;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.menu.file.PrintProxy;
import com.trollworks.toolkit.ui.menu.file.Saveable;
import com.trollworks.toolkit.ui.menu.file.SignificantFrame;
import com.trollworks.toolkit.ui.widget.AppWindow;
import com.trollworks.toolkit.ui.widget.BaseWindow;
import com.trollworks.toolkit.ui.widget.DataModifiedListener;
import com.trollworks.toolkit.ui.widget.ModifiedMarker;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.ui.widget.outline.OutlineHeader;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.RowFilter;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.notification.BatchNotifierTarget;

import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.KeyboardFocusManager;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.io.File;
import java.io.IOException;

import javax.swing.DefaultListCellRenderer;
import javax.swing.JComboBox;
import javax.swing.JList;
import javax.swing.JScrollPane;
import javax.swing.JTextField;
import javax.swing.JToolBar;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** The library window. */
public class LibraryWindow extends AppWindow implements SignificantFrame, Saveable, ActionListener, BatchNotifierTarget, RowFilter, DocumentListener, JumpToSearchTarget {
	@Localize("Library")
	private static String		TITLE;
	@Localize("Switches between allowing editing and not")
	private static String		TOGGLE_EDIT_MODE_TOOLTIP;
	@Localize("Opens/closes all hierarchical rows")
	private static String		TOGGLE_ROWS_OPEN_TOOLTIP;
	@Localize("Sets the width of each column to exactly fit its contents")
	private static String		SIZE_COLUMNS_TO_FIT_TOOLTIP;
	@Localize("An error occurred while trying to save the library.")
	private static String		SAVE_ERROR;
	@Localize("Advantages")
	static String				ADVANTAGES;
	@Localize("Skills")
	static String				SKILLS;
	@Localize("Spells")
	static String				SPELLS;
	@Localize("Equipment")
	static String				EQUIPMENT;
	@Localize("Any Category")
	private static String		CHOOSE_CATEGORY;
	@Localize("Enter text here to narrow the list to only those rows containing matching items")
	private static String		SEARCH_FIELD_TOOLTIP;

	static {
		Localization.initialize();
	}

	private static final String	ADVANTAGES_CONFIG	= "AdvantagesConfig";	//$NON-NLS-1$
	private static final String	SKILLS_CONFIG		= "SkillsConfig";		//$NON-NLS-1$
	private static final String	SPELLS_CONFIG		= "SpellsConfig";		//$NON-NLS-1$
	private static final String	EQUIPMENT_CONFIG	= "EquipmentConfig";	//$NON-NLS-1$
	private LibraryFile			mFile;
	private JTextField			mFilterField;
	private IconButton			mToggleLockButton;
	private IconButton			mToggleRowsButton;
	private IconButton			mSizeColumnsButton;
	private boolean				mLocked;
	private JComboBox<Object>	mTypeCombo;
	private JComboBox<Object>	mCategoryCombo;
	private JScrollPane			mScroller;
	/** The current outline. */
	ListOutline					mOutline;
	private ListOutline			mAdvantagesOutline;
	private ListOutline			mSkillsOutline;
	private ListOutline			mSpellsOutline;
	private ListOutline			mEquipmentOutline;

	/**
	 * Looks for an existing {@link LibraryWindow} for the specified {@link LibraryFile}.
	 *
	 * @param file The {@link LibraryFile} to look for.
	 * @return The {@link LibraryWindow} for the specified file, if any.
	 */
	public static LibraryWindow findLibraryWindow(LibraryFile file) {
		for (LibraryWindow window : BaseWindow.getWindows(LibraryWindow.class)) {
			if (window.getLibraryFile() == file) {
				return window;
			}
		}
		return null;
	}

	/**
	 * Looks for an existing {@link LibraryWindow} for the specified {@link LibraryFile}.
	 *
	 * @param file The {@link LibraryFile} to look for.
	 * @return The {@link LibraryWindow} for the specified file, if any.
	 */
	public static LibraryWindow findLibraryWindow(File file) {
		String fullPath = PathUtils.getFullPath(file);
		for (LibraryWindow window : BaseWindow.getWindows(LibraryWindow.class)) {
			File wFile = window.getLibraryFile().getFile();
			if (wFile != null) {
				if (PathUtils.getFullPath(wFile).equals(fullPath)) {
					return window;
				}
			}
		}
		return null;
	}

	/**
	 * Displays a {@link LibraryWindow} for the specified {@link LibraryFile}.
	 *
	 * @param file The {@link LibraryFile} to display.
	 * @return The displayed {@link LibraryWindow}.
	 */
	public static LibraryWindow displayLibraryWindow(LibraryFile file) {
		LibraryWindow window = findLibraryWindow(file);
		if (window == null) {
			window = new LibraryWindow(file);
		}
		window.setVisible(true);
		return window;
	}

	/**
	 * Creates a new {@link LibraryWindow}.
	 *
	 * @param file The file to display.
	 */
	public LibraryWindow(File file) throws IOException {
		this(new LibraryFile(file));
	}

	/**
	 * Creates a new {@link LibraryWindow}.
	 *
	 * @param file The {@link LibraryFile} to display.
	 */
	public LibraryWindow(LibraryFile file) {
		super(TITLE, GCSImages.getLibraryIcons());
		mFile = file;
		mLocked = mFile.getFile() != null;

		ListFile list = mFile.getAdvantageList();
		mAdvantagesOutline = new AdvantageOutline(mFile, list.getModel());
		list.addTarget(this, Advantage.ID_TYPE);
		list.addTarget(this, Advantage.ID_CATEGORY);

		list = mFile.getSkillList();
		mSkillsOutline = new SkillOutline(mFile, list.getModel());
		list.addTarget(this, Skill.ID_CATEGORY);

		list = mFile.getSpellList();
		mSpellsOutline = new SpellOutline(mFile, list.getModel());
		list.addTarget(this, Spell.ID_CATEGORY);

		list = mFile.getEquipmentList();
		mEquipmentOutline = new EquipmentOutline(mFile, list.getModel());
		list.addTarget(this, Equipment.ID_CATEGORY);

		createToolBar();
		initializeOutline(mAdvantagesOutline, ADVANTAGES_CONFIG);
		initializeOutline(mSkillsOutline, SKILLS_CONFIG);
		initializeOutline(mSpellsOutline, SPELLS_CONFIG);
		initializeOutline(mEquipmentOutline, EQUIPMENT_CONFIG);

		mOutline = getOutlineFromTypeCombo();

		mScroller = new JScrollPane(mOutline);
		mScroller.setMinimumSize(new Dimension(200, 150));
		mScroller.setColumnHeaderView(mOutline.getHeaderPanel());
		add(mScroller);

		pack();
		adjustWindowTitle();
		restoreBounds();

		mFile.setModified(mFile.wasImported());
		getUndoManager().discardAllEdits();
		mFile.setUndoManager(getUndoManager());

		jumpToSearchField();
	}

	private void initializeOutline(ListOutline outline, String configKey) {
		Preferences prefs = getWindowPreferences();
		String config = prefs.getStringValue(getWindowPrefsPrefix(), configKey);
		OutlineModel outlineModel = outline.getModel();
		outline.setDynamicRowHeight(true);
		String defConfig = outlineModel.getSortConfig();
		outlineModel.applySortConfig(defConfig);
		defConfig = outline.getDefaultConfig();
		if (config != null) {
			outline.applyConfig(config);
		}
		outlineModel.setLocked(mLocked);
		outlineModel.setRowFilter(this);
	}

	@Override
	public String getWindowPrefsPrefix() {
		return "LibraryWindow:" + mFile.getUniqueID() + "."; //$NON-NLS-1$ //$NON-NLS-2$
	}

	@Override
	public String getSaveTitle() {
		return getTitle();
	}

	/** Called to adjust the window title to reflect the window's contents. */
	protected void adjustWindowTitle() {
		File file = mFile.getFile();
		String title;
		if (mFile.wasImported()) {
			title = mFile.getSuggestedFileNameFromImport();
		} else if (file == null) {
			title = BaseWindow.getNextUntitledWindowName(getClass(), TITLE, this);
		} else {
			title = PathUtils.getLeafName(file.getName(), false);
		}
		setTitle(title);
		getRootPane().putClientProperty("Window.documentFile", file); //$NON-NLS-1$
	}

	@Override
	protected void createToolBarContents(JToolBar toolbar, FlexRow row) {
		FlexColumn column = new FlexColumn();
		row.add(column);
		FlexRow firstRow = new FlexRow();
		column.add(firstRow);
		ModifiedMarker marker = new ModifiedMarker();
		mFile.addDataModifiedListener(marker);
		toolbar.add(marker);
		firstRow.add(marker);
		mToggleLockButton = createToolBarButton(toolbar, firstRow, mLocked ? ToolkitImage.getLockedIcon() : ToolkitImage.getUnlockedIcon(), TOGGLE_EDIT_MODE_TOOLTIP);
		mToggleRowsButton = createToolBarButton(toolbar, firstRow, ToolkitImage.getToggleOpenIcon(), TOGGLE_ROWS_OPEN_TOOLTIP);
		mSizeColumnsButton = createToolBarButton(toolbar, firstRow, ToolkitImage.getSizeToFitIcon(), SIZE_COLUMNS_TO_FIT_TOOLTIP);
		createTypeCombo(toolbar, firstRow);
		createCategoryCombo(toolbar, firstRow);

		FlexRow secondRow = new FlexRow();
		column.add(secondRow);
		mFilterField = new JTextField(20);
		mFilterField.getDocument().addDocumentListener(this);
		mFilterField.setToolTipText(SEARCH_FIELD_TOOLTIP);
		mFilterField.putClientProperty("JTextField.variant", "search"); //$NON-NLS-1$ //$NON-NLS-2$
		toolbar.add(mFilterField);
		secondRow.add(mFilterField);
	}

	private IconButton createToolBarButton(JToolBar toolbar, FlexRow row, ToolkitIcon icon, String tooltip) {
		IconButton button = new IconButton(icon, tooltip);
		button.addActionListener(this);
		toolbar.add(button);
		row.add(button);
		return button;
	}

	private void createCategoryCombo(JToolBar toolbar, FlexRow row) {
		mCategoryCombo = new JComboBox<>();
		mCategoryCombo.setRenderer(new DefaultListCellRenderer() {
			@Override
			public Component getListCellRendererComponent(JList<?> list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
				Component comp = super.getListCellRendererComponent(list, value, index, isSelected, cellHasFocus);
				setFont(getFont().deriveFont(index == 0 ? Font.ITALIC : Font.PLAIN));
				return comp;
			}
		});
		mCategoryCombo.addActionListener(this);
		toolbar.add(mCategoryCombo);
		row.add(mCategoryCombo);
		adjustCategoryCombo();
	}

	private void adjustCategoryCombo() {
		mCategoryCombo.removeAllItems();
		mCategoryCombo.addItem(CHOOSE_CATEGORY);
		for (String category : getListFile().getCategories()) {
			mCategoryCombo.addItem(category);
		}
		mCategoryCombo.setPreferredSize(null);
		mCategoryCombo.setFont(mCategoryCombo.getFont().deriveFont(Font.ITALIC));
		UIUtilities.setOnlySize(mCategoryCombo, mCategoryCombo.getPreferredSize());
		mCategoryCombo.invalidate();
		mCategoryCombo.repaint();
		WindowSizeEnforcer.enforce(this);
	}

	private void createTypeCombo(JToolBar toolbar, FlexRow row) {
		mTypeCombo = new JComboBox<>(new Object[] { ADVANTAGES, SKILLS, SPELLS, EQUIPMENT });
		mTypeCombo.setRenderer(new DefaultListCellRenderer() {
			@Override
			public Component getListCellRendererComponent(JList<?> list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
				Component comp = super.getListCellRendererComponent(list, value, index, isSelected, cellHasFocus);
				if (value == ADVANTAGES) {
					setIcon(GCSImages.getAdvantageIcon(false, true));
				}
				if (value == SKILLS) {
					setIcon(GCSImages.getSkillIcon(false, true));
				}
				if (value == SPELLS) {
					setIcon(GCSImages.getSpellIcon(false, true));
				}
				if (value == EQUIPMENT) {
					setIcon(GCSImages.getEquipmentIcon(false, true));
				}
				return comp;
			}
		});
		UIUtilities.setOnlySize(mTypeCombo, mTypeCombo.getPreferredSize());
		if (mAdvantagesOutline.getModel().getRowCount() == 0) {
			if (mSkillsOutline.getModel().getRowCount() > 0) {
				mTypeCombo.setSelectedItem(SKILLS);
			} else if (mSpellsOutline.getModel().getRowCount() > 0) {
				mTypeCombo.setSelectedItem(SPELLS);
			} else if (mEquipmentOutline.getModel().getRowCount() > 0) {
				mTypeCombo.setSelectedItem(EQUIPMENT);
			}
		}
		mTypeCombo.addActionListener(this);
		toolbar.add(mTypeCombo);
		row.add(mTypeCombo);
	}

	@Override
	public void dispose() {
		Preferences prefs = getWindowPreferences();
		String prefix = getWindowPrefsPrefix();
		prefs.setValue(prefix, ADVANTAGES_CONFIG, mAdvantagesOutline.getConfig());
		prefs.setValue(prefix, SKILLS_CONFIG, mSkillsOutline.getConfig());
		prefs.setValue(prefix, SPELLS_CONFIG, mSpellsOutline.getConfig());
		prefs.setValue(prefix, EQUIPMENT_CONFIG, mEquipmentOutline.getConfig());
		mFile.resetNotifier();
		super.dispose();
	}

	@Override
	public String[] getAllowedExtensions() {
		return new String[] { LibraryFile.EXTENSION };
	}

	/** @return The {@link LibraryFile}. */
	public LibraryFile getLibraryFile() {
		return mFile;
	}

	@Override
	public File getBackingFile() {
		return mFile.getFile();
	}

	@Override
	public PrintProxy getPrintProxy() {
		return null;
	}

	@Override
	public void toFrontAndFocus() {
		toFront();
		mOutline.requestFocusInWindow();
	}

	@Override
	public String getPreferredSavePath() {
		return PathUtils.getFullPath(PathUtils.getParent(PathUtils.getFullPath(getBackingFile())), getTitle());
	}

	@Override
	public File[] saveTo(File file) {
		if (mFile.save(file)) {
			mFile.setFile(file);
			adjustWindowTitle();
			return new File[] { file };
		}
		WindowUtils.showError(this, SAVE_ERROR);
		return new File[0];
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		if (src == mToggleLockButton) {
			mLocked = !mLocked;
			mFile.getAdvantageList().getModel().setLocked(mLocked);
			mFile.getSkillList().getModel().setLocked(mLocked);
			mFile.getSpellList().getModel().setLocked(mLocked);
			mFile.getEquipmentList().getModel().setLocked(mLocked);
			mToggleLockButton.setIcon(mLocked ? ToolkitImage.getLockedIcon() : ToolkitImage.getUnlockedIcon());
		} else if (src == mToggleRowsButton) {
			mOutline.getModel().toggleRowOpenState();
		} else if (src == mSizeColumnsButton) {
			mOutline.sizeColumnsToFit();
		} else if (src == mTypeCombo) {
			ListOutline saved = mOutline;
			mOutline = getOutlineFromTypeCombo();
			if (saved != mOutline) {
				OutlineHeader headerPanel = saved.getHeaderPanel();
				headerPanel.getParent().remove(headerPanel);
				saved.getParent().remove(saved);
				mScroller.setColumnHeaderView(mOutline.getHeaderPanel());
				mScroller.setViewportView(mOutline);
				adjustCategoryCombo();
				UIUtilities.revalidateImmediately(mScroller);
			}
		} else if (src == mCategoryCombo) {
			mCategoryCombo.setFont(mCategoryCombo.getFont().deriveFont(mCategoryCombo.getSelectedIndex() == 0 ? Font.ITALIC : Font.PLAIN));
			mCategoryCombo.invalidate();
			mCategoryCombo.repaint();
			if (mOutline != null) {
				mOutline.reapplyRowFilter();
			}
		}
	}

	/** @return The current outline. */
	public ListOutline getOutline() {
		return mOutline;
	}

	private ListOutline getOutlineFromTypeCombo() {
		Object item = mTypeCombo.getSelectedItem();
		if (item == ADVANTAGES) {
			return mAdvantagesOutline;
		} else if (item == SKILLS) {
			return mSkillsOutline;
		} else if (item == SPELLS) {
			return mSpellsOutline;
		} else if (item == EQUIPMENT) {
			return mEquipmentOutline;
		}
		return mAdvantagesOutline;
	}

	private ListFile getListFile() {
		Object item = mTypeCombo.getSelectedItem();
		if (item == ADVANTAGES) {
			return mFile.getAdvantageList();
		} else if (item == SKILLS) {
			return mFile.getSkillList();
		} else if (item == SPELLS) {
			return mFile.getSpellList();
		} else if (item == EQUIPMENT) {
			return mFile.getEquipmentList();
		}
		return mFile.getAdvantageList();
	}

	@Override
	public boolean isJumpToSearchAvailable() {
		return mFilterField.isEnabled() && mFilterField != KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
	}

	@Override
	public void jumpToSearchField() {
		mFilterField.requestFocusInWindow();
	}

	@Override
	public void enterBatchMode() {
		// Not needed.
	}

	@Override
	public void leaveBatchMode() {
		// Not needed.
	}

	@Override
	public void handleNotification(Object producer, String name, Object data) {
		if (Advantage.ID_TYPE.equals(name)) {
			repaint();
		} else {
			adjustCategoryCombo();
		}
	}

	@Override
	public boolean isModified() {
		return mFile.isModified();
	}

	@Override
	public boolean isRowFiltered(Row row) {
		boolean filtered = false;
		if (row instanceof ListRow) {
			ListRow listRow = (ListRow) row;
			if (mCategoryCombo.getSelectedIndex() != 0) {
				Object selectedItem = mCategoryCombo.getSelectedItem();
				if (selectedItem != null) {
					filtered = !listRow.getCategories().contains(selectedItem);
				}
			}
			if (!filtered) {
				String filter = mFilterField.getText();
				if (filter.length() > 0) {
					filtered = !listRow.contains(filter.toLowerCase(), true);
				}
			}
		}
		return filtered;
	}

	@Override
	public void changedUpdate(DocumentEvent event) {
		documentChanged();
	}

	@Override
	public void insertUpdate(DocumentEvent event) {
		documentChanged();
	}

	@Override
	public void removeUpdate(DocumentEvent event) {
		documentChanged();
	}

	private void documentChanged() {
		mOutline.reapplyRowFilter();
	}

	/** Switch to the unfiltered Advantages view. */
	public void switchToAdvantages() {
		mTypeCombo.setSelectedItem(ADVANTAGES);
		mFilterField.setText(""); //$NON-NLS-1$
	}

	/** Switch to the unfiltered Skills view. */
	public void switchToSkills() {
		mTypeCombo.setSelectedItem(SKILLS);
		mFilterField.setText(""); //$NON-NLS-1$
	}

	/** Switch to the unfiltered Spells view. */
	public void switchToSpells() {
		mTypeCombo.setSelectedItem(SPELLS);
		mFilterField.setText(""); //$NON-NLS-1$
	}

	/** Switch to the unfiltered Equipment view. */
	public void switchToEquipment() {
		mTypeCombo.setSelectedItem(EQUIPMENT);
		mFilterField.setText(""); //$NON-NLS-1$
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}

	@Override
	public void addDataModifiedListener(DataModifiedListener listener) {
		mFile.addDataModifiedListener(listener);
	}

	@Override
	public void removeDataModifiedListener(DataModifiedListener listener) {
		mFile.removeDataModifiedListener(listener);
	}
}
