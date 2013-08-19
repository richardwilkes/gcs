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

package com.trollworks.gcs.model.spell;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.skill.CMSkillDefault;
import com.trollworks.gcs.model.skill.CMSkillLevel;
import com.trollworks.gcs.model.weapon.CMMeleeWeaponStats;
import com.trollworks.gcs.model.weapon.CMRangedWeaponStats;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.spell.CSSpellColumnID;
import com.trollworks.gcs.ui.spell.CSSpellEditor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.widget.outline.TKColumn;

import java.awt.image.BufferedImage;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

/** A GURPS Spell. */
public class CMSpell extends CMRow {
	/** The XML tag used for items. */
	public static final String			TAG_SPELL					= "spell";									//$NON-NLS-1$
	/** The XML tag used for containers. */
	public static final String			TAG_SPELL_CONTAINER			= "spell_container";						//$NON-NLS-1$
	private static final String			TAG_NAME					= "name";									//$NON-NLS-1$
	private static final String			TAG_TECH_LEVEL				= "tech_level";							//$NON-NLS-1$
	private static final String			TAG_COLLEGE					= "college";								//$NON-NLS-1$
	private static final String			TAG_SPELL_CLASS				= "spell_class";							//$NON-NLS-1$
	private static final String			TAG_CASTING_COST			= "casting_cost";							//$NON-NLS-1$
	private static final String			TAG_MAINTENANCE_COST		= "maintenance_cost";						//$NON-NLS-1$
	private static final String			TAG_CASTING_TIME			= "casting_time";							//$NON-NLS-1$
	private static final String			TAG_DURATION				= "duration";								//$NON-NLS-1$
	private static final String			TAG_POINTS					= "points";								//$NON-NLS-1$
	private static final String			TAG_REFERENCE				= "reference";								//$NON-NLS-1$
	private static final String			ATTRIBUTE_VERY_HARD			= "very_hard";								//$NON-NLS-1$
	/** The prefix used in front of all IDs for the spells. */
	public static final String			PREFIX						= CMCharacter.CHARACTER_PREFIX + "spell.";	//$NON-NLS-1$
	/** The field ID for name changes. */
	public static final String			ID_NAME						= PREFIX + "Name";							//$NON-NLS-1$
	/** The field ID for tech level changes. */
	public static final String			ID_TECH_LEVEL				= PREFIX + "TechLevel";					//$NON-NLS-1$
	/** The field ID for college changes. */
	public static final String			ID_COLLEGE					= PREFIX + "College";						//$NON-NLS-1$
	/** The field ID for spell class changes. */
	public static final String			ID_SPELL_CLASS				= PREFIX + "Class";						//$NON-NLS-1$
	/** The field ID for casting cost changes */
	public static final String			ID_CASTING_COST				= PREFIX + "CastingCost";					//$NON-NLS-1$
	/** The field ID for maintainance cost changes */
	public static final String			ID_MAINTENANCE_COST			= PREFIX + "MaintenanceCost";				//$NON-NLS-1$
	/** The field ID for casting time changes */
	public static final String			ID_CASTING_TIME				= PREFIX + "CastingTime";					//$NON-NLS-1$
	/** The field ID for duration changes */
	public static final String			ID_DURATION					= PREFIX + "Duration";						//$NON-NLS-1$
	/** The field ID for point changes. */
	public static final String			ID_POINTS					= PREFIX + "Points";						//$NON-NLS-1$
	/** The field ID for level changes. */
	public static final String			ID_LEVEL					= PREFIX + "Level";						//$NON-NLS-1$
	/** The field ID for page reference changes. */
	public static final String			ID_REFERENCE				= PREFIX + "Reference";					//$NON-NLS-1$
	/** The field ID for difficulty changes. */
	public static final String			ID_IS_VERY_HARD				= PREFIX + "Difficulty";					//$NON-NLS-1$
	/** The field ID for when the row hierarchy changes. */
	public static final String			ID_LIST_CHANGED				= PREFIX + "ListChanged";					//$NON-NLS-1$
	/** The field ID for when the spell becomes or stops being a weapon. */
	public static final String			ID_WEAPON_STATUS_CHANGED	= PREFIX + "WeaponStatus";					//$NON-NLS-1$
	private static final String			EMPTY						= "";										//$NON-NLS-1$
	private static final String			NEWLINE						= "\n";									//$NON-NLS-1$
	private static final String			SPACE						= " ";										//$NON-NLS-1$
	private String						mName;
	private String						mTechLevel;
	private String						mCollege;
	private String						mSpellClass;
	private String						mCastingCost;
	private String						mMaintenance;
	private String						mCastingTime;
	private String						mDuration;
	private int							mPoints;
	private int							mLevel;
	private int							mRelativeLevel;
	private String						mReference;
	private boolean						mIsVeryHard;
	private ArrayList<CMWeaponStats>	mWeapons;

