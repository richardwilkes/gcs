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
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.TKWidgetBorderPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
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
	private static final String			MODULE					= "CSSheetPreferences"; //$NON-NLS-1$
	private static final int			DEFAULT_PNG_RESOLUTION	= 200;
	private static final String			PNG_RESOLUTION			= "PNGResolution";		//$NON-NLS-1$
	private TKTextField					mPlayerName;
	private TKTextField					mCampaign;
	private TKTextField					mTechLevel;
	private CSPortraitPreferencePanel	mPortrait;
	private TKPopupMenu					mPNGResolutionPopup;

	/** @return The resolution to use when saving the sheet as a PNG. */
	public static int getPNGResolution() {
		return TKPreferences.getInstance().getIntValue(MODULE, PNG_RESOLUTION, DEFAULT_PNG_RESOLUTION);
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

		wrapper = new TKPanel(new TKColumnLayout(1, 0, 5));
		wrapper.add(panel);
		wrapper.setBorder(new TKEmptyBorder(5));
		return new TKWidgetBorderPanel(new TKLabel(Msgs.MISCELLANEOUS, TKFont.CONTROL_FONT_KEY), wrapper);
	}

	private TKPopupMenu createPNGResolutionPopup(TKPanel wrapper) {
		TKPanel panel = new TKPanel(new TKColumnLayout(3));
		TKLabel label = new TKLabel(Msgs.PNG_RESOLUTION, TKAlignment.RIGHT);
		TKMenu menu = new TKMenu();
		TKPopupMenu popup;

		label.setToolTipText(Msgs.PNG_RESOLUTION_TOOLTIP);
		panel.add(label);

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
		TKLabel label = new TKLabel(name, TKAlignment.RIGHT);

		field.setToolTipText(tooltip);
		label.setToolTipText(tooltip);
		field.addActionListener(this);
		wrapper.add(label);
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
		}
		adjustResetButton();
	}

	@Override public void reset() {
		mPlayerName.setText(System.getProperty("user.name")); //$NON-NLS-1$
		mCampaign.setText(""); //$NON-NLS-1$
		mTechLevel.setText(CMCharacter.DEFAULT_TECH_LEVEL);
		setPortrait(CMCharacter.DEFAULT_PORTRAIT);
		mPNGResolutionPopup.setSelectedUserObject(new Integer(DEFAULT_PNG_RESOLUTION));
	}

	@Override public boolean isSetToDefaults() {
		return CMCharacter.getDefaultPlayerName().equals(System.getProperty("user.name")) && CMCharacter.getDefaultCampaign().equals("") && CMCharacter.getDefaultPortraitPath().equals(CMCharacter.DEFAULT_PORTRAIT) && CMCharacter.getDefaultTechLevel().equals(CMCharacter.DEFAULT_TECH_LEVEL) && getPNGResolution() == DEFAULT_PNG_RESOLUTION; //$NON-NLS-1$ //$NON-NLS-2$
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
