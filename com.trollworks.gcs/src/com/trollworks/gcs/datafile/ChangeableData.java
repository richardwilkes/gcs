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

package com.trollworks.gcs.datafile;

import java.util.HashSet;
import java.util.Set;

public class ChangeableData implements ChangeNotifier {
    private Set<DataChangeListener> mChangeListeners;

    public final synchronized void addChangeListener(DataChangeListener listener) {
        if (mChangeListeners == null) {
            mChangeListeners = new HashSet<>();
        }
        mChangeListeners.add(listener);
    }

    public final synchronized void removeChangeListener(DataChangeListener listener) {
        mChangeListeners.remove(listener);
        if (mChangeListeners.isEmpty()) {
            mChangeListeners = null;
        }
    }

    @Override
    public void notifyOfChange() {
        DataChangeListener[] listeners;
        synchronized (this) {
            if (mChangeListeners == null) {
                return;
            }
            listeners = mChangeListeners.toArray(new DataChangeListener[0]);
        }
        for (DataChangeListener listener : listeners) {
            listener.dataWasChanged();
        }
    }
}
