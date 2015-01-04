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
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.NumberFilter;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.TextUtility;

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

/** The detailed editor for {@link Technique}s. */
public class TechniqueEditor extends RowEditor<Technique> implements ActionListener, DocumentListener {
	@Localize("Name")
	@Localize(locale = "de", value = "Name")
	@Localize(locale = "ru", value = "Название")
	private static String		NAME;
	@Localize("Notes")
	@Localize(locale = "de", value = "Anmerkungen")
	@Localize(locale = "ru", value = "Заметка")
	private static String		NOTES;
	@Localize("Page Reference")
	@Localize(locale = "de", value = "Seitenangabe")
	@Localize(locale = "ru", value = "Ссылка на страницу")
	private static String		EDITOR_REFERENCE;
	@Localize("Points")
	@Localize(locale = "de", value = "Punkte")
	@Localize(locale = "ru", value = "Очки")
	private static String		EDITOR_POINTS;
	@Localize("Level")
	@Localize(locale = "de", value = "Fertigkeitswert")
	@Localize(locale = "ru", value = "Уровень")
	private static String		EDITOR_LEVEL;
	@Localize("The skill level and relative skill level to roll against")
	@Localize(locale = "de", value = "Der Fertigkeitswert und relativer Fertigkeitswert, gegen die gewürfelt werden muss")
	@Localize(locale = "ru", value = "Уровень умения и относительный уровень умения для повторного броска")
	private static String		EDITOR_LEVEL_TOOLTIP;
	@Localize("Difficulty")
	@Localize(locale = "de", value = "Schwierigkeit")
	@Localize(locale = "ru", value = "Сложность")
	private static String		EDITOR_DIFFICULTY;
	@Localize("The base name of the technique, without any notes or specialty information")
	@Localize(locale = "de", value = "Der Name der Technik ohne Anmerkungen oder Spezialisierung")
	@Localize(locale = "ru", value = "Базовое название техники, без заметок или информации по специализации")
	private static String		TECHNIQUE_NAME_TOOLTIP;
	@Localize("The name field may not be empty")
	@Localize(locale = "de", value = "Der Name darf nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Название\" не может быть пустым")
	private static String		TECHNIQUE_NAME_CANNOT_BE_EMPTY;
	@Localize("Any notes that you would like to show up in the list along with this technique")
	@Localize(locale = "de", value = "Anmerkungen, die in der Liste neben der Technik erscheinen sollen")
	@Localize(locale = "ru", value = "Заметки, которые показываются в списке рядом с техникой")
	private static String		TECHNIQUE_NOTES_TOOLTIP;
	@Localize("The difficulty of learning this technique")
	@Localize(locale = "de", value = "Die Schwierigkeit dieser Fertikgeit")
	@Localize(locale = "ru", value = "Сложность изучения техники")
	private static String		TECHNIQUE_DIFFICULTY_TOOLTIP;
	@Localize("The relative difficulty of learning this technique")
	@Localize(locale = "de", value = "Die relative Schwierigkeit dieser Fertigkeit")
	@Localize(locale = "ru", value = "Относительная сложность изучения техники")
	private static String		TECHNIQUE_DIFFICULTY_POPUP_TOOLTIP;
	@Localize("The number of points spent on this technique")
	@Localize(locale = "de", value = "Die Punkte, die für diese Fertigkeit aufgewendet wurden")
	@Localize(locale = "ru", value = "Потрачено на технику количество очков")
	private static String		TECHNIQUE_POINTS_TOOLTIP;
	@Localize("A reference to the book and page this technique appears\non (e.g. B22 would refer to \"Basic Set\", page 22)")
	@Localize(locale = "de", value = "Eine Referenz auf das Buch und die Seite, auf der diese Technik beschrieben wird (z.B. B22 würde auf \"Basic Set\" Seite 22 verweisen)")
	@Localize(locale = "ru", value = "Ссылка на страницу и книгу, описывающая технику\n (например B22 - книга \"Базовые правила\", страница 22)")
	private static String		TECHNIQUE_REFERENCE_TOOLTIP;
	@Localize("Defaults To")
	@Localize(locale = "de", value = "Grundwert auf")
	@Localize(locale = "ru", value = "По умолчанию к")
	private static String		DEFAULTS_TO;
	@Localize("The name of the skill this technique defaults from")
	@Localize(locale = "de", value = "Der Name der Fertigkeit, von welcher diese Technik ihren Grundwert bezieht")
	@Localize(locale = "ru", value = "Название умения, для которой предназначена техника")
	private static String		DEFAULTS_TO_TOOLTIP;
	@Localize("The default name field may not be empty")
	@Localize(locale = "de", value = "Der Grundwert darf nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Название по умолчанию\" не может быть пустым")
	private static String		DEFAULT_NAME_CANNOT_BE_EMPTY;
	@Localize("The specialization of the skill, if any, this technique defaults from")
	@Localize(locale = "de", value = "Die Spezialisierung für diese Fertigkeit, wenn eine genommen wurde, von welcher diese Technik ihren Grundwert bezieht")
	@Localize(locale = "ru", value = "Специализация умения, если есть, для которой предназначена техника")
	private static String		DEFAULT_SPECIALIZATION_TOOLTIP;
	@Localize("The amount to adjust the default skill level by")
	@Localize(locale = "de", value = "Um wieviel der Fertigkeitswert des Grundwerts angepasst wird")
	@Localize(locale = "ru", value = "Значение уровня умения по умолчанию")
	private static String		DEFAULT_MODIFIER_TOOLTIP;
	@Localize("Cannot exceed default skill level by more than")
	@Localize(locale = "de", value = "Kann den Fertigkeitswert des Grundwerts nicht übersteigen um mehr als")
	@Localize(locale = "ru", value = "Не может превышать уровень умения по умолчанию, больше чем на")
	private static String		LIMIT;
	@Localize("Whether to limit the maximum level that can be achieved or not")
	@Localize(locale = "de", value = "Ob der Fertigkeitswert der Technik eine Obergrenze hat")
	@Localize(locale = "ru", value = "Ограничить максимальный достижимый уровень")
	private static String		LIMIT_TOOLTIP;
	@Localize("The maximum amount above the default skill level that this technique can be raised")
	@Localize(locale = "de", value = "Die maximale Erhöhung über den Fertigkeitswert des Grundwerts.")
	@Localize(locale = "ru", value = "Максимальное значение выше уровня умения по умолчанию, при котором эта техника может быть использована")
	private static String		LIMIT_AMOUNT_TOOLTIP;
	@Localize("Categories")
	@Localize(locale = "de", value = "Kategorie")
	@Localize(locale = "ru", value = "Категории")
	private static String		CATEGORIES;
	@Localize("The category or categories the technique belongs to (separate multiple categories with a comma)")
	@Localize(locale = "de", value = "Die Kategorie oder Kategorien, denen diese Fertigkeit angehört (trenne mehrere Kategorien mit einem Komma)")
	@Localize(locale = "ru", value = "Категория или категории, к которым относится техника (перечислить через запятую)")
	private static String		CATEGORIES_TOOLTIP;

