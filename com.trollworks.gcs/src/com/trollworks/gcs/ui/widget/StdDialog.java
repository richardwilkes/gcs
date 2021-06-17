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

import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.WindowSizeEnforcer;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.EventQueue;
import java.awt.KeyboardFocusManager;
import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyAdapter;
import java.awt.event.KeyEvent;
import java.awt.event.WindowAdapter;
import java.awt.event.WindowEvent;
import javax.swing.AbstractAction;
import javax.swing.ActionMap;
import javax.swing.InputMap;
import javax.swing.JComponent;
import javax.swing.JDialog;
import javax.swing.JRootPane;
import javax.swing.KeyStroke;
import javax.swing.SwingUtilities;

public class StdDialog extends JDialog {
    public static final int       CLOSED = 0;
    public static final int       OK     = 1;
    public static final int       CANCEL = 2;
    public static final int       MARGIN = 20;
    private             Component mOwner;
    private             StdPanel  mButtons;
    private             StdButton mOKButton;
    private             StdButton mCancelButton;
    private             int       mResult;
    private             int       mInitialFocusAttemptsRemaining;

    public StdDialog(Component owner, String title) {
        super(WindowUtils.getWindowForComponent(owner), title, ModalityType.APPLICATION_MODAL);
        setResizable(true);
        StdPanel content = new StdPanel(new BorderLayout());
        StdPanel buttons = new StdPanel(new PrecisionLayout().setMargins(MARGIN).setHorizontalAlignment(PrecisionLayoutAlignment.MIDDLE));
        mButtons = new StdPanel(new PrecisionLayout().setEqualColumns(true).setMargins(0).setHorizontalSpacing(10));
        buttons.add(mButtons);
        content.add(buttons, BorderLayout.SOUTH);
        setContentPane(content);
        mOwner = owner;
        installActions();
    }

    public StdButton getOKButton() {
        return mOKButton;
    }

    public StdButton getCancelButton() {
        return mCancelButton;
    }

    private void installActions() {
        final String pressOKKey     = "press_ok";
        final String pressCancelKey = "press_cancel";
        ActionMap    actionMap      = getRootPane().getActionMap();
        actionMap.put(pressOKKey, new KeyAction(this::getOKButton));
        actionMap.put(pressCancelKey, new KeyAction(this::getCancelButton));
        InputMap inputMap = getRootPane().getInputMap(JComponent.WHEN_IN_FOCUSED_WINDOW);
        inputMap.put(KeyStroke.getKeyStroke("ENTER"), pressOKKey);
        inputMap.put(KeyStroke.getKeyStroke("ESCAPE"), pressCancelKey);
    }

    private interface KeyButton {
        StdButton getButton();
    }

    private static class KeyAction extends AbstractAction {
        private KeyButton mAccessor;

        KeyAction(KeyButton accessor) {
            mAccessor = accessor;
        }

        @Override
        public void actionPerformed(ActionEvent event) {
            StdButton button = mAccessor.getButton();
            if (button != null && SwingUtilities.getRootPane(button) == event.getSource()) {
                button.click();
            }
        }

        @Override
        public boolean accept(Object sender) {
            if (sender instanceof JRootPane) {
                StdButton button = mAccessor.getButton();
                return button != null && button.isEnabled() && button.isShowing();
            }
            return true;
        }
    }

    public void presentToUser() {
        addWindowListener(new WindowAdapter() {
            @Override
            public void windowOpened(WindowEvent event) {
                mInitialFocusAttemptsRemaining = 5;
                tryInitialFocus();
            }
        });
        addKeyListener(new KeyAdapter() {
            @Override
            public void keyTyped(KeyEvent event) {
                if (event.getModifiersEx() == 0) {
                    char ch = event.getKeyChar();
                    if (ch == 27) {
                        if (mCancelButton != null && mCancelButton.isEnabled()) {
                            mCancelButton.click();
                        }
                    } else if (ch == '\n') {
                        if (mOKButton != null && mOKButton.isEnabled()) {
                            mOKButton.click();
                        }
                    }
                }
            }
        });
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

    private void tryInitialFocus() {
        if (--mInitialFocusAttemptsRemaining > 0 && !hasFocus()) {
            requestFocus();
            EventQueue.invokeLater(this::tryInitialFocus);
        } else {
            getContentPane().transferFocus();
        }
    }

    public int getResult() {
        return mResult;
    }

    public StdPanel getButtonsPanel() {
        return mButtons;
    }

    public StdButton addButton(String title, int value) {
        StdButton button = addButton(title, (btn) -> {
            mResult = value;
            setVisible(false);
        });
        if (value == OK) {
            mOKButton = button;
        } else if (value == CANCEL) {
            mCancelButton = button;
        }
        return button;
    }

    public StdButton addButton(String title, StdButton.ClickFunction clickHandler) {
        StdButton button = new StdButton(title, clickHandler);
        mButtons.add(button, new PrecisionLayoutData().setFillHorizontalAlignment().setMinimumWidth(60));
        ((PrecisionLayout) mButtons.getLayout()).setColumns(mButtons.getComponentCount());
        return button;
    }

    @SuppressWarnings("UnusedReturnValue")
    public StdButton addCancelRemainingButton() {
        return addButton(I18n.text("Cancel Remaining"), CLOSED);
    }

    @SuppressWarnings("UnusedReturnValue")
    public StdButton addCancelButton() {
        return addButton(I18n.text("Cancel"), CANCEL);
    }

    public StdButton addNoButton() {
        return addButton(I18n.text("No"), CANCEL);
    }

    public StdButton addApplyButton() {
        return addButton(I18n.text("Apply"), OK);
    }

    @SuppressWarnings("UnusedReturnValue")
    public StdButton addOKButton() {
        return addButton(I18n.text("OK"), OK);
    }

    public StdButton addYesButton() {
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
        StdDialog dialog  = new StdDialog(comp, title);
        StdPanel  content = new StdPanel(new BorderLayout(MARGIN, 0));
        content.setBorder(new EmptyBorder(MARGIN, MARGIN, 0, MARGIN));
        String  iconValue = msgType.getText();
        boolean hasIcon   = !iconValue.isEmpty();
        if (hasIcon) {
            StdPanel        left = new StdPanel(new BorderLayout());
            FontAwesomeIcon icon = new FontAwesomeIcon(iconValue, ThemeFont.LABEL_PRIMARY.getFont().getSize() * 3, 0, null);
            icon.setForeground(msgType.getColor());
            left.add(icon, BorderLayout.NORTH);
            content.add(left, BorderLayout.WEST);
        }
        Component msgComp;
        if (msg instanceof Component) {
            msgComp = (Component) msg;
        } else {
            StdLabel label = new StdLabel(msg.toString());
            label.setTruncationPolicy(StdLabel.WRAP);
            msgComp = label;
        }
        content.add(msgComp, BorderLayout.CENTER);
        dialog.getContentPane().add(content, BorderLayout.CENTER);
        dialog.setResizable(false);
        return dialog;
    }
}
