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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.modifier.EquipmentModifierListEditor;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.FilteredList;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.NumberFilter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.units.WeightValue;
import com.trollworks.gcs.weapon.MeleeWeaponListEditor;
import com.trollworks.gcs.weapon.RangedWeaponListEditor;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Container;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JCheckBox;
import javax.swing.JComponent;
import javax.swing.JPanel;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** The detailed editor for {@link Equipment}s. */
public class EquipmentEditor extends RowEditor<Equipment> implements ActionListener, DocumentListener, FocusListener {
    private JCheckBox                   mEquippedCheckBox;
    private JCheckBox                   mIgnoreWeightForSkillsCheckBox;
    private JTextField                  mDescriptionField;
    private JTextField                  mTechLevelField;
    private JTextField                  mLegalityClassField;
    private JTextField                  mQtyField;
    private JTextField                  mUsesField;
    private JTextField                  mMaxUsesField;
    private JTextField                  mValueField;
    private JTextField                  mExtValueField;
    private JTextField                  mWeightField;
    private JTextField                  mExtWeightField;
    private MultiLineTextField          mNotesField;
    private JTextField                  mCategoriesField;
    private JTextField                  mReferenceField;
    private PrereqsPanel                mPrereqs;
    private FeaturesPanel               mFeatures;
    private MeleeWeaponListEditor       mMeleeWeapons;
    private RangedWeaponListEditor      mRangedWeapons;
    private EquipmentModifierListEditor mModifiers;
    private Fixed6                      mContainedValue;
    private WeightValue                 mContainedWeight;
    private boolean                     mCarried;

    /**
     * Creates a new {@link Equipment} editor.
     *
     * @param equipment The {@link Equipment} to edit.
     * @param carried   {@code true} for the carried equipment, {@code false} for the other
     *                  equipment.
     */
    public EquipmentEditor(Equipment equipment, boolean carried) {
        super(equipment);
        mCarried = carried;
        addContent();
    }

