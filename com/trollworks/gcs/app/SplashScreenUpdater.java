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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.app;

import com.trollworks.ttk.utility.App;
import com.trollworks.ttk.utility.LocalizedMessages;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import java.awt.SplashScreen;

import javax.swing.UIManager;

public class SplashScreenUpdater {
	private static String		MSG_VERSION_FORMAT;
	private static String		MSG_COPYRIGHT_FORMAT;
	private static String		MSG_PLATFORM_FORMAT;
	private static String		MSG_LICENSES;
	private static final String	SEPARATOR	= "\n";	//$NON-NLS-1$
	private static final int	HMARGIN		= 4;

	static {
		LocalizedMessages.initialize(SplashScreenUpdater.class);
	}

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
		int y = draw(gc, MSG_LICENSES, size.height - HMARGIN, right, true, true);
		int y2 = draw(gc, String.format(MSG_COPYRIGHT_FORMAT, App.getCopyrightYears(), App.getCopyrightOwner()), y, right, false, true);
		draw(gc, String.format(MSG_PLATFORM_FORMAT, System.getProperty("os.name"), System.getProperty("os.version"), System.getProperty("os.arch"), System.getProperty("java.version")), y, right, false, false);//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
		gc.setFont(baseFont.deriveFont(Font.BOLD, 11f));
		draw(gc, String.format(MSG_VERSION_FORMAT, App.getVersion()), y2, right, false, true);
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
