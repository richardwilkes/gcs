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

package com.trollworks.gcs.utility;

import java.awt.EventQueue;
import java.util.Timer;
import java.util.TimerTask;

/**
 * A delayed task that will queue a {@link Runnable} onto the event queue thread for later
 * invocation.
 */
public class DelayedTask extends TimerTask {
	private static Timer	TIMER	= null;
	private Runnable		mRunnable;
	private boolean			mOnEventQueue;

	/**
	 * Creates a new {@link DelayedTask}. The {@link Runnable} passed in will be invoked on the
	 * event queue thread.
	 * 
	 * @param runnable The {@link Runnable} to invoke.
	 */
	public DelayedTask(Runnable runnable) {
		this(runnable, true);
	}

	/**
	 * Creates a new {@link DelayedTask}.
	 * 
	 * @param runnable The {@link Runnable} to invoke.
	 * @param onEventQueue Pass in <code>true</code> to cause the runnable task to be queued on
	 *            the primary event queue when it fires rather than being called directly.
	 */
	public DelayedTask(Runnable runnable, boolean onEventQueue) {
		mRunnable = runnable;
		mOnEventQueue = onEventQueue;
	}

	/** @return A common {@link Timer} object. */
	public static synchronized Timer getCommonTimer() {
		if (TIMER == null) {
			TIMER = new Timer(true);
		}
		return TIMER;
	}

	/** Puts the {@link Runnable} assigned to this task onto the event queue. */
	@Override public void run() {
		if (mRunnable != null) {
			if (mOnEventQueue) {
				EventQueue.invokeLater(mRunnable);
			} else {
				try {
					mRunnable.run();
				} catch (Exception exception) {
					assert false : Debug.throwableToString(exception);
				}
			}
		}
	}

	/**
	 * Schedules the specified {@link Runnable} for execution on the primary event queue thread
	 * after the specified delay.
	 * 
	 * @param runnable The runnable to schedule.
	 * @param delay The number of milliseconds to wait.
	 */
	public static void schedule(Runnable runnable, long delay) {
		schedule(runnable, delay, true);
	}

	/**
	 * Schedules the specified {@link Runnable} for execution after the specified delay.
	 * 
	 * @param runnable The runnable to schedule.
	 * @param delay The number of milliseconds to wait.
	 * @param onEventQueue Pass in <code>true</code> to cause the runnable task to be queued on
	 *            the primary event queue when it fires rather than being called directly.
	 */
	public static void schedule(Runnable runnable, long delay, boolean onEventQueue) {
		getCommonTimer().schedule(new DelayedTask(runnable, onEventQueue), delay);
	}
}
