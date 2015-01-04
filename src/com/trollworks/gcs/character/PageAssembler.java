/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.layout.RowDistribution;
import com.trollworks.toolkit.ui.widget.Wrapper;

import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;

/** Assembles pages in a sheet. */
public class PageAssembler {
	private CharacterSheet	mSheet;
	private Page			mPage;
	private Wrapper			mContent;
	private int				mRemaining;
	private int				mContentHeight;
	private int				mContentWidth;

	/**
	 * Create a new page assembler.
	 * 
	 * @param sheet The sheet to assemble pages within.
	 */
	PageAssembler(CharacterSheet sheet) {
		mSheet = sheet;
		addPageInternal();
	}

	/** Adds the notes field to the sheet. */
	public void addNotes() {
		NotesPanel notes = new NotesPanel(mSheet.getCharacter().getDescription().getNotes(), false);
		Insets insets = notes.getInsets();
		int width = mContentWidth - (insets.left + insets.right);
		boolean addPage = mRemaining < notes.getMinimumSize().height;
		notes.addActionListener(mSheet);
		notes.setWrapWidth(width);
		while (true) {
			if (addPage) {
				addPageInternal();
			}
			String text = notes.setMaxHeight(mRemaining);
			addPage = !addToContent(notes, null, null);
			if (text == null) {
				break;
			}
			notes = new NotesPanel(text, true);
			notes.addActionListener(mSheet);
			notes.setWrapWidth(width);
		}
	}

	/** @return The content width. */
	public int getContentWidth() {
		return mContentWidth;
	}

	private void addPageInternal() {
		mPage = new Page(mSheet);
		mSheet.add(mPage);
		if (mContentHeight < 1) {
			Insets insets = mPage.getInsets();
			Dimension size = mPage.getSize();
			mContentWidth = size.width - (insets.left + insets.right);
			mContentHeight = size.height - (insets.top + insets.bottom);
		}
		mContent = new Wrapper(new ColumnLayout(1, 2, 2, RowDistribution.GIVE_EXCESS_TO_LAST));
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
	public boolean addToContent(Container panel, OutlineInfo leftInfo, OutlineInfo rightInfo) {
		boolean isOutline = panel instanceof SingleOutlinePanel || panel instanceof DoubleOutlinePanel;
		int height = 0;
		int minLeft = 0;
		int minRight = 0;

		if (isOutline) {
			minLeft = leftInfo.getMinimumHeight();
			if (panel instanceof SingleOutlinePanel) {
				height += minLeft;
			} else {
				minRight = rightInfo.getMinimumHeight();
				height += minLeft < minRight ? minRight : minLeft;
			}
		} else {
			height += panel.getPreferredSize().height;
		}
		if (mRemaining < height) {
			addPageInternal();
		}
		mContent.add(panel);

		if (isOutline) {
			int savedRemaining = mRemaining;
			boolean hasMore;

			if (panel instanceof SingleOutlinePanel) {
				int startIndex = leftInfo.getRowIndex() + 1;
				int amt = leftInfo.determineHeightForOutline(mRemaining);
				((SingleOutlinePanel) panel).setOutlineRowRange(startIndex, leftInfo.getRowIndex());
				if (amt < minLeft) {
					amt = minLeft;
				}
				mRemaining = savedRemaining - amt;
				hasMore = leftInfo.hasMore();
			} else {
				DoubleOutlinePanel panel2 = (DoubleOutlinePanel) panel;
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

		if (mRemaining >= height) {
			mRemaining -= height;
			return false;
		}
		addPageInternal();
		return true;
	}
}
