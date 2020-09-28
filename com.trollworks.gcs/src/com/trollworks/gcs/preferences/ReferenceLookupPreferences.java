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

import com.trollworks.gcs.pdfview.PDFRef;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.FlexLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.IntegerFormatter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.event.ActionListener;
import java.awt.event.ActionEvent;
import java.awt.event.ItemListener;
import java.awt.event.ItemEvent;
import javax.swing.JButton;
import javax.swing.JLabel;
import javax.swing.JScrollPane;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.DefaultFormatterFactory;
import javax.swing.text.Document;

/** The page reference lookup preferences panel. */
public class ReferenceLookupPreferences extends PreferencePanel implements DocumentListener, ItemListener {
    private JTextField  mPdfViewerCommandLine;
    private BandedPanel mPanel;

    /**
     * Creates a new {@link ReferenceLookupPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public ReferenceLookupPreferences(PreferencesWindow owner) {
        super(I18n.Text("Page References"), owner);
        setLayout(new PrecisionLayout().setColumns(2));
        Preferences prefs = Preferences.getInstance();

        // Command line to use to open PDFs
        String pdfViewerCommandLineTooltip = I18n.Text("The command to launch a PDF viewer to view a page reference. %f for file, %p for page.") + "\n" + I18n.Text("Example") + ":" + "\"C:\\Program Files\\SumatraPDF\\SumatraPDF.exe\" -page %p \"%f\"\n\n" + I18n.Text("Clear field to revert to default behavior.");
        addLabel(I18n.Text("PDF Viewer Launch String"), pdfViewerCommandLineTooltip);
        mPdfViewerCommandLine = addTextField(prefs.getPdfViewerString(), pdfViewerCommandLineTooltip);

        // List of page references encountered and set so far
        mPanel = new BandedPanel(I18n.Text("Page References"));
        mPanel.setLayout(new PrecisionLayout().setColumns(4));
        mPanel.setBorder(new EmptyBorder(2, 5, 2, 5));
        mPanel.setOpaque(true);
        mPanel.setBackground(Color.WHITE);

        Color background = new Color(255, 255, 224);
        for (PDFRef ref : prefs.allPdfRefs(false)) {
            JButton button = new JButton(I18n.Text("Remove"));
            button.addActionListener(event -> {
                prefs.removePdfRef(ref);
                Component[] children = mPanel.getComponents();
                int         length   = children.length;
                for (int i = 0; i < length; i++) {
                    if (children[i] == button) {
                        for (int j = i + 4; --j >= i; ) {
                            mPanel.remove(j);
                        }
                        mPanel.setSize(mPanel.getPreferredSize());
                        break;
                    }
                }
            });
            mPanel.add(button);
            JLabel idLabel = new JLabel(ref.getID(), SwingConstants.CENTER);
            idLabel.setBorder(new CompoundBorder(new LineBorder(), new EmptyBorder(1, 4, 1, 4)));
            idLabel.setOpaque(true);
            idLabel.setBackground(background);
            mPanel.add(idLabel, new PrecisionLayoutData().setFillHorizontalAlignment());
            EditorField field = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(-9999, 9999, true)), event -> ref.setPageToIndexOffset(((Integer) event.getNewValue()).intValue()), SwingConstants.RIGHT, Integer.valueOf(ref.getPageToIndexOffset()), Integer.valueOf(-9999), I18n.Text("If your PDF is opening up to the wrong page when opening page references, enter an offset here to compensate."));
            mPanel.add(field);
            mPanel.add(new JLabel(ref.getPath().normalize().toAbsolutePath().toString()));
        }
        mPanel.setSize(mPanel.getPreferredSize());
        JScrollPane scroller      = new JScrollPane(mPanel);
        Dimension   preferredSize = scroller.getPreferredSize();
        if (preferredSize.height > 200) {
            preferredSize.height = 200;
        }
        scroller.setPreferredSize(preferredSize);
        add(scroller, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalSpan(2).setFillHorizontalAlignment());
    }

    private void addLabel(String title, String tooltip) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private JTextField addTextField(String value, String tooltip) {
        JTextField field = new JTextField(value);
        field.getDocument().addDocumentListener(this);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        return field;
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
    public void changedUpdate(DocumentEvent event) {
        Preferences prefs    = Preferences.getInstance();
        Document    document = event.getDocument();
        if (mPdfViewerCommandLine.getDocument() == document) {
            prefs.setPdfViewerString(mPdfViewerCommandLine.getText());
        }
        adjustResetButton();
    }

    @Override
    public boolean isSetToDefaults() {
        return Preferences.getInstance().arePdfRefsSetToDefault();
    }

    @Override
    public void reset() {
        Preferences.getInstance().clearPdfLaunchString();
        mPdfViewerCommandLine.setText("");
        Preferences.getInstance().clearPdfRefs();
        mPanel.removeAll();
        mPanel.setSize(mPanel.getPreferredSize());
    }

    @Override
    public void itemStateChanged(ItemEvent event) {
        Preferences prefs  = Preferences.getInstance();
        Object      source = event.getSource();
        if (source == mPdfViewerCommandLine) {
            prefs.setPdfViewerString(mPdfViewerCommandLine.getText());
        }
        adjustResetButton();
    }
}
