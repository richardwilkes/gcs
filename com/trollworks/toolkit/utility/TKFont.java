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

package com.trollworks.toolkit.utility;

import com.trollworks.toolkit.io.TKPreferences;

import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.GraphicsEnvironment;
import java.awt.font.FontRenderContext;
import java.util.HashMap;
import java.util.StringTokenizer;

/** Provides standardized font access and utilities. */
public class TKFont {
	private static final String					COMMA				= ",";							//$NON-NLS-1$
	private static final String					MODULE				= "Font";						//$NON-NLS-1$
	/** The default control font. */
	public static final String					CONTROL_FONT_KEY	= "Control";					//$NON-NLS-1$
	/** The default text font. */
	public static final String					TEXT_FONT_KEY		= "Text";						//$NON-NLS-1$
	/** The default monospaced text font. */
	public static final String					MONO_TEXT_FONT_KEY	= "Monospaced";				//$NON-NLS-1$
	/** The default menu item font. */
	public static final String					MENU_FONT_KEY		= "Menu";						//$NON-NLS-1$
	/** The default menu item key equivalents font. */
	public static final String					MENU_KEY_FONT_KEY	= "MenuKey";					//$NON-NLS-1$
	private static final HashMap<String, Font>	MAP					= new HashMap<String, Font>();
	private static final HashMap<String, Font>	DEFAULT_MAP			= new HashMap<String, Font>();
	private static String						DEFAULT_FONT_NAME;
	private static Font							LAST_GASP_FONT;
	private static FontRenderContext			RENDER_CONTEXT;

	static {
		int fontSize;
		String name;

		if (TKPlatform.isMacintosh()) {
			name = "Lucida Grande"; //$NON-NLS-1$
			fontSize = 11;
		} else if (TKPlatform.isWindows()) {
			name = "Tahoma"; //$NON-NLS-1$
			fontSize = 11;
		} else {
			name = "Lucida Sans"; //$NON-NLS-1$
			fontSize = 12;
		}
		if (!isValidFontName(name)) {
			name = "SansSerif"; //$NON-NLS-1$
		}
		DEFAULT_FONT_NAME = name;
		LAST_GASP_FONT = new Font(name, Font.PLAIN, fontSize);
		register(CONTROL_FONT_KEY, new Font(name, TKPlatform.isLinux() ? Font.BOLD : Font.PLAIN, fontSize));
		register(TEXT_FONT_KEY, LAST_GASP_FONT);
		register(MONO_TEXT_FONT_KEY, new Font("Monospaced", Font.PLAIN, fontSize)); //$NON-NLS-1$
		register(MENU_FONT_KEY, new Font(name, TKPlatform.isLinux() ? Font.BOLD : Font.PLAIN, fontSize));
		register(MENU_KEY_FONT_KEY, new Font(name, Font.PLAIN, 9));
	}

	/** Restores the default fonts. */
	public static void restoreDefaults() {
		MAP.clear();
		MAP.putAll(DEFAULT_MAP);
	}

	/** @return Whether the fonts are currently at their default values or not. */
	public static boolean isSetToDefaults() {
		return MAP.equals(DEFAULT_MAP);
	}

	/** @return The default font name to use. */
	public static String getDefaultFontName() {
		return DEFAULT_FONT_NAME;
	}

	/**
	 * Looks up a font by keyed name.
	 * 
	 * @param name The name the font was registered under.
	 * @return The font.
	 */
	public static Font lookup(String name) {
		Font font = MAP.get(name);

		return font != null ? font : LAST_GASP_FONT;
	}

	/**
	 * Registers a font under a specific name.
	 * 
	 * @param name The name to register the font with.
	 * @param font The font.
	 * @return The previous value of <code>name</code>.
	 */
	public static Font register(String name, Font font) {
		Font old = MAP.get(name);

		if (font == null) {
			MAP.remove(name);
			DEFAULT_MAP.remove(name);
		} else {
			MAP.put(name, font);
			if (old == null) {
				DEFAULT_MAP.put(name, font);
			}
		}
		return old;
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
		Graphics2D g2d = TKGraphics.getGraphics();
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
			defaultValue = LAST_GASP_FONT;
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
			RENDER_CONTEXT = new FontRenderContext(null, TKGraphics.isAntiAliasFontsOn(), TKGraphics.isFractionalMetricsOn());
		}
		return RENDER_CONTEXT;
	}

	/** Resets the global font render context. */
	public static void resetRenderContext() {
		RENDER_CONTEXT = null;
	}

	/**
	 * @param font The font to check.
	 * @return <code>true</code> if the specified font is monospaced.
	 */
	public static boolean isMonospaced(Font font) {
		font = font.deriveFont(12f);
		if (font.canDisplay('m') && font.canDisplay('i')) {
			FontRenderContext frc = getRenderContext();

			return font.getStringBounds("i", frc).getWidth() == font.getStringBounds("m", frc).getWidth(); //$NON-NLS-1$ //$NON-NLS-2$
		}
		return false;
	}

	/**
	 * @param name The name to check.
	 * @return <code>true</code> if the specified name is a valid font name.
	 */
	public static boolean isValidFontName(String name) {
		String[] fontList = GraphicsEnvironment.getLocalGraphicsEnvironment().getAvailableFontFamilyNames();

		for (String element : fontList) {
			if (element.equalsIgnoreCase(name)) {
				return true;
			}
		}
		return false;
	}

	/** Loads the current font settings from the preferences file. */
	public static void loadFromPreferences() {
		TKPreferences prefs = TKPreferences.getInstance();
		HashMap<String, Font> map = new HashMap<String, Font>();

		for (String name : prefs.getModuleKeys(MODULE)) {
			Font font = prefs.getFontValue(MODULE, name);

			if (font != null) {
				map.put(name, font);
			}
		}

		for (String name : DEFAULT_MAP.keySet()) {
			if (!map.containsKey(name)) {
				map.put(name, DEFAULT_MAP.get(name));
			}
		}

		for (String name : map.keySet()) {
			if (!DEFAULT_MAP.containsKey(name)) {
				DEFAULT_MAP.put(name, map.get(name));
			}
		}

		MAP.clear();
		MAP.putAll(map);
	}

	/** Saves the current font settings to the preferences file. */
	public static void saveToPreferences() {
		TKPreferences prefs = TKPreferences.getInstance();

		prefs.removePreferences(MODULE);
		for (String name : MAP.keySet()) {
			prefs.setValue(MODULE, name, MAP.get(name));
		}
	}
}
