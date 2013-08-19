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

import com.trollworks.toolkit.undo.TKUndoManager;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.TKRectUtils;
import com.trollworks.toolkit.widget.border.TKBorder;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKTooltip;
import com.trollworks.toolkit.window.TKUserInputManager;

import java.awt.AWTEvent;
import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Cursor;
import java.awt.Dialog;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.GraphicsConfiguration;
import java.awt.Insets;
import java.awt.LayoutManager;
import java.awt.LayoutManager2;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Shape;
import java.awt.Transparency;
import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.awt.event.MouseWheelEvent;
import java.awt.image.BufferedImage;
import java.util.ArrayList;
import java.util.List;

/** The base class for all graphical components. */
public class TKPanel extends Container {
	/** The default maximum size. */
	public static final int					MAX_SIZE		= Integer.MAX_VALUE / 8;
	/** A sequence counter to allow us to drop outdated focus requests. */
	protected static int					FOCUS_SEQUENCE	= 0;
	private TKBorder						mBorder;
	private boolean							mAlignmentXSet;
	private boolean							mAlignmentYSet;
	private float							mAlignmentX;
	private float							mAlignmentY;
	private boolean							mInGetMaximumSize;
	private boolean							mInGetMinimumSize;
	private Dimension						mMaximumSize;
	private Dimension						mMinimumSize;
	private boolean							mOpaque;
	private Dimension						mPreferredSize;
	private String							mToolTipText;
	private ArrayList<TKReshapeListener>	mReshapeListeners;
	private TKPanel							mExcludeChild;
	private Dimension						mCalcedMaximumSize;
	private Dimension						mCalcedMinimumSize;
	private Dimension						mCalcedPreferredSize;
	private long							mAWTEventMask;
	private boolean							mInLayerChange;
	private ArrayList<ActionListener>		mActionListeners;
	private String							mActionCommand;
	private boolean							mGetMouseEventsWhenDisabled;
	private String							mFontKey;

	/** Creates a new panel with no layout. */
	public TKPanel() {
		this(null);
	}

	/**
	 * Creates a new panel with the specified layout.
	 * 
	 * @param layout The layout manager to use. May be <code>null</code>.
	 */
	public TKPanel(LayoutManager layout) {
		super();
		mAlignmentX = CENTER_ALIGNMENT;
		mAlignmentY = CENTER_ALIGNMENT;
		setLayout(layout);
		setForeground(Color.black);
		setBackground(TKColor.LIGHT_BACKGROUND);
		setCursor(Cursor.getDefaultCursor());
		enableAWTEvents(AWTEvent.MOUSE_EVENT_MASK);
	}

	/**
	 * Called by the owning window when it gains/loses focus and this component is the current
	 * keyboard focus.
	 * 
	 * @param gained <code>true</code> will be passed in when the focus has been gained.
	 *            <code>false</code> will be passed in when the focus has been lost.
	 */
	public void windowFocus(@SuppressWarnings("unused") boolean gained) {
		// Does nothing by default.
	}

	/**
	 * Adds a reshape listener to this panel.
	 * 
	 * @param listener The listener to add.
	 */
	public void addReshapeListener(TKReshapeListener listener) {
		if (mReshapeListeners == null) {
			mReshapeListeners = new ArrayList<TKReshapeListener>(1);
		}
		if (!mReshapeListeners.contains(listener)) {
			mReshapeListeners.add(listener);
		}
	}

	/**
	 * Removes a reshape listener from this panel.
	 * 
	 * @param listener The listener to remove.
	 */
	public void removeReshapeListener(TKReshapeListener listener) {
		if (mReshapeListeners != null) {
			mReshapeListeners.remove(listener);
			if (mReshapeListeners.isEmpty()) {
				mReshapeListeners = null;
			}
		}
	}

	/**
	 * Adds an action listener.
	 * 
	 * @param listener The listener to add.
	 */
	public void addActionListener(ActionListener listener) {
		if (mActionListeners == null) {
			mActionListeners = new ArrayList<ActionListener>(1);
		}
		if (!mActionListeners.contains(listener)) {
			mActionListeners.add(listener);
		}
	}

	/**
	 * Removes an action listener.
	 * 
	 * @param listener The listener to remove.
	 */
	public void removeActionListener(ActionListener listener) {
		if (mActionListeners != null) {
			mActionListeners.remove(listener);
			if (mActionListeners.isEmpty()) {
				mActionListeners = null;
			}
		}
	}

	/** Notifies all action listeners with the standard action command. */
	public void notifyActionListeners() {
		notifyActionListeners(new ActionEvent(this, ActionEvent.ACTION_PERFORMED, getActionCommand()));
	}

	/**
	 * Notifies all action listeners.
	 * 
	 * @param event The action event to notify with.
	 */
	public void notifyActionListeners(ActionEvent event) {
		if (mActionListeners != null && event != null) {
			for (ActionListener listener : new ArrayList<ActionListener>(mActionListeners)) {
				listener.actionPerformed(event);
			}
		}
	}

	/** @return The action command. */
	public String getActionCommand() {
		return mActionCommand;
	}

	/** @param command The action command. */
	public void setActionCommand(String command) {
		mActionCommand = command;
	}

	/**
	 * Adjust the passed in size to account for minimum/maximum sizes on this component.
	 * 
	 * @param size The size.
	 */
	public void adjustForMinMaxSize(Dimension size) {
		Dimension mySize = getMaximumSize();

		if (size.width > mySize.width) {
			size.width = mySize.width;
		}
		if (size.height > mySize.height) {
			size.height = mySize.height;
		}
		mySize = getMinimumSize();
		if (size.width < mySize.width) {
			size.width = mySize.width;
		}
		if (size.height < mySize.height) {
			size.height = mySize.height;
		}
	}

	/**
	 * Converts a {@link Point} from one component's coordinate system to another's.
	 * 
	 * @param pt The point to convert.
	 * @param from The component the point originated in.
	 * @param to The component the point should be translated to.
	 */
	static public void convertPoint(Point pt, Component from, Component to) {
		convertPointToScreen(pt, from);
		convertPointFromScreen(pt, to);
	}

