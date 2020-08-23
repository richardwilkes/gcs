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
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.IntegerFormatter;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import javax.swing.JButton;
import javax.swing.JLabel;
import javax.swing.JScrollPane;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;
import javax.swing.text.DefaultFormatterFactory;

/** The page reference lookup preferences panel. */
public class ReferenceLookupPreferences extends PreferencePanel {
    private BandedPanel mPanel;

    /**
     * Creates a new {@link ReferenceLookupPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public ReferenceLookupPreferences(PreferencesWindow owner) {
        super(I18n.Text("Page References"), owner);
        setLayout(new BorderLayout());
        mPanel = new BandedPanel(I18n.Text("Page References"));
        mPanel.setLayout(new PrecisionLayout().setColumns(4));
        mPanel.setBorder(new EmptyBorder(2, 5, 2, 5));
        mPanel.setOpaque(true);
        mPanel.setBackground(Color.WHITE);
        Preferences prefs      = Preferences.getInstance();
        Color       background = new Color(255, 255, 224);
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
        add(scroller);
    }

    @Override
    public boolean isSetToDefaults() {
        return Preferences.getInstance().arePdfRefsSetToDefault();
    }

    @Override
    public void reset() {
        Preferences.getInstance().clearPdfRefs();
        mPanel.removeAll();
        mPanel.setSize(mPanel.getPreferredSize());
    }
}
