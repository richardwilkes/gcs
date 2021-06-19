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

import com.trollworks.gcs.body.HitLocationTable;
import com.trollworks.gcs.body.HitLocationTablePanel;
import com.trollworks.gcs.body.LibraryHitLocationTables;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataChangeListener;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.BorderLayout;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Point;
import java.awt.Window;
import java.awt.event.WindowEvent;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import javax.swing.JMenuItem;
import javax.swing.JPopupMenu;

/** A window for editing hit location settings. */
public final class HitLocationSettingsWindow extends BaseWindow implements CloseHandler, DataChangeListener {
    private static final Map<UUID, HitLocationSettingsWindow> INSTANCES = new HashMap<>();
    private              GURPSCharacter                       mCharacter;
    private              HitLocationTablePanel                mLocationsPanel;
    private              FontAwesomeButton                    mResetButton;
    private              FontAwesomeButton                    mMenuButton;
    private              ScrollPanel                          mScroller;
    private              boolean                              mUpdatePending;

    /** Displays the hit location settings window. */
    public static void display(GURPSCharacter gchar) {
        if (!UIUtilities.inModalState()) {
            HitLocationSettingsWindow wnd;
            synchronized (INSTANCES) {
                UUID key = gchar == null ? null : gchar.getID();
                wnd = INSTANCES.get(key);
                if (wnd == null) {
                    wnd = new HitLocationSettingsWindow(gchar);
                    INSTANCES.put(key, wnd);
                }
            }
            wnd.setVisible(true);
        }
    }

    /** Closes the HitLocationSettingsWindow for the given character if it is open. */
    public static void closeFor(GURPSCharacter gchar) {
        for (Window window : Window.getWindows()) {
            if (window.isShowing() && window instanceof HitLocationSettingsWindow) {
                HitLocationSettingsWindow wnd = (HitLocationSettingsWindow) window;
                if (wnd.mCharacter == gchar) {
                    wnd.attemptClose();
                }
            }
        }
    }

    private static String createTitle(GURPSCharacter gchar) {
        return gchar == null ? I18n.text("Default Hit Locations") : String.format(I18n.text("Hit Locations for %s"), gchar.getProfile().getName());
    }

    private HitLocationSettingsWindow(GURPSCharacter gchar) {
        super(createTitle(gchar));
        mCharacter = gchar;
        Container content = getContentPane();
        Panel     header  = new Panel(new PrecisionLayout().setColumns(2).setMargins(5, 10, 5, 10).setHorizontalSpacing(10).setHorizontalAlignment(PrecisionLayoutAlignment.END));
        mResetButton = new FontAwesomeButton("\uf011", mCharacter == null ? I18n.text("Reset to Factory Defaults") : I18n.text("Reset to Global Defaults"), this::reset);
        header.add(mResetButton);
        mMenuButton = new FontAwesomeButton("\uf0c9", I18n.text("Menu"), this::actionMenu);
        header.add(mMenuButton);
        content.add(header, BorderLayout.NORTH);
        mLocationsPanel = new HitLocationTablePanel(getHitLocations(), () -> {
            getHitLocations().update();
            if (mCharacter == null) {
                Settings.getInstance().notifyOfChange();
            } else {
                mCharacter.notifyOfChange();
            }
            adjustResetButton();
        });
        mScroller = new ScrollPanel(mLocationsPanel);
        content.add(mScroller, BorderLayout.CENTER);
        if (mCharacter != null) {
            mCharacter.addChangeListener(this);
            Settings.getInstance().addChangeListener(this);
        }
        adjustResetButton();
        establishSizing();
        WindowUtils.packAndCenterWindowOn(this, null);
        EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
    }

    @Override
    public void establishSizing() {
        Dimension min = getMinimumSize();
        setMinimumSize(new Dimension(Math.max(min.width, 600), min.height));
    }

    private HitLocationTable getHitLocations() {
        return SheetSettings.get(mCharacter).getHitLocations();
    }

