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

package com.trollworks.gcs.ui;

import java.util.BitSet;

/** Provides standardized handling of selections. */
public class Selection {
    /** The constant if no relevant keys were held down during a mouse-click. */
    public static final int            MOUSE_NONE   = 0;
    /** The constant if the shift key was held down during a mouse-click. */
    public static final int            MOUSE_EXTEND = 1;
    /** The constant if the platform command key was held down during a mouse-click. */
    public static final int            MOUSE_FLIP   = 2;
    private             BitSet         mSelection;
    private             int            mSize;
    private             int            mAnchor;
    private             SelectionOwner mOwner;

    /**
     * Creates a new selection.
     *
     * @param owner The owner of this selection.
     */
    public Selection(SelectionOwner owner) {
        mOwner = owner;
        mSelection = new BitSet();
        mAnchor = -1;
    }

    /**
     * Creates a new selection.
     *
     * @param owner The owner of this selection.
     * @param size  The initial size.
     */
    public Selection(SelectionOwner owner, int size) {
        this(owner);
        mSize = size;
    }

    /**
     * Creates a clone of the specified selection.
     *
     * @param other The selection to clone.
     */
    public Selection(Selection other) {
        mOwner = other.mOwner;
        mSelection = (BitSet) other.mSelection.clone();
        mSize = other.mSize;
        mAnchor = other.mAnchor;
    }

    /** @return The anchor index (used when extending selections). */
    public int getAnchor() {
        return mAnchor;
    }

    /** @param index Sets the anchor index (used when extending selections). */
    public void setAnchor(int index) {
        mAnchor = index < 0 || index >= mSize ? -1 : index;
    }

    /** @param size The number of valid indexes. */
    public void setSize(int size) {
        int lastPlusOne = mSelection.length();
        if (size < 0) {
            size = 0;
        }
        if (lastPlusOne >= size) {
            mSelection.clear(size, lastPlusOne);
        }
        mSize = size;
        if (mAnchor >= mSize) {
            mAnchor = firstSelectedIndex();
        }
    }

    /**
     * Checks the specified index to determine if it is selected.
     *
     * @param index The index to check.
     * @return {@code true} if the specified index is selected, {@code false} if it isn't.
     */
    public boolean isSelected(int index) {
        if (index < 0 || index >= mSize) {
            return false;
        }
        return mSelection.get(index);
    }

    /** @return Whether the selection is empty or not. */
    public boolean isEmpty() {
        return lastSelectedIndex() == -1;
    }

    /** @return The number of selected indexes. */
    public int getCount() {
        if (mSelection.length() > mSize) {
            // If the last selected index was outside our valid range, we create
            // a new BitSet that only contains bits in our valid range.
            mSelection = mSelection.get(0, mSize);
        }
        return mSelection.cardinality();
    }

    /** @return The first selected index, or {@code -1} if there is none. */
    public int firstSelectedIndex() {
        return nextSelectedIndex(0);
    }

    /**
     * @param fromIndex The index to start checking at.
     * @return The index of the first selected index on or after {@code fromIndex}, or {@code -1} if
     *         there is none.
     */
    public int nextSelectedIndex(int fromIndex) {
        int index = mSelection.nextSetBit(fromIndex);
        return index < mSize ? index : -1;
    }

    /** @return The last selected index, or {@code -1} if there is none. */
    public int lastSelectedIndex() {
        int index = mSelection.length() - 1;
        return index < mSize ? index : -1;
    }

    /** @return An array of selected indexes. */
    public int[] getSelectedIndexes() {
        int[] array = new int[getCount()];
        int   i     = 0;
        int   index = firstSelectedIndex();
        while (index != -1) {
            array[i++] = index;
            index = nextSelectedIndex(index + 1);
        }
        return array;
    }

    /** @return Whether calling {@link #select()} will do anything. */
    public boolean canSelectAll() {
        return mSize != getCount();
    }

    /** Selects all indexes. Sets the anchor to the first index. */
    public void select() {
        if (mSize > 0 && canSelectAll()) {
            selectionAboutToChange();
            mSelection.set(0, mSize);
            mAnchor = 0;
            selectionDidChange();
        }
    }

