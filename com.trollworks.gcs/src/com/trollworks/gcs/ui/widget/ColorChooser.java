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
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.DoubleFormatter;
import com.trollworks.gcs.utility.text.IntegerFormatter;

import java.awt.Color;
import java.awt.Component;
import java.awt.Window;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

public final class ColorChooser extends Panel {
    private ColorWell mOriginal;
    private ColorWell mCurrent;
    private Button    mSetButton;

    public static Color presentToUser(Component comp, String title, Color current) {
        ColorChooser cc = new ColorChooser(current);
        Modal modal = Modal.prepareToShowMessage(comp, title != null ? title :
                I18n.text("Choose a color…"), MessageType.QUESTION, cc);
        modal.addCancelButton();
        cc.mSetButton = modal.addButton(I18n.text("Set"), Modal.OK);
        cc.adjustButtons();
        modal.presentToUser();
        switch (modal.getResult()) {
        case Modal.OK:
            Commitable.sendCommitToFocusOwner();
            return cc.mCurrent.getWellColor();
        default: // Close or cancel
            return null;
        }
    }

    private ColorChooser(Color color) {
        super(new PrecisionLayout().setColumns(2).setMargins(0).setHorizontalSpacing(LayoutConstants.WINDOW_BORDER_INSET));
        PopupMenu<Choice> popup = new PopupMenu<>(Choice.values(), (p) -> {
            Choice choice = p.getSelectedItem();
            if (choice != null) {
                if (getComponentCount() > 2) {
                    remove(2);
                }
                add(choice.createPanel(this, new Color(mCurrent.getWellColor().getRGB(), true)));
                invalidate();
                Window wnd = UIUtilities.getAncestorOfType(this, Window.class);
                if (wnd != null) {
                    wnd.setMinimumSize(null);
                    wnd.setMaximumSize(null);
                    wnd.setPreferredSize(null);
                    wnd.pack();
                }
            }
        });
        popup.setSelectedItem(Choice.RGB, false);
        add(popup, new PrecisionLayoutData().setHorizontalSpan(2).setMiddleHorizontalAlignment().setBottomMargin(LayoutConstants.TOOLBAR_VERTICAL_INSET));
        add(createBeforeAfterWells(color));
        add(Choice.RGB.createPanel(this, new Color(mOriginal.getWellColor().getRGB(), true)));
    }

    private Panel createBeforeAfterWells(Color color) {
        Panel panel = new Panel(new PrecisionLayout().setColumns(2).setMargins(0));
        mCurrent = addWell(panel, I18n.text("Updated"), color);
        mOriginal = addWell(panel, I18n.text("Original"), color);
        return panel;
    }

    private static ColorWell addWell(Panel panel, String title, Color color) {
        ColorWell well = new ColorWell(new Color(color.getRGB(), true), null);
        well.setEnabled(false);
        panel.add(new Label(title), new PrecisionLayoutData().setEndHorizontalAlignment());
        panel.add(well);
        return well;
    }

    private static void addLabel(Panel panel, String label) {
        panel.add(new Label(label), new PrecisionLayoutData().setEndHorizontalAlignment());
    }

    private interface ColorAdjuster {
        Color adjustColor(Color color, int value);
    }

    private void adjustButtons() {
        mSetButton.setEnabled(mOriginal.getWellColor().getRGB() != mCurrent.getWellColor().getRGB());
    }