	static {
		Localization.initialize();
	}

	private JTextField			mNameField;
	private JTextField			mNotesField;
	private JTextField			mCategoriesField;
	private JTextField			mReferenceField;
	private JComboBox<Object>	mDifficultyCombo;
	private JTextField			mPointsField;
	private JTextField			mLevelField;
	private JPanel				mDefaultPanel;
	private LinkedLabel			mDefaultPanelLabel;
	private JComboBox<Object>	mDefaultTypeCombo;
	private JTextField			mDefaultNameField;
	private JTextField			mDefaultSpecializationField;
	private JTextField			mDefaultModifierField;
	private JCheckBox			mLimitCheckbox;
	private JTextField			mLimitField;
	private JTabbedPane			mTabPanel;
	private PrereqsPanel		mPrereqs;
	private FeaturesPanel		mFeatures;
	private SkillDefaultType	mLastDefaultType;
	private MeleeWeaponEditor	mMeleeWeapons;
	private RangedWeaponEditor	mRangedWeapons;

	/**
	 * Creates a new {@link Technique} editor.
	 *
	 * @param technique The {@link Technique} to edit.
	 */
	public TechniqueEditor(Technique technique) {
		super(technique);

		JPanel content = new JPanel(new ColumnLayout(2));
		JPanel fields = new JPanel(new ColumnLayout(2));
		JLabel icon = new JLabel(technique.getIcon(true));
		Container wrapper;

		mNameField = createCorrectableField(fields, fields, NAME, technique.getName(), TECHNIQUE_NAME_TOOLTIP);
		mNotesField = createField(fields, fields, NOTES, technique.getNotes(), TECHNIQUE_NOTES_TOOLTIP, 0);
		mCategoriesField = createField(fields, fields, CATEGORIES, technique.getCategoriesAsString(), CATEGORIES_TOOLTIP, 0);
		createDefaults(fields);
		createLimits(fields);
		wrapper = createDifficultyPopups(fields);
		mReferenceField = createField(wrapper, wrapper, EDITOR_REFERENCE, mRow.getReference(), TECHNIQUE_REFERENCE_TOOLTIP, 6);
		icon.setVerticalAlignment(SwingConstants.TOP);
		icon.setAlignmentY(-1f);
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
		mDefaultPanelLabel = new LinkedLabel(DEFAULTS_TO);
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
		UIUtilities.setOnlySize(combo, combo.getPreferredSize());
		parent.add(combo);
		return combo;
	}

