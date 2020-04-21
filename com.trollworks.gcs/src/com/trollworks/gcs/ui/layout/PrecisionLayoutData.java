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

package com.trollworks.gcs.ui.layout;

import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Component;
import java.awt.Dimension;

/**
 * Data for components within a {@link PrecisionLayout}. Do not re-use {@link PrecisionLayoutData}
 * objects. Each component should have its own.
 */
public final class PrecisionLayoutData {
    public static final int                      DEFAULT     = -1;
    private             int                      mCacheMinWidth;
    private             int                      mCacheWidth;
    private             int                      mCacheHeight;
    private             PrecisionLayoutAlignment mVAlign     = PrecisionLayoutAlignment.MIDDLE;
    private             PrecisionLayoutAlignment mHAlign     = PrecisionLayoutAlignment.BEGINNING;
    private             int                      mMarginTop;
    private             int                      mMarginLeft;
    private             int                      mMarginBottom;
    private             int                      mMarginRight;
    private             int                      mWidthHint  = DEFAULT;
    private             int                      mHeightHint = DEFAULT;
    private             int                      mHSpan      = 1;
    private             int                      mVSpan      = 1;
    private             int                      mMinWidth   = DEFAULT;
    private             int                      mMinHeight  = DEFAULT;
    private             boolean                  mHGrab;
    private             boolean                  mVGrab;
    private             boolean                  mExclude;

    /**
     * Position the component at the left of the cell. This is the default.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setBeginningHorizontalAlign() {
        mHAlign = PrecisionLayoutAlignment.BEGINNING;
        return this;
    }

    /**
     * Position the component in the horizontal center of the cell.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setMiddleHorizontalAlignment() {
        mHAlign = PrecisionLayoutAlignment.MIDDLE;
        return this;
    }

    /**
     * Position the component at the right of the cell.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setEndHorizontalAlignment() {
        mHAlign = PrecisionLayoutAlignment.END;
        return this;
    }

    /**
     * Resize the component to fill the cell horizontally.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setFillHorizontalAlignment() {
        mHAlign = PrecisionLayoutAlignment.FILL;
        return this;
    }

    /** @return The horizontal positioning of components within the container. */
    public PrecisionLayoutAlignment getHorizontalAlignment() {
        return mHAlign;
    }

    /**
     * @param alignment Specifies how components will be positioned horizontally within a cell. The
     *                  default value is {@link PrecisionLayoutAlignment#BEGINNING}.
     * @return This layout data.
     */
    public PrecisionLayoutData setHorizontalAlignment(PrecisionLayoutAlignment alignment) {
        mHAlign = alignment;
        return this;
    }

    /**
     * Position the component at the top of the cell.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setBeginningVerticalAlignment() {
        mVAlign = PrecisionLayoutAlignment.BEGINNING;
        return this;
    }

    /**
     * Position the component in the vertical center of the cell. This is the default.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setMiddleVerticalAlignment() {
        mVAlign = PrecisionLayoutAlignment.MIDDLE;
        return this;
    }

    /**
     * Position the component at the bottom of the cell.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setEndVerticalAlignment() {
        mVAlign = PrecisionLayoutAlignment.END;
        return this;
    }

    /**
     * Resize the component to fill the cell vertically.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setFillVerticalAlignment() {
        mVAlign = PrecisionLayoutAlignment.FILL;
        return this;
    }

    /** @return The vertical positioning of components within the container. */
    public PrecisionLayoutAlignment getVerticalAlignment() {
        return mVAlign;
    }

    /**
     * @param alignment Specifies how components will be positioned vertically within a cell. The
     *                  default value is {@link PrecisionLayoutAlignment#MIDDLE}.
     * @return This layout data.
     */
    public PrecisionLayoutData setVerticalAlignment(PrecisionLayoutAlignment alignment) {
        mVAlign = alignment;
        return this;
    }

    /**
     * Position the component at the top left of the cell.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setBeginningAlignment() {
        mHAlign = PrecisionLayoutAlignment.BEGINNING;
        mVAlign = PrecisionLayoutAlignment.BEGINNING;
        return this;
    }

    /**
     * Position the component in the center of the cell.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setMiddleAlignment() {
        mHAlign = PrecisionLayoutAlignment.MIDDLE;
        mVAlign = PrecisionLayoutAlignment.MIDDLE;
        return this;
    }

    /**
     * Position the component at the bottom right of the cell.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setEndAlignment() {
        mHAlign = PrecisionLayoutAlignment.END;
        mVAlign = PrecisionLayoutAlignment.END;
        return this;
    }

    /**
     * Resize the component to fill the cell horizontally and vertically.
     *
     * @return This layout data.
     */
    public PrecisionLayoutData setFillAlignment() {
        mHAlign = PrecisionLayoutAlignment.FILL;
        mVAlign = PrecisionLayoutAlignment.FILL;
        return this;
    }

