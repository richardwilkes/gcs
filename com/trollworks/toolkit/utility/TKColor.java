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

import java.awt.Color;
import java.lang.reflect.Field;

/** Provides standardized color access and utilities. */
public class TKColor extends Color {
	/** The constant for the control line color. */
	public static final TKColor	CONTROL_LINE				= new TKColor(90, 90, 90);
	/** The constant for the control shadow color. */
	public static final TKColor	CONTROL_SHADOW				= new TKColor(205, 205, 205);
	/** The constant for the control highlight color. */
	public static final TKColor	CONTROL_HIGHLIGHT			= new TKColor(Color.white);
	/** The constant for the control rollover color. */
	public static final TKColor	CONTROL_ROLL				= new TKColor(246, 246, 246);
	/** The constant for the control fill color. */
	public static final TKColor	CONTROL_FILL				= new TKColor(230, 230, 230);
	/** The constant for the control pressed fill color. */
	public static final TKColor	CONTROL_PRESSED_FILL		= new TKColor(CONTROL_SHADOW);
	/** The constant for the control icon color. */
	public static final TKColor	CONTROL_ICON				= new TKColor(41, 41, 41);
	/** The constant for the default button indicator color. */
	public static final TKColor	DEFAULT_BUTTON_INDICATOR	= new TKColor(Color.lightGray);
	/** The constant for the progress bar color. */
	public static final TKColor	PROGRESS_BAR				= new TKColor(160, 160, 224);
	/** The constant for the scroll bar fill color. */
	public static final TKColor	SCROLL_BAR_FILL				= new TKColor(197, 197, 197);
	/** The constant for the scroll bar line color. */
	public static final TKColor	SCROLL_BAR_LINE				= new TKColor(156, 156, 156);
	/** The constant for the slider track shadow. */
	public static final TKColor	SLIDER_SHADOW				= new TKColor(176, 176, 176);
	/** The constant for the menu bar shadow color. */
	public static final TKColor	MENU_BAR_SHADOW				= new TKColor(SCROLL_BAR_FILL);
	/** The constant for the menu background color. */
	public static final TKColor	MENU_BACKGROUND				= new TKColor(CONTROL_FILL);
	/** The constant for the menu selection fill color. */
	public static final TKColor	MENU_SELECTION_FILL			= new TKColor(82, 97, 172);
	/** The constant for the menu selection line color. */
	public static final TKColor	MENU_SELECTION_LINE			= new TKColor(65, 68, 106);
	/** The constant for the menu selection shadow color. */
	public static final TKColor	MENU_SELECTION_SHADOW		= new TKColor(70, 88, 170);
	/** The constant for the menu selection highlight color. */
	public static final TKColor	MENU_SELECTION_HIGHLIGHT	= new TKColor(164, 170, 238);
	/** The constant for the menu outline color. */
	public static final TKColor	MENU_OUTLINE				= new TKColor(SCROLL_BAR_LINE);
	/** The constant for the menu highlight color. */
	public static final TKColor	MENU_HIGHLIGHT				= new TKColor(CONTROL_HIGHLIGHT);
	/** The constant for the menu shadow color. */
	public static final TKColor	MENU_SHADOW					= new TKColor(CONTROL_SHADOW);
	/** The constant for the menu keystroke color. */
	public static final TKColor	MENU_KEYSTROKE				= new TKColor(MENU_SELECTION_SHADOW);
	/** The constant for the menu disabled keystroke color. */
	public static final TKColor	MENU_DISABLED_KEYSTROKE		= new TKColor(Color.gray);
	/** The constant for the menu separator color. */
	public static final TKColor	MENU_SEPARATOR				= new TKColor(64, 64, 102);
	/** The constant for the primary banding color. */
	public static final TKColor	PRIMARY_BANDING				= new TKColor(240, 240, 255);
	/** The constant for the secondary banding color. */
	public static final TKColor	SECONDARY_BANDING			= new TKColor(240, 255, 240);
	/** The constant for the light background color. */
	public static final TKColor	LIGHT_BACKGROUND			= new TKColor(CONTROL_FILL);
	/** The constant for the dark background color. */
	public static final TKColor	DARK_BACKGROUND				= new TKColor(132, 132, 132);
	/** The constant for the tooltip background color. */
	public static final TKColor	TOOLTIP_BACKGROUND			= new TKColor(255, 255, 204);
	/** The constant for the highlight color. */
	public static final TKColor	HIGHLIGHT					= new TKColor(74, 89, 164);
	/** The constant for the inactive highlight color. */
	public static final TKColor	INACTIVE_HIGHLIGHT			= new TKColor(156, 165, 189);
	/** The constant for the object highlight color. */
	public static final TKColor	OBJECT_HIGHLIGHT			= new TKColor(255, 255, 150);
	/** The constant for the marquee line color. */
	public static final TKColor	MARQUEE_LINE				= new TKColor(CONTROL_SHADOW);
	/** The constant for the invalid marker color. */
	public static final TKColor	INVALID_MARKER				= new TKColor(180, 0, 0);
	/** The constant for the tool bar shadow color. */
	public static final TKColor	TOOLBAR_SHADOW				= new TKColor(MENU_BAR_SHADOW);
	/** The constant for the text color. */
	public static final TKColor	TEXT						= new TKColor(Color.black);
	/** The constant for the text background color. */
	public static final TKColor	TEXT_BACKGROUND				= new TKColor(Color.white);
	/** The constant for the disabled text background color. */
	public static final TKColor	DISABLED_TEXT_BACKGROUND	= new TKColor(Color.lightGray);
	/** The constant for the highlighted text color. */
	public static final TKColor	HIGHLIGHTED_TEXT			= new TKColor(Color.white);
	/** The constant for the expand bar title fill color. */
	public static final TKColor	EXPANDBAR_FILL				= new TKColor(214, 214, 214);
	/** The constant for the expand bar title line color. */
	public static final TKColor	EXPANDBAR_LINE				= new TKColor(MENU_BAR_SHADOW);
	private static final String	NO_COLOR					= "null";								//$NON-NLS-1$

