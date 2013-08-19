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

package com.trollworks.toolkit.widget.border;

import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.Insets;

/** A beveled border which can be raised or lowered. */
public class TKBevelBorder implements TKBorder {
	private static TKBevelBorder	RAISED	= null;
	private static TKBevelBorder	LOWERED	= null;
	private Color					mOuterHighlight;
	private Color					mInnerHighlight;
	private Color					mOuterShadow;
	private Color					mInnerShadow;
	private boolean					mRaised;

	/**
	 * Create a new beveled border.
	 * 
	 * @param raised Pass in <code>true</code> to get a raised border, <code>false</code> to get
	 *            a lowered one.
	 * @param highlight The highlight color to use.
	 * @param shadow The shadow color to use.
	 */
	public TKBevelBorder(boolean raised, Color highlight, Color shadow) {
		this(raised, TKColor.lighter(highlight, 30), highlight, shadow, TKColor.lighter(shadow, 30));
	}

	/**
	 * Create a new beveled border.
	 * 
	 * @param raised Pass in <code>true</code> to get a raised border, <code>false</code> to get
	 *            a lowered one.
	 * @param outerHighlight The outer highlight color to use.
	 * @param innerHighlight The inner highlight color to use.
	 * @param outerShadow The outer shadow color to use.
	 * @param innerShadow The inner shadow color to use.
	 */
	public TKBevelBorder(boolean raised, Color outerHighlight, Color innerHighlight, Color outerShadow, Color innerShadow) {
		mRaised = raised;
		mOuterHighlight = outerHighlight;
		mInnerHighlight = innerHighlight;
		mOuterShadow = outerShadow;
		mInnerShadow = innerShadow;
	}

	/**
	 * @param raised Pass in <code>true</code> to get a raised border, <code>false</code> to get
	 *            a lowered one.
	 * @return A default shared border, which can be either raised or lowered.
	 */
	public static TKBevelBorder getSharedBorder(boolean raised) {
		if (raised) {
			if (RAISED == null) {
				RAISED = new TKBevelBorder(true, null, null, null, null);
			}
			return RAISED;
		}

		if (LOWERED == null) {
			LOWERED = new TKBevelBorder(false, null, null, null, null);
		}
		return LOWERED;
	}

	public void paintBorder(TKPanel panel, Graphics2D g2d, int x, int y, int width, int height) {
		Color savedColor = g2d.getColor();
		Color outerHighlight = mOuterHighlight != null ? mOuterHighlight : TKColor.lighter(panel.getBackground(), 50);
		Color innerHighlight = mInnerHighlight != null ? mInnerHighlight : TKColor.lighter(panel.getBackground(), 30);
		Color outerShadow = mOuterShadow != null ? mOuterShadow : TKColor.darker(panel.getBackground(), 50);
		Color innerShadow = mInnerShadow != null ? mInnerShadow : TKColor.darker(panel.getBackground(), 30);

		g2d.translate(x, y);

		g2d.setColor(mRaised ? outerHighlight : innerShadow);
		g2d.drawLine(0, 0, 0, height - 1);
		g2d.drawLine(1, 0, width - 1, 0);

		g2d.setColor(mRaised ? innerHighlight : outerShadow);
		g2d.drawLine(1, 1, 1, height - 2);
		g2d.drawLine(2, 1, width - 2, 1);

		g2d.setColor(mRaised ? outerShadow : outerHighlight);
		g2d.drawLine(1, height - 1, width - 1, height - 1);
		g2d.drawLine(width - 1, 1, width - 1, height - 2);

		g2d.setColor(mRaised ? innerShadow : innerHighlight);
		g2d.drawLine(2, height - 2, width - 2, height - 2);
		g2d.drawLine(width - 2, 2, width - 2, height - 3);

		g2d.translate(-x, -y);
		g2d.setColor(savedColor);
	}

	public Insets getBorderInsets(TKPanel panel) {
		return new Insets(2, 2, 2, 2);
	}
}
