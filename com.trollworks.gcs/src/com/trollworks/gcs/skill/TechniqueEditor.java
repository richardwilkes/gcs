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
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
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
import java.util.ArrayList;
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
import javax.swing.text.Document;

/** The detailed editor for {@link Technique}s. */
public class TechniqueEditor extends RowEditor<Technique> implements ActionListener, DocumentListener {
    private JTextField         mNameField;
    private JTextField         mNotesField;
    private JTextField         mCategoriesField;
    private JTextField         mReferenceField;
    private JComboBox<Object>  mDifficultyCombo;
    private JTextField         mPointsField;
    private JTextField         mLevelField;
    private JPanel             mDefaultPanel;
    private LinkedLabel        mDefaultPanelLabel;
    private JComboBox<Object>  mDefaultTypeCombo;
    private JTextField         mDefaultNameField;
    private JTextField         mDefaultSpecializationField;
    private JTextField         mDefaultModifierField;
    private JCheckBox          mLimitCheckbox;
    private JTextField         mLimitField;
    private JTabbedPane        mTabPanel;
    private PrereqsPanel       mPrereqs;
    private FeaturesPanel      mFeatures;
    private SkillDefaultType   mLastDefaultType;
    private MeleeWeaponEditor  mMeleeWeapons;
    private RangedWeaponEditor mRangedWeapons;

