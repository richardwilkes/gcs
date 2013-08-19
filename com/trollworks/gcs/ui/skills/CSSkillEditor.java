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

package com.trollworks.gcs.ui.skills;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMSkillAttribute;
import com.trollworks.gcs.model.skill.CMSkillDifficulty;
import com.trollworks.gcs.model.skill.CMSkillLevel;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.editor.defaults.CSDefaults;
import com.trollworks.gcs.ui.editor.feature.CSFeatures;
import com.trollworks.gcs.ui.editor.prereq.CSPrereqs;
import com.trollworks.gcs.ui.weapon.CSMeleeWeaponEditor;
import com.trollworks.gcs.ui.weapon.CSRangedWeaponEditor;
import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKFont;
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
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.tab.TKTabbedPanel;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashSet;

/** The detailed editor for {@link CMSkill}s. */
public class CSSkillEditor extends CSRowEditor<CMSkill> implements ActionListener {
	private TKCorrectableField		mNameField;
	private TKTextField				mSpecializationField;
	private TKTextField				mNotesField;
	private TKTextField				mReferenceField;
	private TKCheckbox				mHasTechLevel;
	private TKTextField				mTechLevel;
	private String					mSavedTechLevel;
	private TKPopupMenu				mAttributePopup;
	private TKPopupMenu				mDifficultyPopup;
	private TKTextField				mPointsField;
	private TKTextField				mLevelField;
	private TKPopupMenu				mEncPenaltyPopup;
	private TKTabbedPanel			mTabPanel;
	private CSPrereqs				mPrereqs;
	private CSFeatures				mFeatures;
	private CSDefaults				mDefaults;
	private CSMeleeWeaponEditor		mMeleeWeapons;
	private CSRangedWeaponEditor	mRangedWeapons;

	/**
	 * Creates a new {@link CMSkill} editor.
	 * 
	 * @param skill The {@link CMSkill} to edit.
	 */
	public CSSkillEditor(CMSkill skill) {
		super(skill);

		TKPanel content = new TKPanel(new TKColumnLayout(2));
		TKPanel fields = new TKPanel(new TKColumnLayout(2));
		TKLabel icon = new TKLabel(skill.getImage(true));
		boolean notContainer = !skill.canHaveChildren();
		TKPanel wrapper;

		mNameField = createCorrectableField(fields, Msgs.NAME, skill.getName(), Msgs.NAME_TOOLTIP);
		if (notContainer) {
			wrapper = new TKPanel(new TKColumnLayout(2));
			mSpecializationField = createField(fields, wrapper, Msgs.SPECIALIZATION, skill.getSpecialization(), Msgs.SPECIALIZATION_TOOLTIP, 0);
			createTechLevelFields(wrapper);
			fields.add(wrapper);
			mEncPenaltyPopup = createEncumbrancePenaltyMultiplierPopup(fields);
		}
		mNotesField = createField(fields, fields, Msgs.NOTES, skill.getNotes(), Msgs.NOTES_TITLE, 0);
		if (notContainer) {
			wrapper = createDifficultyPopups(fields);
		} else {
			wrapper = fields;
		}
		mReferenceField = createField(wrapper, wrapper, Msgs.EDITOR_REFERENCE, mRow.getReference(), Msgs.REFERENCE_TOOLTIP, 6);
		icon.setVerticalAlignment(TKAlignment.TOP);
		icon.setAlignmentY(-1f);
		content.add(icon);
		content.add(fields);
		add(content);

		if (notContainer) {
			ArrayList<TKPanel> panels = new ArrayList<TKPanel>();

			mPrereqs = new CSPrereqs(mRow, mRow.getPrereqs());
			mMeleeWeapons = CSMeleeWeaponEditor.createEditor(mRow);
			mRangedWeapons = CSRangedWeaponEditor.createEditor(mRow);
			mFeatures = new CSFeatures(mRow, mRow.getFeatures());
			mDefaults = new CSDefaults(mRow.getDefaults());
			mDefaults.addActionListener(this);
			panels.add(embedEditor(mDefaults));
			panels.add(embedEditor(mPrereqs));
			panels.add(embedEditor(mFeatures));
			panels.add(mMeleeWeapons);
			panels.add(mRangedWeapons);
			if (!mIsEditable) {
				disableControls(mMeleeWeapons);
				disableControls(mRangedWeapons);
			}
			mTabPanel = new TKTabbedPanel(panels);
			mTabPanel.setSelectedPanelByName(getLastTabName());
			add(mTabPanel);
		}
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

	private void createPointsFields(TKPanel parent, boolean forCharacter) {
		mPointsField = createField(parent, parent, Msgs.EDITOR_POINTS, Integer.toString(mRow.getPoints()), Msgs.EDITOR_POINTS_TOOLTIP, 4);
		mPointsField.setKeyEventFilter(new TKNumberFilter(false, false, false, 4));
		mPointsField.addActionListener(this);

		if (forCharacter) {
			mLevelField = createField(parent, parent, Msgs.EDITOR_LEVEL, CMSkill.getSkillDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel(), mRow.getAttribute(), mRow.canHaveChildren()), Msgs.EDITOR_LEVEL_TOOLTIP, 8);
			mLevelField.setEnabled(false);
		}
	}