	/**
	 * @param color1 The first color.
	 * @param color2 The second color.
	 * @param percentage How much of the second color to use.
	 * @return A color that is a blended version of the two passed in.
	 */
	public static final Color blend(Color color1, Color color2, int percentage) {
		int remaining = 100 - percentage;

		return new Color((color1.getRed() * remaining + color2.getRed() * percentage) / 100, (color1.getGreen() * remaining + color2.getGreen() * percentage) / 100, (color1.getBlue() * remaining + color2.getBlue() * percentage) / 100);
	}

	/**
	 * @param color Return an intensified version of this color.
	 * @return A color that is brighter than the color passed in.
	 */
	public static final Color brighten(Color color) {
		if (threshold(color, 50)) {
			return darker(color, 65).brighter();
		}
		return color;
	}

	/**
	 * @param color Return a darker version of this color.
	 * @param percentage How much darker.
	 * @return A color that is darker than the color passed in by the given percentage.
	 */
	public static final Color darker(Color color, int percentage) {
		return blend(color, Color.black, percentage);
	}

	/**
	 * @param color Return a lighter version of this color.
	 * @param percentage How much lighter.
	 * @return A color that is lighter than the color passed in by the given percentage.
	 */
	public static final Color lighter(Color color, int percentage) {
		return blend(color, Color.white, percentage);
	}

	/**
	 * Extract a color from the specified string.
	 * 
	 * @param stringRepresentation The string representation of the color.
	 * @return The extracted color, or <code>null</code>.
	 */
	public static Color extractColor(String stringRepresentation) {
		if (stringRepresentation == null || NO_COLOR.equals(stringRepresentation)) {
			return null;
		}

		try {
			int[] rgba = TKConversion.extractIntegers(stringRepresentation);

			if (rgba.length != 3 && rgba.length != 4) {
				return null;
			}

			for (int i = 0; i < rgba.length; i++) {
				if (rgba[i] < 0) {
					rgba[i] = 0;
				} else if (rgba[i] > 255) {
					rgba[i] = 255;
				}
			}

			if (rgba.length == 3) {
				return new Color(rgba[0], rgba[1], rgba[2]);
			}
			return new Color(rgba[0], rgba[1], rgba[2], rgba[3]);
		} catch (NumberFormatException nfe) {
			// If any of the above fails, we merely return null
		}
		return null;
	}

