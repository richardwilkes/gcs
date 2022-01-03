/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.library;

import com.trollworks.gcs.datafile.DataFileDockable;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.LayoutConstants;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ProgressBar;
import com.trollworks.gcs.ui.widget.Workspace;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.RecursiveDirectoryRemover;
import com.trollworks.gcs.utility.Release;

import java.awt.EventQueue;
import java.awt.GraphicsEnvironment;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.FutureTask;

public final class LibraryUpdater implements Runnable {
    private static final ExecutorService QUEUE = Executors.newSingleThreadExecutor();
    private              String          mResult;
    private              Modal           mModal;
    private              Library         mLibrary;
    private              Release         mRelease;
    private              boolean         mUpdateComplete;

    public static List<Object> collectFiles() {
        FutureTask<List<Object>> task = new FutureTask<>(() -> {
            Set<Path>    dirs = new HashSet<>();
            List<Object> list = new ArrayList<>();
            list.add("GCS");
            for (Library library : Library.LIBRARIES) {
                list.add(LibraryCollector.list(library.getTitle(), library.getPath(), dirs));
            }
            LibraryWatcher.INSTANCE.watchDirs(dirs);
            return list;
        });
        QUEUE.submit(task);
        try {
            return task.get();
        } catch (Exception exception) {
            Log.error(exception);
            List<Object> list   = new ArrayList<>();
            List<Object> master = new ArrayList<>();
            List<Object> user   = new ArrayList<>();
            master.add(Library.MASTER.getTitle());
            list.add(master);
            user.add(Library.USER.getTitle());
            list.add(user);
            return list;
        }
    }

    public static void download(Library library, Release release) {
        LibraryUpdater lib = new LibraryUpdater(library, release);
        if (GraphicsEnvironment.isHeadless()) {
            FutureTask<Object> task = new FutureTask<>(lib, null);
            QUEUE.submit(task);
            try {
                task.get();
            } catch (Exception exception) {
                Log.error(exception);
            }
        } else {
            // Close any open files that come from the library
            Workspace workspace = Workspace.get();
            Path      prefix    = library.getPath();
            String    title     = library.getTitle();
            for (Dockable dockable : workspace.getDock().getDockables()) {
                if (dockable instanceof DataFileDockable dfd) {
                    Path path = dfd.getBackingFile();
                    if (path != null && path.toAbsolutePath().startsWith(prefix)) {
                        if (dfd.mayAttemptClose()) {
                            if (!dfd.attemptClose()) {
                                Modal.showMessage(null, I18n.text("Canceled!"), MessageType.NONE, String.format(I18n.text("GCS %s update was canceled."), title));
                                return;
                            }
                        }
                    }
                }
            }

            // Put up a progress dialog
            Panel msgPanel = new Panel(new PrecisionLayout().setHorizontalAlignment(PrecisionLayoutAlignment.MIDDLE));
            msgPanel.add(new Label(String.format(I18n.text("Downloading and installing the %s…"), title)));
            msgPanel.add(new ProgressBar(0), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setTopMargin(LayoutConstants.TOOLBAR_VERTICAL_INSET));
            Modal modal = Modal.prepareToShowMessage(workspace, String.format(I18n.text("Update %s"), title), MessageType.NONE, msgPanel);
            lib.mModal = modal;
            QUEUE.submit(lib);
            modal.presentToUser();
        }
    }

    private LibraryUpdater(Library library, Release release) {
        mLibrary = library;
        mRelease = release;
    }

    @Override
    public void run() {
        if (mUpdateComplete) {
            doCleanup();
        } else {
            doDownload();
        }
    }

    private void doDownload() {
        if (mRelease.getVersion().isZero()) {
            mResult = "No releases available";
        } else {
            try {
                LibraryWatcher.INSTANCE.watchDirs(new HashSet<>());
                Path    root           = mLibrary.getPath();
                boolean shouldContinue = true;
                Path    saveRoot       = root.resolveSibling(root.getFileName() + ".save");
                if (Files.exists(root)) {
                    try {
                        Files.move(root, saveRoot);
                    } catch (IOException exception) {
                        shouldContinue = false;
                        Log.error(exception);
                        mResult = exception.getMessage();
                        if (mResult == null) {
                            mResult = "exception";
                        }
                    }
                }
                if (shouldContinue) {
                    try {
                        mLibrary.download(mRelease);
                    } catch (IOException exception) {
                        Log.error(exception);
                        mResult = exception.getMessage();
                        if (mResult == null) {
                            mResult = "exception";
                        }
                    }
                    if (mResult == null) {
                        if (Files.exists(saveRoot)) {
                            RecursiveDirectoryRemover.remove(saveRoot, true);
                        }
                    } else {
                        RecursiveDirectoryRemover.remove(root, true);
                        if (Files.exists(saveRoot)) {
                            try {
                                Files.move(saveRoot, root);
                            } catch (IOException exception) {
                                Log.error(exception);
                            }
                        } else {
                            mLibrary.getPath(); // will recreate the dir
                        }
                    }
                }
            } catch (Throwable throwable) {
                Log.error(throwable);
                if (mResult == null) {
                    Throwable t = throwable;
                    do {
                        mResult = t.getMessage();
                        t = t.getCause();
                    } while (mResult == null && t != null);
                    if (mResult == null) {
                        mResult = "exception";
                    }
                }
            }
        }
        mUpdateComplete = true;
        if (!GraphicsEnvironment.isHeadless()) {
            EventQueue.invokeLater(this);
        }
    }

    private void doCleanup() {
        // Refresh the library view and let the user know what happened
        LibraryExplorerDockable libraryDockable = LibraryExplorerDockable.get();
        if (libraryDockable != null) {
            libraryDockable.refresh();
        }
        mModal.dispose();
        String title = mLibrary.getTitle();
        if (mResult == null) {
            Modal.showMessage(null, I18n.text("Success!"), MessageType.NONE, String.format(I18n.text("%s update was successful."), title));
        } else {
            Modal.showError(null, String.format(I18n.text("An error occurred while trying to update the %s:\n\n"), title) + mResult);
        }
    }
}
