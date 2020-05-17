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

import com.trollworks.gcs.datafile.DataFileDockable;
import com.trollworks.gcs.io.Log;
import com.trollworks.gcs.io.RecursiveDirectoryRemover;
import com.trollworks.gcs.io.UrlUtils;
import com.trollworks.gcs.io.json.Json;
import com.trollworks.gcs.io.json.JsonArray;
import com.trollworks.gcs.io.json.JsonMap;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.ui.widget.Workspace;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Preferences;
import com.trollworks.gcs.utility.Version;
import com.trollworks.gcs.utility.task.Tasks;

import java.awt.BorderLayout;
import java.awt.EventQueue;
import java.awt.GraphicsEnvironment;
import java.io.BufferedInputStream;
import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;
import javax.swing.JComponent;
import javax.swing.JDialog;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JProgressBar;
import javax.swing.WindowConstants;

public class Library implements Runnable {
    private static final String  MODULE          = "Libraries";
    private static final String  MASTER_PATH_KEY = "MasterLibraryPath";
    private static final String  USER_PATH_KEY   = "UserLibraryPath";
    private static final String  SHA_PREFIX      = "\"sha\": \"";
    private static final String  SHA_SUFFIX      = "\",";
    private static final String  ROOT_PREFIX     = "richardwilkes-gcs_library-";
    private static final String  VERSION_FILE    = "version.txt";
    private              String  mResult;
    private              JDialog mDialog;
    private              boolean mUpdateComplete;

    public static Path getDefaultMasterRootPath() {
        return Paths.get(System.getProperty("user.home", "."), "GCS", "Master Library").normalize();
    }

    public static Path getDefaultUserRootPath() {
        return Paths.get(System.getProperty("user.home", "."), "GCS", "User Library").normalize();
    }

    /** @return The path to the master GCS library files. */
    public static Path getMasterRootPath() {
        Path path = Paths.get(Preferences.getInstance().getStringValue(MODULE, MASTER_PATH_KEY, getDefaultMasterRootPath().toString())).normalize();
        if (!Files.exists(path)) {
            try {
                Files.createDirectories(path);
            } catch (IOException exception) {
                Log.error(exception);
            }
        }
        return path;
    }

    public static void setMasterRootPath(Path path) {
        Preferences prefs = Preferences.getInstance();
        path = path.toAbsolutePath().normalize();
        if (path.equals(getDefaultMasterRootPath())) {
            prefs.removePreference(MODULE, MASTER_PATH_KEY);
        } else {
            prefs.setValue(MODULE, MASTER_PATH_KEY, path.toString());
        }
        prefs.save();
    }

    /** @return The path to the user GCS library files. */
    public static Path getUserRootPath() {
        Path path = Paths.get(Preferences.getInstance().getStringValue(MODULE, USER_PATH_KEY, getDefaultUserRootPath().toString())).normalize();
        if (!Files.exists(path)) {
            try {
                Files.createDirectories(path);
            } catch (IOException exception) {
                Log.error(exception);
            }
        }
        return path;
    }

    public static void setUserRootPath(Path path) {
        Preferences prefs = Preferences.getInstance();
        path = path.toAbsolutePath().normalize();
        if (path.equals(getDefaultUserRootPath())) {
            prefs.removePreference(MODULE, USER_PATH_KEY);
        } else {
            prefs.setValue(MODULE, USER_PATH_KEY, path.toString());
        }
        prefs.save();
    }

    public static final String getRecordedCommit() {
        try (BufferedReader in = Files.newBufferedReader(getMasterRootPath().resolve(VERSION_FILE))) {
            String line = in.readLine();
            while (line != null) {
                line = line.trim();
                if (!line.isEmpty()) {
                    return line;
                }
                line = in.readLine();
            }
        } catch (IOException exception) {
            // Ignore
        }
        return "";
    }

    public static final String getLatestCommit() {
        String sha = "";
        try {
            JsonArray array = Json.asArray(Json.parse(new URL("https://api.github.com/repos/richardwilkes/gcs_library/commits?per_page=1")), false);
            JsonMap   map   = array.getMap(0, false);
            sha = map.getString("sha", false);
            if (sha.length() > 7) {
                sha = sha.substring(0, 7);
            }
        } catch (IOException exception) {
            Log.error(exception);
        }
        return sha;
    }

