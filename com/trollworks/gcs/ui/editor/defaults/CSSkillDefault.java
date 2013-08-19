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

package com.trollworks.gcs.ui.editor.defaults;

import com.trollworks.gcs.model.skill.CMSkillDefault;
import com.trollworks.gcs.model.skill.CMSkillDefaultType;
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

import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;

/** A skill default editor panel. */
public class CSSkillDefault extends TKPanel implements ActionListener {
	private static final String			ADD				= "Add";					//$NON-NLS-1$
	private static final String			REMOVE			= "Remove";				//$NON-NLS-1$
	private static CMSkillDefaultType	LAST_ITEM_TYPE	= CMSkillDefaultType.DX;
	private CMSkillDefault				mDefault;
	/** The center panel. */
	protected TKPanel					mCenter;
	/** The right-side panel. */
	protected TKPanel					mRight;

	/** @param type The last item type created or switched to. */
	public static void setLastItemType(CMSkillDefaultType type) {
		LAST_ITEM_TYPE = type;
	}

	/**
	 * Creates a new skill default editor panel.
	 * 
	 * @param skillDefault The skill default to edit.
	 */
	public CSSkillDefault(CMSkillDefault skillDefault) {
		super(new TKColumnLayout(2));
		setBorder(new TKEmptyBorder(0, TKColumnLayout.DEFAULT_H_GAP_SIZE, 0, 0));
		mDefault = skillDefault;
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

		addButton(CSImage.getAddIcon(), ADD, Msgs.ADD_DEFAULT);
		if (mDefault != null) {
			addButton(CSImage.getRemoveIcon(), REMOVE, Msgs.REMOVE_DEFAULT);
		}

		mRight.setLayout(new TKColumnLayout(Math.max(mRight.getComponentCount(), 1)));
		mCenter.invalidate();
		mRight.invalidate();
		revalidate();
	}

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

	/** Editable fields go here. */
	protected void rebuildSelf() {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4));
		TKMenu menu = new TKMenu();
		CMSkillDefaultType current = mDefault.getType();
		TKPopupMenu popup;
		TKTextField field;

		wrapper.add(new OrLabel());

		for (CMSkillDefaultType type : CMSkillDefaultType.values()) {
			TKMenuItem item = new TKMenuItem(type.toString());

			item.setUserObject(type);
			menu.add(item);
		}
		popup = new TKPopupMenu(menu);
		popup.setSelectedUserObject(mDefault.getType());
		popup.setActionCommand(CMSkillDefault.TAG_TYPE);
		popup.addActionListener(this);
		popup.setOnlySize(popup.getPreferredSize());
		wrapper.add(popup);

		if (CMSkillDefaultType.Skill == current) {
			wrapper.add(new TKPanel());
			mCenter.add(wrapper);

			wrapper = new TKPanel(new TKColumnLayout(3));
			wrapper.setBorder(new TKEmptyBorder(0, 40, 0, 0));
			field = new TKTextField(mDefault.getName(), 50);
			field.setActionCommand(CMSkillDefault.TAG_NAME);
			field.addActionListener(this);
			wrapper.add(field);

			field = new TKTextField(mDefault.getSpecialization(), 50);
			field.setImprint(Msgs.SPECIALIZATION_IMPRINT);
			field.setActionCommand(CMSkillDefault.TAG_SPECIALIZATION);
			field.addActionListener(this);
			wrapper.add(field);
		}

		field = new TKTextField("+" + TKTextUtility.makeFiller(2, 'M')); //$NON-NLS-1$
		field.setOnlySize(field.getPreferredSize());
		field.setText(TKNumberUtils.format(mDefault.getModifier(), true));
		field.setKeyEventFilter(new TKNumberFilter(false, true, 2));
		field.setActionCommand(CMSkillDefault.TAG_MODIFIER);
		field.addActionListener(this);
		wrapper.add(field);
		if (CMSkillDefaultType.Skill != current) {
			wrapper.add(new TKPanel());
		}
		mCenter.add(wrapper);
	}

	/** @return The underlying skill default. */
	public CMSkillDefault getSkillDefault() {
		return mDefault;
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		String command = event.getActionCommand();
		TKPanel parent = (TKPanel) getParent();

		if (ADD.equals(command)) {
			CMSkillDefault skillDefault = new CMSkillDefault(LAST_ITEM_TYPE, LAST_ITEM_TYPE == CMSkillDefaultType.Skill ? "" : null, null, 0); //$NON-NLS-1$

			parent.add(new CSSkillDefault(skillDefault));
			if (mDefault == null) {
				removeFromParent();
			}
			parent.revalidate();
			notifyActionListeners();
		} else if (REMOVE.equals(command)) {
			removeFromParent();
			if (parent.getComponentCount() == 0) {
				parent.add(new CSNoDefault());
			}
			parent.revalidate();
			notifyActionListeners();
		} else if (CMSkillDefault.TAG_NAME.equals(command)) {
			mDefault.setName(((TKTextField) src).getText());
			notifyActionListeners();
		} else if (CMSkillDefault.TAG_SPECIALIZATION.equals(command)) {
			mDefault.setSpecialization(((TKTextField) src).getText());
			notifyActionListeners();
		} else if (CMSkillDefault.TAG_MODIFIER.equals(command)) {
			mDefault.setModifier(TKNumberUtils.getInteger(((TKTextField) src).getText(), 0));
			notifyActionListeners();
		} else if (CMSkillDefault.TAG_TYPE.equals(command)) {
			CMSkillDefaultType current = mDefault.getType();
			CMSkillDefaultType value = (CMSkillDefaultType) ((TKPopupMenu) src).getSelectedItemUserObject();

			if (!current.equals(value)) {
				forceFocusToAccept();
				mDefault.setType(value);
				rebuild();
				notifyActionListeners();
			}
			setLastItemType(value);
		}
	}

	private class OrLabel extends TKLabel {
		/** Create a new or label. */
		OrLabel() {
			super(Msgs.OR, TKFont.CONTROL_FONT_KEY, TKAlignment.RIGHT);
			setOnlySize(getPreferredSize());
		}

		@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
			CSDefaults parent = (CSDefaults) CSSkillDefault.this.getParent();

			if (parent != null && parent.getComponent(0) != CSSkillDefault.this) {
				setText(Msgs.OR);
			} else {
				setText(""); //$NON-NLS-1$
			}
			super.paintPanel(g2d, clips);
		}
	}
}
