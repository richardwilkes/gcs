/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.template;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.collections.FilteredIterator;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.image.StdImageSet;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.RowIterator;
import com.trollworks.toolkit.utility.FileType;
import com.trollworks.toolkit.utility.text.Text;

import java.io.File;
import java.io.IOException;
import java.util.Iterator;

/** A template. */
public class Template extends DataFile {
    /** The extension for templates. */
    public static final String  EXTENSION              = "gct"; //$NON-NLS-1$
    private static final int    CURRENT_VERSION        = 2;
    private static final String TAG_ROOT               = "template"; //$NON-NLS-1$
    private static final String TAG_OLD_NOTES          = "notes"; //$NON-NLS-1$
    /** The prefix for all template IDs. */
    public static final String  TEMPLATE_PREFIX        = "gct."; //$NON-NLS-1$
    /**
     * The prefix used to indicate a point value is requested from {@link #getValueForID(String)}.
     */
    public static final String  POINTS_PREFIX          = TEMPLATE_PREFIX + "points."; //$NON-NLS-1$
    /** The field ID for point total changes. */
    public static final String  ID_TOTAL_POINTS        = POINTS_PREFIX + "Total"; //$NON-NLS-1$
    /** The field ID for advantage point summary changes. */
    public static final String  ID_ADVANTAGE_POINTS    = POINTS_PREFIX + "Advantages"; //$NON-NLS-1$
    /** The field ID for disadvantage point summary changes. */
    public static final String  ID_DISADVANTAGE_POINTS = POINTS_PREFIX + "Disadvantages"; //$NON-NLS-1$
    /** The field ID for quirk point summary changes. */
    public static final String  ID_QUIRK_POINTS        = POINTS_PREFIX + "Quirks"; //$NON-NLS-1$
    /** The field ID for skill point summary changes. */
    public static final String  ID_SKILL_POINTS        = POINTS_PREFIX + "Skills"; //$NON-NLS-1$
    /** The field ID for spell point summary changes. */
    public static final String  ID_SPELL_POINTS        = POINTS_PREFIX + "Spells"; //$NON-NLS-1$
    private OutlineModel        mAdvantages;
    private OutlineModel        mSkills;
    private OutlineModel        mSpells;
    private OutlineModel        mEquipment;
    private OutlineModel        mNotes;
    private boolean             mNeedAdvantagesPointCalculation;
    private boolean             mNeedSkillPointCalculation;
    private boolean             mNeedSpellPointCalculation;
    private int                 mCachedAdvantagePoints;
    private int                 mCachedDisadvantagePoints;
    private int                 mCachedQuirkPoints;
    private int                 mCachedSkillPoints;
    private int                 mCachedSpellPoints;

    /** Creates a new character with only default values set. */
    public Template() {
        super();
        mAdvantages = new OutlineModel();
        mSkills     = new OutlineModel();
        mSpells     = new OutlineModel();
        mEquipment  = new OutlineModel();
        mNotes      = new OutlineModel();
    }

    /**
     * Creates a new character from the specified file.
     *
     * @param file The file to load the data from.
     * @throws IOException if the data cannot be read or the file doesn't contain a valid character
     *                     sheet.
     */
    public Template(File file) throws IOException {
        this();
        load(file);
    }

    @Override
    public FileType getFileType() {
        return FileType.getByExtension(EXTENSION);
    }

    @Override
    public StdImageSet getFileIcons() {
        return GCSImages.getTemplateDocumentIcons();
    }

