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

package com.trollworks.gcs.ui.weapon;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMSkillDefault;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.model.weapon.CMRangedWeaponStats;
import com.trollworks.gcs.model.weapon.CMWeaponColumnID;
import com.trollworks.gcs.model.weapon.CMWeaponDisplayRow;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.gcs.ui.editor.defaults.CSDefaults;
import com.trollworks.toolkit.collections.TKFilteredIterator;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKLinkedLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKSelection;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKBaseButton;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.widget.layout.TKRowDistribution;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;

/** An editor for ranged weapon statistics. */
public class CSRangedWeaponEditor extends TKPanel implements ActionListener, TKMenuTarget {
	private static final String	EMPTY	= "";	//$NON-NLS-1$
	private CMRow				mOwner;
	private TKOutline			mOutline;
	private TKButton			mAddButton;
	private TKPanel				mEditorPanel;
	private TKTextField			mUsage;
	private TKTextField			mDamage;
	private TKTextField			mStrength;
	private TKTextField			mAccuracy;
	private TKTextField			mRange;
	private TKTextField			mRateOfFire;
	private TKTextField			mShots;
	private TKTextField			mBulk;
	private TKTextField			mRecoil;
	private CSDefaults			mDefaults;
	private CMRangedWeaponStats	mWeapon;

	/**
	 * Creates a new ranged weapon editor for the specified row.
	 * 
	 * @param row The row to edit ranged weapon statistics for.
	 * @return The editor, or <code>null</code> if the row is not appropriate.
	 */
	static public CSRangedWeaponEditor createEditor(CMRow row) {
		if (row instanceof CMEquipment) {
			return new CSRangedWeaponEditor(row, ((CMEquipment) row).getWeapons());
		} else if (row instanceof CMAdvantage) {
			return new CSRangedWeaponEditor(row, ((CMAdvantage) row).getWeapons());
		} else if (row instanceof CMSpell) {
			return new CSRangedWeaponEditor(row, ((CMSpell) row).getWeapons());
		} else if (row instanceof CMSkill) {
			return new CSRangedWeaponEditor(row, ((CMSkill) row).getWeapons());
		}
		return null;
	}

	/**
	 * Creates a new {@link CMRangedWeaponStats} editor.
	 * 
	 * @param owner The owning row.
	 * @param weapons The weapons to modify.
	 */
	public CSRangedWeaponEditor(CMRow owner, List<CMWeaponStats> weapons) {
		super(new TKCompassLayout());
		mOwner = owner;
		add(createOutline(weapons), TKCompassPosition.NORTH);
		add(createEditorPanel(), TKCompassPosition.CENTER);
	}

	/** @return The weapons in this editor. */
	public List<CMWeaponStats> getWeapons() {
		ArrayList<CMWeaponStats> weapons = new ArrayList<CMWeaponStats>();

		for (TKRow row : mOutline.getModel().getRows()) {
			CMRangedWeaponStats weapon = (CMRangedWeaponStats) ((CMWeaponDisplayRow) row).getWeapon();

			weapons.add(new CMRangedWeaponStats(mOwner, weapon));
		}
		return weapons;
	}

	private TKPanel createOutline(List<CMWeaponStats> weapons) {
		TKScrollPanel scroller;
		TKOutlineModel model;
		TKPanel borderView;
		Dimension minSize;

		mAddButton = new TKButton(CSImage.getAddIcon());
		mAddButton.addActionListener(this);

		mOutline = new TKOutline(false);
		model = mOutline.getModel();
		CMWeaponColumnID.addColumns(mOutline, CMRangedWeaponStats.class, true);
		mOutline.setAllowColumnDrag(false);
		mOutline.setAllowColumnResize(false);
		mOutline.setAllowRowDrag(false);
		mOutline.setMenuTargetDelegate(this);

		for (CMRangedWeaponStats weapon : new TKFilteredIterator<CMRangedWeaponStats>(weapons, CMRangedWeaponStats.class)) {
			model.addRow(new CMWeaponDisplayRow(new CMRangedWeaponStats(mOwner, weapon)));
		}
		mOutline.addActionListener(this);

		scroller = new TKScrollPanel(mOutline);
		borderView = scroller.getContentBorderView();
		scroller.setHorizontalHeader(mOutline.getHeaderPanel());
		scroller.setTopRightCorner(mAddButton);
		borderView.setBorder(new TKLineBorder(TKColor.SCROLL_BAR_LINE, 1, TKLineBorder.TOP_EDGE, false));
		scroller.setBorder(new TKLineBorder(TKColor.SCROLL_BAR_LINE, 1, TKLineBorder.LEFT_EDGE | TKLineBorder.TOP_EDGE | TKLineBorder.RIGHT_EDGE, false));
		minSize = borderView.getMinimumSize();
		if (minSize.height < 48) {
			minSize.height = 48;
		}
		borderView.setMinimumSize(minSize);
		return scroller;
	}

