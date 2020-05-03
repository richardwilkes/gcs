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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.Fixed4;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.NumberFilter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.units.Units;
import com.trollworks.gcs.utility.units.WeightValue;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTabbedPane;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** Editor for {@link EquipmentModifier}s. */
public class EquipmentModifierEditor extends RowEditor<EquipmentModifier> implements ActionListener, DocumentListener, FocusListener {
    private JTextField        mNameField;
    private JCheckBox         mEnabledField;
    private JTextField        mNotesField;
    private JTextField        mReferenceField;
    private FeaturesPanel     mFeatures;
    private JTabbedPane       mTabPanel;
    private JComboBox<Object> mCostType;
    private JTextField        mCostAmountField;
    private JComboBox<Object> mWeightType;
    private JTextField        mWeightAmountField;

    /**
     * Creates a new {@link EquipmentModifierEditor}.
     *
     * @param modifier The {@link EquipmentModifier} to edit.
     */
    public EquipmentModifierEditor(EquipmentModifier modifier) {
        super(modifier);

        JPanel     content = new JPanel(new ColumnLayout(2));
        JPanel     fields  = new JPanel(new ColumnLayout(2));
        JLabel     iconLabel;
        RetinaIcon icon    = modifier.getIcon(true);
        iconLabel = icon != null ? new JLabel(icon) : new JLabel();

        if (modifier.canHaveChildren()) {
            mNameField = createCorrectableField(fields, fields, I18n.Text("Name"), modifier.getName(), I18n.Text("Name of container"));
            mNotesField = createField(fields, fields, I18n.Text("Notes"), modifier.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this modifier"), 0);
            mReferenceField = createField(fields, fields, I18n.Text("Ref"), mRow.getReference(), I18n.Text("A reference to the book and page this modifier appears on (e.g. B22 would refer to \"Basic Set\", page 22)"), 6);
        } else {
            JPanel wrapper = new JPanel(new ColumnLayout(2));
            mNameField = createCorrectableField(fields, wrapper, I18n.Text("Name"), modifier.getName(), I18n.Text("Name of Modifier"));
            mEnabledField = new JCheckBox(I18n.Text("Enabled"), modifier.isEnabled());
            mEnabledField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Whether this modifier has been enabled or not")));
            mEnabledField.setEnabled(mIsEditable);
            wrapper.add(mEnabledField);
            fields.add(wrapper);

            createCostAdjustmentFields(fields);
            createWeightAdjustmentFields(fields);

            wrapper = new JPanel(new ColumnLayout(3));
            mNotesField = createField(fields, wrapper, I18n.Text("Notes"), modifier.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this modifier"), 0);
            mReferenceField = createField(wrapper, wrapper, I18n.Text("Ref"), mRow.getReference(), I18n.Text("A reference to the book and page this modifier appears on (e.g. B22 would refer to \"Basic Set\", page 22)"), 6);
            fields.add(wrapper);
        }

        iconLabel.setVerticalAlignment(SwingConstants.TOP);
        iconLabel.setAlignmentY(-1.0f);
        content.add(iconLabel);
        content.add(fields);
        add(content);

        if (!modifier.canHaveChildren()) {
            mTabPanel = new JTabbedPane();
            mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
            Component panel = embedEditor(mFeatures);
            mTabPanel.addTab(panel.getName(), panel);
            UIUtilities.selectTab(mTabPanel, getLastTabName());
            add(mTabPanel);
        }
    }

    @Override
    protected boolean applyChangesSelf() {
        boolean modified = mRow.setName(mNameField.getText());
        modified |= mRow.setReference(mReferenceField.getText());
        modified |= mRow.setNotes(mNotesField.getText());
        if (!mRow.canHaveChildren()) {
            modified |= mRow.setEnabled(mEnabledField.isSelected());
            if (mFeatures != null) {
                modified |= mRow.setFeatures(mFeatures.getFeatures());
            }
            modified |= mRow.setCostAdjType(getCostType());
            modified |= mRow.setCostAdjAmount(getCostAmount());
            modified |= mRow.setWeightAdjType(getWeightType());
            switch (mRow.getWeightAdjType()) {
            case BASE_ADDITION:
            case FINAL_ADDITION:
                modified |= mRow.setWeightAdjAddition(getWeightAddition());
                break;
            case MULTIPLIER:
            default:
                modified |= mRow.setWeightAdjMultiplier(getWeightMultiplier());
                break;
            }
        }
        return modified;
    }

    @Override
    public void finished() {
        if (mTabPanel != null) {
            updateLastTabName(mTabPanel.getTitleAt(mTabPanel.getSelectedIndex()));
        }
    }

    private JTextField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        field.getDocument().addDocumentListener(this);
        LinkedLabel label = new LinkedLabel(title);
        label.setLink(field);
        labelParent.add(label);
        fieldParent.add(field);
        return field;
    }

    private JTextField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
        JTextField field = new JTextField(maxChars > 0 ? Text.makeFiller(maxChars, 'M') : text);
        if (maxChars > 0) {
            UIUtilities.setToPreferredSizeOnly(field);
            field.setText(text);
        }
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
        return field;
    }

