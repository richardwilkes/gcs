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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.attribute.AttributeChoice;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.NumberFilter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.weapon.MeleeWeaponListEditor;
import com.trollworks.gcs.weapon.RangedWeaponListEditor;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Container;
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
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** The detailed editor for {@link Skill}s. */
public class SkillEditor extends RowEditor<Skill> implements ActionListener, DocumentListener, FocusListener {
    private JTextField                 mNameField;
    private JTextField                 mSpecializationField;
    private MultiLineTextField         mNotesField;
    private JTextField                 mCategoriesField;
    private JTextField                 mReferenceField;
    private JCheckBox                  mHasTechLevel;
    private JTextField                 mTechLevel;
    private String                     mSavedTechLevel;
    private JComboBox<AttributeChoice> mAttributePopup;
    private JComboBox<Object>          mDifficultyPopup;
    private JTextField                 mPointsField;
    private JTextField                 mLevelField;
    private JComboBox<Object>          mEncPenaltyPopup;
    private PrereqsPanel               mPrereqs;
    private FeaturesPanel              mFeatures;
    private Defaults                   mDefaults;
    private MeleeWeaponListEditor      mMeleeWeapons;
    private RangedWeaponListEditor     mRangedWeapons;

    /**
     * Creates a new {@link Skill} editor.
     *
     * @param skill The {@link Skill} to edit.
     */
    public SkillEditor(Skill skill) {
        super(skill);
        addContent();
    }

