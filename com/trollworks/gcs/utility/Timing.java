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

import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.text.MessageFormat;

/** Provides basic timing facilities. */
public class Timing {
	private static String	MSG_ONE_SECOND;
	private static String	MSG_SECONDS_FORMAT;
	private long			mBase;

	static {
		LocalizedMessages.initialize(Timing.class);
	}

	/** Creates a new {@link Timing}. */
	public Timing() {
		reset();
	}

	/**
	 * Waits until {@link #delta()} is greater than the specified number, then resets this timing
	 * object.
	 * 
	 * @param nanoseconds The amount to delay before reseting.
	 */
	public void delayUntilThenReset(long nanoseconds) {
		long delta = delta();

		if (delta < nanoseconds) {
			delta = nanoseconds - delta;
			if (delta > 20000000) {
				try {
					Thread.sleep((delta - 10000000) / 1000000);
				} catch (Exception exception) {
					assert false : Debug.throwableToString(exception);
				}
			}

			while (delta() < nanoseconds) {
				// Nothing to do.
			}
		}
		reset();
	}

	/**
	 * @return The delta, in nanoseconds, since the timing object was created or last reset, then
	 *         resets it.
	 */
	public long deltaThenReset() {
		long now = System.nanoTime();
		long oldBase = mBase;

		mBase = now;
		return now - oldBase;
	}

	/**
	 * @return The delta, in nanoseconds, since the timing object was created or last reset.
	 */
	public long delta() {
		return System.nanoTime() - mBase;
	}

	/** Resets the base time to this instant. */
	public void reset() {
		mBase = System.nanoTime();
	}

	@Override public String toString() {
		double amt = delta() / 1000000000.0;
		if (amt == 1.0) {
			return MSG_ONE_SECOND;
		}
		return MessageFormat.format(MSG_SECONDS_FORMAT, new Double(amt));
	}
}
