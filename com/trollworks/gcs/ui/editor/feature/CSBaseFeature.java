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

package com.trollworks.gcs.ui.editor.feature;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.criteria.CMStringCompareType;
import com.trollworks.gcs.model.criteria.CMStringCriteria;
import com.trollworks.gcs.model.feature.CMAttributeBonus;
import com.trollworks.gcs.model.feature.CMBonus;
import com.trollworks.gcs.model.feature.CMCostReduction;
import com.trollworks.gcs.model.feature.CMDRBonus;
import com.trollworks.gcs.model.feature.CMFeature;
import com.trollworks.gcs.model.feature.CMLeveledAmount;
import com.trollworks.gcs.model.feature.CMSkillBonus;
import com.trollworks.gcs.model.feature.CMSpellBonus;
import com.trollworks.gcs.model.feature.CMWeaponBonus;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;

/** A generic feature editor panel. */
public abstract class CSBaseFeature extends TKPanel implements ActionListener, TKMenuTarget {
	private static final String	ADD					= "Add";					//$NON-NLS-1$
	private static final String	REMOVE				= "Remove";				//$NON-NLS-1$
	private static final String	CHANGE_BASE_TYPE	= "ChangeBaseType";		//$NON-NLS-1$
	private static final String	QUALIFIER			= "Qualifier";				//$NON-NLS-1$
	private static String		sLastItemType		= CMSkillBonus.TAG_ROOT;
	private CMRow				mRow;
	private CMFeature			mFeature;
	/** The center panel. */
	protected TKPanel			mCenter;
	/** The right-side panel. */
	protected TKPanel			mRight;

	/** @param type The last item type created or switched to. */
	public static void setLastItemType(String type) {
		sLastItemType = type;
	}

	/**
	 * Creates a new feature editor panel.
	 * 
	 * @param row The row this feature will belong to.
	 * @param feature The feature to edit.
	 * @return The newly created editor panel.
	 */
	public static CSBaseFeature create(CMRow row, CMFeature feature) {
		if (feature instanceof CMAttributeBonus) {
			return new CSAttributeBonus(row, (CMAttributeBonus) feature);
		}
		if (feature instanceof CMDRBonus) {
			return new CSDRBonus(row, (CMDRBonus) feature);
		}
		if (feature instanceof CMSkillBonus) {
			return new CSSkillBonus(row, (CMSkillBonus) feature);
		}
		if (feature instanceof CMSpellBonus) {
			return new CSSpellBonus(row, (CMSpellBonus) feature);
		}
		if (feature instanceof CMWeaponBonus) {
			return new CSWeaponBonus(row, (CMWeaponBonus) feature);
		}
		if (feature instanceof CMCostReduction) {
			return new CSCostReduction(row, (CMCostReduction) feature);
		}
		return null;
	}

	/**
	 * Creates a new feature editor panel.
	 * 
	 * @param row The row this feature will belong to.
	 * @param feature The feature to edit.
	 */
	public CSBaseFeature(CMRow row, CMFeature feature) {
		super(new TKColumnLayout(2));
		setBorder(new TKEmptyBorder(0, TKColumnLayout.DEFAULT_H_GAP_SIZE, 0, 0));
		mRow = row;
		mFeature = feature;
		mCenter = new TKPanel(new TKColumnLayout());
		mRight = new TKPanel();
		mCenter.setAlignmentY(TOP_ALIGNMENT);
		mRight.setAlignmentY(CENTER_ALIGNMENT);
		add(mCenter);
		add(mRight);
		rebuild();
	}

	/** Rebuilds the contents of this panel with the current feature settings. */
	protected void rebuild() {
		mCenter.removeAll();
		mRight.removeAll();

		rebuildSelf();

		addButton(CSImage.getAddIcon(), ADD, Msgs.ADD_FEATURE_TOOLTIP);
		if (mFeature != null) {
			addButton(CSImage.getRemoveIcon(), REMOVE, Msgs.REMOVE_FEATURE_TOOLTIP);
		}

		mRight.setLayout(new TKColumnLayout(Math.max(mRight.getComponentCount(), 1)));
		mCenter.invalidate();
		mRight.invalidate();
		revalidate();
	}

	/**
	 * Sub-classes must implement this method to add any components they want to be visible.
	 */
	protected abstract void rebuildSelf();

	/**
	 * Adds a button.
	 * 
	 * @param icon The image to use.
	 * @param command The command to use.
	 * @param tooltip The tooltip to use.
	 */
	protected void addButton(BufferedImage icon, String command, String tooltip) {
		TKButton button = new TKButton(icon);

		button.setActionCommand(command);
		button.addActionListener(this);
		button.setToolTipText(tooltip);
		button.setOnlySize(16, 16);
		mRight.add(button);
	}

