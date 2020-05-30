/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.Alignment;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.layout.FlexComponent;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.layout.RowDistribution;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.FilteredList;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.IntegerFormatter;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Component;
import java.awt.Dimension;
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
import javax.swing.JComponent;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTabbedPane;
import javax.swing.JTextArea;
import javax.swing.ScrollPaneConstants;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

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
    private EditorField                           mNotesField;
    private EditorField                           mCategoriesField;
    private EditorField                           mReferenceField;
    private JTabbedPane                           mTabPanel;
    private PrereqsPanel                          mPrereqs;
    private FeaturesPanel                         mFeatures;
    private MeleeWeaponEditor                     mMeleeWeapons;
    private RangedWeaponEditor                    mRangedWeapons;
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

        FlexGrid outerGrid = new FlexGrid();

        JLabel icon = new JLabel(advantage.getIcon(true));
        UIUtilities.setOnlySize(icon, icon.getPreferredSize());
        add(icon);
        outerGrid.add(new FlexComponent(icon, Alignment.LEFT_TOP, Alignment.LEFT_TOP), 0, 0);

        FlexGrid innerGrid = new FlexGrid();
        int      ri        = 0;
        outerGrid.add(innerGrid, 0, 1);

        FlexRow row = new FlexRow();

        mNameField = createField(advantage.getName(), null, I18n.Text("The name of the advantage, without any notes"));
        mNameField.getDocument().addDocumentListener(this);
        innerGrid.add(new FlexComponent(createLabel(I18n.Text("Name"), mNameField), Alignment.RIGHT_BOTTOM, null), ri, 0);
        innerGrid.add(row, ri++, 1);
        row.add(mNameField);

        mEnabledCheckBox = new JCheckBox(I18n.Text("Enabled"));
        mEnabledCheckBox.setSelected(advantage.isSelfEnabled());
        mEnabledCheckBox.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("If checked, this advantage is treated normally. If not checked, it is treated as if it didn't exist.")));
        mEnabledCheckBox.setEnabled(mIsEditable);
        mEnabledCheckBox.addActionListener(this);
        UIUtilities.setToPreferredSizeOnly(mEnabledCheckBox);
        add(mEnabledCheckBox);
        row.add(mEnabledCheckBox);

        boolean notContainer = !advantage.canHaveChildren();
        if (notContainer) {
            mLastLevel = mRow.getLevels();
            mLastHalfLevel = mRow.hasHalfLevel();
            mLastPointsPerLevel = mRow.getPointsPerLevel();
            if (mLastLevel < 0) {
                mLastLevel = 1;
                mLastHalfLevel = false;
            }

            row = new FlexRow();

            mBasePointsField = createField(-9999, 9999, mRow.getPoints(), I18n.Text("The base point cost of this advantage"));
            row.add(mBasePointsField);
            innerGrid.add(new FlexComponent(createLabel(I18n.Text("Base Point Cost"), mBasePointsField), Alignment.RIGHT_BOTTOM, null), ri, 0);
            innerGrid.add(row, ri++, 1);

            mLevelTypeCombo = new JComboBox<>(Levels.values());
            Levels levels = mRow.allowHalfLevels() ? Levels.HAS_HALF_LEVELS : Levels.HAS_LEVELS;
            mLevelTypeCombo.setSelectedItem(mRow.isLeveled() ? levels : Levels.NO_LEVELS);
            UIUtilities.setToPreferredSizeOnly(mLevelTypeCombo);
            mLevelTypeCombo.setEnabled(mIsEditable);
            mLevelTypeCombo.addActionListener(this);
            add(mLevelTypeCombo);
            row.add(mLevelTypeCombo);

            mLevelField = createField(0, 999, mLastLevel, I18n.Text("The level of this advantage"));
            row.add(createLabel(I18n.Text("Level"), mLevelField));
            row.add(mLevelField);

            mHalfLevel = new JCheckBox("+½");
            mHalfLevel.setSelected(mLastHalfLevel);
            mHalfLevel.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Add a half Level")));
            mHalfLevel.setEnabled(mIsEditable && advantage.allowHalfLevels());
            mHalfLevel.addActionListener(this);
            UIUtilities.setToPreferredSizeOnly(mHalfLevel);
            add(mHalfLevel);
            row.add(mHalfLevel);

            mLevelPointsField = createField(-9999, 9999, mLastPointsPerLevel, I18n.Text("The per level cost of this advantage. If this is set to zero and there is a value other than zero in the level field, then the value in the base points field will be used"));
            row.add(createLabel(I18n.Text("Point Cost Per Level"), mLevelPointsField));
            row.add(mLevelPointsField);

            mShouldRoundCostDown = new JCheckBox(I18n.Text("Round Down"));
            mShouldRoundCostDown.setSelected(advantage.shouldRoundCostDown());
            mShouldRoundCostDown.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Round point costs down if selected, round them up if not (most things in GURPS round up)")));
            mShouldRoundCostDown.setEnabled(mIsEditable);
            mShouldRoundCostDown.addActionListener(this);
            UIUtilities.setToPreferredSizeOnly(mShouldRoundCostDown);
            add(mShouldRoundCostDown);
            row.add(mShouldRoundCostDown);

            row.add(new FlexSpacer(0, 0, true, false));

            mPointsField = createField(-9999999, 9999999, mRow.getAdjustedPoints(), I18n.Text("The total point cost of this advantage"));
            mPointsField.setEnabled(false);
            row.add(createLabel(I18n.Text("Total"), mPointsField));
            row.add(mPointsField);

            if (!mRow.isLeveled()) {
                mLevelField.setText("");
                mLevelField.setEnabled(false);
                mLevelPointsField.setText("");
                mLevelPointsField.setEnabled(false);
            }
        }

        mNotesField = createField(advantage.getNotes(), null, I18n.Text("Any notes that you would like to show up in the list along with this advantage"));
        add(mNotesField);
        innerGrid.add(new FlexComponent(createLabel(I18n.Text("Notes"), mNotesField), Alignment.RIGHT_BOTTOM, null), ri, 0);
        innerGrid.add(mNotesField, ri++, 1);

        mCategoriesField = createField(advantage.getCategoriesAsString(), null, I18n.Text("The category or categories the advantage belongs to (separate multiple categories with a comma)"));
        innerGrid.add(new FlexComponent(createLabel(I18n.Text("Categories"), mCategoriesField), Alignment.RIGHT_BOTTOM, null), ri, 0);
        innerGrid.add(mCategoriesField, ri++, 1);

        mCRCombo = new JComboBox<>(SelfControlRoll.values());
        mCRCombo.setSelectedIndex(mRow.getCR().ordinal());
        UIUtilities.setToPreferredSizeOnly(mCRCombo);
        mCRCombo.setEnabled(mIsEditable);
        mCRCombo.addActionListener(this);
        add(mCRCombo);
        mCRAdjCombo = new JComboBox<>(SelfControlRollAdjustments.values());
        mCRAdjCombo.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Adjustments that are applied due to Self-Control Roll limitations")));
        mCRAdjCombo.setSelectedIndex(mRow.getCRAdj().ordinal());
        UIUtilities.setToPreferredSizeOnly(mCRAdjCombo);
        mCRAdjCombo.setEnabled(mIsEditable && mRow.getCR() != SelfControlRoll.NONE_REQUIRED);
        add(mCRAdjCombo);
        innerGrid.add(new FlexComponent(createLabel(I18n.Text("Self-Control Roll"), mCRCombo), Alignment.RIGHT_BOTTOM, null), ri, 0);
        row = new FlexRow();
        row.add(mCRCombo);
        row.add(mCRAdjCombo);
        innerGrid.add(row, ri++, 1);

        row = new FlexRow();
        innerGrid.add(row, ri, 1);
        if (notContainer) {
            JLabel label = new JLabel(I18n.Text("Type"), SwingConstants.RIGHT);
            label.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The type of advantage this is")));
            add(label);
            innerGrid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, null), ri, 0);

            mMentalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_MENTAL) == Advantage.TYPE_MASK_MENTAL, I18n.Text("Mental"));
            row.add(mMentalType);
            row.add(createTypeLabel(Images.MENTAL_TYPE, mMentalType));

            mPhysicalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_PHYSICAL) == Advantage.TYPE_MASK_PHYSICAL, I18n.Text("Physical"));
            row.add(mPhysicalType);
            row.add(createTypeLabel(Images.PHYSICAL_TYPE, mPhysicalType));

            mSocialType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_SOCIAL) == Advantage.TYPE_MASK_SOCIAL, I18n.Text("Social"));
            row.add(mSocialType);
            row.add(createTypeLabel(Images.SOCIAL_TYPE, mSocialType));

            mExoticType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_EXOTIC) == Advantage.TYPE_MASK_EXOTIC, I18n.Text("Exotic"));
            row.add(mExoticType);
            row.add(createTypeLabel(Images.EXOTIC_TYPE, mExoticType));

            mSupernaturalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_SUPERNATURAL) == Advantage.TYPE_MASK_SUPERNATURAL, I18n.Text("Supernatural"));
            row.add(mSupernaturalType);
            row.add(createTypeLabel(Images.SUPERNATURAL_TYPE, mSupernaturalType));
        } else {
            mContainerTypeCombo = new JComboBox<>(AdvantageContainerType.values());
            mContainerTypeCombo.setSelectedItem(mRow.getContainerType());
            UIUtilities.setToPreferredSizeOnly(mContainerTypeCombo);
            mContainerTypeCombo.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("The type of container this is")));
            add(mContainerTypeCombo);
            row.add(mContainerTypeCombo);
            innerGrid.add(new FlexComponent(new LinkedLabel(I18n.Text("Container Type"), mContainerTypeCombo), Alignment.RIGHT_BOTTOM, null), ri, 0);
        }

        row.add(new FlexSpacer(0, 0, true, false));

        mReferenceField = createField(mRow.getReference(), "MMMMMM", I18n.Text("Page Reference"));
        row.add(createLabel(I18n.Text("Ref"), mReferenceField));
        row.add(mReferenceField);

        mTabPanel = new JTabbedPane();
        mModifiers = AdvantageModifierListEditor.createEditor(mRow);
        mModifiers.addActionListener(this);
        if (notContainer) {
            mPrereqs = new PrereqsPanel(mRow, mRow.getPrereqs());
            mFeatures = new FeaturesPanel(mRow, mRow.getFeatures());
            mMeleeWeapons = MeleeWeaponEditor.createEditor(mRow);
            mRangedWeapons = RangedWeaponEditor.createEditor(mRow);
            Component panel = embedEditor(mPrereqs);
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

        if (mRow.getDataFile() instanceof GURPSCharacter) {
            mUserDesc = mRow.getUserDesc();
            mTabPanel.addTab(I18n.Text("User Description"), createUserDescEditor());
        }

        if (!mIsEditable) {
            UIUtilities.disableControls(mModifiers);
        }

        UIUtilities.selectTab(mTabPanel, getLastTabName());

        add(mTabPanel);
        outerGrid.add(mTabPanel, 1, 0, 1, 2);
        outerGrid.apply(this);
    }

    private JPanel createUserDescEditor() {
        JPanel    content = new JPanel(new ColumnLayout(2, RowDistribution.GIVE_EXCESS_TO_LAST));
        JLabel    icon    = new JLabel(Images.NOT_MARKER);
        JTextArea editor  = new JTextArea(mUserDesc);
        editor.setLineWrap(true);
        editor.setWrapStyleWord(true);
        editor.setEnabled(mIsEditable);
        JScrollPane scroller = new JScrollPane(editor, ScrollPaneConstants.VERTICAL_SCROLLBAR_ALWAYS, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_ALWAYS);
        scroller.setMinimumSize(new Dimension(500, 300));
        icon.setVerticalAlignment(SwingConstants.TOP);
        icon.setAlignmentY(-1);
        content.add(icon);
        content.add(scroller);

        // editor.setPreferredSize(new Dimension(600, 400));
        editor.getDocument().addDocumentListener(new DocumentListener() {

            @Override
            public void removeUpdate(DocumentEvent e) {
                mUserDesc = editor.getText();
            }

            @Override
            public void insertUpdate(DocumentEvent e) {
                mUserDesc = editor.getText();
            }

            @Override
            public void changedUpdate(DocumentEvent e) {
                mUserDesc = editor.getText();
            }
        });
        return content;
    }

    private JCheckBox createTypeCheckBox(boolean selected, String tooltip) {
        JCheckBox button = new JCheckBox();
        button.setSelected(selected);
        button.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        button.setEnabled(mIsEditable);
        UIUtilities.setToPreferredSizeOnly(button);
        add(button);
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
        EditorField field = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(min, max, false)), this, SwingConstants.LEFT, Integer.valueOf(value), Integer.valueOf(proto), tooltip);
        field.setEnabled(mIsEditable);
        add(field);
        return field;
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
        modified |= mRow.setNotes((String) mNotesField.getValue());
        modified |= mRow.setCategories((String) mCategoriesField.getValue());
        if (mUserDesc != null) {
            modified |= mRow.setUserDesc(mUserDesc);
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

    @Override
    public void changedUpdate(DocumentEvent event) {
        nameChanged();
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        nameChanged();
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        nameChanged();
    }

    private void nameChanged() {
        LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().isEmpty() ? I18n.Text("The name field may not be empty") : null);
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
