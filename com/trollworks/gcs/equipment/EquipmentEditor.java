/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.text.NumberFilter;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.text.TextUtility;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.LinkedLabel;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;

import javax.swing.ImageIcon;
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
public class EquipmentEditor extends RowEditor<Equipment> implements ActionListener, DocumentListener {
	private static String		MSG_REFERENCE_TOOLTIP;
	private static String		MSG_VALUE_TOOLTIP;
	private static String		MSG_EXT_VALUE_TOOLTIP;
	private static String		MSG_NAME;
	private static String		MSG_NAME_TOOLTIP;
	private static String		MSG_NAME_CANNOT_BE_EMPTY;
	private static String		MSG_EDITOR_TECH_LEVEL;
	private static String		MSG_EDITOR_TECH_LEVEL_TOOLTIP;
	private static String		MSG_EDITOR_LEGALITY_CLASS;
	private static String		MSG_EDITOR_LEGALITY_CLASS_TOOLTIP;
	private static String		MSG_EDITOR_QUANTITY;
	private static String		MSG_EDITOR_QUANTITY_TOOLTIP;
	private static String		MSG_EDITOR_VALUE;
	private static String		MSG_EDITOR_EXTENDED_VALUE;
	private static String		MSG_EDITOR_WEIGHT;
	private static String		MSG_EDITOR_WEIGHT_TOOLTIP;
	private static String		MSG_EDITOR_EXTENDED_WEIGHT;
	private static String		MSG_EDITOR_EXTENDED_WEIGHT_TOOLTIP;
	private static String		MSG_CATEGORIES;
	private static String		MSG_CATEGORIES_TOOLTIP;
	private static String		MSG_NOTES;
	private static String		MSG_NOTES_TOOLTIP;
	private static String		MSG_EDITOR_REFERENCE;
	private static String		MSG_POUNDS;
	private static String		MSG_STATE_TOOLTIP;
	private JComboBox			mStateCombo;
	private JTextField			mDescriptionField;
	private JTextField			mTechLevelField;
	private JTextField			mLegalityClassField;
	private JTextField			mQtyField;
	private JTextField			mValueField;
	private JTextField			mExtValueField;
	private JTextField			mWeightField;
	private JTextField			mExtWeightField;
	private JTextField			mNotesField;
	private JTextField			mCategoriesField;
	private JTextField			mReferenceField;
	private JTabbedPane			mTabPanel;
	private PrereqsPanel		mPrereqs;
	private FeaturesPanel		mFeatures;
	private MeleeWeaponEditor	mMeleeWeapons;
	private RangedWeaponEditor	mRangedWeapons;
	private double				mContainedValue;
	private double				mContainedWeight;

	static {
		LocalizedMessages.initialize(EquipmentEditor.class);
	}