	/**
	 * Converts a {@link Point} from on the screen to a position within the component.
	 * 
	 * @param pt The point to convert.
	 * @param component The component the point should be translated to.
	 */
	static public void convertPointFromScreen(Point pt, Component component) {
		do {
			pt.x -= component.getX();
			pt.y -= component.getY();
			if (component instanceof TKBaseWindow) {
				break;
			}
			component = component.getParent();
		} while (component != null);
	}

	/**
	 * Converts a {@link Point} in a component to its position on the screen.
	 * 
	 * @param pt The point to convert.
	 * @param component The component the point originated in.
	 */
	static public void convertPointToScreen(Point pt, Component component) {
		do {
			pt.x += component.getX();
			pt.y += component.getY();
			if (component instanceof TKBaseWindow) {
				break;
			}
			component = component.getParent();
		} while (component != null);
	}

	/**
	 * Converts a {@link Rectangle} from one component's coordinate system to another's.
	 * 
	 * @param bounds The rectangle to convert.
	 * @param from The component the rectangle originated in.
	 * @param to The component the rectangle should be translated to.
	 */
	static public void convertRectangle(Rectangle bounds, Component from, Component to) {
		convertRectangleToScreen(bounds, from);
		convertRectangleFromScreen(bounds, to);
	}

	/**
	 * Converts a {@link Rectangle} from on the screen to a position within the component.
	 * 
	 * @param bounds The rectangle to convert.
	 * @param component The component the rectangle should be translated to.
	 */
	static public void convertRectangleFromScreen(Rectangle bounds, Component component) {
		do {
			bounds.x -= component.getX();
			bounds.y -= component.getY();
			if (component instanceof TKBaseWindow) {
				break;
			}
			component = component.getParent();
		} while (component != null);
	}

	/**
	 * Converts a {@link Rectangle} in a component to its position on the screen.
	 * 
	 * @param bounds The rectangle to convert.
	 * @param component The component the rectangle originated in.
	 */
	static public void convertRectangleToScreen(Rectangle bounds, Component component) {
		do {
			bounds.x += component.getX();
			bounds.y += component.getY();
			if (component instanceof TKBaseWindow) {
				break;
			}
			component = component.getParent();
		} while (component != null);
	}

	/**
	 * @return Whether to deliver mouse events to the component even it has been disabled.
	 *         <code>false</code> by default.
	 */
	public boolean getMouseEventsEvenWhenDisabled() {
		return mGetMouseEventsWhenDisabled;
	}

	/**
	 * @param enabled Whether to deliver mouse events to the component even it has been disabled.
	 *            <code>false</code> by default.
	 */
	public void setGetMouseEventsEvenWhenDisabled(boolean enabled) {
		mGetMouseEventsWhenDisabled = enabled;
	}

	/**
	 * Expose the {@link Component#disableEvents(long)} method.
	 * 
	 * @param mask The AWT event mask.
	 */
	public void disableAWTEvents(long mask) {
		disableEvents(mask);
		mAWTEventMask &= ~mask;
	}

	/**
	 * Expose the {@link Component#enableEvents(long)} method.
	 * 
	 * @param mask The AWT event mask.
	 */
	public void enableAWTEvents(long mask) {
		enableEvents(mask);
		mAWTEventMask |= mask;
	}

	/**
	 * @param mask The AWT event mask.
	 * @return <code>true</code> if the specified AWT mask is enabled.
	 */
	public boolean isAWTEventEnabled(long mask) {
		long tmpMask = mAWTEventMask;

		if ((mask & AWTEvent.MOUSE_EVENT_MASK) != 0 && (tmpMask & AWTEvent.MOUSE_EVENT_MASK) == 0 && getListeners(MouseListener.class).length > 0) {
			tmpMask |= AWTEvent.MOUSE_EVENT_MASK;
		}

		if ((mask & AWTEvent.MOUSE_MOTION_EVENT_MASK) != 0 && (tmpMask & AWTEvent.MOUSE_MOTION_EVENT_MASK) == 0 && getListeners(MouseMotionListener.class).length > 0) {
			tmpMask |= AWTEvent.MOUSE_MOTION_EVENT_MASK;
		}

		if ((mask & AWTEvent.KEY_EVENT_MASK) != 0 && (tmpMask & AWTEvent.KEY_EVENT_MASK) == 0 && getListeners(KeyListener.class).length > 0) {
			tmpMask |= AWTEvent.KEY_EVENT_MASK;
		}

		return (tmpMask & mask) == mask;
	}

	@Override public float getAlignmentX() {
		if (!mAlignmentXSet) {
			LayoutManager mgr = getLayout();

			if (mgr instanceof LayoutManager2) {
				synchronized (getTreeLock()) {
					return ((LayoutManager2) mgr).getLayoutAlignmentX(this);
				}
			}
		}
		return mAlignmentX;
	}

	@Override public float getAlignmentY() {
		if (!mAlignmentYSet) {
			LayoutManager mgr = getLayout();

			if (mgr instanceof LayoutManager2) {
				synchronized (getTreeLock()) {
					return ((LayoutManager2) mgr).getLayoutAlignmentY(this);
				}
			}
		}
		return mAlignmentY;
	}

	/** @param alignment The panel x-alignment. */
	public void setAlignmentX(float alignment) {
		mAlignmentX = alignment;
		mAlignmentXSet = true;
	}

	/** @param alignment The panel y-alignment. */
	public void setAlignmentY(float alignment) {
		mAlignmentY = alignment;
		mAlignmentYSet = true;
	}

	/**
	 * @param type The class type look for.
	 * @return The first ancestor of this panel that has the specified type.
	 */
	public Container getAncestorOfType(Class<?> type) {
		return getAncestorOfType(type, false);
	}

	/**
	 * @param type The class type look for.
	 * @param includeSelf Pass in <code>true</code> if the panel itself should be considered in
	 *            the search.
	 * @return The first ancestor of this panel that has the specified type.
	 */
	public Container getAncestorOfType(Class<?> type, boolean includeSelf) {
		Container parent = includeSelf ? this : getParent();

		while (parent != null && !type.isInstance(parent)) {
			parent = parent.getParent();
		}
		return parent;
	}

	/**
	 * @return The border of this component or <code>null</code> if no border is currently set.
	 */
	public TKBorder getBorder() {
		return mBorder;
	}

