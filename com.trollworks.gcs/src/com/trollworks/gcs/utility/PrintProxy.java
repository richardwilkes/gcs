/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility;

import com.trollworks.gcs.page.PageSettings;

import java.awt.print.Printable;

/** Objects that want to be printable must implement this interface. */
public interface PrintProxy extends Printable {
    /** @return The title of the print job. */
    String getPrintJobTitle();

    /** @return The {@link PageSettings} to use. */
    PageSettings getPageSettings();

    /** @return {@code true} when printing is in progress. */
    boolean isPrinting();

    /** @param printing The current state to return from a call to {@link #isPrinting()}. */
    void setPrinting(boolean printing);
}