	/**
	 * Adds the popup that allows the base feature type to be changed.
	 * 
	 * @param parent The panel to add the popup to.
	 */
	protected void addChangeBaseTypePopup(TKPanel parent) {
		String[] keys = { CMAttributeBonus.TAG_ROOT, CMDRBonus.TAG_ROOT, CMSkillBonus.TAG_ROOT, CMSpellBonus.TAG_ROOT, CMWeaponBonus.TAG_ROOT, CMCostReduction.TAG_ROOT };
		String[] titles = { Msgs.ATTRIBUTE_BONUS, Msgs.DR_BONUS, Msgs.SKILL_BONUS, Msgs.SPELL_BONUS, Msgs.WEAPON_BONUS, Msgs.COST_REDUCTION };
		TKMenu menu = new TKMenu();
		int selection = 0;
		String current = mFeature.getXMLTag();
		TKPopupMenu popup;

		for (int i = 0; i < keys.length; i++) {
			addChangeBaseTypeMenuItem(menu, titles[i], keys[i]);
			if (keys[i].equals(current)) {
				selection = i;
			}
		}
		popup = new TKPopupMenu(menu, this, false, selection);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	private void addChangeBaseTypeMenuItem(TKMenu menu, String title, String key) {
		TKMenuItem item = new TKMenuItem(title, CHANGE_BASE_TYPE);

		item.setUserObject(key);
		menu.add(item);
	}

	/**
	 * Adds the popup that allows an integer comparison to be changed.
	 * 
	 * @param parent The panel to add the text field and popup to.
	 * @param amt The current leveled amount object.
	 * @param maxDigits The maximum number of digits to allow.
	 * @param decimal Whether decimal values need to be allowed.
	 */
	protected void addLeveledAmountPopups(TKPanel parent, CMLeveledAmount amt, int maxDigits, boolean decimal) {
		addLeveledAmountPopups(parent, amt, maxDigits, decimal, false);
	}

	/**
	 * Adds the popup that allows an integer comparison to be changed.
	 * 
	 * @param parent The panel to add the text field and popup to.
	 * @param amt The current leveled amount object.
	 * @param maxDigits The maximum number of digits to allow.
	 * @param decimal Whether decimal values need to be allowed.
	 * @param usePerDie Whether to use the "per die" message or the "per level" message.
	 */
	protected void addLeveledAmountPopups(TKPanel parent, CMLeveledAmount amt, int maxDigits, boolean decimal, boolean usePerDie) {
		TKMenu menu = new TKMenu();
		TKTextField field = new TKTextField((decimal ? "+." : "+") + TKTextUtility.makeFiller(maxDigits, 'M')); //$NON-NLS-1$ //$NON-NLS-2$
		TKPopupMenu popup;

		field.setOnlySize(field.getPreferredSize());
		field.setText(amt.getAmountAsString());
		field.setKeyEventFilter(new TKNumberFilter(decimal, true, maxDigits));
		field.setActionCommand(CMBonus.TAG_AMOUNT);
		field.addActionListener(this);
		parent.add(field);

		addLeveledAmountMenuItem(menu, " ", Boolean.FALSE); //$NON-NLS-1$
		addLeveledAmountMenuItem(menu, usePerDie ? Msgs.PER_DIE : Msgs.PER_LEVEL, Boolean.TRUE);
		popup = new TKPopupMenu(menu, this, false, amt.isPerLevel() ? 1 : 0);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);

	}

	private void addLeveledAmountMenuItem(TKMenu menu, String msg, Boolean perLevel) {
		TKMenuItem item = new TKMenuItem(msg, CMLeveledAmount.ATTRIBUTE_PER_LEVEL);

		item.setUserObject(perLevel);
		menu.add(item);
	}

	/**
	 * Adds the popup that allows a string comparison to be changed.
	 * 
	 * @param parent The panel to add the popup and text field to.
	 * @param compare The current string compare object.
	 * @param extra The extra text to add to the menu item.
	 * @param keyPrefix The prefix to add in front of all command keys.
	 */
	protected void addStringComparePopups(TKPanel parent, CMStringCriteria compare, String extra, String keyPrefix) {
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;
		TKTextField field;

		for (CMStringCompareType type : CMStringCompareType.values()) {
			String title = type.toString();

			if (extra != null) {
				title = extra + title;
			}
			menu.add(new TKMenuItem(title, keyPrefix + type.name()));
		}
		popup = new TKPopupMenu(menu, this, false, compare.getType().ordinal());
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);

