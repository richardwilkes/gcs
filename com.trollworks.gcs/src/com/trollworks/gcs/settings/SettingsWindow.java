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

package com.trollworks.gcs.settings;

import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.LayoutConstants;
import com.trollworks.gcs.ui.widget.Menu;
import com.trollworks.gcs.ui.widget.MenuItem;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.Toolbar;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.awt.BorderLayout;
import java.awt.EventQueue;
import java.awt.Point;
import java.awt.event.WindowEvent;
import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import javax.swing.border.CompoundBorder;
import javax.swing.filechooser.FileNameExtensionFilter;

public abstract class SettingsWindow<T> extends BaseWindow implements CloseHandler {
    private FontAwesomeButton mResetButton;
    private FontAwesomeButton mMenuButton;
    private ScrollPanel       mScroller;

    protected SettingsWindow(String title) {
        super(title);
    }

    protected final void fill() {
        addToolBar();
        mScroller = new ScrollPanel(createContent());
        getContentPane().add(mScroller, BorderLayout.CENTER);
        adjustResetButton();
        establishSizing();
        WindowUtils.packAndCenterWindowOn(this, null);
        scrollToTop();
    }

    public final void scrollToTop() {
        EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
    }

    private void addToolBar() {
        Toolbar toolbar = new Toolbar();
        toolbar.setBorder(new CompoundBorder(new LineBorder(Colors.DIVIDER, 0, 0, 1, 0),
                new EmptyBorder(LayoutConstants.TOOLBAR_VERTICAL_INSET,
                        LayoutConstants.WINDOW_BORDER_INSET, LayoutConstants.TOOLBAR_VERTICAL_INSET,
                        LayoutConstants.WINDOW_BORDER_INSET)));
        addToToolBar(toolbar);
        mResetButton = new FontAwesomeButton("\uf011", getResetButtonTitle(), this::reset);
        toolbar.add(mResetButton, Toolbar.LAYOUT_EXTRA_BEFORE);
        mMenuButton = new FontAwesomeButton("\uf0c9", I18n.text("Menu"),
                () -> createActionMenu().presentToUser(mMenuButton, 0, mMenuButton::updateRollOver));
        toolbar.add(mMenuButton);
        getContentPane().add(toolbar, BorderLayout.NORTH);
    }

    protected void addToToolBar(Toolbar toolbar) {
        // Here for sub-classes
    }

    protected String getResetButtonTitle() {
        return I18n.text("Reset to Factory Defaults");
    }

    protected abstract Panel createContent();

    protected final void adjustResetButton() {
        mResetButton.setEnabled(shouldResetBeEnabled());
    }

    protected abstract boolean shouldResetBeEnabled();

    public final void reset() {
        resetTo(getResetData());
    }

    protected abstract T getResetData();

    protected final void resetTo(T data) {
        doResetTo(data);
        transferFocus();
    }

    protected abstract void doResetTo(T data);

    protected abstract Dirs getDir();

    protected abstract FileType getFileType();

    private void importSettings(MenuItem item) {
        FileNameExtensionFilter filter = getFileType().getFilter();
        Path                    path   = Modal.presentOpenFileDialog(this, item.getText(), getDir(), filter);
        if (path != null) {
            try {
                resetTo(createSettingsFrom(path));
            } catch (IOException ioe) {
                Log.error(ioe);
                Modal.showError(this, String.format(I18n.text("Unable to import %s."),
                        filter.getDescription()));
            }
        }
    }

    protected abstract T createSettingsFrom(Path path) throws IOException;

    private void exportSettings(MenuItem item) {
        FileType                fileType = getFileType();
        FileNameExtensionFilter filter   = fileType.getFilter();
        Path path = Modal.presentSaveFileDialog(this, item.getText(), getDir(),
                fileType.getUntitledDefaultFileName(), filter);
        if (path != null) {
            try {
                exportSettingsTo(path);
            } catch (IOException exception) {
                Log.error(exception);
                Modal.showError(this, String.format(I18n.text("Unable to export %s."),
                        filter.getDescription()));
            }
        }
    }

    protected abstract void exportSettingsTo(Path path) throws IOException;

    private Menu createActionMenu() {
        Menu menu = new Menu();
        menu.addItem(new MenuItem(I18n.text("Import…"), this::importSettings));
        menu.addItem(new MenuItem(I18n.text("Export…"), this::exportSettings));
        Settings.getInstance(); // Just to ensure the libraries list is initialized
        FileType fileType = getFileType();
        Dirs     lastDir  = getDir();
        for (Library lib : Library.LIBRARIES) {
            Path dir = lib.getPath().resolve(lastDir.getDefaultPath().getFileName());
            if (Files.isDirectory(dir)) {
                List<SetData<T>> list = new ArrayList<>();
                // IMPORTANT: On Windows, calling any of the older methods to list the contents of a
                // directory results in leaving state around that prevents future move & delete
                // operations. Only use this style of access for directory listings to avoid that.
                try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                    for (Path path : stream) {
                        if (fileType.matchExtension(PathUtils.getExtension(path))) {
                            try {
                                list.add(new SetData<>(PathUtils.getLeafName(path, false), createSettingsFrom(path)));
                            } catch (IOException ioe) {
                                Log.error("unable to load " + path, ioe);
                            }
                        }
                    }
                } catch (IOException exception) {
                    Log.error(exception);
                }
                if (!list.isEmpty()) {
                    Collections.sort(list);
                    menu.addSeparator();
                    MenuItem item = new MenuItem(dir.getParent().getFileName().toString(), null);
                    item.setEnabled(false);
                    menu.addItem(item);
                    for (SetData<T> choice : list) {
                        menu.add(new MenuItem(choice.toString(), (p) -> resetTo(choice.mData)));
                    }
                }
            }
        }
        return menu;
    }

    @Override
    public final boolean mayAttemptClose() {
        return true;
    }

    @Override
    public final boolean attemptClose() {
        windowClosing(new WindowEvent(this, WindowEvent.WINDOW_CLOSING));
        return true;
    }

    @Override
    public final void dispose() {
        preDispose();
        super.dispose();
    }

    protected abstract void preDispose();

    private static class SetData<T> implements Comparable<SetData<T>> {
        String mName;
        T      mData;

        SetData(String name, T data) {
            mName = name;
            mData = data;
        }

        @Override
        public int compareTo(SetData other) {
            return NumericComparator.CASELESS_COMPARATOR.compare(mName, other.mName);
        }

        @Override
        public String toString() {
            return mName;
        }
    }
}
