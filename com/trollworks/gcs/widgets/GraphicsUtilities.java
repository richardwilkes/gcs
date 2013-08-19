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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.utility.Geometry;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.GraphicsDevice;
import java.awt.GraphicsEnvironment;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.RenderingHints;
import java.awt.Toolkit;
import java.awt.Window;
import java.awt.image.BufferedImage;
import java.io.File;
import java.lang.reflect.Method;
import java.net.URL;
import java.net.URLClassLoader;

/** Provides general graphics settings and manipulation. */
public class GraphicsUtilities {
	private static String			MSG_HEADLESS;
	private static Frame			HIDDEN_FRAME						= null;
	private static BufferedImage	HIDDEN_FRAME_ICON					= null;
	private static boolean			HEADLESS_PRINT_MODE					= false;
	private static int				HEADLESS_CHECK_RESULT				= 0;
	private static Object			MAC_NSAPPLICATION					= null;
	private static Method			MAC_ACTIVATE_IGNORING_OTHER_APPS	= null;
	private static boolean			OK_TO_USE_FULLSCREEN_TRICK			= true;

	static {
		LocalizedMessages.initialize(GraphicsUtilities.class);
	}

	/** @return Whether the headless print mode is enabled. */
	public static boolean inHeadlessPrintMode() {
		return HEADLESS_PRINT_MODE;
	}

	/** @param inHeadlessPrintMode Whether the headless print mode is enabled. */
	public static void setHeadlessPrintMode(boolean inHeadlessPrintMode) {
		HEADLESS_PRINT_MODE = inHeadlessPrintMode;
	}

	/**
	 * Looks for the screen device that contains the largest part of the specified window.
	 * 
	 * @param window The window to determine the preferred screen device for.
	 * @return The preferred screen device.
	 */
	public static GraphicsDevice getPreferredScreenDevice(Window window) {
		return getPreferredScreenDevice(window.getBounds());
	}

	/**
	 * Looks for the screen device that contains the largest part of the specified global bounds.
	 * 
	 * @param bounds The global bounds to determine the preferred screen device for.
	 * @return The preferred screen device.
	 */
	public static GraphicsDevice getPreferredScreenDevice(Rectangle bounds) {
		GraphicsEnvironment ge = GraphicsEnvironment.getLocalGraphicsEnvironment();
		GraphicsDevice best = ge.getDefaultScreenDevice();
		Rectangle overlapBounds = Geometry.intersection(bounds, best.getDefaultConfiguration().getBounds());
		int bestOverlap = overlapBounds.width * overlapBounds.height;

		for (GraphicsDevice gd : ge.getScreenDevices()) {
			if (gd.getType() == GraphicsDevice.TYPE_RASTER_SCREEN) {
				overlapBounds = Geometry.intersection(bounds, gd.getDefaultConfiguration().getBounds());
				if (overlapBounds.width * overlapBounds.height > bestOverlap) {
					best = gd;
				}
			}
		}
		return best;
	}

	/** @return The maximum bounds that fits on the main screen. */
	public static Rectangle getMaximumWindowBounds() {
		return getMaximumWindowBounds(GraphicsEnvironment.getLocalGraphicsEnvironment().getDefaultScreenDevice().getDefaultConfiguration().getBounds());
	}

	/**
	 * Determines the screen that most contains the specified window and returns the maximum size
	 * the window can be on that screen.
	 * 
	 * @param window The window to determine a maximum bounds for.
	 * @return The maximum bounds that fits on a screen.
	 */
	public static Rectangle getMaximumWindowBounds(Window window) {
		return getMaximumWindowBounds(window.getBounds());
	}

	/**
	 * Determines the screen that most contains the specified panel area and returns the maximum
	 * size a window can be on that screen.
	 * 
	 * @param panel The panel that contains the area.
	 * @param area The area within the panel to use when determining the maximum bounds for a
	 *            window.
	 * @return The maximum bounds that fits on a screen.
	 */
	public static Rectangle getMaximumWindowBounds(Component panel, Rectangle area) {
		area = new Rectangle(area);
		UIUtilities.convertRectangleToScreen(area, panel);
		return getMaximumWindowBounds(area);
	}

