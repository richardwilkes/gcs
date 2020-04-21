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

import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Component;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import javax.swing.Icon;

public class RetinaIcon implements Icon {
    private Img mNormal;
    private Img mRetina;

    public RetinaIcon(Img normal, Img retina) {
        mNormal = normal;
        mRetina = retina;
    }

    public Img getNormal() {
        return mNormal;
    }

    public Img getRetina() {
        return mRetina;
    }

    @Override
    public void paintIcon(Component component, Graphics g, int x, int y) {
        Graphics2D     gc    = (Graphics2D) g;
        RenderingHints saved = GraphicsUtilities.setMaximumQualityForGraphics(gc);
        Scale          scale = Scale.get(component);
        Img            img   = mRetina != null && (scale.getScale() > 1 || GraphicsUtilities.isRetinaDisplay(g)) ? mRetina : mNormal;
        gc.drawImage(img, x, y, scale.scale(getIconWidth()), scale.scale(getIconHeight()), component);
        gc.setRenderingHints(saved);
    }

    @Override
    public int getIconWidth() {
        return mNormal.getWidth();
    }

    @Override
    public int getIconHeight() {
        return mNormal.getHeight();
    }

    public RetinaIcon createDisabled() {
        return new RetinaIcon(mNormal.translucent(0.3f), mRetina != null ? mRetina.translucent(0.3f) : null);
    }
}
