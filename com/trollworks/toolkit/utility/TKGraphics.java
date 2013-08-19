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

package com.trollworks.toolkit.utility;

import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.GraphicsDevice;
import java.awt.GraphicsEnvironment;
import java.awt.Insets;
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
public class TKGraphics {
	/** The constant for bi-cubic interpolation. */
	public static final int			BICUBIC_INTERPOLATION				= 0;
	/** The constant for bi-linear interpolation. */
	public static final int			BILINEAR_INTERPOLATION				= 1;
	/** The constant for nearest interpolation. */
	public static final int			NEAREST_INTERPOLATION				= 2;
	private static boolean			ANTIALIAS_FONTS						= true;
	private static boolean			FRACTIONAL_METRICS					= false;
	private static boolean			QUALITY_RENDERING					= false;
	private static boolean			STROKE_NORMALIZATION				= true;
	private static boolean			ANTIALIAS							= false;
	private static boolean			QUALITY_COLOR_RENDERING				= false;
	private static boolean			DITHERING							= false;
	private static boolean			QUALITY_ALPHA_INTERPOLATION			= false;
	private static int				INTERPOLATION						= NEAREST_INTERPOLATION;
	private static boolean			SHOW_DRAWING						= false;
	private static boolean			SHOW_NO_TOOLTIPS					= false;
	private static boolean			USE_DOUBLE_BUFFERING				= true;
	private static Frame			HIDDEN_FRAME						= null;
	private static BufferedImage	HIDDEN_FRAME_ICON					= null;
	private static boolean			HEADLESS_PRINT_MODE					= false;
	private static int				HEADLESS_CHECK_RESULT				= 0;
	private static Object			MAC_NSAPPLICATION					= null;
	private static Method			MAC_ACTIVATE_IGNORING_OTHER_APPS	= null;
	private static boolean			OK_TO_USE_FULLSCREEN_TRICK			= true;

	/** @return Whether the headless print mode is enabled. */
	public static boolean inHeadlessPrintMode() {
		return HEADLESS_PRINT_MODE;
	}

	/** @param inHeadlessPrintMode Whether the headless print mode is enabled. */
	public static void setHeadlessPrintMode(boolean inHeadlessPrintMode) {
		HEADLESS_PRINT_MODE = inHeadlessPrintMode;
	}

	/**
	 * Configures the graphics object rendering hints for our defaults.
	 * 
	 * @param graphics The graphics object to configure.
	 * @return The originally passed-in graphics object, for convenience.
	 */
	public static Graphics2D configureGraphics(Graphics graphics) {
		return configureGraphics((Graphics2D) graphics);
	}

	/**
	 * Configures the graphics object rendering hints for our defaults.
	 * 
	 * @param g2d The graphics object to configure.
	 * @return The originally passed-in graphics object, for convenience.
	 */
	public static Graphics2D configureGraphics(Graphics2D g2d) {
		if (g2d != null) {
			g2d.setRenderingHint(RenderingHints.KEY_ALPHA_INTERPOLATION, isQualityAlphaInterpolationOn() ? RenderingHints.VALUE_ALPHA_INTERPOLATION_QUALITY : RenderingHints.VALUE_ALPHA_INTERPOLATION_SPEED);
			g2d.setRenderingHint(RenderingHints.KEY_ANTIALIASING, isAntiAliasOn() ? RenderingHints.VALUE_ANTIALIAS_ON : RenderingHints.VALUE_ANTIALIAS_OFF);
			g2d.setRenderingHint(RenderingHints.KEY_COLOR_RENDERING, isQualityColorRenderingOn() ? RenderingHints.VALUE_COLOR_RENDER_QUALITY : RenderingHints.VALUE_COLOR_RENDER_SPEED);
			g2d.setRenderingHint(RenderingHints.KEY_DITHERING, isDitheringOn() ? RenderingHints.VALUE_DITHER_ENABLE : RenderingHints.VALUE_DITHER_DISABLE);
			g2d.setRenderingHint(RenderingHints.KEY_FRACTIONALMETRICS, isFractionalMetricsOn() ? RenderingHints.VALUE_FRACTIONALMETRICS_ON : RenderingHints.VALUE_FRACTIONALMETRICS_OFF);
			g2d.setRenderingHint(RenderingHints.KEY_INTERPOLATION, getInterpolation());
			g2d.setRenderingHint(RenderingHints.KEY_RENDERING, isQualityRenderingOn() ? RenderingHints.VALUE_RENDER_QUALITY : RenderingHints.VALUE_RENDER_SPEED);
			g2d.setRenderingHint(RenderingHints.KEY_STROKE_CONTROL, isStrokeNormalizationOn() ? RenderingHints.VALUE_STROKE_NORMALIZE : RenderingHints.VALUE_STROKE_PURE);
			g2d.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, isAntiAliasFontsOn() ? RenderingHints.VALUE_TEXT_ANTIALIAS_ON : RenderingHints.VALUE_TEXT_ANTIALIAS_OFF);
		}
		return g2d;
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
		Rectangle overlapBounds = TKRectUtils.intersection(bounds, best.getDefaultConfiguration().getBounds());
		int bestOverlap = overlapBounds.width * overlapBounds.height;

