/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.io.Log;
import com.trollworks.gcs.menu.file.RecentFilesMenu;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.NewerDataFileVersionException;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Preferences;

import java.awt.Component;
import java.awt.Dialog;
import java.awt.Frame;
import java.io.File;
import java.text.MessageFormat;
import java.util.List;
import javax.swing.JComponent;
import javax.swing.JFileChooser;
import javax.swing.filechooser.FileFilter;
import javax.swing.filechooser.FileNameExtensionFilter;

/** Provides standard file dialog handling. */
public class StdFileDialog {
    private static final String MODULE   = "StdFileDialog";
    private static final String LAST_DIR = "LastDir";

    /** @return The last directory used by the StdFileDialog. May return {@code null}. */
    public static final String getLastDir() {
        String last = Preferences.getInstance().getStringValue(MODULE, LAST_DIR);
        if (last != null) {
            if (!new File(last).isDirectory()) {
                last = null;
            }
        }
        return last;
    }

    /** @param path The path to use as the last directory. */
    public static final void setLastDir(String path) {
        Preferences.getInstance().setValue(MODULE, LAST_DIR, path);
    }

    /**
     * Creates a new {@link StdFileDialog}.
     *
     * @param comp    The parent {@link Component} of the dialog. May be {@code null}.
     * @param title   The title to use. May be {@code null}.
     * @param filters The file filters to make available. If there are none, then the {@code
     *                showAllFilter} flag will be forced to {@code true}.
     * @return The chosen {@link File} or {@code null}.
     */
    public static File showOpenDialog(Component comp, String title, List<FileNameExtensionFilter> filters) {
        return showOpenDialog(comp, title, filters != null ? filters.toArray(new FileNameExtensionFilter[0]) : null);
    }

    /**
     * Creates a new {@link StdFileDialog}.
     *
     * @param comp    The parent {@link Component} of the dialog. May be {@code null}.
     * @param title   The title to use. May be {@code null}.
     * @param filters The file filters to make available. If there are none, then the {@code
     *                showAllFilter} flag will be forced to {@code true}.
     * @return The chosen {@link File} or {@code null}.
     */
    public static File showOpenDialog(Component comp, String title, FileNameExtensionFilter... filters) {
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
     * @return The chosen {@link File} or {@code null}.
     */
    public static File showOpenDialog(Component comp, String title, JComponent accessoryPanel, FileNameExtensionFilter... filters) {
        JFileChooser dialog = new JFileChooser(getLastDir());
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
                setLastDir(current.getAbsolutePath());
            }
        }
        if (result == JFileChooser.APPROVE_OPTION) {
            File file = dialog.getSelectedFile();
            RecentFilesMenu.addRecent(file);
            return file;
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
     * @return The chosen {@link File} or {@code null}.
     */
    public static File showSaveDialog(Component comp, String title, File suggestedFile, List<FileNameExtensionFilter> filters) {
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
     * @return The chosen {@link File} or {@code null}.
     */
    public static File showSaveDialog(Component comp, String title, File suggestedFile, FileNameExtensionFilter... filters) {
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
     * @return The chosen {@link File} or {@code null}.
     */
    public static File showSaveDialog(Component comp, String title, File suggestedFile, JComponent accessoryPanel, FileNameExtensionFilter... filters) {
        JFileChooser dialog = new JFileChooser(suggestedFile != null ? suggestedFile.getParent() : getLastDir());
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
            dialog.setSelectedFile(suggestedFile);
        }
        if (accessoryPanel != null) {
            dialog.setAccessory(accessoryPanel);
        }
        int result = dialog.showSaveDialog(comp);
        if (result != JFileChooser.ERROR_OPTION) {
            File current = dialog.getCurrentDirectory();
            if (current != null) {
                setLastDir(current.getAbsolutePath());
            }
        }
        if (result == JFileChooser.APPROVE_OPTION) {
            File file = dialog.getSelectedFile();
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
            RecentFilesMenu.addRecent(file);
            return file;
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
