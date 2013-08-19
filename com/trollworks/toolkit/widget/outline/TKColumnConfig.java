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

package com.trollworks.toolkit.widget.outline;

/** Stores the configuration of an {@link TKColumn}. */
public class TKColumnConfig {
	/** The id of the column */
	public int		mID;
	/** The visibility of the column */
	public boolean	mVisible;
	/** The width of the column */
	public int		mWidth;
	/** the sort sequence of the column */
	public int		mSortSequence;
	/** <code>true</code> if the sort is ascending */
	public boolean	mSortAscending;

	/**
	 * Creates a new {@link TKColumnConfig} with the given id.
	 * 
	 * @param id The id of the column.
	 */
	public TKColumnConfig(int id) {
		this(id, true);
	}

	/**
	 * Creates a new {@link TKColumnConfig} with the given id and visibility.
	 * 
	 * @param id The id of the column.
	 * @param visible The visiblity of the column.
	 */
	public TKColumnConfig(int id, boolean visible) {
		this(id, visible, -1, false);
	}

	/**
	 * Creates a new {@link TKColumnConfig} with the given id, visibility, sort sequence and if the
	 * sort is ascending.
	 * 
	 * @param id The id of the column.
	 * @param visible The visiblity of the column.
	 * @param sortSequence The sort sequence of the column.
	 * @param sortAscending <code>true</code> if the sort is ascending.
	 */
	public TKColumnConfig(int id, boolean visible, int sortSequence, boolean sortAscending) {
		this(id, visible, -1, sortSequence, sortAscending);
	}

	/**
	 * Creates a new {@link TKColumnConfig} with the given id, visibility, width, sort sequence and
	 * if the sort is ascending.
	 * 
	 * @param id The id of the column.
	 * @param visible The visiblity of the column.
	 * @param width The width of the column.
	 * @param sortSequence The sort sequence of the column.
	 * @param sortAscending <code>true</code> if the sort is ascending.
	 */
	public TKColumnConfig(int id, boolean visible, int width, int sortSequence, boolean sortAscending) {
		mID = id;
		mVisible = visible;
		mWidth = width;
		mSortSequence = sortSequence;
		mSortAscending = sortAscending;
	}
}
