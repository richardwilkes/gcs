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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.WindowUtils;
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
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;

/** Asks the user to enable/disable equipment modifiers. */
public class EquipmentModifierEnabler extends JPanel {
    private Equipment           mEquipment;
    private JCheckBox[]         mEnabled;
    private EquipmentModifier[] mModifiers;
    private JComboBox<String>   mCRCombo;

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
            Equipment                eqp         = list.get(i);
            boolean                  hasMore     = i != count - 1;
            EquipmentModifierEnabler panel       = new EquipmentModifierEnabler(eqp, count - i - 1);
            String                   applyTitle  = I18n.Text("Apply");
            String                   cancelTitle = I18n.Text("Cancel");
            switch (WindowUtils.showOptionDialog(comp, panel, I18n.Text("Enable Modifiers"), true, hasMore ? JOptionPane.YES_NO_CANCEL_OPTION : JOptionPane.YES_NO_OPTION, JOptionPane.INFORMATION_MESSAGE, eqp.getIcon(true), hasMore ? new String[]{applyTitle, cancelTitle, I18n.Text("Cancel Remaining")} : new String[]{applyTitle, cancelTitle}, applyTitle)) {
            case JOptionPane.YES_OPTION:
                panel.applyChanges();
                modified = true;
                break;
            case JOptionPane.NO_OPTION:
                break;
            case JOptionPane.CANCEL_OPTION:
            case JOptionPane.CLOSED_OPTION:
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
        JScrollPane scrollPanel = new JScrollPane(createCenter());
        scrollPanel.setMinimumSize(new Dimension(500, 120));
        add(scrollPanel, BorderLayout.CENTER);
    }

    private static Container createTop(Equipment equipment, int remaining) {
        JPanel top   = new JPanel(new ColumnLayout());
        JLabel label = new JLabel(Text.truncateIfNecessary(equipment.toString(), 80, SwingConstants.RIGHT), SwingConstants.LEFT);
        top.setBorder(new EmptyBorder(0, 0, 15, 0));
        if (remaining > 0) {
            top.add(new JLabel(MessageFormat.format(I18n.Text("{0} equipment remaining to be processed."), Integer.valueOf(remaining)), SwingConstants.CENTER));
        }
        label.setBorder(new CompoundBorder(new LineBorder(), new EmptyBorder(0, 2, 0, 2)));
        label.setOpaque(true);
        top.add(new JPanel());
        top.add(label);
        return top;
    }

    private Container createCenter() {
        JPanel wrapper = new JPanel(new ColumnLayout());
        wrapper.setBackground(Color.WHITE);
        mModifiers = mEquipment.getModifiers().toArray(new EquipmentModifier[0]);
        Arrays.sort(mModifiers);

        int length = mModifiers.length;
        mEnabled = new JCheckBox[length];
        for (int i = 0; i < length; i++) {
            mEnabled[i] = new JCheckBox(mModifiers[i].getFullDescription(), mModifiers[i].isEnabled());
            wrapper.add(mEnabled[i]);
        }
        return wrapper;
    }

    private void applyChanges() {
        int     length   = mModifiers.length;
        boolean modified = false;
        for (int i = 0; i < length; i++) {
            modified |= mModifiers[i].setEnabled(mEnabled[i].isSelected());
        }
        if (modified) {
            mEquipment.update();
        }
    }
}
