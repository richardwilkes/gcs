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

package com.trollworks.gcs.model.advantage;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.modifier.CMModifier;
import com.trollworks.gcs.model.skill.CMSkillDefault;
import com.trollworks.gcs.model.weapon.CMMeleeWeaponStats;
import com.trollworks.gcs.model.weapon.CMOldWeapon;
import com.trollworks.gcs.model.weapon.CMRangedWeaponStats;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.gcs.ui.advantage.CSAdvantageColumnID;
import com.trollworks.gcs.ui.advantage.CSAdvantageEditor;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.toolkit.collections.TKEnumExtractor;
import com.trollworks.toolkit.collections.TKFilteredIterator;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.widget.outline.TKColumn;

import java.awt.image.BufferedImage;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

/** A GURPS (Dis)Advantage. */
public class CMAdvantage extends CMRow {
	/** The XML tag used for items. */
	public static final String			TAG_ADVANTAGE				= "advantage";									//$NON-NLS-1$
	/** The XML tag used for containers. */
	public static final String			TAG_ADVANTAGE_CONTAINER		= "advantage_container";						//$NON-NLS-1$
	private static final String			TAG_REFERENCE				= "reference";									//$NON-NLS-1$
	private static final String			TAG_OLD_POINTS				= "points";									//$NON-NLS-1$
	private static final String			TAG_BASE_POINTS				= "base_points";								//$NON-NLS-1$
	private static final String			TAG_POINTS_PER_LEVEL		= "points_per_level";							//$NON-NLS-1$
	private static final String			TAG_LEVELS					= "levels";									//$NON-NLS-1$
	private static final String			TAG_TYPE					= "type";										//$NON-NLS-1$
	private static final String			TAG_NAME					= "name";										//$NON-NLS-1$
	private static final String			TYPE_MENTAL					= "Mental";									//$NON-NLS-1$
	private static final String			TYPE_PHYSICAL				= "Physical";									//$NON-NLS-1$
	private static final String			TYPE_SOCIAL					= "Social";									//$NON-NLS-1$
	private static final String			TYPE_EXOTIC					= "Exotic";									//$NON-NLS-1$
	private static final String			TYPE_SUPERNATURAL			= "Supernatural";								//$NON-NLS-1$
	/** The prefix used in front of all IDs for the (dis)advantages. */
	public static final String			PREFIX						= CMCharacter.CHARACTER_PREFIX + "advantage.";	//$NON-NLS-1$
	/** The field ID for type changes. */
	public static final String			ID_TYPE						= PREFIX + "Type";								//$NON-NLS-1$
	/** The field ID for container type changes. */
	public static final String			ID_CONTAINER_TYPE			= PREFIX + "ContainerType";					//$NON-NLS-1$
	/** The field ID for name changes. */
	public static final String			ID_NAME						= PREFIX + "Name";								//$NON-NLS-1$
	/** The field ID for level changes. */
	public static final String			ID_LEVELS					= PREFIX + "Levels";							//$NON-NLS-1$
	/** The field ID for point changes. */
	public static final String			ID_POINTS					= PREFIX + "Points";							//$NON-NLS-1$
	/** The field ID for page reference changes. */
	public static final String			ID_REFERENCE				= PREFIX + "Reference";						//$NON-NLS-1$
	/** The field ID for when the row hierarchy changes. */
	public static final String			ID_LIST_CHANGED				= PREFIX + "ListChanged";						//$NON-NLS-1$
	/** The field ID for when the advantage becomes or stops being a weapon. */
	public static final String			ID_WEAPON_STATUS_CHANGED	= PREFIX + "WeaponStatus";						//$NON-NLS-1$
	/** The field ID for when the advantage gets Modifiers. */
	public static final String			ID_MODIFIER_STATUS_CHANGED	= PREFIX + "Modifier";							//$NON-NLS-1$
	/** The type mask for mental (dis)advantages. */
	public static final int				TYPE_MASK_MENTAL			= 1 << 0;
	/** The type mask for physical (dis)advantages. */
	public static final int				TYPE_MASK_PHYSICAL			= 1 << 1;
	/** The type mask for social (dis)advantages. */
	public static final int				TYPE_MASK_SOCIAL			= 1 << 2;
	/** The type mask for exotic (dis)advantages. */
	public static final int				TYPE_MASK_EXOTIC			= 1 << 3;
	/** The type mask for supernatural (dis)advantages. */
	public static final int				TYPE_MASK_SUPERNATURAL		= 1 << 4;
	private int							mType;
	private String						mName;
	private int							mLevels;
	private int							mPoints;
	private int							mPointsPerLevel;
	private String						mReference;
	private String						mOldPointsString;
	private CMAdvantageContainerType	mContainerType;
	private ArrayList<CMWeaponStats>	mWeapons;
	private ArrayList<CMModifier>		mModifiers;
	// For load-time conversion only
	private CMOldWeapon					mOldWeapon;

