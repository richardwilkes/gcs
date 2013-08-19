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

package com.trollworks.gcs.ui.preferences;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKBoxedDropShadowBorder;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;

/** The character portrait. */
public class CSPortraitPreferencePanel extends TKPanel {
	private BufferedImage	mImage;

	/**
	 * Creates a new character portrait.
	 * 
	 * @param image The image to display.
	 */
	public CSPortraitPreferencePanel(BufferedImage image) {
		super();
		mImage = image;
		setBorder(new TKBoxedDropShadowBorder(TKFont.lookup(CSFont.KEY_LABEL), Msgs.PORTRAIT));
		Insets insets = getInsets();
		setOnlySize(new Dimension(insets.left + insets.right + CMCharacter.PORTRAIT_WIDTH, insets.top + insets.bottom + CMCharacter.PORTRAIT_HEIGHT));
		setToolTipText(Msgs.PORTRAIT_TOOLTIP);
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		if (event.getID() == MouseEvent.MOUSE_CLICKED && event.getClickCount() == 2) {
			notifyActionListeners();
		}
	}

	/** @param image The new portrait. */
	public void setPortrait(BufferedImage image) {
		mImage = image;
		repaint();
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Rectangle bounds = getLocalInsetBounds();

		g2d.setColor(Color.white);
		g2d.fill(bounds);
		if (mImage != null) {
			g2d.drawImage(mImage, bounds.x, bounds.y, null);
		}
	}
}
