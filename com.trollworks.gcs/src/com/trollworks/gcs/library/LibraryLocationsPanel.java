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

package com.trollworks.gcs.library;

import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.menu.file.QuitCommand;
import com.trollworks.gcs.menu.library.ChangeLibraryLocationsCommand;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.ui.widget.Workspace;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import java.awt.Dimension;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import javax.swing.JButton;
import javax.swing.JFileChooser;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JTextArea;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.UIManager;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

public class LibraryLocationsPanel extends JPanel implements DocumentListener {
    private JTextField mMasterLibraryPath;
    private JTextField mUserLibraryPath;
    private JButton    mApplyButton;
    private JButton    mCancelButton;
    private Color      mNormalForeground;
    private Color      mNormalBackground;

    public static void showDialog() {
        // Close all documents
        Workspace workspace = Workspace.get();
        for (Dockable dockable : workspace.getDock().getDockables()) {
            if (dockable instanceof CloseHandler) {
                CloseHandler handler = (CloseHandler) dockable;
                if (handler.mayAttemptClose()) {
                    if (!handler.attemptClose()) {
                        JOptionPane.showMessageDialog(null, I18n.Text("No documents may be open when setting the library locations."), I18n.Text("Canceled!"), JOptionPane.INFORMATION_MESSAGE);
                        return;
                    }
                }
            }
        }

        LibraryLocationsPanel panel = new LibraryLocationsPanel();
        if (WindowUtils.showOptionDialog(Workspace.get(), panel, ChangeLibraryLocationsCommand.INSTANCE.getTitle(), true, JOptionPane.OK_CANCEL_OPTION, JOptionPane.PLAIN_MESSAGE, null, new JButton[]{panel.mApplyButton, panel.mCancelButton}, panel.mCancelButton) == JOptionPane.OK_OPTION) {
            Library.MASTER.setPath(Paths.get(panel.mMasterLibraryPath.getText()));
            Library.USER.setPath(Paths.get(panel.mUserLibraryPath.getText()));
            QuitCommand.INSTANCE.attemptQuit();
        }
    }

    private LibraryLocationsPanel() {
        super(new PrecisionLayout().setColumns(4));
        String    masterLibraryTitle = Library.MASTER.getTitle();
        JTextArea warning            = new JTextArea(String.format(I18n.Text("Warning: The directory chosen for the %1$s Path will have its contents deleted and replaced when GCS starts up again if it does not already contain the %1$s data."), masterLibraryTitle));
        warning.setEditable(false);
        warning.setWrapStyleWord(true);
        warning.setLineWrap(true);
        warning.setFont(UIManager.getFont("Label.font"));
        warning.setOpaque(false);
        add(warning, new PrecisionLayoutData().setHorizontalSpan(4).setFillHorizontalAlignment().setWidthHint(400).setBottomMargin(10));
        mMasterLibraryPath = createField(String.format("%s Path:", masterLibraryTitle), Library.MASTER.getPath(), Library.getDefaultMasterLibraryPath());
        mUserLibraryPath = createField(String.format("%s Path:", Library.USER.getTitle()), Library.USER.getPath(), Library.getDefaultUserLibraryPath());
        mApplyButton = createDialogButton(I18n.Text("Apply & Quit"));
        mApplyButton.setEnabled(false);
        mCancelButton = createDialogButton(I18n.Text("Cancel"));
        mNormalForeground = mMasterLibraryPath.getForeground();
        mNormalBackground = mMasterLibraryPath.getBackground();
    }

    private JTextField createField(String title, Path value, Path def) {
        add(new JLabel(title, SwingConstants.RIGHT), new PrecisionLayoutData().setEndHorizontalAlignment());
        JTextField field    = new JTextField(value.toString());
        Dimension  prefSize = field.getPreferredSize();
        if (prefSize.width < 400) {
            prefSize.width = 400;
            field.setPreferredSize(prefSize);
        }
        field.getDocument().addDocumentListener(this);
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        JButton button = new JButton(I18n.Text("Locate"));
        button.addActionListener(e -> {
            Path         current = Paths.get(field.getText()).toAbsolutePath();
            JFileChooser dialog  = new JFileChooser(current.getParent().toString());
            dialog.setDialogTitle(title);
            dialog.setFileSelectionMode(JFileChooser.DIRECTORIES_ONLY);
            if (dialog.showDialog(this, I18n.Text("Select")) == JFileChooser.APPROVE_OPTION) {
                field.setText(dialog.getSelectedFile().getAbsolutePath());
                field.requestFocusInWindow();
                field.selectAll();
            }
        });
        add(button);
        button = new JButton(I18n.Text("Use Default"));
        button.addActionListener(e -> {
            field.setText(def.toString());
            field.requestFocusInWindow();
            field.selectAll();
        });
        add(button);
        return field;
    }

    private JButton createDialogButton(String title) {
        JButton button = new JButton(title);
        button.addActionListener(e -> {
            JOptionPane pane = UIUtilities.getAncestorOfType(button, JOptionPane.class);
            if (pane != null) {
                pane.setValue(button);
            }
        });
        return button;
    }

    private void contentsChanged() {
        Path masterPath = getPath(mMasterLibraryPath);
        Path userPath   = getPath(mUserLibraryPath);
        if (masterPath != null && userPath != null) {
            if (masterPath.startsWith(userPath) || userPath.startsWith(masterPath)) {
                masterPath = null;
                userPath = null;
            }
        }
        setColors(mMasterLibraryPath, masterPath != null);
        setColors(mUserLibraryPath, userPath != null);
        Preferences prefs = Preferences.getInstance();
        mApplyButton.setEnabled(masterPath != null && userPath != null && (!masterPath.equals(Library.MASTER.getPathNoCreate()) || !userPath.equals(Library.USER.getPathNoCreate())));
    }

    private Path getPath(JTextField field) {
        String text = field.getText();
        if (text.isBlank()) {
            return null;
        }
        Path path = Paths.get(text).toAbsolutePath().normalize();
        if (!Files.isDirectory(path)) {
            return null;
        }
        return path;
    }

    private void setColors(JTextField field, boolean valid) {
        if (valid) {
            field.setForeground(mNormalForeground);
            field.setBackground(mNormalBackground);
        } else {
            field.setForeground(Color.WHITE);
            field.setBackground(Color.RED);
        }
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        contentsChanged();
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        contentsChanged();
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        contentsChanged();
    }
}
