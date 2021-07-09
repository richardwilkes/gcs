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

import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.FontIconButton;
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
import com.trollworks.gcs.utility.NamedData;

import java.awt.BorderLayout;
import java.awt.EventQueue;
import java.awt.Point;
import java.awt.event.WindowEvent;
import java.io.IOException;
import java.nio.file.Path;
import java.util.List;
import javax.swing.border.CompoundBorder;
import javax.swing.filechooser.FileNameExtensionFilter;

public abstract class SettingsWindow<T> extends BaseWindow implements CloseHandler {
    private FontIconButton mResetButton;
    private FontIconButton mMenuButton;
    private ScrollPanel    mScroller;

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
        getContentPane().add(toolbar, BorderLayout.NORTH);
    }

    protected void addToToolBar(Toolbar toolbar) {
        addResetButton(toolbar);
        addActionMenu(toolbar);
    }

    protected final void addResetButton(Toolbar toolbar) {
        mResetButton = new FontIconButton(getResetButtonGlyph(), getResetButtonTooltip(), (b) -> reset());
        toolbar.add(mResetButton, Toolbar.LAYOUT_EXTRA_BEFORE);
    }

    protected final void addActionMenu(Toolbar toolbar) {
        mMenuButton = new FontIconButton(FontAwesome.BARS, I18n.text("Menu"), (b) -> {
            Menu menu = new Menu();
            populateActionMenu(menu);
            menu.presentToUser(mMenuButton, 0, mMenuButton::updateRollOver);
        });
        toolbar.add(mMenuButton);
    }

    protected void populateActionMenu(Menu menu) {
        menu.addItem(new MenuItem(I18n.text("Import…"), this::importSettings));
        menu.addItem(new MenuItem(I18n.text("Export…"), this::exportSettings));
        for (NamedData<List<NamedData<T>>> oneDir : NamedData.scanLibraries(getFileType(), getDir(),
                this::createSettingsFrom)) {
            menu.addSeparator();
            MenuItem item = new MenuItem(oneDir.getName(), null);
            item.setEnabled(false);
            menu.addItem(item);
            for (NamedData<T> choice : oneDir.getData()) {
                menu.add(new MenuItem(choice.toString(), (p) -> resetTo(choice.getData())));
            }
        }
    }

    protected String getResetButtonGlyph() {
        return FontAwesome.POWER_OFF;
    }

    protected String getResetButtonTooltip() {
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
        adjustResetButton();
        scrollToTop();
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
                String msg = String.format(I18n.text("Unable to import %s."),
                        filter.getDescription());
                String exMsg = ioe.getMessage();
                if (exMsg != null && !exMsg.isBlank()) {
                    msg += "\n\n" + exMsg;
                }
                Modal.showError(this, msg);
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
}
