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

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
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
import java.util.ArrayList;

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
import javax.swing.text.Document;

/** The detailed editor for {@link Spell}s. */
public class SpellEditor extends RowEditor<Spell> implements ActionListener, DocumentListener {
	private static String		MSG_NAME;
	private static String		MSG_NAME_TOOLTIP;
	private static String		MSG_NAME_CANNOT_BE_EMPTY;
	private static String		MSG_TECH_LEVEL;
	private static String		MSG_TECH_LEVEL_TOOLTIP;
	private static String		MSG_TECH_LEVEL_REQUIRED;
	private static String		MSG_TECH_LEVEL_REQUIRED_TOOLTIP;
	private static String		MSG_COLLEGE;
	private static String		MSG_COLLEGE_TOOLTIP;
	private static String		MSG_POWER_SOURCE;
	private static String		MSG_POWER_SOURCE_TOOLTIP;
	private static String		MSG_CLASS;
	private static String		MSG_CLASS_ONLY_TOOLTIP;
	private static String		MSG_CLASS_CANNOT_BE_EMPTY;
	private static String		MSG_CASTING_COST;
	private static String		MSG_CASTING_COST_TOOLTIP;
	private static String		MSG_CASTING_COST_CANNOT_BE_EMPTY;
	private static String		MSG_MAINTENANCE_COST;
	private static String		MSG_MAINTENANCE_COST_TOOLTIP;
	private static String		MSG_CASTING_TIME;
	private static String		MSG_CASTING_TIME_TOOLTIP;
	private static String		MSG_CASTING_TIME_CANNOT_BE_EMPTY;
	private static String		MSG_DURATION;
	private static String		MSG_DURATION_TOOLTIP;
	private static String		MSG_DURATION_CANNOT_BE_EMPTY;
	private static String		MSG_CATEGORIES;
	private static String		MSG_CATEGORIES_TOOLTIP;
	private static String		MSG_NOTES;
	private static String		MSG_NOTES_TOOLTIP;
	private static String		MSG_EDITOR_POINTS;
	private static String		MSG_EDITOR_POINTS_TOOLTIP;
	private static String		MSG_EDITOR_LEVEL;
	private static String		MSG_EDITOR_LEVEL_TOOLTIP;
	private static String		MSG_DIFFICULTY;
	private static String		MSG_DIFFICULTY_TOOLTIP;
	private static String		MSG_EDITOR_REFERENCE;
	private static String		MSG_REFERENCE_TOOLTIP;
	private JTextField			mNameField;
	private JTextField			mCollegeField;
	private JTextField			mPowerSourceField;
	private JTextField			mClassField;
	private JTextField			mCastingCostField;
	private JTextField			mMaintenanceField;
	private JTextField			mCastingTimeField;
	private JTextField			mDurationField;
	private JComboBox			mDifficultyCombo;
	private JTextField			mNotesField;
	private JTextField			mCategoriesField;
	private JTextField			mPointsField;
	private JTextField			mLevelField;
	private JTextField			mReferenceField;
	private JTabbedPane			mTabPanel;
	private PrereqsPanel		mPrereqs;
	private JCheckBox			mHasTechLevel;
	private JTextField			mTechLevel;
	private String				mSavedTechLevel;
	private MeleeWeaponEditor	mMeleeWeapons;
	private RangedWeaponEditor	mRangedWeapons;

	static {
		LocalizedMessages.initialize(SpellEditor.class);
	}

