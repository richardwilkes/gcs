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

package com.trollworks.gcs.menu.edit;

/**
 * Objects that have their Skill Level change (mainly just skills)
 */
public interface SkillLevelIncrementable {
    /** @return The title to use for the increment skill level menu item. */
    String getIncrementSkillLevelTitle();

    /** @return The title to use for the decrement skill level menu item. */
    String getDecrementSkillLevelTitle();

    /** @return Whether the level can be incremented. */
    boolean canIncrementSkillLevel();

    /** @return Whether the level can be decremented. */
    boolean canDecrementSkillLevel();

    /** Call to increment the level. */
    void incrementSkillLevel();

    /** Call to decrement the level. */
    void decrementSkillLevel();
}
