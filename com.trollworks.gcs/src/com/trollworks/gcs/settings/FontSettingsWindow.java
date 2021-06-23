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
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.FontPanel;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.LayoutConstants;
import com.trollworks.gcs.ui.widget.Menu;
import com.trollworks.gcs.ui.widget.MenuItem;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.awt.Font;
import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/** A window for editing font settings. */
public final class FontSettingsWindow extends SettingsWindow {
    private List<FontTracker> mFontPanels;
    private boolean           mIgnore;

    public FontSettingsWindow() {
        super(I18n.text("Font Settings"));
    }

    @Override
    protected Panel createContent() {
        Panel panel = new Panel(new PrecisionLayout().setColumns(2).
                setMargins(LayoutConstants.WINDOW_BORDER_INSET), false);
        mFontPanels = new ArrayList<>();
        for (ThemeFont font : Fonts.ALL) {
            if (font.isEditable()) {
                FontTracker tracker = new FontTracker(font);
                panel.add(new Label(font.toString()), new PrecisionLayoutData().setFillHorizontalAlignment());
                panel.add(tracker);
                mFontPanels.add(tracker);
            }
        }
        return panel;
    }

    @Override
    protected boolean shouldResetBeEnabled() {
        for (ThemeFont font : Fonts.ALL) {
            if (font.isEditable() && !font.getFont().equals(Fonts.defaultThemeFonts().getFont(font.getIndex()))) {
                return true;
            }
        }
        return false;
    }

    @Override
    protected void reset() {
        resetTo(Fonts.defaultThemeFonts());
    }

    private void resetTo(Fonts fonts) {
        mIgnore = true;
        for (FontTracker tracker : mFontPanels) {
            tracker.resetTo(fonts);
        }
        mIgnore = false;
        BaseWindow.forceRevalidateAndRepaint();
        adjustResetButton();
    }

    @Override
    protected Menu createActionMenu() {
        Menu menu = new Menu();
        menu.addItem(new MenuItem(I18n.text("Import…"), (p) -> {
            Path path = Modal.presentOpenFileDialog(this, I18n.text("Import…"),
                    Dirs.THEME, FileType.FONT_SETTINGS.getFilter());
            if (path != null) {
                try {
                    resetTo(new Fonts(path));
                } catch (IOException ioe) {
                    Log.error(ioe);
                    Modal.showError(this, I18n.text("Unable to import font settings."));
                }
            }
        }));
        menu.addItem(new MenuItem(I18n.text("Export…"), (p) -> {
            Path path = Modal.presentSaveFileDialog(this, I18n.text("Export…"), Dirs.THEME,
                    FileType.FONT_SETTINGS.getUntitledDefaultFileName(),
                    FileType.FONT_SETTINGS.getFilter());
            if (path != null) {
                try {
                    Fonts.currentThemeFonts().save(path);
                } catch (Exception exception) {
                    Log.error(exception);
                    Modal.showError(this, I18n.text("Unable to export font settings."));
                }
            }
        }));
        Settings.getInstance(); // Just to ensure the libraries list is initialized
        for (Library lib : Library.LIBRARIES) {
            Path dir = lib.getPath().resolve(Dirs.THEME.getDefaultPath().getFileName());
            if (Files.isDirectory(dir)) {
                List<FontSetData> list = new ArrayList<>();
                // IMPORTANT: On Windows, calling any of the older methods to list the contents of a
                // directory results in leaving state around that prevents future move & delete
                // operations. Only use this style of access for directory listings to avoid that.
                try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                    for (Path path : stream) {
                        if (FileType.FONT_SETTINGS.matchExtension(PathUtils.getExtension(path))) {
                            try {
                                list.add(new FontSetData(PathUtils.getLeafName(path, false), new Fonts(path)));
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
                    for (FontSetData choice : list) {
                        menu.add(new MenuItem(choice.toString(),
                                (p) -> resetTo(choice.mFonts)));
                    }
                }
            }
        }
        return menu;
    }

    private static class FontSetData implements Comparable<FontSetData> {
        String mName;
        Fonts  mFonts;

        FontSetData(String name, Fonts fonts) {
            mName = name;
            mFonts = fonts;
        }

        @Override
        public int compareTo(FontSetData other) {
            return NumericComparator.CASELESS_COMPARATOR.compare(mName, other.mName);
        }

        @Override
        public String toString() {
            return mName;
        }
    }

    private class FontTracker extends FontPanel {
        private int mIndex;

        FontTracker(ThemeFont font) {
            super(font.getFont());
            mIndex = font.getIndex();
            addActionListener((evt) -> {
                if (!mIgnore) {
                    Fonts.currentThemeFonts().setFont(mIndex, getCurrentFont());
                    adjustResetButton();
                    BaseWindow.forceRevalidateAndRepaint();
                }
            });
        }

        void resetTo(Fonts fonts) {
            Font font = fonts.getFont(mIndex);
            setCurrentFont(font);
            Fonts.currentThemeFonts().setFont(mIndex, font);
        }
    }
}
