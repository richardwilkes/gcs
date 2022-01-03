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

package com.trollworks.gcs.datafile;

import com.trollworks.gcs.utility.units.WeightUnits;

/** Temporary storage for data needed at load time. */
public class LoadState {
    /** The data file version. */
    public int         mDataFileVersion;
    /** Whether the load is happening to restore undo state. */
    public boolean     mForUndo;
    /** The default weight units to use. */
    public WeightUnits mDefWeightUnits;
}
