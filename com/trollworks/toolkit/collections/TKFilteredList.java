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

/**
 * A list that filters an {@link Iterable} to only contain objects of a specific type.
 * 
 * @param <T> The type of object the list should contain.
 */
public class TKFilteredList<T> extends ArrayList<T> {
	/**
	 * Creates a new {@link TKFilteredList}. Will not include <code>null</code> values.
	 * 
	 * @param iterable The {@link Iterable} to filter by type.
	 * @param contentClass The class of objects to extract from the collection.
	 */
	public TKFilteredList(Iterable<?> iterable, Class<T> contentClass) {
		this(iterable, contentClass, true);
	}

	/**
	 * Creates a new {@link TKFilteredList}.
	 * 
	 * @param iterable The {@link Iterable} to filter by type.
	 * @param contentClass The class of objects to extract from the collection.
	 * @param omitNulls Whether to omit <code>null</code> values or not.
	 */
	public TKFilteredList(Iterable<?> iterable, Class<T> contentClass, boolean omitNulls) {
		super();
		for (T item : new TKFilteredIterator<T>(iterable, contentClass, omitNulls)) {
			add(item);
		}
	}
}
