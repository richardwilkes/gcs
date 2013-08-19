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
import com.trollworks.gcs.model.skill.CMSkillDefault;
import com.trollworks.gcs.model.skill.CMSkillDefaultType;
import com.trollworks.gcs.model.skill.CMSkillDifficulty;
import com.trollworks.gcs.model.skill.CMSkillLevel;
import com.trollworks.gcs.model.skill.CMTechnique;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.gcs.ui.editor.CSRowEditor;
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
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.tab.TKTabbedPanel;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;

/** The detailed editor for {@link CMTechnique}s. */
public class CSTechniqueEditor extends CSRowEditor<CMTechnique> implements ActionListener {
	private TKCorrectableField		mNameField;
	private TKTextField				mNotesField;
	private TKTextField				mReferenceField;
	private TKPopupMenu				mDifficultyPopup;
	private TKTextField				mPointsField;
	private TKTextField				mLevelField;
	private TKPanel					mDefaultPanel;
	private TKCorrectableLabel		mDefaultPanelLabel;
	private TKPopupMenu				mDefaultTypePopup;
	private TKCorrectableField		mDefaultNameField;
	private TKTextField				mDefaultSpecializationField;
	private TKTextField				mDefaultModifierField;
	private TKCheckbox				mLimitCheckbox;
	private TKTextField				mLimitField;
	private TKTabbedPanel			mTabPanel;
	private CSPrereqs				mPrereqs;
	private CSFeatures				mFeatures;
	private CMSkillDefaultType		mLastDefaultType;
	private CSMeleeWeaponEditor		mMeleeWeapons;
	private CSRangedWeaponEditor	mRangedWeapons;

	/**
	 * Creates a new {@link CMTechnique} editor.
	 * 
	 * @param technique The {@link CMTechnique} to edit.
	 */
	public CSTechniqueEditor(CMTechnique technique) {
		super(technique);

		TKPanel content = new TKPanel(new TKColumnLayout(2));
		TKPanel fields = new TKPanel(new TKColumnLayout(2));
		TKLabel icon = new TKLabel(technique.getImage(true));
		TKPanel wrapper;

		mNameField = createCorrectableField(fields, fields, Msgs.NAME, technique.getName(), Msgs.TECHNIQUE_NAME_TOOLTIP);
		mNotesField = createField(fields, fields, Msgs.NOTES, technique.getNotes(), Msgs.TECHNIQUE_NOTES_TOOLTIP, 0);
		createDefaults(fields);
		createLimits(fields);
		wrapper = createDifficultyPopups(fields);
		mReferenceField = createField(wrapper, wrapper, Msgs.EDITOR_REFERENCE, mRow.getReference(), Msgs.TECHNIQUE_REFERENCE_TOOLTIP, 6);
		icon.setVerticalAlignment(TKAlignment.TOP);
		icon.setAlignmentY(-1f);
		content.add(icon);
		content.add(fields);
		add(content);

		ArrayList<TKPanel> panels = new ArrayList<TKPanel>();

		mPrereqs = new CSPrereqs(mRow, mRow.getPrereqs());
		mFeatures = new CSFeatures(mRow, mRow.getFeatures());
		mMeleeWeapons = CSMeleeWeaponEditor.createEditor(mRow);
		mRangedWeapons = CSRangedWeaponEditor.createEditor(mRow);
		panels.add(embedEditor(mPrereqs));
		panels.add(embedEditor(mFeatures));
		panels.add(mMeleeWeapons);
		panels.add(mRangedWeapons);
		mTabPanel = new TKTabbedPanel(panels);
		mTabPanel.setSelectedPanelByName(getLastTabName());
		add(mTabPanel);
	}

	private void createDefaults(TKPanel parent) {
		TKMenu menu = new TKMenu();

		for (CMSkillDefaultType type : CMSkillDefaultType.values()) {
			TKMenuItem item = new TKMenuItem(type.toString());

			item.setUserObject(type);
			menu.add(item);
		}
		mDefaultPanel = new TKPanel(new TKColumnLayout(4));
		mDefaultPanelLabel = new TKCorrectableLabel(Msgs.DEFAULTS_TO);
		mDefaultTypePopup = new TKPopupMenu(menu);
		mDefaultTypePopup.setSelectedUserObject(mRow.getDefault().getType());
		mDefaultTypePopup.setOnlySize(mDefaultTypePopup.getPreferredSize());
		mDefaultTypePopup.setEnabled(mIsEditable);
		mDefaultTypePopup.addActionListener(this);
		mDefaultPanel.add(mDefaultTypePopup);

		parent.add(mDefaultPanelLabel);
		parent.add(mDefaultPanel);
		rebuildDefaultPanel();
	}

