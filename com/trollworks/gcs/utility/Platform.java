/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.utility;

/** Defines constants for each of the various platforms that can be detected. */
public enum Platform {
	/** The constant used for the Macintosh platform. */
	MACINTOSH,
	/** The constant used for the Windows platform. */
	WINDOWS,
	/** The constant used for the Linux platform. */
	LINUX,
	/** The constant used for the Solaris platform. */
	SOLARIS,
	/** The constant used for unknown platforms. */
	UNKNOWN;

	private static final Platform	CURRENT;

	static {
		String osName = System.getProperty("os.name"); //$NON-NLS-1$

		if (osName.startsWith("Mac")) { //$NON-NLS-1$
			CURRENT = MACINTOSH;
		} else if (osName.startsWith("Win")) { //$NON-NLS-1$
			CURRENT = WINDOWS;
		} else if (osName.startsWith("Linux")) { //$NON-NLS-1$
			CURRENT = LINUX;
		} else if (osName.startsWith("Sun")) { //$NON-NLS-1$
			CURRENT = SOLARIS;
		} else {
			CURRENT = UNKNOWN;
		}
	}

	/** @return The platform being run on. */
	public static final Platform current() {
		return CURRENT;
	}

	/** @return <code>true</code> if Macintosh is the platform being run on. */
	public static final boolean isMacintosh() {
		return CURRENT == MACINTOSH;
	}

	/** @return <code>true</code> if Windows is the platform being run on. */
	public static final boolean isWindows() {
		return CURRENT == WINDOWS;
	}

	/** @return <code>true</code> if Linux is the platform being run on. */
	public static final boolean isLinux() {
		return CURRENT == LINUX;
	}

	/** @return <code>true</code> if Solaris is the platform being run on. */
	public static final boolean isSolaris() {
		return CURRENT == SOLARIS;
	}

	/** @return <code>true</code> if platform being run on is Unix-based. */
	public static final boolean isUnix() {
		return isMacintosh() || isLinux() || isSolaris();
	}
}
