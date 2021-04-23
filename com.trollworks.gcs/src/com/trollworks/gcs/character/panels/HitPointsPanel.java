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

/** The character hit points panel. */
public class HitPointsPanel extends HPFPPanel {
    /**
     * Creates a new hit points panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public HitPointsPanel(CharacterSheet sheet) {
        super(I18n.Text("Hit Points"));
        GURPSCharacter gch = sheet.getCharacter();
        createEditableField(sheet, null, FieldFactory.INT7, gch.getCurrentHitPoints(), (c, v) -> c.setHitPointsDamage(-Math.min(((Integer) v).intValue() - gch.getHitPointsAdj(), 0)), "current HP", I18n.Text("Current"), I18n.Text("Current hit points"));
        createEditableField(sheet, Integer.valueOf(gch.getHitPointPoints()), FieldFactory.POSINT6, gch.getHitPointsAdj(), (c, v) -> c.setHitPointsAdj(((Integer) v).intValue()), "maximum HP", I18n.Text("Basic"), I18n.Text("Normal (i.e. unharmed) hit points"));
        boolean dead      = gch.isDead();
        boolean check4    = !dead && gch.isDeathCheck4();
        boolean check3    = !dead && !check4 && gch.isDeathCheck3();
        boolean check2    = !dead && !check4 && !check3 && gch.isDeathCheck2();
        boolean check1    = !dead && !check4 && !check3 && !check2 && gch.isDeathCheck1();
        boolean collapsed = !dead && !check4 && !check3 && !check2 && !check1 && gch.isCollapsedFromHP();
        boolean reeling   = !dead && !check4 && !check3 && !check2 && !check1 && !collapsed && gch.isReeling();
        createField(sheet, gch.getReelingHitPoints(), I18n.Text("Reeling"), I18n.Text("Current hit points at or below this point indicate the character is reeling from the pain, halving move, speed and dodge"), reeling);
        createField(sheet, gch.getUnconsciousChecksHitPoints(), I18n.Text("Collapse"), I18n.Text("<html><body>Current hit points at or below this point indicate the character<br>is on the verge of collapse, causing the character to <b>roll vs. HT</b><br>(at -1 per full multiple of HP below zero) every second to avoid<br>falling unconscious</body></html>"), collapsed);
        String deathCheckTooltip = I18n.Text("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>");
        createField(sheet, gch.getDeathCheck1HitPoints(), I18n.Text("Check #1"), deathCheckTooltip, check1);
        createField(sheet, gch.getDeathCheck2HitPoints(), I18n.Text("Check #2"), deathCheckTooltip, check2);
        createField(sheet, gch.getDeathCheck3HitPoints(), I18n.Text("Check #3"), deathCheckTooltip, check3);
        createField(sheet, gch.getDeathCheck4HitPoints(), I18n.Text("Check #4"), deathCheckTooltip, check4);
        createField(sheet, gch.getDeadHitPoints(), I18n.Text("Dead"), I18n.Text("Current hit points at or below this point cause the character to die"), dead);
    }
}
