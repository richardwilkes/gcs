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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.AppWindow;

import java.awt.BorderLayout;
import java.awt.Container;
import java.awt.FlowLayout;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

import javax.swing.JButton;
import javax.swing.JPanel;
import javax.swing.JTabbedPane;
import javax.swing.event.ChangeEvent;
import javax.swing.event.ChangeListener;

/** A window for editing application preferences. */
public class PreferencesWindow extends AppWindow implements ActionListener, ChangeListener {
	private static String				MSG_PREFERENCES;
	private static String				MSG_RESET;
	private static final String			PREFIX		= "CSPreferencesWindow.";	//$NON-NLS-1$
	private static PreferencesWindow	INSTANCE	= null;
	private JTabbedPane					mTabPanel;
	private JButton						mResetButton;

	static {
		LocalizedMessages.initialize(PreferencesWindow.class);
	}

	/** Initializes the services controlled by these preferences. */
	public static void initialize() {
		SheetPreferences.initialize();
	}

	/** Displays the preferences window. */
	public static void display() {
		if (INSTANCE == null) {
			INSTANCE = new PreferencesWindow();
		}
		INSTANCE.setVisible(true);
	}

	private PreferencesWindow() {
		super(MSG_PREFERENCES, Images.getPreferencesIcon(), Images.getPreferencesIcon());
		Container content = getContentPane();
		mTabPanel = new JTabbedPane();
		addTab(new GeneralPreferences(this));
		addTab(new SheetPreferences(this));
		addTab(new FontPreferences(this));
		mTabPanel.addChangeListener(this);
		content.add(mTabPanel);
		content.add(createResetPanel(), BorderLayout.SOUTH);
		adjustResetButton();
		restoreBounds();
	}

	private void addTab(PreferencePanel panel) {
		mTabPanel.addTab(panel.toString(), panel);
	}

	@Override public void dispose() {
		INSTANCE = null;
		super.dispose();
	}

	private JPanel createResetPanel() {
		JPanel panel = new JPanel(new FlowLayout(FlowLayout.CENTER));
		mResetButton = new JButton(MSG_RESET);
		mResetButton.addActionListener(this);
		panel.add(mResetButton);
		return panel;
	}

	/** Call to adjust the reset button to the current panel. */
	public void adjustResetButton() {
		mResetButton.setEnabled(!((PreferencePanel) mTabPanel.getSelectedComponent()).isSetToDefaults());
	}

	public void actionPerformed(ActionEvent event) {
		((PreferencePanel) mTabPanel.getSelectedComponent()).reset();
		adjustResetButton();
	}

	@Override public String getWindowPrefsPrefix() {
		return PREFIX;
	}

	public void stateChanged(ChangeEvent event) {
		adjustResetButton();
	}
}
