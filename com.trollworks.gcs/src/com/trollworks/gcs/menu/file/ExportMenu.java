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
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.event.MenuEvent;
import javax.swing.event.MenuListener;

public class ExportMenu extends JMenu implements MenuListener {
    public ExportMenu() {
        super(I18n.text("Export To…"));
        addMenuListener(this);
    }

    @Override
    public void menuCanceled(MenuEvent event) {
        // Nothing to do.
    }

    @Override
    public void menuDeselected(MenuEvent event) {
        // Nothing to do.
    }

    @Override
    public void menuSelected(MenuEvent event) {
        removeAll();
        boolean shouldEnable = Command.getTarget(SheetDockable.class) != null;
        ExportToGURPSCalculatorCommand.INSTANCE.setEnabled(shouldEnable);
        ExportToPNGCommand.INSTANCE.setEnabled(shouldEnable);
        add(ExportToGURPSCalculatorCommand.INSTANCE);
        add(ExportToPNGCommand.INSTANCE);
        for (Library lib : Library.LIBRARIES) {
            List<Command> cmds = new ArrayList<>();
            Path          dir  = lib.getPath().resolve("Output Templates");
            if (Files.isDirectory(dir)) {
                // IMPORTANT: On Windows, calling any of the older methods to list the contents of a
                // directory results in leaving state around that prevents future move & delete
                // operations. Only use this style of access for directory listings to avoid that.
                try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                    for (Path path : stream) {
                        cmds.add(new ExportToTextTemplateCommand(path, lib));
                    }
                } catch (IOException exception) {
                    Log.error(exception);
                }
                cmds.sort((c1, c2) -> NumericComparator.caselessCompareStrings(PathUtils.getLeafName(c1.getTitle(), true), PathUtils.getLeafName(c2.getTitle(), true)));
            }
            if (!cmds.isEmpty()) {
                addSeparator();
                JMenuItem header = new JMenuItem(String.format(I18n.text("%s Output Templates"), lib.getTitle()));
                header.setEnabled(false);
                add(header);
                for (Command cmd : cmds) {
                    cmd.setEnabled(shouldEnable);
                    add(cmd);
                }
            }
        }
    }
}
