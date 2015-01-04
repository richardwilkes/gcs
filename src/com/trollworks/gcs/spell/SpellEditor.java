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

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.SkillAttribute;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
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
	@Localize("Name")
	@Localize(locale = "de", value = "Name")
	@Localize(locale = "ru", value = "Название")
	private static String		NAME;
	@Localize("The name of the spell, without any notes")
	@Localize(locale = "de", value = "Der Name des Zaubers ohne Anmerkungen")
	@Localize(locale = "ru", value = "Название заклинания без заметок")
	private static String		NAME_TOOLTIP;
	@Localize("The name field may not be empty")
	@Localize(locale = "de", value = "Der Name darf nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Название\" не может быть пустым")
	private static String		NAME_CANNOT_BE_EMPTY;
	@Localize("Tech Level")
	@Localize(locale = "de", value = "Techlevel")
	@Localize(locale = "ru", value = "Технологический уровень")
	private static String		TECH_LEVEL;
	@Localize("Whether this spell requires tech level specialization,\nand, if so, at what tech level it was learned")
	@Localize(locale = "de", value = "Ob dieser Zauber auf einen bestimmten Techlevel spezialisiert ist\nund wenn, mit welchem Techlevel er gelernt wurde")
	@Localize(locale = "ru", value = "Для заклинания необходима специализация с технологическим уровнем \nс указанием уровня изучения")
	private static String		TECH_LEVEL_TOOLTIP;
	@Localize("Tech Level Required")
	@Localize(locale = "de", value = "Techlevel benötigt")
	@Localize(locale = "ru", value = "Необходимый технологический уровень")
	private static String		TECH_LEVEL_REQUIRED;
	@Localize("Whether this spell requires tech level specialization")
	@Localize(locale = "de", value = "Ob dieser Zauber auf einen bestimmten Techlevel spezialisiert ist")
	@Localize(locale = "ru", value = "Для заклинания необходима специализация с технологическим уровнем")
	private static String		TECH_LEVEL_REQUIRED_TOOLTIP;
	@Localize("College")
	@Localize(locale = "de", value = "Schule")
	@Localize(locale = "ru", value = "Школа")
	private static String		COLLEGE;
	@Localize("The college the spell belongs to")
	@Localize(locale = "de", value = "Die Schule, die dieser Zauber angehört")
	@Localize(locale = "ru", value = "Школа, к которой относится заклинание")
	private static String		COLLEGE_TOOLTIP;
	@Localize("Power Source")
	@Localize(locale = "de", value = "Energiequelle")
	@Localize(locale = "ru", value = "Источник силы")
	private static String		POWER_SOURCE;
	@Localize("The source of power for the spell")
	@Localize(locale = "de", value = "Die Quelle der Energie für den Zauber")
	@Localize(locale = "ru", value = "Источник силы для заклинания")
	private static String		POWER_SOURCE_TOOLTIP;
	@Localize("Class")
	@Localize(locale = "de", value = "Klasse")
	@Localize(locale = "ru", value = "Класс")
	private static String		CLASS;
	@Localize("The class of spell (Area, Missile, etc.)")
	@Localize(locale = "de", value = "Die Klasse des Zaubers (Gebiet, Geschoss, usw.)")
	@Localize(locale = "ru", value = "Класс заклинания (областные, метательные и т.д.)")
	private static String		CLASS_ONLY_TOOLTIP;
	@Localize("The class field may not be empty")
	@Localize(locale = "de", value = "Die Klasse darf nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Класс\" не может быть пустым")
	private static String		CLASS_CANNOT_BE_EMPTY;
	@Localize("Casting Cost")
	@Localize(locale = "de", value = "Zauberkosten")
	@Localize(locale = "ru", value = "Стоимость заклинания")
	private static String		CASTING_COST;
	@Localize("The casting cost of the spell")
	@Localize(locale = "de", value = "Die Kosten, um den Zauber zu wirken")
	@Localize(locale = "ru", value = "Стоимость сотворения заклинания")
	private static String		CASTING_COST_TOOLTIP;
	@Localize("The casting cost field may not be empty")
	@Localize(locale = "de", value = "Die Zauberkosten dürfen nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Мана-Стоимость\" не может быть пустым")
	private static String		CASTING_COST_CANNOT_BE_EMPTY;
	@Localize("Maintenance Cost")
	@Localize(locale = "de", value = "Erhaltungskosten")
	@Localize(locale = "ru", value = "Стоимость обслуживания")
	private static String		MAINTENANCE_COST;
	@Localize("The cost to maintain a spell after its initial duration")
	@Localize(locale = "de", value = "Die Kosten, um den Zauber aufrecht zu erhalten")
	@Localize(locale = "ru", value = "Стоимость поддержки заклинания свыше исходной длительности")
	private static String		MAINTENANCE_COST_TOOLTIP;
	@Localize("Casting Time")
	@Localize(locale = "de", value = "Zauberzeit")
	@Localize(locale = "ru", value = "Время сотворения")
	private static String		CASTING_TIME;
	@Localize("The casting time of the spell")
	@Localize(locale = "de", value = "Die Zeit, um den Zauber zu wirken")
	@Localize(locale = "ru", value = "Время сотворения заклинания")
	private static String		CASTING_TIME_TOOLTIP;
	@Localize("The casting time field may not be empty")
	@Localize(locale = "de", value = "Die Zauberzeit darf nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Время сотворения\" не может быть пустым")
	private static String		CASTING_TIME_CANNOT_BE_EMPTY;
	@Localize("Duration")
	@Localize(locale = "de", value = "Dauer")
	@Localize(locale = "ru", value = "Длительность")
	private static String		DURATION;
	@Localize("The duration of the spell once its cast")
	@Localize(locale = "de", value = "Die Dauer des Zaubers, nachdem er gewirkt wurde")
	@Localize(locale = "ru", value = "Длительность заклинания после сотворения")
	private static String		DURATION_TOOLTIP;
	@Localize("The duration field may not be empty")
	@Localize(locale = "de", value = "Die Dauer darf nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Длительность\" не может быть пустым")
	private static String		DURATION_CANNOT_BE_EMPTY;
	@Localize("Categories")
	@Localize(locale = "de", value = "Kategorie")
	@Localize(locale = "ru", value = "Категории")
	private static String		CATEGORIES;
	@Localize("The category or categories the spell belongs to (separate multiple categories with a comma)")
	@Localize(locale = "de", value = "Die Kategorie oder Kategorien, denen dieser Zauber angehört (trenne mehrere Kategorien mit einem Komma)")
	@Localize(locale = "ru", value = "Категория или категории, к которым относится заклинание (перечислить через запятую)")
	private static String		CATEGORIES_TOOLTIP;
	@Localize("Notes")
	@Localize(locale = "de", value = "Anmerkungen")
	@Localize(locale = "ru", value = "Заметка")
	private static String		NOTES;
	@Localize("Any notes that you would like to show up in the list along with this spell")
	@Localize(locale = "de", value = "Anmerkungen, die in der Liste neben dem Zauber erscheinen sollen")
	@Localize(locale = "ru", value = "Заметки, которые показываются в списке рядом с заклинанием")
	private static String		NOTES_TOOLTIP;
	@Localize("Points")
	@Localize(locale = "de", value = "Punkte")
	@Localize(locale = "ru", value = "Очки")
	private static String		EDITOR_POINTS;
	@Localize("The number of points spent on this spell")
	@Localize(locale = "de", value = "Die Punkte, die für diesen Zauber aufgewendet wurden")
	@Localize(locale = "ru", value = "Потрачено на заклинание количество очков")
	private static String		EDITOR_POINTS_TOOLTIP;
	@Localize("Level")
	@Localize(locale = "de", value = "Fertigkeitswert")
	@Localize(locale = "ru", value = "Уровень")
	private static String		EDITOR_LEVEL;
	@Localize("The spell level and relative spell level to roll against")
	@Localize(locale = "de", value = "Der Fertigkeitswert und relativer Fertigkeitswert des Zaubers, gegen die gewürfelt werden muss")
	@Localize(locale = "ru", value = "Уровень заклинания и относительный уровень заклинания для повторного броска")
	private static String		EDITOR_LEVEL_TOOLTIP;
	@Localize("Difficulty")
	@Localize(locale = "de", value = "Schwierigkeit")
	@Localize(locale = "ru", value = "Сложность")
	private static String		DIFFICULTY;
	@Localize("The difficulty of the spell")
	@Localize(locale = "de", value = "Die Schwierigkeit des Zaubers")
	@Localize(locale = "ru", value = "Сложность заклинания")
	private static String		DIFFICULTY_TOOLTIP;
	@Localize("Page Reference")
	@Localize(locale = "de", value = "Seitenangabe")
	@Localize(locale = "ru", value = "Ссылка на страницу")
	private static String		EDITOR_REFERENCE;
	@Localize("A reference to the book and page this spell appears\non (e.g. B22 would refer to \"Basic Set\", page 22)")
	@Localize(locale = "de", value = "Eine Referenz auf das Buch und die Seite, auf der dieser Zauber beschrieben wird (z.B. B22 würde auf \"Basic Set\" Seite 22 verweisen)")
	@Localize(locale = "ru", value = "Ссылка на страницу и книгу, описывающая заклинание\n (например B22 - книга \"Базовые правила\", страница 22)")
	private static String		REFERENCE_TOOLTIP;

	static {
		Localization.initialize();
	}

	private JTextField			mNameField;
	private JTextField			mCollegeField;
	private JTextField			mPowerSourceField;
	private JTextField			mClassField;
	private JTextField			mCastingCostField;
	private JTextField			mMaintenanceField;
	private JTextField			mCastingTimeField;
	private JTextField			mDurationField;
	private JComboBox<Object>	mDifficultyCombo;
	private JTextField			mNotesField;
	private JTextField			mCategoriesField;
	private JTextField			mPointsField;
	private JTextField			mLevelField;
	private JTextField			mReferenceField;
	private JTabbedPane			mTabPanel;
	private PrereqsPanel		mPrereqs;
	private JCheckBox			mHasTechLevel;
	private JTextField			mTechLevel;
	private String				mSavedTechLevel;
	private MeleeWeaponEditor	mMeleeWeapons;
	private RangedWeaponEditor	mRangedWeapons;

	/**
	 * Creates a new {@link Spell} editor.
	 *
	 * @param spell The {@link Spell} to edit.
	 */
	public SpellEditor(Spell spell) {
		super(spell);

		boolean notContainer = !spell.canHaveChildren();
		Container content = new JPanel(new ColumnLayout(2));
		Container fields = new JPanel(new ColumnLayout());
		Container wrapper1 = new JPanel(new ColumnLayout(notContainer ? 3 : 2));
		Container wrapper2 = new JPanel(new ColumnLayout(4));
		Container wrapper3 = new JPanel(new ColumnLayout(2));
		Container noGapWrapper = new JPanel(new ColumnLayout(2, 0, 0));
		Container ptsPanel = null;
		JLabel icon = new JLabel(spell.getIcon(true));
		Dimension size = new Dimension();
		Container refParent = wrapper3;

		mNameField = createCorrectableField(wrapper1, wrapper1, NAME, spell.getName(), NAME_TOOLTIP);
		fields.add(wrapper1);
		if (notContainer) {
			createTechLevelFields(wrapper1);
			mCollegeField = createField(wrapper2, wrapper2, COLLEGE, spell.getCollege(), COLLEGE_TOOLTIP, 0);
			mPowerSourceField = createField(wrapper2, wrapper2, POWER_SOURCE, spell.getPowerSource(), POWER_SOURCE_TOOLTIP, 0);
			mClassField = createCorrectableField(wrapper2, wrapper2, CLASS, spell.getSpellClass(), CLASS_ONLY_TOOLTIP);
			mCastingCostField = createCorrectableField(wrapper2, wrapper2, CASTING_COST, spell.getCastingCost(), CASTING_COST_TOOLTIP);
			mMaintenanceField = createField(wrapper2, wrapper2, MAINTENANCE_COST, spell.getMaintenance(), MAINTENANCE_COST_TOOLTIP, 0);
			mCastingTimeField = createCorrectableField(wrapper2, wrapper2, CASTING_TIME, spell.getCastingTime(), CASTING_TIME_TOOLTIP);
			mDurationField = createCorrectableField(wrapper2, wrapper2, DURATION, spell.getDuration(), DURATION_TOOLTIP);
			fields.add(wrapper2);

			ptsPanel = createPointsFields();
			fields.add(ptsPanel);
			refParent = ptsPanel;
		}
		mNotesField = createField(wrapper3, wrapper3, NOTES, spell.getNotes(), NOTES_TOOLTIP, 0);
		mCategoriesField = createField(wrapper3, wrapper3, CATEGORIES, spell.getCategoriesAsString(), CATEGORIES_TOOLTIP, 0);
		mReferenceField = createField(refParent, noGapWrapper, EDITOR_REFERENCE, mRow.getReference(), REFERENCE_TOOLTIP, 6);
		noGapWrapper.add(new JPanel());
		refParent.add(noGapWrapper);
		fields.add(wrapper3);

		determineLargest(wrapper1, 2, size);
		determineLargest(wrapper2, 4, size);
		if (ptsPanel != null) {
			determineLargest(ptsPanel, 100, size);
		}
		determineLargest(wrapper3, 2, size);
		applySize(wrapper1, 2, size);
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

	@SuppressWarnings("unused")
	private Container createPointsFields() {
		boolean forCharacter = mRow.getCharacter() != null;
		boolean forTemplate = mRow.getTemplate() != null;
		JPanel panel = new JPanel(new ColumnLayout(forCharacter ? 8 : forTemplate ? 6 : 4));

		mDifficultyCombo = new JComboBox<>(new Object[] { SkillDifficulty.H.toString(), SkillDifficulty.VH.toString() });
		mDifficultyCombo.setSelectedIndex(mRow.isVeryHard() ? 1 : 0);
		mDifficultyCombo.setToolTipText(DIFFICULTY_TOOLTIP);
		UIUtilities.setOnlySize(mDifficultyCombo, mDifficultyCombo.getPreferredSize());
		mDifficultyCombo.addActionListener(this);
		mDifficultyCombo.setEnabled(mIsEditable);
		panel.add(new LinkedLabel(DIFFICULTY, mDifficultyCombo));
		panel.add(mDifficultyCombo);

		if (forCharacter || mRow.getTemplate() != null) {
			mPointsField = createField(panel, panel, EDITOR_POINTS, Integer.toString(mRow.getPoints()), EDITOR_POINTS_TOOLTIP, 4);
			new NumberFilter(mPointsField, false, false, false, 4);
			mPointsField.addActionListener(this);

			if (forCharacter) {
				mLevelField = createField(panel, panel, EDITOR_LEVEL, getDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel()), EDITOR_LEVEL_TOOLTIP, 5);
				mLevelField.setEnabled(false);
			}
		}
		return panel;
	}

	private static String getDisplayLevel(int level, int relativeLevel) {
		if (level < 0) {
			return "-"; //$NON-NLS-1$
		}
		return Numbers.format(level) + "/" + SkillAttribute.IQ + Numbers.formatWithForcedSign(relativeLevel); //$NON-NLS-1$
	}

	private JTextField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
		JTextField field = new JTextField(text);
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.getDocument().addDocumentListener(this);

		LinkedLabel label = new LinkedLabel(title);
		label.setLink(field);

		labelParent.add(label);
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
		labelParent.add(new LinkedLabel(title, field));
		fieldParent.add(field);
		return field;
	}

	@Override
	public boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());
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
			modified |= mRow.setIsVeryHard(isVeryHard());
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
				mTechLevel.setText(""); //$NON-NLS-1$
			}
		} else if (src == mPointsField || src == mDifficultyCombo) {
			recalculateLevel();
		}
	}

	private void recalculateLevel() {
		if (mLevelField != null) {
			SkillLevel level = Spell.calculateLevel(mRow.getCharacter(), getSpellPoints(), isVeryHard(), mCollegeField.getText(), mPowerSourceField.getText(), mNameField.getText());
			mLevelField.setText(getDisplayLevel(level.mLevel, level.mRelativeLevel));
		}
	}

	private int getSpellPoints() {
		return Numbers.getLocalizedInteger(mPointsField.getText(), 0);
	}

	private boolean isVeryHard() {
		return mDifficultyCombo.getSelectedIndex() == 1;
	}

	@Override
	public void changedUpdate(DocumentEvent event) {
		Document doc = event.getDocument();
		if (doc == mNameField.getDocument()) {
			LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().length() != 0 ? null : NAME_CANNOT_BE_EMPTY);
		} else if (doc == mClassField.getDocument()) {
			LinkedLabel.setErrorMessage(mClassField, mClassField.getText().trim().length() != 0 ? null : CLASS_CANNOT_BE_EMPTY);
		} else if (doc == mClassField.getDocument()) {
			LinkedLabel.setErrorMessage(mCastingCostField, mCastingCostField.getText().trim().length() != 0 ? null : CASTING_COST_CANNOT_BE_EMPTY);
		} else if (doc == mClassField.getDocument()) {
			LinkedLabel.setErrorMessage(mCastingTimeField, mCastingTimeField.getText().trim().length() != 0 ? null : CASTING_TIME_CANNOT_BE_EMPTY);
		} else if (doc == mClassField.getDocument()) {
			LinkedLabel.setErrorMessage(mDurationField, mDurationField.getText().trim().length() != 0 ? null : DURATION_CANNOT_BE_EMPTY);
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
