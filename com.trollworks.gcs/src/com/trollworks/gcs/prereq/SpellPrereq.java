/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashSet;
import java.util.Map;
import java.util.Objects;
import java.util.Set;

/** A Spell prerequisite. */
public class SpellPrereq extends HasPrereq {
    public static final  String KEY_ROOT          = "spell_prereq";
    public static final  String KEY_NAME          = "name";
    public static final  String KEY_ANY           = "any";
    public static final  String KEY_CATEGORY      = "category";
    public static final  String KEY_COLLEGE       = "college";
    public static final  String KEY_COLLEGE_COUNT = "college_count";
    private static final String KEY_QUANTITY      = "quantity";
    private static final String KEY_SUB_TYPE      = "sub_type";
    private static final String KEY_QUALIFIER     = "qualifier";

    private String          mType;
    private StringCriteria  mStringCriteria;
    private IntegerCriteria mQuantityCriteria;

    /**
     * Creates a new prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     */
    public SpellPrereq(PrereqList parent) {
        super(parent);
        mType = KEY_NAME;
        mStringCriteria = new StringCriteria(StringCompareType.IS, "");
        mQuantityCriteria = new IntegerCriteria(NumericCompareType.AT_LEAST, 1);
    }

    /**
     * Loads a prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param m      The {@link JsonMap} to load from.
     */
    public SpellPrereq(PrereqList parent, JsonMap m) throws IOException {
        this(parent);
        loadSelf(m, new LoadState());
    }

    /**
     * Creates a copy of the specified prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param prereq The prerequisite to clone.
     */
    protected SpellPrereq(PrereqList parent, SpellPrereq prereq) {
        super(parent, prereq);
        mType = prereq.mType;
        mStringCriteria = new StringCriteria(prereq.mStringCriteria);
        mQuantityCriteria = new IntegerCriteria(prereq.mQuantityCriteria);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof SpellPrereq sp && super.equals(obj)) {
            return mType.equals(sp.mType) && mStringCriteria.equals(sp.mStringCriteria) && mQuantityCriteria.equals(sp.mQuantityCriteria);
        }
        return false;
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    @Override
    public Prereq clone(PrereqList parent) {
        return new SpellPrereq(parent, this);
    }

