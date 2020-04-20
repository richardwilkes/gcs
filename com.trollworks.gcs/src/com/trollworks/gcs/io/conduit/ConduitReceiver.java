/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.io.conduit;

/** Clients that want to receive messages from a {@link Conduit} must implement this interface. */
public interface ConduitReceiver {
    /**
     * Called when a message is received.
     *
     * @param msg The message.
     */
    void conduitMessageReceived(ConduitMessage msg);

    /**
     * Called to get the filter to apply to incoming message IDs, if any. This method is only called
     * once, when the {@link Conduit} is starting up.
     *
     * @return The string to match IDs against, or {@code null} if any ID is OK.
     */
    String getConduitMessageIDFilter();

    /**
     * Called to get the filter to apply to incoming message users, if any. This method is only
     * called once, when the {@link Conduit} is starting up.
     *
     * @return The string to match users against, or {@code null} if any user is OK.
     */
    String getConduitMessageUserFilter();
}