	/**
	 * @param bounds The bounds to check.
	 * @param above Only children above this one are considered.
	 * @return An array of rectangles which are not obscured by children above the specified child.
	 */
	public Rectangle[] getNonObscuredBounds(Rectangle bounds, Component above) {
		int count = getComponentCount();
		Rectangle[] remains = new Rectangle[] { bounds };
		ArrayList<Rectangle> list = new ArrayList<Rectangle>(32);
		Rectangle orig = new Rectangle(bounds);

		for (int i = 0; i < count; i++) {
			Component component = getComponent(i);
			int j;
			int k;
			Rectangle[] smallerRemains;

			if (above == component) {
				break;
			}

			if (component.isVisible()) {
				bounds = TKRectUtils.intersection(orig, component.getBounds());
				if (bounds.width > 0 && bounds.height > 0) {
					if (component.isOpaque()) {
						for (j = 0; j < remains.length; j++) {
							smallerRemains = TKRectUtils.computeDifference(remains[j], bounds);
							for (k = 0; k < smallerRemains.length; k++) {
								list.add(smallerRemains[k]);
							}
						}

						if (list.size() == 0) {
							return new Rectangle[0];
						}

						remains = list.toArray(new Rectangle[0]);
						list.clear();
					} else if (component instanceof TKPanel) {
						TKPanel comp = (TKPanel) component;
						Rectangle[] blockers = comp.getObscuredBounds();

						for (k = 0; k < blockers.length; k++) {
							blockers[k].x += comp.getX();
							blockers[k].y += comp.getY();
							for (j = 0; j < remains.length; j++) {
								smallerRemains = TKRectUtils.computeDifference(remains[j], blockers[k]);

								for (Rectangle element : smallerRemains) {
									list.add(element);
								}
							}

							if (list.size() == 0) {
								return new Rectangle[0];
							}

							remains = list.toArray(new Rectangle[0]);
							list.clear();
						}
					}
				}
			}
		}
		return remains;
	}

	/** @return An array of rectangles which are obscured by descendants. */
	public Rectangle[] getObscuredBounds() {
		int count = getComponentCount();
		ArrayList<Rectangle> list = new ArrayList<Rectangle>(64);

		for (int i = 0; i < count; i++) {
			Component comp = getComponent(i);

			if (comp.isVisible() && comp.getWidth() > 0 && comp.getHeight() > 0) {
				if (comp.isOpaque()) {
					list.add(comp.getBounds());
				} else if (comp instanceof TKPanel) {
					TKPanel cfComp = (TKPanel) comp;
					Rectangle[] obscured = cfComp.getObscuredBounds();

					for (Rectangle element : obscured) {
						element.x += cfComp.getX();
						element.y += cfComp.getY();
						list.add(element);
					}
				}
			}
		}
		return list.toArray(new Rectangle[0]);
	}

	/**
	 * @param bounds The bounds that must intersect.
	 * @return A list of components intersecting the specified bounds.
	 */
	public List<Component> getComponentsIntersecting(Rectangle bounds) {
		return getComponentsIntersecting(bounds, null);
	}

	/**
	 * @param bounds The bounds that must intersect.
	 * @param above Only components above this one are considered.
	 * @return A list of components intersecting the specified bounds which are above the specified
	 *         component.
	 */
	public List<Component> getComponentsIntersecting(Rectangle bounds, Component above) {
		int count = getComponentCount();
		ArrayList<Component> list = new ArrayList<Component>(count);

		for (int i = 0; i < count; i++) {
			Component component = getComponent(i);

			if (above == component) {
				break;
			}
			if (bounds.intersects(component.getBounds())) {
				list.add(component);
			}
		}
		return list;
	}

	@Override public Graphics getGraphics() {
		return TKGraphics.configureGraphics((Graphics2D) super.getGraphics());
	}

	/** @return A new graphics context for this panel. */
	public Graphics2D getGraphics2D() {
		return TKGraphics.configureGraphics((Graphics2D) super.getGraphics());
	}

	@Override public GraphicsConfiguration getGraphicsConfiguration() {
		GraphicsConfiguration gc = super.getGraphicsConfiguration();

		if (gc == null) {
			Graphics2D g2d = TKGraphics.getGraphics();

			gc = g2d.getDeviceConfiguration();
			g2d.dispose();
		}
		return gc;
	}

	/**
	 * @param exclude A child panel to exclude from drawing.
	 * @return An {@link BufferedImage} containing the current contents of this component, minus the
	 *         specified component and its children.
	 */
	public BufferedImage getImage(TKPanel exclude) {
		BufferedImage offscreen = null;

		synchronized (getTreeLock()) {
			Graphics2D g2d = null;

			try {
				Rectangle bounds = getLocalBounds();
				Color saved;

				mExcludeChild = exclude;
				offscreen = getGraphicsConfiguration().createCompatibleImage(bounds.width, bounds.height, Transparency.TRANSLUCENT);
				g2d = (Graphics2D) offscreen.getGraphics();
				saved = g2d.getBackground();
				g2d.setClip(bounds);
				g2d.setBackground(new Color(0, true));
				g2d.clearRect(0, 0, bounds.width, bounds.height);
				g2d.setBackground(saved);
				paint(g2d);
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			} finally {
				if (g2d != null) {
					g2d.dispose();
				}
				mExcludeChild = null;
			}
		}
		return offscreen;
	}

	/**
	 * @param component The component to work on.
	 * @return The index of the specified component. -1 will be returned if the component isn't a
	 *         direct child.
	 */
	public int getIndexOf(Component component) {
		int count = getComponentCount();

		for (int i = 0; i < count; i++) {
			if (component == getComponent(i)) {
				return i;
			}
		}
		return -1;
	}

	@Override public Insets getInsets() {
		if (mBorder != null) {
			return mBorder.getBorderInsets(this);
		}
		return super.getInsets();
	}

	/** @return The bounds of this panel minus its insets. */
	public Rectangle getInsetBounds() {
		Rectangle bounds = getBounds();
		Insets insets = getInsets();

		bounds.x += insets.left;
		bounds.y += insets.top;
		bounds.width -= insets.left + insets.right;
		bounds.height -= insets.top + insets.bottom;
		return bounds;
	}

	/** @return The bounds of this panel in local coordinates. */
	public Rectangle getLocalBounds() {
		return new Rectangle(0, 0, getWidth(), getHeight());
	}

