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
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.Separator;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.units.LengthValue;

import javax.swing.SwingConstants;

public class PageSettingsEditor extends Panel {
    private static final Fixed6                     HUNDRED = new Fixed6(100);
    private              PageSettings               mSettings;
    private              Runnable                   mAdjustCallback;
    private              EditorField                mTopMargin;
    private              EditorField                mLeftMargin;
    private              EditorField                mBottomMargin;
    private              EditorField                mRightMargin;
    private              PopupMenu<PaperSize>       mPaperSize;
    private              PopupMenu<PageOrientation> mOrientation;

    public PageSettingsEditor(PageSettings settings, Runnable adjustCallback) {
        super(new PrecisionLayout().setColumns(4).setMargins(4, 0, 4, 0), false);
        Label header = new Label(I18n.text("Page Settings"));
        header.setThemeFont(Fonts.HEADER);
        add(header, new PrecisionLayoutData().setHorizontalSpan(4));
        add(new Separator(), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(4));
        mSettings = settings;
        mAdjustCallback = adjustCallback;
        mPaperSize = addPopupMenu(I18n.text("Paper Size"), PaperSize.getPaperSizes(), mSettings.getPaperSize(), (p) -> {
            mSettings.setPaperSize(mPaperSize.getSelectedItem());
            mAdjustCallback.run();
        });
        mOrientation = addPopupMenu(I18n.text("Orientation"), PageOrientation.values(), mSettings.getPageOrientation(), (p) -> {
            mSettings.setPageOrientation(mOrientation.getSelectedItem());
            mAdjustCallback.run();
        });
        LengthValue proto = new LengthValue(new Fixed6(99.99), LengthUnits.IN);
        mTopMargin = addField(I18n.text("Top Margin"), mSettings.getTopMargin(), proto, (f) -> {
            mSettings.setTopMargin((LengthValue) f.getValue());
            mAdjustCallback.run();
        });
        mBottomMargin = addField(I18n.text("Bottom Margin"), mSettings.getBottomMargin(), proto, (f) -> {
            mSettings.setBottomMargin((LengthValue) f.getValue());
            mAdjustCallback.run();
        });
        mLeftMargin = addField(I18n.text("Left Margin"), mSettings.getLeftMargin(), proto, (f) -> {
            mSettings.setLeftMargin((LengthValue) f.getValue());
            mAdjustCallback.run();
        });
        mRightMargin = addField(I18n.text("Right Margin"), mSettings.getRightMargin(), proto, (f) -> {
            mSettings.setRightMargin((LengthValue) f.getValue());
            mAdjustCallback.run();
        });
    }


    private static Fixed6 convert(Fixed6 value, LengthUnits oldUnits, LengthUnits newUnits) {
        return newUnits.convert(oldUnits, value).mul(HUNDRED).round().div(HUNDRED);
    }

    private EditorField addField(String title, Object value, Object protoValue, EditorField.ChangeListener listener) {
        EditorField field = new EditorField(FieldFactory.LENGTH, listener, SwingConstants.LEFT, value, protoValue, null);
        add(new Label(title), new PrecisionLayoutData().setFillHorizontalAlignment());
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment());
        return field;
    }

    private <T> PopupMenu<T> addPopupMenu(String title, T[] values, T selection, PopupMenu.SelectionListener<T> listener) {
        PopupMenu<T> popup = new PopupMenu<>(values, listener);
        popup.setSelectedItem(selection, false);
        if (title != null) {
            add(new Label(title), new PrecisionLayoutData().setFillHorizontalAlignment());
        }
        add(popup);
        return popup;
    }

    public void resetTo(PageSettings settings) {
        mSettings.copy(settings);
        mPaperSize.setSelectedItem(settings.getPaperSize(), true);
        mOrientation.setSelectedItem(settings.getPageOrientation(), true);
        mTopMargin.setValue(settings.getTopMargin());
        mLeftMargin.setValue(settings.getLeftMargin());
        mBottomMargin.setValue(settings.getBottomMargin());
        mRightMargin.setValue(settings.getRightMargin());
    }
}
