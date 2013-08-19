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

package com.trollworks.gcs.ui.editor.prereq;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.criteria.CMDoubleCriteria;
import com.trollworks.gcs.model.criteria.CMIntegerCriteria;
import com.trollworks.gcs.model.criteria.CMNumericCriteria;
import com.trollworks.gcs.model.criteria.CMStringCompareType;
import com.trollworks.gcs.model.criteria.CMStringCriteria;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.prereq.CMAdvantagePrereq;
import com.trollworks.gcs.model.prereq.CMAttributePrereq;
import com.trollworks.gcs.model.prereq.CMContainedWeightPrereq;
import com.trollworks.gcs.model.prereq.CMHasPrereq;
import com.trollworks.gcs.model.prereq.CMPrereq;
import com.trollworks.gcs.model.prereq.CMPrereqList;
import com.trollworks.gcs.model.prereq.CMSkillPrereq;
import com.trollworks.gcs.model.prereq.CMSpellPrereq;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;

import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.text.MessageFormat;

/** A generic prerequisite editor panel. */
public abstract class CSBasePrereq extends TKPanel implements ActionListener, TKMenuTarget {
	private static final String		REMOVE				= "Remove";		//$NON-NLS-1$
	private static final String		HAS					= "Has";			//$NON-NLS-1$
	private static final String		DOES_NOT_HAVE		= "DoesNotHave";	//$NON-NLS-1$
	private static final String		CHANGE_BASE_TYPE	= "ChangeBaseType"; //$NON-NLS-1$
	/** The key for the qualifier action. */
	protected static final String	QUALIFIER_KEY		= "Qualifier";		//$NON-NLS-1$
	/** The prerequisite this panel represents. */
	protected CMPrereq				mPrereq;
	/** The row this prerequisite will be attached to. */
	protected CMRow					mRow;
	private int						mDepth;
	/** The left-side panel. */
	protected TKPanel				mLeft;
	/** The center panel. */
	protected TKPanel				mCenter;
	/** The right-side panel. */
	protected TKPanel				mRight;

	/**
	 * Creates a new prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 * @return The newly created editor panel.
	 */
	public static CSBasePrereq create(CMRow row, CMPrereq prereq, int depth) {
		if (prereq instanceof CMPrereqList) {
			return new CSListPrereq(row, (CMPrereqList) prereq, depth);
		}
		if (prereq instanceof CMAdvantagePrereq) {
			return new CSAdvantagePrereq(row, (CMAdvantagePrereq) prereq, depth);
		}
		if (prereq instanceof CMSkillPrereq) {
			return new CSSkillPrereq(row, (CMSkillPrereq) prereq, depth);
		}
		if (prereq instanceof CMSpellPrereq) {
			return new CSSpellPrereq(row, (CMSpellPrereq) prereq, depth);
		}
		if (prereq instanceof CMAttributePrereq) {
			return new CSAttributePrereq(row, (CMAttributePrereq) prereq, depth);
		}
		if (prereq instanceof CMContainedWeightPrereq) {
			return new CSContainedWeightPrereq(row, (CMContainedWeightPrereq) prereq, depth);
		}
		return null;
	}

	/**
	 * Creates a new generic prerequisite editor panel.
	 * 
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	protected CSBasePrereq(CMRow row, CMPrereq prereq, int depth) {
		super(new TKColumnLayout(3));
		mRow = row;
		mPrereq = prereq;
		mDepth = depth;
		mLeft = new TKPanel();
		mCenter = new TKPanel(new TKColumnLayout());
		mRight = new TKPanel();
		mLeft.setAlignmentY(TOP_ALIGNMENT);
		mCenter.setAlignmentY(TOP_ALIGNMENT);
		mRight.setAlignmentY(CENTER_ALIGNMENT);
		add(mLeft);
		add(mCenter);
		add(mRight);
		setBorder(new TKEmptyBorder(0, TKColumnLayout.DEFAULT_H_GAP_SIZE + mDepth * 20, 0, 0));
		rebuild();
	}

	/** Rebuilds the contents of this panel with the current prereq settings. */
	protected void rebuild() {
		mLeft.removeAll();
		mCenter.removeAll();
		mRight.removeAll();

		if (mPrereq.getParent() != null) {
			mLeft.add(new AndOrLabel());
		}
		rebuildSelf();
		if (mDepth > 0) {
			addButton(CSImage.getRemoveIcon(), REMOVE, mPrereq instanceof CMPrereqList ? Msgs.REMOVE_PREREQ_LIST_TOOLTIP : Msgs.REMOVE_PREREQ_TOOLTIP);
		}
		mLeft.setLayout(new TKColumnLayout(Math.max(mLeft.getComponentCount(), 1)));
		mRight.setLayout(new TKColumnLayout(Math.max(mRight.getComponentCount(), 1)));
		mLeft.invalidate();
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
	 * Adds the popup that allows the "has" attribute to be changed.
	 * 
	 * @param parent The panel to add the popup to.
	 * @param has The current value of the "has" attribute.
	 */
	protected void addHasPopup(TKPanel parent, boolean has) {
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;

		addHasMenuItem(menu, Msgs.HAS, HAS);
		addHasMenuItem(menu, Msgs.DOES_NOT_HAVE, DOES_NOT_HAVE);
		popup = new TKPopupMenu(menu, this, false, has ? 0 : 1);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);
	}

