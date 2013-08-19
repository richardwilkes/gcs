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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
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
import com.trollworks.ttk.widgets.outline.OutlineModel;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashSet;

import javax.swing.ImageIcon;
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

/** The detailed editor for {@link Skill}s. */
public class SkillEditor extends RowEditor<Skill> implements ActionListener, DocumentListener {
	private static String		MSG_NAME;
	private static String		MSG_NAME_TOOLTIP;
	private static String		MSG_NAME_CANNOT_BE_EMPTY;
	private static String		MSG_SPECIALIZATION;
	private static String		MSG_SPECIALIZATION_TOOLTIP;
	private static String		MSG_CATEGORIES;
	private static String		MSG_CATEGORIES_TOOLTIP;
	private static String		MSG_NOTES;
	private static String		MSG_NOTES_TOOLTIP;
	private static String		MSG_TECH_LEVEL;
	private static String		MSG_TECH_LEVEL_TOOLTIP;
	private static String		MSG_TECH_LEVEL_REQUIRED;
	private static String		MSG_TECH_LEVEL_REQUIRED_TOOLTIP;
	private static String		MSG_EDITOR_DIFFICULTY;
	private static String		MSG_EDITOR_DIFFICULTY_TOOLTIP;
	private static String		MSG_EDITOR_DIFFICULTY_POPUP_TOOLTIP;
	private static String		MSG_EDITOR_LEVEL;
	private static String		MSG_EDITOR_LEVEL_TOOLTIP;
	private static String		MSG_ATTRIBUTE_POPUP_TOOLTIP;
	private static String		MSG_EDITOR_POINTS;
	private static String		MSG_EDITOR_POINTS_TOOLTIP;
	private static String		MSG_EDITOR_REFERENCE;
	private static String		MSG_REFERENCE_TOOLTIP;
	private static String		MSG_ENC_PENALTY_MULT;
	private static String		MSG_ENC_PENALTY_MULT_TOOLTIP;
	private static String		MSG_NO_ENC_PENALTY;
	private static String		MSG_ONE_ENC_PENALTY;
	private static String		MSG_ENC_PENALTY_FORMAT;
	private JTextField			mNameField;
	private JTextField			mSpecializationField;
	private JTextField			mNotesField;
	private JTextField			mCategoriesField;
	private JTextField			mReferenceField;
	private JCheckBox			mHasTechLevel;
	private JTextField			mTechLevel;
	private String				mSavedTechLevel;
	private JComboBox			mAttributePopup;
	private JComboBox			mDifficultyPopup;
	private JTextField			mPointsField;
	private JTextField			mLevelField;
	private JComboBox			mEncPenaltyPopup;
	private JTabbedPane			mTabPanel;
	private PrereqsPanel		mPrereqs;
	private FeaturesPanel		mFeatures;
	private Defaults			mDefaults;
	private MeleeWeaponEditor	mMeleeWeapons;
	private RangedWeaponEditor	mRangedWeapons;

	static {
		LocalizedMessages.initialize(SkillEditor.class);
	}