    private void actionMenu() {
        JPopupMenu menu = new JPopupMenu();
        menu.add(createMenuItem(I18n.text("Import…"), this::importData, true));
        menu.add(createMenuItem(I18n.text("Export…"), this::exportData, true));
        for (LibraryHitLocationTables tables : LibraryHitLocationTables.get()) {
            menu.addSeparator();
            menu.add(createMenuItem(tables.toString(), null, false));
            for (HitLocationTable choice : tables.getTables()) {
                menu.add(createMenuItem(choice.getName(), () -> reset(choice), true));
            }
        }
        menu.show(mMenuButton, 0, 0);
    }

    private static JMenuItem createMenuItem(String title, Runnable onSelection, boolean enabled) {
        JMenuItem item = new JMenuItem(title);
        item.addActionListener((evt) -> onSelection.run());
        item.setEnabled(enabled);
        return item;
    }

    private void importData() {
        Path path = Modal.presentOpenFileDialog(this, I18n.text("Import…"),
                FileType.HIT_LOCATIONS.getFilter());
        if (path != null) {
            try {
                reset(new HitLocationTable(path));
            } catch (IOException ioe) {
                Log.error(ioe);
                Modal.showError(this, I18n.text("Unable to import hit locations."));
            }
        }
    }

    private void exportData() {
        Path path = Modal.presentSaveFileDialog(this, I18n.text("Export…"),
                Settings.getInstance().getLastDir().resolve(I18n.text("hit_locations")),
                FileType.HIT_LOCATIONS.getFilter());
        if (path != null) {
            SafeFileUpdater transaction = new SafeFileUpdater();
            transaction.begin();
            try {
                File transactionFile = transaction.getTransactionFile(path.toFile());
                try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(transactionFile, StandardCharsets.UTF_8)), "\t")) {
                    w.startMap();
                    w.keyValue(DataFile.TYPE, HitLocationTable.JSON_TYPE_NAME);
                    w.keyValue(DataFile.VERSION, DataFile.CURRENT_VERSION);
                    w.key(HitLocationTable.JSON_TYPE_NAME);
                    mLocationsPanel.getHitLocations().toJSON(w, null);
                    w.endMap();
                }
                transaction.commit();
            } catch (Exception exception) {
                Log.error(exception);
                transaction.abort();
                Modal.showError(this, I18n.text("Unable to export hit locations."));
            }
        }
    }

    private void reset() {
        HitLocationTable locations;
        if (mCharacter == null) {
            locations = LibraryHitLocationTables.getHumanoid();
        } else {
            locations = Settings.getInstance().getSheetSettings().getHitLocations();
        }
        reset(locations);
        adjustResetButton();
    }

    private void reset(HitLocationTable locations) {
        mLocationsPanel.reset(locations.clone());
        mLocationsPanel.getAdjustCallback().run();
        revalidate();
        repaint();
        EventQueue.invokeLater(() -> mScroller.getViewport().setViewPosition(new Point(0, 0)));
    }

    private void adjustResetButton() {
        HitLocationTable prefsLocations = Settings.getInstance().getSheetSettings().getHitLocations();
        if (mCharacter == null) {
            mResetButton.setEnabled(!prefsLocations.equals(LibraryHitLocationTables.getHumanoid()));
        } else {
            mResetButton.setEnabled(!mCharacter.getSheetSettings().getHitLocations().equals(prefsLocations));
        }
    }

    @Override
    public boolean mayAttemptClose() {
        return true;
    }

    @Override
    public boolean attemptClose() {
        windowClosing(new WindowEvent(this, WindowEvent.WINDOW_CLOSING));
        return true;
    }

    @Override
    public void dispose() {
        synchronized (INSTANCES) {
            INSTANCES.remove(mCharacter == null ? null : mCharacter.getID());
        }
        if (mCharacter != null) {
            mCharacter.removeChangeListener(this);
            Settings.getInstance().removeChangeListener(this);
        }
        super.dispose();
    }

    @Override
    public void dataWasChanged() {
        if (!mUpdatePending) {
            mUpdatePending = true;
            EventQueue.invokeLater(() -> {
                setTitle(createTitle(mCharacter));
                adjustResetButton();
                mUpdatePending = false;
            });
        }
    }
}
