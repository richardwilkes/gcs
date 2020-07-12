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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.prereq.PrereqsPanel;
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
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
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

/** The detailed editor for {@link Skill}s. */
public class SkillEditor extends RowEditor<Skill> implements ActionListener, DocumentListener, FocusListener {
    private JTextField         mNameField;
    private JTextField         mSpecializationField;
    private JTextField         mNotesField;
    private JTextField         mCategoriesField;
    private JTextField         mReferenceField;
    private JCheckBox          mHasTechLevel;
    private JTextField         mTechLevel;
    private String             mSavedTechLevel;
    private JComboBox<Object>  mAttributePopup;
    private JComboBox<Object>  mDifficultyPopup;
    private JTextField         mPointsField;
    private JTextField         mLevelField;
    private JComboBox<Object>  mEncPenaltyPopup;
    private JTabbedPane        mTabPanel;
    private PrereqsPanel       mPrereqs;
    private FeaturesPanel      mFeatures;
    private Defaults           mDefaults;
    private MeleeWeaponEditor  mMeleeWeapons;
    private RangedWeaponEditor mRangedWeapons;

    /**
     * Creates a new {@link Skill} editor.
     *
     * @param skill The {@link Skill} to edit.
     */
    public SkillEditor(Skill skill) {
        super(skill);

        JPanel    content      = new JPanel(new ColumnLayout(2));
        JPanel    fields       = new JPanel(new ColumnLayout(2));
        JLabel    icon         = new JLabel(skill.getIcon(true));
        boolean   notContainer = !skill.canHaveChildren();
        Container wrapper;

        mNameField = createCorrectableField(fields, I18n.Text("Name"), skill.getName(), I18n.Text("The base name of the skill, without any notes or specialty information"));
        if (notContainer) {
            wrapper = new JPanel(new ColumnLayout(2));
            mSpecializationField = createField(fields, wrapper, I18n.Text("Specialization"), skill.getSpecialization(), I18n.Text("The specialization, if any, taken for this skill"), 0);
            createTechLevelFields(wrapper);
            fields.add(wrapper);
            mEncPenaltyPopup = createEncumbrancePenaltyMultiplierPopup(fields);
        }
        mNotesField = createField(fields, fields, I18n.Text("Notes"), skill.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this skill"), 0);
        mCategoriesField = createField(fields, fields, I18n.Text("Categories"), skill.getCategoriesAsString(), I18n.Text("The category or categories the skill belongs to (separate multiple categories with a comma)"), 0);
        wrapper = notContainer ? createDifficultyPopups(fields) : fields;
        mReferenceField = createField(wrapper, wrapper, I18n.Text("Page Reference"), mRow.getReference(), I18n.Text("A reference to the book and page this skill appears on (e.g. B22 would refer to \"Basic Set\", page 22)"), 6);
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
            mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
            mDefaults = new Defaults(mRow.getDefaults());
            mDefaults.addActionListener(this);
            Component panel = embedEditor(mDefaults);
            mTabPanel.addTab(panel.getName(), panel);
            panel = embedEditor(mPrereqs);
            mTabPanel.addTab(panel.getName(), panel);
            panel = embedEditor(mFeatures);
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

    private JScrollPane embedEditor(Container editor) {
        JScrollPane scrollPanel = new JScrollPane(editor);
        scrollPanel.setMinimumSize(new Dimension(500, 120));
        scrollPanel.setName(editor.toString());
        if (!mIsEditable) {
            UIUtilities.disableControls(editor);
        }
        return scrollPanel;
    }

    private JTextField createCorrectableField(Container parent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        field.getDocument().addDocumentListener(this);

        LinkedLabel label = new LinkedLabel(title);
        label.setLink(field);

        parent.add(label);
        parent.add(field);
        return field;
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
        field.addActionListener(this);
        field.addFocusListener(this);
        return field;
    }

    private static String editorLevelTooltip() {
        return I18n.Text("The skill level and relative skill level to roll against.\n");
    }

    @SuppressWarnings("unused")
    private void createPointsFields(Container parent, boolean forCharacter) {
        mPointsField = createField(parent, parent, I18n.Text("Points"), Integer.toString(mRow.getRawPoints()), I18n.Text("The number of points spent on this skill"), 4);
        new NumberFilter(mPointsField, false, false, false, 4);
        if (forCharacter) {
            mLevelField = createField(parent, parent, I18n.Text("Level"), Skill.getSkillDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel(), mRow.getAttribute(), mRow.canHaveChildren()), editorLevelTooltip() + mRow.getLevelToolTip(), 8);
            mLevelField.setEnabled(false);
        }
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

            String tlTooltip = I18n.Text("Whether this skill requires tech level specialization, and, if so, at what tech level it was learned");
            mHasTechLevel = new JCheckBox(I18n.Text("Tech Level"), hasTL);
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
            parent.add(wrapper);

            if (!hasTL) {
                mSavedTechLevel = character.getProfile().getTechLevel();
            }
        } else {
            mTechLevel = new JTextField(mSavedTechLevel);
            mHasTechLevel = new JCheckBox(I18n.Text("Tech Level Required"), hasTL);
            mHasTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Whether this skill requires tech level specialization")));
            mHasTechLevel.setEnabled(enabled);
            mHasTechLevel.addActionListener(this);
            parent.add(mHasTechLevel);
        }
    }

