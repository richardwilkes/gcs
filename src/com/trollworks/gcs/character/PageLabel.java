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

package com.trollworks.gcs.character;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.toolkit.ui.widget.LinkedLabel;

import java.awt.Color;

import javax.swing.JComponent;

/** A label for a field in a page. */
public class PageLabel extends LinkedLabel {
	/**
	 * Creates a new label for the specified field.
	 * 
	 * @param title The title of the field.
	 * @param field The field.
	 */
	public PageLabel(String title, JComponent field) {
		super(title, GCSFonts.KEY_LABEL, field);
		setForeground(Color.BLACK);
	}

	/**
	 * Creates a new label for the specified field.
	 * 
	 * @param title The title of the field.
	 * @param field The field.
	 * @param alignment The horizontal alignment to use.
	 */
	public PageLabel(String title, JComponent field, int alignment) {
		super(title, GCSFonts.KEY_LABEL, field, alignment);
		setForeground(Color.BLACK);
	}
}