	/**
	 * Creates a new {@link Skill} editor.
	 * 
	 * @param skill The {@link Skill} to edit.
	 */
	public SkillEditor(Skill skill) {
		super(skill);

		JPanel content = new JPanel(new ColumnLayout(2));
		JPanel fields = new JPanel(new ColumnLayout(2));
		JLabel icon = new JLabel(new ImageIcon(skill.getImage(true)));
		boolean notContainer = !skill.canHaveChildren();
		Container wrapper;

		mNameField = createCorrectableField(fields, MSG_NAME, skill.getName(), MSG_NAME_TOOLTIP);
		if (notContainer) {
			wrapper = new JPanel(new ColumnLayout(2));
			mSpecializationField = createField(fields, wrapper, MSG_SPECIALIZATION, skill.getSpecialization(), MSG_SPECIALIZATION_TOOLTIP, 0);
			createTechLevelFields(wrapper);
			fields.add(wrapper);
			mEncPenaltyPopup = createEncumbrancePenaltyMultiplierPopup(fields);
		}
		mNotesField = createField(fields, fields, MSG_NOTES, skill.getNotes(), MSG_NOTES_TOOLTIP, 0);
		mCategoriesField = createField(fields, fields, MSG_CATEGORIES, skill.getCategoriesAsString(), MSG_CATEGORIES_TOOLTIP, 0);
		if (notContainer) {
			wrapper = createDifficultyPopups(fields);
		} else {
			wrapper = fields;
		}
		mReferenceField = createField(wrapper, wrapper, MSG_EDITOR_REFERENCE, mRow.getReference(), MSG_REFERENCE_TOOLTIP, 6);
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
			mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
			mDefaults = new Defaults(mRow.getDefaults());
			mDefaults.addActionListener(this);
			Component panel = embedEditor(mDefaults);
			mTabPanel.addTab(panel.getName(), panel);
			panel = embedEditor(mPrereqs);
			mTabPanel.addTab(panel.getName(), panel);
			panel = embedEditor(mFeatures);
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

	private void createPointsFields(Container parent, boolean forCharacter) {
		mPointsField = createField(parent, parent, MSG_EDITOR_POINTS, Integer.toString(mRow.getPoints()), MSG_EDITOR_POINTS_TOOLTIP, 4);
		new NumberFilter(mPointsField, false, false, false, 4);
		mPointsField.addActionListener(this);

		if (forCharacter) {
			mLevelField = createField(parent, parent, MSG_EDITOR_LEVEL, Skill.getSkillDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel(), mRow.getAttribute(), mRow.canHaveChildren()), MSG_EDITOR_LEVEL_TOOLTIP, 8);
			mLevelField.setEnabled(false);
		}
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

			mHasTechLevel = new JCheckBox(MSG_TECH_LEVEL, hasTL);
			mHasTechLevel.setToolTipText(MSG_TECH_LEVEL_TOOLTIP);
			mHasTechLevel.setEnabled(enabled);
			mHasTechLevel.addActionListener(this);
			wrapper.add(mHasTechLevel);

			mTechLevel = new JTextField("9999"); //$NON-NLS-1$
			UIUtilities.setOnlySize(mTechLevel, mTechLevel.getPreferredSize());
			mTechLevel.setText(mSavedTechLevel);
			mTechLevel.setToolTipText(MSG_TECH_LEVEL_TOOLTIP);
			mTechLevel.setEnabled(enabled && hasTL);
			wrapper.add(mTechLevel);
			parent.add(wrapper);

			if (!hasTL) {
				mSavedTechLevel = character.getDescription().getTechLevel();
			}
		} else {
			mTechLevel = new JTextField(mSavedTechLevel);
			mHasTechLevel = new JCheckBox(MSG_TECH_LEVEL_REQUIRED, hasTL);
			mHasTechLevel.setToolTipText(MSG_TECH_LEVEL_REQUIRED_TOOLTIP);
			mHasTechLevel.setEnabled(enabled);
			mHasTechLevel.addActionListener(this);
			parent.add(mHasTechLevel);
		}
	}

	private JComboBox createEncumbrancePenaltyMultiplierPopup(Container parent) {
		Object[] items = new Object[10];
		items[0] = MSG_NO_ENC_PENALTY;
		items[1] = MSG_ONE_ENC_PENALTY;
		for (int i = 2; i < 10; i++) {
			items[i] = MessageFormat.format(MSG_ENC_PENALTY_FORMAT, new Integer(i));
		}
		LinkedLabel label = new LinkedLabel(MSG_ENC_PENALTY_MULT);
		parent.add(label);
		JComboBox popup = createComboBox(parent, items, items[mRow.getEncumbrancePenaltyMultiplier()], MSG_ENC_PENALTY_MULT_TOOLTIP);
		label.setLink(popup);
		return popup;
	}

	private Container createDifficultyPopups(Container parent) {
		GURPSCharacter character = mRow.getCharacter();
		boolean forCharacterOrTemplate = character != null || mRow.getTemplate() != null;
		JLabel label = new JLabel(MSG_EDITOR_DIFFICULTY, SwingConstants.RIGHT);
		JPanel wrapper = new JPanel(new ColumnLayout(forCharacterOrTemplate ? character != null ? 10 : 8 : 6));

		label.setToolTipText(MSG_EDITOR_DIFFICULTY_TOOLTIP);

		mAttributePopup = createComboBox(wrapper, SkillAttribute.values(), mRow.getAttribute(), MSG_ATTRIBUTE_POPUP_TOOLTIP);
		wrapper.add(new JLabel(" /")); //$NON-NLS-1$
		mDifficultyPopup = createComboBox(wrapper, SkillDifficulty.values(), mRow.getDifficulty(), MSG_EDITOR_DIFFICULTY_POPUP_TOOLTIP);

		if (forCharacterOrTemplate) {
			createPointsFields(wrapper, character != null);
		}
		wrapper.add(new JPanel());

		parent.add(label);
		parent.add(wrapper);
		return wrapper;
	}

	private JComboBox createComboBox(Container parent, Object[] items, Object selection, String tooltip) {
		JComboBox combo = new JComboBox(items);
		combo.setToolTipText(tooltip);
		combo.setSelectedItem(selection);
		combo.addActionListener(this);
		combo.setMaximumRowCount(items.length);
		UIUtilities.setOnlySize(combo, combo.getPreferredSize());
		combo.setEnabled(mIsEditable);
		parent.add(combo);
		return combo;
	}

	private void recalculateLevel() {
		if (mLevelField != null) {
			SkillAttribute attribute = getSkillAttribute();
			SkillLevel level = Skill.calculateLevel(mRow.getCharacter(), mRow, mNameField.getText(), mSpecializationField.getText(), mDefaults.getDefaults(), attribute, getSkillDifficulty(), getSkillPoints(), new HashSet<String>(), getEncumbrancePenaltyMultiplier());
			mLevelField.setText(Skill.getSkillDisplayLevel(level.mLevel, level.mRelativeLevel, attribute, false));
		}
	}

	private SkillAttribute getSkillAttribute() {
		return (SkillAttribute) mAttributePopup.getSelectedItem();
	}

	private SkillDifficulty getSkillDifficulty() {
		return (SkillDifficulty) mDifficultyPopup.getSelectedItem();
	}

	private int getSkillPoints() {
		return Numbers.getLocalizedInteger(mPointsField.getText(), 0);
	}

	private int getEncumbrancePenaltyMultiplier() {
		return mEncPenaltyPopup.getSelectedIndex();
	}

	@Override
	public boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());
		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setNotes(mNotesField.getText());
		modified |= mRow.setCategories(mCategoriesField.getText());
		if (mSpecializationField != null) {
			modified |= mRow.setSpecialization(mSpecializationField.getText());
		}
		if (mHasTechLevel != null) {
			modified |= mRow.setTechLevel(mHasTechLevel.isSelected() ? mTechLevel.getText() : null);
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
		} else if (src == mAttributePopup || src == mDifficultyPopup || src == mPointsField || src == mDefaults || src == mEncPenaltyPopup) {
			recalculateLevel();
		}
	}

	public void changedUpdate(DocumentEvent event) {
		nameChanged();
	}

	public void insertUpdate(DocumentEvent event) {
		nameChanged();
	}

	public void removeUpdate(DocumentEvent event) {
		nameChanged();
	}

	private void nameChanged() {
		LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().length() != 0 ? null : MSG_NAME_CANNOT_BE_EMPTY);
	}
}
