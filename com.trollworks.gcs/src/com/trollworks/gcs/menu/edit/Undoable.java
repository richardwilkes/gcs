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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.utility.undo.StdUndoManager;

/**
 * Windows that want to participate in the standard {@link UndoCommand} and {@link RedoCommand}
 * processing must implement this interface.
 */
public interface Undoable {
    /** @return The {@link StdUndoManager} to use. */
    StdUndoManager getUndoManager();
}
