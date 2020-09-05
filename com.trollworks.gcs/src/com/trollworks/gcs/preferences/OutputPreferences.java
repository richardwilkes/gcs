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

import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Desktop;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.net.URI;
import java.text.MessageFormat;
import javax.swing.JButton;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
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
        setLayout(new PrecisionLayout().setColumns(3));
        Preferences prefs = Preferences.getInstance();

        String gcalcTitle = I18n.Text("GURPS Calculator Key");
        addLabel(gcalcTitle, null);
        mGurpsCalculatorKey = addTextField(prefs.getGURPSCalculatorKey(), gcalcTitle);
        mGurpsCalculatorLink = addButton(I18n.Text("Find mine"), gcalcTitle);

        addLabel(I18n.Text("Image Resolution"), pngDPIMsg());
        mPNGResolutionCombo = addPNGResolutionPopup();

        mUseNativePrinter = addCheckBox(I18n.Text("Use platform native print dialogs (settings cannot be saved)"), I18n.Text("<html><body>Whether or not the native print dialogs should be used.<br>Choosing this option will prevent the program from saving<br>and restoring print settings with the document.</body></html>"), prefs.useNativePrintDialogs());
    }

    private void addLabel(String text, String tooltip) {
        JLabel label = new JLabel(text, SwingConstants.RIGHT);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        label.setOpaque(false);
        add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private JTextField addTextField(String text, String tooltip) {
        JTextField field = new JTextField(text);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return field;
    }

    private JButton addButton(String title, String tooltip) {
        JButton button = new JButton(title);
        button.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        button.addActionListener(this);
        add(button);
        return button;
    }

    private JCheckBox addCheckBox(String title, String tooltip, boolean checked) {
        JCheckBox checkbox = new JCheckBox(title, checked);
        checkbox.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        checkbox.setOpaque(false);
        checkbox.addItemListener(this);
        add(checkbox, new PrecisionLayoutData().setHorizontalSpan(3));
        return checkbox;
    }

    private JComboBox<String> addPNGResolutionPopup() {
        int               selection  = 0;
        int               resolution = Preferences.getInstance().getPNGResolution();
        JComboBox<String> combo      = new JComboBox<>();
        combo.setOpaque(false);
        combo.setToolTipText(Text.wrapPlainTextForToolTip(pngDPIMsg()));
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
        add(combo, new PrecisionLayoutData().setHorizontalSpan(2));
        return combo;
    }

    private static String pngDPIMsg() {
        return I18n.Text("The resolution, in dots-per-inch, to use when saving sheets as PNG files");
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
