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

package com.trollworks.toolkit.widget.tab;

import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.button.TKToggleButton;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;

/**
 * Provides a panel that contains a list of other panels which a user may access by clicking on
 * tabs.
 */
public class TKTabbedPanel extends TKPanel {
	private static final int	MINIMUM_TAB_WIDTH			= 5;
	private static final int	MINIMUM_PRIMARY_TAB_WIDTH	= 20;
	private ArrayList<TKPanel>	mPanels;
	private TabPanel			mTabPanel;
	private TKTabGroup			mTabs;
	private TKPanel				mHoldingPanel;
	private TKPanel				mCurrentPanel;

	/**
	 * Creates a tabbed panel.
	 * 
	 * @param panels The panels to add.
	 * @param titles The titles for the tabs.
	 */
	public TKTabbedPanel(TKPanel[] panels, String[] titles) {
		this(Arrays.asList(panels), Arrays.asList(titles));
	}

	/**
	 * Creates a tabbed panel.
	 * 
	 * @param panels The panels to add.
	 */
	public TKTabbedPanel(TKPanel[] panels) {
		this(Arrays.asList(panels), null);
	}

	/**
	 * Creates a tabbed panel.
	 * 
	 * @param panels The panels to add.
	 */
	public TKTabbedPanel(List<TKPanel> panels) {
		this(panels, null);
	}

	/**
	 * Constructs a tabbed panel which will allow switching between sub-panels in the given panel
	 * list. The title list should contain Strings which will be used for tab titles. Both
	 * ArrayLists must be the same size. The tabbed panel is made up of two primary parts: the
	 * 'page' display panel, and the tab panel. The tab panel contains the tabs for switching
	 * between pages, and the page panel shows the currently selected panel from the list passed to
	 * this constructor.
	 * 
	 * @param panels The list of panels which should be displayed as separate tabbed pages.
	 * @param titles The list of String titles for the tabs.
	 */
	public TKTabbedPanel(List<TKPanel> panels, List<String> titles) {
		this(panels, titles, 10, false);
	}

	/**
	 * Constructs a tabbed panel which will allow switching between sub-panels in the given panel
	 * list. The title list should contain Strings which will be used for tab titles. Both
	 * ArrayLists must be the same size. The tabbed panel is made up of two primary parts: the
	 * 'page' display panel, and the tab panel. The tab panel contains the tabs for switching
	 * between pages, and the page panel shows the currently selected panel from the list passed to
	 * this constructor.
	 * 
	 * @param panels The list of panels which should be displayed as separate tabbed pages.
	 * @param titles The list of titles for the tabs.
	 * @param inset The amount to inset the page
	 * @param hasScrollbars if <code>true</code>, the page being outlined has scrollbars pushed
	 *            all the way to the edge
	 */
	public TKTabbedPanel(List<TKPanel> panels, List<String> titles, int inset, boolean hasScrollbars) {
		if (titles == null) {
			titles = new ArrayList<String>();
			for (TKPanel panel : panels) {
				titles.add(panel.toString());
			}
		}

		if (panels.size() != titles.size()) {
			throw new IllegalStateException("Panel count is not equal to title count"); //$NON-NLS-1$
		}

		setLayout(new TKCompassLayout());

		mPanels = new ArrayList<TKPanel>(panels);
		mTabPanel = new TabPanel();
		mTabs = new TKTabGroup(this);
		mHoldingPanel = new TKPanel(new TKCompassLayout());
		mHoldingPanel.setBorder(new TKTabPageBorder(mTabs, inset, TKColor.CONTROL_LINE, hasScrollbars));

		boolean first = true;
		for (String title : titles) {
			TKTab tab = new TKTab(title, null, first, TKAlignment.CENTER);

			tab.setFirst(first);
			first = false;
			mTabPanel.add(tab);
			mTabs.add(tab);
		}

		add(mTabPanel, TKCompassPosition.NORTH);
		add(mHoldingPanel, TKCompassPosition.CENTER);
		resetSelectedPanel();
	}

