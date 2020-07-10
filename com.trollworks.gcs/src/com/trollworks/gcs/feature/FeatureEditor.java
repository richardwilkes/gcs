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

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.EditorPanel;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.text.DoubleFormatter;
import com.trollworks.gcs.utility.text.IntegerFormatter;

import java.awt.event.ActionEvent;
import java.beans.PropertyChangeEvent;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComboBox;
import javax.swing.JComponent;
import javax.swing.JFormattedTextField.AbstractFormatter;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

/** A generic feature editor panel. */
public abstract class FeatureEditor extends EditorPanel {
    private static final String            CHANGE_BASE_TYPE  = "ChangeBaseType";
    private static       FeatureType       LAST_FEATURE_TYPE = FeatureType.SKILL_LEVEL_BONUS;
    private              ListRow           mRow;
    private              Feature           mFeature;
    private              JComboBox<Object> mBaseTypeCombo;
    private              JComboBox<Object> mLeveledAmountCombo;

    /**
     * Creates a new {@link FeatureEditor}.
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
     * Creates a new {@link FeatureEditor}.
     *
     * @param row     The row this feature will belong to.
     * @param feature The feature to edit.
     */
    public FeatureEditor(ListRow row, Feature feature) {
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
            IconButton button = new IconButton(Images.REMOVE, I18n.Text("Remove this feature"), () -> removeFeature());
            add(button);
            right.add(button);
        }
        IconButton button = new IconButton(Images.ADD, I18n.Text("Add a feature"), () -> addFeature());
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

    /** @return The {@link JComboBox} that allows the base feature type to be changed. */
    protected JComboBox<Object> addChangeBaseTypeCombo() {
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
        mBaseTypeCombo = addComboBox(CHANGE_BASE_TYPE, choices.toArray(), current);
        return mBaseTypeCombo;
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
        if (amt.isIntegerOnly()) {
            formatter = new IntegerFormatter(min, max, true);
            value = Integer.valueOf(amt.getIntegerAmount());
            prototype = Integer.valueOf(max);
        } else {
            formatter = new DoubleFormatter(min, max, true);
            value = Double.valueOf(amt.getAmount());
            prototype = Double.valueOf(max + 0.25);
        }
        EditorField field = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, value, prototype, null);
        field.putClientProperty(LeveledAmount.class, amt);
        UIUtilities.setToPreferredSizeOnly(field);
        add(field);
        return field;
    }

    /**
     * @param amt       The current leveled amount object.
     * @param usePerDie Whether to use the "per die" message or the "per level" message.
     * @return The {@link JComboBox} that allows a {@link LeveledAmount} to be changed.
     */
    protected JComboBox<Object> addLeveledAmountCombo(LeveledAmount amt, boolean usePerDie) {
        String per = usePerDie ? I18n.Text("per die") : I18n.Text("per level");
        mLeveledAmountCombo = addComboBox(LeveledAmount.ATTRIBUTE_PER_LEVEL, new Object[]{" ", per}, amt.isPerLevel() ? per : " ");
        mLeveledAmountCombo.putClientProperty(LeveledAmount.class, amt);
        return mLeveledAmountCombo;
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
    public void actionPerformed(ActionEvent event) {
        String     command = event.getActionCommand();
        JComponent parent  = (JComponent) getParent();
        if (LeveledAmount.ATTRIBUTE_PER_LEVEL.equals(command)) {
            ((Bonus) mFeature).getAmount().setPerLevel(mLeveledAmountCombo.getSelectedIndex() == 1);
        } else if (CHANGE_BASE_TYPE.equals(command)) {
            LAST_FEATURE_TYPE = (FeatureType) mBaseTypeCombo.getSelectedItem();
            if (LAST_FEATURE_TYPE != null && !LAST_FEATURE_TYPE.matches(mFeature)) {
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
        } else {
            super.actionPerformed(event);
        }
    }

    @Override
    public void propertyChange(PropertyChangeEvent event) {
        if ("value".equals(event.getPropertyName())) {
            EditorField   field = (EditorField) event.getSource();
            LeveledAmount amt   = (LeveledAmount) field.getClientProperty(LeveledAmount.class);
            if (amt != null) {
                if (amt.isIntegerOnly()) {
                    amt.setAmount(((Integer) field.getValue()).intValue());
                } else {
                    amt.setAmount(((Double) field.getValue()).doubleValue());
                }
                notifyActionListeners();
            } else {
                super.propertyChange(event);
            }
        } else {
            super.propertyChange(event);
        }
    }
}
