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

package com.trollworks.gcs.app;

import com.trollworks.ttk.text.TextDrawing;
import com.trollworks.ttk.text.Version;
import com.trollworks.ttk.utility.App;
import com.trollworks.ttk.utility.Fonts;
import com.trollworks.ttk.utility.GraphicsUtilities;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.font.FontRenderContext;
import java.awt.image.BufferedImage;
import java.text.MessageFormat;

import javax.swing.JPanel;
import javax.swing.SwingConstants;

/** The about box contents. */
public class AboutPanel extends JPanel {
	private static String		MSG_VERSION;
	private static String		MSG_JAVA_VERSION;
	private static String		MSG_GURPS_LICENSE;
	private static String		MSG_ITEXT_LICENSE;
	private static final int	MARGIN		= 5;
	private static final int	EXTRA_SPACE	= 110;

	static {
		LocalizedMessages.initialize(AboutPanel.class);
	}

	/** Creates a new about panel. */
	public AboutPanel() {
		super();
		setOpaque(true);
		setBackground(Color.black);
		BufferedImage img = GCSImages.getSplash();
		UIUtilities.setOnlySize(this, new Dimension(img.getWidth(), img.getHeight() + EXTRA_SPACE));
	}

	@Override
	protected void paintComponent(Graphics gc) {
		super.paintComponent(GraphicsUtilities.prepare(gc));

		// Draw the background
		BufferedImage img = GCSImages.getSplash();
		gc.drawImage(img, 0, 0, null);

		// Version
		gc.setColor(Color.white);
		Font font = new Font(Fonts.getDefaultFontName(), Font.BOLD, 10);
		gc.setFont(font);
		String version = MessageFormat.format(MSG_VERSION, Version.getHumanReadableVersion(Version.extractVersion(App.getVersion())));
		FontRenderContext frc = ((Graphics2D) gc).getFontRenderContext();
		Dimension size = TextDrawing.getPreferredSize(font, frc, version);
		Rectangle bounds = new Rectangle(MARGIN, img.getHeight() + MARGIN, size.width, size.height);
		TextDrawing.draw(gc, bounds, version, SwingConstants.LEFT, SwingConstants.TOP);

		// Copyright
		String copyright = Main.getCopyrightBanner(true);
		font = new Font(Fonts.getDefaultFontName(), Font.PLAIN, 10);
		gc.setFont(font);
		size = TextDrawing.getPreferredSize(font, frc, copyright);
		bounds.width = size.width;
		bounds.y += bounds.height;
		bounds.height = size.height;
		TextDrawing.draw(gc, bounds, copyright, SwingConstants.LEFT, SwingConstants.TOP);
		String javaVersion = MessageFormat.format(MSG_JAVA_VERSION, System.getProperty("java.version")); //$NON-NLS-1$
		size = TextDrawing.getPreferredSize(font, frc, javaVersion);
		bounds.width = size.width;
		int compWidth = getWidth();
		bounds.x = compWidth - (size.width + MARGIN);
		TextDrawing.draw(gc, bounds, javaVersion, SwingConstants.RIGHT, SwingConstants.TOP);
		bounds.x = MARGIN;

		// iText License
		font = new Font(Fonts.getDefaultFontName(), Font.PLAIN, 9);
		String license = TextDrawing.wrapToPixelWidth(font, frc, MSG_ITEXT_LICENSE, compWidth - MARGIN * 2);
		gc.setFont(font);
		size = TextDrawing.getPreferredSize(font, frc, license);
		bounds.width = size.width;
		bounds.y = getHeight() - (size.height + MARGIN);
		bounds.height = size.height;
		TextDrawing.draw(gc, bounds, license, SwingConstants.LEFT, SwingConstants.TOP);

		// SJG License
		license = TextDrawing.wrapToPixelWidth(font, frc, MSG_GURPS_LICENSE, compWidth - MARGIN * 2);
		size = TextDrawing.getPreferredSize(font, frc, license);
		bounds.width = size.width;
		bounds.y -= size.height + MARGIN;
		bounds.height = size.height;
		TextDrawing.draw(gc, bounds, license, SwingConstants.LEFT, SwingConstants.TOP);
	}
}