	@Override protected Dimension getMinimumSizeSelf() {
		Insets insets = mHoldingPanel.getInsets();
		Dimension tabSize = mTabPanel.getMinimumSize();
		Dimension contentSize = new Dimension();

		for (TKPanel panel : mPanels) {
			Dimension size = panel.getMinimumSize();

			if (size.width > contentSize.width) {
				contentSize.width = size.width;
			}
			if (size.height > contentSize.height) {
				contentSize.height = size.height;
			}
		}

		contentSize.width += insets.left + insets.right;
		contentSize.height += insets.top + insets.bottom;

		if (tabSize.width > contentSize.width) {
			contentSize.width = tabSize.width;
		}
		contentSize.height += tabSize.height;

		insets = getInsets();
		contentSize.width += insets.left + insets.right;
		contentSize.height += insets.top + insets.bottom;

		return sanitizeSize(contentSize);
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = mHoldingPanel.getInsets();
		Dimension tabSize = mTabPanel.getPreferredSize();
		Dimension contentSize = new Dimension();

		for (TKPanel panel : mPanels) {
			Dimension size = panel.getPreferredSize();

			if (size.width > contentSize.width) {
				contentSize.width = size.width;
			}
			if (size.height > contentSize.height) {
				contentSize.height = size.height;
			}
		}

		contentSize.width += insets.left + insets.right;
		contentSize.height += insets.top + insets.bottom;

		if (tabSize.width > contentSize.width) {
			contentSize.width = tabSize.width;
		}
		contentSize.height += tabSize.height;

		insets = getInsets();
		contentSize.width += insets.left + insets.right;
		contentSize.height += insets.top + insets.bottom;

		return sanitizeSize(contentSize);
	}

	@Override protected Dimension getMaximumSizeSelf() {
		Insets insets = mHoldingPanel.getInsets();
		Dimension tabSize = mTabPanel.getMaximumSize();
		Dimension contentSize = new Dimension();

		for (TKPanel panel : mPanels) {
			Dimension size = panel.getMaximumSize();

			if (size.width > contentSize.width) {
				contentSize.width = size.width;
			}
			if (size.height > contentSize.height) {
				contentSize.height = size.height;
			}
		}

		contentSize.width += insets.left + insets.right;
		contentSize.height += insets.top + insets.bottom;

		if (tabSize.width > contentSize.width) {
			contentSize.width = tabSize.width;
		}
		contentSize.height += tabSize.height;

		insets = getInsets();
		contentSize.width += insets.left + insets.right;
		contentSize.height += insets.top + insets.bottom;

		return sanitizeSize(contentSize);
	}

	/** @return The tab panel. */
	TKPanel getTabPanel() {
		return mTabPanel;
	}

	/** @return The tab group. */
	TKTabGroup getTabGroup() {
		return mTabs;
	}

	/** @return The selected panel's name (in the tab). */
	public String getSelectedPanelName() {
		return mTabs.getSelection().getText();
	}

	/**
	 * Sets the selected panel.
	 * 
	 * @param panel The panel to select.
	 */
	public void setSelectedPanel(TKPanel panel) {
		int size = mPanels.size();

		for (int i = 0; i < size; i++) {
			if (mPanels.get(i) == panel) {
				mTabs.getButtons()[i].doClick();
				break;
			}
		}
	}

	/**
	 * @param panel The panel to return the tab name of.
	 * @return The name of the tab this panel is within.
	 */
	public String getPanelTabName(TKPanel panel) {
		int size = mPanels.size();

		for (int i = 0; i < size; i++) {
			if (mPanels.get(i) == panel) {
				return mTabs.getButtons()[i].getText();
			}
		}
		return null;
	}

	/**
	 * @param tabName The name of the tab to return the panel for.
	 * @return The panel, or <code>null</code>.
	 */
	public TKPanel getPanelByTabName(String tabName) {
		TKToggleButton[] buttons = mTabs.getButtons();

		for (int i = 0; i < buttons.length; i++) {
			if (buttons[i].getText().equals(tabName)) {
				return mPanels.get(i);
			}
		}
		return null;
	}

	/**
	 * Sets the selected panel.
	 * 
	 * @param name The name to select.
	 */
	public void setSelectedPanelByName(String name) {
		for (TKToggleButton button : mTabs.getButtons()) {
			if (button.getText().equals(name)) {
				button.doClick();
				break;
			}
		}
	}

	/** @return An unmodifiable list of panels contained within this tab panel. */
	public List<TKPanel> getPanels() {
		return Collections.unmodifiableList(mPanels);
	}

