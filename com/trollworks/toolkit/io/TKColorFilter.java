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

import java.awt.Color;
import java.awt.Toolkit;
import java.awt.image.BufferedImage;
import java.awt.image.FilteredImageSource;
import java.awt.image.ImageProducer;
import java.awt.image.RGBImageFilter;

/** Colorizes an image by blending any solid pixels with the specified color. */
public class TKColorFilter extends RGBImageFilter {
	private static final int	OPAQUE	= 0xFF000000;
	private int					mRed;
	private int					mGreen;
	private int					mBlue;

	/**
	 * Constructs an image filter that colorizes an image by blending any solid pixels with the
	 * specified color.
	 * 
	 * @param color The color to blend with.
	 */
	public TKColorFilter(Color color) {
		mRed = color.getRed();
		mGreen = color.getGreen();
		mBlue = color.getBlue();
		canFilterIndexColorModel = true;
	}

	/**
	 * Creates a colorized image.
	 * 
	 * @param image The image to colorize.
	 * @param color The color to apply.
	 * @return The colorized image.
	 */
	public static BufferedImage createColorizedImage(BufferedImage image, Color color) {
		TKColorFilter filter = new TKColorFilter(color);
		ImageProducer producer = new FilteredImageSource(image.getSource(), filter);

		return TKImage.getBufferedImage(Toolkit.getDefaultToolkit().createImage(producer));
	}

	/**
	 * {@inheritDoc} We leave pixels with an alpha channel alone.
	 */
	@Override public int filterRGB(int x, int y, int argb) {
		if ((argb & OPAQUE) == OPAQUE) {
			int darkenBy = (int) (((argb >> 16 & 255) * 0.3 + (argb >> 8 & 255) * 0.59 + (argb & 255) * 0.11) / 2.55);

			argb = OPAQUE | mRed * darkenBy / 100 << 16 | mGreen * darkenBy / 100 << 8 | mBlue * darkenBy / 100;
		}
		return argb;
	}
}