	/** @return The bounds of this panel in local coordinates minus its insets. */
	public Rectangle getLocalInsetBounds() {
		Insets insets = getInsets();

		return new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
	}

	@Override public final Dimension getMaximumSize() {
		mInGetMaximumSize = true;
		try {
			Dimension size;

			if (mMaximumSize == null) {
				if (mCalcedMaximumSize == null) {
					mCalcedMaximumSize = getMaximumSizeSelf();
				}
				size = new Dimension(mCalcedMaximumSize);
			} else {
				size = new Dimension(mMaximumSize);
			}
			return sanitizeSize(size);
		} finally {
			mInGetMaximumSize = false;
		}
	}

	/** @return The maximum size of the panel. */
	protected Dimension getMaximumSizeSelf() {
		synchronized (getTreeLock()) {
			LayoutManager mgr = getLayout();

			return mgr != null && mgr instanceof LayoutManager2 ? ((LayoutManager2) mgr).maximumLayoutSize(this) : new Dimension(MAX_SIZE, MAX_SIZE);
		}
	}

	/**
	 * Sets the maximum size of the panel to a constant value. Subsequent calls to
	 * {@link #getMaximumSize()} will always return this value. Setting the maximum size to
	 * <code>null</code> restores the default behavior.
	 * 
	 * @param maximumSize The new maximum size.
	 */
	@Override public void setMaximumSize(Dimension maximumSize) {
		mMaximumSize = maximumSize != null ? new Dimension(maximumSize) : null;
	}

	@Override public final Dimension getMinimumSize() {
		mInGetMinimumSize = true;
		try {
			Dimension size;

			if (mMinimumSize == null) {
				if (mCalcedMinimumSize == null) {
					mCalcedMinimumSize = getMinimumSizeSelf();
				}
				size = new Dimension(mCalcedMinimumSize);
			} else {
				size = new Dimension(mMinimumSize);
			}
			return sanitizeSize(size);
		} finally {
			mInGetMinimumSize = false;
		}
	}

	/** @return The minimum size of the panel. */
	protected Dimension getMinimumSizeSelf() {
		synchronized (getTreeLock()) {
			LayoutManager mgr = getLayout();

			return mgr != null ? mgr.minimumLayoutSize(this) : new Dimension(0, 0);
		}
	}

	/**
	 * Sets the minimum size of the panel to a constant value. Subsequent calls to
	 * {@link #getMinimumSize()} will always return this value. Setting the minimum size to
	 * <code>null</code> restores the default behavior.
	 * 
	 * @param minimumSize The new minimum size.
	 */
	@Override public void setMinimumSize(Dimension minimumSize) {
		mMinimumSize = minimumSize != null ? new Dimension(minimumSize) : null;
	}

	@Override public final Dimension getPreferredSize() {
		Dimension pSize;
		Dimension mSize;

		if (mPreferredSize == null) {
			if (mCalcedPreferredSize == null) {
				mCalcedPreferredSize = getPreferredSizeSelf();
			}
			pSize = new Dimension(mCalcedPreferredSize);
		} else {
			pSize = new Dimension(mPreferredSize);
		}

		if (!mInGetMaximumSize) {
			mSize = getMaximumSize();
			if (pSize.width > mSize.width) {
				pSize.width = mSize.width;
			}
			if (pSize.height > mSize.height) {
				pSize.height = mSize.height;
			}
		}

		if (!mInGetMinimumSize) {
			mSize = getMinimumSize();
			if (pSize.width < mSize.width) {
				pSize.width = mSize.width;
			}
			if (pSize.height < mSize.height) {
				pSize.height = mSize.height;
			}
		}

		return sanitizeSize(pSize);
	}

	/** @return The preferred size of the panel. */
	protected Dimension getPreferredSizeSelf() {
		synchronized (getTreeLock()) {
			LayoutManager mgr = getLayout();

			return mgr != null ? mgr.preferredLayoutSize(this) : new Dimension(0, 0);
		}
	}

	/**
	 * Sets the preferred size of the panel to a constant value. Subsequent calls to
	 * {@link #getPreferredSize()} will always return this value. Setting the preferred size to
	 * <code>null</code> restores the default behavior.
	 * 
	 * @param preferredSize The new preferred size.
	 */
	@Override public void setPreferredSize(Dimension preferredSize) {
		mPreferredSize = preferredSize != null ? new Dimension(preferredSize) : null;
	}

	/**
	 * @param component The component to work on.
	 * @return The root container of a hierarchy. A root component is defined as one which is either
	 *         a {@link TKBaseWindow} or not a {@link TKPanel}.
	 */
	static public Container getRoot(Component component) {
		Container parent = component instanceof Container ? (Container) component : component.getParent();

		while (parent != null) {
			if (parent instanceof TKBaseWindow || !(parent instanceof TKPanel)) {
				return parent;
			}
			parent = parent.getParent();
		}
		return null;
	}

	/**
	 * @param event The {@link MouseEvent} that caused the tooltip to be shown.
	 * @return The area, in the panel's local coordinates, that should not be obscured by a tooltip.
	 */
	public Rectangle getToolTipProhibitedArea(@SuppressWarnings("unused") MouseEvent event) {
		Rectangle bounds = getLocalInsetBounds();

		bounds.grow(5, 5);
		return bounds;
	}

	/**
	 * If a panel provides a custom tooltip (i.e. a sub-class or custom-configured {@link TKTooltip}),
	 * this method should be overridden.
	 * 
	 * @param event The event that is triggering the tooltip.
	 * @return The tooltip for the specified {@link MouseEvent}.
	 */
	public TKTooltip getToolTip(MouseEvent event) {
		return new TKTooltip(getBaseWindow(), getToolTipText(event));
	}

	/** @return The tooltip string for this component. */
	public String getToolTipText() {
		return mToolTipText;
	}

	/**
	 * By default this calls {@link #getToolTipText()}. If a panel provides more extensive API to
	 * support differing tooltips at different locations, this method should be overridden.
	 * 
	 * @param event The event that is triggering the tooltip.
	 * @return The string to be used as the tooltip for the specified {@link MouseEvent}.
	 */
	public String getToolTipText(@SuppressWarnings("unused") MouseEvent event) {
		return getToolTipText();
	}

