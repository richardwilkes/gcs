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

package com.trollworks.gcs.library;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.menu.edit.JumpToSearchTarget;
import com.trollworks.gcs.menu.file.Saveable;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.utility.notification.BatchNotifierTarget;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.ModifiedMarker;
import com.trollworks.gcs.widgets.ToolBarIconButton;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.WindowSizeEnforcer;
import com.trollworks.gcs.widgets.WindowUtils;
import com.trollworks.gcs.widgets.layout.FlexColumn;
import com.trollworks.gcs.widgets.layout.FlexRow;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.OutlineHeader;
import com.trollworks.gcs.widgets.outline.OutlineModel;
import com.trollworks.gcs.widgets.outline.Row;
import com.trollworks.gcs.widgets.outline.RowFilter;

import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.io.File;

import javax.swing.DefaultListCellRenderer;
import javax.swing.ImageIcon;
import javax.swing.JComboBox;
import javax.swing.JList;
import javax.swing.JScrollPane;
import javax.swing.JTextField;
import javax.swing.JToolBar;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** The library window. */
public class LibraryWindow extends AppWindow implements Saveable, ActionListener, BatchNotifierTarget, RowFilter, DocumentListener, JumpToSearchTarget {
	private static String		MSG_TITLE;
	private static String		MSG_TOGGLE_EDIT_MODE_TOOLTIP;
	private static String		MSG_TOGGLE_ROWS_OPEN_TOOLTIP;
	private static String		MSG_SIZE_COLUMNS_TO_FIT_TOOLTIP;
	private static String		MSG_SAVE_ERROR;
	private static String		MSG_CHOOSE_CATEGORY;
	private static String		MSG_SEARCH_FIELD_TOOLTIP;
	private static final String	ADVANTAGES_CONFIG	= "AdvantagesConfig";	//$NON-NLS-1$
	private static final String	SKILLS_CONFIG		= "SkillsConfig";		//$NON-NLS-1$
	private static final String	SPELLS_CONFIG		= "SpellsConfig";		//$NON-NLS-1$
	private static final String	EQUIPMENT_CONFIG	= "EquipmentConfig";	//$NON-NLS-1$
	static String				MSG_ADVANTAGES;
	static String				MSG_SKILLS;
	static String				MSG_SPELLS;
	static String				MSG_EQUIPMENT;
	private LibraryFile			mFile;
	private JTextField			mFilterField;
	private ToolBarIconButton	mToggleLockButton;
	private ToolBarIconButton	mToggleRowsButton;
	private ToolBarIconButton	mSizeColumnsButton;
	private boolean				mLocked;
	private JComboBox			mTypeCombo;
	private JComboBox			mCategoryCombo;
	private JScrollPane			mScroller;
	/** The current outline. */
	ListOutline					mOutline;
	private ListOutline			mAdvantagesOutline;
	private ListOutline			mSkillsOutline;
	private ListOutline			mSpellsOutline;
	private ListOutline			mEquipmentOutline;

	static {
		LocalizedMessages.initialize(LibraryWindow.class);
	}

