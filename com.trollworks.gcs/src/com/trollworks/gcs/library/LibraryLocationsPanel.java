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

import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.menu.library.ChangeLibraryLocationsCommand;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.StdDialog;
import com.trollworks.gcs.ui.widget.StdPanel;
import com.trollworks.gcs.ui.widget.StdScrollPanel;
import com.trollworks.gcs.ui.widget.Workspace;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.UpdateChecker;

import java.awt.EventQueue;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import javax.swing.JButton;
import javax.swing.JOptionPane;
import javax.swing.JTextField;

public final class LibraryLocationsPanel extends StdPanel {
    private List<LibraryFields> mFields;
    private JButton             mApplyButton;

    public static void showDialog() {
        // Close all documents
        Workspace workspace = Workspace.get();
        for (Dockable dockable : workspace.getDock().getDockables()) {
            if (dockable instanceof CloseHandler) {
                CloseHandler handler = (CloseHandler) dockable;
                if (handler.mayAttemptClose()) {
                    if (!handler.attemptClose()) {
                        JOptionPane.showMessageDialog(null, I18n.text("No documents may be open when setting library locations."), I18n.text("Canceled!"), JOptionPane.INFORMATION_MESSAGE);
                        return;
                    }
                }
            }
        }

        // Remove all library watchers
        LibraryWatcher.INSTANCE.watchDirs(new HashSet<>());

        // Ask the user to make changes
        LibraryLocationsPanel panel    = new LibraryLocationsPanel();
        StdScrollPanel        scroller = new StdScrollPanel(panel);
        StdDialog             dialog   = StdDialog.prepareToShowMessage(Workspace.get(), ChangeLibraryLocationsCommand.INSTANCE.getTitle(), MessageType.QUESTION, scroller);
        dialog.addButton(I18n.text("Add"), (evt) -> panel.addLibraryRow());
        dialog.addCancelButton();
        panel.mApplyButton = dialog.addApplyButton();
        panel.mFields.get(0).contentsChanged();
        dialog.presentToUser();
        int result = dialog.getResult();
        if (result == StdDialog.OK) {
            Library.LIBRARIES.clear();
            for (LibraryFields fields : panel.mFields) {
                switch (fields.getLibraryType()) {
                case MASTER -> {
                    Library.MASTER.setPath(fields.getPath());
                    Library.LIBRARIES.add(Library.MASTER);
                }
                case USER -> {
                    Library.USER.setPath(fields.getPath());
                    Library.LIBRARIES.add(Library.USER);
                }
                default -> Library.LIBRARIES.add(fields.createLibrary());
                }
            }
            Collections.sort(Library.LIBRARIES);
        }

        // Refresh the library view
        LibraryExplorerDockable libraryDockable = LibraryExplorerDockable.get();
        if (libraryDockable != null) {
            libraryDockable.refresh();
        }

        // Check to see if the new library locations need updating
        if (result == StdDialog.OK) {
            UpdateChecker.check();
        }
    }

    private LibraryLocationsPanel() {
        super(new PrecisionLayout().setColumns(8).setVerticalSpacing(1));
        mFields = new ArrayList<>();
        for (Library library : Library.LIBRARIES) {
            LibraryFields.LibraryType libType;
            if (library == Library.MASTER) {
                libType = LibraryFields.LibraryType.MASTER;
            } else if (library == Library.USER) {
                libType = LibraryFields.LibraryType.USER;
            } else {
                libType = LibraryFields.LibraryType.EXTRA;
            }
            mFields.add(new LibraryFields(this, library.getTitle(), library.getGitHubAccountName(), library.getRepoName(), library.getPathNoCreate().toString(), libType));
        }
    }

    public List<LibraryFields> getFields() {
        return mFields;
    }

    private void addLibraryRow() {
        mFields.add(new LibraryFields(this, "", "", "", "", LibraryFields.LibraryType.EXTRA));
        mFields.get(0).contentsChanged();
        revalidate();
        repaint();
        EventQueue.invokeLater(() -> {
            JTextField field = mFields.get(mFields.size() - 1).getTitleField();
            scrollRectToVisible(field.getBounds());
            field.requestFocus();
        });
    }

    public void setApplyState(boolean enabled) {
        if (mApplyButton != null) {
            mApplyButton.setEnabled(enabled);
        }
    }
}
