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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.skill.Defaults;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.Selection;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.layout.PrecisionLayout;
import com.trollworks.toolkit.ui.layout.RowDistribution;
import com.trollworks.toolkit.ui.widget.EditorField;
import com.trollworks.toolkit.ui.widget.IconButton;
import com.trollworks.toolkit.ui.widget.LinkedLabel;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.utility.Localization;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Panel;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.ArrayList;
import java.util.List;

import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.SwingConstants;
import javax.swing.border.EmptyBorder;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** An abstract editor for weapon statistics. */
public abstract class WeaponEditor extends JPanel implements ActionListener, PropertyChangeListener {
	@Localize("Usage")
	@Localize(locale = "de", value = "Nutzungsart")
	@Localize(locale = "ru", value = "Применение")
	private static String					USAGE;
	@Localize("Damage")
	@Localize(locale = "de", value = "Schaden")
	@Localize(locale = "ru", value = "Повреждения")
	private static String					DAMAGE;
	@Localize("Minimum Strength")
	@Localize(locale = "de", value = "Mindeststärke")
	@Localize(locale = "ru", value = "Минимальная сила")
	private static String					MINIMUM_STRENGTH;
	@Localize("Add an attack")
	@Localize(locale = "de", value = "Angriff hinzufügen")
	@Localize(locale = "ru", value = "Добавить атаку")
	private static String					ADD_TOOLTIP;
	@Localize("Remove the selected attacks")
	@Localize(locale = "de", value = "Ausgewählte Angriffe entfernen")
	@Localize(locale = "ru", value = "Удалить выбранные атаки")
	private static String					REMOVE_TOOLTIP;

	static {
		Localization.initialize();
	}

	protected static final String			EMPTY	= "";		//$NON-NLS-1$
	private ListRow							mOwner;
	private WeaponOutline					mOutline;
	private IconButton						mAddButton;
	private IconButton						mDeleteButton;
	private JPanel							mEditorPanel;
	private EditorField						mUsage;
	private EditorField						mDamage;
	private EditorField						mStrength;
	private Defaults						mDefaults;
	private WeaponStats						mWeapon;
	private Class<? extends WeaponStats>	mWeaponClass;
	private boolean							mRespond;

	/**
	 * Creates a new {@link WeaponEditor}.
	 *
	 * @param owner The owning row.
	 * @param weapons The weapons to modify.
	 * @param weaponClass The {@link Class} of weapons.
	 */
	public WeaponEditor(ListRow owner, List<WeaponStats> weapons, Class<? extends WeaponStats> weaponClass) {
		super(new BorderLayout());
		mOwner = owner;
		mWeaponClass = weaponClass;
		mAddButton = new IconButton(StdImage.ADD, ADD_TOOLTIP, () -> addWeapon());
		mDeleteButton = new IconButton(StdImage.REMOVE, REMOVE_TOOLTIP, () -> mOutline.deleteSelection());
		mDeleteButton.setEnabled(false);
		Panel top = new Panel(new BorderLayout());
		Panel left = new Panel(new PrecisionLayout());
		left.add(mAddButton);
		left.add(mDeleteButton);
		top.add(left, BorderLayout.WEST);
		top.add(createOutline(weapons, weaponClass), BorderLayout.CENTER);
		add(top, BorderLayout.NORTH);
		add(createEditorPanel(), BorderLayout.CENTER);
		setName(toString());
	}

	/** @return The owner. */
	public ListRow getOwner() {
		return mOwner;
	}

	/** @return The weapon. */
	protected WeaponStats getWeapon() {
		return mWeapon;
	}

	/** @return The weapons in this editor. */
	public List<WeaponStats> getWeapons() {
		ArrayList<WeaponStats> weapons = new ArrayList<>();
		for (Row row : mOutline.getModel().getRows()) {
			weapons.add(((WeaponDisplayRow) row).getWeapon().clone(mOwner));
		}
		return weapons;
	}

	private Component createOutline(List<WeaponStats> weapons, Class<? extends WeaponStats> weaponClass) {
		mOutline = new WeaponOutline();
		OutlineModel model = mOutline.getModel();
		WeaponColumn.addColumns(mOutline, weaponClass, true);
		mOutline.setAllowColumnDrag(false);
		mOutline.setAllowColumnResize(false);
		mOutline.setAllowRowDrag(false);

		for (WeaponStats weapon : weapons) {
			if (mWeaponClass.isInstance(weapon)) {
				model.addRow(new WeaponDisplayRow(weapon.clone(mOwner)));
			}
		}

		mOutline.addActionListener(this);
		Dimension size = mOutline.getMinimumSize();
		if (size.height < 50) {
			size.height = 50;
			mOutline.setMinimumSize(size);
		}

		JScrollPane scroller = new JScrollPane(mOutline);
		scroller.setColumnHeaderView(mOutline.getHeaderPanel());
		return scroller;
	}

	private Container createEditorPanel() {
		JPanel wrapper = new JPanel(new ColumnLayout(4));
		wrapper.setBorder(new EmptyBorder(5, 5, 5, 5));
		mEditorPanel = new JPanel(new ColumnLayout(1, RowDistribution.GIVE_EXCESS_TO_LAST));
		mEditorPanel.add(wrapper);
		mUsage = createTextField(wrapper, USAGE, EMPTY);
		mDamage = createTextField(wrapper, DAMAGE, EMPTY);
		createFields(wrapper);
		mStrength = createTextField(wrapper, MINIMUM_STRENGTH, EMPTY);
		createDefaults(mEditorPanel);
		setWeaponState(false);
		return mEditorPanel;
	}

