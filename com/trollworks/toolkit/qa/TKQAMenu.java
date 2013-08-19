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

package com.trollworks.toolkit.qa;

import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.widget.menu.TKBaseMenu;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuBar;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Graphics2D;
import java.awt.event.KeyEvent;
import java.text.MessageFormat;

/** Provides a QA menu within a base window. */
public class TKQAMenu extends TKMenu implements TKMenuTarget {
	/** Prefix to identify QA-only commands. */
	public static final String	QA_CMD_PREFIX			= "QA-";									//$NON-NLS-1$
	private static final String	MONITOR_UNDO			= QA_CMD_PREFIX + "MonitorUndo";			//$NON-NLS-1$
	private static final String	DIAGNOSE_LOAD_SAVE		= QA_CMD_PREFIX + "DiagnoseLoadSave";		//$NON-NLS-1$
	private static final String	NO_DOUBLE_BUFFERING		= QA_CMD_PREFIX + "NoDoubleBuffering";		//$NON-NLS-1$
	private static final String	SHOW_DRAWING			= QA_CMD_PREFIX + "ShowDrawing";			//$NON-NLS-1$
	private static final String	SHOW_NO_TOOLTIPS		= QA_CMD_PREFIX + "ShowNoToolTips";		//$NON-NLS-1$
	private static final String	ANTIALIAS_FONTS			= QA_CMD_PREFIX + "AntiAliasFonts";		//$NON-NLS-1$
	private static final String	ALPHA_INTERPOLATION		= QA_CMD_PREFIX + "AlphaInterpolation";	//$NON-NLS-1$
	private static final String	ANTIALIAS				= QA_CMD_PREFIX + "AntiAlias";				//$NON-NLS-1$
	private static final String	COLOR_RENDERING			= QA_CMD_PREFIX + "ColorRendering";		//$NON-NLS-1$
	private static final String	DITHERING				= QA_CMD_PREFIX + "Dithering";				//$NON-NLS-1$
	private static final String	FRACTIONAL_METRICS		= QA_CMD_PREFIX + "FractionalMetrics";		//$NON-NLS-1$
	private static final String	BICUBIC_INTERPOLATION	= QA_CMD_PREFIX + "BiCubicInterpolation";	//$NON-NLS-1$
	private static final String	BILINEAR_INTERPOLATION	= QA_CMD_PREFIX + "BiLinearInterpolation";	//$NON-NLS-1$
	private static final String	NEAREST_INTERPOLATION	= QA_CMD_PREFIX + "NearestInterpolation";	//$NON-NLS-1$
	private static final String	RENDERING				= QA_CMD_PREFIX + "Rendering";				//$NON-NLS-1$
	private static final String	STROKE_CONTROL			= QA_CMD_PREFIX + "StrokeControl";			//$NON-NLS-1$
	private static final String	MAX_QUALITY				= QA_CMD_PREFIX + "MaxQuality";			//$NON-NLS-1$
	private static final String	MIN_QUALITY				= QA_CMD_PREFIX + "MinQuality";			//$NON-NLS-1$
	private static final String	SHOW_RENDERING_HINTS	= QA_CMD_PREFIX + "ShowHints";				//$NON-NLS-1$
	private static final String	SHOW_WINDOW_SIZE		= QA_CMD_PREFIX + "ShowWindowSize";		//$NON-NLS-1$

