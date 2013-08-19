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

package com.trollworks.gcs.ui.advantage;

import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.advantage.CMAdvantageContainerType;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.editor.defaults.CSDefaults;
import com.trollworks.gcs.ui.editor.feature.CSFeatures;
import com.trollworks.gcs.ui.editor.prereq.CSPrereqs;
import com.trollworks.gcs.ui.modifiers.CSModifierListEditor;
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
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKBaseButton;
import com.trollworks.toolkit.widget.button.TKCheckbox;
import com.trollworks.toolkit.widget.button.TKToggleButton;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.tab.TKTabbedPanel;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.util.ArrayList;

/** The detailed editor for {@link CMAdvantage}s. */
public class CSAdvantageEditor extends CSRowEditor<CMAdvantage> implements ActionListener {
	private static final String		EMPTY	= "";			//$NON-NLS-1$
	private TKCorrectableField		mNameField;
	private TKPopupMenu				mLevelType;
	private TKTextField				mBasePointsField;
	private TKTextField				mLevelField;
	private TKTextField				mLevelPointsField;
	private TKTextField				mPointsField;
	private TKTextField				mNotesField;
	private TKTextField				mReferenceField;
	private TKTabbedPanel			mTabPanel;
	private CSPrereqs				mPrereqs;
	private CSFeatures				mFeatures;
	private CSDefaults				mDefaults;
	private CSMeleeWeaponEditor		mMeleeWeapons;
	private CSRangedWeaponEditor	mRangedWeapons;
	private CSModifierListEditor	mModifiers;
	private int						mLastLevel;
	private int						mLastPointsPerLevel;
	private TKToggleButton			mMentalType;
	private TKToggleButton			mPhysicalType;
	private TKToggleButton			mSocialType;
	private TKToggleButton			mExoticType;
	private TKToggleButton			mSupernaturalType;
	private TKPopupMenu				mContainerTypePopup;

	/**
	 * Creates a new {@link CMAdvantage} editor.
	 * 
	 * @param advantage The {@link CMAdvantage} to edit.
	 */
	public CSAdvantageEditor(CMAdvantage advantage) {
		super(advantage);

		TKPanel content = new TKPanel(new TKColumnLayout(2));
		TKPanel fields = new TKPanel(new TKColumnLayout(2));
		TKLabel icon = new TKLabel(advantage.getImage(true));
		boolean notContainer = !advantage.canHaveChildren();

		mNameField = createCorrectableField(fields, Msgs.NAME, advantage.getName(), Msgs.NAME_TOOLTIP);
		if (notContainer) {
			createPointsFields(fields);
		}
		mNotesField = createField(fields, fields, Msgs.NOTES, advantage.getNotes(), Msgs.NOTES_TOOLTIP, 0);
		if (notContainer) {
			TKPanel wrapper = new TKPanel(new TKColumnLayout(8));
			TKLabel label = new TKLabel(Msgs.TYPE, TKAlignment.RIGHT);

			label.setToolTipText(Msgs.EDTIOR_TYPE_TOOLTIP);
			fields.add(label);
			mMentalType = createType(wrapper, CSImage.getMentalTypeIcon(), CSImage.getMentalTypeSelectedIcon(), (mRow.getType() & CMAdvantage.TYPE_MASK_MENTAL) == CMAdvantage.TYPE_MASK_MENTAL, Msgs.MENTAL);
			mPhysicalType = createType(wrapper, CSImage.getPhysicalTypeIcon(), CSImage.getPhysicalTypeSelectedIcon(), (mRow.getType() & CMAdvantage.TYPE_MASK_PHYSICAL) == CMAdvantage.TYPE_MASK_PHYSICAL, Msgs.PHYSICAL);
			mSocialType = createType(wrapper, CSImage.getSocialTypeIcon(), CSImage.getSocialTypeSelectedIcon(), (mRow.getType() & CMAdvantage.TYPE_MASK_SOCIAL) == CMAdvantage.TYPE_MASK_SOCIAL, Msgs.SOCIAL);
			mExoticType = createType(wrapper, CSImage.getExoticTypeIcon(), CSImage.getExoticTypeSelectedIcon(), (mRow.getType() & CMAdvantage.TYPE_MASK_EXOTIC) == CMAdvantage.TYPE_MASK_EXOTIC, Msgs.EXOTIC);
			mSupernaturalType = createType(wrapper, CSImage.getSupernaturalTypeIcon(), CSImage.getSupernaturalTypeSelectedIcon(), (mRow.getType() & CMAdvantage.TYPE_MASK_SUPERNATURAL) == CMAdvantage.TYPE_MASK_SUPERNATURAL, Msgs.SUPERNATURAL);
			wrapper.add(new TKPanel());
			mReferenceField = createField(wrapper, wrapper, Msgs.REFERENCE, mRow.getReference(), Msgs.EDITOR_REFERENCE_TOOLTIP, 6);
			fields.add(wrapper);
		} else {
			TKPanel wrapper = new TKPanel(new TKColumnLayout(5));

			createContainerTypePopup(fields, wrapper);
			mReferenceField = createField(wrapper, wrapper, Msgs.REFERENCE, mRow.getReference(), Msgs.EDITOR_REFERENCE_TOOLTIP, 6);
			wrapper.add(new TKPanel());
			fields.add(wrapper);
		}
		icon.setVerticalAlignment(TKAlignment.TOP);
		icon.setAlignmentY(-1f);
		content.add(icon);
		content.add(fields);
		add(content);

		ArrayList<TKPanel> panels = new ArrayList<TKPanel>();
		mModifiers = CSModifierListEditor.createEditor(mRow);
		mModifiers.addActionListener(this);
		if (notContainer) {
			mPrereqs = new CSPrereqs(mRow, mRow.getPrereqs());
			mFeatures = new CSFeatures(mRow, mRow.getFeatures());
			mDefaults = new CSDefaults(mRow.getDefaults());
			mMeleeWeapons = CSMeleeWeaponEditor.createEditor(mRow);
			mRangedWeapons = CSRangedWeaponEditor.createEditor(mRow);
			mDefaults.addActionListener(this);
			panels.add(embedEditor(mDefaults));
			panels.add(embedEditor(mPrereqs));
			panels.add(embedEditor(mFeatures));
			panels.add(mModifiers);
			panels.add(mMeleeWeapons);
			panels.add(mRangedWeapons);

			if (!mIsEditable) {
				disableControls(mMeleeWeapons);
				disableControls(mRangedWeapons);
			}
			updatePoints();
		} else {
			panels.add(mModifiers);
		}
		if (!mIsEditable) {
			disableControls(mModifiers);
		}
		mTabPanel = new TKTabbedPanel(panels);
		mTabPanel.setSelectedPanelByName(getLastTabName());
		add(mTabPanel);
	}

