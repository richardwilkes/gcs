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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.text.NumberFilter;
import com.trollworks.ttk.text.NumberUtils;
import com.trollworks.ttk.text.TextUtility;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.LinkedLabel;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.text.MessageFormat;

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

/** Editor for {@link Modifier}s. */
public class ModifierEditor extends RowEditor<Modifier> implements ActionListener, DocumentListener {
	private static String		MSG_NAME;
	private static String		MSG_NAME_TOOLTIP;
	private static String		MSG_NOTES;
	private static String		MSG_NOTES_TOOLTIP;
	private static String		MSG_NAME_CANNOT_BE_EMPTY;
	private static String		MSG_COST;
	private static String		MSG_COST_TOOLTIP;
	private static String		MSG_LEVELS;
	private static String		MSG_LEVELS_TOOLTIP;
	private static String		MSG_TOTAL_COST;
	private static String		MSG_TOTAL_COST_TOOLTIP;
	private static String		MSG_HAS_LEVELS;
	private static String		MSG_ENABLED;
	private static String		MSG_ENABLED_TOOLTIP;
	private static String		MSG_REFERENCE;
	private static String		MSG_REFERENCE_TOOLTIP;
	private static final String	EMPTY	= "";				//$NON-NLS-1$
	private JTextField			mNameField;
	private JCheckBox			mEnabledField;
	private JTextField			mNotesField;
	private JTextField			mReferenceField;
	private JTextField			mCostField;
	private JTextField			mLevelField;
	private JTextField			mCostModifierField;
	private FeaturesPanel		mFeatures;
	private JTabbedPane			mTabPanel;
	private JComboBox			mCostType;
	private JComboBox			mAffects;
	private int					mLastLevel;

	static {
		LocalizedMessages.initialize(ModifierEditor.class);
	}

	/**
	 * Creates a new {@link ModifierEditor}.
	 * 
	 * @param modifier The {@link Modifier} to edit.
	 */
	public ModifierEditor(Modifier modifier) {
		super(modifier);

		JPanel content = new JPanel(new ColumnLayout(2));
		JPanel fields = new JPanel(new ColumnLayout(2));
		JLabel icon;
		BufferedImage image = modifier.getImage(true);
		if (image != null) {
			icon = new JLabel(new ImageIcon(image));
		} else {
			icon = new JLabel();
		}

		JPanel wrapper = new JPanel(new ColumnLayout(2));
		mNameField = createCorrectableField(fields, wrapper, MSG_NAME, modifier.getName(), MSG_NAME_TOOLTIP);
		mEnabledField = new JCheckBox(MSG_ENABLED, modifier.isEnabled());
		mEnabledField.setToolTipText(MSG_ENABLED_TOOLTIP);
		mEnabledField.setEnabled(mIsEditable);
		wrapper.add(mEnabledField);
		fields.add(wrapper);

		createCostModifierFields(fields);

		wrapper = new JPanel(new ColumnLayout(3));
		mNotesField = createField(fields, wrapper, MSG_NOTES, modifier.getNotes(), MSG_NOTES_TOOLTIP, 0);
		mReferenceField = createField(wrapper, wrapper, MSG_REFERENCE, mRow.getReference(), MSG_REFERENCE_TOOLTIP, 6);
		fields.add(wrapper);

		icon.setVerticalAlignment(SwingConstants.TOP);
		icon.setAlignmentY(-1f);
		content.add(icon);
		content.add(fields);
		add(content);

		mTabPanel = new JTabbedPane();
		mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
		Component panel = embedEditor(mFeatures);
		mTabPanel.addTab(panel.getName(), panel);
		UIUtilities.selectTab(mTabPanel, getLastTabName());
		add(mTabPanel);
	}

	@Override
	protected boolean applyChangesSelf() {
		boolean modified = mRow.setName(mNameField.getText());

		modified |= mRow.setReference(mReferenceField.getText());
		modified |= mRow.setNotes(mNotesField.getText());
		if (getCostType() == CostType.MULTIPLIER) {
			modified |= mRow.setCostMultiplier(getCostMultiplier());
		} else {
			modified |= mRow.setCost(getCost());
		}
		if (hasLevels()) {
			modified |= mRow.setLevels(getLevels());
			modified |= mRow.setCostType(CostType.PERCENTAGE);
		} else {
			modified |= mRow.setLevels(0);
			modified |= mRow.setCostType(getCostType());
		}
		modified |= mRow.setAffects((Affects) mAffects.getSelectedItem());
		modified |= mRow.setEnabled(mEnabledField.isSelected());

		if (mFeatures != null) {
			modified |= mRow.setFeatures(mFeatures.getFeatures());
		}
		return modified;
	}

	private boolean hasLevels() {
		return mCostType.getSelectedIndex() == 0;
	}

