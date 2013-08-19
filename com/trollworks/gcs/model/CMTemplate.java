/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.model;

import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.advantage.CMAdvantageList;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.equipment.CMEquipmentList;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMSkillList;
import com.trollworks.gcs.model.skill.CMTechnique;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.model.spell.CMSpellList;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.toolkit.collections.TKFilteredIterator;
import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.outline.TKRowIterator;

import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;
import java.util.Iterator;

/** A template. */
public class CMTemplate extends CMDataFile {
	private static final String	TAG_ROOT				= "template";						//$NON-NLS-1$
	private static final String	TAG_NOTES				= "notes";							//$NON-NLS-1$
	/** The prefix for all template IDs. */
	public static final String	TEMPLATE_PREFIX			= "gct.";							//$NON-NLS-1$
	/** The prefix used to indicate a point value is requested from {@link #getValueForID(String)}. */
	public static final String	POINTS_PREFIX			= TEMPLATE_PREFIX + "points.";		//$NON-NLS-1$
	/** The field ID for point total changes. */
	public static final String	ID_TOTAL_POINTS			= POINTS_PREFIX + "Total";			//$NON-NLS-1$
	/** The field ID for advantage point summary changes. */
	public static final String	ID_ADVANTAGE_POINTS		= POINTS_PREFIX + "Advantages";	//$NON-NLS-1$
	/** The field ID for disadvantage point summary changes. */
	public static final String	ID_DISADVANTAGE_POINTS	= POINTS_PREFIX + "Disadvantages";	//$NON-NLS-1$
	/** The field ID for quirk point summary changes. */
	public static final String	ID_QUIRK_POINTS			= POINTS_PREFIX + "Quirks";		//$NON-NLS-1$
	/** The field ID for skill point summary changes. */
	public static final String	ID_SKILL_POINTS			= POINTS_PREFIX + "Skills";		//$NON-NLS-1$
	/** The field ID for spell point summary changes. */
	public static final String	ID_SPELL_POINTS			= POINTS_PREFIX + "Spells";		//$NON-NLS-1$
	/** The field ID for notes changes. */
	public static final String	ID_NOTES				= TEMPLATE_PREFIX + "Notes";		//$NON-NLS-1$
	private TKOutlineModel		mAdvantages;
	private TKOutlineModel		mSkills;
	private TKOutlineModel		mSpells;
	private TKOutlineModel		mEquipment;
	private String				mNotes;
	private boolean				mNeedAdvantagesPointCalculation;
	private boolean				mNeedSkillPointCalculation;
	private boolean				mNeedSpellPointCalculation;
	private int					mCachedAdvantagePoints;
	private int					mCachedDisadvantagePoints;
	private int					mCachedQuirkPoints;
	private int					mCachedSkillPoints;
	private int					mCachedSpellPoints;

	/** Creates a new character with only default values set. */
	public CMTemplate() {
		super();
		characterInitialize();
		initialize();
		calculateAdvantagePoints();
		calculateSkillPoints();
		calculateSpellPoints();
	}

	/**
	 * Creates a new character from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @throws IOException if the data cannot be read or the file doesn't contain a valid character
	 *             sheet.
	 */
	public CMTemplate(File file) throws IOException {
		super(file);
		initialize();
	}

	private void characterInitialize() {
		mAdvantages = new TKOutlineModel();
		mSkills = new TKOutlineModel();
		mSpells = new TKOutlineModel();
		mEquipment = new TKOutlineModel();
		mNotes = ""; //$NON-NLS-1$
	}

	@Override public BufferedImage getFileIcon(boolean large) {
		return CSImage.getTemplateIcon(large);
	}

