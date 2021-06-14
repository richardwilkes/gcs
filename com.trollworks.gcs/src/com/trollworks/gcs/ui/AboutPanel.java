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

package com.trollworks.gcs.ui;

import com.trollworks.gcs.GCS;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.StdPanel;
import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.LinearGradientPaint;
import java.awt.RenderingHints;
import javax.swing.UIManager;

/** The about box contents. */
public class AboutPanel extends StdPanel {
    private static final int MARGIN = 8;

    /** Creates a new about panel. */
    public AboutPanel() {
        setPreferredSize(new Dimension(Images.ABOUT.getIconWidth(), Images.ABOUT.getIconHeight()));
    }

    @Override
    protected void setStdColors() {
        setBackground(Color.BLACK);
        setForeground(Color.WHITE);
    }

    @Override
    protected void paintComponent(Graphics g) {
        Graphics2D gc = GraphicsUtilities.prepare(g);
        super.paintComponent(gc);
        Images.ABOUT.paintIcon(this, gc, 0, 0);
        //noinspection IntegerDivisionInFloatingPointContext
        gc.setPaint(new LinearGradientPaint(0, Images.ABOUT.getIconHeight() / 2, 0, Images.ABOUT.getIconHeight(), new float[]{0, 1}, new Color[]{Colors.TRANSPARENT, getBackground()}));
        gc.fillRect(0, 0, Images.ABOUT.getIconWidth(), Images.ABOUT.getIconHeight());
        RenderingHints saved = (RenderingHints) gc.getRenderingHints().clone();
        gc.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, RenderingHints.VALUE_TEXT_ANTIALIAS_ON);
        gc.setRenderingHint(RenderingHints.KEY_FRACTIONALMETRICS, RenderingHints.VALUE_FRACTIONALMETRICS_ON);
        Font baseFont = UIManager.getFont("TextField.font");
        gc.setFont(baseFont.deriveFont(10.0f));
        gc.setColor(getForeground());
        int right = getWidth() - MARGIN;
        int y     = draw(gc, I18n.text("GURPS is a trademark of Steve Jackson Games, used by permission. All rights reserved.\nThis product includes copyrighted material from the GURPS game, which is used by permission of Steve Jackson Games."), getHeight() - MARGIN, right, true, true);
        int y2    = draw(gc, GCS.COPYRIGHT + "\nAll rights reserved", y, right, false, true);
        draw(gc, String.format(I18n.text("%s %s\n%s Architecture\nJava %s"), System.getProperty("os.name"), System.getProperty("os.version"), System.getProperty("os.arch"), System.getProperty("java.version")), y, right, false, false);  //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
        gc.setFont(baseFont.deriveFont(Font.BOLD, 12.0f));
        draw(gc, GCS.VERSION.isZero() ? I18n.text("Development Version") : I18n.text("Version ") + GCS.VERSION, y2, right, false, true);
        gc.setRenderingHints(saved);
    }

    private static int draw(Graphics2D gc, String text, int y, int right, boolean addGap, boolean onLeft) {
        String[]    one     = text.split("\n");
        FontMetrics fm      = gc.getFontMetrics();
        int         fHeight = fm.getAscent() + fm.getDescent();
        for (int i = one.length - 1; i >= 0; i--) {
            gc.drawString(one[i], onLeft ? MARGIN : right - fm.stringWidth(one[i]), y);
            y -= fHeight;
        }
        if (addGap) {
            y -= fHeight / 2;
        }
        return y;
    }
}
