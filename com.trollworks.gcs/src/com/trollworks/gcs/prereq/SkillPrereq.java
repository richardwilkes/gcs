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

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.xml.XMLReader;
import com.trollworks.gcs.utility.xml.XMLWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.Map;
import java.util.Set;

/** A Skill prerequisite. */
public class SkillPrereq extends NameLevelPrereq {
    /** The XML tag for this class. */
    public static final  String         TAG_ROOT           = "skill_prereq";
    private static final String         TAG_SPECIALIZATION = "specialization";
    private              StringCriteria mSpecializationCriteria;

    /**
     * Creates a new prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     */
    public SkillPrereq(PrereqList parent) {
        super(TAG_ROOT, parent);
        mSpecializationCriteria = new StringCriteria(StringCompareType.IS_ANYTHING, "");
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
        mSpecializationCriteria = new StringCriteria(StringCompareType.IS_ANYTHING, "");
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
        boolean         satisfied     = false;
        String          techLevel     = null;
        StringCriteria  nameCriteria  = getNameCriteria();
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
                if (satisfied) {
                    break;
                }
            }
        }
        if (!has()) {
            satisfied = !satisfied;
        }
        if (!satisfied && builder != null) {
            builder.append(MessageFormat.format(I18n.Text("{0}{1} a skill whose name {2}"), prefix, hasText(), nameCriteria.toString()));
            boolean notAnySpecialization = mSpecializationCriteria.getType() != StringCompareType.IS_ANYTHING;

            if (notAnySpecialization) {
                builder.append(MessageFormat.format(I18n.Text(", specialization {0},"), mSpecializationCriteria.toString()));
            }
            if (techLevel == null) {
                builder.append(MessageFormat.format(I18n.Text(" and level {0}"), levelCriteria.toString()));
            } else {
                if (notAnySpecialization) {
                    builder.append(",");
                }
                builder.append(MessageFormat.format(I18n.Text(" level {0} and tech level matches\n"), levelCriteria.toString()));
            }
        }
        return satisfied;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        super.fillWithNameableKeys(set);
        ListRow.extractNameables(set, mSpecializationCriteria.getQualifier());
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        super.applyNameableKeys(map);
        mSpecializationCriteria.setQualifier(ListRow.nameNameables(map, mSpecializationCriteria.getQualifier()));
    }

    /** @return The specialization comparison object. */
    public StringCriteria getSpecializationCriteria() {
        return mSpecializationCriteria;
    }
}