		for (GraphicsDevice gd : ge.getScreenDevices()) {
			if (gd.getType() == GraphicsDevice.TYPE_RASTER_SCREEN) {
				overlapBounds = TKRectUtils.intersection(bounds, gd.getDefaultConfiguration().getBounds());
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
	public static Rectangle getMaximumWindowBounds(TKPanel panel, Rectangle area) {
		area = new Rectangle(area);
		TKPanel.convertRectangleToScreen(area, panel);
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
			if (!TKPlatform.isMacintosh() && !TKPlatform.isWindows()) {
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

		window.setBounds(bounds);
		window.validate();
	}

	/** Forces a full repaint of all windows, disposing of any window buffers. */
	public static void forceRepaint() {
		for (TKWindow window : TKWindow.getAllWindows()) {
			window.forceRepaint();
		}
	}

	/**
	 * Forces a full repaint and invalidate on all windows, disposing of any window buffers.
	 */
	public static void forceRepaintAndInvalidate() {
		for (TKWindow window : TKWindow.getAllWindows()) {
			window.forceRepaintAndInvalidate();
		}
	}

	/** @return <code>true</code> if overall anti-aliasing of fonts is turned on. */
	public static boolean isAntiAliasFontsOn() {
		return ANTIALIAS_FONTS;
	}

	/** @param antiAliasFonts Whether overall anti-aliased fonts should be on or not. */
	public static void setAntiAliasFonts(boolean antiAliasFonts) {
		if (ANTIALIAS_FONTS != antiAliasFonts) {
			ANTIALIAS_FONTS = antiAliasFonts;
			TKFont.resetRenderContext();
			forceRepaint();
		}
	}

	/** @param on Whether double-buffering is on or not. */
	public static void setDoubleBuffering(boolean on) {
		if (USE_DOUBLE_BUFFERING != on) {
			USE_DOUBLE_BUFFERING = on;
			forceRepaint();
		}
	}

	/** @return The current double buffering setting. */
	public static boolean useDoubleBuffering() {
		return USE_DOUBLE_BUFFERING;
	}

	/** @return <code>true</code> if show drawing is enabled. */
	public static boolean isShowDrawingOn() {
		return SHOW_DRAWING;
	}

	/** @return <code>true</code> if show components with no tooltips is enabled. */
	public static boolean isShowComponentsWithNoToolTipsOn() {
		return SHOW_NO_TOOLTIPS;
	}

	/** @return <code>true</code> if anti-aliasing is enabled. */
	public static boolean isAntiAliasOn() {
		return ANTIALIAS;
	}

	/** @return <code>true</code> if dithering is enabled. */
	public static boolean isDitheringOn() {
		return DITHERING;
	}

	/** @return <code>true</code> if fractional metrics are enabled. */
	public static boolean isFractionalMetricsOn() {
		return FRACTIONAL_METRICS;
	}

	/** @return <code>true</code> if alpha interpolation is enabled. */
	public static boolean isQualityAlphaInterpolationOn() {
		return QUALITY_ALPHA_INTERPOLATION;
	}

	/** @return <code>true</code> if bicubic interpolation is enabled. */
	public static boolean isBiCubicInterpolationOn() {
		return INTERPOLATION == BICUBIC_INTERPOLATION;
	}

	/** @return <code>true</code> if bilinear interpolation is enabled. */
	public static boolean isBiLinearInterpolationOn() {
		return INTERPOLATION == BILINEAR_INTERPOLATION;
	}

	/** @return <code>true</code> if nearest interpolation is enabled. */
	public static boolean isNearestInterpolationOn() {
		return INTERPOLATION == NEAREST_INTERPOLATION;
	}

	/** @return <code>true</code> if high-quality color rendering is enabled. */
	public static boolean isQualityColorRenderingOn() {
		return QUALITY_COLOR_RENDERING;
	}

	/** @return <code>true</code> if high-quality rendering is enabled. */
	public static boolean isQualityRenderingOn() {
		return QUALITY_RENDERING;
	}

	/** @return <code>true</code> if stroke normalization is enabled. */
	public static boolean isStrokeNormalizationOn() {
		return STROKE_NORMALIZATION;
	}

	/** @param antiAlias The anti-alias rendering hint. */
	public static void setAntiAlias(boolean antiAlias) {
		if (ANTIALIAS != antiAlias) {
			ANTIALIAS = antiAlias;
			forceRepaint();
		}
	}

	/** @param dithering The dithering rendering hint. */
	public static void setDithering(boolean dithering) {
		if (DITHERING != dithering) {
			DITHERING = dithering;
			forceRepaint();
		}
	}

	/** @param fractionalMetrics The fractional metrics rendering hint. */
	public static void setFractionalMetrics(boolean fractionalMetrics) {
		if (FRACTIONAL_METRICS != fractionalMetrics) {
			FRACTIONAL_METRICS = fractionalMetrics;
			TKFont.resetRenderContext();
			forceRepaint();
		}
	}

	/**
	 * @return A graphics context obtained by looking for an existing window and asking it for a
	 *         graphics context.
	 */
	public static Graphics2D getGraphics() {
		Frame frame = TKWindow.getTopWindow();
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
		return g2d;
	}

	/** @return A {@link Frame} to use when a valid frame of any sort is all that is needed. */
	public static Frame getAnyFrame() {
		Frame frame = TKWindow.getTopWindow();

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

		if (TKPlatform.isMacintosh()) {
			useFallBack = false;
			try {
				if (MAC_NSAPPLICATION == null) {
					URLClassLoader loader = new URLClassLoader(new URL[] { (new File("/System/Library/Java/")).toURL() }); //$NON-NLS-1$
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
			TKWindow window = new TKWindow(null, null);

			window.setUndecorated(true);
			window.getContent().setBackground(new Color(0, 0, 0, 0));

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
		titleIcon = TKWindow.getDefaultWindowIcon();
		if (HIDDEN_FRAME_ICON != titleIcon) {
			HIDDEN_FRAME_ICON = titleIcon;
			HIDDEN_FRAME.setIconImage(titleIcon);
		}
		return HIDDEN_FRAME;
	}

	private static Object getInterpolation() {
		switch (INTERPOLATION) {
			case BICUBIC_INTERPOLATION:
				return RenderingHints.VALUE_INTERPOLATION_BICUBIC;
			case BILINEAR_INTERPOLATION:
				return RenderingHints.VALUE_INTERPOLATION_BILINEAR;
			case NEAREST_INTERPOLATION:
				return RenderingHints.VALUE_INTERPOLATION_NEAREST_NEIGHBOR;
			default:
				return RenderingHints.VALUE_INTERPOLATION_BICUBIC;
		}
	}

	/**
	 * @param interpolation The interpolation rendering hint. Must be one of
	 *            {@link #BICUBIC_INTERPOLATION}, {@link #BILINEAR_INTERPOLATION}, or
	 *            {@link #NEAREST_INTERPOLATION}.
	 */
	public static void setInterpolation(int interpolation) {
		if (interpolation == BICUBIC_INTERPOLATION || interpolation == BILINEAR_INTERPOLATION || interpolation == NEAREST_INTERPOLATION) {
			if (INTERPOLATION != interpolation) {
				INTERPOLATION = interpolation;
				forceRepaint();
			}
		}
	}

	/** @param qualityAlphaInterpolation The quality alpha interpolation rendering hint. */
	public static void setQualityAlphaInterpolation(boolean qualityAlphaInterpolation) {
		if (QUALITY_ALPHA_INTERPOLATION != qualityAlphaInterpolation) {
			QUALITY_ALPHA_INTERPOLATION = qualityAlphaInterpolation;
			forceRepaint();
		}
	}

	/** @param qualityColorRendering The quality color rendering hint. */
	public static void setQualityColorRendering(boolean qualityColorRendering) {
		if (QUALITY_COLOR_RENDERING != qualityColorRendering) {
			QUALITY_COLOR_RENDERING = qualityColorRendering;
			forceRepaint();
		}
	}

	/** @param qualityRendering The quality rendering hint. */
	public static void setQualityRendering(boolean qualityRendering) {
		if (QUALITY_RENDERING != qualityRendering) {
			QUALITY_RENDERING = qualityRendering;
			forceRepaint();
		}
	}

	/** @param strokeNormalization The stroke normalization rendering hint. */
	public static void setStrokeNormalization(boolean strokeNormalization) {
		if (STROKE_NORMALIZATION != strokeNormalization) {
			STROKE_NORMALIZATION = strokeNormalization;
			forceRepaint();
		}
	}

	/** @param on Whether or not drawing is shown or not. */
	public static void setShowDrawing(boolean on) {
		SHOW_DRAWING = on;
	}

	/** @param on Whether or not drawing is shown or not. */
	public static void setShowComponentsWithNoToolTips(boolean on) {
		if (SHOW_NO_TOOLTIPS != on) {
			SHOW_NO_TOOLTIPS = on;
			forceRepaint();
		}
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
		return Msgs.HEADLESS;
	}
}
