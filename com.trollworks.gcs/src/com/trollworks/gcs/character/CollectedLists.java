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

package com.trollworks.gcs.character;

import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.notes.NoteOutline;
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.ui.scale.ScaleRoot;

/**
 * CollectedLists defines the methods necessary for use of a DataFile with a
 * CollectedListsDockable.
 */
public interface CollectedLists extends ScaleRoot {
    /** @return The outline containing the Advantages, Disadvantages & Quirks. */
    AdvantageOutline getAdvantageOutline();

    /** @return The outline containing the skills. */
    SkillOutline getSkillOutline();

    /** @return The outline containing the spells. */
    SpellOutline getSpellOutline();

    /** @return The outline containing the carried equipment. */
    EquipmentOutline getEquipmentOutline();

    /** @return The outline containing the other equipment. */
    EquipmentOutline getOtherEquipmentOutline();

    /** @return The outline containing the notes. */
    NoteOutline getNoteOutline();
}