	/** Creates a QA menu. */
	public TKQAMenu() {
		super(Msgs.MENU_TITLE);

		add(new TKMenuItem("", SHOW_WINDOW_SIZE)); //$NON-NLS-1$
		addSeparator();
		add(new TKMenuItem(Msgs.MONITOR_UNDO, MONITOR_UNDO));
		addSeparator();
		add(new TKMenuItem(Msgs.DIAGNOSE_LOAD_SAVE, DIAGNOSE_LOAD_SAVE));
		addSeparator();
		add(new TKMenuItem(Msgs.DOUBLE_BUFFERED, NO_DOUBLE_BUFFERING));
		add(new TKMenuItem(Msgs.SHOW_DRAWING, new TKKeystroke(KeyEvent.VK_F11, 0), SHOW_DRAWING));
		add(new TKMenuItem(Msgs.SHOW_PANELS_WITHOUT_TOOLTIPS, new TKKeystroke(KeyEvent.VK_F10, 0), SHOW_NO_TOOLTIPS));
		addSeparator();
		add(new TKMenuItem(Msgs.ANTIALIASED_FONTS, ANTIALIAS_FONTS));
		add(new TKMenuItem(Msgs.FRACTIONAL_FONT_METRICS, FRACTIONAL_METRICS));
		add(new TKMenuItem(Msgs.HIGH_QUALITY_RENDERING, RENDERING));
		add(new TKMenuItem(Msgs.NORMALIZE_STROKES, STROKE_CONTROL));
		add(new TKMenuItem(Msgs.ANTIALIASING, ANTIALIAS));
		add(new TKMenuItem(Msgs.HIGH_QUALITY_COLOR_RENDERING, COLOR_RENDERING));
		add(new TKMenuItem(Msgs.DITHER_WHEN_NEEDED, DITHERING));

		TKMenu subMenu = new TKMenu(Msgs.INTERPOLATION);
		subMenu.add(new TKMenuItem(Msgs.BICUBIC, BICUBIC_INTERPOLATION));
		subMenu.add(new TKMenuItem(Msgs.BILINEAR, BILINEAR_INTERPOLATION));
		subMenu.add(new TKMenuItem(Msgs.NEAREST_NEIGHBOR, NEAREST_INTERPOLATION));
		add(subMenu);

		add(new TKMenuItem(Msgs.HIGH_QUALITY_ALPHA_INTERPOLATION, ALPHA_INTERPOLATION));
		addSeparator();
		add(new TKMenuItem(Msgs.SET_MAXIMUM_QUALITY, MAX_QUALITY));
		add(new TKMenuItem(Msgs.SET_MINIMUM_QUALITY, MIN_QUALITY));
		addSeparator();
		add(new TKMenuItem(Msgs.SHOW_RENDERING_HINTS, SHOW_RENDERING_HINTS));
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (BICUBIC_INTERPOLATION.equals(command)) {
			item.setMarked(TKGraphics.isBiCubicInterpolationOn());
		} else if (BILINEAR_INTERPOLATION.equals(command)) {
			item.setMarked(TKGraphics.isBiLinearInterpolationOn());
		} else if (NEAREST_INTERPOLATION.equals(command)) {
			item.setMarked(TKGraphics.isNearestInterpolationOn());
		} else if (ANTIALIAS_FONTS.equals(command)) {
			item.setMarked(TKGraphics.isAntiAliasFontsOn());
		} else if (ALPHA_INTERPOLATION.equals(command)) {
			item.setMarked(TKGraphics.isQualityAlphaInterpolationOn());
		} else if (ANTIALIAS.equals(command)) {
			item.setMarked(TKGraphics.isAntiAliasOn());
		} else if (COLOR_RENDERING.equals(command)) {
			item.setMarked(TKGraphics.isQualityColorRenderingOn());
		} else if (DITHERING.equals(command)) {
			item.setMarked(TKGraphics.isDitheringOn());
		} else if (FRACTIONAL_METRICS.equals(command)) {
			item.setMarked(TKGraphics.isFractionalMetricsOn());
		} else if (RENDERING.equals(command)) {
			item.setMarked(TKGraphics.isQualityRenderingOn());
		} else if (STROKE_CONTROL.equals(command)) {
			item.setMarked(TKGraphics.isStrokeNormalizationOn());
		} else if (NO_DOUBLE_BUFFERING.equals(command)) {
			item.setMarked(TKGraphics.useDoubleBuffering());
		} else if (SHOW_DRAWING.equals(command)) {
			item.setMarked(TKGraphics.isShowDrawingOn());
		} else if (SHOW_NO_TOOLTIPS.equals(command)) {
			item.setMarked(TKGraphics.isShowComponentsWithNoToolTipsOn());
		} else if (MAX_QUALITY.equals(command) || MIN_QUALITY.equals(command) || SHOW_RENDERING_HINTS.equals(command) || MONITOR_UNDO.equals(command)) {
			// Nothing to do...
		} else if (SHOW_WINDOW_SIZE.equals(command)) {
			TKWindow window = getOwningWindow();

			item.setTitle(MessageFormat.format(Msgs.CURRENT_WINDOW_SIZE_FORMAT, new Integer(window.getWidth()), new Integer(window.getHeight())));
		} else if (DIAGNOSE_LOAD_SAVE.equals(command)) {
			item.setMarked(TKDebug.isKeySet(TKDebug.KEY_DIAGNOSE_LOAD_SAVE));
		} else {
			return false;
		}
		return true;
	}

