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

import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.SkillAttribute;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JTabbedPane;
import javax.swing.JTextField;
import javax.swing.SwingConstants;

/** The detailed editor for {@link Spell}s. */
public class SpellEditor extends BaseSpellEditor<Spell> {
    protected JComboBox<SkillAttribute> mAttributePopup;

    /**
     * Creates a new {@link Spell} editor.
     *
     * @param spell The {@link Spell} to edit.
     */
    public SpellEditor(Spell spell) {
        super(spell);

        boolean   notContainer = !spell.canHaveChildren();
        Container content      = new JPanel(new ColumnLayout(2));
        Container fields       = new JPanel(new ColumnLayout());
        Container wrapper1     = new JPanel(new ColumnLayout(notContainer ? 3 : 2));
        Container wrapper2     = new JPanel(new ColumnLayout(4));
        Container wrapper3     = new JPanel(new ColumnLayout(2));
        Container noGapWrapper = new JPanel(new ColumnLayout(2, 0, 0));
        Container ptsPanel     = null;
        JLabel    icon         = new JLabel(spell.getIcon(true));
        Dimension size         = new Dimension();
        Container refParent    = wrapper3;

        mNameField = createCorrectableField(wrapper1, wrapper1, I18n.Text("Name"), spell.getName(), I18n.Text("The name of the spell, without any notes"));
        fields.add(wrapper1);
        if (notContainer) {
            createTechLevelFields(wrapper1);
            mCollegeField = createField(wrapper2, wrapper2, I18n.Text("College"), spell.getCollege(), I18n.Text("The college the spell belongs to"), 0);
            mPowerSourceField = createField(wrapper2, wrapper2, I18n.Text("Power Source"), spell.getPowerSource(), I18n.Text("The source of power for the spell"), 0);
            mClassField = createCorrectableField(wrapper2, wrapper2, I18n.Text("Class"), spell.getSpellClass(), I18n.Text("The class of spell (Area, Missile, etc.)"));
            mCastingCostField = createCorrectableField(wrapper2, wrapper2, I18n.Text("Casting Cost"), spell.getCastingCost(), I18n.Text("The casting cost of the spell"));
            mMaintenanceField = createField(wrapper2, wrapper2, I18n.Text("Maintenance Cost"), spell.getMaintenance(), I18n.Text("The cost to maintain a spell after its initial duration"), 0);
            mCastingTimeField = createCorrectableField(wrapper2, wrapper2, I18n.Text("Casting Time"), spell.getCastingTime(), I18n.Text("The casting time of the spell"));
            mResistField = createCorrectableField(wrapper2, wrapper2, I18n.Text("Resist"), spell.getResist(), I18n.Text("The resistance roll, if any"));
            mDurationField = createCorrectableField(wrapper2, wrapper2, I18n.Text("Duration"), spell.getDuration(), I18n.Text("The duration of the spell once its cast"));
            fields.add(wrapper2);

            ptsPanel = createPointsFields();
            fields.add(ptsPanel);
            refParent = ptsPanel;
        }
        mNotesField = createField(wrapper3, wrapper3, I18n.Text("Notes"), spell.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this spell"), 0);
        mCategoriesField = createField(wrapper3, wrapper3, I18n.Text("Categories"), spell.getCategoriesAsString(), I18n.Text("The category or categories the spell belongs to (separate multiple categories with a comma)"), 0);
        mReferenceField = createField(refParent, noGapWrapper, I18n.Text("Page Reference"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.Text("spell")), 6);
        noGapWrapper.add(new JPanel());
        refParent.add(noGapWrapper);
        fields.add(wrapper3);

        determineLargest(wrapper1, 3, size);
        determineLargest(wrapper2, 4, size);
        if (ptsPanel != null) {
            determineLargest(ptsPanel, 100, size);
        }
        determineLargest(wrapper3, 2, size);
        applySize(wrapper1, 3, size);
        applySize(wrapper2, 4, size);
        if (ptsPanel != null) {
            applySize(ptsPanel, 100, size);
        }
        applySize(wrapper3, 2, size);

        icon.setVerticalAlignment(SwingConstants.TOP);
        icon.setAlignmentY(-1.0f);
        content.add(icon);
        content.add(fields);
        add(content);

        if (notContainer) {
            mTabPanel = new JTabbedPane();
            mPrereqs = new PrereqsPanel(mRow, mRow.getPrereqs());
            mMeleeWeapons = MeleeWeaponEditor.createEditor(mRow);
            mRangedWeapons = RangedWeaponEditor.createEditor(mRow);
            Component panel = embedEditor(mPrereqs);
            mTabPanel.addTab(panel.getName(), panel);
            mTabPanel.addTab(mMeleeWeapons.getName(), mMeleeWeapons);
            mTabPanel.addTab(mRangedWeapons.getName(), mRangedWeapons);
            if (!mIsEditable) {
                UIUtilities.disableControls(mMeleeWeapons);
                UIUtilities.disableControls(mRangedWeapons);
            }
            UIUtilities.selectTab(mTabPanel, getLastTabName());
            add(mTabPanel);
        }
    }

    protected Container createPointsFields() {
        boolean forCharacter = mRow.getCharacter() != null;
        boolean forTemplate  = mRow.getTemplate() != null;
        int     columns      = forTemplate ? 8 : 6;
        JPanel  panel        = new JPanel(new ColumnLayout(forCharacter ? 10 : columns));

        JLabel label = new JLabel(I18n.Text("Difficulty"), SwingConstants.RIGHT);
        label.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The difficulty of the spell")));
        panel.add(label);

        mAttributePopup = createComboBox(panel, SkillAttribute.values(), mRow.getAttribute(), I18n.Text("The attribute this spell is based on"));
        panel.add(new JLabel(" /"));
        mDifficultyCombo = createComboBox(panel, SkillDifficulty.values(), mRow.getDifficulty(), I18n.Text("The difficulty of the spell"));

        if (forCharacter || forTemplate) {
            mPointsField = createNumberField(panel, panel, I18n.Text("Points"), I18n.Text("The number of points spent on this spell"), mRow.getRawPoints(), 4);
            if (forCharacter) {
                mLevelField = createField(panel, panel, I18n.Text("Level"), getDisplayLevel(mRow.getAttribute(), mRow.getLevel(), mRow.getRelativeLevel()), I18n.Text("The spell level and relative spell level to roll against.\n") + mRow.getLevelToolTip(), 7);
                mLevelField.setEnabled(false);
            }
        }
        return panel;
    }

    protected SkillAttribute getAttribute() {
        return (SkillAttribute) mAttributePopup.getSelectedItem();
    }

    public static String getDisplayLevel(SkillAttribute attribute, int level, int relativeLevel) {
        if (level < 0) {
            return "-";
        }
        return Numbers.format(level) + "/" + attribute + Numbers.formatWithForcedSign(relativeLevel);
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
            modified |= mRow.setCollege(mCollegeField.getText());
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
        SkillAttribute attribute = getAttribute();
        SkillLevel     level     = Spell.calculateLevel(mRow.getCharacter(), getAdjustedPoints(), attribute, getDifficulty(), mCollegeField.getText(), mPowerSourceField.getText(), mNameField.getText(), ListRow.createCategoriesList(mCategoriesField.getText()));
        levelField.setText(getDisplayLevel(attribute, level.mLevel, level.mRelativeLevel));
        levelField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The spell level and relative spell level to roll against.\n") + level.getToolTip()));
    }
}
