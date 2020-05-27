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
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

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
    private JTextField        mTechLevelField;
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
            JPanel wrapper = new JPanel(new ColumnLayout(4));
            mNameField = createCorrectableField(fields, wrapper, I18n.Text("Name"), modifier.getName(), I18n.Text("Name of Modifier"));
            mTechLevelField = createField(wrapper, wrapper, I18n.Text("Tech Level"), mRow.getTechLevel(), I18n.Text("The first Tech Level this equipment is available at"), 3);
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
            modified |= mRow.setWeightAdjAmount(weightType.format(mWeightAmountField.getText(), mRow.getDataFile().defaultWeightUnits(), false));
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
        mCostAmountField.setText(mRow.getCostAdjType().format(mRow.getCostAdjAmount(), true));
        mCostAmountField.setToolTipText(I18n.Text("The cost modifier"));
        mCostAmountField.setEnabled(mIsEditable);
        mCostAmountField.addFocusListener(this);
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
            obj = EquipmentModifierCostType.TO_ORIGINAL_COST;
        }
        return (EquipmentModifierCostType) obj;
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
        mWeightAmountField.setText(mRow.getWeightAdjType().format(mRow.getWeightAdjAmount(), mRow.getDataFile().defaultWeightUnits(), true));
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
        String revised = getWeightType().format(text, mRow.getDataFile().defaultWeightUnits(), true);
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
