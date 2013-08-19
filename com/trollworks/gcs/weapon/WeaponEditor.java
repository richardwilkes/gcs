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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.skill.Defaults;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.EditorField;
import com.trollworks.gcs.widgets.IconButton;
import com.trollworks.gcs.widgets.LinkedLabel;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.WindowUtils;
import com.trollworks.gcs.widgets.layout.ColumnLayout;
import com.trollworks.gcs.widgets.layout.RowDistribution;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.Outline;
import com.trollworks.gcs.widgets.outline.OutlineModel;
import com.trollworks.gcs.widgets.outline.Row;
import com.trollworks.gcs.widgets.outline.Selection;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
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
	private static String					MSG_USAGE;
	private static String					MSG_DAMAGE;
	private static String					MSG_MINIMUM_STRENGTH;
	static final String						EMPTY	= "";			//$NON-NLS-1$
	private ListRow							mOwner;
	private Outline							mOutline;
	private IconButton						mAddButton;
	private JPanel							mEditorPanel;
	private EditorField						mUsage;
	private EditorField						mDamage;
	private EditorField						mStrength;
	private Defaults						mDefaults;
	private WeaponStats						mWeapon;
	private Class<? extends WeaponStats>	mWeaponClass;
	private boolean							mRespond;

	static {
		LocalizedMessages.initialize(WeaponEditor.class);
	}

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
		mAddButton = new IconButton(Images.getAddIcon());
		mAddButton.addActionListener(this);
		add(mAddButton, BorderLayout.WEST);
		add(createOutline(weapons, weaponClass), BorderLayout.NORTH);
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
		ArrayList<WeaponStats> weapons = new ArrayList<WeaponStats>();
		for (Row row : mOutline.getModel().getRows()) {
			weapons.add(((WeaponDisplayRow) row).getWeapon().clone(mOwner));
		}
		return weapons;
	}

	private Component createOutline(List<WeaponStats> weapons, Class<? extends WeaponStats> weaponClass) {
		mOutline = new Outline(false);
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
		mUsage = createTextField(wrapper, MSG_USAGE, EMPTY);
		mDamage = createTextField(wrapper, MSG_DAMAGE, EMPTY);
		createFields(wrapper);
		mStrength = createTextField(wrapper, MSG_MINIMUM_STRENGTH, EMPTY);
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
			mOutline.setMinimumSize(size);
		}
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

	public final void actionPerformed(ActionEvent event) {
		Object source = event.getSource();
		if (mAddButton == source) {
			addWeapon();
		} else if (mOutline == source) {
			handleOutline(event.getActionCommand());
		} else if (mRespond) {
			if (mDefaults == source) {
				changeDefaults();
			}
		}
	}

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
			if (selection.getCount() == 1) {
				setWeapon(((WeaponDisplayRow) model.getRowAtIndex(selection.firstSelectedIndex())).getWeapon());
			} else {
				setWeapon(null);
			}
		}
	}

	private void setWeapon(WeaponStats weapon) {
		if (weapon != mWeapon) {
			WindowUtils.forceFocusToAccept();
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
