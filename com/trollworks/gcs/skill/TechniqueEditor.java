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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.text.NumberUtils;
import com.trollworks.gcs.utility.text.TextUtility;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.CommitEnforcer;
import com.trollworks.gcs.widgets.LinkedLabel;
import com.trollworks.gcs.widgets.NumberFilter;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.layout.ColumnLayout;
import com.trollworks.gcs.widgets.outline.RowEditor;

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

/** The detailed editor for {@link Technique}s. */
public class TechniqueEditor extends RowEditor<Technique> implements ActionListener, DocumentListener {
	private static String		MSG_NAME;
	private static String		MSG_NOTES;
	private static String		MSG_EDITOR_REFERENCE;
	private static String		MSG_EDITOR_POINTS;
	private static String		MSG_EDITOR_LEVEL;
	private static String		MSG_EDITOR_LEVEL_TOOLTIP;
	private static String		MSG_EDITOR_DIFFICULTY;
	private static String		MSG_TECHNIQUE_NAME_TOOLTIP;
	private static String		MSG_TECHNIQUE_NAME_CANNOT_BE_EMPTY;
	private static String		MSG_TECHNIQUE_NOTES_TOOLTIP;
	private static String		MSG_TECHNIQUE_DIFFICULTY_TOOLTIP;
	private static String		MSG_TECHNIQUE_DIFFICULTY_POPUP_TOOLTIP;
	private static String		MSG_TECHNIQUE_POINTS_TOOLTIP;
	private static String		MSG_TECHNIQUE_REFERENCE_TOOLTIP;
	private static String		MSG_DEFAULTS_TO;
	private static String		MSG_DEFAULTS_TO_TOOLTIP;
	private static String		MSG_DEFAULT_NAME_CANNOT_BE_EMPTY;
	private static String		MSG_DEFAULT_SPECIALIZATION_TOOLTIP;
	private static String		MSG_DEFAULT_MODIFIER_TOOLTIP;
	private static String		MSG_LIMIT;
	private static String		MSG_LIMIT_TOOLTIP;
	private static String		MSG_LIMIT_AMOUNT_TOOLTIP;
	private static String		MSG_CATEGORIES;
	private static String		MSG_CATEGORIES_TOOLTIP;
	private JTextField			mNameField;
	private JTextField			mNotesField;
	private JTextField			mCategoriesField;
	private JTextField			mReferenceField;
	private JComboBox			mDifficultyCombo;
	private JTextField			mPointsField;
	private JTextField			mLevelField;
	private JPanel				mDefaultPanel;
	private LinkedLabel			mDefaultPanelLabel;
	private JComboBox			mDefaultTypeCombo;
	private JTextField			mDefaultNameField;
	private JTextField			mDefaultSpecializationField;
	private JTextField			mDefaultModifierField;
	private JCheckBox			mLimitCheckbox;
	private JTextField			mLimitField;
	private JTabbedPane			mTabPanel;
	private PrereqsPanel		mPrereqs;
	private FeaturesPanel		mFeatures;
	private SkillDefaultType	mLastDefaultType;
	private MeleeWeaponEditor	mMeleeWeapons;
	private RangedWeaponEditor	mRangedWeapons;

	static {
		LocalizedMessages.initialize(TechniqueEditor.class);
	}

	/**
	 * Creates a new {@link Technique} editor.
	 * 
	 * @param technique The {@link Technique} to edit.
	 */
	public TechniqueEditor(Technique technique) {
		super(technique);

		JPanel content = new JPanel(new ColumnLayout(2));
		JPanel fields = new JPanel(new ColumnLayout(2));
		JLabel icon = new JLabel(new ImageIcon(technique.getImage(true)));
		Container wrapper;

		mNameField = createCorrectableField(fields, fields, MSG_NAME, technique.getName(), MSG_TECHNIQUE_NAME_TOOLTIP);
		mNotesField = createField(fields, fields, MSG_NOTES, technique.getNotes(), MSG_TECHNIQUE_NOTES_TOOLTIP, 0);
		mCategoriesField = createField(fields, fields, MSG_CATEGORIES, technique.getCategoriesAsString(), MSG_CATEGORIES_TOOLTIP, 0);
		createDefaults(fields);
		createLimits(fields);
		wrapper = createDifficultyPopups(fields);
		mReferenceField = createField(wrapper, wrapper, MSG_EDITOR_REFERENCE, mRow.getReference(), MSG_TECHNIQUE_REFERENCE_TOOLTIP, 6);
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
		Component panel = embedEditor(mPrereqs);
		mTabPanel.addTab(panel.getName(), panel);
		panel = embedEditor(mFeatures);
		mTabPanel.addTab(panel.getName(), panel);
		mTabPanel.addTab(mMeleeWeapons.getName(), mMeleeWeapons);
		mTabPanel.addTab(mRangedWeapons.getName(), mRangedWeapons);
		UIUtilities.selectTab(mTabPanel, getLastTabName());
		add(mTabPanel);
	}

