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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character.names;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.URL;
import java.util.Random;

/** An abstract base class for name generation. */
public abstract class Names {
	/** A random number generator that can be shared amongst instances of the {@link Names} class. */
	protected static final Random	RANDOM	= new Random();

	/** @return A newly generated male first name. */
	public abstract String getMaleFirstName();

	/** @return A newly generated female first name. */
	public abstract String getFemaleFirstName();

	/**
	 * @param male Whether to generate a male or female name.
	 * @return A newly generated first name.
	 */
	public String getGivenName(boolean male) {
		return male ? getMaleFirstName() : getFemaleFirstName();
	}

	/** @return A newly generated last name. */
	public abstract String getLastName();

	/**
	 * @param male Whether to generate a male or female name.
	 * @return A newly generated full (first and last) name.
	 */
	public String getFullName(boolean male) {
		return getGivenName(male) + " " + getLastName(); //$NON-NLS-1$
	}

	/**
	 * Loads names from a file into an array of strings. The names in the file should be listed one
	 * per line.
	 * 
	 * @param url The {@link URL} to load the name data from.
	 * @param fallback A single name to use in case the file couldn't be loaded.
	 * @return An array of strings representing the contents of the file.
	 */
	protected static final String[] loadNames(URL url, String fallback) {
		String[] names;

		try {
			BufferedReader in = new BufferedReader(new InputStreamReader(url.openStream()));
			String line = in.readLine();
			int count = 0;

			while (line != null) {
				if (line.trim().length() > 0) {
					count++;
				}
				line = in.readLine();
			}
			in.close();

			names = new String[count];
			in = new BufferedReader(new InputStreamReader(url.openStream()));
			line = in.readLine();
			count = 0;
			while (line != null) {
				line = line.trim();
				if (line.length() > 0) {
					names[count++] = line;
				}
				line = in.readLine();
			}
			in.close();
		} catch (Exception exception) {
			// This should never occur... but if it does, we won't fail.
			names = new String[] { fallback };
		}

		return names;
	}
}
