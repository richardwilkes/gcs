/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui;

/**
 * Owners of a {@link Selection} that want notification when the selection is modified must
 * implement this interface.
 */
public interface SelectionOwner {
    /** Called whenever the selection is about to change. */
    void selectionAboutToChange();

    /** Called whenever the selection changes. */
    void selectionDidChange();
}