	/**
	 * Determines the screen that most contains the specified global bounds and returns the maximum
	 * size a window can be on that screen.
	 * 
	 * @param bounds The global bounds to use when determining the maximum bounds for a window.
	 * @return The maximum bounds that fits on a screen.
	 */
	public static Rectangle getMaximumWindowBounds(Rectangle bounds) {
		GraphicsEnvironment ge = GraphicsEnvironment.getLocalGraphicsEnvironment();
		GraphicsDevice gd = getPreferredScreenDevice(bounds);

		if (gd == ge.getDefaultScreenDevice()) {
			bounds = ge.getMaximumWindowBounds();
			// The Mac (and now Windows as of Java 5) already return the correct
			// value... try to fix it up for the other platforms. This doesn't
			// currently work, either, since the other platforms seem to always
			// return empty insets.
			if (!Platform.isMacintosh() && !Platform.isWindows()) {
				Insets insets = Toolkit.getDefaultToolkit().getScreenInsets(gd.getDefaultConfiguration());

				// Since this is failing to do the right thing anyway, we're going
				// to try and come up with some reasonable limitations...
				if (insets.top == 0 && insets.bottom == 0) {
					insets.bottom = 48;
				}

				bounds.x += insets.left;
				bounds.y += insets.top;
				bounds.width -= insets.left + insets.right;
				bounds.height -= insets.top + insets.bottom;
			}
			return bounds;
		}
		return gd.getDefaultConfiguration().getBounds();
	}

	/**
	 * Forces the specified window onscreen.
	 * 
	 * @param window The window to force onscreen.
	 */
	public static void forceOnScreen(Window window) {
		Rectangle maxBounds = getMaximumWindowBounds(window);
		Rectangle bounds = window.getBounds();
		Point location = new Point(bounds.x, bounds.y);
		Dimension size = window.getMinimumSize();

		if (bounds.width < size.width) {
			bounds.width = size.width;
		}
		if (bounds.height < size.height) {
			bounds.height = size.height;
		}

		if (bounds.x < maxBounds.x) {
			bounds.x = maxBounds.x;
		} else if (bounds.x >= maxBounds.x + maxBounds.width) {
			bounds.x = maxBounds.x + maxBounds.width - 1;
		}

		if (bounds.x + bounds.width >= maxBounds.x + maxBounds.width) {
			bounds.x = maxBounds.x + maxBounds.width - bounds.width;
			if (bounds.x < maxBounds.x) {
				bounds.x = maxBounds.x;
				bounds.width = maxBounds.width;
			}
		}

		if (bounds.y < maxBounds.y) {
			bounds.y = maxBounds.y;
		} else if (bounds.y >= maxBounds.y + maxBounds.height) {
			bounds.y = maxBounds.y + maxBounds.height - 1;
		}

		if (bounds.y + bounds.height >= maxBounds.y + maxBounds.height) {
			bounds.y = maxBounds.y + maxBounds.height - bounds.height;
			if (bounds.y < maxBounds.y) {
				bounds.y = maxBounds.y;
				bounds.height = maxBounds.height;
			}
		}

		if (location.x != bounds.x || location.y != bounds.y) {
			window.setBounds(bounds);
		} else {
			window.setSize(bounds.width, bounds.height);
		}
		window.validate();
	}

	/** Forces a full repaint of all windows, disposing of any window buffers. */
	public static void forceRepaint() {
		for (AppWindow window : AppWindow.getAllWindows()) {
			window.repaint();
		}
	}

	/** Forces a full repaint and invalidate on all windows, disposing of any window buffers. */
	public static void forceRepaintAndInvalidate() {
		for (AppWindow window : AppWindow.getAllWindows()) {
			window.invalidate(window.getRootPane());
		}
	}

