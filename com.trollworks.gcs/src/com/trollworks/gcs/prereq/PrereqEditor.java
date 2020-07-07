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

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.widget.EditorPanel;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import java.awt.event.ActionEvent;
import javax.swing.JComboBox;
import javax.swing.JComponent;

/** A generic prerequisite editor panel. */
public abstract class PrereqEditor extends EditorPanel {
    private static final String            CHANGE_BASE_TYPE = "ChangeBaseType";
    private static final String            CHANGE_HAS       = "ChangeHas";
    private static final Class<?>[]        BASE_TYPES       = new Class<?>[]{AttributePrereq.class, AdvantagePrereq.class, SkillPrereq.class, SpellPrereq.class, ContainedWeightPrereq.class, ContainedQuantityPrereq.class};
    /** The prerequisite this panel represents. */
    protected            Prereq            mPrereq;
    /** The row this prerequisite will be attached to. */
    protected            ListRow           mRow;
    private              int               mDepth;
    private              JComboBox<Object> mBaseTypeCombo;

    /**
     * Creates a new prerequisite editor panel.
     *
     * @param row    The owning row.
     * @param prereq The prerequisite to edit.
     * @param depth  The depth of this prerequisite.
     * @return The newly created editor panel.
     */
    public static PrereqEditor create(ListRow row, Prereq prereq, int depth) {
        if (prereq instanceof PrereqList) {
            return new ListPrereqEditor(row, (PrereqList) prereq, depth);
        }
        if (prereq instanceof AdvantagePrereq) {
            return new AdvantagePrereqEditor(row, (AdvantagePrereq) prereq, depth);
        }
        if (prereq instanceof SkillPrereq) {
            return new SkillPrereqEditor(row, (SkillPrereq) prereq, depth);
        }
        if (prereq instanceof SpellPrereq) {
            return new SpellPrereqEditor(row, (SpellPrereq) prereq, depth);
        }
        if (prereq instanceof AttributePrereq) {
            return new AttributePrereqEditor(row, (AttributePrereq) prereq, depth);
        }
        if (prereq instanceof ContainedWeightPrereq) {
            return new ContainedWeightPrereqEditor(row, (ContainedWeightPrereq) prereq, depth);
        }
        if (prereq instanceof ContainedQuantityPrereq) {
            return new ContainedQuantityPrereqEditor(row, (ContainedQuantityPrereq) prereq, depth);
        }
        return null;
    }

    /**
     * Creates a new generic prerequisite editor panel.
     *
     * @param row    The owning row.
     * @param prereq The prerequisite to edit.
     * @param depth  The depth of this prerequisite.
     */
    protected PrereqEditor(ListRow row, Prereq prereq, int depth) {
        super(20 * depth);
        mRow = row;
        mPrereq = prereq;
        mDepth = depth;
        rebuild();
    }

    /** Rebuilds the contents of this panel with the current prerequisite settings. */
    protected final void rebuild() {
        removeAll();
        FlexGrid grid  = new FlexGrid();
        FlexRow  left  = new FlexRow();
        FlexRow  right = new FlexRow();
        if (mPrereq.getParent() != null) {
            AndOrLabel andOrLabel = new AndOrLabel(mPrereq);
            add(andOrLabel);
            left.add(andOrLabel);
        }
        grid.add(left, 0, 0);
        rebuildSelf(left, grid, right);
        if (mDepth > 0) {
            IconButton button = new IconButton(Images.REMOVE, mPrereq instanceof PrereqList ? I18n.Text("Remove this prerequisite list") : I18n.Text("Remove this prerequisite"), () -> remove());
            add(button);
            right.add(button);
        }
        grid.add(right, 0, 2);
        grid.apply(this);
        revalidate();
        repaint();
    }

    /**
     * Sub-classes must implement this method to add any components they want to be visible.
     *
     * @param left  The left-side {@link FlexRow}, situated in grid row 0, column 0.
     * @param grid  The general {@link FlexGrid}. Add items in column 1.
     * @param right The right-side {@link FlexRow}, situated in grid row 0, column 2.
     */
    protected abstract void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right);

    /**
     * @param has The current value of the "has" attribute.
     * @return The {@link JComboBox} that allows the "has" attribute to be changed.
     */
    protected JComboBox<Object> addHasCombo(boolean has) {
        String hasText         = I18n.Text("has");
        String doesNotHaveText = I18n.Text("doesn't have");
        return addComboBox(CHANGE_HAS, new Object[]{hasText, doesNotHaveText}, has ? hasText : doesNotHaveText);
    }

    /** @return The {@link JComboBox} that allows the base prereq type to be changed. */
    protected JComboBox<Object> addChangeBaseTypeCombo() {
        Object[] choices = {I18n.Text("attribute"), I18n.Text("advantage"), I18n.Text("skill"), I18n.Text("spell(s)"), I18n.Text("contained weight"), I18n.Text("contained quantity of")};
        Class<?> type    = mPrereq.getClass();
        Object   current = choices[0];
        int      length  = BASE_TYPES.length;
        for (int i = 0; i < length; i++) {
            if (type.equals(BASE_TYPES[i])) {
                current = choices[i];
                break;
            }
        }
        mBaseTypeCombo = addComboBox(CHANGE_BASE_TYPE, choices, current);
        return mBaseTypeCombo;
    }

    /** @return The depth of this prerequisite. */
    public int getDepth() {
        return mDepth;
    }

    /** @return The underlying prerequisite. */
    public Prereq getPrereq() {
        return mPrereq;
    }

    private void remove() {
        JComponent parent = (JComponent) getParent();
        int        index  = UIUtilities.getIndexOf(parent, this);
        int        count  = countSelfAndDescendents(mPrereq);
        for (int i = 0; i < count; i++) {
            parent.remove(index);
        }
        mPrereq.removeFromParent();
        parent.revalidate();
        parent.repaint();
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (CHANGE_BASE_TYPE.equals(command)) {
            Class<?> type = BASE_TYPES[mBaseTypeCombo.getSelectedIndex()];
            if (!mPrereq.getClass().equals(type)) {
                JComponent parent    = (JComponent) getParent();
                PrereqList list      = mPrereq.getParent();
                int        listIndex = list.getIndexOf(mPrereq);
                try {
                    Prereq prereq;
                    if (type == ContainedWeightPrereq.class) {
                        prereq = new ContainedWeightPrereq(list, mRow.getDataFile().defaultWeightUnits());
                    } else {
                        prereq = (Prereq) type.getConstructor(PrereqList.class).newInstance(list);
                    }
                    if (prereq instanceof HasPrereq && mPrereq instanceof HasPrereq) {
                        ((HasPrereq) prereq).has(((HasPrereq) mPrereq).has());
                    }
                    list.add(listIndex, prereq);
                    list.remove(mPrereq);
                    parent.add(create(mRow, prereq, mDepth), UIUtilities.getIndexOf(parent, this));
                } catch (Exception exception) {
                    // Shouldn't have a failure...
                    Log.error(exception);
                }
                parent.remove(this);
                parent.revalidate();
                parent.repaint();
                ListPrereqEditor.setLastItemType(type);
            }
        } else if (CHANGE_HAS.equals(command)) {
            ((HasPrereq) mPrereq).has(((JComboBox<?>) event.getSource()).getSelectedIndex() == 0);
        } else {
            super.actionPerformed(event);
        }
    }

    private int countSelfAndDescendents(Prereq prereq) {
        int count = 1;
        if (prereq instanceof PrereqList) {
            for (Prereq one : ((PrereqList) prereq).getChildren()) {
                count += countSelfAndDescendents(one);
            }
        }
        return count;
    }
}
