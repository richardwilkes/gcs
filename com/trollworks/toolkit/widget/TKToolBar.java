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

import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.search.TKSearch;
import com.trollworks.toolkit.widget.search.TKSearchTarget;
import com.trollworks.toolkit.window.TKBaseWindow;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;

/** Creates a tool bar for a window. */
public class TKToolBar extends TKPanel implements ActionListener, TKSearchTarget, LayoutManager2 {
	private TKToolBarTarget				mTarget;
	private HashMap<TKPanel, String>	mItemMap;
	private TKPanel						mFocus;
	private TKProgressBar				mProgressBar;
	private TKBusyIndicator				mBusyIndicator;

	/**
	 * Creates an empty tool bar.
	 * 
	 * @param target The target of this tool bar.
	 */
	public TKToolBar(TKToolBarTarget target) {
		super();
		mTarget = target;
		mItemMap = new HashMap<TKPanel, String>();
		setLayout(this);
		setOpaque(true);
		setBorder(new TKCompoundBorder(new TKLineBorder(TKColor.TOOLBAR_SHADOW, 1, TKLineBorder.BOTTOM_EDGE), new TKEmptyBorder(4)));
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		Object cmd = mItemMap.get(src);

		if (cmd != null) {
			mTarget.obeyToolBarCommand((String) cmd, (TKPanel) src);
		}
	}

	/**
	 * Adds a control to the toolbar.
	 * 
	 * @param item The control to add.
	 * @param index The position on the toolbar to add this control. Pass in <code>-1</code> to
	 *            put it at the end.
	 * @param command The command to associate with this item.
	 */
	public void addControl(TKPanel item, int index, String command) {
		item.addActionListener(this);
		item.setEnabled(item.getDefaultToolbarEnabledState());
		add(item, index);
		mItemMap.put(item, command);
		if (mProgressBar == null && item instanceof TKProgressBar) {
			mProgressBar = (TKProgressBar) item;
		}
		if (mBusyIndicator == null && item instanceof TKBusyIndicator) {
			mBusyIndicator = (TKBusyIndicator) item;
		}
	}

	public void addLayoutComponent(String name, Component component) {
		// Nothing to do...
	}

	public void addLayoutComponent(Component component, Object constraints) {
		// Nothing to do...
	}

	/** Adds a 10-pixel spacer. */
	public void addSpacer() {
		add(new Spacer(10));
	}

	/**
	 * Adds a spacer.
	 * 
	 * @param width The width of the spacer.
	 */
	public void addSpacer(int width) {
		add(new Spacer(width));
	}

	/** Adds a flexible spacer that will consume any remaining space. */
	public void addFlexibleSpacer() {
		add(new FlexibleSpacer());
	}

	/** @return The first progress bar found in the toolbar, if any. */
	public TKProgressBar getProgressBar() {
		return mProgressBar;
	}

	/** @return The first busy indicator found in the toolbar, if any. */
	public TKBusyIndicator getBusyIndicator() {
		return mBusyIndicator;
	}

	/** Force the tool bar to adjust itself to the current target. */
	public void adjustToolBar() {
		for (Map.Entry<TKPanel, String> entry : mItemMap.entrySet()) {
			TKPanel panel = entry.getKey();

			if (!mTarget.adjustToolBarItem(entry.getValue(), panel)) {
				panel.setEnabled(panel.getDefaultToolbarEnabledState());
			}
		}
	}

	/**
	 * @param command The command to look for.
	 * @return The item that issues the specified command.
	 */
	public TKPanel getItemForCommand(String command) {
		for (Map.Entry<TKPanel, String> entry : mItemMap.entrySet()) {
			if (command.equals(entry.getValue())) {
				return entry.getKey();
			}
		}
		return null;
	}

	/**
	 * @return The current keyboard focus in this tool bar. <code>null</code> will be returned if
	 *         there is no focused component.
	 */
	public TKPanel getKeyboardFocus() {
		return mFocus;
	}

	public float getLayoutAlignmentX(Container target) {
		return CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container target) {
		return CENTER_ALIGNMENT;
	}

