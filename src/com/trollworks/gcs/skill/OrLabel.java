/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.skill;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.GraphicsUtilities;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Component;
import java.awt.Graphics;

import javax.swing.JLabel;
import javax.swing.SwingConstants;

/** A label that displays the "or" message, or nothing if it is the first one. */
public class OrLabel extends JLabel {
	@Localize("or")
	@Localize(locale = "de", value = "oder")
	@Localize(locale = "ru", value = "или")
	private static String	OR;

	static {
		Localization.initialize();
	}

	private Component		mOwner;

	/**
	 * Creates a new {@link OrLabel}.
	 *
	 * @param owner The owning component.
	 */
	public OrLabel(Component owner) {
		super(OR, SwingConstants.RIGHT);
		mOwner = owner;
		UIUtilities.setOnlySize(this, getPreferredSize());
	}

	@Override
	protected void paintComponent(Graphics gc) {
		setText(UIUtilities.getIndexOf(mOwner.getParent(), mOwner) != 0 ? OR : ""); //$NON-NLS-1$
		super.paintComponent(GraphicsUtilities.prepare(gc));
	}
}
