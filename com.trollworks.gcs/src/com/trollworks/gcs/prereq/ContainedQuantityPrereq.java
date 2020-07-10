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
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.text.MessageFormat;

/** An equipment contained quantity prerequisite. */
public class ContainedQuantityPrereq extends HasPrereq {
    /** The XML tag for this class. */
    public static final  String          TAG_ROOT          = "contained_quantity_prereq";
    private static final String          ATTRIBUTE_COMPARE = "compare";
    private static final String          KEY_QUALIFIER     = "qualifier";
    private              IntegerCriteria mQuantityCompare;

    /**
     * Creates a new prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     */
    public ContainedQuantityPrereq(PrereqList parent) {
        super(parent);
        mQuantityCompare = new IntegerCriteria(NumericCompareType.AT_MOST, 1);
    }

    /**
     * Loads a prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param m      The {@link JsonMap} to load from.
     */
    public ContainedQuantityPrereq(PrereqList parent, JsonMap m) throws IOException {
        this(parent);
        loadSelf(m, new LoadState());
    }

    /**
     * Loads a prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param reader The XML reader to load from.
     */
    public ContainedQuantityPrereq(PrereqList parent, XMLReader reader) throws IOException {
        this(parent);
        loadHasAttribute(reader);
        mQuantityCompare.setType(Enums.extract(reader.getAttribute(ATTRIBUTE_COMPARE), NumericCompareType.values(), NumericCompareType.AT_LEAST));
        mQuantityCompare.setQualifier(reader.readInteger(0));
    }

    /**
     * Creates a copy of the specified prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param prereq The prerequisite to clone.
     */
    protected ContainedQuantityPrereq(PrereqList parent, ContainedQuantityPrereq prereq) {
        super(parent, prereq);
        mQuantityCompare = new IntegerCriteria(prereq.mQuantityCompare);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof ContainedQuantityPrereq && super.equals(obj)) {
            return mQuantityCompare.equals(((ContainedQuantityPrereq) obj).mQuantityCompare);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    @Override
    public String getJSONTypeName() {
        return TAG_ROOT;
    }

    @Override
    public String getXMLTag() {
        return TAG_ROOT;
    }

    @Override
    public Prereq clone(PrereqList parent) {
        return new ContainedQuantityPrereq(parent, this);
    }

    @Override
    public void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        mQuantityCompare.load(m.getMap(KEY_QUALIFIER));
    }

    @Override
    public void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        mQuantityCompare.save(w, KEY_QUALIFIER);
    }

    /** @return The quantity comparison object. */
    public IntegerCriteria getQuantityCompare() {
        return mQuantityCompare;
    }

    @Override
    public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
        boolean satisfied = false;
        if (exclude instanceof Equipment) {
            Equipment equipment = (Equipment) exclude;
            satisfied = !equipment.canHaveChildren();
            if (!satisfied) {
                int qty = 0;
                for (Row child : equipment.getChildren()) {
                    if (child instanceof Equipment) {
                        qty += ((Equipment) child).getQuantity();
                    }
                }
                satisfied = mQuantityCompare.matches(qty);
            }
        }
        if (!has()) {
            satisfied = !satisfied;
        }
        if (!satisfied && builder != null) {
            builder.append(MessageFormat.format(I18n.Text("{0}{1} a contained quantity which {2}\n"), prefix, hasText(), mQuantityCompare));
        }
        return satisfied;
    }
}
