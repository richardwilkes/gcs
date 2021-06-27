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

package com.trollworks.gcs.library;

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Separator;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Platform;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import javax.swing.JComponent;
import javax.swing.JFileChooser;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

public class LibraryFields implements DocumentListener {
    private LibraryLocationsPanel mOwner;
    private LibraryType           mLibraryType;
    private EditorField           mTitle;
    private EditorField           mGitHubAccountName;
    private EditorField           mRepoName;
    private EditorField           mPath;
    private List<JComponent>      mComps;

    public enum LibraryType {
        MASTER,
        USER,
        EXTRA
    }

    public LibraryFields(LibraryLocationsPanel owner, String title, String account, String repo, String path, LibraryType libType) {
        mOwner = owner;
        mLibraryType = libType;
        mComps = new ArrayList<>();

        if (owner.getComponentCount() > 0) {
            Separator sep = new Separator();
            mOwner.add(sep, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment().setHorizontalSpan(7).setTopMargin(5).setBottomMargin(5));
            mComps.add(sep);
        }

        mTitle = addLabelAndField(I18n.text("Name:"), title, 1);
        mGitHubAccountName = addLabelAndField(I18n.text("GitHub Account:"), account, 1);
        mRepoName = addLabelAndField(I18n.text("Repo:"), repo, 1);
        FontIconButton button = new FontIconButton(FontAwesome.SEARCH_LOCATION, I18n.text("Locate"), (b) -> {
            String       currentPath = mPath.getText();
            Path         current     = currentPath.isBlank() ? Dirs.GENERAL.get() : Path.of(currentPath).getParent().toAbsolutePath();
            JFileChooser dialog      = new JFileChooser(current.toString());
            dialog.setDialogTitle(mTitle.getText());
            dialog.setFileSelectionMode(JFileChooser.DIRECTORIES_ONLY);
            if (dialog.showDialog(mOwner, I18n.text("Select")) == JFileChooser.APPROVE_OPTION) {
                File selectedFile = dialog.getSelectedFile();
                Dirs.GENERAL.set(selectedFile.toPath().getParent());
                mPath.setText(selectedFile.getAbsolutePath());
                mPath.requestFocus();
                mPath.selectAll();
            }
        });
        mOwner.add(button);
        mComps.add(button);

        mPath = addLabelAndField(I18n.text("Path:"), path, 5);
        if (mLibraryType == LibraryType.EXTRA) {
            button = new FontIconButton(FontAwesome.TRASH, I18n.text("Remove"), (b) -> {
                for (JComponent comp : mComps) {
                    mOwner.remove(comp);
                }
                mOwner.getFields().remove(this);
                contentsChanged();
                mOwner.revalidate();
                mOwner.repaint();
            });
        } else {
            button = new FontIconButton(FontAwesome.POWER_OFF, I18n.text("Use Default"), (b) -> {
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
                mPath.requestFocus();
                mPath.selectAll();
            });
        }
        mOwner.add(button);
        mComps.add(button);

        if (mLibraryType != LibraryType.EXTRA) {
            mTitle.setEnabled(false);
            mGitHubAccountName.setEnabled(false);
            mRepoName.setEnabled(false);
        }
    }

    public LibraryType getLibraryType() {
        return mLibraryType;
    }

    public EditorField getTitleField() {
        return mTitle;
    }

    private EditorField addLabelAndField(String title, String value, int columns) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, value, null);
        field.getDocument().addDocumentListener(this);
        Label label = new Label(title);
        mOwner.add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
        mOwner.add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment().setHorizontalSpan(columns));
        mComps.add(label);
        mComps.add(field);
        return field;
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

    private static boolean adjustColors(EditorField field, boolean valid) {
        field.setForeground(valid ? Colors.ON_EDITABLE : Colors.ON_ERROR);
        field.setBackground(valid ? Colors.EDITABLE : Colors.ERROR);
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
        Path path = Path.of(text).toAbsolutePath().normalize();
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
