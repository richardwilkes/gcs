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

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.utility.Platform;

import java.awt.Dimension;
import java.awt.image.BufferedImage;

import javax.swing.ImageIcon;
import javax.swing.JButton;

/** A button with an icon. */
public class ToolBarIconButton extends JButton {
	/**
	 * Creates a new {@link ToolBarIconButton}.
	 * 
	 * @param image The image to use for the icon.
	 * @param tooltip The tooltip to use.
	 */
	public ToolBarIconButton(BufferedImage image, String tooltip) {
		this(new ImageIcon(image), tooltip);
	}

	/**
	 * Creates a new {@link ToolBarIconButton}.
	 * 
	 * @param image The image to use for the icon.
	 * @param tooltip The tooltip to use.
	 */
	public ToolBarIconButton(ImageIcon image, String tooltip) {
		super(image);
		setToolTipText(tooltip);
		putClientProperty("JButton.buttonType", "textured"); //$NON-NLS-1$ //$NON-NLS-2$
		if (Platform.isLinux()) {
			Dimension size = getPreferredSize();
			size.width = size.height;
			UIUtilities.setOnlySize(this, size);
		}
	}
}
