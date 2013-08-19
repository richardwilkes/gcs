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

package com.trollworks.toolkit.widget.search;

import com.trollworks.toolkit.widget.TKItemList;
import com.trollworks.toolkit.widget.TKItemRenderer;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKReshapeListener;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;

import java.awt.Point;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.util.ArrayList;
import java.util.Collection;

/**
 * The drop-down panel used by {@link TKSearch}. This class is exposed to the outside world as a
 * side-effect. Do not use it directly!
 */
public class TKSearchDropDown extends TKPanel implements ActionListener, TKReshapeListener {
	private TKItemList<Object>	mList;
	private TKScrollPanel		mScrollPane;
	private TKTextField			mFilterField;
	private TKSearchTarget		mTarget;

	/**
	 * Creates a new drop-down panel for use with {@link TKSearch}.
	 * 
	 * @param renderer The item renderer to use in the list.
	 * @param filterField The text field the drop-down will appear to be attached to.
	 * @param target The search target.
	 */
	public TKSearchDropDown(TKItemRenderer renderer, TKTextField filterField, TKSearchTarget target) {
		super(new TKCompassLayout());
		setOpaque(true);
		mFilterField = filterField;
		mTarget = target;
		mList = new TKItemList<Object>();
		mList.setFocusable(false);
		mList.addActionListener(this);
		mList.setItemRenderer(renderer);
		mScrollPane = new TKScrollPanel(TKScrollPanel.VERTICAL, mList);
		mScrollPane.getContentBorderView().setBorder(new TKLineBorder(TKLineBorder.LEFT_EDGE | TKLineBorder.BOTTOM_EDGE));
		add(mScrollPane, TKCompassPosition.CENTER);
	}

	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (TKItemList.CMD_OPEN_SELECTION.equals(command)) {
			openSelection();
		}
	}

	private void adjustToField() {
		Point where = new Point(0, mFilterField.getHeight());
		int count = mList.getItemCount();
		int height = count > 0 ? mList.getItemBounds(0).height * 5 + 1 : 0;

		convertPoint(where, mFilterField, getParent());
		setBounds(where.x, where.y, mFilterField.getWidth(), height);
	}

	public void panelReshaped(TKPanel component, boolean moved, boolean resized) {
		adjustToField();
	}

	/**
	 * Adjust the list to the specified collection of hits.
	 * 
	 * @param hits The current collection of hits.
	 */
	void adjustToHits(Collection<Object> hits) {
		mList.replaceItems(hits);
		revalidate();
		adjustToField();
	}

	/**
	 * Process a key event.
	 * 
	 * @param event The key event to process.
	 * @return <code>true</code> if it is handled.
	 */
	boolean doKey(KeyEvent event) {
		int count = mList.getItemCount();

		if (count > 0) {
			switch (event.getID()) {
				case KeyEvent.KEY_PRESSED:
					mList.processKeyEvent(event);
					break;
				case KeyEvent.KEY_TYPED:
					char ch = event.getKeyChar();

					if (mList.getSelectionCount() > 0 && (ch == '\r' || ch == '\n')) {
						openSelection();
						return true;
					}
					break;
			}
		}
		return false;
	}

	private void openSelection() {
		int[] selection = mList.getSelectedIndexes();
		ArrayList<Object> list = new ArrayList<Object>(selection.length);

		for (int element : selection) {
			list.add(mList.getItem(element));
		}
		mTarget.searchSelect(list);
	}
}
