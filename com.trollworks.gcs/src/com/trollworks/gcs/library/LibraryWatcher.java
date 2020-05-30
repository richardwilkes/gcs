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

package com.trollworks.gcs.library;

import com.trollworks.gcs.utility.Log;

import java.io.IOException;
import java.nio.file.FileSystems;
import java.nio.file.Path;
import java.nio.file.StandardWatchEventKinds;
import java.nio.file.WatchKey;
import java.nio.file.WatchService;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;
import javax.swing.SwingUtilities;

public class LibraryWatcher implements Runnable {
    public static final LibraryWatcher      INSTANCE = new LibraryWatcher();
    private             WatchService        mWatcher;
    private             Map<Path, WatchKey> mPathKeyMap;

    private LibraryWatcher() {
        mPathKeyMap = new HashMap<>();
        try {
            mWatcher = FileSystems.getDefault().newWatchService();
            Thread thread = new Thread(this, "Library Watcher");
            thread.setDaemon(true);
            thread.start();
        } catch (IOException exception) {
            Log.error(exception);
        }
    }

    public void run() {
        if (mWatcher == null) {
            return;
        }
        while (true) {
            WatchKey key;
            try {
                key = mWatcher.take();
                key.pollEvents();
                key.reset();
                while (true) {
                    key = mWatcher.poll();
                    if (key == null) {
                        break;
                    }
                    key.pollEvents();
                    key.reset();
                }
            } catch (InterruptedException iex) {
                return;
            }
            SwingUtilities.invokeLater(() -> {
                LibraryExplorerDockable explorer = LibraryExplorerDockable.get();
                if (explorer != null) {
                    explorer.refresh();
                }
            });
        }
    }

    public void watchDirs(Set<Path> dirs) {
        if (mWatcher == null) {
            return;
        }
        Map<Path, WatchKey> keep = new HashMap<>();
        for (Path p : dirs) {
            WatchKey key = mPathKeyMap.get(p);
            if (key == null) {
                try {
                    keep.put(p, p.register(mWatcher, StandardWatchEventKinds.ENTRY_CREATE, StandardWatchEventKinds.ENTRY_DELETE));
                } catch (IOException exception) {
                    Log.error(exception);
                }
            } else {
                keep.put(p, key);
                mPathKeyMap.remove(p);
            }
        }
        for (WatchKey watchKey : mPathKeyMap.values()) {
            watchKey.cancel();
        }
        mPathKeyMap = keep;
    }
}
