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

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.character.Namer;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.edit.ConvertToContainer;
import com.trollworks.gcs.menu.edit.Duplicatable;
import com.trollworks.gcs.menu.edit.MoveEquipmentCommand;
import com.trollworks.gcs.menu.item.ApplyTemplateCommand;
import com.trollworks.gcs.menu.item.CopyToSheetCommand;
import com.trollworks.gcs.menu.item.CopyToTemplateCommand;
import com.trollworks.gcs.menu.item.OpenEditorCommand;
import com.trollworks.gcs.menu.item.OpenPageReferenceCommand;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.widget.Menu;
import com.trollworks.gcs.ui.widget.MenuItem;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.utility.Filtered;
import com.trollworks.gcs.utility.I18n;

import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Rectangle;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.MouseEvent;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import javax.swing.JComponent;
import javax.swing.border.CompoundBorder;
import javax.swing.undo.StateEdit;

/** Base outline class. */
public class ListOutline extends Outline implements Runnable, ActionListener, Duplicatable {
    public static final String   OWNING_LIST = "owning_list";
    /** The owning data file. */
    protected           DataFile mDataFile;

    /**
     * Create a new outline.
     *
     * @param dataFile The owning data file.
     * @param model    The outline model to use.
     */
    public ListOutline(DataFile dataFile, OutlineModel model) {
        super(model);
        model.setProperty(OWNING_LIST, this);
        mDataFile = dataFile;
        addActionListener(this);
    }

    /** @return The owning data file. */
    public DataFile getDataFile() {
        return mDataFile;
    }

    @Override
    public void rowsAdded(OutlineModel model, Row[] rows) {
        super.rowsAdded(model, rows);
        mDataFile.notifyOfChange();
    }

    @Override
    public void rowsWereRemoved(OutlineModel model, Row[] rows) {
        super.rowsWereRemoved(model, rows);
        mDataFile.notifyOfChange();
    }

    @Override
    public boolean canDeleteSelection() {
        OutlineModel model = getModel();
        return !model.isLocked() && model.hasSelection();
    }

    @Override
    public void deleteSelection() {
        if (canDeleteSelection()) {
            OutlineModel model = getModel();
            StateEdit    edit  = new StateEdit(model, I18n.text("Remove Rows"));
            Row[]        rows  = model.getSelectionAsList(true).toArray(new Row[0]);
            model.removeSelection();
            for (int i = rows.length - 1; i >= 0; i--) {
                rows[i].removeFromParent();
            }
            if (model.getRowCount() > 0) {
                updateAllRows();
            }
            mDataFile.notifyOfChange();
            edit.end();
            postUndo(edit);
        }
    }

    /**
     * Adds a row at the "best" place (i.e. looks at the selection).
     *
     * @param row     The row to add.
     * @param name    The name for the undo event.
     * @param sibling If the current selection is a container, whether to insert into it, or as a
     *                sibling.
     */
    public void addRow(ListRow row, String name, boolean sibling) {
        addRow(new ListRow[]{row}, name, sibling);
    }

    /**
     * Adds rows at the "best" place (i.e. looks at the selection).
     *
     * @param rows    The rows to add.
     * @param name    The name for the undo event.
     * @param sibling If the current selection is a container, whether to insert into it, or as a
     *                sibling.
     */
    public void addRow(ListRow[] rows, String name, boolean sibling) {
        OutlineModel model = getModel();
        StateEdit    edit  = new StateEdit(model, name);
        List<Row>    sel   = model.getSelectionAsList(true);
        int          count = sel.size();
        int          insertAt;
        int          i;
        Row          parentRow;
        if (count > 0) {
            insertAt = model.getIndexOfRow(sel.get(count == 1 ? 0 : count - 1));
            parentRow = model.getRowAtIndex(insertAt++);
            if (!parentRow.canHaveChildren() || !parentRow.isOpen()) {
                parentRow = parentRow.getParent();
            } else if (sibling) {
                Set<Row> set = new HashSet<>();
                model.collectRowAndDescendantsAtIndex(set, insertAt - 1);
                insertAt += set.size() - 1;
                parentRow = parentRow.getParent();
            }
            if (parentRow != null && parentRow.canHaveChildren()) {
                for (ListRow row : rows) {
                    parentRow.addChild(row);
                }
            }
        } else {
            insertAt = model.getRowCount();
        }
        i = insertAt;
        for (ListRow row : rows) {
            model.addRow(i++, row, true);
        }
        updateAllRows();
        edit.end();
        postUndo(edit);
        revalidate();
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object source  = event.getSource();
        String command = event.getActionCommand();

        if (source instanceof OutlineProxy) {
            source = ((OutlineProxy) source).getRealOutline();
        }
        if (source == this && Outline.CMD_OPEN_SELECTION.equals(command)) {
            openDetailEditor(false);
        }
    }

