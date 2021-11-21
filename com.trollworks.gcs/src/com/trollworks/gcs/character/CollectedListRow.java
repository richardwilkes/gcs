/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.ListRow;

public abstract class CollectedListRow extends ListRow {
    /**
     * Creates a new data row.
     *
     * @param dataFile    The data file to associate it with.
     * @param isContainer Whether or not this row allows children.
     */
    protected CollectedListRow(DataFile dataFile, boolean isContainer) {
        super(dataFile, isContainer);
    }

    /**
     * Creates a clone of an existing data row and associates it with the specified data file.
     *
     * @param dataFile   The data file to associate it with.
     * @param rowToClone The data row to clone.
     */
    protected CollectedListRow(DataFile dataFile, ListRow rowToClone) {
        super(dataFile, rowToClone);
    }

    /**
     * @param outlines The {@link CollectedOutlines} to use.
     * @return The {@link ListOutline} associated with this row.
     */
    public abstract ListOutline getOutlineFromCollectedOutlines(CollectedOutlines outlines);
}
