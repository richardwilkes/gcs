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
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.tab.TKTabbedPanel;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;

/** Editor for {@link CMModifier}s. */
public class CSModifierEditor extends CSRowEditor<CMModifier> implements ActionListener {
	private static final String	EMPTY	= "";		//$NON-NLS-1$
	private TKCorrectableField	mNameField;
	private TKTextField			mNotesField;
	private TKTextField			mReferenceField;
	private TKTextField			mCostField;
	private TKTextField			mLevelField;
	private TKTextField			mCostModifierField;
	private CSFeatures			mFeatures;
	private TKTabbedPanel		mTabPanel;
	private TKPopupMenu			mLevelType;
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
		TKPanel wrapper;

		TKLabel icon = new TKLabel(modifier.getImage(true));

		mNameField = createCorrectableField(fields, Msgs.NAME, modifier.getName(), Msgs.NAME_TOOLTIP);
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
		modified |= mRow.setCost(getCost());
		modified |= mRow.setLevels(hasLevels() ? getLevels() : 0);

		if (mFeatures != null) {
			modified |= mRow.setFeatures(mFeatures.getFeatures());
		}
		return modified;
	}

	private boolean hasLevels() {
		return ((Boolean) mLevelType.getSelectedItemUserObject()).booleanValue();
	}

	@Override public void finished() {
		if (mTabPanel != null) {
			updateLastTabName(mTabPanel.getSelectedPanelName());
		}
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

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();

		if (src == mNameField) {
			nameChanged();
		} else if (src == mLevelType) {
			levelTypeChanged();
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

		if (panel instanceof TKBaseButton || panel instanceof TKTextField || panel instanceof TKPopupMenu) {
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

	private void createCostModifierFields(TKPanel parent) {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(6));

		mLastLevel = mRow.getLevels();
		if (mLastLevel < 1) {
			mLastLevel = 1;
		}
		mCostField = createNumberField(parent, wrapper, Msgs.COST, true, mRow.getCost(), Msgs.COST_TOOLTIP, 5);
		createLevelType(wrapper);
		mLevelField = createNumberField(wrapper, wrapper, Msgs.LEVELS, false, mLastLevel, Msgs.LEVELS_TOOLTIP, 3);
		mCostModifierField = createNumberField(wrapper, wrapper, Msgs.TOTAL_COST, true, mRow.getCostModifier(), Msgs.TOTAL_COST_TOOLTIP, 9);
		mCostModifierField.setEnabled(false);
		if (!mRow.hasLevels()) {
			mLevelField.setText(EMPTY);
			mLevelField.setEnabled(false);
		}
		parent.add(wrapper);
	}

	private void createLevelType(TKPanel parent) {
		TKMenu menu = new TKMenu();
		TKMenuItem item = new TKMenuItem(Msgs.NO_LEVELS);

		item.setUserObject(Boolean.FALSE);
		menu.add(item);
		item = new TKMenuItem(Msgs.HAS_LEVELS);
		item.setUserObject(Boolean.TRUE);
		menu.add(item);

		mLevelType = new TKPopupMenu(menu);
		mLevelType.setSelectedUserObject(mRow.hasLevels() ? Boolean.TRUE : Boolean.FALSE);
		mLevelType.setOnlySize(mLevelType.getPreferredSize());
		mLevelType.addActionListener(this);
		parent.add(mLevelType);
	}

	private void levelTypeChanged() {
		boolean hasLevels = hasLevels();

		if (hasLevels) {
			mLevelField.setText(TKNumberUtils.format(mLastLevel));
		} else {
			mLastLevel = TKNumberUtils.getInteger(mLevelField.getText(), 0);
			mLevelField.setText(EMPTY);
		}
		mLevelField.setEnabled(hasLevels);
		updateCostModifier();
	}

	private void updateCostModifier() {
		mCostModifierField.setText(TKNumberUtils.format(hasLevels() ? getCost() * getLevels() : getCost(), true) + "%"); //$NON-NLS-1$
	}

	private int getCost() {
		return TKNumberUtils.getInteger(mCostField.getText(), 0);
	}

	private int getLevels() {
		return TKNumberUtils.getInteger(mLevelField.getText(), 0);
	}
}