    @Override
    protected void addContentSelf(ScrollContent outer) {
        outer.add(createTopSection(), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mPrereqs = new PrereqsPanel(mRow, mRow.getPrereqs());
        addSection(outer, mPrereqs);
        mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
        addSection(outer, mFeatures);
        mModifiers = EquipmentModifierListEditor.createEditor(mRow);
        addSection(outer, mModifiers);
        List<WeaponStats> weapons = mRow.getWeapons();
        mMeleeWeapons = new MeleeWeaponListEditor(mRow, weapons);
        addSection(outer, mMeleeWeapons);
        mRangedWeapons = new RangedWeaponListEditor(mRow, weapons);
        addSection(outer, mRangedWeapons);
    }

    private JPanel createTopSection() {
        JPanel panel = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        mDescriptionField = createCorrectableField(panel, I18n.Text("Name"), mRow.getDescription(), I18n.Text("The name/description of the equipment, without any notes"));
        createSecondLineFields(panel);
        createValueAndWeightFields(panel);
        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this equipment"), this);
        panel.add(new LinkedLabel(I18n.Text("Notes"), mNotesField), new PrecisionLayoutData().setBeginningVerticalAlignment().setFillHorizontalAlignment().setTopMargin(2));
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mCategoriesField = createField(panel, panel, I18n.Text("Categories"), mRow.getCategoriesAsString(), I18n.Text("The category or categories the equipment belongs to (separate multiple categories with a comma)"), 0);

        boolean    forCharacterOrTemplate = mRow.getCharacter() != null || mRow.getTemplate() != null;
        JPanel     wrapper                = new JPanel(new PrecisionLayout().setMargins(0).setColumns(forCharacterOrTemplate ? 5 : 3));
        JComponent labelParent            = panel;
        if (forCharacterOrTemplate) {
            mUsesField = createIntegerNumberField(panel, wrapper, I18n.Text("Uses"), mRow.getUses(), I18n.Text("The number of uses remaining for this equipment"), 5);
            labelParent = wrapper;
        }
        mMaxUsesField = createIntegerNumberField(labelParent, wrapper, I18n.Text("Max Uses"), mRow.getMaxUses(), I18n.Text("The maximum number of uses for this equipment"), 5);
        mReferenceField = createField(wrapper, wrapper, I18n.Text("Page Reference"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.Text("equipment")), 0);
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return panel;
    }

    private void createSecondLineFields(Container parent) {
        boolean isContainer = mRow.canHaveChildren();
        JPanel  wrapper     = new JPanel(new PrecisionLayout().setMargins(0).setColumns((isContainer ? 4 : 6) + (showEquipmentState() ? 1 : 0)));
        if (!isContainer) {
            mQtyField = createIntegerNumberField(parent, wrapper, I18n.Text("Quantity"), mRow.getQuantity(), I18n.Text("The number of this equipment present"), 9);
        }
        mTechLevelField = createField(isContainer ? parent : wrapper, wrapper, I18n.Text("Tech Level"), mRow.getTechLevel(), I18n.Text("The first Tech Level this equipment is available at"), 3);
        mLegalityClassField = createField(wrapper, wrapper, I18n.Text("Legality Class"), mRow.getLegalityClass(), I18n.Text("The legality class of this equipment"), 3);
        if (showEquipmentState()) {
            mEquippedCheckBox = new JCheckBox(I18n.Text("Equipped"));
            mEquippedCheckBox.setSelected(mRow.isEquipped());
            mEquippedCheckBox.setEnabled(mIsEditable);
            mEquippedCheckBox.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Items that are not equipped do not apply any features they may normally contribute to the character.")));
            wrapper.add(mEquippedCheckBox);
        }
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private boolean showEquipmentState() {
        return mCarried && mRow.getCharacter() != null;
    }

    private void createValueAndWeightFields(Container parent) {
        JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(3));
        mContainedValue = mRow.getExtendedValue().sub(mRow.getAdjustedValue().mul(new Fixed6(mRow.getQuantity())));
        mValueField = createValueField(parent, wrapper, I18n.Text("Value"), mRow.getValue(), I18n.Text("The base value of one of these pieces of equipment before modifiers"), 13);
        mExtValueField = createValueField(wrapper, wrapper, I18n.Text("Extended"), mRow.getExtendedValue(), I18n.Text("The value of all of these pieces of equipment, plus the value of any contained equipment"), 13);
        mExtValueField.setEnabled(false);
        parent.add(wrapper);

        wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(4));
        mContainedWeight = new WeightValue(mRow.getExtendedWeight(false));
        WeightValue weight = new WeightValue(mRow.getAdjustedWeight(false));
        weight.setValue(weight.getValue().mul(new Fixed6(mRow.getQuantity())));
        mContainedWeight.subtract(weight);
        mWeightField = createWeightField(parent, wrapper, I18n.Text("Weight"), mRow.getWeight(), I18n.Text("The weight of one of these pieces of equipment"), 13);
        mExtWeightField = createWeightField(wrapper, wrapper, I18n.Text("Extended"), mRow.getExtendedWeight(false), I18n.Text("The total weight of this quantity of equipment, plus everything contained by it"), 13);
        mExtWeightField.setEnabled(false);
        mIgnoreWeightForSkillsCheckBox = new JCheckBox(I18n.Text("Ignore for Skills"));
        mIgnoreWeightForSkillsCheckBox.setSelected(mRow.isWeightIgnoredForSkills());
        UIUtilities.setToPreferredSizeOnly(mIgnoreWeightForSkillsCheckBox);
        mIgnoreWeightForSkillsCheckBox.setEnabled(mIsEditable);
        mIgnoreWeightForSkillsCheckBox.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("If checked, the weight of this item is not considered when calculating encumbrance penalties for skills")));
        wrapper.add(mIgnoreWeightForSkillsCheckBox);
        parent.add(wrapper);
    }

    private JTextField createCorrectableField(Container parent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        field.getDocument().addDocumentListener(this);
        field.addFocusListener(this);
        parent.add(new LinkedLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
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
        labelParent.add(new LinkedLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        PrecisionLayoutData ld = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (maxChars == 0) {
            ld.setGrabHorizontalSpace(true);
        }
        fieldParent.add(field, ld);
        return field;
    }

    private JTextField createIntegerNumberField(Container labelParent, Container fieldParent, String title, int value, String tooltip, int maxDigits) {
        JTextField field = new JTextField(Text.makeFiller(maxDigits, '9') + Text.makeFiller(maxDigits / 3, ','));
        UIUtilities.setToPreferredSizeOnly(field);
        field.setText(Numbers.format(value));
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        NumberFilter.apply(field, false, false, true, maxDigits);
        field.addActionListener(this);
        field.addFocusListener(this);
        labelParent.add(new LinkedLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    private JTextField createValueField(Container labelParent, Container fieldParent, String title, Fixed6 value, String tooltip, int maxDigits) {
        JTextField field = new JTextField(Text.makeFiller(maxDigits, '9') + Text.makeFiller(maxDigits / 3, ',') + ".");
        UIUtilities.setToPreferredSizeOnly(field);
        field.setText(value.toLocalizedString());
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        NumberFilter.apply(field, true, false, true, maxDigits);
        field.addActionListener(this);
        field.addFocusListener(this);
        labelParent.add(new LinkedLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
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
        labelParent.add(new LinkedLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified = mRow.setDescription(mDescriptionField.getText());
        modified |= mRow.setReference(mReferenceField.getText());
        modified |= mRow.setTechLevel(mTechLevelField.getText());
        modified |= mRow.setLegalityClass(mLegalityClassField.getText());
        modified |= mRow.setQuantity(getQty());
        modified |= mRow.setValue(new Fixed6(mValueField.getText(), Fixed6.ZERO, true));
        modified |= mRow.setWeight(WeightValue.extract(mWeightField.getText(), true));
        modified |= mRow.setWeightIgnoredForSkills(mIgnoreWeightForSkillsCheckBox.isSelected());
        modified |= mRow.setMaxUses(Numbers.extractInteger(mMaxUsesField.getText(), 0, true));
        modified |= mUsesField != null ? mRow.setUses(Numbers.extractInteger(mUsesField.getText(), 0, true)) : mRow.setUses(mRow.getMaxUses());
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
            List<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
            list.addAll(mRangedWeapons.getWeapons());
            modified |= mRow.setWeapons(list);
        }
        if (mModifiers.wasModified()) {
            modified = true;
            mRow.setModifiers(mModifiers.getModifiers());
        }
        return modified;
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
        int    qty   = getQty();
        Fixed6 value = qty < 1 ? Fixed6.ZERO : new Fixed6(qty).mul(Equipment.getValueAdjustedForModifiers(new Fixed6(mValueField.getText(), Fixed6.ZERO, true), new FilteredList<>(mModifiers.getAllModifiers(), EquipmentModifier.class))).add(mContainedValue);
        mExtValueField.setText(value.toLocalizedString());
    }

    private void weightChanged() {
        int         qty    = getQty();
        WeightValue weight = mRow.getWeightAdjustedForModifiers(WeightValue.extract(qty < 1 ? "0" : mWeightField.getText(), true), new FilteredList<>(mModifiers.getAllModifiers(), EquipmentModifier.class));
        if (qty > 0) {
            weight.setValue(weight.getValue().mul(new Fixed6(qty)));
            weight.add(mContainedWeight);
        } else {
            weight.setValue(Fixed6.ZERO);
        }
        mExtWeightField.setText(weight.toString());
    }

    private void docChanged(DocumentEvent event) {
        if (mDescriptionField.getDocument() == event.getDocument()) {
            LinkedLabel.setErrorMessage(mDescriptionField, mDescriptionField.getText().trim().isEmpty() ? I18n.Text("The name field may not be empty") : null);
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
        adjustForChange(event.getSource());
    }

    private void adjustForChange(Object field) {
        if (field == mValueField) {
            valueChanged();
        } else if (field == mWeightField) {
            weightChanged();
        } else if (field == mQtyField || field == mModifiers) {
            valueChanged();
            weightChanged();
        }
    }
}
