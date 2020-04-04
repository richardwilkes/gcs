/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.character.names.Namer;
import com.trollworks.gcs.advmod.AdvantageModifierEnabler;
import com.trollworks.toolkit.collections.FilteredList;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.widget.outline.Outline;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/** Helper for causing the row post-processing to occur. */
public class RowPostProcessor implements Runnable {
    private Map<Outline, List<ListRow>> mMap;

    /**
     * Creates a new post processor for name substitution.
     *
     * @param map The map to process.
     */
    public RowPostProcessor(Map<Outline, List<ListRow>> map) {
        mMap = map;
    }

    /**
     * Creates a new post processor for name substitution.
     *
     * @param outline The outline containing the rows.
     * @param list    The list to process.
     */
    public RowPostProcessor(Outline outline, List<ListRow> list) {
        mMap = new HashMap<>();
        mMap.put(outline, list);
    }

    @Override
    public void run() {
        for (Map.Entry<Outline, List<ListRow>> entry : mMap.entrySet()) {
            Outline       outline  = entry.getKey();
            List<ListRow> rows     = entry.getValue();
            boolean       modified = AdvantageModifierEnabler.process(outline, new FilteredList<>(rows, Advantage.class));
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
