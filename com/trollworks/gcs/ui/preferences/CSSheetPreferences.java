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

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDice;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKLinkedLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.TKWidgetBorderPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.button.TKCheckbox;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKFileDialog;

import java.awt.Frame;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.text.MessageFormat;

/** The sheet preferences panel. */
public class CSSheetPreferences extends CSPreferencePanel implements ActionListener {
	private static final String			MODULE							= "CSSheetPreferences";									//$NON-NLS-1$
	private static final String			USE_HTML_TEMPLATE_OVERRIDE		= "UseHTMLTemplateOverride";								//$NON-NLS-1$
	private static final String			HTML_TEMPLATE_OVERRIDE			= "HTMLTemplateOverride";									//$NON-NLS-1$
	private static final int			DEFAULT_PNG_RESOLUTION			= 200;
	private static final String			PNG_RESOLUTION					= "PNGResolution";											//$NON-NLS-1$
	private static final boolean		DEFAULT_OPTIONAL_DICE_RULES		= false;
	private static final String			OPTIONAL_DICE_RULES				= "UseOptionDiceRules";									//$NON-NLS-1$
	/** The optional dice rules preference key. */
	public static final String			OPTIONAL_DICE_RULES_PREF_KEY	= TKPreferences.getModuleKey(MODULE, OPTIONAL_DICE_RULES);
	private TKTextField					mPlayerName;
	private TKTextField					mCampaign;
	private TKTextField					mTechLevel;
	private CSPortraitPreferencePanel	mPortrait;
	private TKPopupMenu					mPNGResolutionPopup;
	private TKCheckbox					mUseOptionalDiceRules;
	private TKCheckbox					mUseHTMLTemplateOverride;
	private TKTextField					mHTMLTemplatePath;
	private TKButton					mHTMLTemplatePicker;

	/** Initializes the services controlled by these preferences. */
	public static void initialize() {
		adjustOptionalDiceRulesProperty(isOptionalDiceRulesUsed());
	}

	private static void adjustOptionalDiceRulesProperty(boolean use) {
		if (use) {
			System.setProperty(CMDice.USE_OPTIONAL_GURPS_DICE_ADDS, Boolean.TRUE.toString());
		} else {
			System.clearProperty(CMDice.USE_OPTIONAL_GURPS_DICE_ADDS);
		}
	}

	/** @return Whether the default HTML template has been overridden. */
	public static boolean isHTMLTemplateOverridden() {
		return TKPreferences.getInstance().getBooleanValue(MODULE, USE_HTML_TEMPLATE_OVERRIDE);
	}

	/** @return The HTML template to use when exporting to HTML. */
	public static String getHTMLTemplate() {
		return isHTMLTemplateOverridden() ? getHTMLTemplateOverride() : getDefaultHTMLTemplate();
	}

	private static String getHTMLTemplateOverride() {
		return TKPreferences.getInstance().getStringValue(MODULE, HTML_TEMPLATE_OVERRIDE);
	}

	/** @return The default HTML template to use when exporting to HTML. */
	public static String getDefaultHTMLTemplate() {
		return TKPath.normalizeFullPath(TKPath.getFullPath(System.getProperty("app.home", "."), "data/template.html")); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
	}

	/** @return The resolution to use when saving the sheet as a PNG. */
	public static int getPNGResolution() {
		return TKPreferences.getInstance().getIntValue(MODULE, PNG_RESOLUTION, DEFAULT_PNG_RESOLUTION);
	}

	/** @return Whether the optional dice rules from B269 are in use. */
	public static boolean isOptionalDiceRulesUsed() {
		return TKPreferences.getInstance().getBooleanValue(MODULE, OPTIONAL_DICE_RULES, DEFAULT_OPTIONAL_DICE_RULES);
	}

	/** Creates the general preferences panel. */
	public CSSheetPreferences() {
		super(Msgs.SHEET);
		add(createSheetDefaultsPanel());
		add(createMiscellaneousPanel());
	}

	private TKPanel createSheetDefaultsPanel() {
		TKPanel outerWrapper = new TKPanel(new TKColumnLayout(2));
		TKPanel wrapper = new TKPanel(new TKColumnLayout(2));

		mPortrait = createPortrait(outerWrapper, CMCharacter.getPortraitFromPortraitPath(CMCharacter.getDefaultPortraitPath()));
		mPlayerName = createTextField(wrapper, Msgs.PLAYER, Msgs.PLAYER_TOOLTIP, CMCharacter.getDefaultPlayerName());
		mCampaign = createTextField(wrapper, Msgs.CAMPAIGN, Msgs.CAMPAIGN_TOOLTIP, CMCharacter.getDefaultCampaign());
		mTechLevel = createTextField(wrapper, Msgs.TECH_LEVEL, Msgs.TECH_LEVEL_TOOLTIP, CMCharacter.getDefaultTechLevel());
		outerWrapper.add(wrapper);
		wrapper = new TKPanel(new TKColumnLayout(1, 0, 5));
		wrapper.add(outerWrapper);
		wrapper.setBorder(new TKEmptyBorder(5));
		return new TKWidgetBorderPanel(new TKLabel(Msgs.SHEET_DEFAULTS, TKFont.CONTROL_FONT_KEY), wrapper);
	}

