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

package com.trollworks.gcs.skill;

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
import com.trollworks.gcs.utility.xml.XMLNodeType;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;

/** A list of skills. */
public class SkillList extends ListFile {
    private static final int    CURRENT_JSON_VERSION = 1;
    /** The current version. */
    public static final  int    CURRENT_VERSION      = 1;
    /** The XML tag for {@link SkillList}s. */
    public static final  String TAG_ROOT             = "skill_list";

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
    }

    @Override
    public String getJSONTypeName() {
        return TAG_ROOT;
    }

    @Override
    public int getXMLTagVersion() {
        return CURRENT_VERSION;
    }

    @Override
    public String getXMLTagName() {
        return TAG_ROOT;
    }

    @Override
    public FileType getFileType() {
        return FileType.SKILL;
    }

    @Override
    public RetinaIcon getFileIcons() {
        return Images.SKL_FILE;
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
            if (Skill.TAG_SKILL.equals(type) || Skill.TAG_SKILL_CONTAINER.equals(type)) {
                model.addRow(new Skill(file, m1, state), true);
            } else if (Technique.TAG_TECHNIQUE.equals(type)) {
                model.addRow(new Technique(file, m1, state), true);
            } else {
                Log.warn("invalid skill type: " + type);
            }
        }
    }

    @Override
    protected void loadList(XMLReader reader, LoadState state) throws IOException {
        OutlineModel model  = getModel();
        String       marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();

                if (Skill.TAG_SKILL.equals(name) || Skill.TAG_SKILL_CONTAINER.equals(name)) {
                    model.addRow(new Skill(this, reader, state), true);
                } else if (Technique.TAG_TECHNIQUE.equals(name)) {
                    model.addRow(new Technique(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }
}
