/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.menu.edit.Duplicatable;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;

import java.util.ArrayList;
import java.util.List;

class WeaponOutline extends Outline implements Duplicatable {
    private ListRow mOwner;

    WeaponOutline(ListRow owner) {
        super(new OutlineModel());
        getModel().setShowIndent(false);
        mOwner = owner;
        setAllowColumnResize(false);
        setAllowRowDrag(false);
    }

    @Override
    public boolean canDeleteSelection() {
        return getModel().hasSelection();
    }

    @Override
    public void deleteSelection() {
        OutlineModel model = getModel();
        if (model.hasSelection()) {
            model.removeSelection();
            sizeColumnsToFit();
        }
    }

    @Override
    public boolean canDuplicateSelection() {
        return getModel().hasSelection();
    }

    @Override
    public void duplicateSelection() {
        OutlineModel model = getModel();
        if (model.hasSelection()) {
            List<WeaponDisplayRow> added = new ArrayList<>();
            for (Row row : model.getSelectionAsList()) {
                WeaponDisplayRow weapon = new WeaponDisplayRow(((WeaponDisplayRow) row).getWeapon().clone(mOwner));
                model.addRow(weapon);
                added.add(weapon);
            }
            sizeColumnsToFit();
            model.select(added, false);
            revalidate();
            scrollSelectionIntoView();
            requestFocus();
        }
    }
}
