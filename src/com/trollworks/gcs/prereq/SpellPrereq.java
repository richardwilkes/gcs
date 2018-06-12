/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.Localization;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashMap;
import java.util.HashSet;

/** A Spell prerequisite. */
public class SpellPrereq extends HasPrereq {
    @Localize("spell")
    @Localize(locale = "de", value = "Zauber")
    @Localize(locale = "ru", value = "заклинание")
    @Localize(locale = "es", value = "sortilegio")
    private static String ONE_SPELL;
    @Localize("spells")
    @Localize(locale = "de", value = "Zauber")
    @Localize(locale = "ru", value = "заклинания")
    @Localize(locale = "es", value = "sortilegios")
    private static String MULTIPLE_SPELLS;
    @Localize("{0}{1} {2} {3} whose name {4}\n")
    @Localize(locale = "de", value = "{0}{1} {2} {3}, deren/dessen Namen {4}\n")
    @Localize(locale = "ru", value = "{0}{1} {2} {3} с названием {4}\n")
    @Localize(locale = "es", value = "{0}{1} {2} {3}, cuyo nombre es {4}\n")
    private static String WHOSE_NAME;
    @Localize("{0}{1} {2} {3} of any kind\n")
    @Localize(locale = "de", value = "{0}{1} {2} {3} jeglicher Art\n")
    @Localize(locale = "ru", value = "{0}{1} {2} {3} любого вида\n ")
    @Localize(locale = "es", value = "{0}{1} {2} {3} de cualquier tipo\n")
    private static String OF_ANY_KIND;
    @Localize("{0}{1} {2} {3} whose college {4}\n")
    @Localize(locale = "de", value = "{0}{1} {2} {3}, deren/dessen Schule {4}\n")
    @Localize(locale = "ru", value = "{0}{1} {2} {3} со школой {4}\n")
    @Localize(locale = "es", value = "{0}{1} {2} {3} cuya escuela se llama {4}\n")
    private static String WHOSE_COLLEGE;
    @Localize("{0}{1} college count which {2}\n")
    @Localize(locale = "de", value = "{0}{1} Zauber von {4} unterschiedlichen Schulen\n")
    @Localize(locale = "ru", value = "{0}{1} заклинаний школы {2}\n")
    @Localize(locale = "es", value = "{0}{1} Escuela que cuenta como {2}\n")
    private static String COLLEGE_COUNT;

    static {
        Localization.initialize();
    }

    /** The XML tag for this class. */
    public static final String  TAG_ROOT          = "spell_prereq"; //$NON-NLS-1$
    /** The tag/type for name comparison. */
    public static final String  TAG_NAME          = "name"; //$NON-NLS-1$
    /** The tag/type for any. */
    public static final String  TAG_ANY           = "any"; //$NON-NLS-1$
    /** The tag/type for college name comparison. */
    public static final String  TAG_COLLEGE       = "college"; //$NON-NLS-1$
    /** The tag/type for college count comparison. */
    public static final String  TAG_COLLEGE_COUNT = "college_count"; //$NON-NLS-1$
    private static final String TAG_QUANTITY      = "quantity"; //$NON-NLS-1$
    private static final String EMPTY             = ""; //$NON-NLS-1$
    private String              mType;
    private StringCriteria      mStringCriteria;
    private IntegerCriteria     mQuantityCriteria;

    /**
     * Creates a new prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     */
    public SpellPrereq(PrereqList parent) {
        super(parent);
        mType = TAG_NAME;
        mStringCriteria = new StringCriteria(StringCompareType.IS, EMPTY);
        mQuantityCriteria = new IntegerCriteria(NumericCompareType.AT_LEAST, 1);
    }

