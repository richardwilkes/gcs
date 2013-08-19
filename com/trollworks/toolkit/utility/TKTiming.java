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

package com.trollworks.toolkit.utility;

import java.text.MessageFormat;

/** Provides basic timing facilities. */
public class TKTiming {
	private long	mBase;

	/** Creates a new {@link TKTiming}. */
	public TKTiming() {
		reset();
	}

	/**
	 * Waits until {@link #delta()} is greater than the specified number, then resets this timing
	 * object.
	 * 
	 * @param millis The amount to delay before reseting.
	 */
	public void delayUntilThenReset(long millis) {
		long delta = delta();

		if (delta < millis) {
			delta = millis - delta;
			if (delta > 50) {
				try {
					Thread.sleep(delta - 20);
				} catch (Exception exception) {
					assert false : TKDebug.throwableToString(exception);
				}
			}

			// To eliminate a warning about an empty control-flow statement,
			// the line:
			//
			// while (delta() < millis);
			//
			// has been changed to this:
			boolean keepGoing;
			do {
				keepGoing = delta() < millis;
			} while (keepGoing);
		}
		reset();
	}

	/**
	 * @return The delta, in milliseconds, since the timing object was created or last reset, then
	 *         resets it.
	 */
	public long deltaThenReset() {
		long now = System.currentTimeMillis();
		long oldBase = mBase;

		mBase = now;
		return now - oldBase;
	}

	/**
	 * @return The delta, in milliseconds, since the timing object was created or last reset.
	 */
	public long delta() {
		return System.currentTimeMillis() - mBase;
	}

	/** Resets the base time to this instant. */
	public void reset() {
		mBase = System.currentTimeMillis();
	}

	@Override public String toString() {
		return delta() + "ms"; //$NON-NLS-1$
	}

	/** @return The delta, in seconds. */
	public String toSeconds() {
		double amt = delta() / 1000.0;

		if (amt == 1.0) {
			return Msgs.ONE_SECOND;
		}
		return MessageFormat.format(Msgs.SECONDS_FORMAT, new Double(amt));
	}
}
