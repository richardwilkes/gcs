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

package com.trollworks.gcs.character;

import com.trollworks.gcs.settings.AttributeSettingsWindow;
import com.trollworks.gcs.settings.BodyTypeSettingsWindow;
import com.trollworks.gcs.settings.QuickExport;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.settings.SheetSettingsWindow;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.FontAdjustable;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.Toolbar;
import com.trollworks.gcs.ui.widget.dock.Dock;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.undo.StdUndoManager;

import java.awt.BorderLayout;
import java.awt.Component;
import java.nio.file.Path;
import javax.swing.JViewport;

/** A list of advantages and disadvantages from a library. */
public class SheetDockable extends CollectedOutlinesDockable implements FontAdjustable {
    private static SheetDockable  LAST_ACTIVATED;
    private        CharacterSheet mSheet;
    private        FontIconButton mQuickExportButton;

    /** Creates a new SheetDockable. */
    public SheetDockable(GURPSCharacter character) {
        super(character);
        mSheet = new CharacterSheet(character);
        long modifiedOn = character.getModifiedOn();
        createToolbar();
        ScrollPanel scroller = new ScrollPanel(mSheet);
        JViewport   viewport = scroller.getViewport();
        viewport.setBackground(Colors.PAGE_VOID);
        viewport.addChangeListener(mSheet);
        add(scroller, BorderLayout.CENTER);
        mSheet.rebuild();
        character.setModifiedOn(modifiedOn);
        character.setModified(false);
        StdUndoManager undoManager = getUndoManager();
        undoManager.discardAllEdits();
        character.setUndoManager(undoManager);
    }

    @Override
    protected Toolbar createToolbar() {
        Toolbar toolbar = super.createToolbar();
        mQuickExportButton = new FontIconButton(FontAwesome.FILE_EXPORT,
                I18n.text("Quick Export\nExport to the same location using the last used output template for this sheet"),
                this::quickExport);
        toolbar.add(new FontIconButton(FontAwesome.COG, I18n.text("Settings"),
                (b) -> SheetSettingsWindow.display(getDataFile())), 0);
        toolbar.add(mQuickExportButton, 1);
        updateQuickExport();
        return toolbar;
    }

    @Override
    public CollectedOutlines getCollectedOutlines() {
        return mSheet;
    }

    @Override
    public boolean attemptClose() {
        boolean closed = super.attemptClose();
        if (closed) {
            GURPSCharacter gurpsCharacter = getDataFile();
            SheetSettingsWindow.closeFor(gurpsCharacter);
            AttributeSettingsWindow.closeFor(gurpsCharacter);
            BodyTypeSettingsWindow.closeFor(gurpsCharacter);
            mSheet.dispose();
        }
        return closed;
    }

    @Override
    public Component getRetargetedFocus() {
        return mSheet;
    }

    /** @return The last activated SheetDockable. */
    public static SheetDockable getLastActivated() {
        if (LAST_ACTIVATED != null) {
            Dock dock = UIUtilities.getAncestorOfType(LAST_ACTIVATED, Dock.class);
            if (dock == null) {
                LAST_ACTIVATED = null;
            }
        }
        return LAST_ACTIVATED;
    }

    @Override
    public void activated() {
        super.activated();
        LAST_ACTIVATED = this;
    }

    @Override
    public GURPSCharacter getDataFile() {
        return (GURPSCharacter) super.getDataFile();
    }

    /** @return The {@link CharacterSheet}. */
    public CharacterSheet getSheet() {
        return mSheet;
    }

    @Override
    protected String getUntitledName() {
        String name = mSheet.getCharacter().getProfile().getName().trim();
        return (name.isBlank()) ? super.getUntitledName() : name;
    }

    @Override
    protected String getUntitledBaseName() {
        return I18n.text("Untitled Sheet");
    }

    @Override
    public PrintProxy getPrintProxy() {
        return mSheet;
    }

    public void updateQuickExport() {
        boolean enabled = false;
        Path    path    = mSheet.getCharacter().getPath();
        if (path != null) {
            QuickExport qe = Settings.getInstance().getQuickExport(path.toAbsolutePath().toString());
            if (qe != null) {
                enabled = qe.isValid();
            }
        }
        mQuickExportButton.setEnabled(enabled);
    }

    private void quickExport(Component comp) {
        updateQuickExport();
        if (mQuickExportButton.isEnabled()) {
            Path        path = mSheet.getCharacter().getPath();
            QuickExport qe   = Settings.getInstance().getQuickExport(path.toAbsolutePath().toString());
            if (qe != null) {
                qe.export(this);
            }
        }
    }

    public void recordQuickExport(QuickExport qe) {
        Settings.getInstance().putQuickExport(mSheet.getCharacter(), qe);
        updateQuickExport();
    }

    @Override
    public void adjustToFontChanges() {
        mSheet.markForRebuild();
    }
}
