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

package com.trollworks.gcs.ui;

import java.awt.Desktop;
import java.awt.desktop.AppForegroundEvent;
import java.awt.desktop.AppForegroundListener;
import java.util.ArrayList;
import java.util.List;

public final class SystemEventHandler implements AppForegroundListener {
    public static final SystemEventHandler          INSTANCE = new SystemEventHandler();
    private             List<AppForegroundListener> mListeners;

    private SystemEventHandler() {
        mListeners = new ArrayList<>();
        if (Desktop.isDesktopSupported()) {
            Desktop desktop = Desktop.getDesktop();
            if (desktop.isSupported(Desktop.Action.APP_EVENT_FOREGROUND)) {
                desktop.addAppEventListener(this);
            }
        }
    }

    public synchronized void addAppForegroundListener(AppForegroundListener listener) {
        mListeners.add(listener);
    }

    public synchronized void removeAppForegroundListener(AppForegroundListener listener) {
        mListeners.remove(listener);
    }

    @Override
    public void appRaisedToForeground(AppForegroundEvent event) {
        List<AppForegroundListener> listeners;
        synchronized (this) {
            listeners = new ArrayList<>(mListeners);
        }
        for (AppForegroundListener listener : listeners) {
            listener.appRaisedToForeground(event);
        }
    }

    @Override
    public void appMovedToBackground(AppForegroundEvent event) {
        List<AppForegroundListener> listeners;
        synchronized (this) {
            listeners = new ArrayList<>(mListeners);
        }
        for (AppForegroundListener listener : listeners) {
            listener.appMovedToBackground(event);
        }
    }
}