	/**
	 * Registers the text to display in a tool tip. The text displays when the cursor lingers over
	 * the component.
	 * 
	 * @param text The string to display; if the text is <code>null</code>, the tool tip is
	 *            turned off for this component.
	 */
	public void setToolTipText(String text) {
		if (mToolTipText == null ? text != null : !mToolTipText.equals(text)) {
			TKBaseWindow baseWindow = getBaseWindow();

			mToolTipText = text;
			if (baseWindow != null) {
				TKUserInputManager mgr = baseWindow.getUserInputManager();

				if (mgr.getCurrentToolTipOwner() == this) {
					MouseEvent trigger = mgr.getToolTipTrigger();

					mgr.hideToolTip();
					mgr.showToolTip(trigger, this);
				}
			}
		}
	}

	/**
	 * @param bounds The bounds to check.
	 * @return <code>true</code> if the entirety of the specified bounds can be seen through its
	 *         containment.
	 */
	public boolean isVisible(Rectangle bounds) {
		if (!getLocalBounds().contains(bounds)) {
			return false;
		}

		Container parent = getParent();
		while (parent != null) {
			Rectangle pBounds = new Rectangle(0, 0, parent.getWidth(), parent.getHeight());

			convertRectangle(pBounds, parent, this);
			if (!pBounds.contains(bounds)) {
				return false;
			}
			parent = parent.getParent();
		}

		return true;
	}

	/**
	 * @return The panel's visible bounds - the intersection of the visible bounds for this panel
	 *         and all of its ancestors.
	 */
	public Rectangle getVisibleBounds() {
		Rectangle visibleRect = getLocalBounds();
		Container parent = getParent();

		while (parent != null) {
			Rectangle bounds = new Rectangle(0, 0, parent.getWidth(), parent.getHeight());

			convertRectangle(bounds, parent, this);
			visibleRect = TKRectUtils.intersection(visibleRect, bounds);
			parent = parent.getParent();
		}

		return visibleRect;
	}

	/** @return The {@link TKBaseWindow} this panel resides in, if any. */
	public TKBaseWindow getBaseWindow() {
		Container parent = getParent();

		while (parent != null) {
			if (parent instanceof TKBaseWindow) {
				return (TKBaseWindow) parent;
			}
			parent = parent.getParent();
		}
		return null;
	}

	/** @return The {@link TKBaseWindow} this panel resides in, if any. */
	public Window getBaseWindowAsWindow() {
		return (Window) getBaseWindow();
	}

	/**
	 * Call to force any existing text field with focus to notify its action listeners.
	 */
	public void forceFocusToAccept() {
		TKBaseWindow window = getBaseWindow();

		if (window != null) {
			window.forceFocusToAccept();
		}
	}

	@Override public void invalidate() {
		synchronized (getTreeLock()) {
			mCalcedMaximumSize = null;
			mCalcedMinimumSize = null;
			mCalcedPreferredSize = null;
			super.invalidate();
		}
	}

	@Override public boolean isShowing() {
		if (isPrinting()) {
			return true;
		}
		return super.isShowing();
	}

	/**
	 * @param bounds The bounds to check.
	 * @param above Only children above this one are considered.
	 * @return <code>true</code> if the specified bounds within the panel is completely obscured
	 *         by children above the specified child.
	 */
	public boolean isCompletelyObscured(Rectangle bounds, Component above) {
		return getNonObscuredBounds(bounds, above).length == 0;
	}

	/**
	 * @return <code>true</code> if default button actions should be processed when this component
	 *         is the keyboard focus, otherwise returns <code>false</code>. By default, this
	 *         method is implemented to return <code>true</code>.
	 */
	public boolean isDefaultButtonProcessingAllowed() {
		return true;
	}

	/** @return <code>true</code> if a layer change is occurring. */
	public boolean isInLayerChange() {
		boolean inChange = mInLayerChange;

		if (!inChange) {
			Container parent = getParent();

			if (parent != null && parent instanceof TKPanel) {
				inChange = ((TKPanel) parent).isInLayerChange();
			}
		}
		return inChange;
	}

	/** @param in Whether a layer change is occurring. */
	public void setInLayerChange(boolean in) {
		mInLayerChange = in;
	}

	/** @return <code>true</code> if the maximum size has been set to a fixed value. */
	@Override public boolean isMaximumSizeSet() {
		return mMaximumSize != null;
	}

	/** @return <code>true</code> if the minimum size has been set to a fixed value. */
	@Override public boolean isMinimumSizeSet() {
		return mMinimumSize != null;
	}

	/** @return <code>true</code> if the preferred size has been set to a fixed value. */
	@Override public boolean isPreferredSizeSet() {
		return mPreferredSize != null;
	}

	@Override public boolean isOpaque() {
		return mOpaque;
	}

	/** @return Whether this panel should be flagged as having no tooltip. */
	public boolean flagForNoToolTip() {
		return false;
	}

	/** @return <code>true</code> if the panel is in a window that is being printed. */
	public boolean isPrinting() {
		TKBaseWindow window = getBaseWindow();

		if (window != null) {
			return window.isPrinting();
		}
		return TKGraphics.inHeadlessPrintMode();
	}

	@Override public void paint(Graphics graphics) {
		paint((Graphics2D) graphics, null);
	}

	/**
	 * This method is draws the panel. Applications should not invoke paint directly, but should
	 * instead use one of the <code>repaint<code> methods
	 * to schedule the panel for redrawing.
	 *
	 * @param g2d	The graphics context.
	 * @param clips	The clipping bounds.
	 */
	public void paint(Graphics2D g2d, Rectangle[] clips) {
		synchronized (getTreeLock()) {
			Rectangle bounds = getLocalBounds();

			if (bounds.width > 0 && bounds.height > 0) {
				Rectangle clip;

				if (clips == null) {
					Rectangle realClip = g2d.getClipBounds();

					clips = new Rectangle[] { realClip == null ? getVisibleBounds() : realClip };
				}
				clip = TKRectUtils.unionIntersection(clips, bounds);
				if (!clip.isEmpty()) {
					Graphics2D tg2d = null;

					TKGraphics.configureGraphics(g2d);
					g2d.setClip(clip);
					g2d.setColor(getForeground());
					g2d.setFont(getFont());

					try {
						tg2d = TKGraphics.configureGraphics((Graphics2D) g2d.create());
						paintMinimally(tg2d, clips);
						tg2d.dispose();

						tg2d = TKGraphics.configureGraphics((Graphics2D) g2d.create());
						paintChildren(tg2d, clips);
						tg2d.dispose();

						tg2d = TKGraphics.configureGraphics((Graphics2D) g2d.create());
						paintBorder(tg2d);
						if (TKGraphics.isShowComponentsWithNoToolTipsOn() && flagForNoToolTip()) {
							tg2d.setColor(new Color(255, 0, 0, 50));
							for (Rectangle one : clips) {
								tg2d.fill(one);
							}
						}
					} catch (Exception exception) {
						assert false : TKDebug.throwableToString(exception);
					} finally {
						if (tg2d != null) {
							tg2d.dispose();
						}
					}
				}
			}
		}
	}

