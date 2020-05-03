/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.io.Log;
import com.trollworks.gcs.io.xml.XMLNodeType;
import com.trollworks.gcs.io.xml.XMLReader;
import com.trollworks.gcs.io.xml.XMLWriter;
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.FlexColumn;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.print.PageOrientation;
import com.trollworks.gcs.ui.print.PrintManager;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Preferences;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.units.LengthUnits;

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
    static final         String            MODULE                    = "Output";
    private static final int               DEFAULT_PNG_RESOLUTION    = 200;
    private static final String            PNG_RESOLUTION_KEY        = "PNGResolution";
    private static final int[]             DPI                       = {72, 96, 144, 150, 200, 300};
    private static final String            GURPS_CALCULATOR_KEY_KEY  = "GurpsCalculatorKey";
    public static final  String            BASE_GURPS_CALCULATOR_URL = "http://www.gurpscalculator.com";
    public static final  String            GURPS_CALCULATOR_URL      = BASE_GURPS_CALCULATOR_URL + "/Character/ImportGCS";
    private static final String            DEFAULT_PAGE_SETTINGS_KEY = "DefaultPageSettings";
    private              JComboBox<String> mPNGResolutionCombo;
    private              JButton           mGurpsCalculatorLink;
    private              JTextField        mGurpsCalculatorKey;
    private              JCheckBox         mUseNativePrinter;

    /** Initializes the services controlled by these preferences. */
    public static void initialize() {
        // Converting preferences that used to live in SheetPreferences
        Preferences prefs = Preferences.getInstance();
        for (String key : prefs.getModuleKeys(SheetPreferences.MODULE)) {
            if (PNG_RESOLUTION_KEY.equals(key)) {
                prefs.setValue(MODULE, PNG_RESOLUTION_KEY, prefs.getIntValue(SheetPreferences.MODULE, PNG_RESOLUTION_KEY, DEFAULT_PNG_RESOLUTION));
                prefs.removePreference(SheetPreferences.MODULE, PNG_RESOLUTION_KEY);
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

    public static String getGurpsCalculatorKey() {
        return Preferences.getInstance().getStringValue(MODULE, GURPS_CALCULATOR_KEY_KEY);
    }

    /**
     * @return The default page settings to use. May return {@code null} if no printer has been
     *         defined.
     */
    public static PrintManager getDefaultPageSettings() {
        String settings = Preferences.getInstance().getStringValue(MODULE, DEFAULT_PAGE_SETTINGS_KEY);
        if (settings != null && !settings.isEmpty()) {
            try (XMLReader in = new XMLReader(new StringReader(settings))) {
                XMLNodeType type = in.next();
                while (type != XMLNodeType.END_DOCUMENT) {
                    if (type == XMLNodeType.START_TAG) {
                        String name = in.getName();
                        if (PrintManager.TAG_ROOT.equals(name)) {
                            return new PrintManager(in);
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

    /** @param mgr The {@link PrintManager} to record the settings for. */
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

        FlexGrid grid = new FlexGrid();
        column.add(grid);

        FlexRow row        = new FlexRow();
        String  gcalcTitle = I18n.Text("GURPS Calculator Key");
        row.add(createLabel(gcalcTitle, gcalcTitle, Images.GCALC_LOGO));
        mGurpsCalculatorKey = createTextField(gcalcTitle, getGurpsCalculatorKey());
        row.add(mGurpsCalculatorKey);
        mGurpsCalculatorLink = createHyperlinkButton(I18n.Text("Find mine"), GURPS_CALCULATOR_URL);
        if (Desktop.isDesktopSupported()) {
            row.add(mGurpsCalculatorLink);
        }
        column.add(row);

        mUseNativePrinter = createCheckBox(I18n.Text("Use platform native print dialogs (settings cannot be saved)"), I18n.Text("<html><body>Whether or not the native print dialogs should be used.<br>Choosing this option will prevent the program from saving<br>and restoring print settings with the document.</body></html>"), PrintManager.useNativeDialogs());
        column.add(mUseNativePrinter);

        row = new FlexRow();
        row.add(createLabel(I18n.Text("Use"), pngDPIMsg()));
        mPNGResolutionCombo = createPNGResolutionPopup();
        row.add(mPNGResolutionCombo);
        row.add(createLabel(I18n.Text("when saving sheets to PNG"), pngDPIMsg(), SwingConstants.LEFT));
        column.add(row);

        column.add(new FlexSpacer(0, 0, false, true));

        column.apply(this);
    }

    private static String pngDPIMsg() {
        return I18n.Text("The resolution, in dots-per-inch, to use when saving sheets as PNG files");
    }

    private JButton createButton(String title, String tooltip) {
        JButton button = new JButton(title);
        button.setOpaque(false);
        button.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        button.addActionListener(this);
        UIUtilities.setToPreferredSizeOnly(button);
        add(button);
        return button;
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
        UIUtilities.setToPreferredSizeOnly(button);
        add(button);
        return button;
    }

    private JComboBox<String> createPNGResolutionPopup() {
        int               selection  = 0;
        int               resolution = getPNGResolution();
        JComboBox<String> combo      = new JComboBox<>();
        setupCombo(combo, pngDPIMsg());
        int length = DPI.length;
        for (int i = 0; i < length; i++) {
            combo.addItem(MessageFormat.format(I18n.Text("{0} dpi"), Integer.valueOf(DPI[i])));
            if (DPI[i] == resolution) {
                selection = i;
            }
        }
        combo.setSelectedIndex(selection);
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        UIUtilities.setToPreferredSizeOnly(combo);
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
        int length = DPI.length;
        for (int i = 0; i < length; i++) {
            if (DPI[i] == DEFAULT_PNG_RESOLUTION) {
                mPNGResolutionCombo.setSelectedIndex(i);
                break;
            }
        }
        mUseNativePrinter.setSelected(false);
    }

    @Override
    public boolean isSetToDefaults() {
        return getPNGResolution() == DEFAULT_PNG_RESOLUTION && !PrintManager.useNativeDialogs() && mGurpsCalculatorKey.getText() != null && mGurpsCalculatorKey.getText().isEmpty();
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        if (mGurpsCalculatorKey.getDocument() == event.getDocument()) {
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
        if (source == mUseNativePrinter) {
            PrintManager.useNativeDialogs(mUseNativePrinter.isSelected());
        }
        adjustResetButton();
    }
}
