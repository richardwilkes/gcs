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

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.advantage.AdvantagesDockable;
import com.trollworks.gcs.app.Workspace;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.common.ListCollectionListener;
import com.trollworks.gcs.common.ListCollectionThread;
import com.trollworks.gcs.equipment.EquipmentDockable;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.SkillsDockable;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.spell.SpellsDockable;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.widgets.search.Search;
import com.trollworks.gcs.widgets.search.SearchTarget;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.Log;
import com.trollworks.toolkit.ui.image.ToolkitIcon;
import com.trollworks.toolkit.ui.image.ToolkitImage;
import com.trollworks.toolkit.ui.menu.edit.Openable;
import com.trollworks.toolkit.ui.menu.file.FileProxy;
import com.trollworks.toolkit.ui.menu.file.RecentFilesMenu;
import com.trollworks.toolkit.ui.widget.IconButton;
import com.trollworks.toolkit.ui.widget.StdFileDialog;
import com.trollworks.toolkit.ui.widget.Toolbar;
import com.trollworks.toolkit.ui.widget.dock.Dock;
import com.trollworks.toolkit.ui.widget.dock.DockLocation;
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
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.notification.Notifier;

import java.awt.BorderLayout;
import java.awt.KeyboardFocusManager;
import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

import javax.swing.Icon;
import javax.swing.ListCellRenderer;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** A list of available library files. */
public class LibraryExplorerDockable extends Dockable implements DocumentListener, SearchTarget, ListCollectionListener, FieldAccessor, IconAccessor, Openable {
	@Localize("Library Explorer")
	private static String	TITLE;
	@Localize("Enter text here to narrow the list to only those rows containing matching items")
	private static String	SEARCH_FIELD_TOOLTIP;
	@Localize("Opens/closes all hierarchical rows")
	private static String	TOGGLE_ROWS_OPEN_TOOLTIP;

	static {
		Localization.initialize();
	}

	private Toolbar			mToolbar;
	private Search			mSearch;
	private TreePanel		mTreePanel;
	private Notifier		mNotifier;

	public static LibraryExplorerDockable get() {
		for (Dockable dockable : Workspace.get().getDock().getDockables()) {
			if (dockable instanceof LibraryExplorerDockable) {
				return (LibraryExplorerDockable) dockable;
			}
		}
		// Shouldn't be possible
		return null;
	}

	public LibraryExplorerDockable() {
		super(new BorderLayout());
		ListCollectionThread listCollectionThread = ListCollectionThread.get();
		mNotifier = new Notifier();
		TreeRoot root = new TreeRoot(mNotifier);
		fillTree(listCollectionThread.getLists(), root);
		mTreePanel = new TreePanel(root);
		mTreePanel.setShowHeader(false);
		mTreePanel.addColumn(new TextTreeColumn(TITLE, this, this));
		mTreePanel.setAllowColumnDrag(false);
		mTreePanel.setAllowColumnResize(false);
		mTreePanel.setAllowColumnContextMenu(false);
		mTreePanel.setAllowRowDropFromExternal(false);
		mTreePanel.setAllowedRowDragTypes(0); // Turns off row dragging
		mTreePanel.setShowRowDivider(false);
		mTreePanel.setShowColumnDivider(false);
		mTreePanel.setUseBanding(false);
		mTreePanel.setUserSortable(false);
		mTreePanel.setOpenableProxy(this);
		mToolbar = new Toolbar();
		mSearch = new Search(this);
		mToolbar.add(mSearch, Toolbar.LAYOUT_FILL);
		mToolbar.add(new IconButton(ToolkitImage.getToggleOpenIcon(), TOGGLE_ROWS_OPEN_TOOLTIP, () -> mTreePanel.toggleDisclosure()));
		add(mToolbar, BorderLayout.NORTH);
		add(mTreePanel, BorderLayout.CENTER);
		listCollectionThread.addListener(this);
	}

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
	public String getField(TreeRow row) {
		return ((LibraryExplorerRow) row).getName();
	}