	/**
	 * Creates a new (dis)advantage.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param isContainer Whether or not this row allows children.
	 */
	public CMAdvantage(CMDataFile dataFile, boolean isContainer) {
		super(dataFile, isContainer);
		mType = TYPE_MASK_PHYSICAL;
		mName = Msgs.DEFAULT_NAME;
		mLevels = -1;
		mReference = ""; //$NON-NLS-1$
		mContainerType = CMAdvantageContainerType.GROUP;
		mWeapons = new ArrayList<CMWeaponStats>();
		mModifiers = new ArrayList<CMModifier>();
	}

	/**
	 * Creates a clone of an existing (dis)advantage and associates it with the specified data file.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param advantage The (dis)advantage to clone.
	 * @param deep Whether or not to clone the children, grandchildren, etc.
	 */
	public CMAdvantage(CMDataFile dataFile, CMAdvantage advantage, boolean deep) {
		super(dataFile, advantage);
		mType = advantage.mType;
		mName = advantage.mName;
		mLevels = advantage.mLevels;
		mPoints = advantage.mPoints;
		mPointsPerLevel = advantage.mPointsPerLevel;
		mReference = advantage.mReference;
		mContainerType = advantage.mContainerType;
		mWeapons = new ArrayList<CMWeaponStats>(advantage.mWeapons.size());
		for (CMWeaponStats weapon : advantage.mWeapons) {
			if (weapon instanceof CMMeleeWeaponStats) {
				mWeapons.add(new CMMeleeWeaponStats(this, (CMMeleeWeaponStats) weapon));
			} else if (weapon instanceof CMRangedWeaponStats) {
				mWeapons.add(new CMRangedWeaponStats(this, (CMRangedWeaponStats) weapon));
			}
		}
		mModifiers = new ArrayList<CMModifier>(advantage.mModifiers.size());
		for (CMModifier modifier : advantage.mModifiers) {
			mModifiers.add(new CMModifier(mDataFile, modifier));
		}
		if (deep) {
			int count = advantage.getChildCount();

			for (int i = 0; i < count; i++) {
				addChild(new CMAdvantage(dataFile, (CMAdvantage) advantage.getChild(i), true));
			}
		}
	}

