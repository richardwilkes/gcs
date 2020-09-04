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

package com.trollworks.gcs.ui;

import com.trollworks.gcs.ui.widget.BaseWindow;

import java.awt.Frame;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import java.awt.geom.AffineTransform;
import java.awt.image.BufferedImage;
import java.awt.print.PrinterGraphics;

/** Provides general graphics settings and manipulation. */
public final class GraphicsUtilities {
    private static boolean       HEADLESS_PRINT_MODE;
    private static BufferedImage FALLBACK_GRAPHICS_BACKING_STORE;

    private GraphicsUtilities() {
    }

    /** @return Whether the headless print mode is enabled. */
    public static boolean inHeadlessPrintMode() {
        return HEADLESS_PRINT_MODE;
    }

    /** @param inHeadlessPrintMode Whether the headless print mode is enabled. */
    public static void setHeadlessPrintMode(boolean inHeadlessPrintMode) {
        HEADLESS_PRINT_MODE = inHeadlessPrintMode;
    }

    /**
     * @param gc The {@link Graphics} to prepare for use.
     * @return The passed-in {@link Graphics2D}.
     */
    public static Graphics2D prepare(Graphics gc) {
        Graphics2D g2d = (Graphics2D) gc;
        g2d.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, RenderingHints.VALUE_TEXT_ANTIALIAS_ON);
        return g2d;
    }

    /**
     * @param gc The {@link Graphics2D} to configure.
     * @return The {@link RenderingHints} as they were prior to this call.
     */
    public static RenderingHints setMaximumQualityForGraphics(Graphics2D gc) {
        RenderingHints saved = (RenderingHints) gc.getRenderingHints().clone();
        gc.setRenderingHint(RenderingHints.KEY_ALPHA_INTERPOLATION, RenderingHints.VALUE_ALPHA_INTERPOLATION_QUALITY);
        gc.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_ON);
        gc.setRenderingHint(RenderingHints.KEY_COLOR_RENDERING, RenderingHints.VALUE_COLOR_RENDER_QUALITY);
        gc.setRenderingHint(RenderingHints.KEY_INTERPOLATION, RenderingHints.VALUE_INTERPOLATION_BICUBIC);
        gc.setRenderingHint(RenderingHints.KEY_RENDERING, RenderingHints.VALUE_RENDER_QUALITY);
        return saved;
    }

    /**
     * @return A graphics context obtained by looking for an existing window and asking it for a
     *         graphics context.
     */
    public static Graphics2D getGraphics() {
        Frame      frame = BaseWindow.getTopWindow();
        Graphics2D g2d   = frame == null ? null : (Graphics2D) frame.getGraphics();

        if (g2d == null) {
            Frame[] frames = Frame.getFrames();

            for (Frame element : frames) {
                if (element.isDisplayable()) {
                    g2d = (Graphics2D) element.getGraphics();
                    if (g2d != null) {
                        return g2d;
                    }
                }
            }
            BufferedImage fallback;
            synchronized (GraphicsUtilities.class) {
                if (FALLBACK_GRAPHICS_BACKING_STORE == null) {
                    FALLBACK_GRAPHICS_BACKING_STORE = new BufferedImage(32, 1, BufferedImage.TYPE_INT_ARGB);
                }
                fallback = FALLBACK_GRAPHICS_BACKING_STORE;
            }
            return fallback.createGraphics();
        }
        return prepare(g2d);
    }

    /**
     * On the Mac (and perhaps Windows now, too), the graphics context will have a scale transform
     * of 2x if being drawn onto a retina display.
     *
     * @param gc The {@link Graphics} to check.
     * @return {@code true} if the specified graphics context is set to a 2x scale transform or is a
     *         printer context.
     */
    public static boolean isRetinaDisplay(Graphics gc) {
        if (gc instanceof PrinterGraphics) {
            return true;
        }
        AffineTransform transform = ((Graphics2D) gc).getFontRenderContext().getTransform();
        return transform.getScaleX() == 2 && transform.getScaleY() == 2;
    }
}
