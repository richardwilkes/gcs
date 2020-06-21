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

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;
import java.util.List;

/** Provides the "Duplicate" command. */
// TODO: This should be reimplemented in terms of the Duplicatable interface
public class DuplicateCommand extends Command {
    /** The action command this command will issue. */
    public static final String           CMD_DUPLICATE = "Duplicate";
    /** The singleton {@link DuplicateCommand}. */
    public static final DuplicateCommand INSTANCE      = new DuplicateCommand();

    private DuplicateCommand() {
        super(I18n.Text("Duplicate"), CMD_DUPLICATE, KeyEvent.VK_U);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof ListOutline) {
            OutlineModel model = ((ListOutline) focus).getModel();
            setEnabled(!model.isLocked() && model.hasSelection());
        } else {
            setEnabled(false);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        ListOutline  outline = (ListOutline) focus;
        OutlineModel model   = outline.getModel();
        if (!model.isLocked() && model.hasSelection()) {
            List<Row>     rows     = new ArrayList<>();
            List<ListRow> topRows  = new ArrayList<>();
            DataFile      dataFile = outline.getDataFile();
            dataFile.startNotify();
            model.setDragRows(model.getSelectionAsList(true).toArray(new Row[0]));
            outline.convertDragRowsToSelf(rows);
            model.setDragRows(null);
            for (Row row : rows) {
                if (row.getDepth() == 0 && row instanceof ListRow) {
                    topRows.add((ListRow) row);
                }
            }
            outline.addRow(topRows.toArray(new ListRow[0]), I18n.Text("Duplicate Rows"), true);
            dataFile.endNotify();
            model.select(topRows, false);
            outline.scrollSelectionIntoView();
        }
    }
}