    private JScrollPane embedEditor(Container editor) {
        JScrollPane scrollPanel = new JScrollPane(editor);
        scrollPanel.setMinimumSize(new Dimension(500, 120));
        scrollPanel.setName(editor.toString());
        if (!mIsEditable) {
            UIUtilities.disableControls(editor);
        }
        return scrollPanel;
    }

    private void createCostAdjustmentFields(Container parent) {
        JPanel wrapper = new JPanel(new ColumnLayout(2));
        createCostAdjustmentField(parent, wrapper);
        createCostTypeCombo(wrapper);
        parent.add(wrapper);
    }

    private void createCostAdjustmentField(Container labelParent, Container fieldParent) {
        mCostAmountField = new JTextField("-999,999,999.00");
        UIUtilities.setToPreferredSizeOnly(mCostAmountField);
        mCostAmountField.setText(mRow.getCostAdjType().format(mRow.getCostAdjAmount()));
        mCostAmountField.setToolTipText(I18n.Text("The cost modifier"));
        mCostAmountField.setEnabled(mIsEditable);
        new NumberFilter(mCostAmountField, true, true, true, 11);
        mCostAmountField.addActionListener(this);
        labelParent.add(new LinkedLabel("", mCostAmountField));
        fieldParent.add(mCostAmountField);
    }

    private void createCostTypeCombo(Container parent) {
        EquipmentModifierCostType[] types = EquipmentModifierCostType.values();
        mCostType = new JComboBox<>(types);
        mCostType.setSelectedItem(mRow.getCostAdjType());
        mCostType.addActionListener(this);
        mCostType.setMaximumRowCount(types.length);
        UIUtilities.setToPreferredSizeOnly(mCostType);
        parent.add(mCostType);
    }

    private EquipmentModifierCostType getCostType() {
        Object obj = mCostType.getSelectedItem();
        if (!(obj instanceof EquipmentModifierCostType)) {
            obj = EquipmentModifierCostType.COST_FACTOR;
        }
        return (EquipmentModifierCostType) obj;
    }

    private Fixed4 getCostAmount() {
        return getCostType().extract(mCostAmountField.getText(), true);
    }

    private void createWeightAdjustmentFields(Container parent) {
        JPanel wrapper = new JPanel(new ColumnLayout(2));
        createWeightAdjustmentField(parent, wrapper);
        createWeightTypeCombo(wrapper);
        parent.add(wrapper);
    }

