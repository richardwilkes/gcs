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

package com.trollworks.gcs.equipment;

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
import com.trollworks.toolkit.utility.units.WeightValue;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.util.ArrayList;

import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTabbedPane;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** The detailed editor for {@link Equipment}s. */
public class EquipmentEditor extends RowEditor<Equipment> implements ActionListener, DocumentListener, FocusListener {
	@Localize("A reference to the book and page this equipment appears\non (e.g. B22 would refer to \"Basic Set\", page 22)")
	@Localize(locale = "de", value = "Eine Referenz auf das Buch und die Seite, auf der diese Ausrüstung beschrieben wird (z.B. B22 würde auf \"Basic Set\" Seite 22 verweisen).")
	@Localize(locale = "ru", value = "Ссылка на страницу и книгу, описывающая снаряжение\n (например B22 - книга \"Базовые правила\", страница 22)")
	private static String				REFERENCE_TOOLTIP;
	@Localize("The value of one of these pieces of equipment")
	@Localize(locale = "de", value = "Der Wert eines einzelnen Ausrüstungsgegenstandes")
	@Localize(locale = "ru", value = "Цена снаряжения")
	private static String				VALUE_TOOLTIP;
	@Localize("The value of all of these pieces of equipment,\nplus the value of any contained equipment")
	@Localize(locale = "de", value = "Der Wert aller dieser Ausrüstungsgegenstände\nund der Wert der darin enthaltenen Ausrüstung")
	@Localize(locale = "ru", value = "Цена всего снаряжения,\nплюс цена любого входящего в него снаряжения")
	private static String				EXT_VALUE_TOOLTIP;
	@Localize("Name")
	@Localize(locale = "de", value = "Name")
	@Localize(locale = "ru", value = "Название")
	private static String				NAME;
	@Localize("The name/description of the equipment, without any notes")
	@Localize(locale = "de", value = "Der Name des Ausrüstungsgegenstands, ohne Anmerkungen")
	@Localize(locale = "ru", value = "Название/описание снаряжения без заметок")
	private static String				NAME_TOOLTIP;
	@Localize("The name field may not be empty")
	@Localize(locale = "de", value = "Der Name darf nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Название\" не может быть пустым")
	private static String				NAME_CANNOT_BE_EMPTY;
	@Localize("Tech Level")
	@Localize(locale = "de", value = "Techlevel")
	@Localize(locale = "ru", value = "Технологический уровень")
	private static String				EDITOR_TECH_LEVEL;
	@Localize("The first Tech Level this equipment is available at")
	@Localize(locale = "de", value = "Der Techlevel, ab dem diese Ausrüstung zur Verfügung steht")
	@Localize(locale = "ru", value = "Первый тех. уровень этого снаряжения доступен с")
	private static String				EDITOR_TECH_LEVEL_TOOLTIP;
	@Localize("Legality Class")
	@Localize(locale = "de", value = "Legalitätsklasse")
	@Localize(locale = "ru", value = "Клас легальности")
	private static String				EDITOR_LEGALITY_CLASS;
	@Localize("The legality class of this equipment")
	@Localize(locale = "de", value = "Die Legalitätsklasse des Ausrüstungsgegenstandes")
	@Localize(locale = "ru", value = "Класс легальности снаряжения")
	private static String				EDITOR_LEGALITY_CLASS_TOOLTIP;
	@Localize("Quantity")
	@Localize(locale = "de", value = "Anzahl")
	@Localize(locale = "ru", value = "Количество")
	private static String				EDITOR_QUANTITY;
	@Localize("The number of this equipment present")
	@Localize(locale = "de", value = "")
	@Localize(locale = "ru", value = "Количество этого снаряжения")
	private static String				EDITOR_QUANTITY_TOOLTIP;
	@Localize("Value")
	@Localize(locale = "de", value = "Wert")
	@Localize(locale = "ru", value = "Цена")
	private static String				EDITOR_VALUE;
	@Localize("Extended Value")
	@Localize(locale = "de", value = "Gesamtwert")
	@Localize(locale = "ru", value = "Полная цена")
	private static String				EDITOR_EXTENDED_VALUE;
	@Localize("Weight")
	@Localize(locale = "de", value = "Gewicht")
	@Localize(locale = "ru", value = "Вес")
	private static String				EDITOR_WEIGHT;
	@Localize("The weight of one of these pieces of equipment")
	@Localize(locale = "de", value = "Das Gewicht eines einzelnen Ausrüstungsgegenstandes")
	@Localize(locale = "ru", value = "Вес снаряжения")
	private static String				EDITOR_WEIGHT_TOOLTIP;
	@Localize("Extended Weight")
	@Localize(locale = "de", value = "Gesamtgewicht")
	@Localize(locale = "ru", value = "Полный вес")
	private static String				EDITOR_EXTENDED_WEIGHT;
	@Localize("The total weight of this quantity of equipment, plus everything contained by it")
	@Localize(locale = "de", value = "Das Gewicht aller dieser Ausrüstungsgegenstände\nund das Gewicht der darin enthaltenen Ausrüstung")
	@Localize(locale = "ru", value = "Общий вес имеющегося снаряжения и его содержимого")
	private static String				EDITOR_EXTENDED_WEIGHT_TOOLTIP;
	@Localize("Categories")
	@Localize(locale = "de", value = "Kategorien")
	@Localize(locale = "ru", value = "Категории")
	private static String				CATEGORIES;
	@Localize("The category or categories the equipment belongs to (separate multiple categories with a comma)")
	@Localize(locale = "de", value = "Die Kategorie oder Kategorien, denen diese Ausrüstung angehört (trenne mehrere Kategorien mit einem Komma)")
	@Localize(locale = "ru", value = "Категория или категории снаряжения, к которым оно принадлежит (несколько категорий разделяются точкой с запятой)")
	private static String				CATEGORIES_TOOLTIP;
	@Localize("Notes")
	@Localize(locale = "de", value = "Anmerkungen")
	@Localize(locale = "ru", value = "Заметка")
	private static String				NOTES;
	@Localize("Any notes that you would like to show up in the list along with this equipment")
	@Localize(locale = "de", value = "Anmerkungen, die in der Liste neben der Ausrüstung erscheinen sollen")
	@Localize(locale = "ru", value = "Заметки, которые показываются в списке рядом с снаряжением")
	private static String				NOTES_TOOLTIP;
	@Localize("Page Reference")
	@Localize(locale = "de", value = "Seitenangabe")
	@Localize(locale = "ru", value = "Ссылка на страницу")
	private static String				EDITOR_REFERENCE;
	@Localize("Items that are not equipped do not apply any features they may\nnormally contribute to the character.")
	@Localize(locale = "de", value = "Gegenstände, die nicht ausgerüstet sind, haben keine Auswirkungen auf den Charakter.")
	@Localize(locale = "ru", value = "Не экипированные предметы не добавляют свойств, которые обычно\n может использовать персонаж.")
	private static String				STATE_TOOLTIP;

