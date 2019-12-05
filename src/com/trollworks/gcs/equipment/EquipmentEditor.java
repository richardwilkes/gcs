/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.widget.LinkedLabel;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.text.NumberFilter;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.Text;
import com.trollworks.toolkit.utility.units.WeightValue;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.util.ArrayList;

import javax.swing.JCheckBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTabbedPane;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** The detailed editor for {@link Equipment}s. */
public class EquipmentEditor extends RowEditor<Equipment> implements ActionListener, DocumentListener, FocusListener {
    private JCheckBox          mEquippedCheckBox;
    private JTextField         mDescriptionField;
    private JTextField         mTechLevelField;
    private JTextField         mLegalityClassField;
    private JTextField         mQtyField;
    private JTextField         mUsesField;
    private JTextField         mMaxUsesField;
    private JTextField         mValueField;
    private JTextField         mExtValueField;
    private JTextField         mWeightField;
    private JTextField         mExtWeightField;
    private JTextField         mNotesField;
    private JTextField         mCategoriesField;
    private JTextField         mReferenceField;
    private JTabbedPane        mTabPanel;
    private PrereqsPanel       mPrereqs;
    private FeaturesPanel      mFeatures;
    private MeleeWeaponEditor  mMeleeWeapons;
    private RangedWeaponEditor mRangedWeapons;
    private double             mContainedValue;
    private WeightValue        mContainedWeight;
    private boolean            mCarried;