	private TKToggleButton createType(TKPanel parent, BufferedImage icon, BufferedImage selectedIcon, boolean selected, String tooltip) {
		TKToggleButton button = new TKToggleButton(icon, selected);

		button.setPressedIcon(selectedIcon);
		button.setToolTipText(tooltip);
		button.setEnabled(mIsEditable);
		parent.add(button);
		return button;
	}

	private void createContainerTypePopup(TKPanel labelParent, TKPanel popupParent) {
		TKMenu menu = new TKMenu();

		for (CMAdvantageContainerType type : CMAdvantageContainerType.values()) {
			TKMenuItem item = new TKMenuItem(type.toString());

			item.setUserObject(type);
			menu.add(item);
		}

		mContainerTypePopup = new TKPopupMenu(menu);
		mContainerTypePopup.setSelectedUserObject(mRow.getContainerType());
		mContainerTypePopup.setOnlySize(mContainerTypePopup.getPreferredSize());
		mContainerTypePopup.setToolTipText(Msgs.CONTAINER_TYPE_TOOLTIP);

		labelParent.add(new TKLinkedLabel(mContainerTypePopup, Msgs.CONTAINER_TYPE));
		popupParent.add(mContainerTypePopup);
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
		mLevelType.setSelectedUserObject(mRow.isLeveled() ? Boolean.TRUE : Boolean.FALSE);
		mLevelType.setOnlySize(mLevelType.getPreferredSize());
		mLevelType.setEnabled(mIsEditable);
		mLevelType.addActionListener(this);
		parent.add(mLevelType);
	}

	private void createPointsFields(TKPanel parent) {
		boolean isLeveled = mRow.isLeveled();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(8));
		TKPanel lastWrapper = new TKPanel(new TKColumnLayout(2, 0, 0));