	/**
	 * Creates a new {@link Spell} editor.
	 * 
	 * @param spell The {@link Spell} to edit.
	 */
	public SpellEditor(Spell spell) {
		super(spell);

		boolean notContainer = !spell.canHaveChildren();
		Container content = new JPanel(new ColumnLayout(2));
		Container fields = new JPanel(new ColumnLayout());
		Container wrapper1 = new JPanel(new ColumnLayout(notContainer ? 3 : 2));
		Container wrapper2 = new JPanel(new ColumnLayout(4));
		Container wrapper3 = new JPanel(new ColumnLayout(2));
		Container noGapWrapper = new JPanel(new ColumnLayout(2, 0, 0));
		Container ptsPanel = null;
		JLabel icon = new JLabel(new ImageIcon(spell.getImage(true)));
		Dimension size = new Dimension();
		Container refParent = wrapper3;

		mNameField = createCorrectableField(wrapper1, wrapper1, MSG_NAME, spell.getName(), MSG_NAME_TOOLTIP);
		fields.add(wrapper1);
		if (notContainer) {
			createTechLevelFields(wrapper1);
			mCollegeField = createField(wrapper2, wrapper2, MSG_COLLEGE, spell.getCollege(), MSG_COLLEGE_TOOLTIP, 0);
			mPowerSourceField = createField(wrapper2, wrapper2, MSG_POWER_SOURCE, spell.getPowerSource(), MSG_POWER_SOURCE_TOOLTIP, 0);
			mClassField = createCorrectableField(wrapper2, wrapper2, MSG_CLASS, spell.getSpellClass(), MSG_CLASS_ONLY_TOOLTIP);
			mCastingCostField = createCorrectableField(wrapper2, wrapper2, MSG_CASTING_COST, spell.getCastingCost(), MSG_CASTING_COST_TOOLTIP);
			mMaintenanceField = createField(wrapper2, wrapper2, MSG_MAINTENANCE_COST, spell.getMaintenance(), MSG_MAINTENANCE_COST_TOOLTIP, 0);
			mCastingTimeField = createCorrectableField(wrapper2, wrapper2, MSG_CASTING_TIME, spell.getCastingTime(), MSG_CASTING_TIME_TOOLTIP);
			mDurationField = createCorrectableField(wrapper2, wrapper2, MSG_DURATION, spell.getDuration(), MSG_DURATION_TOOLTIP);
			fields.add(wrapper2);

			ptsPanel = createPointsFields();
			fields.add(ptsPanel);
			refParent = ptsPanel;
		}
		mNotesField = createField(wrapper3, wrapper3, MSG_NOTES, spell.getNotes(), MSG_NOTES_TOOLTIP, 0);
		mCategoriesField = createField(wrapper3, wrapper3, MSG_CATEGORIES, spell.getCategoriesAsString(), MSG_CATEGORIES_TOOLTIP, 0);
		mReferenceField = createField(refParent, noGapWrapper, MSG_EDITOR_REFERENCE, mRow.getReference(), MSG_REFERENCE_TOOLTIP, 6);
		noGapWrapper.add(new JPanel());
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
			Component panel = embedEditor(mPrereqs);
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

	private void determineLargest(Container panel, int every, Dimension size) {
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

	private void applySize(Container panel, int every, Dimension size) {
		int count = panel.getComponentCount();

		for (int i = 0; i < count; i += every) {
			UIUtilities.setOnlySize(panel.getComponent(i), size);
		}
	}

	private JScrollPane embedEditor(Component editor) {
		JScrollPane scrollPanel = new JScrollPane(editor);

		scrollPanel.setMinimumSize(new Dimension(500, 120));
		scrollPanel.setName(editor.toString());
		if (!mIsEditable) {
			UIUtilities.disableControls(editor);
		}
		return scrollPanel;
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

	private Container createPointsFields() {
		boolean forCharacter = mRow.getCharacter() != null;
		boolean forTemplate = mRow.getTemplate() != null;
		JPanel panel = new JPanel(new ColumnLayout(forCharacter ? 8 : forTemplate ? 6 : 4));

		mDifficultyCombo = new JComboBox(new Object[] { SkillDifficulty.H.name(), SkillDifficulty.VH.name() });
		mDifficultyCombo.setSelectedIndex(mRow.isVeryHard() ? 1 : 0);
		mDifficultyCombo.setToolTipText(MSG_DIFFICULTY_TOOLTIP);
		UIUtilities.setOnlySize(mDifficultyCombo, mDifficultyCombo.getPreferredSize());
		mDifficultyCombo.addActionListener(this);
		mDifficultyCombo.setEnabled(mIsEditable);
		panel.add(new LinkedLabel(MSG_DIFFICULTY, mDifficultyCombo));
		panel.add(mDifficultyCombo);

		if (forCharacter || mRow.getTemplate() != null) {
			mPointsField = createField(panel, panel, MSG_EDITOR_POINTS, Integer.toString(mRow.getPoints()), MSG_EDITOR_POINTS_TOOLTIP, 4);
			new NumberFilter(mPointsField, false, false, false, 4);
			mPointsField.addActionListener(this);

			if (forCharacter) {
				mLevelField = createField(panel, panel, MSG_EDITOR_LEVEL, getDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel()), MSG_EDITOR_LEVEL_TOOLTIP, 5);
				mLevelField.setEnabled(false);
			}
		}
		return panel;
	}

	private String getDisplayLevel(int level, int relativeLevel) {
		if (level < 0) {
			return "-"; //$NON-NLS-1$
		}
		return Numbers.format(level) + "/IQ" + Numbers.formatWithForcedSign(relativeLevel); //$NON-NLS-1$
	}

	private JTextField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
		JTextField field = new JTextField(text);
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.getDocument().addDocumentListener(this);

		LinkedLabel label = new LinkedLabel(title);
		label.setLink(field);

		labelParent.add(label);
		fieldParent.add(field);
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

	@Override
	public boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());
		boolean notContainer = !mRow.canHaveChildren();

		modified |= mRow.setReference(mReferenceField.getText());
		if (notContainer) {
			if (mHasTechLevel != null) {
				modified |= mRow.setTechLevel(mHasTechLevel.isSelected() ? mTechLevel.getText() : null);
			}
			modified |= mRow.setCollege(mCollegeField.getText());
			modified |= mRow.setPowerSource(mPowerSourceField.getText());
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
		modified |= mRow.setCategories(mCategoriesField.getText());
		if (mPrereqs != null) {
			modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
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

	@Override
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
		} else if (src == mPointsField || src == mDifficultyCombo) {
			recalculateLevel();
		}
	}

	private void recalculateLevel() {
		if (mLevelField != null) {
			SkillLevel level = Spell.calculateLevel(mRow.getCharacter(), getSpellPoints(), isVeryHard(), mCollegeField.getText(), mPowerSourceField.getText(), mNameField.getText());
			mLevelField.setText(getDisplayLevel(level.mLevel, level.mRelativeLevel));
		}
	}

	private int getSpellPoints() {
		return Numbers.getLocalizedInteger(mPointsField.getText(), 0);
	}

	private boolean isVeryHard() {
		return mDifficultyCombo.getSelectedIndex() == 1;
	}

	@Override
	public void changedUpdate(DocumentEvent event) {
		Document doc = event.getDocument();
		if (doc == mNameField.getDocument()) {
			LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().length() != 0 ? null : MSG_NAME_CANNOT_BE_EMPTY);
		} else if (doc == mClassField.getDocument()) {
			LinkedLabel.setErrorMessage(mClassField, mClassField.getText().trim().length() != 0 ? null : MSG_CLASS_CANNOT_BE_EMPTY);
		} else if (doc == mClassField.getDocument()) {
			LinkedLabel.setErrorMessage(mCastingCostField, mCastingCostField.getText().trim().length() != 0 ? null : MSG_CASTING_COST_CANNOT_BE_EMPTY);
		} else if (doc == mClassField.getDocument()) {
			LinkedLabel.setErrorMessage(mCastingTimeField, mCastingTimeField.getText().trim().length() != 0 ? null : MSG_CASTING_TIME_CANNOT_BE_EMPTY);
		} else if (doc == mClassField.getDocument()) {
			LinkedLabel.setErrorMessage(mDurationField, mDurationField.getText().trim().length() != 0 ? null : MSG_DURATION_CANNOT_BE_EMPTY);
		}
	}

	@Override
	public void insertUpdate(DocumentEvent event) {
		changedUpdate(event);
	}

	@Override
	public void removeUpdate(DocumentEvent event) {
		changedUpdate(event);
	}
}
