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

package com.trollworks.gcs.template;

import com.trollworks.gcs.character.CollectedModels;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.nio.file.Path;

/** A template. */
public class Template extends CollectedModels {
    /** Creates a new template with only default values set. */
    public Template() {
    }

    /**
     * Creates a new template from the specified file.
     *
     * @param path The path to load the data from.
     * @throws IOException if the data cannot be read or the file doesn't contain a valid template.
     */
    public Template(Path path) throws IOException {
        this();
        load(path);
    }

    @Override
    public String getJSONTypeName() {
        return "template";
    }

    @Override
    public FileType getFileType() {
        return FileType.TEMPLATE;
    }

    @Override
    public RetinaIcon getFileIcons() {
        return Images.GCT_FILE;
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        loadModels(m, state);
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        saveModels(w, saveType);
    }

    public void recalculate() {
        for (Row one : getEquipmentModel().getTopLevelRows()) {
            ((Equipment) one).update();
        }
        for (Row one : getOtherEquipmentModel().getTopLevelRows()) {
            ((Equipment) one).update();
        }
    }
}
