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
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.StdLabel;
import com.trollworks.gcs.ui.widget.StdPanel;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Container;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** Editor for {@link EquipmentModifier}s. */
public class EquipmentModifierEditor extends RowEditor<EquipmentModifier> implements ActionListener, DocumentListener, FocusListener {
    private JTextField         mNameField;
    private JTextField         mTechLevelField;
    private JCheckBox          mEnabledField;
    private MultiLineTextField mNotesField;
    private JTextField         mReferenceField;
    private FeaturesPanel      mFeatures;
    private JComboBox<Object>  mCostType;
    private JTextField         mCostAmountField;
    private JComboBox<Object>  mWeightType;
    private JTextField         mWeightAmountField;

    /**
     * Creates a new EquipmentModifierEditor.
     *
     * @param modifier The {@link EquipmentModifier} to edit.
     */
    public EquipmentModifierEditor(EquipmentModifier modifier) {
        super(modifier);
        addContent();
    }

    @Override
    protected void addContentSelf(ScrollContent outer) {
        StdPanel panel = new StdPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        if (mRow.canHaveChildren()) {
            mNameField = createCorrectableField(panel, panel, I18n.text("Name"), mRow.getName(), I18n.text("Name of container"));
        } else {
            StdPanel wrapper = new StdPanel(new PrecisionLayout().setMargins(0).setColumns(4));
            mNameField = createCorrectableField(panel, wrapper, I18n.text("Name"), mRow.getName(), I18n.text("Name of Modifier"));
            mTechLevelField = createField(wrapper, wrapper, I18n.text("Tech Level"), mRow.getTechLevel(), I18n.text("The first Tech Level this equipment is available at"), 3);
            mEnabledField = new JCheckBox(I18n.text("Enabled"), mRow.isEnabled());
            mEnabledField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.text("Whether this modifier has been enabled or not")));
            wrapper.add(mEnabledField);
            panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

            createCostAdjustmentFields(panel);
            createWeightAdjustmentFields(panel);
        }

        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.text("Any notes that you would like to show up in the list along with this modifier"), this);
        panel.add(new StdLabel(I18n.text("Notes"), mNotesField), new PrecisionLayoutData().setFillHorizontalAlignment().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2));
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        mReferenceField = createField(panel, panel, I18n.text("Ref"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.text("equipment modifier")), 6);
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
            modified |= mRow.setTechLevel(mTechLevelField.getText());
            modified |= mRow.setEnabled(mEnabledField.isSelected());
            if (mFeatures != null) {
                modified |= mRow.setFeatures(mFeatures.getFeatures());
            }
            EquipmentModifierCostType costType = getCostType();
            modified |= mRow.setCostAdjType(costType);
            modified |= mRow.setCostAdjAmount(costType.format(mCostAmountField.getText(), false));
            EquipmentModifierWeightType weightType = getWeightType();
            modified |= mRow.setWeightAdjType(weightType);
            modified |= mRow.setWeightAdjAmount(weightType.format(mWeightAmountField.getText(), mRow.getDataFile().getSheetSettings().defaultWeightUnits(), false));
        }
        return modified;
    }

    private JTextField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        addLabel(labelParent, title, field);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
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

    private void createCostAdjustmentFields(Container parent) {
        StdPanel wrapper = new StdPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        createCostAdjustmentField(parent, wrapper);
        createCostTypeCombo(wrapper);
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void createCostAdjustmentField(Container labelParent, Container fieldParent) {
        mCostAmountField = new JTextField("-999,999,999.00");
        UIUtilities.setToPreferredSizeOnly(mCostAmountField);
        mCostAmountField.setText(mRow.getCostAdjType().format(mRow.getCostAdjAmount(), true));
        mCostAmountField.setToolTipText(I18n.text("The cost modifier"));
        mCostAmountField.addFocusListener(this);
        mCostAmountField.addActionListener(this);
        addLabel(labelParent, "", mCostAmountField);
        fieldParent.add(mCostAmountField);
    }

    private void createCostTypeCombo(Container parent) {
        EquipmentModifierCostType[] types = EquipmentModifierCostType.values();
        mCostType = new JComboBox<>(types);
        mCostType.setSelectedItem(mRow.getCostAdjType());
        mCostType.addActionListener(this);
        mCostType.setMaximumRowCount(types.length);
        parent.add(mCostType);
    }

    private EquipmentModifierCostType getCostType() {
        Object obj = mCostType.getSelectedItem();
        if (!(obj instanceof EquipmentModifierCostType)) {
            obj = EquipmentModifierCostType.TO_ORIGINAL_COST;
        }
        return (EquipmentModifierCostType) obj;
    }

    private void createWeightAdjustmentFields(Container parent) {
        StdPanel wrapper = new StdPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        createWeightAdjustmentField(parent, wrapper);
        createWeightTypeCombo(wrapper);
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void createWeightAdjustmentField(Container labelParent, Container fieldParent) {
        mWeightAmountField = new JTextField("-999,999,999.00");
        UIUtilities.setToPreferredSizeOnly(mWeightAmountField);
        mWeightAmountField.setText(mRow.getWeightAdjType().format(mRow.getWeightAdjAmount(), mRow.getDataFile().getSheetSettings().defaultWeightUnits(), true));
        mWeightAmountField.setToolTipText(I18n.text("The weight modifier"));
        mWeightAmountField.addActionListener(this);
        mWeightAmountField.addFocusListener(this);
        labelParent.add(new StdLabel("", mWeightAmountField));
        fieldParent.add(mWeightAmountField);
    }

    private void createWeightTypeCombo(Container parent) {
        EquipmentModifierWeightType[] types = EquipmentModifierWeightType.values();
        mWeightType = new JComboBox<>(types);
        mWeightType.setSelectedItem(mRow.getWeightAdjType());
        mWeightType.addActionListener(this);
        mWeightType.setMaximumRowCount(types.length);
        parent.add(mWeightType);
    }

    private EquipmentModifierWeightType getWeightType() {
        Object obj = mWeightType.getSelectedItem();
        if (!(obj instanceof EquipmentModifierWeightType)) {
            obj = EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT;
        }
        return (EquipmentModifierWeightType) obj;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (!mRow.canHaveChildren()) {
            Object src = event.getSource();
            if (src == mCostAmountField || src == mCostType) {
                costChanged();
            } else if (src == mWeightAmountField || src == mWeightType) {
                weightChanged();
            }
        }
    }

    private void docChanged(DocumentEvent event) {
        if (mNameField.getDocument() == event.getDocument()) {
            StdLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.text("The name field may not be empty") : null);
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

    @Override
    public void focusGained(FocusEvent event) {
        // Not used.
    }

    @Override
    public void focusLost(FocusEvent event) {
        if (!mRow.canHaveChildren()) {
            Object src = event.getSource();
            if (src == mCostAmountField) {
                costChanged();
            } else if (src == mWeightAmountField) {
                weightChanged();
            }
        }
    }

    private void weightChanged() {
        String text    = mWeightAmountField.getText();
        String revised = getWeightType().format(text, mRow.getDataFile().getSheetSettings().defaultWeightUnits(), true);
        if (!text.equals(revised)) {
            mWeightAmountField.setText(revised);
        }
    }

    private void costChanged() {
        String text    = mCostAmountField.getText();
        String revised = getCostType().format(text, true);
        if (!text.equals(revised)) {
            mCostAmountField.setText(revised);
        }
    }
}
