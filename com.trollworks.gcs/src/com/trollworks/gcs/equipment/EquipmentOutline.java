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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.edit.Incrementable;
import com.trollworks.gcs.menu.edit.TechLevelIncrementable;
import com.trollworks.gcs.menu.edit.UsesIncrementable;
import com.trollworks.gcs.modifier.EquipmentModifier;
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
import com.trollworks.gcs.utility.text.Numbers;

import java.awt.EventQueue;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.util.ArrayList;
import java.util.List;
import javax.swing.undo.StateEdit;

/** An outline specifically for equipment. */
public class EquipmentOutline extends ListOutline implements Incrementable, UsesIncrementable, TechLevelIncrementable {
    /**
     * Create a new equipment outline.
     *
     * @param dataFile The owning data file.
     * @param model    The {@link OutlineModel} to use.
     */
    public EquipmentOutline(DataFile dataFile, OutlineModel model) {
        super(dataFile, model, Equipment.ID_LIST_CHANGED);
        EquipmentColumn.addColumns(this, dataFile, model.getProperty(EquipmentList.TAG_OTHER_ROOT) == null);
    }

    @Override
    public String getDecrementTitle() {
        return I18n.Text("Decrement Quantity");
    }

    @Override
    public String getIncrementTitle() {
        return I18n.Text("Increment Quantity");
    }

    @Override
    public boolean canDecrement() {
        return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasIncrementableLeafRows(true);
    }

    @Override
    public boolean canIncrement() {
        return (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) && selectionHasIncrementableLeafRows(false);
    }

