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
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.WeightUnits;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Pattern;

/** A prerequisite list. */
public class PrereqList extends Prereq {
    public static final  String  KEY_ROOT          = "prereq_list";
    private static final String  KEY_WHEN_TL       = "when_tl";
    private static final String  KEY_ALL           = "all";
    private static final String  KEY_PREREQS       = "prereqs";
    private static final Pattern LINE_FEED_MATCHER = Pattern.compile("\n");

    private IntegerCriteria mWhenTLCriteria;
    private List<Prereq>    mPrereqs;
    private boolean         mWhenEnabled;
    private boolean         mAll;

    /**
     * Creates a new prerequisite list.
     *
     * @param parent The owning prerequisite list, if any.
     * @param all    Whether only one criteria in this list has to be met, or all of them must be
     *               met.
     */
    public PrereqList(PrereqList parent, boolean all) {
        super(parent);
        mAll = all;
        mWhenEnabled = false;
        mWhenTLCriteria = new IntegerCriteria(NumericCompareType.AT_LEAST, 0);
        mPrereqs = new ArrayList<>();
    }

    /**
     * Loads a prerequisite list.
     *
     * @param parent         The owning prerequisite list, if any.
     * @param m              The {@link JsonMap} to load from.
     * @param defWeightUnits The default weight units to use.
     */
    public PrereqList(PrereqList parent, WeightUnits defWeightUnits, JsonMap m) throws IOException {
        this(parent, true);
        LoadState state = new LoadState();
        state.mDefWeightUnits = defWeightUnits;
        loadSelf(m, state);
    }

    /**
     * Creates a clone of the specified prerequisite list.
     *
     * @param parent     The new owning prerequisite list, if any.
     * @param prereqList The prerequisite to clone.
     */
    public PrereqList(PrereqList parent, PrereqList prereqList) {
        super(parent);
        mAll = prereqList.mAll;
        mWhenEnabled = prereqList.mWhenEnabled;
        mWhenTLCriteria = new IntegerCriteria(prereqList.mWhenTLCriteria);
        mPrereqs = new ArrayList<>(prereqList.mPrereqs.size());
        for (Prereq prereq : prereqList.mPrereqs) {
            mPrereqs.add(prereq.clone(this));
        }
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof PrereqList list) {
            return mAll == list.mAll && mWhenEnabled == list.mWhenEnabled && mWhenTLCriteria.equals(list.mWhenTLCriteria) && mPrereqs.equals(list.mPrereqs);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    @Override
    public void loadSelf(JsonMap m, LoadState state) throws IOException {
        mAll = m.getBoolean(KEY_ALL);
        mWhenEnabled = m.has(KEY_WHEN_TL);
        if (mWhenEnabled) {
            mWhenTLCriteria.load(m.getMap(KEY_WHEN_TL));
        }
        if (m.has(KEY_PREREQS)) {
            JsonArray a     = m.getArray(KEY_PREREQS);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                JsonMap m1 = a.getMap(i);
                switch (m1.getString(DataFile.TYPE)) {
                case KEY_ROOT -> mPrereqs.add(new PrereqList(this, state.mDefWeightUnits, m1));
                case AdvantagePrereq.KEY_ROOT -> mPrereqs.add(new AdvantagePrereq(this, m1));
                case AttributePrereq.KEY_ROOT -> mPrereqs.add(new AttributePrereq(this, m1));
                case ContainedWeightPrereq.KEY_ROOT -> mPrereqs.add(new ContainedWeightPrereq(this, state.mDefWeightUnits, m1));
                case ContainedQuantityPrereq.KEY_ROOT -> mPrereqs.add(new ContainedQuantityPrereq(this, m1));
                case SkillPrereq.KEY_ROOT -> mPrereqs.add(new SkillPrereq(this, m1));
                case SpellPrereq.KEY_ROOT -> mPrereqs.add(new SpellPrereq(this, m1));
                }
            }
        }
    }

