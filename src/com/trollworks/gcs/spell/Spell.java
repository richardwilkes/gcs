/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.spell;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.utility.Localization;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

/** A GURPS Spell. */
public class Spell extends ListRow {
	@Localize("Spell")
	@Localize(locale = "de", value = "Zauber")
	@Localize(locale = "ru", value = "Заклинание")
	private static String			DEFAULT_NAME;
	@Localize("Arcane")
	@Localize(locale = "de", value = "Arkan")
	@Localize(locale = "ru", value = "Тайный")
	private static String			DEFAULT_POWER_SOURCE;
	@Localize("Regular")
	@Localize(locale = "de", value = "Regulär")
	@Localize(locale = "ru", value = "Обычный")
	private static String			DEFAULT_SPELL_CLASS;
	@Localize("1")
	@Localize(locale = "de", value = "1")
	private static String			DEFAULT_CASTING_COST;
	@Localize("1 sec")
	@Localize(locale = "de", value = "1 Sek.")
	@Localize(locale = "ru", value = "1 сек")
	private static String			DEFAULT_CASTING_TIME;
	@Localize("Instant")
	@Localize(locale = "de", value = "Sofort")
	@Localize(locale = "ru", value = "Мгновенное")
	private static String			DEFAULT_DURATION;

	static {
		Localization.initialize();
	}

	private static final int		CURRENT_VERSION				= 2;
	/** The extension for Spell lists. */
	public static final String		OLD_SPELL_EXTENSION			= "spl";										//$NON-NLS-1$
	/** The XML tag used for items. */
	public static final String		TAG_SPELL					= "spell";										//$NON-NLS-1$
	/** The XML tag used for containers. */
	public static final String		TAG_SPELL_CONTAINER			= "spell_container";							//$NON-NLS-1$
	private static final String		TAG_NAME					= "name";										//$NON-NLS-1$
	private static final String		TAG_TECH_LEVEL				= "tech_level";								//$NON-NLS-1$
	private static final String		TAG_COLLEGE					= "college";									//$NON-NLS-1$
	private static final String		TAG_POWER_SOURCE			= "power_source";								//$NON-NLS-1$
	private static final String		TAG_SPELL_CLASS				= "spell_class";								//$NON-NLS-1$
	private static final String		TAG_CASTING_COST			= "casting_cost";								//$NON-NLS-1$
	private static final String		TAG_MAINTENANCE_COST		= "maintenance_cost";							//$NON-NLS-1$
	private static final String		TAG_CASTING_TIME			= "casting_time";								//$NON-NLS-1$
	private static final String		TAG_DURATION				= "duration";									//$NON-NLS-1$
	private static final String		TAG_POINTS					= "points";									//$NON-NLS-1$
	private static final String		TAG_REFERENCE				= "reference";									//$NON-NLS-1$
	private static final String		ATTRIBUTE_VERY_HARD			= "very_hard";									//$NON-NLS-1$
	/** The prefix used in front of all IDs for the spells. */
	public static final String		PREFIX						= GURPSCharacter.CHARACTER_PREFIX + "spell.";	//$NON-NLS-1$
	/** The field ID for name changes. */
	public static final String		ID_NAME						= PREFIX + "Name";								//$NON-NLS-1$
	/** The field ID for tech level changes. */
	public static final String		ID_TECH_LEVEL				= PREFIX + "TechLevel";						//$NON-NLS-1$
	/** The field ID for college changes. */
	public static final String		ID_COLLEGE					= PREFIX + "College";							//$NON-NLS-1$
	/** The field ID for power source changes. */
	public static final String		ID_POWER_SOURCE				= PREFIX + "PowerSource";						//$NON-NLS-1$
	/** The field ID for spell class changes. */
	public static final String		ID_SPELL_CLASS				= PREFIX + "Class";							//$NON-NLS-1$
	/** The field ID for casting cost changes */
	public static final String		ID_CASTING_COST				= PREFIX + "CastingCost";						//$NON-NLS-1$
	/** The field ID for maintainance cost changes */
	public static final String		ID_MAINTENANCE_COST			= PREFIX + "MaintenanceCost";					//$NON-NLS-1$
	/** The field ID for casting time changes */
	public static final String		ID_CASTING_TIME				= PREFIX + "CastingTime";						//$NON-NLS-1$
	/** The field ID for duration changes */
	public static final String		ID_DURATION					= PREFIX + "Duration";							//$NON-NLS-1$
	/** The field ID for point changes. */
	public static final String		ID_POINTS					= PREFIX + "Points";							//$NON-NLS-1$
	/** The field ID for level changes. */
	public static final String		ID_LEVEL					= PREFIX + "Level";							//$NON-NLS-1$
	/** The field ID for page reference changes. */
	public static final String		ID_REFERENCE				= PREFIX + "Reference";						//$NON-NLS-1$
	/** The field ID for difficulty changes. */
	public static final String		ID_IS_VERY_HARD				= PREFIX + "Difficulty";						//$NON-NLS-1$
	/** The field ID for when the categories change. */
	public static final String		ID_CATEGORY					= PREFIX + "Category";							//$NON-NLS-1$
	/** The field ID for when the row hierarchy changes. */
	public static final String		ID_LIST_CHANGED				= PREFIX + "ListChanged";						//$NON-NLS-1$
	/** The field ID for when the spell becomes or stops being a weapon. */
	public static final String		ID_WEAPON_STATUS_CHANGED	= PREFIX + "WeaponStatus";						//$NON-NLS-1$
	private static final String		EMPTY						= "";											//$NON-NLS-1$
	private static final String		NEWLINE						= "\n";										//$NON-NLS-1$
	private static final String		SPACE						= " ";											//$NON-NLS-1$
	private String					mName;
	private String					mTechLevel;
	private String					mCollege;
	private String					mPowerSource;
	private String					mSpellClass;
	private String					mCastingCost;
	private String					mMaintenance;
	private String					mCastingTime;
	private String					mDuration;
	private int						mPoints;
	private int						mLevel;
	private int						mRelativeLevel;
	private String					mReference;
	private boolean					mIsVeryHard;
	private ArrayList<WeaponStats>	mWeapons;

