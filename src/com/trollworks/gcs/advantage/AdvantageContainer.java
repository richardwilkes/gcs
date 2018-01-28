/*
 * Copyright (c) 1998-2018 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.advantage;

import com.trollworks.toolkit.collections.FilteredIterator;
import com.trollworks.toolkit.io.xml.XMLWriter;

import java.util.ArrayList;

interface AdvantageContainer {
    int getAdjustedPoints(Advantage advantage);

    void saveAttributes(XMLWriter out);

    boolean checkSatisfied(Advantage advantage, StringBuilder builder, String prefix);
}

class SummativeAdvantageContainer implements AdvantageContainer {

    static SummativeAdvantageContainer advantageContainer = null;

    static SummativeAdvantageContainer getInstance() {
        if (advantageContainer == null) {
            advantageContainer = new SummativeAdvantageContainer();
        }
        return advantageContainer;
    }

    @Override
    public int getAdjustedPoints(Advantage advantage) {
        int points = 0;
        for (Advantage child : new FilteredIterator<>(advantage.getChildren(), Advantage.class)) {
            points += child.getAdjustedPoints();
        }
        return points;
    }

    @Override
    public void saveAttributes(XMLWriter out) {
        // No Additional attributes required
    }

    @Override
    public boolean checkSatisfied(Advantage advantage, StringBuilder builder, String prefix) {
        return true;
    }
}

class AlternativeAbilitiesAdvantageContainer implements AdvantageContainer {

    @Override
    public int getAdjustedPoints(Advantage advantage) {
        int points = 0;
        ArrayList<Integer> values = new ArrayList<>();
        for (Advantage child : new FilteredIterator<>(advantage.getChildren(), Advantage.class)) {
            int pts = child.getAdjustedPoints();
            values.add(Integer.valueOf(pts));
            if (pts > points) {
                points = pts;
            }
        }
        int max = points;
        boolean found = false;
        for (Integer one : values) {
            int value = one.intValue();
            if (!found && max == value) {
                found = true;
            } else {
                points += Advantage.applyRounding(Advantage.calculateModifierPoints(value, 20), advantage.mRoundCostDown);
            }
        }
        return points;
    }

    @Override
    public void saveAttributes(XMLWriter out) {
        // No Additional attributes required
    }

    @Override
    public boolean checkSatisfied(Advantage advantage, StringBuilder builder, String prefix) {
        return true;
    }
}