	/**
	 * Looks for an existing {@link LibraryWindow} for the specified {@link LibraryFile}.
	 * 
	 * @param file The {@link LibraryFile} to look for.
	 * @return The {@link LibraryWindow} for the specified file, if any.
	 */
	public static LibraryWindow findLibraryWindow(LibraryFile file) {
		for (LibraryWindow window : AppWindow.getWindows(LibraryWindow.class)) {
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
		for (LibraryWindow window : AppWindow.getWindows(LibraryWindow.class)) {
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
	 * @param file The {@link LibraryFile} to display.
	 */
	public LibraryWindow(LibraryFile file) {
		super(MSG_TITLE, Images.getLibraryIcon(true), Images.getLibraryIcon(false));
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
		outline.setRowFilter(this);
	}

	@Override public String getWindowPrefsPrefix() {
		return "LibraryWindow:" + mFile.getUniqueID() + "."; //$NON-NLS-1$ //$NON-NLS-2$ 
	}

	/** Called to adjust the window title to reflect the window's contents. */
	protected void adjustWindowTitle() {
		File file = mFile.getFile();
		String title;
		if (mFile.wasImported()) {
			title = mFile.getSuggestedFileNameFromImport();
		} else if (file == null) {
			title = AppWindow.getNextUntitledWindowName(getClass(), MSG_TITLE, this);
		} else {
			title = Path.getLeafName(file.getName(), false);
		}
		setTitle(title);
		getRootPane().putClientProperty("Window.documentFile", file); //$NON-NLS-1$
	}

	@Override protected void createToolBarContents(JToolBar toolbar, FlexRow row) {
		FlexColumn column = new FlexColumn();
		row.add(column);
		FlexRow firstRow = new FlexRow();
		column.add(firstRow);
		ModifiedMarker marker = new ModifiedMarker();
		mFile.addDataModifiedListener(marker);
		toolbar.add(marker);
		firstRow.add(marker);
		mToggleLockButton = createToolBarButton(toolbar, firstRow, mLocked ? Images.getLockedIcon() : Images.getUnlockedIcon(), MSG_TOGGLE_EDIT_MODE_TOOLTIP);
		mToggleRowsButton = createToolBarButton(toolbar, firstRow, Images.getToggleOpenIcon(), MSG_TOGGLE_ROWS_OPEN_TOOLTIP);
		mSizeColumnsButton = createToolBarButton(toolbar, firstRow, Images.getSizeToFitIcon(), MSG_SIZE_COLUMNS_TO_FIT_TOOLTIP);
		createTypeCombo(toolbar, firstRow);
		createCategoryCombo(toolbar, firstRow);

		FlexRow secondRow = new FlexRow();
		column.add(secondRow);
		mFilterField = new JTextField(20);
		mFilterField.getDocument().addDocumentListener(this);
		mFilterField.setToolTipText(MSG_SEARCH_FIELD_TOOLTIP);
		mFilterField.putClientProperty("JTextField.variant", "search"); //$NON-NLS-1$ //$NON-NLS-2$
		toolbar.add(mFilterField);
		secondRow.add(mFilterField);
	}

	private ToolBarIconButton createToolBarButton(JToolBar toolbar, FlexRow row, BufferedImage image, String tooltip) {
		ToolBarIconButton button = new ToolBarIconButton(image, tooltip);
		button.addActionListener(this);
		toolbar.add(button);
		row.add(button);
		return button;
	}

	private void createCategoryCombo(JToolBar toolbar, FlexRow row) {
		mCategoryCombo = new JComboBox();
		mCategoryCombo.setRenderer(new DefaultListCellRenderer() {
			@Override public Component getListCellRendererComponent(JList list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
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
		mCategoryCombo.addItem(MSG_CHOOSE_CATEGORY);
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
		mTypeCombo = new JComboBox(new Object[] { MSG_ADVANTAGES, MSG_SKILLS, MSG_SPELLS, MSG_EQUIPMENT });
		mTypeCombo.setRenderer(new DefaultListCellRenderer() {
			@Override public Component getListCellRendererComponent(JList list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
				Component comp = super.getListCellRendererComponent(list, value, index, isSelected, cellHasFocus);
				if (value == MSG_ADVANTAGES) {
					setIcon(new ImageIcon(Images.getAdvantageIcon(false, true)));
				}
				if (value == MSG_SKILLS) {
					setIcon(new ImageIcon(Images.getSkillIcon(false, true)));
				}
				if (value == MSG_SPELLS) {
					setIcon(new ImageIcon(Images.getSpellIcon(false, true)));
				}
				if (value == MSG_EQUIPMENT) {
					setIcon(new ImageIcon(Images.getEquipmentIcon(false, true)));
				}
				return comp;
			}
		});
		UIUtilities.setOnlySize(mTypeCombo, mTypeCombo.getPreferredSize());
		if (mAdvantagesOutline.getModel().getRowCount() == 0) {
			if (mSkillsOutline.getModel().getRowCount() > 0) {
				mTypeCombo.setSelectedItem(MSG_SKILLS);
			} else if (mSpellsOutline.getModel().getRowCount() > 0) {
				mTypeCombo.setSelectedItem(MSG_SPELLS);
			} else if (mEquipmentOutline.getModel().getRowCount() > 0) {
				mTypeCombo.setSelectedItem(MSG_EQUIPMENT);
			}
		}
		mTypeCombo.addActionListener(this);
		toolbar.add(mTypeCombo);
		row.add(mTypeCombo);
	}

	@Override public void dispose() {
		Preferences prefs = getWindowPreferences();
		String prefix = getWindowPrefsPrefix();
		prefs.setValue(prefix, ADVANTAGES_CONFIG, mAdvantagesOutline.getConfig());
		prefs.setValue(prefix, SKILLS_CONFIG, mSkillsOutline.getConfig());
		prefs.setValue(prefix, SPELLS_CONFIG, mSpellsOutline.getConfig());
		prefs.setValue(prefix, EQUIPMENT_CONFIG, mEquipmentOutline.getConfig());
		mFile.resetNotifier();
		super.dispose();
	}

	public String[] getAllowedExtensions() {
		return new String[] { LibraryFile.EXTENSION };
	}

	/** @return The {@link LibraryFile}. */
	public LibraryFile getLibraryFile() {
		return mFile;
	}

	public File getBackingFile() {
		return mFile.getFile();
	}

	public String getPreferredSavePath() {
		return getTitle();
	}

	public File[] saveTo(File file) {
		if (mFile.save(file)) {
			mFile.setFile(file);
			adjustWindowTitle();
			return new File[] { file };
		}
		WindowUtils.showError(this, MSG_SAVE_ERROR);
		return new File[0];
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		if (src == mToggleLockButton) {
			mLocked = !mLocked;
			mFile.getAdvantageList().getModel().setLocked(mLocked);
			mFile.getSkillList().getModel().setLocked(mLocked);
			mFile.getSpellList().getModel().setLocked(mLocked);
			mFile.getEquipmentList().getModel().setLocked(mLocked);
			mToggleLockButton.setIcon(new ImageIcon(mLocked ? Images.getLockedIcon() : Images.getUnlockedIcon()));
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
		if (item == MSG_ADVANTAGES) {
			return mAdvantagesOutline;
		} else if (item == MSG_SKILLS) {
			return mSkillsOutline;
		} else if (item == MSG_SPELLS) {
			return mSpellsOutline;
		} else if (item == MSG_EQUIPMENT) {
			return mEquipmentOutline;
		}
		return mAdvantagesOutline;
	}

	private ListFile getListFile() {
		Object item = mTypeCombo.getSelectedItem();
		if (item == MSG_ADVANTAGES) {
			return mFile.getAdvantageList();
		} else if (item == MSG_SKILLS) {
			return mFile.getSkillList();
		} else if (item == MSG_SPELLS) {
			return mFile.getSpellList();
		} else if (item == MSG_EQUIPMENT) {
			return mFile.getEquipmentList();
		}
		return mFile.getAdvantageList();
	}

	public void jumpToSearchField() {
		mFilterField.requestFocusInWindow();
	}

	public void enterBatchMode() {
		// Not needed.
	}

	public void leaveBatchMode() {
		// Not needed.
	}

	public void handleNotification(Object producer, String name, Object data) {
		if (Advantage.ID_TYPE.equals(name)) {
			repaint();
		} else {
			adjustCategoryCombo();
		}
	}

	public boolean isModified() {
		return mFile.isModified();
	}

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

	public void changedUpdate(DocumentEvent event) {
		documentChanged();
	}

	public void insertUpdate(DocumentEvent event) {
		documentChanged();
	}

	public void removeUpdate(DocumentEvent event) {
		documentChanged();
	}

	private void documentChanged() {
		mOutline.reapplyRowFilter();
	}

	/** Switch to the unfiltered Advantages view. */
	public void switchToAdvantages() {
		mTypeCombo.setSelectedItem(MSG_ADVANTAGES);
		mFilterField.setText(""); //$NON-NLS-1$
	}

	/** Switch to the unfiltered Skills view. */
	public void switchToSkills() {
		mTypeCombo.setSelectedItem(MSG_SKILLS);
		mFilterField.setText(""); //$NON-NLS-1$
	}

	/** Switch to the unfiltered Spells view. */
	public void switchToSpells() {
		mTypeCombo.setSelectedItem(MSG_SPELLS);
		mFilterField.setText(""); //$NON-NLS-1$
	}

	/** Switch to the unfiltered Equipment view. */
	public void switchToEquipment() {
		mTypeCombo.setSelectedItem(MSG_EQUIPMENT);
		mFilterField.setText(""); //$NON-NLS-1$
	}
}
