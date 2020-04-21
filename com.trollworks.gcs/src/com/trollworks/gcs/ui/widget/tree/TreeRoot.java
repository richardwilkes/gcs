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

import com.trollworks.gcs.utility.notification.Notifier;

/** Provides the root of a tree of {@link TreeRow}s. */
public class TreeRoot extends TreeContainerRow {
    private Notifier mNotifier;

    /**
     * Creates a new {@link TreeRoot}.
     *
     * @param notifier The {@link Notifier} to use. Must not be {@code null}.
     */
    public TreeRoot(Notifier notifier) {
        mNotifier = notifier;
    }

    /** @return The {@link Notifier} being used. */
    public Notifier getNotifier() {
        return mNotifier;
    }
}