    private enum Choice {
        RGB {
            @Override
            public String toString() {
                return I18n.text("RGB");
            }

            Panel createPanel(ColorChooser chooser, Color color) {
                Panel panel = new Panel(new PrecisionLayout().setColumns(2).setMargins(0));
                addChannelField(chooser, panel, I18n.text("Red"), color.getRed(),
                        I18n.text("The value of the red channel, from 0 to 255"),
                        (c, v) -> new Color(v, c.getGreen(), c.getBlue(), c.getAlpha()));
                addChannelField(chooser, panel, I18n.text("Green"), color.getGreen(),
                        I18n.text("The value of the green channel, from 0 to 255"),
                        (c, v) -> new Color(c.getRed(), v, c.getBlue(), c.getAlpha()));
                addChannelField(chooser, panel, I18n.text("Blue"), color.getBlue(),
                        I18n.text("The value of the blue channel, from 0 to 255"),
                        (c, v) -> new Color(c.getRed(), c.getGreen(), v, c.getAlpha()));
                addChannelField(chooser, panel, I18n.text("Alpha"), color.getAlpha(),
                        I18n.text("The value of the alpha channel, from 0 to 255"),
                        (c, v) -> new Color(c.getRed(), c.getGreen(), c.getBlue(), v));
                return panel;
            }
        },
        HEX_RGB {
            @Override
            public String toString() {
                return I18n.text("Hex RGB");
            }

            Panel createPanel(ColorChooser chooser, Color color) {
                Panel panel = new Panel(new PrecisionLayout().setColumns(2).setMargins(0));
                addLabel(panel, toString());
                panel.add(new EditorField(FieldFactory.STRING, (f) -> {
                    chooser.mCurrent.setWellColor(Colors.decode(f.getText()));
                    f.setText(Colors.encodeToHex(chooser.mCurrent.getWellColor()));
                    chooser.adjustButtons();
                }, SwingConstants.LEFT, Colors.encodeToHex(color), "#FFFFFFFFF",
                        I18n.text("The color encoded as either a 6 or 8 digit hex value")));
                return panel;
            }
        },
        HSB {
            @Override
            public String toString() {
                return I18n.text("HSB");
            }

            Panel createPanel(ColorChooser chooser, Color color) {
                Panel   panel = new Panel(new PrecisionLayout().setColumns(2).setMargins(0));
                float[] hsb   = Color.RGBtoHSB(color.getRed(), color.getGreen(), color.getBlue(), null);
                addLabel(panel, I18n.text("Hue"));
                panel.add(new EditorField(new DefaultFormatterFactory(new IntegerFormatter(0, 360, false)),
                        (f) -> {
                            float hue    = (((Integer) f.getValue()).intValue()) / 360.0f;
                            float sat    = ((Double) ((EditorField) panel.getComponent(3)).getValue()).floatValue();
                            float bright = ((Double) ((EditorField) panel.getComponent(5)).getValue()).floatValue();
                            chooser.mCurrent.setWellColor(new Color(Color.HSBtoRGB(hue, sat, bright) |
                                    (chooser.mCurrent.getWellColor().getAlpha() << 24), true));
                            chooser.adjustButtons();
                        }, SwingConstants.LEFT, Integer.valueOf((int) (hsb[0] * 360)),
                        Integer.valueOf(360),
                        I18n.text("The value to use for the hue, from 0 to 360")));
                DefaultFormatterFactory ff = new DefaultFormatterFactory(new DoubleFormatter(0, 1, false));
                addLabel(panel, I18n.text("Saturation"));
                panel.add(new EditorField(ff, (f) -> {
                    float hue    = (((Integer) ((EditorField) panel.getComponent(1)).getValue()).intValue()) / 360.0f;
                    float sat    = ((Double) f.getValue()).floatValue();
                    float bright = ((Double) ((EditorField) panel.getComponent(5)).getValue()).floatValue();
                    chooser.mCurrent.setWellColor(new Color(Color.HSBtoRGB(hue, sat, bright) |
                            (chooser.mCurrent.getWellColor().getAlpha() << 24), true));
                    chooser.adjustButtons();
                }, SwingConstants.LEFT, Double.valueOf(hsb[1]), Double.valueOf(0.99999),
                        I18n.text("The value to use for the saturation, from 0 to 1")));
                addLabel(panel, I18n.text("Brightness"));
                panel.add(new EditorField(ff, (f) -> {
                    float hue    = (((Integer) ((EditorField) panel.getComponent(1)).getValue()).intValue()) / 360.0f;
                    float sat    = ((Double) ((EditorField) panel.getComponent(3)).getValue()).floatValue();
                    float bright = ((Double) f.getValue()).floatValue();
                    chooser.mCurrent.setWellColor(new Color(Color.HSBtoRGB(hue, sat, bright) |
                            (chooser.mCurrent.getWellColor().getAlpha() << 24), true));
                    chooser.adjustButtons();
                }, SwingConstants.LEFT, Double.valueOf(hsb[2]), Double.valueOf(0.99999),
                        I18n.text("The value to use for the brightness, from 0 to 1")));
                addChannelField(chooser, panel, I18n.text("Alpha"), color.getAlpha(),
                        I18n.text("The value of the alpha channel, from 0 to 255"),
                        (c, v) -> new Color(c.getRed(), c.getGreen(), c.getBlue(), v));
                return panel;
            }
        };

        abstract Panel createPanel(ColorChooser chooser, Color color);

        void addChannelField(ColorChooser chooser, Panel panel, String label, int value, String tooltip, ColorAdjuster adjuster) {
            addLabel(panel, label);
            panel.add(new EditorField(FieldFactory.BYTE, (f) -> {
                chooser.mCurrent.setWellColor(adjuster.adjustColor(chooser.mCurrent.getWellColor(),
                        ((Integer) f.getValue()).intValue()));
                chooser.adjustButtons();
            }, SwingConstants.LEFT, Integer.valueOf(value), Integer.valueOf(255), tooltip));
        }
    }
}
