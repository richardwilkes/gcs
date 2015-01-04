/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.widget.LinkedLabel;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.NumberFilter;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.TextUtility;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashSet;

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
public class SkillEditor extends RowEditor<Skill> implements ActionListener, DocumentListener {
	@Localize("Name")
	@Localize(locale = "de", value = "Name")
	@Localize(locale = "ru", value = "Название")
	private static String		NAME;
	@Localize("The base name of the skill, without any notes or specialty information")
	@Localize(locale = "de", value = "Der Name der Fertigkeit ohne Anmerkungen oder Spezialisierung")
	@Localize(locale = "ru", value = "Базовое название умения, без заметок или информации по специализации")
	private static String		NAME_TOOLTIP;
	@Localize("The name field may not be empty")
	@Localize(locale = "de", value = "Der Name darf nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Название\" не может быть пустым")
	private static String		NAME_CANNOT_BE_EMPTY;
	@Localize("Specialization")
	@Localize(locale = "de", value = "Spezialisierung")
	@Localize(locale = "ru", value = "Специализация")
	private static String		SPECIALIZATION;
	@Localize("The specialization, if any, taken for this skill")
	@Localize(locale = "de", value = "Die Spezialisierung, wenn für diese Fertigkeit eine genommen wurde")
	@Localize(locale = "ru", value = "Специализация, если есть, используемая для этого умения")
	private static String		SPECIALIZATION_TOOLTIP;
	@Localize("Categories")
	@Localize(locale = "de", value = "Kategorie")
	@Localize(locale = "ru", value = "Категории")
	private static String		CATEGORIES;
	@Localize("The category or categories the skill belongs to (separate multiple categories with a comma)")
	@Localize(locale = "de", value = "Die Kategorie oder Kategorien, denen diese Fertigkeit angehört (trenne mehrere Kategorien mit einem Komma)")
	@Localize(locale = "ru", value = "Категория или категории, к которым относится умение (перечислить через запятую)")
	private static String		CATEGORIES_TOOLTIP;
	@Localize("Notes")
	@Localize(locale = "de", value = "Anmerkungen")
	@Localize(locale = "ru", value = "Заметка")
	private static String		NOTES;
	@Localize("Any notes that you would like to show up in the list along with this skill")
	@Localize(locale = "de", value = "Anmerkungen, die in der Liste neben der Fertigkeit erscheinen sollen")
	@Localize(locale = "ru", value = "Заметки, которые показываются в списке рядом с умением")
	private static String		NOTES_TOOLTIP;
	@Localize("Tech Level")
	@Localize(locale = "de", value = "Techlevel")
	@Localize(locale = "ru", value = "Технологический уровень")
	private static String		TECH_LEVEL;
	@Localize("Whether this skill requires tech level specialization,\nand, if so, at what tech level it was learned")
	@Localize(locale = "de", value = "Ob diese Fertigkeit auf einen bestimmten Techlevel spezialisiert ist\nund wenn, mit welchem Techlevel sie gelernt wurde")
	@Localize(locale = "ru", value = "Для умения необходима специализации с технологическим уровнем \nс указанием уровня изучения")
	private static String		TECH_LEVEL_TOOLTIP;
	@Localize("Tech Level Required")
	@Localize(locale = "de", value = "Techlevel benötigt")
	@Localize(locale = "ru", value = "Необходимый технологический уровень")
	private static String		TECH_LEVEL_REQUIRED;
	@Localize("Whether this skill requires tech level specialization")
	@Localize(locale = "de", value = "Ob diese Fertigkeit auf einen bestimmten Techlevel spezialisiert ist")
	@Localize(locale = "ru", value = "Для умения необходима специализация с технологическим уровнем")
	private static String		TECH_LEVEL_REQUIRED_TOOLTIP;
	@Localize("Difficulty")
	@Localize(locale = "de", value = "Schwierigkeit")
	@Localize(locale = "ru", value = "Сложность")
	private static String		EDITOR_DIFFICULTY;
	@Localize("The difficulty of learning this skill")
	@Localize(locale = "de", value = "Die Schwierigkeit dieser Fertikgeit")
	@Localize(locale = "ru", value = "Сложность изучения умения")
	private static String		EDITOR_DIFFICULTY_TOOLTIP;
	@Localize("The relative difficulty of learning this skill")
	@Localize(locale = "de", value = "Die relative Schwierigkeit dieser Fertigkeit")
	@Localize(locale = "ru", value = "Относительная сложность изучения умения")
	private static String		EDITOR_DIFFICULTY_POPUP_TOOLTIP;
	@Localize("Level")
	@Localize(locale = "de", value = "Fertigkeitswert")
	@Localize(locale = "ru", value = "Уровень")
	private static String		EDITOR_LEVEL;
	@Localize("The skill level and relative skill level to roll against")
	@Localize(locale = "de", value = "Der Fertigkeitswert und relativer Fertigkeitswert, gegen die gewürfelt werden muss")
	@Localize(locale = "ru", value = "Уровень умения и относительный уровень умения для повторного броска")
	private static String		EDITOR_LEVEL_TOOLTIP;
	@Localize("The attribute this skill is based on")
	@Localize(locale = "de", value = "Das Attribut, auf dem diese Fertigkeit basiert")
	@Localize(locale = "ru", value = "Базовый атрибут умения ")
	private static String		ATTRIBUTE_POPUP_TOOLTIP;
	@Localize("Points")
	@Localize(locale = "de", value = "Punkte")
	@Localize(locale = "ru", value = "Очки")
	private static String		EDITOR_POINTS;
	@Localize("The number of points spent on this skill")
	@Localize(locale = "de", value = "Die Punkte, die für diese Fertigkeit aufgewendet wurden")
	@Localize(locale = "ru", value = "Потрачено на умение количество очков")
	private static String		EDITOR_POINTS_TOOLTIP;
	@Localize("Page Reference")
	@Localize(locale = "de", value = "Seitenangabe")
	@Localize(locale = "ru", value = "Ссылка на страницу")
	private static String		EDITOR_REFERENCE;
	@Localize("A reference to the book and page this skill appears\non (e.g. B22 would refer to \"Basic Set\", page 22)")
	@Localize(locale = "de", value = "Eine Referenz auf das Buch und die Seite, auf der diese Fertigkeit beschrieben wird (z.B. B22 würde auf \"Basic Set\" Seite 22 verweisen)")
	@Localize(locale = "ru", value = "Ссылка на страницу и книгу, описывающая умение\n (например B22 - книга \"Базовые правила\", страница 22)")
	private static String		REFERENCE_TOOLTIP;
	@Localize("Encumbrance")
	@Localize(locale = "de", value = "Belastung")
	@Localize(locale = "ru", value = "Нагрузка")
	private static String		ENC_PENALTY_MULT;
	@Localize("The encumbrance penalty multiplier")
	@Localize(locale = "de", value = "Der Belastungs-Malus-Multiplikator")
	@Localize(locale = "ru", value = "Множитель штрафа за нагрузку")
	private static String		ENC_PENALTY_MULT_TOOLTIP;
	@Localize("No penalty due to encumbrance")
	@Localize(locale = "de", value = "Kein Malus durch Belastung")
	@Localize(locale = "ru", value = "Нет штрафа за нагрузку")
	private static String		NO_ENC_PENALTY;
	@Localize("Penalty equal to the current encumbrance level")
	@Localize(locale = "de", value = "Malus gleich der aktuellen Belastungsstufe")
	@Localize(locale = "ru", value = "Штраф равен текущему уровню нагрузки")
	private static String		ONE_ENC_PENALTY;
	@Localize("Penalty equal to {0} times the current encumbrance level")
	@Localize(locale = "de", value = "Malus gleich dem {0}-fachen der aktuellen Belastungsstufe")
	@Localize(locale = "ru", value = "Штраф в {0} раза выше нынешнего уровня нагрузки")
	private static String		ENC_PENALTY_FORMAT;

