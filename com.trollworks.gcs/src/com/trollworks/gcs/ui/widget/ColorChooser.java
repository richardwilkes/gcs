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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import java.awt.Component;
import javax.swing.SwingConstants;

public final class ColorChooser extends Panel {
    private ColorWell mOriginal;
    private ColorWell mCurrent;
    private Button    mSetButton;

    public static Color presentToUser(Component comp, String title, Color current) {
        ColorChooser cc = new ColorChooser(current);
        Modal dialog = Modal.prepareToShowMessage(comp, title != null ? title :
                I18n.text("Choose a color…"), MessageType.QUESTION, cc);
        dialog.addCancelButton();
        cc.mSetButton = dialog.addButton(I18n.text("Set"), Modal.OK);
        cc.adjustButtons();
        dialog.presentToUser();
        switch (dialog.getResult()) {
        case Modal.OK:
            Commitable.sendCommitToFocusOwner();
            return cc.mCurrent.getWellColor();
        default: // Close or cancel
            return null;
        }
    }

    private ColorChooser(Color current) {
        super(new PrecisionLayout().setColumns(3));

        Integer proto = Integer.valueOf(255);
        current = new Color(current.getRGB(), true);

        add(new Label(I18n.text("Red")), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(new EditorField(FieldFactory.BYTE, (f2) -> {
            Color color2 = mCurrent.getWellColor();
            mCurrent.setWellColor(new Color(((Integer) f2.getValue()).intValue(), color2.getGreen(),
                    color2.getBlue(), color2.getAlpha()));
            adjustButtons();
        }, SwingConstants.RIGHT, Integer.valueOf(current.getRed()), proto,
                I18n.text("The value of the red channel, from 0 to 255")));

        mCurrent = new ColorWell(current, null);
        mCurrent.setEnabled(false);
        add(mCurrent, new PrecisionLayoutData().setMiddleHorizontalAlignment().
                setLeftMargin(LayoutConstants.WINDOW_BORDER_INSET));

        add(new Label(I18n.text("Green")), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(new EditorField(FieldFactory.BYTE, (f1) -> {
            Color color1 = mCurrent.getWellColor();
            mCurrent.setWellColor(new Color(color1.getRed(), ((Integer) f1.getValue()).intValue(),
                    color1.getBlue(), color1.getAlpha()));
            adjustButtons();
        }, SwingConstants.RIGHT, Integer.valueOf(current.getGreen()), proto,
                I18n.text("The value of the green channel, from 0 to 255")));

        add(new Label(I18n.text("New")), new PrecisionLayoutData().setMiddleHorizontalAlignment().
                setLeftMargin(LayoutConstants.WINDOW_BORDER_INSET));

        add(new Label(I18n.text("Blue")), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(new EditorField(FieldFactory.BYTE, (f1) -> {
            Color color1 = mCurrent.getWellColor();
            mCurrent.setWellColor(new Color(color1.getRed(), color1.getGreen(),
                    ((Integer) f1.getValue()).intValue(), color1.getAlpha()));
            adjustButtons();
        }, SwingConstants.RIGHT, Integer.valueOf(current.getBlue()), proto,
                I18n.text("The value of the blue channel, from 0 to 255")));

        mOriginal = new ColorWell(current, null);
        mOriginal.setEnabled(false);
        add(mOriginal, new PrecisionLayoutData().setMiddleHorizontalAlignment().
                setLeftMargin(LayoutConstants.WINDOW_BORDER_INSET));

        add(new Label(I18n.text("Alpha")), new PrecisionLayoutData().setEndHorizontalAlignment());
        add(new EditorField(FieldFactory.BYTE, (f) -> {
            Color color = mCurrent.getWellColor();
            mCurrent.setWellColor(new Color(color.getRed(), color.getGreen(),
                    color.getBlue(), ((Integer) f.getValue()).intValue()));
            adjustButtons();
        }, SwingConstants.RIGHT, Integer.valueOf(current.getAlpha()), proto,
                I18n.text("The value of the alpha channel, from 0 to 255")));

        add(new Label(I18n.text("Original")), new PrecisionLayoutData().
                setMiddleHorizontalAlignment().setLeftMargin(LayoutConstants.WINDOW_BORDER_INSET));
    }

    private void adjustButtons() {
        mSetButton.setEnabled(mOriginal.getWellColor().getRGB() != mCurrent.getWellColor().getRGB());
    }
}
