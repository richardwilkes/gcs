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

import java.awt.EventQueue;
import java.util.HashSet;
import java.util.Set;

/** Provides synchronization of an outline to its data. */
public class OutlineSyncer implements Runnable {
    private static final OutlineSyncer INSTANCE = new OutlineSyncer();
    private static final Set<Outline>  OUTLINES = new HashSet<>();
    private static       boolean       PENDING;

    /**
     * @param outline The {@link Outline} to add to the set of outlines that need to be
     *                synchronized.
     */
    public static final void add(Outline outline) {
        if (outline != null) {
            synchronized (OUTLINES) {
                OUTLINES.add(outline);
                if (!PENDING) {
                    PENDING = true;
                    EventQueue.invokeLater(INSTANCE);
                }
            }
        }
    }

    /**
     * @param outline The {@link Outline} to remove from the set of outlines that need to be
     *                synchronized.
     */
    public static final void remove(Outline outline) {
        synchronized (OUTLINES) {
            OUTLINES.remove(outline);
        }
    }

    @Override
    public void run() {
        Set<Outline> outlines;
        synchronized (OUTLINES) {
            PENDING = false;
            outlines = new HashSet<>(OUTLINES);
            OUTLINES.clear();
        }
        for (Outline outline : outlines) {
            outline.sizeColumnsToFit();
            outline.repaint();
        }
    }
}