	static {
		Localization.initialize();
	}

	private JTextField			mNameField;
	private JTextField			mSpecializationField;
	private JTextField			mNotesField;
	private JTextField			mCategoriesField;
	private JTextField			mReferenceField;
	private JCheckBox			mHasTechLevel;
	private JTextField			mTechLevel;
	private String				mSavedTechLevel;
	private JComboBox<Object>	mAttributePopup;
	private JComboBox<Object>	mDifficultyPopup;
	private JTextField			mPointsField;
	private JTextField			mLevelField;
	private JComboBox<Object>	mEncPenaltyPopup;
	private JTabbedPane			mTabPanel;
	private PrereqsPanel		mPrereqs;
	private FeaturesPanel		mFeatures;
	private Defaults			mDefaults;
	private MeleeWeaponEditor	mMeleeWeapons;
	private RangedWeaponEditor	mRangedWeapons;

	/**
	 * Creates a new {@link Skill} editor.
	 *
	 * @param skill The {@link Skill} to edit.
	 */
	public SkillEditor(Skill skill) {
		super(skill);

		JPanel content = new JPanel(new ColumnLayout(2));
		JPanel fields = new JPanel(new ColumnLayout(2));
		JLabel icon = new JLabel(skill.getIcon(true));
		boolean notContainer = !skill.canHaveChildren();
		Container wrapper;

		mNameField = createCorrectableField(fields, NAME, skill.getName(), NAME_TOOLTIP);
		if (notContainer) {
			wrapper = new JPanel(new ColumnLayout(2));
			mSpecializationField = createField(fields, wrapper, SPECIALIZATION, skill.getSpecialization(), SPECIALIZATION_TOOLTIP, 0);
			createTechLevelFields(wrapper);
			fields.add(wrapper);
			mEncPenaltyPopup = createEncumbrancePenaltyMultiplierPopup(fields);
		}
		mNotesField = createField(fields, fields, NOTES, skill.getNotes(), NOTES_TOOLTIP, 0);
		mCategoriesField = createField(fields, fields, CATEGORIES, skill.getCategoriesAsString(), CATEGORIES_TOOLTIP, 0);
		if (notContainer) {
			wrapper = createDifficultyPopups(fields);
		} else {
			wrapper = fields;
		}
		mReferenceField = createField(wrapper, wrapper, EDITOR_REFERENCE, mRow.getReference(), REFERENCE_TOOLTIP, 6);
		icon.setVerticalAlignment(SwingConstants.TOP);
		icon.setAlignmentY(-1f);
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
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.getDocument().addDocumentListener(this);

		LinkedLabel label = new LinkedLabel(title);
		label.setLink(field);

		parent.add(label);
		parent.add(field);
		return field;
	}

