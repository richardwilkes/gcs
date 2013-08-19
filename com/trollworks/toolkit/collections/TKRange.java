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

package com.trollworks.toolkit.collections;

/** Provides a way to specify an integer range. */
public class TKRange {
	private int	mPosition;
	private int	mLength;

	/** Create an empty range. */
	public TKRange() {
		this(0, 0);
	}

	/**
	 * Create a new range.
	 * 
	 * @param position The starting position of this range.
	 * @param length The length of the range.
	 */
	public TKRange(int position, int length) {
		mPosition = position;
		mLength = length;
	}

	/**
	 * @return The last position in the range. This is the same as calling
	 *         <code>{@link #getPosition()} + {@link #getLength()} - 1</code>.
	 */
	public int getLastPosition() {
		return mPosition + mLength - 1;
	}

	/** @return The length of the range. */
	public int getLength() {
		return mLength;
	}

	/** @return The starting position of the range. */
	public int getPosition() {
		return mPosition;
	}

	@Override public String toString() {
		return "Range[Position: " + getPosition() + ", Length: " + getLength() + ", End Position: " + getLastPosition() + "]"; //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
	}
}
