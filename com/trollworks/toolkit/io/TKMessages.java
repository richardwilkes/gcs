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

package com.trollworks.toolkit.io;

import java.io.IOException;
import java.io.InputStream;
import java.lang.reflect.Field;
import java.lang.reflect.Modifier;
import java.security.AccessController;
import java.security.PrivilegedAction;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Locale;
import java.util.Properties;

/**
 * Loads localized messages into classes. This provides similar functionality to the
 * org.eclipse.osgi.util.NLS class, but this way we can get localized strings without requiring any
 * part of Eclipse.
 */
public class TKMessages extends Properties implements PrivilegedAction<Object> {
	private static final String		IN				= "\" in: ";							//$NON-NLS-1$
	private static final String		EXTENSION		= ".properties";						//$NON-NLS-1$
	private static final int		MOD_EXPECTED	= Modifier.PUBLIC | Modifier.STATIC;
	private static final int		MOD_MASK		= MOD_EXPECTED | Modifier.FINAL;
	private static final String[]	SUFFIXES;
	private Class<?>				mClass;
	private HashSet<Field>			mInitialized;
	private HashMap<String, Field>	mFields;
	private String					mBundleName;
	private boolean					mIsAccessible;

	static {
		String nl = Locale.getDefault().toString();
		ArrayList<String> result = new ArrayList<String>(4);
		int lastSeparator;

		while (true) {
			result.add('_' + nl + EXTENSION);
			lastSeparator = nl.lastIndexOf('_');
			if (lastSeparator == -1) {
				break;
			}
			nl = nl.substring(0, lastSeparator);
		}

		result.add(EXTENSION);
		SUFFIXES = result.toArray(new String[result.size()]);
	}

	/**
	 * Initialize the specified class with the values from its message bundle.
	 * 
	 * @param theClass The class to process.
	 */
	public static void initialize(final Class<?> theClass) {
		TKMessages nls = new TKMessages(theClass);

		if (System.getSecurityManager() == null) {
			nls.run();
		} else {
			AccessController.doPrivileged(nls);
		}
	}

	private TKMessages(Class<?> theClass) {
		mClass = theClass;
		mInitialized = new HashSet<Field>();
		mFields = new HashMap<String, Field>();
		mBundleName = theClass.getName();
		mIsAccessible = (mClass.getModifiers() & Modifier.PUBLIC) != 0;
	}

	public Object run() {
		Field[] fieldArray = mClass.getDeclaredFields();
		ClassLoader loader = mClass.getClassLoader();
		String root = mBundleName.replace('.', '/');
		String[] variants = new String[SUFFIXES.length];
		int len = fieldArray.length;

		for (int i = 0; i < len; i++) {
			mFields.put(fieldArray[i].getName(), fieldArray[i]);
		}

		for (int i = 0; i < variants.length; i++) {
			variants[i] = root + SUFFIXES[i];
		}

		for (String variant : variants) {
			InputStream input = loader == null ? ClassLoader.getSystemResourceAsStream(variant) : loader.getResourceAsStream(variant);

			if (input != null) {
				try {
					load(input);
				} catch (IOException exception) {
					System.err.println("Error: Unable to load " + variant); //$NON-NLS-1$
					exception.printStackTrace(System.err);
				} finally {
					try {
						input.close();
					} catch (IOException exception) {
						// Ignore
					}
					clear();
				}
			}
		}

		for (int i = 0; i < len; i++) {
			Field field = fieldArray[i];

			if ((field.getModifiers() & MOD_MASK) == MOD_EXPECTED && !mInitialized.contains(field)) {
				try {
					String warning = "Warning: Missing message for \"" + field.getName() + IN + mBundleName; //$NON-NLS-1$

					System.err.println(warning);
					if (!mIsAccessible) {
						field.setAccessible(true);
					}
					field.set(null, warning);
				} catch (Exception exception) {
					// Should not be possible
				}
			}
		}

		return null;
	}

	@Override public synchronized Object put(Object key, Object value) {
		Field field = mFields.get(key);

		if (field == null) {
			System.err.println("Warning: Unused message for \"" + key + IN + mBundleName); //$NON-NLS-1$
			return null;
		}

		if (!mInitialized.contains(field)) {
			mInitialized.add(field);
			if ((field.getModifiers() & MOD_MASK) != MOD_EXPECTED) {
				System.err.println("Warning: Incorrect field modifiers for \"" + field.getName() + IN + mBundleName); //$NON-NLS-1$
				return null;
			}
			try {
				if (!mIsAccessible) {
					field.setAccessible(true);
				}
				field.set(null, value);
			} catch (Exception e) {
				// Should not be possible
			}
		}
		return null;
	}
}
