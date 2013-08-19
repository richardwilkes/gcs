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

package com.trollworks.gcs.widgets;

import java.awt.Dimension;
import java.awt.image.BufferedImage;

import javax.swing.ImageIcon;
import javax.swing.JButton;

/** A button with an icon. */
public class IconButton extends JButton {
	/**
	 * Creates a new {@link IconButton}.
	 * 
	 * @param image The image to use for the icon.
	 */
	public IconButton(BufferedImage image) {
		this(new ImageIcon(image));
	}

	/**
	 * Creates a new {@link IconButton}.
	 * 
	 * @param image The image to use for the icon.
	 * @param tooltip The tooltip to use.
	 */
	public IconButton(BufferedImage image, String tooltip) {
		this(new ImageIcon(image), tooltip);
	}

	/**
	 * Creates a new {@link IconButton}.
	 * 
	 * @param icon The icon to use.
	 */
	public IconButton(ImageIcon icon) {
		this(icon, null);
	}

	/**
	 * Creates a new {@link IconButton}.
	 * 
	 * @param icon The icon to use.
	 * @param tooltip The tooltip to use.
	 */
	public IconButton(ImageIcon icon, String tooltip) {
		super(icon);
		setOpaque(false);
		setToolTipText(tooltip);

		// This is done since Linux & Windows both do screwy things with the width of icon-only
		// buttons.
		Dimension size = getPreferredSize();
		size.width = size.height;
		UIUtilities.setOnlySize(this, size);
	}
}
