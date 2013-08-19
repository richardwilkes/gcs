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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.ui.common.CSDropPanel;
import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.notification.TKNotifierTarget;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.border.TKBoxedDropShadowBorder;
import com.trollworks.toolkit.window.TKOptionDialog;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKFileDialog;

import java.awt.Dimension;
import java.awt.Frame;
import java.awt.Graphics2D;
import java.awt.Image;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.awt.geom.AffineTransform;
import java.text.MessageFormat;

/** The character portrait. */
public class CSPortraitPanel extends CSDropPanel implements TKNotifierTarget {
	private static final String	EXTENSIONS	= ".png .jpg .jpeg .gif";	//$NON-NLS-1$
	private CMCharacter			mCharacter;

	/**
	 * Creates a new character portrait.
	 * 
	 * @param character The owning character.
	 */
	public CSPortraitPanel(CMCharacter character) {
		super(null, true);
		setBorder(new TKBoxedDropShadowBorder(TKFont.lookup(CSFont.KEY_LABEL), Msgs.PORTRAIT));
		mCharacter = character;
		Insets insets = getInsets();
		setOnlySize(new Dimension(insets.left + insets.right + CMCharacter.PORTRAIT_WIDTH, insets.top + insets.bottom + CMCharacter.PORTRAIT_HEIGHT));
		setToolTipText(MessageFormat.format(Msgs.PORTRAIT_TOOLTIP, new Integer(CMCharacter.PORTRAIT_WIDTH * 2), new Integer(CMCharacter.PORTRAIT_HEIGHT * 2)));
		mCharacter.addTarget(this, CMCharacter.ID_PORTRAIT);
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		if (event.getID() == MouseEvent.MOUSE_CLICKED && event.getClickCount() == 2) {
			choosePortrait();
		}
	}

	/** Allows the user to choose a portrait for their character. */
	public void choosePortrait() {
		TKFileDialog dialog = new TKFileDialog((Frame) getBaseWindow(), true);
		TKFileFilter filter = new TKFileFilter(Msgs.IMAGE_FILES, EXTENSIONS);

		dialog.addFileFilter(filter);
		dialog.setActiveFileFilter(filter);
		if (dialog.doModal() == TKDialog.OK) {
			try {
				mCharacter.setPortrait(TKImage.loadImage(dialog.getSelectedItem()));
			} catch (Exception exception) {
				TKOptionDialog.error(getBaseWindow(), MessageFormat.format(Msgs.BAD_IMAGE, TKPath.getFullPath(dialog.getSelectedItem())));
			}
		}
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		boolean isPrinting = isPrinting();
		Image portrait = mCharacter.getPortrait(isPrinting);

		if (portrait != null) {
			Rectangle bounds = getLocalInsetBounds();
			AffineTransform transform = null;

			if (isPrinting) {
				transform = g2d.getTransform();
				g2d.scale(0.5, 0.5);
				bounds.x *= 2;
				bounds.y *= 2;
			}
			g2d.drawImage(portrait, bounds.x, bounds.y, null);
			if (isPrinting) {
				g2d.setTransform(transform);
			}
		}
	}

	public void handleNotification(Object producer, String type, Object data) {
		repaint();
	}
}
