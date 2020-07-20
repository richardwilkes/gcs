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

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.NumberFilter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.util.Set;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTabbedPane;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** An editor implementing functionalities common to all spell implementations. */
public abstract class BaseSpellEditor<T extends Spell> extends RowEditor<T> implements ActionListener, DocumentListener, FocusListener {
    protected JTextField                 mNameField;
    protected JTextField                 mCollegeField;
    protected JTextField                 mPowerSourceField;
    protected JTextField                 mResistField;
    protected JTextField                 mClassField;
    protected JTextField                 mCastingCostField;
    protected JTextField                 mMaintenanceField;
    protected JTextField                 mCastingTimeField;
    protected JTextField                 mDurationField;
    protected JComboBox<SkillDifficulty> mDifficultyCombo;
    protected JTextField                 mNotesField;
    protected JTextField                 mCategoriesField;
    protected JTextField                 mPointsField;
    protected JTextField                 mLevelField;
    protected JTextField                 mReferenceField;
    protected JTabbedPane                mTabPanel;
    protected PrereqsPanel               mPrereqs;
    protected JCheckBox                  mHasTechLevel;
    protected JTextField                 mTechLevel;
    protected String                     mSavedTechLevel;
    protected MeleeWeaponEditor          mMeleeWeapons;
    protected RangedWeaponEditor         mRangedWeapons;

    /**
     * Creates a new {@link BaseSpellEditor}.
     *
     * @param row The row being edited.
     */
    protected BaseSpellEditor(T row) {
        super(row);
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
        return Numbers.extractInteger(mPointsField.getText(), 0, true);
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

    protected JScrollPane embedEditor(Component editor) {
        JScrollPane scrollPanel = new JScrollPane(editor);
        scrollPanel.setMinimumSize(new Dimension(500, 120));
        scrollPanel.setName(editor.toString());
        if (!mIsEditable) {
            UIUtilities.disableControls(editor);
        }
        return scrollPanel;
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
    protected JTextField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
        JTextField field = new JTextField(maxChars > 0 ? Text.makeFiller(maxChars, 'M') : text);
        if (maxChars > 0) {
            UIUtilities.setToPreferredSizeOnly(field);
            field.setText(text);
        }
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
        field.addActionListener(this);
        field.addFocusListener(this);
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
     * @param maxDigits   The maximum number of digits to display.
     */
    protected JTextField createNumberField(Container labelParent, Container fieldParent, String title, String tooltip, int value, int maxDigits) {
        JTextField field = createField(labelParent, fieldParent, title, Numbers.format(value), tooltip, maxDigits + 1);
        NumberFilter.apply(field, false, false, false, maxDigits);
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
    protected JTextField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        field.getDocument().addDocumentListener(this);
        field.addActionListener(this);
        field.addFocusListener(this);

        LinkedLabel label = new LinkedLabel(title);
        label.setLink(field);

        labelParent.add(label);
        fieldParent.add(field);
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
        UIUtilities.setToPreferredSizeOnly(combo);
        combo.setEnabled(mIsEditable);
        parent.add(combo);
        return combo;
    }

    /**
     * Creates the widgets for the tech level using a standard layout.
     *
     * @param parent Container for the widgets.
     */
    protected void createTechLevelFields(Container parent) {
        OutlineModel   owner     = mRow.getOwner();
        GURPSCharacter character = mRow.getCharacter();
        boolean        enabled   = !owner.isLocked();
        boolean        hasTL;

        mSavedTechLevel = mRow.getTechLevel();
        hasTL = mSavedTechLevel != null;
        if (!hasTL) {
            mSavedTechLevel = "";
        }

        if (character != null) {
            JPanel wrapper = new JPanel(new ColumnLayout(2));

            mHasTechLevel = new JCheckBox(I18n.Text("Tech Level"), hasTL);
            UIUtilities.setToPreferredSizeOnly(mHasTechLevel);
            String tlTooltip = I18n.Text("Whether this spell requires tech level specialization, and, if so, at what tech level it was learned");
            mHasTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(tlTooltip));
            mHasTechLevel.setEnabled(enabled);
            mHasTechLevel.addActionListener(this);
            wrapper.add(mHasTechLevel);

            mTechLevel = new JTextField("9999");
            UIUtilities.setToPreferredSizeOnly(mTechLevel);
            mTechLevel.setText(mSavedTechLevel);
            mTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(tlTooltip));
            mTechLevel.setEnabled(enabled && hasTL);
            wrapper.add(mTechLevel);
            UIUtilities.setToPreferredSizeOnly(wrapper);
            parent.add(wrapper);

            if (!hasTL) {
                mSavedTechLevel = character.getProfile().getTechLevel();
            }
        } else {
            mTechLevel = new JTextField(mSavedTechLevel);
            mHasTechLevel = new JCheckBox(I18n.Text("Tech Level Required"), hasTL);
            mHasTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Whether this spell requires tech level specialization")));
            mHasTechLevel.setEnabled(enabled);
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
    protected abstract void recalculateLevel(JTextField levelField);

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

    @Override
    public void finished() {
        if (mTabPanel != null) {
            updateLastTabName(mTabPanel.getTitleAt(mTabPanel.getSelectedIndex()));
        }
    }

    /** Always call the super implementation when overriding this method. */
    @Override
    public void changedUpdate(DocumentEvent event) {
        Document doc = event.getDocument();
        if (doc == mNameField.getDocument()) {
            LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.Text("The name field may not be empty") : null);
        } else if (doc == mClassField.getDocument()) {
            LinkedLabel.setErrorMessage(mClassField, mClassField.getText().trim().isEmpty() ? I18n.Text("The class field may not be empty") : null);
        } else if (doc == mCastingCostField.getDocument()) {
            LinkedLabel.setErrorMessage(mCastingCostField, mCastingCostField.getText().trim().isEmpty() ? I18n.Text("The casting cost field may not be empty") : null);
        } else if (doc == mCastingTimeField.getDocument()) {
            LinkedLabel.setErrorMessage(mCastingTimeField, mCastingTimeField.getText().trim().isEmpty() ? I18n.Text("The casting time field may not be empty") : null);
        } else if (doc == mDurationField.getDocument()) {
            LinkedLabel.setErrorMessage(mDurationField, mDurationField.getText().trim().isEmpty() ? I18n.Text("The duration field may not be empty") : null);
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
    public void focusGained(FocusEvent event) {
        // Nothing to do
    }

    @Override
    public void focusLost(FocusEvent event) {
        adjustForSource(event.getSource());
    }
}
