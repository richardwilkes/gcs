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

import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

public class ConduitMessage {
    // No repeating characters allowed!
    private static final byte[]       ID = {'#', 'W', 'i', 'l', 'k', 'e', 's', '!'};
    protected            String       mUser;
    protected            String       mApp;
    protected            long         mID;
    protected            State        mState;
    protected            List<String> mFiles;

    public ConduitMessage(String app, long id, State state, List<String> files) {
        mUser = System.getProperty("user.name");
        mApp = app;
        mID = id;
        mState = state;
        mFiles = files;
    }

    public ConduitMessage(DataInputStream stream) throws IOException {
        int i      = 0;
        int length = ID.length;
        while (i < length) {
            byte value = stream.readByte();
            if (value == ID[i]) {
                i++;
            } else if (value == ID[0]) {
                i = 1;
            } else {
                i = 0;
            }
        }
        mUser = stream.readUTF();
        mApp = stream.readUTF();
        mID = stream.readLong();
        int state = stream.readInt();
        if (state >= State.LAUNCH.ordinal() && state <= State.TOOK_OVER_FOR.ordinal()) {
            mState = State.values()[state];
        } else {
            mState = State.INVALID;
        }
        if (mState == State.LAUNCH) {
            int count = stream.readInt();
            if (count > 100) {
                count = 100;
            }
            mFiles = new ArrayList<>(count);
            for (i = 0; i < count; i++) {
                mFiles.add(stream.readUTF());
            }
        }
    }

    void send(DataOutputStream stream) throws IOException {
        stream.write(ID);
        stream.writeUTF(mUser);
        stream.writeUTF(mApp);
        stream.writeLong(mID);
        stream.writeInt(mState.ordinal());
        if (mState == State.LAUNCH) {
            if (mFiles != null) {
                stream.writeInt(mFiles.size());
                for (String file : mFiles) {
                    stream.writeUTF(file);
                }
            } else {
                stream.writeInt(0);
            }
        }
        stream.flush();
    }
}
