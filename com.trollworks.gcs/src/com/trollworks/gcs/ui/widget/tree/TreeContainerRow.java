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

package com.trollworks.gcs.ui.widget.tree;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Set;

/** A {@link TreeRow} which can have children. */
public class TreeContainerRow extends TreeRow {
    private ArrayList<TreeRow> mChildren = new ArrayList<>();

    @Override
    protected TreeContainerRow clone() {
        TreeContainerRow   other    = (TreeContainerRow) super.clone();
        ArrayList<TreeRow> children = new ArrayList<>(mChildren.size());
        for (TreeRow row : mChildren) {
            row = row.clone();
            children.add(row);
            row.mParent = other;
        }
        other.mChildren = children;
        other.renumber(0);
        return other;
    }

    /** @return The number of direct children this {@link TreeContainerRow} contains. */
    public int getChildCount() {
        return mChildren.size();
    }

    public int getRecursiveChildCount() {
        int count = mChildren.size();
        if (count > 0) {
            for (TreeRow row : mChildren) {
                if (row instanceof TreeContainerRow) {
                    count += ((TreeContainerRow) row).getRecursiveChildCount();
                }
            }
        }
        return count;
    }

    /**
     * @return An unmodifiable {@link List} containing the children of this {@link TreeContainerRow}
     *         .
     */
    public List<TreeRow> getChildren() {
        return Collections.unmodifiableList(mChildren);
    }

    /**
     * @param index The index of the child to return.
     * @return The child {@link TreeRow} at the specified index.
     */
    public TreeRow getChild(int index) {
        return mChildren.get(index);
    }

    /**
     * Adds the specified {@link TreeRow} as a direct child to this {@link TreeContainerRow}. If the
     * {@link TreeRow} already has a parent, it will first be removed from that parent.
     *
     * @param row The {@link TreeRow} to add as a child.
     */
    public void addRow(TreeRow row) {
        addRow(mChildren.size(), row);
    }

    /**
     * Adds the specified {@link TreeRow}s as direct children to this {@link TreeContainerRow}. If
     * the {@link TreeRow} already has a parent, it will first be removed from that parent.
     *
     * @param rows The {@link TreeRow}s to add as children.
     */
    public void addRow(List<TreeRow> rows) {
        addRow(mChildren.size(), rows);
    }

    /**
     * Adds the specified {@link TreeRow} as a direct child to this {@link TreeContainerRow} at the
     * specified index. If the {@link TreeRow} already has a parent, it will first be removed from
     * that parent. If the {@link TreeRow} is already a child, but not in the specified position, it
     * will be moved into that position.
     *
     * @param index The index to insert at.
     * @param row   The {@link TreeRow} to add as a child.
     */
    public void addRow(int index, TreeRow row) {
        List<TreeRow> rows = new ArrayList<>(1);
        rows.add(row);
        addRow(index, rows);
    }

    /**
     * Adds the specified {@link TreeRow}s as direct children to this {@link TreeContainerRow} at
     * the specified index. If the {@link TreeRow} already has a parent, it will first be removed
     * from that parent. If the {@link TreeRow} is already a child, but not in the specified
     * position, it will be moved into that position.
     *
     * @param index The index to insert at.
     * @param rows  The {@link TreeRow}s to add as children.
     */
    public void addRow(int index, List<TreeRow> rows) {
        if (!rows.isEmpty()) {
            Map<TreeContainerRow, Set<TreeRow>> map    = new HashMap<>();
            Set<TreeRow>                        exists = new HashSet<>();
            List<TreeRow>                       list   = new ArrayList<>();
            for (TreeRow row : rows) {
                if (row.mParent != null) {
                    Set<TreeRow> set = map.get(row.mParent);
                    if (set == null) {
                        set = new HashSet<>();
                        map.put(row.mParent, set);
                    }
                    set.add(row);
                }
                if (!exists.contains(row)) {
                    exists.add(row);
                    list.add(row);
                }
            }
            for (Entry<TreeContainerRow, Set<TreeRow>> entry : map.entrySet()) {
                entry.getKey().removeRow(entry.getValue());
            }
            index = Math.min(index, mChildren.size());
            int start = index;
            for (TreeRow row : list) {
                mChildren.add(index++, row);
                row.mParent = this;
            }
            renumber(start);
            notify(TreeNotificationKeys.ROW_ADDED, rows.toArray(new TreeRow[list.size()]));
        }
    }

    /**
     * Removes a child {@link TreeRow} from this {@link TreeContainerRow}. If the specified {@link
     * TreeRow} is not an immediate child of this {@link TreeContainerRow}, nothing will be done.
     *
     * @param row The child {@link TreeRow} to remove.
     */
    public void removeRow(TreeRow row) {
        List<TreeRow> rows = new ArrayList<>(1);
        rows.add(row);
        removeRow(rows);
    }

    /**
     * Removes the specified child {@link TreeRow}s from this {@link TreeContainerRow}. If a {@link
     * TreeRow} is not an immediate child of this {@link TreeContainerRow}, nothing will be done to
     * it.
     *
     * @param rows The child {@link TreeRow}s to remove.
     */
    public void removeRow(Collection<TreeRow> rows) {
        if (!rows.isEmpty()) {
            HashSet<TreeRow> set = new HashSet<>();
            for (TreeRow row : rows) {
                if (row.mParent == this) {
                    set.add(row);
                }
            }
            int count = set.size();
            if (count > 0) {
                TreeRow[] removed = set.toArray(new TreeRow[count]);
                Arrays.sort(removed, TreeRowIndexComparator.INSTANCE);
                int start = removed[0].getIndex();
                for (TreeRow row : removed) {
                    mChildren.remove(row);
                    row.mParent = null;
                }
                renumber(start);
                notify(TreeNotificationKeys.ROW_REMOVED, removed);
            }
        }
    }

    private void renumber(int start) {
        int max = mChildren.size();
        while (start < max) {
            mChildren.get(start).setIndex(start);
            start++;
        }
    }

    /** @param sorter The {@link TreeSorter} to use. */
    public void sort(TreeSorter sorter) {
        if (!mChildren.isEmpty()) {
            for (TreeRow child : mChildren) {
                if (child instanceof TreeContainerRow) {
                    ((TreeContainerRow) child).sort(sorter);
                }
            }
            mChildren.sort(sorter);
            renumber(0);
        }
    }

    /**
     * @param list A {@link List} to add the child containers into. May be {@code null}.
     * @return The child containers. This list will be the one passed in, or a newly created one, if
     *         {@code null} was passed for the {@code list} parameter.
     */
    public List<TreeContainerRow> getRecursiveChildContainers(List<TreeContainerRow> list) {
        if (list == null) {
            list = new ArrayList<>();
        }
        for (TreeRow child : mChildren) {
            if (child instanceof TreeContainerRow) {
                TreeContainerRow container = (TreeContainerRow) child;
                list.add(container);
                container.getRecursiveChildContainers(list);
            }
        }
        return list;
    }
}
