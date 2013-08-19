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

package com.trollworks.toolkit.widget;

import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.FlowLayout;
import java.awt.Font;
import java.awt.GraphicsEnvironment;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

/** A standard font selection panel. */
public class TKFontPanel extends TKPanel implements ActionListener {
	private static final String[]	STD_FONT_SIZES	= { "6", "7", "8", "9", "10", "11", "12", "14", "16", "18", "20", "24", "30", "36" };	//$NON-NLS-1$ //$NON-NLS-2$$ //$NON-NLS-3$ //$NON-NLS-4$ //$NON-NLS-5$ //$NON-NLS-6$ //$NON-NLS-7$ //$NON-NLS-8$ //$NON-NLS-9$ //$NON-NLS-10$ //$NON-NLS-11$ //$NON-NLS-12$ //$NON-NLS-13$ //$NON-NLS-14$
	private TKPopupMenu				mFontSizeMenu;
	private TKPopupMenu				mFontNameMenu;
	private TKPopupMenu				mFontStyleMenu;
	private boolean					mNoNotify;

	/** Creates a new font panel. */
	public TKFontPanel() {
		this(TKFont.lookup(TKFont.TEXT_FONT_KEY));
	}

	/**
	 * Creates a new font panel.
	 * 
	 * @param font The font to start with.
	 */
	public TKFontPanel(Font font) {
		super(new FlowLayout(FlowLayout.LEFT, 5, 2));

		String[] fontList = GraphicsEnvironment.getLocalGraphicsEnvironment().getAvailableFontFamilyNames();
		TKMenu menu = new TKMenu();
		int i;

		for (i = 0; i < fontList.length; i++) {
			menu.add(new TKMenuItem(fontList[i]));
		}

		mFontNameMenu = new TKPopupMenu(menu);
		mFontNameMenu.setToolTipText(Msgs.NAME_TOOLTIP);
		mFontNameMenu.addActionListener(this);
		add(mFontNameMenu);

		menu = new TKMenu();
		for (i = 0; i < STD_FONT_SIZES.length; i++) {
			menu.add(new TKMenuItem(STD_FONT_SIZES[i]));
		}

		mFontSizeMenu = new TKPopupMenu(menu, true);
		mFontSizeMenu.setToolTipText(Msgs.SIZE_TOOLTIP);
		mFontSizeMenu.addActionListener(this);
		add(mFontSizeMenu);

		menu = new TKMenu();
		menu.add(new TKMenuItem(Msgs.PLAIN));
		menu.add(new TKMenuItem(Msgs.BOLD));
		menu.add(new TKMenuItem(Msgs.ITALIC));
		menu.add(new TKMenuItem(Msgs.BOLD_ITALIC));
		mFontStyleMenu = new TKPopupMenu(menu);
		mFontStyleMenu.setToolTipText(Msgs.STYLE_TOOLTIP);
		mFontStyleMenu.addActionListener(this);
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
		int size;

		try {
			size = Integer.parseInt(mFontSizeMenu.getSelectedItem().getTitle());
		} catch (NumberFormatException nfe) {
			size = 12;
		}

		if (size > 72) {
			size = 72;
		} else if (size < 1) {
			size = 1;
		}

		return new Font(mFontNameMenu.getSelectedItem().getTitle(), mFontStyleMenu.getSelectedItemIndex(), size);
	}

	/** @param font The new font. */
	public void setCurrentFont(Font font) {
		mNoNotify = true;
		mFontNameMenu.setSelectedItem(font.getName());
		if (mFontNameMenu.getSelectedItem() == null) {
			mFontNameMenu.setSelectedItem(0);
		}
		mFontSizeMenu.setSelectedItem("" + font.getSize()); //$NON-NLS-1$
		if (mFontSizeMenu.getSelectedItem() == null) {
			mFontSizeMenu.setSelectedItem(0);
		}
		mFontStyleMenu.setSelectedItem(font.getStyle());
		if (mFontStyleMenu.getSelectedItem() == null) {
			mFontStyleMenu.setSelectedItem(0);
		}
		mNoNotify = false;
		notifyActionListeners();
	}
}
