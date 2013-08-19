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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.utility.io.Browser;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.utility.io.cmdline.CmdLineParser;
import com.trollworks.gcs.utility.io.print.PrintManager;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.layout.FlexColumn;
import com.trollworks.gcs.widgets.layout.FlexRow;
import com.trollworks.gcs.widgets.layout.FlexSpacer;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.text.MessageFormat;
import java.util.ArrayList;

import javax.swing.ImageIcon;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** The general preferences panel. */
public class GeneralPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
	private static String		MSG_GENERAL;
	private static String		MSG_DISPLAY_SPLASH;
	private static String		MSG_DISPLAY_SPLASH_TOOLTIP;
	private static String		MSG_NATIVE_PRINTER;
	private static String		MSG_NATIVE_PRINTER_TOOLTIP;
	private static String		MSG_BROWSER;
	private static String		MSG_BROWSER_TOOLTIP;
	private static String		MSG_CUSTOM;
	private static String		MSG_COMMAND;
	private static String		MSG_COMMAND_TOOLTIP;
	private static String		MSG_WARNING;
	private static String		MSG_WARNING_FORMAT;
	private static final String	MODULE			= "General";									//$NON-NLS-1$
	private static final String	DISPLAY_SPLASH	= "DisplaySplash";								//$NON-NLS-1$
	private ImageIcon			mWarningIcon	= new ImageIcon(Images.getMiniWarningIcon());
	private JCheckBox			mDisplaySplash;
	private JCheckBox			mUseNativePrinter;
	private JComboBox			mBrowserPopup;
	private JTextField			mCommand;
	private JLabel				mCommandLabel;
	private Browser				mDefaultBrowser;
	private boolean				mSwitchingBrowser;

	static {
		LocalizedMessages.initialize(GeneralPreferences.class);
	}

	/** @return Whether the splash screen should be displayed at startup. */
	public static boolean shouldDisplaySplash() {
		return Preferences.getInstance().getBooleanValue(MODULE, DISPLAY_SPLASH, true);
	}

	/**
	 * Creates a new {@link GeneralPreferences}.
	 * 
	 * @param owner The owning {@link PreferencesWindow}.
	 */
	public GeneralPreferences(PreferencesWindow owner) {
		super(MSG_GENERAL, owner);

		FlexColumn column = new FlexColumn();

		mDisplaySplash = createCheckBox(MSG_DISPLAY_SPLASH, MSG_DISPLAY_SPLASH_TOOLTIP, shouldDisplaySplash());
		column.add(mDisplaySplash);
		mUseNativePrinter = createCheckBox(MSG_NATIVE_PRINTER, MSG_NATIVE_PRINTER_TOOLTIP, PrintManager.useNativeDialogs());
		column.add(mUseNativePrinter);

		addSeparator(column);

		FlexRow row = new FlexRow();
		JLabel label = createLabel(MSG_BROWSER, MSG_BROWSER_TOOLTIP);
		UIUtilities.setOnlySize(label, label.getPreferredSize());
		row.add(label);
		mBrowserPopup = createBrowserCombo();
		row.add(mBrowserPopup);
		mCommandLabel = createLabel(MSG_COMMAND, MSG_COMMAND_TOOLTIP, mWarningIcon);
		row.add(mCommandLabel);
		mCommand = createCommandField();
		row.add(mCommand);
		column.add(row);

		column.add(new FlexSpacer(0, 0, false, true));

		adjustBrowserIcon();

		column.apply(this);
	}

	private JComboBox createBrowserCombo() {
		mDefaultBrowser = Browser.getDefaultBrowser();
		Browser browser = Browser.getPreferredBrowser();
		JComboBox combo = createCombo(MSG_BROWSER_TOOLTIP);
		for (Browser one : Browser.getStandardBrowsers()) {
			combo.addItem(one);
		}
		Browser custom = new Browser(Browser.getPreferredBrowser(), true);
		if (!custom.getTitle().equals(MSG_CUSTOM)) {
			custom = new Browser(mDefaultBrowser, true);
			custom.setTitle(MSG_CUSTOM);
		}
		combo.addItem(custom);
		if (getMatchingItemIndex(combo, browser) == -1) {
			browser = mDefaultBrowser;
		}
		combo.setSelectedIndex(getMatchingItemIndex(combo, browser));
		combo.addActionListener(this);
		combo.setMaximumRowCount(combo.getItemCount());
		UIUtilities.setOnlySize(combo, combo.getPreferredSize());
		add(combo);
		return combo;
	}

	private int getMatchingItemIndex(JComboBox combo, Object obj) {
		int count = combo.getItemCount();
		for (int i = 0; i < count; i++) {
			if (combo.getItemAt(i).equals(obj)) {
				return i;
			}
		}
		return -1;
	}

	private JTextField createCommandField() {
		Browser browser = getSelectedBrowser();
		JTextField field = new JTextField(browser.getCommandLine());
		field.setToolTipText(MSG_COMMAND_TOOLTIP);
		field.setEnabled(browser.getTitle().equals(MSG_CUSTOM));
		field.getDocument().addDocumentListener(this);
		Dimension size = field.getPreferredSize();
		Dimension maxSize = field.getMaximumSize();
		maxSize.height = size.height;
		field.setMaximumSize(maxSize);
		add(field);
		return field;
	}

	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();
		if (source == mBrowserPopup) {
			Browser browser = getSelectedBrowser();
			Browser.setPreferredBrowser(browser);
			mCommand.setEnabled(browser.getTitle().equals(MSG_CUSTOM));
			mSwitchingBrowser = true;
			mCommand.setText(browser.getCommandLine());
			mSwitchingBrowser = false;
		}
		adjustResetButton();
	}

	private Browser getSelectedBrowser() {
		return (Browser) mBrowserPopup.getSelectedItem();
	}

	private void syncBrowser() {
		if (!mSwitchingBrowser) {
			Browser browser = getSelectedBrowser();
			if (browser.getTitle().equals(MSG_CUSTOM)) {
				ArrayList<String> list = CmdLineParser.parseIntoList(mCommand.getText());

				if (list.isEmpty()) {
					browser.setCommandPath(""); //$NON-NLS-1$
					browser.setArguments(null);
				} else if (list.size() == 1) {
					browser.setCommandPath(list.get(0));
					browser.setArguments(null);
				} else {
					browser.setCommandPath(list.get(0));
					list.remove(0);
					browser.setArguments(list.toArray(new String[0]));
				}
				Browser.setPreferredBrowser(browser);
			}
			adjustBrowserIcon();
		}
	}

	@Override public void reset() {
		mDisplaySplash.setSelected(true);
		mUseNativePrinter.setSelected(false);
		mBrowserPopup.setSelectedIndex(getMatchingItemIndex(mBrowserPopup, mDefaultBrowser));
		adjustBrowserIcon();
	}

	@Override public boolean isSetToDefaults() {
		return shouldDisplaySplash() && !PrintManager.useNativeDialogs() && mDefaultBrowser.isEquivalent(getSelectedBrowser());
	}

	private void adjustBrowserIcon() {
		Browser browser = getSelectedBrowser();
		String tooltip = MSG_COMMAND_TOOLTIP;
		boolean needsIcon = Path.isCommandPathViable(browser.getCommandPath(), null) == null;

		if (needsIcon) {
			tooltip = MessageFormat.format(MSG_WARNING_FORMAT, tooltip, MSG_WARNING);
		}
		mCommandLabel.setIcon(needsIcon ? new ImageIcon(Images.getMiniWarningIcon()) : null);
		mCommandLabel.setToolTipText(tooltip);
		mCommand.setToolTipText(tooltip);
	}

	public void changedUpdate(DocumentEvent event) {
		Document document = event.getDocument();
		if (mCommand.getDocument() == document) {
			syncBrowser();
		}
		adjustResetButton();
	}

	public void insertUpdate(DocumentEvent event) {
		changedUpdate(event);
	}

	public void removeUpdate(DocumentEvent event) {
		changedUpdate(event);
	}

	public void itemStateChanged(ItemEvent event) {
		Object source = event.getSource();
		if (source == mDisplaySplash) {
			Preferences.getInstance().setValue(MODULE, DISPLAY_SPLASH, mDisplaySplash.isSelected());
		} else if (source == mUseNativePrinter) {
			PrintManager.useNativeDialogs(mUseNativePrinter.isSelected());
		}
		adjustResetButton();
	}
}
