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

package com.trollworks.gcs.character;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.datafile.Updatable;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.RowIterator;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.UUID;

public abstract class CollectedModels extends DataFile {
    public static final String       KEY_ADVANTAGES      = "advantages";
    public static final String       KEY_SKILLS          = "skills";
    public static final String       KEY_SPELLS          = "spells";
    public static final String       KEY_EQUIPMENT       = "equipment";
    public static final String       KEY_OTHER_EQUIPMENT = "other_equipment";
    public static final String       KEY_NOTES           = "notes";
    private             OutlineModel mAdvantages;
    private             OutlineModel mSkills;
    private             OutlineModel mSpells;
    private             OutlineModel mEquipment;
    private             OutlineModel mOtherEquipment;
    private             OutlineModel mNotes;

    public CollectedModels() {
        mAdvantages = new OutlineModel();
        mSkills = new OutlineModel();
        mSpells = new OutlineModel();
        mEquipment = new OutlineModel();
        mOtherEquipment = new OutlineModel();
        mOtherEquipment.setProperty(EquipmentList.TAG_OTHER_ROOT, Boolean.TRUE);
        mNotes = new OutlineModel();
    }

    protected void loadModels(JsonMap m, LoadState state) throws IOException {
        AdvantageList.loadIntoModel(this, m.getArray(KEY_ADVANTAGES), mAdvantages, state);
        SkillList.loadIntoModel(this, m.getArray(KEY_SKILLS), mSkills, state);
        SpellList.loadIntoModel(this, m.getArray(KEY_SPELLS), mSpells, state);
        EquipmentList.loadIntoModel(this, m.getArray(KEY_EQUIPMENT), mEquipment, state);
        EquipmentList.loadIntoModel(this, m.getArray(KEY_OTHER_EQUIPMENT), mOtherEquipment, state);
        NoteList.loadIntoModel(this, m.getArray(KEY_NOTES), mNotes, state);
    }

    protected void saveModels(JsonWriter w, SaveType saveType) throws IOException {
        ListRow.saveList(w, KEY_ADVANTAGES, mAdvantages.getTopLevelRows(), saveType);
        ListRow.saveList(w, KEY_SKILLS, mSkills.getTopLevelRows(), saveType);
        ListRow.saveList(w, KEY_SPELLS, mSpells.getTopLevelRows(), saveType);
        ListRow.saveList(w, KEY_EQUIPMENT, mEquipment.getTopLevelRows(), saveType);
        ListRow.saveList(w, KEY_OTHER_EQUIPMENT, mOtherEquipment.getTopLevelRows(), saveType);
        ListRow.saveList(w, KEY_NOTES, mNotes.getTopLevelRows(), saveType);
    }

    /** @return The outline model for the advantages. */
    public OutlineModel getAdvantagesModel() {
        return mAdvantages;
    }

    /**
     * @param includeDisabled {@code true} if disabled entries should be included.
     * @return A recursive iterator over the advantages.
     */
    public RowIterator<Advantage> getAdvantagesIterator(boolean includeDisabled) {
        if (includeDisabled) {
            return new RowIterator<>(mAdvantages);
        }
        return new RowIterator<>(mAdvantages, Advantage::isEnabled);
    }

    /** @return The outline model for the skills. */
    public OutlineModel getSkillsModel() {
        return mSkills;
    }

    /** @return A recursive iterable for the skills. */
    public RowIterator<Skill> getSkillsIterator() {
        return new RowIterator<>(mSkills);
    }

    /** @return The outline model for the spells. */
    public OutlineModel getSpellsModel() {
        return mSpells;
    }

    /** @return A recursive iterator over the spells. */
    public RowIterator<Spell> getSpellsIterator() {
        return new RowIterator<>(mSpells);
    }

    /** @return The outline model for the equipment. */
    public OutlineModel getEquipmentModel() {
        return mEquipment;
    }

    /** @return A recursive iterator over the equipment. */
    public RowIterator<Equipment> getEquipmentIterator() {
        return new RowIterator<>(mEquipment);
    }

    /** @return The outline model for the other equipment. */
    public OutlineModel getOtherEquipmentModel() {
        return mOtherEquipment;
    }

    /** @return A recursive iterator over the other equipment. */
    public RowIterator<Equipment> getOtherEquipmentIterator() {
        return new RowIterator<>(mOtherEquipment);
    }

    /** @return The outline model for the notes. */
    public OutlineModel getNotesModel() {
        return mNotes;
    }

    /** @return A recursive iterator over the notes. */
    public RowIterator<Note> getNoteIterator() {
        return new RowIterator<>(mNotes);
    }

    @Override
    public void getContainedUpdatables(Map<UUID, Updatable> updatables) {
        ListFile.getContainedUpdatables(mAdvantages, updatables);
        ListFile.getContainedUpdatables(mSkills, updatables);
        ListFile.getContainedUpdatables(mSpells, updatables);
        ListFile.getContainedUpdatables(mEquipment, updatables);
        ListFile.getContainedUpdatables(mOtherEquipment, updatables);
        ListFile.getContainedUpdatables(mNotes, updatables);
    }
}
