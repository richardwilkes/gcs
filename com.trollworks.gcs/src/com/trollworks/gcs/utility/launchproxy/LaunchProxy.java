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

package com.trollworks.gcs.utility.launchproxy;

import com.trollworks.gcs.menu.file.OpenCommand;
import com.trollworks.gcs.menu.file.OpenDataFileCommand;
import com.trollworks.gcs.ui.widget.WindowUtils;

import java.awt.EventQueue;
import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.List;

/**
 * Provides the ability for an application to be launched from different terminals and still use
 * only a single instance.
 */
public class LaunchProxy {
    private InetSocketAddress mSocketAddress;
    private Server            mServer;
    private Socket            mSocket;
    private DataInputStream   mInput;
    private DataOutputStream  mOutput;
    private String            mUserFilter;
    private long              mID;
    private boolean           mReady;

    public LaunchProxy(List<Path> paths) {
        mSocketAddress = new InetSocketAddress(InetAddress.getLoopbackAddress(), 13321);
        mUserFilter = System.getProperty("user.name", "*");
        mID = new SecureRandom().nextLong();
        ArrayList<String> files = new ArrayList<>();
        for (Path p : paths) {
            files.add(p.toAbsolutePath().toString());
        }
        reconnect();
        Thread receptionThread = new Thread(() -> {
            while (true) {
                try {
                    ConduitMessage msg = new ConduitMessage(mInput);
                    if (mUserFilter.equals(msg.mUser) && "GCS".equals(msg.mApp)) {
                        switch (msg.mState) {
                        case LAUNCH:
                            boolean ready;
                            synchronized (this) {
                                ready = mReady;
                            }
                            if (ready) {
                                if (mID != msg.mID) {
                                    send(new ConduitMessage("GCS", msg.mID, State.TOOK_OVER_FOR, null));
                                    WindowUtils.forceAppToFront();
                                    if (msg.mFiles != null && !msg.mFiles.isEmpty()) {
                                        for (String file : msg.mFiles) {
                                            OpenDataFileCommand.open(Paths.get(file));
                                        }
                                    } else {
                                        EventQueue.invokeLater(OpenCommand::open);
                                    }
                                }
                            }
                            break;
                        case TOOK_OVER_FOR:
                            if (mID == msg.mID) {
                                System.exit(0);
                            }
                            break;
                        default:
                            break;
                        }
                    }
                } catch (Exception exception) {
                    reconnect();
                }
            }
        }, "LaunchProxy");
        receptionThread.setDaemon(true);
        receptionThread.start();
        send(new ConduitMessage("GCS", mID, State.LAUNCH, files));
        try {
            // Give it a chance to terminate this run...
            Thread.sleep(1500);
        } catch (Exception exception) {
            // Ignore
        }
    }

    private void reconnect() {
        shutdownSocket();
        if (mServer != null) {
            mServer.shutdown();
            mServer = null;
        }

        while (true) {
            try {
                mServer = new Server(mSocketAddress);
                mServer.start();
            } catch (Exception exception) {
                // Someone else is already the server, just start a client.
            }

            mSocket = new Socket();
            try {
                mSocket.connect(mSocketAddress);
                mInput = new DataInputStream(mSocket.getInputStream());
                mOutput = new DataOutputStream(mSocket.getOutputStream());
                return;
            } catch (Exception exception) {
                // The server is no longer around or hasn't quite started up
                // yet, so loop and try again.
                shutdownSocket();
            }
        }
    }

    private void shutdownSocket() {
        if (mSocket != null) {
            try {
                mSocket.close();
            } catch (Exception exception) {
                // Ignore.
            }
            mSocket = null;
            mInput = null;
            mOutput = null;
        }
    }

    private void send(ConduitMessage msg) {
        while (true) {
            try {
                msg.send(mOutput);
                return;
            } catch (Exception exception) {
                reconnect();
            }
        }
    }

    /**
     * Sets whether this application is ready to take over responsibility for other copies being
     * launched.
     *
     * @param ready Whether the application is ready or not.
     */
    public synchronized void setReady(boolean ready) {
        mReady = ready;
    }
}
