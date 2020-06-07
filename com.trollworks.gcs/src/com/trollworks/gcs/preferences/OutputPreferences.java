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

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.FlexColumn;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Desktop;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.net.URI;
import java.text.MessageFormat;
import javax.swing.JButton;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** The sheet preferences panel. */
public class OutputPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
    private static final int[]             DPI                       = {72, 96, 144, 150, 200, 300};
    public static final  String            BASE_GURPS_CALCULATOR_URL = "http://www.gurpscalculator.com";
    public static final  String            GURPS_CALCULATOR_URL      = BASE_GURPS_CALCULATOR_URL + "/Character/ImportGCS";
    private              JComboBox<String> mPNGResolutionCombo;
    private              JButton           mGurpsCalculatorLink;
    private              JTextField        mGurpsCalculatorKey;
    private              JCheckBox         mUseNativePrinter;

    /**
     * Creates a new {@link OutputPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public OutputPreferences(PreferencesWindow owner) {
        super(I18n.Text("Output"), owner);
        Preferences prefs  = Preferences.getInstance();
        FlexColumn  column = new FlexColumn();

        FlexGrid grid = new FlexGrid();
        column.add(grid);

        FlexRow row        = new FlexRow();
        String  gcalcTitle = I18n.Text("GURPS Calculator Key");
        row.add(createLabel(gcalcTitle, gcalcTitle, Images.GCALC_LOGO));
        mGurpsCalculatorKey = createTextField(gcalcTitle, prefs.getGURPSCalculatorKey());
        row.add(mGurpsCalculatorKey);
        mGurpsCalculatorLink = createHyperlinkButton(I18n.Text("Find mine"), GURPS_CALCULATOR_URL);
        if (Desktop.isDesktopSupported()) {
            row.add(mGurpsCalculatorLink);
        }
        column.add(row);

        mUseNativePrinter = createCheckBox(I18n.Text("Use platform native print dialogs (settings cannot be saved)"), I18n.Text("<html><body>Whether or not the native print dialogs should be used.<br>Choosing this option will prevent the program from saving<br>and restoring print settings with the document.</body></html>"), prefs.useNativePrintDialogs());
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
        int               resolution = Preferences.getInstance().getPNGResolution();
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
            Preferences.getInstance().setPNGResolution(DPI[mPNGResolutionCombo.getSelectedIndex()]);
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
            if (DPI[i] == Preferences.DEFAULT_PNG_RESOLUTION) {
                mPNGResolutionCombo.setSelectedIndex(i);
                break;
            }
        }
        mUseNativePrinter.setSelected(false);
    }

    @Override
    public boolean isSetToDefaults() {
        Preferences prefs      = Preferences.getInstance();
        boolean     atDefaults = prefs.getPNGResolution() == Preferences.DEFAULT_PNG_RESOLUTION;
        atDefaults = atDefaults && prefs.useNativePrintDialogs() == Preferences.DEFAULT_USE_NATIVE_PRINT_DIALOGS;
        atDefaults = atDefaults && mGurpsCalculatorKey.getText() != null && mGurpsCalculatorKey.getText().isEmpty();
        return atDefaults;
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        if (mGurpsCalculatorKey.getDocument() == event.getDocument()) {
            Preferences.getInstance().setGURPSCalculatorKey(mGurpsCalculatorKey.getText());
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
            Preferences.getInstance().setUseNativePrintDialogs(mUseNativePrinter.isSelected());
        }
        adjustResetButton();
    }
}