	/**
	 * Paints the panel's border.
	 * 
	 * @param g2d The graphics context.
	 */
	protected void paintBorder(Graphics2D g2d) {
		if (mBorder != null) {
			mBorder.paintBorder(this, g2d, 0, 0, getWidth(), getHeight());
		}
	}

	/**
	 * Paints this panel's children.
	 * 
	 * @param g2d The graphics context.
	 * @param clips The clipping bounds.
	 */
	protected void paintChildren(Graphics2D g2d, Rectangle[] clips) {
		for (int i = getComponentCount() - 1; i >= 0; i--) {
			TKPanel comp = (TKPanel) getComponent(i);

			if (comp.isVisible() && comp != mExcludeChild) {
				Rectangle bounds = comp.getBounds();

				if (TKRectUtils.intersects(clips, bounds)) {
					Rectangle[] newClips = TKRectUtils.intersection(clips, bounds);
					Rectangle compClip = TKRectUtils.unionIntersection(newClips, bounds);
					Graphics2D cg2d = TKGraphics.configureGraphics((Graphics2D) g2d.create(compClip.x, compClip.y, compClip.width, compClip.height));

					cg2d.translate(bounds.x - compClip.x, bounds.y - compClip.y);
					compClip.x -= bounds.x;
					compClip.y -= bounds.y;
					cg2d.setClip(compClip);

					TKRectUtils.offsetRectangles(newClips, -bounds.x, -bounds.y);

					try {
						comp.paint(cg2d, newClips);
					} catch (Exception exception) {
						assert false : TKDebug.throwableToString(exception);
					} finally {
						cg2d.dispose();
					}
				}
			}
		}
	}

	/**
	 * Paints the minimal portions of the panel that are necessary.
	 * 
	 * @param g2d The graphics context.
	 * @param clips The clipping bounds.
	 */
	protected void paintMinimally(Graphics2D g2d, Rectangle[] clips) {
		Rectangle bounds = getLocalBounds();

		clips = TKRectUtils.removeFrom(clips, getObscuredBounds());
		if (TKRectUtils.intersects(clips, bounds)) {
			Rectangle compClip = TKRectUtils.unionIntersection(clips, bounds);
			Graphics2D cg2d = TKGraphics.configureGraphics((Graphics2D) g2d.create(compClip.x, compClip.y, compClip.width, compClip.height));

			cg2d.translate(bounds.x - compClip.x, bounds.y - compClip.y);
			compClip.x -= bounds.x;
			compClip.y -= bounds.y;

			try {
				if (isOpaque()) {
					for (Rectangle element : clips) {
						if (element.intersects(compClip)) {
							Rectangle clip = TKRectUtils.intersection(element, compClip);

							cg2d.setClip(clip);
							cg2d.setColor(getBackground());
							cg2d.fill(clip);
						}
					}
				}

				cg2d.setColor(getForeground());
				cg2d.setFont(getFont());
				cg2d.setClip(compClip);

				paintPanel(cg2d, clips);
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			} finally {
				cg2d.dispose();
			}
		}
	}

	/**
	 * Paints the panel. By default, nothing is done.
	 * 
	 * @param g2d The graphics context.
	 * @param clips The clipping bounds.
	 */
	protected void paintPanel(@SuppressWarnings("unused") Graphics2D g2d, @SuppressWarnings("unused") Rectangle[] clips) {
		// Nothing to do...
	}

	/**
	 * Immediately paints the panel and all of its descendants that overlap the region without
	 * waiting for the normal event loop repaint management.
	 */
	public void paintImmediately() {
		paintImmediately(getLocalBounds());
	}

	/**
	 * Immediately paints the specified region in this panel and all of its descendants that overlap
	 * the region without waiting for the normal event loop repaint management.
	 * 
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @param width The width.
	 * @param height The height.
	 */
	public void paintImmediately(int x, int y, int width, int height) {
		paintImmediately(new Rectangle(x, y, width, height));
	}

	/**
	 * Immediately paints the specified region in this panel and all of its descendants that overlap
	 * the region without waiting for the normal event loop repaint management.
	 * 
	 * @param bounds The bounds.
	 */
	public void paintImmediately(Rectangle bounds) {
		paintImmediately(new Rectangle[] { bounds });
	}

	/**
	 * Immediately paints the specified region in this panel and all of its descendants that overlap
	 * the region without waiting for the normal event loop repaint management.
	 * 
	 * @param bounds The bounds.
	 */
	public void paintImmediately(Rectangle[] bounds) {
		Container parent = getParent();

		if (parent != null) {
			Rectangle[] compBounds = new Rectangle[bounds.length];

			for (int i = 0; i < bounds.length; i++) {
				compBounds[i] = new Rectangle(bounds[i]);
			}
			TKRectUtils.offsetRectangles(compBounds, getX(), getY());
			compBounds = TKRectUtils.intersection(compBounds, getBounds());
			if (compBounds.length != 0) {
				if (parent instanceof TKPanel) {
					((TKPanel) parent).paintImmediately(compBounds);
				} else if (parent instanceof TKBaseWindow) {
					((TKBaseWindow) parent).getRepaintManager().paintImmediately(compBounds);
				}
			}
		}
	}

	@Override public final void processMouseEvent(MouseEvent event) {
		TKBaseWindow window = getBaseWindow();

		if (window != null) {
			window.getUserInputManager().processMouseEvent(event);
		}
	}

