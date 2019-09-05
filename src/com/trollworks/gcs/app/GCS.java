/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.toolkit.ui.App;

/** The main entry point for the character sheet. */
public class GCS {
    /**
     * The main entry point for the character sheet.
     *
     * @param args Arguments to the program.
     */
    public static void main(String[] args) {
        // The call to App.setup() must be done first and no other code, including static variable
        // assignment, should be present in this file other than the call to GCSCmdLine.start().
        App.setup(GCS.class);
        GCSCmdLine.start(args);
    }
}
