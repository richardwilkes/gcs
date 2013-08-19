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

package com.trollworks.gcs.model.skill;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.skills.CSSkillColumnID;
import com.trollworks.gcs.ui.skills.CSSkillEditor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKRow;

import java.awt.image.BufferedImage;
import java.io.IOException;
import java.util.Collection;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

/** A GURPS Skill. */
public class CMSkill extends CMRow {
	/** The XML tag used for items. */
	public static final String	TAG_SKILL			= "skill";									//$NON-NLS-1$
	/** The XML tag used for containers. */
	public static final String	TAG_SKILL_CONTAINER	= "skill_container";						//$NON-NLS-1$
	private static final String	TAG_NAME			= "name";									//$NON-NLS-1$
	private static final String	TAG_SPECIALIZATION	= "specialization";						//$NON-NLS-1$
	private static final String	TAG_TECH_LEVEL		= "tech_level";							//$NON-NLS-1$
	private static final String	TAG_DIFFICULTY		= "difficulty";							//$NON-NLS-1$
	private static final String	TAG_POINTS			= "points";								//$NON-NLS-1$
	private static final String	TAG_REFERENCE		= "reference";								//$NON-NLS-1$
	/** The prefix used in front of all IDs for the skills. */
	public static final String	PREFIX				= CMCharacter.CHARACTER_PREFIX + "skill.";	//$NON-NLS-1$
	/** The field ID for name changes. */
	public static final String	ID_NAME				= PREFIX + "Name";							//$NON-NLS-1$
	/** The field ID for specialization changes. */
	public static final String	ID_SPECIALIZATION	= PREFIX + "Specialization";				//$NON-NLS-1$
	/** The field ID for tech level changes. */
	public static final String	ID_TECH_LEVEL		= PREFIX + "TechLevel";					//$NON-NLS-1$
	/** The field ID for level changes. */
	public static final String	ID_LEVEL			= PREFIX + "Level";						//$NON-NLS-1$
	/** The field ID for relative level changes. */
	public static final String	ID_RELATIVE_LEVEL	= PREFIX + "RelativeLevel";				//$NON-NLS-1$
	/** The field ID for difficulty changes. */
	public static final String	ID_DIFFICULTY		= PREFIX + "Difficulty";					//$NON-NLS-1$
	/** The field ID for point changes. */
	public static final String	ID_POINTS			= PREFIX + "Points";						//$NON-NLS-1$
	/** The field ID for page reference changes. */
	public static final String	ID_REFERENCE		= PREFIX + "Reference";					//$NON-NLS-1$
	/** The field ID for when the row hierarchy changes. */
	public static final String	ID_LIST_CHANGED		= PREFIX + "ListChanged";					//$NON-NLS-1$
	private static final String	NEWLINE				= "\n";									//$NON-NLS-1$
	private static final String	SPACE				= " ";										//$NON-NLS-1$
	private static final String	EMPTY				= "";										//$NON-NLS-1$
	private static final String	ASTERISK			= "*";										//$NON-NLS-1$
	private static final String	SLASH				= "/";										//$NON-NLS-1$
	private String				mName;
	private String				mSpecialization;
	private String				mTechLevel;
	/** The level. */
	protected int				mLevel;
	/** The relative level. */
	protected int				mRelativeLevel;
	private CMSkillAttribute	mAttribute;
	private CMSkillDifficulty	mDifficulty;
	/** The points spent. */
	protected int				mPoints;
	private String				mReference;

	/**
	 * Creates a string suitable for displaying the level.
	 * 
	 * @param level The skill level.
	 * @param relativeLevel The relative skill level.
	 * @param attribute The attribute the skill is based on.
	 * @param isContainer Whether this skill is a container or not.
	 * @return The formatted string.
	 */
	public static String getSkillDisplayLevel(int level, int relativeLevel, CMSkillAttribute attribute, boolean isContainer) {
		if (isContainer) {
			return EMPTY;
		}
		if (level < 0) {
			return "-"; //$NON-NLS-1$
		}
		return TKNumberUtils.format(level) + SLASH + attribute + TKNumberUtils.format(relativeLevel, true);
	}

	/**
	 * Creates a new skill.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param isContainer Whether or not this row allows children.
	 */
	public CMSkill(CMDataFile dataFile, boolean isContainer) {
		super(dataFile, isContainer);
		mName = getLocalizedName();
		mSpecialization = EMPTY;
		mTechLevel = null;
		mAttribute = CMSkillAttribute.DX;
		mDifficulty = CMSkillDifficulty.A;
		mPoints = 1;
		mReference = EMPTY;
		updateLevel(false);
	}