	/**
	 * Creates a new {@link Equipment} editor.
	 * 
	 * @param equipment The {@link Equipment} to edit.
	 */
	public EquipmentEditor(Equipment equipment) {
		super(equipment);

		JPanel content = new JPanel(new ColumnLayout(2));
		JPanel fields = new JPanel(new ColumnLayout(2));
		JLabel icon = new JLabel(new ImageIcon(equipment.getImage(true)));
		JPanel wrapper = new JPanel(new ColumnLayout(2));

		mDescriptionField = createCorrectableField(fields, MSG_NAME, equipment.getDescription(), MSG_NAME_TOOLTIP);
		createSecondLineFields(fields);
		createValueAndWeightFields(fields);
		mNotesField = createField(fields, fields, MSG_NOTES, equipment.getNotes(), MSG_NOTES_TOOLTIP, 0);
		mCategoriesField = createField(fields, fields, MSG_CATEGORIES, equipment.getCategoriesAsString(), MSG_CATEGORIES_TOOLTIP, 0);
		mReferenceField = createField(fields, wrapper, MSG_EDITOR_REFERENCE, mRow.getReference(), MSG_REFERENCE_TOOLTIP, 6);
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
			mQtyField = createIntegerNumberField(parent, wrapper, MSG_EDITOR_QUANTITY, mRow.getQuantity(), MSG_EDITOR_QUANTITY_TOOLTIP, 9);
		}
		mTechLevelField = createField(isContainer ? parent : wrapper, wrapper, MSG_EDITOR_TECH_LEVEL, mRow.getTechLevel(), MSG_EDITOR_TECH_LEVEL_TOOLTIP, 3);
		mLegalityClassField = createField(wrapper, wrapper, MSG_EDITOR_LEGALITY_CLASS, mRow.getLegalityClass(), MSG_EDITOR_LEGALITY_CLASS_TOOLTIP, 3);
		if (showEquipmentState()) {
			mStateCombo = new JComboBox(EquipmentState.values());
			mStateCombo.setSelectedItem(mRow.getState());
			UIUtilities.setOnlySize(mStateCombo, mStateCombo.getPreferredSize());
			mStateCombo.setEnabled(mIsEditable);
			mStateCombo.setToolTipText(MSG_STATE_TOOLTIP);
			wrapper.add(mStateCombo);
		}
		wrapper.add(new JPanel());
		parent.add(wrapper);
	}

	private void createValueAndWeightFields(Container parent) {
		JPanel wrapper = new JPanel(new ColumnLayout(4));
		Component first;

		mContainedValue = mRow.getExtendedValue() - mRow.getValue() * mRow.getQuantity();
		mValueField = createNumberField(parent, wrapper, MSG_EDITOR_VALUE, mRow.getValue(), MSG_VALUE_TOOLTIP, 13);
		mExtValueField = createNumberField(wrapper, wrapper, MSG_EDITOR_EXTENDED_VALUE, mRow.getExtendedValue(), MSG_EXT_VALUE_TOOLTIP, 13);
		first = wrapper.getComponent(1);
		mExtValueField.setEnabled(false);
		wrapper.add(new JPanel());
		parent.add(wrapper);

		wrapper = new JPanel(new ColumnLayout(5));
		mContainedWeight = mRow.getExtendedWeight() - mRow.getWeight() * mRow.getQuantity();
		mWeightField = createNumberField(parent, wrapper, MSG_EDITOR_WEIGHT, mRow.getWeight(), MSG_EDITOR_WEIGHT_TOOLTIP, 13);
		mExtWeightField = createNumberField(wrapper, wrapper, MSG_EDITOR_EXTENDED_WEIGHT, mRow.getExtendedWeight(), MSG_EDITOR_EXTENDED_WEIGHT_TOOLTIP, 13);
		mExtWeightField.setEnabled(false);
		UIUtilities.adjustToSameSize(new Component[] { first, wrapper.getComponent(1) });
		wrapper.add(new JLabel(MSG_POUNDS));
		wrapper.add(new JPanel());
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

	private JTextField createIntegerNumberField(Container labelParent, Container fieldParent, String title, int value, String tooltip, int maxDigits) {
		JTextField field = new JTextField(TextUtility.makeFiller(maxDigits, '9') + TextUtility.makeFiller(maxDigits / 3, ','));

		UIUtilities.setOnlySize(field, field.getPreferredSize());
		field.setText(Numbers.format(value));
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		new NumberFilter(field, false, false, true, maxDigits);
		field.addActionListener(this);
		labelParent.add(new LinkedLabel(title, field));
		fieldParent.add(field);
		return field;
	}

	private JTextField createNumberField(Container labelParent, Container fieldParent, String title, double value, String tooltip, int maxDigits) {
		JTextField field = new JTextField(TextUtility.makeFiller(maxDigits, '9') + TextUtility.makeFiller(maxDigits / 3, ',') + "."); //$NON-NLS-1$

		UIUtilities.setOnlySize(field, field.getPreferredSize());
		field.setText(Numbers.format(value));
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		new NumberFilter(field, true, false, true, maxDigits);
		field.addActionListener(this);
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
		modified |= mRow.setWeight(Numbers.getLocalizedDouble(mWeightField.getText(), 0.0));
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
			ArrayList<WeaponStats> list = new ArrayList<WeaponStats>(mMeleeWeapons.getWeapons());

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

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		if (src == mValueField) {
			valueChanged();
		} else if (src == mWeightField) {
			weightChanged();
		} else if (src == mQtyField) {
			valueChanged();
			weightChanged();
		}
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
		int qty = getQty();
		double value;

		if (qty < 1) {
			value = 0;
		} else {
			value = qty * Numbers.getLocalizedDouble(mWeightField.getText(), 0.0) + mContainedWeight;
		}
		mExtWeightField.setText(Numbers.format(value));
	}

	public void changedUpdate(DocumentEvent event) {
		descriptionChanged();
	}

	public void insertUpdate(DocumentEvent event) {
		descriptionChanged();
	}

	public void removeUpdate(DocumentEvent event) {
		descriptionChanged();
	}

	private void descriptionChanged() {
		LinkedLabel.setErrorMessage(mDescriptionField, mDescriptionField.getText().trim().length() != 0 ? null : MSG_NAME_CANNOT_BE_EMPTY);
	}
}
