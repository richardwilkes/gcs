/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.utility.BundleInfo;

import java.io.File;

/** Utility to generate the Info.plist file for GCS. */
public class GCSInfoPlistCreator {
	public static void main(String[] args) {
		if (args.length != 1) {
			System.err.println("You must provide the path to the Info.plist you'd like created as the sole argument."); //$NON-NLS-1$
			System.exit(1);
		}
		App.setup(GCS.class);
		GCS.registerFileTypes(null);
		BundleInfo.getDefault().write(new File(args[0]));
	}
}