	private void createTechLevelFields(TKPanel parent) {
		TKOutlineModel owner = mRow.getOwner();
		CMCharacter character = mRow.getCharacter();
		boolean enabled = !owner.isLocked();
		boolean hasTL;

		mSavedTechLevel = mRow.getTechLevel();
		hasTL = mSavedTechLevel != null;
		if (!hasTL) {
			mSavedTechLevel = ""; //$NON-NLS-1$
		}

		if (character != null) {
			TKPanel wrapper = new TKPanel(new TKColumnLayout(2));

			mHasTechLevel = new TKCheckbox(Msgs.TECH_LEVEL, hasTL);
			mHasTechLevel.setFontKey(TKFont.TEXT_FONT_KEY);
			mHasTechLevel.setToolTipText(Msgs.TECH_LEVEL_TOOLTIP);
			mHasTechLevel.setEnabled(enabled);
			mHasTechLevel.addActionListener(this);
			wrapper.add(mHasTechLevel);

			mTechLevel = new TKTextField("9999"); //$NON-NLS-1$
			mTechLevel.setOnlySize(mTechLevel.getPreferredSize());
			mTechLevel.setText(mSavedTechLevel);
			mTechLevel.setToolTipText(Msgs.TECH_LEVEL_TOOLTIP);
			mTechLevel.setEnabled(enabled && hasTL);
			wrapper.add(mTechLevel);
			parent.add(wrapper);

			if (!hasTL) {
				mSavedTechLevel = character.getTechLevel();
			}
		} else {
			mTechLevel = new TKTextField(mSavedTechLevel);
			mHasTechLevel = new TKCheckbox(Msgs.TECH_LEVEL_REQUIRED, hasTL);
			mHasTechLevel.setFontKey(TKFont.TEXT_FONT_KEY);
			mHasTechLevel.setToolTipText(Msgs.TECH_LEVEL_REQUIRED_TOOLTIP);
			mHasTechLevel.setEnabled(enabled);
			mHasTechLevel.addActionListener(this);
			parent.add(mHasTechLevel);
		}
	}

	private TKPopupMenu createEncumbrancePenaltyMultiplierPopup(TKPanel parent) {
		TKMenu menu = new TKMenu();
		TKMenuItem item = new TKMenuItem(Msgs.NO_ENC_PENALTY);
		item.setUserObject(new Integer(0));
		menu.add(item);
		item = new TKMenuItem(Msgs.ONE_ENC_PENALTY);
		item.setUserObject(new Integer(1));
		menu.add(item);
		for (int i = 2; i < 10; i++) {
			Integer value = new Integer(i);
			item = new TKMenuItem(MessageFormat.format(Msgs.ENC_PENALTY_FORMAT, value));
			item.setUserObject(value);
			menu.add(item);
		}
		TKPopupMenu popup = new TKPopupMenu(menu);
		popup.setSelectedUserObject(new Integer(mRow.getEncumbrancePenaltyMultiplier()));
		popup.setOnlySize(popup.getPreferredSize());
		popup.setToolTipText(Msgs.ENC_PENALTY_MULT_TOOLTIP);
		popup.setEnabled(mIsEditable);
		popup.addActionListener(this);
		parent.add(new TKLinkedLabel(popup, Msgs.ENC_PENALTY_MULT));
		parent.add(popup);
		return popup;
	}

