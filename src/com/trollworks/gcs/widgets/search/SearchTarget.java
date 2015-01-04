/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.widgets.search;

import com.trollworks.gcs.menu.edit.JumpToSearchTarget;

import java.util.List;

import javax.swing.ListCellRenderer;

/** Defines the methods which must be implemented to be the target of a {@link Search} control. */
public interface SearchTarget extends JumpToSearchTarget {
	/**
	 * Called to obtain a {@link ListCellRenderer} for displaying in the drop-down list.
	 * 
	 * @return The item renderer.
	 */
	ListCellRenderer<Object> getSearchRenderer();

	/**
	 * Called to have the target search itself with the specified filter and return the matching
	 * objects.
	 * 
	 * @param filter The filter to apply.
	 * @return The matching objects.
	 */
	List<Object> search(String filter);

	/**
	 * Called to have the target select the objects specified.
	 * 
	 * @param selection The objects to select.
	 */
	void searchSelect(List<Object> selection);
}