	static {
		Localization.initialize();
	}

	private JComboBox<EquipmentState>	mStateCombo;
	private JTextField					mDescriptionField;
	private JTextField					mTechLevelField;
	private JTextField					mLegalityClassField;
	private JTextField					mQtyField;
	private JTextField					mValueField;
	private JTextField					mExtValueField;
	private JTextField					mWeightField;
	private JTextField					mExtWeightField;
	private JTextField					mNotesField;
	private JTextField					mCategoriesField;
	private JTextField					mReferenceField;
	private JTabbedPane					mTabPanel;
	private PrereqsPanel				mPrereqs;
	private FeaturesPanel				mFeatures;
	private MeleeWeaponEditor			mMeleeWeapons;
	private RangedWeaponEditor			mRangedWeapons;
	private double						mContainedValue;
	private WeightValue					mContainedWeight;

	/**
	 * Creates a new {@link Equipment} editor.
	 *
	 * @param equipment The {@link Equipment} to edit.
	 */
	public EquipmentEditor(Equipment equipment) {
		super(equipment);

		JPanel content = new JPanel(new ColumnLayout(2));
		JPanel fields = new JPanel(new ColumnLayout(2));
		JLabel icon = new JLabel(equipment.getIcon(true));
		JPanel wrapper = new JPanel(new ColumnLayout(2));

		mDescriptionField = createCorrectableField(fields, NAME, equipment.getDescription(), NAME_TOOLTIP);
		createSecondLineFields(fields);
		createValueAndWeightFields(fields);
		mNotesField = createField(fields, fields, NOTES, equipment.getNotes(), NOTES_TOOLTIP, 0);
		mCategoriesField = createField(fields, fields, CATEGORIES, equipment.getCategoriesAsString(), CATEGORIES_TOOLTIP, 0);
		mReferenceField = createField(fields, wrapper, EDITOR_REFERENCE, mRow.getReference(), REFERENCE_TOOLTIP, 6);
		wrapper.add(new JPanel());
		fields.add(wrapper);
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
		mTabPanel.addTab(mMeleeWeapons.getName(), mMeleeWeapons);
		mTabPanel.addTab(mRangedWeapons.getName(), mRangedWeapons);
		Component panel = embedEditor(mPrereqs);
		mTabPanel.addTab(panel.getName(), panel);
		panel = embedEditor(mFeatures);
		mTabPanel.addTab(panel.getName(), panel);
		if (!mIsEditable) {
			UIUtilities.disableControls(mMeleeWeapons);
			UIUtilities.disableControls(mRangedWeapons);
		}
		UIUtilities.selectTab(mTabPanel, getLastTabName());
		add(mTabPanel);
	}

