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

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Container;
import javax.swing.JComboBox;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** Editor for {@link EquipmentModifier}s. */
public class EquipmentModifierEditor extends RowEditor<EquipmentModifier> implements DocumentListener {
    private EditorField        mNameField;
    private EditorField        mTechLevelField;
    private Checkbox           mEnabledField;
    private MultiLineTextField mNotesField;
    private EditorField        mReferenceField;
    private FeaturesPanel      mFeatures;
    private JComboBox<Object>  mCostType;
    private EditorField        mCostAmountField;
    private JComboBox<Object>  mWeightType;
    private EditorField        mWeightAmountField;

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
        Panel panel = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
        if (mRow.canHaveChildren()) {
            mNameField = createCorrectableField(panel, panel, I18n.text("Name"), mRow.getName(), I18n.text("Name of container"));
        } else {
            Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(4));
            mNameField = createCorrectableField(panel, wrapper, I18n.text("Name"), mRow.getName(), I18n.text("Name of Modifier"));
            mTechLevelField = createField(wrapper, wrapper, I18n.text("Tech Level"), mRow.getTechLevel(), I18n.text("The first Tech Level this equipment is available at"), 3);
            mEnabledField = new Checkbox(I18n.text("Enabled"), mRow.isEnabled(), null);
            mEnabledField.setToolTipText(I18n.text("Whether this modifier has been enabled or not"));
            wrapper.add(mEnabledField);
            panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

            createCostAdjustmentFields(panel);
            createWeightAdjustmentFields(panel);
        }

        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.text("Any notes that you would like to show up in the list along with this modifier"), this);
        panel.add(new Label(I18n.text("Notes"), mNotesField), new PrecisionLayoutData().setFillHorizontalAlignment().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2));
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
            modified |= mRow.setEnabled(mEnabledField.isChecked());
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

    private EditorField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text, tooltip);
        field.getDocument().addDocumentListener(this);
        addLabel(labelParent, title, field);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    private static EditorField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text,
                maxChars > 0 ? Text.makeFiller(maxChars, 'M') : null, tooltip);
        addLabel(labelParent, title, field);
        PrecisionLayoutData ld = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (maxChars == 0) {
            ld.setGrabHorizontalSpace(true);
        }
        fieldParent.add(field, ld);
        return field;
    }

    private void createCostAdjustmentFields(Container parent) {
        Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
        createCostAdjustmentField(parent, wrapper);
        createCostTypeCombo(wrapper);
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void createCostAdjustmentField(Container labelParent, Container fieldParent) {
        mCostAmountField = new EditorField(FieldFactory.STRING, (f) -> costChanged(),
                SwingConstants.LEFT, mRow.getCostAdjType().format(mRow.getCostAdjAmount(), true),
                "-999,999,999.00", I18n.text("The cost modifier"));
        addLabel(labelParent, "", mCostAmountField);
        fieldParent.add(mCostAmountField);
    }

    private void createCostTypeCombo(Container parent) {
        EquipmentModifierCostType[] types = EquipmentModifierCostType.values();
        mCostType = new JComboBox<>(types);
        mCostType.setSelectedItem(mRow.getCostAdjType());
        mCostType.addActionListener((evt) -> costChanged());
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
        Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
        createWeightAdjustmentField(parent, wrapper);
        createWeightTypeCombo(wrapper);
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void createWeightAdjustmentField(Container labelParent, Container fieldParent) {
        mWeightAmountField = new EditorField(FieldFactory.STRING, (f) -> weightChanged(),
                SwingConstants.LEFT, mRow.getWeightAdjType().format(mRow.getWeightAdjAmount(),
                mRow.getDataFile().getSheetSettings().defaultWeightUnits(), true),
                "-999,999,999.00", I18n.text("The weight modifier"));
        labelParent.add(new Label("", mWeightAmountField));
        fieldParent.add(mWeightAmountField);
    }

    private void createWeightTypeCombo(Container parent) {
        EquipmentModifierWeightType[] types = EquipmentModifierWeightType.values();
        mWeightType = new JComboBox<>(types);
        mWeightType.setSelectedItem(mRow.getWeightAdjType());
        mWeightType.addActionListener((evt) -> weightChanged());
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

    private void docChanged(DocumentEvent event) {
        if (mNameField.getDocument() == event.getDocument()) {
            Label.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.text("The name field may not be empty") : null);
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