    /**
     * Selects a specific index. If the selection is replaced or there was no prior selection, the
     * anchor is set to the {@code index}.
     *
     * @param index The index to select.
     * @param add   Pass in {@code true} to add the index to the current selection, {@code false} to
     *              replace the current selection.
     */
    public void select(int index, boolean add) {
        BitSet newSel = (BitSet) mSelection.clone();
        if (!add) {
            newSel.clear();
        }
        int last = newSel.length() - 1;
        last = last < mSize ? last : -1;
        if (last == -1) {
            mAnchor = index < 0 || index >= mSize ? -1 : index;
        }
        if (index >= 0 && index < mSize) {
            newSel.set(index);
        }
        applySelectionChange(newSel);
    }

    /**
     * Selects the range {@code fromIndex} (inclusive) to {@code toIndex} (inclusive). If the
     * selection is replaced or there was no prior selection, the anchor is set to {@code
     * fromIndex}.
     *
     * @param fromIndex The first index to select.
     * @param toIndex   The last index to select.
     * @param add       Pass in {@code true} to add the range to the current selection, {@code
     *                  false} to replace the current selection.
     */
    public void select(int fromIndex, int toIndex, boolean add) {
        BitSet newSel = (BitSet) mSelection.clone();
        if (!add) {
            newSel.clear();
        }
        int index = newSel.length() - 1;
        index = index < mSize ? index : -1;
        if (index == -1) {
            mAnchor = fromIndex < 0 || fromIndex >= mSize ? -1 : fromIndex;
        }
        if (fromIndex > toIndex) {
            int tmp = fromIndex;
            fromIndex = toIndex;
            toIndex = tmp;
        }
        if (fromIndex < mSize) {
            if (toIndex >= mSize) {
                toIndex = mSize - 1;
            }
            newSel.set(fromIndex, toIndex + 1);
        }
        applySelectionChange(newSel);
    }

    /**
     * Selects the specified indexes. If the selection is replaced or there was no prior selection,
     * the anchor is set to the first index, if any.
     *
     * @param indexes The indexes to select.
     * @param add     Pass in {@code true} to add the indexes to the current selection, {@code
     *                false} to replace the current selection.
     */
    public void select(int[] indexes, boolean add) {
        BitSet newSel = (BitSet) mSelection.clone();
        if (!add) {
            newSel.clear();
        }
        int index = newSel.length() - 1;
        index = index < mSize ? index : -1;
        if (index == -1) {
            if (indexes.length > 0) {
                index = indexes[0];
                mAnchor = index < 0 || index >= mSize ? -1 : index;
            } else {
                mAnchor = -1;
            }
        }
        for (int element : indexes) {
            index = element;
            if (index > -1 && index < mSize) {
                newSel.set(index);
            }
        }
        applySelectionChange(newSel);
    }

    /**
     * Does a logical "select up" (such as from the keyboard up-arrow).
     *
     * @param extend Whether to extend the current selection or replace it.
     * @return The index that should be scrolled into view, or {@code -1} if there isn't one.
     */
    public int selectUp(boolean extend) {
        int index = -1;
        if (mSize > 0) {
            int count = getCount();
            if (extend && count > 0) {
                index = lastSelectedIndex();
                if (index > mAnchor) {
                    deselect(index--);
                } else {
                    index = firstSelectedIndex() - 1;
                    if (index >= 0) {
                        select(index, true);
                    }
                }
            } else {
                if (count == 0) {
                    select(mSize - 1, false);
                } else {
                    index = firstSelectedIndex();
                    if (count == 1) {
                        index--;
                    }
                    if (index >= 0) {
                        select(index, false);
                    }
                }
            }
        }
        return index < 0 || index >= mSize ? -1 : index;
    }

    /**
     * Does a logical "select down" (such as from the keyboard down-arrow).
     *
     * @param extend Whether to extend the current selection or replace it.
     * @return The index that should be scrolled into view, or {@code -1} if there isn't one.
     */
    public int selectDown(boolean extend) {
        int index = -1;
        if (mSize > 0) {
            int count = getCount();
            if (extend && count > 0) {
                index = firstSelectedIndex();
                if (index < mAnchor) {
                    deselect(index++);
                } else {
                    index = lastSelectedIndex() + 1;
                    if (index < mSize) {
                        select(index, true);
                    }
                }
            } else {
                if (count == 0) {
                    select(0, false);
                } else {
                    index = lastSelectedIndex();
                    if (count == 1) {
                        index++;
                    }
                    if (index < mSize) {
                        select(index, false);
                    }
                }
            }
        }
        return index < 0 || index >= mSize ? -1 : index;
    }

