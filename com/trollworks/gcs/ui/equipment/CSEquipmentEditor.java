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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.equipment;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.editor.feature.CSFeatures;
import com.trollworks.gcs.ui.editor.prereq.CSPrereqs;
import com.trollworks.gcs.ui.weapon.CSMeleeWeaponEditor;
import com.trollworks.gcs.ui.weapon.CSRangedWeaponEditor;
import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKCorrectableField;
import com.trollworks.toolkit.widget.TKCorrectableLabel;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKLinkedLabel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKBaseButton;
import com.trollworks.toolkit.widget.button.TKCheckbox;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.tab.TKTabbedPanel;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;

/** The detailed editor for {@link CMEquipment}s. */
public class CSEquipmentEditor extends CSRowEditor<CMEquipment> implements ActionListener {
	private TKCorrectableField		mDescriptionField;
	private TKTextField				mTechLevelField;
	private TKTextField				mLegalityClassField;
	private TKTextField				mQtyField;
	private TKCheckbox				mEquipped;
	private TKTextField				mValueField;
	private TKTextField				mExtValueField;
	private TKTextField				mWeightField;
	private TKTextField				mExtWeightField;
	private TKTextField				mNotesField;
	private TKTextField				mReferenceField;
	private TKTabbedPanel			mTabPanel;
	private CSPrereqs				mPrereqs;
	private CSFeatures				mFeatures;
	private CSMeleeWeaponEditor		mMeleeWeapons;
	private CSRangedWeaponEditor	mRangedWeapons;
	private double					mContainedValue;
	private double					mContainedWeight;

	/**
	 * Creates a new {@link CMEquipment} editor.
	 * 
	 * @param equipment The {@link CMEquipment} to edit.
	 */
	public CSEquipmentEditor(CMEquipment equipment) {
		super(equipment);

		TKPanel content = new TKPanel(new TKColumnLayout(2));
		TKPanel fields = new TKPanel(new TKColumnLayout(2));
		TKLabel icon = new TKLabel(equipment.getImage(true));
		ArrayList<TKPanel> panels = new ArrayList<TKPanel>();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(2));

		mDescriptionField = createCorrectableField(fields, Msgs.NAME, equipment.getDescription(), Msgs.NAME_TOOLTIP);
		createSecondLineFields(fields);
		createValueAndWeightFields(fields);
		mNotesField = createField(fields, fields, Msgs.NOTES, equipment.getNotes(), Msgs.NOTES_TOOLTIP, 0);
		mReferenceField = createField(fields, wrapper, Msgs.EDITOR_REFERENCE, mRow.getReference(), Msgs.REFERENCE_TOOLTIP, 6);
		wrapper.add(new TKPanel());
		fields.add(wrapper);
		icon.setVerticalAlignment(TKAlignment.TOP);
		icon.setAlignmentY(-1f);
		content.add(icon);
		content.add(fields);
		add(content);

