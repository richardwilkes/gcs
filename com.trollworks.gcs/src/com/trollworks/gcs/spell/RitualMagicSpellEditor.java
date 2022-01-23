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

import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.weapon.MeleeWeaponListEditor;
import com.trollworks.gcs.weapon.RangedWeaponListEditor;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Container;
import java.util.ArrayList;
import java.util.List;
import javax.swing.event.DocumentEvent;
import javax.swing.text.Document;

/** The detailed editor for {@link RitualMagicSpell}s. */
public class RitualMagicSpellEditor extends BaseSpellEditor<RitualMagicSpell> {
    private EditorField mBaseSkillNameField;
    private EditorField mPrerequisiteSpellsCountField;

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

    private Panel createTop() {
        Panel panel   = new Panel(new PrecisionLayout().setMargins(0).setColumns(4));
        Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
        addLabel(panel, I18n.text("Name"));
        mNameField = createCorrectableField(wrapper, mRow.getName(),
                I18n.text("The name of the spell, without any notes"),
                (f) -> recalculateLevel(mLevelField));
        createTechLevelFields(wrapper);
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        addLabel(panel, I18n.text("Base Skill"));
        mBaseSkillNameField = createCorrectableField(panel, mRow.getBaseSkillName(),
                I18n.text("The name of the base skill, such as \"Ritual Magic\" or \"Thaumatology\""),
                (f) -> recalculateLevel(mLevelField));
        addInteriorLabel(panel, I18n.text("Prerequisite Count"));
        mPrerequisiteSpellsCountField = createNumberField(panel,
                I18n.text("The penalty to skill level based on the number of prerequisite spells"),
                mRow.getPrerequisiteSpellsCount(), 99, (f) -> recalculateLevel(mLevelField));
        addLabel(panel, I18n.text("College"));
        mCollegeField = createField(panel, String.join(", ", mRow.getColleges()),
                I18n.text("The college(s) the spell belongs to; separate multiple colleges with a comma"),
                0);
        addInteriorLabel(panel, I18n.text("Power Source"));
        mPowerSourceField = createField(panel, mRow.getPowerSource(),
                I18n.text("The source of power for the spell"), 0);
        addLabel(panel, I18n.text("Class"));
        mClassField = createCorrectableField(panel, mRow.getSpellClass(),
                I18n.text("The class of spell (Area, Missile, etc.)"), null);
        addInteriorLabel(panel, I18n.text("Resistance"));
        mResistField = createCorrectableField(panel, mRow.getResist(),
                I18n.text("The resistance roll, if any"), null);
        addLabel(panel, I18n.text("Casting Cost"));
        mCastingCostField = createCorrectableField(panel, mRow.getCastingCost(),
                I18n.text("The casting cost of the spell"), null);
        addInteriorLabel(panel, I18n.text("Casting Time"));
        mCastingTimeField = createCorrectableField(panel, mRow.getCastingTime(),
                I18n.text("The casting time of the spell"), null);
        addLabel(panel, I18n.text("Maintenance Cost"));
        mMaintenanceField = createField(panel, mRow.getMaintenance(),
                I18n.text("The cost to maintain a spell after its initial duration"), 0);
        addInteriorLabel(panel, I18n.text("Duration"));
        mDurationField = createCorrectableField(panel, mRow.getDuration(),
                I18n.text("The duration of the spell once its cast"), null);
        createPointsFields(panel);
        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.text("Any notes that you would like to show up in the list along with this spell"), this);
        addLabel(panel, I18n.text("Notes")).setBeginningVerticalAlignment().setTopMargin(2);
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        mVTTNotesField = addVTTNotesField(panel, 3, this);
        addLabel(panel, I18n.text("Categories"));
        wrapper = new Panel(new PrecisionLayout().setMargins(0));
        mCategoriesField = createField(wrapper, mRow.getCategoriesAsString(),
                I18n.text("The category or categories the spell belongs to (separate multiple categories with a comma)"),
                0);
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        addLabel(panel, I18n.text("Page Reference"));
        wrapper = new Panel(new PrecisionLayout().setMargins(0));
        mReferenceField = createField(wrapper, mRow.getReference(),
                PageRefCell.getStdToolTip(I18n.text("spell")), 0);
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
        Panel panel = new Panel(new PrecisionLayout().setMargins(0).setColumns(columns));
        mDifficultyPopup = createPopupMenu(panel, new SkillDifficulty[]{SkillDifficulty.A,
                        SkillDifficulty.H}, mRow.getDifficulty(),
                I18n.text("The difficulty of the spell"), (p) -> recalculateLevel(mLevelField));
        if (forCharacter || forTemplate) {
            addInteriorLabel(panel, I18n.text("Points"));
            mPointsField = createNumberField(panel,
                    I18n.text("The number of points spent on this spell"), mRow.getRawPoints(),
                    9999, (f) -> recalculateLevel(mLevelField));
            if (forCharacter) {
                addInteriorLabel(panel, I18n.text("Level"));
                mLevelField = createField(panel, getDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel()),
                        I18n.text("The spell level and relative spell level to roll against.\n") + mRow.getLevelToolTip(),
                        7);
                mLevelField.setEnabled(false);
            }
        }
        addLabel(parent, I18n.text("Difficulty"));
        parent.add(panel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
    }

    public static String getDisplayLevel(int level, int relativeLevel) {
        if (level < 0) {
            return "-";
        }
        return Numbers.format(level) + "/" + Numbers.formatWithForcedSign(relativeLevel);
    }

    protected int getPrerequisiteSpellsCount() {
        return ((Integer) mPrerequisiteSpellsCountField.getValue()).intValue();
    }

    @Override
    protected boolean applyChangesSelf() {
        boolean modified = mRow.setName(mNameField.getText());
        modified |= mRow.setBaseSkillName(mBaseSkillNameField.getText());
        modified |= mRow.setReference(mReferenceField.getText());
        if (mHasTechLevel != null) {
            modified |= mRow.setTechLevel(mHasTechLevel.isChecked() ? mTechLevel.getText() : null);
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
        modified |= mRow.setVTTNotes(mVTTNotesField.getText());
        modified |= mRow.setCategories(mCategoriesField.getText());
        modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
        List<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
        list.addAll(mRangedWeapons.getWeapons());
        modified |= mRow.setWeapons(list);
        return modified;
    }

    protected void recalculateLevel(EditorField levelField) {
        if (levelField != null) {
            SkillLevel level = RitualMagicSpell.calculateLevel(mRow.getCharacter(),
                    mNameField.getText(), mBaseSkillNameField.getText(), getColleges(),
                    mPowerSourceField.getText(),
                    ListRow.createCategoriesList(mCategoriesField.getText()), getDifficulty(),
                    getPrerequisiteSpellsCount(), getAdjustedPoints());
            levelField.setText(getDisplayLevel(level.getLevel(), level.getRelativeLevel()));
            levelField.setToolTipText(I18n.text("The spell level and relative spell level to roll against.\n") + level.getToolTip());
        }
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document doc = event.getDocument();
        if (doc == mBaseSkillNameField.getDocument()) {
            mBaseSkillNameField.setErrorMessage(mBaseSkillNameField.getText().trim().isEmpty() ? I18n.text("The base skill field may not be empty") : null);
        } else {
            super.changedUpdate(event);
        }
    }
}
