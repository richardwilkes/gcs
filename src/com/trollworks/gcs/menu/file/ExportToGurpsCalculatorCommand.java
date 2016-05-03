/*
 * Copyright (c) 1998-2016 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.common.GurpsCalculatorExportable;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.services.NotImplementedException;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Color;
import java.awt.Component;
import java.awt.Font;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.io.IOException;
import java.net.URISyntaxException;

import javax.swing.JEditorPane;
import javax.swing.JLabel;
import javax.swing.event.HyperlinkEvent;
import javax.swing.event.HyperlinkListener;

public class ExportToGurpsCalculatorCommand extends Command {
	@Localize("Export to GURPS Calculator\u2026")
	private static String	EXPORT_TO_GURPS_CALCULATOR;
	@Localize("Export to GURPS Calculator was successful.")
	private static String	SUCCESS_MESSAGE;
	@Localize("There was an error exporting to GURPS Calculator. Please try again later.")
	private static String	ERROR_MESSAGE;
	@Localize("You need to set a valid GURPS Calculator Key in sheet preferences.<br><a href='%s'>Click here</a> for more information.")
	private static String	KEY_MISSING_MESSAGE;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String							CMD_EXPORT_TO_GUPRS_CALCULATOR	= "ExportToGurpsCalculator";			//$NON-NLS-1$

	/** The singleton {@link ExportToGurpsCalculatorCommand}. */
	public static final ExportToGurpsCalculatorCommand	INSTANCE						= new ExportToGurpsCalculatorCommand();

	private ExportToGurpsCalculatorCommand() {
		super(EXPORT_TO_GURPS_CALCULATOR, CMD_EXPORT_TO_GUPRS_CALCULATOR, KeyEvent.VK_L);
	}

	@Override
	public void adjust() {
		setEnabled(getTarget(GurpsCalculatorExportable.class) != null);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		export(getTarget(GurpsCalculatorExportable.class));
	}

	/**
	 * Allows the user to save the file under another name.
	 *
	 * @param exportable The {@link GurpsCalculatorExportable} to work on.
	 * @return The file(s) actually written to. May be empty.
	 */
	@SuppressWarnings("static-method")
	public boolean export(GurpsCalculatorExportable exportable) {
		if (exportable == null) {
			return false;
		}
		boolean result;
		try {
			result = exportable.exportToGurpsCalculator();
			if (!result) {
				return result;
			}
		} catch (IOException | NotImplementedException exception) {
			result = false;
		}
		Component frame = getFocusOwner();
		String message = result ? SUCCESS_MESSAGE : ERROR_MESSAGE;
		String key = SheetPreferences.getGurpsCalculatorKey();
		if (key == null || !key.matches("[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89ab][0-9a-fA-F]{3}-[0-9a-fA-F]{12}")) { //$NON-NLS-1$
			message = String.format(KEY_MISSING_MESSAGE, SheetPreferences.GURPS_CALCULATOR_URL);
		}
		JEditorPane messagePane = new JEditorPane("text/html", String.format("<html><body style='%s'>%s</body></html>", getStyle(), message)); //$NON-NLS-1$//$NON-NLS-2$
		messagePane.setEditable(false);
		messagePane.setBorder(null);
		messagePane.addHyperlinkListener(new HyperlinkListener() {
			@Override
			public void hyperlinkUpdate(HyperlinkEvent event) {
				if (event.getEventType().equals(HyperlinkEvent.EventType.ACTIVATED) && java.awt.Desktop.isDesktopSupported()) {
					try {
						java.awt.Desktop.getDesktop().browse(event.getURL().toURI());
					} catch (IOException | URISyntaxException exception) {
						exception.printStackTrace();
					}
				}
			}
		});
		WindowUtils.showError(frame, messagePane);
		return result;
	}

	static String getStyle() {
		// for copying style
		JLabel label = new JLabel();
		Font font = label.getFont();
		Color color = label.getBackground();

		// create some css from the label's font
		StringBuffer style = new StringBuffer(String.format("font-family:%s;", font.getFamily())); //$NON-NLS-1$
		style.append(String.format("font-weight:%s;", font.isBold() ? "bold" : "normal")); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
		style.append(String.format("font-size:%dpt;", Integer.valueOf(font.getSize()))); //$NON-NLS-1$
		style.append(String.format("background-color: rgb(%d,%d,%d);", Integer.valueOf(color.getRed()), Integer.valueOf(color.getGreen()), Integer.valueOf(color.getBlue()))); //$NON-NLS-1$
		return style.toString();
	}
}
