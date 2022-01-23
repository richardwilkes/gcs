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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.attribute.AttributeChoice;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
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
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** The detailed editor for {@link Skill}s. */
public class SkillEditor extends RowEditor<Skill> implements ActionListener, DocumentListener, FocusListener {
    private EditorField                mNameField;
    private EditorField                mSpecializationField;
    private MultiLineTextField         mNotesField;
    private MultiLineTextField         mVTTNotesField;
    private EditorField                mCategoriesField;
    private EditorField                mReferenceField;
    private Checkbox                   mHasTechLevel;
    private EditorField                mTechLevel;
    private String                     mSavedTechLevel;
    private PopupMenu<AttributeChoice> mAttributePopup;
    private PopupMenu<Object>          mDifficultyPopup;
    private EditorField                mPointsField;
    private EditorField                mLevelField;
    private PopupMenu<String>          mEncPenaltyPopup;
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
        Panel   panel       = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
        boolean isContainer = mRow.canHaveChildren();
        addLabel(panel, I18n.text("Name"));
        mNameField = createCorrectableField(panel, mRow.getName(),
                I18n.text("The base name of the skill, without any notes or specialty information"));
        if (!isContainer) {
            addLabel(panel, I18n.text("Specialization"));
            Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
            mSpecializationField = createField(wrapper, mRow.getSpecialization(),
                    I18n.text("The specialization, if any, taken for this skill"), 0, null);
            createTechLevelFields(wrapper);
            panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        }
        mNotesField = new MultiLineTextField(mRow.getNotes(),
                I18n.text("Any notes that you would like to show up in the list along with this skill"), this);
        addLabel(panel, I18n.text("Notes")).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2);
        panel.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mVTTNotesField = addVTTNotesField(panel, this);
        addLabel(panel, I18n.text("Categories"));
        mCategoriesField = createField(panel, mRow.getCategoriesAsString(),
                I18n.text("The category or categories the skill belongs to (separate multiple categories with a comma)"),
                0, null);
        if (!isContainer) {
            createDifficultyPopups(panel);
            mEncPenaltyPopup = createEncumbrancePenaltyMultiplierPopup(panel);
        }
        addLabel(panel, I18n.text("Page Reference"));
        mReferenceField = createField(panel, mRow.getReference(),
                PageRefCell.getStdToolTip(I18n.text("skill")), 0, null);
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

    private EditorField createCorrectableField(Container parent, String text, String tooltip) {
        EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, text, tooltip);
        field.getDocument().addDocumentListener(this);
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    private static EditorField createField(Container parent, String text, String tooltip, int maxChars, EditorField.ChangeListener listener) {
        EditorField field = new EditorField(FieldFactory.STRING, listener, SwingConstants.LEFT, text,
                maxChars > 0 ? Text.makeFiller(maxChars, 'M') : null, tooltip);
        PrecisionLayoutData ld = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (maxChars == 0) {
            ld.setGrabHorizontalSpace(true);
        }
        parent.add(field, ld);
        return field;
    }

    private static String editorLevelTooltip() {
        return I18n.text("The skill level and relative skill level to roll against.\n");
    }

    private void createPointsFields(Container parent, boolean forCharacter) {
        mPointsField = new EditorField(FieldFactory.POSINT3, (f) -> recalculateLevel(),
                SwingConstants.LEFT, Integer.valueOf(mRow.getRawPoints()), Integer.valueOf(999),
                I18n.text("The number of points spent on this skill"));
        addInteriorLabel(parent, I18n.text("Points"));
        parent.add(mPointsField, new PrecisionLayoutData().setFillHorizontalAlignment());
        if (forCharacter) {
            addInteriorLabel(parent, I18n.text("Level"));
            mLevelField = createField(parent, Skill.getSkillDisplayLevel(mRow.getDataFile(),
                    mRow.getLevel(), mRow.getRelativeLevel(), mRow.getAttribute(),
                    mRow.canHaveChildren()), editorLevelTooltip() + mRow.getLevelToolTip(), 8, null);
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
            Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));

            String tlTooltip = I18n.text("Whether this skill requires tech level specialization, and, if so, at what tech level it was learned");
            mHasTechLevel = new Checkbox(I18n.text("Tech Level"), hasTL, this::clickedOnHasTechLevelCheckbox);
            mHasTechLevel.setToolTipText(tlTooltip);
            wrapper.add(mHasTechLevel, new PrecisionLayoutData().setLeftMargin(4));

            mTechLevel = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT,
                    mSavedTechLevel, "9999", tlTooltip);
            mTechLevel.setEnabled(hasTL);
            wrapper.add(mTechLevel);
            parent.add(wrapper);

            if (!hasTL) {
                mSavedTechLevel = character.getProfile().getTechLevel();
            }
        } else {
            // mTechLevel is created here -- despite not being used in this case -- so that all the
            // places where it is used don't have to be conditionalized.
            mTechLevel = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT,
                    mSavedTechLevel, "9999", null);
            mHasTechLevel = new Checkbox(I18n.text("Tech Level Required"), hasTL, this::clickedOnHasTechLevelCheckbox);
            mHasTechLevel.setToolTipText(I18n.text("Whether this skill requires tech level specialization"));
            parent.add(mHasTechLevel, new PrecisionLayoutData().setLeftMargin(4));
        }
    }

    private void clickedOnHasTechLevelCheckbox(Checkbox checkbox) {
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

    private PopupMenu<String> createEncumbrancePenaltyMultiplierPopup(Container parent) {
        String[] items = new String[10];
        items[0] = I18n.text("No penalty due to encumbrance");
        items[1] = I18n.text("Penalty equal to the current encumbrance level");
        for (int i = 2; i < 10; i++) {
            items[i] = MessageFormat.format(I18n.text("Penalty equal to {0} times the current encumbrance level"),
                    Integer.valueOf(i));
        }
        addLabel(parent, I18n.text("Encumbrance"));
        return createPopup(parent, items,
                items[mRow.getEncumbrancePenaltyMultiplier()],
                I18n.text("The encumbrance penalty multiplier"), (p) -> recalculateLevel());
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
        Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(columns));
        mAttributePopup = createPopup(wrapper, list.toArray(new AttributeChoice[0]), current,
                I18n.text("The attribute this skill is based on"), (p) -> recalculateLevel());
        addLabel(wrapper, "/");
        mDifficultyPopup = createPopup(wrapper, SkillDifficulty.values(), mRow.getDifficulty(),
                I18n.text("The relative difficulty of learning this skill"),
                (p) -> recalculateLevel());

        if (forCharacterOrTemplate) {
            createPointsFields(wrapper, character != null);
        }

        addLabel(parent, I18n.text("Difficulty"));
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private static <T> PopupMenu<T> createPopup(Container parent, T[] items, T selection, String tooltip, PopupMenu.SelectionListener<T> listener) {
        PopupMenu<T> popup = new PopupMenu<>(items, listener);
        popup.setToolTipText(tooltip);
        popup.setSelectedItem(selection, false);
        parent.add(popup);
        return popup;
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
            mLevelField.setToolTipText(editorLevelTooltip() + level.getToolTip());
        }
    }

    private String getSkillAttribute() {
        AttributeChoice choice = mAttributePopup.getSelectedItem();
        return choice != null ? choice.getAttribute() : Skill.getDefaultAttribute("dx");
    }

    private SkillDifficulty getSkillDifficulty() {
        return (SkillDifficulty) mDifficultyPopup.getSelectedItem();
    }

    private int getSkillPoints() {
        return ((Integer) mPointsField.getValue()).intValue();
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
        modified |= mRow.setVTTNotes(mVTTNotesField.getText());
        modified |= mRow.setCategories(mCategoriesField.getText());
        if (mSpecializationField != null) {
            modified |= mRow.setSpecialization(mSpecializationField.getText());
        }
        if (mHasTechLevel != null) {
            modified |= mRow.setTechLevel(mHasTechLevel.isChecked() ? mTechLevel.getText() : null);
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
        if (src == mDefaults) {
            recalculateLevel();
        }
    }

    private void docChanged(DocumentEvent event) {
        if (mNameField.getDocument() == event.getDocument()) {
            mNameField.setErrorMessage(mNameField.getText().trim().isEmpty() ? I18n.text("The name field may not be empty") : null);
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
