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
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.NewerDataFileVersionException;
import com.trollworks.gcs.utility.PathUtils;

import java.awt.Component;
import java.awt.Dialog;
import java.awt.Frame;
import java.io.File;
import java.nio.file.Path;
import java.text.MessageFormat;
import java.util.List;
import javax.swing.JComponent;
import javax.swing.JFileChooser;
import javax.swing.filechooser.FileFilter;
import javax.swing.filechooser.FileNameExtensionFilter;

/** Provides standard file dialog handling. */
public class StdFileDialog {
    /**
     * Creates a new {@link StdFileDialog}.
     *
     * @param comp    The parent {@link Component} of the dialog. May be {@code null}.
     * @param title   The title to use. May be {@code null}.
     * @param filters The file filters to make available. If there are none, then the {@code
     *                showAllFilter} flag will be forced to {@code true}.
     * @return The chosen {@link Path} or {@code null}.
     */
    public static Path showOpenDialog(Component comp, String title, List<FileNameExtensionFilter> filters) {
        return showOpenDialog(comp, title, filters != null ? filters.toArray(new FileNameExtensionFilter[0]) : null);
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
        return showOpenDialog(comp, title, null, filters);
    }

    /**
     * Creates a new {@link StdFileDialog}.
     *
     * @param comp           The parent {@link Component} of the dialog. May be {@code null}.
     * @param title          The title to use. May be {@code null}.
     * @param accessoryPanel An extra panel to show. May be {@code null}.
     * @param filters        The file filters to make available. If there are none, then the {@code
     *                       showAllFilter} flag will be forced to {@code true}.
     * @return The chosen {@link Path} or {@code null}.
     */
    public static Path showOpenDialog(Component comp, String title, JComponent accessoryPanel, FileNameExtensionFilter... filters) {
        Preferences  prefs  = Preferences.getInstance();
        JFileChooser dialog = new JFileChooser(prefs.getLastDir().toFile());
        dialog.setDialogTitle(title);
        if (filters != null && filters.length > 0) {
            dialog.setAcceptAllFileFilterUsed(false);
            for (FileNameExtensionFilter filter : filters) {
                dialog.addChoosableFileFilter(filter);
            }
        } else {
            dialog.setAcceptAllFileFilterUsed(true);
        }
        if (accessoryPanel != null) {
            dialog.setAccessory(accessoryPanel);
        }
        int result = dialog.showOpenDialog(comp);
        if (result != JFileChooser.ERROR_OPTION) {
            File current = dialog.getCurrentDirectory();
            if (current != null) {
                prefs.setLastDir(current.toPath());
            }
        }
        if (result == JFileChooser.APPROVE_OPTION) {
            Path path = dialog.getSelectedFile().toPath().normalize().toAbsolutePath();
            prefs.addRecentFile(path);
            return path;
        }
        return null;
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
    public static Path showSaveDialog(Component comp, String title, Path suggestedFile, List<FileNameExtensionFilter> filters) {
        return showSaveDialog(comp, title, suggestedFile, filters != null ? filters.toArray(new FileNameExtensionFilter[0]) : null);
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
        return showSaveDialog(comp, title, suggestedFile, null, filters);
    }

    /**
     * Creates a new {@link StdFileDialog}.
     *
     * @param comp           The parent {@link Component} of the dialog. May be {@code null}.
     * @param title          The title to use. May be {@code null}.
     * @param suggestedFile  The suggested file to save as. May be {@code null}.
     * @param accessoryPanel An extra panel to show. May be {@code null}.
     * @param filters        The file filters to make available. If there are none, then the {@code
     *                       showAllFilter} flag will be forced to {@code true}.
     * @return The chosen {@link Path} or {@code null}.
     */
    public static Path showSaveDialog(Component comp, String title, Path suggestedFile, JComponent accessoryPanel, FileNameExtensionFilter... filters) {
        Preferences prefs = Preferences.getInstance();
        Path        path  = suggestedFile != null ? suggestedFile.getParent() : null;
        if (path == null) {
            path = prefs.getLastDir();
        }
        JFileChooser dialog = new JFileChooser(path.toFile());
        dialog.setDialogTitle(title);
        if (filters != null && filters.length > 0) {
            dialog.setAcceptAllFileFilterUsed(false);
            for (FileNameExtensionFilter filter : filters) {
                dialog.addChoosableFileFilter(filter);
            }
        } else {
            dialog.setAcceptAllFileFilterUsed(true);
        }
        if (suggestedFile != null) {
            dialog.setSelectedFile(suggestedFile.toFile());
        }
        if (accessoryPanel != null) {
            dialog.setAccessory(accessoryPanel);
        }
        int result = dialog.showSaveDialog(comp);
        if (result != JFileChooser.ERROR_OPTION) {
            File current = dialog.getCurrentDirectory();
            if (current != null) {
                prefs.setLastDir(current.toPath().normalize().toAbsolutePath());
            }
        }
        if (result == JFileChooser.APPROVE_OPTION) {
            File file = dialog.getSelectedFile();
            if (!PathUtils.isNameValidForFile(file.getName())) {
                WindowUtils.showError(comp, I18n.Text("Invalid file name"));
                return null;
            }
            if (filters != null) {
                FileFilter fileFilter = dialog.getFileFilter();
                if (!fileFilter.accept(file)) {
                    for (FileNameExtensionFilter filter : filters) {
                        if (filter == fileFilter) {
                            file = new File(file.getParentFile(), PathUtils.enforceExtension(file.getName(), filter.getExtensions()[0]));
                            break;
                        }
                    }
                }
            }
            path = file.toPath().normalize().toAbsolutePath();
            prefs.addRecentFile(path);
            return path;
        }
        return null;
    }

    /**
     * @param comp      The {@link Component} to use for determining the parent {@link Frame} or
     *                  {@link Dialog}.
     * @param name      The name of the file that cannot be opened.
     * @param throwable The {@link Throwable}, if any, that caused the failure.
     */
    public static void showCannotOpenMsg(Component comp, String name, Throwable throwable) {
        if (throwable instanceof NewerDataFileVersionException) {
            WindowUtils.showError(comp, MessageFormat.format(I18n.Text("Unable to open \"{0}\"\n{1}"), name, throwable.getMessage()));
        } else {
            if (throwable != null) {
                Log.error(throwable);
            }
            WindowUtils.showError(comp, MessageFormat.format(I18n.Text("Unable to open \"{0}\"."), name));
        }
    }
}
