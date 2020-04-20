/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.gcs.menu.file.OpenCommand;
import com.trollworks.gcs.menu.file.OpenDataFileCommand;
import com.trollworks.gcs.io.conduit.Conduit;
import com.trollworks.gcs.io.conduit.ConduitMessage;
import com.trollworks.gcs.io.conduit.ConduitReceiver;
import com.trollworks.gcs.ui.widget.WindowUtils;

import java.awt.EventQueue;
import java.io.File;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.StringTokenizer;

/**
 * Provides the ability for an application to be launched from different terminals and still use
 * only a single instance.
 */
class LaunchProxy implements ConduitReceiver {
    private static final String     LAUNCH_ID        = "Launched";
    private static final String     TOOK_OVER_FOR_ID = "TookOverFor";
    private              Conduit    mConduit;
    private              long       mTimeStamp;
    private              boolean    mReady;
    private              List<File> mFiles;

    LaunchProxy(List<Path> files) {
        StringBuilder buffer = new StringBuilder();
        mFiles = new ArrayList<>();
        mTimeStamp = System.currentTimeMillis();
        buffer.append(LAUNCH_ID);
        buffer.append(' ');
        buffer.append(mTimeStamp);
        buffer.append(' ');
        boolean needComma = false;
        for (Path file : files) {
            if (needComma) {
                buffer.append(',');
            } else {
                needComma = true;
            }
            buffer.append(file.toAbsolutePath().toString().replaceAll("@", "@!").replaceAll(" ", "@%").replaceAll(",", "@#"));
        }
        mConduit = new Conduit(this, false);
        mConduit.send(new ConduitMessage("GCS", buffer.toString()));
        try {
            // Give it a chance to terminate this run...
            Thread.sleep(1500);
        } catch (Exception exception) {
            // Ignore
        }
    }

    /**
     * Sets whether this application is ready to take over responsibility for other copies being
     * launched.
     *
     * @param ready Whether the application is ready or not.
     */
    void setReady(boolean ready) {
        mReady = ready;
    }

    @Override
    public void conduitMessageReceived(ConduitMessage msg) {
        StringTokenizer tokenizer = new StringTokenizer(msg.getMessage(), " ");
        if (tokenizer.hasMoreTokens()) {
            String token = tokenizer.nextToken();
            if (mReady && LAUNCH_ID.equals(token)) {
                if (tokenizer.hasMoreTokens()) {
                    long timeStamp = getLong(tokenizer);
                    if (timeStamp != mTimeStamp) {
                        mConduit.send(new ConduitMessage("GCS", TOOK_OVER_FOR_ID + ' ' + timeStamp));
                        WindowUtils.forceAppToFront();
                        if (tokenizer.hasMoreTokens()) {
                            tokenizer = new StringTokenizer(tokenizer.nextToken(), ",");
                            while (tokenizer.hasMoreTokens()) {
                                synchronized (mFiles) {
                                    mFiles.add(new File(tokenizer.nextToken().replaceAll("@#", ",").replaceAll("@%", " ").replaceAll("@!", "@")));
                                }
                            }
                            synchronized (mFiles) {
                                if (!mFiles.isEmpty()) {
                                    for (File file : mFiles) {
                                        OpenDataFileCommand.open(file);
                                    }
                                    mFiles.clear();
                                }
                            }
                        } else {
                            EventQueue.invokeLater(() -> OpenCommand.open());
                        }
                    }
                }
            } else if (TOOK_OVER_FOR_ID.equals(token)) {
                if (tokenizer.hasMoreTokens()) {
                    if (getLong(tokenizer) == mTimeStamp) {
                        System.exit(0);
                    }
                }
            }
        }
    }

    private static long getLong(StringTokenizer tokenizer) {
        try {
            return Long.parseLong(tokenizer.nextToken().trim());
        } catch (Exception exception) {
            return -1;
        }
    }

    @Override
    public String getConduitMessageIDFilter() {
        return "GCS";
    }

    @Override
    public String getConduitMessageUserFilter() {
        return System.getProperty("user.name");
    }
}