		field = new TKTextField(compare.getQualifier(), 50);
		field.setActionCommand(keyPrefix + QUALIFIER);
		field.addActionListener(this);
		parent.add(field);
	}

	/**
	 * Call to handle a change in one of the string comparison controls.
	 * 
	 * @param criteria The current string compare object.
	 * @param command The command without the prefix.
	 * @param event In the case of a change to the qualifier field, the {@link ActionEvent}
	 *            associated with the notification.
	 */
	protected void handleStringCompareChange(CMStringCriteria criteria, String command, ActionEvent event) {
		if (QUALIFIER.equals(command)) {
			criteria.setQualifier(((TKTextField) event.getSource()).getText());
		} else {
			criteria.setType(CMStringCompareType.get(command));
		}
	}

	/** @return The underlying feature. */
	public CMFeature getFeature() {
		return mFeature;
	}

	public void menusWillBeAdjusted() {
		// Not used.
	}

	public boolean adjustMenuItem(String command, TKMenuItem item) {
		return true;
	}

	public void menusWereAdjusted() {
		// Not used.
	}

	public boolean obeyCommand(String command, TKMenuItem item) {
		if (CHANGE_BASE_TYPE.equals(command)) {
			String current = mFeature.getXMLTag();
			String value = (String) item.getUserObject();

			if (!current.equals(value)) {
				forceFocusToAccept();
				replace(transform(value));
			}
			setLastItemType(value);
		} else if (CMLeveledAmount.ATTRIBUTE_PER_LEVEL.equals(command)) {
			((CMBonus) mFeature).getAmount().setPerLevel(((Boolean) item.getUserObject()).booleanValue());
		} else {
			return false;
		}
		return true;
	}

	private void replace(CMFeature prereq) {
		TKPanel parent = (TKPanel) getParent();
		int pIndex = parent.getIndexOf(this);

		parent.add(create(mRow, prereq), pIndex);
		removeFromParent();
		parent.revalidate();
	}

	private CMFeature transform(String to) {
		CMFeature feature;

		if (CMAttributeBonus.TAG_ROOT.equals(to)) {
			feature = new CMAttributeBonus();
		} else if (CMDRBonus.TAG_ROOT.equals(to)) {
			feature = new CMDRBonus();
		} else if (CMSkillBonus.TAG_ROOT.equals(to)) {
			feature = new CMSkillBonus();
		} else if (CMSpellBonus.TAG_ROOT.equals(to)) {
			feature = new CMSpellBonus();
		} else if (CMWeaponBonus.TAG_ROOT.equals(to)) {
			feature = new CMWeaponBonus();
		} else if (CMCostReduction.TAG_ROOT.equals(to)) {
			feature = new CMCostReduction();
		} else {
			return null;
		}
		return feature;
	}

	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		TKPanel parent = (TKPanel) getParent();

		if (ADD.equals(command)) {
			CMFeature feature;

			if (CMAttributeBonus.TAG_ROOT == sLastItemType) {
				feature = new CMAttributeBonus();
			} else if (CMDRBonus.TAG_ROOT == sLastItemType) {
				feature = new CMDRBonus();
			} else if (CMSkillBonus.TAG_ROOT == sLastItemType) {
				feature = new CMSkillBonus();
			} else if (CMSpellBonus.TAG_ROOT == sLastItemType) {
				feature = new CMSpellBonus();
			} else if (CMWeaponBonus.TAG_ROOT == sLastItemType) {
				feature = new CMWeaponBonus();
			} else if (CMCostReduction.TAG_ROOT == sLastItemType) {
				feature = new CMCostReduction();
			} else {
				// Default
				feature = new CMSkillBonus();
			}
			parent.add(create(mRow, feature));
			if (mFeature == null) {
				removeFromParent();
			}
			parent.revalidate();
		} else if (REMOVE.equals(command)) {
			removeFromParent();
			if (parent.getComponentCount() == 0) {
				parent.add(new CSNoFeature(mRow));
			}
			parent.revalidate();
		} else if (CMBonus.TAG_AMOUNT.equals(command)) {
			CMLeveledAmount amt = ((CMBonus) mFeature).getAmount();
			String text = ((TKTextField) event.getSource()).getText();

			if (amt.isIntegerOnly()) {
				amt.setAmount(TKNumberUtils.getInteger(text, 0));
			} else {
				amt.setAmount(TKNumberUtils.getDouble(text, 0.0));
			}
		}
	}
}
