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
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.ColorWell;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.LayoutConstants;
import com.trollworks.gcs.ui.widget.Menu;
import com.trollworks.gcs.ui.widget.MenuItem;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.awt.Color;
import java.awt.Container;
import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/** A window for editing color settings. */
public final class ColorSettingsWindow extends SettingsWindow {
    private List<ColorTracker> mColorWells;
    private boolean            mIgnore;

    public ColorSettingsWindow() {
        super(I18n.text("Color Settings"));
    }

    protected Panel createContent() {
        int cols = 8;
        Panel panel = new Panel(new PrecisionLayout().setColumns(cols).
                setMargins(LayoutConstants.WINDOW_BORDER_INSET), false);
        mColorWells = new ArrayList<>();
        int max = Colors.ALL.size();
        cols /= 2;
        int maxPerCol  = max / cols;
        int excess     = max % (maxPerCol * cols);
        int iterations = maxPerCol;
        if (excess != 0) {
            iterations++;
        }
        for (int i = 0; i < iterations; i++) {
            addColorTracker(panel, Colors.ALL.get(i), 0);
            int index = i;
            for (int j = 1; j < ((i == maxPerCol) ? excess : cols); j++) {
                index += maxPerCol;
                if (j - 1 < excess) {
                    index++;
                }
                if (index < max) {
                    addColorTracker(panel, Colors.ALL.get(index), 8);
                }
            }
        }
        return panel;
    }

    private void addColorTracker(Container parent, ThemeColor color, int leftMargin) {
        ColorTracker tracker = new ColorTracker(color);
        mColorWells.add(tracker);
        parent.add(new Label(color.toString()), new PrecisionLayoutData().
                setFillHorizontalAlignment().setLeftMargin(leftMargin));
        parent.add(tracker, new PrecisionLayoutData().setLeftMargin(4));
    }

    @Override
    protected boolean shouldResetBeEnabled() {
        for (ThemeColor color : Colors.ALL) {
            if (color.getRGB() != Colors.defaultThemeColors().getColor(color.getIndex()).getRGB()) {
                return true;
            }
        }
        return false;
    }

    @Override
    protected void reset() {
        resetTo(Colors.defaultThemeColors());
    }

    private void resetTo(Colors colors) {
        mIgnore = true;
        for (ColorTracker tracker : mColorWells) {
            tracker.resetTo(colors);
        }
        mIgnore = false;
        WindowUtils.repaintAll();
        adjustResetButton();
    }

    @Override
    protected Menu createActionMenu() {
        Menu menu = new Menu();
        menu.addItem(new MenuItem(I18n.text("Import…"), (p) -> {
            Path path = Modal.presentOpenFileDialog(this, I18n.text("Import…"),
                    Dirs.THEME, FileType.COLOR_SETTINGS.getFilter());
            if (path != null) {
                try {
                    resetTo(new Colors(path));
                } catch (IOException ioe) {
                    Log.error(ioe);
                    Modal.showError(this, I18n.text("Unable to import color settings."));
                }
            }
        }));
        menu.addItem(new MenuItem(I18n.text("Export…"), (p) -> {
            Path path = Modal.presentSaveFileDialog(this, I18n.text("Export…"), Dirs.THEME,
                    FileType.COLOR_SETTINGS.getUntitledDefaultFileName(),
                    FileType.COLOR_SETTINGS.getFilter());
            if (path != null) {
                try {
                    Colors.currentThemeColors().save(path);
                } catch (Exception exception) {
                    Log.error(exception);
                    Modal.showError(this, I18n.text("Unable to export color settings."));
                }
            }
        }));
        Settings.getInstance(); // Just to ensure the libraries list is initialized
        for (Library lib : Library.LIBRARIES) {
            Path dir = lib.getPath().resolve(Dirs.THEME.getDefaultPath().getFileName());
            if (Files.isDirectory(dir)) {
                List<ColorSetData> list = new ArrayList<>();
                // IMPORTANT: On Windows, calling any of the older methods to list the contents of a
                // directory results in leaving state around that prevents future move & delete
                // operations. Only use this style of access for directory listings to avoid that.
                try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                    for (Path path : stream) {
                        if (FileType.COLOR_SETTINGS.matchExtension(PathUtils.getExtension(path))) {
                            try {
                                list.add(new ColorSetData(PathUtils.getLeafName(path, false), new Colors(path)));
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
                    for (ColorSetData choice : list) {
                        menu.add(new MenuItem(choice.toString(),
                                (p) -> resetTo(choice.mColors)));
                    }
                }
            }
        }
        return menu;
    }

    private static class ColorSetData implements Comparable<ColorSetData> {
        String mName;
        Colors mColors;

        ColorSetData(String name, Colors colors) {
            mName = name;
            mColors = colors;
        }

        @Override
        public int compareTo(ColorSetData other) {
            return NumericComparator.CASELESS_COMPARATOR.compare(mName, other.mName);
        }

        @Override
        public String toString() {
            return mName;
        }
    }

    private class ColorTracker extends ColorWell implements ColorWell.ColorChangedListener {
        private int mIndex;

        ColorTracker(ThemeColor color) {
            super(new Color(color.getRGB(), true), null);
            mIndex = color.getIndex();
            setColorChangedListener(this);
        }

        void resetTo(Colors colors) {
            Color color = colors.getColor(mIndex);
            setWellColor(color);
            Colors.currentThemeColors().setColor(mIndex, color);
        }

        @Override
        public void colorChanged(Color color) {
            if (!mIgnore) {
                Colors.currentThemeColors().setColor(mIndex, color);
                adjustResetButton();
                WindowUtils.repaintAll();
            }
        }
    }
}