	private CMSkillDefaultType getDefaultType() {
		return (CMSkillDefaultType) mDefaultTypePopup.getSelectedItemUserObject();
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
		CMSkillDefault def = mRow.getDefault();
		boolean skillBased;

		mLastDefaultType = getDefaultType();
		skillBased = mLastDefaultType == CMSkillDefaultType.Skill;
		forceFocusToAccept();
		while (mDefaultPanel.getComponentCount() > 1) {
			mDefaultPanel.remove(1);
		}
		if (skillBased) {
			mDefaultNameField = createCorrectableField(null, mDefaultPanel, Msgs.DEFAULTS_TO, def.getName(), Msgs.DEFAULTS_TO_TOOLTIP);
			mDefaultSpecializationField = createField(null, mDefaultPanel, null, def.getSpecialization(), Msgs.DEFAULT_SPECIALIZATION_TOOLTIP, 0);
			mDefaultSpecializationField.setImprint(Msgs.SPECIALIZATION_IMPRINT);
			mDefaultNameField.setLabel(mDefaultPanelLabel);
		}
		mDefaultModifierField = createNumberField(null, mDefaultPanel, null, Msgs.DEFAULT_MODIFIER_TOOLTIP, def.getModifier(), 2);
		if (!skillBased) {
			mDefaultPanel.add(new TKPanel());
			mDefaultPanel.add(new TKPanel());
		}
		mDefaultPanel.revalidate();
	}

	private void createLimits(TKPanel parent) {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(3));

		mLimitCheckbox = new TKCheckbox(Msgs.LIMIT, mRow.isLimited());
		mLimitCheckbox.setFontKey(TKFont.TEXT_FONT_KEY);
		mLimitCheckbox.setToolTipText(Msgs.LIMIT_TOOLTIP);
		mLimitCheckbox.addActionListener(this);
		mLimitCheckbox.setEnabled(mIsEditable);

		mLimitField = createNumberField(null, wrapper, null, Msgs.LIMIT_AMOUNT_TOOLTIP, mRow.getLimitModifier(), 2);
		mLimitField.setEnabled(mIsEditable && mLimitCheckbox.isChecked());
		mLimitField.addActionListener(this);

		wrapper.add(mLimitCheckbox);
		wrapper.add(mLimitField);
		wrapper.add(new TKPanel());
		parent.add(new TKLabel());
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

	private TKCorrectableField createCorrectableField(TKPanel labelParent, TKPanel fieldParent, String title, String text, String tooltip) {
		TKCorrectableLabel label = new TKCorrectableLabel(title);
		TKCorrectableField field = new TKCorrectableField(label, text);

		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.addActionListener(this);
		if (labelParent != null) {
			labelParent.add(label);
		}
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
		field.addActionListener(this);
		if (labelParent != null) {
			labelParent.add(new TKLinkedLabel(field, title));
		}
		fieldParent.add(field);
		return field;
	}

	private TKTextField createNumberField(TKPanel labelParent, TKPanel fieldParent, String title, String tooltip, int value, int maxDigits) {
		TKTextField field = createField(labelParent, fieldParent, title, TKNumberUtils.format(value, true), tooltip, maxDigits + 1);

		field.setKeyEventFilter(new TKNumberFilter(false, true, false, maxDigits));
		return field;
	}