	/**
	 * Creates a new spell.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param isContainer Whether or not this row allows children.
	 */
	public CMSpell(CMDataFile dataFile, boolean isContainer) {
		super(dataFile, isContainer);
		mName = Msgs.DEFAULT_NAME;
		mTechLevel = null;
		mCollege = EMPTY;
		mSpellClass = isContainer ? EMPTY : Msgs.DEFAULT_SPELL_CLASS;
		mCastingCost = isContainer ? EMPTY : Msgs.DEFAULT_CASTING_COST;
		mMaintenance = EMPTY;
		mCastingTime = isContainer ? EMPTY : Msgs.DEFAULT_CASTING_TIME;
		mDuration = isContainer ? EMPTY : Msgs.DEFAULT_DURATION;
		mPoints = 1;
		mReference = EMPTY;
		mIsVeryHard = false;
		mWeapons = new ArrayList<CMWeaponStats>();
		updateLevel(false);
	}

	/**
	 * Creates a clone of an existing spell and associates it with the specified data file.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param spell The spell to clone.
	 * @param deep Whether or not to clone the children, grandchildren, etc.
	 * @param forSheet Whether this is for a character sheet or a list.
	 */
	public CMSpell(CMDataFile dataFile, CMSpell spell, boolean deep, boolean forSheet) {
		super(dataFile, spell);
		mName = spell.mName;
		mTechLevel = spell.mTechLevel;
		mCollege = spell.mCollege;
		mSpellClass = spell.mSpellClass;
		mCastingCost = spell.mCastingCost;
		mMaintenance = spell.mMaintenance;
		mCastingTime = spell.mCastingTime;
		mDuration = spell.mDuration;
		mPoints = forSheet ? spell.mPoints : 1;
		mReference = spell.mReference;
		mIsVeryHard = spell.mIsVeryHard;
		if (forSheet && dataFile instanceof CMCharacter) {
			if (mTechLevel != null) {
				mTechLevel = ((CMCharacter) dataFile).getTechLevel();
			}
		} else {
			if (mTechLevel != null && mTechLevel.trim().length() > 0) {
				mTechLevel = EMPTY;
			}
		}
		mWeapons = new ArrayList<CMWeaponStats>(spell.mWeapons.size());
		for (CMWeaponStats weapon : spell.mWeapons) {
			if (weapon instanceof CMMeleeWeaponStats) {
				mWeapons.add(new CMMeleeWeaponStats(this, (CMMeleeWeaponStats) weapon));
			} else if (weapon instanceof CMRangedWeaponStats) {
				mWeapons.add(new CMRangedWeaponStats(this, (CMRangedWeaponStats) weapon));
			}
		}
		updateLevel(false);
		if (deep) {
			int count = spell.getChildCount();

			for (int i = 0; i < count; i++) {
				addChild(new CMSpell(dataFile, (CMSpell) spell.getChild(i), true, forSheet));
			}
		}
	}