	private TKPanel createEditorPanel() {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));

		wrapper.setBorder(new TKEmptyBorder(5));

		mEditorPanel = new TKPanel(new TKColumnLayout(1, TKRowDistribution.GIVE_EXCESS_TO_LAST));
		mEditorPanel.setBorder(new TKLineBorder(TKColor.SCROLL_BAR_LINE, 1, TKLineBorder.LEFT_EDGE | TKLineBorder.BOTTOM_EDGE | TKLineBorder.RIGHT_EDGE, false));
		mEditorPanel.add(wrapper);

		mUsage = createTextField(wrapper, Msgs.USAGE);
		mDamage = createTextField(wrapper, Msgs.DAMAGE);
		mAccuracy = createTextField(wrapper, Msgs.ACCURACY);
		mRange = createTextField(wrapper, Msgs.RANGE);
		mRateOfFire = createTextField(wrapper, Msgs.RATE_OF_FIRE);
		mShots = createTextField(wrapper, Msgs.SHOTS);
		mStrength = createTextField(wrapper, Msgs.MINIMUM_STRENGTH);
		mRecoil = createTextField(wrapper, Msgs.RECOIL);
		mBulk = createTextField(wrapper, Msgs.BULK);

		createDefaults(mEditorPanel);
		setWeaponState(false);
		return mEditorPanel;
	}

	private void createDefaults(TKPanel parent) {
		TKScrollPanel scrollPanel;
		TKPanel borderView;
		Dimension minSize;

		mDefaults = new CSDefaults(new ArrayList<CMSkillDefault>());
		mDefaults.removeAll();
		scrollPanel = new TKScrollPanel(mDefaults);
		borderView = scrollPanel.getContentBorderView();
		borderView.setBorder(new TKLineBorder(TKLineBorder.LEFT_EDGE | TKLineBorder.TOP_EDGE));
		minSize = borderView.getMinimumSize();
		if (minSize.height < 60) {
			minSize.height = 60;
		}
		borderView.setMinimumSize(minSize);
		parent.add(scrollPanel);
	}

	private TKTextField createTextField(TKPanel panel, String name) {
		TKTextField field = new TKTextField(200);

		panel.add(new TKLinkedLabel(field, name));
		panel.add(field);
		return field;
	}

	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();

		if (mAddButton == source) {
			addWeapon();
		} else if (mOutline == source) {
			handleOutline(event.getActionCommand());
		} else if (mUsage == source) {
			changeUsage();
		} else if (mDamage == source) {
			changeDamage();
		} else if (mStrength == source) {
			changeStrength();
		} else if (mAccuracy == source) {
			changeAccuracy();
		} else if (mRange == source) {
			changeRange();
		} else if (mRateOfFire == source) {
			changeRateOfFire();
		} else if (mShots == source) {
			changeShots();
		} else if (mBulk == source) {
			changeBulk();
		} else if (mRecoil == source) {
			changeRecoil();
		} else if (mDefaults == source) {
			changeDefaults();
		}
	}

	private void changeDefaults() {
		mWeapon.setDefaults(mDefaults.getDefaults());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void changeUsage() {
		mWeapon.setUsage(mUsage.getText());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void changeDamage() {
		mWeapon.setDamage(mDamage.getText());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void changeStrength() {
		mWeapon.setStrength(mStrength.getText());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void changeAccuracy() {
		mWeapon.setAccuracy(mAccuracy.getText());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void changeRange() {
		mWeapon.setRange(mRange.getText());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void changeRateOfFire() {
		mWeapon.setRateOfFire(mRateOfFire.getText());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void changeShots() {
		mWeapon.setShots(mShots.getText());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void changeBulk() {
		mWeapon.setBulk(mBulk.getText());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void changeRecoil() {
		mWeapon.setRecoil(mRecoil.getText());
		mOutline.sizeColumnsToFit();
		mOutline.repaintSelection();
	}

	private void addWeapon() {
		CMWeaponDisplayRow weapon = new CMWeaponDisplayRow(new CMRangedWeaponStats(mOwner));
		TKOutlineModel model = mOutline.getModel();

		model.addRow(weapon);
		mOutline.sizeColumnsToFit();
		model.select(weapon, false);
		mOutline.revalidateImmediately();
		mOutline.scrollSelectionIntoView();
		mOutline.requestFocus();
	}

	private void handleOutline(String cmd) {
		if (TKOutline.CMD_SELECTION_CHANGED.equals(cmd)) {
			TKOutlineModel model = mOutline.getModel();
			TKSelection selection = model.getSelection();

			if (selection.getCount() == 1) {
				setWeapon((CMRangedWeaponStats) ((CMWeaponDisplayRow) model.getRowAtIndex(selection.firstSelectedIndex())).getWeapon());
			} else {
				setWeapon(null);
			}
		} else if (TKOutline.CMD_DELETE_SELECTION.equals(cmd)) {
			if (canClear()) {
				clear();
			} else {
				getToolkit().beep();
			}
		}
	}

	private void setWeapon(CMRangedWeaponStats weapon) {
		if (weapon != mWeapon) {
			forceFocusToAccept();
			mWeapon = weapon;
			setWeaponState(mWeapon != null);
		}
	}

	private void setWeaponState(boolean enabled) {
		mUsage.removeActionListener(this);
		mDamage.removeActionListener(this);
		mStrength.removeActionListener(this);
		mAccuracy.removeActionListener(this);
		mRange.removeActionListener(this);
		mRateOfFire.removeActionListener(this);
		mShots.removeActionListener(this);
		mBulk.removeActionListener(this);
		mRecoil.removeActionListener(this);
		mDefaults.removeActionListener(this);

		if (enabled) {
			mUsage.setText(mWeapon.getUsage());
			mDamage.setText(mWeapon.getDamage());
			mStrength.setText(mWeapon.getStrength());
			mAccuracy.setText(mWeapon.getAccuracy());
			mRange.setText(mWeapon.getRange());
			mRateOfFire.setText(mWeapon.getRateOfFire());
			mShots.setText(mWeapon.getShots());
			mBulk.setText(mWeapon.getBulk());
			mRecoil.setText(mWeapon.getRecoil());
			mDefaults.setDefaults(mWeapon.getDefaults());

			if (mAddButton.isEnabled()) {
				mUsage.addActionListener(this);
				mDamage.addActionListener(this);
				mStrength.addActionListener(this);
				mAccuracy.addActionListener(this);
				mRange.addActionListener(this);
				mRateOfFire.addActionListener(this);
				mShots.addActionListener(this);
				mBulk.addActionListener(this);
				mRecoil.addActionListener(this);
				mDefaults.addActionListener(this);
			} else {
				disableControls(mDefaults);
			}
		} else {
			mUsage.setText(EMPTY);
			mDamage.setText(EMPTY);
			mStrength.setText(EMPTY);
			mAccuracy.setText(EMPTY);
			mRange.setText(EMPTY);
			mRateOfFire.setText(EMPTY);
			mShots.setText(EMPTY);
			mBulk.setText(EMPTY);
			mRecoil.setText(EMPTY);
			mDefaults.removeAll();
			mDefaults.revalidate();
			mDefaults.repaint();
		}

		if (!mAddButton.isEnabled()) {
			enabled = false;
		}
		mUsage.setEnabled(enabled);
		mDamage.setEnabled(enabled);
		mStrength.setEnabled(enabled);
		mAccuracy.setEnabled(enabled);
		mRange.setEnabled(enabled);
		mRateOfFire.setEnabled(enabled);
		mShots.setEnabled(enabled);
		mBulk.setEnabled(enabled);
		mRecoil.setEnabled(enabled);
	}

	private void disableControls(TKPanel panel) {
		int count = panel.getComponentCount();

		for (int i = 0; i < count; i++) {
			disableControls((TKPanel) panel.getComponent(i));
		}

		if (panel instanceof TKBaseButton || panel instanceof TKTextField || panel instanceof TKPopupMenu) {
			panel.setEnabled(false);
		}
	}

	@Override public String toString() {
		return Msgs.RANGED_WEAPON;
	}

	public void menusWillBeAdjusted() {
		// Not used.
	}

	public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (TKWindow.CMD_CLEAR.equals(command)) {
			item.setEnabled(canClear());
			return true;
		}
		return false;
	}

	public void menusWereAdjusted() {
		// Not used.
	}

	public boolean obeyCommand(String command, TKMenuItem item) {
		if (TKWindow.CMD_CLEAR.equals(command)) {
			clear();
			return true;
		}
		return false;
	}

	private boolean canClear() {
		return mOutline.getModel().hasSelection();
	}

	private void clear() {
		if (canClear()) {
			mOutline.getModel().removeSelection();
			mOutline.sizeColumnsToFit();
		}
	}
}
