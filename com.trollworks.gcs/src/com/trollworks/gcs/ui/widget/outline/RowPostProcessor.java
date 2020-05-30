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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.character.names.Namer;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.modifier.AdvantageModifierEnabler;
import com.trollworks.gcs.modifier.EquipmentModifierEnabler;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.FilteredList;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/** Helper for causing the row post-processing to occur. */
public class RowPostProcessor implements Runnable {
    private Map<Outline, List<ListRow>> mMap;
    private boolean                     mRunModifierEnabler;

    /**
     * Creates a new post processor for name substitution.
     *
     * @param map The map to process.
     */
    public RowPostProcessor(Map<Outline, List<ListRow>> map) {
        mMap = map;
        mRunModifierEnabler = true;
    }

    /**
     * Creates a new post processor for name substitution.
     *
     * @param outline The outline containing the rows.
     * @param list    The list to process.
     */
    public RowPostProcessor(Outline outline, List<ListRow> list) {
        this(outline, list, true);
    }

    /**
     * Creates a new post processor for name substitution.
     *
     * @param outline The outline containing the rows.
     * @param list    The list to process.
     */
    public RowPostProcessor(Outline outline, List<ListRow> list, boolean runModifierEnabler) {
        mMap = new HashMap<>();
        mMap.put(outline, list);
        mRunModifierEnabler = runModifierEnabler;
    }

    @Override
    public void run() {
        for (Map.Entry<Outline, List<ListRow>> entry : mMap.entrySet()) {
            Outline       outline  = entry.getKey();
            List<ListRow> rows     = entry.getValue();
            boolean       modified = false;
            if (mRunModifierEnabler) {
                modified = AdvantageModifierEnabler.process(outline, new FilteredList<>(rows, Advantage.class));
                modified |= EquipmentModifierEnabler.process(outline, new FilteredList<>(rows, Equipment.class));
            }
            modified |= Namer.name(outline, rows);
            if (modified) {
                outline.updateRowHeights(rows);
                outline.repaint();
                SheetDockable dockable = UIUtilities.getAncestorOfType(outline, SheetDockable.class);
                if (dockable != null) {
                    dockable.notifyOfPrereqOrFeatureModification();
                }
            }
        }
    }
}