	private TKWindow getOwningWindow() {
		TKBaseMenu menu = getParentMenu();

		while (!(menu instanceof TKMenuBar)) {
			menu = menu.getParentMenu();
		}
		return (TKWindow) ((TKMenuBar) menu).getBaseWindow();
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (MONITOR_UNDO.equals(command)) {
			(new TKUndoMonitorWindow(getOwningWindow())).setVisible(true);
		} else if (SHOW_WINDOW_SIZE.equals(command)) {
			// Nothing to do...
		} else if (NO_DOUBLE_BUFFERING.equals(command)) {
			TKGraphics.setDoubleBuffering(!TKGraphics.useDoubleBuffering());
		} else if (SHOW_DRAWING.equals(command)) {
			TKGraphics.setShowDrawing(!TKGraphics.isShowDrawingOn());
		} else if (SHOW_NO_TOOLTIPS.equals(command)) {
			TKGraphics.setShowComponentsWithNoToolTips(!TKGraphics.isShowComponentsWithNoToolTipsOn());
		} else if (ANTIALIAS_FONTS.equals(command)) {
			TKGraphics.setAntiAliasFonts(!TKGraphics.isAntiAliasFontsOn());
		} else if (ALPHA_INTERPOLATION.equals(command)) {
			TKGraphics.setQualityAlphaInterpolation(!TKGraphics.isQualityAlphaInterpolationOn());
		} else if (ANTIALIAS.equals(command)) {
			TKGraphics.setAntiAlias(!TKGraphics.isAntiAliasOn());
		} else if (COLOR_RENDERING.equals(command)) {
			TKGraphics.setQualityColorRendering(!TKGraphics.isQualityColorRenderingOn());
		} else if (DITHERING.equals(command)) {
			TKGraphics.setDithering(!TKGraphics.isDitheringOn());
		} else if (FRACTIONAL_METRICS.equals(command)) {
			TKGraphics.setFractionalMetrics(!TKGraphics.isFractionalMetricsOn());
		} else if (RENDERING.equals(command)) {
			TKGraphics.setQualityRendering(!TKGraphics.isQualityRenderingOn());
		} else if (STROKE_CONTROL.equals(command)) {
			TKGraphics.setStrokeNormalization(!TKGraphics.isStrokeNormalizationOn());
		} else if (BICUBIC_INTERPOLATION.equals(command)) {
			TKGraphics.setInterpolation(TKGraphics.BICUBIC_INTERPOLATION);
		} else if (BILINEAR_INTERPOLATION.equals(command)) {
			TKGraphics.setInterpolation(TKGraphics.BILINEAR_INTERPOLATION);
		} else if (NEAREST_INTERPOLATION.equals(command)) {
			TKGraphics.setInterpolation(TKGraphics.NEAREST_INTERPOLATION);
		} else if (MAX_QUALITY.equals(command)) {
			TKGraphics.setAntiAliasFonts(true);
			TKGraphics.setQualityAlphaInterpolation(true);
			TKGraphics.setAntiAlias(true);
			TKGraphics.setQualityColorRendering(true);
			TKGraphics.setDithering(true);
			TKGraphics.setFractionalMetrics(true);
			TKGraphics.setQualityRendering(true);
			TKGraphics.setStrokeNormalization(true);
			TKGraphics.setInterpolation(TKGraphics.BICUBIC_INTERPOLATION);
		} else if (MIN_QUALITY.equals(command)) {
			TKGraphics.setAntiAliasFonts(false);
			TKGraphics.setQualityAlphaInterpolation(false);
			TKGraphics.setAntiAlias(false);
			TKGraphics.setQualityColorRendering(false);
			TKGraphics.setDithering(false);
			TKGraphics.setFractionalMetrics(false);
			TKGraphics.setQualityRendering(false);
			TKGraphics.setStrokeNormalization(false);
			TKGraphics.setInterpolation(TKGraphics.NEAREST_INTERPOLATION);
		} else if (SHOW_RENDERING_HINTS.equals(command)) {
			Graphics2D g2d = TKGraphics.getGraphics();

			System.err.println(g2d.getRenderingHints().toString().replace(',', '\n'));
			System.err.println();
			g2d.dispose();
		} else if (DIAGNOSE_LOAD_SAVE.equals(command)) {
			System.setProperty(TKDebug.KEY_DIAGNOSE_LOAD_SAVE, Boolean.toString(!TKDebug.isKeySet(TKDebug.KEY_DIAGNOSE_LOAD_SAVE)));
		} else {
			return false;
		}
		return true;
	}
}
