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
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.page.PagePoints;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.notification.NotifierTarget;

import java.awt.Color;
import java.awt.Component;
import javax.swing.SwingConstants;

/** The character fatigue points panel. */
public class FatiguePointsPanel extends DropPanel implements NotifierTarget {
    private static final Color          CURRENT_THRESHOLD_COLOR = new Color(255, 224, 224);
    private              CharacterSheet mSheet;
    private              PageField      mTiredField;
    private              PageField      mCollapsedField;
    private              PageField      mUnconsciousField;
    private              State          mState;

    enum State {
        UNAFFECTED, TIRED, COLLAPSED, UNCONSCIOUS
    }

    /**
     * Creates a new hit points panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public FatiguePointsPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(2, 0).setAlignment(PrecisionLayoutAlignment.FILL, PrecisionLayoutAlignment.FILL), I18n.Text("Fatigue Points"));
        mSheet = sheet;
        mState = State.UNAFFECTED;
        addLabelAndField(sheet, GURPSCharacter.ID_CURRENT_FP, I18n.Text("Current"), I18n.Text("Current fatigue points"), true);
        addLabelAndField(sheet, GURPSCharacter.ID_FATIGUE_POINTS, I18n.Text("Basic"), I18n.Text("Normal (i.e. fully rested) fatigue points"), true);
        mTiredField = addLabelAndField(sheet, GURPSCharacter.ID_TIRED_FATIGUE_POINTS, I18n.Text("Tired"), I18n.Text("Current fatigue points at or below this point indicate the character is very tired, halving move, dodge and strength"), false);
        mCollapsedField = addLabelAndField(sheet, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, I18n.Text("Collapse"), I18n.Text("Current fatigue points at or below this point indicate the character is on the verge of collapse, causing the character to roll vs. Will to do anything besides talk or rest"), false);
        mUnconsciousField = addLabelAndField(sheet, GURPSCharacter.ID_UNCONSCIOUS_FATIGUE_POINTS, I18n.Text("Unconscious"), I18n.Text("Current fatigue points at or below this point cause the character to fall unconscious"), false);
        adjustBackgrounds();
        sheet.getCharacter().addTarget(this, GURPSCharacter.ID_CURRENT_FP);
    }

    private PageField addLabelAndField(CharacterSheet sheet, String key, String title, String tooltip, boolean enabled) {
        add(new PagePoints(sheet, key), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        PageField field = new PageField(sheet, key, SwingConstants.RIGHT, enabled, tooltip);
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, field));
        return field;
    }

    private void adjustBackgrounds() {
        GURPSCharacter character = mSheet.getCharacter();
        if (character.isUnconscious()) {
            if (mState == State.UNCONSCIOUS) {
                return;
            }
            mState = State.UNCONSCIOUS;
        } else if (character.isCollapsedFromFP()) {
            if (mState == State.COLLAPSED) {
                return;
            }
            mState = State.COLLAPSED;
        } else if (character.isTired()) {
            if (mState == State.TIRED) {
                return;
            }
            mState = State.TIRED;
        } else if (mState == State.UNAFFECTED) {
            return;
        } else {
            mState = State.UNAFFECTED;
        }
        applyBackground(mTiredField, mState == State.TIRED);
        applyBackground(mCollapsedField, mState == State.COLLAPSED);
        applyBackground(mUnconsciousField, mState == State.UNCONSCIOUS);
    }

    private void applyBackground(Component field, boolean add) {
        if (add) {
            addHorizontalBackground(field, CURRENT_THRESHOLD_COLOR);
        } else {
            removeHorizontalBackground(field);
        }
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        adjustBackgrounds();
    }
}