	/**
	 * Creates a clone of an existing skill and associates it with the specified data file.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param skill The skill to clone.
	 * @param deep Whether or not to clone the children, grandchildren, etc.
	 * @param forSheet Whether this is for a character sheet or a list.
	 */
	public CMSkill(CMDataFile dataFile, CMSkill skill, boolean deep, boolean forSheet) {
		super(dataFile, skill);
		mName = skill.mName;
		mSpecialization = skill.mSpecialization;
		mTechLevel = skill.mTechLevel;
		mAttribute = skill.mAttribute;
		mDifficulty = skill.mDifficulty;
		mPoints = forSheet ? skill.mPoints : 1;
		mReference = skill.mReference;
		if (forSheet && dataFile instanceof CMCharacter) {
			if (mTechLevel != null) {
				mTechLevel = ((CMCharacter) dataFile).getTechLevel();
			}
		} else {
			if (mTechLevel != null && mTechLevel.trim().length() > 0) {
				mTechLevel = EMPTY;
			}
		}
		updateLevel(false);
		if (deep) {
			int count = skill.getChildCount();

			for (int i = 0; i < count; i++) {
				TKRow row = skill.getChild(i);

				if (row instanceof CMSkill) {
					addChild(new CMSkill(dataFile, (CMSkill) row, true, forSheet));
				} else {
					addChild(new CMTechnique(dataFile, (CMTechnique) row, forSheet));
				}
			}
		}
	}

	/**
	 * Loads a skill and associates it with the specified data file.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMSkill(CMDataFile dataFile, TKXMLReader reader) throws IOException {
		this(dataFile, TAG_SKILL_CONTAINER.equals(reader.getName()));
		load(reader, false);
	}

	@Override public String getLocalizedName() {
		return Msgs.DEFAULT_NAME;
	}

	@Override public String getListChangedID() {
		return ID_LIST_CHANGED;
	}

	@Override public String getXMLTagName() {
		return canHaveChildren() ? TAG_SKILL_CONTAINER : TAG_SKILL;
	}

	@Override public String getRowType() {
		return "Skill"; //$NON-NLS-1$
	}

	@Override protected void prepareForLoad(boolean forUndo) {
		super.prepareForLoad(forUndo);
		mName = getLocalizedName();
		mSpecialization = EMPTY;
		mTechLevel = null;
		mAttribute = CMSkillAttribute.DX;
		mDifficulty = CMSkillDifficulty.A;
		mPoints = 1;
		mReference = EMPTY;
	}

	@Override protected void loadSubElement(TKXMLReader reader, boolean forUndo) throws IOException {
		String name = reader.getName();

		if (TAG_NAME.equals(name)) {
			mName = reader.readText().replace(NEWLINE, SPACE);
		} else if (TAG_SPECIALIZATION.equals(name)) {
			mSpecialization = reader.readText().replace(NEWLINE, SPACE);
		} else if (TAG_TECH_LEVEL.equals(name)) {
			mTechLevel = reader.readText().replace(NEWLINE, SPACE);
			if (mTechLevel != null && getDataFile() instanceof CMListFile) {
				mTechLevel = EMPTY;
			}
		} else if (TAG_REFERENCE.equals(name)) {
			mReference = reader.readText().replace(NEWLINE, SPACE);
		} else if (!forUndo && (TAG_SKILL.equals(name) || TAG_SKILL_CONTAINER.equals(name))) {
			addChild(new CMSkill(mDataFile, reader));
		} else if (!forUndo && CMTechnique.TAG_TECHNIQUE.equals(name)) {
			addChild(new CMTechnique(mDataFile, reader));
		} else if (!canHaveChildren()) {
			if (TAG_DIFFICULTY.equals(name)) {
				setDifficultyFromText(reader.readText().replace(NEWLINE, SPACE));
			} else if (TAG_POINTS.equals(name)) {
				mPoints = reader.readInteger(1);
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

	@Override public void saveSelf(TKXMLWriter out, boolean forUndo) {
		out.simpleTag(TAG_NAME, mName);
		if (!canHaveChildren()) {
			out.simpleTag(TAG_SPECIALIZATION, mSpecialization);
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
			out.simpleTag(TAG_DIFFICULTY, getDifficultyAsText());
			out.simpleTag(TAG_POINTS, mPoints);
		}
		out.simpleTag(TAG_REFERENCE, mReference);
	}

	/** @return The level. */
	public int getLevel() {
		return mLevel;
	}

