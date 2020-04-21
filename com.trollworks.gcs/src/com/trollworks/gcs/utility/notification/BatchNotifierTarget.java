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

package com.trollworks.gcs.utility.notification;

/**
 * Objects that want to be the target of notifications from a {@link Notifier} and want to be
 * notified when a batch change occurs must implement this interface.
 */
public interface BatchNotifierTarget extends NotifierTarget {
    /**
     * Called when a series of notifications is about to be broadcast. The {@link
     * BatchNotifierTarget} may or may not have intervening calls to {@link
     * NotifierTarget#handleNotification(Object, String, Object)} made to it.
     */
    void enterBatchMode();

    /** Called after a series of notifications was broadcast. */
    void leaveBatchMode();
}
