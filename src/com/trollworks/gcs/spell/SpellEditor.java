/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.skill.SkillAttribute;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
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
import com.trollworks.toolkit.utility.Platform;
import com.trollworks.toolkit.utility.text.NumberFilter;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.Text;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;

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

/** The detailed editor for {@link Spell}s. */
public class SpellEditor extends RowEditor<Spell> implements ActionListener, DocumentListener {
    private JTextField         mNameField;
    private JTextField         mCollegeField;
    private JTextField         mPowerSourceField;
    private JTextField         mClassField;
    private JTextField         mCastingCostField;
    private JTextField         mMaintenanceField;
    private JTextField         mCastingTimeField;
    private JTextField         mDurationField;
    private JComboBox<Object>  mAttributePopup;
    private JComboBox<Object>  mDifficultyCombo;
    private JTextField         mNotesField;
    private JTextField         mCategoriesField;
    private JTextField         mPointsField;
    private JTextField         mLevelField;
    private JTextField         mReferenceField;
    private JTabbedPane        mTabPanel;
    private PrereqsPanel       mPrereqs;
    private JCheckBox          mHasTechLevel;
    private JTextField         mTechLevel;
    private String             mSavedTechLevel;
    private MeleeWeaponEditor  mMeleeWeapons;
    private RangedWeaponEditor mRangedWeapons;

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
            mCollegeField     = createField(wrapper2, wrapper2, I18n.Text("College"), spell.getCollege(), I18n.Text("The college the spell belongs to"), 0);
            mPowerSourceField = createField(wrapper2, wrapper2, I18n.Text("Power Source"), spell.getPowerSource(), I18n.Text("The source of power for the spell"), 0);
            mClassField       = createCorrectableField(wrapper2, wrapper2, I18n.Text("Class"), spell.getSpellClass(), I18n.Text("The class of spell (Area, Missile, etc.)"));
            mCastingCostField = createCorrectableField(wrapper2, wrapper2, I18n.Text("Casting Cost"), spell.getCastingCost(), I18n.Text("The casting cost of the spell"));
            mMaintenanceField = createField(wrapper2, wrapper2, I18n.Text("Maintenance Cost"), spell.getMaintenance(), I18n.Text("The cost to maintain a spell after its initial duration"), 0);
            mCastingTimeField = createCorrectableField(wrapper2, wrapper2, I18n.Text("Casting Time"), spell.getCastingTime(), I18n.Text("The casting time of the spell"));
            mDurationField    = createCorrectableField(wrapper2, wrapper2, I18n.Text("Duration"), spell.getDuration(), I18n.Text("The duration of the spell once its cast"));
            fields.add(wrapper2);

