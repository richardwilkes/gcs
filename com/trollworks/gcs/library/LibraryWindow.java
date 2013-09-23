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

package com.trollworks.gcs.library;

import static com.trollworks.gcs.library.LibraryWindow_LS.*;

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
import com.trollworks.gcs.widgets.GCSWindow;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.image.ToolkitImage;
import com.trollworks.ttk.layout.FlexColumn;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.menu.file.Saveable;
import com.trollworks.ttk.notification.BatchNotifierTarget;
import com.trollworks.ttk.preferences.Preferences;
import com.trollworks.ttk.utility.Path;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.utility.WindowSizeEnforcer;
import com.trollworks.ttk.widgets.BaseWindow;
import com.trollworks.ttk.widgets.IconButton;
import com.trollworks.ttk.widgets.ModifiedMarker;
import com.trollworks.ttk.widgets.WindowUtils;
import com.trollworks.ttk.widgets.outline.OutlineHeader;
import com.trollworks.ttk.widgets.outline.OutlineModel;
import com.trollworks.ttk.widgets.outline.Row;
import com.trollworks.ttk.widgets.outline.RowFilter;

import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;

import javax.swing.DefaultListCellRenderer;
import javax.swing.ImageIcon;
import javax.swing.JComboBox;
import javax.swing.JList;
import javax.swing.JScrollPane;
import javax.swing.JTextField;
import javax.swing.JToolBar;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

@Localized({
				@LS(key = "TITLE", msg = "Library"),
				@LS(key = "TOGGLE_EDIT_MODE_TOOLTIP", msg = "Switches between allowing editing and not"),
				@LS(key = "TOGGLE_ROWS_OPEN_TOOLTIP", msg = "Opens/closes all hierarchical rows"),
				@LS(key = "SIZE_COLUMNS_TO_FIT_TOOLTIP", msg = "Sets the width of each column to exactly fit its contents"),
				@LS(key = "SAVE_ERROR", msg = "An error occurred while trying to save the library."),
				@LS(key = "ADVANTAGES", msg = "Advantages"),
				@LS(key = "SKILLS", msg = "Skills"),
				@LS(key = "SPELLS", msg = "Spells"),
				@LS(key = "EQUIPMENT", msg = "Equipment"),
				@LS(key = "CHOOSE_CATEGORY", msg = "Any Category"),
				@LS(key = "SEARCH_FIELD_TOOLTIP", msg = "Enter text here to narrow the list to only those rows containing matching items"),
})
/** The library window. */
public class LibraryWindow extends GCSWindow implements Saveable, ActionListener, BatchNotifierTarget, RowFilter, DocumentListener, JumpToSearchTarget {
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
		String fullPath = Path.getFullPath(file);
		for (LibraryWindow window : BaseWindow.getWindows(LibraryWindow.class)) {
			File wFile = window.getLibraryFile().getFile();
			if (wFile != null) {
				if (Path.getFullPath(wFile).equals(fullPath)) {
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
		super(TITLE, GCSImages.getLibraryIcon(true), GCSImages.getLibraryIcon(false));
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

	/** Called to adjust the window title to reflect the window's contents. */
	protected void adjustWindowTitle() {
		File file = mFile.getFile();
		String title;
		if (mFile.wasImported()) {
			title = mFile.getSuggestedFileNameFromImport();
		} else if (file == null) {
			title = BaseWindow.getNextUntitledWindowName(getClass(), TITLE, this);
		} else {
			title = Path.getLeafName(file.getName(), false);
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

	private IconButton createToolBarButton(JToolBar toolbar, FlexRow row, BufferedImage image, String tooltip) {
		IconButton button = new IconButton(image, tooltip);
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
		for (String category : mFile.getCategoriesFor(getListFile())) {
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
					setIcon(new ImageIcon(GCSImages.getAdvantageIcon(false, true)));
				}
				if (value == SKILLS) {
					setIcon(new ImageIcon(GCSImages.getSkillIcon(false, true)));
				}
				if (value == SPELLS) {
					setIcon(new ImageIcon(GCSImages.getSpellIcon(false, true)));
				}
				if (value == EQUIPMENT) {
					setIcon(new ImageIcon(GCSImages.getEquipmentIcon(false, true)));
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
	public String getPreferredSavePath() {
		return Path.getFullPath(Path.getParent(Path.getFullPath(getBackingFile())), getTitle());
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
			mToggleLockButton.setIcon(new ImageIcon(mLocked ? ToolkitImage.getLockedIcon() : ToolkitImage.getUnlockedIcon()));
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
}
