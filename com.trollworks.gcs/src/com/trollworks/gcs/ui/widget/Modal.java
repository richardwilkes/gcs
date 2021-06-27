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

import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.WindowSizeEnforcer;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.FileDialog;
import java.awt.Frame;
import java.awt.KeyboardFocusManager;
import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyAdapter;
import java.awt.event.KeyEvent;
import java.awt.event.WindowAdapter;
import java.awt.event.WindowEvent;
import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import java.text.MessageFormat;
import javax.swing.AbstractAction;
import javax.swing.ActionMap;
import javax.swing.InputMap;
import javax.swing.JComponent;
import javax.swing.JDialog;
import javax.swing.JRootPane;
import javax.swing.KeyStroke;
import javax.swing.SwingUtilities;
import javax.swing.filechooser.FileNameExtensionFilter;

public class Modal extends JDialog {
    public static final int       CLOSED = 0;
    public static final int       OK     = 1;
    public static final int       CANCEL = 2;
    private             Component mOwner;
    private             Panel     mButtons;
    private             Button    mOKButton;
    private             Button    mCancelButton;
    private             int       mResult;

    public Modal(Component owner, String title) {
        super(WindowUtils.getWindowForComponent(owner), title, ModalityType.APPLICATION_MODAL);
        setResizable(true);
        Panel content = new Panel(new BorderLayout());
        Panel buttons = new Panel(new PrecisionLayout().setMargins(LayoutConstants.WINDOW_BORDER_INSET).setHorizontalAlignment(PrecisionLayoutAlignment.MIDDLE));
        mButtons = new Panel(new PrecisionLayout().setEqualColumns(true).setMargins(0).setHorizontalSpacing(10));
        buttons.add(mButtons);
        content.add(buttons, BorderLayout.SOUTH);
        setContentPane(content);
        mOwner = owner;
        installActions();
    }

    public Button getOKButton() {
        return mOKButton;
    }

    public Button getCancelButton() {
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
        Button getButton();
    }

    private static class KeyAction extends AbstractAction {
        private KeyButton mAccessor;

        KeyAction(KeyButton accessor) {
            mAccessor = accessor;
        }

        @Override
        public void actionPerformed(ActionEvent event) {
            Button button = mAccessor.getButton();
            if (button != null && SwingUtilities.getRootPane(button) == event.getSource()) {
                button.click();
            }
        }

        @Override
        public boolean accept(Object sender) {
            if (sender instanceof JRootPane) {
                Button button = mAccessor.getButton();
                return button != null && button.isEnabled() && button.isShowing();
            }
            return true;
        }
    }

