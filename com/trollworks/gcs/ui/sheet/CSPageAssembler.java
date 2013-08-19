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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKRowDistribution;

import java.awt.Dimension;
import java.awt.Insets;
import java.awt.Rectangle;

/** Assembles pages in a sheet. */
public class CSPageAssembler {
	private CSSheet	mSheet;
	private CSPage	mPage;
	private TKPanel	mContent;
	private int		mRemaining;
	private int		mContentHeight;
	private int		mContentWidth;

	/**
	 * Create a new page assembler.
	 * 
	 * @param sheet The sheet to assemble pages within.
	 */
	CSPageAssembler(CSSheet sheet) {
		mSheet = sheet;
		addPageInternal();
	}

	/** Adds the notes field to the sheet. */
	public void addNotes() {
		CSNotesPanel notes = new CSNotesPanel(mSheet.getCharacter().getNotes(), false);
		Insets insets = notes.getInsets();
		int width = mContentWidth - (insets.left + insets.right);

		if (mRemaining < notes.getMinimumSize().height) {
			addPageInternal();
		}
		notes.addActionListener(mSheet);
		notes.setWrapWidth(width);
		while (true) {
			String text = notes.setMaxHeight(mRemaining);

			addToContent(notes, null, null);
			if (text == null) {
				break;
			}
			notes = new CSNotesPanel(text, true);
			notes.addActionListener(mSheet);
			notes.setWrapWidth(width);
			addPageInternal();
		}
	}

	private void addPageInternal() {
		mPage = new CSPage(mSheet);

		mSheet.add(mPage);
		mPage.setSize(mPage.getPreferredSize());
		if (mContentHeight < 1) {
			Rectangle bounds = mPage.getLocalInsetBounds();

			mContentHeight = bounds.height;
			mContentWidth = bounds.width;
		}
		mContent = new TKPanel(new TKColumnLayout(1, 2, 2, TKRowDistribution.GIVE_EXCESS_TO_LAST));
		mContent.setAlignmentY(-1f);
		mRemaining = mContentHeight;
		mPage.add(mContent);
	}

	/**
	 * Add a panel to the content of the page.
	 * 
	 * @param panel The panel to add.
	 * @param leftInfo Outline info for the left outline.
	 * @param rightInfo Outline info for the right outline.
	 * @return <code>true</code> if the panel was too big to fit on a single page.
	 */
	public boolean addToContent(TKPanel panel, CSOutlineInfo leftInfo, CSOutlineInfo rightInfo) {
		boolean isOutline = panel instanceof CSSingleOutlinePanel || panel instanceof CSDoubleOutlinePanel;
		int height = 0;
		int minLeft = 0;
		int minRight = 0;

		if (isOutline) {
			minLeft = leftInfo.getMinimumHeight();
			if (panel instanceof CSSingleOutlinePanel) {
				height += minLeft;
			} else {
				minRight = rightInfo.getMinimumHeight();
				height += minLeft < minRight ? minRight : minLeft;
			}
		} else {
			height += panel.getPreferredSize().height;
		}
		if (mRemaining <= height) {
			addPageInternal();
		}
		panel.setMinimumSize(new Dimension(panel.getMinimumSize().width, height));
		mContent.add(panel);

		if (isOutline) {
			int savedRemaining = --mRemaining;
			boolean hasMore;

			if (panel instanceof CSSingleOutlinePanel) {
				int startIndex = leftInfo.getRowIndex() + 1;
				int amt = leftInfo.determineHeightForOutline(mRemaining);

				((CSSingleOutlinePanel) panel).setOutlineRowRange(startIndex, leftInfo.getRowIndex());
				if (amt < minLeft) {
					amt = minLeft;
				}
				mRemaining = savedRemaining - amt;
				hasMore = leftInfo.hasMore();
			} else {
				CSDoubleOutlinePanel panel2 = (CSDoubleOutlinePanel) panel;
				int leftStart = leftInfo.getRowIndex() + 1;
				int leftHeight = leftInfo.determineHeightForOutline(mRemaining);
				int rightStart = rightInfo.getRowIndex() + 1;
				int rightHeight = rightInfo.determineHeightForOutline(mRemaining);

				panel2.setOutlineRowRange(false, leftStart, leftInfo.getRowIndex());
				panel2.setOutlineRowRange(true, rightStart, rightInfo.getRowIndex());
				if (leftHeight < minLeft) {
					leftHeight = minLeft;
				}
				if (rightHeight < minRight) {
					rightHeight = minRight;
				}
				if (leftHeight < rightHeight) {
					mRemaining = savedRemaining - rightHeight;
				} else {
					mRemaining = savedRemaining - leftHeight;
				}
				hasMore = leftInfo.hasMore() || rightInfo.hasMore();
			}
			if (hasMore) {
				addPageInternal();
				return true;
			}
			return false;
		}

		if (mRemaining > height) {
			mRemaining -= height;
			return false;
		}
		addPageInternal();
		return true;
	}
}
