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
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageHeader;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.ThemeColor;
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
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComponent;
import javax.swing.SwingConstants;

/** The character encumbrance panel. */
public class EncumbrancePanel extends DropPanel implements NotifierTarget {
    private CharacterSheet   mSheet;
    private PageLabel[]      mMarkers;
    private List<JComponent> mFieldsForColorUpdate;

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
        addHorizontalBackground(bulletHeader, ThemeColor.HEADER);
        PageHeader header = createHeader(this, I18n.Text("Level"), encLevelTooltip);
        addVerticalBackground(createDivider(), ThemeColor.DIVIDER);
        String maxLoadTooltip = I18n.Text("The maximum load a character can carry and still remain within a specific encumbrance level");
        createHeader(this, I18n.Text("Max Load"), maxLoadTooltip);
        addVerticalBackground(createDivider(), ThemeColor.DIVIDER);
        String moveTooltip = I18n.Text("The character's ground movement rate for a specific encumbrance level");
        createHeader(this, I18n.Text("Move"), moveTooltip);
        addVerticalBackground(createDivider(), ThemeColor.DIVIDER);
        String dodgeTooltip = I18n.Text("The character's dodge for a specific encumbrance level");
        createHeader(this, I18n.Text("Dodge"), dodgeTooltip);
        mFieldsForColorUpdate = new ArrayList<>();
        GURPSCharacter character = mSheet.getCharacter();
        Encumbrance    current   = character.getEncumbranceLevel(false);
        boolean        band      = false;
        for (Encumbrance encumbrance : encumbranceValues) {
            boolean warn;
            Color   textColor;
            if (current == encumbrance) {
                warn = character.isCarryingGreaterThanMaxLoad(false);
                textColor = warn ? ThemeColor.ON_WARN : ThemeColor.ON_CURRENT;
            } else {
                warn = false;
                textColor = ThemeColor.ON_PAGE;
            }
            int index = encumbrance.ordinal();
            mMarkers[index] = new PageLabel(encumbrance == current ? "•" : "", textColor, header);
            UIUtilities.setOnlySize(mMarkers[index], prefSize);
            addField(encumbrance, mMarkers[index]);
            if (current == encumbrance) {
                addHorizontalBackground(mMarkers[index], warn ? ThemeColor.WARN : ThemeColor.CURRENT);
            } else if (band) {
                addHorizontalBackground(mMarkers[index], ThemeColor.BANDING);
            }
            band = !band;
            addField(encumbrance, new PageLabel(MessageFormat.format("{0} {1}", Numbers.format(-encumbrance.getEncumbrancePenalty()), encumbrance), textColor, header));
            createDivider();
            addField(encumbrance, new PageField(mSheet, GURPSCharacter.MAXIMUM_CARRY_PREFIX + index, SwingConstants.RIGHT, false, maxLoadTooltip, textColor));
            createDivider();
            addField(encumbrance, new PageField(mSheet, GURPSCharacter.MOVE_PREFIX + index, SwingConstants.RIGHT, false, moveTooltip, textColor));
            createDivider();
            addField(encumbrance, new PageField(mSheet, GURPSCharacter.DODGE_PREFIX + index, SwingConstants.RIGHT, false, dodgeTooltip, textColor));
        }
        character.addTarget(this, GURPSCharacter.ID_CARRIED_WEIGHT, GURPSCharacter.ID_BASIC_LIFT, GURPSCharacter.ID_CURRENT_HP, GURPSCharacter.ID_CURRENT_FP);
    }

    private void addField(Encumbrance enc, JComponent field) {
        field.putClientProperty(Encumbrance.class, enc);
        add(field);
        mFieldsForColorUpdate.add(field);
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
        Encumbrance current = character.getEncumbranceLevel(false);
        boolean band = false;
        for (Encumbrance encumbrance : Encumbrance.values()) {
            int index = encumbrance.ordinal();
            if (encumbrance == current) {
                addHorizontalBackground(mMarkers[index], character.isCarryingGreaterThanMaxLoad(false) ? ThemeColor.WARN : ThemeColor.CURRENT);
            } else if (band) {
                addHorizontalBackground(mMarkers[index], ThemeColor.BANDING);
            } else {
                removeHorizontalBackground(mMarkers[index]);
            }
            mMarkers[index].setText(encumbrance == current ? "•" : "");
            band = !band;
        }
        for (JComponent comp : mFieldsForColorUpdate) {
            Color       textColor;
            Encumbrance enc = (Encumbrance) comp.getClientProperty(Encumbrance.class);
            if (current == enc) {
                textColor = character.isCarryingGreaterThanMaxLoad(false) ? ThemeColor.ON_WARN : ThemeColor.ON_CURRENT;
            } else {
                textColor = ThemeColor.ON_PAGE;
            }
            if (comp instanceof PageField) {
                ((PageField) comp).setDisabledTextColor(textColor);
            } else {
                comp.setForeground(textColor);
            }
        }
        revalidate();
        repaint();
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }
}
