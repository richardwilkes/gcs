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

package com.trollworks.gcs.spell;

import com.trollworks.gcs.library.LibraryDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.I18n;

/** A list of spells from a library. */
public class SpellsDockable extends LibraryDockable {
    /** Creates a new {@link SpellsDockable}. */
    public SpellsDockable(SpellList list) {
        super(list);
    }

    @Override
    public SpellList getDataFile() {
        return (SpellList) super.getDataFile();
    }

    @Override
    protected String getUntitledBaseName() {
        return I18n.Text("Untitled Spells");
    }

    @Override
    protected ListOutline createOutline() {
        SpellList list = getDataFile();
        list.addTarget(this, Spell.ID_CATEGORY);
        return new SpellOutline(list, list.getModel());
    }
}
