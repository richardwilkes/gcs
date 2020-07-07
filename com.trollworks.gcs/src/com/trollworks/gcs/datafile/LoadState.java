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

package com.trollworks.gcs.datafile;

import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.util.HashSet;

/** Temporary storage for data needed at load time. */
public class LoadState {
    /** The attribute used for versioning. */
    public static final String             ATTRIBUTE_VERSION = "version";
    /** The data file version. */
    public              int                mDataFileVersion;
    /** The data item version. Used for individual items within a file. */
    public              int                mDataItemVersion;
    /** Whether the load is happening to restore undo state. */
    public              boolean            mForUndo;
    /** Used to move uncarried equipment out of the carried equipment list for old sheets. */
    public              HashSet<Equipment> mUncarriedEquipment;
    /** The default weight units to use. */
    public              WeightUnits        mDefWeightUnits;
}