    public void presentToUser() {
        addWindowListener(new WindowAdapter() {
            @Override
            public void windowOpened(WindowEvent event) {
                FocusHelper.focusOn(Modal.this);
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

    public int getResult() {
        return mResult;
    }

    public Panel getButtonsPanel() {
        return mButtons;
    }

    public Button addButton(String title, int value) {
        Button button = addButton(title, (btn) -> {
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

    public Button addButton(String title, Button.ClickFunction clickHandler) {
        Button button = new Button(title, clickHandler);
        mButtons.add(button, new PrecisionLayoutData().setFillHorizontalAlignment().setMinimumWidth(60));
        ((PrecisionLayout) mButtons.getLayout()).setColumns(mButtons.getComponentCount());
        return button;
    }

    @SuppressWarnings("UnusedReturnValue")
    public Button addCancelRemainingButton() {
        return addButton(I18n.text("Cancel Remaining"), CLOSED);
    }

    @SuppressWarnings("UnusedReturnValue")
    public Button addCancelButton() {
        return addButton(I18n.text("Cancel"), CANCEL);
    }

    public Button addNoButton() {
        return addButton(I18n.text("No"), CANCEL);
    }

    public Button addApplyButton() {
        return addButton(I18n.text("Apply"), OK);
    }

    @SuppressWarnings("UnusedReturnValue")
    public Button addOKButton() {
        return addButton(I18n.text("OK"), OK);
    }

    public Button addYesButton() {
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
        Modal dialog = prepareToShowMessage(comp, title, msgType, msg);
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
    public static Modal prepareToShowMessage(Component comp, String title, MessageType msgType, Object msg) {
        Modal dialog  = new Modal(comp, title);
        Panel content = new Panel(new BorderLayout(LayoutConstants.WINDOW_BORDER_INSET, 0));
        content.setBorder(new EmptyBorder(LayoutConstants.WINDOW_BORDER_INSET, LayoutConstants.WINDOW_BORDER_INSET, 0, LayoutConstants.WINDOW_BORDER_INSET));
        String  iconValue = msgType.getText();
        boolean hasIcon   = !iconValue.isEmpty();
        if (hasIcon) {
            Panel left = new Panel(new BorderLayout());
            Label icon = new Label(iconValue);
            icon.setThemeFont(Fonts.FONT_ICON_HUGE);
            icon.setForeground(msgType.getColor());
            left.add(icon, BorderLayout.NORTH);
            content.add(left, BorderLayout.WEST);
        }
        Component msgComp;
        if (msg instanceof Component) {
            msgComp = (Component) msg;
        } else {
            Label label = new Label(msg.toString());
            label.setTruncationPolicy(Label.WRAP);
            msgComp = label;
        }
        content.add(msgComp, BorderLayout.CENTER);
        dialog.getContentPane().add(content, BorderLayout.CENTER);
        dialog.setResizable(false);
        return dialog;
    }

    /**
     * Creates a new open file dialog.
     *
     * @param comp    The parent {@link Component} of the dialog. May be {@code null}.
     * @param title   The title to use. May be {@code null}.
     * @param dirs    The directory to start in. May be {@code null}.
     * @param filters The file filters to make available. If there are none, then the {@code
     *                showAllFilter} flag will be forced to {@code true}.
     * @return The chosen {@link Path} or {@code null}.
     */
    public static Path presentOpenFileDialog(Component comp, String title, Dirs dirs, FileNameExtensionFilter... filters) {
        FileDialog dialog = createFileDialog(comp, title, FileDialog.LOAD, dirs, filters);
        Path       path   = presentFileDialog(dialog, dirs);
        if (path != null) {
            Settings.getInstance().addRecentFile(path);
        }
        return path;
    }

    /**
     * Creates a new save file dialog.
     *
     * @param comp              The parent {@link Component} of the dialog. May be {@code null}.
     * @param title             The title to use. May be {@code null}.
     * @param dirs              The directory to show initially. May be {@code null}.
     * @param suggestedFileName The suggested file name to save as. May be {@code null}.
     * @param filter            The file filter to use.
     * @return The chosen {@link Path} or {@code null}.
     */
    public static Path presentSaveFileDialog(Component comp, String title, Dirs dirs, String suggestedFileName, FileNameExtensionFilter filter) {
        if (filter == null) {
            throw new IllegalArgumentException("must supply a file filter");
        }
        FileDialog dialog = createFileDialog(comp, title, FileDialog.SAVE, dirs, filter);
        if (suggestedFileName != null && !suggestedFileName.isBlank()) {
            dialog.setFile(PathUtils.getLeafName(suggestedFileName, true));
        }
        Path path = presentFileDialog(dialog, dirs);
        if (path != null) {
            if (!PathUtils.isNameValidForFile(PathUtils.getLeafName(path, true))) {
                showError(comp, I18n.text("Invalid file name"));
                return null;
            }
            if (!filter.accept(path.toFile())) {
                path = path.resolveSibling(PathUtils.enforceExtension(PathUtils.getLeafName(path, true), filter.getExtensions()[0]));
                // Only check for existance if we had to modify the path. Otherwise, we'd get a
                // double prompt since the platform's save dialog should have already prompted if
                // there was a match.
                if (Files.exists(path)) {
                    Modal question = prepareToShowMessage(comp,
                            I18n.text("Already Exists!"),
                            MessageType.QUESTION,
                            String.format(I18n.text("%s already exists!\nDo you want to replace it?"), path));
                    question.addCancelButton();
                    question.addButton(I18n.text("Replace"), OK);
                    question.presentToUser();
                    if (question.getResult() == CANCEL) {
                        return null;
                    }
                }
            }
            Settings.getInstance().addRecentFile(path);
        }
        return path;
    }

    /**
     * @param comp      The {@link Component} to use for determining the parent {@link Frame} or
     *                  {@link java.awt.Dialog}.
     * @param name      The name of the file that cannot be opened.
     * @param throwable The {@link Throwable}, if any, that caused the failure.
     */
    public static void showCannotOpenMsg(Component comp, String name, Throwable throwable) {
        if (throwable != null) {
            Log.error(throwable);
            showError(comp, MessageFormat.format(I18n.text("Unable to open \"{0}\"\n{1}"), name, throwable.getMessage()));
        } else {
            showError(comp, MessageFormat.format(I18n.text("Unable to open \"{0}\"."), name));
        }
    }

    private static FileDialog createFileDialog(Component comp, String title, int mode, Dirs dirs, FileNameExtensionFilter... filters) {
        FileDialog dialog;
        Window     window = UIUtilities.getSelfOrAncestorOfType(comp, Window.class);
        if (window instanceof Frame) {
            dialog = new FileDialog((Frame) window, title, mode);
        } else {
            dialog = new FileDialog((java.awt.Dialog) window, title, mode);
        }
        if (dirs == null) {
            dirs = Dirs.GENERAL;
        }
        dialog.setDirectory(dirs.get().toString());
        if (filters != null) {
            dialog.setFilenameFilter((File dir, String name) -> {
                File f = new File(dir, name);
                for (FileNameExtensionFilter filter : filters) {
                    if (filter.accept(f)) {
                        return true;
                    }
                }
                return false;
            });
        }
        return dialog;
    }

    private static Path presentFileDialog(FileDialog dialog, Dirs dirs) {
        dialog.setVisible(true);
        String dir = dialog.getDirectory();
        if (dir == null) {
            return null;
        }
        if (dirs == null) {
            dirs = Dirs.GENERAL;
        }
        dirs.set(Path.of(dir));
        String file = dialog.getFile();
        if (file == null) {
            return null;
        }
        return Path.of(dir, file).normalize().toAbsolutePath();
    }
}
