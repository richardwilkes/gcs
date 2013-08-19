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

package com.trollworks.gcs.utility.io;

import com.trollworks.gcs.app.Main;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.Fonts;
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.text.Conversion;

import java.awt.Dimension;
import java.awt.Font;
import java.awt.Point;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.Enumeration;
import java.util.List;
import java.util.Properties;

/** Provides the implementation of preferences. */
public class Preferences {
	private static String		MSG_UNINITIALIZED;
	private static final String	VERSION_KEY	= "Version";	//$NON-NLS-1$
	private static final String	DOT			= ".";			//$NON-NLS-1$
	private static Preferences	INSTANCE	= null;
	private boolean				mDirty;
	private Properties			mPrefs;
	private File				mFile;
	private Notifier			mNotifier;

	static {
		LocalizedMessages.initialize(Preferences.class);
	}

	/** @return The default, global, preferences. */
	public static synchronized Preferences getInstance() {
		if (INSTANCE == null) {
			throw new RuntimeException(MSG_UNINITIALIZED);
		}
		return INSTANCE;
	}

	/**
	 * Sets the preference file to use. The default, global, preferences are not operational until
	 * this is done.
	 * 
	 * @param prefsFile The file containing preferences.
	 */
	public static synchronized void setPreferenceFile(File prefsFile) {
		if (INSTANCE == null) {
			INSTANCE = new Preferences(prefsFile);
		}
	}

	/**
	 * Sets the preference file to use. The default, global, preferences are not operational until
	 * this is done.
	 * 
	 * @param leafName The leaf file name to use, such as "app.prf".
	 */
	public static synchronized void setPreferenceFile(String leafName) {
		setPreferenceFile(getDefaultPreferenceFile(leafName));
	}

	/**
	 * Sets the preference file to use. The default, global, preferences are not operational until
	 * this is done.
	 * 
	 * @param packageName The package name to use, if this preference is part of a set of
	 *            preferences.
	 * @param leafName The leaf file name to use, such as "app.prf".
	 */
	public static synchronized void setPreferenceFile(String packageName, String leafName) {
		setPreferenceFile(getDefaultPreferenceFile(packageName, leafName));
	}

	/**
	 * @param leafName The leaf file name to use, such as "appname.prf".
	 * @return <code>true</code> if the module has preferences set.
	 */
	public static File getDefaultPreferenceFile(String leafName) {
		return getDefaultPreferenceFile(null, leafName);
	}

	/**
	 * @param packageDir The package name to use, if this preference is part of a set of
	 *            preferences.
	 * @param leafName The leaf file name to use, such as "appname.prf".
	 * @return The default preference file for this platform.
	 */
	public static File getDefaultPreferenceFile(String packageDir, String leafName) {
		File base = new File(System.getProperty("user.home", DOT)); //$NON-NLS-1$

		if (Platform.isMacintosh()) {
			base = new File(base, "Library/Preferences"); //$NON-NLS-1$
		} else if (Platform.isWindows()) {
			base = new File(base, "Local Settings/Application Data"); //$NON-NLS-1$
		} else {
			if (packageDir != null) {
				if (!packageDir.startsWith(DOT)) {
					packageDir = DOT + packageDir;
				}
			} else {
				if (!leafName.startsWith(DOT)) {
					leafName = DOT + leafName;
				}
			}
		}

		if (packageDir != null) {
			base = new File(base, packageDir);
		}

		base.mkdirs(); // Ensure the directory exists...

		return new File(base, leafName);
	}

	/**
	 * @param module The module to use.
	 * @param key The key to use.
	 * @return The key as it is used within the properties file.
	 */
	public static String getModuleKey(String module, String key) {
		return module + '.' + key;
	}

	/**
	 * Loads/creates the preferences database.
	 * 
	 * @param prefsFile The file containing preferences.
	 */
	public Preferences(File prefsFile) {
		mFile = prefsFile;
		mPrefs = new Properties();
		mNotifier = new Notifier();
		if (mFile.exists()) {
			try {
				InputStream in = new FileInputStream(mFile);

				mPrefs.loadFromXML(in);
				in.close();
			} catch (Exception exception) {
				// Throw away anything we loaded, since the file must be corrupted.
				mPrefs = new Properties();
			}
		}
	}

