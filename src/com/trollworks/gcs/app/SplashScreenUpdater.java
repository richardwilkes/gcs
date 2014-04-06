/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.BundleInfo;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.Version;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import java.awt.SplashScreen;

import javax.swing.UIManager;

public class SplashScreenUpdater {
	@Localize("Version %s")
	private static String		VERSION_FORMAT;
	@Localize("%s %s\n%s Architecture\nJava %s")
	private static String		PLATFORM_FORMAT;
	@Localize("GURPS is a trademark of Steve Jackson Games, used by permission. All rights reserved.\nThis product includes copyrighted material from the GURPS game, which is used by permission of Steve Jackson Games.\nThe iText Library is licensed under LGPL 2.1 by Bruno Lowagie and Paulo Soares.\nThe Trove Library is licensed under LGPL 2.1 by Eric D. Friedman and Rob Eden.")
	private static String		LICENSES;
	@Localize("Unknown build date")
	private static String		UNKNOWN_BUILD_DATE;
	@Localize("Development Version")
	private static String		DEVELOPMENT;

	static {
		Localization.initialize();
	}

	private static final String	SEPARATOR	= "\n"; //$NON-NLS-1$
	private static final int	HMARGIN		= 4;

	public static final void update() {
		SplashScreen splashScreen = SplashScreen.getSplashScreen();
		if (splashScreen != null) {
			Graphics2D gc = splashScreen.createGraphics();
			drawOverlay(gc, splashScreen.getSize());
			gc.dispose();
			splashScreen.update();
		}
	}

	public static void drawOverlay(Graphics2D gc, Dimension size) {
		Object savedTextAA = gc.getRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING);
		gc.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, RenderingHints.VALUE_TEXT_ANTIALIAS_ON);
		Font baseFont = UIManager.getFont("TextField.font"); //$NON-NLS-1$
		gc.setFont(baseFont.deriveFont(9f));
		gc.setColor(Color.WHITE);
		int right = size.width - HMARGIN;
		int y = draw(gc, LICENSES, size.height - HMARGIN, right, true, true);
		BundleInfo bundleInfo = BundleInfo.getDefault();
		long version = bundleInfo.getVersion();
		int y2 = draw(gc, bundleInfo.getCopyrightBanner(), y, right, false, true);
		draw(gc, String.format(PLATFORM_FORMAT, System.getProperty("os.name"), System.getProperty("os.version"), System.getProperty("os.arch"), System.getProperty("java.version")), y, right, false, false);//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
		y2 = draw(gc, version != 0 ? Version.toBuildTimestamp(version) : UNKNOWN_BUILD_DATE, y2, right, false, true);
		gc.setFont(baseFont.deriveFont(Font.BOLD, 11f));
		draw(gc, version != 0 ? String.format(VERSION_FORMAT, Version.toString(version, false)) : DEVELOPMENT, y2, right, false, true);
		gc.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, savedTextAA);
	}

	private static int draw(Graphics2D gc, String text, int y, int right, boolean addGap, boolean onLeft) {
		String[] one = text.split(SEPARATOR);
		FontMetrics fm = gc.getFontMetrics();
		int fHeight = fm.getAscent() + fm.getDescent();
		for (int i = one.length - 1; i >= 0; i--) {
			gc.drawString(one[i], onLeft ? HMARGIN : right - fm.stringWidth(one[i]), y);
			y -= fHeight;
		}
		if (addGap) {
			y -= fHeight / 2;
		}
		return y;
	}
}
