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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.preferences;

import static com.trollworks.gcs.preferences.SheetPreferences_LS.*;

import com.trollworks.gcs.character.Profile;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.image.Images;
import com.trollworks.ttk.layout.Alignment;
import com.trollworks.ttk.layout.FlexColumn;
import com.trollworks.ttk.layout.FlexComponent;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;
import com.trollworks.ttk.preferences.PreferencePanel;
import com.trollworks.ttk.preferences.Preferences;
import com.trollworks.ttk.preferences.PreferencesWindow;
import com.trollworks.ttk.print.PrintManager;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.units.LengthUnits;
import com.trollworks.ttk.units.WeightUnits;
import com.trollworks.ttk.utility.App;
import com.trollworks.ttk.utility.Dice;
import com.trollworks.ttk.utility.Path;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.StdFileDialog;

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

@Localized({
				@LS(key = "SHEET", msg = "Sheet"),
				@LS(key = "PLAYER", msg = "Player"),
				@LS(key = "PLAYER_TOOLTIP", msg = "The player name to use when a new character sheet is created"),
				@LS(key = "CAMPAIGN", msg = "Campaign"),
				@LS(key = "CAMPAIGN_TOOLTIP", msg = "The campaign to use when a new character sheet is created"),
				@LS(key = "TECH_LEVEL", msg = "Tech Level"),
				@LS(key = "TECH_LEVEL_TOOLTIP", msg = "<html><body>TL0: Stone Age<br>TL1: Bronze Age<br>TL2: Iron Age<br>TL3: Medieval<br>TL4: Age of Sail<br>TL5: Industrial Revolution<br>TL6: Mechanized Age<br>TL7: Nuclear Age<br>TL8: Digital Age<br>TL9: Microtech Age<br>TL10: Robotic Age<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>"),
				@LS(key = "INITIAL_POINTS", msg = "Initial Points"),
				@LS(key = "INITIAL_POINTS_TOOLTIP", msg = "The initial number of character points to start with"),
				@LS(key = "SELECT_PORTRAIT", msg = "Select A Portrait"),
				@LS(key = "OPTIONAL_IQ_RULES", msg = "Use optional (house) rule: Will and Perception are not based upon IQ"),
				@LS(key = "OPTIONAL_MODIFIER_RULES", msg = "Use optional rule \"Multiplicative Modifiers\" from PW102 (note: changes point value)"),
				@LS(key = "OPTIONAL_DICE_RULES", msg = "Use optional rule \"Modifying Dice + Adds\" from B269"),
				@LS(key = "PNG_RESOLUTION_POST", msg = "when saving sheets to PNG"),
				@LS(key = "PNG_RESOLUTION_TOOLTIP", msg = "The resolution, in dots-per-inch, to use when saving sheets as PNG files"),
				@LS(key = "DPI_FORMAT", msg = "{0} dpi"),
				@LS(key = "HTML_TEMPLATE_OVERRIDE", msg = "HTML Template Override"),
				@LS(key = "HTML_TEMPLATE_PICKER", msg = "Choose..."),
				@LS(key = "HTML_TEMPLATE_OVERRIDE_TOOLTIP", msg = "Specify a file to use as the template when exporting to HTML"),
				@LS(key = "SELECT_HTML_TEMPLATE", msg = "Select A HTML Template"),
				@LS(key = "NATIVE_PRINTER", msg = "Use platform native print dialogs (settings cannot be saved)"),
				@LS(key = "NATIVE_PRINTER_TOOLTIP", msg = "<html><body>Whether or not the native print dialogs should be used.<br>Choosing this option will prevent the program from saving<br>and restoring print settings with the document.</body></html>"),
				@LS(key = "AUTO_NAME", msg = "Automatically name new characters"),
				@LS(key = "LENGTH_UNITS_TOOLTIP", msg = "The units to use for display of generated lengths"),
				@LS(key = "WEIGHT_UNITS_TOOLTIP", msg = "The units to use for display of generated weights"),
				@LS(key = "USE", msg = "Use"),
				@LS(key = "AND", msg = "and"),
				@LS(key = "FOR_UNIT_DISPLAY", msg = "for display of generated units"),
})
/** The sheet preferences panel. */
public class SheetPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
	private static final String			MODULE								= "Sheet";														//$NON-NLS-1$
	private static final String			OPTIONAL_DICE_RULES					= "UseOptionDiceRules";										//$NON-NLS-1$
	/** The optional dice rules preference key. */
	public static final String			OPTIONAL_DICE_RULES_PREF_KEY		= Preferences.getModuleKey(MODULE, OPTIONAL_DICE_RULES);
	private static final boolean		DEFAULT_OPTIONAL_DICE_RULES			= false;
	private static final String			OPTIONAL_IQ_RULES					= "UseOptionIQRules";											//$NON-NLS-1$
	/** The optional IQ rules preference key. */
	public static final String			OPTIONAL_IQ_RULES_PREF_KEY			= Preferences.getModuleKey(MODULE, OPTIONAL_IQ_RULES);
	private static final boolean		DEFAULT_OPTIONAL_IQ_RULES			= false;
	private static final String			OPTIONAL_MODIFIER_RULES				= "UseOptionModifierRules";									//$NON-NLS-1$
	/** The optional modifier rules preference key. */
	public static final String			OPTIONAL_MODIFIER_RULES_PREF_KEY	= Preferences.getModuleKey(MODULE, OPTIONAL_MODIFIER_RULES);
	private static final boolean		DEFAULT_OPTIONAL_MODIFIER_RULES		= false;
	private static final String			AUTO_NAME							= "AutoNameNewCharacters";										//$NON-NLS-1$
	/** The auto-naming preference key. */
	public static final String			AUTO_NAME_PREF_KEY					= Preferences.getModuleKey(MODULE, AUTO_NAME);
	private static final boolean		DEFAULT_AUTO_NAME					= true;
	private static final String			LENGTH_UNITS						= "LengthUnits";												//$NON-NLS-1$
	/** The default length units preference key. */
	public static final String			LENGTH_UNITS_PREF_KEY				= Preferences.getModuleKey(MODULE, LENGTH_UNITS);
	private static final LengthUnits	DEFAULT_LENGTH_UNITS				= LengthUnits.FEET_AND_INCHES;
	private static final String			WEIGHT_UNITS						= "WeightUnits";												//$NON-NLS-1$
	/** The default weight units preference key. */
	public static final String			WEIGHT_UNITS_PREF_KEY				= Preferences.getModuleKey(MODULE, WEIGHT_UNITS);
	private static final WeightUnits	DEFAULT_WEIGHT_UNITS				= WeightUnits.POUNDS;
	private static final int			DEFAULT_PNG_RESOLUTION				= 200;
	private static final String			PNG_RESOLUTION						= "PNGResolution";												//$NON-NLS-1$
	private static final int[]			DPI									= { 72, 96, 144, 150, 200, 300 };
	private static final String			USE_HTML_TEMPLATE_OVERRIDE			= "UseHTMLTemplateOverride";									//$NON-NLS-1$
	private static final String			HTML_TEMPLATE_OVERRIDE				= "HTMLTemplateOverride";										//$NON-NLS-1$
	private static final String			INITIAL_POINTS_KEY					= "InitialPoints";												//$NON-NLS-1$
	private static final int			DEFAULT_INITIAL_POINTS				= 100;
	private JTextField					mPlayerName;
	private JTextField					mCampaign;
	private JTextField					mTechLevel;
	private JTextField					mInitialPoints;
	private PortraitPreferencePanel		mPortrait;
	private JComboBox<String>			mPNGResolutionCombo;
	private JComboBox<String>			mLengthUnitsCombo;
	private JComboBox<String>			mWeightUnitsCombo;
	private JCheckBox					mUseHTMLTemplateOverride;
	private JTextField					mHTMLTemplatePath;
	private JButton						mHTMLTemplatePicker;
	private JCheckBox					mUseOptionalDiceRules;
	private JCheckBox					mUseOptionalIQRules;
	private JCheckBox					mUseOptionalModifierRules;
	private JCheckBox					mAutoName;
	private JCheckBox					mUseNativePrinter;

	/** Initializes the services controlled by these preferences. */
	public static void initialize() {
		adjustOptionalDiceRulesProperty(areOptionalDiceRulesUsed());
	}

	/** @return The default length units to use. */
	public static LengthUnits getLengthUnits() {
		try {
			return LengthUnits.valueOf(Preferences.getInstance().getStringValue(MODULE, LENGTH_UNITS, DEFAULT_LENGTH_UNITS.name()));
		} catch (Exception exception) {
			return DEFAULT_LENGTH_UNITS;
		}
	}

	/** @return The default weight units to use. */
	public static WeightUnits getWeightUnits() {
		try {
			return WeightUnits.valueOf(Preferences.getInstance().getStringValue(MODULE, WEIGHT_UNITS, DEFAULT_WEIGHT_UNITS.name()));
		} catch (Exception exception) {
			return DEFAULT_WEIGHT_UNITS;
		}
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

	/** @return Whether the optional IQ rules (Will &amp; Perception are not based on IQ) are in use. */
	public static boolean areOptionalIQRulesUsed() {
		return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_IQ_RULES, DEFAULT_OPTIONAL_IQ_RULES);
	}

	/** @return Whether the optional modifier rules from PW102 are in use. */
	public static boolean areOptionalModifierRulesUsed() {
		return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_MODIFIER_RULES, DEFAULT_OPTIONAL_MODIFIER_RULES);
	}

	/** @return Whether a new character should be automatically named. */
	public static boolean isNewCharacterAutoNamed() {
		return Preferences.getInstance().getBooleanValue(MODULE, AUTO_NAME, DEFAULT_AUTO_NAME);
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
		return Path.getFullPath(new File(new File(App.APP_HOME_DIR, "data"), "template.html")); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/** @return The initial points to start a new character with. */
	public static int getInitialPoints() {
		return Preferences.getInstance().getIntValue(MODULE, INITIAL_POINTS_KEY, DEFAULT_INITIAL_POINTS);
	}

	/**
	 * Creates a new {@link SheetPreferences}.
	 * 
	 * @param owner The owning {@link PreferencesWindow}.
	 */
	public SheetPreferences(PreferencesWindow owner) {
		super(SHEET, owner);
		FlexColumn column = new FlexColumn();

		FlexGrid grid = new FlexGrid();
		column.add(grid);

		int rowIndex = 0;
		mPortrait = createPortrait();
		FlexComponent comp = new FlexComponent(mPortrait, Alignment.LEFT_TOP, Alignment.LEFT_TOP);
		grid.add(comp, rowIndex, 0, 4, 1);

		grid.add(createFlexLabel(PLAYER, PLAYER_TOOLTIP), rowIndex, 1);
		mPlayerName = createTextField(PLAYER_TOOLTIP, Profile.getDefaultPlayerName());
		grid.add(mPlayerName, rowIndex++, 2);

		grid.add(createFlexLabel(CAMPAIGN, CAMPAIGN_TOOLTIP), rowIndex, 1);
		mCampaign = createTextField(CAMPAIGN_TOOLTIP, Profile.getDefaultCampaign());
		grid.add(mCampaign, rowIndex++, 2);

		grid.add(createFlexLabel(TECH_LEVEL, TECH_LEVEL_TOOLTIP), rowIndex, 1);
		mTechLevel = createTextField(TECH_LEVEL_TOOLTIP, Profile.getDefaultTechLevel());
		grid.add(mTechLevel, rowIndex++, 2);

		grid.add(createFlexLabel(INITIAL_POINTS, INITIAL_POINTS_TOOLTIP), rowIndex, 1);
		mInitialPoints = createTextField(INITIAL_POINTS_TOOLTIP, Integer.toString(getInitialPoints()));
		grid.add(mInitialPoints, rowIndex++, 2);

		grid.add(new FlexSpacer(0, 0, false, true), rowIndex, 1);
		grid.add(new FlexSpacer(0, 0, true, true), rowIndex, 2);

		addSeparator(column);

		FlexRow row = new FlexRow();
		row.add(createLabel(USE, null));
		mLengthUnitsCombo = createLengthUnitsPopup();
		row.add(mLengthUnitsCombo);
		row.add(createLabel(AND, null));
		mWeightUnitsCombo = createWeightUnitsPopup();
		row.add(mWeightUnitsCombo);
		row.add(createLabel(FOR_UNIT_DISPLAY, null));
		column.add(row);

		mAutoName = createCheckBox(AUTO_NAME, null, isNewCharacterAutoNamed());
		column.add(mAutoName);

		mUseOptionalIQRules = createCheckBox(OPTIONAL_IQ_RULES, null, areOptionalIQRulesUsed());
		column.add(mUseOptionalIQRules);

		mUseOptionalModifierRules = createCheckBox(OPTIONAL_MODIFIER_RULES, null, areOptionalModifierRulesUsed());
		column.add(mUseOptionalModifierRules);

		mUseOptionalDiceRules = createCheckBox(OPTIONAL_DICE_RULES, null, areOptionalDiceRulesUsed());
		column.add(mUseOptionalDiceRules);

		row = new FlexRow();
		mUseHTMLTemplateOverride = createCheckBox(HTML_TEMPLATE_OVERRIDE, HTML_TEMPLATE_OVERRIDE_TOOLTIP, isHTMLTemplateOverridden());
		row.add(mUseHTMLTemplateOverride);
		mHTMLTemplatePath = createHTMLTemplatePathField();
		row.add(mHTMLTemplatePath);
		mHTMLTemplatePicker = createButton(HTML_TEMPLATE_PICKER, HTML_TEMPLATE_OVERRIDE_TOOLTIP);
		mHTMLTemplatePicker.setEnabled(isHTMLTemplateOverridden());
		row.add(mHTMLTemplatePicker);
		column.add(row);

		row = new FlexRow();
		row.add(createLabel(USE, PNG_RESOLUTION_TOOLTIP));
		mPNGResolutionCombo = createPNGResolutionPopup();
		row.add(mPNGResolutionCombo);
		row.add(createLabel(PNG_RESOLUTION_POST, PNG_RESOLUTION_TOOLTIP, SwingConstants.LEFT));
		column.add(row);

		mUseNativePrinter = createCheckBox(NATIVE_PRINTER, NATIVE_PRINTER_TOOLTIP, PrintManager.useNativeDialogs());
		column.add(mUseNativePrinter);

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
		field.setToolTipText(HTML_TEMPLATE_OVERRIDE_TOOLTIP);
		field.setEnabled(isHTMLTemplateOverridden());
		field.getDocument().addDocumentListener(this);
		Dimension size = field.getPreferredSize();
		Dimension maxSize = field.getMaximumSize();
		maxSize.height = size.height;
		field.setMaximumSize(maxSize);
		add(field);
		return field;
	}

	private JComboBox<String> createPNGResolutionPopup() {
		int selection = 0;
		int resolution = getPNGResolution();
		JComboBox<String> combo = new JComboBox<>();
		setupCombo(combo, PNG_RESOLUTION_TOOLTIP);
		for (int i = 0; i < DPI.length; i++) {
			combo.addItem(MessageFormat.format(DPI_FORMAT, new Integer(DPI[i])));
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

	private JComboBox<String> createLengthUnitsPopup() {
		JComboBox<String> combo = new JComboBox<>();
		setupCombo(combo, LENGTH_UNITS_TOOLTIP);
		for (LengthUnits unit : LengthUnits.values()) {
			combo.addItem(unit.getDescription());
		}
		combo.setSelectedIndex(getLengthUnits().ordinal());
		combo.addActionListener(this);
		combo.setMaximumRowCount(combo.getItemCount());
		UIUtilities.setOnlySize(combo, combo.getPreferredSize());
		return combo;
	}

	private JComboBox<String> createWeightUnitsPopup() {
		JComboBox<String> combo = new JComboBox<>();
		setupCombo(combo, WEIGHT_UNITS_TOOLTIP);
		for (WeightUnits unit : WeightUnits.values()) {
			combo.addItem(unit.getDescription());
		}
		combo.setSelectedIndex(getWeightUnits().ordinal());
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

	@Override
	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();
		if (source == mPortrait) {
			File file = StdFileDialog.choose(this, true, SELECT_PORTRAIT, null, null, "png", "jpg", "gif", "jpeg"); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
			if (file != null) {
				setPortrait(Path.getFullPath(file));
			}
		} else if (source == mPNGResolutionCombo) {
			Preferences.getInstance().setValue(MODULE, PNG_RESOLUTION, DPI[mPNGResolutionCombo.getSelectedIndex()]);
		} else if (source == mLengthUnitsCombo) {
			Preferences.getInstance().setValue(MODULE, LENGTH_UNITS, LengthUnits.values()[mLengthUnitsCombo.getSelectedIndex()].name());
		} else if (source == mWeightUnitsCombo) {
			Preferences.getInstance().setValue(MODULE, WEIGHT_UNITS, WeightUnits.values()[mWeightUnitsCombo.getSelectedIndex()].name());
		} else if (source == mHTMLTemplatePicker) {
			File file = StdFileDialog.choose(this, true, SELECT_HTML_TEMPLATE, null, null, "html", "htm"); //$NON-NLS-1$ //$NON-NLS-2$
			if (file != null) {
				mHTMLTemplatePath.setText(Path.getFullPath(file));
			}
		}
		adjustResetButton();
	}

	@Override
	public void reset() {
		mPlayerName.setText(System.getProperty("user.name")); //$NON-NLS-1$
		mCampaign.setText(""); //$NON-NLS-1$
		mTechLevel.setText(Profile.DEFAULT_TECH_LEVEL);
		mInitialPoints.setText(Integer.toString(DEFAULT_INITIAL_POINTS));
		setPortrait(Profile.DEFAULT_PORTRAIT);
		for (int i = 0; i < DPI.length; i++) {
			if (DPI[i] == DEFAULT_PNG_RESOLUTION) {
				mPNGResolutionCombo.setSelectedIndex(i);
				break;
			}
		}
		mLengthUnitsCombo.setSelectedIndex(DEFAULT_LENGTH_UNITS.ordinal());
		mWeightUnitsCombo.setSelectedIndex(DEFAULT_WEIGHT_UNITS.ordinal());
		mUseHTMLTemplateOverride.setSelected(false);
		mAutoName.setSelected(DEFAULT_AUTO_NAME);
		mUseOptionalDiceRules.setSelected(DEFAULT_OPTIONAL_DICE_RULES);
		mUseOptionalIQRules.setSelected(DEFAULT_OPTIONAL_IQ_RULES);
		mUseOptionalModifierRules.setSelected(DEFAULT_OPTIONAL_MODIFIER_RULES);
		mUseNativePrinter.setSelected(false);
	}

	@Override
	public boolean isSetToDefaults() {
		return Profile.getDefaultPlayerName().equals(System.getProperty("user.name")) && Profile.getDefaultCampaign().equals("") && Profile.getDefaultPortraitPath().equals(Profile.DEFAULT_PORTRAIT) && Profile.getDefaultTechLevel().equals(Profile.DEFAULT_TECH_LEVEL) && getInitialPoints() == DEFAULT_INITIAL_POINTS && getPNGResolution() == DEFAULT_PNG_RESOLUTION && isHTMLTemplateOverridden() == false && areOptionalDiceRulesUsed() == DEFAULT_OPTIONAL_DICE_RULES && areOptionalIQRulesUsed() == DEFAULT_OPTIONAL_IQ_RULES && areOptionalModifierRulesUsed() == DEFAULT_OPTIONAL_MODIFIER_RULES && isNewCharacterAutoNamed() == DEFAULT_AUTO_NAME && !PrintManager.useNativeDialogs(); //$NON-NLS-1$ //$NON-NLS-2$
	}

	private void setPortrait(String path) {
		BufferedImage image = Profile.getPortraitFromPortraitPath(path);
		Profile.setDefaultPortraitPath(path);
		if (image != null) {
			image = Images.scale(image, Profile.PORTRAIT_WIDTH, Profile.PORTRAIT_HEIGHT);
		}
		mPortrait.setPortrait(image);
	}

	@Override
	public void changedUpdate(DocumentEvent event) {
		Document document = event.getDocument();
		if (mPlayerName.getDocument() == document) {
			Profile.setDefaultPlayerName(mPlayerName.getText());
		} else if (mCampaign.getDocument() == document) {
			Profile.setDefaultCampaign(mCampaign.getText());
		} else if (mTechLevel.getDocument() == document) {
			Profile.setDefaultTechLevel(mTechLevel.getText());
		} else if (mInitialPoints.getDocument() == document) {
			Preferences.getInstance().setValue(MODULE, INITIAL_POINTS_KEY, Numbers.getLocalizedInteger(mInitialPoints.getText(), 0));
		} else if (mHTMLTemplatePath.getDocument() == document) {
			Preferences.getInstance().setValue(MODULE, HTML_TEMPLATE_OVERRIDE, mHTMLTemplatePath.getText());
		}
		adjustResetButton();
	}

	@Override
	public void insertUpdate(DocumentEvent event) {
		changedUpdate(event);
	}

	@Override
	public void removeUpdate(DocumentEvent event) {
		changedUpdate(event);
	}

	@Override
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
		} else if (source == mUseOptionalIQRules) {
			Preferences.getInstance().setValue(MODULE, OPTIONAL_IQ_RULES, mUseOptionalIQRules.isSelected());
		} else if (source == mUseOptionalModifierRules) {
			Preferences.getInstance().setValue(MODULE, OPTIONAL_MODIFIER_RULES, mUseOptionalModifierRules.isSelected());
		} else if (source == mAutoName) {
			Preferences.getInstance().setValue(MODULE, AUTO_NAME, mAutoName.isSelected());
		} else if (source == mUseNativePrinter) {
			PrintManager.useNativeDialogs(mUseNativePrinter.isSelected());
		}
		adjustResetButton();
	}
}