    public static final long getMinimumGCSVersion() {
        String version = "";
        try (BufferedReader in = new BufferedReader(new InputStreamReader(UrlUtils.setupConnection("https://raw.githubusercontent.com/richardwilkes/gcs_library/master/minimum_version.txt").getInputStream(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = in.readLine()) != null) {
                if (version.isBlank()) {
                    line = line.trim();
                    if (!line.isBlank()) {
                        version = line;
                    }
                }
            }
        } catch (IOException exception) {
            Log.error(exception);
        }
        return Version.extract(version, 0);
    }

    public static final void downloadIfNotPresent() {
        if (getRecordedCommit().isBlank()) {
            download();
        }
    }

    public static final void download() {
        Library lib = new Library();
        if (GraphicsEnvironment.isHeadless()) {
            Tasks.callOnBackgroundThread(lib);
            synchronized (lib) {
                while (!lib.mUpdateComplete) {
                    try {
                        //noinspection WaitOrAwaitWithoutTimeout
                        lib.wait();
                    } catch (InterruptedException exception) {
                        break;
                    }
                }
            }
        } else {
            // Close any open files that come from the master library
            Workspace workspace = Workspace.get();
            String    prefix    = getMasterRootPath().toAbsolutePath().toString();
            for (Dockable dockable : workspace.getDock().getDockables()) {
                if (dockable instanceof DataFileDockable) {
                    DataFileDockable dfd  = (DataFileDockable) dockable;
                    File             file = dfd.getBackingFile();
                    if (file != null) {
                        if (file.getAbsolutePath().startsWith(prefix)) {
                            if (dfd.mayAttemptClose()) {
                                if (!dfd.attemptClose()) {
                                    JOptionPane.showMessageDialog(null, I18n.Text("GCS Master Library update was canceled."), I18n.Text("Canceled!"), JOptionPane.INFORMATION_MESSAGE);
                                    return;
                                }
                            }
                        }
                    }
                }
            }

            // Put up a progress dialog
            JDialog dialog = new JDialog(workspace, I18n.Text("Update Master Library"), true);
            dialog.setResizable(false);
            dialog.setDefaultCloseOperation(WindowConstants.DO_NOTHING_ON_CLOSE);
            dialog.setUndecorated(true);
            JComponent content = (JComponent) dialog.getContentPane();
            content.setLayout(new BorderLayout());
            content.setBorder(new EmptyBorder(10));
            content.add(new JLabel(I18n.Text("Downloading and installing the Master Library…")), BorderLayout.NORTH);
            JProgressBar bar = new JProgressBar();
            bar.setIndeterminate(true);
            content.add(bar);
            dialog.pack();
            dialog.setLocationRelativeTo(workspace);
            StdMenuBar.SUPRESS_MENUS = true;
            lib.mDialog = dialog;
            Tasks.callOnBackgroundThread(lib);
            dialog.setVisible(true);
        }
    }

    private Library() {
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
        try {
            Path    root           = getMasterRootPath();
            boolean shouldContinue = true;
            Path    saveRoot       = root.resolveSibling(root.getFileName().toString() + ".save");
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
                getMasterRootPath(); // will recreate the dir
                try (ZipInputStream in = new ZipInputStream(new BufferedInputStream(UrlUtils.setupConnection("https://api.github.com/repos/richardwilkes/gcs_library/zipball/master").getInputStream()))) {
                    byte[]   buffer = new byte[8192];
                    ZipEntry entry;
                    String   sha    = "unknown";
                    while ((entry = in.getNextEntry()) != null) {
                        if (entry.isDirectory()) {
                            continue;
                        }
                        Path entryPath = Paths.get(entry.getName());
                        int  nameCount = entryPath.getNameCount();
                        if (nameCount < 3 || !entryPath.getName(0).toString().startsWith(ROOT_PREFIX) || !"Library".equals(entryPath.getName(1).toString())) {
                            continue;
                        }
                        long size = entry.getSize();
                        if (size < 1) {
                            continue;
                        }
                        sha = entryPath.getName(0).toString().substring(ROOT_PREFIX.length());
                        entryPath = entryPath.subpath(2, nameCount);
                        Path path = root.resolve(entryPath);
                        Files.createDirectories(path.getParent());
                        try (OutputStream out = Files.newOutputStream(path)) {
                            while (size > 0) {
                                int amt = in.read(buffer);
                                if (amt < 0) {
                                    break;
                                }
                                if (amt > 0) {
                                    size -= amt;
                                    out.write(buffer, 0, amt);
                                }
                            }
                        }
                    }
                    if (sha.length() > 7) {
                        sha = sha.substring(0, 7);
                    }
                    Files.writeString(root.resolve(VERSION_FILE), sha + "\n");
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
                        getMasterRootPath(); // will recreate the dir
                    }
                }
            }
        } catch (Throwable throwable) {
            Log.error(throwable);
            if (mResult == null) {
                mResult = throwable.getMessage();
                if (mResult == null) {
                    mResult = "exception";
                }
            }
        }

        synchronized (this) {
            mUpdateComplete = true;
            notifyAll();
        }
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
        mDialog.dispose();
        StdMenuBar.SUPRESS_MENUS = false;
        if (mResult == null) {
            JOptionPane.showMessageDialog(null, I18n.Text("GCS Master Library update was successful."), I18n.Text("Success!"), JOptionPane.INFORMATION_MESSAGE);
        } else {
            WindowUtils.showError(null, I18n.Text("An error occurred while trying to update the GCS Master Library:") + "\n\n" + mResult);
        }
    }
}