	/**
	 * Called to add the necessary fields.
	 *
	 * @param parent The parent to add the fields to.
	 */
	protected abstract void createFields(Container parent);

	private void createDefaults(Container parent) {
		mDefaults = new Defaults(new ArrayList<SkillDefault>());
		mDefaults.removeAll();
		mDefaults.addActionListener(this);
		JScrollPane scrollPanel = new JScrollPane(mDefaults);
		Dimension size = mDefaults.getMinimumSize();
		if (size.height < 50) {
			size.height = 50;
			mDefaults.setMinimumSize(size);
		}
		scrollPanel.setPreferredSize(new Dimension(50, 50));
		parent.add(scrollPanel);
	}

	/**
	 * Creates a new text field.
	 *
	 * @param parent The parent.
	 * @param title The title of the field.
	 * @param value The initial value.
	 * @return The newly created field.
	 */
	protected EditorField createTextField(Container parent, String title, Object value) {
		DefaultFormatter formatter = new DefaultFormatter();
		formatter.setOverwriteMode(false);
		EditorField field = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, value, null);
		parent.add(new LinkedLabel(title, field));
		parent.add(field);
		return field;
	}

	@Override
	public final void actionPerformed(ActionEvent event) {
		Object source = event.getSource();
		if (mOutline == source) {
			handleOutline(event.getActionCommand());
		} else if (mRespond) {
			if (mDefaults == source) {
				changeDefaults();
			}
		}
	}

	@Override
	public void propertyChange(PropertyChangeEvent event) {
		if (mRespond) {
			Object source = event.getSource();
			if (mUsage == source) {
				changeUsage();
			} else if (mDamage == source) {
				changeDamage();
			} else if (mStrength == source) {
				changeStrength();
			} else {
				updateFromField(source);
			}
		}
	}

	/**
	 * Called to update from a field.
	 *
	 * @param field The field that was altered.
	 */
	protected abstract void updateFromField(Object field);

	private void changeDefaults() {
		mWeapon.setDefaults(mDefaults.getDefaults());
		adjustOutlineToContent();
	}

	private void changeUsage() {
		mWeapon.setUsage((String) mUsage.getValue());
		adjustOutlineToContent();
	}

	private void changeDamage() {
		mWeapon.setDamage((String) mDamage.getValue());
		adjustOutlineToContent();
	}

	private void changeStrength() {
		mWeapon.setStrength((String) mStrength.getValue());
		adjustOutlineToContent();
	}

	/** Call to adjust the {@link Outline} to its new content. */
	protected void adjustOutlineToContent() {
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void addWeapon() {
		WeaponDisplayRow weapon = new WeaponDisplayRow(createWeaponStats());
		OutlineModel model = mOutline.getModel();
		model.addRow(weapon);
		mOutline.sizeColumnsToFit();
		model.select(weapon, false);
		mOutline.revalidate();
		mOutline.scrollSelectionIntoView();
		mOutline.requestFocus();
	}

	/** @return A newly created {@link WeaponStats}. */
	protected abstract WeaponStats createWeaponStats();

	private void handleOutline(String cmd) {
		if (Outline.CMD_SELECTION_CHANGED.equals(cmd)) {
			OutlineModel model = mOutline.getModel();
			Selection selection = model.getSelection();
			int count = selection.getCount();
			if (count == 1) {
				setWeapon(((WeaponDisplayRow) model.getRowAtIndex(selection.firstSelectedIndex())).getWeapon());
			} else {
				setWeapon(null);
			}
			mDeleteButton.setEnabled(count > 0);
		}
	}

	private void setWeapon(WeaponStats weapon) {
		if (weapon != mWeapon) {
			UIUtilities.forceFocusToAccept();
			mWeapon = weapon;
			setWeaponState(mWeapon != null);
		}
	}

	private void setWeaponState(boolean enabled) {
		mRespond = false;
		if (enabled) {
			updateFields();
			if (mAddButton.isEnabled()) {
				mRespond = true;
			} else {
				UIUtilities.disableControls(mDefaults);
			}
		} else {
			blankFields();
		}
		if (!mAddButton.isEnabled()) {
			enabled = false;
		}
		enableFields(enabled);
	}

	/** Called to update the contents of the fields. */
	protected void updateFields() {
		mUsage.setValue(mWeapon.getUsage());
		mDamage.setValue(mWeapon.getDamage());
		mStrength.setValue(mWeapon.getStrength());
		mDefaults.setDefaults(mWeapon.getDefaults());
	}

	/**
	 * Called to enable/disable all fields.
	 *
	 * @param enabled Whether to enable the fields.
	 */
	protected void enableFields(boolean enabled) {
		mUsage.setEnabled(enabled);
		mDamage.setEnabled(enabled);
		mStrength.setEnabled(enabled);
	}

	/** Called to blank all fields. */
	protected void blankFields() {
		mUsage.setValue(EMPTY);
		mDamage.setValue(EMPTY);
		mStrength.setValue(EMPTY);
		mDefaults.removeAll();
		mDefaults.revalidate();
		mDefaults.repaint();
	}
}
