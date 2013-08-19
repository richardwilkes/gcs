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
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.skills.CSTechniqueEditor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.utility.TKNumberUtils;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashMap;
import java.util.HashSet;

/** A GURPS Technique. */
public class CMTechnique extends CMSkill {
	/** The XML tag used for items. */
	public static final String	TAG_TECHNIQUE	= "technique";	//$NON-NLS-1$
	private static final String	ATTRIBUTE_LIMIT	= "limit";		//$NON-NLS-1$
	private CMSkillDefault		mDefault;
	private boolean				mLimited;
	private int					mLimitModifier;

	/**
	 * Calculates the technique level.
	 * 
	 * @param character The character the technique will be attached to.
	 * @param name The name of the technique.
	 * @param specialization The specialization of the technique.
	 * @param def The default the technique is based on.
	 * @param difficulty The difficulty of the technique.
	 * @param points The number of points spent in the technique.
	 * @param limited Whether the technique has been limited or not.
	 * @param limitModifier The maximum bonus the technique can grant.
	 * @return The calculated technique level.
	 */
	public static CMSkillLevel calculateTechniqueLevel(CMCharacter character, String name, String specialization, CMSkillDefault def, CMSkillDifficulty difficulty, int points, boolean limited, int limitModifier) {
		int relativeLevel = 0;
		int level = getBaseLevel(character, def);

		if (level != Integer.MIN_VALUE) {
			int baseLevel = level;

			level += def.getModifier();
			if (difficulty == CMSkillDifficulty.H) {
				points--;
			}
			if (points > 0) {
				relativeLevel = points;
			}

			if (level != Integer.MIN_VALUE) {
				level += relativeLevel + character.getIntegerBonusFor(ID_NAME + "/" + name.toLowerCase()) + character.getSkillComparedIntegerBonusFor(ID_NAME + "*", name, specialization); //$NON-NLS-1$ //$NON-NLS-2$
			}
			if (limited) {
				int max = baseLevel + limitModifier;

				if (level > max) {
					relativeLevel -= level - max;
					level = max;
				}
			}
		}
		return new CMSkillLevel(level, relativeLevel);
	}

	private static int getBaseLevel(CMCharacter character, CMSkillDefault def) {
		CMSkillDefaultType type = def.getType();

		if (CMSkillDefaultType.Skill == type) {
			CMSkill skill = character != null ? character.getBestSkillNamed(def.getName(), def.getSpecialization(), false, new HashSet<CMSkill>()) : null;

			return skill != null ? skill.getLevel() : Integer.MIN_VALUE;
		}
		// Take the modifier back out, as we wanted the base, not the final value.
		return type.getSkillLevelFast(character, def, null) - def.getModifier();
	}

	/**
	 * Creates a string suitable for displaying the level.
	 * 
	 * @param level The skill level.
	 * @param relativeLevel The relative skill level.
	 * @param modifier The modifer to the skill level.
	 * @return The formatted string.
	 */
	public static String getTechniqueDisplayLevel(int level, int relativeLevel, int modifier) {
		if (level < 0) {
			return "-"; //$NON-NLS-1$
		}
		return TKNumberUtils.format(level) + "/" + TKNumberUtils.format(relativeLevel + modifier, true); //$NON-NLS-1$
	}

	/**
	 * Creates a new technique.
	 * 
	 * @param dataFile The data file to associate it with.
	 */
	public CMTechnique(CMDataFile dataFile) {
		super(dataFile, false);
		mDefault = new CMSkillDefault(CMSkillDefaultType.Skill, Msgs.DEFAULT_NAME, null, 0);
		updateLevel(false);
	}

	/**
	 * Creates a clone of an existing technique and associates it with the specified data file.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param technique The technique to clone.
	 * @param forSheet Whether this is for a character sheet or a list.
	 */
	public CMTechnique(CMDataFile dataFile, CMTechnique technique, boolean forSheet) {
		super(dataFile, technique, false, forSheet);
		mPoints = forSheet ? technique.mPoints : getDifficulty() == CMSkillDifficulty.A ? 1 : 2;
		mDefault = new CMSkillDefault(technique.mDefault);
		mLimited = technique.mLimited;
		mLimitModifier = technique.mLimitModifier;
		updateLevel(false);
	}