    /**
     * Does a logical "select home" (such as from the keyboard HOME key).
     *
     * @param extend Whether to extend the current selection or replace it.
     * @return The index that should be scrolled into view, or {@code -1} if there isn't one.
     */
    public int selectToHome(boolean extend) {
        if (mSize > 0) {
            if (extend && !isEmpty()) {
                if (mAnchor < 0) {
                    mAnchor = lastSelectedIndex();
                }
                select(0, mAnchor, true);
            } else {
                select(0, false);
            }
            return 0;
        }
        return -1;
    }

    /**
     * Does a logical "select end" (such as from the keyboard END key).
     *
     * @param extend Whether to extend the current selection or replace it.
     * @return The index that should be scrolled into view, or {@code -1} if there isn't one.
     */
    public int selectToEnd(boolean extend) {
        if (mSize > 0) {
            if (extend && !isEmpty()) {
                if (mAnchor < 0) {
                    mAnchor = firstSelectedIndex();
                }
                select(mAnchor, mSize - 1, true);
            } else {
                select(mSize - 1, false);
            }
            return mSize - 1;
        }
        return -1;
    }

    /**
     * Helper method for selecting one or more indexes by mouse.
     *
     * @param index  The index that was clicked on.
     * @param method The type of click, one of {@link #MOUSE_NONE}, {@link #MOUSE_EXTEND}, or {@link
     *               #MOUSE_FLIP}.
     * @return {@code -1} if no selection on mouse-up should occur, or the passed in index if
     *         selection on mouse-up is a possibility.
     */
    public int selectByMouse(int index, int method) {
        if (index == -1) {
            deselect();
        } else {
            if (mAnchor >= 0 && (method & MOUSE_EXTEND) == MOUSE_EXTEND) {
                select(mAnchor, index, true);
            } else if ((method & MOUSE_FLIP) == MOUSE_FLIP) {
                if (isSelected(index)) {
                    deselect(index);
                } else {
                    select(index, true);
                }
            } else if (!isSelected(index)) {
                select(index, false);
            } else if (getCount() != 1) {
                // Index was selected prior to this... note
                // that it might be eligible for becoming
                // the sole selection if it wasn't already
                return index;
            }
        }
        return -1;
    }

    /** Deselects all indexes. */
    public void deselect() {
        if (!isEmpty()) {
            selectionAboutToChange();
            mSelection.clear();
            mAnchor = -1;
            selectionDidChange();
        }
    }

    /**
     * Deselects a specific index.
     *
     * @param index The index to deselect.
     */
    public void deselect(int index) {
        if (mSelection.get(index)) {
            selectionAboutToChange();
            mSelection.clear(index);
            if (mAnchor == index) {
                mAnchor = firstSelectedIndex();
            }
            selectionDidChange();
        }
    }

    /**
     * Deselects the range {@code fromIndex} (inclusive) to {@code toIndex} (inclusive).
     *
     * @param fromIndex The first index to deselect.
     * @param toIndex   The last index to deselect.
     */
    public void deselect(int fromIndex, int toIndex) {
        BitSet newSel = (BitSet) mSelection.clone();
        if (fromIndex > toIndex) {
            int tmp = fromIndex;
            fromIndex = toIndex;
            toIndex = tmp;
        }
        newSel.clear(fromIndex, toIndex + 1);
        if (mAnchor >= fromIndex && mAnchor <= toIndex) {
            int index = newSel.nextSetBit(0);
            mAnchor = index < mSize ? index : -1;
        }
        applySelectionChange(newSel);
    }

    /**
     * Deselects the specified indexes.
     *
     * @param indexes The indexes to deselect.
     */
    public void deselect(int[] indexes) {
        BitSet newSel = (BitSet) mSelection.clone();
        for (int element : indexes) {
            newSel.clear(element);
            if (mAnchor == element) {
                mAnchor = -1;
            }
        }
        if (mAnchor == -1) {
            int index = newSel.nextSetBit(0);
            mAnchor = index < mSize ? index : -1;
        }
        applySelectionChange(newSel);
    }

    private void applySelectionChange(BitSet newSel) {
        if (!newSel.equals(mSelection)) {
            selectionAboutToChange();
            mSelection = newSel;
            selectionDidChange();
        }
    }

    private void selectionDidChange() {
        if (mOwner != null) {
            mOwner.selectionDidChange();
        }
    }

    private void selectionAboutToChange() {
        if (mOwner != null) {
            mOwner.selectionAboutToChange();
        }
    }
}
