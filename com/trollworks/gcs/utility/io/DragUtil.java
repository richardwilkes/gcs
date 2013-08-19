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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.utility.io;

import java.awt.dnd.DragGestureEvent;
import java.awt.dnd.DragGestureListener;

import sun.awt.dnd.SunDragSourceContextPeer;

/** Helper methods for drag & drop. */
public class DragUtil {
	/**
	 * This method should be called prior to doing <b>anything</b> else within
	 * {@link DragGestureListener#dragGestureRecognized(DragGestureEvent)}. It works around a bug
	 * on the Mac as of December 18, 2005 which sometimes allowed that method to be called prior to
	 * the drag &amp; drop state being reset.
	 */
	public static final void prepDrag() {
		try {
			SunDragSourceContextPeer.setDragDropInProgress(false);
		} catch (Exception exception) {
			// Don't care... this code is here to force the drag & drop state to
			// false if it wasn't already, which seems to happen on the Mac under
			// some conditions.
		}
	}
}
