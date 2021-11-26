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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.SelfControlRoll;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;

/** Asks the user to enable/disable advantage modifiers. */
public final class AdvantageModifierEnabler extends Panel {
    private Advantage           mAdvantage;
    private Checkbox[]          mEnabled;
    private AdvantageModifier[] mModifiers;
    private PopupMenu<String>   mCRPopup;

    /**
     * Brings up a modal dialog that allows {@link AdvantageModifier}s to be enabled or disabled for
     * the specified {@link Advantage}s.
     *
     * @param comp       The component to open the dialog over.
     * @param advantages The {@link Advantage}s to process.
     * @return Whether anything was modified.
     */
    public static boolean process(Component comp, List<Advantage> advantages) {
        List<Advantage> list     = new ArrayList<>();
        boolean         modified = false;
        int             count;

        for (Advantage advantage : advantages) {
            if (advantage.getCR() != SelfControlRoll.NONE_REQUIRED || !advantage.getModifiers().isEmpty()) {
                list.add(advantage);
            }
        }

        count = list.size();
        for (int i = 0; i < count; i++) {
            AdvantageModifierEnabler panel  = new AdvantageModifierEnabler(list.get(i), count - i - 1);
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

    private AdvantageModifierEnabler(Advantage advantage, int remaining) {
        super(new BorderLayout());
        mAdvantage = advantage;
        add(createTop(advantage, remaining), BorderLayout.NORTH);
        ScrollPanel scrollPanel = new ScrollPanel(createCenter());
        scrollPanel.setMinimumSize(new Dimension(500, 120));
        add(scrollPanel, BorderLayout.CENTER);
    }

    private static Container createTop(Advantage advantage, int remaining) {
        Panel top   = new Panel(new ColumnLayout());
        Label label = new Label(Text.truncateIfNecessary(advantage.toString(), 80, SwingConstants.CENTER), SwingConstants.LEFT);

        top.setBorder(new EmptyBorder(0, 0, 15, 0));
        if (remaining > 0) {
            String msg;
            msg = remaining == 1 ? I18n.text("1 advantage remaining to be processed.") : MessageFormat.format(I18n.text("{0} advantages remaining to be processed."), Integer.valueOf(remaining));
            top.add(new Label(msg, SwingConstants.CENTER));
        }
        label.setBorder(new CompoundBorder(new LineBorder(), new EmptyBorder(0, 2, 0, 2)));
        label.setBackground(Colors.HEADER);
        label.setForeground(Colors.ON_HEADER);
        label.setOpaque(true);
        top.add(new Panel());
        top.add(label);
        return top;
    }

    private Container createCenter() {
        Panel panel = new Panel(new ColumnLayout());
        panel.setBorder(new EmptyBorder(0, 0, 16, 0));
        SelfControlRoll cr = mAdvantage.getCR();
        if (cr != SelfControlRoll.NONE_REQUIRED) {
            ArrayList<String> possible = new ArrayList<>();
            for (SelfControlRoll one : SelfControlRoll.values()) {
                if (one != SelfControlRoll.NONE_REQUIRED) {
                    possible.add(one.getDescriptionWithCost());
                }
            }
            mCRPopup = new PopupMenu<>(possible, null);
            mCRPopup.setSelectedItem(cr.getDescriptionWithCost(), false);
            panel.add(mCRPopup);
        }

        mModifiers = mAdvantage.getModifiers().toArray(new AdvantageModifier[0]);
        Arrays.sort(mModifiers);

        int length = mModifiers.length;
        mEnabled = new Checkbox[length];
        for (int i = 0; i < length; i++) {
            mEnabled[i] = new Checkbox(mModifiers[i].getFullDescription() + ", " + mModifiers[i].getCostDescription(), mModifiers[i].isEnabled(), null);
            panel.add(mEnabled[i]);
        }
        return panel;
    }

    private void applyChanges() {
        if (mAdvantage.getCR() != SelfControlRoll.NONE_REQUIRED) {
            String selectedItem = mCRPopup.getSelectedItem();
            if (selectedItem != null) {
                for (SelfControlRoll one : SelfControlRoll.values()) {
                    if (one != SelfControlRoll.NONE_REQUIRED) {
                        if (selectedItem.equals(one.getDescriptionWithCost())) {
                            mAdvantage.setCR(one);
                            break;
                        }
                    }
                }
            }
        }
        int length = mModifiers.length;
        for (int i = 0; i < length; i++) {
            mModifiers[i].setEnabled(mEnabled[i].isChecked());
        }
    }
}
