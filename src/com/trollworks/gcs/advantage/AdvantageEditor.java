/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.feature.FeaturesPanel;
import com.trollworks.gcs.modifier.ModifierListEditor;
import com.trollworks.gcs.prereq.PrereqsPanel;
import com.trollworks.gcs.skill.Defaults;
import com.trollworks.gcs.weapon.MeleeWeaponEditor;
import com.trollworks.gcs.weapon.RangedWeaponEditor;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.layout.Alignment;
import com.trollworks.toolkit.ui.layout.FlexComponent;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.widget.EditorField;
import com.trollworks.toolkit.ui.widget.LinkedLabel;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.IntegerFormatter;

import java.awt.Component;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.ArrayList;

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
	@Localize("Name")
	@Localize(locale = "de", value = "Name")
	@Localize(locale = "ru", value = "Название")
	private static String							NAME;
	@Localize("The name of the advantage, without any notes")
	@Localize(locale = "de", value = "Der Name des Vorteils ohne Anmerkungen")
	@Localize(locale = "ru", value = "Название преимущества без заметок")
	private static String							NAME_TOOLTIP;
	@Localize("The name field may not be empty")
	@Localize(locale = "de", value = "Der Name darf nicht leer sein")
	@Localize(locale = "ru", value = "Поле \"Название\" не может быть пустым")
	private static String							NAME_CANNOT_BE_EMPTY;
	@Localize("Self-Control Roll")
	@Localize(locale = "de", value = "Selbstbeherrschungs-Probe")
	@Localize(locale = "ru", value = "Бросок самоконтроля")
	private static String							CR;
	@Localize("Adjustments that are applied due to Self-Control Roll limitations")
	@Localize(locale = "de", value = "Anpassungen, die auf den Wert der Selbstbeherrschungs-Probe basieren")
	@Localize(locale = "ru", value = "Настройки, которые применяются для ограничений бросков СамоКонтроля (СК)")
	private static String							CR_ADJ_TOOLTIP;
	@Localize("Total")
	@Localize(locale = "de", value = "Gesamt")
	@Localize(locale = "ru", value = "Всего")
	private static String							TOTAL_POINTS;
	@Localize("The total point cost of this advantage")
	@Localize(locale = "de", value = "Die Gesamtkosten dieses Vortiels")
	@Localize(locale = "ru", value = "Сумарная стоимость преимущества")
	private static String							TOTAL_POINTS_TOOLTIP;
	@Localize("Base Point Cost")
	@Localize(locale = "de", value = "Grundkosten")
	@Localize(locale = "ru", value = "Базовая стоимость")
	private static String							BASE_POINTS;
	@Localize("The base point cost of this advantage")
	@Localize(locale = "de", value = "Die Grundkosten dieses Vorteils")
	@Localize(locale = "ru", value = "Базовая стоимость преимущества")
	private static String							BASE_POINTS_TOOLTIP;
	@Localize("Point Cost Per Level")
	@Localize(locale = "de", value = "Kosten pro Stufe")
	@Localize(locale = "ru", value = "Количество очков за уровень")
	private static String							LEVEL_POINTS;
	@Localize("The per level cost of this advantage. If this is set to zero\nand there is a value other than zero in the level field, then the\nvalue in the base points field will be used")
	@Localize(locale = "de", value = "Die Kosten pro Stufe dieses Vorteils.  Wenn dieses Feld leer ist\nund im Stufen-Feld etwas anderes als Null steht, dann wird\nder Wert im Grundkosten-Feld verwendet")
	@Localize(locale = "ru", value = "Стоимость одного уровня преимущества. Если этот параметр установлен в ноль\n и есть значение, отличное от нуля в поле Уровень, то\nбудет использоваться значение из поля Базовая стоимость")
	private static String							LEVEL_POINTS_TOOLTIP;
	@Localize("Level")
	@Localize(locale = "de", value = "Stufe")
	@Localize(locale = "ru", value = "Уровень")
	private static String							LEVEL;
	@Localize("The level of this advantage")
	@Localize(locale = "de", value = "Die Stufe dieses Vorteils")
	@Localize(locale = "ru", value = "Уровень преимущества")
	private static String							LEVEL_TOOLTIP;
	@Localize("Categories")
	@Localize(locale = "de", value = "Kategorie")
	@Localize(locale = "ru", value = "Категории")
	private static String							CATEGORIES;
	@Localize("The category or categories the advantage belongs to (separate multiple categories with a comma)")
	@Localize(locale = "de", value = "Die Kategorie oder Kategorien, denen dieser Vorteil angehört (trenne mehrere Kategorien mit einem Komma)")
	@Localize(locale = "ru", value = "Категория или категории, к которым относится преимущество (перечислить через запятую)")
	private static String							CATEGORIES_TOOLTIP;
	@Localize("Notes")
	@Localize(locale = "de", value = "Anmerkungen")
	@Localize(locale = "ru", value = "Заметка")
	private static String							NOTES;
	@Localize("Any notes that you would like to show up in the list along with this advantage")
	@Localize(locale = "de", value = "Anmerkungen, die in der Liste neben dem Vorteil erscheinen sollen")
	@Localize(locale = "ru", value = "Заметки, которые показываются в списке рядом с преимуществом")
	private static String							NOTES_TOOLTIP;
	@Localize("Type")
	@Localize(locale = "de", value = "Typ")
	@Localize(locale = "ru", value = "Тип")
	private static String							TYPE;
	@Localize("The type of advantage this is")
	@Localize(locale = "de", value = "Der Typ dieses Vorteils")
	@Localize(locale = "ru", value = "Тип этого преимущества")
	private static String							TYPE_TOOLTIP;
	@Localize("Container Type")
	@Localize(locale = "de", value = "Container-Typ")
	@Localize(locale = "ru", value = "Тип контейнера")
	private static String							CONTAINER_TYPE;
	@Localize("The type of container this is")
	@Localize(locale = "de", value = "Der Container-Typ dieses Vorteils")
	@Localize(locale = "ru", value = "Тип этого контейнера")
	private static String							CONTAINER_TYPE_TOOLTIP;
	@Localize("Ref")
	@Localize(locale = "de", value = "Seitenangabe")
	@Localize(locale = "ru", value = "Ссыл")
	private static String							REFERENCE;
	@Localize("Page Reference")
	@Localize(locale = "de", value = "Eine Referenz auf das Buch und die Seite, auf der dieser Vorteil beschrieben wird (z.B. B22 würde auf \"Basic Set\" Seite 22 verweisen)")
	@Localize(locale = "ru", value = "Ссылка на страницу")
	private static String							REFERENCE_TOOLTIP;
	@Localize("Has No Levels")
	@Localize(locale = "de", value = "Hat keine Stufen")
	@Localize(locale = "ru", value = "Не имеет уровни")
	private static String							NO_LEVELS;
	@Localize("Has Levels")
	@Localize(locale = "de", value = "Hat Stufen")
	@Localize(locale = "ru", value = "Имеет уровни")
	private static String							HAS_LEVELS;
	@Localize("Mental")
	@Localize(locale = "de", value = "Mental")
	@Localize(locale = "ru", value = "Ментальный")
	private static String							MENTAL;
	@Localize("Physical")
	@Localize(locale = "de", value = "Physisch")
	@Localize(locale = "ru", value = "Физическая")
	private static String							PHYSICAL;
	@Localize("Social")
	@Localize(locale = "de", value = "Sozial")
	@Localize(locale = "ru", value = "Социальная")
	private static String							SOCIAL;
	@Localize("Exotic")
	@Localize(locale = "de", value = "Exotisch")
	@Localize(locale = "ru", value = "Экзотические")
	private static String							EXOTIC;
	@Localize("Supernatural")
	@Localize(locale = "de", value = "Übernatürlich")
	@Localize(locale = "ru", value = "Сверхъестественное")
	private static String							SUPERNATURAL;

	static {
		Localization.initialize();
	}

	private EditorField								mNameField;
	private JComboBox<String>						mLevelTypeCombo;
	private EditorField								mBasePointsField;
	private EditorField								mLevelField;
	private EditorField								mLevelPointsField;
	private EditorField								mPointsField;
	private EditorField								mNotesField;
	private EditorField								mCategoriesField;
	private EditorField								mReferenceField;
	private JTabbedPane								mTabPanel;
	private PrereqsPanel							mPrereqs;
	private FeaturesPanel							mFeatures;
	private Defaults								mDefaults;
	private MeleeWeaponEditor						mMeleeWeapons;
	private RangedWeaponEditor						mRangedWeapons;
	private ModifierListEditor						mModifiers;
	private int										mLastLevel;
	private int										mLastPointsPerLevel;
	private JCheckBox								mMentalType;
	private JCheckBox								mPhysicalType;
	private JCheckBox								mSocialType;
	private JCheckBox								mExoticType;
	private JCheckBox								mSupernaturalType;
	private JComboBox<AdvantageContainerType>		mContainerTypeCombo;
	private JComboBox<SelfControlRoll>				mCRCombo;
	private JComboBox<SelfControlRollAdjustments>	mCRAdjCombo;

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
		int ri = 0;
		outerGrid.add(innerGrid, 0, 1);

		mNameField = createField(advantage.getName(), null, NAME_TOOLTIP);
		mNameField.getDocument().addDocumentListener(this);
		innerGrid.add(new FlexComponent(createLabel(NAME, mNameField), Alignment.RIGHT_BOTTOM, null), ri, 0);
		innerGrid.add(mNameField, ri++, 1);

		boolean notContainer = !advantage.canHaveChildren();
		if (notContainer) {
			mLastLevel = mRow.getLevels();
			mLastPointsPerLevel = mRow.getPointsPerLevel();
			if (mLastLevel < 0) {
				mLastLevel = 1;
			}

			FlexRow row = new FlexRow();

			mBasePointsField = createField(-9999, 9999, mRow.getPoints(), BASE_POINTS_TOOLTIP);
			row.add(mBasePointsField);
			innerGrid.add(new FlexComponent(createLabel(BASE_POINTS, mBasePointsField), Alignment.RIGHT_BOTTOM, null), ri, 0);
			innerGrid.add(row, ri++, 1);

			mLevelTypeCombo = new JComboBox<>(new String[] { NO_LEVELS, HAS_LEVELS });
			mLevelTypeCombo.setSelectedIndex(mRow.isLeveled() ? 1 : 0);
			UIUtilities.setOnlySize(mLevelTypeCombo, mLevelTypeCombo.getPreferredSize());
			mLevelTypeCombo.setEnabled(mIsEditable);
			mLevelTypeCombo.addActionListener(this);
			add(mLevelTypeCombo);
			row.add(mLevelTypeCombo);

			mLevelField = createField(0, 999, mLastLevel, LEVEL_TOOLTIP);
			row.add(createLabel(LEVEL, mLevelField));
			row.add(mLevelField);

			mLevelPointsField = createField(-9999, 9999, mLastPointsPerLevel, LEVEL_POINTS_TOOLTIP);
			row.add(createLabel(LEVEL_POINTS, mLevelPointsField));
			row.add(mLevelPointsField);

			row.add(new FlexSpacer(0, 0, true, false));

			mPointsField = createField(-9999999, 9999999, mRow.getAdjustedPoints(), TOTAL_POINTS_TOOLTIP);
			mPointsField.setEnabled(false);
			row.add(createLabel(TOTAL_POINTS, mPointsField));
			row.add(mPointsField);

			if (!mRow.isLeveled()) {
				mLevelField.setText(""); //$NON-NLS-1$
				mLevelField.setEnabled(false);
				mLevelPointsField.setText(""); //$NON-NLS-1$
				mLevelPointsField.setEnabled(false);
			}
		}

		mNotesField = createField(advantage.getNotes(), null, NOTES_TOOLTIP);
		innerGrid.add(new FlexComponent(createLabel(NOTES, mNotesField), Alignment.RIGHT_BOTTOM, null), ri, 0);
		innerGrid.add(mNotesField, ri++, 1);

		mCategoriesField = createField(advantage.getCategoriesAsString(), null, CATEGORIES_TOOLTIP);
		innerGrid.add(new FlexComponent(createLabel(CATEGORIES, mCategoriesField), Alignment.RIGHT_BOTTOM, null), ri, 0);
		innerGrid.add(mCategoriesField, ri++, 1);

		mCRCombo = new JComboBox<>(SelfControlRoll.values());
		mCRCombo.setSelectedIndex(mRow.getCR().ordinal());
		UIUtilities.setOnlySize(mCRCombo, mCRCombo.getPreferredSize());
		mCRCombo.setEnabled(mIsEditable);
		mCRCombo.addActionListener(this);
		add(mCRCombo);
		mCRAdjCombo = new JComboBox<>(SelfControlRollAdjustments.values());
		mCRAdjCombo.setToolTipText(CR_ADJ_TOOLTIP);
		mCRAdjCombo.setSelectedIndex(mRow.getCRAdj().ordinal());
		UIUtilities.setOnlySize(mCRAdjCombo, mCRAdjCombo.getPreferredSize());
		mCRAdjCombo.setEnabled(mIsEditable && mRow.getCR() != SelfControlRoll.NONE_REQUIRED);
		add(mCRAdjCombo);
		innerGrid.add(new FlexComponent(createLabel(CR, mCRCombo), Alignment.RIGHT_BOTTOM, null), ri, 0);
		FlexRow row = new FlexRow();
		row.add(mCRCombo);
		row.add(mCRAdjCombo);
		innerGrid.add(row, ri++, 1);

		row = new FlexRow();
		innerGrid.add(row, ri, 1);
		if (notContainer) {
			JLabel label = new JLabel(TYPE, SwingConstants.RIGHT);
			label.setToolTipText(TYPE_TOOLTIP);
			add(label);
			innerGrid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, null), ri++, 0);

			mMentalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_MENTAL) == Advantage.TYPE_MASK_MENTAL, MENTAL);
			row.add(mMentalType);
			row.add(createTypeLabel(GCSImages.getMentalTypeIcon(), mMentalType));

			mPhysicalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_PHYSICAL) == Advantage.TYPE_MASK_PHYSICAL, PHYSICAL);
			row.add(mPhysicalType);
			row.add(createTypeLabel(GCSImages.getPhysicalTypeIcon(), mPhysicalType));

			mSocialType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_SOCIAL) == Advantage.TYPE_MASK_SOCIAL, SOCIAL);
			row.add(mSocialType);
			row.add(createTypeLabel(GCSImages.getSocialTypeIcon(), mSocialType));

			mExoticType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_EXOTIC) == Advantage.TYPE_MASK_EXOTIC, EXOTIC);
			row.add(mExoticType);
			row.add(createTypeLabel(GCSImages.getExoticTypeIcon(), mExoticType));

			mSupernaturalType = createTypeCheckBox((mRow.getType() & Advantage.TYPE_MASK_SUPERNATURAL) == Advantage.TYPE_MASK_SUPERNATURAL, SUPERNATURAL);
			row.add(mSupernaturalType);
			row.add(createTypeLabel(GCSImages.getSupernaturalTypeIcon(), mSupernaturalType));
		} else {
			mContainerTypeCombo = new JComboBox<>(AdvantageContainerType.values());
			mContainerTypeCombo.setSelectedItem(mRow.getContainerType());
			UIUtilities.setOnlySize(mContainerTypeCombo, mContainerTypeCombo.getPreferredSize());
			mContainerTypeCombo.setToolTipText(CONTAINER_TYPE_TOOLTIP);
			add(mContainerTypeCombo);
			row.add(mContainerTypeCombo);
			innerGrid.add(new FlexComponent(new LinkedLabel(CONTAINER_TYPE, mContainerTypeCombo), Alignment.RIGHT_BOTTOM, null), ri++, 0);
		}

		row.add(new FlexSpacer(0, 0, true, false));

		mReferenceField = createField(mRow.getReference(), "MMMMMM", REFERENCE_TOOLTIP); //$NON-NLS-1$
		row.add(createLabel(REFERENCE, mReferenceField));
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

	private LinkedLabel createTypeLabel(StdImage icon, final JCheckBox linkTo) {
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
		EditorField field = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(min, max, false)), this, SwingConstants.LEFT, new Integer(value), new Integer(proto), tooltip);
		field.setEnabled(mIsEditable);
		add(field);
		return field;
	}

	@Override
	public boolean applyChangesSelf() {
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
				ArrayList<WeaponStats> list = new ArrayList<>(mMeleeWeapons.getWeapons());
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
		} else if (src == mModifiers) {
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
		return mLevelTypeCombo.getSelectedItem() == HAS_LEVELS;
	}

	private void levelTypeChanged() {
		boolean isLeveled = isLeveled();

		if (isLeveled) {
			mLevelField.setValue(new Integer(mLastLevel));
			mLevelPointsField.setValue(new Integer(mLastPointsPerLevel));
		} else {
			mLastLevel = getLevels();
			mLastPointsPerLevel = getPointsPerLevel();
			mLevelField.setText(""); //$NON-NLS-1$
			mLevelPointsField.setText(""); //$NON-NLS-1$
		}
		mLevelField.setEnabled(isLeveled);
		mLevelPointsField.setEnabled(isLeveled);
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
		return Advantage.getAdjustedPoints(getBasePoints(), isLeveled() ? getLevels() : 0, getPointsPerLevel(), getCR(), mModifiers.getAllModifiers());
	}

	private void updatePoints() {
		if (mPointsField != null) {
			mPointsField.setValue(new Integer(getPoints()));
		}
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
		LinkedLabel.setErrorMessage(mNameField, mNameField.getText().trim().length() != 0 ? null : NAME_CANNOT_BE_EMPTY);
	}

	@Override
	public void propertyChange(PropertyChangeEvent event) {
		if ("value".equals(event.getPropertyName())) { //$NON-NLS-1$
			Object src = event.getSource();
			if (src == mLevelField || src == mLevelPointsField || src == mBasePointsField) {
				updatePoints();
			}
		}
	}
}
