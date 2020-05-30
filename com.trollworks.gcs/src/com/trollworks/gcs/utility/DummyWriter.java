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

package com.trollworks.gcs.utility;

import java.io.IOException;
import java.io.Writer;

public class DummyWriter extends Writer {
    @Override
    public void write(int c) throws IOException {
        // Unused
    }

    @Override
    public void write(char[] cbuf) throws IOException {
        // Unused
    }

    @Override
    public void write(char[] cbuf, int off, int len) throws IOException {
        // Unused
    }

    @Override
    public void write(String str) throws IOException {
        // Unused
    }

    @Override
    public void write(String str, int off, int len) throws IOException {
        // Unused
    }

    @Override
    public Writer append(CharSequence csq) throws IOException {
        return this;
    }

    @Override
    public Writer append(CharSequence csq, int start, int end) throws IOException {
        return this;
    }

    @Override
    public Writer append(char c) throws IOException {
        return this;
    }

    @Override
    public void flush() throws IOException {
        // Unused
    }

    @Override
    public void close() throws IOException {
        // Unused
    }
}
