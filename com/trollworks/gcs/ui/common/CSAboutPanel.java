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

package com.trollworks.gcs.ui.common;

import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKApp;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKVersion;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.window.TKAboutWindow;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.RenderingHints;
import java.awt.font.FontRenderContext;
import java.awt.image.BufferedImage;
import java.text.MessageFormat;

/** The about box contents. */
public class CSAboutPanel extends TKPanel {
	private static final int	MARGIN		= 5;
	private static final int	EXTRA_SPACE	= 110;
	private BufferedImage		mImage;
	private String				mVersion;

	/** @return The about window. */
	public static final TKAboutWindow createAboutWindow() {
		return new TKAboutWindow(MessageFormat.format(Msgs.TITLE, TKApp.getName()), null, new CSAboutPanel());
	}

	/** Creates a new about panel. */
	public CSAboutPanel() {
		super();
		setOpaque(true);
		setBackground(Color.black);
		mImage = CSImage.getSplash();
		mVersion = MessageFormat.format(Msgs.VERSION, TKVersion.getHumanReadableVersion(TKVersion.extractVersion(TKApp.getVersion()), true));
		setOnlySize(new Dimension(mImage.getWidth(), mImage.getHeight() + EXTRA_SPACE));
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		RenderingHints savedHints = g2d.getRenderingHints();
		Font versionFont = new Font(TKFont.getDefaultFontName(), Font.BOLD, 10);
		Font copyrightFont = new Font(TKFont.getDefaultFontName(), Font.PLAIN, 10);
		Font titleFont = new Font(TKFont.getDefaultFontName(), Font.BOLD, 9);
		Font testerFont = new Font(TKFont.getDefaultFontName(), Font.PLAIN, 9);
		Font licenseFont = new Font(TKFont.getDefaultFontName(), Font.PLAIN, 8);
		int top = mImage.getHeight() + MARGIN;
		FontRenderContext frc;
		Dimension size;
		Dimension size2;
		Rectangle bounds;

		g2d.setRenderingHint(RenderingHints.KEY_ALPHA_INTERPOLATION, RenderingHints.VALUE_ALPHA_INTERPOLATION_QUALITY);
		g2d.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_ON);
		g2d.setRenderingHint(RenderingHints.KEY_COLOR_RENDERING, RenderingHints.VALUE_COLOR_RENDER_QUALITY);
		g2d.setRenderingHint(RenderingHints.KEY_DITHERING, RenderingHints.VALUE_DITHER_DISABLE);
		g2d.setRenderingHint(RenderingHints.KEY_FRACTIONALMETRICS, RenderingHints.VALUE_FRACTIONALMETRICS_OFF);
		g2d.setRenderingHint(RenderingHints.KEY_INTERPOLATION, RenderingHints.VALUE_INTERPOLATION_BICUBIC);
		g2d.setRenderingHint(RenderingHints.KEY_RENDERING, RenderingHints.VALUE_RENDER_QUALITY);
		g2d.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, RenderingHints.VALUE_TEXT_ANTIALIAS_ON);
		frc = g2d.getFontRenderContext();

		g2d.drawImage(mImage, 0, 0, null);

		// Version
		g2d.setColor(Color.white);
		g2d.setFont(versionFont);
		size = TKTextDrawing.getPreferredSize(versionFont, frc, mVersion);
		bounds = new Rectangle(MARGIN, top, size.width, size.height);
		TKTextDrawing.draw(g2d, bounds, mVersion, TKAlignment.LEFT, TKAlignment.TOP);

		// Copyright
		String copyright = TKApp.getCopyrightBanner(true);
		g2d.setFont(copyrightFont);
		size = TKTextDrawing.getPreferredSize(copyrightFont, frc, copyright);
		bounds.width = size.width;
		bounds.y += bounds.height;
		bounds.height = size.height;
		TKTextDrawing.draw(g2d, bounds, copyright, TKAlignment.LEFT, TKAlignment.TOP);

		// iText License
		g2d.setFont(licenseFont);
		size = TKTextDrawing.getPreferredSize(licenseFont, frc, Msgs.ITEXT_LICENSE);
		bounds.width = size.width;
		bounds.y = getHeight() - (size.height + MARGIN);
		bounds.height = size.height;
		TKTextDrawing.draw(g2d, bounds, Msgs.ITEXT_LICENSE, TKAlignment.LEFT, TKAlignment.TOP);

		// SJG License
		g2d.setFont(licenseFont);
		size = TKTextDrawing.getPreferredSize(licenseFont, frc, Msgs.GURPS_LICENSE);
		bounds.width = size.width;
		bounds.y -= size.height + MARGIN;
		bounds.height = size.height;
		TKTextDrawing.draw(g2d, bounds, Msgs.GURPS_LICENSE, TKAlignment.LEFT, TKAlignment.TOP);

		// Testers
		size = TKTextDrawing.getPreferredSize(titleFont, frc, Msgs.THANKS);
		size2 = TKTextDrawing.getPreferredSize(testerFont, frc, Msgs.TESTERS);
		bounds.x += bounds.width + MARGIN;
		bounds.width = getWidth() - (bounds.x + MARGIN);
		bounds.y = mImage.getHeight() + (EXTRA_SPACE - (size.height + MARGIN + size2.height)) / 2;
		bounds.height = size.height;
		g2d.setFont(titleFont);
		TKTextDrawing.draw(g2d, bounds, Msgs.THANKS, TKAlignment.CENTER, TKAlignment.TOP);
		g2d.setFont(testerFont);
		bounds.y += size.height + MARGIN;
		bounds.height = size2.height;
		TKTextDrawing.draw(g2d, bounds, Msgs.TESTERS, TKAlignment.CENTER, TKAlignment.TOP);

		g2d.setRenderingHints(savedHints);
	}
}