    /**
     * Creates a new {@link Equipment} editor.
     *
     * @param equipment The {@link Equipment} to edit.
     * @param carried   <code>true</code> for the carried equipment, <code>false</code> for the
     *                  other equipment.
     */
    public EquipmentEditor(Equipment equipment, boolean carried) {
        super(equipment);
        mCarried = carried;

        JPanel content = new JPanel(new ColumnLayout(2));
        JPanel fields  = new JPanel(new ColumnLayout(2));
        JLabel icon    = new JLabel(equipment.getIcon(true));

        mDescriptionField = createCorrectableField(fields, I18n.Text("Name"), equipment.getDescription(), I18n.Text("The name/description of the equipment, without any notes"));
        createSecondLineFields(fields);
        createValueAndWeightFields(fields);
        mNotesField      = createField(fields, fields, I18n.Text("Notes"), equipment.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this equipment"), 0);
        mCategoriesField = createField(fields, fields, I18n.Text("Categories"), equipment.getCategoriesAsString(), I18n.Text("The category or categories the equipment belongs to (separate multiple categories with a comma)"), 0);

        boolean forCharacterOrTemplate = equipment.getCharacter() != null || equipment.getTemplate() != null;
        JPanel  wrapper                = new JPanel(new ColumnLayout(forCharacterOrTemplate ? 5 : 3));
        if (forCharacterOrTemplate) {
            mUsesField    = createIntegerNumberField(fields, wrapper, I18n.Text("Uses"), mRow.getUses(), I18n.Text("The number of uses remaining for this equipment"), 5);
            mMaxUsesField = createIntegerNumberField(wrapper, wrapper, I18n.Text("Max Uses"), mRow.getMaxUses(), I18n.Text("The maximum number of uses for this equipment"), 5);
        } else {
            mMaxUsesField = createIntegerNumberField(fields, wrapper, I18n.Text("Max Uses"), mRow.getMaxUses(), I18n.Text("The maximum number of uses for this equipment"), 5);
        }
        mReferenceField = createField(wrapper, wrapper, I18n.Text("Page Reference"), mRow.getReference(), I18n.Text("A reference to the book and page this equipment appears on (e.g. B22 would refer to \"Basic Set\", page 22)"), 6);
        fields.add(wrapper);

        icon.setVerticalAlignment(SwingConstants.TOP);
        icon.setAlignmentY(-1f);
        content.add(icon);
        content.add(fields);
        add(content);

        mTabPanel      = new JTabbedPane();
        mPrereqs       = new PrereqsPanel(mRow, mRow.getPrereqs());
        mFeatures      = new FeaturesPanel(mRow, mRow.getFeatures());
        mMeleeWeapons  = MeleeWeaponEditor.createEditor(mRow);
        mRangedWeapons = RangedWeaponEditor.createEditor(mRow);
        mTabPanel.addTab(mMeleeWeapons.getName(), mMeleeWeapons);
        mTabPanel.addTab(mRangedWeapons.getName(), mRangedWeapons);
        Component panel = embedEditor(mPrereqs);
        mTabPanel.addTab(panel.getName(), panel);
        panel = embedEditor(mFeatures);
        mTabPanel.addTab(panel.getName(), panel);
        if (!mIsEditable) {
            UIUtilities.disableControls(mMeleeWeapons);
            UIUtilities.disableControls(mRangedWeapons);
        }
        UIUtilities.selectTab(mTabPanel, getLastTabName());
        add(mTabPanel);
    }

    private boolean showEquipmentState() {
        return mCarried && mRow.getCharacter() != null;
    }

    private void createSecondLineFields(Container parent) {
        boolean isContainer = mRow.canHaveChildren();
        JPanel  wrapper     = new JPanel(new ColumnLayout((isContainer ? 4 : 6) + (showEquipmentState() ? 1 : 0)));

        if (!isContainer) {
            mQtyField = createIntegerNumberField(parent, wrapper, I18n.Text("Quantity"), mRow.getQuantity(), I18n.Text("The number of this equipment present"), 9);
        }
        mTechLevelField     = createField(isContainer ? parent : wrapper, wrapper, I18n.Text("Tech Level"), mRow.getTechLevel(), I18n.Text("The first Tech Level this equipment is available at"), 3);
        mLegalityClassField = createField(wrapper, wrapper, I18n.Text("Legality Class"), mRow.getLegalityClass(), I18n.Text("The legality class of this equipment"), 3);
        if (showEquipmentState()) {
            mEquippedCheckBox = new JCheckBox(I18n.Text("Equipped"));
            mEquippedCheckBox.setSelected(mRow.isEquipped());
            UIUtilities.setToPreferredSizeOnly(mEquippedCheckBox);
            mEquippedCheckBox.setEnabled(mIsEditable);
            mEquippedCheckBox.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Items that are not equipped do not apply any features they may normally contribute to the character.")));
            wrapper.add(mEquippedCheckBox);
        }
        wrapper.add(new JPanel());
        parent.add(wrapper);
    }

    private void createValueAndWeightFields(Container parent) {
        JPanel    wrapper = new JPanel(new ColumnLayout(4));
        Component first;

        mContainedValue = mRow.getExtendedValue() - mRow.getValue() * mRow.getQuantity();
        mValueField     = createNumberField(parent, wrapper, I18n.Text("Value"), mRow.getValue(), I18n.Text("The value of one of these pieces of equipment"), 13);
        mExtValueField  = createNumberField(wrapper, wrapper, I18n.Text("Extended Value"), mRow.getExtendedValue(), I18n.Text("The value of all of these pieces of equipment, plus the value of any contained equipment"), 13);
        first           = wrapper.getComponent(1);
        mExtValueField.setEnabled(false);
        wrapper.add(new JPanel());
        parent.add(wrapper);

        wrapper          = new JPanel(new ColumnLayout(3));
        mContainedWeight = new WeightValue(mRow.getExtendedWeight());
        WeightValue weight = new WeightValue(mRow.getWeight());
        weight.setValue(weight.getValue() * mRow.getQuantity());
        mContainedWeight.subtract(weight);
        mWeightField    = createWeightField(parent, wrapper, I18n.Text("Weight"), mRow.getWeight(), I18n.Text("The weight of one of these pieces of equipment"), 13);
        mExtWeightField = createWeightField(wrapper, wrapper, I18n.Text("Extended Weight"), mRow.getExtendedWeight(), I18n.Text("The total weight of this quantity of equipment, plus everything contained by it"), 13);
        mExtWeightField.setEnabled(false);
        UIUtilities.adjustToSameSize(new Component[] { first, wrapper.getComponent(1) });
        parent.add(wrapper);
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

    private JTextField createCorrectableField(Container parent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        field.getDocument().addDocumentListener(this);
        field.addFocusListener(this);

        LinkedLabel label = new LinkedLabel(title);
        label.setLink(field);

        parent.add(label);
        parent.add(field);
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
        field.addFocusListener(this);
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
        return field;
    }

    @SuppressWarnings("unused")
    private JTextField createIntegerNumberField(Container labelParent, Container fieldParent, String title, int value, String tooltip, int maxDigits) {
        JTextField field = new JTextField(Text.makeFiller(maxDigits, '9') + Text.makeFiller(maxDigits / 3, ','));
        UIUtilities.setToPreferredSizeOnly(field);
        field.setText(Numbers.format(value));
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        new NumberFilter(field, false, false, true, maxDigits);
        field.addActionListener(this);
        field.addFocusListener(this);
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
        return field;
    }

    @SuppressWarnings("unused")
    private JTextField createNumberField(Container labelParent, Container fieldParent, String title, double value, String tooltip, int maxDigits) {
        JTextField field = new JTextField(Text.makeFiller(maxDigits, '9') + Text.makeFiller(maxDigits / 3, ',') + ".");
        UIUtilities.setToPreferredSizeOnly(field);
        field.setText(Numbers.format(value));
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        new NumberFilter(field, true, false, true, maxDigits);
        field.addActionListener(this);
        field.addFocusListener(this);
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
        return field;
    }

    private JTextField createWeightField(Container labelParent, Container fieldParent, String title, WeightValue value, String tooltip, int maxDigits) {
        JTextField field = new JTextField(Text.makeFiller(maxDigits, '9') + Text.makeFiller(maxDigits / 3, ',') + ".");
        UIUtilities.setToPreferredSizeOnly(field);
        field.setText(value.toString());
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        field.addActionListener(this);
        field.addFocusListener(this);
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
        return field;
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified = mRow.setDescription(mDescriptionField.getText());
        modified |= mRow.setReference(mReferenceField.getText());
        modified |= mRow.setTechLevel(mTechLevelField.getText());
        modified |= mRow.setLegalityClass(mLegalityClassField.getText());
        modified |= mRow.setQuantity(getQty());
        modified |= mRow.setValue(Numbers.extractDouble(mValueField.getText(), 0.0, true));
        modified |= mRow.setWeight(WeightValue.extract(mWeightField.getText(), true));
        modified |= mRow.setMaxUses(Numbers.extractInteger(mMaxUsesField.getText(), 0, true));
        if (mUsesField != null) {
            modified |= mRow.setUses(Numbers.extractInteger(mUsesField.getText(), 0, true));
        } else {
            modified |= mRow.setUses(mRow.getMaxUses());
        }
        if (showEquipmentState()) {
            modified |= mRow.setEquipped(mEquippedCheckBox.isSelected());
        }
        modified |= mRow.setNotes(mNotesField.getText());
        modified |= mRow.setCategories(mCategoriesField.getText());
        if (mPrereqs != null) {
            modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
        }
        if (mFeatures != null) {
            modified |= mRow.setFeatures(mFeatures.getFeatures());
        }
        if (mMeleeWeapons != null) {
            ArrayList<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
            list.addAll(mRangedWeapons.getWeapons());
            modified |= mRow.setWeapons(list);
        }
        return modified;
    }

    @Override
    public void finished() {
        if (mTabPanel != null) {
            updateLastTabName(mTabPanel.getTitleAt(mTabPanel.getSelectedIndex()));
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        adjustForChange(event.getSource());
    }

    private int getQty() {
        if (mQtyField != null) {
            return Numbers.extractInteger(mQtyField.getText(), 0, true);
        }
        return 1;
    }

    private void valueChanged() {
        int    qty = getQty();
        double value;

        if (qty < 1) {
            value = 0;
        } else {
            value = qty * Numbers.extractDouble(mValueField.getText(), 0.0, true) + mContainedValue;
        }
        mExtValueField.setText(Numbers.format(value));
    }

    private void weightChanged() {
        WeightValue weight = WeightValue.extract(mWeightField.getText(), true);
        weight.setValue(weight.getValue() * Math.max(getQty(), 0));
        mExtWeightField.setText(weight.toString());
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        descriptionChanged();
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        descriptionChanged();
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        descriptionChanged();
    }

    private void descriptionChanged() {
        LinkedLabel.setErrorMessage(mDescriptionField, mDescriptionField.getText().trim().length() != 0 ? null : I18n.Text("The name field may not be empty"));
    }

    @Override
    public void focusGained(FocusEvent event) {
        // Not used.
    }

    @Override
    public void focusLost(FocusEvent event) {
        adjustForChange(event.getSource());
    }

    private void adjustForChange(Object field) {
        if (field == mValueField) {
            valueChanged();
        } else if (field == mWeightField) {
            weightChanged();
        } else if (field == mQtyField) {
            valueChanged();
            weightChanged();
        }
    }
}