	/**
	 * Processes any mouse events that the panel itself recognizes. This is called after any
	 * interested listeners have been given a chance to steal away the event. This method is called
	 * only if the event has not yet been consumed.
	 * 
	 * @param event The mouse event to process.
	 */
	public void processMouseEventSelf(@SuppressWarnings("unused") MouseEvent event) {
		// Nothing to do...
	}

	@Override public final void processMouseMotionEvent(MouseEvent event) {
		processMouseEvent(event);
	}

	@Override public final void processMouseWheelEvent(MouseWheelEvent event) {
		processMouseEvent(event);
	}

	/** Removes this panel from its parent, if any. */
	public void removeFromParent() {
		Container parent = getParent();

		if (parent != null) {
			parent.remove(this);
		}
	}

	@Override public void removeNotify() {
		if (!isInLayerChange()) {
			TKBaseWindow window = getBaseWindow();

			if (window != null) {
				TKUserInputManager uiMgr = window.getUserInputManager();

				if (uiMgr.getCurrentToolTipOwner() == this) {
					uiMgr.hideToolTip();
				}
			}
		}
		try {
			Component comp = getParent();

			if (comp != null) {
				Rectangle bounds = getBounds();

				comp.repaint(bounds.x, bounds.y, bounds.width, bounds.height);
			}
			super.removeNotify();
		} catch (Exception exception) {
			// Hide the exception that might happen if something was altered by
			// the extra code above.
		}
	}

	@Override public void repaint(long maxDelay, int x, int y, int width, int height) {
		repaint(new Rectangle(x, y, width, height));
	}

	/**
	 * Repaint the specified area.
	 * 
	 * @param bounds The area to repaint.
	 */
	public void repaint(Rectangle bounds) {
		if (bounds != null) {
			Container parent = getParent();

			if (parent != null) {
				bounds = TKRectUtils.intersection(new Rectangle(bounds.x + getX(), bounds.y + getY(), bounds.width, bounds.height), getBounds());
				if (!bounds.isEmpty()) {
					if (parent instanceof TKPanel) {
						((TKPanel) parent).repaint(bounds);
					} else if (parent instanceof TKBaseWindow) {
						((TKBaseWindow) parent).getRepaintManager().repaint(bounds);
					} else {
						parent.repaint(bounds.x, bounds.y, bounds.width, bounds.height);
					}
				}
			}
		}
	}

	/**
	 * Repaint the specified area.
	 * 
	 * @param bounds The area to repaint.
	 */
	public void repaint(Rectangle[] bounds) {
		for (Rectangle element : bounds) {
			repaint(element);
		}
	}

	/** @return <code>true</code> if this is a validation root. */
	public boolean isValidateRoot() {
		return false;
	}

	/**
	 * Revalidate the component. Calls {@link #invalidate()} followed by
	 * {@link Container#validate()}.
	 */
	public void revalidate() {
		TKBaseWindow window = getBaseWindow();

		invalidate();
		if (window != null) {
			Container parent = getParent();
			Component comp = this;

			while (parent != null) {
				comp = parent;
				if (comp instanceof Dialog || comp instanceof Frame || comp instanceof TKPanel && ((TKPanel) comp).isValidateRoot()) {
					break;
				}
				parent = parent.getParent();
			}
			window.getRepaintManager().addInvalidComponent(comp);
		}
	}

	/**
	 * Revalidates and repaints the component immediately.
	 */
	public void revalidateImmediately() {
		TKBaseWindow window = getBaseWindow();

		revalidate();
		if (window != null) {
			window.getRepaintManager().validate();
		}
		repaint();
		updateImmediately();
	}

	/**
	 * Runs the mouse listeners for this panel.
	 * 
	 * @param event The mouse event.
	 */
	public final void runMouseListeners(MouseEvent event) {
		int id = event.getID();

		if (id == MouseEvent.MOUSE_MOVED || id == MouseEvent.MOUSE_DRAGGED) {
			super.processMouseMotionEvent(event);
		} else if (id == MouseEvent.MOUSE_WHEEL) {
			super.processMouseWheelEvent((MouseWheelEvent) event);
		} else {
			super.processMouseEvent(event);
		}
	}

	/**
	 * Ensures the size is within reasonable parameters.
	 * 
	 * @param size The size to check.
	 * @return The passed-in {@link Dimension} object, for convenience.
	 */
	public static Dimension sanitizeSize(Dimension size) {
		if (size.width < 0) {
			size.width = 0;
		} else if (size.width > MAX_SIZE) {
			size.width = MAX_SIZE;
		}
		if (size.height < 0) {
			size.height = 0;
		} else if (size.height > MAX_SIZE) {
			size.height = MAX_SIZE;
		}
		return size;
	}

	/** @param panels The panels to set to the same size. */
	public static void adjustToSameSize(TKPanel[] panels) {
		Dimension best = new Dimension();

		for (TKPanel panel : panels) {
			Dimension size = panel.getPreferredSize();

			if (size.width > best.width) {
				best.width = size.width;
			}
			if (size.height > best.height) {
				best.height = size.height;
			}
		}
		for (TKPanel panel : panels) {
			panel.setOnlySize(best);
		}
	}

	/**
	 * Forwards the {@link #scrollRectIntoView(Rectangle)} call to the panel's parent. Panels that
	 * can service the request override this method and perform the scrolling.
	 * 
	 * @param bounds The area to scroll into view.
	 */
	public void scrollRectIntoView(Rectangle bounds) {
		Container parent = getParent();

		if (parent != null && parent instanceof TKPanel) {
			bounds.x += getX();
			bounds.y += getY();
			((TKPanel) parent).scrollRectIntoView(bounds);
			bounds.x -= getX();
			bounds.y -= getY();
		}
	}

	@Override public void setBackground(Color bg) {
		Color oldBg = getBackground();

		super.setBackground(bg);
		if (oldBg != null && !oldBg.equals(bg) || bg != null && !bg.equals(oldBg)) {
			repaint();
		}
	}

	/**
	 * Sets the border of this panel. The {@link TKBorder} object is responsible for defining the
	 * insets for the component (overriding any insets set directly on the panel) and for rendering
	 * any border decorations within the bounds of those insets. Borders should be used (rather than
	 * insets) for creating both decorative and non-decorative (such as margins and padding) regions
	 * for a panel.
	 * 
	 * @param border The border to set.
	 */
	public void setBorder(TKBorder border) {
		setBorder(border, true, true);
	}