	private boolean showEquipmentState() {
		return mRow.getCharacter() != null;
	}

	private void createSecondLineFields(Container parent) {
		boolean isContainer = mRow.canHaveChildren();
		JPanel wrapper = new JPanel(new ColumnLayout((isContainer ? 4 : 6) + (showEquipmentState() ? 1 : 0)));

		if (!isContainer) {
			mQtyField = createIntegerNumberField(parent, wrapper, EDITOR_QUANTITY, mRow.getQuantity(), EDITOR_QUANTITY_TOOLTIP, 9);
		}
		mTechLevelField = createField(isContainer ? parent : wrapper, wrapper, EDITOR_TECH_LEVEL, mRow.getTechLevel(), EDITOR_TECH_LEVEL_TOOLTIP, 3);
		mLegalityClassField = createField(wrapper, wrapper, EDITOR_LEGALITY_CLASS, mRow.getLegalityClass(), EDITOR_LEGALITY_CLASS_TOOLTIP, 3);
		if (showEquipmentState()) {
			mStateCombo = new JComboBox<>(EquipmentState.values());
			mStateCombo.setSelectedItem(mRow.getState());
			UIUtilities.setOnlySize(mStateCombo, mStateCombo.getPreferredSize());
			mStateCombo.setEnabled(mIsEditable);
			mStateCombo.setToolTipText(STATE_TOOLTIP);
			wrapper.add(mStateCombo);
		}
		wrapper.add(new JPanel());
		parent.add(wrapper);
	}

