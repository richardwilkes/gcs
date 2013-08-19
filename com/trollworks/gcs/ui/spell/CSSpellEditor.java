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

package com.trollworks.gcs.ui.spell;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.skill.CMSkillDifficulty;
import com.trollworks.gcs.model.skill.CMSkillLevel;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.editor.prereq.CSPrereqs;
import com.trollworks.gcs.ui.weapon.CSMeleeWeaponEditor;
import com.trollworks.gcs.ui.weapon.CSRangedWeaponEditor;
import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.utility.TKFont;
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
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.tab.TKTabbedPanel;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;

/** The detailed editor for {@link CMSpell}s. */
public class CSSpellEditor extends CSRowEditor<CMSpell> implements ActionListener {
	private TKCorrectableField		mNameField;
	private TKTextField				mCollegeField;
	private TKCorrectableField		mClassField;
	private TKCorrectableField		mCastingCostField;
	private TKTextField				mMaintenanceField;
	private TKCorrectableField		mCastingTimeField;
	private TKCorrectableField		mDurationField;
	private TKPopupMenu				mDifficultyPopup;
	private TKTextField				mNotesField;
	private TKTextField				mPointsField;
	private TKTextField				mLevelField;
	private TKTextField				mReferenceField;
	private TKTabbedPanel			mTabPanel;
	private CSPrereqs				mPrereqs;
	private TKCheckbox				mHasTechLevel;
	private TKTextField				mTechLevel;
	private String					mSavedTechLevel;
	private CSMeleeWeaponEditor		mMeleeWeapons;
	private CSRangedWeaponEditor	mRangedWeapons;