	/**
	 * @param color The color to use.
	 * @return A string representation of the specified color.
	 */
	public static String getStringValue(Color color) {
		if (color != null) {
			StringBuilder buffer = new StringBuilder();
			int alpha = color.getAlpha();

			buffer.append(color.getRed());
			buffer.append(',');
			buffer.append(color.getGreen());
			buffer.append(',');
			buffer.append(color.getBlue());
			if (alpha != 255) {
				buffer.append(',');
				buffer.append(alpha);
			}
			return buffer.toString();
		}
		return NO_COLOR;
	}

	/**
	 * @param color The color being tested.
	 * @param threshold The percent of threshold (0 to 100).
	 * @return <code>true</code> if the specified color is above the threshold. For example, if
	 *         you pass in a dark color with a threshold of 50, it will return <code>false</code>
	 *         because black is 0.
	 */
	public static boolean threshold(Color color, int threshold) {
		if (threshold < 0 || threshold > 100) {
			return false;
		}
		return (int) ((color.getRed() + color.getGreen() + color.getBlue()) / 7.65) > threshold;
	}

	/**
	 * Creates an opaque RGB color.
	 * 
	 * @param rgb The combined RGB components. The alpha channel will be ignored.
	 * @see Color#Color(int)
	 */
	public TKColor(int rgb) {
		super(rgb);
	}

	/**
	 * Creates a RGB color.
	 * 
	 * @param rgb The combined RGB components.
	 * @param hasAlpha Whether or not the rgb value has an alpha channel.
	 * @see Color#Color(int,boolean)
	 */
	public TKColor(int rgb, boolean hasAlpha) {
		super(rgb, hasAlpha);
	}

	/**
	 * Creates an opaque RGB color.
	 * 
	 * @param r The red component of the color to use.
	 * @param g The green component of the color to use.
	 * @param b The blue component of the color to use.
	 * @see Color#Color(int,int,int)
	 */
	public TKColor(int r, int g, int b) {
		super(r, g, b);
	}

	/**
	 * Creates a RGB color.
	 * 
	 * @param r The red component of the color to use.
	 * @param g The green component of the color to use.
	 * @param b The blue component of the color to use.
	 * @param a The alpha component of the color to use.
	 * @see Color#Color(int,int,int)
	 */
	public TKColor(int r, int g, int b, int a) {
		super(r, g, b, a);
	}

	/**
	 * Creates a RGB color.
	 * 
	 * @param color The color to use.
	 */
	public TKColor(Color color) {
		super(color.getRGB(), true);
	}

	/**
	 * This method allows us to bypass the normal limitation of accessing the package-private field
	 * "value" in our superclass.
	 * 
	 * @param argb The new ARGB value to set.
	 */
	private void adjustColor(int argb) {
		if (argb != getRGB()) {
			try {
				Field field = Color.class.getDeclaredField("value"); //$NON-NLS-1$

				field.setAccessible(true);
				field.set(this, new Integer(argb));

				// Blow out the paint context cache, as our color is different
				field = Color.class.getDeclaredField("theContext"); //$NON-NLS-1$
				field.setAccessible(true);
				field.set(this, null);
			} catch (NoSuchFieldException nsfex) {
				assert false : TKDebug.throwableToString(nsfex);
			} catch (IllegalAccessException iaex) {
				assert false : TKDebug.throwableToString(iaex);
			}
		}
	}

	/** @param color The new color to use. */
	public void setColor(Color color) {
		adjustColor(color.getRGB());
		TKGraphics.forceRepaint();
	}

	/**
	 * @param r The new red component of the color to use.
	 * @param g The new green component of the color to use.
	 * @param b The new blue component of the color to use.
	 */
	public void setColor(int r, int g, int b) {
		adjustColor(0xFF << 24 | (r & 0xFF) << 16 | (g & 0xFF) << 8 | b & 0xFF);
		TKGraphics.forceRepaint();
	}

	/**
	 * @param r The new red component of the color to use.
	 * @param g The new green component of the color to use.
	 * @param b The new blue component of the color to use.
	 * @param a The new alpha component of the color to use.
	 */
	public void setColor(int r, int g, int b, int a) {
		adjustColor((a & 0xFF) << 24 | (r & 0xFF) << 16 | (g & 0xFF) << 8 | b & 0xFF);
		TKGraphics.forceRepaint();
	}
}