	/** @return The file used to store these preferences. */
	public File getFile() {
		return mFile;
	}

	/** @return The preference broker. */
	public Notifier getNotifier() {
		return mNotifier;
	}

	/** Mark the start of a batch change. */
	public void startBatch() {
		mNotifier.startBatch();
	}

	/** Mark the end of a batch change. */
	public void endBatch() {
		mNotifier.endBatch();
	}

	/**
	 * Disposes of this preference instance.
	 * 
	 * @param save Whether or not to save the current contents to disk.
	 */
	public void dispose(boolean save) {
		if (save) {
			save();
		}
		// Can't really dispose of the global preferences,
		// so only do the remaining cleanup for non-global preferences.
		if (INSTANCE != this) {
			mPrefs = null;
			mFile = null;
			mNotifier = null;
		}
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @return The specified preference as a boolean. <code>false</code> will be returned if the
	 *         key cannot be extracted.
	 */
	public boolean getBooleanValue(String module, String key) {
		return getBooleanValue(module, key, false);
	}

	/**
	 * @param module The tag representing the module this preference is associated with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @return The specified preference as a boolean.
	 */
	public boolean getBooleanValue(String module, String key, boolean defaultValue) {
		String buffer = getStringValueForced(module, key);

		if (Boolean.TRUE.toString().equals(buffer)) {
			return true;
		}
		if (Boolean.FALSE.toString().equals(buffer)) {
			return false;
		}
		return defaultValue;
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @return The specified preference as a {@link Dimension}. <code>null</code> will be
	 *         returned if the key cannot be extracted.
	 */
	public Dimension getDimensionValue(String module, String key) {
		return getDimensionValue(module, key, null);
	}

	/**
	 * @param module The tag representing the module this preference is associated with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @return The specified preference as a {@link Dimension}.
	 */
	public Dimension getDimensionValue(String module, String key, Dimension defaultValue) {
		Dimension dim = Conversion.extractDimension(getStringValue(module, key));

		return dim == null ? defaultValue : dim;
	}

	/**
	 * @param module The tag representing the module this preference is associated with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @return The specified preference as a double.
	 */
	public double getDoubleValue(String module, String key, double defaultValue) {
		return getDoubleValue(module, key, defaultValue, Double.MIN_VALUE, Double.MAX_VALUE);
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @param minimum The minimum value to return.
	 * @param maximum The maximum value to return.
	 * @return The specified preference as a double.
	 */
	public double getDoubleValue(String module, String key, double defaultValue, double minimum, double maximum) {
		String buffer = getStringValue(module, key);
		double value = defaultValue;

		if (buffer != null) {
			try {
				value = Double.parseDouble(buffer);
			} catch (NumberFormatException nfe) {
				// Use the default value instead
			}
		}

		if (value < minimum) {
			value = minimum;
		}
		if (value > maximum) {
			value = maximum;
		}

		return value;
	}

	/**
	 * @param module The tag representing the module this preference is associated with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @return The specified preference as a float.
	 */
	public float getFloatValue(String module, String key, float defaultValue) {
		return getFloatValue(module, key, defaultValue, Float.MIN_VALUE, Float.MAX_VALUE);
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @param minimum The minimum value to return.
	 * @param maximum The maximum value to return.
	 * @return The specified preference as a float.
	 */
	public float getFloatValue(String module, String key, float defaultValue, float minimum, float maximum) {
		String buffer = getStringValue(module, key);
		float value = defaultValue;

		if (buffer != null) {
			try {
				value = Float.parseFloat(buffer);
			} catch (NumberFormatException nfe) {
				// Use the default value instead
			}
		}

		if (value < minimum) {
			value = minimum;
		}
		if (value > maximum) {
			value = maximum;
		}

		return value;
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @return The specified preference as a {@link Font}. <code>null</code> will be returned if
	 *         the key cannot be extracted.
	 */
	public Font getFontValue(String module, String key) {
		String value = getStringValue(module, key);

		return value == null ? null : Fonts.create(value);
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @return The specified preference as a {@link Font}.
	 */
	public Font getFontValue(String module, String key, Font defaultValue) {
		return Fonts.create(getStringValue(module, key), defaultValue);
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @return The specified preference as an integer.
	 */
	public int getIntValue(String module, String key, int defaultValue) {
		return getIntValue(module, key, defaultValue, Integer.MIN_VALUE, Integer.MAX_VALUE);
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @param minimum The minimum value to return.
	 * @param maximum The maximum value to return.
	 * @return The specified preference as an integer.
	 */
	public int getIntValue(String module, String key, int defaultValue, int minimum, int maximum) {
		String buffer = getStringValue(module, key);
		int value = defaultValue;

		if (buffer != null) {
			try {
				value = Integer.parseInt(buffer);
			} catch (NumberFormatException nfe) {
				// Use the default value instead
			}
		}

		if (value < minimum) {
			value = minimum;
		}
		if (value > maximum) {
			value = maximum;
		}

		return value;
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @return The specified preference as a long.
	 */
	public long getLongValue(String module, String key, long defaultValue) {
		return getLongValue(module, key, defaultValue, Long.MIN_VALUE, Long.MAX_VALUE);
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @param minimum The minimum value to return.
	 * @param maximum The maximum value to return.
	 * @return The specified preference as a long.
	 */
	public long getLongValue(String module, String key, long defaultValue, long minimum, long maximum) {
		String buffer = getStringValue(module, key);
		long value = defaultValue;

		if (buffer != null) {
			try {
				value = Long.parseLong(buffer);
			} catch (NumberFormatException nfe) {
				// Use the default value instead
			}
		}

		if (value < minimum) {
			value = minimum;
		}
		if (value > maximum) {
			value = maximum;
		}

		return value;
	}

	/**
	 * @param module The module to use.
	 * @return All of the keys for the specified module.
	 */
	public List<String> getModuleKeys(String module) {
		ArrayList<String> keys = new ArrayList<String>();
		Enumeration<Object> keyEnum = mPrefs.keys();

		while (keyEnum.hasMoreElements()) {
			String key = (String) keyEnum.nextElement();

			if (key.startsWith(module)) {
				key = key.substring(key.indexOf('.') + 1);
				if (key != null && key.length() > 0) {
					keys.add(key);
				}
			}
		}

		return keys;
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @return The specified preference as a {@link Point}.<code>null</code> will be returned if
	 *         the key cannot be extracted.
	 */
	public Point getPointValue(String module, String key) {
		return getPointValue(module, key, null);
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @return The specified preference as a {@link Point}.
	 */
	public Point getPointValue(String module, String key, Point defaultValue) {
		Point pt = Conversion.extractPoint(getStringValue(module, key));

		return pt == null ? defaultValue : pt;
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @return The specified general preference as a {@link String}. <code>null</code> may be
	 *         returned.
	 */
	public String getStringValue(String module, String key) {
		return mPrefs.getProperty(getModuleKey(module, key));
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param defaultValue The default value to return if the key cannot be extracted.
	 * @return The specified preference as a {@link String}.
	 */
	public String getStringValue(String module, String key, String defaultValue) {
		String value = getStringValue(module, key);

		return value == null ? defaultValue : value;
	}

	/**
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @return The specified preference as a {@link String}. If a <code>null</code> value would
	 *         normally be returned, returns an empty string instead.
	 */
	public String getStringValueForced(String module, String key) {
		return getStringValue(module, key, ""); //$NON-NLS-1$
	}

	/**
	 * @param module The module to work with.
	 * @return The version of the specified module.
	 */
	public int getVersion(String module) {
		return getIntValue(module, VERSION_KEY, 1);
	}

	/**
	 * @param module The module to check.
	 * @return <code>true</code> if the module has preferences set.
	 */
	public boolean hasPreferences(String module) {
		module += '.';

		for (Enumeration<Object> keys = mPrefs.keys(); keys.hasMoreElements();) {
			String key = (String) keys.nextElement();

			if (key.startsWith(module)) {
				return true;
			}
		}
		return false;
	}

	/**
	 * Removes the specified preference.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 */
	public void removePreference(String module, String key) {
		setValue(module, key, (String) null);
	}

	/**
	 * Removes all preferences for the specified module.
	 * 
	 * @param module The module to remove preferences from.
	 */
	public void removePreferences(String module) {
		String prefix = module + '.';
		ArrayList<String> list = new ArrayList<String>();
		int length = prefix.length();

		for (Enumeration<Object> keys = mPrefs.keys(); keys.hasMoreElements();) {
			String key = (String) keys.nextElement();

			if (key.startsWith(prefix)) {
				list.add(key.substring(length));
			}
		}

		if (!list.isEmpty()) {
			startBatch();
			mDirty = true;
			for (String keyToRemove : list) {
				removePreference(module, keyToRemove);
			}
			endBatch();
		}
	}

	/**
	 * Removes the preferences for the specified module if the version passed in does not match
	 * those in preferences.
	 * 
	 * @param module The module to check.
	 * @param version The current version.
	 */
	public void resetIfVersionMisMatch(String module, int version) {
		if (getVersion(module) != version) {
			startBatch();
			removePreferences(module);
			setVersion(module, version);
			endBatch();
		}
	}

	/**
	 * Saves the preference information to disk.
	 * 
	 * @return Whether the save was successful or not.
	 */
	public boolean save() {
		if (mDirty) {
			try {
				SafeFileUpdater trans = new SafeFileUpdater();

				trans.begin();
				try {
					File file = trans.getTransactionFile(mFile);
					OutputStream out = new FileOutputStream(file);

					mPrefs.storeToXML(out, Main.getName());
					out.close();
				} catch (IOException ioe) {
					trans.abort();
					throw ioe;
				}
				trans.commit();
				mDirty = false;
			} catch (Exception exception) {
				return false;
			}
		}
		return true;
	}

	/**
	 * Sets the specified preference as a boolean.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param value The value to set the key to.
	 */
	public void setValue(String module, String key, boolean value) {
		setValue(module, key, new Boolean(value).toString());
	}

	/**
	 * Sets the specified preference.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param value The value to set the key to. Pass <code>null</code> to remove the key.
	 */
	public void setValue(String module, String key, String value) {
		key = getModuleKey(module, key);
		if (value != null) {
			if (!value.equals(mPrefs.getProperty(key))) {
				mPrefs.setProperty(key, value);
				mDirty = true;
				mNotifier.notify(this, key, value);
			}
		} else if (mPrefs.getProperty(key) != null) {
			mPrefs.remove(key);
			mDirty = true;
			mNotifier.notify(this, key, null);
		}
	}

	/**
	 * Sets the specified preference as a <code>Dimension</code>.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param value The value to set the key to.
	 */
	public void setValue(String module, String key, Dimension value) {
		setValue(module, key, value != null ? Conversion.createString(value) : null);
	}

	/**
	 * Sets the specified preference as a <code>Font</code>.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param value The value to set the key to.
	 */
	public void setValue(String module, String key, Font value) {
		setValue(module, key, value != null ? Fonts.getStringValue(value) : null);
	}

	/**
	 * Sets the specified preference as an integer.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param value The value to set the key to.
	 */
	public void setValue(String module, String key, int value) {
		setValue(module, key, Integer.toString(value));
	}

	/**
	 * Sets the specified preference as a long.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param value The value to set the key to.
	 */
	public void setValue(String module, String key, long value) {
		setValue(module, key, Long.toString(value));
	}

	/**
	 * Sets the specified preference as a double.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param value The value to set the key to.
	 */
	public void setValue(String module, String key, double value) {
		setValue(module, key, Double.toString(value));
	}

	/**
	 * Sets the specified preference as a float.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param value The value to set the key to.
	 */
	public void setValue(String module, String key, float value) {
		setValue(module, key, Float.toString(value));
	}

	/**
	 * Sets the specified preference as a <code>Point</code>.
	 * 
	 * @param module The module to work with.
	 * @param key The key this preference is tied to.
	 * @param value The value to set the key to.
	 */
	public void setValue(String module, String key, Point value) {
		setValue(module, key, value != null ? Conversion.createString(value) : null);
	}

	/**
	 * Sets the version of the specified module.
	 * 
	 * @param module The module to work with.
	 * @param version The value to set the version to.
	 */
	public void setVersion(String module, int version) {
		setValue(module, VERSION_KEY, version);
	}
}
