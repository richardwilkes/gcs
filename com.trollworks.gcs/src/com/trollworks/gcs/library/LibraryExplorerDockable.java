/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.library;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.advantage.AdvantagesDockable;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.equipment.EquipmentDockable;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.menu.edit.Deletable;
import com.trollworks.gcs.modifier.AdvantageModifierList;
import com.trollworks.gcs.modifier.AdvantageModifiersDockable;
import com.trollworks.gcs.modifier.EquipmentModifierList;
import com.trollworks.gcs.modifier.EquipmentModifiersDockable;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.notes.NotesDockable;
import com.trollworks.gcs.pageref.PDFServer;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.SkillsDockable;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.spell.SpellsDockable;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.FontIcon;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.Search;
import com.trollworks.gcs.ui.widget.SearchTarget;
import com.trollworks.gcs.ui.widget.Toolbar;
import com.trollworks.gcs.ui.widget.Workspace;
import com.trollworks.gcs.ui.widget.dock.Dock;
import com.trollworks.gcs.ui.widget.dock.DockContainer;
import com.trollworks.gcs.ui.widget.dock.DockLayout;
import com.trollworks.gcs.ui.widget.dock.DockLocation;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.FileProxy;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.awt.BorderLayout;
import java.awt.KeyboardFocusManager;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ComponentAdapter;
import java.awt.event.ComponentEvent;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Objects;
import java.util.Set;
import javax.swing.Icon;
import javax.swing.ListCellRenderer;

