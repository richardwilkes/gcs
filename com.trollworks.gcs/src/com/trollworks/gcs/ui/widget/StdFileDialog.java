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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;

import java.awt.Component;
import java.awt.Dialog;
import java.awt.FileDialog;
import java.awt.Frame;
import java.awt.Window;
import java.io.File;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.text.MessageFormat;
import javax.swing.filechooser.FileNameExtensionFilter;

/** Provides standard file dialog handling. */
public final class StdFileDialog {
    private StdFileDialog() {
    }

    /**
     * Creates a new {@link StdFileDialog}.
     *
     * @param comp    The parent {@link Component} of the dialog. May be {@code null}.
     * @param title   The title to use. May be {@code null}.
     * @param filters The file filters to make available. If there are none, then the {@code
     *                showAllFilter} flag will be forced to {@code true}.
     * @return The chosen {@link Path} or {@code null}.
     */
    public static Path showOpenDialog(Component comp, String title, FileNameExtensionFilter... filters) {
        FileDialog dialog = createFileDialog(comp, title, FileDialog.LOAD, null, filters);
        Path       path   = show(dialog);
        if (path != null) {
            Preferences.getInstance().addRecentFile(path);
        }
        return path;
    }

    /**
     * Creates a new {@link StdFileDialog}.
     *
     * @param comp          The parent {@link Component} of the dialog. May be {@code null}.
     * @param title         The title to use. May be {@code null}.
     * @param suggestedFile The suggested file to save as. May be {@code null}.
     * @param filters       The file filters to make available. If there are none, then the {@code
     *                      showAllFilter} flag will be forced to {@code true}.
     * @return The chosen {@link Path} or {@code null}.
     */
    public static Path showSaveDialog(Component comp, String title, Path suggestedFile, FileNameExtensionFilter... filters) {
        FileDialog dialog = createFileDialog(comp, title, FileDialog.SAVE, suggestedFile != null ? suggestedFile.getParent() : null, filters);
        if (suggestedFile != null) {
            dialog.setFile(PathUtils.getLeafName(suggestedFile, true));
        }
        Path path = show(dialog);
        if (path != null) {
            if (!PathUtils.isNameValidForFile(PathUtils.getLeafName(path, true))) {
                WindowUtils.showError(comp, I18n.Text("Invalid file name"));
                return null;
            }
            if (filters != null) {
                FileNameExtensionFilter filter = filters[0];
                if (!filter.accept(path.toFile())) {
                    path = path.resolveSibling(PathUtils.enforceExtension(PathUtils.getLeafName(path, true), filter.getExtensions()[0]));
                }
            }
            Preferences.getInstance().addRecentFile(path);
        }
        return path;
    }

    /**
     * @param comp      The {@link Component} to use for determining the parent {@link Frame} or
     *                  {@link Dialog}.
     * @param name      The name of the file that cannot be opened.
     * @param throwable The {@link Throwable}, if any, that caused the failure.
     */
    public static void showCannotOpenMsg(Component comp, String name, Throwable throwable) {
        if (throwable != null) {
            Log.error(throwable);
            WindowUtils.showError(comp, MessageFormat.format(I18n.Text("Unable to open \"{0}\"\n{1}"), name, throwable.getMessage()));
        } else {
            WindowUtils.showError(comp, MessageFormat.format(I18n.Text("Unable to open \"{0}\"."), name));
        }
    }

    private static FileDialog createFileDialog(Component comp, String title, int mode, Path startingDir, FileNameExtensionFilter... filters) {
        FileDialog dialog;
        Window     window = UIUtilities.getSelfOrAncestorOfType(comp, Window.class);
        if (window instanceof Frame) {
            dialog = new FileDialog((Frame) window, title, mode);
        } else {
            dialog = new FileDialog((Dialog) window, title, mode);
        }
        if (startingDir == null) {
            startingDir = Preferences.getInstance().getLastDir();
        }
        dialog.setDirectory(startingDir.toString());
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

    private static Path show(FileDialog dialog) {
        dialog.setVisible(true);
        String dir = dialog.getDirectory();
        if (dir == null) {
            return null;
        }
        Preferences.getInstance().setLastDir(Paths.get(dir).normalize().toAbsolutePath());
        String file = dialog.getFile();
        if (file == null) {
            return null;
        }
        return Paths.get(dir, file).normalize().toAbsolutePath();
    }
}
