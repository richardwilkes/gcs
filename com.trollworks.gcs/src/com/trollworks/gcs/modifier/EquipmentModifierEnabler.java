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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;

/** Asks the user to enable/disable equipment modifiers. */
public final class EquipmentModifierEnabler extends Panel {
    private Equipment           mEquipment;
    private Checkbox[]          mEnabled;
    private EquipmentModifier[] mModifiers;

    /**
     * Brings up a modal dialog that allows {@link EquipmentModifier}s to be enabled or disabled for
     * the specified {@link Equipment}s.
     *
     * @param comp      The component to open the dialog over.
     * @param equipment The {@link Equipment}s to process.
     * @return Whether anything was modified.
     */
    public static boolean process(Component comp, List<Equipment> equipment) {
        List<Equipment> list     = new ArrayList<>();
        boolean         modified = false;
        int             count;

        for (Equipment eqp : equipment) {
            if (!eqp.getModifiers().isEmpty()) {
                list.add(eqp);
            }
        }

        count = list.size();
        for (int i = 0; i < count; i++) {
            EquipmentModifierEnabler panel  = new EquipmentModifierEnabler(list.get(i), count - i - 1);
            Modal                    dialog = Modal.prepareToShowMessage(comp, I18n.text("Enable Modifiers"), MessageType.QUESTION, panel);
            if (i != count - 1) {
                dialog.addCancelRemainingButton();
            }
            dialog.addCancelButton();
            dialog.addApplyButton();
            dialog.presentToUser();
            switch (dialog.getResult()) {
            case Modal.OK:
                panel.applyChanges();
                modified = true;
                break;
            case Modal.CANCEL:
                break;
            case Modal.CLOSED:
            default:
                return modified;
            }
        }
        return modified;
    }

    private EquipmentModifierEnabler(Equipment equipment, int remaining) {
        super(new BorderLayout());
        mEquipment = equipment;
        add(createTop(equipment, remaining), BorderLayout.NORTH);
        ScrollPanel scrollPanel = new ScrollPanel(createCenter());
        scrollPanel.setMinimumSize(new Dimension(500, 120));
        add(scrollPanel, BorderLayout.CENTER);
    }

    private static Container createTop(Equipment equipment, int remaining) {
        Panel top   = new Panel(new ColumnLayout());
        Label label = new Label(Text.truncateIfNecessary(equipment.toString(), 80, SwingConstants.CENTER), SwingConstants.LEFT);
        top.setBorder(new EmptyBorder(0, 0, 15, 0));
        if (remaining > 0) {
            top.add(new Label(MessageFormat.format(I18n.text("{0} equipment remaining to be processed."), Integer.valueOf(remaining)), SwingConstants.CENTER));
        }
        label.setBorder(new CompoundBorder(new LineBorder(), new EmptyBorder(0, 2, 0, 2)));
        label.setBackground(Color.BLACK);
        label.setForeground(Color.WHITE);
        label.setOpaque(true);
        top.add(new Panel());
        top.add(label);
        return top;
    }

    private Container createCenter() {
        Panel panel = new Panel(new ColumnLayout());
        mModifiers = mEquipment.getModifiers().toArray(new EquipmentModifier[0]);
        Arrays.sort(mModifiers);

        int length = mModifiers.length;
        mEnabled = new Checkbox[length];
        for (int i = 0; i < length; i++) {
            mEnabled[i] = new Checkbox(mModifiers[i].getFullDescription(), mModifiers[i].isEnabled(), null);
            panel.add(mEnabled[i]);
        }
        return panel;
    }

    private void applyChanges() {
        int     length   = mModifiers.length;
        boolean modified = false;
        for (int i = 0; i < length; i++) {
            modified |= mModifiers[i].setEnabled(mEnabled[i].isChecked());
        }
        if (modified) {
            mEquipment.update();
        }
    }
}