	@Override protected final void loadSelf(TKXMLReader reader, Object param) throws IOException {
		String marker = reader.getMarker();

		characterInitialize();
		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMAdvantageList.TAG_ROOT.equals(name)) {
					loadAdvantageList(reader);
				} else if (CMSkillList.TAG_ROOT.equals(name)) {
					loadSkillList(reader);
				} else if (CMSpellList.TAG_ROOT.equals(name)) {
					loadSpellList(reader);
				} else if (CMEquipmentList.TAG_ROOT.equals(name)) {
					loadEquipmentList(reader);
				} else if (TAG_NOTES.equals(name)) {
					mNotes = reader.readText();
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));

		calculateAdvantagePoints();
		calculateSkillPoints();
		calculateSpellPoints();
	}

	private void loadAdvantageList(TKXMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMAdvantage.TAG_ADVANTAGE.equals(name) || CMAdvantage.TAG_ADVANTAGE_CONTAINER.equals(name)) {
					mAdvantages.addRow(new CMAdvantage(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	private void loadSkillList(TKXMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMSkill.TAG_SKILL.equals(name) || CMSkill.TAG_SKILL_CONTAINER.equals(name)) {
					mSkills.addRow(new CMSkill(this, reader), true);
				} else if (CMTechnique.TAG_TECHNIQUE.equals(name)) {
					mSkills.addRow(new CMTechnique(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	private void loadSpellList(TKXMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMSpell.TAG_SPELL.equals(name) || CMSpell.TAG_SPELL_CONTAINER.equals(name)) {
					mSpells.addRow(new CMSpell(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	private void loadEquipmentList(TKXMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMEquipment.TAG_EQUIPMENT.equals(name) || CMEquipment.TAG_EQUIPMENT_CONTAINER.equals(name)) {
					mEquipment.addRow(new CMEquipment(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	@Override public String getXMLTagName() {
		return TAG_ROOT;
	}

	@Override protected void saveSelf(TKXMLWriter out) {
		Iterator<TKRow> iterator;

		if (mAdvantages.getRowCount() > 0) {
			out.startSimpleTagEOL(CMAdvantageList.TAG_ROOT);
			for (iterator = mAdvantages.getTopLevelRows().iterator(); iterator.hasNext();) {
				((CMAdvantage) iterator.next()).save(out, false);
			}
			out.endTagEOL(CMAdvantageList.TAG_ROOT, true);
		}

		if (mSkills.getRowCount() > 0) {
			out.startSimpleTagEOL(CMSkillList.TAG_ROOT);
			for (iterator = mSkills.getTopLevelRows().iterator(); iterator.hasNext();) {
				((CMRow) iterator.next()).save(out, false);
			}
			out.endTagEOL(CMSkillList.TAG_ROOT, true);
		}

		if (mSpells.getRowCount() > 0) {
			out.startSimpleTagEOL(CMSpellList.TAG_ROOT);
			for (iterator = mSpells.getTopLevelRows().iterator(); iterator.hasNext();) {
				((CMSpell) iterator.next()).save(out, false);
			}
			out.endTagEOL(CMSpellList.TAG_ROOT, true);
		}

		if (mEquipment.getRowCount() > 0) {
			out.startSimpleTagEOL(CMEquipmentList.TAG_ROOT);
			for (iterator = mEquipment.getTopLevelRows().iterator(); iterator.hasNext();) {
				((CMEquipment) iterator.next()).save(out, false);
			}
			out.endTagEOL(CMEquipmentList.TAG_ROOT, true);
		}
		out.simpleTag(TAG_NOTES, mNotes);
	}

	/**
	 * @param id The field ID to retrieve the data for.
	 * @return The value of the specified field ID, or <code>null</code> if the field ID is
	 *         invalid.
	 */
	public Object getValueForID(String id) {
		if (ID_ADVANTAGE_POINTS.equals(id)) {
			return new Integer(getAdvantagePoints());
		} else if (ID_DISADVANTAGE_POINTS.equals(id)) {
			return new Integer(getDisadvantagePoints());
		} else if (ID_QUIRK_POINTS.equals(id)) {
			return new Integer(getQuirkPoints());
		} else if (ID_SKILL_POINTS.equals(id)) {
			return new Integer(getSkillPoints());
		} else if (ID_SPELL_POINTS.equals(id)) {
			return new Integer(getSpellPoints());
		}
		return null;
	}

	@Override protected void startNotifyAtBatchLevelZero() {
		mNeedAdvantagesPointCalculation = false;
		mNeedSkillPointCalculation = false;
		mNeedSpellPointCalculation = false;
	}

	@Override public void notify(String type, Object data) {
		super.notify(type, data);
		if (CMAdvantage.ID_POINTS.equals(type) || CMAdvantage.ID_LEVELS.equals(type) || CMAdvantage.ID_LIST_CHANGED.equals(type)) {
			mNeedAdvantagesPointCalculation = true;
		}
		if (CMSkill.ID_POINTS.equals(type) || CMSkill.ID_LIST_CHANGED.equals(type)) {
			mNeedSkillPointCalculation = true;
		}
		if (CMSpell.ID_POINTS.equals(type) || CMSpell.ID_LIST_CHANGED.equals(type)) {
			mNeedSpellPointCalculation = true;
		}
	}

	@Override protected void endNotifyAtBatchLevelOne() {
		if (mNeedAdvantagesPointCalculation) {
			calculateAdvantagePoints();
			notify(ID_ADVANTAGE_POINTS, new Integer(getAdvantagePoints()));
			notify(ID_DISADVANTAGE_POINTS, new Integer(getDisadvantagePoints()));
			notify(ID_QUIRK_POINTS, new Integer(getQuirkPoints()));
		}
		if (mNeedSkillPointCalculation) {
			calculateSkillPoints();
			notify(ID_SKILL_POINTS, new Integer(getSkillPoints()));
		}
		if (mNeedSpellPointCalculation) {
			calculateSpellPoints();
			notify(ID_SPELL_POINTS, new Integer(getSpellPoints()));
		}
		if (mNeedAdvantagesPointCalculation || mNeedSkillPointCalculation || mNeedSpellPointCalculation) {
			notify(ID_TOTAL_POINTS, new Integer(getTotalPoints()));
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
		mCachedAdvantagePoints = 0;
		mCachedDisadvantagePoints = 0;
		mCachedQuirkPoints = 0;

		for (CMAdvantage advantage : getAdvantagesIterator()) {
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
		for (CMSkill skill : getSkillsIterable()) {
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
		for (CMSpell spell : getSpellsIterator()) {
			if (!spell.canHaveChildren()) {
				mCachedSpellPoints += spell.getPoints();
			}
		}
	}

	/** @return The outline model for the (dis)advantages. */
	public TKOutlineModel getAdvantagesModel() {
		return mAdvantages;
	}

	/** @return A recursive iterator over the (dis)advantages. */
	public TKRowIterator<CMAdvantage> getAdvantagesIterator() {
		return new TKRowIterator<CMAdvantage>(mAdvantages);
	}

	/** @return The outline model for the skills. */
	public TKOutlineModel getSkillsModel() {
		return mSkills;
	}

	/** @return A recursive iterable for the template's skills. */
	public Iterable<CMSkill> getSkillsIterable() {
		return new TKFilteredIterator<CMSkill>(new TKRowIterator<CMRow>(mSkills), CMSkill.class);
	}

	/** @return The outline model for the spells. */
	public TKOutlineModel getSpellsModel() {
		return mSpells;
	}

	/** @return A recursive iterator over the spells. */
	public TKRowIterator<CMSpell> getSpellsIterator() {
		return new TKRowIterator<CMSpell>(mSpells);
	}

	/** @return The outline model for the equipment. */
	public TKOutlineModel getEquipmentModel() {
		return mEquipment;
	}

	/** @return A recursive iterator over the equipment. */
	public TKRowIterator<CMEquipment> getEquipmentIterator() {
		return new TKRowIterator<CMEquipment>(mEquipment);
	}

	/** @return The notes. */
	public String getNotes() {
		return mNotes;
	}

	/**
	 * Sets the notes.
	 * 
	 * @param notes The new notes.
	 */
	public void setNotes(String notes) {
		if (!mNotes.equals(notes)) {
			if (!isUndoBeingApplied()) {
				addEdit(new CMTemplateNotesUndo(this, mNotes, notes));
			}
			mNotes = notes;
			notifySingle(ID_NOTES, mNotes);
		}
	}
}