	private void createPointsFields(TKPanel parent, boolean forCharacter) {
		mPointsField = createField(parent, parent, Msgs.EDITOR_POINTS, Integer.toString(mRow.getPoints()), Msgs.TECHNIQUE_POINTS_TOOLTIP, 4);
		mPointsField.setKeyEventFilter(new TKNumberFilter(false, false, false, 4));
		mPointsField.addActionListener(this);

		if (forCharacter) {
			mLevelField = createField(parent, parent, Msgs.EDITOR_LEVEL, CMTechnique.getTechniqueDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel(), mRow.getDefault().getModifier()), Msgs.EDITOR_LEVEL_TOOLTIP, 6);
			mLevelField.setEnabled(false);
		}
	}

	private TKPanel createDifficultyPopups(TKPanel parent) {
		CMCharacter character = mRow.getCharacter();
		boolean forCharacterOrTemplate = character != null || mRow.getTemplate() != null;
		TKLabel label = new TKLabel(Msgs.EDITOR_DIFFICULTY, TKAlignment.RIGHT);
		TKPanel wrapper = new TKPanel(new TKColumnLayout(forCharacterOrTemplate ? character != null ? 8 : 6 : 4));
		TKMenu menu = new TKMenu();
		TKMenuItem item = new TKMenuItem(CMSkillDifficulty.A.toString());

		label.setToolTipText(Msgs.TECHNIQUE_DIFFICULTY_TOOLTIP);

		item.setUserObject(CMSkillDifficulty.A);
		menu.add(item);
		item = new TKMenuItem(CMSkillDifficulty.H.toString());
		item.setUserObject(CMSkillDifficulty.H);
		menu.add(item);

		mDifficultyPopup = new TKPopupMenu(menu);
		mDifficultyPopup.setSelectedUserObject(mRow.getDifficulty());
		mDifficultyPopup.setOnlySize(mDifficultyPopup.getPreferredSize());
		mDifficultyPopup.setToolTipText(Msgs.TECHNIQUE_DIFFICULTY_POPUP_TOOLTIP);
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
			CMSkillLevel level = CMTechnique.calculateTechniqueLevel(mRow.getCharacter(), mNameField.getText(), getSpecialization(), createNewDefault(), getSkillDifficulty(), getPoints(), mLimitCheckbox.isChecked(), getLimitModifier());

			mLevelField.setText(CMTechnique.getTechniqueDisplayLevel(level.mLevel, level.mRelativeLevel, getDefaultModifier()));
		}
	}

	private CMSkillDefault createNewDefault() {
		CMSkillDefaultType type = getDefaultType();

		if (type == CMSkillDefaultType.Skill) {
			return new CMSkillDefault(type, mDefaultNameField.getText(), mDefaultSpecializationField.getText(), getDefaultModifier());
		}
		return new CMSkillDefault(type, null, null, getDefaultModifier());
	}

	private CMSkillDifficulty getSkillDifficulty() {
		return (CMSkillDifficulty) mDifficultyPopup.getSelectedItemUserObject();
	}

	private int getPoints() {
		return TKNumberUtils.getInteger(mPointsField.getText(), 0);
	}

	private int getDefaultModifier() {
		return TKNumberUtils.getInteger(mDefaultModifierField.getText(), 0);
	}

	private int getLimitModifier() {
		return TKNumberUtils.getInteger(mLimitField.getText(), 0);
	}

	@Override public boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());

		modified |= mRow.setDefault(createNewDefault());
		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setNotes(mNotesField.getText());
		if (mPointsField != null) {
			modified |= mRow.setPoints(getPoints());
		}
		modified |= mRow.setLimited(mLimitCheckbox.isChecked());
		modified |= mRow.setLimitModifier(getLimitModifier());
		modified |= mRow.setDifficulty(getSkillDifficulty());
		modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
		modified |= mRow.setFeatures(mFeatures.getFeatures());

		ArrayList<CMWeaponStats> list = new ArrayList<CMWeaponStats>(mMeleeWeapons.getWeapons());
		list.addAll(mRangedWeapons.getWeapons());
		modified |= mRow.setWeapons(list);

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
				mNameField.correct(corrected, Msgs.TECHNIQUE_NAME_CANNOT_BE_EMPTY);
			}
		} else if (src == mDefaultNameField) {
			if (mDefaultNameField.getText().trim().length() == 0) {
				String corrected = mRow.getDefault().getName().trim();

				if (corrected.length() == 0) {
					corrected = mRow.getLocalizedName();
				}
				mDefaultNameField.correct(corrected, Msgs.DEFAULT_NAME_CANNOT_BE_EMPTY);
			}
		} else if (src == mLimitCheckbox) {
			mLimitField.setEnabled(mLimitCheckbox.isChecked());
		} else if (src == mDefaultTypePopup) {
			if (mLastDefaultType != getDefaultType()) {
				rebuildDefaultPanel();
			}
		}

		if (src == mDifficultyPopup || src == mPointsField || src == mDefaultNameField || src == mDefaultModifierField || src == mLimitCheckbox || src == mLimitField || src == mDefaultSpecializationField || src == mDefaultTypePopup) {
			recalculateLevel();
		}
	}
}
