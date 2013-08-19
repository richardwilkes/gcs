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

package com.trollworks.toolkit.io.conduit;

/**
 * Clients that want to receive messages from a {@link TKConduit} must implement this interface.
 */
public interface TKConduitReceiver {
	/**
	 * Called when a message is received.
	 * 
	 * @param msg The message.
	 */
	public void conduitMessageReceived(TKConduitMessage msg);

	/**
	 * Called to get the filter to apply to incoming message IDs, if any. This method is only called
	 * once, when the broker is starting up.
	 * 
	 * @return The string to match IDs against, or <code>null</code> if any ID is OK.
	 */
	public String getConduitMessageIDFilter();

	/**
	 * Called to get the filter to apply to incoming message users, if any. This method is only
	 * called once, when the broker is starting up.
	 * 
	 * @return The string to match users against, or <code>null</code> if any user is OK.
	 */
	public String getConduitMessageUserFilter();
}
