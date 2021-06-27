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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.EditorPanel;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.text.DoubleFormatter;
import com.trollworks.gcs.utility.text.IntegerFormatter;

import java.util.ArrayList;
import java.util.List;
import javax.swing.JComponent;
import javax.swing.JFormattedTextField.AbstractFormatter;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

/** A generic feature editor panel. */
public abstract class FeatureEditor extends EditorPanel {
    private static FeatureType LAST_FEATURE_TYPE = FeatureType.SKILL_LEVEL_BONUS;

    private ListRow                mRow;
    private Feature                mFeature;
    private PopupMenu<FeatureType> mBaseTypePopup;
    private PopupMenu<String>      mLeveledAmountPopup;

    /**
     * Creates a new FeatureEditor.
     *
     * @param row     The row this feature will belong to.
     * @param feature The feature to edit.
     * @return The newly created editor panel.
     */
    public static FeatureEditor create(ListRow row, Feature feature) {
        for (FeatureType featureType : FeatureType.values()) {
            if (featureType.matches(feature)) {
                return featureType.createFeatureEditor(row, feature);
            }
        }
        return null;
    }

    /**
     * Creates a new FeatureEditor.
     *
     * @param row     The row this feature will belong to.
     * @param feature The feature to edit.
     */
    protected FeatureEditor(ListRow row, Feature feature) {
        mRow = row;
        mFeature = feature;
        rebuild();
    }

    public ListRow getRow() {
        return mRow;
    }

    /** Rebuilds the contents of this panel with the current feature settings. */
    protected void rebuild() {
        removeAll();
        FlexGrid grid  = new FlexGrid();
        FlexRow  right = new FlexRow();
        rebuildSelf(grid, right);
        if (mFeature != null) {
            FontIconButton button = new FontIconButton(FontAwesome.TRASH,
                    I18n.text("Remove this feature"), (b) -> removeFeature());
            add(button);
            right.add(button);
        }
        FontIconButton button = new FontIconButton(FontAwesome.PLUS_CIRCLE,
                I18n.text("Add a feature"), (b) -> addFeature());
        add(button);
        right.add(button);
        grid.add(right, 0, 1);
        grid.apply(this);
        revalidate();
        repaint();
    }

    /**
     * Sub-classes must implement this method to add any components they want to be visible.
     *
     * @param grid  The general {@link FlexGrid}. Add items in column 0.
     * @param right The right-side {@link FlexRow}, situated in grid row 0, column 1.
     */
    protected abstract void rebuildSelf(FlexGrid grid, FlexRow right);

    /** @return The {@link PopupMenu} that allows the base feature type to be changed. */
    protected PopupMenu<FeatureType> addChangeBaseTypePopup() {
        List<FeatureType> choices = new ArrayList<>();
        FeatureType       current = null;
        for (FeatureType featureType : FeatureType.values()) {
            if (featureType.validRow(mRow)) {
                choices.add(featureType);
                if (featureType.matches(mFeature)) {
                    current = featureType;
                }
            }
        }
        if (current == null) {
            current = choices.get(0);
        }
        mBaseTypePopup = new PopupMenu<>(choices, (p) -> {
            LAST_FEATURE_TYPE = mBaseTypePopup.getSelectedItem();
            if (LAST_FEATURE_TYPE != null && !LAST_FEATURE_TYPE.matches(mFeature)) {
                JComponent parent = (JComponent) getParent();
                Commitable.sendCommitToFocusOwner();
                try {
                    parent.add(create(mRow, LAST_FEATURE_TYPE.createFeature()), UIUtilities.getIndexOf(parent, this));
                } catch (Exception exception) {
                    // Shouldn't have a failure...
                    Log.error(exception);
                }
                parent.remove(this);
                parent.revalidate();
                parent.repaint();
            }
        });
        mBaseTypePopup.setSelectedItem(current, false);
        add(mBaseTypePopup);
        return mBaseTypePopup;
    }

    /**
     * @param amt The current {@link LeveledAmount}.
     * @param min The minimum value to allow.
     * @param max The maximum value to allow.
     * @return The {@link EditorField} that allows a {@link LeveledAmount} to be changed.
     */
    protected EditorField addLeveledAmountField(LeveledAmount amt, int min, int max) {
        AbstractFormatter formatter;
        Object            value;
        Object            prototype;
        if (amt.isDecimal()) {
            formatter = new DoubleFormatter(min, max, true);
            value = Double.valueOf(amt.getAmount());
            prototype = Double.valueOf(max + 0.25);
        } else {
            formatter = new IntegerFormatter(min, max, true);
            value = Integer.valueOf(amt.getIntegerAmount());
            prototype = Integer.valueOf(max);
        }
        EditorField field = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, value, prototype, null);
        field.putClientProperty(LeveledAmount.class, amt);
        add(field);
        return field;
    }

    protected PopupMenu<String> addLeveledAmountPopup(LeveledAmount amt, boolean usePerDie) {
        String per = usePerDie ? I18n.text("per die") : I18n.text("per level");
        mLeveledAmountPopup = new PopupMenu<>(new String[]{" ", per},
                (p) -> ((Bonus) mFeature).getAmount().setPerLevel(mLeveledAmountPopup.getSelectedIndex() == 1));
        mLeveledAmountPopup.setSelectedItem(amt.isPerLevel() ? per : " ", false);
        add(mLeveledAmountPopup);
        return mLeveledAmountPopup;
    }

    /** @return The underlying feature. */
    public Feature getFeature() {
        return mFeature;
    }

    private void addFeature() {
        JComponent parent = (JComponent) getParent();
        try {
            parent.add(create(mRow, LAST_FEATURE_TYPE.createFeature()));
        } catch (Exception exception) {
            // Shouldn't have a failure...
            Log.error(exception);
        }
        if (mFeature == null) {
            parent.remove(this);
        }
        parent.revalidate();
    }

    private void removeFeature() {
        JComponent parent = (JComponent) getParent();
        parent.remove(this);
        if (parent.getComponentCount() == 0) {
            parent.add(new NoFeature(mRow));
        }
        parent.revalidate();
        parent.repaint();
    }

    @Override
    public void editorFieldChanged(EditorField field) {
        LeveledAmount amt = (LeveledAmount) field.getClientProperty(LeveledAmount.class);
        if (amt != null) {
            if (amt.isDecimal()) {
                amt.setAmount(((Double) field.getValue()).doubleValue());
            } else {
                amt.setAmount(((Integer) field.getValue()).intValue());
            }
            notifyActionListeners();
        } else {
            super.editorFieldChanged(field);
        }
    }
}
