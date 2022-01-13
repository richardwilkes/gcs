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

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.weapon.MeleeWeaponListEditor;
import com.trollworks.gcs.weapon.RangedWeaponListEditor;

import java.awt.Container;
import java.awt.Dimension;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Set;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** An editor implementing functionalities common to all spell implementations. */
public abstract class BaseSpellEditor<T extends Spell> extends RowEditor<T> implements DocumentListener {
    protected EditorField                mNameField;
    protected EditorField                mCollegeField;
    protected EditorField                mPowerSourceField;
    protected EditorField                mResistField;
    protected EditorField                mClassField;
    protected EditorField                mCastingCostField;
    protected EditorField                mMaintenanceField;
    protected EditorField                mCastingTimeField;
    protected EditorField                mDurationField;
    protected PopupMenu<SkillDifficulty> mDifficultyPopup;
    protected MultiLineTextField         mNotesField;
    protected MultiLineTextField         mVTTNotesField;
    protected EditorField                mCategoriesField;
    protected EditorField                mPointsField;
    protected EditorField                mLevelField;
    protected EditorField                mReferenceField;
    protected PrereqsPanel               mPrereqs;
    protected Checkbox                   mHasTechLevel;
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

    /** @return The selected item of the difficulty popup, as a SkillDifficulty. */
    protected SkillDifficulty getDifficulty() {
        return mDifficultyPopup.getSelectedItem();
    }

    /**
     * Utility function to create a text field and set a few properties.
     *
     * @param parent   Container for the text field.
     * @param text     The text of the text field.
     * @param tooltip  The tooltip of the text field.
     * @param maxChars The maximum number of characters that can be written in the text field.
     */
    protected EditorField createField(Container parent, String text, String tooltip, int maxChars) {
        EditorField         field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text, maxChars > 0 ? Text.makeFiller(maxChars, 'M') : null, tooltip);
        PrecisionLayoutData ld    = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (maxChars == 0) {
            ld.setGrabHorizontalSpace(true);
        }
        parent.add(field, ld);
        return field;
    }

    /**
     * Utility function to create a text field that accepts only integral, unsigned numbers.
     *
     * @param parent   Container for the text field.
     * @param tooltip  The tooltip of the text field.
     * @param value    The number display in the text field.
     * @param maxValue The maximum value the field will hold.
     */
    protected EditorField createNumberField(Container parent, String tooltip, int value, int maxValue, EditorField.ChangeListener listener) {
        EditorField field = new EditorField(FieldFactory.POSINT5, listener, SwingConstants.LEFT, Integer.valueOf(value), Integer.valueOf(maxValue), tooltip);
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    /**
     * Utility function to create a text field and set a few properties.
     *
     * @param parent  Container for the text field.
     * @param text    The text of the text field.
     * @param tooltip The tooltip of the text field.
     */
    protected EditorField createCorrectableField(Container parent, String text, String tooltip, EditorField.ChangeListener listener) {
        EditorField field = new EditorField(FieldFactory.STRING, listener, SwingConstants.LEFT, text, tooltip);
        field.getDocument().addDocumentListener(this);
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    /**
     * Utility function to create a PopupMenu, populate it, and set a few properties.
     *
     * @param parent    Container for the widget.
     * @param items     Items of the PopupMenu.
     * @param selection The item initialliy selected.
     * @param tooltip   The tooltip of the PopupMenu.
     */
    protected <E> PopupMenu<E> createPopupMenu(Container parent, E[] items, E selection, String tooltip, PopupMenu.SelectionListener<E> listener) {
        PopupMenu<E> popup = new PopupMenu<>(items, listener);
        popup.setToolTipText(tooltip);
        popup.setSelectedItem(selection, false);
        parent.add(popup);
        return popup;
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
            Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));

            mHasTechLevel = new Checkbox(I18n.text("Tech Level"), hasTL, this::clickedOnHasTechLevel);
            String tlTooltip = I18n.text("Whether this spell requires tech level specialization, and, if so, at what tech level it was learned");
            mHasTechLevel.setToolTipText(tlTooltip);
            wrapper.add(mHasTechLevel, new PrecisionLayoutData().setLeftMargin(4));

            mTechLevel = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, mSavedTechLevel, "9999", tlTooltip);
            mTechLevel.setEnabled(hasTL);
            wrapper.add(mTechLevel);
            parent.add(wrapper);

            if (!hasTL) {
                mSavedTechLevel = character.getProfile().getTechLevel();
            }
        } else {
            mTechLevel = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, mSavedTechLevel, "9999", null);
            mHasTechLevel = new Checkbox(I18n.text("Tech Level Required"), hasTL, this::clickedOnHasTechLevel);
            mHasTechLevel.setToolTipText(I18n.text("Whether this spell requires tech level specialization"));
            parent.add(mHasTechLevel, new PrecisionLayoutData().setLeftMargin(4));
        }
    }

    private void clickedOnHasTechLevel(Checkbox checkbox) {
        boolean enabled = checkbox.isChecked();
        mTechLevel.setEnabled(enabled);
        if (enabled) {
            mTechLevel.setText(mSavedTechLevel);
            mTechLevel.requestFocus();
        } else {
            mSavedTechLevel = mTechLevel.getText();
            mTechLevel.setText("");
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

    /** Always call the super implementation when overriding this method. */
    @Override
    public void changedUpdate(DocumentEvent event) {
        Document doc = event.getDocument();
        if (doc == mNameField.getDocument()) {
            mNameField.setErrorMessage(mNameField.getText().trim().isEmpty() ? I18n.text("The name field may not be empty") : null);
        } else if (mClassField != null && doc == mClassField.getDocument()) {
            mClassField.setErrorMessage(mClassField.getText().trim().isEmpty() ? I18n.text("The class field may not be empty") : null);
        } else if (mCastingCostField != null && doc == mCastingCostField.getDocument()) {
            mCastingCostField.setErrorMessage(mCastingCostField.getText().trim().isEmpty() ? I18n.text("The casting cost field may not be empty") : null);
        } else if (mCastingTimeField != null && doc == mCastingTimeField.getDocument()) {
            mCastingTimeField.setErrorMessage(mCastingTimeField.getText().trim().isEmpty() ? I18n.text("The casting time field may not be empty") : null);
        } else if (mDurationField != null && doc == mDurationField.getDocument()) {
            mDurationField.setErrorMessage(mDurationField.getText().trim().isEmpty() ? I18n.text("The duration field may not be empty") : null);
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
