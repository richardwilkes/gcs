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

import com.trollworks.toolkit.widget.layout.TKColumnLayout;

import java.awt.Color;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;

/** Provides an editable color swatch. */
public class TKEditableColorSwatch extends TKPanel implements MouseListener {
	private TKColorSwatch	mSwatch;

	/**
	 * Create a new editable color swatch.
	 * 
	 * @param label The label for the color swatch.
	 * @param color The color to use.
	 */
	public TKEditableColorSwatch(String label, Color color) {
		super(new TKColumnLayout(2, 5, 5));
		TKLabel labelComp = new TKLabel(label);
		mSwatch = new TKColorSwatch(color);
		mSwatch.addMouseListener(this);
		labelComp.addMouseListener(this);
		addMouseListener(this);
		add(mSwatch);
		add(labelComp);
	}

	/** @return The current color. */
	public Color getColor() {
		return mSwatch.getColor();
	}

	/** @param color The new current color. */
	public void setColor(Color color) {
		if (color != null && !color.equals(getColor())) {
			mSwatch.setColor(color);
			notifyActionListeners();
		}
	}

	public void mouseClicked(MouseEvent event) {
		setColor(TKColorChooser.chooseColor(getBaseWindow(), getColor()));
		event.consume();
	}

	public void mouseEntered(MouseEvent event) {
		// Nothing to do...
	}

	public void mouseExited(MouseEvent event) {
		// Nothing to do...
	}

	public void mousePressed(MouseEvent event) {
		// Nothing to do...
	}

	public void mouseReleased(MouseEvent event) {
		// Nothing to do...
	}
}
