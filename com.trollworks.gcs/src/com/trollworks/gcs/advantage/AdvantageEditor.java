/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.AdvantageModifierListEditor;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.FilteredList;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.IntegerFormatter;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.weapon.MeleeWeaponListEditor;
import com.trollworks.gcs.weapon.RangedWeaponListEditor;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Container;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;
import javax.swing.text.Document;

/** The detailed editor for {@link Advantage}s. */
public class AdvantageEditor extends RowEditor<Advantage> implements ActionListener, DocumentListener, PropertyChangeListener {
    private EditorField                           mNameField;
    private JCheckBox                             mShouldRoundCostDown;
    private JComboBox<Levels>                     mLevelTypeCombo;
    private EditorField                           mBasePointsField;
    private EditorField                           mLevelField;
    private JCheckBox                             mHalfLevel;
    private EditorField                           mLevelPointsField;
    private EditorField                           mPointsField;
    private MultiLineTextField                    mNotesField;
    private MultiLineTextField                    mUserDescField;
    private EditorField                           mCategoriesField;
    private EditorField                           mReferenceField;
    private PrereqsPanel                          mPrereqs;
    private FeaturesPanel                         mFeatures;
    private MeleeWeaponListEditor                 mMeleeWeapons;
    private RangedWeaponListEditor                mRangedWeapons;
    private AdvantageModifierListEditor           mModifiers;
    private int                                   mLastLevel;
    private int                                   mLastPointsPerLevel;
    private boolean                               mLastHalfLevel;
    private JCheckBox                             mMentalType;
    private JCheckBox                             mPhysicalType;
    private JCheckBox                             mSocialType;
    private JCheckBox                             mExoticType;
    private JCheckBox                             mSupernaturalType;
    private JCheckBox                             mEnabledCheckBox;
    private JComboBox<AdvantageContainerType>     mContainerTypeCombo;
    private JComboBox<SelfControlRoll>            mCRCombo;
    private JComboBox<SelfControlRollAdjustments> mCRAdjCombo;
    private String                                mUserDesc;

    /**
     * Creates a new {@link Advantage} editor.
     *
     * @param advantage The {@link Advantage} to edit.
     */
    public AdvantageEditor(Advantage advantage) {
        super(advantage);
        addContent();
    }

    @Override
    protected void addContentSelf(ScrollContent outer) {
        outer.add(createTopSection(), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        boolean isContainer = mRow.canHaveChildren();
        if (!isContainer) {
            mPrereqs = new PrereqsPanel(mRow, mRow.getPrereqs());
            addSection(outer, mPrereqs);
            mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
            addSection(outer, mFeatures);
        }
        mModifiers = AdvantageModifierListEditor.createEditor(mRow);
        mModifiers.addActionListener(this);
        addSection(outer, mModifiers);
        if (!isContainer) {
            List<WeaponStats> weapons = mRow.getWeapons();
            mMeleeWeapons = new MeleeWeaponListEditor(mRow, weapons);
            addSection(outer, mMeleeWeapons);
            mRangedWeapons = new RangedWeaponListEditor(mRow, weapons);
            addSection(outer, mRangedWeapons);
            updatePoints();
        }
    }

    private JPanel createTopSection() {
        JPanel panel = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        addPrimaryCommonFields(panel);
        if (mRow.canHaveChildren()) {
            addSecondaryCommonFields(panel);
            addContainerTypeFields(panel);
        } else {
            addPointFields(panel);
            addSecondaryCommonFields(panel);
            addTypeFields(panel);
        }
        return panel;
    }

    private void addPrimaryCommonFields(Container parent) {
        mNameField = createField(mRow.getName(), null, I18n.Text("The name of the advantage, without any notes"));
        mNameField.getDocument().addDocumentListener(this);
        addLabel(parent, I18n.Text("Name"), mNameField);
        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(2).setMargins(0));
        wrapper.add(mNameField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        mEnabledCheckBox = new JCheckBox(I18n.Text("Enabled"));
        mEnabledCheckBox.setSelected(mRow.isSelfEnabled());
        mEnabledCheckBox.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("If checked, this advantage is treated normally. If not checked, it is treated as if it didn't exist.")));
        mEnabledCheckBox.addActionListener(this);
        wrapper.add(mEnabledCheckBox);
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void addPointFields(Container parent) {
        mLastLevel = mRow.getLevels();
        mLastHalfLevel = mRow.hasHalfLevel();
        mLastPointsPerLevel = mRow.getPointsPerLevel();
        if (mLastLevel < 0) {
            mLastLevel = 1;
            mLastHalfLevel = false;
        }

        mPointsField = createField(-9999999, 9999999, mRow.getAdjustedPoints(), I18n.Text("The total point cost of this advantage"));
        mPointsField.setEnabled(false);
        addLabel(parent, I18n.Text("Point Cost"), mPointsField);
        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(10).setMargins(0));
        wrapper.add(mPointsField, new PrecisionLayoutData().setFillHorizontalAlignment());

