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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.WindowSizeEnforcer;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.KeyboardFocusManager;
import javax.swing.JButton;
import javax.swing.JDialog;

public class StdDialog extends JDialog {
    public static final int       CLOSED = 0;
    public static final int       OK     = 1;
    public static final int       CANCEL = 2;
    public static final int       MARGIN = 10;
    private             Component mOwner;
    private             Panel     mButtons;
    private             int       mResult;

    public StdDialog(Component owner, String title) {
        super(WindowUtils.getWindowForComponent(owner), title, ModalityType.APPLICATION_MODAL);
        setResizable(true);
        Panel content = new Panel(new BorderLayout());
        Panel buttons = new Panel(new PrecisionLayout().setMargins(MARGIN).setHorizontalAlignment(PrecisionLayoutAlignment.MIDDLE));
        mButtons = new Panel(new PrecisionLayout().setEqualColumns(true).setMargins(0));
        buttons.add(mButtons);
        content.add(buttons, BorderLayout.SOUTH);
        setContentPane(content);
        mOwner = owner;
    }

    public void presentToUser() {
        WindowSizeEnforcer.monitor(this);
        Component focusOwner = KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
        if (focusOwner == null) {
            focusOwner = mOwner;
        }
        WindowUtils.packAndCenterWindowOn(this, WindowUtils.getWindowForComponent(mOwner));
        setVisible(true);
        dispose();
        if (focusOwner != null) {
            focusOwner.requestFocus();
        }
    }

    public int getResult() {
        return mResult;
    }

    public void addButton(String title, int value) {
        JButton button = new JButton(title);
        button.addActionListener((evt) -> {
            mResult = value;
            setVisible(false);
        });
        mButtons.add(button, new PrecisionLayoutData().setFillHorizontalAlignment());
        ((PrecisionLayout) mButtons.getLayout()).setColumns(mButtons.getComponentCount());
    }

    public void addCancelRemainingButton() {
        addButton(I18n.text("Cancel Remaining"), CLOSED);
    }

    public void addCancelButton() {
        addButton(I18n.text("Cancel"), CANCEL);
    }

    public void addApplyButton() {
        addButton(I18n.text("Apply"), OK);
    }
}