	/**
	 * Loads an (dis)advantage and associates it with the specified data file.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMAdvantage(CMDataFile dataFile, TKXMLReader reader) throws IOException {
		this(dataFile, TAG_ADVANTAGE_CONTAINER.equals(reader.getName()));
		load(reader, false);
	}

	@Override public String getListChangedID() {
		return ID_LIST_CHANGED;
	}

	@Override public String getRowType() {
		return "(Dis)Advantage"; //$NON-NLS-1$
	}

	@Override protected void prepareForLoad(boolean forUndo) {
		super.prepareForLoad(forUndo);
		mType = TYPE_MASK_PHYSICAL;
		mName = Msgs.DEFAULT_NAME;
		mLevels = -1;
		mReference = ""; //$NON-NLS-1$
		mContainerType = CMAdvantageContainerType.GROUP;
		mPoints = 0;
		mPointsPerLevel = 0;
		mOldPointsString = null;
		mWeapons = new ArrayList<CMWeaponStats>();
		mModifiers = new ArrayList<CMModifier>();
	}

	@Override protected void loadAttributes(TKXMLReader reader, boolean forUndo) {
		super.loadAttributes(reader, forUndo);
		if (canHaveChildren()) {
			mContainerType = (CMAdvantageContainerType) TKEnumExtractor.extract(reader.getAttribute(TAG_TYPE), CMAdvantageContainerType.values(), CMAdvantageContainerType.GROUP);
		}
	}

	@Override protected void loadSubElement(TKXMLReader reader, boolean forUndo) throws IOException {
		String name = reader.getName();

		if (TAG_NAME.equals(name)) {
			mName = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
		} else if (TAG_REFERENCE.equals(name)) {
			mReference = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
		} else if (!forUndo && (TAG_ADVANTAGE.equals(name) || TAG_ADVANTAGE_CONTAINER.equals(name))) {
			addChild(new CMAdvantage(mDataFile, reader));
		} else if (CMModifier.TAG_MODIFIER.equals(name)) {
			mModifiers.add(new CMModifier(getDataFile(), reader));
		} else if (!canHaveChildren()) {
			if (TAG_TYPE.equals(name)) {
				mType = getTypeFromText(reader.readText());
			} else if (TAG_LEVELS.equals(name)) {
				mLevels = reader.readInteger(-1);
			} else if (TAG_OLD_POINTS.equals(name)) {
				mOldPointsString = reader.readText();
			} else if (TAG_BASE_POINTS.equals(name)) {
				mPoints = reader.readInteger(0);
			} else if (TAG_POINTS_PER_LEVEL.equals(name)) {
				mPointsPerLevel = reader.readInteger(0);
			} else if (CMMeleeWeaponStats.TAG_ROOT.equals(name)) {
				mWeapons.add(new CMMeleeWeaponStats(this, reader));
			} else if (CMRangedWeaponStats.TAG_ROOT.equals(name)) {
				mWeapons.add(new CMRangedWeaponStats(this, reader));
			} else if (CMOldWeapon.TAG_ROOT.equals(name)) {
				mOldWeapon = new CMOldWeapon(reader);
			} else {
				super.loadSubElement(reader, forUndo);
			}
		} else {
			super.loadSubElement(reader, forUndo);
		}
	}

	@Override protected void finishedLoading() {
		if (mOldPointsString != null) {
			// All this is here solely to support loading old data files
			int slash;

			mOldPointsString = mOldPointsString.trim();
			slash = mOldPointsString.indexOf('/');
			if (slash == -1) {
				mPoints = getSimpleNumber(mOldPointsString);
				mPointsPerLevel = 0;
			} else {
				mPoints = getSimpleNumber(mOldPointsString.substring(0, slash));
				if (mOldPointsString.length() > slash) {
					mPointsPerLevel = getSimpleNumber(mOldPointsString.substring(slash + 1));
				} else {
					mPointsPerLevel = 0;
				}
				if (mPoints == 0) {
					mPoints = mPointsPerLevel;
					mPointsPerLevel = 0;
				}
			}
			if (mLevels >= 0 && mPointsPerLevel == 0) {
				mPointsPerLevel = mPoints;
				mPoints = 0;
			}
			mOldPointsString = null;
		}
		if (mOldWeapon != null) {
			mWeapons.addAll(mOldWeapon.getWeapons(this));
			mOldWeapon = null;
		}
		// We no longer have defaults... that was solely for the weapons
		setDefaults(new ArrayList<CMSkillDefault>());
		super.finishedLoading();
	}

	@Override public String getXMLTagName() {
		return canHaveChildren() ? TAG_ADVANTAGE_CONTAINER : TAG_ADVANTAGE;
	}

	@Override protected void saveAttributes(TKXMLWriter out, boolean forUndo) {
		super.saveAttributes(out, forUndo);
		if (canHaveChildren() && mContainerType != CMAdvantageContainerType.GROUP) {
			out.writeAttribute(TAG_TYPE, mContainerType.name().toLowerCase());
		}
	}

	@Override public void saveSelf(TKXMLWriter out, boolean forUndo) {
		out.simpleTag(TAG_NAME, mName);
		if (!canHaveChildren()) {
			out.simpleTag(TAG_TYPE, getTypeAsText());
			if (mLevels != -1) {
				out.simpleTag(TAG_LEVELS, mLevels);
			}
			if (mPoints != 0) {
				out.simpleTag(TAG_BASE_POINTS, mPoints);
			}
			if (mPointsPerLevel != 0) {
				out.simpleTag(TAG_POINTS_PER_LEVEL, mPointsPerLevel);
			}

			for (CMWeaponStats weapon : mWeapons) {
				weapon.save(out);
			}
		}
		for (CMModifier modifier : mModifiers) {
			modifier.save(out, forUndo);
		}
		out.simpleTagNotEmpty(TAG_REFERENCE, mReference);
	}

	/** @return The container type. */
	public CMAdvantageContainerType getContainerType() {
		return mContainerType;
	}

