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

package com.trollworks.gcs.utility.undo;

import com.trollworks.gcs.utility.I18n;

import javax.swing.undo.CompoundEdit;

/** Provides a convenient way to collect multiple undos into a single undo. */
public class MultipleUndo extends CompoundEdit {
    private String mName;

    /**
     * Create a multiple undo edit.
     *
     * @param name The name of the undo edit.
     */
    public MultipleUndo(String name) {
        mName = name;
    }

    @Override
    public String getPresentationName() {
        return mName;
    }

    @Override
    public String getRedoPresentationName() {
        return I18n.Text("Redo ") + mName;
    }

    @Override
    public String getUndoPresentationName() {
        return I18n.Text("Undo ") + mName;
    }
}
