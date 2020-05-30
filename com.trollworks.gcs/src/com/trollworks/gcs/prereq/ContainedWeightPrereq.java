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
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.criteria.WeightCriteria;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.utility.units.WeightValue;
import com.trollworks.gcs.utility.xml.XMLReader;
import com.trollworks.gcs.utility.xml.XMLWriter;

import java.io.IOException;
import java.text.MessageFormat;

/** An equipment contained weight prerequisite. */
public class ContainedWeightPrereq extends HasPrereq {
    /** The XML tag for this class. */
    public static final  String         TAG_ROOT          = "contained_weight_prereq";
    private static final String         ATTRIBUTE_COMPARE = "compare";
    private              WeightCriteria mWeightCompare;

    /**
     * Creates a new prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     */
    public ContainedWeightPrereq(PrereqList parent, WeightUnits defUnits) {
        super(parent);
        mWeightCompare = new WeightCriteria(NumericCompareType.AT_MOST, new WeightValue(new Fixed6(5), defUnits));
    }

    /**
     * Loads a prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param reader The XML reader to load from.
     */
    public ContainedWeightPrereq(PrereqList parent, WeightUnits defUnits, XMLReader reader) throws IOException {
        this(parent, defUnits);
        loadHasAttribute(reader);
        mWeightCompare.setType(Enums.extract(reader.getAttribute(ATTRIBUTE_COMPARE), NumericCompareType.values(), NumericCompareType.AT_LEAST));
        mWeightCompare.setQualifier(WeightValue.extract(reader.readText(), false));
    }

    /**
     * Creates a copy of the specified prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param prereq The prerequisite to clone.
     */
    protected ContainedWeightPrereq(PrereqList parent, ContainedWeightPrereq prereq) {
        super(parent, prereq);
        mWeightCompare = new WeightCriteria(prereq.mWeightCompare);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof ContainedWeightPrereq && super.equals(obj)) {
            return mWeightCompare.equals(((ContainedWeightPrereq) obj).mWeightCompare);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    @Override
    public String getXMLTag() {
        return TAG_ROOT;
    }

    @Override
    public Prereq clone(PrereqList parent) {
        return new ContainedWeightPrereq(parent, this);
    }

    @Override
    public void save(XMLWriter out) {
        out.startTag(TAG_ROOT);
        saveHasAttribute(out);
        out.writeAttribute(ATTRIBUTE_COMPARE, Enums.toId(mWeightCompare.getType()));
        out.finishTag();
        out.writeEncodedData(mWeightCompare.getQualifier().toString(false));
        out.endTagEOL(TAG_ROOT, false);
    }

    /** @return The weight comparison object. */
    public WeightCriteria getWeightCompare() {
        return mWeightCompare;
    }

    @Override
    public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
        boolean satisfied = false;
        if (exclude instanceof Equipment) {
            Equipment equipment = (Equipment) exclude;
            satisfied = !equipment.canHaveChildren();
            if (!satisfied) {
                WeightValue weight = new WeightValue(equipment.getExtendedWeight());
                weight.subtract(equipment.getAdjustedWeight());
                satisfied = mWeightCompare.matches(weight);
            }
        }
        if (!has()) {
            satisfied = !satisfied;
        }
        if (!satisfied && builder != null) {
            builder.append(MessageFormat.format(I18n.Text("{0}{1} a contained weight which {2}\n"), prefix, hasText(), mWeightCompare));
        }
        return satisfied;
    }
}
