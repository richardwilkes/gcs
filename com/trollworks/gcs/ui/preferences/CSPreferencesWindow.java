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

package com.trollworks.gcs.ui.preferences;

import com.trollworks.gcs.ui.common.CSWindow;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.widget.tab.TKTabbedPanel;

import java.awt.FlowLayout;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

/** A window for editing application preferences. */
public class CSPreferencesWindow extends CSWindow implements ActionListener {
	private static final String			PREFIX		= "CSPreferencesWindow.";	//$NON-NLS-1$
	private static CSPreferencesWindow	INSTANCE	= null;
	private TKTabbedPanel				mTabPanel;
	private TKButton					mResetButton;

	/** Initializes the services controlled by these preferences. */
	public static void initialize() {
		CSGeneralPreferences.initialize();
		CSFontPreferences.initialize();
	}

	/** Displays the preferences window. */
	public static void display() {
		if (INSTANCE == null) {
			INSTANCE = new CSPreferencesWindow();
		}
		INSTANCE.setVisible(true);
	}

	private CSPreferencesWindow() {
		super(Msgs.PREFERENCES, TKImage.getPreferencesIcon(), TKImage.getPreferencesIcon());
		TKPanel content = getContent();
		content.setBorder(new TKEmptyBorder(10));
		mTabPanel = new TKTabbedPanel(new TKPanel[] { new CSGeneralPreferences(), new CSSheetPreferences(), new CSFontPreferences(), new CSMenuKeyPreferences() });
		mTabPanel.addActionListener(this);
		content.add(mTabPanel);
		content.add(createResetPanel(), TKCompassPosition.SOUTH);
		adjustResetButton();
		restoreBounds();
	}

	@Override public void dispose() {
		INSTANCE = null;
		super.dispose();
	}

	private TKPanel createResetPanel() {
		TKPanel panel = new TKPanel(new FlowLayout(FlowLayout.CENTER));

		mResetButton = new TKButton(Msgs.RESET);
		mResetButton.addActionListener(this);
		panel.add(mResetButton);
		return panel;
	}

	/** Call to adjust the reset button to the current panel. */
	public void adjustResetButton() {
		mResetButton.setEnabled(!((CSPreferencePanel) mTabPanel.getCurrentPanel()).isSetToDefaults());
	}

	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();

		if (source == mResetButton) {
			((CSPreferencePanel) mTabPanel.getCurrentPanel()).reset();
		} else if (source == mTabPanel) {
			adjustResetButton();
		}
	}

	@Override public String getWindowPrefsPrefix() {
		return PREFIX;
	}
}
