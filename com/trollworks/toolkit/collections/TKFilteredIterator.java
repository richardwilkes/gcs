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

import java.util.Iterator;
import java.util.NoSuchElementException;

/**
 * An {@link Iterator} that filters an {@link Iterable} to return only objects of a specific type.
 * 
 * @param <T> The type of object the {@link Iterator} should return.
 */
public class TKFilteredIterator<T> implements Iterator<T>, Iterable<T> {
	private Iterator<?>	mIterator;
	private T			mNext;
	private Class<T>	mContentClass;
	private boolean		mNextValid;
	private boolean		mOmitNulls;

	/**
	 * Creates a new {@link TKFilteredIterator}. Will not include <code>null</code> values.
	 * 
	 * @param iterable The {@link Iterable} to filter by type.
	 * @param contentClass The class of objects to extract from the collection.
	 */
	public TKFilteredIterator(Iterable<?> iterable, Class<T> contentClass) {
		this(iterable, contentClass, true);
	}

	/**
	 * Creates a new {@link TKFilteredIterator}.
	 * 
	 * @param iterable The {@link Iterable} to filter by type.
	 * @param contentClass The class of objects to extract from the collection.
	 * @param omitNulls Whether to omit <code>null</code> values or not.
	 */
	public TKFilteredIterator(Iterable<?> iterable, Class<T> contentClass, boolean omitNulls) {
		mIterator = iterable.iterator();
		mContentClass = contentClass;
		mOmitNulls = omitNulls;
	}

	public boolean hasNext() {
		if (mNextValid) {
			return true;
		}
		while (mIterator.hasNext()) {
			Object obj = mIterator.next();

			if (obj == null) {
				if (!mOmitNulls) {
					mNext = null;
					mNextValid = true;
					return true;
				}
			} else if (mContentClass.isInstance(obj)) {
				mNext = mContentClass.cast(obj);
				mNextValid = true;
				return true;
			}
		}
		return false;
	}

	public T next() {
		if (!mNextValid) {
			if (!hasNext()) {
				throw new NoSuchElementException();
			}
		}
		mNextValid = false;
		return mNext;
	}

	/** Not supported. */
	public void remove() {
		throw new UnsupportedOperationException();
	}

	public Iterator<T> iterator() {
		return this;
	}
}
