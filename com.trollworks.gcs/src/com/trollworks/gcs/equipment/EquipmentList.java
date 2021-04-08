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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;

import java.io.IOException;

/** A list of equipment. */
public class EquipmentList extends ListFile {
    private static final int    CURRENT_JSON_VERSION = 1;
    /** The XML tag for {@link EquipmentList}s. */
    public static final  String TAG_CARRIED_ROOT     = "equipment_list";
    /** The XML tag for {@link EquipmentList}s. */
    public static final  String TAG_OTHER_ROOT       = "other_equipment_list";

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
    }

    @Override
    public String getJSONTypeName() {
        return TAG_CARRIED_ROOT;
    }

    @Override
    public FileType getFileType() {
        return FileType.EQUIPMENT;
    }

    @Override
    public RetinaIcon getFileIcons() {
        return Images.EQP_FILE;
    }

    @Override
    protected void loadList(JsonArray a, LoadState state) throws IOException {
        loadIntoModel(this, a, getModel(), state);
    }

    public static void loadIntoModel(DataFile file, JsonArray a, OutlineModel model, LoadState state) throws IOException {
        int count = a.size();
        for (int i = 0; i < count; i++) {
            JsonMap m1   = a.getMap(i);
            String  type = m1.getString(DataFile.KEY_TYPE);
            if (Equipment.TAG_EQUIPMENT.equals(type) || Equipment.TAG_EQUIPMENT_CONTAINER.equals(type)) {
                model.addRow(new Equipment(file, m1, state), true);
            } else {
                Log.warn("invalid equipment type: " + type);
            }
        }
    }
}
