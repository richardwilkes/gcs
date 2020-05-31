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
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import javax.swing.SwingConstants;

/** The character attributes panel. */
public class AttributesPanel extends DropPanel {
    /**
     * Creates a new attributes panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public AttributesPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(2, 0).setAlignment(PrecisionLayoutAlignment.FILL, PrecisionLayoutAlignment.FILL), I18n.Text("Attributes"));
        addLabelAndField(sheet, GURPSCharacter.ID_STRENGTH, I18n.Text("Strength (ST)"), true);
        addLabelAndField(sheet, GURPSCharacter.ID_DEXTERITY, I18n.Text("Dexterity (DX)"), true);
        addLabelAndField(sheet, GURPSCharacter.ID_INTELLIGENCE, I18n.Text("Intelligence (IQ)"), true);
        addLabelAndField(sheet, GURPSCharacter.ID_HEALTH, I18n.Text("Health (HT)"), true);
        addDivider();
        addLabelAndField(sheet, GURPSCharacter.ID_WILL, I18n.Text("Will"), true);
        addLabelAndField(sheet, GURPSCharacter.ID_FRIGHT_CHECK, I18n.Text("Fright Check"), false);
        addDivider();
        addLabelAndField(sheet, GURPSCharacter.ID_BASIC_SPEED, I18n.Text("Basic Speed"), true);
        addLabelAndField(sheet, GURPSCharacter.ID_BASIC_MOVE, I18n.Text("Basic Move"), true);
        addDivider();
        addLabelAndField(sheet, GURPSCharacter.ID_PERCEPTION, I18n.Text("Perception (Per)"), true);
        addLabelAndField(sheet, GURPSCharacter.ID_VISION, I18n.Text("Vision"), false);
        addLabelAndField(sheet, GURPSCharacter.ID_HEARING, I18n.Text("Hearing"), false);
        addLabelAndField(sheet, GURPSCharacter.ID_TASTE_AND_SMELL, I18n.Text("Taste & Smell"), false);
        addLabelAndField(sheet, GURPSCharacter.ID_TOUCH, I18n.Text("Touch"), false);
        addDivider();
        addLabelAndDamageField(sheet, GURPSCharacter.ID_BASIC_THRUST, I18n.Text("Basic Thrust"));
        addLabelAndDamageField(sheet, GURPSCharacter.ID_BASIC_SWING, I18n.Text("Basic Swing"));
    }

    private void addLabelAndField(CharacterSheet sheet, String key, String title, boolean enabled) {
        add(new PagePoints(sheet, key), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        PageField field = new PageField(sheet, key, SwingConstants.RIGHT, enabled, null);
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, field));
    }

    private void addLabelAndDamageField(CharacterSheet sheet, String key, String title) {
        PageField field = new PageField(sheet, key, SwingConstants.RIGHT, false, null);
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL).setHorizontalSpan(2));
        add(new PageLabel(title, field));
    }

    private void addDivider() {
        Wrapper panel = new Wrapper();
        add(panel, new PrecisionLayoutData().setHorizontalSpan(3).setHeightHint(1));
        addHorizontalBackground(panel, Color.black);
    }
}
