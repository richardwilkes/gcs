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

import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.io.xml.XMLNodeType;
import com.trollworks.gcs.io.xml.XMLReader;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.FileType;

import java.io.IOException;

/** Data Object to hold several {@link EquipmentModifier}s */
public class EquipmentModifierList extends ListFile {
    private static final int    CURRENT_VERSION = 1;
    /** The XML tag for equipment modifier lists. */
    public static final  String TAG_ROOT        = "eqp_modifier_list";

    /** Creates new {@link EquipmentModifierList}. */
    public EquipmentModifierList() {
    }

    /**
     * Creates a new {@link EquipmentModifierList}.
     *
     * @param modifiers The {@link EquipmentModifierList} to clone.
     */
    public EquipmentModifierList(EquipmentModifierList modifiers) {
        this();
        for (Row Row : modifiers.getModel().getRows()) {
            getModel().getRows().add(Row);
        }
    }

    @Override
    protected void loadList(XMLReader reader, LoadState state) throws IOException {
        OutlineModel model  = getModel();
        String       marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();

                if (EquipmentModifier.TAG_MODIFIER.equals(name)) {
                    model.addRow(new EquipmentModifier(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }

    @Override
    // Not used
    public FileType getFileType() {
        return null;
    }

    @Override
    // Not used
    public RetinaIcon getFileIcons() {
        return null;
    }

    @Override
    public int getXMLTagVersion() {
        return CURRENT_VERSION;
    }

    @Override
    public String getXMLTagName() {
        return TAG_ROOT;
    }
}
