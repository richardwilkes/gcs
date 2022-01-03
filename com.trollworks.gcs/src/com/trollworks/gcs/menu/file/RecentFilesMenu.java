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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.event.MenuEvent;
import javax.swing.event.MenuListener;

/** The standard "Recent Files" menu. */
public class RecentFilesMenu extends JMenu implements MenuListener {
    private int mLastSeenRecentFilesUpdateCounter;

    /** Creates a new RecentFilesMenu. */
    public RecentFilesMenu() {
        super(I18n.text("Recent Files"));
        addMenuListener(this);
        mLastSeenRecentFilesUpdateCounter = Settings.getInstance().getLastRecentFilesUpdateCounter() - 1;
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
        Settings prefs                        = Settings.getInstance();
        int      lastRecentFilesUpdateCounter = prefs.getLastRecentFilesUpdateCounter();
        if (mLastSeenRecentFilesUpdateCounter != lastRecentFilesUpdateCounter) {
            mLastSeenRecentFilesUpdateCounter = lastRecentFilesUpdateCounter;
            removeAll();
            List<Path>  list            = new ArrayList<>();
            Set<String> set             = new HashSet<>();
            Set<String> needFullPathSet = new HashSet<>();
            for (Path path : prefs.getRecentFiles()) {
                if (Files.isReadable(path)) {
                    list.add(path);
                    String leaf = PathUtils.getLeafName(path, false);
                    if (set.contains(leaf)) {
                        needFullPathSet.add(leaf);
                    } else {
                        set.add(leaf);
                    }
                    if (list.size() == Settings.MAX_RECENT_FILES) {
                        break;
                    }
                }
            }
            for (Path path : list) {
                String title = PathUtils.getLeafName(path, false);
                if (needFullPathSet.contains(title)) {
                    title = path.toAbsolutePath().toString();
                }
                OpenDataFileCommand cmd = new OpenDataFileCommand(title, path);
                cmd.adjust();
                add(new JMenuItem(cmd));
            }
            prefs.setRecentFiles(list);
            if (!list.isEmpty()) {
                addSeparator();
            }
            JMenuItem item = new JMenuItem(ClearRecentFilesMenuCommand.INSTANCE);
            ClearRecentFilesMenuCommand.INSTANCE.adjust();
            add(item);
        }
    }
}