    private JComboBox<Object> createEncumbrancePenaltyMultiplierPopup(Container parent) {
        Object[] items = new Object[10];
        items[0] = I18n.Text("No penalty due to encumbrance");
        items[1] = I18n.Text("Penalty equal to the current encumbrance level");
        for (int i = 2; i < 10; i++) {
            items[i] = MessageFormat.format(I18n.Text("Penalty equal to {0} times the current encumbrance level"), Integer.valueOf(i));
        }
        LinkedLabel label = new LinkedLabel(I18n.Text("Encumbrance"));
        parent.add(label);
        JComboBox<Object> popup = createComboBox(parent, items, items[mRow.getEncumbrancePenaltyMultiplier()], I18n.Text("The encumbrance penalty multiplier"));
        label.setLink(popup);
        return popup;
    }

    private Container createDifficultyPopups(Container parent) {
        GURPSCharacter character              = mRow.getCharacter();
        boolean        forCharacterOrTemplate = character != null || mRow.getTemplate() != null;
        JLabel         label                  = new JLabel(I18n.Text("Difficulty"), SwingConstants.RIGHT);
        int            columns                = character != null ? 10 : 8;
        JPanel         wrapper                = new JPanel(new ColumnLayout(forCharacterOrTemplate ? columns : 6));

        label.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The difficulty of learning this skill")));

        mAttributePopup = createComboBox(wrapper, SkillAttribute.values(), mRow.getAttribute(), I18n.Text("The attribute this skill is based on"));
        wrapper.add(new JLabel(" /"));
        mDifficultyPopup = createComboBox(wrapper, SkillDifficulty.values(), mRow.getDifficulty(), I18n.Text("The relative difficulty of learning this skill"));

        if (forCharacterOrTemplate) {
            createPointsFields(wrapper, character != null);
        }
        wrapper.add(new JPanel());

        parent.add(label);
        parent.add(wrapper);
        return wrapper;
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

    private void recalculateLevel() {
        if (mLevelField != null) {
            SkillAttribute attribute = getSkillAttribute();
            SkillLevel     level     = mRow.calculateLevel(mRow.getCharacter(), mNameField.getText(), mSpecializationField.getText(), ListRow.createCategoriesList(mCategoriesField.getText()), mDefaults.getDefaults(), attribute, getSkillDifficulty(), getAdjustedSkillPoints(), new HashSet<>(), getEncumbrancePenaltyMultiplier());
            mLevelField.setText(Skill.getSkillDisplayLevel(level.mLevel, level.mRelativeLevel, attribute, false));
            mLevelField.setToolTipText(Text.wrapPlainTextForToolTip(editorLevelTooltip() + level.getToolTip()));
        }
    }

    private SkillAttribute getSkillAttribute() {
        return (SkillAttribute) mAttributePopup.getSelectedItem();
    }

    private SkillDifficulty getSkillDifficulty() {
        return (SkillDifficulty) mDifficultyPopup.getSelectedItem();
    }

    private int getSkillPoints() {
        return Numbers.extractInteger(mPointsField.getText(), 0, true);
    }

    private int getAdjustedSkillPoints() {
        int            points    = getSkillPoints();
        GURPSCharacter character = mRow.getCharacter();
        if (character != null) {
            String name = mNameField.getText();
            points += character.getSkillPointComparedIntegerBonusFor(Skill.ID_POINTS + "*", name, mSpecializationField.getText(), ListRow.createCategoriesList(mCategoriesField.getText()));
            points += character.getIntegerBonusFor(Skill.ID_POINTS + "/" + name.toLowerCase());
            if (points < 0) {
                points = 0;
            }
        }
        return points;
    }

    private int getEncumbrancePenaltyMultiplier() {
        return mEncPenaltyPopup.getSelectedIndex();
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified = mRow.setName(mNameField.getText());
        modified |= mRow.setReference(mReferenceField.getText());
        modified |= mRow.setNotes(mNotesField.getText());
        modified |= mRow.setCategories(mCategoriesField.getText());
        if (mSpecializationField != null) {
            modified |= mRow.setSpecialization(mSpecializationField.getText());
        }
        if (mHasTechLevel != null) {
            modified |= mRow.setTechLevel(mHasTechLevel.isSelected() ? mTechLevel.getText() : null);
        }
        if (mAttributePopup != null) {
            modified |= mRow.setDifficulty(getSkillAttribute(), getSkillDifficulty());
        }
        if (mEncPenaltyPopup != null) {
            modified |= mRow.setEncumbrancePenaltyMultiplier(getEncumbrancePenaltyMultiplier());
        }
        if (mPointsField != null) {
            modified |= mRow.setRawPoints(getSkillPoints());
        }
        if (mDefaults != null) {
            modified |= mRow.setDefaults(mDefaults.getDefaults());
        }
        if (mPrereqs != null) {
            modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
        }
        if (mFeatures != null) {
            modified |= mRow.setFeatures(mFeatures.getFeatures());
        }
        if (mMeleeWeapons != null) {
            List<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
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
        } else if (src == mAttributePopup || src == mDifficultyPopup || src == mPointsField || src == mDefaults || src == mEncPenaltyPopup) {
            recalculateLevel();
        }
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        nameChanged();
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        nameChanged();
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        nameChanged();
    }

    private void nameChanged() {
        LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.Text("The name field may not be empty") : null);
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