    @Override
    public void saveSelf(JsonWriter w) throws IOException {
        w.keyValue(KEY_ALL, mAll);
        if (mWhenEnabled) {
            mWhenTLCriteria.save(w, KEY_WHEN_TL);
        }
        if (!mPrereqs.isEmpty()) {
            w.key(KEY_PREREQS);
            w.startArray();
            for (Prereq prereq : mPrereqs) {
                prereq.save(w);
            }
            w.endArray();
        }
    }

    public boolean isEmpty() {
        return mPrereqs.isEmpty();
    }

    /** @return The character's TL criteria. */
    public IntegerCriteria getWhenTLCriteria() {
        return mWhenTLCriteria;
    }

    /** @return Whether the character's TL criteria check is enabled. */
    public boolean isWhenTLEnabled() {
        return mWhenEnabled;
    }

    /** @param enabled Whether the character's TL criteria check is enabled. */
    public void setWhenTLEnabled(boolean enabled) {
        mWhenEnabled = enabled;
    }

    /** @return Whether only one criteria in this list has to be met, or all of them must be met. */
    public boolean requiresAll() {
        return mAll;
    }

    /**
     * @param requiresAll Whether only one criteria in this list has to be met, or all of them must
     *                    be met.
     */
    public void setRequiresAll(boolean requiresAll) {
        mAll = requiresAll;
    }

    /**
     * @param prereq The prerequisite to work on.
     * @return The index of the specified prerequisite. -1 will be returned if the component isn't a
     *         direct child.
     */
    public int getIndexOf(Prereq prereq) {
        return mPrereqs.indexOf(prereq);
    }

    /** @return The number of children in this list. */
    public int getChildCount() {
        return mPrereqs.size();
    }

    /** @return The children of this list. */
    public List<Prereq> getChildren() {
        return Collections.unmodifiableList(mPrereqs);
    }

    /**
     * Adds the specified prerequisite to this list.
     *
     * @param index  The index to add the list at.
     * @param prereq The prerequisite to add.
     */
    public void add(int index, Prereq prereq) {
        mPrereqs.add(index, prereq);
    }

    /**
     * Removes the specified prerequisite from this list.
     *
     * @param prereq The prerequisite to remove.
     */
    public void remove(Prereq prereq) {
        if (mPrereqs.contains(prereq)) {
            mPrereqs.remove(prereq);
            prereq.mParent = null;
        }
    }

    @Override
    public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
        if (mWhenEnabled) {
            if (!mWhenTLCriteria.matches(Numbers.extractInteger(character.getProfile().getTechLevel(), 0, false))) {
                return true;
            }
        }

        int           satisfiedCount = 0;
        int           total          = mPrereqs.size();
        boolean       requiresAll    = requiresAll();
        StringBuilder localBuilder   = builder != null ? new StringBuilder() : null;
        for (Prereq prereq : mPrereqs) {
            if (prereq.satisfied(character, exclude, localBuilder, prefix)) {
                satisfiedCount++;
            }
        }
        if (localBuilder != null && !localBuilder.isEmpty()) {
            String indented = LINE_FEED_MATCHER.matcher(localBuilder.toString()).replaceAll("\n\u00a0\u00a0");
            localBuilder.setLength(0);
            localBuilder.append(indented);
        }

        boolean satisfied = satisfiedCount == total || !requiresAll && satisfiedCount > 0;
        if (!satisfied && localBuilder != null) {
            builder.append("\n");
            builder.append(prefix);
            builder.append(requiresAll ? I18n.text("Requires all of:") : I18n.text("Requires at least one of:"));
            builder.append(localBuilder);
        }
        return satisfied;
    }

    @Override
    public Prereq clone(PrereqList parent) {
        return new PrereqList(parent, this);
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        for (Prereq prereq : mPrereqs) {
            prereq.fillWithNameableKeys(set);
        }
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        for (Prereq prereq : mPrereqs) {
            prereq.applyNameableKeys(map);
        }
    }
}
