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

import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Img;

public interface IconAccessor {
    /**
     * @param row The {@link TreeRow} to operate on.
     * @return The {@link Img} for the field, or {@code null}.
     */
    RetinaIcon getIcon(TreeRow row);
}
