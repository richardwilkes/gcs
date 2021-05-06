/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.page;

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.border.Edge;
import com.trollworks.gcs.ui.border.TitledBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Fixed6Formatter;
import com.trollworks.gcs.utility.units.LengthUnits;

import java.awt.Container;
import java.awt.event.ActionListener;
import java.beans.PropertyChangeListener;
import javax.swing.JComboBox;
import javax.swing.JFormattedTextField;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

public class PageSettingsEditor extends JPanel {
    private PageSettings               mSettings;
    private Runnable                   mAdjustCallback;
    private EditorField                mTopMargin;
    private EditorField                mLeftMargin;
    private EditorField                mBottomMargin;
    private EditorField                mRightMargin;
    private JComboBox<PaperSize>       mPaperSize;
    private JComboBox<PageOrientation> mOrientation;
    private JComboBox<LengthUnits>     mUnits;

    public PageSettingsEditor(PageSettings settings, Runnable adjustCallback) {
        super(new PrecisionLayout().setMargins(0));
        setOpaque(false);
        mSettings = settings;
        mAdjustCallback = adjustCallback;
        TitledBorder border = new TitledBorder(Fonts.getDefaultFont(), I18n.Text("Page Settings"));
        border.setColorAndThickness(Edge.LEFT, Colors.TRANSPARENT, 8);
        border.setColorAndThickness(Edge.RIGHT, Colors.TRANSPARENT, 8);
        border.setColorAndThickness(Edge.BOTTOM, Colors.TRANSPARENT, 0);
        setBorder(border);

        Wrapper wrapper = new Wrapper(new PrecisionLayout().setColumns(4).setMargins(0));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        mPaperSize = addCombo(wrapper, I18n.Text("Paper Size"), PaperSize.getPaperSizes(), mSettings.getPaperSize(), (evt) -> {
            mSettings.setPaperSize(((PaperSize) mPaperSize.getSelectedItem()));
            mAdjustCallback.run();
        });
        mOrientation = addCombo(wrapper, I18n.Text("Orientation"), PageOrientation.values(), mSettings.getPageOrientation(), (evt) -> {
            mSettings.setPageOrientation((PageOrientation) mOrientation.getSelectedItem());
            mAdjustCallback.run();
        });

        wrapper = new Wrapper(new PrecisionLayout().setColumns(9).setMargins(0));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        Fixed6                  proto   = new Fixed6(9999.999999);
        DefaultFormatterFactory factory = new DefaultFormatterFactory(new Fixed6Formatter(Fixed6.ZERO, new Fixed6(9999), false));
        mTopMargin = addField(wrapper, I18n.Text("Margins: Top"), null, mSettings.getTopMargin(), proto, factory, (evt) -> {
            mSettings.setTopMargin((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mLeftMargin = addField(wrapper, I18n.Text("Left"), null, mSettings.getLeftMargin(), proto, factory, (evt) -> {
            mSettings.setLeftMargin((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mBottomMargin = addField(wrapper, I18n.Text("Bottom"), null, mSettings.getBottomMargin(), proto, factory, (evt) -> {
            mSettings.setBottomMargin((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mRightMargin = addField(wrapper, I18n.Text("Right"), null, mSettings.getRightMargin(), proto, factory, (evt) -> {
            mSettings.setRightMargin((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mUnits = addCombo(wrapper, null, new LengthUnits[]{LengthUnits.IN, LengthUnits.MM}, mSettings.getUnits(), (evt) -> {
            LengthUnits oldUnits = mSettings.getUnits();
            mSettings.setUnits((LengthUnits) mUnits.getSelectedItem());
            LengthUnits newUnits = mSettings.getUnits();
            if (oldUnits != newUnits) {
                mTopMargin.setValue(convert(mSettings.getTopMargin(), oldUnits, newUnits));
                mLeftMargin.setValue(convert(mSettings.getLeftMargin(), oldUnits, newUnits));
                mBottomMargin.setValue(convert(mSettings.getBottomMargin(), oldUnits, newUnits));
                mRightMargin.setValue(convert(mSettings.getRightMargin(), oldUnits, newUnits));
            }
            mAdjustCallback.run();
        });
    }

    private static final Fixed6 HUNDRED = new Fixed6(100);

    private Fixed6 convert(Fixed6 value, LengthUnits oldUnits, LengthUnits newUnits) {
        return newUnits.convert(oldUnits, value).mul(HUNDRED).round().div(HUNDRED);
    }

    private void addLabel(Container container, String title) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        container.add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private EditorField addField(Container container, String title, String tooltip, Object value, Object protoValue, JFormattedTextField.AbstractFormatterFactory formatter, PropertyChangeListener listener) {
        addLabel(container, title);
        EditorField         field      = new EditorField(formatter, listener, SwingConstants.LEFT, value, protoValue, tooltip);
        PrecisionLayoutData layoutData = new PrecisionLayoutData().setFillHorizontalAlignment();
        container.add(field, layoutData);
        return field;
    }

    private <T> JComboBox<T> addCombo(Container container, String title, T[] values, T selection, ActionListener listener) {
        if (title != null) {
            addLabel(container, title);
        }
        JComboBox<T> combo = new JComboBox<>(values);
        combo.setOpaque(false);
        combo.setSelectedItem(selection);
        combo.addActionListener(listener);
        combo.setMaximumRowCount(combo.getItemCount());
        container.add(combo);
        return combo;
    }

    public void reset() {
        mSettings.reset();
        mPaperSize.setSelectedItem(mSettings.getPaperSize());
        mUnits.setSelectedItem(mSettings.getUnits());
        mOrientation.setSelectedItem(mSettings.getPageOrientation());
        mTopMargin.setValue(mSettings.getTopMargin());
        mLeftMargin.setValue(mSettings.getLeftMargin());
        mBottomMargin.setValue(mSettings.getBottomMargin());
        mRightMargin.setValue(mSettings.getRightMargin());
    }

    public boolean isSetToDefaults() {
        return mSettings.equals(new PageSettings(null));
    }
}
