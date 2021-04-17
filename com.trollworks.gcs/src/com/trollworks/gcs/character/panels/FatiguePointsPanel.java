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

package com.trollworks.gcs.character.panels;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.utility.I18n;

/** The character fatigue points panel. */
public class FatiguePointsPanel extends HPFPPanel {
    /**
     * Creates a new fatigue points panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public FatiguePointsPanel(CharacterSheet sheet) {
        super(I18n.Text("Fatigue Points"));
        GURPSCharacter gch = sheet.getCharacter();
        createEditableField(sheet, null, FieldFactory.INT7, gch.getCurrentFatiguePoints(), (c, v) -> c.setFatiguePointsDamage(-Math.min(((Integer) v).intValue() - gch.getFatiguePoints(), 0)), "current FP", I18n.Text("Current"), I18n.Text("Current fatigue points"));
        createEditableField(sheet, Integer.valueOf(gch.getFatiguePointPoints()), FieldFactory.POSINT6, gch.getFatiguePoints(), (c, v) -> c.setFatiguePoints(((Integer) v).intValue()), "maximum FP", I18n.Text("Basic"), I18n.Text("Normal (i.e. fully rested) fatigue points"));
        boolean unconscious = gch.isUnconscious();
        boolean collapsed   = !unconscious && gch.isCollapsedFromFP();
        boolean tired       = !unconscious && !collapsed && gch.isTired();
        createField(sheet, gch.getTiredFatiguePoints(), I18n.Text("Tired"), I18n.Text("Current fatigue points at or below this point indicate the character is very tired, halving move, dodge and strength"), tired);
        createField(sheet, gch.getUnconsciousChecksFatiguePoints(), I18n.Text("Collapse"), I18n.Text("Current fatigue points at or below this point indicate the character is on the verge of collapse, causing the character to roll vs. Will to do anything besides talk or rest"), collapsed);
        createField(sheet, gch.getUnconsciousFatiguePoints(), I18n.Text("Unconscious"), I18n.Text("Current fatigue points at or below this point cause the character to fall unconscious"), unconscious);
    }
}