    /**
     * Creates a new {@link Technique} editor.
     *
     * @param technique The {@link Technique} to edit.
     */
    public TechniqueEditor(Technique technique) {
        super(technique);

        JPanel    content = new JPanel(new ColumnLayout(2));
        JPanel    fields  = new JPanel(new ColumnLayout(2));
        JLabel    icon    = new JLabel(technique.getIcon(true));
        Container wrapper;

        mNameField = createCorrectableField(fields, fields, I18n.Text("Name"), technique.getName(), I18n.Text("The base name of the technique, without any notes or specialty information"));
        mNotesField = createField(fields, fields, I18n.Text("Notes"), technique.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this technique"), 0);
        mCategoriesField = createField(fields, fields, I18n.Text("Categories"), technique.getCategoriesAsString(), I18n.Text("The category or categories the technique belongs to (separate multiple categories with a comma)"), 0);
        createDefaults(fields);
        createLimits(fields);
        wrapper = createDifficultyPopups(fields);
        mReferenceField = createField(wrapper, wrapper, I18n.Text("Page Reference"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.Text("technique")), 6);
        icon.setVerticalAlignment(SwingConstants.TOP);
        icon.setAlignmentY(-1.0f);
        content.add(icon);
        content.add(fields);
        add(content);

        mTabPanel = new JTabbedPane();
        mPrereqs = new PrereqsPanel(mRow, mRow.getPrereqs());
        mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
        mMeleeWeapons = MeleeWeaponEditor.createEditor(mRow);
        mRangedWeapons = RangedWeaponEditor.createEditor(mRow);
        Component panel = embedEditor(mPrereqs);
        mTabPanel.addTab(panel.getName(), panel);
        panel = embedEditor(mFeatures);
        mTabPanel.addTab(panel.getName(), panel);
        mTabPanel.addTab(mMeleeWeapons.getName(), mMeleeWeapons);
        mTabPanel.addTab(mRangedWeapons.getName(), mRangedWeapons);
        UIUtilities.selectTab(mTabPanel, getLastTabName());
        add(mTabPanel);
    }

    private void createDefaults(Container parent) {
        mDefaultPanel = new JPanel(new ColumnLayout(4));
        mDefaultPanelLabel = new LinkedLabel(I18n.Text("Defaults To"));
        mDefaultTypeCombo = createComboBox(mDefaultPanel, SkillDefaultType.values(), mRow.getDefault().getType());
        mDefaultTypeCombo.setEnabled(mIsEditable);

        parent.add(mDefaultPanelLabel);
        parent.add(mDefaultPanel);
        rebuildDefaultPanel();
    }

    private JComboBox<Object> createComboBox(Container parent, Object[] items, Object selection) {
        JComboBox<Object> combo = new JComboBox<>(items);
        combo.setSelectedItem(selection);
        combo.addActionListener(this);
        combo.setMaximumRowCount(items.length);
        UIUtilities.setToPreferredSizeOnly(combo);
        parent.add(combo);
        return combo;
    }

    private SkillDefaultType getDefaultType() {
        return (SkillDefaultType) mDefaultTypeCombo.getSelectedItem();
    }

    private String getSpecialization() {
        StringBuilder builder        = new StringBuilder();
        String        specialization = mDefaultSpecializationField.getText();

        builder.append(mDefaultNameField.getText());
        if (!specialization.isEmpty()) {
            builder.append(" (");
            builder.append(specialization);
            builder.append(')');
        }
        return builder.toString();
    }

    private void rebuildDefaultPanel() {
        SkillDefault def = mRow.getDefault();
        boolean      skillBased;

        mLastDefaultType = getDefaultType();
        skillBased = mLastDefaultType.isSkillBased();
        Commitable.sendCommitToFocusOwner();
        while (mDefaultPanel.getComponentCount() > 1) {
            mDefaultPanel.remove(1);
        }
        if (skillBased) {
            mDefaultNameField = createCorrectableField(null, mDefaultPanel, I18n.Text("Defaults To"), def.getName(), I18n.Text("The name of the skill this technique defaults from"));
            mDefaultSpecializationField = createField(null, mDefaultPanel, null, def.getSpecialization(), I18n.Text("The specialization of the skill, if any, this technique defaults from"), 0);
            mDefaultPanelLabel.setLink(mDefaultNameField);
        }
        mDefaultModifierField = createNumberField(mDefaultPanel, I18n.Text("The amount to adjust the default skill level by"), def.getModifier());
        if (!skillBased) {
            mDefaultPanel.add(new JPanel());
            mDefaultPanel.add(new JPanel());
        }
        mDefaultPanel.revalidate();
    }

    private void createLimits(Container parent) {
        JPanel wrapper = new JPanel(new ColumnLayout(3));

        mLimitCheckbox = new JCheckBox(I18n.Text("Cannot exceed default skill level by more than"), mRow.isLimited());
        mLimitCheckbox.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Whether to limit the maximum level that can be achieved or not")));
        mLimitCheckbox.addActionListener(this);
        mLimitCheckbox.setEnabled(mIsEditable);

        mLimitField = createNumberField(wrapper, I18n.Text("The maximum amount above the default skill level that this technique can be raised"), mRow.getLimitModifier());
        mLimitField.setEnabled(mIsEditable && mLimitCheckbox.isSelected());
        mLimitField.addActionListener(this);

        wrapper.add(mLimitCheckbox);
        wrapper.add(mLimitField);
        wrapper.add(new JPanel());
        parent.add(new JLabel());
        parent.add(wrapper);
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

    private JTextField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.setEnabled(mIsEditable);
        field.getDocument().addDocumentListener(this);

        if (labelParent != null) {
            LinkedLabel label = new LinkedLabel(title);
            label.setLink(field);
            labelParent.add(label);
        }

        fieldParent.add(field);
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
        field.addActionListener(this);
        if (labelParent != null) {
            labelParent.add(new LinkedLabel(title, field));
        }
        fieldParent.add(field);
        return field;
    }

    private JTextField createNumberField(Container fieldParent, String tooltip, int value) {
        JTextField field = createField(null, fieldParent, null, Numbers.formatWithForcedSign(value), tooltip, 3);
        NumberFilter.apply(field, false, true, false, 2);
        return field;
    }

    private static String editorLevelTooltip() {
        return I18n.Text("The skill level and relative skill level to roll against.\n");
    }

    private void createPointsFields(Container parent, boolean forCharacter) {
        mPointsField = createField(parent, parent, I18n.Text("Points"), Integer.toString(mRow.getRawPoints()), I18n.Text("The number of points spent on this technique"), 4);
        NumberFilter.apply(mPointsField, false, false, false, 4);
        mPointsField.addActionListener(this);

        if (forCharacter) {
            mLevelField = createField(parent, parent, I18n.Text("Level"), Technique.getTechniqueDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel(), mRow.getDefault().getModifier()), editorLevelTooltip() + mRow.getLevelToolTip(), 6);
            mLevelField.setEnabled(false);
        }
    }

