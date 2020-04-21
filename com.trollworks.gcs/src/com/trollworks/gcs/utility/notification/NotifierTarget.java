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
 * Objects that want to be the target of notifications from a {@link Notifier} must implement this
 * interface.
 */
public interface NotifierTarget {
    /** @return The relative notification priority. Higher gets delivered first. */
    int getNotificationPriority();

    /**
     * Called when a notification is delivered.
     *
     * @param producer The producer of the notification.
     * @param name     The notification name.
     * @param data     Extra data specific to the notification.
     */
    void handleNotification(Object producer, String name, Object data);
}