        mBasePointsField = createField(-9999, 9999, mRow.getPoints(), I18n.Text("The base point cost of this advantage"));
        addLabel(wrapper, I18n.Text("Base"), mBasePointsField);
        wrapper.add(mBasePointsField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        mLevelTypeCombo = new JComboBox<>(Levels.values());
        Levels levels = mRow.allowHalfLevels() ? Levels.HAS_HALF_LEVELS : Levels.HAS_LEVELS;
        mLevelTypeCombo.setSelectedItem(mRow.isLeveled() ? levels : Levels.NO_LEVELS);
        mLevelTypeCombo.addActionListener(this);
        wrapper.add(mLevelTypeCombo);

        mLevelField = createField(0, 9999, mLastLevel, I18n.Text("The level of this advantage"));
        addLabel(wrapper, I18n.Text("Level"), mLevelField);
        wrapper.add(mLevelField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        mHalfLevel = new JCheckBox("+½");
        mHalfLevel.setSelected(mLastHalfLevel);
        mHalfLevel.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Add a half Level")));
        mHalfLevel.setEnabled(mRow.allowHalfLevels());
        mHalfLevel.addActionListener(this);
        wrapper.add(mHalfLevel);

        mLevelPointsField = createField(-9999, 9999, mLastPointsPerLevel, I18n.Text("The per level cost of this advantage. If this is set to zero and there is a value other than zero in the level field, then the value in the base points field will be used"));
        addLabel(wrapper, I18n.Text("Per Level"), mLevelPointsField);
        wrapper.add(mLevelPointsField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        mShouldRoundCostDown = new JCheckBox(I18n.Text("Round Down"));
        mShouldRoundCostDown.setSelected(mRow.shouldRoundCostDown());
        mShouldRoundCostDown.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Round point costs down if selected, round them up if not (most things in GURPS round up)")));
        mShouldRoundCostDown.addActionListener(this);
        wrapper.add(mShouldRoundCostDown);

        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        if (!mRow.isLeveled()) {
            mLevelField.setText("");
            mLevelField.setEnabled(false);
            mLevelPointsField.setText("");
            mLevelPointsField.setEnabled(false);
        }
    }

    private void addSecondaryCommonFields(Container parent) {
        mNotesField = new MultiLineTextField(mRow.getNotes(), I18n.Text("Any notes that you would like to show up in the list along with this advantage"), this);
        parent.add(new LinkedLabel(I18n.Text("Notes"), mNotesField), new PrecisionLayoutData().setFillHorizontalAlignment().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2));
        parent.add(mNotesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        if (mRow.getDataFile() instanceof GURPSCharacter) {
            mUserDesc = mRow.getUserDesc();
            mUserDescField = new MultiLineTextField(mUserDesc, I18n.Text("Additional notes for your own reference. These only exist in character sheets and will be removed if transferred to a data list or template"), this);
            parent.add(new LinkedLabel(I18n.Text("User Description"), mUserDescField), new PrecisionLayoutData().setFillHorizontalAlignment().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2));
            parent.add(mUserDescField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        }

        mCategoriesField = createField(mRow.getCategoriesAsString(), null, I18n.Text("The category or categories the advantage belongs to (separate multiple categories with a comma)"));
        parent.add(new LinkedLabel(I18n.Text("Categories"), mCategoriesField), new PrecisionLayoutData().setFillHorizontalAlignment());
        parent.add(mCategoriesField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        mCRCombo = new JComboBox<>(SelfControlRoll.values());
        mCRCombo.setSelectedIndex(mRow.getCR().ordinal());
        mCRCombo.addActionListener(this);
        parent.add(new LinkedLabel(I18n.Text("Self-Control Roll"), mCRCombo), new PrecisionLayoutData().setFillHorizontalAlignment());
        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(2).setMargins(0));
        wrapper.add(mCRCombo);
        mCRAdjCombo = new JComboBox<>(SelfControlRollAdjustments.values());
        mCRAdjCombo.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Adjustments that are applied due to Self-Control Roll limitations")));
        mCRAdjCombo.setSelectedIndex(mRow.getCRAdj().ordinal());
        mCRAdjCombo.setEnabled(mRow.getCR() != SelfControlRoll.NONE_REQUIRED);
        wrapper.add(mCRAdjCombo);
        parent.add(wrapper);
    }

    private void addTypeFields(Container parent) {
        JLabel label = new JLabel(I18n.Text("Type"), SwingConstants.RIGHT);
        label.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The type of advantage this is")));
        parent.add(label, new PrecisionLayoutData().setFillHorizontalAlignment());

        mMentalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_MENTAL) == Advantage.TYPE_MASK_MENTAL, I18n.Text("Mental"));
        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(12).setMargins(0));
        wrapper.add(mMentalType);
        wrapper.add(createTypeLabel(Images.MENTAL_TYPE, mMentalType));

        mPhysicalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_PHYSICAL) == Advantage.TYPE_MASK_PHYSICAL, I18n.Text("Physical"));
        wrapper.add(mPhysicalType);
        wrapper.add(createTypeLabel(Images.PHYSICAL_TYPE, mPhysicalType));

        mSocialType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_SOCIAL) == Advantage.TYPE_MASK_SOCIAL, I18n.Text("Social"));
        wrapper.add(mSocialType);
        wrapper.add(createTypeLabel(Images.SOCIAL_TYPE, mSocialType));

        mExoticType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_EXOTIC) == Advantage.TYPE_MASK_EXOTIC, I18n.Text("Exotic"));
        wrapper.add(mExoticType);
        wrapper.add(createTypeLabel(Images.EXOTIC_TYPE, mExoticType));

        mSupernaturalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_SUPERNATURAL) == Advantage.TYPE_MASK_SUPERNATURAL, I18n.Text("Supernatural"));
        wrapper.add(mSupernaturalType);
        wrapper.add(createTypeLabel(Images.SUPERNATURAL_TYPE, mSupernaturalType));
        addRefField(wrapper);

        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void addContainerTypeFields(Container parent) {
        mContainerTypeCombo = new JComboBox<>(AdvantageContainerType.values());
        mContainerTypeCombo.setSelectedItem(mRow.getContainerType());
        mContainerTypeCombo.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The type of container this is")));
        parent.add(new LinkedLabel(I18n.Text("Container Type"), mContainerTypeCombo), new PrecisionLayoutData().setFillHorizontalAlignment());
        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(3).setMargins(0));
        wrapper.add(mContainerTypeCombo);
        addRefField(wrapper);
        parent.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void addRefField(Container parent) {
        mReferenceField = createField(mRow.getReference(), "MMMMMM", I18n.Text("Page Reference"));
        parent.add(new LinkedLabel(I18n.Text("Ref"), mReferenceField), new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(10));
        parent.add(mReferenceField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private JCheckBox createTypeCheckBox(boolean selected, String tooltip) {
        JCheckBox button = new JCheckBox();
        button.setSelected(selected);
        button.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        return button;
    }

    private LinkedLabel createTypeLabel(RetinaIcon icon, JCheckBox linkTo) {
        LinkedLabel label = new LinkedLabel(icon, linkTo);
        label.addMouseListener(new MouseAdapter() {
            @Override
            public void mouseClicked(MouseEvent event) {
                linkTo.doClick();
            }
        });
        return label;
    }

    private EditorField createField(String text, String prototype, String tooltip) {
        DefaultFormatter formatter = new DefaultFormatter();
        formatter.setOverwriteMode(false);
        return new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, text, prototype, tooltip);
    }

    private EditorField createField(int min, int max, int value, String tooltip) {
        int proto = Math.max(Math.abs(min), Math.abs(max));
        if (min < 0 || max < 0) {
            proto = -proto;
        }
        return new EditorField(new DefaultFormatterFactory(new IntegerFormatter(min, max, false)), this, SwingConstants.LEFT, Integer.valueOf(value), Integer.valueOf(proto), tooltip);
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified = mRow.setName((String) mNameField.getValue());
        modified |= mRow.setEnabled(enabled());
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
            modified |= mRow.setShouldRoundCostDown(shouldRoundCostDown());
            modified |= mRow.setAllowHalfLevels(allowHalfLevels());
            modified |= mRow.setPoints(getBasePoints());
            if (isLeveled()) {
                modified |= mRow.setPointsPerLevel(getPointsPerLevel());
                modified |= mRow.setLevels(getLevels());
            } else {
                modified |= mRow.setPointsPerLevel(0);
                modified |= mRow.setLevels(-1);
            }
            modified |= isLeveled() && allowHalfLevels() ? mRow.setHalfLevel(getHalfLevel()) : mRow.setHalfLevel(false);
            if (mPrereqs != null) {
                modified |= mRow.setPrereqs(mPrereqs.getPrereqList());
            }
            if (mFeatures != null) {
                modified |= mRow.setFeatures(mFeatures.getFeatures());
            }
            if (mMeleeWeapons != null) {
                List<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
                list.addAll(mRangedWeapons.getWeapons());
                modified |= mRow.setWeapons(list);
            }
        }
        modified |= mRow.setCR(getCR());
        modified |= mRow.setCRAdj(getCRAdj());
        if (mModifiers.wasModified()) {
            modified = true;
            mRow.setModifiers(mModifiers.getModifiers());
        }
        modified |= mRow.setReference((String) mReferenceField.getValue());
        modified |= mRow.setNotes(mNotesField.getText());
        modified |= mRow.setCategories((String) mCategoriesField.getValue());
        if (mUserDesc != null) {
            modified |= mRow.setUserDesc(mUserDesc);
        }
        return modified;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object src = event.getSource();
        if (src == mLevelTypeCombo) {
            levelTypeChanged();
        } else if (src == mModifiers || src == mShouldRoundCostDown || src == mHalfLevel || src == mEnabledCheckBox) {
            updatePoints();
        } else if (src == mCRCombo) {
            SelfControlRoll cr = getCR();
            if (cr == SelfControlRoll.NONE_REQUIRED) {
                mCRAdjCombo.setSelectedItem(SelfControlRollAdjustments.NONE);
                mCRAdjCombo.setEnabled(false);
            } else {
                mCRAdjCombo.setEnabled(mIsEditable);
            }
            updatePoints();
        }
    }

    private boolean isLeveled() {
        return mLevelTypeCombo.getSelectedItem() != Levels.NO_LEVELS;
    }

    private boolean allowHalfLevels() {
        return mLevelTypeCombo.getSelectedItem() == Levels.HAS_HALF_LEVELS;
    }

    private void levelTypeChanged() {
        boolean isLeveled       = isLeveled();
        boolean allowHalfLevels = allowHalfLevels();

        if (isLeveled) {
            mLevelField.setValue(Integer.valueOf(mLastLevel));
            mLevelPointsField.setValue(Integer.valueOf(mLastPointsPerLevel));
            mHalfLevel.setSelected(mLastHalfLevel && allowHalfLevels);
        } else {
            mLastLevel = getLevels();
            mLastHalfLevel = getHalfLevel();
            mLastPointsPerLevel = getPointsPerLevel();
            mLevelField.setText("");
            mHalfLevel.setSelected(false);
            mLevelPointsField.setText("");
        }
        mLevelField.setEnabled(isLeveled);
        mLevelPointsField.setEnabled(isLeveled);
        mHalfLevel.setEnabled(isLeveled && allowHalfLevels);
        updatePoints();
    }

    private SelfControlRoll getCR() {
        return (SelfControlRoll) mCRCombo.getSelectedItem();
    }

    private SelfControlRollAdjustments getCRAdj() {
        return (SelfControlRollAdjustments) mCRAdjCombo.getSelectedItem();
    }

    private int getLevels() {
        return ((Integer) mLevelField.getValue()).intValue();
    }

    private boolean getHalfLevel() {
        return mHalfLevel.isSelected();
    }

    private int getPointsPerLevel() {
        return ((Integer) mLevelPointsField.getValue()).intValue();
    }

    private int getBasePoints() {
        return ((Integer) mBasePointsField.getValue()).intValue();
    }

    private int getPoints() {
        if (mModifiers == null || !enabled()) {
            return 0;
        }
        List<AdvantageModifier> modifiers = new FilteredList<>(mModifiers.getAllModifiers(), AdvantageModifier.class);
        return mRow.getAdjustedPoints(getBasePoints(), isLeveled() ? getLevels() : 0, allowHalfLevels() && getHalfLevel(), getPointsPerLevel(), getCR(), modifiers, shouldRoundCostDown());
    }

    private void updatePoints() {
        if (mPointsField != null) {
            mPointsField.setValue(Integer.valueOf(getPoints()));
        }
    }

    private boolean shouldRoundCostDown() {
        return mShouldRoundCostDown.isSelected();
    }

    private boolean enabled() {
        return mEnabledCheckBox.isSelected();
    }

    private void docChanged(DocumentEvent event) {
        Document doc = event.getDocument();
        if (mNameField.getDocument() == doc) {
            LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.Text("The name field may not be empty") : null);
        } else if (mUserDescField.getDocument() == doc) {
            mUserDesc = mUserDescField.getText();
        }
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        docChanged(event);
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        docChanged(event);
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        docChanged(event);
    }

    @Override
    public void propertyChange(PropertyChangeEvent event) {
        if ("value".equals(event.getPropertyName())) {
            Object src = event.getSource();
            if (src == mLevelField || src == mLevelPointsField || src == mBasePointsField) {
                updatePoints();
            }
        }
    }
}