	private JTextField createField(Container labelParent, Container fieldParent, String title, String text, String tooltip, int maxChars) {
		JTextField field = new JTextField(maxChars > 0 ? TextUtility.makeFiller(maxChars, 'M') : text);

		if (maxChars > 0) {
			UIUtilities.setOnlySize(field, field.getPreferredSize());
			field.setText(text);
		}
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		labelParent.add(new LinkedLabel(title, field));
		fieldParent.add(field);
		return field;
	}

	@SuppressWarnings("unused")
	private void createPointsFields(Container parent, boolean forCharacter) {
		mPointsField = createField(parent, parent, EDITOR_POINTS, Integer.toString(mRow.getPoints()), EDITOR_POINTS_TOOLTIP, 4);
		new NumberFilter(mPointsField, false, false, false, 4);
		mPointsField.addActionListener(this);

		if (forCharacter) {
			mLevelField = createField(parent, parent, EDITOR_LEVEL, Skill.getSkillDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel(), mRow.getAttribute(), mRow.canHaveChildren()), EDITOR_LEVEL_TOOLTIP, 8);
			mLevelField.setEnabled(false);
		}
	}

	private void createTechLevelFields(Container parent) {
		OutlineModel owner = mRow.getOwner();
		GURPSCharacter character = mRow.getCharacter();
		boolean enabled = !owner.isLocked();
		boolean hasTL;

		mSavedTechLevel = mRow.getTechLevel();
		hasTL = mSavedTechLevel != null;
		if (!hasTL) {
			mSavedTechLevel = ""; //$NON-NLS-1$
		}

		if (character != null) {
			JPanel wrapper = new JPanel(new ColumnLayout(2));

			mHasTechLevel = new JCheckBox(TECH_LEVEL, hasTL);
			mHasTechLevel.setToolTipText(TECH_LEVEL_TOOLTIP);
			mHasTechLevel.setEnabled(enabled);
			mHasTechLevel.addActionListener(this);
			wrapper.add(mHasTechLevel);

			mTechLevel = new JTextField("9999"); //$NON-NLS-1$
			UIUtilities.setOnlySize(mTechLevel, mTechLevel.getPreferredSize());
			mTechLevel.setText(mSavedTechLevel);
			mTechLevel.setToolTipText(TECH_LEVEL_TOOLTIP);
			mTechLevel.setEnabled(enabled && hasTL);
			wrapper.add(mTechLevel);
			parent.add(wrapper);

			if (!hasTL) {
				mSavedTechLevel = character.getDescription().getTechLevel();
			}
		} else {
			mTechLevel = new JTextField(mSavedTechLevel);
			mHasTechLevel = new JCheckBox(TECH_LEVEL_REQUIRED, hasTL);
			mHasTechLevel.setToolTipText(TECH_LEVEL_REQUIRED_TOOLTIP);
			mHasTechLevel.setEnabled(enabled);
			mHasTechLevel.addActionListener(this);
			parent.add(mHasTechLevel);
		}
	}

	private JComboBox<Object> createEncumbrancePenaltyMultiplierPopup(Container parent) {
		Object[] items = new Object[10];
		items[0] = NO_ENC_PENALTY;
		items[1] = ONE_ENC_PENALTY;
		for (int i = 2; i < 10; i++) {
			items[i] = MessageFormat.format(ENC_PENALTY_FORMAT, new Integer(i));
		}
		LinkedLabel label = new LinkedLabel(ENC_PENALTY_MULT);
		parent.add(label);
		JComboBox<Object> popup = createComboBox(parent, items, items[mRow.getEncumbrancePenaltyMultiplier()], ENC_PENALTY_MULT_TOOLTIP);
		label.setLink(popup);
		return popup;
	}

	private Container createDifficultyPopups(Container parent) {
		GURPSCharacter character = mRow.getCharacter();
		boolean forCharacterOrTemplate = character != null || mRow.getTemplate() != null;
		JLabel label = new JLabel(EDITOR_DIFFICULTY, SwingConstants.RIGHT);
		JPanel wrapper = new JPanel(new ColumnLayout(forCharacterOrTemplate ? character != null ? 10 : 8 : 6));

		label.setToolTipText(EDITOR_DIFFICULTY_TOOLTIP);

		mAttributePopup = createComboBox(wrapper, SkillAttribute.values(), mRow.getAttribute(), ATTRIBUTE_POPUP_TOOLTIP);
		wrapper.add(new JLabel(" /")); //$NON-NLS-1$
		mDifficultyPopup = createComboBox(wrapper, SkillDifficulty.values(), mRow.getDifficulty(), EDITOR_DIFFICULTY_POPUP_TOOLTIP);

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
		combo.setToolTipText(tooltip);
		combo.setSelectedItem(selection);
		combo.addActionListener(this);
		combo.setMaximumRowCount(items.length);
		UIUtilities.setOnlySize(combo, combo.getPreferredSize());
		combo.setEnabled(mIsEditable);
		parent.add(combo);
		return combo;
	}

	private void recalculateLevel() {
		if (mLevelField != null) {
			SkillAttribute attribute = getSkillAttribute();
			SkillLevel level = mRow.calculateLevel(mRow.getCharacter(), mNameField.getText(), mSpecializationField.getText(), mDefaults.getDefaults(), attribute, getSkillDifficulty(), getSkillPoints(), new HashSet<String>(), getEncumbrancePenaltyMultiplier());
			mLevelField.setText(Skill.getSkillDisplayLevel(level.mLevel, level.mRelativeLevel, attribute, false));
		}
	}

	private SkillAttribute getSkillAttribute() {
		return (SkillAttribute) mAttributePopup.getSelectedItem();
	}

	private SkillDifficulty getSkillDifficulty() {
		return (SkillDifficulty) mDifficultyPopup.getSelectedItem();
	}

	private int getSkillPoints() {
		return Numbers.getLocalizedInteger(mPointsField.getText(), 0);
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
			modified |= mRow.setPoints(getSkillPoints());
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
				mTechLevel.setText(""); //$NON-NLS-1$
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
		LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().length() != 0 ? null : NAME_CANNOT_BE_EMPTY);
	}
}
