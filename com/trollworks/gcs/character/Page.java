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

package com.trollworks.gcs.character;

import com.trollworks.gcs.utility.io.print.PrintManager;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.widgets.GraphicsUtilities;
import com.trollworks.gcs.widgets.UIUtilities;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Insets;

import javax.swing.JPanel;
import javax.swing.border.EmptyBorder;

/** A printer page. */
public class Page extends JPanel {
	private PageOwner	mOwner;

	/**
	 * Creates a new page.
	 * 
	 * @param owner The page owner.
	 */
	public Page(PageOwner owner) {
		super(new BorderLayout());
		mOwner = owner;
		setOpaque(true);
		setBackground(Color.white);
		PrintManager pageSettings = mOwner.getPageSettings();
		Insets insets = mOwner.getPageAdornmentsInsets(this);
		double[] size = pageSettings != null ? pageSettings.getPageSize(LengthUnits.POINTS) : new double[] { 8.5 * 72.0, 11.0 * 72.0 };
		double[] margins = pageSettings != null ? pageSettings.getPageMargins(LengthUnits.POINTS) : new double[] { 36.0, 36.0, 36.0, 36.0 };
		setBorder(new EmptyBorder(insets.top + (int) margins[0], insets.left + (int) margins[1], insets.bottom + (int) margins[2], insets.right + (int) margins[3]));
		Dimension pageSize = new Dimension((int) size[0], (int) size[1]);
		UIUtilities.setOnlySize(this, pageSize);
		setSize(pageSize);
	}

	@Override protected void paintComponent(Graphics gc) {
		super.paintComponent(GraphicsUtilities.prepare(gc));
		mOwner.drawPageAdornments(this, gc);
	}
}
