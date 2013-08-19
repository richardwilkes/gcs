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

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashSet;
import java.util.List;

/** Sorts rows by the sort sequence specified in the associated columns. */
public class RowSorter implements Comparator<Row> {
	private Column[]	mSortingOrder;

	private RowSorter(ArrayList<Column> columns) {
		int count = columns.size();
		Column[] orig = new Column[count];
		int pos = -1;
		int i;

		mSortingOrder = new Column[count];

		for (i = 0; i < count; i++) {
			Column column = columns.get(i);
			int order = column.getSortSequence();

			if (order >= 0 && order < count) {
				mSortingOrder[order] = column;
			} else {
				orig[i] = column;
				if (pos == -1) {
					pos = i;
				}
			}
		}

		if (pos != -1) {
			for (i = 0; i < count; i++) {
				if (mSortingOrder[i] == null) {
					mSortingOrder[i] = orig[pos++];
					while (pos < count && orig[pos] == null) {
						pos++;
					}
					if (pos >= count) {
						break;
					}
				}
			}
		}
	}

	/**
	 * Used to sort an outline.
	 * 
	 * @param columns The columns in the {@link Outline}.
	 * @param rows The rows in the {@link Outline}.
	 */
	public static void sort(ArrayList<Column> columns, ArrayList<Row> rows) {
		sort(columns, rows, false);
	}

	/**
	 * Used to sort an outline.
	 * 
	 * @param columns The columns in the {@link Outline}.
	 * @param rows The rows in the {@link Outline}.
	 * @param internal Pass in <code>true</code> if the actual row child storage should also be
	 *            sorted.
	 */
	public static void sort(ArrayList<Column> columns, ArrayList<Row> rows, boolean internal) {
		for (Column column : columns) {
			if (column.getSortSequence() != -1) {
				RowSorter rowSorter = new RowSorter(columns);

				Collections.sort(rows, rowSorter);
				if (internal) {
					for (Row row : collectContainerRows(rows, new HashSet<Row>())) {
						if (row.hasChildren()) {
							Collections.sort(row.getChildList(), rowSorter);
						}
					}
				}
				return;
			}
		}
	}

	/**
	 * Collects all container rows from the passed in rows and their children.
	 * 
	 * @param rows The rows to collect container rows from.
	 * @param containers The set to add the container rows to.
	 * @return The passed in set.
	 */
	public static HashSet<Row> collectContainerRows(List<Row> rows, HashSet<Row> containers) {
		for (Row row : rows) {
			if (row.canHaveChildren()) {
				containers.add(row);
				if (row.hasChildren()) {
					collectContainerRows(row.getChildren(), containers);
				}
			}
		}
		return containers;
	}

	public int compare(Row rowOne, Row rowTwo) {
		if (rowOne.getParent() == rowTwo.getParent()) {
			for (Column column : mSortingOrder) {
				int result;

				if (column == null) {
					return 0;
				}
				result = column.getRowCell(null).compare(column, rowOne, rowTwo);
				if (result != 0) {
					return column.isSortAscending() ? result : -result;
				}
			}
		} else {
			if (rowOne.isDescendentOf(rowTwo)) {
				return 1;
			}
			if (rowTwo.isDescendentOf(rowOne)) {
				return -1;
			}

			// Find common parents and compare them...
			Row[] oneParents = rowOne.getPath();
			Row[] twoParents = rowTwo.getPath();
			int max = Math.min(oneParents.length, twoParents.length);
			int i = 0;

			for (i = 0; i < max; i++) {
				if (oneParents[i] != twoParents[i]) {
					break;
				}
			}

			return compare(oneParents[i], twoParents[i]);
		}
		return 0;
	}
}
