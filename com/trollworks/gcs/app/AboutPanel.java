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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.app;

import com.trollworks.ttk.utility.GraphicsUtilities;
import com.trollworks.ttk.utility.UIUtilities;

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
