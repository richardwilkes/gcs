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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateSheet;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowUndo;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.undo.MultipleUndo;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.util.ArrayList;
import java.util.List;
import javax.swing.undo.StateEdit;

/** Provides the "Toggle State" command. */
public class MoveEquipmentCommand extends Command {
    public static final String  CMD_TO_CARRIED_EQUIPMENT = "ToCarriedEquipment";
    public static final String  CMD_TO_OTHER_EQUIPMENT   = "ToOtherEquipment";
    private             boolean mToCarried;

    public MoveEquipmentCommand(boolean toCarried) {
        super(toCarried ? I18n.Text("Move to Carried Equipment") : I18n.Text("Move to Other Equipment"), toCarried ? CMD_TO_CARRIED_EQUIPMENT : CMD_TO_OTHER_EQUIPMENT);
        mToCarried = toCarried;
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof EquipmentOutline) {
            ListOutline  outline  = (ListOutline) focus;
            DataFile     dataFile = outline.getDataFile();
            OutlineModel model    = outline.getModel();
            boolean      isOther  = model.getProperty(EquipmentList.TAG_OTHER_ROOT) != null;
            setEnabled((dataFile instanceof GURPSCharacter || dataFile instanceof Template) && isOther == mToCarried && model.hasSelection());
        } else {
            setEnabled(false);
        }
    }

    @SuppressWarnings("unused")
    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof EquipmentOutline) {
            EquipmentOutline outline  = (EquipmentOutline) focus;
            EquipmentOutline other    = null;
            DataFile         dataFile = outline.getDataFile();
            OutlineModel     model    = outline.getModel();
            boolean          isOther  = model.getProperty(EquipmentList.TAG_OTHER_ROOT) != null;
            if ((dataFile instanceof GURPSCharacter || dataFile instanceof Template) && isOther == mToCarried && model.hasSelection()) {
                CharacterSheet csheet = UIUtilities.getAncestorOfType(outline, CharacterSheet.class);
                if (csheet != null) {
                    other = mToCarried ? csheet.getEquipmentOutline() : csheet.getOtherEquipmentOutline();
                } else {
                    TemplateSheet tsheet = UIUtilities.getAncestorOfType(outline, TemplateSheet.class);
                    if (tsheet != null) {
                        other = mToCarried ? tsheet.getEquipmentOutline() : tsheet.getOtherEquipmentOutline();
                    }
                }
                if (other != null) {
                    MultipleUndo undo = new MultipleUndo(getTitle());
                    outline.postUndo(undo);
                    List<RowUndo> undos   = new ArrayList<>();
                    List<Row>     rows    = new ArrayList<>();
                    List<ListRow> topRows = new ArrayList<>();
                    OutlineModel  target  = other.getModel();
                    StateEdit     edit1   = new StateEdit(model, getTitle());
                    StateEdit     edit2   = new StateEdit(target, getTitle());
                    dataFile.startNotify();
                    target.setDragRows(outline.getModel().getSelectionAsList(true).toArray(new Row[0]));
                    other.convertDragRowsToSelf(rows);
                    target.setDragRows(null);
                    for (Row row : rows) {
                        if (row.getDepth() == 0 && row instanceof ListRow) {
                            topRows.add((ListRow) row);
                        }
                    }
                    other.addRow(topRows.toArray(new ListRow[0]), getTitle(), true);
                    dataFile.endNotify();
                    edit1.end();
                    edit2.end();
                    undo.addEdit(edit1);
                    undo.addEdit(edit2);
                    undo.end();
                    model.select(topRows, false);
                    other.scrollSelectionIntoView();
                }
            }
        }
    }
}
