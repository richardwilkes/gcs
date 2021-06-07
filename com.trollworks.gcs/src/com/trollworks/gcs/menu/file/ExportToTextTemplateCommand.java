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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.character.TextTemplate;
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.settings.QuickExport;
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

    public ExportToTextTemplateCommand(Path templatePath, Library library) {
        super(PathUtils.getLeafName(templatePath, false) + "…", "ExportTextTemplate-" + library.getKey() + "-" + PathUtils.getLeafName(templatePath, true));
        mTemplatePath = templatePath;
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        SheetDockable dockable = getTarget(SheetDockable.class);
        if (dockable != null) {
            String name = PathUtils.cleanNameForFile(dockable.getSheet().getCharacter().getProfile().getName());
            if (name.isBlank()) {
                name = "untitled";
            }
            String ext  = PathUtils.getExtension(mTemplatePath);
            Path   path = StdFileDialog.showSaveDialog(UIUtilities.getComponentForDialog(dockable), getTitle(), Settings.getInstance().getLastDir().resolve(name), new FileNameExtensionFilter(ext + I18n.text(" Files"), ext));
            if (path != null) {
                performExport(dockable, mTemplatePath, path);
            }
        }
    }

    public static void performExport(SheetDockable dockable, Path templatePath, Path exportPath) {
        if (new TextTemplate(dockable.getSheet()).export(exportPath, templatePath)) {
            dockable.recordQuickExport(new QuickExport(templatePath, exportPath));
        } else {
            WindowUtils.showError(dockable, String.format(I18n.text("An error occurred while trying to export the sheet as %s."), PathUtils.getLeafName(templatePath, false)));
        }
    }
}
