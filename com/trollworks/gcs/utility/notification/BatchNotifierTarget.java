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

package com.trollworks.gcs.utility.notification;

/**
 * Objects that want to be the target of notifications from a {@link Notifier} and want to be
 * notified when a batch change occurs must implement this interface.
 */
public interface BatchNotifierTarget extends NotifierTarget {
	/**
	 * Called when a series of notifications is about to be broadcast. The
	 * {@link BatchNotifierTarget} may or may not have intervening calls to
	 * {@link NotifierTarget#handleNotification(Object,String,Object)} made to it.
	 */
	public void enterBatchMode();

	/** Called after a series of notifications was broadcast. */
	public void leaveBatchMode();
}
