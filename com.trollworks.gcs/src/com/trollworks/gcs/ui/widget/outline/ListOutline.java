/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.character.names.Namer;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.edit.ConvertToContainer;
import com.trollworks.gcs.menu.edit.MoveEquipmentCommand;
import com.trollworks.gcs.menu.item.ApplyTemplateCommand;
import com.trollworks.gcs.menu.item.CopyToSheetCommand;
import com.trollworks.gcs.menu.item.CopyToTemplateCommand;
import com.trollworks.gcs.menu.item.OpenEditorCommand;
import com.trollworks.gcs.menu.item.OpenPageReferenceCommand;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.utility.FilteredList;
import com.trollworks.gcs.utility.I18n;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.MouseEvent;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import javax.swing.JMenu;
import javax.swing.JPopupMenu;
import javax.swing.undo.StateEdit;

/** Base outline class. */
public class ListOutline extends Outline implements Runnable, ActionListener {
    public static final String   OWNING_LIST = "owning_list";
    /** The owning data file. */
    protected           DataFile mDataFile;
    private             String   mRowSetChangedID;

    /**
     * Create a new outline.
     *
     * @param dataFile        The owning data file.
     * @param model           The outline model to use.
     * @param rowSetChangedID The notification ID to use when the row set changes.
     */
    public ListOutline(DataFile dataFile, OutlineModel model, String rowSetChangedID) {
        super(model);
        model.setProperty(OWNING_LIST, this);
        mDataFile = dataFile;
        mRowSetChangedID = rowSetChangedID;
        addActionListener(this);
    }

    /** @return The owning data file. */
    public DataFile getDataFile() {
        return mDataFile;
    }

    @Override
    public void rowsAdded(OutlineModel model, Row[] rows) {
        super.rowsAdded(model, rows);
        mDataFile.notifySingle(mRowSetChangedID, null);
    }

    @Override
    public void rowsWereRemoved(OutlineModel model, Row[] rows) {
        super.rowsWereRemoved(model, rows);
        mDataFile.notifySingle(mRowSetChangedID, null);
    }

    /** @return The notification ID to use when the row set changes. */
    public String getRowSetChangedID() {
        return mRowSetChangedID;
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
            StateEdit    edit  = new StateEdit(model, I18n.Text("Remove Rows"));
            Row[]        rows  = model.getSelectionAsList(true).toArray(new Row[0]);
            mDataFile.startNotify();
            model.removeSelection();
            for (int i = rows.length - 1; i >= 0; i--) {
                rows[i].removeFromParent();
            }
            if (model.getRowCount() > 0) {
                updateAllRows();
            }
            // Send it out again, since we have a few chicken-and-egg
            // scenarios to deal with... <sigh>
            mDataFile.notify(mRowSetChangedID, null);
            mDataFile.endNotify();
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
     * @return The index of the row that was added.
     */
    public int addRow(ListRow row, String name, boolean sibling) {
        return addRow(new ListRow[]{row}, name, sibling);
    }

    /**
     * Adds rows at the "best" place (i.e. looks at the selection).
     *
     * @param rows    The rows to add.
     * @param name    The name for the undo event.
     * @param sibling If the current selection is a container, whether to insert into it, or as a
     *                sibling.
     * @return The index of the first row that was added.
     */
    public int addRow(ListRow[] rows, String name, boolean sibling) {
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
        return insertAt;
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
        mRowsToEdit = new FilteredList<>(getModel().getSelectionAsList(), ListRow.class);
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
            mDataFile.notifySingle(mRowSetChangedID, null);
        }
        mRowsToEdit = null;
    }

    @Override
    public void undoDidHappen(OutlineModel model) {
        super.undoDidHappen(model);
        updateAllRows();
        mDataFile.notifySingle(mRowSetChangedID, null);
    }

    @Override
    protected void rowsWereDropped() {
        updateAllRows();
        mDataFile.notifySingle(mRowSetChangedID, null);
        requestFocus();
    }

    public void updateAllRows() {
        updateRows(new FilteredList<>(getModel().getTopLevelRows(), ListRow.class));
    }

    private void updateRows(Collection<ListRow> rows) {
        for (ListRow row : rows) {
            updateRow(row);
        }
    }

    private void updateRow(ListRow row) {
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
    }

    @Override
    public void drop(DropTargetDropEvent dtde) {
        super.drop(dtde);
        revalidateView();
    }

    @Override
    protected void showContextMenu(MouseEvent event) {
        if (getModel().hasSelection()) {
            JPopupMenu menu = new JPopupMenu();
            if (this instanceof EquipmentOutline && (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template)) {
                menu.add(new MoveEquipmentCommand(getModel().getProperty(EquipmentList.TAG_OTHER_ROOT) != null));
                menu.add(new ConvertToContainer());
                menu.addSeparator();
            }
            menu.add(new OpenEditorCommand(this));
            LibraryExplorerDockable library = LibraryExplorerDockable.get();
            if (library != null) {
                JMenu          subMenu   = new JMenu(CopyToSheetCommand.INSTANCE.getTitle());
                List<Dockable> dockables = library.getDockContainer().getDock().getDockables();
                for (Dockable dockable : dockables) {
                    if (dockable instanceof SheetDockable) {
                        subMenu.add(new CopyToSheetCommand(this, (SheetDockable) dockable));
                    }
                }
                if (subMenu.getItemCount() == 0) {
                    subMenu.setEnabled(false);
                }
                menu.add(subMenu);

                subMenu = new JMenu(CopyToTemplateCommand.INSTANCE.getTitle());
                for (Dockable dockable : dockables) {
                    if (dockable instanceof TemplateDockable) {
                        subMenu.add(new CopyToTemplateCommand(this, (TemplateDockable) dockable));
                    }
                }
                if (subMenu.getItemCount() == 0) {
                    subMenu.setEnabled(false);
                }
                menu.add(subMenu);

                TemplateDockable template = UIUtilities.getAncestorOfType(this, TemplateDockable.class);
                if (template != null) {
                    subMenu = new JMenu(ApplyTemplateCommand.INSTANCE.getTitle());
                    for (Dockable dockable : dockables) {
                        if (dockable instanceof SheetDockable) {
                            subMenu.add(new ApplyTemplateCommand(template, (SheetDockable) dockable));
                        }
                    }
                    if (subMenu.getItemCount() == 0) {
                        subMenu.setEnabled(false);
                    }
                    menu.add(subMenu);
                }
            }
            menu.addSeparator();
            menu.add(new OpenPageReferenceCommand(this, true));
            menu.add(new OpenPageReferenceCommand(this, false));
            Command.adjustMenuTree(menu);
            menu.show(event.getComponent(), event.getX(), event.getY());
        }
    }
}