	private void createDefaults(Container parent) {
		mDefaultPanel = new JPanel(new ColumnLayout(4));
		mDefaultPanelLabel = new LinkedLabel(MSG_DEFAULTS_TO);
		mDefaultTypeCombo = createComboBox(mDefaultPanel, SkillDefaultType.values(), mRow.getDefault().getType());
		mDefaultTypeCombo.setEnabled(mIsEditable);

		parent.add(mDefaultPanelLabel);
		parent.add(mDefaultPanel);
		rebuildDefaultPanel();
	}

	private JComboBox createComboBox(Container parent, Object[] items, Object selection) {
		JComboBox combo = new JComboBox(items);
		combo.setSelectedItem(selection);
		combo.addActionListener(this);
		combo.setMaximumRowCount(items.length);
		UIUtilities.setOnlySize(combo, combo.getPreferredSize());
		parent.add(combo);
		return combo;
	}

	private SkillDefaultType getDefaultType() {
		return (SkillDefaultType) mDefaultTypeCombo.getSelectedItem();
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
		SkillDefault def = mRow.getDefault();
		boolean skillBased;

		mLastDefaultType = getDefaultType();
		skillBased = mLastDefaultType.isSkillBased();
		CommitEnforcer.forceFocusToAccept();
		while (mDefaultPanel.getComponentCount() > 1) {
			mDefaultPanel.remove(1);
		}
		if (skillBased) {
			mDefaultNameField = createCorrectableField(null, mDefaultPanel, MSG_DEFAULTS_TO, def.getName(), MSG_DEFAULTS_TO_TOOLTIP);
			mDefaultSpecializationField = createField(null, mDefaultPanel, null, def.getSpecialization(), MSG_DEFAULT_SPECIALIZATION_TOOLTIP, 0);
			mDefaultPanelLabel.setLink(mDefaultNameField);
		}
		mDefaultModifierField = createNumberField(null, mDefaultPanel, null, MSG_DEFAULT_MODIFIER_TOOLTIP, def.getModifier(), 2);
		if (!skillBased) {
			mDefaultPanel.add(new JPanel());
			mDefaultPanel.add(new JPanel());
		}
		mDefaultPanel.revalidate();
	}

