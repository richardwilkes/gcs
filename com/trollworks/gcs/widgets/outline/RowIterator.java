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

package com.trollworks.gcs.widgets.outline;

import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;

/**
 * Provides an iterator that will iterate over all rows (disclosed or not) in an outline model.
 * 
 * @param <T> The type of row being iterated over.
 */
public class RowIterator<T extends Row> implements Iterator<T>, Iterable<T> {
	private List<Row>			mList;
	private int					mIndex;
	private RowIterator<T>	mIterator;

	/**
	 * Creates an iterator that will iterate over all rows (disclosed or not) in the specified
	 * outline model.
	 * 
	 * @param model The model to iterator over.
	 */
	public RowIterator(OutlineModel model) {
		this(model.getTopLevelRows());
	}

	private RowIterator(List<Row> rows) {
		mList = rows;
	}

	public void remove() {
		throw new UnsupportedOperationException();
	}

	public boolean hasNext() {
		boolean hasNext = mIterator != null && mIterator.hasNext();

		if (!hasNext) {
			mIterator = null;
			hasNext = mIndex < mList.size();
		}
		return hasNext;
	}

	@SuppressWarnings("unchecked") public T next() {
		if (hasNext()) {
			if (mIterator == null) {
				Row row = mList.get(mIndex++);

				if (row.hasChildren()) {
					mIterator = new RowIterator<T>(row.getChildren());
				}
				return (T) row;
			}
			return mIterator.next();
		}
		throw new NoSuchElementException();
	}

	public Iterator<T> iterator() {
		return this;
	}
}
