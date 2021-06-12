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

import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.WindowSizeEnforcer;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.KeyboardFocusManager;
import java.awt.Window;
import java.awt.event.ActionListener;
import javax.swing.JButton;
import javax.swing.JDialog;

public class StdDialog extends JDialog {
    public static final int       CLOSED = 0;
    public static final int       OK     = 1;
    public static final int       CANCEL = 2;
    public static final int       MARGIN = 10;
    private             Component mOwner;
    private             StdPanel  mButtons;
    private             int       mResult;

    public StdDialog(Component owner, String title) {
        super(WindowUtils.getWindowForComponent(owner), title, ModalityType.APPLICATION_MODAL);
        setResizable(true);
        StdPanel content = new StdPanel(new BorderLayout());
        StdPanel buttons = new StdPanel(new PrecisionLayout().setMargins(MARGIN).setHorizontalAlignment(PrecisionLayoutAlignment.MIDDLE));
        mButtons = new StdPanel(new PrecisionLayout().setEqualColumns(true).setMargins(0));
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

    public StdPanel getButtonsPanel() {
        return mButtons;
    }

    public JButton addButton(String title, int value) {
        return addButton(title, (evt) -> {
            mResult = value;
            setVisible(false);
        });
    }

    public JButton addButton(String title, ActionListener listener) {
        JButton button = new JButton(title);
        button.addActionListener(listener);
        mButtons.add(button, new PrecisionLayoutData().setFillHorizontalAlignment());
        ((PrecisionLayout) mButtons.getLayout()).setColumns(mButtons.getComponentCount());
        return button;
    }

    @SuppressWarnings("UnusedReturnValue")
    public JButton addCancelRemainingButton() {
        return addButton(I18n.text("Cancel Remaining"), CLOSED);
    }

    @SuppressWarnings("UnusedReturnValue")
    public JButton addCancelButton() {
        return addButton(I18n.text("Cancel"), CANCEL);
    }

    public JButton addApplyButton() {
        return addButton(I18n.text("Apply"), OK);
    }

    @SuppressWarnings("UnusedReturnValue")
    public JButton addOKButton() {
        return addButton(I18n.text("OK"), OK);
    }

    public JButton addYesButton() {
        return addButton(I18n.text("Yes"), OK);
    }

    /**
     * @param comp The {@link Component} to use for determining the parent {@link Window}.
     * @param msg  The message to display. If this is not a sub-class of {@link Component}, {@code
     *             toString()} will be called on it and the result will be used to create a label.
     */
    public static void showError(Component comp, Object msg) {
        showMessage(comp, I18n.text("Error"), MessageType.ERROR, msg);
    }

    /**
     * @param comp The {@link Component} to use for determining the parent {@link Window}.
     * @param msg  The message to display. If this is not a sub-class of {@link Component}, {@code
     *             toString()} will be called on it and the result will be used to create a label.
     */
    public static void showWarning(Component comp, Object msg) {
        showMessage(comp, I18n.text("Warning"), MessageType.WARNING, msg);
    }

    /**
     * @param comp    The {@link Component} to use for determining the parent {@link Window}.
     * @param title   The title to use.
     * @param msgType The {@link MessageType} this is.
     * @param msg     The message to display. If this is not a sub-class of {@link Component},
     *                {@code toString()} will be called on it and the result will be used to create
     *                a label.
     */
    @SuppressWarnings("UnusedReturnValue")
    public static int showMessage(Component comp, String title, MessageType msgType, Object msg) {
        StdDialog dialog = prepareToShowMessage(comp, title, msgType, msg);
        if (msgType == MessageType.QUESTION) {
            dialog.addCancelButton();
        }
        dialog.addOKButton();
        dialog.presentToUser();
        return dialog.getResult();
    }

    /**
     * @param comp    The {@link Component} to use for determining the parent {@link Window}.
     * @param title   The title to use.
     * @param msgType The {@link MessageType} this is.
     * @param msg     The message to display. If this is not a sub-class of {@link Component},
     *                {@code toString()} will be called on it and the result will be used to create
     *                a label.
     */
    public static StdDialog prepareToShowMessage(Component comp, String title, MessageType msgType, Object msg) {
        StdDialog dialog    = new StdDialog(comp, title);
        String    iconValue = msgType.getText();
        boolean   hasIcon   = !iconValue.isEmpty();
        StdPanel  content   = new StdPanel(new PrecisionLayout().setColumns(hasIcon ? 2 : 1).setHorizontalSpacing(20).setMargins(20));
        int       offset    = 0;
        if (hasIcon) {
            FontAwesomeIcon icon = new FontAwesomeIcon(iconValue, ThemeFont.LABEL_PRIMARY.getFont().getSize() * 3, 0, null);
            icon.setForeground(msgType.getColor());
            content.add(icon, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
            offset = (icon.getPreferredSize().height - TextDrawing.getFontHeight(ThemeFont.LABEL_PRIMARY.getFont())) / 2;
        }
        Component msgComp;
        if (msg instanceof Component) {
            msgComp = (Component) msg;
        } else {
            StdLabel label = new StdLabel(msg.toString());
            label.setTruncationPolicy(StdLabel.WRAP);
            msgComp = label;
        }
        content.add(msgComp, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(offset));
        dialog.getContentPane().add(content, BorderLayout.CENTER);
        dialog.setResizable(false);
        return dialog;
    }
}
