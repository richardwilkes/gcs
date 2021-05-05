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

import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.List;
import java.util.Map;
import java.util.Objects;

/** A Attribute prerequisite. */
public class AttributePrereq extends HasPrereq {
    public static final  String KEY_ROOT          = "attribute_prereq";
    private static final String KEY_WHICH         = "which";
    private static final String KEY_COMBINED_WITH = "combined_with";
    private static final String KEY_QUALIFIER     = "qualifier";

    private String          mWhich;
    private String          mCombinedWith;
    private IntegerCriteria mValueCompare;

    /**
     * Creates a new prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     */
    public AttributePrereq(PrereqList parent) {
        super(parent);
        mValueCompare = new IntegerCriteria(NumericCompareType.AT_LEAST, 10);
        List<AttributeDef> list = AttributeDef.getOrdered(Preferences.getInstance().getAttributes());
        mWhich = list.isEmpty() ? "st" : list.get(0).getID();
    }

    /**
     * Loads a prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param m      The {@link JsonMap} to load from.
     */
    public AttributePrereq(PrereqList parent, JsonMap m) throws IOException {
        this(parent);
        loadSelf(m, new LoadState());
    }

    /**
     * Creates a copy of the specified prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param prereq The prerequisite to clone.
     */
    protected AttributePrereq(PrereqList parent, AttributePrereq prereq) {
        super(parent, prereq);
        mWhich = prereq.mWhich;
        mCombinedWith = prereq.mCombinedWith;
        mValueCompare = new IntegerCriteria(prereq.mValueCompare);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof AttributePrereq && super.equals(obj)) {
            AttributePrereq ap = (AttributePrereq) obj;
            return mWhich.equals(ap.mWhich) && Objects.equals(mCombinedWith, ap.mCombinedWith) && mValueCompare.equals(ap.mValueCompare);
        }
        return false;
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    @Override
    public Prereq clone(PrereqList parent) {
        return new AttributePrereq(parent, this);
    }

    @Override
    public void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        mWhich = m.getString(KEY_WHICH);
        mCombinedWith = m.has(KEY_COMBINED_WITH) ? m.getString(KEY_COMBINED_WITH) : null;
        mValueCompare.load(m.getMap(KEY_QUALIFIER));
    }

    @Override
    public void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(KEY_WHICH, mWhich);
        if (mCombinedWith != null) {
            w.keyValue(KEY_COMBINED_WITH, mCombinedWith);
        }
        mValueCompare.save(w, KEY_QUALIFIER);
    }

    /** @return The type of comparison to make. */
    public String getWhich() {
        return mWhich;
    }

    /** @param which The type of comparison to make. */
    public void setWhich(String which) {
        mWhich = which;
    }

    /** @return The type of comparison to make. */
    public String getCombinedWith() {
        return mCombinedWith;
    }

    /** @param which The type of comparison to make. */
    public void setCombinedWith(String which) {
        mCombinedWith = which;
    }

    /** @return The value comparison object. */
    public IntegerCriteria getValueCompare() {
        return mValueCompare;
    }

    @Override
    public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
        boolean satisfied = mValueCompare.matches(character.getAttributeIntValue(mWhich) + (mCombinedWith != null ? character.getAttributeIntValue(mCombinedWith) : 0));
        if (!has()) {
            satisfied = !satisfied;
        }
        if (!satisfied && builder != null) {
            Map<String, AttributeDef> attributes = character.getSettings().getAttributes();
            AttributeDef              def        = attributes.get(mWhich);
            String                    text       = def != null ? def.getName() : "<unknown>";
            if (mCombinedWith != null) {
                def = attributes.get(mCombinedWith);
                text += "+" + (def != null ? def.getName() : "<unknown>");
            }
            builder.append(MessageFormat.format(I18n.Text("{0}{1} {2} which {3}"), prefix, hasText(), text, mValueCompare.toString()));
        }
        return satisfied;
    }
}
