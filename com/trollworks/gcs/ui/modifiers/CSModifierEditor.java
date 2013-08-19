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

package com.trollworks.gcs.ui.modifiers;

import com.trollworks.gcs.model.modifier.CMAffects;
import com.trollworks.gcs.model.modifier.CMCostType;
import com.trollworks.gcs.model.modifier.CMModifier;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.editor.feature.CSFeatures;
import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKCorrectableField;
import com.trollworks.toolkit.widget.TKCorrectableLabel;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKLinkedLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKBaseButton;
import com.trollworks.toolkit.widget.button.TKCheckbox;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.tab.TKTabbedPanel;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.text.MessageFormat;
import java.util.ArrayList;

/** Editor for {@link CMModifier}s. */
public class CSModifierEditor extends CSRowEditor<CMModifier> implements ActionListener {
	private static final String	EMPTY	= "";		//$NON-NLS-1$
	private TKCorrectableField	mNameField;
	private TKCheckbox			mEnabledField;
	private TKTextField			mNotesField;
	private TKTextField			mReferenceField;
	private TKTextField			mCostField;
	private TKTextField			mLevelField;
	private TKTextField			mCostModifierField;
	private CSFeatures			mFeatures;
	private TKTabbedPanel		mTabPanel;
	private TKPopupMenu			mCostType;
	private TKPopupMenu			mAffects;
	private int					mLastLevel;

	/**
	 * Creates a new {@link CSModifierEditor}.
	 * 
	 * @param modifier The {@link CMModifier} to edit.
	 */
	public CSModifierEditor(CMModifier modifier) {
		super(modifier);

		TKPanel content = new TKPanel(new TKColumnLayout(2));
		TKPanel fields = new TKPanel(new TKColumnLayout(2));
		TKLabel icon = new TKLabel(modifier.getImage(true));

		TKPanel wrapper = new TKPanel(new TKColumnLayout(2));
		mNameField = createCorrectableField(fields, wrapper, Msgs.NAME, modifier.getName(), Msgs.NAME_TOOLTIP);
		mEnabledField = new TKCheckbox(Msgs.ENABLED, modifier.isEnabled());
		mEnabledField.setToolTipText(Msgs.ENABLED_TOOLTIP);
		mEnabledField.setEnabled(mIsEditable);
		wrapper.add(mEnabledField);
		fields.add(wrapper);

		createCostModifierFields(fields);

		wrapper = new TKPanel(new TKColumnLayout(3));
		mNotesField = createField(fields, wrapper, Msgs.NOTES, modifier.getNotes(), Msgs.NOTES_TOOLTIP, 0);
		mReferenceField = createField(wrapper, wrapper, Msgs.REFERENCE, mRow.getReference(), Msgs.REFERENCE_TOOLTIP, 6);
		fields.add(wrapper);

		icon.setVerticalAlignment(TKAlignment.TOP);
		icon.setAlignmentY(-1f);
		content.add(icon);
		content.add(fields);
		add(content);

		ArrayList<TKPanel> panels = new ArrayList<TKPanel>();

		mFeatures = new CSFeatures(mRow, mRow.getFeatures());

		panels.add(embedEditor(mFeatures));

		mTabPanel = new TKTabbedPanel(panels);
		mTabPanel.setSelectedPanelByName(getLastTabName());
		add(mTabPanel);
	}