/** A list of available library files. */
public class LibraryExplorerDockable extends Dockable implements SearchTarget, Deletable, ActionListener {
    private Search  mSearch;
    private Outline mOutline;

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
        OutlineModel model = new OutlineModel();
        model.addColumn(new Column(0, "", "", new LibraryExplorerCell()));
        LibraryDirectoryRow root = new LibraryDirectoryRow("");
        fillTree(LibraryUpdater.collectFiles(), root);
        transferRowsToModel(model, root);
        restoreOpenRows(model, new HashSet<>(Settings.getInstance().getLibraryExplorerOpenRowKeys()));
        mOutline = new Outline(model);
        mOutline.setUseBanding(false);
        mOutline.setAllowRowDrag(false);
        mOutline.setAllowColumnResize(false);
        mOutline.setDeletableProxy(this);
        mOutline.addActionListener(this);
        mOutline.addComponentListener(new ComponentAdapter() {
            @Override
            public void componentResized(ComponentEvent event) {
                mOutline.getModel().getColumnWithID(0).setWidth(mOutline, mOutline.getWidth());
            }
        });
        Toolbar toolbar = new Toolbar();
        mSearch = new Search(this);
        toolbar.add(new FontIconButton(FontAwesome.SITEMAP,
                I18n.text("Opens/closes all hierarchical rows"),
                (b) -> mOutline.getModel().toggleRowOpenState()));
        toolbar.add(new FontIconButton(FontAwesome.SYNC_ALT, I18n.text("Refresh"),
                (b) -> refresh()));
        toolbar.add(mSearch, Toolbar.LAYOUT_FILL);
        add(toolbar, BorderLayout.NORTH);
        ScrollContent content = new ScrollContent(new BorderLayout());
        content.setScrollableTracksViewportWidth(true);
        content.add(mOutline, BorderLayout.CENTER);
        ScrollPanel scrollPanel = new ScrollPanel(content);
        scrollPanel.getViewport().setBackground(mOutline.getBackground());
        add(scrollPanel, BorderLayout.CENTER);
    }

    public void savePreferences() {
        List<String> list = new ArrayList<>(collectOpenRowKeys());
        list.sort(NumericComparator.CASELESS_COMPARATOR);
        Settings prefs = Settings.getInstance();
        prefs.setLibraryExplorerOpenRowKeys(list);
        prefs.setLibraryExplorerDividerPosition(getDockContainer().getDock().getLayout().findLayout(getDockContainer()).getRawDividerPosition());
    }

    @Override
    public Icon getTitleIcon() {
        return new FontIcon(FontAwesome.FOLDER, Fonts.FONT_ICON_LABEL_PRIMARY);
    }

    @Override
    public String getTitle() {
        return I18n.text("Library Explorer");
    }

    @Override
    public String getTitleTooltip() {
        return getTitle();
    }

    private static void fillTree(List<?> lists, LibraryDirectoryRow parent) {
        int count = lists.size();
        for (int i = 1; i < count; i++) {
            Object entry = lists.get(i);
            if (entry instanceof List<?>) {
                List<?>             subList = (List<?>) entry;
                LibraryDirectoryRow dir     = new LibraryDirectoryRow((String) subList.get(0));
                fillTree(subList, dir);
                parent.addChild(dir);
            } else {
                parent.addChild(new LibraryFileRow((Path) entry));
            }
        }
    }

    private static void transferRowsToModel(OutlineModel model, Row root) {
        model.removeAllRows();
        List<Row> rows = new ArrayList<>(root.getChildren());
        for (Row child : rows) {
            child.removeFromParent();
            model.addRow(child, true);
        }
    }

    public void refresh() {
        OutlineModel model    = mOutline.getModel();
        Set<String>  selected = new HashSet<>();
        for (Row row : model.getSelectionAsList()) {
            selected.add(((LibraryExplorerRow) row).getSelectionKey());
        }
        Set<String>         openSet = collectOpenRowKeys();
        LibraryDirectoryRow root    = new LibraryDirectoryRow("");
        fillTree(LibraryUpdater.collectFiles(), root);
        transferRowsToModel(model, root);
        restoreOpenRows(model, openSet);
        restoreSelectedRows(model, selected);
    }

    private Set<String> collectOpenRowKeys() {
        Set<String> openSet = new HashSet<>();
        for (Row row : mOutline.getModel().getTopLevelRows()) {
            if (row instanceof LibraryDirectoryRow) {
                collectOpenRowKeys((LibraryDirectoryRow) row, openSet);
            }
        }
        return openSet;
    }

    private static void collectOpenRowKeys(LibraryDirectoryRow dirRow, Set<String> openSet) {
        if (dirRow.isOpen()) {
            openSet.add(dirRow.getSelectionKey());
            for (Row row : dirRow.getChildren()) {
                if (row instanceof LibraryDirectoryRow) {
                    collectOpenRowKeys((LibraryDirectoryRow) row, openSet);
                }
            }
        }
    }

    private static void restoreOpenRows(OutlineModel model, Set<String> openSet) {
        if (!openSet.isEmpty()) {
            for (Row row : model.getTopLevelRows()) {
                if (row instanceof LibraryDirectoryRow) {
                    restoreOpenRows((LibraryDirectoryRow) row, openSet);
                }
            }
        }
    }

    private static void restoreOpenRows(LibraryDirectoryRow dirRow, Set<String> openSet) {
        dirRow.setOpen(openSet.contains(dirRow.getSelectionKey()));
        for (Row row : dirRow.getChildren()) {
            if (row instanceof LibraryDirectoryRow) {
                restoreOpenRows((LibraryDirectoryRow) row, openSet);
            }
        }
    }

    private static void restoreSelectedRows(OutlineModel model, Set<String> selectedSet) {
        if (!selectedSet.isEmpty()) {
            Set<Row> toSelect = new HashSet<>();
            for (Row row : model.getTopLevelRows()) {
                restoreSelectedRows(row, selectedSet, toSelect);
            }
        }
    }

    private static void restoreSelectedRows(Row row, Set<String> selectedSet, Set<Row> toSelect) {
        if (row instanceof LibraryExplorerRow) {
            if (selectedSet.contains(((LibraryExplorerRow) row).getSelectionKey())) {
                toSelect.add(row);
            }
        }
        if (row instanceof LibraryDirectoryRow && row.isOpen()) {
            for (Row child : row.getChildren()) {
                restoreSelectedRows(child, selectedSet, toSelect);
            }
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (Outline.CMD_OPEN_SELECTION.equals(event.getActionCommand())) {
            List<Row> containers = new ArrayList<>();
            boolean   hadFile    = false;
            for (Row row : mOutline.getModel().getSelectionAsList()) {
                if (row instanceof LibraryFileRow) {
                    open(((LibraryFileRow) row).getFilePath());
                    hadFile = true;
                } else {
                    containers.add(row);
                }
            }
            if (!hadFile) {
                for (Row row : containers) {
                    row.setOpen(!row.isOpen());
                }
            }
        }
    }

    public Dockable getDockableFor(Path path) {
        for (Dockable dockable : getDockContainer().getDock().getDockables()) {
            if (dockable instanceof FileProxy) {
                Path backing = ((FileProxy) dockable).getBackingFile();
                if (backing != null) {
                    try {
                        if (Files.isSameFile(path, backing)) {
                            return dockable;
                        }
                    } catch (IOException ioe) {
                        Log.error(ioe);
                    }
                }
            }
        }
        return null;
    }

    public FileProxy open(Path path) {
        // See if it is already open
        FileProxy proxy = (FileProxy) getDockableFor(path);
        if (proxy == null) {
            // If it wasn't, load it and put it into the dock
            try {
                String ext = PathUtils.getExtension(path);
                if (FileType.ADVANTAGE.matchExtension(ext)) {
                    proxy = openAdvantageList(path);
                } else if (FileType.ADVANTAGE_MODIFIER.matchExtension(ext)) {
                    proxy = openAdvantageModifierList(path);
                } else if (FileType.EQUIPMENT.matchExtension(ext)) {
                    proxy = openEquipmentList(path);
                } else if (FileType.EQUIPMENT_MODIFIER.matchExtension(ext)) {
                    proxy = openEquipmentModifierList(path);
                } else if (FileType.SKILL.matchExtension(ext)) {
                    proxy = openSkillList(path);
                } else if (FileType.SPELL.matchExtension(ext)) {
                    proxy = openSpellList(path);
                } else if (FileType.NOTE.matchExtension(ext)) {
                    proxy = openNoteList(path);
                } else if (FileType.SHEET.matchExtension(ext)) {
                    proxy = dockSheet(new SheetDockable(new GURPSCharacter(path)));
                } else if (FileType.TEMPLATE.matchExtension(ext)) {
                    proxy = dockTemplate(new TemplateDockable(new Template(path)));
                } else if (FileType.PDF.matchExtension(ext)) {
                    PDFServer.showPDF(path, 0);
                }
            } catch (Throwable throwable) {
                Modal.showCannotOpenMsg(this, PathUtils.getLeafName(path, true), throwable);
                proxy = null;
            }
        } else {
            Dockable dockable = (Dockable) proxy;
            dockable.getDockContainer().setCurrentDockable(dockable);
        }
        if (proxy != null) {
            Path backing = proxy.getBackingFile();
            if (backing != null) {
                Settings.getInstance().addRecentFile(backing);
            }
        }
        return proxy;
    }

    private FileProxy openAdvantageList(Path path) throws IOException {
        AdvantageList list = new AdvantageList();
        list.load(path);
        list.getModel().setLocked(true);
        return dockLibrary(new AdvantagesDockable(list));
    }

    private FileProxy openAdvantageModifierList(Path path) throws IOException {
        AdvantageModifierList list = new AdvantageModifierList();
        list.load(path);
        list.getModel().setLocked(true);
        return dockLibrary(new AdvantageModifiersDockable(list));
    }

    private FileProxy openEquipmentList(Path path) throws IOException {
        EquipmentList list = new EquipmentList();
        list.load(path);
        list.getModel().setLocked(true);
        return dockLibrary(new EquipmentDockable(list));
    }

    private FileProxy openEquipmentModifierList(Path path) throws IOException {
        EquipmentModifierList list = new EquipmentModifierList();
        list.load(path);
        list.getModel().setLocked(true);
        return dockLibrary(new EquipmentModifiersDockable(list));
    }

    private FileProxy openSkillList(Path path) throws IOException {
        SkillList list = new SkillList();
        list.load(path);
        list.getModel().setLocked(true);
        return dockLibrary(new SkillsDockable(list));
    }

    private FileProxy openSpellList(Path path) throws IOException {
        SpellList list = new SpellList();
        list.load(path);
        list.getModel().setLocked(true);
        return dockLibrary(new SpellsDockable(list));
    }

    private FileProxy openNoteList(Path path) throws IOException {
        NoteList list = new NoteList();
        list.load(path);
        list.getModel().setLocked(true);
        return dockLibrary(new NotesDockable(list));
    }

    /**
     * @param library The {@link LibraryDockable} to dock.
     * @return The {@link LibraryDockable} that was passed in.
     */
    public LibraryDockable dockLibrary(LibraryDockable library) {
        // Order of docking:
        // 1. Stack with another library
        // 2. Dock to the top of a template
        // 2. Dock to the right of a sheet
        // 3. Dock to the right of the library explorer
        Dockable template = null;
        Dockable sheet    = null;
        Dock     dock     = getDockContainer().getDock();
        for (Dockable dockable : dock.getDockables()) {
            if (dockable instanceof LibraryDockable) {
                dockable.getDockContainer().stack(library);
                return library;
            }
            if (template == null && dockable instanceof TemplateDockable) {
                template = dockable;
            }
            if (sheet == null && dockable instanceof SheetDockable) {
                sheet = dockable;
            }
        }
        if (template != null) {
            dock.dock(library, template, DockLocation.NORTH);
        } else {
            dock.dock(library, Objects.requireNonNullElse(sheet, this), DockLocation.EAST);
        }
        return library;
    }

    /**
     * @param sheet The {@link SheetDockable} to dock.
     * @return The {@link SheetDockable} that was passed in.
     */
    public SheetDockable dockSheet(SheetDockable sheet) {
        // Order of docking:
        // 1. Stack with another sheet
        // 2. Dock to the left of a library or template
        // 3. Dock to the right of the library explorer
        Dockable other = null;
        Dock     dock  = getDockContainer().getDock();
        for (Dockable dockable : dock.getDockables()) {
            if (dockable instanceof SheetDockable) {
                dockable.getDockContainer().stack(sheet);
                return sheet;
            }
            if (other == null && (dockable instanceof TemplateDockable || dockable instanceof LibraryDockable)) {
                other = dockable;
            }
        }
        if (other != null) {
            DockContainer dc     = other.getDockContainer();
            DockLayout    layout = dc.getDock().getLayout().findLayout(dc);
            if (layout.isVertical()) {
                dock.dock(sheet, layout, DockLocation.WEST);
            } else {
                dock.dock(sheet, other, DockLocation.WEST);
            }
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
        // 2. Dock to the bottom of a library
        // 3. Dock to the right of a sheet
        // 4. Dock to the right of the library explorer
        Dockable sheet   = null;
        Dockable library = null;
        Dock     dock    = getDockContainer().getDock();
        for (Dockable dockable : dock.getDockables()) {
            if (dockable instanceof TemplateDockable) {
                dockable.getDockContainer().stack(template);
                return template;
            }
            if (sheet == null && dockable instanceof SheetDockable) {
                sheet = dockable;
            }
            if (library == null && dockable instanceof LibraryDockable) {
                library = dockable;
            }
        }
        if (library != null) {
            dock.dock(template, library, DockLocation.SOUTH);
        } else {
            dock.dock(template, Objects.requireNonNullElse(sheet, this), DockLocation.EAST);
        }
        return template;
    }

    @Override
    public boolean isJumpToSearchAvailable() {
        return mSearch.isEnabled() && mSearch != KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
    }

    @Override
    public void jumpToSearchField() {
        mSearch.requestFocus();
    }

    @Override
    public ListCellRenderer<Object> getSearchRenderer() {
        return new LibraryExplorerRowRenderer();
    }

    @Override
    public List<Object> search(String filter) {
        List<LibraryExplorerSearchResult> list = new ArrayList<>();
        collect(filter.toLowerCase(), list);
        Set<String> titles     = new HashSet<>();
        Set<String> duplicates = new HashSet<>();
        for (LibraryExplorerSearchResult one : list) {
            String title = one.getTitle();
            if (titles.contains(title)) {
                duplicates.add(title);
            } else {
                titles.add(title);
            }
        }
        List<Object> result = new ArrayList<>();
        for (LibraryExplorerSearchResult one : list) {
            if (duplicates.contains(one.getTitle())) {
                one.useFullPath();
            }
            result.add(one);
        }
        return result;
    }

    private void collect(String text, List<LibraryExplorerSearchResult> list) {
        for (Row row : mOutline.getModel().getTopLevelRows()) {
            collect(row, text, list);
        }
    }

    private static void collect(Row row, String text, List<LibraryExplorerSearchResult> list) {
        if (row instanceof LibraryExplorerRow) {
            LibraryExplorerRow libRow = (LibraryExplorerRow) row;
            if (libRow.getName().toLowerCase().contains(text)) {
                list.add(new LibraryExplorerSearchResult(libRow));
            }
        }
        if (row instanceof LibraryDirectoryRow) {
            for (Row child : row.getChildren()) {
                collect(child, text, list);
            }
        }
    }

    @Override
    public void searchSelect(List<Object> selection) {
        List<Row> list = new ArrayList<>();
        for (Object one : selection) {
            if (one instanceof LibraryExplorerSearchResult) {
                LibraryExplorerRow row = ((LibraryExplorerSearchResult) one).getRow();
                if (row instanceof Row) {
                    list.add((Row) row);
                }
            }
        }
        mOutline.getModel().openAllParents(list);
        mOutline.getModel().select(list, false);
        mOutline.requestFocus();
    }

    @Override
    public boolean canDeleteSelection() {
        return !collectSelectedFilePaths().isEmpty();
    }

    @Override
    public void deleteSelection() {
        List<Path> paths = collectSelectedFilePaths();
        if (!paths.isEmpty()) {
            Modal dialog = Modal.prepareToShowMessage(this,
                    paths.size() == 1 ?
                            I18n.text("Delete File") :
                            String.format(I18n.text("Delete {0} Files"), Integer.valueOf(paths.size())),
                    MessageType.QUESTION,
                    paths.size() == 1 ?
                            I18n.text("Are you sure you want to delete this file?") :
                            I18n.text("Are you sure you want to delete these files?"));
            dialog.addCancelButton();
            dialog.addButton(I18n.text("Delete"), Modal.OK);
            dialog.presentToUser();
            if (dialog.getResult() == Modal.OK) {
                int failed = 0;
                for (Path p : paths) {
                    FileProxy proxy = (FileProxy) getDockableFor(p);
                    if (proxy != null) {
                        Dockable      dockable = (Dockable) proxy;
                        DockContainer dc       = dockable.getDockContainer();
                        dc.close(dockable);
                    }
                    try {
                        Files.deleteIfExists(p);
                    } catch (IOException exception) {
                        failed++;
                    }
                }
                refresh();
                if (failed != 0) {
                    Modal.showError(this, failed == 1 ? I18n.text("A file could not be deleted.") : String.format(I18n.text("{0} files could not be deleted."), Integer.valueOf(failed)));
                }
            }
        }
    }

    private List<Path> collectSelectedFilePaths() {
        Set<Path> set = new HashSet<>();
        for (Row row : mOutline.getModel().getSelectionAsList()) {
            collectFilePaths(row, set);
        }
        List<Path> list = new ArrayList<>(set);
        Collections.sort(list);
        return list;
    }

    private static void collectFilePaths(Row row, Set<Path> set) {
        if (row instanceof LibraryDirectoryRow) {
            for (Row child : row.getChildren()) {
                collectFilePaths(child, set);
            }
        } else {
            set.add(((LibraryFileRow) row).getFilePath());
        }
    }
}
