/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.page.Page;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.Wrapper;

import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;

/** Assembles pages in a sheet. */
public class PageAssembler {
    private static final int            GAP = 2;
    private              CharacterSheet mSheet;
    private              Wrapper        mContent;
    private              int            mRemaining;
    private              int            mContentHeight;
    private              int            mContentWidth;

    /**
     * Create a new page assembler.
     *
     * @param sheet The sheet to assemble pages within.
     */
    PageAssembler(CharacterSheet sheet) {
        mSheet = sheet;
        Scale.setOverride(mSheet.getScale());
        addPageInternal();
    }

    /** @return The content width. */
    public int getContentWidth() {
        return mContentWidth;
    }

    private void addPageInternal() {
        Page page = new Page(mSheet);
        mSheet.add(page);
        if (mContentHeight < 1) {
            Insets    insets = page.getInsets();
            Dimension size   = page.getSize();
            mContentWidth = size.width - (insets.left + insets.right);
            mContentHeight = size.height - (insets.top + insets.bottom);
        }
        if (mContent != null) {
            ((PrecisionLayout) mContent.getLayout()).getLayoutData(mContent.getComponent(mContent.getComponentCount() - 1)).setFillVerticalAlignment().setGrabVerticalSpace(true);
        }
        mContent = new Wrapper(new PrecisionLayout().setFillAlignment().setMargins(0, 0, 0, 0).setSpacing(GAP, GAP));
        mRemaining = mContentHeight;
        page.add(mContent);
    }

    /**
     * Add a panel to the content of the page.
     *
     * @param panel     The panel to add.
     * @param leftInfo  Outline info for the left outline.
     * @param rightInfo Outline info for the right outline.
     * @return {@code true} if the panel was too big to fit on a single page.
     */
    public boolean addToContent(Container panel, OutlineInfo leftInfo, OutlineInfo rightInfo) {
        boolean isOutline = panel instanceof SingleOutlinePanel || panel instanceof DoubleOutlinePanel;
        int     minLeft   = 0;
        int     minRight  = 0;
        int     height;

        if (mContent.getComponentCount() > 0) {
            mRemaining -= Scale.get(mContent).scale(GAP);
        }

        if (isOutline) {
            minLeft = leftInfo.getMinimumHeight();
            if (panel instanceof SingleOutlinePanel) {
                height = minLeft;
            } else {
                minRight = rightInfo.getMinimumHeight();
                height = Math.max(minLeft, minRight);
            }
        } else {
            height = panel.getPreferredSize().height;
        }
        if (mRemaining < height) {
            addPageInternal();
        }
        mContent.add(panel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        if (isOutline) {
            boolean hasMore;
            if (panel instanceof SingleOutlinePanel) {
                int startIndex = leftInfo.getRowIndex() + 1;
                int amt        = leftInfo.determineHeightForOutline(mRemaining);
                ((SingleOutlinePanel) panel).setOutlineRowRange(startIndex, leftInfo.getRowIndex());
                if (amt < minLeft) {
                    amt = minLeft;
                }
                mRemaining -= amt;
                hasMore = leftInfo.hasMore();
            } else {
                DoubleOutlinePanel panel2      = (DoubleOutlinePanel) panel;
                int                leftStart   = leftInfo.getRowIndex() + 1;
                int                leftHeight  = leftInfo.determineHeightForOutline(mRemaining);
                int                rightStart  = rightInfo.getRowIndex() + 1;
                int                rightHeight = rightInfo.determineHeightForOutline(mRemaining);
                panel2.setOutlineRowRange(false, leftStart, leftInfo.getRowIndex());
                panel2.setOutlineRowRange(true, rightStart, rightInfo.getRowIndex());
                if (leftHeight < minLeft) {
                    leftHeight = minLeft;
                }
                if (rightHeight < minRight) {
                    rightHeight = minRight;
                }
                mRemaining -= Math.max(leftHeight, rightHeight);
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

    @SuppressWarnings("static-method")
    public void finish() {
        if (mContent != null) {
            ((PrecisionLayout) mContent.getLayout()).getLayoutData(mContent.getComponent(mContent.getComponentCount() - 1)).setFillVerticalAlignment().setGrabVerticalSpace(true);
        }
        Scale.setOverride(null);
    }
}