	@Override protected boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());

		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setNotes(mNotesField.getText());
		if (getCostType() == CMCostType.MULTIPLIER) {
			modified |= mRow.setCostMultiplier(getCostMultiplier());
		} else {
			modified |= mRow.setCost(getCost());
		}
		if (hasLevels()) {
			modified |= mRow.setLevels(getLevels());
			modified |= mRow.setCostType(CMCostType.PERCENTAGE);
		} else {
			modified |= mRow.setLevels(0);
			modified |= mRow.setCostType(getCostType());
		}
		modified |= mRow.setAffects((CMAffects) mAffects.getSelectedItemUserObject());
		modified |= mRow.setEnabled(mEnabledField.isChecked());

		if (mFeatures != null) {
			modified |= mRow.setFeatures(mFeatures.getFeatures());
		}
		return modified;
	}

	private boolean hasLevels() {
		return mCostType.getSelectedItemUserObject() instanceof Boolean;
	}

	@Override public void finished() {
		if (mTabPanel != null) {
			updateLastTabName(mTabPanel.getSelectedPanelName());
		}
	}

	private TKCorrectableField createCorrectableField(TKPanel labelParent, TKPanel fieldParent, String title, String text, String tooltip) {
		TKCorrectableLabel label = new TKCorrectableLabel(title);
		TKCorrectableField field = new TKCorrectableField(label, text);

		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.addActionListener(this);
		labelParent.add(label);
		fieldParent.add(field);
		return field;
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();

		if (src == mNameField) {
			nameChanged();
		} else if (src == mCostType) {
			costTypeChanged();
		} else if (src == mCostField || src == mLevelField) {
			updateCostModifier();
		}
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

	private void disableControls(TKPanel panel) {
		int count = panel.getComponentCount();

		for (int i = 0; i < count; i++) {
			disableControls((TKPanel) panel.getComponent(i));
		}

		if (panel instanceof TKBaseButton || panel instanceof TKTextField || panel instanceof TKPopupMenu || panel instanceof TKCheckbox) {
			panel.setEnabled(false);
		}
	}

	private void nameChanged() {
		if (mNameField.getText().trim().length() == 0) {
			String corrected = mRow.getName().trim();

			if (corrected.length() == 0) {
				corrected = mRow.getLocalizedName();
			}
			mNameField.correct(corrected, Msgs.NAME_CANNOT_BE_EMPTY);
		}
	}

	private TKTextField createNumberField(TKPanel labelParent, TKPanel fieldParent, String title, boolean allowSign, int value, String tooltip, int maxDigits) {
		TKTextField field = new TKTextField(TKTextUtility.makeFiller(maxDigits, '9') + TKTextUtility.makeFiller(maxDigits / 3, ',') + (allowSign ? "-" : EMPTY)); //$NON-NLS-1$

		field.setOnlySize(field.getPreferredSize());
		field.setText(TKNumberUtils.format(value));
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.setKeyEventFilter(new TKNumberFilter(false, allowSign, true, maxDigits));
		field.addActionListener(this);
		labelParent.add(new TKLinkedLabel(field, title));
		fieldParent.add(field);
		return field;
	}

	private TKTextField createNumberField(TKPanel labelParent, TKPanel fieldParent, String title, double value, String tooltip, int maxDigits) {
		TKTextField field = new TKTextField(TKTextUtility.makeFiller(maxDigits, '9') + TKTextUtility.makeFiller(maxDigits / 3, ',') + '.');

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

	private void createCostModifierFields(TKPanel parent) {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(7));

		mLastLevel = mRow.getLevels();
		if (mLastLevel < 1) {
			mLastLevel = 1;
		}
		if (mRow.getCostType() == CMCostType.MULTIPLIER) {
			mCostField = createNumberField(parent, wrapper, Msgs.COST, mRow.getCostMultiplier(), Msgs.COST_TOOLTIP, 5);
		} else {
			mCostField = createNumberField(parent, wrapper, Msgs.COST, true, mRow.getCost(), Msgs.COST_TOOLTIP, 5);
		}
		createCostType(wrapper);
		mLevelField = createNumberField(wrapper, wrapper, Msgs.LEVELS, false, mLastLevel, Msgs.LEVELS_TOOLTIP, 3);
		mCostModifierField = createNumberField(wrapper, wrapper, Msgs.TOTAL_COST, true, 0, Msgs.TOTAL_COST_TOOLTIP, 9);
		createAffects(wrapper);
		mCostModifierField.setEnabled(false);
		if (!mRow.hasLevels()) {
			mLevelField.setText(EMPTY);
			mLevelField.setEnabled(false);
		}
		parent.add(wrapper);
	}

	private void createAffects(TKPanel parent) {
		TKMenu menu = new TKMenu();

		for (CMAffects affects : CMAffects.values()) {
			TKMenuItem item = new TKMenuItem(affects.toString());
			item.setUserObject(affects);
			menu.add(item);
		}

		mAffects = new TKPopupMenu(menu);
		mAffects.setSelectedUserObject(mRow.getAffects());
		mAffects.setOnlySize(mAffects.getPreferredSize());
		mAffects.setEnabled(mIsEditable);
		parent.add(mAffects);
	}

	private void createCostType(TKPanel parent) {
		TKMenu menu = new TKMenu();

		for (CMCostType type : CMCostType.values()) {
			TKMenuItem item = new TKMenuItem(type.toString());
			item.setUserObject(type);
			menu.add(item);
			if (type == CMCostType.PERCENTAGE) {
				item = new TKMenuItem(MessageFormat.format(Msgs.HAS_LEVELS, type.toString()));
				item.setUserObject(Boolean.TRUE);
				menu.add(item);
			}
		}

		mCostType = new TKPopupMenu(menu);
		mCostType.setSelectedUserObject(mRow.hasLevels() ? Boolean.TRUE : mRow.getCostType());
		mCostType.setOnlySize(mCostType.getPreferredSize());
		mCostType.setEnabled(mIsEditable);
		mCostType.addActionListener(this);
		parent.add(mCostType);
	}

	private void costTypeChanged() {
		boolean hasLevels = hasLevels();

		if (hasLevels) {
			mLevelField.setText(TKNumberUtils.format(mLastLevel));
		} else {
			mLastLevel = TKNumberUtils.getInteger(mLevelField.getText(), 0);
			mLevelField.setText(EMPTY);
		}
		mLevelField.setEnabled(hasLevels);
		updateCostField();
		updateCostModifier();
	}

	private void updateCostField() {
		if (getCostType() == CMCostType.MULTIPLIER) {
			mCostField.setKeyEventFilter(new TKNumberFilter(true, false, true, 5));
			mCostField.setText(TKNumberUtils.format(Math.abs(TKNumberUtils.getDouble(mCostField.getText(), 0))));
		} else {
			mCostField.setKeyEventFilter(new TKNumberFilter(false, true, true, 5));
			mCostField.setText(TKNumberUtils.format(TKNumberUtils.getInteger(mCostField.getText(), 0), true));
		}
	}

	private void updateCostModifier() {
		boolean enabled = true;

		if (hasLevels()) {
			mCostModifierField.setText(TKNumberUtils.format(getCost() * getLevels(), true) + "%"); //$NON-NLS-1$
		} else {
			switch (getCostType()) {
				case PERCENTAGE:
					mCostModifierField.setText(TKNumberUtils.format(getCost(), true) + "%"); //$NON-NLS-1$
					break;
				case POINTS:
					mCostModifierField.setText(TKNumberUtils.format(getCost(), true));
					break;
				case MULTIPLIER:
					mCostModifierField.setText("x" + TKNumberUtils.format(getCostMultiplier())); //$NON-NLS-1$
					mAffects.setSelectedUserObject(CMAffects.TOTAL);
					enabled = false;
					break;
			}
		}
		mAffects.setEnabled(mIsEditable && enabled);
	}

	private CMCostType getCostType() {
		Object obj = mCostType.getSelectedItemUserObject();
		if (obj instanceof Boolean) {
			obj = CMCostType.PERCENTAGE;
		}
		return (CMCostType) obj;
	}

	private int getCost() {
		return TKNumberUtils.getInteger(mCostField.getText(), 0);
	}

	private double getCostMultiplier() {
		return TKNumberUtils.getDouble(mCostField.getText(), 0);
	}

	private int getLevels() {
		return TKNumberUtils.getInteger(mLevelField.getText(), 0);
	}
}
