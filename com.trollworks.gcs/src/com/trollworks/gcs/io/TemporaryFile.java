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

package com.trollworks.gcs.io;

import java.io.Closeable;
import java.io.File;
import java.io.IOException;

/** Used to create a file for temporary use. The file will be deleted when closed */
public class TemporaryFile extends File implements Closeable {
    public TemporaryFile(String prefix, String extension) throws IOException {
        super(File.createTempFile(prefix, extension).getAbsolutePath());
    }

    @Override
    public void close() throws IOException {
        delete();
    }
}
