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

package com.trollworks.gcs.io.conduit;

import com.trollworks.gcs.io.Log;

import java.awt.EventQueue;
import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.UnknownHostException;

/** Provides a conduit through which messages from external processes can be received. */
public class Conduit implements Runnable {
    /** The default port used by the conduit. */
    public static final int               DEFAULT_PORT = 13321;
    private             InetSocketAddress mSocketAddress;
    private             Server            mServer;
    private             Socket            mSocket;
    private             DataInputStream   mInput;
    private             DataOutputStream  mOutput;
    private             ConduitReceiver   mReceiver;
    private             boolean           mOnEventThread;
    private             String            mUserFilter;
    private             String            mIDFilter;

    /**
     * Creates a new conduit with the default port on the loopback address.
     *
     * @param receiver      The object that wants the messages from this conduit.
     * @param onEventThread Pass in {@code true} to receive the messages on the event thread.
     */
    public Conduit(ConduitReceiver receiver, boolean onEventThread) {
        this(null, receiver, onEventThread);
    }

    /**
     * Creates a new conduit with the specified port on the loopback address.
     *
     * @param port          The port to communicate on.
     * @param receiver      The object that wants the messages from this conduit.
     * @param onEventThread Pass in {@code true} to receive the messages on the event thread.
     */
    public Conduit(int port, ConduitReceiver receiver, boolean onEventThread) {
        this(new InetSocketAddress(getLoopBackAddress(), port), receiver, onEventThread);
    }

    /**
     * Creates a new conduit with the specified port on the loopback address.
     *
     * @param address       The address to communicate on.
     * @param port          The port to communicate on.
     * @param receiver      The object that wants the messages from this conduit.
     * @param onEventThread Pass in {@code true} to receive the messages on the event thread.
     */
    public Conduit(InetAddress address, int port, ConduitReceiver receiver, boolean onEventThread) {
        this(new InetSocketAddress(address, port), receiver, onEventThread);
    }

    /**
     * Creates a new conduit with the specified socket address.
     *
     * @param socketAddress The socket address to use.
     * @param receiver      The object that wants the messages from this conduit.
     * @param onEventThread Pass in {@code true} to receive the messages on the event thread.
     */
    public Conduit(InetSocketAddress socketAddress, ConduitReceiver receiver, boolean onEventThread) {
        if (socketAddress == null) {
            socketAddress = new InetSocketAddress(getLoopBackAddress(), DEFAULT_PORT);
        }
        mSocketAddress = socketAddress;
        mReceiver = receiver;
        mOnEventThread = onEventThread;
        Thread receptionThread = new Thread(this, Conduit.class.getSimpleName() + '@' + mSocketAddress);
        mUserFilter = mReceiver.getConduitMessageUserFilter();
        mIDFilter = mReceiver.getConduitMessageIDFilter();
        reconnect();
        receptionThread.setPriority(Thread.NORM_PRIORITY);
        receptionThread.setDaemon(true);
        receptionThread.start();
    }

    private static InetAddress getLoopBackAddress() {
        try {
            return InetAddress.getByName(null);
        } catch (UnknownHostException uhe) {
            // This can't occur, as the loopback address is always valid.
            Log.error(uhe);
            return null;
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
                mServer.setDaemon(true);
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
            } catch (Exception ex2) {
                // The server is no longer around or hasn't quite started up
                // yet, so loop and try again.
                shutdownSocket();
            }
        }
    }

    /**
     * Sends a message to all clients connected to the conduit.
     *
     * @param msg The message.
     */
    public void send(ConduitMessage msg) {
        while (true) {
            try {
                msg.send(mOutput);
                return;
            } catch (Exception exception) {
                reconnect();
            }
        }
    }

    @Override
    public void run() {
        while (true) {
            try {
                ConduitMessage msg = new ConduitMessage(mInput);
                if ((mUserFilter == null || mUserFilter.equals(msg.getUser())) && (mIDFilter == null || mIDFilter.equals(msg.getID()))) {
                    if (mOnEventThread) {
                        msg.setReceiver(mReceiver);
                        EventQueue.invokeLater(msg);
                    } else {
                        mReceiver.conduitMessageReceived(msg);
                    }
                }
            } catch (Exception exception) {
                reconnect();
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
}