	/**
	 * Loads a spell and associates it with the specified data file.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMSpell(CMDataFile dataFile, TKXMLReader reader) throws IOException {
		this(dataFile, TAG_SPELL_CONTAINER.equals(reader.getName()));
		load(reader, false);
	}

	@Override public String getLocalizedName() {
		return Msgs.DEFAULT_NAME;
	}

	@Override public String getListChangedID() {
		return ID_LIST_CHANGED;
	}

	@Override public String getXMLTagName() {
		return canHaveChildren() ? TAG_SPELL_CONTAINER : TAG_SPELL;
	}

	@Override public String getRowType() {
		return "Spell"; //$NON-NLS-1$
	}

	@Override protected void prepareForLoad(boolean forUndo) {
		boolean isContainer = canHaveChildren();

		super.prepareForLoad(forUndo);
		mName = Msgs.DEFAULT_NAME;
		mTechLevel = null;
		mCollege = EMPTY;
		mSpellClass = isContainer ? EMPTY : Msgs.DEFAULT_SPELL_CLASS;
		mCastingCost = isContainer ? EMPTY : Msgs.DEFAULT_CASTING_COST;
		mMaintenance = EMPTY;
		mCastingTime = isContainer ? EMPTY : Msgs.DEFAULT_CASTING_TIME;
		mDuration = isContainer ? EMPTY : Msgs.DEFAULT_DURATION;
		mPoints = 1;
		mReference = EMPTY;
		mIsVeryHard = false;
		mWeapons = new ArrayList<CMWeaponStats>();
	}

	@Override protected void loadAttributes(TKXMLReader reader, boolean forUndo) {
		mIsVeryHard = reader.isAttributeSet(ATTRIBUTE_VERY_HARD);
	}

	@Override protected void loadSubElement(TKXMLReader reader, boolean forUndo) throws IOException {
		String name = reader.getName();

		if (TAG_NAME.equals(name)) {
			mName = reader.readText().replace(NEWLINE, SPACE);
			// Fix for legacy format...
			if (mName.toLowerCase().endsWith("(vh)")) { //$NON-NLS-1$
				mName = mName.substring(0, mName.length() - 4).trim();
				mIsVeryHard = true;
			}
		} else if (TAG_TECH_LEVEL.equals(name)) {
			mTechLevel = reader.readText();
			if (mTechLevel != null && getDataFile() instanceof CMListFile) {
				mTechLevel = EMPTY;
			}
		} else if (TAG_REFERENCE.equals(name)) {
			mReference = reader.readText().replace(NEWLINE, SPACE);
		} else if (!forUndo && (TAG_SPELL.equals(name) || TAG_SPELL_CONTAINER.equals(name))) {
			addChild(new CMSpell(mDataFile, reader));
		} else if (!canHaveChildren()) {
			if (TAG_COLLEGE.equals(name)) {
				mCollege = reader.readText().replace(NEWLINE, SPACE).replace("/ ", "/"); //$NON-NLS-1$ //$NON-NLS-2$
			} else if (TAG_SPELL_CLASS.equals(name)) {
				mSpellClass = reader.readText().replace(NEWLINE, SPACE);
			} else if (TAG_CASTING_COST.equals(name)) {
				mCastingCost = reader.readText().replace(NEWLINE, SPACE);
			} else if (TAG_MAINTENANCE_COST.equals(name)) {
				mMaintenance = reader.readText().replace(NEWLINE, SPACE);
			} else if (TAG_CASTING_TIME.equals(name)) {
				mCastingTime = reader.readText().replace(NEWLINE, SPACE);
			} else if (TAG_DURATION.equals(name)) {
				mDuration = reader.readText().replace(NEWLINE, SPACE);
			} else if (TAG_POINTS.equals(name)) {
				mPoints = reader.readInteger(1);
			} else if (CMMeleeWeaponStats.TAG_ROOT.equals(name)) {
				mWeapons.add(new CMMeleeWeaponStats(this, reader));
			} else if (CMRangedWeaponStats.TAG_ROOT.equals(name)) {
				mWeapons.add(new CMRangedWeaponStats(this, reader));
			} else {
				super.loadSubElement(reader, forUndo);
			}
		} else {
			super.loadSubElement(reader, forUndo);
		}
	}

	@Override protected void finishedLoading() {
		updateLevel(false);
	}

	@Override protected void saveAttributes(TKXMLWriter out, boolean forUndo) {
		if (mIsVeryHard) {
			out.writeAttribute(ATTRIBUTE_VERY_HARD, mIsVeryHard);
		}
	}

	@Override public void saveSelf(TKXMLWriter out, boolean forUndo) {
		out.simpleTag(TAG_NAME, mName);
		if (!canHaveChildren()) {
			if (mTechLevel != null) {
				if (getCharacter() != null) {
					if (mTechLevel.length() > 0) {
						out.simpleTag(TAG_TECH_LEVEL, mTechLevel);
					}
				} else {
					out.startTag(TAG_TECH_LEVEL);
					out.finishEmptyTagEOL();
				}
			}
			out.simpleTag(TAG_COLLEGE, mCollege);
			out.simpleTag(TAG_SPELL_CLASS, mSpellClass);
			out.simpleTag(TAG_CASTING_COST, mCastingCost);
			out.simpleTag(TAG_MAINTENANCE_COST, mMaintenance);
			out.simpleTag(TAG_CASTING_TIME, mCastingTime);
			out.simpleTag(TAG_DURATION, mDuration);
			if (mPoints != 1) {
				out.simpleTag(TAG_POINTS, mPoints);
			}
			for (CMWeaponStats weapon : mWeapons) {
				weapon.save(out);
			}
		}
		out.simpleTag(TAG_REFERENCE, mReference);
	}

	/** @return The weapon list. */
	public List<CMWeaponStats> getWeapons() {
		return Collections.unmodifiableList(mWeapons);
	}

