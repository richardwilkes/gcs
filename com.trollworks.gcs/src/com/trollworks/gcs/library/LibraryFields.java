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

import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Platform;

import java.awt.Color;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import javax.swing.JButton;
import javax.swing.JComponent;
import javax.swing.JFileChooser;
import javax.swing.JLabel;
import javax.swing.JSeparator;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

public class LibraryFields implements DocumentListener {
    private LibraryLocationsPanel mOwner;
    private LibraryType           mLibraryType;
    private JTextField            mTitle;
    private JTextField            mGitHubAccountName;
    private JTextField            mRepoName;
    private JTextField            mPath;
    private List<JComponent>      mComps;
    private Color                 mNormalForeground;
    private Color                 mNormalBackground;

    public enum LibraryType {
        MASTER, USER, EXTRA
    }

    public LibraryFields(LibraryLocationsPanel owner, String title, String account, String repo, String path, LibraryType libType) {
        mOwner = owner;
        mLibraryType = libType;
        mComps = new ArrayList<>();

        addPlaceholder();
        mTitle = addLabelAndField(I18n.Text("Name:"), title, 1);
        mGitHubAccountName = addLabelAndField(I18n.Text("GitHub Account:"), account, 1);
        mRepoName = addLabelAndField(I18n.Text("Repo:"), repo, 1);
        addLocateButton();

        addRemoveButton();
        mPath = addLabelAndField(I18n.Text("Path:"), path, 5);
        addUseDefaultButton();

        JSeparator sep = new JSeparator(SwingConstants.HORIZONTAL);
        mOwner.add(sep, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment().setHorizontalSpan(8));
        mComps.add(sep);

        if (mLibraryType != LibraryType.EXTRA) {
            mTitle.setEnabled(false);
            mGitHubAccountName.setEnabled(false);
            mRepoName.setEnabled(false);
        }
        mNormalForeground = mPath.getForeground();
        mNormalBackground = mPath.getBackground();
    }

    public LibraryType getLibraryType() {
        return mLibraryType;
    }

    public JTextField getTitleField() {
        return mTitle;
    }

    private JTextField addLabelAndField(String title, String value, int columns) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        mOwner.add(label, new PrecisionLayoutData().setEndHorizontalAlignment());
        JTextField field = new JTextField(value);
        mOwner.add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment().setHorizontalSpan(columns));
        field.getDocument().addDocumentListener(this);
        mComps.add(label);
        mComps.add(field);
        return field;
    }

    private void addRemoveButton() {
        if (mLibraryType == LibraryType.EXTRA) {
            JButton button = new JButton(I18n.Text("Remove"));
            button.addActionListener(e -> {
                for (JComponent comp : mComps) {
                    mOwner.remove(comp);
                }
                mOwner.getFields().remove(this);
                contentsChanged();
                mOwner.revalidate();
                mOwner.repaint();
            });
            mOwner.add(button);
            mComps.add(button);
        } else {
            addPlaceholder();
        }
    }

    private void addPlaceholder() {
        Wrapper wrapper = new Wrapper();
        mOwner.add(wrapper);
        mComps.add(wrapper);
    }

    private JButton addLocateButton() {
        JButton button = new JButton(I18n.Text("Locate"));
        button.addActionListener(e -> {
            String       path    = mPath.getText();
            Path         current = path.isBlank() ? Preferences.getInstance().getLastDir() : Paths.get(path).getParent().toAbsolutePath();
            JFileChooser dialog  = new JFileChooser(current.toString());
            dialog.setDialogTitle(mTitle.getText());
            dialog.setFileSelectionMode(JFileChooser.DIRECTORIES_ONLY);
            if (dialog.showDialog(mOwner, I18n.Text("Select")) == JFileChooser.APPROVE_OPTION) {
                mPath.setText(dialog.getSelectedFile().getAbsolutePath());
                mPath.requestFocusInWindow();
                mPath.selectAll();
            }
        });
        mOwner.add(button);
        mComps.add(button);
        return button;
    }

    private void addUseDefaultButton() {
        if (mLibraryType == LibraryType.EXTRA) {
            addPlaceholder();
        } else {
            JButton button = new JButton(I18n.Text("Use Default"));
            button.addActionListener(e -> {
                Path def;
                switch (mLibraryType) {
                case MASTER:
                    def = Library.getDefaultMasterLibraryPath();
                    break;
                case USER:
                    def = Library.getDefaultUserLibraryPath();
                    break;
                default: // Shouldn't ever reach this
                    return;
                }
                mPath.setText(def.toString());
                mPath.requestFocusInWindow();
                mPath.selectAll();
            });
            mOwner.add(button);
            mComps.add(button);
        }
    }

    public void contentsChanged() {
        Set<String> titles          = new HashSet<>();
        Set<String> accountAndRepos = new HashSet<>();
        Set<String> paths           = new HashSet<>();
        boolean     hasProblem      = false;
        for (LibraryFields one : mOwner.getFields()) {
            hasProblem |= one.adjustTitleField(titles);
            hasProblem |= one.adjustAccountAndRepoFields(accountAndRepos);
            hasProblem |= one.adjustPathField(paths);
        }
        mOwner.setApplyState(!hasProblem);
    }

    private boolean adjustTitleField(Set<String> set) {
        String title = mTitle.getText().trim().toLowerCase();
        return adjustColors(mTitle, !title.isBlank() && set.add(title));
    }

    private boolean adjustAccountAndRepoFields(Set<String> set) {
        String  account = mGitHubAccountName.getText().trim();
        String  repo    = mRepoName.getText().trim();
        boolean valid   = !account.isBlank() && !repo.isBlank() && set.add(account + "/" + repo);
        adjustColors(mGitHubAccountName, valid);
        return adjustColors(mRepoName, valid);
    }

    private boolean adjustPathField(Set<String> set) {
        String path = mPath.getText().trim();
        if (!Platform.isLinux()) {
            path = path.toLowerCase();
        }
        return adjustColors(mPath, !path.isBlank() && set.add(path) && getPath() != null);
    }

    private boolean adjustColors(JTextField field, boolean valid) {
        field.setForeground(valid ? mNormalForeground : Color.BLACK);
        field.setBackground(valid ? mNormalBackground : Color.ORANGE);
        return !valid;
    }

    public Library createLibrary() {
        return new Library(mTitle.getText().trim(), mGitHubAccountName.getText().trim(), mRepoName.getText().trim(), getPath());
    }

    public Path getPath() {
        String text = mPath.getText();
        if (text.isBlank()) {
            return null;
        }
        Path path = Paths.get(text).toAbsolutePath().normalize();
        if (!Files.isDirectory(path)) {
            return null;
        }
        return path;
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
