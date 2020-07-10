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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.xml.XMLNodeType;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

/** Describes a bonus. */
public abstract class Bonus extends Feature {
    /** The "amount" tag. */
    public static final String        TAG_AMOUNT = "amount";
    private             LeveledAmount mAmount;
    // The "parent" item that is providing this particular bonus (for information only).
    private             ListRow       mParent;

    /**
     * Creates a new bonus.
     *
     * @param amount The initial amount.
     */
    public Bonus(double amount) {
        mAmount = new LeveledAmount(amount);
    }

    /**
     * Creates a new bonus.
     *
     * @param amount The initial amount.
     */
    public Bonus(int amount) {
        mAmount = new LeveledAmount(amount);
    }

    /**
     * Creates a clone of the specified bonus.
     *
     * @param other The bonus to clone.
     */
    public Bonus(Bonus other) {
        mAmount = new LeveledAmount(other.mAmount);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Bonus) {
            return mAmount.equals(((Bonus) obj).mAmount);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    /** @param reader The XML reader to use. */
    protected final void load(XMLReader reader) throws IOException {
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                loadSelf(reader);
            }
        } while (reader.withinMarker(marker));
    }

    /** @param reader The XML reader to use. */
    protected void loadSelf(XMLReader reader) throws IOException {
        String tag = reader.getName();

        if (TAG_AMOUNT.equals(tag)) {
            mAmount.load(reader);
        } else {
            reader.skipTag(tag);
        }
    }

    protected void loadSelf(JsonMap m) throws IOException {
        mAmount.load(m);
    }

    protected void saveSelf(JsonWriter w) throws IOException {
        mAmount.saveInline(w);
    }

    /** @return The leveled amount. */
    public LeveledAmount getAmount() {
        return mAmount;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        // Nothing to do.
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        // Nothing to do.
    }

    public void setParent(ListRow parent) {
        mParent = parent;
    }

    public ListRow getParent() {
        return mParent;
    }

    public String getParentName() {
        return mParent == null ? I18n.Text("Unknown") : mParent.toString();
    }

    public void addToToolTip(StringBuilder toolTip) {
        if (toolTip != null) {
            toolTip.append("\n").append(getParentName()).append(" [").append(getToolTipAmount()).append("]");
        }
    }

    public String getToolTipAmount() {
        return getAmount().getAmountAsString();
    }

    protected static boolean matchesCategories(StringCriteria criteria, Set<String> categories) {
        if (categories != null) {
            for (String category : categories) {
                if (criteria.matches(category)) {
                    return true;
                }
            }
        }
        return criteria.matches("");
    }
}
