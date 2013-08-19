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

import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.widgets.StdFileDialog;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.layout.Alignment;
import com.trollworks.gcs.widgets.layout.FlexColumn;
import com.trollworks.gcs.widgets.layout.FlexComponent;
import com.trollworks.gcs.widgets.layout.FlexGrid;
import com.trollworks.gcs.widgets.layout.FlexRow;
import com.trollworks.gcs.widgets.layout.FlexSpacer;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.awt.image.BufferedImage;
import java.io.File;
import java.text.MessageFormat;

import javax.swing.JButton;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;

/** The sheet preferences panel. */
public class SheetPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
	private static String			MSG_SHEET;
	private static String			MSG_PLAYER;
	private static String			MSG_PLAYER_TOOLTIP;
	private static String			MSG_CAMPAIGN;
	private static String			MSG_CAMPAIGN_TOOLTIP;
	private static String			MSG_TECH_LEVEL;
	private static String			MSG_TECH_LEVEL_TOOLTIP;
	private static String			MSG_SELECT_PORTRAIT;
	private static String			MSG_OPTIONAL_DICE_RULES;
	private static String			MSG_PNG_RESOLUTION_PRE;
	private static String			MSG_PNG_RESOLUTION_POST;
	private static String			MSG_PNG_RESOLUTION_TOOLTIP;
	private static String			MSG_DPI;
	private static String			MSG_HTML_TEMPLATE_OVERRIDE;
	private static String			MSG_HTML_TEMPLATE_PICKER;
	private static String			MSG_HTML_TEMPLATE_OVERRIDE_TOOLTIP;
	private static String			MSG_SELECT_HTML_TEMPLATE;
	private static final String		MODULE							= "Sheet";													//$NON-NLS-1$
	private static final String		OPTIONAL_DICE_RULES				= "UseOptionDiceRules";									//$NON-NLS-1$
	/** The optional dice rules preference key. */
	public static final String		OPTIONAL_DICE_RULES_PREF_KEY	= Preferences.getModuleKey(MODULE, OPTIONAL_DICE_RULES);
	private static final boolean	DEFAULT_OPTIONAL_DICE_RULES		= false;
	private static final int		DEFAULT_PNG_RESOLUTION			= 200;
	private static final String		PNG_RESOLUTION					= "PNGResolution";											//$NON-NLS-1$
	private static final int[]		DPI								= { 72, 96, 144, 150, 200, 300 };
	private static final String		USE_HTML_TEMPLATE_OVERRIDE		= "UseHTMLTemplateOverride";								//$NON-NLS-1$
	private static final String		HTML_TEMPLATE_OVERRIDE			= "HTMLTemplateOverride";									//$NON-NLS-1$
	private JTextField				mPlayerName;
	private JTextField				mCampaign;
	private JTextField				mTechLevel;
	private PortraitPreferencePanel	mPortrait;
	private JComboBox				mPNGResolutionCombo;
	private JCheckBox				mUseHTMLTemplateOverride;
	private JTextField				mHTMLTemplatePath;
	private JButton					mHTMLTemplatePicker;
	private JCheckBox				mUseOptionalDiceRules;

	static {
		LocalizedMessages.initialize(SheetPreferences.class);
	}

	/** Initializes the services controlled by these preferences. */
	public static void initialize() {
		adjustOptionalDiceRulesProperty(areOptionalDiceRulesUsed());
	}

	private static void adjustOptionalDiceRulesProperty(boolean use) {
		if (use) {
			System.setProperty(Dice.USE_OPTIONAL_GURPS_DICE_ADDS, Boolean.TRUE.toString());
		} else {
			System.clearProperty(Dice.USE_OPTIONAL_GURPS_DICE_ADDS);
		}
	}

	/** @return Whether the optional dice rules from B269 are in use. */
	public static boolean areOptionalDiceRulesUsed() {
		return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_DICE_RULES, DEFAULT_OPTIONAL_DICE_RULES);
	}

	/** @return The resolution to use when saving the sheet as a PNG. */
	public static int getPNGResolution() {
		return Preferences.getInstance().getIntValue(MODULE, PNG_RESOLUTION, DEFAULT_PNG_RESOLUTION);
	}

	/** @return Whether the default HTML template has been overridden. */
	public static boolean isHTMLTemplateOverridden() {
		return Preferences.getInstance().getBooleanValue(MODULE, USE_HTML_TEMPLATE_OVERRIDE);
	}

	/** @return The HTML template to use when exporting to HTML. */
	public static String getHTMLTemplate() {
		return isHTMLTemplateOverridden() ? getHTMLTemplateOverride() : getDefaultHTMLTemplate();
	}

	private static String getHTMLTemplateOverride() {
		return Preferences.getInstance().getStringValue(MODULE, HTML_TEMPLATE_OVERRIDE);
	}

	/** @return The default HTML template to use when exporting to HTML. */
	public static String getDefaultHTMLTemplate() {
		return Path.normalizeFullPath(Path.getFullPath(System.getProperty("app.home", "."), "data/template.html")); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
	}

	/**
	 * Creates a new {@link SheetPreferences}.
	 * 
	 * @param owner The owning {@link PreferencesWindow}.
	 */
	public SheetPreferences(PreferencesWindow owner) {
		super(MSG_SHEET, owner);
		FlexColumn column = new FlexColumn();

		FlexGrid grid = new FlexGrid();
		column.add(grid);

		mPortrait = createPortrait();
		FlexComponent comp = new FlexComponent(mPortrait, Alignment.LEFT_TOP, Alignment.LEFT_TOP);
		grid.add(comp, 0, 0, 4, 1);

		grid.add(createFlexLabel(MSG_PLAYER, MSG_PLAYER_TOOLTIP), 0, 1);
		mPlayerName = createTextField(MSG_PLAYER_TOOLTIP, Profile.getDefaultPlayerName());
		grid.add(mPlayerName, 0, 2);

		grid.add(createFlexLabel(MSG_CAMPAIGN, MSG_CAMPAIGN_TOOLTIP), 1, 1);
		mCampaign = createTextField(MSG_CAMPAIGN_TOOLTIP, Profile.getDefaultCampaign());
		grid.add(mCampaign, 1, 2);

		grid.add(createFlexLabel(MSG_TECH_LEVEL, MSG_TECH_LEVEL_TOOLTIP), 2, 1);
		mTechLevel = createTextField(MSG_TECH_LEVEL_TOOLTIP, Profile.getDefaultTechLevel());
		grid.add(mTechLevel, 2, 2);

		grid.add(new FlexSpacer(0, 0, false, true), 3, 1);
		grid.add(new FlexSpacer(0, 0, true, true), 3, 2);

		addSeparator(column);

		mUseOptionalDiceRules = createCheckBox(MSG_OPTIONAL_DICE_RULES, null, areOptionalDiceRulesUsed());
		column.add(mUseOptionalDiceRules);

		FlexRow row = new FlexRow();
		mUseHTMLTemplateOverride = createCheckBox(MSG_HTML_TEMPLATE_OVERRIDE, MSG_HTML_TEMPLATE_OVERRIDE_TOOLTIP, isHTMLTemplateOverridden());
		row.add(mUseHTMLTemplateOverride);
		mHTMLTemplatePath = createHTMLTemplatePathField();
		row.add(mHTMLTemplatePath);
		mHTMLTemplatePicker = createButton(MSG_HTML_TEMPLATE_PICKER, MSG_HTML_TEMPLATE_OVERRIDE_TOOLTIP);
		mHTMLTemplatePicker.setEnabled(isHTMLTemplateOverridden());
		row.add(mHTMLTemplatePicker);
		column.add(row);

		row = new FlexRow();
		row.add(createLabel(MSG_PNG_RESOLUTION_PRE, MSG_PNG_RESOLUTION_TOOLTIP));
		mPNGResolutionCombo = createPNGResolutionPopup();
		row.add(mPNGResolutionCombo);
		row.add(createLabel(MSG_PNG_RESOLUTION_POST, MSG_PNG_RESOLUTION_TOOLTIP, SwingConstants.LEFT));
		column.add(row);

		column.add(new FlexSpacer(0, 0, false, true));

		column.apply(this);
	}

	private JButton createButton(String title, String tooltip) {
		JButton button = new JButton(title);
		button.setOpaque(false);
		button.setToolTipText(tooltip);
		button.addActionListener(this);
		add(button);
		return button;
	}

	private FlexComponent createFlexLabel(String title, String tooltip) {
		return new FlexComponent(createLabel(title, tooltip), Alignment.RIGHT_BOTTOM, Alignment.CENTER);
	}

	private PortraitPreferencePanel createPortrait() {
		BufferedImage image = Profile.getPortraitFromPortraitPath(Profile.getDefaultPortraitPath());
		if (image != null) {
			image = Images.scale(image, Profile.PORTRAIT_WIDTH, Profile.PORTRAIT_HEIGHT);
		}
		PortraitPreferencePanel panel = new PortraitPreferencePanel(image);
		panel.addActionListener(this);
		add(panel);
		return panel;
	}

	private JTextField createHTMLTemplatePathField() {
		JTextField field = new JTextField(getHTMLTemplate());
		field.setToolTipText(MSG_HTML_TEMPLATE_OVERRIDE_TOOLTIP);
		field.setEnabled(isHTMLTemplateOverridden());
		field.getDocument().addDocumentListener(this);
		Dimension size = field.getPreferredSize();
		Dimension maxSize = field.getMaximumSize();
		maxSize.height = size.height;
		field.setMaximumSize(maxSize);
		add(field);
		return field;
	}

	private JComboBox createPNGResolutionPopup() {
		int selection = 0;
		int resolution = getPNGResolution();
		JComboBox combo = createCombo(MSG_PNG_RESOLUTION_TOOLTIP);
		for (int i = 0; i < DPI.length; i++) {
			combo.addItem(MessageFormat.format(MSG_DPI, new Integer(DPI[i])));
			if (DPI[i] == resolution) {
				selection = i;
			}
		}
		combo.setSelectedIndex(selection);
		combo.addActionListener(this);
		combo.setMaximumRowCount(combo.getItemCount());
		UIUtilities.setOnlySize(combo, combo.getPreferredSize());
		return combo;
	}

	private JTextField createTextField(String tooltip, String value) {
		JTextField field = new JTextField(value);
		field.setToolTipText(tooltip);
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
		if (source == mPortrait) {
			File file = StdFileDialog.choose(this, true, MSG_SELECT_PORTRAIT, null, null, "png", "jpg", "gif", "jpeg"); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
			if (file != null) {
				setPortrait(Path.getFullPath(file));
			}
		} else if (source == mPNGResolutionCombo) {
			Preferences.getInstance().setValue(MODULE, PNG_RESOLUTION, DPI[mPNGResolutionCombo.getSelectedIndex()]);
		} else if (source == mHTMLTemplatePicker) {
			File file = StdFileDialog.choose(this, true, MSG_SELECT_HTML_TEMPLATE, null, null, "html", "htm"); //$NON-NLS-1$ //$NON-NLS-2$
			if (file != null) {
				mHTMLTemplatePath.setText(Path.getFullPath(file));
			}
		}
		adjustResetButton();
	}

	@Override public void reset() {
		mPlayerName.setText(System.getProperty("user.name")); //$NON-NLS-1$
		mCampaign.setText(""); //$NON-NLS-1$
		mTechLevel.setText(Profile.DEFAULT_TECH_LEVEL);
		setPortrait(Profile.DEFAULT_PORTRAIT);
		for (int i = 0; i < DPI.length; i++) {
			if (DPI[i] == DEFAULT_PNG_RESOLUTION) {
				mPNGResolutionCombo.setSelectedIndex(i);
				break;
			}
		}
		mUseHTMLTemplateOverride.setSelected(false);
		mUseOptionalDiceRules.setSelected(DEFAULT_OPTIONAL_DICE_RULES);
	}

	@Override public boolean isSetToDefaults() {
		return Profile.getDefaultPlayerName().equals(System.getProperty("user.name")) && Profile.getDefaultCampaign().equals("") && Profile.getDefaultPortraitPath().equals(Profile.DEFAULT_PORTRAIT) && Profile.getDefaultTechLevel().equals(Profile.DEFAULT_TECH_LEVEL) && getPNGResolution() == DEFAULT_PNG_RESOLUTION && isHTMLTemplateOverridden() == false && areOptionalDiceRulesUsed() == DEFAULT_OPTIONAL_DICE_RULES; //$NON-NLS-1$ //$NON-NLS-2$
	}

	private void setPortrait(String path) {
		BufferedImage image = Profile.getPortraitFromPortraitPath(path);
		Profile.setDefaultPortraitPath(path);
		if (image != null) {
			image = Images.scale(image, Profile.PORTRAIT_WIDTH, Profile.PORTRAIT_HEIGHT);
		}
		mPortrait.setPortrait(image);
	}

	public void changedUpdate(DocumentEvent event) {
		Document document = event.getDocument();
		if (mPlayerName.getDocument() == document) {
			Profile.setDefaultPlayerName(mPlayerName.getText());
		} else if (mCampaign.getDocument() == document) {
			Profile.setDefaultCampaign(mCampaign.getText());
		} else if (mTechLevel.getDocument() == document) {
			Profile.setDefaultTechLevel(mTechLevel.getText());
		} else if (mHTMLTemplatePath.getDocument() == document) {
			Preferences.getInstance().setValue(MODULE, HTML_TEMPLATE_OVERRIDE, mHTMLTemplatePath.getText());
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
		if (source == mUseHTMLTemplateOverride) {
			boolean checked = mUseHTMLTemplateOverride.isSelected();
			Preferences.getInstance().setValue(MODULE, USE_HTML_TEMPLATE_OVERRIDE, checked);
			mHTMLTemplatePath.setEnabled(checked);
			mHTMLTemplatePicker.setEnabled(checked);
			mHTMLTemplatePath.setText(getHTMLTemplate());
		} else if (source == mUseOptionalDiceRules) {
			boolean checked = mUseOptionalDiceRules.isSelected();
			adjustOptionalDiceRulesProperty(checked);
			Preferences.getInstance().setValue(MODULE, OPTIONAL_DICE_RULES, checked);
		}
		adjustResetButton();
	}
}