	/**
	 * Loads a technique and associates it with the specified data file.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMTechnique(CMDataFile dataFile, TKXMLReader reader) throws IOException {
		this(dataFile);
		load(reader, false);
		if (!(dataFile instanceof CMCharacter)) {
			mPoints = getDifficulty() == CMSkillDifficulty.A ? 1 : 2;
		}
	}

	@Override public String getLocalizedName() {
		return Msgs.TECHNIQUE_DEFAULT_NAME;
	}

	@Override public String getXMLTagName() {
		return TAG_TECHNIQUE;
	}

	@Override public String getRowType() {
		return "Technique"; //$NON-NLS-1$
	}

	@Override protected void prepareForLoad(boolean forUndo) {
		super.prepareForLoad(forUndo);
		mDefault = new CMSkillDefault(CMSkillDefaultType.Skill, Msgs.DEFAULT_NAME, null, 0);
		mLimited = false;
		mLimitModifier = 0;
	}

	@Override protected void loadAttributes(TKXMLReader reader, boolean forUndo) {
		String value = reader.getAttribute(ATTRIBUTE_LIMIT);

		if (value != null && value.length() > 0) {
			mLimited = true;
			try {
				mLimitModifier = Integer.parseInt(value);
			} catch (Exception exception) {
				mLimited = false;
				mLimitModifier = 0;
			}
		}
		super.loadAttributes(reader, forUndo);
	}

	@Override protected void loadSubElement(TKXMLReader reader, boolean forUndo) throws IOException {
		if (CMSkillDefault.TAG_ROOT.equals(reader.getName())) {
			mDefault = new CMSkillDefault(reader);
		} else {
			super.loadSubElement(reader, forUndo);
		}
	}

	@Override public void saveSelf(TKXMLWriter out, boolean forUndo) {
		super.saveSelf(out, forUndo);
		mDefault.save(out);
	}

	@Override protected void saveAttributes(TKXMLWriter out, boolean forUndo) {
		if (mLimited) {
			out.writeAttribute(ATTRIBUTE_LIMIT, mLimitModifier);
		}
	}

	/**
	 * @param builder The {@link StringBuilder} to append this technique's satisfied/unsatisfied
	 *            description to. May be <code>null</code>.
	 * @param prefix The prefix to add to each line appended to the builder.
	 * @return <code>true</code> if this technique has its default satisfied.
	 */
	public boolean satisfied(StringBuilder builder, String prefix) {
		if (mDefault.getType() == CMSkillDefaultType.Skill) {
			CMSkill skill = getCharacter().getBestSkillNamed(mDefault.getName(), mDefault.getSpecialization(), true, new HashSet<CMSkill>());
			boolean satisfied = skill != null && skill.getPoints() > 0;

			if (!satisfied && builder != null) {
				if (skill != null) {
					builder.append(MessageFormat.format(Msgs.REQUIRES_SKILL, prefix, mDefault.getFullName()));
				} else {
					builder.append(MessageFormat.format(Msgs.REQUIRES_POINTS, prefix, mDefault.getFullName()));
				}
			}
			return satisfied;
		}
		return true;
	}

	@Override protected CMSkillLevel calculateLevelSelf() {
		return calculateTechniqueLevel(getCharacter(), getName(), getSpecialization(), getDefault(), getDifficulty(), getPoints(), isLimited(), getLimitModifier());
	}

	@Override public void updateLevel(boolean notify) {
		if (mDefault != null) {
			super.updateLevel(notify);
		}
	}

	/**
	 * @param difficulty The difficulty to set.
	 * @return Whether it was modified or not.
	 */
	public boolean setDifficulty(CMSkillDifficulty difficulty) {
		return setDifficulty(getAttribute(), difficulty);
	}

	@Override public String getSpecialization() {
		return mDefault.getFullName();
	}

	@Override public boolean setSpecialization(String specialization) {
		return false;
	}

	@Override public String getTechLevel() {
		return null;
	}

	@Override public boolean setTechLevel(String techLevel) {
		return false;
	}

	/** @return The default to base the technique on. */
	public CMSkillDefault getDefault() {
		return mDefault;
	}

	/**
	 * @param def The new default to base the technique on.
	 * @return Whether anything was changed.
	 */
	public boolean setDefault(CMSkillDefault def) {
		if (!mDefault.equals(def)) {
			mDefault = new CMSkillDefault(def);
			return true;
		}
		return false;
	}

	@Override public void setDifficultyFromText(String text) {
		text = text.trim();
		if (CMSkillDifficulty.A.toString().equalsIgnoreCase(text)) {
			setDifficulty(CMSkillDifficulty.A);
		} else if (CMSkillDifficulty.H.toString().equalsIgnoreCase(text)) {
			setDifficulty(CMSkillDifficulty.H);
		}
	}

	@Override public String getDifficultyAsText() {
		return getDifficulty().toString();
	}

	/** @return Whether the maximum level is limited. */
	public boolean isLimited() {
		return mLimited;
	}

	/**
	 * Sets whether the maximum level is limited.
	 * 
	 * @param limited The value to set.
	 * @return Whether anything was changed.
	 */
	public boolean setLimited(boolean limited) {
		if (limited != mLimited) {
			mLimited = limited;
			return true;
		}
		return false;
	}

	/** @return The limit modifier. */
	public int getLimitModifier() {
		return mLimitModifier;
	}

	/**
	 * Sets the value of limit modifier.
	 * 
	 * @param limitModifier The value to set.
	 * @return Whether anything was changed.
	 */
	public boolean setLimitModifier(int limitModifier) {
		if (mLimitModifier != limitModifier) {
			mLimitModifier = limitModifier;
			return true;
		}
		return false;
	}

	@Override public CSRowEditor<? extends CMRow> createEditor() {
		return new CSTechniqueEditor(this);
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		mDefault.fillWithNameableKeys(set);
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mDefault.applyNameableKeys(map);
	}
}
