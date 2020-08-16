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

import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.utility.I18n;

/** The character fatigue points panel. */
public class FatiguePointsPanel extends HPFPPanel {
    private PageField mTiredField;
    private PageField mCollapsedField;
    private PageField mUnconsciousField;
    private State     mState;

    enum State {
        UNAFFECTED,
        TIRED,
        COLLAPSED,
        UNCONSCIOUS
    }

    /**
     * Creates a new fatigue points panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public FatiguePointsPanel(CharacterSheet sheet) {
        super(sheet, I18n.Text("Fatigue Points"));
        mState = State.UNAFFECTED;
        addLabelAndField(sheet, GURPSCharacter.ID_CURRENT_FP, I18n.Text("Current"), I18n.Text("Current fatigue points"), true);
        addLabelAndField(sheet, GURPSCharacter.ID_FATIGUE_POINTS, I18n.Text("Basic"), I18n.Text("Normal (i.e. fully rested) fatigue points"), true);
        mTiredField = addLabelAndField(sheet, GURPSCharacter.ID_TIRED_FATIGUE_POINTS, I18n.Text("Tired"), I18n.Text("Current fatigue points at or below this point indicate the character is very tired, halving move, dodge and strength"), false);
        mCollapsedField = addLabelAndField(sheet, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, I18n.Text("Collapse"), I18n.Text("Current fatigue points at or below this point indicate the character is on the verge of collapse, causing the character to roll vs. Will to do anything besides talk or rest"), false);
        mUnconsciousField = addLabelAndField(sheet, GURPSCharacter.ID_UNCONSCIOUS_FATIGUE_POINTS, I18n.Text("Unconscious"), I18n.Text("Current fatigue points at or below this point cause the character to fall unconscious"), false);
        adjustColors();
        sheet.getCharacter().addTarget(this, GURPSCharacter.ID_CURRENT_FP);
    }

    protected void adjustColors() {
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
        adjustColor(mTiredField, mState == State.TIRED);
        adjustColor(mCollapsedField, mState == State.COLLAPSED);
        adjustColor(mUnconsciousField, mState == State.UNCONSCIOUS);
    }
}
