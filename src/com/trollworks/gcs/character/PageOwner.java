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

package com.trollworks.gcs.character;

import com.trollworks.toolkit.ui.print.PrintManager;

import java.awt.Graphics;
import java.awt.Insets;

/** Objects which control printable pages must implement this interface. */
public interface PageOwner {
	/** @return The page settings. */
	public PrintManager getPageSettings();

	/**
	 * Called so the page owner can draw headers, footers, etc.
	 * 
	 * @param page The page to work on.
	 * @param gc The graphics object to work with.
	 */
	public void drawPageAdornments(Page page, Graphics gc);

	/**
	 * @param page The page to work on.
	 * @return The insets required for the page adornments.
	 */
	public Insets getPageAdornmentsInsets(Page page);
}