	/**
	 * @param type The container type to set.
	 * @return Whether it was modified.
	 */
	public boolean setContainerType(CMAdvantageContainerType type) {
		if (mContainerType != type) {
			mContainerType = type;
			notifySingle(ID_CONTAINER_TYPE);
			return true;
		}
		return false;
	}

	/** @return The type. */
	public int getType() {
		return mType;
	}

	/**
	 * @param type The type to set.
	 * @return Whether it was modified.
	 */
	public boolean setType(int type) {
		if (mType != type) {
			mType = type;
			notifySingle(ID_TYPE);
			return true;
		}
		return false;
	}

	@Override public String getLocalizedName() {
		return Msgs.DEFAULT_NAME;
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

	/** @return Whether this advantage is leveled or not. */
	public boolean isLeveled() {
		return mLevels >= 0;
	}

	/** @return The levels. */
	public int getLevels() {
		return mLevels;
	}

	/**
	 * @param levels The levels to set.
	 * @return Whether it was modified.
	 */
	public boolean setLevels(int levels) {
		if (mLevels != levels) {
			mLevels = levels;
			notifySingle(ID_LEVELS);
			return true;
		}
		return false;
	}

	/** @return The total points, taking levels into account. */
	public int getAdjustedPoints() {
		if (canHaveChildren()) {
			int points = 0;
			for (CMAdvantage child : new TKFilteredIterator<CMAdvantage>(getChildren(), CMAdvantage.class)) {
				points += child.getAdjustedPoints();
			}
			return points;
		}
		return getAdjustedPoints(mPoints, mLevels, mPointsPerLevel, getAllModifiers());
	}

	/**
	 * @param basePoints The base point cost.
	 * @param levels The number of levels.
	 * @param pointsPerLevel The point cost per level.
	 * @param modifiers The {@link CMModifier}s to apply.
	 * @return The total points, taking levels and modifiers into account.
	 */
	public static int getAdjustedPoints(int basePoints, int levels, int pointsPerLevel, Collection<CMModifier> modifiers) {
		int baseMod = 0;
		int levelMod = 0;
		double multiplier = 1.0;

		for (CMModifier one : modifiers) {
			if (one.isEnabled()) {
				int modifier = one.getCostModifier();
				switch (one.getCostType()) {
					case PERCENTAGE:
						switch (one.getAffects()) {
							case TOTAL:
								baseMod += modifier;
								levelMod += modifier;
								break;
							case BASE_ONLY:
								baseMod += modifier;
								break;
							case LEVELS_ONLY:
								levelMod += modifier;
								break;
						}
						break;
					case POINTS:
						switch (one.getAffects()) {
							case TOTAL:
							case BASE_ONLY:
								basePoints += modifier;
								break;
							case LEVELS_ONLY:
								pointsPerLevel += modifier;
								break;
						}
						break;
					case MULTIPLIER:
						multiplier *= one.getCostMultiplier();
						break;
				}
			}
		}

		int leveledPoints = levels > 0 ? pointsPerLevel * levels : 0;
		if (baseMod != 0 || levelMod != 0) {
			if (baseMod < -80) {
				baseMod = -80;
			}
			if (levelMod < -80) {
				levelMod = -80;
			}
			if (baseMod == levelMod) {
				basePoints = modifyPoints(basePoints + leveledPoints, baseMod);
			} else {
				basePoints = modifyPoints(basePoints, baseMod) + modifyPoints(leveledPoints, levelMod);
			}
		} else {
			basePoints += leveledPoints;
		}

		if (basePoints > 0) {
			basePoints = (int) (basePoints * multiplier + 0.5);
			if (basePoints < 1) {
				basePoints = 1;
			}
		} else if (basePoints < 0) {
			basePoints = (int) (basePoints * multiplier);
		}
		return basePoints;
	}

	private static int modifyPoints(int points, int modifier) {
		modifier *= points;
		if (modifier > 0) {
			modifier = (modifier + 50) / 100;
		} else {
			modifier /= 100;
		}
		return points + modifier;
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
			notifySingle(ID_POINTS);
			return true;
		}
		return false;
	}

