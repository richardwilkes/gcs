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

import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.NumberFilter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.text.MessageFormat;
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

/** Editor for {@link AdvantageModifier}s. */
public class AdvantageModifierEditor extends RowEditor<AdvantageModifier> implements ActionListener, DocumentListener {
    private JTextField        mNameField;
    private JCheckBox         mEnabledField;
    private JTextField        mNotesField;
    private JTextField        mReferenceField;
    private JTextField        mCostField;
    private JTextField        mLevelField;
    private JTextField        mCostModifierField;
    private FeaturesPanel     mFeatures;
    private JTabbedPane       mTabPanel;
    private JComboBox<Object> mCostType;
    private JComboBox<Object> mAffects;
    private int               mLastLevel;

    /**
     * Creates a new {@link AdvantageModifierEditor}.
     *
     * @param modifier The {@link AdvantageModifier} to edit.
     */
    public AdvantageModifierEditor(AdvantageModifier modifier) {
        super(modifier);

        JPanel     content = new JPanel(new ColumnLayout(2));
        JPanel     fields  = new JPanel(new ColumnLayout(2));
        JLabel     iconLabel;
        RetinaIcon icon    = modifier.getIcon(true);
        iconLabel = icon != null ? new JLabel(icon) : new JLabel();

        if (modifier.canHaveChildren()) {
            mNameField = createCorrectableField(fields, fields, I18n.Text("Name"), modifier.getName(), I18n.Text("Name of container"));
            mNotesField = createField(fields, fields, I18n.Text("Notes"), modifier.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this modifier"), 0);
            mReferenceField = createField(fields, fields, I18n.Text("Ref"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.Text("advantage modifier")), 6);
        } else {
            JPanel wrapper = new JPanel(new ColumnLayout(2));
            mNameField = createCorrectableField(fields, wrapper, I18n.Text("Name"), modifier.getName(), I18n.Text("Name of Modifier"));
            mEnabledField = new JCheckBox(I18n.Text("Enabled"), modifier.isEnabled());
            mEnabledField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Whether this modifier has been enabled or not")));
            mEnabledField.setEnabled(mIsEditable);
            wrapper.add(mEnabledField);
            fields.add(wrapper);

            createCostModifierFields(fields);

            wrapper = new JPanel(new ColumnLayout(3));
            mNotesField = createField(fields, wrapper, I18n.Text("Notes"), modifier.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this modifier"), 0);
            mReferenceField = createField(wrapper, wrapper, I18n.Text("Ref"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.Text("advantage modifier")), 6);
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

    private JTextField createNumberField(Container labelParent, Container fieldParent, String title, boolean allowSign, int value, String tooltip, int maxDigits) {
        JTextField field = new JTextField(Text.makeFiller(maxDigits, '9') + Text.makeFiller(maxDigits / 3, ',') + (allowSign ? "-" : ""));
        UIUtilities.setToPreferredSizeOnly(field);
        field.setText(Numbers.format(value));
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        NumberFilter.apply(field, false, allowSign, true, maxDigits);
        field.addActionListener(this);
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
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
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
        return field;
    }

    private void createCostModifierFields(Container parent) {
        JPanel wrapper = new JPanel(new ColumnLayout(7));
        mLastLevel = mRow.getLevels();
        if (mLastLevel < 1) {
            mLastLevel = 1;
        }
        String costTitle   = I18n.Text("Cost");
        String costTooltip = I18n.Text("The base cost modifier");
        mCostField = mRow.getCostType() == AdvantageModifierCostType.MULTIPLIER ? createNumberField(parent, wrapper, costTitle, mRow.getCostMultiplier(), costTooltip, 5) : createNumberField(parent, wrapper, costTitle, true, mRow.getCost(), costTooltip, 5);
        createCostType(wrapper);
        mLevelField = createNumberField(wrapper, wrapper, I18n.Text("Levels"), false, mLastLevel, I18n.Text("The number of levels this modifier has"), 3);
        mCostModifierField = createNumberField(wrapper, wrapper, I18n.Text("Total"), true, 0, I18n.Text("The cost modifier's total value"), 9);
        mAffects = createComboBox(wrapper, Affects.values(), mRow.getAffects());
        mCostModifierField.setEnabled(false);
        if (!mRow.hasLevels()) {
            mLevelField.setText("");
            mLevelField.setEnabled(false);
        }
        parent.add(wrapper);
    }

    private JComboBox<Object> createComboBox(Container parent, Object[] items, Object selection) {
        JComboBox<Object> combo = new JComboBox<>(items);
        combo.setSelectedItem(selection);
        combo.addActionListener(this);
        combo.setMaximumRowCount(items.length);
        UIUtilities.setToPreferredSizeOnly(combo);
        parent.add(combo);
        return combo;
    }

    private void createCostType(Container parent) {
        AdvantageModifierCostType[] types  = AdvantageModifierCostType.values();
        Object[]                    values = new Object[types.length + 1];
        values[0] = MessageFormat.format(I18n.Text("{0} Per Level"), AdvantageModifierCostType.PERCENTAGE.toString());
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
}
