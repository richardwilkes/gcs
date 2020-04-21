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

package com.trollworks.gcs.ui.layout;

/** The options for row distribution within a {@link ColumnLayout}. */
public enum RowDistribution {
    /** Gives each component its preferred height. */
    USE_PREFERRED_HEIGHT,
    /** Distributes the height equally among the components. */
    DISTRIBUTE_HEIGHT,
    /** Gives excess height to the last component. */
    GIVE_EXCESS_TO_LAST
}