	private TKPanel createMiscellaneousPanel() {
		TKPanel panel = new TKPanel(new TKColumnLayout(1, 0, 5));
		TKPanel wrapper;

		mPNGResolutionPopup = createPNGResolutionPopup(panel);
		createHTMLTemplatePanel(panel);
		mUseOptionalDiceRules = createOptionalDiceRulesCheckbox(panel);

		wrapper = new TKPanel(new TKColumnLayout(1, 0, 5));
		wrapper.add(panel);
		wrapper.setBorder(new TKEmptyBorder(5));
		return new TKWidgetBorderPanel(new TKLabel(Msgs.MISCELLANEOUS, TKFont.CONTROL_FONT_KEY), wrapper);
	}

	private void createHTMLTemplatePanel(TKPanel wrapper) {
		TKPanel panel = new TKPanel(new TKColumnLayout(3));

		boolean overridden = isHTMLTemplateOverridden();
		mUseHTMLTemplateOverride = new TKCheckbox(Msgs.HTML_TEMPLATE_OVERRIDE, overridden);
		mUseHTMLTemplateOverride.setToolTipText(Msgs.HTML_TEMPLATE_OVERRIDE_TOOLTIP);
		mUseHTMLTemplateOverride.addActionListener(this);
		mHTMLTemplatePath = new TKTextField(getHTMLTemplate());
		mHTMLTemplatePath.setToolTipText(Msgs.HTML_TEMPLATE_OVERRIDE_TOOLTIP);
		mHTMLTemplatePath.addActionListener(this);
		mHTMLTemplatePicker = new TKButton(TKImage.get("TKFile")); //$NON-NLS-1$
		mHTMLTemplatePicker.setToolTipText(Msgs.HTML_TEMPLATE_OVERRIDE_TOOLTIP);
		mHTMLTemplatePicker.addActionListener(this);
		if (!overridden) {
			mHTMLTemplatePath.setEnabled(false);
			mHTMLTemplatePicker.setEnabled(false);
		}
		panel.add(mUseHTMLTemplateOverride);
		panel.add(mHTMLTemplatePath);
		panel.add(mHTMLTemplatePicker);
		wrapper.add(panel);
	}

	private TKCheckbox createOptionalDiceRulesCheckbox(TKPanel wrapper) {
		TKCheckbox checkbox = new TKCheckbox(Msgs.OPTIONAL_DICE_RULES, isOptionalDiceRulesUsed());
		checkbox.addActionListener(this);
		wrapper.add(checkbox);
		return checkbox;
	}

	private TKPopupMenu createPNGResolutionPopup(TKPanel wrapper) {
		TKPanel panel = new TKPanel(new TKColumnLayout(3));
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;

		for (int dpi : new int[] { 72, 96, 144, 150, 200, 300 }) {
			TKMenuItem item = new TKMenuItem(MessageFormat.format(Msgs.DPI, new Integer(dpi)));

			item.setUserObject(new Integer(dpi));
			menu.add(item);
		}
		popup = new TKPopupMenu(menu);
		popup.setToolTipText(Msgs.PNG_RESOLUTION_TOOLTIP);
		popup.setOnlySize(popup.getPreferredSize());
		popup.setSelectedUserObject(new Integer(getPNGResolution()));
		popup.addActionListener(this);
		panel.add(new TKLinkedLabel(popup, Msgs.PNG_RESOLUTION));
		panel.add(popup);
		panel.add(new TKPanel());
		wrapper.add(panel);
		return popup;
	}

	private CSPortraitPreferencePanel createPortrait(TKPanel wrapper, BufferedImage image) {
		CSPortraitPreferencePanel panel;

		if (image != null) {
			image = TKImage.scale(image, CMCharacter.PORTRAIT_WIDTH, CMCharacter.PORTRAIT_HEIGHT);
		}
		panel = new CSPortraitPreferencePanel(image);
		panel.addActionListener(this);
		wrapper.add(panel);
		return panel;
	}

