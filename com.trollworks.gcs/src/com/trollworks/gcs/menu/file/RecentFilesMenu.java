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

import com.trollworks.gcs.preferences.Preferences;
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

    /** Creates a new {@link RecentFilesMenu}. */
    public RecentFilesMenu() {
        super(I18n.Text("Recent Files"));
        addMenuListener(this);
        mLastSeenRecentFilesUpdateCounter = Preferences.getInstance().getLastRecentFilesUpdateCounter() - 1;
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
        Preferences prefs                        = Preferences.getInstance();
        int         lastRecentFilesUpdateCounter = prefs.getLastRecentFilesUpdateCounter();
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
                    if (list.size() == Preferences.MAX_RECENT_FILES) {
                        break;
                    }
                }
            }
            for (Path path : list) {
                String title = PathUtils.getLeafName(path, false);
                if (needFullPathSet.contains(title)) {
                    title = path.toAbsolutePath().toString();
                }
                add(new JMenuItem(new OpenDataFileCommand(title, path)));
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
