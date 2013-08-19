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

package com.trollworks.gcs.widgets;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Window;
import java.awt.event.MouseEvent;
import java.awt.event.MouseWheelEvent;

import javax.swing.AbstractButton;
import javax.swing.JComboBox;
import javax.swing.JTabbedPane;
import javax.swing.JTextField;
import javax.swing.JViewport;
import javax.swing.RepaintManager;

/** Various utility methods for the UI. */
public class UIUtilities {
	/**
	 * Selects the tab with the specified title.
	 * 
	 * @param pane The {@link JTabbedPane} to use.
	 * @param title The title to select.
	 */
	public static void selectTab(JTabbedPane pane, String title) {
		int count = pane.getTabCount();
		for (int i = 0; i < count; i++) {
			if (pane.getTitleAt(i).equals(title)) {
				pane.setSelectedIndex(i);
				break;
			}
		}
	}

	/**
	 * Disables all controls in the specified component and all its children.
	 * 
	 * @param comp The {@link Component} to work on.
	 */
	public static void disableControls(Component comp) {
		if (comp instanceof Container) {
			Container container = (Container) comp;
			int count = container.getComponentCount();

			for (int i = 0; i < count; i++) {
				disableControls(container.getComponent(i));
			}
		}

		if (comp instanceof AbstractButton || comp instanceof JComboBox || comp instanceof JTextField) {
			comp.setEnabled(false);
		}
	}

	/**
	 * Sets a {@link Component}'s min, max & preferred sizes to a specific size.
	 * 
	 * @param comp The {@link Component} to work on.
	 * @param size The size to set the component to.
	 */
	public static void setOnlySize(Component comp, Dimension size) {
		comp.setMinimumSize(size);
		comp.setMaximumSize(size);
		comp.setPreferredSize(size);
	}

	/** @param comps The {@link Component}s to set to the same size. */
	public static void adjustToSameSize(Component... comps) {
		Dimension best = new Dimension();
		for (Component comp : comps) {
			Dimension size = comp.getPreferredSize();
			if (size.width > best.width) {
				best.width = size.width;
			}
			if (size.height > best.height) {
				best.height = size.height;
			}
		}
		for (Component comp : comps) {
			setOnlySize(comp, best);
		}
	}

	/**
	 * Converts a {@link Point} from one component's coordinate system to another's.
	 * 
	 * @param pt The point to convert.
	 * @param from The component the point originated in.
	 * @param to The component the point should be translated to.
	 */
	public static void convertPoint(Point pt, Component from, Component to) {
		convertPointToScreen(pt, from);
		convertPointFromScreen(pt, to);
	}

