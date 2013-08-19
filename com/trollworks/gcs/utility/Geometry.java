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

package com.trollworks.gcs.utility;

import java.awt.Rectangle;

/** Utility methods for dealing with geometry. */
public class Geometry {
	/**
	 * Intersects two rectangles, producing a third. Unlike the
	 * {@link Rectangle#intersection(Rectangle)} method, the resulting rectangle's width & height
	 * will not be set to less than zero when there is no overlap.
	 * 
	 * @param first The first rectangle.
	 * @param second The second rectangle.
	 * @return The intersection of the two rectangles.
	 */
	public static Rectangle intersection(Rectangle first, Rectangle second) {
		if (first.width < 1 || first.height < 1 || second.width < 1 || second.height < 1) {
			return new Rectangle();
		}

		int x = Math.max(first.x, second.x);
		int y = Math.max(first.y, second.y);
		int w = Math.min(first.x + first.width, second.x + second.width) - x;
		int h = Math.min(first.y + first.height, second.y + second.height) - y;

		if (w < 0 || h < 0) {
			return new Rectangle();
		}

		return new Rectangle(x, y, w, h);
	}

	/**
	 * Unions two rectangles, producing a third. Unlike the {@link Rectangle#union(Rectangle)}
	 * method, an empty rectangle will not cause the rectangle's boundary to extend to the 0,0
	 * point.
	 * 
	 * @param first The first rectangle.
	 * @param second The second rectangle.
	 * @return The resulting rectangle.
	 */
	public static Rectangle union(Rectangle first, Rectangle second) {
		boolean firstEmpty = first.width < 1 || first.height < 1;
		boolean secondEmpty = second.width < 1 || second.height < 1;

		if (firstEmpty && secondEmpty) {
			return new Rectangle();
		}
		if (firstEmpty) {
			return new Rectangle(second);
		}
		if (secondEmpty) {
			return new Rectangle(first);
		}
		return first.union(second);
	}
}
