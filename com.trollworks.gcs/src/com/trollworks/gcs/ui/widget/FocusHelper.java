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

package com.trollworks.gcs.ui.widget;

import java.awt.Component;
import java.awt.EventQueue;
import java.awt.KeyboardFocusManager;

public final class FocusHelper {
    private Component mTarget;
    private int       mRemaining;
    private Runnable  mRunnable;

    public static void focusOn(Component target) {
        focusOn(target, null);
    }

    public static void focusOn(Component target, Runnable runnable) {
        FocusHelper helper = new FocusHelper(target, runnable);
        helper.tryInitialFocus();
    }

    private FocusHelper(Component target, Runnable runnable) {
        mTarget = target;
        mRemaining = 5;
        mRunnable = runnable;
    }

    private void tryInitialFocus() {
        if (--mRemaining > 0 && !isTargetFocusOwner()) {
            mTarget.requestFocus();
            EventQueue.invokeLater(this::tryInitialFocus);
        } else if (mRunnable != null) {
            mRunnable.run();
        }
    }

    private boolean isTargetFocusOwner() {
        Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();
        while (focus != null && focus != mTarget) {
            focus = focus.getParent();
        }
        return focus == mTarget;
    }
}
