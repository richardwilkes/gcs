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

package com.trollworks.gcs.ui.image;

import com.trollworks.gcs.ui.GraphicsUtilities;

import java.awt.AlphaComposite;
import java.awt.Color;
import java.awt.Component;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.GraphicsConfiguration;
import java.awt.Image;
import java.awt.Transparency;
import java.awt.image.BufferedImage;
import java.awt.image.ColorModel;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;
import javax.imageio.ImageIO;
import javax.swing.Icon;

/**
 * Provides a {@link BufferedImage} that implements Swing's {@link Icon} interface for convenience.
 */
public class Img extends BufferedImage implements Icon {
    private Map<Object, Img> mDerived = new HashMap<>();

    /**
     * @param path The path to load the image from.
     * @return The image.
     */
    public static Img create(Path path) throws IOException {
        try (InputStream in = Files.newInputStream(path)) {
            return create(in);
        }
    }

    /**
     * @param file The file to load the image from.
     * @return The image.
     */
    public static Img create(File file) throws IOException {
        try (InputStream in = new FileInputStream(file)) {
            return create(in);
        }
    }

    /**
     * @param in The input stream to load the image from.
     * @return The image.
     */
    public static Img create(InputStream in) throws IOException {
        return create(ImageIO.read(in));
    }

    /**
     * @param in The input image.
     * @return A new image.
     */
    public static Img create(Image in) {
        Img        img = create(in.getWidth(null), in.getHeight(null), Transparency.TRANSLUCENT);
        Graphics2D gc  = img.getGraphics();
        gc.drawImage(in, 0, 0, null);
        gc.dispose();
        return img;
    }

    /**
     * @param width        The width to create.
     * @param height       The height to create.
     * @param transparency A constant from {@link Transparency}.
     * @return A new {@link Img} of the given size.
     */
    public static Img create(int width, int height, int transparency) {
        Graphics2D            g2d = GraphicsUtilities.getGraphics();
        GraphicsConfiguration gc  = g2d.getDeviceConfiguration();
        Img                   img = create(gc, width, height, transparency);
        g2d.dispose();
        return img;
    }

    /**
     * @param gc           The {@link GraphicsConfiguration} to make the image compatible with.
     * @param width        The width to create.
     * @param height       The height to create.
     * @param transparency A constant from {@link Transparency}.
     * @return A new {@link Img} of the given size.
     */
    public static Img create(GraphicsConfiguration gc, int width, int height, int transparency) {
        return new Img(gc.getColorModel(transparency), width, height);
    }

    private Img(ColorModel cm, int width, int height) {
        super(cm, cm.createCompatibleWritableRaster(width, height), cm.isAlphaPremultiplied(), null);
        Graphics2D gc = getGraphics();
        gc.setBackground(new Color(0, getTransparency() != OPAQUE));
        gc.clearRect(0, 0, getWidth(), getHeight());
        gc.dispose();
    }

    @Override
    public Graphics2D getGraphics() {
        Graphics2D gc = (Graphics2D) super.getGraphics();
        gc.setClip(0, 0, getWidth(), getHeight());
        return gc;
    }

    @Override
    public void paintIcon(Component component, Graphics gc, int x, int y) {
        gc.drawImage(this, x, y, component);
    }

    @Override
    public int getIconWidth() {
        return getWidth();
    }

    @Override
    public int getIconHeight() {
        return getHeight();
    }

    /**
     * Creates a scaled version of this image.
     *
     * @param width  The width to scale the image to.
     * @param height The height to scale the image to.
     * @return A new image.
     */
    public Img scale(int width, int height) {
        Img        buffer = create(width, height, getTransparency());
        Graphics2D gc     = buffer.getGraphics();
        GraphicsUtilities.setMaximumQualityForGraphics(gc);
        gc.drawImage(this, 0, 0, width, height, null);
        gc.dispose();
        return buffer;
    }

    /**
     * Creates a translucent version of this image.
     *
     * @param alpha The amount of alpha to use.
     * @return A new image.
     */
    public synchronized Img translucent(float alpha) {
        Float key = Float.valueOf(alpha);
        Img   img = mDerived.get(key);
        if (img == null) {
            img = create(getWidth(), getHeight(), Transparency.TRANSLUCENT);
            Graphics2D gc = img.getGraphics();
            gc.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC, alpha));
            gc.drawImage(this, 0, 0, null);
            gc.dispose();
            mDerived.put(key, img);
        }
        return img;
    }
}