            ptsPanel = createPointsFields();
            fields.add(ptsPanel);
            refParent = ptsPanel;
        }
        mNotesField      = createField(wrapper3, wrapper3, I18n.Text("Notes"), spell.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this spell"), 0);
        mCategoriesField = createField(wrapper3, wrapper3, I18n.Text("Categories"), spell.getCategoriesAsString(), I18n.Text("The category or categories the spell belongs to (separate multiple categories with a comma)"), 0);
        mReferenceField  = createField(refParent, noGapWrapper, I18n.Text("Page Reference"), mRow.getReference(), I18n.Text("A reference to the book and page this spell appears on (e.g. B22 would refer to \"Basic Set\", page 22)"), 6);
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
        icon.setAlignmentY(-1f);
        content.add(icon);
        content.add(fields);
        add(content);

        if (notContainer) {
            mTabPanel      = new JTabbedPane();
            mPrereqs       = new PrereqsPanel(mRow, mRow.getPrereqs());
            mMeleeWeapons  = MeleeWeaponEditor.createEditor(mRow);
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

    private void createTechLevelFields(Container parent) {
        OutlineModel   owner     = mRow.getOwner();
        GURPSCharacter character = mRow.getCharacter();
        boolean        enabled   = !owner.isLocked();
        boolean        hasTL;

        mSavedTechLevel = mRow.getTechLevel();
        hasTL           = mSavedTechLevel != null;
        if (!hasTL) {
            mSavedTechLevel = "";
        }

        if (character != null) {
            JPanel wrapper = new JPanel(new ColumnLayout(2));

            mHasTechLevel = new JCheckBox(I18n.Text("Tech Level"), hasTL);
            Dimension size = mHasTechLevel.getPreferredSize();
            if (Platform.isWindows()) {
                size.width++; // Text measurement is off on Windows for some reason
            }
            UIUtilities.setOnlySize(mHasTechLevel, size);
            String tlTooltip = I18n.Text("Whether this spell requires tech level specialization, and, if so, at what tech level it was learned");
            mHasTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(tlTooltip));
            mHasTechLevel.setEnabled(enabled);
            mHasTechLevel.addActionListener(this);
            wrapper.add(mHasTechLevel);

            mTechLevel = new JTextField("9999");
            UIUtilities.setOnlySize(mTechLevel, mTechLevel.getPreferredSize());
            mTechLevel.setText(mSavedTechLevel);
            mTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(tlTooltip));
            mTechLevel.setEnabled(enabled && hasTL);
            wrapper.add(mTechLevel);
            UIUtilities.setOnlySize(wrapper, wrapper.getPreferredSize());
            parent.add(wrapper);

            if (!hasTL) {
                mSavedTechLevel = character.getDescription().getTechLevel();
            }
        } else {
            mTechLevel    = new JTextField(mSavedTechLevel);
            mHasTechLevel = new JCheckBox(I18n.Text("Tech Level Required"), hasTL);
            mHasTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Whether this spell requires tech level specialization")));
            mHasTechLevel.setEnabled(enabled);
            mHasTechLevel.addActionListener(this);
            parent.add(mHasTechLevel);
        }
    }

    @SuppressWarnings("unused")
    private Container createPointsFields() {
        boolean forCharacter = mRow.getCharacter() != null;
        boolean forTemplate  = mRow.getTemplate() != null;
        JPanel  panel        = new JPanel(new ColumnLayout(forCharacter ? 10 : forTemplate ? 8 : 6));

        JLabel  label        = new JLabel(I18n.Text("Difficulty"), SwingConstants.RIGHT);
        label.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The difficulty of the spell")));
        panel.add(label);

        mAttributePopup = createComboBox(panel, SkillAttribute.values(), mRow.getAttribute(), I18n.Text("The attribute this spell is based on"));
        panel.add(new JLabel(" /"));
        mDifficultyCombo = createComboBox(panel, new Object[] { SkillDifficulty.H, SkillDifficulty.VH }, mRow.isVeryHard() ? SkillDifficulty.VH : SkillDifficulty.H, I18n.Text("The difficulty of the spell"));

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

    private static final String editorLevelTooltip() {
        return I18n.Text("The spell level and relative spell level to roll against.\n");
    }

    private JComboBox<Object> createComboBox(Container parent, Object[] items, Object selection, String tooltip) {
        JComboBox<Object> combo = new JComboBox<>(items);
        combo.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        combo.setSelectedItem(selection);
        combo.addActionListener(this);
        combo.setMaximumRowCount(items.length);
        UIUtilities.setOnlySize(combo, combo.getPreferredSize());
        combo.setEnabled(mIsEditable);
        parent.add(combo);
        return combo;
    }

    private static String getDisplayLevel(SkillAttribute attribute, int level, int relativeLevel) {
        if (level < 0) {
            return "-";
        }
        return Numbers.format(level) + "/" + attribute + Numbers.formatWithForcedSign(relativeLevel);
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

    private JTextField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
        JTextField field = new JTextField(maxChars > 0 ? Text.makeFiller(maxChars, 'M') : text);

        if (maxChars > 0) {
            UIUtilities.setOnlySize(field, field.getPreferredSize());
            field.setText(text);
        }
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        labelParent.add(new LinkedLabel(title, field));
        fieldParent.add(field);
        return field;
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
            modified |= mRow.setDuration(mDurationField.getText());
            modified |= mRow.setDifficulty(getAttribute(), isVeryHard());
            if (mRow.getCharacter() != null || mRow.getTemplate() != null) {
                modified |= mRow.setPoints(getSpellPoints());
            }
        }
        modified |= mRow.setNotes(mNotesField.getText());
        modified |= mRow.setCategories(mCategoriesField.getText());
        if (mPrereqs != null) {
            modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
        }
        if (mMeleeWeapons != null) {
            ArrayList<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());

            list.addAll(mRangedWeapons.getWeapons());
            modified |= mRow.setWeapons(list);
        }
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
        } else if (src == mPointsField || src == mAttributePopup || src == mDifficultyCombo) {
            recalculateLevel();
        }
    }

    private void recalculateLevel() {
        if (mLevelField != null) {
            SkillAttribute attribute = getAttribute();
            SkillLevel     level     = Spell.calculateLevel(mRow.getCharacter(), getSpellPoints(), attribute, isVeryHard(), mCollegeField.getText(), mPowerSourceField.getText(), mNameField.getText(), ListRow.createCategoriesList(mCategoriesField.getText()));
            mLevelField.setText(getDisplayLevel(attribute, level.mLevel, level.mRelativeLevel));
            mLevelField.setToolTipText(Text.wrapPlainTextForToolTip(editorLevelTooltip() + level.getToolTip()));
        }
    }

    private int getSpellPoints() {
        return Numbers.extractInteger(mPointsField.getText(), 0, true);
    }

    private SkillAttribute getAttribute() {
        return (SkillAttribute) mAttributePopup.getSelectedItem();
    }

    private boolean isVeryHard() {
        return mDifficultyCombo.getSelectedItem() == SkillDifficulty.VH;
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document doc = event.getDocument();
        if (doc == mNameField.getDocument()) {
            LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().length() != 0 ? null : I18n.Text("The name field may not be empty"));
        } else if (doc == mClassField.getDocument()) {
            LinkedLabel.setErrorMessage(mClassField, mClassField.getText().trim().length() != 0 ? null : I18n.Text("The class field may not be empty"));
        } else if (doc == mClassField.getDocument()) {
            LinkedLabel.setErrorMessage(mCastingCostField, mCastingCostField.getText().trim().length() != 0 ? null : I18n.Text("The casting cost field may not be empty"));
        } else if (doc == mClassField.getDocument()) {
            LinkedLabel.setErrorMessage(mCastingTimeField, mCastingTimeField.getText().trim().length() != 0 ? null : I18n.Text("The casting time field may not be empty"));
        } else if (doc == mClassField.getDocument()) {
            LinkedLabel.setErrorMessage(mDurationField, mDurationField.getText().trim().length() != 0 ? null : I18n.Text("The duration field may not be empty"));
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
}
