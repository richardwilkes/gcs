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

import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageHeader;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.Numbers;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;
import javax.swing.SwingConstants;

/** The character encumbrance panel. */
public class EncumbrancePanel extends DropPanel implements NotifierTarget {
    private static final Color          CURRENT_ENCUMBRANCE_COLOR            = new Color(252, 242, 196);
    private static final Color          CURRENT_ENCUMBRANCE_OVERLOADED_COLOR = new Color(255, 224, 224);
    private              CharacterSheet mSheet;
    private              PageLabel[]    mMarkers;

    /**
     * Creates a new encumbrance panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public EncumbrancePanel(CharacterSheet sheet) {
        super(new ColumnLayout(8, 2, 0), I18n.Text("Encumbrance, Move & Dodge"), true);
        mSheet = sheet;
        Encumbrance[] encumbranceValues = Encumbrance.values();
        Dimension     prefSize          = new PageLabel("•", null).getPreferredSize();
        mMarkers = new PageLabel[encumbranceValues.length];
        String     encLevelTooltip = I18n.Text("The encumbrance level");
        PageHeader bulletHeader    = createHeader(this, "", encLevelTooltip);
        UIUtilities.setOnlySize(bulletHeader, prefSize);
        addHorizontalBackground(bulletHeader, Color.black);
        PageHeader header = createHeader(this, I18n.Text("Level"), encLevelTooltip);
        addVerticalBackground(createDivider(), Color.black);
        String maxLoadTooltip = I18n.Text("The maximum load a character can carry and still remain within a specific encumbrance level");
        createHeader(this, I18n.Text("Max Load"), maxLoadTooltip);
        addVerticalBackground(createDivider(), Color.black);
        String moveTooltip = I18n.Text("The character's ground movement rate for a specific encumbrance level");
        createHeader(this, I18n.Text("Move"), moveTooltip);
        addVerticalBackground(createDivider(), Color.black);
        String dodgeTooltip = I18n.Text("The character's dodge for a specific encumbrance level");
        createHeader(this, I18n.Text("Dodge"), dodgeTooltip);
        GURPSCharacter character = mSheet.getCharacter();
        Encumbrance    current   = character.getEncumbranceLevel();
        for (Encumbrance encumbrance : encumbranceValues) {
            int index = encumbrance.ordinal();
            mMarkers[index] = new PageLabel(encumbrance == current ? "•" : "", header);
            UIUtilities.setOnlySize(mMarkers[index], prefSize);
            add(mMarkers[index]);
            if (current == encumbrance) {
                addHorizontalBackground(mMarkers[index], character.isCarryingGreaterThanMaxLoad() ? CURRENT_ENCUMBRANCE_OVERLOADED_COLOR : CURRENT_ENCUMBRANCE_COLOR);
            }
            add(new PageLabel(MessageFormat.format("{0} {1}", Numbers.format(-encumbrance.getEncumbrancePenalty()), encumbrance), header));
            createDivider();
            createDisabledField(this, mSheet, GURPSCharacter.MAXIMUM_CARRY_PREFIX + index, maxLoadTooltip, SwingConstants.RIGHT);
            createDivider();
            createDisabledField(this, mSheet, GURPSCharacter.MOVE_PREFIX + index, moveTooltip, SwingConstants.RIGHT);
            createDivider();
            createDisabledField(this, mSheet, GURPSCharacter.DODGE_PREFIX + index, dodgeTooltip, SwingConstants.RIGHT);
        }
        character.addTarget(this, GURPSCharacter.ID_CARRIED_WEIGHT, GURPSCharacter.ID_BASIC_LIFT, GURPSCharacter.ID_CURRENT_HP, GURPSCharacter.ID_CURRENT_FP);
    }

    private Container createDivider() {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        return panel;
    }

    @Override
    public void handleNotification(Object producer, String type, Object data) {
        GURPSCharacter character = mSheet.getCharacter();
        if (GURPSCharacter.ID_CURRENT_HP.equals(type) || GURPSCharacter.ID_CURRENT_FP.equals(type)) {
            character.notifyMoveAndDodge();
            character.notifyBasicLift();
        }
        Encumbrance current = character.getEncumbranceLevel();
        for (Encumbrance encumbrance : Encumbrance.values()) {
            int index = encumbrance.ordinal();
            if (encumbrance == current) {
                addHorizontalBackground(mMarkers[index], character.isCarryingGreaterThanMaxLoad() ? CURRENT_ENCUMBRANCE_OVERLOADED_COLOR : CURRENT_ENCUMBRANCE_COLOR);
            } else {
                removeHorizontalBackground(mMarkers[index]);
            }
            mMarkers[index].setText(encumbrance == current ? "•" : "");
        }
        revalidate();
        repaint();
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }
}
