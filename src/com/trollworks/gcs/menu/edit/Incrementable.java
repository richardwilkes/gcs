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

package com.trollworks.gcs.menu.edit;

/** Objects that can have their data incremented/decremented should implement this interface. */
public interface Incrementable {
	/** @return The title to use for the increment menu item. */
	String getIncrementTitle();

	/** @return The title to use for the decrement menu item. */
	String getDecrementTitle();

	/** @return Whether the data can be incremented. */
	boolean canIncrement();

	/** @return Whether the data can be decremented. */
	boolean canDecrement();

	/** Call to increment the data. */
	void increment();

	/** Call to decrement the data. */
	void decrement();
}