	private SkillDefaultType getDefaultType() {
		return (SkillDefaultType) mDefaultTypeCombo.getSelectedItem();
	}

	private String getSpecialization() {
		StringBuilder builder = new StringBuilder();
		String specialization = mDefaultSpecializationField.getText();

		builder.append(mDefaultNameField.getText());
		if (specialization.length() > 0) {
			builder.append(" ("); //$NON-NLS-1$
			builder.append(specialization);
			builder.append(')');
		}
		return builder.toString();
	}

	private void rebuildDefaultPanel() {
		SkillDefault def = mRow.getDefault();
		boolean skillBased;

		mLastDefaultType = getDefaultType();
		skillBased = mLastDefaultType.isSkillBased();
		UIUtilities.forceFocusToAccept();
		while (mDefaultPanel.getComponentCount() > 1) {
			mDefaultPanel.remove(1);
		}
		if (skillBased) {
			mDefaultNameField = createCorrectableField(null, mDefaultPanel, DEFAULTS_TO, def.getName(), DEFAULTS_TO_TOOLTIP);
			mDefaultSpecializationField = createField(null, mDefaultPanel, null, def.getSpecialization(), DEFAULT_SPECIALIZATION_TOOLTIP, 0);
			mDefaultPanelLabel.setLink(mDefaultNameField);
		}
		mDefaultModifierField = createNumberField(null, mDefaultPanel, null, DEFAULT_MODIFIER_TOOLTIP, def.getModifier(), 2);
		if (!skillBased) {
			mDefaultPanel.add(new JPanel());
			mDefaultPanel.add(new JPanel());
		}
		mDefaultPanel.revalidate();
	}

	private void createLimits(Container parent) {
		JPanel wrapper = new JPanel(new ColumnLayout(3));

		mLimitCheckbox = new JCheckBox(LIMIT, mRow.isLimited());
		mLimitCheckbox.setToolTipText(LIMIT_TOOLTIP);
		mLimitCheckbox.addActionListener(this);
		mLimitCheckbox.setEnabled(mIsEditable);

		mLimitField = createNumberField(null, wrapper, null, LIMIT_AMOUNT_TOOLTIP, mRow.getLimitModifier(), 2);
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
		field.setToolTipText(tooltip);
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
		JTextField field = new JTextField(maxChars > 0 ? TextUtility.makeFiller(maxChars, 'M') : text);

		if (maxChars > 0) {
			UIUtilities.setOnlySize(field, field.getPreferredSize());
			field.setText(text);
		}
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.addActionListener(this);
		if (labelParent != null) {
			labelParent.add(new LinkedLabel(title, field));
		}
		fieldParent.add(field);
		return field;
	}