    @Override
    protected final void loadSelf(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (AdvantageList.TAG_ROOT.equals(name)) {
                    loadAdvantageList(reader, state);
                } else if (SkillList.TAG_ROOT.equals(name)) {
                    loadSkillList(reader, state);
                } else if (SpellList.TAG_ROOT.equals(name)) {
                    loadSpellList(reader, state);
                } else if (EquipmentList.TAG_ROOT.equals(name)) {
                    loadEquipmentList(reader, state);
                } else if (NoteList.TAG_ROOT.equals(name)) {
                    loadNotesList(reader, state);
                } else if (TAG_OLD_NOTES.equals(name)) {
                    Note note = new Note(this, false);
                    note.setDescription(Text.standardizeLineEndings(reader.readText()));
                    mNotes.addRow(note, false);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
        calculateAdvantagePoints();
        calculateSkillPoints();
        calculateSpellPoints();
    }

    private void loadAdvantageList(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Advantage.TAG_ADVANTAGE.equals(name) || Advantage.TAG_ADVANTAGE_CONTAINER.equals(name)) {
                    mAdvantages.addRow(new Advantage(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }

    private void loadSkillList(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Skill.TAG_SKILL.equals(name) || Skill.TAG_SKILL_CONTAINER.equals(name)) {
                    mSkills.addRow(new Skill(this, reader, state), true);
                } else if (Technique.TAG_TECHNIQUE.equals(name)) {
                    mSkills.addRow(new Technique(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }

    private void loadSpellList(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Spell.TAG_SPELL.equals(name) || Spell.TAG_SPELL_CONTAINER.equals(name)) {
                    mSpells.addRow(new Spell(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }

    private void loadEquipmentList(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Equipment.TAG_EQUIPMENT.equals(name) || Equipment.TAG_EQUIPMENT_CONTAINER.equals(name)) {
                    mEquipment.addRow(new Equipment(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
    }

    private void loadNotesList(XMLReader reader, LoadState state) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                if (Note.TAG_NOTE.equals(name) || Note.TAG_NOTE_CONTAINER.equals(name)) {
                    mNotes.addRow(new Note(this, reader, state), true);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
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
    protected void saveSelf(XMLWriter out) {
        Iterator<Row> iterator;

        if (mAdvantages.getRowCount() > 0) {
            out.startSimpleTagEOL(AdvantageList.TAG_ROOT);
            for (iterator = mAdvantages.getTopLevelRows().iterator(); iterator.hasNext();) {
                ((Advantage) iterator.next()).save(out, false);
            }
            out.endTagEOL(AdvantageList.TAG_ROOT, true);
        }

        if (mSkills.getRowCount() > 0) {
            out.startSimpleTagEOL(SkillList.TAG_ROOT);
            for (iterator = mSkills.getTopLevelRows().iterator(); iterator.hasNext();) {
                ((ListRow) iterator.next()).save(out, false);
            }
            out.endTagEOL(SkillList.TAG_ROOT, true);
        }

        if (mSpells.getRowCount() > 0) {
            out.startSimpleTagEOL(SpellList.TAG_ROOT);
            for (iterator = mSpells.getTopLevelRows().iterator(); iterator.hasNext();) {
                ((Spell) iterator.next()).save(out, false);
            }
            out.endTagEOL(SpellList.TAG_ROOT, true);
        }

        if (mEquipment.getRowCount() > 0) {
            out.startSimpleTagEOL(EquipmentList.TAG_ROOT);
            for (iterator = mEquipment.getTopLevelRows().iterator(); iterator.hasNext();) {
                ((Equipment) iterator.next()).save(out, false);
            }
            out.endTagEOL(EquipmentList.TAG_ROOT, true);
        }

        if (mNotes.getRowCount() > 0) {
            out.startSimpleTagEOL(NoteList.TAG_ROOT);
            for (iterator = mNotes.getTopLevelRows().iterator(); iterator.hasNext();) {
                ((Note) iterator.next()).save(out, false);
            }
            out.endTagEOL(NoteList.TAG_ROOT, true);
        }
    }

    /**
     * @param id The field ID to retrieve the data for.
     * @return The value of the specified field ID, or <code>null</code> if the field ID is invalid.
     */
    public Object getValueForID(String id) {
        if (ID_ADVANTAGE_POINTS.equals(id)) {
            return Integer.valueOf(getAdvantagePoints());
        } else if (ID_DISADVANTAGE_POINTS.equals(id)) {
            return Integer.valueOf(getDisadvantagePoints());
        } else if (ID_QUIRK_POINTS.equals(id)) {
            return Integer.valueOf(getQuirkPoints());
        } else if (ID_SKILL_POINTS.equals(id)) {
            return Integer.valueOf(getSkillPoints());
        } else if (ID_SPELL_POINTS.equals(id)) {
            return Integer.valueOf(getSpellPoints());
        }
        return null;
    }

    @Override
    protected void startNotifyAtBatchLevelZero() {
        mNeedAdvantagesPointCalculation = false;
        mNeedSkillPointCalculation      = false;
        mNeedSpellPointCalculation      = false;
    }

    @Override
    public void notify(String type, Object data) {
        super.notify(type, data);
        if (Advantage.ID_POINTS.equals(type) || Advantage.ID_ROUND_COST_DOWN.equals(type) || Advantage.ID_LEVELS.equals(type) || Advantage.ID_LIST_CHANGED.equals(type)) {
            mNeedAdvantagesPointCalculation = true;
        }
        if (Skill.ID_POINTS.equals(type) || Skill.ID_LIST_CHANGED.equals(type)) {
            mNeedSkillPointCalculation = true;
        }
        if (Spell.ID_POINTS.equals(type) || Spell.ID_LIST_CHANGED.equals(type)) {
            mNeedSpellPointCalculation = true;
        }
    }

    @Override
    protected void endNotifyAtBatchLevelOne() {
        if (mNeedAdvantagesPointCalculation) {
            calculateAdvantagePoints();
            notify(ID_ADVANTAGE_POINTS, Integer.valueOf(getAdvantagePoints()));
            notify(ID_DISADVANTAGE_POINTS, Integer.valueOf(getDisadvantagePoints()));
            notify(ID_QUIRK_POINTS, Integer.valueOf(getQuirkPoints()));
        }
        if (mNeedSkillPointCalculation) {
            calculateSkillPoints();
            notify(ID_SKILL_POINTS, Integer.valueOf(getSkillPoints()));
        }
        if (mNeedSpellPointCalculation) {
            calculateSpellPoints();
            notify(ID_SPELL_POINTS, Integer.valueOf(getSpellPoints()));
        }
        if (mNeedAdvantagesPointCalculation || mNeedSkillPointCalculation || mNeedSpellPointCalculation) {
            notify(ID_TOTAL_POINTS, Integer.valueOf(getTotalPoints()));
        }
    }

    private int getTotalPoints() {
        return getAdvantagePoints() + getDisadvantagePoints() + getQuirkPoints() + getSkillPoints() + getSpellPoints();
    }

    /** @return The number of points spent on advantages. */
    public int getAdvantagePoints() {
        return mCachedAdvantagePoints;
    }

    /** @return The number of points spent on disadvantages. */
    public int getDisadvantagePoints() {
        return mCachedDisadvantagePoints;
    }

    /** @return The number of points spent on quirks. */
    public int getQuirkPoints() {
        return mCachedQuirkPoints;
    }

    private void calculateAdvantagePoints() {
        mCachedAdvantagePoints    = 0;
        mCachedDisadvantagePoints = 0;
        mCachedQuirkPoints        = 0;

        for (Advantage advantage : getAdvantagesIterator()) {
            if (!advantage.canHaveChildren()) {
                int pts = advantage.getAdjustedPoints();

                if (pts > 0) {
                    mCachedAdvantagePoints += pts;
                } else if (pts < -1) {
                    mCachedDisadvantagePoints += pts;
                } else if (pts == -1) {
                    mCachedQuirkPoints--;
                }
            }
        }
    }

    /** @return The number of points spent on skills. */
    public int getSkillPoints() {
        return mCachedSkillPoints;
    }

    private void calculateSkillPoints() {
        mCachedSkillPoints = 0;
        for (Skill skill : getSkillsIterable()) {
            if (!skill.canHaveChildren()) {
                mCachedSkillPoints += skill.getPoints();
            }
        }
    }

    /** @return The number of points spent on spells. */
    public int getSpellPoints() {
        return mCachedSpellPoints;
    }

    private void calculateSpellPoints() {
        mCachedSpellPoints = 0;
        for (Spell spell : getSpellsIterator()) {
            if (!spell.canHaveChildren()) {
                mCachedSpellPoints += spell.getPoints();
            }
        }
    }

    /** @return The outline model for the advantages. */
    public OutlineModel getAdvantagesModel() {
        return mAdvantages;
    }

    /** @return A recursive iterator over the advantages. */
    public RowIterator<Advantage> getAdvantagesIterator() {
        return new RowIterator<>(mAdvantages);
    }

    /** @return The outline model for the skills. */
    public OutlineModel getSkillsModel() {
        return mSkills;
    }

    /** @return A recursive iterable for the template's skills. */
    public Iterable<Skill> getSkillsIterable() {
        return new FilteredIterator<>(new RowIterator<ListRow>(mSkills), Skill.class);
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

    /** @return The outline model for the notes. */
    public OutlineModel getNotesModel() {
        return mNotes;
    }

    /** @return A recursive iterator over the notes. */
    public RowIterator<Note> getNoteIterator() {
        return new RowIterator<>(mNotes);
    }
}
