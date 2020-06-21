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

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.library.LibraryDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Copy to Character Sheet" command. */
public class CopyToSheetCommand extends Command {
    /** The action command this command will issue. */
    public static final String             CMD_COPY_TO_SHEET = "CopyToSheet";
    /** The singleton {@link CopyToSheetCommand} for the menu bar. */
    public static final CopyToSheetCommand INSTANCE          = new CopyToSheetCommand();
    private             ListOutline        mOutline;
    private             SheetDockable      mSheet;

    private CopyToSheetCommand() {
        super(I18n.Text("Copy to Character Sheet"), CMD_COPY_TO_SHEET, KeyEvent.VK_C, SHIFTED_COMMAND_MODIFIER);
    }

    /**
     * Creates a new {@link CopyToSheetCommand}.
     *
     * @param outline The outline to work against.
     * @param sheet   The sheet to target.
     */
    public CopyToSheetCommand(ListOutline outline, SheetDockable sheet) {
        super(sheet.getTitle(), CMD_COPY_TO_SHEET);
        mOutline = outline;
        mSheet = sheet;
    }

    @Override
    public void adjust() {
        boolean     shouldEnable = false;
        ListOutline outline      = getOutline();
        if (outline != null && outline.getModel().hasSelection()) {
            SheetDockable lastSheet = getSheet();
            shouldEnable = lastSheet != null && lastSheet != UIUtilities.getSelfOrAncestorOfType(outline, SheetDockable.class);
        }
        setEnabled(shouldEnable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ListOutline outline = getOutline();
        if (outline != null) {
            OutlineModel outlineModel = outline.getModel();
            if (outlineModel.hasSelection()) {
                SheetDockable sheet = getSheet();
                if (sheet != null) {
                    sheet.addRows(outlineModel.getSelectionAsList(true));
                }
            }
        }
    }

    private ListOutline getOutline() {
        ListOutline outline = mOutline;
        if (outline == null) {
            LibraryDockable library = getTarget(LibraryDockable.class);
            if (library != null) {
                outline = library.getOutline();
            }
        }
        return outline;
    }

    private SheetDockable getSheet() {
        SheetDockable sheet = mSheet;
        if (sheet == null) {
            sheet = SheetDockable.getLastActivated();
        }
        return sheet;
    }
}
