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

import com.trollworks.gcs.datafile.DataFileDockable;
import com.trollworks.gcs.library.DataUpdater;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import java.awt.event.ActionEvent;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public class UpdateCommand extends Command {
    public static final UpdateCommand INSTANCE = new UpdateCommand();

    private UpdateCommand() {
        super(I18n.Text("Update Data…"), "UpdateData");
    }

    @Override
    public void adjust() {
        setEnabled(Command.getTarget(DataFileDockable.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (getTarget(DataFileDockable.class) != null) {
            try {
                long        start   = System.currentTimeMillis();
                DataUpdater du      = new DataUpdater();
                long        elapsed = System.currentTimeMillis() - start;
                System.out.println(elapsed + "ms");
                System.out.println(du.updatables.size() + " loaded objects");
                System.out.println("ignored:");
                List<Path> keys = new ArrayList<>(du.ignoreMap.keySet());
                Collections.sort(keys);
                for (Path k : keys) {
                    System.out.println("  " + k + ":");
                    for (Path p : du.ignoreMap.get(k)) {
                        System.out.println("    " + p);
                    }
                }
            } catch (Exception ex) {
                Log.error(ex);
            }
        }
    }
}
