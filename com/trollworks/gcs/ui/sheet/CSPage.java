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

import com.trollworks.toolkit.print.TKPrintManager;
import com.trollworks.toolkit.utility.units.TKLengthUnits;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;

/** A printer page. */
public class CSPage extends TKPanel {
	private CSPageOwner	mOwner;

	/**
	 * Creates a new page.
	 * 
	 * @param owner The page owner.
	 */
	public CSPage(CSPageOwner owner) {
		super(new TKCompassLayout());
		mOwner = owner;
		updateBorder();
		setBackground(Color.white);
		setOpaque(true);
	}

	private void updateBorder() {
		TKPrintManager pageSettings = mOwner.getPageSettings();
		Insets insets = mOwner.getPageAdornmentsInsets(this);
		double[] margins = pageSettings.getPageMargins(TKLengthUnits.POINTS);

		setBorder(new TKEmptyBorder(insets.top + (int) margins[0], insets.left + (int) margins[1], insets.bottom + (int) margins[2], insets.right + (int) margins[3]));
	}

	@Override protected Dimension getPreferredSizeSelf() {
		double[] size = mOwner.getPageSettings().getPageSize(TKLengthUnits.POINTS);

		updateBorder();
		return new Dimension((int) size[0], (int) size[1]);
	}

	@Override protected Dimension getMinimumSizeSelf() {
		return getPreferredSizeSelf();
	}

	@Override protected Dimension getMaximumSizeSelf() {
		return getPreferredSizeSelf();
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		mOwner.drawPageAdornments(this, g2d);
	}
}
