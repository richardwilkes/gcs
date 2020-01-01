/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
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
import java.nio.charset.StandardCharsets;

/** Utility to generate the Info.plist file for GCS. */
public class GCSInfoPlistCreator {
    private static final String PLIST   = "Info.plist";
    private static final String PKGINFO = "PkgInfo";

    public static void main(String[] args) {
        File plist = new File(PLIST);
        File pkg   = new File(PKGINFO);
        if (args.length > 0) {
            plist = new File(args[0], PLIST);
            pkg = new File(args[0], PKGINFO);
        }
        App.setup(GCS.class);
        new GCSInfoPlistCreator().run(plist, pkg);
    }

    @SuppressWarnings("static-method")
    private void run(File plist, File pkg) {
        GCSCmdLine.registerFileTypes(null);
        BundleInfo info = BundleInfo.getDefault();
        info.write(plist, "app.icns");
        try (PrintWriter out = new PrintWriter(new BufferedOutputStream(new FileOutputStream(pkg)), false, StandardCharsets.UTF_8)) {
            out.print("APPL");
            out.print(info.getSignature());
        } catch (Exception exception) {
            Log.error(exception);
        }
    }
}