    private void createWeightAdjustmentField(Container labelParent, Container fieldParent) {
        mWeightAmountField = new JTextField("-999,999,999.00");
        UIUtilities.setToPreferredSizeOnly(mWeightAmountField);
        switch (mRow.getWeightAdjType()) {
        case BASE_ADDITION:
        case FINAL_ADDITION:
            String addition = mRow.getWeightAdjAddition().toString();
            if (!addition.startsWith("-")) {
                addition = "+" + addition;
            }
            mWeightAmountField.setText(addition);
            break;
        case MULTIPLIER:
        default:
            mWeightAmountField.setText(Numbers.format(mRow.getWeightAdjMultiplier()));
            break;
        }
        mWeightAmountField.setToolTipText(I18n.Text("The weight modifier"));
        mWeightAmountField.setEnabled(mIsEditable);
        mWeightAmountField.addActionListener(this);
        mWeightAmountField.addFocusListener(this);
        labelParent.add(new LinkedLabel("", mWeightAmountField));
        fieldParent.add(mWeightAmountField);
    }

    private void createWeightTypeCombo(Container parent) {
        EquipmentModifierWeightType[] types = EquipmentModifierWeightType.values();
        mWeightType = new JComboBox<>(types);
        mWeightType.setSelectedItem(mRow.getWeightAdjType());
        mWeightType.addActionListener(this);
        mWeightType.setMaximumRowCount(types.length);
        UIUtilities.setToPreferredSizeOnly(mWeightType);
        parent.add(mWeightType);
    }

    private EquipmentModifierWeightType getWeightType() {
        Object obj = mWeightType.getSelectedItem();
        if (!(obj instanceof EquipmentModifierWeightType)) {
            obj = EquipmentModifierWeightType.MULTIPLIER;
        }
        return (EquipmentModifierWeightType) obj;
    }

    private double getWeightMultiplier() {
        double value = Numbers.extractDouble(mWeightAmountField.getText(), 0, true);
        if (value <= 0 && getWeightType() == EquipmentModifierWeightType.MULTIPLIER) {
            value = 1;
        }
        return value;
    }

    private WeightValue getWeightAddition() {
        return WeightValue.extract(mWeightAmountField.getText(), true);
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

    @Override
    public void changedUpdate(DocumentEvent event) {
        nameChanged();
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        nameChanged();
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        nameChanged();
    }

    private void nameChanged() {
        LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.Text("The name field may not be empty") : null);
    }

    @Override
    public void focusGained(FocusEvent event) {
        // Not used.
    }

    @Override
    public void focusLost(FocusEvent event) {
        if (!mRow.canHaveChildren()) {
            weightChanged();
        }
    }

    private void weightChanged() {
        String      text   = mWeightAmountField.getText().trim();
        WeightValue weight = WeightValue.extract(text, true);
        if (getWeightType() == EquipmentModifierWeightType.MULTIPLIER) {
            double value = weight.getValue();
            if (value <= 0) {
                mWeightAmountField.setText("1");
            } else {
                if (text.startsWith("+")) {
                    text = text.substring(1);
                }
                if (text.toLowerCase().endsWith(weight.getUnits().getAbbreviation().toLowerCase())) {
                    text = text.substring(0, text.length() - weight.getUnits().getAbbreviation().length()).trim();
                }
                mWeightAmountField.setText(text);
            }
        } else {
            String  lowered = text.toLowerCase();
            boolean found   = false;
            for (Units unit : weight.getUnits().getCompatibleUnits()) {
                if (lowered.endsWith(unit.getAbbreviation().toLowerCase())) {
                    found = true;
                    break;
                }
            }
            if (!found) {
                text += " " + weight.getUnits().getAbbreviation();
            }
            if (!text.startsWith("+") && !text.startsWith("-")) {
                text = "+" + text;
            }
            mWeightAmountField.setText(text);
        }
    }

    private void costChanged() {
        String text    = mCostAmountField.getText().trim();
        String revised = getCostType().adjustText(text);
        if (!text.equals(revised)) {
            mCostAmountField.setText(revised);
        }
    }
}
