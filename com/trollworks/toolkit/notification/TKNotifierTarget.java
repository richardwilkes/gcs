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

package com.trollworks.toolkit.notification;

/**
 * Objects that want to be the target of notifications from a {@link TKNotifier} must implement this
 * interface.
 */
public interface TKNotifierTarget {
	/**
	 * Called when a notification is delivered.
	 * 
	 * @param producer The producer of the notification.
	 * @param name The notification name.
	 * @param data Extra data specific to the notification.
	 */
	public void handleNotification(Object producer, String name, Object data);
}