	@SuppressWarnings("unused")
	private JTextField createNumberField(Container labelParent, Container fieldParent, String title, String tooltip, int value, int maxDigits) {
		JTextField field = createField(labelParent, fieldParent, title, Numbers.formatWithForcedSign(value), tooltip, maxDigits + 1);
		new NumberFilter(field, false, true, false, maxDigits);
		return field;
	}

	@SuppressWarnings("unused")
	private void createPointsFields(Container parent, boolean forCharacter) {
		mPointsField = createField(parent, parent, EDITOR_POINTS, Integer.toString(mRow.getPoints()), TECHNIQUE_POINTS_TOOLTIP, 4);
		new NumberFilter(mPointsField, false, false, false, 4);
		mPointsField.addActionListener(this);

		if (forCharacter) {
			mLevelField = createField(parent, parent, EDITOR_LEVEL, Technique.getTechniqueDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel(), mRow.getDefault().getModifier()), EDITOR_LEVEL_TOOLTIP, 6);
			mLevelField.setEnabled(false);
		}
	}

	private Container createDifficultyPopups(Container parent) {
		GURPSCharacter character = mRow.getCharacter();
		boolean forCharacterOrTemplate = character != null || mRow.getTemplate() != null;
		JLabel label = new JLabel(EDITOR_DIFFICULTY, SwingConstants.RIGHT);
		JPanel wrapper = new JPanel(new ColumnLayout(forCharacterOrTemplate ? character != null ? 8 : 6 : 4));

		label.setToolTipText(TECHNIQUE_DIFFICULTY_TOOLTIP);

		mDifficultyCombo = createComboBox(wrapper, new Object[] { SkillDifficulty.A, SkillDifficulty.H }, mRow.getDifficulty());
		mDifficultyCombo.setToolTipText(TECHNIQUE_DIFFICULTY_POPUP_TOOLTIP);
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
			SkillLevel level = Technique.calculateTechniqueLevel(mRow.getCharacter(), mNameField.getText(), getSpecialization(), createNewDefault(), getSkillDifficulty(), getPoints(), mLimitCheckbox.isSelected(), getLimitModifier());

			mLevelField.setText(Technique.getTechniqueDisplayLevel(level.mLevel, level.mRelativeLevel, getDefaultModifier()));
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
		return Numbers.getLocalizedInteger(mPointsField.getText(), 0);
	}

	private int getDefaultModifier() {
		return Numbers.getLocalizedInteger(mDefaultModifierField.getText(), 0);
	}

	private int getLimitModifier() {
		return Numbers.getLocalizedInteger(mLimitField.getText(), 0);
	}

	@Override
	public boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());

		modified |= mRow.setDefault(createNewDefault());
		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setNotes(mNotesField.getText());
		modified |= mRow.setCategories(mCategoriesField.getText());
		if (mPointsField != null) {
			modified |= mRow.setPoints(getPoints());
		}
		modified |= mRow.setLimited(mLimitCheckbox.isSelected());
		modified |= mRow.setLimitModifier(getLimitModifier());
		modified |= mRow.setDifficulty(getSkillDifficulty());
		modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
		modified |= mRow.setFeatures(mFeatures.getFeatures());

		ArrayList<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
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
			LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().length() != 0 ? null : TECHNIQUE_NAME_CANNOT_BE_EMPTY);
		} else if (doc == mDefaultNameField.getDocument()) {
			LinkedLabel.setErrorMessage(mDefaultNameField, mDefaultNameField.getText().trim().length() != 0 ? null : DEFAULT_NAME_CANNOT_BE_EMPTY);
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