	private void createLimits(Container parent) {
		JPanel wrapper = new JPanel(new ColumnLayout(3));

		mLimitCheckbox = new JCheckBox(MSG_LIMIT, mRow.isLimited());
		mLimitCheckbox.setToolTipText(MSG_LIMIT_TOOLTIP);
		mLimitCheckbox.addActionListener(this);
		mLimitCheckbox.setEnabled(mIsEditable);

		mLimitField = createNumberField(null, wrapper, null, MSG_LIMIT_AMOUNT_TOOLTIP, mRow.getLimitModifier(), 2);
		mLimitField.setEnabled(mIsEditable && mLimitCheckbox.isSelected());
		mLimitField.addActionListener(this);

		wrapper.add(mLimitCheckbox);
		wrapper.add(mLimitField);
		wrapper.add(new JPanel());
		parent.add(new JLabel());
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

	private JTextField createCorrectableField(Container labelParent, Container fieldParent, String title, String text, String tooltip) {
		JTextField field = new JTextField(text);
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		field.getDocument().addDocumentListener(this);

		if (labelParent != null) {
			LinkedLabel label = new LinkedLabel(title);
			label.setLink(field);
			labelParent.add(label);
		}

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
		field.addActionListener(this);
		if (labelParent != null) {
			labelParent.add(new LinkedLabel(title, field));
		}
		fieldParent.add(field);
		return field;
	}

	private JTextField createNumberField(Container labelParent, Container fieldParent, String title, String tooltip, int value, int maxDigits) {
		JTextField field = createField(labelParent, fieldParent, title, NumberUtils.format(value, true), tooltip, maxDigits + 1);
		new NumberFilter(field, false, true, false, maxDigits);
		return field;
	}

	private void createPointsFields(Container parent, boolean forCharacter) {
		mPointsField = createField(parent, parent, MSG_EDITOR_POINTS, Integer.toString(mRow.getPoints()), MSG_TECHNIQUE_POINTS_TOOLTIP, 4);
		new NumberFilter(mPointsField, false, false, false, 4);
		mPointsField.addActionListener(this);

		if (forCharacter) {
			mLevelField = createField(parent, parent, MSG_EDITOR_LEVEL, Technique.getTechniqueDisplayLevel(mRow.getLevel(), mRow.getRelativeLevel(), mRow.getDefault().getModifier()), MSG_EDITOR_LEVEL_TOOLTIP, 6);
			mLevelField.setEnabled(false);
		}
	}

	private Container createDifficultyPopups(Container parent) {
		GURPSCharacter character = mRow.getCharacter();
		boolean forCharacterOrTemplate = character != null || mRow.getTemplate() != null;
		JLabel label = new JLabel(MSG_EDITOR_DIFFICULTY, SwingConstants.RIGHT);
		JPanel wrapper = new JPanel(new ColumnLayout(forCharacterOrTemplate ? character != null ? 8 : 6 : 4));

		label.setToolTipText(MSG_TECHNIQUE_DIFFICULTY_TOOLTIP);

		mDifficultyCombo = createComboBox(wrapper, new Object[] { SkillDifficulty.A, SkillDifficulty.H }, mRow.getDifficulty());
		mDifficultyCombo.setToolTipText(MSG_TECHNIQUE_DIFFICULTY_POPUP_TOOLTIP);
		mDifficultyCombo.setEnabled(mIsEditable);

		if (forCharacterOrTemplate) {
			createPointsFields(wrapper, character != null);
		}
		wrapper.add(new JPanel());

		parent.add(label);
		parent.add(wrapper);
		return wrapper;
	}

	private void recalculateLevel() {
		if (mLevelField != null) {
			SkillLevel level = Technique.calculateTechniqueLevel(mRow.getCharacter(), mNameField.getText(), getSpecialization(), createNewDefault(), getSkillDifficulty(), getPoints(), mLimitCheckbox.isSelected(), getLimitModifier());

			mLevelField.setText(Technique.getTechniqueDisplayLevel(level.mLevel, level.mRelativeLevel, getDefaultModifier()));
		}
	}

	private SkillDefault createNewDefault() {
		SkillDefaultType type = getDefaultType();
		if (type.isSkillBased()) {
			return new SkillDefault(type, mDefaultNameField.getText(), mDefaultSpecializationField.getText(), getDefaultModifier());
		}
		return new SkillDefault(type, null, null, getDefaultModifier());
	}

	private SkillDifficulty getSkillDifficulty() {
		return (SkillDifficulty) mDifficultyCombo.getSelectedItem();
	}

	private int getPoints() {
		return NumberUtils.getInteger(mPointsField.getText(), 0);
	}

	private int getDefaultModifier() {
		return NumberUtils.getInteger(mDefaultModifierField.getText(), 0);
	}

	private int getLimitModifier() {
		return NumberUtils.getInteger(mLimitField.getText(), 0);
	}

	@Override public boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());

		modified |= mRow.setDefault(createNewDefault());
		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setNotes(mNotesField.getText());
		modified |= mRow.setCategories(mCategoriesField.getText());
		if (mPointsField != null) {
			modified |= mRow.setPoints(getPoints());
		}
		modified |= mRow.setLimited(mLimitCheckbox.isSelected());
		modified |= mRow.setLimitModifier(getLimitModifier());
		modified |= mRow.setDifficulty(getSkillDifficulty());
		modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
		modified |= mRow.setFeatures(mFeatures.getFeatures());

		ArrayList<WeaponStats> list = new ArrayList<WeaponStats>(mMeleeWeapons.getWeapons());
		list.addAll(mRangedWeapons.getWeapons());
		modified |= mRow.setWeapons(list);

		return modified;
	}

	@Override public void finished() {
		if (mTabPanel != null) {
			updateLastTabName(mTabPanel.getTitleAt(mTabPanel.getSelectedIndex()));
		}
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();

		if (src == mLimitCheckbox) {
			mLimitField.setEnabled(mLimitCheckbox.isSelected());
		} else if (src == mDefaultTypeCombo) {
			if (mLastDefaultType != getDefaultType()) {
				rebuildDefaultPanel();
			}
		}

		if (src == mDifficultyCombo || src == mPointsField || src == mDefaultNameField || src == mDefaultModifierField || src == mLimitCheckbox || src == mLimitField || src == mDefaultSpecializationField || src == mDefaultTypeCombo) {
			recalculateLevel();
		}
	}

	public void changedUpdate(DocumentEvent event) {
		Document doc = event.getDocument();
		if (doc == mNameField.getDocument()) {
			LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().length() != 0 ? null : MSG_TECHNIQUE_NAME_CANNOT_BE_EMPTY);
		} else if (doc == mDefaultNameField.getDocument()) {
			LinkedLabel.setErrorMessage(mDefaultNameField, mDefaultNameField.getText().trim().length() != 0 ? null : MSG_DEFAULT_NAME_CANNOT_BE_EMPTY);
		}
	}

	public void insertUpdate(DocumentEvent event) {
		changedUpdate(event);
	}

	public void removeUpdate(DocumentEvent event) {
		changedUpdate(event);
	}
}
