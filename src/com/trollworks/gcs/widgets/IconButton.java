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

package com.trollworks.gcs.widgets;

import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.ToolkitIcon;

import java.awt.Dimension;

import javax.swing.Action;
import javax.swing.JButton;

/** A button with an icon. */
// RAW: Should replace all uses of this with the Toolkit's IconButton
public class IconButton extends JButton {
	/**
	 * Creates a new {@link IconButton}.
	 *
	 * @param action The {@link Action} to use.
	 */
	public IconButton(Action action) {
		super(action);
		setToolTipText(getText());
		setText(null);
		initialize();
	}

	/**
	 * Creates a new {@link IconButton}.
	 *
	 * @param icon The image to use for the icon.
	 */
	public IconButton(ToolkitIcon icon) {
		this(icon, null);
	}

	/**
	 * Creates a new {@link IconButton}.
	 *
	 * @param icon The image to use for the icon.
	 * @param tooltip The tooltip to use.
	 */
	public IconButton(ToolkitIcon icon, String tooltip) {
		super(icon);
		setToolTipText(tooltip);
		initialize();
	}

	private void initialize() {
		setOpaque(false);
		putClientProperty("JButton.buttonType", "textured"); //$NON-NLS-1$ //$NON-NLS-2$
		Dimension size = getPreferredSize();
		size.width = size.height;
		UIUtilities.setOnlySize(this, size);
	}
}
