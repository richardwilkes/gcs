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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.library.LibraryDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.I18n;

/** A list of advantage modifiers from a library. */
public class AdvantageModifiersDockable extends LibraryDockable {
    public AdvantageModifiersDockable(AdvantageModifierList list) {
        super(list);
    }

    @Override
    public AdvantageModifierList getDataFile() {
        return (AdvantageModifierList) super.getDataFile();
    }

    @Override
    protected String getUntitledBaseName() {
        return I18n.Text("Untitled Advantage Modifiers");
    }

    @Override
    protected ListOutline createOutline() {
        AdvantageModifierList list = getDataFile();
        list.addTarget(this, AdvantageModifier.ID_CATEGORY);
        return new AdvantageModifiersOutline(list);
    }
}
