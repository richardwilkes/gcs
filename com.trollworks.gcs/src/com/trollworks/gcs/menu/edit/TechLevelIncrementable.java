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

package com.trollworks.gcs.menu.edit;

/**
 * Objects that can have their Tech Level incremented/decremented should implement this interface.
 */
public interface TechLevelIncrementable {
    /** @return Whether the data can be incremented. */
    boolean canIncrementTechLevel();

    /** @return Whether the data can be decremented. */
    boolean canDecrementTechLevel();

    /** Call to increment the data. */
    void incrementTechLevel();

    /** Call to decrement the data. */
    void decrementTechLevel();
}
