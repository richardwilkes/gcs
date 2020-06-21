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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.nio.file.Path;
import java.util.ArrayList;

public class ExportToPNGCommand extends Command {
    public static final ExportToPNGCommand INSTANCE = new ExportToPNGCommand();

    private ExportToPNGCommand() {
        super(I18n.Text("Export to PNG Image(s)…"), "ToPNG");
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        SheetDockable sheet = getTarget(SheetDockable.class);
        if (sheet != null) {
            String name = sheet.getSheet().getCharacter().getProfile().getName();
            if (name.isBlank()) {
                name = "untitled";
            }
            Path path = StdFileDialog.showSaveDialog(UIUtilities.getComponentForDialog(sheet), getTitle(), Preferences.getInstance().getLastDir().resolve(name), FileType.PNG.getFilter());
            if (path != null) {
                if (!sheet.getSheet().saveAsPNG(path, new ArrayList<>())) {
                    WindowUtils.showError(sheet, I18n.Text("An error occurred while trying to export the sheet as PNG."));
                }
            }
        }
    }
}
