/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Container;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** Editor for {@link EquipmentModifier}s. */
public class EquipmentModifierEditor extends RowEditor<EquipmentModifier> implements DocumentListener {
    private EditorField                            mNameField;
    private EditorField                            mTechLevelField;
    private Checkbox                               mEnabledField;
    private MultiLineTextField                     mNotesField;
    private MultiLineTextField                     mVTTNotesField;
    private EditorField                            mReferenceField;
    private FeaturesPanel                          mFeatures;
    private PopupMenu<EquipmentModifierCostType>   mCostType;
    private EditorField                            mCostAmountField;
    private PopupMenu<EquipmentModifierWeightType> mWeightType;
    private EditorField                            mWeightAmountField;

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
        addLabel(panel, I18n.text("Name"));
        if (mRow.canHaveChildren()) {
            mNameField = createCorrectableField(panel, mRow.getName(), I18n.text("Name of container"));
        } else {
            Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(4));
            mNameField = createCorrectableField(wrapper, mRow.getName(), I18n.text("Name of Modifier"));
            addInteriorLabel(wrapper, I18n.text("Tech Level"));
            mTechLevelField = createField(wrapper, mRow.getTechLevel(), I18n.text("The first Tech Level this equipment is available at"), 3);
            mEnabledField = new Checkbox(I18n.text("Enabled"), mRow.isEnabled(), null);
            mEnabledField.setToolTipText(I18n.text("Whether this modifier has been enabled or not"));
            wrapper.add(mEnabledField, new PrecisionLayoutData().setLeftMargin(4));
            panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

            wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
            addLabel(panel, I18n.text("Cost Modifier"));
            mCostAmountField = new EditorField(FieldFactory.STRING, (f) -> costChanged(),
                    SwingConstants.LEFT, mRow.getCostAdjType().format(mRow.getCostAdjAmount(), true),
                    "-999,999,999.00", I18n.text("The cost modifier"));
            wrapper.add(mCostAmountField);
            mCostType = new PopupMenu<>(EquipmentModifierCostType.values(), (p) -> costChanged());
            mCostType.setSelectedItem(mRow.getCostAdjType(), false);
            wrapper.add(mCostType);
            panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

            wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
            addLabel(panel, I18n.text("Weight Modifier"));
            mWeightAmountField = new EditorField(FieldFactory.STRING, (f) -> weightChanged(),
                    SwingConstants.LEFT, mRow.getWeightAdjType().format(mRow.getWeightAdjAmount(),
                    mRow.getDataFile().getSheetSettings().defaultWeightUnits(), true),
                    "-999,999,999.00", I18n.text("The weight modifier"));
            wrapper.add(mWeightAmountField);
            mWeightType = new PopupMenu<>(EquipmentModifierWeightType.values(), (p) -> weightChanged());
            mWeightType.setSelectedItem(mRow.getWeightAdjType(), false);
            wrapper.add(mWeightType);
            panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        }

        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.text("Any notes that you would like to show up in the list along with this modifier"), this);
        addLabel(panel, I18n.text("Notes")).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2);
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mVTTNotesField = addVTTNotesField(panel, this);

        addLabel(panel, I18n.text("Page Reference"));
        mReferenceField = createField(panel, mRow.getReference(), PageRefCell.getStdToolTip(I18n.text("equipment modifier")), 6);
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
        modified |= mRow.setVTTNotes(mVTTNotesField.getText());
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

    private EditorField createCorrectableField(Container fieldParent, String text, String tooltip) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text, tooltip);
        field.getDocument().addDocumentListener(this);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    private static EditorField createField(Container fieldParent, String text, String tooltip, int maxChars) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text,
                maxChars > 0 ? Text.makeFiller(maxChars, 'M') : null, tooltip);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    private EquipmentModifierCostType getCostType() {
        EquipmentModifierCostType selection = mCostType.getSelectedItem();
        return selection == null ? EquipmentModifierCostType.TO_ORIGINAL_COST : selection;
    }

    private EquipmentModifierWeightType getWeightType() {
        EquipmentModifierWeightType selection = mWeightType.getSelectedItem();
        return selection == null ? EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT : selection;
    }

    private void docChanged(DocumentEvent event) {
        if (mNameField.getDocument() == event.getDocument()) {
            mNameField.setErrorMessage(mNameField.getText().trim().isEmpty() ? I18n.text("The name field may not be empty") : null);
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