    @Override
    protected void addContentSelf(ScrollContent outer) {
        JPanel  panel       = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        boolean isContainer = mRow.canHaveChildren();
        mNameField = createCorrectableField(panel, I18n.text("Name"), mRow.getName(), I18n.text("The base name of the skill, without any notes or specialty information"));
        if (!isContainer) {
            JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2));
            mSpecializationField = createField(panel, wrapper, I18n.text("Specialization"), mRow.getSpecialization(), I18n.text("The specialization, if any, taken for this skill"), 0);
            createTechLevelFields(wrapper);
            panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        }
        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.text("Any notes that you would like to show up in the list along with this skill"), this);
        panel.add(new LinkedLabel(I18n.text("Notes"), mNotesField), new PrecisionLayoutData().setFillHorizontalAlignment().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2));
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mCategoriesField = createField(panel, panel, I18n.text("Categories"), mRow.getCategoriesAsString(), I18n.text("The category or categories the skill belongs to (separate multiple categories with a comma)"), 0);
        if (!isContainer) {
            createDifficultyPopups(panel);
            mEncPenaltyPopup = createEncumbrancePenaltyMultiplierPopup(panel);
        }
        mReferenceField = createField(panel, panel, I18n.text("Page Reference"), mRow.getReference(), PageRefCell.getStdToolTip(I18n.text("skill")), 0);
        outer.add(panel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        if (!isContainer) {
            mPrereqs = new PrereqsPanel(mRow, mRow.getPrereqs());
            addSection(outer, mPrereqs);
            mDefaults = new Defaults(mRow.getDataFile(), mRow.getDefaults());
            mDefaults.addActionListener(this);
            addSection(outer, mDefaults);
            mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
            addSection(outer, mFeatures);
            List<WeaponStats> weapons = mRow.getWeapons();
            mMeleeWeapons = new MeleeWeaponListEditor(mRow, weapons);
            addSection(outer, mMeleeWeapons);
            mRangedWeapons = new RangedWeaponListEditor(mRow, weapons);
            addSection(outer, mRangedWeapons);
        }
    }

    private JTextField createCorrectableField(Container parent, String title, String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        parent.add(new LinkedLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    private JTextField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
        JTextField field = new JTextField(maxChars > 0 ? Text.makeFiller(maxChars, 'M') : text);
        if (maxChars > 0) {
            UIUtilities.setToPreferredSizeOnly(field);
            field.setText(text);
        }
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.addActionListener(this);
        field.addFocusListener(this);
        labelParent.add(new LinkedLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        PrecisionLayoutData ld = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (maxChars == 0) {
            ld.setGrabHorizontalSpace(true);
        }
        fieldParent.add(field, ld);
        return field;
    }

    private static String editorLevelTooltip() {
        return I18n.text("The skill level and relative skill level to roll against.\n");
    }

    private void createPointsFields(Container parent, boolean forCharacter) {
        mPointsField = createField(parent, parent, I18n.text("Points"), Integer.toString(mRow.getRawPoints()),
                I18n.text("The number of points spent on this skill"), 4);
        NumberFilter.apply(mPointsField, false, false, false, 4);
        if (forCharacter) {
            String level = Skill.getSkillDisplayLevel(mRow.getDataFile(), mRow.getLevel(),
                    mRow.getRelativeLevel(), mRow.getAttribute(), mRow.canHaveChildren());
            mLevelField = createField(parent, parent, I18n.text("Level"), level,
                    editorLevelTooltip() + mRow.getLevelToolTip(), 8);
            mLevelField.setEnabled(false);
        }
    }

    private void createTechLevelFields(Container parent) {
        GURPSCharacter character = mRow.getCharacter();
        boolean        hasTL;

        mSavedTechLevel = mRow.getTechLevel();
        hasTL = mSavedTechLevel != null;
        if (!hasTL) {
            mSavedTechLevel = "";
        }

        if (character != null) {
            JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2));

            String tlTooltip = I18n.text("Whether this skill requires tech level specialization, and, if so, at what tech level it was learned");
            mHasTechLevel = new JCheckBox(I18n.text("Tech Level"), hasTL);
            mHasTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(tlTooltip));
            mHasTechLevel.addActionListener(this);
            wrapper.add(mHasTechLevel);

            mTechLevel = new JTextField("9999");
            UIUtilities.setToPreferredSizeOnly(mTechLevel);
            mTechLevel.setText(mSavedTechLevel);
            mTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(tlTooltip));
            mTechLevel.setEnabled(hasTL);
            wrapper.add(mTechLevel);
            parent.add(wrapper);

            if (!hasTL) {
                mSavedTechLevel = character.getProfile().getTechLevel();
            }
        } else {
            mTechLevel = new JTextField(mSavedTechLevel);
            mHasTechLevel = new JCheckBox(I18n.text("Tech Level Required"), hasTL);
            mHasTechLevel.setToolTipText(Text.wrapPlainTextForToolTip(I18n.text("Whether this skill requires tech level specialization")));
            mHasTechLevel.addActionListener(this);
            parent.add(mHasTechLevel);
        }
    }

    private JComboBox<Object> createEncumbrancePenaltyMultiplierPopup(Container parent) {
        Object[] items = new Object[10];
        items[0] = I18n.text("No penalty due to encumbrance");
        items[1] = I18n.text("Penalty equal to the current encumbrance level");
        for (int i = 2; i < 10; i++) {
            items[i] = MessageFormat.format(I18n.text("Penalty equal to {0} times the current encumbrance level"), Integer.valueOf(i));
        }
        LinkedLabel label = new LinkedLabel(I18n.text("Encumbrance"));
        parent.add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
        JComboBox<Object> popup = createComboBox(parent, items, items[mRow.getEncumbrancePenaltyMultiplier()], I18n.text("The encumbrance penalty multiplier"));
        label.setLink(popup);
        return popup;
    }

    private void createDifficultyPopups(Container parent) {
        List<AttributeChoice> list = new ArrayList<>();
        for (AttributeDef def : AttributeDef.getOrdered(mRow.getDataFile().getSheetSettings().getAttributes())) {
            list.add(new AttributeChoice(def.getID(), "%s", def.getName()));
        }
        list.add(new AttributeChoice("10", "%s", "10"));
        AttributeChoice current     = null;
        String          currentAttr = mRow.getAttribute();
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

        GURPSCharacter character              = mRow.getCharacter();
        int            columns                = 3;
        boolean        forCharacterOrTemplate = character != null || mRow.getTemplate() != null;
        if (forCharacterOrTemplate) {
            columns += 2;
            if (character != null) {
                columns += 2;
            }
        }
        JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(columns));
        mAttributePopup = createComboBox(wrapper, list.toArray(new AttributeChoice[0]), current, I18n.text("The attribute this skill is based on"));
        wrapper.add(new JLabel("/"));
        mDifficultyPopup = createComboBox(wrapper, SkillDifficulty.values(), mRow.getDifficulty(), I18n.text("The relative difficulty of learning this skill"));

        if (forCharacterOrTemplate) {
            createPointsFields(wrapper, character != null);
        }

        addLabel(parent, I18n.text("Difficulty"), null);
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private <T> JComboBox<T> createComboBox(Container parent, T[] items, T selection, String tooltip) {
        JComboBox<T> combo = new JComboBox<>(items);
        combo.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        combo.setSelectedItem(selection);
        combo.addActionListener(this);
        combo.setMaximumRowCount(items.length);
        parent.add(combo);
        return combo;
    }

    private void recalculateLevel() {
        if (mLevelField != null) {
            String attribute = getSkillAttribute();
            SkillLevel level = mRow.calculateLevel(mRow.getCharacter(), mNameField.getText(),
                    mSpecializationField.getText(), ListRow.createCategoriesList(mCategoriesField.getText()),
                    mDefaults.getDefaults(), attribute, getSkillDifficulty(), getAdjustedSkillPoints(),
                    new HashSet<>(), getEncumbrancePenaltyMultiplier());
            mLevelField.setText(Skill.getSkillDisplayLevel(mRow.getDataFile(), level.mLevel,
                    level.mRelativeLevel, attribute, false));
            mLevelField.setToolTipText(Text.wrapPlainTextForToolTip(editorLevelTooltip() + level.getToolTip()));
        }
    }

    private String getSkillAttribute() {
        AttributeChoice choice = (AttributeChoice) mAttributePopup.getSelectedItem();
        return choice != null ? choice.getAttribute() : Skill.getDefaultAttribute("dx");
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

    private void docChanged(DocumentEvent event) {
        if (mNameField.getDocument() == event.getDocument()) {
            LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.text("The name field may not be empty") : null);
        }
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        docChanged(event);
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        docChanged(event);
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        docChanged(event);
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