	@Override
	public ToolkitIcon getIcon(TreeRow row) {
		return ((LibraryExplorerRow) row).getIcon();
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
		for (TreeRow row : mTreePanel.getExplicitlySelectedRows()) {
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
		mTreePanel.select(collectRows(root, selected, null));
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
	public boolean canOpenSelection() {
		return true;
	}

	@Override
	public void openSelection() {
		List<TreeContainerRow> containers = new ArrayList<>();
		boolean hadFile = false;
		for (TreeRow row : mTreePanel.getExplicitlySelectedRows()) {
			if (row instanceof TreeContainerRow) {
				containers.add((TreeContainerRow) row);
			} else {
				open(((LibraryFileRow) row).getPath());
				hadFile = true;
			}
		}
		if (!hadFile) {
			for (TreeContainerRow container : containers) {
				mTreePanel.setOpen(!mTreePanel.isOpen(container), container);
			}
		}
	}

	public FileProxy open(Path path) {
		FileProxy proxy = null;
		// See if it is already open
		for (Dockable dockable : getDockContainer().getDock().getDockables()) {
			if (dockable instanceof FileProxy) {
				proxy = (FileProxy) dockable;
				File file = proxy.getBackingFile();
				if (file != null) {
					try {
						if (Files.isSameFile(path, file.toPath())) {
							dockable.getDockContainer().setCurrentDockable(dockable);
							break;
						}
					} catch (IOException ioe) {
						Log.error(ioe);
					}
				}
				proxy = null;
			}
		}
		if (proxy == null) {
			// If it wasn't, load it and put it into the dock
			try {
				switch (PathUtils.getExtension(path)) {
					case AdvantageList.EXTENSION:
						proxy = openAdvantageList(path);
						break;
					case EquipmentList.EXTENSION:
						proxy = openEquipmentList(path);
						break;
					case SkillList.EXTENSION:
						proxy = openSkillList(path);
						break;
					case SpellList.EXTENSION:
						proxy = openSpellList(path);
						break;
					case LibraryFile.EXTENSION:
						proxy = openLibrary(path);
						break;
					case GURPSCharacter.EXTENSION:
						proxy = dockSheet(new SheetDockable(new GURPSCharacter(path.toFile())));
						break;
					case Template.EXTENSION:
						proxy = dockTemplate(new TemplateDockable(new Template(path.toFile())));
						break;
					default:
						break;
				}
			} catch (Throwable throwable) {
				StdFileDialog.showCannotOpenMsg(this, PathUtils.getLeafName(path, true), throwable);
				proxy = null;
			}
		}
		if (proxy != null) {
			RecentFilesMenu.addRecent(proxy.getBackingFile());
		}
		return proxy;
	}

	private FileProxy openAdvantageList(Path path) throws IOException {
		AdvantageList list = new AdvantageList();
		list.load(path.toFile());
		return dockLibrary(new AdvantagesDockable(list));
	}

	private FileProxy openEquipmentList(Path path) throws IOException {
		EquipmentList list = new EquipmentList();
		list.load(path.toFile());
		return dockLibrary(new EquipmentDockable(list));
	}

	private FileProxy openSkillList(Path path) throws IOException {
		SkillList list = new SkillList();
		list.load(path.toFile());
		return dockLibrary(new SkillsDockable(list));
	}

	private FileProxy openSpellList(Path path) throws IOException {
		SpellList list = new SpellList();
		list.load(path.toFile());
		return dockLibrary(new SpellsDockable(list));
	}

	private FileProxy openLibrary(Path path) throws IOException {
		// RAW: Needs to import these as separate files
		AdvantageList list = new AdvantageList();
		list.load(path.toFile());
		return dockLibrary(new AdvantagesDockable(list));
	}

	/**
	 * @param library The {@link LibraryDockable} to dock.
	 * @return The {@link LibraryDockable} that was passed in.
	 */
	public LibraryDockable dockLibrary(LibraryDockable library) {
		// Order of docking:
		// 1. Stack with another library
		// 2. Dock to the right of a sheet or template
		// 3. Dock to the right of the library explorer
		Dockable other = null;
		Dock dock = getDockContainer().getDock();
		for (Dockable dockable : dock.getDockables()) {
			if (dockable instanceof LibraryDockable) {
				dockable.getDockContainer().stack(library);
				return library;
			}
			if (other == null && (dockable instanceof SheetDockable || dockable instanceof TemplateDockable)) {
				other = dockable;
			}
		}
		if (other == null) {
			other = this;
		}
		dock.dock(library, other, DockLocation.EAST);
		return library;
	}

	/**
	 * @param sheet The {@link SheetDockable} to dock.
	 * @return The {@link SheetDockable} that was passed in.
	 */
	public SheetDockable dockSheet(SheetDockable sheet) {
		// Order of docking:
		// 1. Stack with another sheet
		// 2. Stack with a template
		// 3. Dock to the left of a library
		// 4. Dock to the right of the library explorer
		Dockable template = null;
		Dockable library = null;
		Dock dock = getDockContainer().getDock();
		for (Dockable dockable : dock.getDockables()) {
			if (dockable instanceof SheetDockable) {
				dockable.getDockContainer().stack(sheet);
				return sheet;
			}
			if (template == null && dockable instanceof TemplateDockable) {
				template = dockable;
			} else if (library == null && dockable instanceof LibraryDockable) {
				library = dockable;
			}
		}
		if (template != null) {
			template.getDockContainer().stack(sheet);
		} else if (library != null) {
			dock.dock(sheet, library, DockLocation.WEST);
		} else {
			dock.dock(sheet, this, DockLocation.EAST);
		}
		return sheet;
	}

	/**
	 * @param template The {@link TemplateDockable} to dock.
	 * @return The {@link TemplateDockable} that was passed in.
	 */
	public TemplateDockable dockTemplate(TemplateDockable template) {
		// Order of docking:
		// 1. Stack with another template
		// 2. Stack with a sheet
		// 3. Dock to the left of a library
		// 4. Dock to the right of the library explorer
		Dockable sheet = null;
		Dockable library = null;
		Dock dock = getDockContainer().getDock();
		for (Dockable dockable : dock.getDockables()) {
			if (dockable instanceof TemplateDockable) {
				dockable.getDockContainer().stack(template);
				return template;
			}
			if (sheet == null && dockable instanceof SheetDockable) {
				sheet = dockable;
			} else if (library == null && dockable instanceof LibraryDockable) {
				library = dockable;
			}
		}
		if (sheet != null) {
			sheet.getDockContainer().stack(template);
		} else if (library != null) {
			dock.dock(template, library, DockLocation.WEST);
		} else {
			dock.dock(template, this, DockLocation.EAST);
		}
		return template;
	}

	@Override
	public boolean isJumpToSearchAvailable() {
		return mSearch.isEnabled() && mSearch != KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
	}

	@Override
	public void jumpToSearchField() {
		mSearch.requestFocusInWindow();
	}

	@Override
	public ListCellRenderer<Object> getSearchRenderer() {
		return new LibraryExplorerRowRenderer();
	}

	@Override
	public List<Object> search(String filter) {
		ArrayList<Object> list = new ArrayList<>();
		filter = filter.toLowerCase();
		collect(mTreePanel.getRoot(), filter, list);
		return list;
	}

	private static void collect(TreeRow row, String text, ArrayList<Object> list) {
		if (row instanceof LibraryExplorerRow) {
			if (((LibraryExplorerRow) row).getName().toLowerCase().contains(text)) {
				list.add(row);
			}
		}
		if (row instanceof TreeContainerRow) {
			for (TreeRow child : ((TreeContainerRow) row).getChildren()) {
				collect(child, text, list);
			}
		}
	}

	@Override
	public void searchSelect(List<Object> selection) {
		List<TreeRow> list = new ArrayList<>();
		for (Object one : selection) {
			if (one instanceof TreeRow) {
				list.add((TreeRow) one);

			}
		}
		mTreePanel.setParentsOpen(list);
		mTreePanel.select(list);
		mTreePanel.requestFocus();
	}
}
