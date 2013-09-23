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

package com.trollworks.gcs.preferences;

import static com.trollworks.gcs.preferences.PortraitPreferencePanel_LS.*;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.gcs.character.Profile;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.border.BoxedDropShadowBorder;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.ActionPanel;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;
import java.text.MessageFormat;

import javax.swing.UIManager;

@Localized({
				@LS(key = "PORTRAIT", msg = "Portrait"),
				@LS(key = "PORTRAIT_TOOLTIP", msg = "<html><body>The portrait to use when a new character sheet is created.<br><br>Ideal original portrait size is {0} pixels wide by {1} pixels tall,<br>although the image will be automatically scaled to these<br>dimensions, if necessary.</body></html>"),
})
/** The character portrait. */
public class PortraitPreferencePanel extends ActionPanel {
	private BufferedImage	mImage;

	/**
	 * Creates a new character portrait.
	 * 
	 * @param image The image to display.
	 */
	public PortraitPreferencePanel(BufferedImage image) {
		super();
		mImage = image;
		setBorder(new BoxedDropShadowBorder(UIManager.getFont(GCSFonts.KEY_LABEL), PORTRAIT));
		Insets insets = getInsets();
		UIUtilities.setOnlySize(this, new Dimension(insets.left + insets.right + Profile.PORTRAIT_WIDTH, insets.top + insets.bottom + Profile.PORTRAIT_HEIGHT));
		setToolTipText(MessageFormat.format(PORTRAIT_TOOLTIP, new Integer(Profile.PORTRAIT_WIDTH * 2), new Integer(Profile.PORTRAIT_HEIGHT * 2)));
		addMouseListener(new MouseAdapter() {
			@Override
			public void mouseClicked(MouseEvent event) {
				if (event.getClickCount() == 2) {
					notifyActionListeners();
				}
			}
		});
	}

	/** @param image The new portrait. */
	public void setPortrait(BufferedImage image) {
		mImage = image;
		repaint();
	}

	@Override
	protected void paintComponent(Graphics gc) {
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		gc.setColor(Color.white);
		gc.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
		if (mImage != null) {
			gc.drawImage(mImage, bounds.x, bounds.y, null);
		}
	}
}
