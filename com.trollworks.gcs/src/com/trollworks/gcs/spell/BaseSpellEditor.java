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

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.StdLabel;
import com.trollworks.gcs.ui.widget.StdPanel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.weapon.MeleeWeaponListEditor;
import com.trollworks.gcs.weapon.RangedWeaponListEditor;

import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Set;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** An editor implementing functionalities common to all spell implementations. */
public abstract class BaseSpellEditor<T extends Spell> extends RowEditor<T> implements ActionListener, DocumentListener, EditorField.ChangeListener {
    protected EditorField                mNameField;
    protected EditorField                mCollegeField;
    protected EditorField                mPowerSourceField;
    protected EditorField                mResistField;
    protected EditorField                mClassField;
    protected EditorField                mCastingCostField;
    protected EditorField                mMaintenanceField;
    protected EditorField                mCastingTimeField;
    protected EditorField                mDurationField;
    protected JComboBox<SkillDifficulty> mDifficultyCombo;
    protected MultiLineTextField         mNotesField;
    protected EditorField                mCategoriesField;
    protected EditorField                mPointsField;
    protected EditorField                mLevelField;
    protected EditorField                mReferenceField;
    protected PrereqsPanel               mPrereqs;
    protected JCheckBox                  mHasTechLevel;
    protected EditorField                mTechLevel;
    protected String                     mSavedTechLevel;
    protected MeleeWeaponListEditor      mMeleeWeapons;
    protected RangedWeaponListEditor     mRangedWeapons;

    /**
     * Creates a new BaseSpellEditor.
     *
     * @param row The row being edited.
     */
    protected BaseSpellEditor(T row) {
        super(row);
        addContent();
    }

    protected static void determineLargest(Container panel, int every, Dimension size) {
        int count = panel.getComponentCount();
        for (int i = 0; i < count; i += every) {
            Dimension oneSize = panel.getComponent(i).getPreferredSize();
            if (oneSize.width > size.width) {
                size.width = oneSize.width;
            }
            if (oneSize.height > size.height) {
                size.height = oneSize.height;
            }
        }
    }

    protected static void applySize(Container panel, int every, Dimension size) {
        int count = panel.getComponentCount();
        for (int i = 0; i < count; i += every) {
            UIUtilities.setOnlySize(panel.getComponent(i), size);
        }
    }

    /** @return The points in the points field, as an integer. */
    protected int getPoints() {
        return ((Integer) mPointsField.getValue()).intValue();
    }

    protected int getAdjustedPoints() {
        int            points    = getPoints();
        GURPSCharacter character = mRow.getCharacter();
        if (character != null) {
            Set<String> categories = ListRow.createCategoriesList(mCategoriesField.getText());
            points += Spell.getSpellPointBonusesFor(character, Spell.ID_POINTS_COLLEGE, mCollegeField.getText(), categories, null);
            points += Spell.getSpellPointBonusesFor(character, Spell.ID_POINTS_POWER_SOURCE, mPowerSourceField.getText(), categories, null);
            points += Spell.getSpellPointBonusesFor(character, Spell.ID_POINTS, mNameField.getText(), categories, null);
            if (points < 0) {
                points = 0;
            }
        }
        return points;
    }

    /** @return The selected item of the difficulty combobox, as a SkillDifficulty. */
    protected SkillDifficulty getDifficulty() {
        return (SkillDifficulty) mDifficultyCombo.getSelectedItem();
    }

