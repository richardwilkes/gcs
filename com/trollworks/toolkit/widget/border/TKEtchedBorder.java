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

/** A border which can be etched in or out. */
public class TKEtchedBorder implements TKBorder {
	private static TKEtchedBorder	OUT	= null;
	private static TKEtchedBorder	IN	= null;
	private boolean					mEtchedIn;

	private TKEtchedBorder(boolean etchedIn) {
		mEtchedIn = etchedIn;
	}

	/**
	 * @param etchedIn Pass in <code>true</code> to have the border be etched in,
	 *            <code>false</code> to have the border be etched out.
	 * @return A shared etched border, which can either be etched in or out.
	 */
	public static TKEtchedBorder getSharedBorder(boolean etchedIn) {
		if (etchedIn) {
			if (IN == null) {
				IN = new TKEtchedBorder(true);
			}
			return IN;
		}

		if (OUT == null) {
			OUT = new TKEtchedBorder(false);
		}
		return OUT;
	}

	public void paintBorder(TKPanel panel, Graphics2D g2d, int x, int y, int width, int height) {
		Color oldColor = g2d.getColor();
		Color bgColor = panel.getBackground();
		Color highlight = TKColor.lighter(bgColor, 45);
		Color shadow = TKColor.darker(bgColor, 45);

		height--;
		width--;

		g2d.setColor(mEtchedIn ? shadow : highlight);
		g2d.drawLine(x + 1, y, x + width - 2, y);
		g2d.drawLine(x, y + 1, x, y + height - 2);
		g2d.drawLine(x + width - 1, y + 1, x + width - 1, y + height - 2);
		g2d.drawLine(x + 1, y + height - 1, x + width - 2, y + height - 1);

		g2d.setColor(mEtchedIn ? highlight : shadow);
		g2d.drawLine(x + 1, y + 1, x + width - 2, y + 1);
		g2d.drawLine(x + 1, y + 1, x + 1, y + height - 2);
		g2d.drawLine(x + width, y + 1, x + width, y + height - 2);
		g2d.drawLine(x + 1, y + height, x + width - 2, y + height);

		g2d.setColor(oldColor);
	}

	public Insets getBorderInsets(TKPanel panel) {
		return new Insets(2, 2, 2, 2);
	}
}