	private void addHasMenuItem(TKMenu menu, String title, String key) {
		menu.add(new TKMenuItem(title, key));
	}

	/**
	 * Adds the popup that allows the base prereq type to be changed.
	 * 
	 * @param parent The panel to add the popup to.
	 */
	protected void addChangeBaseTypePopup(TKPanel parent) {
		String[] keys = { CMAttributePrereq.TAG_ROOT, CMAdvantagePrereq.TAG_ROOT, CMSkillPrereq.TAG_ROOT, CMSpellPrereq.TAG_ROOT, CMContainedWeightPrereq.TAG_ROOT };
		String[] titles = { Msgs.ATTRIBUTE, Msgs.ADVANTAGE, Msgs.SKILL, Msgs.SPELL, Msgs.CONTAINED_WEIGHT };
		TKMenu menu = new TKMenu();
		int selection = 0;
		String current = mPrereq.getXMLTag();
		TKPopupMenu popup;

		for (int i = 0; i < keys.length; i++) {
			if (i < keys.length - 1 || mRow instanceof CMEquipment) {
				addChangeBaseTypeMenuItem(menu, titles[i], keys[i]);
				if (keys[i].equals(current)) {
					selection = i;
				}
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
	 * Adds the popup that allows a string comparison to be changed.
	 * 
	 * @param parent The panel to add the popup and text field to.
	 * @param compare The current string compare object.
	 * @param extra The extra text to add to the menu item.
	 * @param keyPrefix The prefix to place on the command keys.
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
		field.setActionCommand(keyPrefix + QUALIFIER_KEY);
		field.addActionListener(this);
		parent.add(field);
	}

	private void addCompareMenuItem(TKMenu menu, String title, String altTitle, String extra, String keyPrefix, String key) {
		String msg;

		if (extra != null) {
			msg = MessageFormat.format(title, extra);
		} else {
			msg = altTitle;
		}
		menu.add(new TKMenuItem(msg, keyPrefix + key));
	}

	/**
	 * Call to handle a change in one of the string comparison controls.
	 * 
	 * @param compare The current string compare object.
	 * @param command The command without the prefix.
	 * @param event In the case of a change to the qualifier field, the {@link ActionEvent}
	 *            associated with the notification.
	 */
	protected void handleStringCompareChange(CMStringCriteria compare, String command, ActionEvent event) {
		if (QUALIFIER_KEY.equals(command)) {
			compare.setQualifier(((TKTextField) event.getSource()).getText());
		} else {
			compare.setType(CMStringCompareType.get(command));
		}
	}

	/**
	 * Adds the popup that allows an integer comparison to be changed.
	 * 
	 * @param parent The panel to add the popup and text field to.
	 * @param compare The current integer compare object.
	 * @param extra The extra text to add to the menu item.
	 * @param keyPrefix The prefix to place on the command keys.
	 * @param maxDigits The maximum number of digits to allow.
	 */
	protected void addNumericComparePopups(TKPanel parent, CMNumericCriteria compare, String extra, String keyPrefix, int maxDigits) {
		String[] keys = { CMNumericCriteria.IS, CMNumericCriteria.AT_LEAST, CMNumericCriteria.NO_MORE_THAN };
		String[] titles = { Msgs.IS, Msgs.AT_LEAST, Msgs.AT_MOST };
		String[] altTitles = { Msgs.ALT_IS, Msgs.ALT_AT_LEAST, Msgs.ALT_AT_MOST };
		TKMenu menu = new TKMenu();
		int selection = 0;
		TKPopupMenu popup;
		TKTextField field;

		for (int i = 0; i < keys.length; i++) {
			addCompareMenuItem(menu, titles[i], altTitles[i], extra, keyPrefix, keys[i]);
			if (compare.getType() == keys[i]) {
				selection = i;
			}
		}
		popup = new TKPopupMenu(menu, this, false, selection);
		popup.setOnlySize(popup.getPreferredSize());
		parent.add(popup);

		field = new TKTextField(TKTextUtility.makeFiller(maxDigits, 'M'));
		field.setOnlySize(field.getPreferredSize());
		field.setText(compare.getQualifierAsString(true));
		field.setKeyEventFilter(new TKNumberFilter(compare instanceof CMDoubleCriteria, false, maxDigits));
		field.setActionCommand(keyPrefix + QUALIFIER_KEY);
		field.addActionListener(this);
		parent.add(field);
	}

	/**
	 * Call to handle a change in one of the numeric comparison controls.
	 * 
	 * @param compare The current numeric compare object.
	 * @param command The command without the prefix.
	 * @param event In the case of a change to the qualifier field, the {@link ActionEvent}
	 *            associated with the notification.
	 */
	protected void handleNumericCompareChange(CMNumericCriteria compare, String command, ActionEvent event) {
		if (CMNumericCriteria.IS.equals(command)) {
			compare.setType(CMNumericCriteria.IS);
		} else if (CMNumericCriteria.AT_LEAST.equals(command)) {
			compare.setType(CMNumericCriteria.AT_LEAST);
		} else if (CMNumericCriteria.NO_MORE_THAN.equals(command)) {
			compare.setType(CMNumericCriteria.NO_MORE_THAN);
		} else if (QUALIFIER_KEY.equals(command)) {
			String text = ((TKTextField) event.getSource()).getText();

			if (compare instanceof CMIntegerCriteria) {
				((CMIntegerCriteria) compare).setQualifier(TKNumberUtils.getInteger(text, 0));
			} else if (compare instanceof CMDoubleCriteria) {
				((CMDoubleCriteria) compare).setQualifier(TKNumberUtils.getDouble(text, 0.0));
			}
		}
	}

	/** @return The depth of this prerequisite. */
	public int getDepth() {
		return mDepth;
	}

	/** @return The underlying prerequisite. */
	public CMPrereq getPrereq() {
		return mPrereq;
	}

	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (REMOVE.equals(command)) {
			TKPanel parent = (TKPanel) getParent();
			int index = parent.getIndexOf(this);
			int count = countSelfAndDescendents(mPrereq);

			for (int i = 0; i < count; i++) {
				parent.remove(index);
			}
			mPrereq.removeFromParent();
			parent.revalidate();
		}
	}

	private int countSelfAndDescendents(CMPrereq prereq) {
		int count = 1;

		if (prereq instanceof CMPrereqList) {
			for (CMPrereq one : ((CMPrereqList) prereq).getChildren()) {
				count += countSelfAndDescendents(one);
			}
		}
		return count;
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
		if (HAS.equals(command)) {
			((CMHasPrereq) mPrereq).has(true);
		} else if (DOES_NOT_HAVE.equals(command)) {
			((CMHasPrereq) mPrereq).has(false);
		} else if (CHANGE_BASE_TYPE.equals(command)) {
			String current = mPrereq.getXMLTag();
			String value = (String) item.getUserObject();

			if (!current.equals(value)) {
				forceFocusToAccept();
				replace(transform(value));
			}
			CSListPrereq.setLastItemType(value);
		} else {
			return false;
		}
		return true;
	}

	private void replace(CMPrereq prereq) {
		TKPanel parent = (TKPanel) getParent();
		int pIndex = parent.getIndexOf(this);
		CMPrereqList list = mPrereq.getParent();
		int lIndex = list.getIndexOf(mPrereq);

		list.add(lIndex, prereq);
		list.remove(mPrereq);

		parent.add(create(mRow, prereq, mDepth), pIndex);
		removeFromParent();
		parent.revalidate();
	}

	private CMPrereq transform(String to) {
		CMPrereqList parent = mPrereq.getParent();
		CMHasPrereq hasPrereq;

		if (CMAdvantagePrereq.TAG_ROOT.equals(to)) {
			hasPrereq = new CMAdvantagePrereq(parent);
		} else if (CMAttributePrereq.TAG_ROOT.equals(to)) {
			hasPrereq = new CMAttributePrereq(parent);
		} else if (CMContainedWeightPrereq.TAG_ROOT.equals(to)) {
			hasPrereq = new CMContainedWeightPrereq(parent);
		} else if (CMSkillPrereq.TAG_ROOT.equals(to)) {
			hasPrereq = new CMSkillPrereq(parent);
		} else if (CMSpellPrereq.TAG_ROOT.equals(to)) {
			hasPrereq = new CMSpellPrereq(parent);
		} else {
			return null;
		}
		hasPrereq.has(((CMHasPrereq) mPrereq).has());
		return hasPrereq;
	}

	private class AndOrLabel extends TKLabel {
		/** Create a new and/or label. */
		AndOrLabel() {
			super(Msgs.AND, TKFont.CONTROL_FONT_KEY, TKAlignment.RIGHT);
			setOnlySize(getPreferredSize());
		}

		@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
			CMPrereqList parent = mPrereq.getParent();

			if (parent != null && parent.getChildren().indexOf(mPrereq) != 0) {
				setText(parent.requiresAll() ? Msgs.AND : Msgs.OR);
			} else {
				setText(""); //$NON-NLS-1$
			}
			super.paintPanel(g2d, clips);
		}
	}
}
