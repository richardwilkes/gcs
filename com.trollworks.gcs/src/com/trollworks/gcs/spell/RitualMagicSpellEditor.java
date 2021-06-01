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

import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.weapon.MeleeWeaponListEditor;
import com.trollworks.gcs.weapon.RangedWeaponListEditor;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Container;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JPanel;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.text.Document;

/** The detailed editor for {@link RitualMagicSpell}s. */
public class RitualMagicSpellEditor extends BaseSpellEditor<RitualMagicSpell> {
    private JTextField mBaseSkillNameField;
    private JTextField mPrerequisiteSpellsCountField;


    /**
     * Creates a new {@link Spell} {@link RowEditor}.
     *
     * @param spell The row being edited.
     */
    protected RitualMagicSpellEditor(RitualMagicSpell spell) {
        super(spell);
    }

    @Override
    protected void addContentSelf(ScrollContent outer) {
        outer.add(createTop(), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mPrereqs = new PrereqsPanel(mRow, mRow.getPrereqs());
        addSection(outer, mPrereqs);
        List<WeaponStats> weapons = mRow.getWeapons();
        mMeleeWeapons = new MeleeWeaponListEditor(mRow, weapons);
        addSection(outer, mMeleeWeapons);
        mRangedWeapons = new RangedWeaponListEditor(mRow, weapons);
        addSection(outer, mRangedWeapons);
    }

    private JPanel createTop() {
        JPanel panel   = new JPanel(new PrecisionLayout().setMargins(0).setColumns(4));
        JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        mNameField = createCorrectableField(panel, wrapper, I18n.Text("Name"), mRow.getName(), I18n.Text("The name of the spell, without any notes"));
        createTechLevelFields(wrapper);
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        mBaseSkillNameField = createCorrectableField(panel, panel, I18n.Text("Base Skill"), mRow.getBaseSkillName(), I18n.Text("The name of the base skill, such as \"Ritual Magic\" or \"Thaumatology\""));
        mPrerequisiteSpellsCountField = createNumberField(panel, panel, I18n.Text("Prerequisite Count"), I18n.Text("The penalty to skill level based on the number of prerequisite spells"), mRow.getPrerequisiteSpellsCount(), 2);
        mCollegeField = createField(panel, panel, I18n.Text("College"), String.join(", ", mRow.getColleges()), I18n.Text("The college(s) the spell belongs to; separate multiple colleges with a comma"), 0);
        mPowerSourceField = createField(panel, panel, I18n.Text("Power Source"), mRow.getPowerSource(), I18n.Text("The source of power for the spell"), 0);
        mClassField = createCorrectableField(panel, panel, I18n.Text("Class"), mRow.getSpellClass(), I18n.Text("The class of spell (Area, Missile, etc.)"));
        mResistField = createCorrectableField(panel, panel, I18n.Text("Resistance"), mRow.getResist(), I18n.Text("The resistance roll, if any"));
        mCastingCostField = createCorrectableField(panel, panel, I18n.Text("Casting Cost"), mRow.getCastingCost(), I18n.Text("The casting cost of the spell"));
        mCastingTimeField = createCorrectableField(panel, panel, I18n.Text("Casting Time"), mRow.getCastingTime(), I18n.Text("The casting time of the spell"));
        mMaintenanceField = createField(panel, panel, I18n.Text("Maintenance Cost"), mRow.getMaintenance(), I18n.Text("The cost to maintain a spell after its initial duration"), 0);
        mDurationField = createCorrectableField(panel, panel, I18n.Text("Duration"), mRow.getDuration(), I18n.Text("The duration of the spell once its cast"));
        createPointsFields(panel);
        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this spell"), this);
        panel.add(new LinkedLabel(I18n.Text("Notes"), mNotesField), new PrecisionLayoutData().setBeginningVerticalAlignment().setFillHorizontalAlignment().setTopMargin(2));
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        wrapper = new JPanel(new PrecisionLayout().setMargins(0));
        mCategoriesField = createField(panel, wrapper, I18n.Text("Categories"), mRow.getCategoriesAsString(), I18n.Text("The category or categories the spell belongs to (separate multiple categories with a comma)"), 0);
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        wrapper = new JPanel(new PrecisionLayout().setMargins(0));
        mReferenceField = createField(panel, wrapper, I18n.Text("Page Reference"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.Text("spell")), 0);
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        return panel;
    }

    private void createPointsFields(Container parent) {
        boolean forCharacter = mRow.getCharacter() != null;
        boolean forTemplate  = mRow.getTemplate() != null;
        int     columns      = 1;
        if (forCharacter || forTemplate) {
            columns += 2;
        }
        if (forCharacter) {
            columns += 2;
        }
        JPanel panel = new JPanel(new PrecisionLayout().setMargins(0).setColumns(columns));
        mDifficultyCombo = createComboBox(panel, new SkillDifficulty[]{SkillDifficulty.A, SkillDifficulty.H}, mRow.getDifficulty(), I18n.Text("The difficulty of the spell"));
        if (forCharacter || forTemplate) {
            mPointsField = createNumberField(panel, panel, I18n.Text("Points"), I18n.Text("The number of points spent on this spell"), mRow.getRawPoints(), 4);
            if (forCharacter) {
                mLevelField = createField(panel, panel, I18n.Text("Level"), getDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel()), I18n.Text("The spell level and relative spell level to roll against.\n") + mRow.getLevelToolTip(), 7);
                mLevelField.setEnabled(false);
            }
        }
        addLabel(parent, I18n.Text("Difficulty"), null);
        parent.add(panel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
    }

    public static String getDisplayLevel(int level, int relativeLevel) {
        if (level < 0) {
            return "-";
        }
        return Numbers.format(level) + "/" + Numbers.formatWithForcedSign(relativeLevel);
    }

    protected int getPrerequisiteSpellsCount() {
        return Numbers.extractInteger(mPrerequisiteSpellsCountField.getText(), 0, true);
    }

    @Override
    protected boolean applyChangesSelf() {
        boolean modified = mRow.setName(mNameField.getText());
        modified |= mRow.setBaseSkillName(mBaseSkillNameField.getText());
        modified |= mRow.setReference(mReferenceField.getText());
        if (mHasTechLevel != null) {
            modified |= mRow.setTechLevel(mHasTechLevel.isSelected() ? mTechLevel.getText() : null);
        }
        modified |= mRow.setColleges(getColleges());
        modified |= mRow.setPowerSource(mPowerSourceField.getText());
        modified |= mRow.setSpellClass(mClassField.getText());
        modified |= mRow.setCastingCost(mCastingCostField.getText());
        modified |= mRow.setMaintenance(mMaintenanceField.getText());
        modified |= mRow.setCastingTime(mCastingTimeField.getText());
        modified |= mRow.setResist(mResistField.getText());
        modified |= mRow.setDuration(mDurationField.getText());
        modified |= mRow.setDifficulty("iq", getDifficulty()); // Attribute is not relevant for Ritual Magic
        modified |= mRow.setPrerequisiteSpellsCount(getPrerequisiteSpellsCount());
        if (mRow.getCharacter() != null || mRow.getTemplate() != null) {
            modified |= mRow.setRawPoints(getPoints());
        }
        modified |= mRow.setNotes(mNotesField.getText());
        modified |= mRow.setCategories(mCategoriesField.getText());
        modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
        List<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
        list.addAll(mRangedWeapons.getWeapons());
        modified |= mRow.setWeapons(list);
        return modified;
    }

    protected void recalculateLevel(JTextField levelField) {
        SkillLevel level = RitualMagicSpell.calculateLevel(mRow.getCharacter(), mNameField.getText(), mBaseSkillNameField.getText(), getColleges(), mPowerSourceField.getText(), ListRow.createCategoriesList(mCategoriesField.getText()), getDifficulty(), getPrerequisiteSpellsCount(), getAdjustedPoints());
        levelField.setText(getDisplayLevel(level.getLevel(), level.getRelativeLevel()));
        levelField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The spell level and relative spell level to roll against.\n") + level.getToolTip()));
    }

    @Override
    protected void adjustForSource(Object src) {
        if (src == mPrerequisiteSpellsCountField || src == mBaseSkillNameField) {
            if (mLevelField != null) {
                recalculateLevel(mLevelField);
            }
        } else {
            super.adjustForSource(src);
        }
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document doc = event.getDocument();
        if (doc == mBaseSkillNameField.getDocument()) {
            LinkedLabel.setErrorMessage(mBaseSkillNameField, mBaseSkillNameField.getText().trim().isEmpty() ? I18n.Text("The base skill field may not be empty") : null);
        } else {
            super.changedUpdate(event);
        }
    }
}
