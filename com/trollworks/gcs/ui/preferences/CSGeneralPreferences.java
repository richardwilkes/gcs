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

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.io.cmdline.TKCmdLineParser;
import com.trollworks.toolkit.print.TKPrintManager;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKBrowser;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKDivider;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKSlider;
import com.trollworks.toolkit.widget.TKSliderFormatter;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.TKWidgetBorderPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.button.TKCheckbox;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.window.TKUserInputManager;

import java.awt.Dimension;
import java.awt.FlowLayout;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.text.MessageFormat;
import java.util.ArrayList;

/** The general preferences panel. */
public class CSGeneralPreferences extends CSPreferencePanel implements ActionListener, TKSliderFormatter {
	private static final String	MODULE					= "CSGeneralPreferences";							//$NON-NLS-1$
	private static final String	DISPLAY_SPLASH			= "DisplaySplash";									//$NON-NLS-1$
	private static final String	TOOLTIPS_ENABLED		= "ToolTipsEnabled";								//$NON-NLS-1$
	private static final String	TOOLTIP_DELAY			= "ToolTipDelay";									//$NON-NLS-1$
	private static final String	TOOLTIP_DURATION		= "ToolTipDuration";								//$NON-NLS-1$
	private static final String	UNDO_LEVELS				= "UndoLevels";									//$NON-NLS-1$
	/** The undo levels preference key. */
	public static final String	UNDO_LEVELS_PREF_KEY	= TKPreferences.getModuleKey(MODULE, UNDO_LEVELS);
	private static final int	DEFAULT_UNDO_LEVELS		= 32;
	private TKCheckbox			mDisplaySplash;
	private TKCheckbox			mUseNativePrintDialogs;
	private TKCheckbox			mToolTipsEnabled;
	private TKSlider			mToolTipDelay;
	private TKSlider			mToolTipDuration;
	private TKPopupMenu			mUndoLevels;
	private TKPopupMenu			mBrowserPopup;
	private TKTextField			mCommand;
	private TKLabel				mCommandLabel;
	private TKBrowser			mDefaultBrowser;
	private boolean				mSwitchingBrowser;

	/** Initializes the services controlled by these preferences. */
	public static void initialize() {
		TKUserInputManager.setToolTipsEnabled(TKPreferences.getInstance().getBooleanValue(MODULE, TOOLTIPS_ENABLED, true));
		TKUserInputManager.setToolTipDelay(TKPreferences.getInstance().getLongValue(MODULE, TOOLTIP_DELAY, TKUserInputManager.DEFAULT_TOOLTIP_DELAY));
		TKUserInputManager.setToolTipDuration(TKPreferences.getInstance().getLongValue(MODULE, TOOLTIP_DURATION, TKUserInputManager.DEFAULT_TOOLTIP_DURATION));
	}

	/** @return Whether the splash screen should be displayed at startup. */
	public static boolean shouldDisplaySplash() {
		return TKPreferences.getInstance().getBooleanValue(MODULE, DISPLAY_SPLASH, true);
	}

	/** @return The maximum number of undo levels per document. */
	public static int getMaximumUndoLevels() {
		return TKPreferences.getInstance().getIntValue(MODULE, UNDO_LEVELS, DEFAULT_UNDO_LEVELS);
	}

	/** Creates the general preferences panel. */
	public CSGeneralPreferences() {
		super(Msgs.GENERAL);
		add(createBrowserPanel());
		add(createToolTipPanel());
		add(createMiscPanel());
	}