	/**
	 * Creates a new spell.
	 *
	 * @param dataFile The data file to associate it with.
	 * @param isContainer Whether or not this row allows children.
	 */
	public Spell(DataFile dataFile, boolean isContainer) {
		super(dataFile, isContainer);
		mName = DEFAULT_NAME;
		mTechLevel = null;
		mCollege = EMPTY;
		mPowerSource = isContainer ? EMPTY : DEFAULT_POWER_SOURCE;
		mSpellClass = isContainer ? EMPTY : DEFAULT_SPELL_CLASS;
		mCastingCost = isContainer ? EMPTY : DEFAULT_CASTING_COST;
		mMaintenance = EMPTY;
		mCastingTime = isContainer ? EMPTY : DEFAULT_CASTING_TIME;
		mDuration = isContainer ? EMPTY : DEFAULT_DURATION;
		mPoints = 1;
		mReference = EMPTY;
		mIsVeryHard = false;
		mWeapons = new ArrayList<>();
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
	public Spell(DataFile dataFile, Spell spell, boolean deep, boolean forSheet) {
		super(dataFile, spell);
		mName = spell.mName;
		mTechLevel = spell.mTechLevel;
		mCollege = spell.mCollege;
		mPowerSource = spell.mPowerSource;
		mSpellClass = spell.mSpellClass;
		mCastingCost = spell.mCastingCost;
		mMaintenance = spell.mMaintenance;
		mCastingTime = spell.mCastingTime;
		mDuration = spell.mDuration;
		mPoints = forSheet ? spell.mPoints : 1;
		mReference = spell.mReference;
		mIsVeryHard = spell.mIsVeryHard;
		if (forSheet && dataFile instanceof GURPSCharacter) {
			if (mTechLevel != null) {
				mTechLevel = ((GURPSCharacter) dataFile).getDescription().getTechLevel();
			}
		} else {
			if (mTechLevel != null && mTechLevel.trim().length() > 0) {
				mTechLevel = EMPTY;
			}
		}
		mWeapons = new ArrayList<>(spell.mWeapons.size());
		for (WeaponStats weapon : spell.mWeapons) {
			if (weapon instanceof MeleeWeaponStats) {
				mWeapons.add(new MeleeWeaponStats(this, (MeleeWeaponStats) weapon));
			} else if (weapon instanceof RangedWeaponStats) {
				mWeapons.add(new RangedWeaponStats(this, (RangedWeaponStats) weapon));
			}
		}
		updateLevel(false);
		if (deep) {
			int count = spell.getChildCount();

			for (int i = 0; i < count; i++) {
				addChild(new Spell(dataFile, (Spell) spell.getChild(i), true, forSheet));
			}
		}
	}

	/**
	 * Loads a spell and associates it with the specified data file.
	 *
	 * @param dataFile The data file to associate it with.
	 * @param reader The XML reader to load from.
	 * @param state The {@link LoadState} to use.
	 */
	public Spell(DataFile dataFile, XMLReader reader, LoadState state) throws IOException {
		this(dataFile, TAG_SPELL_CONTAINER.equals(reader.getName()));
		load(reader, state);
	}

	@Override
	public boolean isEquivalentTo(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof Spell && super.isEquivalentTo(obj)) {
			Spell row = (Spell) obj;
			if (mIsVeryHard == row.mIsVeryHard && mPoints == row.mPoints && mLevel == row.mLevel && mRelativeLevel == row.mRelativeLevel) {
				if (mTechLevel == null ? row.mTechLevel == null : mTechLevel.equals(row.mTechLevel)) {
					if (mName.equals(row.mName) && mCollege.equals(row.mCollege) && mPowerSource.equals(row.mPowerSource) && mSpellClass.equals(row.mSpellClass) && mReference.equals(row.mReference)) {
						if (mCastingCost.equals(row.mCastingCost) && mMaintenance.equals(row.mMaintenance) && mCastingTime.equals(row.mCastingTime) && mDuration.equals(row.mDuration)) {
							return mWeapons.equals(row.mWeapons);
						}
					}
				}
			}
		}
		return false;
	}

