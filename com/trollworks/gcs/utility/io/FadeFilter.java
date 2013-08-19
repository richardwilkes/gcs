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

import java.awt.Toolkit;
import java.awt.image.BufferedImage;
import java.awt.image.FilteredImageSource;
import java.awt.image.ImageProducer;
import java.awt.image.RGBImageFilter;

/** Fades an image by blending the pixels with white or black. */
public class FadeFilter extends RGBImageFilter {
	private static final int	OPAQUE	= 0xFF000000;
	private int					mPercentage;
	private boolean				mUseWhite;

	/**
	 * Constructs an image filter that fades an image by blending the pixels with the specified
	 * percentage of white or black.
	 * 
	 * @param percentage The percentage of white or black to use.
	 * @param useWhite Whether to use white or black.
	 */
	public FadeFilter(int percentage, boolean useWhite) {
		mPercentage = percentage;
		mUseWhite = useWhite;
		canFilterIndexColorModel = true;
	}

	/**
	 * Creates a faded image.
	 * 
	 * @param image The image to fade.
	 * @param percentage The percentage of white or black to use.
	 * @param useWhite Whether to use white or black.
	 * @return The faded image.
	 */
	public static BufferedImage createFadedImage(BufferedImage image, int percentage, boolean useWhite) {
		FadeFilter filter = new FadeFilter(percentage, useWhite);
		ImageProducer producer = new FilteredImageSource(image.getSource(), filter);

		return Images.getBufferedImage(Toolkit.getDefaultToolkit().createImage(producer));
	}

	@Override public int filterRGB(int x, int y, int argb) {
		int red = argb >> 16 & 0xFF;
		int green = argb >> 8 & 0xFF;
		int blue = argb & 0xFF;
		int p1 = 100 - mPercentage;
		int p2 = mUseWhite ? 255 * mPercentage : 0;

		red = (red * p1 + p2) / 100;
		green = (green * p1 + p2) / 100;
		blue = (blue * p1 + p2) / 100;
		return argb & OPAQUE | red << 16 | green << 8 | blue;
	}
}