	/** @return The relative level. */
	public int getRelativeLevel() {
		return mRelativeLevel;
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

	/** @return The specialization. */
	public String getSpecialization() {
		return mSpecialization;
	}

	/**
	 * @param specialization The specialization to set.
	 * @return Whether it was changed.
	 */
	public boolean setSpecialization(String specialization) {
		if (!mSpecialization.equals(specialization)) {
			mSpecialization = specialization;
			notifySingle(ID_SPECIALIZATION);
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

	/** @return The points. */
	public int getPoints() {
		return mPoints;
	}

	/**
	 * @param points The points to set.
	 * @return Whether it was changed.
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

	/**
	 * Call to force an update of the level and relative level for this skill or technique.
	 * 
	 * @param notify Whether or not a notification should be issued on a change.
	 */
	public void updateLevel(boolean notify) {
		int savedLevel = mLevel;
		int savedRelativeLevel = mRelativeLevel;
		CMSkillLevel level = calculateLevelSelf();

		mLevel = level.mLevel;
		mRelativeLevel = level.mRelativeLevel;

		if (notify) {
			startNotify();
			if (savedLevel != mLevel) {
				notify(ID_LEVEL, this);
			}
			if (savedRelativeLevel != mRelativeLevel) {
				notify(ID_RELATIVE_LEVEL, this);
			}
			endNotify();
		}
	}

	/** @return The calculated skill level. */
	protected CMSkillLevel calculateLevelSelf() {
		return calculateLevel(getCharacter(), this, getName(), getSpecialization(), getDefaults(), getAttribute(), getDifficulty(), getPoints(), new HashSet<CMSkill>());
	}

	/**
	 * @param excludes Skills to exclude, other than this one.
	 * @return The calculated level.
	 */
	public int getLevel(HashSet<CMSkill> excludes) {
		return calculateLevel(getCharacter(), this, getName(), getSpecialization(), getDefaults(), getAttribute(), getDifficulty(), getPoints(), excludes).mLevel;
	}

	/** @return The attribute. */
	public CMSkillAttribute getAttribute() {
		return mAttribute;
	}

	/** @return The difficulty. */
	public CMSkillDifficulty getDifficulty() {
		return mDifficulty;
	}

	/**
	 * @param attribute The attribute to set.
	 * @param difficulty The difficulty to set.
	 * @return Whether it was changed.
	 */
	public boolean setDifficulty(CMSkillAttribute attribute, CMSkillDifficulty difficulty) {
		if (mAttribute != attribute || mDifficulty != difficulty) {
			mAttribute = attribute;
			mDifficulty = difficulty;
			startNotify();
			notify(ID_DIFFICULTY, this);
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
			notifySingle(ID_REFERENCE);
			return true;
		}
		return false;
	}

	@Override public boolean contains(String text, boolean lowerCaseOnly) {
		return getName().toLowerCase().indexOf(text) != -1;
	}

	@Override public Object getData(TKColumn column) {
		return CSSkillColumnID.values()[column.getID()].getData(this);
	}

	@Override public String getDataAsText(TKColumn column) {
		return CSSkillColumnID.values()[column.getID()].getDataAsText(this);
	}

	/** @param text The combined attribute/difficulty to set. */
	public void setDifficultyFromText(String text) {
		CMSkillAttribute[] attribute = CMSkillAttribute.values();
		CMSkillDifficulty[] difficulty = CMSkillDifficulty.values();
		String input = text.trim();

		for (CMSkillAttribute element : attribute) {
			// We have to go backwards through the list to avoid the
			// regex grabbing the "H" in "VH".
			for (int j = difficulty.length - 1; j >= 0; j--) {
				if (input.matches("(?i).*" + element + ".*/.*" + difficulty[j] + ".*")) { //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
					setDifficulty(element, difficulty[j]);
					return;
				}
			}
		}
	}

	/** @return The formatted attribute/difficulty. */
	public String getDifficultyAsText() {
		if (canHaveChildren()) {
			return EMPTY;
		}
		return mAttribute + SLASH + mDifficulty;
	}

	@Override public String toString() {
		StringBuilder builder = new StringBuilder();

		builder.append(getName());
		if (!canHaveChildren()) {
			String techLevel = getTechLevel();
			String specialization = getSpecialization();

			if (techLevel != null) {
				builder.append("/TL"); //$NON-NLS-1$
				if (techLevel.length() > 0) {
					builder.append(techLevel);
				}
			}

			if (specialization.length() > 0) {
				builder.append(" ("); //$NON-NLS-1$
				builder.append(specialization);
				builder.append(')');
			}
		}
		return builder.toString();
	}

	@Override public BufferedImage getImage(boolean large) {
		return CSImage.getSkillIcon(large, true);
	}

	@Override public CSRowEditor<? extends CMRow> createEditor() {
		return new CSSkillEditor(this);
	}

	/**
	 * Calculates the skill level.
	 * 
	 * @param character The character the skill will be attached to.
	 * @param exclude A specific skill to exclude from default calculations.
	 * @param name The name of the skill.
	 * @param specialization The specialization of the skill.
	 * @param defaults The defaults the skill has.
	 * @param attribute The attribute the skill is based on.
	 * @param difficulty The difficulty of the skill.
	 * @param points The number of points spent in the skill.
	 * @param excludes The set of skills to exclude from any default calculations.
	 * @return The calculated skill level.
	 */
	public static CMSkillLevel calculateLevel(CMCharacter character, CMSkill exclude, String name, String specialization, List<CMSkillDefault> defaults, CMSkillAttribute attribute, CMSkillDifficulty difficulty, int points, HashSet<CMSkill> excludes) {
		int relativeLevel = difficulty.getBaseRelativeLevel();
		int level = attribute.getBaseSkillLevel(character);

		if (level != Integer.MIN_VALUE) {
			CMSkillDefault best;

			if (difficulty != CMSkillDifficulty.W) {
				best = getBestDefaultWithPoints(character, exclude, defaults, attribute, difficulty, excludes);
				if (best != null && best.getPoints() > 0) {
					points += best.getPoints();
				}
			} else {
				best = null;
				points /= 3;
			}

			if (points > 0) {
				if (points == 1) {
					// relativeLevel is preset to this point value
				} else if (points < 4) {
					relativeLevel++;
				} else {
					relativeLevel += 1 + points / 4;
				}
			} else if (best != null && best.getPoints() < 0) {
				relativeLevel = best.getAdjLevel() - level;
			} else {
				level = Integer.MIN_VALUE;
				relativeLevel = 0;
			}

			if (level != Integer.MIN_VALUE) {
				level += relativeLevel;
				level += character.getSkillComparedIntegerBonusFor(ID_NAME + ASTERISK, name, specialization);
				level += character.getIntegerBonusFor(ID_NAME + SLASH + name.toLowerCase());
				if (best != null) {
					if (level < best.getLevel()) {
						level = best.getLevel();
					}
				}
			}
		}
		return new CMSkillLevel(level, relativeLevel);
	}

	private static CMSkillDefault getBestDefaultWithPoints(CMCharacter character, CMSkill exclude, Collection<CMSkillDefault> defaults, CMSkillAttribute attribute, CMSkillDifficulty difficulty, HashSet<CMSkill> excludes) {
		CMSkillDefault best = getBestDefault(character, exclude, defaults, excludes);

		if (best != null) {
			int baseLine = attribute.getBaseSkillLevel(character) + difficulty.getBaseRelativeLevel();
			int level = best.getLevel();

			if (CMSkillDefaultType.Skill == best.getType()) {
				String name = best.getName();

				level -= character.getSkillComparedIntegerBonusFor(ID_NAME + ASTERISK, name, best.getSpecialization());
				level -= character.getIntegerBonusFor(ID_NAME + SLASH + name.toLowerCase());
			}
			best.setAdjLevel(level);
			if (level == baseLine) {
				best.setPoints(1);
			} else if (level == baseLine + 1) {
				best.setPoints(2);
			} else if (level > baseLine + 1) {
				best.setPoints(4 * (level - (baseLine + 1)));
			} else {
				best.setPoints(-best.getLevel());
			}
		}
		return best;
	}

	private static CMSkillDefault getBestDefault(CMCharacter character, CMSkill exclude, Collection<CMSkillDefault> defaults, HashSet<CMSkill> excludes) {
		if (character != null) {
			if (!defaults.isEmpty()) {
				int best = Integer.MIN_VALUE;
				CMSkillDefault bestSkill = null;

				excludes.add(exclude);
				for (CMSkillDefault skillDefault : defaults) {
					int level = skillDefault.getType().getSkillLevel(character, skillDefault, excludes);

					if (level > best) {
						best = level;
						bestSkill = new CMSkillDefault(skillDefault);
						bestSkill.setLevel(level);
					}
				}
				excludes.remove(exclude);
				return bestSkill;
			}
		}
		return null;
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		extractNameables(set, mName);
		extractNameables(set, mSpecialization);
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mName = nameNameables(map, mName);
		mSpecialization = nameNameables(map, mSpecialization);
	}
}