	/**
	 * @param gc The {@link Graphics} to prepare for use.
	 * @return The passed-in {@link Graphics2D}.
	 */
	public static Graphics2D prepare(Graphics gc) {
		Graphics2D g2d = (Graphics2D) gc;
		g2d.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, RenderingHints.VALUE_TEXT_ANTIALIAS_ON);
		return g2d;
	}

	/**
	 * @return A graphics context obtained by looking for an existing window and asking it for a
	 *         graphics context.
	 */
	public static Graphics2D getGraphics() {
		Frame frame = AppWindow.getTopWindow();
		Graphics2D g2d = frame == null ? null : (Graphics2D) frame.getGraphics();

		if (g2d == null) {
			Frame[] frames = Frame.getFrames();

			for (Frame element : frames) {
				if (element.isDisplayable()) {
					g2d = (Graphics2D) element.getGraphics();
					if (g2d != null) {
						return g2d;
					}
				}
			}
			BufferedImage image = new BufferedImage(32, 1, BufferedImage.TYPE_INT_ARGB);
			return image.createGraphics();
		}
		return prepare(g2d);
	}

	/** @return A {@link Frame} to use when a valid frame of any sort is all that is needed. */
	public static Frame getAnyFrame() {
		Frame frame = AppWindow.getTopWindow();

		if (frame == null) {
			Frame[] frames = Frame.getFrames();

			for (Frame element : frames) {
				if (element.isDisplayable()) {
					return element;
				}
			}
			return getHiddenFrame(true);
		}
		return frame;
	}

	/** @param ok Whether using the momentary fullscreen window trick is OK or not. */
	public static void setOKToUseFullScreenTrick(boolean ok) {
		OK_TO_USE_FULLSCREEN_TRICK = ok;
	}

	/** Attempts to force the app to the front. */
	public static void forceAppToFront() {
		boolean useFallBack = true;

		if (Platform.isMacintosh()) {
			useFallBack = false;
			try {
				if (MAC_NSAPPLICATION == null) {
					URLClassLoader loader = new URLClassLoader(new URL[] { new File("/System/Library/Java/").toURI().toURL() }); //$NON-NLS-1$
					Class<?> appClass = Class.forName("com.apple.cocoa.application.NSApplication", true, loader); //$NON-NLS-1$
					Method method = appClass.getDeclaredMethod("sharedApplication"); //$NON-NLS-1$

					MAC_NSAPPLICATION = method.invoke(null);
					MAC_ACTIVATE_IGNORING_OTHER_APPS = appClass.getDeclaredMethod("activateIgnoringOtherApps", new Class[] { Boolean.TYPE }); //$NON-NLS-1$
				}
				MAC_ACTIVATE_IGNORING_OTHER_APPS.invoke(MAC_NSAPPLICATION, new Object[] { Boolean.TRUE });
			} catch (Exception exception) {
				useFallBack = true;
			}
		}

		if (useFallBack && OK_TO_USE_FULLSCREEN_TRICK) {
			GraphicsEnvironment ge = GraphicsEnvironment.getLocalGraphicsEnvironment();
			GraphicsDevice gd = ge.getDefaultScreenDevice();
			AppWindow window = new AppWindow(null, null, null);

			window.setUndecorated(true);
			window.getContentPane().setBackground(new Color(0, 0, 0, 0));

			gd.setFullScreenWindow(window);
			gd.setFullScreenWindow(null);
			window.dispose();
		}
	}

	/**
	 * @param create Whether it should be created if it doesn't already exist.
	 * @return The single instance of a special, hidden window that can be used for various
	 *         operations that require a window before you actually have one available.
	 */
	public static Frame getHiddenFrame(boolean create) {
		BufferedImage titleIcon;

		if (HIDDEN_FRAME == null) {
			if (!create) {
				return null;
			}
			HIDDEN_FRAME = new Frame();
			HIDDEN_FRAME.setUndecorated(true);
			HIDDEN_FRAME.setBounds(0, 0, 0, 0);
		}
		titleIcon = AppWindow.getDefaultWindowIcon();
		if (HIDDEN_FRAME_ICON != titleIcon) {
			HIDDEN_FRAME_ICON = titleIcon;
			HIDDEN_FRAME.setIconImage(titleIcon);
		}
		return HIDDEN_FRAME;
	}

	/** @return <code>true</code> if the graphics system is safe to use. */
	public static boolean areGraphicsSafeToUse() {
		if (!GraphicsEnvironment.isHeadless()) {
			if (HEADLESS_CHECK_RESULT == 0) {
				// We do the following just in case we're in an X-Windows
				// environment without a valid DISPLAY device...
				try {
					GraphicsEnvironment.getLocalGraphicsEnvironment().getDefaultScreenDevice();
					HEADLESS_CHECK_RESULT = 1;
				} catch (Error error) {
					HEADLESS_CHECK_RESULT = 2;
				}
			}
			if (HEADLESS_CHECK_RESULT == 1) {
				return true;
			}
		}
		return false;
	}

	/** @return The reason <code>areGraphicsSafeToUse()</code> returned <code>false</code>. */
	public static String getReasonForUnsafeGraphics() {
		return MSG_HEADLESS;
	}
}