	/**
	 * @param weapons The weapons to set.
	 * @return Whether it was modified.
	 */
	public boolean setWeapons(List<CMWeaponStats> weapons) {
		if (!mWeapons.equals(weapons)) {
			mWeapons = new ArrayList<CMWeaponStats>(weapons);
			for (CMWeaponStats weapon : mWeapons) {
				weapon.setOwner(this);
			}
			notifySingle(ID_WEAPON_STATUS_CHANGED);
			return true;
		}
		return false;
	}

	/** @return The tech level. */
	public String getTechLevel() {
		return mTechLevel;
	}

	/**
	 * @param techLevel The tech level to set.
	 * @return Whether it was changed.
	 */
	public boolean setTechLevel(String techLevel) {
		if (mTechLevel == null ? techLevel != null : !mTechLevel.equals(techLevel)) {
			mTechLevel = techLevel;
			notifySingle(ID_TECH_LEVEL);
			return true;
		}
		return false;
	}

	/** @return The level. */
	public int getLevel() {
		return mLevel;
	}

	/** @return The relative level. */
	public int getRelativeLevel() {
		return mRelativeLevel;
	}

	/**
	 * Call to force an update of the level and relative level for this spell.
	 * 
	 * @param notify Whether or not a notification should be issued on a change.
	 */
	public void updateLevel(boolean notify) {
		int savedLevel = mLevel;
		int savedRelativeLevel = mRelativeLevel;
		CMSkillLevel level = calculateLevel(getCharacter(), mPoints, mIsVeryHard, mCollege, mName);

		mLevel = level.mLevel;
		mRelativeLevel = level.mRelativeLevel;

		if (notify && (savedLevel != mLevel || savedRelativeLevel != mRelativeLevel)) {
			notify(ID_LEVEL, this);
		}
	}

