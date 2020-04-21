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

package com.trollworks.gcs.menu;

import java.awt.Component;

/**
 * Components that implement this interface may retarget the focus (typically a parent container) to
 * another component when nothing in the normal focus chain would normally be targeted.
 */
public interface RetargetableFocus {
    /**
     * @return The component to use as the 'current' focus, whether it is in the current focus chain
     *         or not.
     */
    Component getRetargetedFocus();
}
