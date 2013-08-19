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

package com.trollworks.gcs.utility;

import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.widgets.GraphicsUtilities;

import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.GraphicsEnvironment;
import java.awt.font.FontRenderContext;
import java.util.HashMap;
import java.util.Map.Entry;
import java.util.StringTokenizer;

import javax.swing.UIManager;

/** Provides standardized font access and utilities. */
public class Fonts {
	private static final String					COMMA					= ",";								//$NON-NLS-1$
	private static final String					MODULE					= "Font";							//$NON-NLS-1$
	/** The standard text field font. */
	public static final String					KEY_STD_TEXT_FIELD		= "TextField.font";				//$NON-NLS-1$
	/** The label font. */
	public static final String					KEY_LABEL				= "trollworks.label";				//$NON-NLS-1$
	/** The field font. */
	public static final String					KEY_FIELD				= "trollworks.field";				//$NON-NLS-1$
	/** The field notes font. */
	public static final String					KEY_FIELD_NOTES			= "trollworks.field.notes";		//$NON-NLS-1$
	/** The technique field font. */
	public static final String					KEY_TECHNIQUE_FIELD		= "trollworks.field.technique";	//$NON-NLS-1$
	/** The primary footer font. */
	public static final String					KEY_PRIMARY_FOOTER		= "trollworks.footer.primary";		//$NON-NLS-1$
	/** The secondary footer font. */
	public static final String					KEY_SECONDARY_FOOTER	= "trollworks.footer.secondary";	//$NON-NLS-1$
	/** The notes font. */
	public static final String					KEY_NOTES				= "trollworks.notes";				//$NON-NLS-1$
	private static final HashMap<String, Font>	DEFAULT_MAP				= new HashMap<String, Font>();
	private static FontRenderContext			RENDER_CONTEXT;

	static {
		String name = getDefaultFontName();
		setDefault(KEY_LABEL, new Font(name, Font.PLAIN, 9));
		setDefault(KEY_FIELD, new Font(name, Font.BOLD, 9));
		setDefault(KEY_FIELD_NOTES, new Font(name, Font.PLAIN, 8));
		setDefault(KEY_TECHNIQUE_FIELD, new Font(name, Font.BOLD + Font.ITALIC, 9));
		setDefault(KEY_PRIMARY_FOOTER, new Font(name, Font.BOLD, 8));
		setDefault(KEY_SECONDARY_FOOTER, new Font(name, Font.PLAIN, 6));
		setDefault(KEY_NOTES, new Font(name, Font.PLAIN, 9));
	}

	private static void setDefault(String key, Font font) {
		UIManager.put(key, font);
		DEFAULT_MAP.put(key, font);
	}

	/** @return The default font name to use. */
	public static String getDefaultFontName() {
		return UIManager.getFont(KEY_STD_TEXT_FIELD).getName();
	}

	/** Restores the default fonts. */
	public static void restoreDefaults() {
		for (Entry<String, Font> entry : DEFAULT_MAP.entrySet()) {
			UIManager.put(entry.getKey(), entry.getValue());
		}
	}

	/** @return Whether the fonts are currently at their default values or not. */
	public static boolean isSetToDefaults() {
		for (Entry<String, Font> entry : DEFAULT_MAP.entrySet()) {
			if (!UIManager.getFont(entry.getKey()).equals(entry.getValue())) {
				return false;
			}
		}
		return true;
	}

	/**
	 * @param font The font to work on.
	 * @return The specified font as a canonical string.
	 */
	public static String getStringValue(Font font) {
		return font.getName() + COMMA + font.getStyle() + COMMA + font.getSize();
	}

	/**
	 * @param font The font to work on.
	 * @return The font metrics for the specified font.
	 */
	public static FontMetrics getFontMetrics(Font font) {
		Graphics2D g2d = GraphicsUtilities.getGraphics();
		FontMetrics fm = g2d.getFontMetrics(font);
		g2d.dispose();
		return fm;
	}

	/**
	 * @param buffer The string to create the font from.
	 * @return A font created from the specified string.
	 */
	public static Font create(String buffer) {
		return create(buffer, null);
	}

	/**
	 * @param buffer The string to create the font from.
	 * @param defaultValue The value to use if the string is invalid.
	 * @return A font created from the specified string.
	 */
	public static Font create(String buffer, Font defaultValue) {
		if (defaultValue == null) {
			defaultValue = UIManager.getFont(KEY_STD_TEXT_FIELD);
		}

		String name = defaultValue.getName();
		int style = defaultValue.getStyle();
		int size = defaultValue.getSize();

		if (buffer != null && buffer.length() > 0) {
			StringTokenizer tokenizer = new StringTokenizer(buffer, COMMA);

			if (tokenizer.hasMoreTokens()) {
				name = tokenizer.nextToken();
				if (!isValidFontName(name)) {
					name = defaultValue.getName();
				}
				if (tokenizer.hasMoreTokens()) {
					buffer = tokenizer.nextToken();
					try {
						style = Integer.parseInt(buffer);
					} catch (NumberFormatException nfe1) {
						// We'll use the default style instead
					}
					if (style < 0 || style > 3) {
						style = defaultValue.getStyle();
					}
					if (tokenizer.hasMoreTokens()) {
						buffer = tokenizer.nextToken();
						try {
							size = Integer.parseInt(buffer);
						} catch (NumberFormatException nfe1) {
							// We'll use the default size instead
						}
						if (size < 1) {
							size = 1;
						} else if (size > 200) {
							size = 200;
						}
					}
				}
			}
		}

		return new Font(name, style, size);
	}

	/** @return A global font rendering context. */
	public static FontRenderContext getRenderContext() {
		if (RENDER_CONTEXT == null) {
			RENDER_CONTEXT = new FontRenderContext(null, Platform.isMacintosh() || Platform.isWindows(), false);
		}
		return RENDER_CONTEXT;
	}

	/** Resets the global font render context. */
	public static void resetRenderContext() {
		RENDER_CONTEXT = null;
	}

	/**
	 * @param name The name to check.
	 * @return <code>true</code> if the specified name is a valid font name.
	 */
	public static boolean isValidFontName(String name) {
		for (String element : GraphicsEnvironment.getLocalGraphicsEnvironment().getAvailableFontFamilyNames()) {
			if (element.equalsIgnoreCase(name)) {
				return true;
			}
		}
		return false;
	}

	/** Loads the current font settings from the preferences file. */
	public static void loadFromPreferences() {
		Preferences prefs = Preferences.getInstance();
		for (String name : DEFAULT_MAP.keySet()) {
			Font font = prefs.getFontValue(MODULE, name);
			if (font != null) {
				UIManager.put(name, font);
			}
		}
	}

	/** Saves the current font settings to the preferences file. */
	public static void saveToPreferences() {
		Preferences prefs = Preferences.getInstance();
		prefs.removePreferences(MODULE);
		for (String name : DEFAULT_MAP.keySet()) {
			Font font = UIManager.getFont(name);
			if (font != null) {
				prefs.setValue(MODULE, name, font);
			}
		}
	}
}