		mLastLevel = mRow.getLevels();
		mLastPointsPerLevel = mRow.getPointsPerLevel();
		if (mLastLevel < 0) {
			mLastLevel = 1;
		}
		mBasePointsField = createNumberField(parent, wrapper, Msgs.BASE_POINTS, true, mRow.getPoints(), Msgs.BASE_POINTS_TOOLTIP, 4);
		createLevelType(wrapper);
		mLevelField = createNumberField(wrapper, wrapper, Msgs.LEVEL, false, mLastLevel, Msgs.LEVEL_TOOLTIP, 3);
		mLevelPointsField = createNumberField(wrapper, wrapper, Msgs.LEVEL_POINTS, true, mLastPointsPerLevel, Msgs.LEVEL_POINTS_TOOLTIP, 4);
		mPointsField = createNumberField(wrapper, lastWrapper, Msgs.TOTAL_POINTS, true, mRow.getAdjustedPoints(), Msgs.TOTAL_POINTS_TOOLTIP, 5);
		mPointsField.setEnabled(false);
		lastWrapper.add(new TKPanel());
		wrapper.add(lastWrapper);
		if (!isLeveled) {
			mLevelField.setText(EMPTY);
			mLevelField.setEnabled(false);
			mLevelPointsField.setText(EMPTY);
			mLevelPointsField.setEnabled(false);
		}
		parent.add(wrapper);
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

	@Override public boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());

		if (mRow.canHaveChildren()) {
			modified |= mRow.setContainerType((CMAdvantageContainerType) mContainerTypePopup.getSelectedItemUserObject());
		} else {
			int type = 0;

			if (mMentalType.isSelected()) {
				type |= CMAdvantage.TYPE_MASK_MENTAL;
			}
			if (mPhysicalType.isSelected()) {
				type |= CMAdvantage.TYPE_MASK_PHYSICAL;
			}
			if (mSocialType.isSelected()) {
				type |= CMAdvantage.TYPE_MASK_SOCIAL;
			}
			if (mExoticType.isSelected()) {
				type |= CMAdvantage.TYPE_MASK_EXOTIC;
			}
			if (mSupernaturalType.isSelected()) {
				type |= CMAdvantage.TYPE_MASK_SUPERNATURAL;
			}
			modified |= mRow.setType(type);
			modified |= mRow.setPoints(getBasePoints());
			if (isLeveled()) {
				modified |= mRow.setPointsPerLevel(getPointsPerLevel());
				modified |= mRow.setLevels(getLevels());
			} else {
				modified |= mRow.setPointsPerLevel(0);
				modified |= mRow.setLevels(-1);
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
				ArrayList<CMWeaponStats> list = new ArrayList<CMWeaponStats>(mMeleeWeapons.getWeapons());

				list.addAll(mRangedWeapons.getWeapons());
				modified |= mRow.setWeapons(list);
			}
		}
		if (mModifiers.wasModified()) {
			modified = true;
			mRow.setModifiers(mModifiers.getModifiers());
		}
		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setNotes(mNotesField.getText());
		return modified;
	}

	@Override public void finished() {
		if (mTabPanel != null) {
			updateLastTabName(mTabPanel.getSelectedPanelName());
		}
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();

		if (src == mNameField) {
			nameChanged();
		} else if (src == mLevelType) {
			levelTypeChanged();
		} else if (src == mBasePointsField || src == mLevelField || src == mLevelPointsField) {
			updatePoints();
		} else if (src == mModifiers) {
			updatePoints();
		}
	}

	private boolean isLeveled() {
		return ((Boolean) mLevelType.getSelectedItemUserObject()).booleanValue();
	}

	private void levelTypeChanged() {
		boolean isLeveled = isLeveled();

		if (isLeveled) {
			mLevelField.setText(TKNumberUtils.format(mLastLevel));
			mLevelPointsField.setText(TKNumberUtils.format(mLastPointsPerLevel));
		} else {
			mLastLevel = TKNumberUtils.getInteger(mLevelField.getText(), 0);
			mLastPointsPerLevel = TKNumberUtils.getInteger(mLevelPointsField.getText(), 0);
			mLevelField.setText(EMPTY);
			mLevelPointsField.setText(EMPTY);
		}
		mLevelField.setEnabled(isLeveled);
		mLevelPointsField.setEnabled(isLeveled);
		updatePoints();
	}

	private int getLevels() {
		return TKNumberUtils.getInteger(mLevelField.getText(), 0);
	}

	private int getPointsPerLevel() {
		return TKNumberUtils.getInteger(mLevelPointsField.getText(), 0);
	}

	private int getBasePoints() {
		return TKNumberUtils.getInteger(mBasePointsField.getText(), 0);
	}

	private int getPoints() {
		if (mModifiers == null) {
			return 0;
		}
		return CMAdvantage.getAdjustedPoints(getBasePoints(), isLeveled() ? getLevels() : 0, getPointsPerLevel(), mModifiers.getAllModifiers());
	}

	private void updatePoints() {
		if (mPointsField != null) {
			mPointsField.setText(TKNumberUtils.format(getPoints()));
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
}