	@Override
	public String getLocalizedName() {
		return DEFAULT_NAME;
	}

	@Override
	public String getListChangedID() {
		return ID_LIST_CHANGED;
	}

	@Override
	public String getXMLTagName() {
		return canHaveChildren() ? TAG_SPELL_CONTAINER : TAG_SPELL;
	}

	@Override
	public int getXMLTagVersion() {
		return CURRENT_VERSION;
	}

	@Override
	public String getRowType() {
		return DEFAULT_NAME;
	}

	@Override
	protected void prepareForLoad(LoadState state) {
		boolean isContainer = canHaveChildren();
		super.prepareForLoad(state);
		mName = DEFAULT_NAME;
		mTechLevel = null;
		mCollege = EMPTY;
		mPowerSource = isContainer ? EMPTY : DEFAULT_POWER_SOURCE;
		mSpellClass = isContainer ? EMPTY : DEFAULT_SPELL_CLASS;
		mCastingCost = isContainer ? EMPTY : DEFAULT_CASTING_COST;
		mMaintenance = EMPTY;
		mCastingTime = isContainer ? EMPTY : DEFAULT_CASTING_TIME;
		mDuration = isContainer ? EMPTY : DEFAULT_DURATION;
		mPoints = 1;
		mReference = EMPTY;
		mIsVeryHard = false;
		mWeapons = new ArrayList<>();
	}

