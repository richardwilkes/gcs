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

package com.trollworks.gcs.character;

import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Row;

import java.util.ArrayList;
import java.util.List;

public class ReactionRow extends Row {
    private String        mFrom;
    private List<Integer> mAmounts;
    private List<String>  mSources;

    public ReactionRow(int bonus, String from, String source) {
        mFrom = from;
        mAmounts = new ArrayList<>();
        mAmounts.add(Integer.valueOf(bonus));
        mSources = new ArrayList<>();
        mSources.add(source);
    }

    public int getTotalAmount() {
        int total = 0;
        for (Integer one : mAmounts) {
            total += one.intValue();
        }
        return total;
    }

    public void addAmount(int amount, String source) {
        mAmounts.add(Integer.valueOf(amount));
        mSources.add(source);
    }

    public String getFrom() {
        return mFrom;
    }

    public List<Integer> getAmounts() {
        return mAmounts;
    }

    public List<String> getSources() {
        return mSources;
    }

    @Override
    public Object getData(Column column) {
        return ReactionColumn.values()[column.getID()].getData(this);
    }

    @Override
    public String getDataAsText(Column column) {
        return ReactionColumn.values()[column.getID()].getDataAsText(this);
    }

    @Override
    public void setData(Column column, Object data) {
        // Not used
    }

    @Override
    public String getToolTip(Column column) {
        return ReactionColumn.values()[column.getID()].getToolTip(this);
    }
}
