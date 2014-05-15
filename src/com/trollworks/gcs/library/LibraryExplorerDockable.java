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

import com.trollworks.gcs.common.ListCollectionListener;
import com.trollworks.gcs.common.ListCollectionThread;
import com.trollworks.gcs.menu.edit.JumpToSearchTarget;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.image.ToolkitIcon;
import com.trollworks.toolkit.ui.image.ToolkitImage;
import com.trollworks.toolkit.ui.widget.IconButton;
import com.trollworks.toolkit.ui.widget.Toolbar;
import com.trollworks.toolkit.ui.widget.dock.Dockable;
import com.trollworks.toolkit.ui.widget.tree.FieldAccessor;
import com.trollworks.toolkit.ui.widget.tree.IconAccessor;
import com.trollworks.toolkit.ui.widget.tree.TextTreeColumn;
import com.trollworks.toolkit.ui.widget.tree.TreeContainerRow;
import com.trollworks.toolkit.ui.widget.tree.TreePanel;
import com.trollworks.toolkit.ui.widget.tree.TreeRoot;
import com.trollworks.toolkit.ui.widget.tree.TreeRow;
import com.trollworks.toolkit.ui.widget.tree.TreeRowViewIterator;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.notification.Notifier;

import java.awt.BorderLayout;
import java.awt.Dimension;
import java.awt.KeyboardFocusManager;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

import javax.swing.Icon;
import javax.swing.JPanel;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** A list of available library files. */
public class LibraryExplorerDockable implements Dockable, DocumentListener, JumpToSearchTarget, ListCollectionListener, FieldAccessor, IconAccessor {
	@Localize("Library Explorer")
	private static String	TITLE;
	@Localize("Enter text here to narrow the list to only those rows containing matching items")
	private static String	SEARCH_FIELD_TOOLTIP;
	@Localize("Opens/closes all hierarchical rows")
	private static String	TOGGLE_ROWS_OPEN_TOOLTIP;

	static {
		Localization.initialize();
	}

	private JPanel			mContent;
	private Toolbar			mToolbar;
	private JTextField		mFilterField;
	private TreePanel		mTreePanel;
	private Notifier		mNotifier;

	@Override
	public String getDescriptor() {
		return "library_explorer"; //$NON-NLS-1$
	}

	@Override
	public Icon getTitleIcon() {
		return ToolkitImage.getFolderIcons().getIcon(16);
	}

	@Override
	public String getTitle() {
		return TITLE;
	}

	@Override
	public String getTitleTooltip() {
		return TITLE;
	}

	@Override
	public JPanel getContent() {
		if (mContent == null) {
			ListCollectionThread listCollectionThread = ListCollectionThread.get();
			mNotifier = new Notifier();
			TreeRoot root = new TreeRoot(mNotifier);
			fillTree(listCollectionThread.getLists(), root);
			mTreePanel = new TreePanel(root);
			mTreePanel.addColumn(new TextTreeColumn(TITLE, this, this));
			mContent = new JPanel(new BorderLayout());
			mToolbar = new Toolbar();
			createFilterField();
			mToolbar.add(new IconButton(ToolkitImage.getToggleOpenIcon(), TOGGLE_ROWS_OPEN_TOOLTIP, () -> mTreePanel.toggleDisclosure()));
			mContent.add(mToolbar, BorderLayout.NORTH);
			mContent.add(mTreePanel, BorderLayout.CENTER);
			listCollectionThread.addListener(this);
		}
		return mContent;
	}

	@Override
	public String getField(TreeRow row) {
		return ((LibraryExplorerRow) row).getName();
	}

	@Override
	public ToolkitIcon getIcon(TreeRow row) {
		return ((LibraryExplorerRow) row).getIcon();
	}

	private void createFilterField() {
		mFilterField = new JTextField(10);
		mFilterField.getDocument().addDocumentListener(this);
		mFilterField.setToolTipText(SEARCH_FIELD_TOOLTIP);
		// This client property is specific to Mac OS X
		mFilterField.putClientProperty("JTextField.variant", "search"); //$NON-NLS-1$ //$NON-NLS-2$
		mFilterField.setMinimumSize(new Dimension(60, mFilterField.getPreferredSize().height));
		mToolbar.add(mFilterField, Toolbar.LAYOUT_FILL);
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
		//mOutline.reapplyRowFilter();
	}

	private void fillTree(List<?> lists, TreeContainerRow parent) {
		int count = lists.size();
		for (int i = 1; i < count; i++) {
			Object entry = lists.get(i);
			if (entry instanceof List<?>) {
				List<?> subList = (List<?>) entry;
				LibraryDirectoryRow dir = new LibraryDirectoryRow((String) subList.get(0));
				fillTree(subList, dir);
				parent.addRow(dir);
			} else {
				parent.addRow(new LibraryFileRow((Path) entry));
			}
		}
	}

	@Override
	public void dataFileListUpdated(List<Object> lists) {
		TreeRoot root = mTreePanel.getRoot();
		Set<String> selected = new HashSet<>();
		for (TreeRow row : mTreePanel.getSelectedRows()) {
			selected.add(((LibraryExplorerRow) row).getSelectionKey());
		}
		Set<String> open = new HashSet<>();
		for (TreeRow row : new TreeRowViewIterator(mTreePanel, root)) {
			if (row instanceof TreeContainerRow && mTreePanel.isOpen((TreeContainerRow) row)) {
				open.add(((LibraryExplorerRow) row).getSelectionKey());
			}
		}
		mNotifier.startBatch();
		root.removeRow(new ArrayList<>(root.getChildren()));
		fillTree(lists, root);
		mNotifier.endBatch();
		mTreePanel.setOpen(true, collectRowsToOpen(root, open, null));
		mTreePanel.select(collectRows(root, open, null));
	}

	private List<TreeContainerRow> collectRowsToOpen(TreeContainerRow parent, Set<String> selectors, List<TreeContainerRow> list) {
		if (list == null) {
			list = new ArrayList<>();
		}
		for (TreeRow row : parent.getChildren()) {
			if (row instanceof TreeContainerRow) {
				TreeContainerRow container = (TreeContainerRow) row;
				if (selectors.contains(((LibraryExplorerRow) row).getSelectionKey())) {
					list.add(container);
				}
				collectRowsToOpen(container, selectors, list);
			}
		}
		return list;
	}

	private List<TreeRow> collectRows(TreeContainerRow parent, Set<String> selectors, List<TreeRow> list) {
		if (list == null) {
			list = new ArrayList<>();
		}
		for (TreeRow row : parent.getChildren()) {
			if (selectors.contains(((LibraryExplorerRow) row).getSelectionKey())) {
				list.add(row);
			}
			if (row instanceof TreeContainerRow) {
				collectRows((TreeContainerRow) row, selectors, list);
			}
		}
		return list;
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