    /**
     * Loads a prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param reader The XML reader to load from.
     */
    public SpellPrereq(PrereqList parent, XMLReader reader) throws IOException {
        this(parent);

        String marker = reader.getMarker();
        loadHasAttribute(reader);

        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();

                if (TAG_NAME.equals(name)) {
                    setType(TAG_NAME);
                    mStringCriteria.load(reader);
                } else if (TAG_ANY.equals(name)) {
                    setType(TAG_ANY);
                    mQuantityCriteria.load(reader);
                } else if (TAG_COLLEGE.equals(name)) {
                    setType(TAG_COLLEGE);
                    mStringCriteria.load(reader);
                } else if (TAG_COLLEGE_COUNT.equals(name)) {
                    setType(TAG_COLLEGE_COUNT);
                    mQuantityCriteria.load(reader);
                } else if (TAG_QUANTITY.equals(name)) {
                    mQuantityCriteria.load(reader);
                } else {
                    reader.skipTag(name);
                }
            }
        } while (reader.withinMarker(marker));
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
        if (obj instanceof SpellPrereq && super.equals(obj)) {
            SpellPrereq sp = (SpellPrereq) obj;
            return mType.equals(sp.mType) && mStringCriteria.equals(sp.mStringCriteria) && mQuantityCriteria.equals(sp.mQuantityCriteria);
        }
        return false;
    }

    @Override
    public String getXMLTag() {
        return TAG_ROOT;
    }

    @Override
    public Prereq clone(PrereqList parent) {
        return new SpellPrereq(parent, this);
    }

    @Override
    public void save(XMLWriter out) {
        out.startTag(TAG_ROOT);
        saveHasAttribute(out);
        out.finishTagEOL();
        if (mType == TAG_NAME || mType == TAG_COLLEGE) {
            mStringCriteria.save(out, mType);
            if (mQuantityCriteria.getType() != NumericCompareType.AT_LEAST || mQuantityCriteria.getQualifier() != 1) {
                mQuantityCriteria.save(out, TAG_QUANTITY);
            }
        } else if (mType == TAG_COLLEGE_COUNT) {
            mQuantityCriteria.save(out, mType);
        } else if (mType == TAG_ANY) {
            out.startTag(TAG_ANY);
            out.finishEmptyTagEOL();
            mQuantityCriteria.save(out, TAG_QUANTITY);
        }
        out.endTagEOL(TAG_ROOT, true);
    }

    /** @return The type of comparison to make. */
    public String getType() {
        return mType;
    }

    /**
     * @param type The type of comparison to make. Must be one of {@link #TAG_NAME},
     *            {@link #TAG_ANY}, {@link #TAG_COLLEGE}, or {@link #TAG_COLLEGE_COUNT}.
     */
    public void setType(String type) {
        if (type == TAG_NAME || type == TAG_COLLEGE || type == TAG_COLLEGE_COUNT || type == TAG_ANY) {
            mType = type;
        } else if (TAG_NAME.equals(type)) {
            mType = TAG_NAME;
        } else if (TAG_ANY.equals(type)) {
            mType = TAG_ANY;
        } else if (TAG_COLLEGE.equals(type)) {
            mType = TAG_COLLEGE;
        } else if (TAG_COLLEGE_COUNT.equals(type)) {
            mType = TAG_COLLEGE_COUNT;
        } else {
            mType = TAG_NAME;
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
        HashSet<String> colleges = new HashSet<>();
        String techLevel = null;
        int count = 0;
        boolean satisfied;

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
                    if (mType == TAG_NAME) {
                        if (mStringCriteria.matches(spell.getName())) {
                            count++;
                        }
                    } else if (mType == TAG_ANY) {
                        count++;
                    } else if (mType == TAG_COLLEGE) {
                        if (mStringCriteria.matches(spell.getCollege())) {
                            count++;
                        }
                    } else if (mType == TAG_COLLEGE_COUNT) {
                        colleges.add(spell.getCollege());
                    }
                }
            }
        }

        if (mType == TAG_COLLEGE_COUNT) {
            count = colleges.size();
        }

        satisfied = mQuantityCriteria.matches(count);
        if (!has()) {
            satisfied = !satisfied;
        }
        if (!satisfied && builder != null) {
            if (mType == TAG_NAME) {
                builder.append(MessageFormat.format(WHOSE_NAME, prefix, has() ? HAS : DOES_NOT_HAVE, mQuantityCriteria.toString(EMPTY), mQuantityCriteria.getQualifier() == 1 ? ONE_SPELL : MULTIPLE_SPELLS, mStringCriteria.toString()));
            } else if (mType == TAG_ANY) {
                builder.append(MessageFormat.format(OF_ANY_KIND, prefix, has() ? HAS : DOES_NOT_HAVE, mQuantityCriteria.toString(EMPTY), mQuantityCriteria.getQualifier() == 1 ? ONE_SPELL : MULTIPLE_SPELLS));
            } else if (mType == TAG_COLLEGE) {
                builder.append(MessageFormat.format(WHOSE_COLLEGE, prefix, has() ? HAS : DOES_NOT_HAVE, mQuantityCriteria.toString(EMPTY), mQuantityCriteria.getQualifier() == 1 ? ONE_SPELL : MULTIPLE_SPELLS, mStringCriteria.toString()));
            } else if (mType == TAG_COLLEGE_COUNT) {
                builder.append(MessageFormat.format(COLLEGE_COUNT, prefix, has() ? HAS : DOES_NOT_HAVE, mQuantityCriteria.toString()));
            }
        }
        return satisfied;
    }

    @Override
    public void fillWithNameableKeys(HashSet<String> set) {
        if (mType != TAG_COLLEGE_COUNT) {
            ListRow.extractNameables(set, mStringCriteria.getQualifier());
        }
    }

    @Override
    public void applyNameableKeys(HashMap<String, String> map) {
        if (mType != TAG_COLLEGE_COUNT) {
            mStringCriteria.setQualifier(ListRow.nameNameables(map, mStringCriteria.getQualifier()));
        }
    }
}