	/**
	 * If the currently displayed panel does not match up with the currently selected tab, switch to
	 * the appropriate panel.
	 */
	public void resetSelectedPanel() {
		TKPanel pagePanel = mPanels.get(mTabs.getSelectedIndex());

		if (mCurrentPanel != pagePanel) {
			if (mCurrentPanel != null) {
				mHoldingPanel.remove(mCurrentPanel);
			}

			mHoldingPanel.add(pagePanel, TKCompassPosition.CENTER);
			mCurrentPanel = pagePanel;
			revalidate();
			Rectangle bounds = mTabPanel.getBounds();
			repaint(bounds.x, bounds.y, bounds.width, bounds.height + 2);
			notifyActionListeners();
		}
	}

	/**
	 * Returns the current panel.
	 * 
	 * @return mCurrentPanel
	 */
	public TKPanel getCurrentPanel() {
		return mCurrentPanel;
	}

	private class TabPanel extends TKPanel implements LayoutManager2 {
		/** Creates a new tab panel. */
		TabPanel() {
			super();
			setLayout(this);
		}

		public void addLayoutComponent(Component comp, Object constraints) {
			// Nothing to do...
		}

		public float getLayoutAlignmentX(Container target) {
			return LEFT_ALIGNMENT;
		}

		public float getLayoutAlignmentY(Container target) {
			return BOTTOM_ALIGNMENT;
		}

		public void invalidateLayout(Container target) {
			// Nothing to do...
		}

		public void addLayoutComponent(String name, Component comp) {
			// Nothing to do...
		}

		public void removeLayoutComponent(Component comp) {
			// Nothing to do...
		}

		public Dimension minimumLayoutSize(Container target) {
			int componentCount = getComponentCount();

			if (componentCount == 0) {
				return new Dimension();
			}

			return sanitizeSize(new Dimension(MINIMUM_TAB_WIDTH * (componentCount - 1) + MINIMUM_PRIMARY_TAB_WIDTH + 1, getComponent(0).getPreferredSize().height));
		}

		public Dimension preferredLayoutSize(Container target) {
			int componentCount = getComponentCount();

			if (componentCount != 0) {
				int width = 1;
				int height = 0;

				for (int i = 0; i < componentCount; i++) {
					Dimension panelSize = getComponent(i).getPreferredSize();

					width += panelSize.width;
					if (i == 0) {
						height = panelSize.height;
					}
				}

				return sanitizeSize(new Dimension(width, height));
			}
			return new Dimension();
		}

		public Dimension maximumLayoutSize(Container target) {
			if (getComponentCount() > 0) {
				return sanitizeSize(new Dimension(MAX_SIZE, getComponent(0).getPreferredSize().height));
			}
			return new Dimension();
		}

		public void layoutContainer(Container target) {
			int componentCount = getComponentCount();

			if (componentCount != 0) {
				Insets insets = getInsets();
				int x = insets.left;
				int y = insets.top;
				int width = getWidth() - (x + insets.right);
				TKTabGroup tabGroup = getTabGroup();
				TKTab[] tabs = tabGroup.getTabs();
				int height = tabs[0].getPreferredSize().height;
				TKTab selectedTab = (TKTab) tabGroup.getSelection();
				int primaryWidth = selectedTab.getPreferredSize().width;
				int remainingWidth = width - primaryWidth;
				int minRemainingWidth = MINIMUM_TAB_WIDTH * (componentCount - 1);
				int secondaryWidth;

				if (remainingWidth < minRemainingWidth) {
					remainingWidth = minRemainingWidth;
					primaryWidth = width - remainingWidth;
					if (primaryWidth < MINIMUM_PRIMARY_TAB_WIDTH) {
						primaryWidth = MINIMUM_PRIMARY_TAB_WIDTH;
					}
				}

				if (componentCount > 1) {
					secondaryWidth = remainingWidth / (componentCount - 1);
				} else {
					secondaryWidth = 1;
				}

				for (int i = 0; i < componentCount; i++) {
					TKTab tab = tabs[i];
					int tabWidth = 0;

					if (tab == selectedTab) {
						tabWidth = primaryWidth;
					} else {
						tabWidth = tab.getPreferredSize().width;

						if (tabWidth > secondaryWidth) {
							tabWidth = secondaryWidth;
						}
					}

					tab.setBounds(x, y, tabWidth, height);
					x += tabWidth - 1;
				}
			}
		}
	}
}