	/**
	 * Sets the border of this panel. The {@link TKBorder} object is responsible for defining the
	 * insets for the component (overriding any insets set directly on the panel) and for rendering
	 * any border decorations within the bounds of those insets. Borders should be used (rather than
	 * insets) for creating both decorative and non-decorative (such as margins and padding) regions
	 * for a panel.
	 * 
	 * @param border The border to set.
	 * @param doRepaint Whether or not to repaint.
	 * @param doLayout Whether or not to re-layout.
	 */
	public void setBorder(TKBorder border, boolean doRepaint, boolean doLayout) {
		if (border != mBorder) {
			TKBorder oldBorder = mBorder;

			mBorder = border;
			if (doLayout) {
				if (border == null || oldBorder == null || !border.getBorderInsets(this).equals(oldBorder.getBorderInsets(this))) {
					revalidate();
				}
			}

			if (doRepaint) {
				repaint();
			}
		}
	}

	@Override public void setBounds(int x, int y, int width, int height) {
		synchronized (getTreeLock()) {
			int oldWidth = getWidth();
			int oldHeight = getHeight();

			if (oldWidth != width || oldHeight != height || getX() != x || getY() != y) {
				int oldX = getX();
				int oldY = getY();
				boolean resized;
				boolean moved;

				super.setBounds(x, y, width, height);
				moved = oldX != getX() || oldY != getY();
				resized = oldWidth != getWidth() || oldHeight != getHeight();

				if ((moved || resized) && mReshapeListeners != null) {
					for (TKReshapeListener listener : mReshapeListeners) {
						listener.panelReshaped(this, moved, resized);
					}
				}
				if (resized) {
					validate();
				}
			}
		}
	}

	@Override public void setEnabled(boolean enabled) {
		if (enabled != isEnabled()) {
			super.setEnabled(enabled);
			repaint();
		}
	}

	@Override public Font getFont() {
		if (mFontKey != null) {
			return TKFont.lookup(mFontKey);
		}
		return super.getFont();
	}

	/** @return The font key. */
	public String getFontKey() {
		if (mFontKey == null) {
			mFontKey = TKFont.TEXT_FONT_KEY;
		}
		return mFontKey;
	}

	@Deprecated @Override public void setFont(Font font) {
		Font oldFont = getFont();

		mFontKey = null;
		super.setFont(font);
		if (font != oldFont) {
			revalidate();
		}
	}

	/** @param fontKey The font key to use for dynamic fonts. */
	public void setFontKey(String fontKey) {
		if (fontKey == null ? mFontKey != null : !fontKey.equals(mFontKey)) {
			mFontKey = fontKey;
			revalidate();
		}
	}

	@Override public void setForeground(Color fg) {
		Color oldFg = getForeground();

		super.setForeground(fg);
		if (oldFg != null && !oldFg.equals(fg) || fg != null && !fg.equals(oldFg)) {
			repaint();
		}
	}

	/**
	 * Call to set minimum, maximum and preferred size all to the same value.
	 * 
	 * @param width The width to set.
	 * @param height The height to set.
	 */
	public void setOnlySize(int width, int height) {
		setOnlySize(new Dimension(width, height));
	}

	/**
	 * Call to set minimum, maximum and preferred size all to the same value.
	 * 
	 * @param size The size to set.
	 */
	public void setOnlySize(Dimension size) {
		setMinimumSize(size);
		setMaximumSize(size);
		setPreferredSize(size);
	}

	/** @param isOpaque Whether the panel is opaque or not. */
	public void setOpaque(boolean isOpaque) {
		boolean oldValue = mOpaque;

		mOpaque = isOpaque;
		if (oldValue != mOpaque) {
			repaint();
		}
	}

	/**
	 * Marks the specified region in this panel and all of its descendants that overlap the region
	 * as not needing to be painted.
	 * 
	 * @param shape The area to mask.
	 */
	public void stopRepaint(Shape shape) {
		Container parent = getParent();

		if (parent != null) {
			Rectangle area = shape.getBounds();

			area.x += getX();
			area.y += getY();
			area = TKRectUtils.intersection(area, getBounds());
			if (!area.isEmpty()) {
				if (parent instanceof TKPanel) {
					((TKPanel) parent).stopRepaint(area);
				} else if (parent instanceof TKBaseWindow) {
					((TKBaseWindow) parent).getRepaintManager().stopRepaint(area);
				}
			}
		}
	}

	@Override public void update(Graphics graphics) {
		paint(graphics);
	}

	/** Updates the window that this panel is in immediately. */
	public void updateImmediately() {
		Container parent = getParent();

		if (parent != null) {
			if (parent instanceof TKPanel) {
				((TKPanel) parent).updateImmediately();
			} else if (parent instanceof TKBaseWindow) {
				((TKBaseWindow) parent).getRepaintManager().updateImmediately();
			}
		}
	}

	/** @return A child panel to ignore during repaints. */
	protected TKPanel getExcludeChild() {
		return mExcludeChild;
	}

	/** @param excludeChild A child panel to ignore during repaints. */
	protected void setExcludeChild(TKPanel excludeChild) {
		mExcludeChild = excludeChild;
	}

	/**
	 * @return The default state this panel should be in if it resides in a toolbar.
	 */
	public boolean getDefaultToolbarEnabledState() {
		return false;
	}

	/**
	 * Looks for a panel with the specified name.
	 * 
	 * @param name The name to look for.
	 * @return The panel with that name, or <code>null</code>.
	 */
	public TKPanel lookupPanelByName(String name) {
		if (name.equals(getName())) {
			return this;
		}
		for (Component comp : getComponents()) {
			TKPanel panel = ((TKPanel) comp).lookupPanelByName(name);

			if (panel != null) {
				return panel;
			}
		}
		return null;
	}

	/** @return Whether an undo is currently being applied. */
	public boolean isUndoBeingApplied() {
		TKBaseWindow window = getBaseWindow();

		if (window != null) {
			TKUndoManager undoMgr = window.getUndoManager();

			if (undoMgr != null) {
				return undoMgr.isUndoBeingApplied();
			}
		}
		return false;
	}

	@Override public String toString() {
		String name = getName();

		return name != null ? name : super.toString();
	}
}
