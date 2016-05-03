/*
 * Copyright (c) 1998-2016 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.services;

/** Thrown when a method is not implemented */
public class NotImplementedException extends Exception {
	/**
	 * Creates a new {@link NotImplementedException}.
	 *
	 * @param message an error message to be thrown
	 */
	public NotImplementedException(String message) {
		super(message);
	}
}
