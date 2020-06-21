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
import com.trollworks.gcs.menu.library.ChangeLibraryLocationsCommand;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.widget.WindowUtils;
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
import javax.swing.JPanel;
import javax.swing.JScrollPane;

public class LibraryLocationsPanel extends JPanel {
    private List<LibraryFields> mFields;
    private JButton             mApplyButton;
    private JButton             mCancelButton;

    public static void showDialog() {
        // Close all documents
        Workspace workspace = Workspace.get();
        for (Dockable dockable : workspace.getDock().getDockables()) {
            if (dockable instanceof CloseHandler) {
                CloseHandler handler = (CloseHandler) dockable;
                if (handler.mayAttemptClose()) {
                    if (!handler.attemptClose()) {
                        JOptionPane.showMessageDialog(null, I18n.Text("No documents may be open when setting library locations."), I18n.Text("Canceled!"), JOptionPane.INFORMATION_MESSAGE);
                        return;
                    }
                }
            }
        }

        // Remove all library watchers
        LibraryWatcher.INSTANCE.watchDirs(new HashSet<>());

        // Ask the user to make changes
        LibraryLocationsPanel panel    = new LibraryLocationsPanel();
        JScrollPane           scroller = new JScrollPane(panel);
        int                   result   = WindowUtils.showOptionDialog(Workspace.get(), scroller, ChangeLibraryLocationsCommand.INSTANCE.getTitle(), true, JOptionPane.OK_CANCEL_OPTION, JOptionPane.PLAIN_MESSAGE, null, new JButton[]{panel.mApplyButton, panel.mCancelButton}, panel.mCancelButton);
        if (result == JOptionPane.OK_OPTION) {
            Library.LIBRARIES.clear();
            for (LibraryFields fields : panel.mFields) {
                switch (fields.getLibraryType()) {
                case MASTER:
                    Library.MASTER.setPath(fields.getPath());
                    Library.LIBRARIES.add(Library.MASTER);
                    break;
                case USER:
                    Library.USER.setPath(fields.getPath());
                    Library.LIBRARIES.add(Library.USER);
                    break;
                default:
                    Library.LIBRARIES.add(fields.createLibrary());
                    break;
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
        if (result == JOptionPane.OK_OPTION) {
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
        createAddButton();
        mApplyButton = createDialogButton(I18n.Text("Apply"));
        mCancelButton = createDialogButton(I18n.Text("Cancel"));
        mFields.get(0).contentsChanged();
    }

    public List<LibraryFields> getFields() {
        return mFields;
    }

    private void createAddButton() {
        JButton button = new JButton(I18n.Text("Add"));
        button.addActionListener(e -> {
            remove(button);
            mFields.add(new LibraryFields(this, "", "", "", "", LibraryFields.LibraryType.EXTRA));
            add(button);
            mFields.get(0).contentsChanged();
            revalidate();
            repaint();
            EventQueue.invokeLater(() -> {
                scrollRectToVisible(button.getBounds());
                mFields.get(mFields.size() - 1).getTitleField().requestFocus();
            });
        });
        add(button);
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

    public void setApplyState(boolean enabled) {
        mApplyButton.setEnabled(enabled);
    }
}
