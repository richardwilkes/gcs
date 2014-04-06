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

import com.trollworks.toolkit.ui.GraphicsUtilities;
import com.trollworks.toolkit.ui.UIUtilities;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.image.BufferedImage;

import javax.swing.JPanel;

/** The about box contents. */
public class AboutPanel extends JPanel {
	/** Creates a new about panel. */
	public AboutPanel() {
		super();
		setOpaque(true);
		setBackground(Color.black);
		BufferedImage img = GCSImages.getSplash();
		UIUtilities.setOnlySize(this, new Dimension(img.getWidth(), img.getHeight()));
	}

	@Override
	protected void paintComponent(Graphics gc) {
		super.paintComponent(GraphicsUtilities.prepare(gc));
		gc.drawImage(GCSImages.getSplash(), 0, 0, null);
		SplashScreenUpdater.drawOverlay((Graphics2D) gc, getSize());
	}
}
