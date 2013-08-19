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

package com.trollworks.toolkit.widget.scroll;

import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Point;
import java.awt.event.MouseEvent;

/** A {@link MouseEvent} generated for auto-scrolling. */
public class TKAutoScrollEvent extends MouseEvent {
	/**
	 * Creates a new auto-scroll event.
	 * 
	 * @param comp The component the event is for.
	 * @param modifiers The modifiers for the event.
	 * @param x The x-coordinate of the mouse.
	 * @param y The y-coordinate of the mouse.
	 */
	public TKAutoScrollEvent(TKPanel comp, int modifiers, int x, int y) {
		this(comp, System.currentTimeMillis(), modifiers, x, y);
	}

	/**
	 * Creates a new auto-scroll event.
	 * 
	 * @param event The event to copy.
	 * @param comp The component the event is for.
	 * @param pt The mouse position.
	 */
	public TKAutoScrollEvent(MouseEvent event, TKPanel comp, Point pt) {
		this(comp, event.getWhen(), event.getModifiers(), pt.x, pt.y);
	}

	/**
	 * Creates a new auto-scroll event.
	 * 
	 * @param comp The component the event is for.
	 * @param when When this event occured.
	 * @param modifiers The modifiers for the event.
	 * @param x The x-coordinate of the mouse.
	 * @param y The y-coordinate of the mouse.
	 */
	public TKAutoScrollEvent(TKPanel comp, long when, int modifiers, int x, int y) {
		super(comp, MouseEvent.MOUSE_DRAGGED, when, modifiers, x, y, 0, false);
	}
}