	/** @return The points per level. */
	public int getPointsPerLevel() {
		return mPointsPerLevel;
	}

	/**
	 * @param points The points per level to set.
	 * @return Whether it was modified.
	 */
	public boolean setPointsPerLevel(int points) {
		if (mPointsPerLevel != points) {
			mPointsPerLevel = points;
			notifySingle(ID_POINTS);
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
			notifySingle(ID_REFERENCE);
			return true;
		}
		return false;
	}

	@Override public Object getData(TKColumn column) {
		return CSAdvantageColumnID.values()[column.getID()].getData(this);
	}

	@Override public String getDataAsText(TKColumn column) {
		return CSAdvantageColumnID.values()[column.getID()].getDataAsText(this);
	}

	private int getSimpleNumber(String buffer) {
		try {
			return Integer.parseInt(buffer);
		} catch (Exception exception) {
			return 0;
		}
	}

	@Override public boolean contains(String text, boolean lowerCaseOnly) {
		return getName().toLowerCase().indexOf(text) != -1;
	}

	private int getTypeFromText(String text) {
		int type = 0;

		if (text.indexOf(TYPE_MENTAL) != -1) {
			type |= TYPE_MASK_MENTAL;
		}
		if (text.indexOf(TYPE_PHYSICAL) != -1) {
			type |= TYPE_MASK_PHYSICAL;
		}
		if (text.indexOf(TYPE_SOCIAL) != -1) {
			type |= TYPE_MASK_SOCIAL;
		}
		if (text.indexOf(TYPE_EXOTIC) != -1) {
			type |= TYPE_MASK_EXOTIC;
		}
		if (text.indexOf(TYPE_SUPERNATURAL) != -1) {
			type |= TYPE_MASK_SUPERNATURAL;
		}
		return type;
	}

