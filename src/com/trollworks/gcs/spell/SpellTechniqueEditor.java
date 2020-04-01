/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.SkillAttribute;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.widget.LinkedLabel;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.text.NumberFilter;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.Text;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTabbedPane;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

public class SpellTechniqueEditor extends RowEditor<SpellTechnique> implements ActionListener, DocumentListener {
    private JTextField         mNameField;
    private JCheckBox          mHasTechLevel;
    private JTextField         mTechLevel;
    private JTextField         mCollegeField;
    private JTextField         mPowerSourceField;
    private JTextField         mClassField;
    private JTextField         mCastingCostField;
    private JTextField         mMaintenanceField;
    private JTextField         mCastingTimeField;
    private JTextField         mDurationField;
    private JComboBox<Object>  mDifficultyCombo;
    private JTextField         mPointsField;
    private JTextField         mDefaultModifierField;
    private JTextField         mLevelField;
    private JTextField         mReferenceField;
    private JTextField         mNotesField;
    private JTextField         mCategoriesField;
    private JTabbedPane        mTabPanel;
    private PrereqsPanel       mPrereqs;
    private MeleeWeaponEditor  mMeleeWeapons;
    private RangedWeaponEditor mRangedWeapons;

    private String             mSavedTechLevel;


    /**
     * Creates a new {@link Spell} {@link RowEditor}.
     *
     * @param spellTechnique The row being edited.
     */
    protected SpellTechniqueEditor(SpellTechnique spellTechnique) {
        super(spellTechnique);

        Container content      = new JPanel(new ColumnLayout(2));
        Container fields       = new JPanel(new ColumnLayout());
        Container wrapper1     = new JPanel(new ColumnLayout(3));
        Container wrapper2     = new JPanel(new ColumnLayout(4));
        Container wrapper3     = new JPanel(new ColumnLayout(2));
        Container noGapWrapper = new JPanel(new ColumnLayout(2, 0, 0));
        JLabel    icon         = new JLabel(spellTechnique.getIcon(true));
        Dimension size         = new Dimension();
        Container ptsPanel;

        mNameField = createCorrectableField(wrapper1, wrapper1, I18n.Text("Name"), spellTechnique.getName(), I18n.Text("The name of the spell, without any notes"));
        fields.add(wrapper1);

        createTechLevelFields(wrapper1);
        mCollegeField         = createField(wrapper2, wrapper2, I18n.Text("College"), spellTechnique.getCollege(), I18n.Text("The college the spell belongs to"), 0);
        mPowerSourceField     = createField(wrapper2, wrapper2, I18n.Text("Power Source"), spellTechnique.getPowerSource(), I18n.Text("The source of power for the spell"), 0);
        mClassField           = createCorrectableField(wrapper2, wrapper2, I18n.Text("Class"), spellTechnique.getSpellClass(), I18n.Text("The class of spell (Area, Missile, etc.)"));
        mCastingCostField     = createCorrectableField(wrapper2, wrapper2, I18n.Text("Casting Cost"), spellTechnique.getCastingCost(), I18n.Text("The casting cost of the spell"));
        mMaintenanceField     = createField(wrapper2, wrapper2, I18n.Text("Maintenance Cost"), spellTechnique.getMaintenance(), I18n.Text("The cost to maintain a spell after its initial duration"), 0);
        mCastingTimeField     = createCorrectableField(wrapper2, wrapper2, I18n.Text("Casting Time"), spellTechnique.getCastingTime(), I18n.Text("The casting time of the spell"));
        mDurationField        = createCorrectableField(wrapper2, wrapper2, I18n.Text("Duration"), spellTechnique.getDuration(), I18n.Text("The duration of the spell once its cast"));
        mDefaultModifierField = createNumberField(wrapper2, wrapper2, I18n.Text("Prerequisite Count"), I18n.Text("The number of prerequisite SPELLS needed to cast this spell"), mRow.getSpellPrerequisiteCount(), 2);
        fields.add(wrapper2);

        ptsPanel = createPointsFields();
        fields.add(ptsPanel);

        mNotesField      = createField(wrapper3, wrapper3, I18n.Text("Notes"), spellTechnique.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this spell"), 0);
        mCategoriesField = createField(wrapper3, wrapper3, I18n.Text("Categories"), spellTechnique.getCategoriesAsString(), I18n.Text("The category or categories the spell belongs to (separate multiple categories with a comma)"), 0);
        mReferenceField  = createField(ptsPanel, noGapWrapper, I18n.Text("Page Reference"), mRow.getReference(), I18n.Text("A reference to the book and page this spell appears on (e.g. B22 would refer to \"Basic Set\", page 22)"), 6);
        noGapWrapper.add(new JPanel());
        ptsPanel.add(noGapWrapper);
        fields.add(wrapper3);

        determineLargest(wrapper1, 3, size);
        determineLargest(wrapper2, 4, size);
        determineLargest(ptsPanel, 100, size);
        determineLargest(wrapper3, 2, size);
        applySize(wrapper1, 3, size);
        applySize(wrapper2, 4, size);
        applySize(ptsPanel, 100, size);
        applySize(wrapper3, 2, size);

        icon.setVerticalAlignment(SwingConstants.TOP);
        icon.setAlignmentY(-1.0f);
        content.add(icon);
        content.add(fields);
        add(content);

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

    @Override
    protected boolean applyChangesSelf() {
        boolean modified = mRow.setName(mNameField.getText());

        modified |= mRow.setReference(mReferenceField.getText());
        if (mHasTechLevel != null) {
            modified |= mRow.setTechLevel(mHasTechLevel.isSelected() ? mTechLevel.getText() : null);
        }
        modified |= mRow.setCollege(mCollegeField.getText());
        modified |= mRow.setPowerSource(mPowerSourceField.getText());
        modified |= mRow.setSpellClass(mClassField.getText());
        modified |= mRow.setCastingCost(mCastingCostField.getText());
        modified |= mRow.setMaintenance(mMaintenanceField.getText());
        modified |= mRow.setCastingTime(mCastingTimeField.getText());
        modified |= mRow.setDuration(mDurationField.getText());
        modified |= mRow.setSpellPrerequisiteCount(Numbers.extractInteger(mDefaultModifierField.getText(), 0, true));
        if (mRow.getCharacter() != null || mRow.getTemplate() != null) {
            modified |= mRow.setPoints(Numbers.extractInteger(mPointsField.getText(), 0, true));
        }
        modified |= mRow.setNotes(mNotesField.getText());
        modified |= mRow.setCategories(mCategoriesField.getText());
        modified |= mRow.setPrereqs(mPrereqs.getPrereqList());

        List<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
        list.addAll(mRangedWeapons.getWeapons());
        modified |= mRow.setWeapons(list);

        return modified;
    }

    @Override
    public void finished() {
        if (mTabPanel != null) {
            updateLastTabName(mTabPanel.getTitleAt(mTabPanel.getSelectedIndex()));
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object src = event.getSource();
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
        } else if (src == mPointsField || src == mDifficultyCombo || src == mCollegeField || src == mDefaultModifierField) {
            recalculateLevel();
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
    public void changedUpdate(DocumentEvent event) {
        Document doc = event.getDocument();
        if (doc == mNameField.getDocument()) {
            LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.Text("The name field may not be empty") : null);
        } else if (doc == mClassField.getDocument()) {
            LinkedLabel.setErrorMessage(mClassField, mClassField.getText().trim().isEmpty() ? I18n.Text("The class field may not be empty") : null);
        } else if (doc == mClassField.getDocument()) {
            LinkedLabel.setErrorMessage(mCastingCostField, mCastingCostField.getText().trim().isEmpty() ? I18n.Text("The casting cost field may not be empty") : null);
        } else if (doc == mClassField.getDocument()) {
            LinkedLabel.setErrorMessage(mCastingTimeField, mCastingTimeField.getText().trim().isEmpty() ? I18n.Text("The casting time field may not be empty") : null);
        } else if (doc == mClassField.getDocument()) {
            LinkedLabel.setErrorMessage(mDurationField, mDurationField.getText().trim().isEmpty() ? I18n.Text("The duration field may not be empty") : null);
        }

    }

    private void recalculateLevel() {
        if (mLevelField != null) {
            String skillName      = I18n.Text("College");
            String skillSpec      = mCollegeField.getText();
            int    prereqModifier = Numbers.extractInteger(mDefaultModifierField.getText(), 0, true);
            int    points         = Numbers.extractInteger(mPointsField.getText(), 0, true);

            String       specialization = String.format("%s (%s)", skillName, skillSpec);
            Set<String>  categories     = ListRow.createCategoriesList(mCategoriesField.getText());
            SkillDefault skillDefault   = new SkillDefault(SkillDefaultType.Skill, skillName, skillSpec, prereqModifier);
            SkillLevel level = Technique.calculateTechniqueLevel(mRow.getCharacter(), mNameField.getText(), specialization, categories, skillDefault, SkillDifficulty.H, points, true, 0);

            mLevelField.setText(Technique.getTechniqueDisplayLevel(level.mLevel, level.mRelativeLevel, prereqModifier));
            mLevelField.setToolTipText(Text.wrapPlainTextForToolTip(editorLevelTooltip() + level.getToolTip()));
        }
    }

    private static String editorLevelTooltip() {
        return I18n.Text("The spell level and relative spell level to roll against.\n");
    }

    private static String getDisplayLevel(SkillAttribute attribute, int level, int relativeLevel) {
        if (level < 0) {
            return "-";
        }
        return Numbers.format(level) + "/" + attribute + Numbers.formatWithForcedSign(relativeLevel);
    }

    private static void determineLargest(Container panel, int every, Dimension size) {
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

    private static void applySize(Container panel, int every, Dimension size) {
        int count = panel.getComponentCount();

        for (int i = 0; i < count; i += every) {
            UIUtilities.setOnlySize(panel.getComponent(i), size);
        }
    }

    private JScrollPane embedEditor(Component editor) {
        JScrollPane scrollPanel = new JScrollPane(editor);

        scrollPanel.setMinimumSize(new Dimension(500, 120));
        scrollPanel.setName(editor.toString());
        if (!mIsEditable) {
            UIUtilities.disableControls(editor);
        }
        return scrollPanel;
    }

    private JTextField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
        JTextField field = new JTextField(maxChars > 0 ? Text.makeFiller(maxChars, 'M') : text);

        if (maxChars > 0) {
            UIUtilities.setToPreferredSizeOnly(field);
            field.setText(text);
        }
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
        return field;
    }

    private JTextField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        field.getDocument().addDocumentListener(this);

        LinkedLabel label = new LinkedLabel(title);
        label.setLink(field);

        labelParent.add(label);
        fieldParent.add(field);
        return field;
    }

    private JTextField createNumberField(Container labelParent, Container fieldParent, String title, String tooltip, int value, int maxDigits) {
        JTextField field = createField(labelParent, fieldParent, title, Numbers.format(value), tooltip, maxDigits + 1);
        new NumberFilter(field, false, false, false, maxDigits);
        return field;
    }

    private JComboBox<Object> createComboBox(Container parent, Object[] items, Object selection, String tooltip) {
        JComboBox<Object> combo = new JComboBox<>(items);
        combo.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        combo.setSelectedItem(selection);
        combo.addActionListener(this);
        combo.setMaximumRowCount(items.length);
        UIUtilities.setToPreferredSizeOnly(combo);
        combo.setEnabled(mIsEditable);
        parent.add(combo);
        return combo;
    }

    private void createTechLevelFields(Container parent) {
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
                mSavedTechLevel = character.getDescription().getTechLevel();
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

    private Container createPointsFields() {
        boolean forCharacter = mRow.getCharacter() != null;
        boolean forTemplate  = mRow.getTemplate() != null;
        int     columns      = forTemplate ? 8 : 6;
        JPanel  panel        = new JPanel(new ColumnLayout(forCharacter ? 10 : columns));

        JLabel label = new JLabel(I18n.Text("Difficulty"), SwingConstants.RIGHT);
        label.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The difficulty of the spell")));
        panel.add(label);

        mDifficultyCombo = createComboBox(panel, new Object[]{SkillDifficulty.A, SkillDifficulty.H}, mRow.getDifficulty(), I18n.Text("The difficulty of the spell"));

        if (forCharacter || forTemplate) {
            mPointsField = createField(panel, panel, I18n.Text("Points"), Integer.toString(mRow.getPoints()), I18n.Text("The number of points spent on this spell"), 4);
            new NumberFilter(mPointsField, false, false, false, 4);
            mPointsField.addActionListener(this);

            if (forCharacter) {
                mLevelField = createField(panel, panel, I18n.Text("Level"), getDisplayLevel(mRow.getAttribute(), mRow.getLevel(), mRow.getRelativeLevel()), editorLevelTooltip() + mRow.getLevelToolTip(), 7);
                mLevelField.setEnabled(false);
            }
        }
        return panel;
    }
}