	public TKItemRenderer getSearchRenderer() {
		TKBaseWindow window = getBaseWindow();

		return window instanceof TKSearchTarget ? ((TKSearchTarget) window).getSearchRenderer() : null;
	}

	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	public void layoutContainer(Container target) {
		Insets insets = getInsets();
		int width = getWidth() - (insets.left + insets.right);
		int height = getHeight() - (insets.top + insets.bottom);
		int count = getComponentCount();
		int x = insets.left;
		Dimension[] sizes = new Dimension[count];
		int[] minWidths = new int[count];
		int widthComps = 0;
		int i;
		TKPanel comp;
		int amt;

		for (i = 0; i < count; i++) {
			comp = (TKPanel) getComponent(i);
			sizes[i] = comp.getPreferredSize();
			minWidths[i] = comp.getMinimumSize().width;
			if (!(comp instanceof TKSearch || comp instanceof FlexibleSpacer)) {
				width -= sizes[i].width;
				widthComps++;
			}
		}

		if (width < 0 && widthComps > 0) {
			int slice = width / widthComps;

			if (slice < 0) {
				for (i = 0; i < count; i++) {
					comp = (TKPanel) getComponent(i);
					if (!(comp instanceof TKSearch || comp instanceof FlexibleSpacer)) {
						amt = sizes[i].width + slice;
						if (amt < minWidths[i]) {
							amt = minWidths[i];
						}
						width += sizes[i].width - amt;
						sizes[i].width = amt;
					}
				}
			}

			for (i = 0; i < count && width < 0; i++) {
				comp = (TKPanel) getComponent(i);
				if (!(comp instanceof TKSearch || comp instanceof FlexibleSpacer)) {
					amt = sizes[i].width + width;
					if (amt < minWidths[i]) {
						amt = minWidths[i];
					}
					width += sizes[i].width - amt;
					sizes[i].width = amt;
				}
			}
		}

		for (i = 0; i < count; i++) {
			comp = (TKPanel) getComponent(i);
			if (!(comp instanceof TKSearch || comp instanceof FlexibleSpacer)) {
				comp.setBounds(x, insets.top + (height - sizes[i].height) / 2, sizes[i].width, sizes[i].height);
				x += sizes[i].width;
			} else {
				amt = width;
				if (amt < minWidths[i]) {
					amt = minWidths[i];
				}

				comp.setBounds(x, insets.top + (height - sizes[i].height) / 2, amt, sizes[i].height);
				x += amt;
			}
		}
	}

	public Dimension maximumLayoutSize(Container target) {
		Insets insets = getInsets();
		int count = getComponentCount();
		Dimension size = new Dimension();

		for (int i = 0; i < count; i++) {
			Component comp = getComponent(i);
			Dimension maxSize = comp instanceof TKSearch ? comp.getMaximumSize() : comp.getPreferredSize();

			if (maxSize.height > size.height) {
				size.height = maxSize.height;
			}
			size.width += maxSize.width;
		}

		size.height += insets.top + insets.bottom;
		size.width += insets.left + insets.right;
		return size;
	}

	public Dimension minimumLayoutSize(Container target) {
		Insets insets = getInsets();
		int count = getComponentCount();
		Dimension size = new Dimension();

		for (int i = 0; i < count; i++) {
			Component comp = getComponent(i);
			Dimension minSize = comp.getMinimumSize();

			if (minSize.height > size.height) {
				size.height = minSize.height;
			}
			size.width += minSize.width;
		}

		size.height += insets.top + insets.bottom;
		size.width += insets.left + insets.right;
		return size;
	}

	public Dimension preferredLayoutSize(Container target) {
		Insets insets = getInsets();
		int count = getComponentCount();
		Dimension size = new Dimension();

		for (int i = 0; i < count; i++) {
			Dimension prefSize = getComponent(i).getPreferredSize();

			if (prefSize.height > size.height) {
				size.height = prefSize.height;
			}
			size.width += prefSize.width;
		}

		size.height += insets.top + insets.bottom;
		size.width += insets.left + insets.right;
		return size;
	}

	public Collection<Object> search(String filter) {
		TKBaseWindow window = getBaseWindow();

		return window instanceof TKSearchTarget ? ((TKSearchTarget) window).search(filter) : new ArrayList<Object>();
	}

	public void searchSelect(Collection<Object> selection) {
		TKBaseWindow window = getBaseWindow();

		if (window instanceof TKSearchTarget) {
			((TKSearchTarget) window).searchSelect(selection);
		}
	}

	public void removeLayoutComponent(Component target) {
		// Nothing to do...
	}

	private class Spacer extends TKPanel {
		/**
		 * Creates a new, fixed size spacer.
		 * 
		 * @param width The width of the spacer.
		 */
		Spacer(int width) {
			super();
			setOnlySize(new Dimension(width, 1));
		}
	}

	private class FlexibleSpacer extends TKPanel {
		/** Creates a new, flexible spacer. */
		FlexibleSpacer() {
			super();
		}
	}
}