	private void createValueAndWeightFields(Container parent) {
		JPanel wrapper = new JPanel(new ColumnLayout(4));
		Component first;

		mContainedValue = mRow.getExtendedValue() - mRow.getValue() * mRow.getQuantity();
		mValueField = createNumberField(parent, wrapper, EDITOR_VALUE, mRow.getValue(), VALUE_TOOLTIP, 13);
		mExtValueField = createNumberField(wrapper, wrapper, EDITOR_EXTENDED_VALUE, mRow.getExtendedValue(), EXT_VALUE_TOOLTIP, 13);
		first = wrapper.getComponent(1);
		mExtValueField.setEnabled(false);
		wrapper.add(new JPanel());
		parent.add(wrapper);

		wrapper = new JPanel(new ColumnLayout(3));
		mContainedWeight = new WeightValue(mRow.getExtendedWeight());
		WeightValue weight = new WeightValue(mRow.getWeight());
		weight.setValue(weight.getValue() * mRow.getQuantity());
		mContainedWeight.subtract(weight);
		mWeightField = createWeightField(parent, wrapper, EDITOR_WEIGHT, mRow.getWeight(), EDITOR_WEIGHT_TOOLTIP, 13);
		mExtWeightField = createWeightField(wrapper, wrapper, EDITOR_EXTENDED_WEIGHT, mRow.getExtendedWeight(), EDITOR_EXTENDED_WEIGHT_TOOLTIP, 13);
		mExtWeightField.setEnabled(false);
		UIUtilities.adjustToSameSize(new Component[] { first, wrapper.getComponent(1) });
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

	private JTextField createCorrectableField(Container parent, String title, String text, String tooltip) {
		JTextField field = new JTextField(text);
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.getDocument().addDocumentListener(this);
		field.addFocusListener(this);

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
		field.addFocusListener(this);
		labelParent.add(new LinkedLabel(title, field));
		fieldParent.add(field);
		return field;
	}

	@SuppressWarnings("unused")
	private JTextField createIntegerNumberField(Container labelParent, Container fieldParent, String title, int value, String tooltip, int maxDigits) {
		JTextField field = new JTextField(TextUtility.makeFiller(maxDigits, '9') + TextUtility.makeFiller(maxDigits / 3, ','));
		UIUtilities.setOnlySize(field, field.getPreferredSize());
		field.setText(Numbers.format(value));
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		new NumberFilter(field, false, false, true, maxDigits);
		field.addActionListener(this);
		field.addFocusListener(this);
		labelParent.add(new LinkedLabel(title, field));
		fieldParent.add(field);
		return field;
	}

	@SuppressWarnings("unused")
	private JTextField createNumberField(Container labelParent, Container fieldParent, String title, double value, String tooltip, int maxDigits) {
		JTextField field = new JTextField(TextUtility.makeFiller(maxDigits, '9') + TextUtility.makeFiller(maxDigits / 3, ',') + "."); //$NON-NLS-1$
		UIUtilities.setOnlySize(field, field.getPreferredSize());
		field.setText(Numbers.format(value));
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		new NumberFilter(field, true, false, true, maxDigits);
		field.addActionListener(this);
		field.addFocusListener(this);
		labelParent.add(new LinkedLabel(title, field));
		fieldParent.add(field);
		return field;
	}

	private JTextField createWeightField(Container labelParent, Container fieldParent, String title, WeightValue value, String tooltip, int maxDigits) {
		JTextField field = new JTextField(TextUtility.makeFiller(maxDigits, '9') + TextUtility.makeFiller(maxDigits / 3, ',') + "."); //$NON-NLS-1$
		UIUtilities.setOnlySize(field, field.getPreferredSize());
		field.setText(value.toString());
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.addActionListener(this);
		field.addFocusListener(this);
		labelParent.add(new LinkedLabel(title, field));
		fieldParent.add(field);
		return field;
	}

	@Override
	public boolean applyChangesSelf() {
		boolean modified = mRow.setDescription(mDescriptionField.getText());
		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setTechLevel(mTechLevelField.getText());
		modified |= mRow.setLegalityClass(mLegalityClassField.getText());
		modified |= mRow.setQuantity(getQty());
		modified |= mRow.setValue(Numbers.getLocalizedDouble(mValueField.getText(), 0.0));
		modified |= mRow.setWeight(WeightValue.extract(mWeightField.getText(), true));
		if (showEquipmentState()) {
			modified |= mRow.setState((EquipmentState) mStateCombo.getSelectedItem());
		}
		modified |= mRow.setNotes(mNotesField.getText());
		modified |= mRow.setCategories(mCategoriesField.getText());
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
		adjustForChange(event.getSource());
	}

	private int getQty() {
		if (mQtyField != null) {
			return Numbers.getLocalizedInteger(mQtyField.getText(), 0);
		}
		return 1;
	}

	private void valueChanged() {
		int qty = getQty();
		double value;

		if (qty < 1) {
			value = 0;
		} else {
			value = qty * Numbers.getLocalizedDouble(mValueField.getText(), 0.0) + mContainedValue;
		}
		mExtValueField.setText(Numbers.format(value));
	}

	private void weightChanged() {
		WeightValue weight = WeightValue.extract(mWeightField.getText(), true);
		weight.setValue(weight.getValue() * Math.max(getQty(), 0));
		mExtWeightField.setText(weight.toString());
	}

	@Override
	public void changedUpdate(DocumentEvent event) {
		descriptionChanged();
	}

	@Override
	public void insertUpdate(DocumentEvent event) {
		descriptionChanged();
	}

	@Override
	public void removeUpdate(DocumentEvent event) {
		descriptionChanged();
	}

	private void descriptionChanged() {
		LinkedLabel.setErrorMessage(mDescriptionField, mDescriptionField.getText().trim().length() != 0 ? null : NAME_CANNOT_BE_EMPTY);
	}

	@Override
	public void focusGained(FocusEvent event) {
		// Not used.
	}

	@Override
	public void focusLost(FocusEvent event) {
		adjustForChange(event.getSource());
	}

	private void adjustForChange(Object field) {
		if (field == mValueField) {
			valueChanged();
		} else if (field == mWeightField) {
			weightChanged();
		} else if (field == mQtyField) {
			valueChanged();
			weightChanged();
		}
	}
}
