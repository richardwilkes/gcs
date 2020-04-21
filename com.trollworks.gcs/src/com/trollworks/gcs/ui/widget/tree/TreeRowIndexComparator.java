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

package com.trollworks.gcs.ui.widget.tree;

import java.util.Comparator;

/** Compares {@link TreeRow} indexes. */
public class TreeRowIndexComparator implements Comparator<TreeRow> {
    /** The one and only instance. */
    public static final TreeRowIndexComparator INSTANCE = new TreeRowIndexComparator();

    private TreeRowIndexComparator() {
        // Here just to prevent multiple copies.
    }

    @Override
    public int compare(TreeRow o1, TreeRow o2) {
        return Integer.compare(o1.getIndex(), o2.getIndex());
    }
}