    private boolean selectionHasIncrementableLeafRows(boolean requireQtyAboveZero) {
        for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
            if (!equipment.canHaveChildren() && (!requireQtyAboveZero || equipment.getQuantity() > 0)) {
                return true;
            }
        }
        return false;
    }

    @SuppressWarnings("unused")
    @Override
    public void decrement() {
        List<RowUndo> undos = new ArrayList<>();
        for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
            if (!equipment.canHaveChildren()) {
                int qty = equipment.getQuantity();
                if (qty > 0) {
                    RowUndo undo = new RowUndo(equipment);
                    equipment.setQuantity(qty - 1);
                    if (undo.finish()) {
                        undos.add(undo);
                    }
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
    public void increment() {
        List<RowUndo> undos = new ArrayList<>();
        for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
            if (!equipment.canHaveChildren()) {
                RowUndo undo = new RowUndo(equipment);
                equipment.setQuantity(equipment.getQuantity() + 1);
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

    @Override
    public boolean canIncrementUses() {
        if (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) {
            for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
                int max = equipment.getMaxUses();
                return max > 0 && equipment.getUses() < max;
            }
        }
        return false;
    }

    @Override
    public boolean canDecrementUses() {
        if (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) {
            for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
                return equipment.getMaxUses() > 0 && equipment.getUses() > 0;
            }
        }
        return false;
    }

    @SuppressWarnings("unused")
    @Override
    public void incrementUses() {
        List<RowUndo> undos = new ArrayList<>();
        for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
            int max  = equipment.getMaxUses();
            int uses = equipment.getUses();
            if (max > 0 && uses < max) {
                RowUndo undo = new RowUndo(equipment);
                equipment.setUses(uses + 1);
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
    public void decrementUses() {
        List<RowUndo> undos = new ArrayList<>();
        for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
            int max  = equipment.getMaxUses();
            int uses = equipment.getUses();
            if (max > 0 && uses > 0) {
                RowUndo undo = new RowUndo(equipment);
                equipment.setUses(uses - 1);
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

    @Override
    public boolean canIncrementTechLevel() {
        return checkTechLevel(0);
    }

    @Override
    public boolean canDecrementTechLevel() {
        return checkTechLevel(1);
    }

    private boolean checkTechLevel(int minLevel) {
        for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
            if (getCurrentTechLevel(equipment) >= minLevel) {
                return true;
            }
        }
        return false;
    }

    private static int getCurrentTechLevel(Equipment equipment) {
        String tl = equipment.getTechLevel();
        if (tl != null) {
            tl = tl.trim();
            if (!tl.isEmpty()) {
                for (int i = tl.length(); --i >= 0; ) {
                    if (!Character.isDigit(tl.charAt(i))) {
                        return -1;
                    }
                }
                return Numbers.extractInteger(tl, -1, 0, Integer.MAX_VALUE, false);
            }
        }
        return -1;
    }

    @SuppressWarnings("unused")
    @Override
    public void incrementTechLevel() {
        List<RowUndo> undos = new ArrayList<>();
        for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
            int tl = getCurrentTechLevel(equipment);
            if (tl >= 0) {
                RowUndo undo = new RowUndo(equipment);
                equipment.setTechLevel(Integer.toString(tl + 1));
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
    public void decrementTechLevel() {
        List<RowUndo> undos = new ArrayList<>();
        for (Equipment equipment : new FilteredIterator<>(getModel().getSelectionAsList(), Equipment.class)) {
            if (!equipment.canHaveChildren()) {
                int tl = getCurrentTechLevel(equipment);
                if (tl >= 1) {
                    RowUndo undo = new RowUndo(equipment);
                    equipment.setTechLevel(Integer.toString(tl - 1));
                    if (undo.finish()) {
                        undos.add(undo);
                    }
                }
            }
        }
        if (!undos.isEmpty()) {
            repaintSelection();
            new MultipleRowUndo(undos);
        }
    }

    @Override
    protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
        return !getModel().isLocked() && rows.length > 0 && (rows[0] instanceof Equipment || rows[0] instanceof EquipmentModifier);
    }

    protected boolean isDropOnRow(Row[] dragRows) {
        return dragRows != null && dragRows.length > 0 && dragRows[0] instanceof EquipmentModifier;
    }

    @Override
    public void convertDragRowsToSelf(List<Row> list) {
        OutlineModel       model              = getModel();
        Row[]              rows               = model.getDragRows();
        boolean            forSheetOrTemplate = mDataFile instanceof GURPSCharacter || mDataFile instanceof Template;
        ArrayList<ListRow> process            = new ArrayList<>();

        if (forSheetOrTemplate) {
            OutlineModel carriedModel;
            OutlineModel uncarriedModel;
            if (mDataFile instanceof GURPSCharacter) {
                GURPSCharacter character = (GURPSCharacter) mDataFile;
                carriedModel = character.getEquipmentRoot();
                uncarriedModel = character.getOtherEquipmentRoot();
            } else {
                Template tmpl = (Template) mDataFile;
                carriedModel = tmpl.getEquipmentModel();
                uncarriedModel = tmpl.getOtherEquipmentModel();
            }
            if (carriedModel != null && uncarriedModel != null) {
                OutlineModel otherModel = rows[0].getOwner();
                OutlineModel selfModel  = getModel();
                if (selfModel != otherModel && (selfModel == carriedModel || selfModel == uncarriedModel) && (otherModel == carriedModel || otherModel == uncarriedModel)) {
                    StateEdit   edit            = new StateEdit(otherModel, I18n.Text("Remove Rows"));
                    ListOutline otherOwningList = (ListOutline) otherModel.getProperty(ListOutline.OWNING_LIST);
                    DataFile    otherDataFile   = otherOwningList.getDataFile();
                    otherDataFile.startNotify();
                    otherModel.removeRows(rows);
                    for (Row row : rows) {
                        Row parent = row.getParent();
                        if (parent != null && !list.contains(parent)) {
                            row.removeFromParent();
                        }
                    }
                    if (otherModel.getRowCount() > 0) {
                        otherOwningList.updateAllRows();
                    }
                    otherDataFile.notify(otherOwningList.getRowSetChangedID(), null);
                    otherDataFile.endNotify();
                    edit.end();
                    postUndo(edit);
                }
            }
        }

        for (Row row : rows) {
            Equipment equipment = new Equipment(mDataFile, (Equipment) row, true);
            model.collectRowsAndSetOwner(list, equipment, false);
            if (forSheetOrTemplate) {
                addRowsToBeProcessed(process, equipment);
            }
        }
        if (forSheetOrTemplate && !process.isEmpty()) {
            EventQueue.invokeLater(new RowPostProcessor(this, process));
        }
    }

    @Override
    protected boolean dropOnRow(DropTargetDropEvent dtde) {
        Row target = getDragTargetRow();
        if (!(target instanceof Equipment)) {
            return false;
        }
        Equipment               targetEquipment = (Equipment) target;
        OutlineModel            model           = getModel();
        ArrayList<RowUndo>      undoList        = new ArrayList<>();
        RowUndo                 undo            = new RowUndo(targetEquipment);
        List<EquipmentModifier> list            = new ArrayList<>(targetEquipment.getModifiers());
        Object                  property        = model.getProperty(ListOutline.OWNING_LIST);
        removeDragHighlight(this);
        if (property instanceof ListOutline) {
            List<EquipmentModifier> collection = new ArrayList<>();
            for (Row row : model.getDragRows()) {
                collectEquipmentModifiers(row, collection);
            }
            DataFile dataFile = ((ListOutline) property).getDataFile();
            for (EquipmentModifier eqpmod : collection) {
                EquipmentModifier modifier = new EquipmentModifier(dataFile, eqpmod, false);
                modifier.setEnabled(true);
                list.add(modifier);
            }
        }
        targetEquipment.setModifiers(list);
        setDragTargetRow(null);
        undo.finish();
        undoList.add(undo);
        new MultipleRowUndo(undoList);
        repaint();
        contentSizeMayHaveChanged();
        model.setDragRows(null);
        if (mDataFile instanceof GURPSCharacter || mDataFile instanceof Template) {
            ArrayList<ListRow> process = new ArrayList<>();
            process.add(targetEquipment);
            EventQueue.invokeLater(new RowPostProcessor(this, process, false));
        }
        return true;
    }

    private void collectEquipmentModifiers(Row row, List<EquipmentModifier> result) {
        if (row instanceof EquipmentModifier) {
            EquipmentModifier eqpmod = (EquipmentModifier) row;
            if (eqpmod.canHaveChildren()) {
                for (Row child : eqpmod.getChildren()) {
                    collectEquipmentModifiers(child, result);
                }
            } else {
                result.add(eqpmod);
            }
        }
    }
}
