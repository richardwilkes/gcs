/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.toolkit.io.Log;
import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.utility.BundleInfo;

import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.PrintWriter;

/** Utility to generate the Info.plist file for GCS. */
public class GCSInfoPlistCreator {
    @SuppressWarnings("nls")
    public static void main(String[] args) {
        File plist = new File("Info.plist");
        File pkg   = new File("PkgInfo");
        if (args.length > 0) {
            plist = new File(args[0], "Info.plist");
            pkg   = new File(args[0], "PkgInfo");
        }
        App.setup(GCS.class);
        GCS.registerFileTypes(null);
        BundleInfo info = BundleInfo.getDefault();
        info.write(plist);
        try (PrintWriter out = new PrintWriter(new BufferedOutputStream(new FileOutputStream(pkg)))) {
            out.print("APPL");
            out.print(info.getSignature());
        } catch (Exception exception) {
            Log.error(exception);
        }
    }
}
