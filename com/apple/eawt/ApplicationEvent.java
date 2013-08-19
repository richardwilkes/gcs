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

package com.apple.eawt;

import java.util.EventObject;

/**
 * This is a stub implementation that allows non-Apple platforms to compile and run. Other than this
 * paragraph, all remaining comments in this file were copied from Apple's online Java documentation
 * for this class.
 * <p>
 * The class of events sent to ApplicationListener callbacks. Since these events are initiated by
 * Apple events they provide additional functionality over the EventObjects that they inherit their
 * basic characteristics from. For those events where it is appropriate, they store the file name of
 * the item that the event corresponds to. They are also unique in that they can be flagged as, and
 * tested for, having been handled.
 */
@SuppressWarnings("serial") public class ApplicationEvent extends EventObject {
	/**
	 * Creates an application event.
	 * 
	 * @param src The source of the event.
	 */
	ApplicationEvent(Object src) {
		super(src);
	}

	/**
	 * Determines whether an ApplicationListener has acted on a particular event. An event is marked
	 * as having been handled with <code>setHandled(true)</code>.
	 * 
	 * @return <code>true</code> if the event has been handled, otherwise <code>false</code>
	 */
	public boolean isHandled() {
		return false;
	}

	/**
	 * Sets the state of the event. After this method handles an ApplicationEvent, it may be useful
	 * to specify that it has been handled. This is usually used in conjunction with
	 * <code>isHandled()</code>. Set to <code>true</code> to designate that this event has been
	 * handled. By default it is <code>false</code>.
	 * 
	 * @param handled <code>true</code> if the event has been handled, otherwise
	 *            <code>false</code>
	 */
	public void setHandled(@SuppressWarnings("unused") boolean handled) {
		// Empty, as this is merely a stub...
	}

	/**
	 * Provides the filename associated with a particular AppleEvent. When the ApplicationEvent
	 * corresponds to an Apple Event that needs to act on a particular file, the ApplicationEvent
	 * carries the name of the specific file with it. For example, the Print and Open events refer
	 * to specific files. For these cases, this returns the appropriate file name.
	 * 
	 * @return The full path to the file associated with the event, if applicable, otherwise
	 *         <code>null</code>
	 */
	public String getFilename() {
		return null;
	}
}