	/**
	 * Converts a {@link Point} from on the screen to a position within the component.
	 * 
	 * @param pt The point to convert.
	 * @param component The component the point should be translated to.
	 */
	public static void convertPointFromScreen(Point pt, Component component) {
		do {
			pt.x -= component.getX();
			pt.y -= component.getY();
			if (component instanceof Window) {
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
	public static void convertPointToScreen(Point pt, Component component) {
		do {
			pt.x += component.getX();
			pt.y += component.getY();
			if (component instanceof Window) {
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
	public static void convertRectangle(Rectangle bounds, Component from, Component to) {
		convertRectangleToScreen(bounds, from);
		convertRectangleFromScreen(bounds, to);
	}

	/**
	 * Converts a {@link Rectangle} from on the screen to a position within the component.
	 * 
	 * @param bounds The rectangle to convert.
	 * @param component The component the rectangle should be translated to.
	 */
	public static void convertRectangleFromScreen(Rectangle bounds, Component component) {
		do {
			bounds.x -= component.getX();
			bounds.y -= component.getY();
			if (component instanceof Window) {
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
	public static void convertRectangleToScreen(Rectangle bounds, Component component) {
		do {
			bounds.x += component.getX();
			bounds.y += component.getY();
			if (component instanceof Window) {
				break;
			}
			component = component.getParent();
		} while (component != null);
	}

	/**
	 * @param parent The parent {@link Container}.
	 * @param child The child {@link Component}.
	 * @return The index of the specified {@link Component}. -1 will be returned if the
	 *         {@link Component} isn't a direct child.
	 */
	public static int getIndexOf(Container parent, Component child) {
		if (parent != null) {
			int count = parent.getComponentCount();
			for (int i = 0; i < count; i++) {
				if (child == parent.getComponent(i)) {
					return i;
				}
			}
		}
		return -1;
	}

	/**
	 * Clones a {@link MouseEvent}.
	 * 
	 * @param event The event to clone.
	 * @return The new {@link MouseEvent}.
	 */
	public static final MouseEvent cloneMouseEvent(MouseEvent event) {
		if (event instanceof MouseWheelEvent) {
			MouseWheelEvent old = (MouseWheelEvent) event;

			return new MouseWheelEvent((Component) old.getSource(), old.getID(), System.currentTimeMillis(), old.getModifiers(), old.getX(), old.getY(), old.getClickCount(), old.isPopupTrigger(), old.getScrollType(), old.getScrollAmount(), old.getWheelRotation());
		}
		return new MouseEvent((Component) event.getSource(), event.getID(), System.currentTimeMillis(), event.getModifiers(), event.getX(), event.getY(), event.getClickCount(), event.isPopupTrigger());
	}

	/**
	 * Clones a {@link MouseEvent}.
	 * 
	 * @param event The event to clone.
	 * @param refreshTime Pass in <code>true</code> to generate a new time stamp.
	 * @return The new {@link MouseEvent}.
	 */
	public static final MouseEvent cloneMouseEvent(MouseEvent event, boolean refreshTime) {
		if (event instanceof MouseWheelEvent) {
			MouseWheelEvent old = (MouseWheelEvent) event;

			return new MouseWheelEvent((Component) old.getSource(), old.getID(), refreshTime ? System.currentTimeMillis() : event.getWhen(), old.getModifiers(), old.getX(), old.getY(), old.getClickCount(), old.isPopupTrigger(), old.getScrollType(), old.getScrollAmount(), old.getWheelRotation());
		}
		return new MouseEvent((Component) event.getSource(), event.getID(), refreshTime ? System.currentTimeMillis() : event.getWhen(), event.getModifiers(), event.getX(), event.getY(), event.getClickCount(), event.isPopupTrigger());
	}

	/**
	 * Clones a {@link MouseEvent}.
	 * 
	 * @param event The event to clone.
	 * @param source Pass in a new source.
	 * @param where Pass in a new location.
	 * @param refreshTime Pass in <code>true</code> to generate a new time stamp.
	 * @return The new {@link MouseEvent}.
	 */
	public static final MouseEvent cloneMouseEvent(MouseEvent event, Component source, Point where, boolean refreshTime) {
		if (event instanceof MouseWheelEvent) {
			MouseWheelEvent old = (MouseWheelEvent) event;

			return new MouseWheelEvent(source, old.getID(), refreshTime ? System.currentTimeMillis() : event.getWhen(), old.getModifiers(), where.x, where.y, old.getClickCount(), old.isPopupTrigger(), old.getScrollType(), old.getScrollAmount(), old.getWheelRotation());
		}
		return new MouseEvent(source, event.getID(), refreshTime ? System.currentTimeMillis() : event.getWhen(), event.getModifiers(), where.x, where.y, event.getClickCount(), event.isPopupTrigger());
	}

	/**
	 * Forwards a {@link MouseEvent} from one component to another.
	 * 
	 * @param event The event to forward.
	 * @param from The component that originally received the event.
	 * @param to The component the event should be forwarded to.
	 */
	static public void forwardMouseEvent(MouseEvent event, Component from, Component to) {
		translateMouseEvent(event, from, to);
		to.dispatchEvent(event);
	}

	/**
	 * Translates a {@link MouseEvent} from one component to another.
	 * 
	 * @param event The event that will be forwarded.
	 * @param from The component that originally received the event.
	 * @param to The component the event should be forwarded to.
	 */
	static public void translateMouseEvent(MouseEvent event, Component from, Component to) {
		Point evtPt = event.getPoint();
		Point pt = new Point(evtPt);

		UIUtilities.convertPoint(pt, from, to);
		event.setSource(to);
		event.translatePoint(pt.x - evtPt.x, pt.y - evtPt.y);
	}

	/**
	 * @param comp The component to work with.
	 * @return Whether the component should be expanded to fit.
	 */
	public static boolean shouldTrackViewportWidth(Component comp) {
		Container parent = comp.getParent();
		if (parent instanceof JViewport) {
			Dimension available = parent.getSize();
			Dimension prefSize = comp.getPreferredSize();
			return prefSize.width < available.width;
		}
		return false;
	}

	/**
	 * @param comp The component to work with.
	 * @return Whether the component should be expanded to fit.
	 */
	public static boolean shouldTrackViewportHeight(Component comp) {
		Container parent = comp.getParent();
		if (parent instanceof JViewport) {
			Dimension available = parent.getSize();
			Dimension prefSize = comp.getPreferredSize();
			return prefSize.height < available.height;
		}
		return false;
	}

	/** @param comp The component to revalidate. */
	public static void revalidateImmediately(Component comp) {
		if (comp != null) {
			RepaintManager mgr = RepaintManager.currentManager(comp);
			mgr.validateInvalidComponents();
			mgr.paintDirtyRegions();
		}
	}

	/**
	 * @param component The component whose ancestor chain is to be looked at.
	 * @param type The type of ancestor being looked for.
	 * @return The ancestor, or <code>null</code>.
	 */
	public static Container getAncestorOfType(Component component, Class<? extends Container> type) {
		Container parent = component.getParent();
		while (parent != null && !type.isAssignableFrom(parent.getClass())) {
			parent = parent.getParent();
		}
		return parent;
	}
}