	/**
	 * Creates a new {@link CMSpell} editor.
	 * 
	 * @param spell The {@link CMSpell} to edit.
	 */
	public CSSpellEditor(CMSpell spell) {
		super(spell);

		boolean notContainer = !spell.canHaveChildren();
		TKPanel content = new TKPanel(new TKColumnLayout(2));
		TKPanel fields = new TKPanel(new TKColumnLayout());
		TKPanel wrapper1 = new TKPanel(new TKColumnLayout(notContainer ? 3 : 2));
		TKPanel wrapper2 = new TKPanel(new TKColumnLayout(4));
		TKPanel wrapper3 = new TKPanel(new TKColumnLayout(2));
		TKPanel noGapWrapper = new TKPanel(new TKColumnLayout(2, 0, 0));
		TKPanel ptsPanel = null;
		TKLabel icon = new TKLabel(spell.getImage(true));
		Dimension size = new Dimension();
		TKPanel refParent = wrapper3;

		mNameField = createCorrectableField(wrapper1, wrapper1, Msgs.NAME, spell.getName(), Msgs.NAME_TOOLTIP);
		fields.add(wrapper1);
		if (notContainer) {
			createTechLevelFields(wrapper1);
			mCollegeField = createField(wrapper2, wrapper2, Msgs.COLLEGE, spell.getCollege(), Msgs.COLLEGE_TOOLTIP, 0);
			mClassField = createCorrectableField(wrapper2, wrapper2, Msgs.CLASS, spell.getSpellClass(), Msgs.CLASS_ONLY_TOOLTIP);
			mCastingCostField = createCorrectableField(wrapper2, wrapper2, Msgs.CASTING_COST, spell.getCastingCost(), Msgs.CASTING_COST_TOOLTIP);
			mMaintenanceField = createField(wrapper2, wrapper2, Msgs.MAINTENANCE_COST, spell.getMaintenance(), Msgs.MAINTENANCE_COST_TOOLTIP, 0);
			mCastingTimeField = createCorrectableField(wrapper2, wrapper2, Msgs.CASTING_TIME, spell.getCastingTime(), Msgs.CASTING_TIME_TOOLTIP);
			mDurationField = createCorrectableField(wrapper2, wrapper2, Msgs.DURATION, spell.getDuration(), Msgs.DURATION_TOOLTIP);
			fields.add(wrapper2);

			ptsPanel = createPointsFields();
			fields.add(ptsPanel);
			refParent = ptsPanel;
		}
		mNotesField = createField(wrapper3, wrapper3, Msgs.NOTES, spell.getNotes(), Msgs.NOTES_TOOLTIP, 0);
		mReferenceField = createField(refParent, noGapWrapper, Msgs.EDITOR_REFERENCE, mRow.getReference(), Msgs.REFERENCE_TOOLTIP, 6);
		noGapWrapper.add(new TKPanel());
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
			panels.add(embedEditor(mPrereqs));
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

	private void determineLargest(TKPanel panel, int every, Dimension size) {
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

	private void applySize(TKPanel panel, int every, Dimension size) {
		int count = panel.getComponentCount();

		for (int i = 0; i < count; i += every) {
			((TKPanel) panel.getComponent(i)).setOnlySize(size);
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

	private TKPanel createPointsFields() {
		boolean forCharacter = mRow.getCharacter() != null;
		boolean forTemplate = mRow.getTemplate() != null;
		TKPanel panel = new TKPanel(new TKColumnLayout(forCharacter ? 8 : forTemplate ? 6 : 4));
		TKMenu menu = new TKMenu();
		TKMenuItem item = new TKMenuItem(CMSkillDifficulty.H.name());

		item.setUserObject(CMSkillDifficulty.H);
		menu.add(item);
		item = new TKMenuItem(CMSkillDifficulty.VH.name());
		item.setUserObject(CMSkillDifficulty.VH);
		menu.add(item);

		mDifficultyPopup = new TKPopupMenu(menu);
		mDifficultyPopup.setSelectedUserObject(mRow.isVeryHard() ? CMSkillDifficulty.VH : CMSkillDifficulty.H);
		mDifficultyPopup.setToolTipText(Msgs.DIFFICULTY_TOOLTIP);
		mDifficultyPopup.setOnlySize(mDifficultyPopup.getPreferredSize());
		mDifficultyPopup.addActionListener(this);
		mDifficultyPopup.setEnabled(mIsEditable);
		panel.add(new TKLinkedLabel(mDifficultyPopup, Msgs.DIFFICULTY));
		panel.add(mDifficultyPopup);

		if (forCharacter || mRow.getTemplate() != null) {
			mPointsField = createField(panel, panel, Msgs.EDITOR_POINTS, Integer.toString(mRow.getPoints()), Msgs.EDITOR_POINTS_TOOLTIP, 4);
			mPointsField.setKeyEventFilter(new TKNumberFilter(false, false, false, 4));
			mPointsField.addActionListener(this);

			if (forCharacter) {
				mLevelField = createField(panel, panel, Msgs.EDITOR_LEVEL, getDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel()), Msgs.EDITOR_LEVEL_TOOLTIP, 5);
				mLevelField.setEnabled(false);
			}
		}
		return panel;
	}

	private String getDisplayLevel(int level, int relativeLevel) {
		if (level < 0) {
			return "-"; //$NON-NLS-1$
		}
		return TKNumberUtils.format(level) + "/IQ" + TKNumberUtils.format(relativeLevel, true); //$NON-NLS-1$
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

	@Override public boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());
		boolean notContainer = !mRow.canHaveChildren();

		modified |= mRow.setReference(mReferenceField.getText());
		if (notContainer) {
			if (mHasTechLevel != null) {
				modified |= mRow.setTechLevel(mHasTechLevel.isChecked() ? mTechLevel.getText() : null);
			}
			modified |= mRow.setCollege(mCollegeField.getText());
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
		if (mPrereqs != null) {
			modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
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
			checkFieldForEmpty(mNameField, mRow.getName(), mRow.getLocalizedName(), Msgs.NAME_CANNOT_BE_EMPTY);
		} else if (src == mClassField) {
			checkFieldForEmpty(mClassField, mRow.getSpellClass(), CMSpell.getDefaultSpellClass(), Msgs.CLASS_CANNOT_BE_EMPTY);
		} else if (src == mCastingCostField) {
			checkFieldForEmpty(mCastingCostField, mRow.getCastingCost(), CMSpell.getDefaultCastingCost(), Msgs.CASTING_COST_CANNOT_BE_EMPTY);
		} else if (src == mCastingTimeField) {
			checkFieldForEmpty(mCastingTimeField, mRow.getCastingTime(), CMSpell.getDefaultCastingTime(), Msgs.CASTING_TIME_CANNOT_BE_EMPTY);
		} else if (src == mDurationField) {
			checkFieldForEmpty(mDurationField, mRow.getDuration(), CMSpell.getDefaultDuration(), Msgs.DURATION_CANNOT_BE_EMPTY);
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
		} else if (src == mPointsField || src == mDifficultyPopup) {
			recalculateLevel();
		}
	}

	private void recalculateLevel() {
		if (mLevelField != null) {
			CMSkillLevel level = CMSpell.calculateLevel(mRow.getCharacter(), getSpellPoints(), isVeryHard(), mCollegeField.getText(), mNameField.getText());

			mLevelField.setText(getDisplayLevel(level.mLevel, level.mRelativeLevel));
		}
	}

	private int getSpellPoints() {
		return TKNumberUtils.getInteger(mPointsField.getText(), 0);
	}

	private boolean isVeryHard() {
		return mDifficultyPopup.getSelectedItemUserObject() == CMSkillDifficulty.VH;
	}

	private void checkFieldForEmpty(TKCorrectableField field, String correction, String defaultCorrection, String reason) {
		if (field.getText().trim().length() == 0) {
			String corrected = correction.trim();

			if (corrected.length() == 0) {
				corrected = defaultCorrection;
			}
			field.correct(corrected, reason);
		}
	}
}
