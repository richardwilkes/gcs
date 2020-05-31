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

/** The character hit points panel. */
public class HitPointsPanel extends DropPanel implements NotifierTarget {
    private static final Color          CURRENT_THRESHOLD_COLOR = new Color(255, 224, 224);
    private              CharacterSheet mSheet;
    private              PageField      mReelingField;
    private              PageField      mCollapsedField;
    private              PageField      mCheck1Field;
    private              PageField      mCheck2Field;
    private              PageField      mCheck3Field;
    private              PageField      mCheck4Field;
    private              PageField      mDeadField;
    private              State          mState;

    enum State {
        UNAFFECTED, REELING, COLLAPSED, CHECK1, CHECK2, CHECK3, CHECK4, DEAD
    }

    /**
     * Creates a new hit points panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public HitPointsPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(2, 0).setAlignment(PrecisionLayoutAlignment.FILL, PrecisionLayoutAlignment.FILL), I18n.Text("Hit Points"));
        mSheet = sheet;
        mState = State.UNAFFECTED;
        addLabelAndField(sheet, GURPSCharacter.ID_CURRENT_HP, I18n.Text("Current"), I18n.Text("Current hit points"), true);
        addLabelAndField(sheet, GURPSCharacter.ID_HIT_POINTS, I18n.Text("Basic"), I18n.Text("Normal (i.e. unharmed) hit points"), true);
        mReelingField = addLabelAndField(sheet, GURPSCharacter.ID_REELING_HIT_POINTS, I18n.Text("Reeling"), I18n.Text("Current hit points at or below this point indicate the character is reeling from the pain, halving move, speed and dodge"), false);
        mCollapsedField = addLabelAndField(sheet, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_HIT_POINTS, I18n.Text("Collapse"), I18n.Text("<html><body>Current hit points at or below this point indicate the character<br>is on the verge of collapse, causing the character to <b>roll vs. HT</b><br>(at -1 per full multiple of HP below zero) every second to avoid<br>falling unconscious</body></html>"), false);
        String deathCheckTooltip = I18n.Text("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>");
        mCheck1Field = addLabelAndField(sheet, GURPSCharacter.ID_DEATH_CHECK_1_HIT_POINTS, I18n.Text("Check #1"), deathCheckTooltip, false);
        mCheck2Field = addLabelAndField(sheet, GURPSCharacter.ID_DEATH_CHECK_2_HIT_POINTS, I18n.Text("Check #2"), deathCheckTooltip, false);
        mCheck3Field = addLabelAndField(sheet, GURPSCharacter.ID_DEATH_CHECK_3_HIT_POINTS, I18n.Text("Check #3"), deathCheckTooltip, false);
        mCheck4Field = addLabelAndField(sheet, GURPSCharacter.ID_DEATH_CHECK_4_HIT_POINTS, I18n.Text("Check #4"), deathCheckTooltip, false);
        mDeadField = addLabelAndField(sheet, GURPSCharacter.ID_DEAD_HIT_POINTS, I18n.Text("Dead"), I18n.Text("Current hit points at or below this point cause the character to die"), false);
        adjustBackgrounds();
        sheet.getCharacter().addTarget(this, GURPSCharacter.ID_CURRENT_HP);
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
        if (character.isDead()) {
            if (mState == State.DEAD) {
                return;
            }
            mState = State.DEAD;
        } else if (character.isDeathCheck4()) {
            if (mState == State.CHECK4) {
                return;
            }
            mState = State.CHECK4;
        } else if (character.isDeathCheck3()) {
            if (mState == State.CHECK3) {
                return;
            }
            mState = State.CHECK3;
        } else if (character.isDeathCheck2()) {
            if (mState == State.CHECK2) {
                return;
            }
            mState = State.CHECK2;
        } else if (character.isDeathCheck1()) {
            if (mState == State.CHECK1) {
                return;
            }
            mState = State.CHECK1;
        } else if (character.isCollapsedFromHP()) {
            if (mState == State.COLLAPSED) {
                return;
            }
            mState = State.COLLAPSED;
        } else if (character.isReeling()) {
            if (mState == State.REELING) {
                return;
            }
            mState = State.REELING;
        } else if (mState == State.UNAFFECTED) {
            return;
        } else {
            mState = State.UNAFFECTED;
        }
        applyBackground(mReelingField, mState == State.REELING);
        applyBackground(mCollapsedField, mState == State.COLLAPSED);
        applyBackground(mCheck1Field, mState == State.CHECK1);
        applyBackground(mCheck2Field, mState == State.CHECK2);
        applyBackground(mCheck3Field, mState == State.CHECK3);
        applyBackground(mCheck4Field, mState == State.CHECK4);
        applyBackground(mDeadField, mState == State.DEAD);
        repaint();
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
