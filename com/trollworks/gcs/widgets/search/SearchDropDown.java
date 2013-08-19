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

package com.trollworks.gcs.widgets.search;

import com.trollworks.gcs.widgets.UIUtilities;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Point;
import java.awt.event.KeyEvent;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;

import javax.swing.DefaultListModel;
import javax.swing.JList;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTextField;
import javax.swing.ListCellRenderer;

/** The drop-down panel used by {@link Search}. */
class SearchDropDown extends JPanel implements MouseListener {
	private JList				mList;
	private JTextField			mFilterField;
	private SearchTarget		mTarget;
	private DefaultListModel	mModel;

	/**
	 * Creates a new drop-down panel for use with {@link Search}.
	 * 
	 * @param renderer The item renderer to use in the list.
	 * @param filterField The text field the drop-down will appear to be attached to.
	 * @param target The search target.
	 */
	SearchDropDown(ListCellRenderer renderer, JTextField filterField, SearchTarget target) {
		super(new BorderLayout());
		setOpaque(true);
		setBackground(Color.red);
		mFilterField = filterField;
		mTarget = target;
		mModel = new DefaultListModel();
		mList = new JList(mModel);
		mList.setFocusable(false);
		mList.addMouseListener(this);
		mList.setCellRenderer(renderer);
		add(new JScrollPane(mList), BorderLayout.CENTER);
	}

	/** @return The currrently selected values. */
	public Object[] getSelectedValues() {
		return mList.getSelectedValues();
	}

	/**
	 * Adjust the list to the specified collection of hits.
	 * 
	 * @param hits The current collection of hits.
	 */
	void adjustToHits(Object[] hits) {
		mModel.removeAllElements();
		for (Object one : hits) {
			mModel.addElement(one);
		}
		Point where = new Point(0, mFilterField.getHeight());
		int count = mModel.getSize();
		int height = count > 0 ? mList.getCellBounds(0, 0).height * 5 + 1 : 0;
		UIUtilities.convertPoint(where, mFilterField, getParent());
		UIUtilities.revalidateImmediately(mFilterField);
		setBounds(where.x, where.y, mFilterField.getWidth(), height);
		revalidate();
	}

	public void mouseClicked(MouseEvent event) {
		if (event.getClickCount() > 1) {
			mTarget.searchSelect(getSelectedValues());
		}
	}

	public void mouseEntered(MouseEvent event) {
		// Not used.
	}

	public void mouseExited(MouseEvent event) {
		// Not used.
	}

	public void mousePressed(MouseEvent event) {
		// Not used.
	}

	public void mouseReleased(MouseEvent event) {
		// Not used.
	}

	void handleKeyPressed(KeyEvent event) {
		mList.dispatchEvent(event);
	}
}
