/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.widget.LinkedLabel;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.text.NumberFilter;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.Text;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
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
public class EquipmentModifierEditor extends RowEditor<EquipmentModifier> implements ActionListener, DocumentListener {
    private JTextField        mNameField;
    private JCheckBox         mEnabledField;
    private JTextField        mNotesField;
    private JTextField        mReferenceField;
    private FeaturesPanel     mFeatures;
    private JTabbedPane       mTabPanel;
    private JComboBox<Object> mCostType;
    private JTextField        mCostAmountField;

    /**
     * Creates a new {@link EquipmentModifierEditor}.
     *
     * @param modifier The {@link EquipmentModifier} to edit.
     */
    public EquipmentModifierEditor(EquipmentModifier modifier) {
        super(modifier);

        JPanel   content = new JPanel(new ColumnLayout(2));
        JPanel   fields  = new JPanel(new ColumnLayout(2));
        JLabel   icon;
        StdImage image   = modifier.getIcon(true);
        icon = image != null ? new JLabel(image) : new JLabel();

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
        mReferenceField = createField(wrapper, wrapper, I18n.Text("Ref"), mRow.getReference(), I18n.Text("A reference to the book and page this modifier appears on (e.g. B22 would refer to \"Basic Set\", page 22)"), 6);
        fields.add(wrapper);

        icon.setVerticalAlignment(SwingConstants.TOP);
        icon.setAlignmentY(-1.0f);
        content.add(icon);
        content.add(fields);
        add(content);

        mTabPanel = new JTabbedPane();
        mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
        Component panel = embedEditor(mFeatures);
        mTabPanel.addTab(panel.getName(), panel);
        UIUtilities.selectTab(mTabPanel, getLastTabName());
        add(mTabPanel);
    }

    @Override
    protected boolean applyChangesSelf() {
        boolean modified = mRow.setName(mNameField.getText());
        modified |= mRow.setReference(mReferenceField.getText());
        modified |= mRow.setNotes(mNotesField.getText());
        modified |= mRow.setEnabled(mEnabledField.isSelected());
        if (mFeatures != null) {
            modified |= mRow.setFeatures(mFeatures.getFeatures());
        }
        modified |= mRow.setCostType(getCostType());
        modified |= mRow.setCostAmount(getCostAmount());
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

    private void createCostModifierFields(Container parent) {
        JPanel wrapper = new JPanel(new ColumnLayout(2));
        createNumberField(parent, wrapper);
        createCostType(wrapper);
        parent.add(wrapper);
    }

    private void createNumberField(Container labelParent, Container fieldParent) {
        mCostAmountField = new JTextField("-999,999,999.00");
        UIUtilities.setToPreferredSizeOnly(mCostAmountField);
        double amt = mRow.getCostAmount();
        mCostAmountField.setText(mRow.getCostType().isMultiplier() ? Numbers.format(amt) : Numbers.formatWithForcedSign(amt));
        mCostAmountField.setToolTipText(I18n.Text("The cost modifier"));
        mCostAmountField.setEnabled(mIsEditable);
        new NumberFilter(mCostAmountField, true, true, true, 11);
        mCostAmountField.addActionListener(this);
        labelParent.add(new LinkedLabel("", mCostAmountField));
        fieldParent.add(mCostAmountField);
    }

    private void createCostType(Container parent) {
        EquipmentModifierCostType[] types = EquipmentModifierCostType.values();
        mCostType = new JComboBox<>(types);
        mCostType.setSelectedItem(mRow.getCostType());
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

    private double getCostAmount() {
        double value = Numbers.extractDouble(mCostAmountField.getText(), 0, true);
        if (value <= 0 && getCostType().isMultiplier()) {
            value = 1;
        }
        return value;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String text = mCostAmountField.getText().trim();
        if (getCostType().isMultiplier()) {
            double value = Numbers.extractDouble(text, 0, true);
            if (value <= 0) {
                mCostAmountField.setText("1");
            } else if (text.startsWith("+")) {
                mCostAmountField.setText(text.substring(1));
            }
        } else if (!text.startsWith("+") && !text.startsWith("-")) {
            mCostAmountField.setText("+" + text);
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
}
