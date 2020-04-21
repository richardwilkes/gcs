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

import java.awt.Component;
import java.awt.KeyboardFocusManager;

public interface Commitable {
    void attemptCommit();

    static void sendCommitToFocusOwner() {
        KeyboardFocusManager focusManager = KeyboardFocusManager.getCurrentKeyboardFocusManager();
        Component            focus        = focusManager.getPermanentFocusOwner();
        if (focus == null) {
            focus = focusManager.getFocusOwner();
        }
        if (focus instanceof Commitable) {
            ((Commitable) focus).attemptCommit();
        }
    }
}