    /**
     * Utility function to create a text field (with a label) and set a few properties.
     *
     * @param labelParent Container for the label.
     * @param fieldParent Container for the text field.
     * @param title       The text of the label.
     * @param text        The text of the text field.
     * @param tooltip     The tooltip of the text field.
     * @param maxChars    The maximum number of characters that can be written in the text field.
     */
    protected EditorField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
        EditorField field = new EditorField(FieldFactory.STRING, this, SwingConstants.LEFT, text, maxChars > 0 ? Text.makeFiller(maxChars, 'M') : null, tooltip);
        addLabel(labelParent, title, field);
        PrecisionLayoutData ld = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (maxChars == 0) {
            ld.setGrabHorizontalSpace(true);
        }
        fieldParent.add(field, ld);
        return field;
    }

    /**
     * Utility function to create a text field (with a label) that accepts only integral, unsigned
     * numbers.
     *
     * @param labelParent Container for the label.
     * @param fieldParent Container for the text field.
     * @param title       The text of the label.
     * @param tooltip     The tooltip of the text field.
     * @param value       The number display in the text field.
     * @param maxValue    The maximum value the field will hold.
     */
    protected EditorField createNumberField(Container labelParent, Container fieldParent, String title, String tooltip, int value, int maxValue) {
        EditorField field = new EditorField(FieldFactory.POSINT5, this, SwingConstants.LEFT, Integer.valueOf(value), Integer.valueOf(maxValue), tooltip);
        addLabel(labelParent, title, field);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    /**
     * Utility function to create a text field (with a label) and set a few properties.
     *
     * @param labelParent Container for the label.
     * @param fieldParent Container for the text field.
     * @param title       The text of the label.
     * @param text        The text of the text field.
     * @param tooltip     The tooltip of the text field.
     */
    protected EditorField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
        EditorField field = new EditorField(FieldFactory.STRING, this, SwingConstants.LEFT, text, tooltip);
        field.getDocument().addDocumentListener(this);
        addLabel(labelParent, title, field);
        fieldParent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    /**
     * Utility function to create a combobox, populate it, and set a few properties.
     *
     * @param parent    Container for the widget.
     * @param items     Items of the combobox.
     * @param selection The item initialliy selected.
     * @param tooltip   The tooltip of the combobox.
     */
    protected <E> JComboBox<E> createComboBox(Container parent, E[] items, Object selection, String tooltip) {
        JComboBox<E> combo = new JComboBox<>(items);
        combo.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        combo.setSelectedItem(selection);
        combo.addActionListener(this);
        combo.setMaximumRowCount(items.length);
        parent.add(combo);
        return combo;
    }

    /**
     * Creates the widgets for the tech level using a standard layout.
     *
     * @param parent Container for the widgets.
     */
    protected void createTechLevelFields(Container parent) {
        mSavedTechLevel = mRow.getTechLevel();
        boolean hasTL = mSavedTechLevel != null;
        if (!hasTL) {
            mSavedTechLevel = "";
        }

        GURPSCharacter character = mRow.getCharacter();
        if (character != null) {
            StdPanel wrapper = new StdPanel(new PrecisionLayout().setMargins(0).setColumns(2));

            mHasTechLevel = new JCheckBox(I18n.text("Tech Level"), hasTL);
            UIUtilities.setToPreferredSizeOnly(mHasTechLevel);
            String tlTooltip = I18n.text("Whether this spell requires tech level specialization, and, if so, at what tech level it was learned");
            mHasTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(tlTooltip));
            mHasTechLevel.addActionListener(this);
            wrapper.add(mHasTechLevel);

            mTechLevel = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, mSavedTechLevel, "9999", tlTooltip);
            mTechLevel.setEnabled(hasTL);
            wrapper.add(mTechLevel);
            parent.add(wrapper);

            if (!hasTL) {
                mSavedTechLevel = character.getProfile().getTechLevel();
            }
        } else {
            mTechLevel = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, mSavedTechLevel, "9999", null);
            mHasTechLevel = new JCheckBox(I18n.text("Tech Level Required"), hasTL);
            mHasTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(I18n.text("Whether this spell requires tech level specialization")));
            mHasTechLevel.addActionListener(this);
            parent.add(mHasTechLevel);
        }
    }

    /**
     * Called by actionPerformed() to update the text and tooltip of the level field. Implement this
     * function by building appropriate strings and assign them to the text and tooltip of the level
     * field
     *
     * @param levelField The text field to update.
     */
    protected abstract void recalculateLevel(EditorField levelField);

    @Override
    public void actionPerformed(ActionEvent event) {
        adjustForSource(event.getSource());
    }

    protected void adjustForSource(Object src) {
        if (src == mHasTechLevel) {
            boolean enabled = mHasTechLevel.isSelected();
            mTechLevel.setEnabled(enabled);
            if (enabled) {
                mTechLevel.setText(mSavedTechLevel);
                mTechLevel.requestFocus();
            } else {
                mSavedTechLevel = mTechLevel.getText();
                mTechLevel.setText("");
            }
        } else if (src == mPointsField || src == mDifficultyCombo || src == mNameField) {
            if (mLevelField != null) {
                recalculateLevel(mLevelField);
            }
        }
    }

    /** Always call the super implementation when overriding this method. */
    @Override
    public void changedUpdate(DocumentEvent event) {
        Document doc = event.getDocument();
        if (doc == mNameField.getDocument()) {
            StdLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.text("The name field may not be empty") : null);
        } else if (doc == mClassField.getDocument()) {
            StdLabel.setErrorMessage(mClassField, mClassField.getText().trim().isEmpty() ? I18n.text("The class field may not be empty") : null);
        } else if (doc == mCastingCostField.getDocument()) {
            StdLabel.setErrorMessage(mCastingCostField, mCastingCostField.getText().trim().isEmpty() ? I18n.text("The casting cost field may not be empty") : null);
        } else if (doc == mCastingTimeField.getDocument()) {
            StdLabel.setErrorMessage(mCastingTimeField, mCastingTimeField.getText().trim().isEmpty() ? I18n.text("The casting time field may not be empty") : null);
        } else if (doc == mDurationField.getDocument()) {
            StdLabel.setErrorMessage(mDurationField, mDurationField.getText().trim().isEmpty() ? I18n.text("The duration field may not be empty") : null);
        }
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        changedUpdate(event);
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        changedUpdate(event);
    }

    @Override
    public void editorFieldChanged(EditorField field) {
        adjustForSource(field);
    }

    protected List<String> getColleges() {
        List<String> colleges = new ArrayList<>();
        for (String college : mCollegeField.getText().split(",")) {
            college = college.trim();
            if (!college.isEmpty()) {
                colleges.add(college);
            }
        }
        Collections.sort(colleges);
        return colleges;
    }
}