    private Container createDifficultyPopups(Container parent) {
        GURPSCharacter character              = mRow.getCharacter();
        int            columns                = character != null ? 8 : 6;
        boolean        forCharacterOrTemplate = character != null || mRow.getTemplate() != null;
        JLabel         label                  = new JLabel(I18n.Text("Difficulty"), SwingConstants.RIGHT);
        JPanel         wrapper                = new JPanel(new ColumnLayout(forCharacterOrTemplate ? columns : 4));

        label.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The difficulty of learning this technique")));

        mDifficultyCombo = createComboBox(wrapper, new Object[]{SkillDifficulty.A, SkillDifficulty.H}, mRow.getDifficulty());
        mDifficultyCombo.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The relative difficulty of learning this technique")));
        mDifficultyCombo.setEnabled(mIsEditable);

        if (forCharacterOrTemplate) {
            createPointsFields(wrapper, character != null);
        }
        wrapper.add(new JPanel());

        parent.add(label);
        parent.add(wrapper);
        return wrapper;
    }

    private void recalculateLevel() {
        if (mLevelField != null) {
            SkillLevel level = Technique.calculateTechniqueLevel(mRow.getCharacter(), mNameField.getText(), getSpecialization(), ListRow.createCategoriesList(mCategoriesField.getText()), createNewDefault(), getSkillDifficulty(), getAdjustedSkillPoints(), true, mLimitCheckbox.isSelected(), getLimitModifier());
            mLevelField.setText(Technique.getTechniqueDisplayLevel(level.mLevel, level.mRelativeLevel, getDefaultModifier()));
            mLevelField.setToolTipText(Text.wrapPlainTextForToolTip(editorLevelTooltip() + level.getToolTip()));
        }
    }

    private SkillDefault createNewDefault() {
        SkillDefaultType type = getDefaultType();
        if (type.isSkillBased()) {
            return new SkillDefault(type, mDefaultNameField.getText(), mDefaultSpecializationField.getText(), getDefaultModifier());
        }
        return new SkillDefault(type, null, null, getDefaultModifier());
    }

    private SkillDifficulty getSkillDifficulty() {
        return (SkillDifficulty) mDifficultyCombo.getSelectedItem();
    }

    private int getPoints() {
        return Numbers.extractInteger(mPointsField.getText(), 0, true);
    }

    private int getAdjustedSkillPoints() {
        int            points    = getPoints();
        GURPSCharacter character = mRow.getCharacter();
        if (character != null) {
            String name = mNameField.getText();
            points += character.getSkillPointComparedIntegerBonusFor(Skill.ID_POINTS + "*", name, mDefaultSpecializationField.getText(), ListRow.createCategoriesList(mCategoriesField.getText()));
            points += character.getIntegerBonusFor(Skill.ID_POINTS + "/" + name.toLowerCase());
            if (points < 0) {
                points = 0;
            }
        }
        return points;
    }

    private int getDefaultModifier() {
        return Numbers.extractInteger(mDefaultModifierField.getText(), 0, true);
    }

    private int getLimitModifier() {
        return Numbers.extractInteger(mLimitField.getText(), 0, true);
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified = mRow.setName(mNameField.getText());
        modified |= mRow.setDefault(createNewDefault());
        modified |= mRow.setReference(mReferenceField.getText());
        modified |= mRow.setNotes(mNotesField.getText());
        modified |= mRow.setCategories(mCategoriesField.getText());
        if (mPointsField != null) {
            modified |= mRow.setRawPoints(getPoints());
        }
        modified |= mRow.setLimited(mLimitCheckbox.isSelected());
        modified |= mRow.setLimitModifier(getLimitModifier());
        modified |= mRow.setDifficulty(getSkillDifficulty());
        modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
        modified |= mRow.setFeatures(mFeatures.getFeatures());
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

        if (src == mLimitCheckbox) {
            mLimitField.setEnabled(mLimitCheckbox.isSelected());
        } else if (src == mDefaultTypeCombo) {
            if (mLastDefaultType != getDefaultType()) {
                rebuildDefaultPanel();
            }
        }

        if (src == mDifficultyCombo || src == mPointsField || src == mDefaultNameField || src == mDefaultModifierField || src == mLimitCheckbox || src == mLimitField || src == mDefaultSpecializationField || src == mDefaultTypeCombo) {
            recalculateLevel();
        }
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document doc = event.getDocument();
        if (doc == mNameField.getDocument()) {
            LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.Text("The name field may not be empty") : null);
        } else if (doc == mDefaultNameField.getDocument()) {
            LinkedLabel.setErrorMessage(mDefaultNameField, mDefaultNameField.getText().trim().isEmpty() ? I18n.Text("The default name field may not be empty") : null);
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