	@Override
	protected void loadAttributes(XMLReader reader, LoadState state) {
		super.loadAttributes(reader, state);
		mIsVeryHard = reader.isAttributeSet(ATTRIBUTE_VERY_HARD);
	}

	@Override
	protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
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
			if (mTechLevel != null) {
				DataFile dataFile = getDataFile();
				if (dataFile instanceof ListFile || dataFile instanceof LibraryFile) {
					mTechLevel = EMPTY;
				}
			}
		} else if (TAG_REFERENCE.equals(name)) {
			mReference = reader.readText().replace(NEWLINE, SPACE);
		} else if (!state.mForUndo && (TAG_SPELL.equals(name) || TAG_SPELL_CONTAINER.equals(name))) {
			addChild(new Spell(mDataFile, reader, state));
		} else if (!canHaveChildren()) {
			if (TAG_COLLEGE.equals(name)) {
				mCollege = reader.readText().replace(NEWLINE, SPACE).replace("/ ", "/"); //$NON-NLS-1$ //$NON-NLS-2$
			} else if (TAG_POWER_SOURCE.equals(name)) {
				mPowerSource = reader.readText().replace(NEWLINE, SPACE);
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
			} else if (MeleeWeaponStats.TAG_ROOT.equals(name)) {
				mWeapons.add(new MeleeWeaponStats(this, reader));
			} else if (RangedWeaponStats.TAG_ROOT.equals(name)) {
				mWeapons.add(new RangedWeaponStats(this, reader));
			} else {
				super.loadSubElement(reader, state);
			}
		} else {
			super.loadSubElement(reader, state);
		}
	}

	@Override
	protected void finishedLoading(LoadState state) {
		updateLevel(false);
		super.finishedLoading(state);
	}

	@Override
	protected void saveAttributes(XMLWriter out, boolean forUndo) {
		if (mIsVeryHard) {
			out.writeAttribute(ATTRIBUTE_VERY_HARD, mIsVeryHard);
		}
	}

	@Override
	public void saveSelf(XMLWriter out, boolean forUndo) {
		out.simpleTag(TAG_NAME, mName);
		if (!canHaveChildren()) {
			if (mTechLevel != null) {
				if (getCharacter() != null) {
					out.simpleTagNotEmpty(TAG_TECH_LEVEL, mTechLevel);
				} else {
					out.startTag(TAG_TECH_LEVEL);
					out.finishEmptyTagEOL();
				}
			}
			out.simpleTagNotEmpty(TAG_COLLEGE, mCollege);
			out.simpleTagNotEmpty(TAG_POWER_SOURCE, mPowerSource);
			out.simpleTagNotEmpty(TAG_SPELL_CLASS, mSpellClass);
			out.simpleTagNotEmpty(TAG_CASTING_COST, mCastingCost);
			out.simpleTagNotEmpty(TAG_MAINTENANCE_COST, mMaintenance);
			out.simpleTagNotEmpty(TAG_CASTING_TIME, mCastingTime);
			out.simpleTagNotEmpty(TAG_DURATION, mDuration);
			if (mPoints != 1) {
				out.simpleTag(TAG_POINTS, mPoints);
			}
			for (WeaponStats weapon : mWeapons) {
				weapon.save(out);
			}
		}
		out.simpleTagNotEmpty(TAG_REFERENCE, mReference);
	}

	/** @return The weapon list. */
	public List<WeaponStats> getWeapons() {
		return Collections.unmodifiableList(mWeapons);
	}

	/**
	 * @param weapons The weapons to set.
	 * @return Whether it was modified.
	 */
	public boolean setWeapons(List<WeaponStats> weapons) {
		if (!mWeapons.equals(weapons)) {
			mWeapons = new ArrayList<>(weapons);
			for (WeaponStats weapon : mWeapons) {
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
		SkillLevel level = calculateLevel(getCharacter(), mPoints, mIsVeryHard, mCollege, mPowerSource, mName);

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
	 * @param powerSource The source of power for the spell.
	 * @param name The name of the spell.
	 * @return The calculated spell level.
	 */
	public static SkillLevel calculateLevel(GURPSCharacter character, int points, boolean isVeryHard, String college, String powerSource, String name) {
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
				relativeLevel += getSpellBonusesFor(character, ID_COLLEGE, college);
				relativeLevel += getSpellBonusesFor(character, ID_POWER_SOURCE, powerSource);
				relativeLevel += getSpellBonusesFor(character, ID_NAME, name);
				level += relativeLevel;
			}
		} else {
			level = -1;
		}

		return new SkillLevel(level, relativeLevel);
	}

	private static int getSpellBonusesFor(GURPSCharacter character, String id, String qualifier) {
		int level = character.getIntegerBonusFor(id);
		level += character.getIntegerBonusFor(id + '/' + qualifier.toLowerCase());
		level += character.getSpellComparedIntegerBonusFor(id + '*', qualifier);
		return level;
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

	/** @return The power source. */
	public String getPowerSource() {
		return mPowerSource;
	}

	/**
	 * @param powerSource The college to set.
	 * @return Whether it was changed.
	 */
	public boolean setPowerSource(String powerSource) {
		if (!mPowerSource.equals(powerSource)) {
			mPowerSource = powerSource;
			notifySingle(ID_POWER_SOURCE);
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

	@Override
	public Object getData(Column column) {
		return SpellColumn.values()[column.getID()].getData(this);
	}

	@Override
	public String getDataAsText(Column column) {
		return SpellColumn.values()[column.getID()].getDataAsText(this);
	}

	@Override
	public boolean contains(String text, boolean lowerCaseOnly) {
		if (getName().toLowerCase().indexOf(text) != -1) {
			return true;
		}
		if (getCollege().toLowerCase().indexOf(text) != -1) {
			return true;
		}
		if (getSpellClass().toLowerCase().indexOf(text) != -1) {
			return true;
		}
		return super.contains(text, lowerCaseOnly);
	}

	@Override
	public String toString() {
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

	@Override
	public StdImage getIcon(boolean large) {
		return GCSImages.getSpellsIcons().getImage(large ? 64 : 16);
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

	@Override
	public RowEditor<? extends ListRow> createEditor() {
		return new SpellEditor(this);
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		extractNameables(set, mName);
		extractNameables(set, mCollege);
		extractNameables(set, mPowerSource);
		extractNameables(set, mCastingCost);
		extractNameables(set, mMaintenance);
		extractNameables(set, mCastingTime);
		extractNameables(set, mDuration);
		for (WeaponStats weapon : mWeapons) {
			for (SkillDefault one : weapon.getDefaults()) {
				one.fillWithNameableKeys(set);
			}
		}
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mName = nameNameables(map, mName);
		mCollege = nameNameables(map, mCollege);
		mPowerSource = nameNameables(map, mPowerSource);
		mSpellClass = nameNameables(map, mSpellClass);
		mCastingCost = nameNameables(map, mCastingCost);
		mMaintenance = nameNameables(map, mMaintenance);
		mCastingTime = nameNameables(map, mCastingTime);
		mDuration = nameNameables(map, mDuration);
		for (WeaponStats weapon : mWeapons) {
			for (SkillDefault one : weapon.getDefaults()) {
				one.applyNameableKeys(map);
			}
		}
	}

	/** @return The default casting cost. */
	public static final String getDefaultCastingCost() {
		return DEFAULT_CASTING_COST;
	}

	/** @return The default casting time. */
	public static final String getDefaultCastingTime() {
		return DEFAULT_CASTING_TIME;
	}

	/** @return The default duration. */
	public static final String getDefaultDuration() {
		return DEFAULT_DURATION;
	}

	/** @return The default spell class. */
	public static final String getDefaultSpellClass() {
		return DEFAULT_SPELL_CLASS;
	}

	@Override
	protected String getCategoryID() {
		return ID_CATEGORY;
	}
}
