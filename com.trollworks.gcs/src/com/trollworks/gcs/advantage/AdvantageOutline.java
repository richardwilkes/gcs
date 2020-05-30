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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.menu.edit.Incrementable;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.MultipleRowUndo;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowPostProcessor;
import com.trollworks.gcs.ui.widget.outline.RowUndo;
import com.trollworks.gcs.utility.FilteredIterator;
import com.trollworks.gcs.utility.I18n;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.util.ArrayList;
import java.util.List;

/** An outline specifically for Advantages. */
public class AdvantageOutline extends ListOutline implements Incrementable {
    private static OutlineModel extractModel(DataFile dataFile) {
        if (dataFile instanceof GURPSCharacter) {
            return ((GURPSCharacter) dataFile).getAdvantagesModel();
        }
        if (dataFile instanceof Template) {
            return ((Template) dataFile).getAdvantagesModel();
        }
        return ((ListFile) dataFile).getModel();
    }

    /**
     * Create a new Advantages, Disadvantages & Quirks outline.
     *
     * @param dataFile The owning data file.
     */
    public AdvantageOutline(DataFile dataFile) {
        this(dataFile, extractModel(dataFile));
    }

    /**
     * Create a new Advantages, Disadvantages & Quirks outline.
     *
     * @param dataFile The owning data file.
     * @param model    The {@link OutlineModel} to use.
     */
    public AdvantageOutline(DataFile dataFile, OutlineModel model) {
        super(dataFile, model, Advantage.ID_LIST_CHANGED);
        AdvantageColumn.addColumns(this, dataFile);
    }

    @Override
    protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
        return !getModel().isLocked() && rows.length > 0 && (rows[0] instanceof Advantage || rows[0] instanceof AdvantageModifier);
    }

    @Override
    public void convertDragRowsToSelf(List<Row> list) {
        OutlineModel       model              = getModel();
        Row[]              rows               = model.getDragRows();
        boolean            forSheetOrTemplate = mDataFile instanceof GURPSCharacter || mDataFile instanceof Template;
        ArrayList<ListRow> process            = new ArrayList<>();
        for (Row element : rows) {
            Advantage advantage = new Advantage(mDataFile, (Advantage) element, true);
            model.collectRowsAndSetOwner(list, advantage, false);
            if (forSheetOrTemplate) {
                addRowsToBeProcessed(process, advantage);
            }
        }
        if (forSheetOrTemplate && !process.isEmpty()) {
            EventQueue.invokeLater(new RowPostProcessor(this, process));
        }
    }

    protected boolean isDropOnRow(Row[] dragRows) {
        return dragRows != null && dragRows.length > 0 && dragRows[0] instanceof AdvantageModifier;
    }

    @Override
    protected boolean dropOnRow(DropTargetDropEvent dtde) {
        Row target = getDragTargetRow();
        if (!(target instanceof Advantage)) {
            return false;
        }
        Advantage               targetAdvantage = (Advantage) target;
        OutlineModel            model           = getModel();
        ArrayList<RowUndo>      undoList        = new ArrayList<>();
        RowUndo                 undo            = new RowUndo(targetAdvantage);
        List<AdvantageModifier> list            = new ArrayList<>(targetAdvantage.getModifiers());
        Object                  property        = model.getProperty(ListOutline.OWNING_LIST);
        removeDragHighlight(this);
        if (property instanceof ListOutline) {
            List<AdvantageModifier> collection = new ArrayList<>();
            for (Row row : model.getDragRows()) {
                collectAdvantageModifiers(row, collection);
            }
            DataFile dataFile = ((ListOutline) property).getDataFile();
            for (AdvantageModifier advmod : collection) {
                AdvantageModifier modifier = new AdvantageModifier(dataFile, advmod, false);
                modifier.setEnabled(true);
                list.add(modifier);
            }
        }
        targetAdvantage.setModifiers(list);
        setDragTargetRow(null);
        undo.finish();
        undoList.add(undo);
        new MultipleRowUndo(undoList);
        repaint();
        contentSizeMayHaveChanged();
        model.setDragRows(null);
        if (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) {
            ArrayList<ListRow> process = new ArrayList<>();
            process.add(targetAdvantage);
            EventQueue.invokeLater(new RowPostProcessor(this, process, false));
        }
        return true;
    }

    private void collectAdvantageModifiers(Row row, List<AdvantageModifier> result) {
        if (row instanceof AdvantageModifier) {
            AdvantageModifier advmod = (AdvantageModifier) row;
            if (advmod.canHaveChildren()) {
                for (Row child : advmod.getChildren()) {
                    collectAdvantageModifiers(child, result);
                }
            } else {
                result.add(advmod);
            }
        }
    }

    @Override
    public String getIncrementTitle() {
        return I18n.Text("Increment Level");
    }

    @Override
    public String getDecrementTitle() {
        return I18n.Text("Decrement Level");
    }

    @Override
    public boolean canIncrement() {
        return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeveledRows(false);
    }

    @Override
    public boolean canDecrement() {
        return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasLeveledRows(true);
    }

    private boolean selectionHasLeveledRows(boolean requireLevelAboveZero) {
        for (Advantage advantage : new FilteredIterator<>(getModel().getSelectionAsList(), Advantage.class)) {
            if (!advantage.canHaveChildren() && advantage.isLeveled() && (!requireLevelAboveZero || advantage.hasLevel())) {
                return true;
            }
        }
        return false;
    }

    @SuppressWarnings("unused")
    @Override
    public void increment() {
        ArrayList<RowUndo> undos = new ArrayList<>();
        for (Advantage advantage : new FilteredIterator<>(getModel().getSelectionAsList(), Advantage.class)) {
            if (!advantage.canHaveChildren() && advantage.isLeveled()) {
                RowUndo undo = new RowUndo(advantage);
                advantage.adjustLevel(1);
                if (undo.finish()) {
                    undos.add(undo);
                }
            }
        }
        if (!undos.isEmpty()) {
            repaintSelection();
            new MultipleRowUndo(undos);
        }
    }

    @SuppressWarnings("unused")
    @Override
    public void decrement() {
        ArrayList<RowUndo> undos = new ArrayList<>();
        for (Advantage advantage : new FilteredIterator<>(getModel().getSelectionAsList(), Advantage.class)) {
            if (!advantage.canHaveChildren() && advantage.isLeveled()) {
                RowUndo undo = new RowUndo(advantage);
                advantage.adjustLevel(-1);
                if (undo.finish()) {
                    undos.add(undo);
                }
            }
        }
        if (!undos.isEmpty()) {
            repaintSelection();
            new MultipleRowUndo(undos);
        }
    }
}