	@Override
	public void finished() {
		if (mTabPanel != null) {
			updateLastTabName(mTabPanel.getTitleAt(mTabPanel.getSelectedIndex()));
		}
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

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		if (src == mCostType) {
			costTypeChanged();
		} else if (src == mCostField || src == mLevelField) {
			updateCostModifier();
		}
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

	private JScrollPane embedEditor(Container editor) {
		JScrollPane scrollPanel = new JScrollPane(editor);
		scrollPanel.setMinimumSize(new Dimension(500, 120));
		scrollPanel.setName(editor.toString());
		if (!mIsEditable) {
			UIUtilities.disableControls(editor);
		}
		return scrollPanel;
	}

	private JTextField createNumberField(Container labelParent, Container fieldParent, String title, boolean allowSign, int value, String tooltip, int maxDigits) {
		JTextField field = new JTextField(TextUtility.makeFiller(maxDigits, '9') + TextUtility.makeFiller(maxDigits / 3, ',') + (allowSign ? "-" : EMPTY)); //$NON-NLS-1$

		UIUtilities.setOnlySize(field, field.getPreferredSize());
		field.setText(NumberUtils.format(value));
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		new NumberFilter(field, false, allowSign, true, maxDigits);
		field.addActionListener(this);
		labelParent.add(new LinkedLabel(title, field));
		fieldParent.add(field);
		return field;
	}

	private JTextField createNumberField(Container labelParent, Container fieldParent, String title, double value, String tooltip, int maxDigits) {
		JTextField field = new JTextField(TextUtility.makeFiller(maxDigits, '9') + TextUtility.makeFiller(maxDigits / 3, ',') + '.');

		UIUtilities.setOnlySize(field, field.getPreferredSize());
		field.setText(NumberUtils.format(value));
		field.setToolTipText(tooltip);
		field.setEnabled(mIsEditable);
		new NumberFilter(field, true, false, true, maxDigits);
		field.addActionListener(this);
		labelParent.add(new LinkedLabel(title, field));
		fieldParent.add(field);
		return field;
	}

	private void createCostModifierFields(Container parent) {
		JPanel wrapper = new JPanel(new ColumnLayout(7));

		mLastLevel = mRow.getLevels();
		if (mLastLevel < 1) {
			mLastLevel = 1;
		}
		if (mRow.getCostType() == CostType.MULTIPLIER) {
			mCostField = createNumberField(parent, wrapper, MSG_COST, mRow.getCostMultiplier(), MSG_COST_TOOLTIP, 5);
		} else {
			mCostField = createNumberField(parent, wrapper, MSG_COST, true, mRow.getCost(), MSG_COST_TOOLTIP, 5);
		}
		createCostType(wrapper);
		mLevelField = createNumberField(wrapper, wrapper, MSG_LEVELS, false, mLastLevel, MSG_LEVELS_TOOLTIP, 3);
		mCostModifierField = createNumberField(wrapper, wrapper, MSG_TOTAL_COST, true, 0, MSG_TOTAL_COST_TOOLTIP, 9);
		mAffects = createComboBox(wrapper, Affects.values(), mRow.getAffects());
		mCostModifierField.setEnabled(false);
		if (!mRow.hasLevels()) {
			mLevelField.setText(EMPTY);
			mLevelField.setEnabled(false);
		}
		parent.add(wrapper);
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

	private void createCostType(Container parent) {
		CostType[] types = CostType.values();
		Object[] values = new Object[types.length + 1];
		values[0] = MessageFormat.format(MSG_HAS_LEVELS, CostType.PERCENTAGE.toString());
		System.arraycopy(types, 0, values, 1, types.length);
		mCostType = createComboBox(parent, values, mRow.hasLevels() ? values[0] : mRow.getCostType());
	}

	private void costTypeChanged() {
		boolean hasLevels = hasLevels();

		if (hasLevels) {
			mLevelField.setText(NumberUtils.format(mLastLevel));
		} else {
			mLastLevel = NumberUtils.getInteger(mLevelField.getText(), 0);
			mLevelField.setText(EMPTY);
		}
		mLevelField.setEnabled(hasLevels);
		updateCostField();
		updateCostModifier();
	}

	private void updateCostField() {
		if (getCostType() == CostType.MULTIPLIER) {
			new NumberFilter(mCostField, true, false, true, 5);
			mCostField.setText(NumberUtils.format(Math.abs(NumberUtils.getDouble(mCostField.getText(), 0))));
		} else {
			new NumberFilter(mCostField, false, true, true, 5);
			mCostField.setText(NumberUtils.format(NumberUtils.getInteger(mCostField.getText(), 0), true));
		}
	}

	private void updateCostModifier() {
		boolean enabled = true;

		if (hasLevels()) {
			mCostModifierField.setText(NumberUtils.format(getCost() * getLevels(), true) + "%"); //$NON-NLS-1$
		} else {
			switch (getCostType()) {
				case PERCENTAGE:
					mCostModifierField.setText(NumberUtils.format(getCost(), true) + "%"); //$NON-NLS-1$
					break;
				case POINTS:
					mCostModifierField.setText(NumberUtils.format(getCost(), true));
					break;
				case MULTIPLIER:
					mCostModifierField.setText("x" + NumberUtils.format(getCostMultiplier())); //$NON-NLS-1$
					mAffects.setSelectedItem(Affects.TOTAL);
					enabled = false;
					break;
			}
		}
		mAffects.setEnabled(mIsEditable && enabled);
	}

	private CostType getCostType() {
		Object obj = mCostType.getSelectedItem();
		if (!(obj instanceof CostType)) {
			obj = CostType.PERCENTAGE;
		}
		return (CostType) obj;
	}

	private int getCost() {
		return NumberUtils.getInteger(mCostField.getText(), 0);
	}

	private double getCostMultiplier() {
		return NumberUtils.getDouble(mCostField.getText(), 0);
	}

	private int getLevels() {
		return NumberUtils.getInteger(mLevelField.getText(), 0);
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
