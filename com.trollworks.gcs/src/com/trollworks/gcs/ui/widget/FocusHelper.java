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
import java.awt.Container;
import java.awt.EventQueue;
import java.awt.KeyboardFocusManager;
import javax.swing.JComponent;

public final class FocusHelper {
    private Component mTarget;
    private int       mRemaining;
    private Runnable  mRunnable;
    private boolean   mUseTransfer;

    public static void focusOn(Component target, boolean useTransfer) {
        focusOn(target, useTransfer, null);
    }

    public static void focusOn(Component target, boolean useTransfer, Runnable runnable) {
        FocusHelper helper = new FocusHelper(target, useTransfer, runnable);
        helper.tryInitialFocus();
    }

    /**
     * Attempts to scroll the component into view by asking its parent. This is hear to help text
     * fields that normally do local scrolling if scrollRectIntoView() is called.
     */
    public static void scrollIntoView(Component comp) {
        Container parent = comp.getParent();
        if (parent instanceof JComponent p) {
            p.scrollRectToVisible(comp.getBounds());
        }
    }

    private FocusHelper(Component target, boolean useTransfer, Runnable runnable) {
        mTarget = target;
        mRemaining = 5;
        mRunnable = runnable;
        mUseTransfer = useTransfer;
    }

    private void tryInitialFocus() {
        if (--mRemaining > 0 && !isTargetFocusOwner()) {
            if (mUseTransfer) {
                if (mTarget instanceof Container c) {
                    KeyboardFocusManager.getCurrentKeyboardFocusManager().downFocusCycle(c);
                } else {
                    mTarget.transferFocus();
                }
            } else {
                mTarget.requestFocus();
            }
            EventQueue.invokeLater(this::tryInitialFocus);
        } else {
            if (mRunnable != null) {
                mRunnable.run();
            }
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