    /**
     * Opens detailed editors for the current selection.
     *
     * @param later Whether to call {@link EventQueue#invokeLater(Runnable)} rather than immediately
     *              opening the editor.
     */
    public void openDetailEditor(boolean later) {
        requestFocus();
        mRowsToEdit = Filtered.list(getModel().getSelectionAsList(), ListRow.class);
        if (later) {
            EventQueue.invokeLater(this);
        } else {
            run();
        }
    }

    private List<ListRow> mRowsToEdit;

    @Override
    public void run() {
        if (RowEditor.edit(this, mRowsToEdit)) {
            if (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) {
                Namer.name(this, mRowsToEdit);
            }
            updateRows(mRowsToEdit);
            updateRowHeights(mRowsToEdit);
            repaint();
            mDataFile.notifyOfChange();
        }
        mRowsToEdit = null;
    }

    @Override
    public void undoDidHappen(OutlineModel model) {
        super.undoDidHappen(model);
        updateAllRows();
        mDataFile.notifyOfChange();
    }

    @Override
    protected void rowsWereDropped() {
        updateAllRows();
        mDataFile.notifyOfChange();
        requestFocus();
    }

    public void updateAllRows() {
        updateRows(Filtered.list(getModel().getTopLevelRows(), ListRow.class));
    }

    private static void updateRows(Collection<ListRow> rows) {
        for (ListRow row : rows) {
            updateRow(row);
        }
    }

    private static void updateRow(ListRow row) {
        if (row != null) {
            int count = row.getChildCount();
            for (int i = 0; i < count; i++) {
                ListRow child = (ListRow) row.getChild(i);
                updateRow(child);
            }
            row.update();
        }
    }

    /**
     * Adds the row to the list. Recursively descends the row's children and does the same with
     * them.
     *
     * @param list The list to add rows to.
     * @param row  The row to check.
     */
    protected void addRowsToBeProcessed(ArrayList<ListRow> list, ListRow row) {
        int count = row.getChildCount();
        list.add(row);
        for (int i = 0; i < count; i++) {
            addRowsToBeProcessed(list, (ListRow) row.getChild(i));
        }
    }

    @Override
    public void sorted(OutlineModel model) {
        super.sorted(model);
        if (mDataFile.sortingMarksDirty()) {
            mDataFile.setModified(true);
        }
        mDataFile.notifyOfChange();
    }

    @Override
    public void drop(DropTargetDropEvent dtde) {
        super.drop(dtde);
        revalidateView();
    }

