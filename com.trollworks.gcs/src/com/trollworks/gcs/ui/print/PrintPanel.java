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

package com.trollworks.gcs.ui.print;

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.IntegerFormatter;

import java.awt.event.ActionEvent;
import javax.print.PrintService;
import javax.print.attribute.PrintRequestAttributeSet;
import javax.print.attribute.standard.Copies;
import javax.print.attribute.standard.PageRanges;
import javax.swing.ButtonGroup;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JRadioButton;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

/** Provides the basic print panel. */
public class PrintPanel extends PageSetupPanel {
    private EditorField  mCopies;
    private JRadioButton mPageRangeAll;
    private JRadioButton mPageRangeSome;
    private EditorField  mPageRangeStart;
    private EditorField  mPageRangeEnd;

    /**
     * Creates a new print panel.
     *
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     */
    public PrintPanel(PrintService service, PrintRequestAttributeSet set) {
        super(service, set);
    }

    @Override
    protected void rebuildSelf(PrintRequestAttributeSet set) {
        createCopiesField(set);
        createPageRangeFields(set);
        super.rebuildSelf(set);
    }

    private void createCopiesField(PrintRequestAttributeSet set) {
        PrintService service = getService();
        if (service.isAttributeCategorySupported(Copies.class)) {
            mCopies = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(1, 999, false)), null, SwingConstants.RIGHT, Integer.valueOf(PrintUtilities.getCopies(service, set)), Integer.valueOf(999), null);
            UIUtilities.setToPreferredSizeOnly(mCopies);
            LinkedLabel label = new LinkedLabel(I18n.Text("Copies"), mCopies);
            add(label, new PrecisionLayoutData().setEndHorizontalAlignment());
            add(mCopies);
        } else {
            mCopies = null;
        }
    }

    private void createPageRangeFields(PrintRequestAttributeSet set) {
        PrintService service = getService();
        if (service.isAttributeCategorySupported(PageRanges.class)) {
            ButtonGroup group      = new ButtonGroup();
            int         start      = 1;
            int         end        = 9999;
            PageRanges  pageRanges = (PageRanges) set.get(PageRanges.class);
            if (pageRanges != null) {
                int[][] ranges = pageRanges.getMembers();
                if (ranges.length > 0 && ranges[0].length > 1) {
                    start = ranges[0][0];
                    end = ranges[0][1];
                } else {
                    pageRanges = null;
                }
            }
            JLabel label = new JLabel(I18n.Text("Print Range"), SwingConstants.CENTER);
            add(label, new PrecisionLayoutData().setEndHorizontalAlignment());
            JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(5));
            mPageRangeAll = new JRadioButton(I18n.Text("All"), pageRanges == null);
            wrapper.add(mPageRangeAll);
            mPageRangeSome = new JRadioButton(I18n.Text("Pages"), pageRanges != null);
            wrapper.add(mPageRangeSome);
            mPageRangeStart = createPageRangeField(start, wrapper);
            wrapper.add(new JLabel(I18n.Text("to"), SwingConstants.CENTER));
            mPageRangeEnd = createPageRangeField(end, wrapper);
            add(wrapper);
            group.add(mPageRangeAll);
            group.add(mPageRangeSome);
            adjustPageRanges();
            mPageRangeAll.addActionListener(this);
            mPageRangeSome.addActionListener(this);
        } else {
            mPageRangeAll = null;
            mPageRangeSome = null;
            mPageRangeStart = null;
            mPageRangeEnd = null;
        }
    }

    private static EditorField createPageRangeField(int value, JPanel parent) {
        EditorField field = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(1, 9999, false)), null, SwingConstants.RIGHT, Integer.valueOf(value), Integer.valueOf(9999), null);
        UIUtilities.setToPreferredSizeOnly(field);
        parent.add(field);
        return field;
    }

    private void adjustPageRanges() {
        boolean enabled = !mPageRangeAll.isSelected();
        mPageRangeStart.setEnabled(enabled);
        mPageRangeEnd.setEnabled(enabled);
    }

    @Override
    public PrintService accept(PrintRequestAttributeSet set) {
        PrintService service = super.accept(set);
        if (mCopies != null) {
            PrintUtilities.setCopies(set, ((Integer) mCopies.getValue()).intValue());
        }
        if (mPageRangeAll != null) {
            if (mPageRangeAll.isSelected()) {
                PrintUtilities.setPageRanges(set, null);
            } else {
                int start = ((Integer) mPageRangeStart.getValue()).intValue();
                int end   = ((Integer) mPageRangeEnd.getValue()).intValue();
                if (start > end) {
                    int tmp = start;
                    start = end;
                    end = tmp;
                }
                PrintUtilities.setPageRanges(set, new PageRanges(start, end));
            }
        }
        return service;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object src = event.getSource();
        if (src == mPageRangeAll || src == mPageRangeSome) {
            adjustPageRanges();
        } else {
            super.actionPerformed(event);
        }
    }
}
