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

package com.trollworks.gcs.character;

import static com.trollworks.gcs.character.PortraitPanel_LS.*;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.gcs.widgets.GCSWindow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.border.BoxedDropShadowBorder;
import com.trollworks.ttk.image.Images;
import com.trollworks.ttk.notification.NotifierTarget;
import com.trollworks.ttk.utility.GraphicsUtilities;
import com.trollworks.ttk.utility.Path;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.StdFileDialog;
import com.trollworks.ttk.widgets.WindowUtils;

import java.awt.Container;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Image;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.awt.geom.AffineTransform;
import java.io.File;
import java.text.MessageFormat;

import javax.swing.UIManager;

@Localized({
				@LS(key = "SELECT_PORTRAIT", msg = "Select A Portrait"),
				@LS(key = "PORTRAIT", msg = "Portrait"),
				@LS(key = "PORTRAIT_TOOLTIP", msg = "<html><body><b>Double-click</b> to set a character portrait.<br><br>The dimensions of the chosen picture should be in a ratio of<br><b>3 pixels wide for every 4 pixels tall</b> to scale without distortion.<br><br>Dimensions of <b>{0}x{1}</b> are ideal.</body></html>"),
				@LS(key = "BAD_IMAGE", msg = "Unable to load\n{0}."),
})
/** The character portrait. */
public class PortraitPanel extends DropPanel implements NotifierTarget {
	private GURPSCharacter	mCharacter;

	/**
	 * Creates a new character portrait.
	 * 
	 * @param character The owning character.
	 */
	public PortraitPanel(GURPSCharacter character) {
		super(null, true);
		setBorder(new BoxedDropShadowBorder(UIManager.getFont(GCSFonts.KEY_LABEL), PORTRAIT));
		mCharacter = character;
		Insets insets = getInsets();
		UIUtilities.setOnlySize(this, new Dimension(insets.left + insets.right + Profile.PORTRAIT_WIDTH, insets.top + insets.bottom + Profile.PORTRAIT_HEIGHT));
		setToolTipText(MessageFormat.format(PORTRAIT_TOOLTIP, new Integer(Profile.PORTRAIT_WIDTH * 2), new Integer(Profile.PORTRAIT_HEIGHT * 2)));
		mCharacter.addTarget(this, Profile.ID_PORTRAIT);
		addMouseListener(new MouseAdapter() {
			@Override
			public void mouseClicked(MouseEvent event) {
				if (event.getClickCount() == 2) {
					choosePortrait();
				}
			}
		});
	}

	/** Allows the user to choose a portrait for their character. */
	public void choosePortrait() {
		File file = StdFileDialog.choose(this, true, SELECT_PORTRAIT, null, null, "png", "jpg", "gif", "jpeg"); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
		if (file != null) {
			try {
				mCharacter.getDescription().setPortrait(Images.loadImage(file));
			} catch (Exception exception) {
				WindowUtils.showError(this, MessageFormat.format(BAD_IMAGE, Path.getFullPath(file)));
			}
		}
	}

	@Override
	protected void paintComponent(Graphics gc) {
		super.paintComponent(GraphicsUtilities.prepare(gc));

		Container top = getTopLevelAncestor();
		boolean isPrinting = top instanceof GCSWindow && ((GCSWindow) top).isPrinting();
		Image portrait = mCharacter.getDescription().getPortrait(isPrinting);

		if (portrait != null) {
			Insets insets = getInsets();
			Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
			AffineTransform transform = null;

			if (isPrinting) {
				transform = ((Graphics2D) gc).getTransform();
				((Graphics2D) gc).scale(0.5, 0.5);
				bounds.x *= 2;
				bounds.y *= 2;
			}
			gc.drawImage(portrait, bounds.x, bounds.y, null);
			if (isPrinting) {
				((Graphics2D) gc).setTransform(transform);
			}
		}
	}

	@Override
	public void handleNotification(Object producer, String type, Object data) {
		repaint();
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}
}