    @Override
    protected void showContextMenu(MouseEvent event) {
        if (getModel().hasSelection()) {
            Command.setFocusOwnerOverride(this);
            Menu menu = new Menu();
            if (this instanceof EquipmentOutline && (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template)) {
                addCmdIfEnabled(menu, new MoveEquipmentCommand(getModel().getProperty(EquipmentList.KEY_OTHER_ROOT) != null));
                addCmdIfEnabled(menu, new ConvertToContainer());
            }
            int count = menu.getComponentCount();
            addCmdIfEnabled(menu, new OpenEditorCommand(this));
            if (count > 0 && count != menu.getComponentCount()) {
                menu.addSeparatorAt(count);
            }
            LibraryExplorerDockable library = LibraryExplorerDockable.get();
            if (library != null) {
                count = menu.getComponentCount();
                List<Dockable> dockables = library.getDockContainer().getDock().getDockables();
                for (Dockable dockable : dockables) {
                    if (dockable instanceof SheetDockable) {
                        addCmdIfEnabled(menu, new CopyToSheetCommand(this, (SheetDockable) dockable), 8);
                    }
                }
                if (count > 0 && count != menu.getComponentCount()) {
                    addMenuHeader(menu, CopyToSheetCommand.INSTANCE.getTitle(), count);
                }

                count = menu.getComponentCount();
                for (Dockable dockable : dockables) {
                    if (dockable instanceof TemplateDockable) {
                        addCmdIfEnabled(menu, new CopyToTemplateCommand(this, (TemplateDockable) dockable), 8);
                    }
                }
                if (count > 0 && count != menu.getComponentCount()) {
                    addMenuHeader(menu, CopyToTemplateCommand.INSTANCE.getTitle(), count);
                }

                TemplateDockable template = UIUtilities.getAncestorOfType(this, TemplateDockable.class);
                if (template != null) {
                    count = menu.getComponentCount();
                    for (Dockable dockable : dockables) {
                        if (dockable instanceof SheetDockable) {
                            addCmdIfEnabled(menu, new ApplyTemplateCommand(template, (SheetDockable) dockable), 8);
                        }
                    }
                    if (count > 0 && count != menu.getComponentCount()) {
                        addMenuHeader(menu, ApplyTemplateCommand.INSTANCE.getTitle(), count);
                    }
                }
            }
            count = menu.getComponentCount();
            addCmdIfEnabled(menu, new OpenPageReferenceCommand(this, true));
            addCmdIfEnabled(menu, new OpenPageReferenceCommand(this, false));
            if (count > 0 && count != menu.getComponentCount()) {
                menu.addSeparatorAt(count);
            }
            if (menu.getComponentCount() > 0) {
                Dimension size = menu.getPreferredSize();
                menu.presentToUser((JComponent) event.getComponent(),
                        new Rectangle(event.getX(), event.getY(), size.width, 16), 0, null);
            }
            Command.setFocusOwnerOverride(null);
        }
    }

    private static void addMenuHeader(Menu menu, String header, int index) {
        MenuItem item = new MenuItem(header, null);
        item.setEnabled(false);
        menu.addItemAt(index, item);
        menu.addSeparatorAt(index);
    }

    private static void addCmdIfEnabled(Menu menu, Command cmd) {
        addCmdIfEnabled(menu, cmd, 0);
    }

    private static void addCmdIfEnabled(Menu menu, Command cmd, int indent) {
        cmd.adjust();
        if (cmd.isEnabled()) {
            MenuItem item = new MenuItem(cmd);
            if (indent > 0) {
                item.setBorder(new CompoundBorder(new EmptyBorder(0, indent, 0, 0), item.getBorder()));
            }
            menu.addItem(item);
        }
    }

    @Override
    public boolean canDuplicateSelection() {
        OutlineModel model = getModel();
        return !model.isLocked() && model.hasSelection();
    }

    @Override
    public void duplicateSelection() {
        OutlineModel model = getModel();
        if (!model.isLocked() && model.hasSelection()) {
            List<Row>     rows    = new ArrayList<>();
            List<ListRow> topRows = new ArrayList<>();
            model.setDragRows(model.getSelectionAsList(true).toArray(new Row[0]));
            convertDragRowsToSelf(rows);
            model.setDragRows(null);
            for (Row row : rows) {
                if (row.getDepth() == 0 && row instanceof ListRow) {
                    topRows.add((ListRow) row);
                }
            }
            addRow(topRows.toArray(new ListRow[0]), I18n.text("Duplicate Rows"), true);
            model.select(topRows, false);
            scrollSelectionIntoView();
            mDataFile.notifyOfChange();
        }
    }
}