	/**
	 * Calculates the spell level.
	 * 
	 * @param character The character the spell will be attached to.
	 * @param points The number of points spent in the spell.
	 * @param isVeryHard Whether the spell is "Very Hard" or not.
	 * @param college The college the spell belongs to.
	 * @param name The name of the spell.
	 * @return The calculated spell level.
	 */
	public static CMSkillLevel calculateLevel(CMCharacter character, int points, boolean isVeryHard, String college, String name) {
		int relativeLevel = isVeryHard ? -3 : -2;
		int level;

		if (character != null) {
			level = character.getIntelligence();
			if (points < 1) {
				level = -1;
				relativeLevel = 0;
			} else if (points == 1) {
				// mRelativeLevel is preset to this point value
			} else if (points < 4) {
				relativeLevel++;
			} else {
				relativeLevel += 1 + points / 4;
			}

			if (level != -1) {
				level += relativeLevel + character.getIntegerBonusFor(ID_COLLEGE) + character.getIntegerBonusFor(ID_COLLEGE + "/" + college.toLowerCase()) + character.getSpellComparedIntegerBonusFor(ID_COLLEGE + "*", college) + character.getIntegerBonusFor(ID_NAME) + character.getIntegerBonusFor(ID_NAME + "/" + name.toLowerCase()) + character.getSpellComparedIntegerBonusFor(ID_NAME + "*", name); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
			}
		} else {
			level = -1;
		}

		return new CMSkillLevel(level, relativeLevel);
	}

	/** @return The name. */
	public String getName() {
		return mName;
	}

	/**
	 * @param name The name to set.
	 * @return Whether it was changed.
	 */
	public boolean setName(String name) {
		if (!mName.equals(name)) {
			mName = name;
			notifySingle(ID_NAME);
			return true;
		}
		return false;
	}

	/** @return The college. */
	public String getCollege() {
		return mCollege;
	}

	/**
	 * @param college The college to set.
	 * @return Whether it was changed.
	 */
	public boolean setCollege(String college) {
		if (!mCollege.equals(college)) {
			mCollege = college;
			notifySingle(ID_COLLEGE);
			return true;
		}
		return false;
	}

	/** @return The class. */
	public String getSpellClass() {
		return mSpellClass;
	}

	/**
	 * @param spellClass The class to set.
	 * @return Whether it was modified.
	 */
	public boolean setSpellClass(String spellClass) {
		if (!mSpellClass.equals(spellClass)) {
			mSpellClass = spellClass;
			return true;
		}
		return false;
	}

	/** @return The casting cost. */
	public String getCastingCost() {
		return mCastingCost;
	}

	/**
	 * @param cost The casting cost to set.
	 * @return Whether it was modified.
	 */
	public boolean setCastingCost(String cost) {
		if (!mCastingCost.equals(cost)) {
			mCastingCost = cost;
			return true;
		}
		return false;
	}

	/** @return The maintainance cost. */
	public String getMaintenance() {
		return mMaintenance;
	}

	/**
	 * @param cost The maintainance cost to set.
	 * @return Whether it was modified.
	 */
	public boolean setMaintenance(String cost) {
		if (!mMaintenance.equals(cost)) {
			mMaintenance = cost;
			return true;
		}
		return false;
	}

	/** @return The casting time. */
	public String getCastingTime() {
		return mCastingTime;
	}

	/**
	 * @param castingTime The casting time to set.
	 * @return Whether it was modified.
	 */
	public boolean setCastingTime(String castingTime) {
		if (!mCastingTime.equals(castingTime)) {
			mCastingTime = castingTime;
			return true;
		}
		return false;
	}

	/** @return The duration. */
	public String getDuration() {
		return mDuration;
	}

	/**
	 * @param duration The duration to set.
	 * @return Whether it was modified.
	 */
	public boolean setDuration(String duration) {
		if (!mDuration.equals(duration)) {
			mDuration = duration;
			return true;
		}
		return false;
	}

	/** @return The points. */
	public int getPoints() {
		return mPoints;
	}

