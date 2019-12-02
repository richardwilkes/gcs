/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.library;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.image.StdImageSet;
import com.trollworks.toolkit.utility.FileType;

import java.io.File;
import java.io.IOException;

/** Holds the contents of a library file. */
public class LibraryFile extends DataFile {
    /** The current version. */
    public static final int    CURRENT_VERSION = 1;
    /** The XML tag for library files. */
    public static final String TAG_ROOT        = "gcs_library";
    /** The extension for library files. */
    public static final String EXTENSION       = "glb";
    private AdvantageList      mAdvantages;
    private SkillList          mSkills;
    private SpellList          mSpells;
    private EquipmentList      mEquipment;
    private NoteList           mNotes;

    /** Creates a new, empty, {@link LibraryFile}. */
    public LibraryFile() {
        super();
        setup();
    }

    /**
     * Creates a new {@link LibraryFile} from the specified file.
     *
     * @param file The file to load the data from.
     * @throws IOException if the data cannot be read or the file doesn't contain valid information.
     */
    public LibraryFile(final File file) throws IOException {
        this();
        load(file);
    }

    private void setup() {
        mAdvantages = new AdvantageList();
        mSkills     = new SkillList();
        mSpells     = new SpellList();
        mEquipment  = new EquipmentList();
        mNotes      = new NoteList();
    }

    @Override
    public FileType getFileType() {
        return FileType.getByExtension(EXTENSION);
    }

    @Override
    public StdImageSet getFileIcons() {
        // Unused
        return null;
    }

    @Override
    public boolean matchesRootTag(String name) {
        return TAG_ROOT.equals(name) || AdvantageList.TAG_ROOT.equals(name) || SkillList.TAG_ROOT.equals(name) || SpellList.TAG_ROOT.equals(name) || EquipmentList.TAG_CARRIED_ROOT.equals(name) || NoteList.TAG_ROOT.equals(name);
    }

    @Override
    public String getXMLTagName() {
        return TAG_ROOT;
    }

    @Override
    public int getXMLTagVersion() {
        return CURRENT_VERSION;
    }

    @Override
    protected void loadSelf(XMLReader reader, LoadState state) throws IOException {
        setup();
        String name = reader.getName();
        if (TAG_ROOT.equals(name)) {
            String marker = reader.getMarker();
            do {
                if (reader.next() == XMLNodeType.START_TAG) {
                    name = reader.getName();
                    if (AdvantageList.TAG_ROOT.equals(name)) {
                        mAdvantages.load(reader, state);
                    } else if (SkillList.TAG_ROOT.equals(name)) {
                        mSkills.load(reader, state);
                    } else if (SpellList.TAG_ROOT.equals(name)) {
                        mSpells.load(reader, state);
                    } else if (EquipmentList.TAG_CARRIED_ROOT.equals(name)) {
                        mEquipment.load(reader, state);
                    } else if (NoteList.TAG_ROOT.equals(name)) {
                        mNotes.load(reader, state);
                    } else {
                        reader.skipTag(name);
                    }
                }
            } while (reader.withinMarker(marker));
        }
    }

    @Override
    protected void saveSelf(XMLWriter out) {
        mAdvantages.save(out, false, true);
        mSkills.save(out, false, true);
        mSpells.save(out, false, true);
        mEquipment.save(out, false, true);
    }

    /** @return The {@link AdvantageList}. */
    public AdvantageList getAdvantageList() {
        return mAdvantages;
    }

    /** @return The {@link SkillList}. */
    public SkillList getSkillList() {
        return mSkills;
    }

    /** @return The {@link SpellList}. */
    public SpellList getSpellList() {
        return mSpells;
    }

    /** @return The {@link EquipmentList}. */
    public EquipmentList getEquipmentList() {
        return mEquipment;
    }

    /** @return The {@link NoteList}. */
    public NoteList getNoteList() {
        return mNotes;
    }
}