    /**
     * @param horizontal Specifies how components will be positioned horizontally within a cell. The
     *                   default value is {@link PrecisionLayoutAlignment#BEGINNING}.
     * @param vertical   Specifies how components will be positioned vertically within a cell. The
     *                   default value is {@link PrecisionLayoutAlignment#MIDDLE}.
     * @return This layout data.
     */
    public PrecisionLayoutData setAlignment(PrecisionLayoutAlignment horizontal, PrecisionLayoutAlignment vertical) {
        mHAlign = horizontal;
        mVAlign = vertical;
        return this;
    }

    /**
     * @return The number of pixels of indentation that will be placed along the top side of the
     *         cell.
     */
    public int getTopMargin() {
        return mMarginTop;
    }

    /**
     * @param top The number of pixels of indentation that will be placed along the top side of the
     *            cell. The default value is 0.
     * @return This layout data.
     */
    public PrecisionLayoutData setTopMargin(int top) {
        mMarginTop = top;
        return this;
    }

    /**
     * @return The number of pixels of indentation that will be placed along the left side of the
     *         cell.
     */
    public int getLeftMargin() {
        return mMarginLeft;
    }

    /**
     * @param left The number of pixels of indentation that will be placed along the left side of
     *             the cell. The default value is 0.
     * @return This layout data.
     */
    public PrecisionLayoutData setLeftMargin(int left) {
        mMarginLeft = left;
        return this;
    }

    /**
     * @return The number of pixels of indentation that will be placed along the bottom side of the
     *         cell.
     */
    public int getBottomMargin() {
        return mMarginBottom;
    }

    /**
     * @param bottom The number of pixels of indentation that will be placed along the bottom side
     *               of the cell. The default value is 0.
     * @return This layout data.
     */
    public PrecisionLayoutData setBottomMargin(int bottom) {
        mMarginBottom = bottom;
        return this;
    }

    /**
     * @return The number of pixels of indentation that will be placed along the right side of the
     *         cell.
     */
    public int getRightMargin() {
        return mMarginRight;
    }

    /**
     * @param right The number of pixels of indentation that will be placed along the right side of
     *              the cell. The default value is 0.
     * @return This layout data.
     */
    public PrecisionLayoutData setRightMargin(int right) {
        mMarginRight = right;
        return this;
    }

    /**
     * @param margins The number of pixels of margin that will be placed along each edge of the
     *                layout. The default value is 4.
     * @return This layout data.
     */
    public PrecisionLayoutData setMargins(int margins) {
        mMarginTop = margins;
        mMarginLeft = margins;
        mMarginBottom = margins;
        mMarginRight = margins;
        return this;
    }

    /**
     * @param top    The number of pixels of indentation that will be placed along the top side of
     *               the cell. The default value is 0.
     * @param left   The number of pixels of indentation that will be placed along the left side of
     *               the cell. The default value is 0.
     * @param bottom The number of pixels of indentation that will be placed along the bottom side
     *               of the cell. The default value is 0.
     * @param right  The number of pixels of indentation that will be placed along the right side of
     *               the cell. The default value is 0.
     * @return This layout data.
     */
    public PrecisionLayoutData setMargins(int top, int left, int bottom, int right) {
        mMarginTop = top;
        mMarginLeft = left;
        mMarginBottom = bottom;
        mMarginRight = right;
        return this;
    }

    /**
     * @return The preferred width in pixels. A value of {@link #DEFAULT} indicates the component
     *         should be asked for its preferred size.
     */
    public int getWidthHint() {
        return mWidthHint;
    }

    /**
     * @param width The preferred width in pixels. A value of {@link #DEFAULT} indicates the
     *              component should be asked for its preferred size. The default value is {@link
     *              #DEFAULT}.
     * @return This layout data.
     */
    public PrecisionLayoutData setWidthHint(int width) {
        mWidthHint = width;
        return this;
    }

    /**
     * @return The minimum width in pixels. This value applies only if {@link
     *         #shouldGrabHorizontalSpace()} is {@code true}. A value of {@link #DEFAULT} means that
     *         the minimum width will be determined by calling {@link Component#getMinimumSize()}.
     */
    public int getMinimumWidth() {
        return mMinWidth;
    }

