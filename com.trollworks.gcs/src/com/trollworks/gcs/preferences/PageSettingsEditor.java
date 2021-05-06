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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.page.PageOrientation;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Fixed6Formatter;
import com.trollworks.gcs.utility.units.LengthUnits;

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
    private EditorField                mWidth;
    private EditorField                mHeight;
    private EditorField                mTopMargin;
    private EditorField                mLeftMargin;
    private EditorField                mBottomMargin;
    private EditorField                mRightMargin;
    private JComboBox<PageOrientation> mOrientation;
    private JComboBox<LengthUnits>     mUnits;

    public PageSettingsEditor(PageSettings settings, Runnable adjustCallback) {
        super(new PrecisionLayout().setColumns(4).setMargins(0));
        setOpaque(false);
        mSettings = settings;
        mAdjustCallback = adjustCallback;
        Fixed6                  max     = new Fixed6(99.99);
        DefaultFormatterFactory factory = new DefaultFormatterFactory(new Fixed6Formatter(Fixed6.ZERO, max, false));
        mWidth = addField(I18n.Text("Page Width"), null, mSettings.getWidth(), max, factory, (evt) -> {
            mSettings.setWidth((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mHeight = addField(I18n.Text("Page Height"), null, mSettings.getHeight(), max, factory, (evt) -> {
            mSettings.setHeight((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mTopMargin = addField(I18n.Text("Top Margin"), null, mSettings.getTopMargin(), max, factory, (evt) -> {
            mSettings.setTopMargin((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mBottomMargin = addField(I18n.Text("Bottom Margin"), null, mSettings.getBottomMargin(), max, factory, (evt) -> {
            mSettings.setBottomMargin((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mLeftMargin = addField(I18n.Text("Left Margin"), null, mSettings.getLeftMargin(), max, factory, (evt) -> {
            mSettings.setLeftMargin((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mRightMargin = addField(I18n.Text("Right Margin"), null, mSettings.getRightMargin(), max, factory, (evt) -> {
            mSettings.setRightMargin((Fixed6) evt.getNewValue());
            mAdjustCallback.run();
        });
        mOrientation = addCombo(I18n.Text("Orientation"), PageOrientation.values(), mSettings.getPageOrientation(), (evt) -> {
            mSettings.setPageOrientation((PageOrientation) mOrientation.getSelectedItem());
            mAdjustCallback.run();
        });
        mUnits = addCombo(I18n.Text("Units"), new LengthUnits[]{LengthUnits.IN, LengthUnits.CM, LengthUnits.MM}, mSettings.getUnits(), (evt) -> {
            LengthUnits oldUnits = mSettings.getUnits();
            mSettings.setUnits((LengthUnits) mUnits.getSelectedItem());
            mWidth.setValue(mSettings.getUnits().convert(oldUnits, mSettings.getWidth()));
            mHeight.setValue(mSettings.getUnits().convert(oldUnits, mSettings.getHeight()));
            mTopMargin.setValue(mSettings.getUnits().convert(oldUnits, mSettings.getTopMargin()));
            mLeftMargin.setValue(mSettings.getUnits().convert(oldUnits, mSettings.getLeftMargin()));
            mBottomMargin.setValue(mSettings.getUnits().convert(oldUnits, mSettings.getBottomMargin()));
            mRightMargin.setValue(mSettings.getUnits().convert(oldUnits, mSettings.getRightMargin()));
            mAdjustCallback.run();
        });
    }

    private void addLabel(String title) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private EditorField addField(String title, String tooltip, Object value, Object protoValue, JFormattedTextField.AbstractFormatterFactory formatter, PropertyChangeListener listener) {
        addLabel(title);
        EditorField         field      = new EditorField(formatter, listener, SwingConstants.LEFT, value, protoValue, tooltip);
        PrecisionLayoutData layoutData = new PrecisionLayoutData().setFillHorizontalAlignment();
        add(field, layoutData);
        return field;
    }

    private <T> JComboBox<T> addCombo(String title, T[] values, T selection, ActionListener listener) {
        addLabel(title);
        JComboBox<T> combo = new JComboBox<>(values);
        combo.setOpaque(false);
        combo.setSelectedItem(selection);
        combo.addActionListener(listener);
        combo.setMaximumRowCount(combo.getItemCount());
        add(combo);
        return combo;
    }

    public void reset() {
        mSettings.reset();
        mUnits.setSelectedItem(mSettings.getUnits());
        mOrientation.setSelectedItem(mSettings.getPageOrientation());
        mWidth.setValue(mSettings.getWidth());
        mHeight.setValue(mSettings.getHeight());
        mTopMargin.setValue(mSettings.getTopMargin());
        mLeftMargin.setValue(mSettings.getLeftMargin());
        mBottomMargin.setValue(mSettings.getBottomMargin());
        mRightMargin.setValue(mSettings.getRightMargin());
    }

    public boolean isSetToDefaults() {
        return mSettings.equals(new PageSettings(null));
    }
}
