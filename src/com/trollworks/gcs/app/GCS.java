/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.toolkit.utility.Launcher;

/**
 * This class is to be compiled with Java 1.1 to allow it to be loaded and executed even on very old
 * JVM's.
 */
public class GCS {
	public static void main(String[] args) {
		Launcher.launch("1.8", "com.trollworks.gcs.app.GCSMain", args); //$NON-NLS-1$ //$NON-NLS-2$
	}
}
