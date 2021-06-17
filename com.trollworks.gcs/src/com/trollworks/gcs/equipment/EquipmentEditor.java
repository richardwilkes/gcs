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

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.modifier.EquipmentModifierListEditor;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.StdCheckbox;
import com.trollworks.gcs.ui.widget.StdLabel;
import com.trollworks.gcs.ui.widget.StdPanel;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.Filtered;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.utility.units.WeightValue;
import com.trollworks.gcs.weapon.MeleeWeaponListEditor;
import com.trollworks.gcs.weapon.RangedWeaponListEditor;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Container;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComponent;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** The detailed editor for {@link Equipment}s. */
public class EquipmentEditor extends RowEditor<Equipment> implements DocumentListener {
    private StdCheckbox                 mEquippedCheckBox;
    private StdCheckbox                 mIgnoreWeightForSkillsCheckBox;
    private EditorField                 mDescriptionField;
    private EditorField                 mTechLevelField;
    private EditorField                 mLegalityClassField;
    private EditorField                 mQtyField;
    private EditorField                 mUsesField;
    private EditorField                 mMaxUsesField;
    private EditorField                 mValueField;
    private EditorField                 mExtValueField;
    private EditorField                 mWeightField;
    private EditorField                 mExtWeightField;
    private MultiLineTextField          mNotesField;
    private EditorField                 mCategoriesField;
    private EditorField                 mReferenceField;
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
        mModifiers.addActionListener((evt) -> {
            valueChanged();
            weightChanged();
        });
        addSection(outer, mModifiers);
        List<WeaponStats> weapons = mRow.getWeapons();
        mMeleeWeapons = new MeleeWeaponListEditor(mRow, weapons);
        addSection(outer, mMeleeWeapons);
        mRangedWeapons = new RangedWeaponListEditor(mRow, weapons);
        addSection(outer, mRangedWeapons);
    }

    private StdPanel createTopSection() {
        StdPanel panel = new StdPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        mDescriptionField = createCorrectableField(panel, I18n.text("Name"), mRow.getDescription(), I18n.text("The name/description of the equipment, without any notes"));
        createSecondLineFields(panel);
        createValueAndWeightFields(panel);
        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.text("Any notes that you would like to show up in the list along with this equipment"), this);
        panel.add(new StdLabel(I18n.text("Notes"), mNotesField), new PrecisionLayoutData().setBeginningVerticalAlignment().setFillHorizontalAlignment().setTopMargin(2));
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mCategoriesField = createField(panel, panel, I18n.text("Categories"), mRow.getCategoriesAsString(), I18n.text("The category or categories the equipment belongs to (separate multiple categories with a comma)"), 0);

        boolean    forCharacterOrTemplate = mRow.getCharacter() != null || mRow.getTemplate() != null;
        StdPanel   wrapper                = new StdPanel(new PrecisionLayout().setMargins(0).setColumns(forCharacterOrTemplate ? 5 : 3));
        JComponent labelParent            = panel;
        if (forCharacterOrTemplate) {
            mUsesField = createIntegerNumberField(panel, wrapper, I18n.text("Uses"), mRow.getUses(),
                    I18n.text("The number of uses remaining for this equipment"), 99999, null);
            labelParent = wrapper;
        }
        mMaxUsesField = createIntegerNumberField(labelParent, wrapper, I18n.text("Max Uses"),
                mRow.getMaxUses(), I18n.text("The maximum number of uses for this equipment"),
                99999, null);
        mReferenceField = createField(wrapper, wrapper, I18n.text("Page Reference"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.text("equipment")), 0);
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return panel;
    }

    private void createSecondLineFields(Container parent) {
        boolean  isContainer = mRow.canHaveChildren();
        StdPanel wrapper     = new StdPanel(new PrecisionLayout().setMargins(0).setColumns((isContainer ? 4 : 6) + (showEquipmentState() ? 1 : 0)));
        if (!isContainer) {
            mQtyField = createIntegerNumberField(parent, wrapper, I18n.text("Quantity"),
                    mRow.getQuantity(), I18n.text("The number of this equipment present"), 999999999,
                    (f) -> {
                        valueChanged();
                        weightChanged();
                    });
        }
        mTechLevelField = createField(isContainer ? parent : wrapper, wrapper, I18n.text("Tech Level"), mRow.getTechLevel(), I18n.text("The first Tech Level this equipment is available at"), 3);
        mLegalityClassField = createField(wrapper, wrapper, I18n.text("Legality Class"), mRow.getLegalityClass(), I18n.text("The legality class of this equipment"), 3);
        if (showEquipmentState()) {
            mEquippedCheckBox = new StdCheckbox(I18n.text("Equipped"), mRow.isEquipped(), null);
            mEquippedCheckBox.setEnabled(mIsEditable);
            mEquippedCheckBox.setToolTipText(Text.wrapPlainTextForToolTip(I18n.text("Items that are not equipped do not apply any features they may normally contribute to the character.")));
            wrapper.add(mEquippedCheckBox);
        }
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private boolean showEquipmentState() {
        return mCarried && mRow.getCharacter() != null;
    }

    private void createValueAndWeightFields(Container parent) {
        StdPanel wrapper = new StdPanel(new PrecisionLayout().setMargins(0).setColumns(3));
        mContainedValue = mRow.getExtendedValue().sub(mRow.getAdjustedValue().mul(new Fixed6(mRow.getQuantity())));
        Fixed6 protoValue = new Fixed6("9999999.999999", false);
        mValueField = createValueField(parent, wrapper, I18n.text("Value"), mRow.getValue(),
                protoValue,
                I18n.text("The base value of one of these pieces of equipment before modifiers"),
                (f) -> valueChanged());
        mExtValueField = createValueField(wrapper, wrapper, I18n.text("Extended"),
                mRow.getExtendedValue(), protoValue,
                I18n.text("The value of all of these pieces of equipment, plus the value of any contained equipment"),
                null);
        mExtValueField.setEnabled(false);
        parent.add(wrapper);

        wrapper = new StdPanel(new PrecisionLayout().setMargins(0).setColumns(4));
        mContainedWeight = new WeightValue(mRow.getExtendedWeight(false));
        WeightValue weight = new WeightValue(mRow.getAdjustedWeight(false));
        weight.setValue(weight.getValue().mul(new Fixed6(mRow.getQuantity())));
        mContainedWeight.subtract(weight);
        WeightValue weightProto = new WeightValue(protoValue, WeightUnits.LB);
        mWeightField = createWeightField(parent, wrapper, I18n.text("Weight"), mRow.getWeight(),
                weightProto, I18n.text("The weight of one of these pieces of equipment"),
                (f) -> weightChanged());
        mExtWeightField = createWeightField(wrapper, wrapper, I18n.text("Extended"),
                mRow.getExtendedWeight(false), weightProto,
                I18n.text("The total weight of this quantity of equipment, plus everything contained by it"),
                null);
        mExtWeightField.setEnabled(false);
        mIgnoreWeightForSkillsCheckBox = new StdCheckbox(I18n.text("Ignore for Skills"), mRow.isWeightIgnoredForSkills(), null);
        mIgnoreWeightForSkillsCheckBox.setEnabled(mIsEditable);
        mIgnoreWeightForSkillsCheckBox.setToolTipText(Text.wrapPlainTextForToolTip(I18n.text("If checked, the weight of this item is not considered when calculating encumbrance penalties for skills")));
        wrapper.add(mIgnoreWeightForSkillsCheckBox);
        parent.add(wrapper);
    }

    private EditorField createCorrectableField(Container parent, String title, String text, String tooltip) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text, tooltip);
        field.setEnabled(mIsEditable);
        field.getDocument().addDocumentListener(this);
        parent.add(new StdLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    private EditorField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text, maxChars > 0 ? Text.makeFiller(maxChars, 'M') : null, tooltip);
        field.setEnabled(mIsEditable);
        labelParent.add(new StdLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        PrecisionLayoutData ld = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (maxChars == 0) {
            ld.setGrabHorizontalSpace(true);
        }
        fieldParent.add(field, ld);
        return field;
    }

    private EditorField createIntegerNumberField(Container labelParent, Container fieldParent, String title, int value, String tooltip, int maxValue, EditorField.ChangeListener listener) {
        EditorField field = new EditorField(maxValue == 99999 ? FieldFactory.POSINT5 : FieldFactory.POSINT9, listener, SwingConstants.LEFT, Integer.valueOf(value), Integer.valueOf(maxValue), tooltip);
        field.setEnabled(mIsEditable);
        labelParent.add(new StdLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    private EditorField createValueField(Container labelParent, Container fieldParent, String title, Fixed6 value, Fixed6 protoValue, String tooltip, EditorField.ChangeListener listener) {
        EditorField field = new EditorField(FieldFactory.FIXED6, listener, SwingConstants.LEFT, value, protoValue, tooltip);
        field.setEnabled(mIsEditable);
        labelParent.add(new StdLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    private EditorField createWeightField(Container labelParent, Container fieldParent, String title, WeightValue value, WeightValue protoValue, String tooltip, EditorField.ChangeListener listener) {
        EditorField field = new EditorField(FieldFactory.WEIGHT, listener, SwingConstants.LEFT, value, protoValue, tooltip);
        field.setEnabled(mIsEditable);
        labelParent.add(new StdLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
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
        modified |= mRow.setWeightIgnoredForSkills(mIgnoreWeightForSkillsCheckBox.isChecked());
        modified |= mRow.setMaxUses(Numbers.extractInteger(mMaxUsesField.getText(), 0, true));
        modified |= mUsesField != null ? mRow.setUses(Numbers.extractInteger(mUsesField.getText(), 0, true)) : mRow.setUses(mRow.getMaxUses());
        if (showEquipmentState()) {
            modified |= mRow.setEquipped(mEquippedCheckBox.isChecked());
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

    private int getQty() {
        if (mQtyField != null) {
            return Numbers.extractInteger(mQtyField.getText(), 0, true);
        }
        return 1;
    }

    private void valueChanged() {
        int    qty   = getQty();
        Fixed6 value = qty < 1 ? Fixed6.ZERO : new Fixed6(qty).mul(Equipment.getValueAdjustedForModifiers(new Fixed6(mValueField.getText(), Fixed6.ZERO, true), Filtered.list(mModifiers.getAllModifiers(), EquipmentModifier.class))).add(mContainedValue);
        mExtValueField.setText(value.toLocalizedString());
    }

    private void weightChanged() {
        int         qty    = getQty();
        WeightValue weight = mRow.getWeightAdjustedForModifiers(WeightValue.extract(qty < 1 ? "0" : mWeightField.getText(), true), Filtered.list(mModifiers.getAllModifiers(), EquipmentModifier.class));
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
            StdLabel.setErrorMessage(mDescriptionField, mDescriptionField.getText().trim().isEmpty() ? I18n.text("The name field may not be empty") : null);
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
