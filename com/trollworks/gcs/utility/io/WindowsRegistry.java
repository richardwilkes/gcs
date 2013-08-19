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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.utility.io;

import com.trollworks.gcs.utility.Platform;

import java.io.File;
import java.io.IOException;
import java.io.PrintWriter;
import java.util.HashMap;

/**
 * Adds registry entries for mapping file extensions to an application and icons to files with those
 * extensions.
 */
public class WindowsRegistry implements Runnable {
	private static final String		DOT				= ".";	//$NON-NLS-1$
	private static final String		UNDERSCORE		= "_";	//$NON-NLS-1$
	private static final String		START_BRACKET	= "[";	//$NON-NLS-1$
	private static final String		END_BRACKET		= "]";	//$NON-NLS-1$
	private String					mPrefix;
	private HashMap<String, String>	mMap;
	private File					mAppFile;
	private File					mIconDir;

	/**
	 * @param prefix The registry prefix to use.
	 * @param map A map of extensions (no leading period) to descriptions.
	 * @param appFile The application to execute for these file extensions.
	 * @param iconDir The icon directory, where Windows .ico files can be found for each extension,
	 *            in the form 'extension.ico'. For example, the extension 'xyz' would need an icon
	 *            file named 'xyz.ico' in this directory.
	 */
	public static final void register(String prefix, HashMap<String, String> map, File appFile, File iconDir) {
		if (Platform.isWindows()) {
			Thread thread = new Thread(new WindowsRegistry(prefix, map, appFile, iconDir), WindowsRegistry.class.getSimpleName());

			thread.setPriority(Thread.NORM_PRIORITY);
			thread.setDaemon(true);
			thread.start();
		}
	}

	private WindowsRegistry(String prefix, HashMap<String, String> map, File appFile, File iconDir) {
		mPrefix = prefix;
		mMap = map;
		mAppFile = appFile;
		mIconDir = iconDir;
	}

	public void run() {
		try {
			String appPath = mAppFile.getCanonicalPath().replaceAll("\\\\", "\\\\\\\\"); //$NON-NLS-1$ //$NON-NLS-2$
			File regFile = File.createTempFile("reg", ".reg"); //$NON-NLS-1$ //$NON-NLS-2$
			PrintWriter writer = new PrintWriter(regFile);

			writer.println("REGEDIT4"); //$NON-NLS-1$
			writer.println();
			for (String key : mMap.keySet()) {
				writeRegistryEntry(writer, appPath, key);
			}
			writer.close();

			(new ProcessBuilder("regedit", "/S", regFile.getCanonicalPath())).start().waitFor(); //$NON-NLS-1$ //$NON-NLS-2$
			regFile.delete();
		} catch (Exception exception) {
			// Ignore
		}
	}

	private void writeRegistryEntry(PrintWriter writer, String appPath, String extension) throws IOException {
		String upper = extension.toUpperCase();
		StringBuilder builder = new StringBuilder("HKEY_CLASSES_ROOT\\"); //$NON-NLS-1$

		writer.println("[-" + builder.toString() + DOT + extension + END_BRACKET); //$NON-NLS-1$
		writer.println();

		writer.println(START_BRACKET + builder.toString() + DOT + extension + END_BRACKET);
		writer.println("@=\"" + mPrefix + UNDERSCORE + upper + "\""); //$NON-NLS-1$ //$NON-NLS-2$
		writer.println();

		builder.append(mPrefix + UNDERSCORE + upper);
		writer.println("[-" + builder.toString() + END_BRACKET); //$NON-NLS-1$
		writer.println();

		writer.println(START_BRACKET + builder.toString() + END_BRACKET);
		writer.println("@=\"" + mMap.get(extension) + "\""); //$NON-NLS-1$ //$NON-NLS-2$
		writer.println();

		writer.println(START_BRACKET + builder.toString() + "\\DefaultIcon]"); //$NON-NLS-1$
		writer.println("@=\"\\\"" + (new File(mIconDir, extension + ".ico")).getCanonicalPath().replaceAll("\\\\", "\\\\\\\\") + "\\\"\""); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$ //$NON-NLS-5$
		writer.println();

		builder.append("\\shell"); //$NON-NLS-1$
		writer.println(START_BRACKET + builder.toString() + END_BRACKET);
		writer.println("@=\"open\""); //$NON-NLS-1$
		writer.println();

		builder.append("\\open"); //$NON-NLS-1$
		writer.println(START_BRACKET + builder.toString() + END_BRACKET);
		writer.println("@=\"&Open\""); //$NON-NLS-1$
		writer.println();

		builder.append("\\command"); //$NON-NLS-1$
		writer.println(START_BRACKET + builder.toString() + END_BRACKET);
		writer.println("@=\"\\\"" + appPath + "\\\" \\\"%1\\\"\""); //$NON-NLS-1$ //$NON-NLS-2$
		writer.println();
	}
}
