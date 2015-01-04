/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.common.CommonDockable;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.menu.edit.JumpToSearchTarget;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.menu.RetargetableFocus;
import com.trollworks.toolkit.ui.widget.IconButton;
import com.trollworks.toolkit.ui.widget.Toolbar;
import com.trollworks.toolkit.ui.widget.dock.Dockable;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.RowFilter;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PrintProxy;
import com.trollworks.toolkit.utility.notification.BatchNotifierTarget;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.KeyboardFocusManager;

import javax.swing.DefaultListCellRenderer;
import javax.swing.JComboBox;
import javax.swing.JList;
import javax.swing.JScrollPane;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** A list from a library. */
public abstract class LibraryDockable extends CommonDockable implements RowFilter, DocumentListener, BatchNotifierTarget, JumpToSearchTarget, RetargetableFocus {
	@Localize("Enter text here to narrow the list to only those rows containing matching items")
	@Localize(locale = "de", value = "Hier Text eingeben, um eine Liste der passenden Einträge anzuzeigen")
	@Localize(locale = "ru", value = "Введите текст здесь, чтобы сузить список до содержащих подходящие элементы")
	private static String		SEARCH_FIELD_TOOLTIP;
	@Localize("Any Category")
	@Localize(locale = "de", value = "Beliebige Kategorie")
	@Localize(locale = "ru", value = "Любая категория")
	private static String		CHOOSE_CATEGORY;
	@Localize("Switches between allowing editing and not")
	@Localize(locale = "de", value = "Bearbeitungserlaubnis umschalten")
	@Localize(locale = "ru", value = "Переключение между разрешать редактировать или нет")
	private static String		TOGGLE_EDIT_MODE_TOOLTIP;
	@Localize("Opens/closes all hierarchical rows")
	@Localize(locale = "de", value = "Öffnet / Schließt alle Untereinträge")
	@Localize(locale = "ru", value = "Развернуть/свернуть все вложенные строки")
	private static String		TOGGLE_ROWS_OPEN_TOOLTIP;
	@Localize("Sets the width of each column to exactly fit its contents")
	@Localize(locale = "de", value = "Setzt die Breite jeder Spalte um den Inhalt komplett darzustellen")
	@Localize(locale = "ru", value = "Подобрать ширину каждого столбца по его содержимому")
	private static String		SIZE_COLUMNS_TO_FIT_TOOLTIP;

	static {
		Localization.initialize();
	}

	private Toolbar				mToolbar;
	private JTextField			mFilterField;
	private JComboBox<Object>	mCategoryCombo;
	private IconButton			mLockButton;
	private JScrollPane			mScroller;
	private ListOutline			mOutline;

	/** Creates a new {@link LibraryDockable}. */
	public LibraryDockable(ListFile file) {
		super(file);
		mOutline = createOutline();
		mOutline.setDynamicRowHeight(true);
		OutlineModel outlineModel = mOutline.getModel();
		outlineModel.applySortConfig(outlineModel.getSortConfig());
		outlineModel.setRowFilter(this);
		mToolbar = new Toolbar();
		createFilterField();
		createCategoryCombo();
		mLockButton = new IconButton(outlineModel.isLocked() ? StdImage.LOCKED : StdImage.UNLOCKED, TOGGLE_EDIT_MODE_TOOLTIP, () -> {
			OutlineModel model = mOutline.getModel();
			model.setLocked(!model.isLocked());
			mLockButton.setIcon(model.isLocked() ? StdImage.LOCKED : StdImage.UNLOCKED);
		});
		mToolbar.add(mLockButton);
		mToolbar.add(new IconButton(StdImage.TOGGLE_OPEN, TOGGLE_ROWS_OPEN_TOOLTIP, () -> mOutline.getModel().toggleRowOpenState()));
		mToolbar.add(new IconButton(StdImage.SIZE_TO_FIT, SIZE_COLUMNS_TO_FIT_TOOLTIP, () -> mOutline.sizeColumnsToFit()));
		add(mToolbar, BorderLayout.NORTH);
		mScroller = new JScrollPane(mOutline);
		mScroller.setBorder(null);
		mScroller.setColumnHeaderView(mOutline.getHeaderPanel());
		add(mScroller, BorderLayout.CENTER);
	}

