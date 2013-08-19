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

package com.trollworks.gcs.utility;

import java.util.Random;

/**
 * Provides a reasonably unique ID.
 * <p>
 * Support is also provided for using the ID as an "upgradable" ID that has both a sub-ID and a time
 * stamp, such that an object being tracked with a unique ID could determine if another object was a
 * newer, more up-to-date version of itself.
 */
public class UniqueID {
	private static final Random	RANDOM	= new Random();
	private long				mTimeStamp;
	private long				mSubID;

	/** Creates a unique ID. */
	public UniqueID() {
		this(System.currentTimeMillis(), RANDOM.nextLong());
	}

	/**
	 * Creates a unique ID.
	 * 
	 * @param uniqueID An ID obtained by called {@link #toString()} on a previous instance of
	 *            {@link UniqueID}.
	 */
	public UniqueID(String uniqueID) {
		try {
			int colon = uniqueID.indexOf(':');

			mTimeStamp = Long.parseLong(uniqueID.substring(0, colon), Character.MAX_RADIX);
			mSubID = Long.parseLong(uniqueID.substring(colon + 1), Character.MAX_RADIX);
		} catch (Exception exception) {
			mTimeStamp = System.currentTimeMillis();
			mSubID = RANDOM.nextLong();
		}
	}

	/**
	 * Creates a unique ID.
	 * 
	 * @param timeStamp The time stamp for this unique ID. Typically, the result of a call to
	 *            {@link System#currentTimeMillis()}.
	 * @param subID The sub-ID for this unique ID. Typically, the result of a call to
	 *            {@link Random#nextLong()}.
	 */
	public UniqueID(long timeStamp, long subID) {
		mTimeStamp = timeStamp;
		mSubID = subID;
	}

	@Override public final boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof UniqueID) {
			UniqueID other = (UniqueID) obj;

			return mTimeStamp == other.mTimeStamp && mSubID == other.mSubID;
		}
		return false;
	}

	/**
	 * @param obj The object to check.
	 * @return <code>true</code> if the passed-in object is also a {@link UniqueID} and their
	 *         sub-ID's match.
	 */
	public final boolean subIDEquals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof UniqueID) {
			UniqueID other = (UniqueID) obj;

			return mSubID == other.mSubID;
		}
		return false;
	}

	/**
	 * @param obj The object to check.
	 * @return <code>true</code> if the passed-in object is also a {@link UniqueID}, their
	 *         sub-ID's match, and the passed-in object's time stamp is newer.
	 */
	public final boolean subIDEqualsAndNewer(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof UniqueID) {
			UniqueID other = (UniqueID) obj;

			return mSubID == other.mSubID && mTimeStamp < other.mTimeStamp;
		}
		return false;
	}

	@Override public final int hashCode() {
		return (int) mSubID;
	}

	@Override public final String toString() {
		return Long.toString(mTimeStamp, Character.MAX_RADIX) + ":" + Long.toString(mSubID, Character.MAX_RADIX); //$NON-NLS-1$
	}

	/** @return The sub-ID. */
	public final long getSubID() {
		return mSubID;
	}

	/** @param subID The sub-ID to set. */
	public final void setSubID(long subID) {
		mSubID = subID;
	}

	/** @return The time stamp. */
	public final long getTimeStamp() {
		return mTimeStamp;
	}

	/** @param timeStamp The time stamp to set. */
	public final void setTimeStamp(long timeStamp) {
		mTimeStamp = timeStamp;
	}
}
