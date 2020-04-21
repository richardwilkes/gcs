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

package com.trollworks.gcs.io;

import java.io.IOException;
import java.io.OutputStream;

public class NonClosingPassThroughOutputStream extends OutputStream {
    private OutputStream mOut;

    public NonClosingPassThroughOutputStream(OutputStream out) {
        mOut = out;
    }

    @Override
    public void write(int b) throws IOException {
        mOut.write(b);
    }

    @Override
    public void write(byte[] b) throws IOException {
        mOut.write(b);
    }

    @Override
    public void write(byte[] b, int off, int len) throws IOException {
        mOut.write(b, off, len);
    }

    @Override
    public void flush() throws IOException {
        mOut.flush();
    }

    @Override
    public void close() throws IOException {
        mOut.flush();
    }
}
