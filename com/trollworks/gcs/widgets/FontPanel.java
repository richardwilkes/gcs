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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.awt.FlowLayout;
import java.awt.Font;
import java.awt.GraphicsEnvironment;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

import javax.swing.JComboBox;

/** A standard font selection panel. */
public class FontPanel extends ActionPanel implements ActionListener {
	private static String			MSG_PLAIN;
	private static String			MSG_BOLD;
	private static String			MSG_ITALIC;
	private static String			MSG_BOLD_ITALIC;
	private static String			MSG_NAME_TOOLTIP;
	private static String			MSG_SIZE_TOOLTIP;
	private static String			MSG_STYLE_TOOLTIP;

	static {
		LocalizedMessages.initialize(FontPanel.class);
	}

	private static final String[]	STD_STYLES	= { MSG_PLAIN, MSG_BOLD, MSG_ITALIC, MSG_BOLD_ITALIC };
	private JComboBox				mFontSizeMenu;
	private JComboBox				mFontNameMenu;
	private JComboBox				mFontStyleMenu;
	private boolean					mNoNotify;

	/**
	 * Creates a new font panel.
	 * 
	 * @param font The font to start with.
	 */
	public FontPanel(Font font) {
		super(new FlowLayout(FlowLayout.LEFT, 5, 0));
		setOpaque(false);

		mFontNameMenu = new JComboBox(GraphicsEnvironment.getLocalGraphicsEnvironment().getAvailableFontFamilyNames());
		mFontNameMenu.setOpaque(false);
		mFontNameMenu.setToolTipText(MSG_NAME_TOOLTIP);
		mFontNameMenu.setMaximumRowCount(25);
		mFontNameMenu.addActionListener(this);
		UIUtilities.setOnlySize(mFontNameMenu, mFontNameMenu.getPreferredSize());
		add(mFontNameMenu);

		Integer[] sizes = new Integer[10];
		for (int i = 0; i < 7; i++) {
			sizes[i] = new Integer(6 + i);
		}
		sizes[7] = new Integer(14);
		sizes[8] = new Integer(16);
		sizes[9] = new Integer(18);
		mFontSizeMenu = new JComboBox(sizes);
		mFontSizeMenu.setOpaque(false);
		mFontSizeMenu.setToolTipText(MSG_SIZE_TOOLTIP);
		mFontSizeMenu.setMaximumRowCount(sizes.length);
		mFontSizeMenu.addActionListener(this);
		UIUtilities.setOnlySize(mFontSizeMenu, mFontSizeMenu.getPreferredSize());
		add(mFontSizeMenu);

		mFontStyleMenu = new JComboBox(STD_STYLES);
		mFontStyleMenu.setOpaque(false);
		mFontStyleMenu.setToolTipText(MSG_STYLE_TOOLTIP);
		mFontStyleMenu.addActionListener(this);
		UIUtilities.setOnlySize(mFontStyleMenu, mFontStyleMenu.getPreferredSize());
		add(mFontStyleMenu);

		setCurrentFont(font);
	}

	public void actionPerformed(ActionEvent event) {
		notifyActionListeners();
	}

	@Override public void notifyActionListeners(ActionEvent event) {
		if (!mNoNotify) {
			super.notifyActionListeners(event);
		}
	}

	/** @return The font this panel has been set to. */
	public Font getCurrentFont() {
		return new Font((String) mFontNameMenu.getSelectedItem(), mFontStyleMenu.getSelectedIndex(), ((Integer) mFontSizeMenu.getSelectedItem()).intValue());
	}

	/** @param font The new font. */
	public void setCurrentFont(Font font) {
		mNoNotify = true;
		mFontNameMenu.setSelectedItem(font.getName());
		if (mFontNameMenu.getSelectedItem() == null) {
			mFontNameMenu.setSelectedIndex(0);
		}
		mFontSizeMenu.setSelectedItem(new Integer(font.getSize()));
		if (mFontSizeMenu.getSelectedItem() == null) {
			mFontSizeMenu.setSelectedIndex(3);
		}
		mFontStyleMenu.setSelectedIndex(font.getStyle());
		if (mFontStyleMenu.getSelectedItem() == null) {
			mFontStyleMenu.setSelectedIndex(0);
		}
		mNoNotify = false;
		notifyActionListeners();
	}
}
