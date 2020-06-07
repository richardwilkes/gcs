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
import com.trollworks.gcs.character.TextTemplate;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;

import java.awt.event.ActionEvent;
import java.nio.file.Path;
import javax.swing.filechooser.FileNameExtensionFilter;

public class ExportToTextTemplateCommand extends Command {
    private Path mTemplatePath;

    public ExportToTextTemplateCommand(Path templatePath, boolean master) {
        super(String.format(I18n.Text("Export to %s…"), PathUtils.getLeafName(templatePath, false)), "ExportTextTemplate-" + (master ? "M-" : "U-") + PathUtils.getLeafName(templatePath, true));
        mTemplatePath = templatePath;
    }

    @Override
    public void adjust() {
        setEnabled(!StdMenuBar.SUPRESS_MENUS && getTarget(SheetDockable.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        SheetDockable sheet = getTarget(SheetDockable.class);
        if (sheet != null) {
            String name = sheet.getSheet().getCharacter().getProfile().getName();
            if (name.isBlank()) {
                name = "untitled";
            }
            String ext  = PathUtils.getExtension(mTemplatePath);
            Path   path = StdFileDialog.showSaveDialog(UIUtilities.getComponentForDialog(sheet), getTitle(), Preferences.getInstance().getLastDir().resolve(name), new FileNameExtensionFilter(ext + I18n.Text(" Files"), ext));
            if (path != null) {
                if (!new TextTemplate(sheet.getSheet()).export(path, mTemplatePath)) {
                    WindowUtils.showError(sheet, String.format(I18n.Text("An error occurred while trying to export the sheet as %s."), PathUtils.getLeafName(mTemplatePath, false)));
                }
            }
        }
    }
}