    /**
     * @param width The minimum width in pixels. This value applies only if {@link
     *              #shouldGrabHorizontalSpace()} is {@code true}. A value of {@link #DEFAULT} means
     *              that the minimum width will be determined by calling {@link
     *              Component#getMinimumSize()}. The default value is {@link #DEFAULT}.
     * @return This layout data.
     */
    public PrecisionLayoutData setMinimumWidth(int width) {
        mMinWidth = width;
        return this;
    }

    /**
     * @return The preferred height in pixels. A value of {@link #DEFAULT} indicates the component
     *         should be asked for its preferred size.
     */
    public int getHeightHint() {
        return mHeightHint;
    }

    /**
     * @param height The preferred height in pixels. A value of {@link #DEFAULT} indicates the
     *               component should be asked for its preferred size. The default value is {@link
     *               #DEFAULT}.
     * @return This layout data.
     */
    public PrecisionLayoutData setHeightHint(int height) {
        mHeightHint = height;
        return this;
    }

    /**
     * @return The minimum height in pixels. This value applies only if {@link
     *         #shouldGrabVerticalSpace()} is true. A value of {@link #DEFAULT} means that the
     *         minimum height will be determined by calling {@link Component#getMinimumSize()}.
     */
    public int getMinimumHeight() {
        return mMinHeight;
    }

    /**
     * @param height The minimum height in pixels. This value applies only if {@link
     *               #shouldGrabVerticalSpace()} is true. A value of {@link #DEFAULT} means that the
     *               minimum height will be determined by calling {@link Component#getMinimumSize()}.
     *               The default value is {@link #DEFAULT}.
     * @return This layout data.
     */
    public PrecisionLayoutData setMinimumHeight(int height) {
        mMinHeight = height;
        return this;
    }

    /**
     * @return The number of column cells that the component will take up.
     */
    public int getHorizontalSpan() {
        return mHSpan;
    }

    /**
     * @param span The number of column cells that the component will take up. The default value is
     *             1.
     * @return This layout data.
     */
    public PrecisionLayoutData setHorizontalSpan(int span) {
        mHSpan = span;
        return this;
    }

    /**
     * @return The number of row cells that the component will take up.
     */
    public int getVerticalSpan() {
        return mVSpan;
    }

    /**
     * @param span The number of row cells that the component will take up. The default value is 1.
     * @return This layout data.
     */
    public PrecisionLayoutData setVerticalSpan(int span) {
        mVSpan = span;
        return this;
    }

    /**
     * @return Whether the width of the cell changes depending on the size of the parent container.
     *         If {@code true}, the following rules apply to the width of the cell:
     *         <ul>
     *         <li>If extra horizontal space is available in the parent, the cell will grow to be
     *         wider than its preferred width. The new width will be "preferred width + delta" where
     *         delta is the extra horizontal space divided by the number of grabbing columns.</li>
     *         <li>If there is not enough horizontal space available in the parent, the cell will
     *         shrink until it reaches its minimum width as specified by {@link #getMinimumWidth()}.
     *         The new width will be the maximum of "{@link #getMinimumWidth()}" and "preferred
     *         width - delta", where delta is the amount of space missing divided by the number of
     *         grabbing columns.</li>
     *         </ul>
     */
    public boolean shouldGrabHorizontalSpace() {
        return mHGrab;
    }

    /**
     * @param grab Whether the width of the cell changes depending on the size of the parent
     *             container. If {@code true}, the following rules apply to the width of the cell:
     *             <ul>
     *             <li>If extra horizontal space is available in the parent, the cell will grow to
     *             be wider than its preferred width. The new width will be "preferred width +
     *             delta" where delta is the extra horizontal space divided by the number of
     *             grabbing columns.</li>
     *             <li>If there is not enough horizontal space available in the parent, the cell
     *             will shrink until it reaches its minimum width as specified by
     *             {@link #getMinimumWidth()}. The new width will be the maximum of "
     *             {@link #getMinimumWidth()}" and "preferred width - delta", where delta is the
     *             amount of space missing divided by the number of grabbing columns.</li>
     *             </ul>
     *             The default value is {@code false}.
     * @return This layout data.
     */
    public PrecisionLayoutData setGrabHorizontalSpace(boolean grab) {
        mHGrab = grab;
        return this;
    }

    /**
     * @return Whether the height of the cell changes depending on the size of the parent container.
     *         If {@code true}, the following rules apply to the height of the cell:
     *         <ul>
     *         <li>If extra vertical space is available in the parent, the cell will grow to be
     *         taller than its preferred height. The new height will be "preferred height + delta"
     *         where delta is the extra vertical space divided by the number of grabbing rows.</li>
     *         <li>If there is not enough vertical space available in the parent, the cell will
     *         shrink until it reaches its minimum height as specified by
     *         {@link #getMinimumHeight()}. The new height will be the maximum of "
     *         {@link #getMinimumHeight()}" and "preferred height - delta", where delta is the
     *         amount of space missing divided by the number of grabbing rows.</li>
     *         </ul>
     */
    public boolean shouldGrabVerticalSpace() {
        return mVGrab;
    }

