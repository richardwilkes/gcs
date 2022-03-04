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
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Container;
import java.text.MessageFormat;
import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** Editor for {@link AdvantageModifier}s. */
public class AdvantageModifierEditor extends RowEditor<AdvantageModifier> implements DocumentListener {
    private EditorField        mNameField;
    private Checkbox           mEnabledField;
    private MultiLineTextField mNotesField;
    private MultiLineTextField mVTTNotesField;
    private EditorField        mReferenceField;
    private EditorField        mCostField;
    private EditorField        mLevelField;
    private EditorField        mCostModifierField;
    private FeaturesPanel      mFeatures;
    private PopupMenu<Object>  mCostType;
    private PopupMenu<Affects> mAffects;
    private int                mLastLevel;

    /**
     * Creates a new AdvantageModifierEditor.
     *
     * @param modifier The {@link AdvantageModifier} to edit.
     */
    public AdvantageModifierEditor(AdvantageModifier modifier) {
        super(modifier);
        addContent();
        if (!modifier.canHaveChildren()) {
            updateCostModifier();
        }
    }

    @Override
    protected void addContentSelf(ScrollContent outer) {
        Panel panel = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
        addLabel(panel, I18n.text("Name"));
        if (mRow.canHaveChildren()) {
            mNameField = createCorrectableField(panel, mRow.getName(), I18n.text("Name of container"));
        } else {
            Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
            mNameField = createCorrectableField(wrapper, mRow.getName(), I18n.text("Name of Modifier"));
            mEnabledField = new Checkbox(I18n.text("Enabled"), mRow.isEnabled(), null);
            mEnabledField.setToolTipText(I18n.text("Whether this modifier has been enabled or not"));
            mEnabledField.setEnabled(mIsEditable);
            wrapper.add(mEnabledField, new PrecisionLayoutData().setLeftMargin(4));
            panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

            createCostModifierFields(panel);
        }

        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.text("Any notes that you would like to show up in the list along with this modifier"), this);
        addLabel(panel, I18n.text("Notes")).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2);
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mVTTNotesField = addVTTNotesField(panel, this);

        addLabel(panel, I18n.text("Page Reference"));
        mReferenceField = createField(panel, mRow.getReference(),
                PageRefCell.getStdToolTip(I18n.text("advantage modifier")), 6);
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
            modified |= getCostType() == AdvantageModifierCostType.MULTIPLIER ?
                    mRow.setCostMultiplier(getCostMultiplier()) : mRow.setCost(getCost());
            if (hasLevels()) {
                modified |= mRow.setLevels(getLevels());
                modified |= mRow.setCostType(AdvantageModifierCostType.PERCENTAGE);
            } else {
                modified |= mRow.setLevels(0);
                modified |= mRow.setCostType(getCostType());
            }
            modified |= mRow.setAffects(mAffects.getSelectedItem());
            modified |= mRow.setEnabled(mEnabledField.isChecked());
            if (mFeatures != null) {
                modified |= mRow.setFeatures(mFeatures.getFeatures());
            }
        }
        return modified;
    }

    private boolean hasLevels() {
        return mCostType.getSelectedIndex() == 0;
    }

    private EditorField createCorrectableField(Container fieldParent, String text, String tooltip) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text, tooltip);
        field.getDocument().addDocumentListener(this);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    private static EditorField createField(Container parent, String text, String tooltip, int maxChars) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text,
                maxChars > 0 ? Text.makeFiller(maxChars, 'M') : null, tooltip);
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    private EditorField createNumberField(Container fieldParent, JFormattedTextField.AbstractFormatterFactory formatter, int value, int protoValue, String tooltip, EditorField.ChangeListener listener) {
        EditorField field = new EditorField(formatter, listener, SwingConstants.LEFT,
                Integer.valueOf(value), Integer.valueOf(protoValue), tooltip);
        field.setEnabled(mIsEditable);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    private EditorField createNumberField(Container fieldParent, double value, String tooltip, EditorField.ChangeListener listener) {
        EditorField field = new EditorField(FieldFactory.FLOAT, listener, SwingConstants.LEFT,
                Double.valueOf(value), Double.valueOf(-99999), tooltip);
        field.setEnabled(mIsEditable);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    private void createCostModifierFields(Container parent) {
        Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(7));
        mLastLevel = mRow.getLevels();
        if (mLastLevel < 1) {
            mLastLevel = 1;
        }
        String costTooltip = I18n.text("The base cost modifier");
        addLabel(parent, I18n.text("Cost"));
        mCostField = mRow.getCostType() == AdvantageModifierCostType.MULTIPLIER ?
                createNumberField(wrapper, mRow.getCostMultiplier(), costTooltip,
                        (f) -> updateCostModifier()) :
                createNumberField(wrapper, FieldFactory.INT5, mRow.getCost(), -99999, costTooltip,
                        (f) -> updateCostModifier());
        createCostType(wrapper);
        addInteriorLabel(wrapper, I18n.text("Levels"));
        mLevelField = createNumberField(wrapper, FieldFactory.POSINT3,
                mLastLevel, 999, I18n.text("The number of levels this modifier has"),
                (f) -> updateCostModifier());
        addInteriorLabel(wrapper, I18n.text("Total"));
        mCostModifierField = createField(wrapper, "", I18n.text("The cost modifier's total value"), 9);
        mAffects = new PopupMenu<>(Affects.values(), null);
        mAffects.setSelectedItem(mRow.getAffects(), false);
        wrapper.add(mAffects);
        mCostModifierField.setEnabled(false);
        if (!mRow.hasLevels()) {
            mLevelField.setText("");
            mLevelField.setEnabled(false);
        }
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void createCostType(Container parent) {
        AdvantageModifierCostType[] types  = AdvantageModifierCostType.values();
        Object[]                    values = new Object[types.length + 1];
        values[0] = MessageFormat.format(I18n.text("{0} per level"), AdvantageModifierCostType.PERCENTAGE.toString());
        System.arraycopy(types, 0, values, 1, types.length);
        mCostType = new PopupMenu<>(values, (p) -> {
            if (!mRow.canHaveChildren()) {
                boolean hasLevels = hasLevels();
                if (hasLevels) {
                    mLevelField.setValue(Integer.valueOf(mLastLevel));
                } else {
                    mLastLevel = ((Integer) mLevelField.getValue()).intValue();
                    mLevelField.setText("");
                }
                mLevelField.setEnabled(hasLevels);

                Object value = mCostField.getValue();
                if (getCostType() == AdvantageModifierCostType.MULTIPLIER) {
                    if (mCostField.getFormatterFactory() != FieldFactory.FLOAT) {
                        mCostField.setFormatterFactory(FieldFactory.FLOAT);
                        mCostField.setValue(Double.valueOf(Math.abs(((Number) value).intValue())));
                    }
                } else {
                    if (mCostField.getFormatterFactory() != FieldFactory.INT5) {
                        mCostField.setFormatterFactory(FieldFactory.INT5);
                        mCostField.setValue(Integer.valueOf(((Number) value).intValue()));
                    }
                }
                updateCostModifier();
            }
        });
        mCostType.setSelectedItem(mRow.hasLevels() ? values[0] : mRow.getCostType(), false);
        parent.add(mCostType, new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private void updateCostModifier() {
        boolean enabled = true;
        if (hasLevels()) {
            mCostModifierField.setText(Integer.valueOf(getCost() * getLevels()) + "%");
        } else {
            AdvantageModifierCostType costType = getCostType();
            switch (costType) {
                case POINTS -> mCostModifierField.setText(Numbers.formatWithForcedSign(getCost()));
                case MULTIPLIER -> {
                    mCostModifierField.setText(costType + Numbers.format(getCostMultiplier()));
                    mAffects.setSelectedItem(Affects.TOTAL, true);
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
        return ((Number) mCostField.getValue()).intValue();
    }

    private double getCostMultiplier() {
        return ((Number) mCostField.getValue()).doubleValue();
    }

    private int getLevels() {
        return ((Number) mLevelField.getValue()).intValue();
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
}
