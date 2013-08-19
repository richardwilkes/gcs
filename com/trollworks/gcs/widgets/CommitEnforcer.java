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

package com.trollworks.gcs.widgets;

import java.awt.Component;
import java.awt.KeyboardFocusManager;
import java.awt.event.FocusEvent;
import java.lang.reflect.Method;
import java.security.AccessController;
import java.security.PrivilegedAction;

/**
 * Sends a 'focus lost' followed by a 'focus gained' event to the current keyboard focus, with the
 * intent that it will cause it to commit any changes it had pending.
 */
public class CommitEnforcer implements PrivilegedAction<Object> {
	private Component	mTarget;

	/**
	 * Sends a 'focus lost' followed by a 'focus gained' event to the current keyboard focus, with
	 * the intent that it will cause it to commit any changes it had pending.
	 */
	public static void forceFocusToAccept() {
		KeyboardFocusManager focusManager = KeyboardFocusManager.getCurrentKeyboardFocusManager();
		Component focus = focusManager.getPermanentFocusOwner();
		if (focus == null) {
			focus = focusManager.getFocusOwner();
		}
		if (focus != null) {
			CommitEnforcer action = new CommitEnforcer(focus);
			if (System.getSecurityManager() == null) {
				action.run();
			} else {
				AccessController.doPrivileged(action);
			}
		}
	}

	private CommitEnforcer(Component comp) {
		mTarget = comp;
	}

	public Object run() {
		try {
			Class<?> cls = mTarget.getClass();
			Method method = null;
			while (method == null && cls != null) {
				try {
					method = cls.getDeclaredMethod("processFocusEvent", FocusEvent.class); //$NON-NLS-1$
				} catch (NoSuchMethodException nsm) {
					cls = cls.getSuperclass();
				}
			}
			if (method != null && cls != null) {
				method.setAccessible(true);
				method.invoke(mTarget, new FocusEvent(mTarget, FocusEvent.FOCUS_LOST, false, null));
				method.invoke(mTarget, new FocusEvent(mTarget, FocusEvent.FOCUS_GAINED, false, null));
			}
		} catch (Exception exception) {
			exception.printStackTrace();
		}
		return null;
	}
}