    /**
     * @param grab Whether the height of the cell changes depending on the size of the parent
     *             container. If {@code true}, the following rules apply to the height of the cell:
     *             <ul>
     *             <li>If extra vertical space is available in the parent, the cell will grow to be
     *             taller than its preferred height. The new height will be "preferred height +
     *             delta" where delta is the extra vertical space divided by the number of grabbing
     *             rows.</li>
     *             <li>If there is not enough vertical space available in the parent, the cell will
     *             shrink until it reaches its minimum height as specified by
     *             {@link #getMinimumHeight()}. The new height will be the maximum of "
     *             {@link #getMinimumHeight()}" and "preferred height - delta", where delta is the
     *             amount of space missing divided by the number of grabbing rows.</li>
     *             </ul>
     *             The default value is {@code false}.
     * @return This layout data.
     */
    public PrecisionLayoutData setGrabVerticalSpace(boolean grab) {
        mVGrab = grab;
        return this;
    }

    /**
     * @param grab Whether the size of the cell changes depending on the size of the parent
     *             container.
     * @return This layout data.
     */
    public PrecisionLayoutData setGrabSpace(boolean grab) {
        mHGrab = grab;
        mVGrab = grab;
        return this;
    }

    /**
     * @return {@code true} if the size and position of the component will not be managed by the
     *         layout. {@code false} ifthe size and position of the component will be computed and
     *         assigned.
     */
    public boolean shouldExclude() {
        return mExclude;
    }

    /**
     * Informs the layout to ignore this component when sizing and positioning components.
     *
     * @param exclude {@code true} if the size and position of the component will not be managed by
     *                the layout. {@code false} if the size and position of the component will be
     *                computed and assigned. The default value is {@code false}.
     */
    public PrecisionLayoutData setExclude(boolean exclude) {
        mExclude = exclude;
        return this;
    }

    /** Clear the height and width caches. */
    void clearCache() {
        mCacheMinWidth = 0;
        mCacheWidth = 0;
        mCacheHeight = 0;
    }

    /** @return The cached width. */
    int getCachedWidth() {
        return mCacheWidth;
    }

    /** @return The cached minimum width. */
    int getCachedMinimumWidth() {
        return mCacheMinWidth;
    }

    /** @return The cached height. */
    int getCachedHeight() {
        return mCacheHeight;
    }

    /** @param height The height to set into the cache. */
    void setCachedHeight(int height) {
        mCacheHeight = height;
    }

    void computeSize(Scale scale, Component component, int wHint, int hHint, boolean useMinimumSize) {
        int       scaledMinWidth  = scale.scale(mMinWidth);
        int       scaledMinHeight = scale.scale(mMinHeight);
        Dimension size;
        if (wHint != DEFAULT || hHint != DEFAULT) {
            size = component.getMinimumSize();
            mCacheMinWidth = mMinWidth == DEFAULT ? size.width : scaledMinWidth;
            if (wHint != DEFAULT && wHint < mCacheMinWidth) {
                wHint = mCacheMinWidth;
            }
            int minHeight = mMinHeight == DEFAULT ? size.height : scaledMinHeight;
            if (hHint != DEFAULT && hHint < minHeight) {
                hHint = minHeight;
            }
            size = component.getMaximumSize();
            if (wHint != DEFAULT && wHint > size.width) {
                wHint = size.width;
            }
            if (hHint != DEFAULT && hHint > size.height) {
                hHint = size.height;
            }
        }
        if (useMinimumSize) {
            size = component.getMinimumSize();
            mCacheMinWidth = mMinWidth == DEFAULT ? size.width : scaledMinWidth;
        } else {
            size = component.getPreferredSize();
        }
        if (mWidthHint != DEFAULT) {
            size.width = scale.scale(mWidthHint);
        }
        if (mMinWidth != DEFAULT && size.width < scaledMinWidth) {
            size.width = scaledMinWidth;
        }
        if (mHeightHint != DEFAULT) {
            size.height = scale.scale(mHeightHint);
        }
        if (mMinHeight != DEFAULT && size.height < scaledMinHeight) {
            size.height = scaledMinHeight;
        }
        if (wHint != DEFAULT) {
            size.width = wHint;
        }
        if (hHint != DEFAULT) {
            size.height = hHint;
        }
        mCacheWidth = size.width;
        mCacheHeight = size.height;
    }
}