	private TKTextField createTextField(TKPanel wrapper, String name, String tooltip, String value) {
		TKTextField field = new TKTextField(value);

		field.setToolTipText(tooltip);
		field.addActionListener(this);
		wrapper.add(new TKLinkedLabel(field, name));
		wrapper.add(field);
		return field;
	}

	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();

		if (source == mPlayerName) {
			CMCharacter.setDefaultPlayerName(mPlayerName.getText());
		} else if (source == mCampaign) {
			CMCharacter.setDefaultCampaign(mCampaign.getText());
		} else if (source == mPortrait) {
			TKFileDialog dialog = new TKFileDialog((Frame) getBaseWindow(), true);
			TKFileFilter filter = new TKFileFilter(Msgs.IMAGE_FILES, ".png .jpg .jpeg .gif"); //$NON-NLS-1$

			dialog.addFileFilter(filter);
			dialog.setActiveFileFilter(filter);
			if (dialog.doModal() == TKDialog.OK) {
				setPortrait(dialog.getSelectedItem().getAbsolutePath());
			}
		} else if (source == mTechLevel) {
			CMCharacter.setDefaultTechLevel(mTechLevel.getText());
		} else if (source == mPNGResolutionPopup) {
			TKPreferences.getInstance().setValue(MODULE, PNG_RESOLUTION, ((Integer) mPNGResolutionPopup.getSelectedItemUserObject()).intValue());
		} else if (source == mUseOptionalDiceRules) {
			boolean checked = mUseOptionalDiceRules.isChecked();
			adjustOptionalDiceRulesProperty(checked);
			TKPreferences.getInstance().setValue(MODULE, OPTIONAL_DICE_RULES, checked);
		} else if (source == mUseHTMLTemplateOverride) {
			boolean checked = mUseHTMLTemplateOverride.isChecked();
			TKPreferences.getInstance().setValue(MODULE, USE_HTML_TEMPLATE_OVERRIDE, checked);
			mHTMLTemplatePath.setEnabled(checked);
			mHTMLTemplatePicker.setEnabled(checked);
			mHTMLTemplatePath.setTextAndImprint(getHTMLTemplate());
		} else if (source == mHTMLTemplatePath) {
			if (isHTMLTemplateOverridden()) {
				TKPreferences.getInstance().setValue(MODULE, HTML_TEMPLATE_OVERRIDE, mHTMLTemplatePath.getTextOrImprint());
			}
		} else if (source == mHTMLTemplatePicker) {
			TKFileDialog dialog = new TKFileDialog((Frame) getBaseWindow(), true);
			TKFileFilter filter = new TKFileFilter(Msgs.HTML_FILES, ".html .htm"); //$NON-NLS-1$

			dialog.addFileFilter(filter);
			dialog.setActiveFileFilter(filter);
			if (dialog.doModal() == TKDialog.OK) {
				mHTMLTemplatePath.setTextAndImprint(TKPath.getFullPath(dialog.getSelectedItem()));
			}
		}
		adjustResetButton();
	}

	@Override public void reset() {
		mPlayerName.setText(System.getProperty("user.name")); //$NON-NLS-1$
		mCampaign.setText(""); //$NON-NLS-1$
		mTechLevel.setText(CMCharacter.DEFAULT_TECH_LEVEL);
		setPortrait(CMCharacter.DEFAULT_PORTRAIT);
		mPNGResolutionPopup.setSelectedUserObject(new Integer(DEFAULT_PNG_RESOLUTION));
		mUseOptionalDiceRules.setCheckedState(DEFAULT_OPTIONAL_DICE_RULES);
		mUseHTMLTemplateOverride.setCheckedState(false);
	}

	@Override public boolean isSetToDefaults() {
		return CMCharacter.getDefaultPlayerName().equals(System.getProperty("user.name")) && CMCharacter.getDefaultCampaign().equals("") && CMCharacter.getDefaultPortraitPath().equals(CMCharacter.DEFAULT_PORTRAIT) && CMCharacter.getDefaultTechLevel().equals(CMCharacter.DEFAULT_TECH_LEVEL) && getPNGResolution() == DEFAULT_PNG_RESOLUTION && isOptionalDiceRulesUsed() == DEFAULT_OPTIONAL_DICE_RULES && isHTMLTemplateOverridden() == false; //$NON-NLS-1$ //$NON-NLS-2$
	}

	private void setPortrait(String path) {
		BufferedImage image = CMCharacter.getPortraitFromPortraitPath(path);

		CMCharacter.setDefaultPortraitPath(path);
		if (image != null) {
			image = TKImage.scale(image, CMCharacter.PORTRAIT_WIDTH, CMCharacter.PORTRAIT_HEIGHT);
		}
		mPortrait.setPortrait(image);
	}
}