    @Override
    public void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        mType = m.getString(KEY_SUB_TYPE);
        mQuantityCriteria.load(m.getMap(KEY_QUANTITY));
        mStringCriteria.load(m.getMap(KEY_QUALIFIER));
    }

    @Override
    public void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(KEY_SUB_TYPE, mType);
        if (KEY_NAME.equals(mType) || KEY_CATEGORY.equals(mType) || KEY_COLLEGE.equals(mType)) {
            mStringCriteria.save(w, KEY_QUALIFIER);
        }
        mQuantityCriteria.save(w, KEY_QUANTITY);
    }

    /** @return The type of comparison to make. */
    public String getType() {
        return mType;
    }

    /**
     * @param type The type of comparison to make. Must be one of {@link #KEY_NAME}, {@link
     *             #KEY_ANY}, {@link #KEY_CATEGORY}, {@link #KEY_COLLEGE}, or {@link
     *             #KEY_COLLEGE_COUNT}.
     */
    public void setType(String type) {
        if (KEY_NAME.equals(type)) {
            mType = KEY_NAME;
        } else if (KEY_ANY.equals(type)) {
            mType = KEY_ANY;
        } else if (KEY_CATEGORY.equals(type)) {
            mType = KEY_CATEGORY;
        } else if (KEY_COLLEGE.equals(type)) {
            mType = KEY_COLLEGE;
        } else if (KEY_COLLEGE_COUNT.equals(type)) {
            mType = KEY_COLLEGE_COUNT;
        } else {
            mType = KEY_NAME;
        }
    }

    /** @return The string comparison object. */
    public StringCriteria getStringCriteria() {
        return mStringCriteria;
    }

    /** @return The quantity comparison object. */
    public IntegerCriteria getQuantityCriteria() {
        return mQuantityCriteria;
    }

    @Override
    public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
        Set<String> colleges  = new HashSet<>();
        String      techLevel = null;
        int         count     = 0;
        boolean     satisfied;
        if (exclude instanceof Spell) {
            techLevel = ((Spell) exclude).getTechLevel();
        }
        for (Spell spell : character.getSpellsIterator()) {
            if (exclude != spell && spell.getPoints() > 0) {
                boolean ok;
                if (techLevel != null) {
                    String otherTL = spell.getTechLevel();

                    ok = otherTL == null || techLevel.equals(otherTL);
                } else {
                    ok = true;
                }
                if (ok) {
                    if (KEY_NAME.equals(mType)) {
                        if (mStringCriteria.matches(spell.getName())) {
                            count++;
                        }
                    } else if (KEY_ANY.equals(mType)) {
                        count++;
                    } else if (KEY_CATEGORY.equals(mType)) {
                        for (String category : spell.getCategories()) {
                            if (mStringCriteria.matches(category)) {
                                count++;
                                break;
                            }
                        }
                    } else if (KEY_COLLEGE.equals(mType)) {
                        for (String college : spell.getColleges()) {
                            if (mStringCriteria.matches(college)) {
                                count++;
                                break;
                            }
                        }
                    } else if (Objects.equals(mType, KEY_COLLEGE_COUNT)) {
                        colleges.addAll(spell.getColleges());
                    }
                }
            }
        }

        if (Objects.equals(mType, KEY_COLLEGE_COUNT)) {
            count = colleges.size();
        }

        satisfied = mQuantityCriteria.matches(count);
        if (!has()) {
            satisfied = !satisfied;
        }
        if (!satisfied && builder != null) {
            String oneSpell       = I18n.text("spell");
            String multipleSpells = I18n.text("spells");
            if (Objects.equals(mType, KEY_NAME)) {
                builder.append(MessageFormat.format(I18n.text("\n{0}{1} {2} {3} whose name {4}"), prefix, getHasText(), mQuantityCriteria.toString(""), mQuantityCriteria.getQualifier() == 1 ? oneSpell : multipleSpells, mStringCriteria.toString()));
            } else if (Objects.equals(mType, KEY_ANY)) {
                builder.append(MessageFormat.format(I18n.text("\n{0}{1} {2} {3} of any kind"), prefix, getHasText(), mQuantityCriteria.toString(""), mQuantityCriteria.getQualifier() == 1 ? oneSpell : multipleSpells));
            } else if (Objects.equals(mType, KEY_CATEGORY)) {
                builder.append(MessageFormat.format(I18n.text("\n{0}{1} {2} {3} whose category {4}"), prefix, getHasText(), mQuantityCriteria.toString(""), mQuantityCriteria.getQualifier() == 1 ? oneSpell : multipleSpells, mStringCriteria.toString()));
            } else if (Objects.equals(mType, KEY_COLLEGE)) {
                builder.append(MessageFormat.format(I18n.text("\n{0}{1} {2} {3} whose college {4}"), prefix, getHasText(), mQuantityCriteria.toString(""), mQuantityCriteria.getQualifier() == 1 ? oneSpell : multipleSpells, mStringCriteria.toString()));
            } else if (Objects.equals(mType, KEY_COLLEGE_COUNT)) {
                builder.append(MessageFormat.format(I18n.text("\n{0}{1} college count which {2}"), prefix, getHasText(), mQuantityCriteria.toString()));
            }
        }
        return satisfied;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        if (!Objects.equals(mType, KEY_COLLEGE_COUNT) && !Objects.equals(mType, KEY_ANY)) {
            ListRow.extractNameables(set, mStringCriteria.getQualifier());
        }
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        if (!Objects.equals(mType, KEY_COLLEGE_COUNT) && !Objects.equals(mType, KEY_ANY)) {
            mStringCriteria.setQualifier(ListRow.nameNameables(map, mStringCriteria.getQualifier()));
        }
    }
}
