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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.NumberFilter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Container;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.text.MessageFormat;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JPanel;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** Editor for {@link AdvantageModifier}s. */
public class AdvantageModifierEditor extends RowEditor<AdvantageModifier> implements ActionListener, DocumentListener {
    private JTextField         mNameField;
    private JCheckBox          mEnabledField;
    private MultiLineTextField mNotesField;
    private JTextField         mReferenceField;
    private JTextField         mCostField;
    private JTextField         mLevelField;
    private JTextField         mCostModifierField;
    private FeaturesPanel      mFeatures;
    private JComboBox<Object>  mCostType;
    private JComboBox<Object>  mAffects;
    private int                mLastLevel;

    /**
     * Creates a new AdvantageModifierEditor.
     *
     * @param modifier The {@link AdvantageModifier} to edit.
     */
    public AdvantageModifierEditor(AdvantageModifier modifier) {
        super(modifier);
        addContent();
    }

    @Override
    protected void addContentSelf(ScrollContent outer) {
        JPanel panel = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        if (mRow.canHaveChildren()) {
            mNameField = createCorrectableField(panel, panel, I18n.text("Name"), mRow.getName(), I18n.text("Name of container"));
        } else {
            JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2));
            mNameField = createCorrectableField(panel, wrapper, I18n.text("Name"), mRow.getName(), I18n.text("Name of Modifier"));
            mEnabledField = new JCheckBox(I18n.text("Enabled"), mRow.isEnabled());
            mEnabledField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.text("Whether this modifier has been enabled or not")));
            mEnabledField.setEnabled(mIsEditable);
            wrapper.add(mEnabledField);
            panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

            createCostModifierFields(panel);
        }

        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.text("Any notes that you would like to show up in the list along with this modifier"), this);
        panel.add(new LinkedLabel(I18n.text("Notes"), mNotesField), new PrecisionLayoutData().setFillHorizontalAlignment().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2));
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        mReferenceField = createField(panel, panel, I18n.text("Ref"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.text("advantage modifier")), 6);
        outer.add(panel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        if (!mRow.canHaveChildren()) {
            mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
            addSection(outer, mFeatures);
        }
    }

    @Override
    protected boolean applyChangesSelf() {
        boolean modified = mRow.setName(mNameField.getText());
        modified |= mRow.setReference(mReferenceField.getText());
        modified |= mRow.setNotes(mNotesField.getText());
        if (!mRow.canHaveChildren()) {
            modified |= getCostType() == AdvantageModifierCostType.MULTIPLIER ? mRow.setCostMultiplier(getCostMultiplier()) : mRow.setCost(getCost());
            if (hasLevels()) {
                modified |= mRow.setLevels(getLevels());
                modified |= mRow.setCostType(AdvantageModifierCostType.PERCENTAGE);
            } else {
                modified |= mRow.setLevels(0);
                modified |= mRow.setCostType(getCostType());
            }
            modified |= mRow.setAffects((Affects) mAffects.getSelectedItem());
            modified |= mRow.setEnabled(mEnabledField.isSelected());
            if (mFeatures != null) {
                modified |= mRow.setFeatures(mFeatures.getFeatures());
            }
        }
        return modified;
    }

    private boolean hasLevels() {
        return mCostType.getSelectedIndex() == 0;
    }

    private JTextField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        addLabel(labelParent, title, field);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (!mRow.canHaveChildren()) {
            Object src = event.getSource();
            if (src == mCostType) {
                costTypeChanged();
            } else if (src == mCostField || src == mLevelField) {
                updateCostModifier();
            }
        }
    }

    private static JTextField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
        JTextField field = new JTextField(maxChars > 0 ? Text.makeFiller(maxChars, 'M') : text);
        if (maxChars > 0) {
            UIUtilities.setToPreferredSizeOnly(field);
            field.setText(text);
        }
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        addLabel(labelParent, title, field);
        PrecisionLayoutData ld = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (maxChars == 0) {
            ld.setGrabHorizontalSpace(true);
        }
        fieldParent.add(field, ld);
        return field;
    }

    private JTextField createNumberField(Container labelParent, Container fieldParent, String title, boolean allowSign, int value, String tooltip, int maxDigits) {
        JTextField field = new JTextField(Text.makeFiller(maxDigits, '9') + Text.makeFiller(maxDigits / 3, ',') + (allowSign ? "-" : ""));
        UIUtilities.setToPreferredSizeOnly(field);
        field.setText(Numbers.format(value));
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        NumberFilter.apply(field, false, allowSign, true, maxDigits);
        field.addActionListener(this);
        addLabel(labelParent, title, field);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    private JTextField createNumberField(Container labelParent, Container fieldParent, String title, double value, String tooltip, int maxDigits) {
        JTextField field = new JTextField(Text.makeFiller(maxDigits, '9') + Text.makeFiller(maxDigits / 3, ',') + '.');
        UIUtilities.setToPreferredSizeOnly(field);
        field.setText(Numbers.format(value));
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        NumberFilter.apply(field, true, false, true, maxDigits);
        field.addActionListener(this);
        addLabel(labelParent, title, field);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    private void createCostModifierFields(Container parent) {
        JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(7));
        mLastLevel = mRow.getLevels();
        if (mLastLevel < 1) {
            mLastLevel = 1;
        }
        String costTitle   = I18n.text("Cost");
        String costTooltip = I18n.text("The base cost modifier");
        mCostField = mRow.getCostType() == AdvantageModifierCostType.MULTIPLIER ? createNumberField(parent, wrapper, costTitle, mRow.getCostMultiplier(), costTooltip, 5) : createNumberField(parent, wrapper, costTitle, true, mRow.getCost(), costTooltip, 5);
        createCostType(wrapper);
        mLevelField = createNumberField(wrapper, wrapper, I18n.text("Levels"), false, mLastLevel, I18n.text("The number of levels this modifier has"), 3);
        mCostModifierField = createNumberField(wrapper, wrapper, I18n.text("Total"), true, 0, I18n.text("The cost modifier's total value"), 9);
        mAffects = createComboBox(wrapper, Affects.values(), mRow.getAffects());
        mCostModifierField.setEnabled(false);
        if (!mRow.hasLevels()) {
            mLevelField.setText("");
            mLevelField.setEnabled(false);
        }
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private JComboBox<Object> createComboBox(Container parent, Object[] items, Object selection) {
        JComboBox<Object> combo = new JComboBox<>(items);
        combo.setSelectedItem(selection);
        combo.addActionListener(this);
        combo.setMaximumRowCount(items.length);
        UIUtilities.setToPreferredSizeOnly(combo);
        parent.add(combo, new PrecisionLayoutData().setFillHorizontalAlignment());
        return combo;
    }

    private void createCostType(Container parent) {
        AdvantageModifierCostType[] types  = AdvantageModifierCostType.values();
        Object[]                    values = new Object[types.length + 1];
        values[0] = MessageFormat.format(I18n.text("{0} Per Level"), AdvantageModifierCostType.PERCENTAGE.toString());
        System.arraycopy(types, 0, values, 1, types.length);
        mCostType = createComboBox(parent, values, mRow.hasLevels() ? values[0] : mRow.getCostType());
    }

    private void costTypeChanged() {
        boolean hasLevels = hasLevels();
        if (hasLevels) {
            mLevelField.setText(Numbers.format(mLastLevel));
        } else {
            mLastLevel = Numbers.extractInteger(mLevelField.getText(), 0, true);
            mLevelField.setText("");
        }
        mLevelField.setEnabled(hasLevels);
        updateCostField();
        updateCostModifier();
    }

    private void updateCostField() {
        if (getCostType() == AdvantageModifierCostType.MULTIPLIER) {
            NumberFilter.apply(mCostField, true, false, true, 5);
            mCostField.setText(Numbers.format(Math.abs(Numbers.extractDouble(mCostField.getText(), 0, true))));
        } else {
            NumberFilter.apply(mCostField, false, true, true, 5);
            mCostField.setText(Numbers.formatWithForcedSign(Numbers.extractInteger(mCostField.getText(), 0, true)));
        }
    }

    private void updateCostModifier() {
        boolean enabled = true;
        if (hasLevels()) {
            mCostModifierField.setText(Numbers.formatWithForcedSign((long) getCost() * getLevels()) + "%");
        } else {
            AdvantageModifierCostType costType = getCostType();
            switch (costType) {
            case POINTS -> mCostModifierField.setText(Numbers.formatWithForcedSign(getCost()));
            case MULTIPLIER -> {
                mCostModifierField.setText(costType + Numbers.format(getCostMultiplier()));
                mAffects.setSelectedItem(Affects.TOTAL);
                enabled = false;
            }
            default -> mCostModifierField.setText(Numbers.formatWithForcedSign(getCost()) + costType);
            }
        }
        mAffects.setEnabled(mIsEditable && enabled);
    }

    private AdvantageModifierCostType getCostType() {
        Object obj = mCostType.getSelectedItem();
        if (!(obj instanceof AdvantageModifierCostType)) {
            obj = AdvantageModifierCostType.PERCENTAGE;
        }
        return (AdvantageModifierCostType) obj;
    }

    private int getCost() {
        return Numbers.extractInteger(mCostField.getText(), 0, true);
    }

    private double getCostMultiplier() {
        return Numbers.extractDouble(mCostField.getText(), 0, true);
    }

    private int getLevels() {
        return Numbers.extractInteger(mLevelField.getText(), 0, true);
    }

    private void docChanged(DocumentEvent event) {
        if (mNameField.getDocument() == event.getDocument()) {
            LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.text("The name field may not be empty") : null);
        }
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        docChanged(event);
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        docChanged(event);
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        docChanged(event);
    }
}
