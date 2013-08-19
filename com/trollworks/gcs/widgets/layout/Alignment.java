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

package com.trollworks.gcs.widgets.layout;

import java.awt.Rectangle;

/** The possible alignments. */
public enum Alignment {
	/** The left/top alignment. */
	LEFT_TOP,
	/** The center alignment. */
	CENTER,
	/** The right/bottom alignment. */
	RIGHT_BOTTOM;

	/**
	 * Positions the inner rectangle within the outer one using the specified alignments.
	 * 
	 * @param outer The outer rectangle.
	 * @param inner The inner rectangle.
	 * @param horizontalAlignment The horizontal alignment.
	 * @param verticalAlignment The vertical alignment.
	 * @return The inner rectangle, which has been adjusted.
	 */
	public static Rectangle position(Rectangle outer, Rectangle inner, Alignment horizontalAlignment, Alignment verticalAlignment) {
		switch (horizontalAlignment) {
			case LEFT_TOP:
				inner.x = outer.x;
				break;
			case CENTER:
				inner.x = outer.x + (outer.width - inner.width) / 2;
				break;
			case RIGHT_BOTTOM:
				inner.x = outer.x + outer.width - inner.width;
				break;
		}
		switch (verticalAlignment) {
			case LEFT_TOP:
				inner.y = outer.y;
				break;
			case CENTER:
				inner.y = outer.y + (outer.height - inner.height) / 2;
				break;
			case RIGHT_BOTTOM:
				inner.y = outer.y + outer.height - inner.height;
				break;
		}
		return inner;
	}
}
