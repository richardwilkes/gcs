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

package com.trollworks.gcs.model.prereq;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.criteria.CMIntegerCriteria;
import com.trollworks.gcs.model.criteria.CMStringCompareType;
import com.trollworks.gcs.model.criteria.CMStringCriteria;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashMap;
import java.util.HashSet;

/** A Skill prerequisite. */
public class CMSkillPrereq extends CMNameLevelPrereq {
	/** The XML tag for this class. */
	public static final String	TAG_ROOT			= "skill_prereq";	//$NON-NLS-1$
	private static final String	TAG_SPECIALIZATION	= "specialization"; //$NON-NLS-1$
	private CMStringCriteria	mSpecializationCriteria;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public CMSkillPrereq(CMPrereqList parent) {
		super(TAG_ROOT, parent);
		mSpecializationCriteria = new CMStringCriteria(CMStringCompareType.IS_ANYTHING, ""); //$NON-NLS-1$
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMSkillPrereq(CMPrereqList parent, TKXMLReader reader) throws IOException {
		super(parent, reader);
	}

	private CMSkillPrereq(CMPrereqList parent, CMSkillPrereq prereq) {
		super(parent, prereq);
		mSpecializationCriteria = new CMStringCriteria(prereq.mSpecializationCriteria);
	}

	@Override protected void initializeForLoad() {
		mSpecializationCriteria = new CMStringCriteria(CMStringCompareType.IS_ANYTHING, ""); //$NON-NLS-1$
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMSkillPrereq) {
			CMSkillPrereq other = (CMSkillPrereq) obj;

			return super.equals(other) && mSpecializationCriteria.equals(other.mSpecializationCriteria);
		}
		return false;
	}

	@Override protected void loadSelf(TKXMLReader reader) throws IOException {
		if (TAG_SPECIALIZATION.equals(reader.getName())) {
			mSpecializationCriteria.load(reader);
		} else {
			super.loadSelf(reader);
		}
	}

	@Override protected void saveSelf(TKXMLWriter out) {
		mSpecializationCriteria.save(out, TAG_SPECIALIZATION);
	}

	@Override public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override public CMPrereq clone(CMPrereqList parent) {
		return new CMSkillPrereq(parent, this);
	}

	@Override public boolean satisfied(CMCharacter character, CMRow exclude, StringBuilder builder, String prefix) {
		boolean satisfied = false;
		String techLevel = null;
		CMStringCriteria nameCriteria = getNameCriteria();
		CMIntegerCriteria levelCriteria = getLevelCriteria();

		if (exclude instanceof CMSkill) {
			techLevel = ((CMSkill) exclude).getTechLevel();
		}

		for (CMSkill skill : character.getSkillsIterator()) {
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
			builder.append(MessageFormat.format(Msgs.SKILL_NAME_PART, prefix, has() ? Msgs.HAS : Msgs.DOES_NOT_HAVE, nameCriteria.toString()));
			boolean notAnySpecialization = mSpecializationCriteria.getType() != CMStringCompareType.IS_ANYTHING;

			if (notAnySpecialization) {
				builder.append(MessageFormat.format(Msgs.SPECIALIZATION_PART, mSpecializationCriteria.toString()));
			}
			if (techLevel == null) {
				builder.append(MessageFormat.format(Msgs.LEVEL_PART, levelCriteria.toString()));
			} else {
				if (notAnySpecialization) {
					builder.append(","); //$NON-NLS-1$
				}
				builder.append(MessageFormat.format(Msgs.LEVEL_AND_TL_PART, levelCriteria.toString()));
			}
		}
		return satisfied;
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		CMRow.extractNameables(set, mSpecializationCriteria.getQualifier());
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mSpecializationCriteria.setQualifier(CMRow.nameNameables(map, mSpecializationCriteria.getQualifier()));
	}

	/** @return The specialization comparison object. */
	public CMStringCriteria getSpecializationCriteria() {
		return mSpecializationCriteria;
	}
}
