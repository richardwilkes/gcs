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

package com.trollworks.gcs.app;

import java.awt.BorderLayout;
import java.awt.Button;
import java.awt.Color;
import java.awt.Frame;
import java.awt.TextArea;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.WindowEvent;
import java.awt.event.WindowListener;
import java.lang.reflect.Method;
import java.text.MessageFormat;

/**
 * The main entry point for the character sheet. This stub class is intended to be compiled with
 * Java 1.1, allowing it to be loaded and executed even on very old JVM's. Should the minimum
 * requirements not be achieved, then a simple error window is displayed, and the program exits.
 * <p>
 * We do not use any classes from the com.trollworks tree, since those are all compiled with a Java
 * version greater than 1.1. This means things like localization is not performed for this class.
 */
public class GCS {
	/**
	 * The main entry point for the character sheet.
	 * 
	 * @param args The command line arguments.
	 */
	public static void main(String[] args) {
		if (checkJavaVersion("1.5")) { //$NON-NLS-1$
			try {
				Method method = Class.forName("com.trollworks.gcs.app.Main").getMethod("main", new Class[] { String[].class }); //$NON-NLS-1$ //$NON-NLS-2$
				method.invoke(null, new Object[] { args });
			} catch (Throwable throwable) {
				throwable.printStackTrace();
				error("Unable to load the main GCS class."); //$NON-NLS-1$
			}
		}
	}

	private static void error(String msg) {
		try {
			Frame frame = new Frame("Error"); //$NON-NLS-1$
			TextArea text = new TextArea(msg);
			Button button = new Button("OK"); //$NON-NLS-1$

			text.setEditable(false);
			text.setBackground(Color.white);
			text.setForeground(Color.black);

			button.addActionListener(new ActionListener() {
				public void actionPerformed(ActionEvent event) {
					System.exit(0);
				}
			});

			frame.setLayout(new BorderLayout());
			frame.add(text, BorderLayout.CENTER);
			frame.add(button, BorderLayout.SOUTH);
			frame.pack();
			frame.addWindowListener(new WindowListener() {
				// Didn't use a WindowAdapter, since that would have required me (with the settings
				// I use in Eclipse) to mark the one method I'm actually using with an override
				// notation, which is not compatible with Java 1.1.
				public void windowClosing(WindowEvent event) {
					System.exit(0);
				}

				public void windowActivated(WindowEvent event) {
					// Not used.
				}

				public void windowClosed(WindowEvent event) {
					// Not used.
				}

				public void windowDeactivated(WindowEvent event) {
					// Not used.
				}

				public void windowDeiconified(WindowEvent event) {
					// Not used.
				}

				public void windowIconified(WindowEvent event) {
					// Not used.
				}

				public void windowOpened(WindowEvent event) {
					// Not used.
				}
			});
			frame.setVisible(true);
		} catch (Throwable throwable) {
			System.err.println(msg);
		}
	}

	private static boolean checkJavaVersion(String minimumVersion) {
		String javaVersion = System.getProperty("java.version"); //$NON-NLS-1$
		if (extractVersion(javaVersion) < extractVersion(minimumVersion)) {
			error(MessageFormat.format("The currently installed version of Java is {0}.\nThis software requires Java version {1} or greater.\n", new Object[] { javaVersion, minimumVersion })); //$NON-NLS-1$
			return false;
		}
		return true;
	}

	private static long extractVersion(String versionString) {
		char[] chars = versionString.toCharArray();
		long version = 0;
		int shift = 48;
		long value = 0;

		for (int i = 0; i < chars.length; i++) {
			char ch = chars[i];
			if (ch >= '0' && ch <= '9') {
				value *= 10;
				value += ch - '0';
			} else if (ch == '.' || ch == '_') {
				if (value > 0xEFFF) {
					value = 0xEFFF;
				}
				version |= value << shift;
				value = 0;
				shift -= 16;
				if (shift < 0) {
					break;
				}
			} else {
				if (value > 0xEFFF) {
					value = 0xEFFF;
				}
				version |= value << shift;
				value = 0;
				shift -= 16;
				break;
			}
		}
		if (shift >= 0) {
			if (value > 0xEFFF) {
				value = 0xEFFF;
			}
			version |= value << shift;
		}
		return version;
	}
}
