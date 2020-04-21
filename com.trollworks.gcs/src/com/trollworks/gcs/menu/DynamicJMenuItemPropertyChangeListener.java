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

import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.lang.ref.ReferenceQueue;
import java.lang.ref.WeakReference;
import javax.swing.Action;
import javax.swing.JMenuItem;
import javax.swing.KeyStroke;

class DynamicJMenuItemPropertyChangeListener implements PropertyChangeListener {
    private static ReferenceQueue<JMenuItem> QUEUE = new ReferenceQueue<>();
    private        WeakReference<JMenuItem>  mTarget;
    private        Action                    mAction;
    private        PropertyChangeListener    mChainedListener;

    DynamicJMenuItemPropertyChangeListener(JMenuItem menuItem, Action action, PropertyChangeListener chainedListener) {
        OwnedWeakReference ref;
        while ((ref = (OwnedWeakReference) QUEUE.poll()) != null) {
            DynamicJMenuItemPropertyChangeListener old       = (DynamicJMenuItemPropertyChangeListener) ref.getOwner();
            Action                                 oldAction = old.mAction;
            if (oldAction != null) {
                oldAction.removePropertyChangeListener(old);
            }
        }
        mTarget = new OwnedWeakReference(menuItem, QUEUE, this);
        mAction = action;
        mChainedListener = chainedListener;
    }

    @Override
    public void propertyChange(PropertyChangeEvent event) {
        JMenuItem mi = mTarget.get();
        if (mi == null) {
            Action action = (Action) event.getSource();
            action.removePropertyChangeListener(this);
        } else {
            if (event.getPropertyName().equals(Action.ACCELERATOR_KEY)) {
                mi.setAccelerator((KeyStroke) event.getNewValue());
                mi.invalidate();
                mi.repaint();
            }
        }
        mChainedListener.propertyChange(event);
    }

    private static class OwnedWeakReference extends WeakReference<JMenuItem> {
        private DynamicJMenuItemPropertyChangeListener mOwner;

        OwnedWeakReference(JMenuItem target, ReferenceQueue<JMenuItem> queue, DynamicJMenuItemPropertyChangeListener owner) {
            super(target, queue);
            mOwner = owner;
        }

        public Object getOwner() {
            return mOwner;
        }
    }
}
