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

package com.trollworks.gcs.prereq;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;


import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashMap;
import java.util.HashSet;

/** A Skill prerequisite. */
public class SkillPrereq extends NameLevelPrereq {
	@Localize("{0}{1} a skill whose name {2}")
	@Localize(locale = "de", value = "{0}{1} eine Fertigkeit, deren Name {2}")
	@Localize(locale = "ru", value = "{0}{1} умение с названием {2}")
	private static String SKILL_NAME_PART;
	@Localize(", specialization {0},")
	@Localize(locale = "de", value = ", Spezialisierung {0},")
	@Localize(locale = "ru", value = ", специализация {0},")
	private static String SPECIALIZATION_PART;
	@Localize(" and level {0}")
	@Localize(locale = "de", value = " und Fertigkeitswert {0}")
	@Localize(locale = "ru", value = " и уровень {0}\n ")
	private static String LEVEL_PART;
	@Localize(" level {0} and tech level matches\n")
	@Localize(locale = "de", value = " Fertigkeitswert {0} und Techlevel stimmt überein")
	@Localize(locale = "ru", value = " уровень {0} и ТУ совпадают\n")
	private static String LEVEL_AND_TL_PART;

	static {
		Localization.initialize();
	}

	/** The XML tag for this class. */
	public static final String	TAG_ROOT			= "skill_prereq";	//$NON-NLS-1$
	private static final String	TAG_SPECIALIZATION	= "specialization"; //$NON-NLS-1$
	private StringCriteria		mSpecializationCriteria;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public SkillPrereq(PrereqList parent) {
		super(TAG_ROOT, parent);
		mSpecializationCriteria = new StringCriteria(StringCompareType.IS_ANYTHING, ""); //$NON-NLS-1$
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 */
	public SkillPrereq(PrereqList parent, XMLReader reader) throws IOException {
		super(parent, reader);
	}

	private SkillPrereq(PrereqList parent, SkillPrereq prereq) {
		super(parent, prereq);
		mSpecializationCriteria = new StringCriteria(prereq.mSpecializationCriteria);
	}

	@Override
	protected void initializeForLoad() {
		mSpecializationCriteria = new StringCriteria(StringCompareType.IS_ANYTHING, ""); //$NON-NLS-1$
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof SkillPrereq && super.equals(obj)) {
			return mSpecializationCriteria.equals(((SkillPrereq) obj).mSpecializationCriteria);
		}
		return false;
	}

	@Override
	protected void loadSelf(XMLReader reader) throws IOException {
		if (TAG_SPECIALIZATION.equals(reader.getName())) {
			mSpecializationCriteria.load(reader);
		} else {
			super.loadSelf(reader);
		}
	}

	@Override
	protected void saveSelf(XMLWriter out) {
		mSpecializationCriteria.save(out, TAG_SPECIALIZATION);
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public Prereq clone(PrereqList parent) {
		return new SkillPrereq(parent, this);
	}

	@Override
	public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
		boolean satisfied = false;
		String techLevel = null;
		StringCriteria nameCriteria = getNameCriteria();
		IntegerCriteria levelCriteria = getLevelCriteria();

		if (exclude instanceof Skill) {
			techLevel = ((Skill) exclude).getTechLevel();
		}

		for (Skill skill : character.getSkillsIterator()) {
			if (exclude != skill && nameCriteria.matches(skill.getName()) && mSpecializationCriteria.matches(skill.getSpecialization())) {
				satisfied = levelCriteria.matches(skill.getLevel());
				if (satisfied && techLevel != null) {
					String otherTL = skill.getTechLevel();
					satisfied = otherTL == null || techLevel.equals(otherTL);
				}
				break;
			}
		}
		if (!has()) {
			satisfied = !satisfied;
		}
		if (!satisfied && builder != null) {
			builder.append(MessageFormat.format(SKILL_NAME_PART, prefix, has() ? HAS : DOES_NOT_HAVE, nameCriteria.toString()));
			boolean notAnySpecialization = mSpecializationCriteria.getType() != StringCompareType.IS_ANYTHING;

			if (notAnySpecialization) {
				builder.append(MessageFormat.format(SPECIALIZATION_PART, mSpecializationCriteria.toString()));
			}
			if (techLevel == null) {
				builder.append(MessageFormat.format(LEVEL_PART, levelCriteria.toString()));
			} else {
				if (notAnySpecialization) {
					builder.append(","); //$NON-NLS-1$
				}
				builder.append(MessageFormat.format(LEVEL_AND_TL_PART, levelCriteria.toString()));
			}
		}
		return satisfied;
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		ListRow.extractNameables(set, mSpecializationCriteria.getQualifier());
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mSpecializationCriteria.setQualifier(ListRow.nameNameables(map, mSpecializationCriteria.getQualifier()));
	}

	/** @return The specialization comparison object. */
	public StringCriteria getSpecializationCriteria() {
		return mSpecializationCriteria;
	}
}
