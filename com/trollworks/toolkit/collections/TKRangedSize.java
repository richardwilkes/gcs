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

import java.util.ArrayList;
import java.util.List;

/** Provides a way to specify discrete ranges that possess specific size values. */
public class TKRangedSize {
	private ArrayList<Measurement>	mSizes;

	/** Creates a new ranged size in which every index has a size of 1. */
	public TKRangedSize() {
		this(1);
	}

	/**
	 * Creates a new ranged size in which every index has the specified size.
	 * 
	 * @param size The starting range size.
	 */
	public TKRangedSize(int size) {
		mSizes = new ArrayList<Measurement>();
		mSizes.add(new Measurement(0, Integer.MAX_VALUE, size));
	}

	/**
	 * @param position The position to retrieve.
	 * @return The index within the range at the specified position, or -1 if it is not within the
	 *         range.
	 */
	public int getIndexAtPosition(int position) {
		int size = 0;

		for (Measurement measurement : mSizes) {
			int span = measurement.mSize * (1 + measurement.mEnd - measurement.mStart);

			if (size + span >= position) {
				if (measurement.mSize == 0) {
					return 0;
				}
				return measurement.mStart + (position - size) / measurement.mSize;
			}
			size += span;
		}
		return -1;
	}

	/**
	 * @param size The size to look for.
	 * @param startIndex The starting index to examine.
	 * @return The first index with the specified size, or -1 if there are none.
	 */
	public int getIndexWithSize(int size, int startIndex) {
		for (Measurement measurement : mSizes) {
			if (measurement.mSize == size) {
				if (measurement.mStart >= startIndex) {
					return measurement.mStart;
				}
				if (measurement.mEnd >= startIndex) {
					return startIndex;
				}
			}
		}
		return -1;
	}

	/**
	 * @param size The size to look for.
	 * @return A copy of the {@link Measurement}s in this {@link TKRangedSize} which have the
	 *         specified size.
	 */
	public List<Measurement> getMeasurementsWithSize(int size) {
		ArrayList<Measurement> list = new ArrayList<Measurement>();

		for (Measurement measurement : mSizes) {
			if (measurement.mSize == size) {
				list.add(new Measurement(measurement));
			}
		}
		return list;
	}

	/**
	 * @param index The index.
	 * @return The size of all indexes before the specified index within the range.
	 */
	public int getSizeBeforeIndex(int index) {
		return getSizeOfRange(-1, index - 1);
	}

	/**
	 * @param index The index.
	 * @return The size of the specified index within the range.
	 */
	public int getSizeOfIndex(int index) {
		return getSizeOfRange(index, index);
	}

	/**
	 * @param startIndex The starting index.
	 * @param endIndex The ending index.
	 * @return The size of all indexes from <code>startIndex</code> to <code>endIndex</code>,
	 *         inclusive.
	 */
	public int getSizeOfRange(int startIndex, int endIndex) {
		int size = 0;

		for (Measurement measurement : mSizes) {
			if (measurement.mStart <= endIndex && measurement.mEnd >= startIndex) {
				int start = measurement.mStart < startIndex ? startIndex : measurement.mStart;
				int end = measurement.mEnd > endIndex ? endIndex : measurement.mEnd;

				size += measurement.mSize * (1 + end - start);
			}

			if (measurement.mEnd >= endIndex) {
				break;
			}
		}

		return size;
	}

	/**
	 * Sets the size of a range, overriding any previous values.
	 * 
	 * @param startIndex The starting index of the range to set, inclusive.
	 * @param endIndex The ending index of the range to set, inclusive.
	 * @param size The new size for each of these indexes.
	 */
	public void setSizeOfRange(int startIndex, int endIndex, int size) {
		int count = mSizes.size();

		for (int i = 0; i < count; i++) {
			Measurement measurement = mSizes.get(i);

			if (measurement.mSize == size && measurement.mStart <= startIndex && measurement.mEnd >= startIndex - 1) {
				if (measurement.mEnd < endIndex) {
					measurement.mEnd = endIndex;
					while (++i < count) {
						measurement = mSizes.get(i);

						if (measurement.mEnd < endIndex + 1) {
							mSizes.remove(i--);
							count--;
						} else {
							measurement.mStart = endIndex + 1;
							break;
						}
					}
				}
				break;
			} else if (measurement.mStart == startIndex && measurement.mEnd == endIndex) {
				measurement.mSize = size;
				break;
			} else if (measurement.mStart < startIndex && measurement.mEnd > endIndex) {
				Measurement split = new Measurement(endIndex + 1, measurement.mEnd, measurement.mSize);

				if (i < count - 1) {
					mSizes.add(i + 1, split);
				} else {
					mSizes.add(split);
				}
				measurement.mEnd = startIndex - 1;
				mSizes.add(i + 1, new Measurement(startIndex, endIndex, size));
				break;
			} else if (measurement.mStart < startIndex && measurement.mEnd >= startIndex) {
				measurement.mEnd = startIndex - 1;
			} else if (measurement.mStart <= endIndex && measurement.mEnd > endIndex) {
				measurement.mStart = endIndex + 1;
				mSizes.add(i, new Measurement(startIndex, endIndex, size));
				break;
			} else if (measurement.mStart >= startIndex && measurement.mEnd <= endIndex) {
				mSizes.remove(i);
				i--;
				count--;
			} else if (measurement.mStart > endIndex) {
				mSizes.add(i, new Measurement(startIndex, endIndex, size));
				break;
			}
		}
	}

	/** A sized measurement. */
	public class Measurement {
		/** The starting index. */
		public int	mStart;
		/** The ending index. */
		public int	mEnd;
		/** The size of the range. */
		public int	mSize;

		/**
		 * Creates a new measurement.
		 * 
		 * @param startIndex The starting index.
		 * @param endIndex The ending index.
		 * @param sizeOfRange The size of the range.
		 */
		public Measurement(int startIndex, int endIndex, int sizeOfRange) {
			mStart = startIndex;
			mEnd = endIndex;
			mSize = sizeOfRange;
		}

		/**
		 * Creates a new measurement from an existing measurement.
		 * 
		 * @param measurement The measurement to clone.
		 */
		public Measurement(Measurement measurement) {
			this(measurement.mStart, measurement.mEnd, measurement.mSize);
		}
	}
}
