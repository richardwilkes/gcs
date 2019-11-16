/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.app.GCSCmdLine;
import com.trollworks.gcs.app.GCSImages;
import com.trollworks.toolkit.io.Log;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.FlexColumn;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.preferences.PreferencePanel;
import com.trollworks.toolkit.ui.preferences.PreferencesWindow;
import com.trollworks.toolkit.ui.print.PageOrientation;
import com.trollworks.toolkit.ui.print.PrintManager;
import com.trollworks.toolkit.ui.widget.StdFileDialog;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.text.Text;
import com.trollworks.toolkit.utility.units.LengthUnits;

import java.awt.Color;
import java.awt.Desktop;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.IOException;
import java.io.StringReader;
import java.net.URI;
import java.nio.charset.StandardCharsets;
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
public class OutputPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
    static final String         MODULE                    = "Output";
    private static final int    DEFAULT_PNG_RESOLUTION    = 200;
    private static final String PNG_RESOLUTION_KEY        = "PNGResolution";
    private static final int[]  DPI                       = { 72, 96, 144, 150, 200, 300 };
    private static final String USE_TEMPLATE_OVERRIDE_KEY = "UseTextTemplateOverride";
    private static final String TEMPLATE_OVERRIDE_KEY     = "TextTemplateOverride";
    private static final String GURPS_CALCULATOR_KEY_KEY  = "GurpsCalculatorKey";
    public static final String  BASE_GURPS_CALCULATOR_URL = "http://www.gurpscalculator.com";
    public static final String  GURPS_CALCULATOR_URL      = BASE_GURPS_CALCULATOR_URL + "/Character/ImportGCS";
    private static final String DEFAULT_PAGE_SETTINGS_KEY = "DefaultPageSettings";
    private JComboBox<String>   mPNGResolutionCombo;
    private JCheckBox           mUseTextTemplateOverride;
    private JTextField          mTextTemplatePath;
    private JButton             mTextTemplatePicker;
    private JButton             mGurpsCalculatorLink;
    private JTextField          mGurpsCalculatorKey;
    private JCheckBox           mUseNativePrinter;

    /** Initializes the services controlled by these preferences. */
    public static void initialize() {
        // Converting preferences that used to live in SheetPreferences
        Preferences prefs = Preferences.getInstance();
        for (String key : prefs.getModuleKeys(SheetPreferences.MODULE)) {
            if (PNG_RESOLUTION_KEY.equals(key)) {
                prefs.setValue(MODULE, PNG_RESOLUTION_KEY, prefs.getIntValue(SheetPreferences.MODULE, PNG_RESOLUTION_KEY, DEFAULT_PNG_RESOLUTION));
                prefs.removePreference(SheetPreferences.MODULE, PNG_RESOLUTION_KEY);
            } else if (USE_TEMPLATE_OVERRIDE_KEY.equals(key)) {
                prefs.setValue(MODULE, USE_TEMPLATE_OVERRIDE_KEY, prefs.getBooleanValue(SheetPreferences.MODULE, USE_TEMPLATE_OVERRIDE_KEY));
                prefs.removePreference(SheetPreferences.MODULE, USE_TEMPLATE_OVERRIDE_KEY);
            } else if (TEMPLATE_OVERRIDE_KEY.equals(key)) {
                prefs.setValue(MODULE, TEMPLATE_OVERRIDE_KEY, prefs.getStringValue(SheetPreferences.MODULE, TEMPLATE_OVERRIDE_KEY));
                prefs.removePreference(SheetPreferences.MODULE, TEMPLATE_OVERRIDE_KEY);
            } else if (GURPS_CALCULATOR_KEY_KEY.equals(key)) {
                prefs.setValue(MODULE, GURPS_CALCULATOR_KEY_KEY, prefs.getStringValue(SheetPreferences.MODULE, GURPS_CALCULATOR_KEY_KEY));
                prefs.removePreference(SheetPreferences.MODULE, GURPS_CALCULATOR_KEY_KEY);
            } else if (DEFAULT_PAGE_SETTINGS_KEY.equals(key)) {
                prefs.setValue(MODULE, DEFAULT_PAGE_SETTINGS_KEY, prefs.getStringValue(SheetPreferences.MODULE, DEFAULT_PAGE_SETTINGS_KEY));
                prefs.removePreference(SheetPreferences.MODULE, DEFAULT_PAGE_SETTINGS_KEY);
            }
        }
    }

    /** @return The resolution to use when saving the sheet as a PNG. */
    public static int getPNGResolution() {
        return Preferences.getInstance().getIntValue(MODULE, PNG_RESOLUTION_KEY, DEFAULT_PNG_RESOLUTION);
    }

    /** @return Whether the default text template has been overridden. */
    public static boolean isTextTemplateOverridden() {
        return Preferences.getInstance().getBooleanValue(MODULE, USE_TEMPLATE_OVERRIDE_KEY);
    }

    /** @return The text template to use when exporting to a text format. */
    public static String getTextTemplate() {
        return isTextTemplateOverridden() ? getTextTemplateOverride() : getDefaultTextTemplate();
    }

    private static String getTextTemplateOverride() {
        return Preferences.getInstance().getStringValue(MODULE, TEMPLATE_OVERRIDE_KEY);
    }

    /** @return The default text template to use when exporting to a text format. */
    public static String getDefaultTextTemplate() {
        return GCSCmdLine.getLibraryRootPath().resolve("Output Templates").resolve("html_template.html").toString();
    }

    public static String getGurpsCalculatorKey() {
        return Preferences.getInstance().getStringValue(MODULE, GURPS_CALCULATOR_KEY_KEY);
    }

    /**
     * @return The default page settings to use. May return <code>null</code> if no printer has been
     *         defined.
     */
    public static PrintManager getDefaultPageSettings() {
        String settings = Preferences.getInstance().getStringValue(MODULE, DEFAULT_PAGE_SETTINGS_KEY);
        if (settings != null && !settings.isEmpty()) {
            try (XMLReader in = new XMLReader(new StringReader(settings))) {
                XMLNodeType type  = in.next();
                boolean     found = false;
                while (type != XMLNodeType.END_DOCUMENT) {
                    if (type == XMLNodeType.START_TAG) {
                        String name = in.getName();
                        if (PrintManager.TAG_ROOT.equals(name)) {
                            if (!found) {
                                found = true;
                                return new PrintManager(in);
                            }
                            throw new IOException();
                        }
                        in.skipTag(name);
                        type = in.getType();
                    } else {
                        type = in.next();
                    }
                }
            } catch (Exception exception) {
                Log.error(exception);
            }
        }
        try {
            return new PrintManager(PageOrientation.PORTRAIT, 0.5, LengthUnits.IN);
        } catch (Exception exception) {
            return null;
        }
    }

    /** @param mgr The {@Link PrintManager} to record the settings for. */
    public static void setDefaultPageSettings(PrintManager mgr) {
        String value = null;
        if (mgr != null) {
            ByteArrayOutputStream baos = new ByteArrayOutputStream();
            try (XMLWriter out = new XMLWriter(baos)) {
                out.writeHeader();
                mgr.save(out, LengthUnits.IN);
            } catch (Exception exception) {
                Log.error(exception);
            }
            if (baos.size() > 0) {
                value = new String(baos.toByteArray(), StandardCharsets.UTF_8);
            }
        }
        Preferences.getInstance().setValue(MODULE, DEFAULT_PAGE_SETTINGS_KEY, value);
    }

    /**
     * Creates a new {@link OutputPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public OutputPreferences(PreferencesWindow owner) {
        super(I18n.Text("Output"), owner);
        FlexColumn column = new FlexColumn();

        FlexGrid   grid   = new FlexGrid();
        column.add(grid);

        FlexRow row        = new FlexRow();
        String  gcalcTitle = I18n.Text("GURPS Calculator Key");
        row.add(createLabel(gcalcTitle, gcalcTitle, GCSImages.getGCalcLogo()));
        mGurpsCalculatorKey = createTextField(gcalcTitle, getGurpsCalculatorKey());
        row.add(mGurpsCalculatorKey);
        mGurpsCalculatorLink = createHyperlinkButton(I18n.Text("Find mine"), GURPS_CALCULATOR_URL);
        if (Desktop.isDesktopSupported()) {
            row.add(mGurpsCalculatorLink);
        }
        column.add(row);

        mUseNativePrinter = createCheckBox(I18n.Text("Use platform native print dialogs (settings cannot be saved)"), I18n.Text("<html><body>Whether or not the native print dialogs should be used.<br>Choosing this option will prevent the program from saving<br>and restoring print settings with the document.</body></html>"), PrintManager.useNativeDialogs());
        column.add(mUseNativePrinter);

        row                      = new FlexRow();
        mUseTextTemplateOverride = createCheckBox(I18n.Text("Text Export Template"), textTemplateOverrideTooltip(), isTextTemplateOverridden());
        row.add(mUseTextTemplateOverride);
        mTextTemplatePath = createTextTemplatePathField();
        row.add(mTextTemplatePath);
        mTextTemplatePicker = createButton(I18n.Text("Choose\u2026"), textTemplateOverrideTooltip());
        mTextTemplatePicker.setEnabled(isTextTemplateOverridden());
        row.add(mTextTemplatePicker);
        column.add(row);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), pngDPIMsg()));
        mPNGResolutionCombo = createPNGResolutionPopup();
        row.add(mPNGResolutionCombo);
        row.add(createLabel(I18n.Text("when saving sheets to PNG"), pngDPIMsg(), SwingConstants.LEFT));
        column.add(row);

        column.add(new FlexSpacer(0, 0, false, true));

        column.apply(this);
    }

    private static final String textTemplateOverrideTooltip() {
        return I18n.Text("Specify a file to use as the template when exporting to a text format, such as HTML");
    }

    private static final String pngDPIMsg() {
        return I18n.Text("The resolution, in dots-per-inch, to use when saving sheets as PNG files");
    }

    private JButton createButton(String title, String tooltip) {
        JButton button = new JButton(title);
        button.setOpaque(false);
        button.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        button.addActionListener(this);
        add(button);
        return button;
    }

    private JTextField createTextTemplatePathField() {
        JTextField field = new JTextField(getTextTemplate());
        field.setToolTipText(Text.wrapPlainTextForToolTip(textTemplateOverrideTooltip()));
        field.setEnabled(isTextTemplateOverridden());
        field.getDocument().addDocumentListener(this);
        Dimension size    = field.getPreferredSize();
        Dimension maxSize = field.getMaximumSize();
        maxSize.height = size.height;
        field.setMaximumSize(maxSize);
        add(field);
        return field;
    }

    private JButton createHyperlinkButton(String linkText, String tooltip) {
        JButton button = new JButton(String.format("<html><body><font color=\"#000099\"><u>%s</u></font></body></html>", linkText));
        button.setFocusPainted(false);
        button.setMargin(new Insets(0, 0, 0, 0));
        button.setContentAreaFilled(false);
        button.setBorderPainted(false);
        button.setOpaque(false);
        button.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        button.setBackground(Color.white);
        button.addActionListener(this);
        UIUtilities.setOnlySize(button, button.getPreferredSize());
        add(button);
        return button;
    }

    private JComboBox<String> createPNGResolutionPopup() {
        int               selection  = 0;
        int               resolution = getPNGResolution();
        JComboBox<String> combo      = new JComboBox<>();
        setupCombo(combo, pngDPIMsg());
        for (int i = 0; i < DPI.length; i++) {
            combo.addItem(MessageFormat.format(I18n.Text("{0} dpi"), Integer.valueOf(DPI[i])));
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
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        Dimension size    = field.getPreferredSize();
        Dimension maxSize = field.getMaximumSize();
        maxSize.height = size.height;
        field.setMaximumSize(maxSize);
        add(field);
        return field;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object source = event.getSource();
        if (source == mPNGResolutionCombo) {
            Preferences.getInstance().setValue(MODULE, PNG_RESOLUTION_KEY, DPI[mPNGResolutionCombo.getSelectedIndex()]);
        } else if (source == mTextTemplatePicker) {
            File file = StdFileDialog.showOpenDialog(this, I18n.Text("Select A Text Template"));
            if (file != null) {
                mTextTemplatePath.setText(PathUtils.getFullPath(file));
            }
        } else if (source == mGurpsCalculatorLink && Desktop.isDesktopSupported()) {
            try {
                Desktop.getDesktop().browse(new URI(GURPS_CALCULATOR_URL));
            } catch (Exception exception) {
                WindowUtils.showError(this, MessageFormat.format(I18n.Text("Unable to open {0}"), GURPS_CALCULATOR_URL));
            }
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mGurpsCalculatorKey.setText("");
        for (int i = 0; i < DPI.length; i++) {
            if (DPI[i] == DEFAULT_PNG_RESOLUTION) {
                mPNGResolutionCombo.setSelectedIndex(i);
                break;
            }
        }
        mUseTextTemplateOverride.setSelected(false);
        mUseNativePrinter.setSelected(false);
    }

    @Override
    public boolean isSetToDefaults() {
        return getPNGResolution() == DEFAULT_PNG_RESOLUTION && isTextTemplateOverridden() == false && !PrintManager.useNativeDialogs() && mGurpsCalculatorKey.getText().equals("");
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document document = event.getDocument();
        if (mTextTemplatePath.getDocument() == document) {
            Preferences.getInstance().setValue(MODULE, TEMPLATE_OVERRIDE_KEY, mTextTemplatePath.getText());
        } else if (mGurpsCalculatorKey.getDocument() == document) {
            Preferences.getInstance().setValue(MODULE, GURPS_CALCULATOR_KEY_KEY, mGurpsCalculatorKey.getText());
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
        if (source == mUseTextTemplateOverride) {
            boolean checked = mUseTextTemplateOverride.isSelected();
            Preferences.getInstance().setValue(MODULE, USE_TEMPLATE_OVERRIDE_KEY, checked);
            mTextTemplatePath.setEnabled(checked);
            mTextTemplatePicker.setEnabled(checked);
            mTextTemplatePath.setText(getTextTemplate());
        } else if (source == mUseNativePrinter) {
            PrintManager.useNativeDialogs(mUseNativePrinter.isSelected());
        }
        adjustResetButton();
    }
}
