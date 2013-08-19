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

import java.awt.event.KeyEvent;

/** Provides a standard way to filter key events. */
public interface TKKeyEventFilter {
	/**
	 * Called to filter the specified key event. If the filter wants to prevent the key event from
	 * being processed, it should return <code>true</code>. It is not necessary to consume the
	 * event, unless the filter also wants to prevent parent components from processing the key
	 * event.
	 * 
	 * @param owner The owning component.
	 * @param event The key event to filter.
	 * @param isReal Will be <code>true</code> if this is a real event, <code>false</code> if
	 *            the event was created just to test what the filter would filter out.
	 * @return <code>true</code> if the key should be filtered.
	 */
	public boolean filterKeyEvent(TKPanel owner, KeyEvent event, boolean isReal);
}