	private TKPanel createBrowserPanel() {
		TKPanel wrapper = new TKPanel(new FlowLayout(FlowLayout.LEFT, 0, 0));
		TKPanel panel = new TKPanel(new TKColumnLayout(2));
		TKLabel label = new TKLabel(Msgs.BROWSER, TKAlignment.RIGHT);
		TKBrowser browser;
		String cmdLine;

		mDefaultBrowser = TKBrowser.getDefaultBrowser();
		browser = TKBrowser.getPreferredBrowser();
		mBrowserPopup = getBrowserMenu();
		mBrowserPopup.setOnlySize(mBrowserPopup.getPreferredSize());
		if (mBrowserPopup.getMatchingItemIndex(browser.getTitle()) == TKMenu.NOTHING_HIT) {
			browser = mDefaultBrowser;
		}
		mBrowserPopup.setSelectedItem(browser.getTitle());
		mBrowserPopup.setToolTipText(Msgs.BROWSER_TOOLTIP);
		mBrowserPopup.addActionListener(this);
		label.setToolTipText(Msgs.BROWSER_TOOLTIP);
		wrapper.add(mBrowserPopup);
		panel.add(label);
		panel.add(wrapper);

		mCommandLabel = new TKLabel(Msgs.COMMAND, TKAlignment.RIGHT);
		cmdLine = browser.getCommandLine();
		mCommand = new TKTextField(cmdLine, 50);
		mCommand.setImprint(cmdLine);
		mCommand.setToolTipText(Msgs.COMMAND_TOOLTIP);
		mCommand.setEnabled(browser.getTitle().equals(Msgs.CUSTOM));
		mCommand.addActionListener(this);
		mCommand.setMinimumSize(new Dimension(mCommand.getMinimumSize().width, mCommand.getPreferredSize().height));
		panel.add(mCommandLabel);
		panel.add(mCommand);

		adjustBrowserIcon();

		wrapper = new TKPanel(new TKColumnLayout(1, 0, 5));
		wrapper.add(panel);
		wrapper.setBorder(new TKEmptyBorder(5));
		return new TKWidgetBorderPanel(new TKLabel(Msgs.BROWSER, TKFont.CONTROL_FONT_KEY), wrapper);
	}

	private TKPopupMenu getBrowserMenu() {
		TKMenu menu = new TKMenu();
		TKBrowser[] browsers = TKBrowser.getStandardBrowsers();
		TKBrowser custom = new TKBrowser(TKBrowser.getPreferredBrowser(), true);

		if (!custom.getTitle().equals(Msgs.CUSTOM)) {
			custom = new TKBrowser(mDefaultBrowser, true);
			custom.setTitle(Msgs.CUSTOM);
		}

		for (TKBrowser element : browsers) {
			addBrowserMenuItem(menu, element);
		}
		if (browsers.length > 0) {
			menu.addSeparator();
		}
		addBrowserMenuItem(menu, custom);
		return new TKPopupMenu(menu, false);
	}

	private void addBrowserMenuItem(TKMenu menu, TKBrowser browser) {
		TKMenuItem item = new TKMenuItem(browser.getTitle());

		item.setUserObject(browser);
		menu.add(item);
	}

	private TKPanel createMiscPanel() {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(1, 0, 5));
		TKPanel undoWrapper = new TKPanel(new TKColumnLayout(3));
		TKLabel label = new TKLabel(Msgs.UNDO_LEVELS_PRE, TKAlignment.RIGHT);
		TKMenu menu = new TKMenu();
		int selection = 0;
		int max = getMaximumUndoLevels();

		wrapper.setBorder(new TKEmptyBorder(5));
		mDisplaySplash = new TKCheckbox(Msgs.DISPLAY_SPLASH, shouldDisplaySplash());
		mDisplaySplash.setToolTipText(Msgs.DISPLAY_SPLASH_TOOLTIP);
		mDisplaySplash.addActionListener(this);
		wrapper.add(mDisplaySplash);

		mUseNativePrintDialogs = new TKCheckbox(Msgs.USE_NATIVE_PRINT_DIALOGS, TKPrintManager.useNativeDialogs());
		mUseNativePrintDialogs.setToolTipText(Msgs.USE_NATIVE_PRINT_DIALOGS_TOOLTIP);
		mUseNativePrintDialogs.addActionListener(this);
		wrapper.add(mUseNativePrintDialogs);

