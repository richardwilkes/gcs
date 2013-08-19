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

package com.trollworks.toolkit.widget;

import com.trollworks.toolkit.io.TKImage;

import java.awt.Graphics2D;
import java.awt.image.BufferedImage;

/** A label for use with {@link TKCorrectableField}s. */
public class TKCorrectableLabel extends TKLinkedLabel {
	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 */
	public TKCorrectableLabel(String text) {
		super(text);
		setImage(TKImage.getMiniWarningIcon());
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 */
	public TKCorrectableLabel(String text, String font) {
		super(text, font);
		setImage(TKImage.getMiniWarningIcon());
	}

	@Override protected void drawImage(Graphics2D g2d, BufferedImage image, int x, int y) {
		TKPanel link = getLink();

		if (link instanceof TKCorrectable && ((TKCorrectable) link).wasCorrected()) {
			super.drawImage(g2d, image, x, y);
		}
	}

	@Override public void setLink(TKPanel link) {
		super.setLink(link);
		repaint();
	}
}
