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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.modifier.ModifierListEditor;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.Defaults;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.text.IntegerFormatter;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.EditorField;
import com.trollworks.gcs.widgets.LinkedLabel;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.layout.Alignment;
import com.trollworks.gcs.widgets.layout.FlexComponent;
import com.trollworks.gcs.widgets.layout.FlexGrid;
import com.trollworks.gcs.widgets.layout.FlexRow;
import com.trollworks.gcs.widgets.layout.FlexSpacer;
import com.trollworks.gcs.widgets.outline.RowEditor;

import java.awt.Component;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.ArrayList;

import javax.swing.ImageIcon;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JComponent;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTabbedPane;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** The detailed editor for {@link Advantage}s. */
public class AdvantageEditor extends RowEditor<Advantage> implements ActionListener, DocumentListener, PropertyChangeListener {
	private static String		MSG_NAME;
	private static String		MSG_NAME_TOOLTIP;
	private static String		MSG_NAME_CANNOT_BE_EMPTY;
	private static String		MSG_TOTAL_POINTS;
	private static String		MSG_TOTAL_POINTS_TOOLTIP;
	private static String		MSG_BASE_POINTS;
	private static String		MSG_BASE_POINTS_TOOLTIP;
	private static String		MSG_LEVEL_POINTS;
	private static String		MSG_LEVEL_POINTS_TOOLTIP;
	private static String		MSG_LEVEL;
	private static String		MSG_LEVEL_TOOLTIP;
	private static String		MSG_NOTES;
	private static String		MSG_NOTES_TOOLTIP;
	private static String		MSG_CATEGORIES;
	private static String		MSG_CATEGORIES_TOOLTIP;
	private static String		MSG_TYPE;
	private static String		MSG_TYPE_TOOLTIP;
	private static String		MSG_CONTAINER_TYPE;
	private static String		MSG_CONTAINER_TYPE_TOOLTIP;
	private static String		MSG_REFERENCE;
	private static String		MSG_REFERENCE_TOOLTIP;
	private static String		MSG_NO_LEVELS;
	private static String		MSG_HAS_LEVELS;
	private static String		MSG_MENTAL;
	private static String		MSG_PHYSICAL;
	private static String		MSG_SOCIAL;
	private static String		MSG_EXOTIC;
	private static String		MSG_SUPERNATURAL;
	private static final String	EMPTY	= "";				//$NON-NLS-1$
	private EditorField			mNameField;
	private JComboBox			mLevelTypeCombo;
	private EditorField			mBasePointsField;
	private EditorField			mLevelField;
	private EditorField			mLevelPointsField;
	private EditorField			mPointsField;
	private EditorField			mNotesField;
	private EditorField			mCategoriesField;
	private EditorField			mReferenceField;
	private JTabbedPane			mTabPanel;
	private PrereqsPanel		mPrereqs;
	private FeaturesPanel		mFeatures;
	private Defaults			mDefaults;
	private MeleeWeaponEditor	mMeleeWeapons;
	private RangedWeaponEditor	mRangedWeapons;
	private ModifierListEditor	mModifiers;
	private int					mLastLevel;
	private int					mLastPointsPerLevel;
	private JCheckBox			mMentalType;
	private JCheckBox			mPhysicalType;
	private JCheckBox			mSocialType;
	private JCheckBox			mExoticType;
	private JCheckBox			mSupernaturalType;
	private JComboBox			mContainerTypeCombo;

	static {
		LocalizedMessages.initialize(AdvantageEditor.class);
	}