		wrapper.add(new TKDivider(false, 0));

		label.setToolTipText(Msgs.UNDO_LEVELS_TOOLTIP);
		undoWrapper.add(label);
		for (int i = 0; i < 8; i++) {
			Integer value = new Integer(1 << i);
			TKMenuItem item = new TKMenuItem(value.toString());

			item.setUserObject(value);
			menu.add(item);
			if (max == value.intValue()) {
				selection = i;
			}
		}
		mUndoLevels = new TKPopupMenu(menu, selection);
		mUndoLevels.addActionListener(this);
		mUndoLevels.setOnlySize(mUndoLevels.getPreferredSize());
		mUndoLevels.setToolTipText(Msgs.UNDO_LEVELS_TOOLTIP);
		undoWrapper.add(mUndoLevels);
		label = new TKLabel(Msgs.UNDO_LEVELS_POST);
		label.setToolTipText(Msgs.UNDO_LEVELS_TOOLTIP);
		undoWrapper.add(label);
		wrapper.add(undoWrapper);

		return new TKWidgetBorderPanel(new TKLabel(Msgs.MISCELLANEOUS, TKFont.CONTROL_FONT_KEY), wrapper);
	}

	private TKPanel createToolTipPanel() {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(2, 2, 5));

		mToolTipsEnabled = new TKCheckbox(Msgs.DISPLAY_TOOLTIPS, TKUserInputManager.getToolTipsEnabled());
		mToolTipsEnabled.setToolTipText(Msgs.DISPLAY_TOOLTIPS_TOOLTIP);
		mToolTipsEnabled.addActionListener(this);

		wrapper.setBorder(new TKEmptyBorder(0, 5, 2, 5));
		mToolTipDelay = addSlider(wrapper, (int) TKUserInputManager.MAXIMUM_TOOLTIP_DELAY, (int) TKUserInputManager.MINIMUM_TOOLTIP_DELAY, (int) TKUserInputManager.getToolTipDelay(), 500, Msgs.DELAY, Msgs.DELAY_TOOLTIP);
		mToolTipDuration = addSlider(wrapper, (int) TKUserInputManager.MAXIMUM_TOOLTIP_DURATION, (int) TKUserInputManager.MINIMUM_TOOLTIP_DURATION, (int) TKUserInputManager.getToolTipDuration(), 5000, Msgs.DURATION, Msgs.DURATION_TOOLTIP);

		return new TKWidgetBorderPanel(mToolTipsEnabled, wrapper);
	}

	private TKSlider addSlider(TKPanel wrapper, int max, int min, int current, int divisor, String name, String tip) {
		int steps = 1 + (max - min) / divisor;
		int[] delay = new int[steps];
		TKLabel label = new TKLabel(name, TKAlignment.RIGHT);
		TKSlider slider;

		for (int i = 0; i < steps; i++) {
			delay[i] = i * divisor + min;
		}

		label.setToolTipText(tip);
		slider = new TKSlider(delay, this, true);
		slider.setToolTipText(tip);
		slider.setValue(current);
		slider.addActionListener(this);
		wrapper.add(label);
		wrapper.add(slider);
		return slider;
	}

	public String getFormattedValue(TKSlider slider, int value, boolean forTickMark) {
		if (forTickMark && value % 1000 == 0) {
			return TKNumberUtils.format(value / 1000);
		}
		return TKNumberUtils.format(value / 100 / 10.0);
	}

	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();

		if (source == mToolTipsEnabled) {
			boolean enabled = mToolTipsEnabled.isChecked();

			mToolTipDelay.setEnabled(enabled);
			mToolTipDuration.setEnabled(enabled);
			TKUserInputManager.setToolTipsEnabled(enabled);
			TKPreferences.getInstance().setValue(MODULE, TOOLTIPS_ENABLED, enabled);
		} else if (source == mToolTipDelay) {
			int delay = mToolTipDelay.getValue();

			TKUserInputManager.setToolTipDelay(delay);
			TKPreferences.getInstance().setValue(MODULE, TOOLTIP_DELAY, delay);
		} else if (source == mToolTipDuration) {
			int duration = mToolTipDuration.getValue();

			TKUserInputManager.setToolTipDuration(duration);
			TKPreferences.getInstance().setValue(MODULE, TOOLTIP_DURATION, duration);
		} else if (source == mDisplaySplash) {
			TKPreferences.getInstance().setValue(MODULE, DISPLAY_SPLASH, mDisplaySplash.isChecked());
		} else if (source == mUseNativePrintDialogs) {
			TKPrintManager.useNativeDialogs(mUseNativePrintDialogs.isChecked());
		} else if (source == mUndoLevels) {
			TKPreferences.getInstance().setValue(MODULE, UNDO_LEVELS, ((Integer) mUndoLevels.getSelectedItem().getUserObject()).intValue());
		} else if (source == mBrowserPopup) {
			TKBrowser browser = getSelectedBrowser();

			TKBrowser.setPreferredBrowser(browser);
			mCommand.setEnabled(browser.getTitle().equals(Msgs.CUSTOM));
			mSwitchingBrowser = true;
			mCommand.setTextAndImprint(browser.getCommandLine());
			mSwitchingBrowser = false;
		} else if (source == mCommand) {
			syncWithGUI();
		}
		adjustResetButton();
	}

	private TKBrowser getSelectedBrowser() {
		return (TKBrowser) mBrowserPopup.getSelectedItem().getUserObject();
	}

	private void syncWithGUI() {
		if (!mSwitchingBrowser) {
			TKBrowser browser = getSelectedBrowser();

			if (browser.getTitle().equals(Msgs.CUSTOM)) {
				ArrayList<String> list = TKCmdLineParser.parseIntoList(mCommand.getTextOrImprint());

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
				TKBrowser.setPreferredBrowser(browser);
			}
			adjustBrowserIcon();
		}
	}

	@Override public void reset() {
		mToolTipsEnabled.setCheckedState(true);
		mToolTipDelay.setValue((int) TKUserInputManager.DEFAULT_TOOLTIP_DELAY);
		mToolTipDuration.setValue((int) TKUserInputManager.DEFAULT_TOOLTIP_DURATION);
		mDisplaySplash.setCheckedState(true);
		mUseNativePrintDialogs.setCheckedState(false);
		mUndoLevels.setSelectedItem("" + DEFAULT_UNDO_LEVELS); //$NON-NLS-1$
		mBrowserPopup.setSelectedItem(mDefaultBrowser.getTitle());
	}

	@Override public boolean isSetToDefaults() {
		return TKUserInputManager.getToolTipsEnabled() && TKUserInputManager.getToolTipDelay() == TKUserInputManager.DEFAULT_TOOLTIP_DELAY && TKUserInputManager.getToolTipDuration() == TKUserInputManager.DEFAULT_TOOLTIP_DURATION && shouldDisplaySplash() && !TKPrintManager.useNativeDialogs() && getMaximumUndoLevels() == DEFAULT_UNDO_LEVELS && mDefaultBrowser.isEquivalent(getSelectedBrowser());
	}

	private void adjustBrowserIcon() {
		TKBrowser browser = getSelectedBrowser();
		String tooltip = Msgs.COMMAND_TOOLTIP;
		boolean needsIcon = TKPath.isCommandPathViable(browser.getCommandPath(), null) == null;

		if (needsIcon) {
			tooltip = MessageFormat.format(Msgs.WARNING_FORMAT, tooltip, Msgs.WARNING);
		}
		mCommandLabel.setImage(needsIcon ? TKImage.getMiniWarningIcon() : null);
		mCommandLabel.setToolTipText(tooltip);
		mCommandLabel.setToolTipText(tooltip);
	}
}
