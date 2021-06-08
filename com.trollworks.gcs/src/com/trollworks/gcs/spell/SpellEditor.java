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

import com.trollworks.gcs.attribute.AttributeChoice;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.weapon.MeleeWeaponListEditor;
import com.trollworks.gcs.weapon.RangedWeaponListEditor;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Container;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JTextField;

/** The detailed editor for {@link Spell}s. */
public class SpellEditor extends BaseSpellEditor<Spell> {
    protected JComboBox<AttributeChoice> mAttributePopup;

    /**
     * Creates a new {@link Spell} editor.
     *
     * @param spell The {@link Spell} to edit.
     */
    public SpellEditor(Spell spell) {
        super(spell);
    }

    @Override
    protected void addContentSelf(ScrollContent outer) {
        outer.add(createTop(), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        if (!mRow.canHaveChildren()) {
            mPrereqs = new PrereqsPanel(mRow, mRow.getPrereqs());
            addSection(outer, mPrereqs);
            List<WeaponStats> weapons = mRow.getWeapons();
            mMeleeWeapons = new MeleeWeaponListEditor(mRow, weapons);
            addSection(outer, mMeleeWeapons);
            mRangedWeapons = new RangedWeaponListEditor(mRow, weapons);
            addSection(outer, mRangedWeapons);
        }
    }

    private Panel createTop() {
        boolean notContainer = !mRow.canHaveChildren();
        Panel   panel        = new Panel(new PrecisionLayout().setMargins(0).setColumns(4));
        Panel   wrapper      = new Panel(new PrecisionLayout().setMargins(0).setColumns(notContainer ? 2 : 1));
        mNameField = createCorrectableField(panel, wrapper, I18n.text("Name"), mRow.getName(), I18n.text("The name of the spell, without any notes"));
        if (notContainer) {
            createTechLevelFields(wrapper);
        }
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        if (notContainer) {
            mCollegeField = createField(panel, panel, I18n.text("College"), String.join(", ", mRow.getColleges()), I18n.text("The college(s) the spell belongs to; separate multiple colleges with a comma"), 0);
            mPowerSourceField = createField(panel, panel, I18n.text("Power Source"), mRow.getPowerSource(), I18n.text("The source of power for the spell"), 0);
            mClassField = createCorrectableField(panel, panel, I18n.text("Class"), mRow.getSpellClass(), I18n.text("The class of spell (Area, Missile, etc.)"));
            mResistField = createCorrectableField(panel, panel, I18n.text("Resistance"), mRow.getResist(), I18n.text("The resistance roll, if any"));
            mCastingCostField = createCorrectableField(panel, panel, I18n.text("Casting Cost"), mRow.getCastingCost(), I18n.text("The casting cost of the spell"));
            mCastingTimeField = createCorrectableField(panel, panel, I18n.text("Casting Time"), mRow.getCastingTime(), I18n.text("The casting time of the spell"));
            mMaintenanceField = createField(panel, panel, I18n.text("Maintenance Cost"), mRow.getMaintenance(), I18n.text("The cost to maintain a spell after its initial duration"), 0);
            mDurationField = createCorrectableField(panel, panel, I18n.text("Duration"), mRow.getDuration(), I18n.text("The duration of the spell once its cast"));
            createPointsFields(panel);
        }
        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.text("Any notes that you would like to show up in the list along with this spell"), this);
        panel.add(new LinkedLabel(I18n.text("Notes"), mNotesField), new PrecisionLayoutData().setBeginningVerticalAlignment().setFillHorizontalAlignment().setTopMargin(2));
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        wrapper = new Panel(new PrecisionLayout().setMargins(0));
        mCategoriesField = createField(panel, wrapper, I18n.text("Categories"), mRow.getCategoriesAsString(), I18n.text("The category or categories the spell belongs to (separate multiple categories with a comma)"), 0);
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        wrapper = new Panel(new PrecisionLayout().setMargins(0));
        mReferenceField = createField(panel, wrapper, I18n.text("Page Reference"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.text("spell")), 0);
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        return panel;
    }