		mPrereqs = new CSPrereqs(mRow, mRow.getPrereqs());
		mFeatures = new CSFeatures(mRow, mRow.getFeatures());
		mMeleeWeapons = CSMeleeWeaponEditor.createEditor(mRow);
		mRangedWeapons = CSRangedWeaponEditor.createEditor(mRow);
		panels.add(mMeleeWeapons);
		panels.add(mRangedWeapons);
		panels.add(embedEditor(mPrereqs));
		panels.add(embedEditor(mFeatures));
		if (!mIsEditable) {
			disableControls(mMeleeWeapons);
			disableControls(mRangedWeapons);
		}
		mTabPanel = new TKTabbedPanel(panels);
		mTabPanel.setSelectedPanelByName(getLastTabName());
		add(mTabPanel);
	}

	private boolean isCarried() {
		CMCharacter character = mRow.getCharacter();

		if (character != null) {
			return character.getCarriedEquipmentRoot() == mRow.getOwner();
		}
		return false;
	}

	private void createSecondLineFields(TKPanel parent) {
		boolean isContainer = mRow.canHaveChildren();
		boolean isCarried = isCarried();
		TKPanel wrapper = new TKPanel(new TKColumnLayout((isContainer ? 4 : 6) + (isCarried ? 1 : 0)));

		if (!isContainer) {
			mQtyField = createIntegerNumberField(parent, wrapper, Msgs.EDITOR_QUANTITY, mRow.getQuantity(), Msgs.EDITOR_QUANTITY_TOOLTIP, 9);
		}
		mTechLevelField = createField(isContainer ? parent : wrapper, wrapper, Msgs.EDITOR_TECH_LEVEL, mRow.getTechLevel(), Msgs.EDITOR_TECH_LEVEL_TOOLTIP, 3);
		mLegalityClassField = createField(wrapper, wrapper, Msgs.EDITOR_LEGALITY_CLASS, mRow.getLegalityClass(), Msgs.EDITOR_LEGALITY_CLASS_TOOLTIP, 3);
		if (isCarried) {
			mEquipped = new TKCheckbox(Msgs.EDITOR_EQUIPPED, mRow.isEquipped());
			mEquipped.setToolTipText(Msgs.EQUIPPED_TOOLTIP);
			wrapper.add(mEquipped);
		}
		wrapper.add(new TKPanel());
		parent.add(wrapper);
	}

	private void createValueAndWeightFields(TKPanel parent) {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));
		TKPanel first;

		mContainedValue = mRow.getExtendedValue() - mRow.getValue() * mRow.getQuantity();
		mValueField = createNumberField(parent, wrapper, Msgs.EDITOR_VALUE, mRow.getValue(), Msgs.VALUE_TOOLTIP, 13);
		mExtValueField = createNumberField(wrapper, wrapper, Msgs.EDITOR_EXTENDED_VALUE, mRow.getExtendedValue(), Msgs.EXT_VALUE_TOOLTIP, 13);
		first = (TKPanel) wrapper.getComponent(1);
		mExtValueField.setEnabled(false);
		wrapper.add(new TKPanel());
		parent.add(wrapper);

		wrapper = new TKPanel(new TKColumnLayout(5));
		mContainedWeight = mRow.getExtendedWeight() - mRow.getWeight() * mRow.getQuantity();
		mWeightField = createNumberField(parent, wrapper, Msgs.EDITOR_WEIGHT, mRow.getWeight(), Msgs.EDITOR_WEIGHT_TOOLTIP, 13);
		mExtWeightField = createNumberField(wrapper, wrapper, Msgs.EDITOR_EXTENDED_WEIGHT, mRow.getExtendedWeight(), Msgs.EDITOR_EXTENDED_WEIGHT_TOOLTIP, 13);
		mExtWeightField.setEnabled(false);
		adjustToSameSize(new TKPanel[] { first, (TKPanel) wrapper.getComponent(1) });
		wrapper.add(new TKLabel(Msgs.POUNDS));
		wrapper.add(new TKPanel());
		parent.add(wrapper);
	}

	private void disableControls(TKPanel panel) {
		int count = panel.getComponentCount();

		for (int i = 0; i < count; i++) {
			disableControls((TKPanel) panel.getComponent(i));
		}

		if (panel instanceof TKBaseButton || panel instanceof TKTextField || panel instanceof TKPopupMenu) {
			panel.setEnabled(false);
		}
	}

	private TKScrollPanel embedEditor(TKPanel editor) {
		TKScrollPanel scrollPanel = new TKScrollPanel(editor);

		scrollPanel.setMinimumSize(new Dimension(500, 120));
		scrollPanel.getContentBorderView().setBorder(new TKLineBorder(TKLineBorder.LEFT_EDGE | TKLineBorder.TOP_EDGE));
		scrollPanel.setName(editor.toString());
		if (!mIsEditable) {
			disableControls(editor);
		}
		return scrollPanel;
	}

	private TKCorrectableField createCorrectableField(TKPanel parent, String title, String text, String tooltip) {
		TKCorrectableLabel label = new TKCorrectableLabel(title);
		TKCorrectableField field = new TKCorrectableField(label, text);

		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.addActionListener(this);
		parent.add(label);
		parent.add(field);
		return field;
	}

	private TKTextField createField(TKPanel labelParent, TKPanel fieldParent, String title, String text, String tooltip, int maxChars) {
		TKTextField field = new TKTextField(maxChars > 0 ? TKTextUtility.makeFiller(maxChars, 'M') : text);

		if (maxChars > 0) {
			field.setOnlySize(field.getPreferredSize());
			field.setText(text);
		}
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		labelParent.add(new TKLinkedLabel(field, title));
		fieldParent.add(field);
		return field;
	}

	private TKTextField createIntegerNumberField(TKPanel labelParent, TKPanel fieldParent, String title, int value, String tooltip, int maxDigits) {
		TKTextField field = new TKTextField(TKTextUtility.makeFiller(maxDigits, '9') + TKTextUtility.makeFiller(maxDigits / 3, ','));

		field.setOnlySize(field.getPreferredSize());
		field.setText(TKNumberUtils.format(value));
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.setKeyEventFilter(new TKNumberFilter(false, false, true, maxDigits));
		field.addActionListener(this);
		labelParent.add(new TKLinkedLabel(field, title));
		fieldParent.add(field);
		return field;
	}

	private TKTextField createNumberField(TKPanel labelParent, TKPanel fieldParent, String title, double value, String tooltip, int maxDigits) {
		TKTextField field = new TKTextField(TKTextUtility.makeFiller(maxDigits, '9') + TKTextUtility.makeFiller(maxDigits / 3, ',') + "."); //$NON-NLS-1$

		field.setOnlySize(field.getPreferredSize());
		field.setText(TKNumberUtils.format(value));
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.setKeyEventFilter(new TKNumberFilter(true, false, true, maxDigits));
		field.addActionListener(this);
		labelParent.add(new TKLinkedLabel(field, title));
		fieldParent.add(field);
		return field;
	}

	@Override public boolean applyChangesSelf() {
		boolean modified = mRow.setDescription(mDescriptionField.getText());

		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setTechLevel(mTechLevelField.getText());
		modified |= mRow.setLegalityClass(mLegalityClassField.getText());
		modified |= mRow.setQuantity(getQty());
		modified |= mRow.setValue(TKNumberUtils.getDouble(mValueField.getText(), 0.0));
		modified |= mRow.setWeight(TKNumberUtils.getDouble(mWeightField.getText(), 0.0));
		if (isCarried()) {
			modified |= mRow.setEquipped(mEquipped.isChecked());
		}
		modified |= mRow.setNotes(mNotesField.getText());
		if (mPrereqs != null) {
			modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
		}
		if (mFeatures != null) {
			modified |= mRow.setFeatures(mFeatures.getFeatures());
		}
		if (mMeleeWeapons != null) {
			ArrayList<CMWeaponStats> list = new ArrayList<CMWeaponStats>(mMeleeWeapons.getWeapons());

			list.addAll(mRangedWeapons.getWeapons());
			modified |= mRow.setWeapons(list);
		}
		return modified;
	}

	@Override public void finished() {
		if (mTabPanel != null) {
			updateLastTabName(mTabPanel.getSelectedPanelName());
		}
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();

		if (src == mDescriptionField) {
			descriptionChanged();
		} else if (src == mValueField) {
			valueChanged();
		} else if (src == mWeightField) {
			weightChanged();
		} else if (src == mQtyField) {
			valueChanged();
			weightChanged();
		}
	}

	private void descriptionChanged() {
		if (mDescriptionField.getText().trim().length() == 0) {
			String corrected = mRow.getDescription().trim();

			if (corrected.length() == 0) {
				corrected = mRow.getLocalizedName();
			}
			mDescriptionField.correct(corrected, Msgs.NAME_CANNOT_BE_EMPTY);
		}
	}

	private int getQty() {
		if (mQtyField != null) {
			return TKNumberUtils.getInteger(mQtyField.getText(), 0);
		}
		return 1;
	}

	private void valueChanged() {
		int qty = getQty();
		double value;

		if (qty < 1) {
			value = 0;
		} else {
			value = qty * TKNumberUtils.getDouble(mValueField.getText(), 0.0) + mContainedValue;
		}
		mExtValueField.setText(TKNumberUtils.format(value));
	}

	private void weightChanged() {
		int qty = getQty();
		double value;

		if (qty < 1) {
			value = 0;
		} else {
			value = qty * TKNumberUtils.getDouble(mWeightField.getText(), 0.0) + mContainedWeight;
		}
		mExtWeightField.setText(TKNumberUtils.format(value));
	}
}
