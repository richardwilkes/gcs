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

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.collections.FilteredIterator;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.Localization;

import java.text.MessageFormat;
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.Queue;
import java.util.Set;

interface AdvantageContainer {
    int getAdjustedPoints(Advantage advantage);

    void saveAttributes(XMLWriter out);

    boolean checkSatisfied(Advantage advantage, StringBuilder builder, String prefix);

    String getNotes();
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

    @Override
    public String getNotes() {
        return ""; //$NON-NLS-1$
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

    @Override
    public String getNotes() {
        return ""; //$NON-NLS-1$
    }
}

class AlternateFormsAdvantageContainer implements AdvantageContainer {
    static final String TAG_COST_MOD_PERCENT = "cost_mod_percent";//$NON-NLS-1$
    static final String TAG_BASE_FORM        = "base_form";       //$NON-NLS-1$
    int                 mCostModPercent;
    String              mBaseForm;

    AlternateFormsAdvantageContainer(int costModPercent, String baseForm) {
        mCostModPercent = costModPercent;
        mBaseForm = baseForm;
    }

    private ArrayList<Advantage> findMatchingName(Advantage advantage) {
        ArrayList<Advantage> ret = new ArrayList<>();
        Set<Advantage> seen = new HashSet<>();
        Queue<Advantage> stack = new ArrayDeque<>();
        seen.add(advantage);
        stack.add(advantage);

        while (!stack.isEmpty()) {
            Advantage current = stack.remove();

            Advantage parent = (Advantage) current.getParent();
            if (parent != null && !seen.contains(parent)) {
                seen.add(parent);
                if (parent.getName().equals(mBaseForm)) {
                    ret.add(parent);
                }
                stack.add(parent);
            }

            if (!current.canHaveChildren()) {
                continue;
            }

            for (Advantage child : new FilteredIterator<>(current.getChildren(), Advantage.class)) {
                if (seen.contains(child)) {
                    continue;
                }
                seen.add(child);

                if (child.getName().equals(mBaseForm)) {
                    ret.add(parent);
                }
                stack.add(child);
            }
        }
        return ret;
    }

    @Override
    public int getAdjustedPoints(Advantage advantage) {
        int points = 0;

        Advantage baseForm = null;
        if (!mBaseForm.equals("")) { //$NON-NLS-1$
            ArrayList<Advantage> matching = findMatchingName(advantage);
            if (matching.size() == 1) {
                baseForm = matching.get(0);
            }
        }

        if (baseForm != null) {
            // Check if the base form is a parent of the Alternate Form Group.
            boolean baseIsParent = false;
            {
                Advantage parent = (Advantage) advantage.getParent();
                while (parent != null) {
                    if (parent == baseForm) {
                        baseIsParent = true;
                        break;
                    }
                    parent = (Advantage) parent.getParent();
                }
            }

            if (baseIsParent) {
                // If it is we need to walk up from the Alternate Form Group,
                // creating a new advantage tree iteratively adding everything which wasn't your
                // advantage.
                Advantage toReplace = advantage;
                Advantage replacementAdvantage = null;
                while (true) {
                    Advantage parent = (Advantage) toReplace.getParent();

                    Advantage newParent = new Advantage(null, parent, false);
                    for (Advantage child : new FilteredIterator<>(parent.getChildren(), Advantage.class)) {
                        if (child == toReplace) {
                            if (replacementAdvantage != null) {
                                newParent.addChild(replacementAdvantage);
                            }
                            continue;
                        }
                        newParent.addChild(new Advantage(null, child, true));
                    }

                    if (parent == baseForm) {
                        baseForm = newParent;
                        break;
                    }

                    toReplace = parent;
                    replacementAdvantage = newParent;
                }

            }
        }

        // NOTE: Requires that Alternate Forms are children since otherwise would have to walk down
        // and subtract costs.
        int max = 0;
        for (Advantage child : new FilteredIterator<>(advantage.getChildren(), Advantage.class)) {
            int pts = child.getAdjustedPoints();
            if (child.canHaveChildren() && child.getContainerType() == AdvantageContainerType.ALTERNATE_FORM) {
                if (pts > max) {
                    max = pts;
                }
            } else {
                points += Advantage.applyRounding(pts, advantage.mRoundCostDown);
            }
        }

        int basePoints = baseForm != null ? baseForm.getAdjustedPoints() : 0;
        if (basePoints < max) {
            points += Advantage.applyRounding(Advantage.calculateModifierPoints(max - basePoints, 100 + mCostModPercent), advantage.mRoundCostDown);
        }
        return points;
    }

    @Override
    public void saveAttributes(XMLWriter out) {
        out.writeAttribute(TAG_COST_MOD_PERCENT, Integer.toString(mCostModPercent));
        out.writeAttribute(TAG_BASE_FORM, mBaseForm);
    }

    @Override
    public boolean checkSatisfied(Advantage advantage, StringBuilder builder, String prefix) {
        if (mBaseForm.equals("")) { //$NON-NLS-1$
            return true;
        }

        ArrayList<Advantage> matching = findMatchingName(advantage);
        if (matching.size() != 1) {
            builder.append(MessageFormat.format(NOT_ONE_ADVANTAGE_NAMED, prefix, Integer.valueOf(matching.size()), mBaseForm));
            return false;
        }

        return true;
    }

    @Override
    public String getNotes() {
        return ""; //$NON-NLS-1$
    }

    @Localize("{0}{1,number} Advantages named {2}")
    private static String NOT_ONE_ADVANTAGE_NAMED;
    static {
        Localization.initialize();
    }
}

class AlternateFormAdvantageContainer implements AdvantageContainer {

    static AlternateFormAdvantageContainer advantageContainer = null;

    static AlternateFormAdvantageContainer getInstance() {
        if (advantageContainer == null) {
            advantageContainer = new AlternateFormAdvantageContainer();
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
        Advantage parent = (Advantage) advantage.getParent();

        if (parent == null) {
            builder.append(MessageFormat.format(NO_PARENT_ADVANTAGE_CONTAINER, prefix));
            return false;
        }

        if (parent.getContainerType() != AdvantageContainerType.ALTERNATE_FORMS) {
            builder.append(MessageFormat.format(PARENT_NOT_ALTERNATE_FORMS_GROUP, prefix));
            return false;
        }

        return true;
    }

    @Override
    public String getNotes() {
        return ""; //$NON-NLS-1$
    }

    @Localize("{0}No parent Advantage Container")
    private static String NO_PARENT_ADVANTAGE_CONTAINER;
    @Localize("{0}Parent Advantage Container is not an Alternate Forms group")
    private static String PARENT_NOT_ALTERNATE_FORMS_GROUP;
    static {
        Localization.initialize();
    }
}