	/** @return The type as a text string. */
	public String getTypeAsText() {
		if (!canHaveChildren()) {
			String separator = ", "; //$NON-NLS-1$
			StringBuilder buffer = new StringBuilder();
			int type = getType();

			if ((type & CMAdvantage.TYPE_MASK_MENTAL) != 0) {
				buffer.append(TYPE_MENTAL);
			}
			if ((type & CMAdvantage.TYPE_MASK_PHYSICAL) != 0) {
				if (buffer.length() > 0) {
					buffer.append("/"); //$NON-NLS-1$
				}
				buffer.append(TYPE_PHYSICAL);
			}
			if ((type & CMAdvantage.TYPE_MASK_SOCIAL) != 0) {
				if (buffer.length() > 0) {
					buffer.append(separator);
				}
				buffer.append(TYPE_SOCIAL);
			}
			if ((type & CMAdvantage.TYPE_MASK_EXOTIC) != 0) {
				if (buffer.length() > 0) {
					buffer.append(separator);
				}
				buffer.append(TYPE_EXOTIC);
			}
			if ((type & CMAdvantage.TYPE_MASK_SUPERNATURAL) != 0) {
				if (buffer.length() > 0) {
					buffer.append(separator);
				}
				buffer.append(TYPE_SUPERNATURAL);
			}
			return buffer.toString();
		}
		return ""; //$NON-NLS-1$
	}

	@Override public String toString() {
		StringBuilder builder = new StringBuilder();

		builder.append(getName());
		if (!canHaveChildren()) {
			int levels = getLevels();

			if (levels > 0) {
				builder.append(' ');
				builder.append(levels);
			}
		}
		return builder.toString();
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

	@Override public BufferedImage getImage(boolean large) {
		return CSImage.getAdvantageIcon(large, true);
	}

	@Override public CSRowEditor<? extends CMRow> createEditor() {
		return new CSAdvantageEditor(this);
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		extractNameables(set, mName);
		for (CMWeaponStats weapon : mWeapons) {
			for (CMSkillDefault one : weapon.getDefaults()) {
				one.fillWithNameableKeys(set);
			}
		}
		for (CMModifier modifier : mModifiers) {
			modifier.fillWithNameableKeys(set);
		}
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mName = nameNameables(map, mName);
		for (CMWeaponStats weapon : mWeapons) {
			for (CMSkillDefault one : weapon.getDefaults()) {
				one.applyNameableKeys(map);
			}
		}
		for (CMModifier modifier : mModifiers) {
			modifier.applyNameableKeys(map);
		}
	}

	/** @return The modifiers. */
	public List<CMModifier> getModifiers() {
		return Collections.unmodifiableList(mModifiers);
	}

	/** @return The modifiers including those inherited from parent row. */
	public List<CMModifier> getAllModifiers() {
		ArrayList<CMModifier> allModifiers = new ArrayList<CMModifier>(mModifiers);
		if (getParent() != null) {
			allModifiers.addAll(((CMAdvantage) getParent()).getAllModifiers());
		}
		return Collections.unmodifiableList(allModifiers);
	}

	/**
	 * @param modifiers The value to set for modifiers.
	 * @return {@code true} if modifiers changed
	 */
	public boolean setModifiers(List<CMModifier> modifiers) {
		if (!mModifiers.equals(modifiers)) {
			mModifiers = new ArrayList<CMModifier>(modifiers);
			notifySingle(ID_MODIFIER_STATUS_CHANGED);
			return true;
		}
		return false;
	}

	@Override public String getModifierNotes() {
		ArrayList<CMModifier> modifiers = new ArrayList<CMModifier>();

		for (CMModifier modifier : mModifiers) {
			if (modifier.isEnabled()) {
				modifiers.add(modifier);
			}
		}
		if (!modifiers.isEmpty()) {
			StringBuilder builder = new StringBuilder();

			for (CMModifier modifier : modifiers) {
				builder.append(modifier.getFullDescription());
				builder.append("; "); //$NON-NLS-1$
			}
			builder.setLength(builder.length() - 2); // Remove the trailing "; "
			builder.append('.');
			return builder.toString();
		}
		return ""; //$NON-NLS-1$
	}
}