	private TKPanel createDifficultyPopups(TKPanel parent) {
		CMCharacter character = mRow.getCharacter();
		boolean forCharacterOrTemplate = character != null || mRow.getTemplate() != null;
		TKLabel label = new TKLabel(Msgs.EDITOR_DIFFICULTY, TKAlignment.RIGHT);
		TKPanel wrapper = new TKPanel(new TKColumnLayout(forCharacterOrTemplate ? character != null ? 10 : 8 : 6));
		TKMenu menu = new TKMenu();

		label.setToolTipText(Msgs.EDITOR_DIFFICULTY_TOOLTIP);

		for (CMSkillAttribute attribute : CMSkillAttribute.values()) {
			TKMenuItem item = new TKMenuItem(attribute.toString());

			item.setUserObject(attribute);
			menu.add(item);
		}
		mAttributePopup = new TKPopupMenu(menu);
		mAttributePopup.setSelectedUserObject(mRow.getAttribute());
		mAttributePopup.setOnlySize(mAttributePopup.getPreferredSize());
		mAttributePopup.setToolTipText(Msgs.ATTRIBUTE_POPUP_TOOLTIP);
		mAttributePopup.setEnabled(mIsEditable);
		mAttributePopup.addActionListener(this);
		wrapper.add(mAttributePopup);
		wrapper.add(new TKLabel(" /")); //$NON-NLS-1$

		menu = new TKMenu();
		for (CMSkillDifficulty difficulty : CMSkillDifficulty.values()) {
			TKMenuItem item = new TKMenuItem(difficulty.toString());

			item.setUserObject(difficulty);
			menu.add(item);
		}
		mDifficultyPopup = new TKPopupMenu(menu);
		mDifficultyPopup.setSelectedUserObject(mRow.getDifficulty());
		mDifficultyPopup.setOnlySize(mDifficultyPopup.getPreferredSize());
		mDifficultyPopup.setToolTipText(Msgs.EDITOR_DIFFICULTY_POPUP_TOOLTIP);
		mDifficultyPopup.setEnabled(mIsEditable);
		mDifficultyPopup.addActionListener(this);
		wrapper.add(mDifficultyPopup);

		if (forCharacterOrTemplate) {
			createPointsFields(wrapper, character != null);
		}
		wrapper.add(new TKPanel());

		parent.add(label);
		parent.add(wrapper);
		return wrapper;
	}

	private void recalculateLevel() {
		if (mLevelField != null) {
			CMSkillAttribute attribute = getSkillAttribute();
			CMSkillLevel level = CMSkill.calculateLevel(mRow.getCharacter(), mRow, mNameField.getText(), mSpecializationField.getText(), mDefaults.getDefaults(), attribute, getSkillDifficulty(), getSkillPoints(), new HashSet<CMSkill>(), getEncumbrancePenaltyMultiplier());

			mLevelField.setText(CMSkill.getSkillDisplayLevel(level.mLevel, level.mRelativeLevel, attribute, false));
		}
	}

	private CMSkillAttribute getSkillAttribute() {
		return (CMSkillAttribute) mAttributePopup.getSelectedItemUserObject();
	}

	private CMSkillDifficulty getSkillDifficulty() {
		return (CMSkillDifficulty) mDifficultyPopup.getSelectedItemUserObject();
	}

	private int getSkillPoints() {
		return TKNumberUtils.getInteger(mPointsField.getText(), 0);
	}

	private int getEncumbrancePenaltyMultiplier() {
		return ((Integer) mEncPenaltyPopup.getSelectedItemUserObject()).intValue();
	}

	@Override public boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());

		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setNotes(mNotesField.getText());
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

		if (src == mNameField) {
			if (mNameField.getText().trim().length() == 0) {
				String corrected = mRow.getName().trim();

				if (corrected.length() == 0) {
					corrected = mRow.getLocalizedName();
				}
				mNameField.correct(corrected, Msgs.NAME_CANNOT_BE_EMPTY);
			}
		} else if (src == mHasTechLevel) {
			boolean enabled = mHasTechLevel.isChecked();

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
}