	@Override
	public Component getRetargetedFocus() {
		return mOutline;
	}

	@Override
	public ListFile getDataFile() {
		return (ListFile) super.getDataFile();
	}

	@Override
	public String[] getAllowedExtensions() {
		return new String[] { getDataFile().getExtension() };
	}

	@Override
	public PrintProxy getPrintProxy() {
		// Don't want to allow printing of the library files
		return null;
	}

	@Override
	public String getDescriptor() {
		return getDataFile().getFile().getAbsolutePath();
	}

	/** @return The {@link Toolbar}. */
	public Toolbar getToolbar() {
		return mToolbar;
	}

	/** @return The {@link JScrollPane}. */
	public JScrollPane getScroller() {
		return mScroller;
	}

	/** @return The {@link ListOutline}. */
	public ListOutline getOutline() {
		return mOutline;
	}

	/**
	 * Called to create the {@link ListOutline} for this {@link Dockable}.
	 *
	 * @return The newly created {@link ListOutline}.
	 */
	protected abstract ListOutline createOutline();

	private void createFilterField() {
		mFilterField = new JTextField(10);
		mFilterField.getDocument().addDocumentListener(this);
		mFilterField.setToolTipText(SEARCH_FIELD_TOOLTIP);
		// This client property is specific to Mac OS X
		mFilterField.putClientProperty("JTextField.variant", "search"); //$NON-NLS-1$ //$NON-NLS-2$
		mFilterField.setMinimumSize(new Dimension(60, mFilterField.getPreferredSize().height));
		mToolbar.add(mFilterField, Toolbar.LAYOUT_FILL);
	}

	private void createCategoryCombo() {
		mCategoryCombo = new JComboBox<>();
		// Next two client properties are specific to Mac OS X
		mCategoryCombo.putClientProperty("JComponent.sizeVariant", "small"); //$NON-NLS-1$ //$NON-NLS-2$
		mCategoryCombo.putClientProperty("JComboBox.isPopDown", Boolean.TRUE); //$NON-NLS-1$
		mCategoryCombo.setMaximumRowCount(20);
		mCategoryCombo.setRenderer(new DefaultListCellRenderer() {
			@Override
			public Component getListCellRendererComponent(JList<?> list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
				Component comp = super.getListCellRendererComponent(list, value, index, isSelected, cellHasFocus);
				setFont(getFont().deriveFont(index == 0 ? Font.ITALIC : Font.PLAIN));
				return comp;
			}
		});
		mCategoryCombo.addActionListener((event) -> {
			mCategoryCombo.setFont(mCategoryCombo.getFont().deriveFont(mCategoryCombo.getSelectedIndex() == 0 ? Font.ITALIC : Font.PLAIN));
			mCategoryCombo.revalidate();
			mCategoryCombo.repaint();
			if (mOutline != null) {
				mOutline.reapplyRowFilter();
			}
		});
		mToolbar.add(mCategoryCombo);
		adjustCategoryCombo();
	}

	private void adjustCategoryCombo() {
		mCategoryCombo.removeAllItems();
		mCategoryCombo.addItem(CHOOSE_CATEGORY);
		for (String category : getDataFile().getCategories()) {
			mCategoryCombo.addItem(category);
		}
		mCategoryCombo.setPreferredSize(null);
		mCategoryCombo.setFont(mCategoryCombo.getFont().deriveFont(Font.ITALIC));
		mCategoryCombo.setMinimumSize(new Dimension(36, mCategoryCombo.getPreferredSize().height));
		mCategoryCombo.revalidate();
		mCategoryCombo.repaint();
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

	@Override
	public int getNotificationPriority() {
		return 0;
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
			getOutline().repaint();
		} else {
			adjustCategoryCombo();
		}
	}

	@Override
	public boolean isJumpToSearchAvailable() {
		return mFilterField.isEnabled() && mFilterField != KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
	}

	@Override
	public void jumpToSearchField() {
		mFilterField.requestFocusInWindow();
	}
}
