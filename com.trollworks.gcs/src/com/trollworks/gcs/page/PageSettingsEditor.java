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

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.StdLabel;
import com.trollworks.gcs.ui.widget.StdPanel;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.LengthValue;

import java.awt.event.ActionListener;
import java.beans.PropertyChangeListener;
import javax.swing.JComboBox;
import javax.swing.JSeparator;
import javax.swing.SwingConstants;

public class PageSettingsEditor extends StdPanel {
    private static final Fixed6                     HUNDRED = new Fixed6(100);
    private              PageSettings               mSettings;
    private              Runnable                   mAdjustCallback;
    private              ResetPageSettings          mResetCallback;
    private              EditorField                mTopMargin;
    private              EditorField                mLeftMargin;
    private              EditorField                mBottomMargin;
    private              EditorField                mRightMargin;
    private              JComboBox<PaperSize>       mPaperSize;
    private              JComboBox<PageOrientation> mOrientation;
    private              JComboBox<LengthUnits>     mUnits;

    public interface ResetPageSettings {
        void resetPageSettings(PageSettings settings);
    }

    public PageSettingsEditor(PageSettings settings, Runnable adjustCallback, ResetPageSettings resetCallback) {
        super(new PrecisionLayout().setColumns(4).setMargins(4, 0, 4, 0), false);
        StdLabel header = new StdLabel(I18n.text("Page Settings"));
        header.setThemeFont(ThemeFont.HEADER);
        add(header, new PrecisionLayoutData().setHorizontalSpan(4));
        add(new JSeparator(SwingConstants.HORIZONTAL), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(4));
        mSettings = settings;
        mAdjustCallback = adjustCallback;
        mResetCallback = resetCallback;
        mPaperSize = addCombo(I18n.text("Paper Size"), PaperSize.getPaperSizes(), mSettings.getPaperSize(), (evt) -> {
            mSettings.setPaperSize(((PaperSize) mPaperSize.getSelectedItem()));
            mAdjustCallback.run();
        });
        mOrientation = addCombo(I18n.text("Orientation"), PageOrientation.values(), mSettings.getPageOrientation(), (evt) -> {
            mSettings.setPageOrientation((PageOrientation) mOrientation.getSelectedItem());
            mAdjustCallback.run();
        });
        LengthValue proto = new LengthValue(new Fixed6(99.99), LengthUnits.IN);
        mTopMargin = addField(I18n.text("Top Margin"), mSettings.getTopMargin(), proto, (evt) -> {
            mSettings.setTopMargin((LengthValue) evt.getNewValue());
            mAdjustCallback.run();
        });
        mBottomMargin = addField(I18n.text("Bottom Margin"), mSettings.getBottomMargin(), proto, (evt) -> {
            mSettings.setBottomMargin((LengthValue) evt.getNewValue());
            mAdjustCallback.run();
        });
        mLeftMargin = addField(I18n.text("Left Margin"), mSettings.getLeftMargin(), proto, (evt) -> {
            mSettings.setLeftMargin((LengthValue) evt.getNewValue());
            mAdjustCallback.run();
        });
        mRightMargin = addField(I18n.text("Right Margin"), mSettings.getRightMargin(), proto, (evt) -> {
            mSettings.setRightMargin((LengthValue) evt.getNewValue());
            mAdjustCallback.run();
        });
    }


    private static Fixed6 convert(Fixed6 value, LengthUnits oldUnits, LengthUnits newUnits) {
        return newUnits.convert(oldUnits, value).mul(HUNDRED).round().div(HUNDRED);
    }

    private EditorField addField(String title, Object value, Object protoValue, PropertyChangeListener listener) {
        EditorField field = new EditorField(FieldFactory.LENGTH, listener, SwingConstants.LEFT, value, protoValue, null);
        add(new StdLabel(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    private <T> JComboBox<T> addCombo(String title, T[] values, T selection, ActionListener listener) {
        JComboBox<T> combo = new JComboBox<>(values);
        combo.setOpaque(false);
        combo.setSelectedItem(selection);
        combo.addActionListener(listener);
        combo.setMaximumRowCount(combo.getItemCount());
        if (title != null) {
            add(new StdLabel(title, combo), new PrecisionLayoutData().setFillHorizontalAlignment());
        }
        add(combo);
        return combo;
    }

    public void reset() {
        mResetCallback.resetPageSettings(mSettings);
        mPaperSize.setSelectedItem(mSettings.getPaperSize());
        mOrientation.setSelectedItem(mSettings.getPageOrientation());
        mTopMargin.setValue(mSettings.getTopMargin());
        mLeftMargin.setValue(mSettings.getLeftMargin());
        mBottomMargin.setValue(mSettings.getBottomMargin());
        mRightMargin.setValue(mSettings.getRightMargin());
    }
}