    private void createPointsFields(Container parent) {
        List<AttributeChoice> list = new ArrayList<>();
        for (AttributeDef def : AttributeDef.getOrdered(mRow.getDataFile().getSheetSettings().getAttributes())) {
            list.add(new AttributeChoice(def.getID(), "%s", def.getName()));
        }
        list.add(new AttributeChoice("10", "%s", "10"));
        String          currentAttr = mRow.getAttribute();
        AttributeChoice current     = null;
        for (AttributeChoice attributeChoice : list) {
            if (attributeChoice.getAttribute().equals(currentAttr)) {
                current = attributeChoice;
                break;
            }
        }
        if (current == null) {
            list.add(new AttributeChoice(currentAttr, "%s", currentAttr));
            current = list.get(list.size() - 1);
        }

        boolean forCharacter = mRow.getCharacter() != null;
        boolean forTemplate  = mRow.getTemplate() != null;
        int     columns      = 3;
        if (forCharacter || forTemplate) {
            columns += 2;
        }
        if (forCharacter) {
            columns += 2;
        }
        Panel panel = new Panel(new PrecisionLayout().setMargins(0).setColumns(columns));
        mAttributePopup = createComboBox(panel, list.toArray(new AttributeChoice[0]), current, I18n.text("The attribute this spell is based on"));
        panel.add(new JLabel("/"));
        mDifficultyCombo = createComboBox(panel, SkillDifficulty.values(), mRow.getDifficulty(), I18n.text("The difficulty of the spell"));
        if (forCharacter || forTemplate) {
            mPointsField = createNumberField(panel, panel, I18n.text("Points"), I18n.text("The number of points spent on this spell"), mRow.getRawPoints(), 4);
            if (forCharacter) {
                mLevelField = createField(panel, panel, I18n.text("Level"), getDisplayLevel(mRow.getAttribute(), mRow.getLevel(), mRow.getRelativeLevel()), I18n.text("The spell level and relative spell level to roll against.\n") + mRow.getLevelToolTip(), 7);
                mLevelField.setEnabled(false);
            }
        }

        addLabel(parent, I18n.text("Difficulty"), null);
        parent.add(panel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
    }

    protected String getAttribute() {
        AttributeChoice choice = (AttributeChoice) mAttributePopup.getSelectedItem();
        return choice != null ? choice.getAttribute() : Skill.getDefaultAttribute("iq");
    }

    private String getDisplayLevel(String attribute, int level, int relativeLevel) {
        if (level < 0) {
            return "-";
        }
        return Numbers.format(level) + "/" + Skill.resolveAttributeName(mRow.getDataFile(), attribute) +
                Numbers.formatWithForcedSign(relativeLevel);
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified     = mRow.setName(mNameField.getText());
        boolean notContainer = !mRow.canHaveChildren();

        modified |= mRow.setReference(mReferenceField.getText());
        if (notContainer) {
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
            modified |= mRow.setDifficulty(getAttribute(), getDifficulty());
            if (mRow.getCharacter() != null || mRow.getTemplate() != null) {
                modified |= mRow.setRawPoints(getPoints());
            }
        }
        modified |= mRow.setNotes(mNotesField.getText());
        modified |= mRow.setCategories(mCategoriesField.getText());
        if (mPrereqs != null) {
            modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
        }
        if (mMeleeWeapons != null) {
            List<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
            list.addAll(mRangedWeapons.getWeapons());
            modified |= mRow.setWeapons(list);
        }
        return modified;
    }

    @Override
    public void adjustForSource(Object src) {
        if (src == mAttributePopup) {
            if (mLevelField != null) {
                recalculateLevel(mLevelField);
            }
        } else {
            super.adjustForSource(src);
        }
    }

    @Override
    protected void recalculateLevel(JTextField levelField) {
        String attribute = getAttribute();
        SkillLevel level = Spell.calculateLevel(mRow.getCharacter(), getAdjustedPoints(),
                attribute, getDifficulty(), getColleges(), mPowerSourceField.getText(),
                mNameField.getText(), ListRow.createCategoriesList(mCategoriesField.getText()));
        levelField.setText(getDisplayLevel(attribute, level.mLevel, level.mRelativeLevel));
        levelField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.text("The spell level and relative spell level to roll against.\n") + level.getToolTip()));
    }
}
