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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;

/** Commonly used icons. */
public final class Icons {
    private Icons() {
    }

    /**
     * @param open {@code true} for the 'open' version.
     * @param roll {@code true} for the highlighted version.
     */
    public static RetinaIcon getDisclosure(boolean open, boolean roll) {
        if (open) {
            return roll ? Images.DISCLOSURE_DOWN_ROLL : Images.DISCLOSURE_DOWN;
        }
        return roll ? Images.DISCLOSURE_RIGHT_ROLL : Images.DISCLOSURE_RIGHT;
    }
}