	/**
	 * Creates a new {@link Advantage} editor.
	 * 
	 * @param advantage The {@link Advantage} to edit.
	 */
	public AdvantageEditor(Advantage advantage) {
		super(advantage);

		FlexGrid outerGrid = new FlexGrid();

		JLabel icon = new JLabel(new ImageIcon(advantage.getImage(true)));
		UIUtilities.setOnlySize(icon, icon.getPreferredSize());
		add(icon);
		outerGrid.add(new FlexComponent(icon, Alignment.LEFT_TOP, Alignment.LEFT_TOP), 0, 0);

		FlexGrid innerGrid = new FlexGrid();
		int ri = 0;
		outerGrid.add(innerGrid, 0, 1);

		mNameField = createField(advantage.getName(), null, MSG_NAME_TOOLTIP);
		mNameField.getDocument().addDocumentListener(this);
		innerGrid.add(new FlexComponent(createLabel(MSG_NAME, mNameField), Alignment.RIGHT_BOTTOM, null), ri, 0);
		innerGrid.add(mNameField, ri++, 1);

		boolean notContainer = !advantage.canHaveChildren();
		if (notContainer) {
			mLastLevel = mRow.getLevels();
			mLastPointsPerLevel = mRow.getPointsPerLevel();
			if (mLastLevel < 0) {
				mLastLevel = 1;
			}

			FlexRow row = new FlexRow();

			mBasePointsField = createField(-9999, 9999, mRow.getPoints(), MSG_BASE_POINTS_TOOLTIP);
			row.add(mBasePointsField);
			innerGrid.add(new FlexComponent(createLabel(MSG_BASE_POINTS, mBasePointsField), Alignment.RIGHT_BOTTOM, null), ri, 0);
			innerGrid.add(row, ri++, 1);

			mLevelTypeCombo = new JComboBox(new Object[] { MSG_NO_LEVELS, MSG_HAS_LEVELS });
			mLevelTypeCombo.setSelectedIndex(mRow.isLeveled() ? 1 : 0);
			UIUtilities.setOnlySize(mLevelTypeCombo, mLevelTypeCombo.getPreferredSize());
			mLevelTypeCombo.setEnabled(mIsEditable);
			mLevelTypeCombo.addActionListener(this);
			add(mLevelTypeCombo);
			row.add(mLevelTypeCombo);

			mLevelField = createField(0, 999, mLastLevel, MSG_LEVEL_TOOLTIP);
			row.add(createLabel(MSG_LEVEL, mLevelField));
			row.add(mLevelField);

			mLevelPointsField = createField(-9999, 9999, mLastPointsPerLevel, MSG_LEVEL_POINTS_TOOLTIP);
			row.add(createLabel(MSG_LEVEL_POINTS, mLevelPointsField));
			row.add(mLevelPointsField);

			row.add(new FlexSpacer(0, 0, true, false));

			mPointsField = createField(-9999999, 9999999, mRow.getAdjustedPoints(), MSG_TOTAL_POINTS_TOOLTIP);
			mPointsField.setEnabled(false);
			row.add(createLabel(MSG_TOTAL_POINTS, mPointsField));
			row.add(mPointsField);

			if (!mRow.isLeveled()) {
				mLevelField.setText(EMPTY);
				mLevelField.setEnabled(false);
				mLevelPointsField.setText(EMPTY);
				mLevelPointsField.setEnabled(false);
			}
		}

		mNotesField = createField(advantage.getNotes(), null, MSG_NOTES_TOOLTIP);
		innerGrid.add(new FlexComponent(createLabel(MSG_NOTES, mNotesField), Alignment.RIGHT_BOTTOM, null), ri, 0);
		innerGrid.add(mNotesField, ri++, 1);

		mCategoriesField = createField(advantage.getCategoriesAsString(), null, MSG_CATEGORIES_TOOLTIP);
		innerGrid.add(new FlexComponent(createLabel(MSG_CATEGORIES, mCategoriesField), Alignment.RIGHT_BOTTOM, null), ri, 0);
		innerGrid.add(mCategoriesField, ri++, 1);

		FlexRow row = new FlexRow();
		innerGrid.add(row, ri, 1);
		if (notContainer) {
			JLabel label = new JLabel(MSG_TYPE, SwingConstants.RIGHT);
			label.setToolTipText(MSG_TYPE_TOOLTIP);
			add(label);
			innerGrid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, null), ri++, 0);

			mMentalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_MENTAL) == Advantage.TYPE_MASK_MENTAL, MSG_MENTAL);
			row.add(mMentalType);
			row.add(createTypeLabel(Images.getMentalTypeIcon(), mMentalType));

			mPhysicalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_PHYSICAL) == Advantage.TYPE_MASK_PHYSICAL, MSG_PHYSICAL);
			row.add(mPhysicalType);
			row.add(createTypeLabel(Images.getPhysicalTypeIcon(), mPhysicalType));

			mSocialType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_SOCIAL) == Advantage.TYPE_MASK_SOCIAL, MSG_SOCIAL);
			row.add(mSocialType);
			row.add(createTypeLabel(Images.getSocialTypeIcon(), mSocialType));

			mExoticType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_EXOTIC) == Advantage.TYPE_MASK_EXOTIC, MSG_EXOTIC);
			row.add(mExoticType);
			row.add(createTypeLabel(Images.getExoticTypeIcon(), mExoticType));

			mSupernaturalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_SUPERNATURAL) == Advantage.TYPE_MASK_SUPERNATURAL, MSG_SUPERNATURAL);
			row.add(mSupernaturalType);
			row.add(createTypeLabel(Images.getSupernaturalTypeIcon(), mSupernaturalType));
		} else {
			mContainerTypeCombo = new JComboBox(AdvantageContainerType.values());
			mContainerTypeCombo.setSelectedItem(mRow.getContainerType());
			UIUtilities.setOnlySize(mContainerTypeCombo, mContainerTypeCombo.getPreferredSize());
			mContainerTypeCombo.setToolTipText(MSG_CONTAINER_TYPE_TOOLTIP);
			add(mContainerTypeCombo);
			row.add(mContainerTypeCombo);
			innerGrid.add(new FlexComponent(new LinkedLabel(MSG_CONTAINER_TYPE, mContainerTypeCombo), Alignment.RIGHT_BOTTOM, null), ri++, 0);
		}

		row.add(new FlexSpacer(0, 0, true, false));

		mReferenceField = createField(mRow.getReference(), "MMMMMM", MSG_REFERENCE_TOOLTIP); //$NON-NLS-1$
		row.add(createLabel(MSG_REFERENCE, mReferenceField));
		row.add(mReferenceField);

		mTabPanel = new JTabbedPane();
		mModifiers = ModifierListEditor.createEditor(mRow);
		mModifiers.addActionListener(this);
		if (notContainer) {
			mPrereqs = new PrereqsPanel(mRow, mRow.getPrereqs());
			mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
			mDefaults = new Defaults(mRow.getDefaults());
			mMeleeWeapons = MeleeWeaponEditor.createEditor(mRow);
			mRangedWeapons = RangedWeaponEditor.createEditor(mRow);
			mDefaults.addActionListener(this);
			Component panel = embedEditor(mDefaults);
			mTabPanel.addTab(panel.getName(), panel);
			panel = embedEditor(mPrereqs);
			mTabPanel.addTab(panel.getName(), panel);
			panel = embedEditor(mFeatures);
			mTabPanel.addTab(panel.getName(), panel);
			mTabPanel.addTab(mModifiers.getName(), mModifiers);
			mTabPanel.addTab(mMeleeWeapons.getName(), mMeleeWeapons);
			mTabPanel.addTab(mRangedWeapons.getName(), mRangedWeapons);

			if (!mIsEditable) {
				UIUtilities.disableControls(mMeleeWeapons);
				UIUtilities.disableControls(mRangedWeapons);
			}
			updatePoints();
		} else {
			mTabPanel.addTab(mModifiers.getName(), mModifiers);
		}
		if (!mIsEditable) {
			UIUtilities.disableControls(mModifiers);
		}

		UIUtilities.selectTab(mTabPanel, getLastTabName());

		add(mTabPanel);
		outerGrid.add(mTabPanel, 1, 0, 1, 2);
		outerGrid.apply(this);
	}

	private JCheckBox createTypeCheckBox(boolean selected, String tooltip) {
		JCheckBox button = new JCheckBox();
		button.setSelected(selected);
		button.setToolTipText(tooltip);
		button.setEnabled(mIsEditable);
		UIUtilities.setOnlySize(button, button.getPreferredSize());
		add(button);
		return button;
	}

	private LinkedLabel createTypeLabel(BufferedImage icon, final JCheckBox linkTo) {
		LinkedLabel label = new LinkedLabel(new ImageIcon(icon), linkTo);
		label.addMouseListener(new MouseAdapter() {
			@Override public void mouseClicked(MouseEvent event) {
				linkTo.doClick();
			}
		});
		add(label);
		return label;
	}

	private JScrollPane embedEditor(JPanel editor) {
		JScrollPane scrollPanel = new JScrollPane(editor);
		scrollPanel.setMinimumSize(new Dimension(500, 120));
		scrollPanel.setName(editor.toString());
		if (!mIsEditable) {
			UIUtilities.disableControls(editor);
		}
		return scrollPanel;
	}

	private LinkedLabel createLabel(String title, JComponent linkTo) {
		LinkedLabel label = new LinkedLabel(title, linkTo);
		add(label);
		return label;
	}

	private EditorField createField(String text, String prototype, String tooltip) {
		DefaultFormatter formatter = new DefaultFormatter();
		formatter.setOverwriteMode(false);
		EditorField field = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, text, prototype, tooltip);
		field.setEnabled(mIsEditable);
		add(field);
		return field;
	}

	private EditorField createField(int min, int max, int value, String tooltip) {
		int proto = Math.max(Math.abs(min), Math.abs(max));
		if (min < 0 || max < 0) {
			proto = -proto;
		}
		EditorField field = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(min, max, false)), this, SwingConstants.LEFT, new Integer(value), new Integer(proto), tooltip);
		field.setEnabled(mIsEditable);
		add(field);
		return field;
	}

	@Override public boolean applyChangesSelf() {
		boolean modified = mRow.setName((String) mNameField.getValue());
		if (mRow.canHaveChildren()) {
			modified |= mRow.setContainerType((AdvantageContainerType) mContainerTypeCombo.getSelectedItem());
		} else {
			int type = 0;

			if (mMentalType.isSelected()) {
				type |= Advantage.TYPE_MASK_MENTAL;
			}
			if (mPhysicalType.isSelected()) {
				type |= Advantage.TYPE_MASK_PHYSICAL;
			}
			if (mSocialType.isSelected()) {
				type |= Advantage.TYPE_MASK_SOCIAL;
			}
			if (mExoticType.isSelected()) {
				type |= Advantage.TYPE_MASK_EXOTIC;
			}
			if (mSupernaturalType.isSelected()) {
				type |= Advantage.TYPE_MASK_SUPERNATURAL;
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
				ArrayList<WeaponStats> list = new ArrayList<WeaponStats>(mMeleeWeapons.getWeapons());
				list.addAll(mRangedWeapons.getWeapons());
				modified |= mRow.setWeapons(list);
			}
		}
		if (mModifiers.wasModified()) {
			modified = true;
			mRow.setModifiers(mModifiers.getModifiers());
		}
		modified |= mRow.setReference((String) mReferenceField.getValue());
		modified |= mRow.setNotes((String) mNotesField.getValue());
		modified |= mRow.setCategories((String) mCategoriesField.getValue());
		return modified;
	}

	@Override public void finished() {
		if (mTabPanel != null) {
			updateLastTabName(mTabPanel.getTitleAt(mTabPanel.getSelectedIndex()));
		}
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		if (src == mLevelTypeCombo) {
			levelTypeChanged();
		} else if (src == mModifiers) {
			updatePoints();
		}
	}

	private boolean isLeveled() {
		return mLevelTypeCombo.getSelectedItem() == MSG_HAS_LEVELS;
	}

	private void levelTypeChanged() {
		boolean isLeveled = isLeveled();

		if (isLeveled) {
			mLevelField.setValue(new Integer(mLastLevel));
			mLevelPointsField.setValue(new Integer(mLastPointsPerLevel));
		} else {
			mLastLevel = getLevels();
			mLastPointsPerLevel = getPointsPerLevel();
			mLevelField.setText(EMPTY);
			mLevelPointsField.setText(EMPTY);
		}
		mLevelField.setEnabled(isLeveled);
		mLevelPointsField.setEnabled(isLeveled);
		updatePoints();
	}

	private int getLevels() {
		return ((Integer) mLevelField.getValue()).intValue();
	}

	private int getPointsPerLevel() {
		return ((Integer) mLevelPointsField.getValue()).intValue();
	}

	private int getBasePoints() {
		return ((Integer) mBasePointsField.getValue()).intValue();
	}

	private int getPoints() {
		if (mModifiers == null) {
			return 0;
		}
		return Advantage.getAdjustedPoints(getBasePoints(), isLeveled() ? getLevels() : 0, getPointsPerLevel(), mModifiers.getAllModifiers());
	}

	private void updatePoints() {
		if (mPointsField != null) {
			mPointsField.setValue(new Integer(getPoints()));
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

	public void propertyChange(PropertyChangeEvent event) {
		if ("value".equals(event.getPropertyName())) { //$NON-NLS-1$
			Object src = event.getSource();
			if (src == mLevelField || src == mLevelPointsField || src == mBasePointsField) {
				updatePoints();
			}
		}
	}
}