	/**
	 * @param points The points to set.
	 * @return Whether it was modified.
	 */
	public boolean setPoints(int points) {
		if (mPoints != points) {
			mPoints = points;
			startNotify();
			notify(ID_POINTS, this);
			updateLevel(true);
			endNotify();
			return true;
		}
		return false;
	}

	/** @return The page reference. */
	public String getReference() {
		return mReference;
	}

	/**
	 * @param reference The page reference to set.
	 * @return Whether it was changed.
	 */
	public boolean setReference(String reference) {
		if (!mReference.equals(reference)) {
			mReference = reference;
			return true;
		}
		return false;
	}

	@Override public Object getData(TKColumn column) {
		return CSSpellColumnID.values()[column.getID()].getData(this);
	}

	@Override public String getDataAsText(TKColumn column) {
		return CSSpellColumnID.values()[column.getID()].getDataAsText(this);
	}

	@Override public boolean contains(String text, boolean lowerCaseOnly) {
		boolean contains = getName().toLowerCase().indexOf(text) != -1;

		if (!contains && canHaveChildren()) {
			contains = getCollege().toLowerCase().indexOf(text) != -1 || getSpellClass().toLowerCase().indexOf(text) != -1;
		}
		return contains;
	}

	@Override public String toString() {
		StringBuilder builder = new StringBuilder();

		builder.append(getName());
		if (!canHaveChildren()) {
			String techLevel = getTechLevel();

			if (techLevel != null) {
				builder.append("/TL"); //$NON-NLS-1$
				if (techLevel.length() > 0) {
					builder.append(techLevel);
				}
			}
		}
		return builder.toString();
	}

	@Override public BufferedImage getImage(boolean large) {
		return CSImage.getSpellIcon(large, true);
	}

	/** @return Whether this is a "Very Hard" spell or not. */
	public boolean isVeryHard() {
		return mIsVeryHard;
	}

	/**
	 * @param isVeryHard Whether this is a "Very Hard" spell or not.
	 * @return Whether it was modified.
	 */
	public boolean setIsVeryHard(boolean isVeryHard) {
		if (mIsVeryHard != isVeryHard) {
			mIsVeryHard = isVeryHard;
			startNotify();
			notify(ID_IS_VERY_HARD, this);
			updateLevel(true);
			endNotify();
			return true;
		}
		return false;
	}

	@Override public CSRowEditor<? extends CMRow> createEditor() {
		return new CSSpellEditor(this);
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		extractNameables(set, mName);
		extractNameables(set, mCollege);
		extractNameables(set, mCastingCost);
		extractNameables(set, mMaintenance);
		extractNameables(set, mCastingTime);
		extractNameables(set, mDuration);
		for (CMWeaponStats weapon : mWeapons) {
			for (CMSkillDefault one : weapon.getDefaults()) {
				one.fillWithNameableKeys(set);
			}
		}
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mName = nameNameables(map, mName);
		mCollege = nameNameables(map, mCollege);
		mSpellClass = nameNameables(map, mSpellClass);
		mCastingCost = nameNameables(map, mCastingCost);
		mMaintenance = nameNameables(map, mMaintenance);
		mCastingTime = nameNameables(map, mCastingTime);
		mDuration = nameNameables(map, mDuration);
		for (CMWeaponStats weapon : mWeapons) {
			for (CMSkillDefault one : weapon.getDefaults()) {
				one.applyNameableKeys(map);
			}
		}
	}

	/** @return The default casting cost. */
	public static final String getDefaultCastingCost() {
		return Msgs.DEFAULT_CASTING_COST;
	}

	/** @return The default casting time. */
	public static final String getDefaultCastingTime() {
		return Msgs.DEFAULT_CASTING_TIME;
	}

	/** @return The default duration. */
	public static final String getDefaultDuration() {
		return Msgs.DEFAULT_DURATION;
	}

	/** @return The default spell class. */
	public static final String getDefaultSpellClass() {
		return Msgs.DEFAULT_SPELL_CLASS;
	}
}
