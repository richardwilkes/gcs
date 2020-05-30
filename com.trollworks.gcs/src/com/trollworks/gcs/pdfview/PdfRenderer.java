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

package com.trollworks.gcs.pdfview;

import com.trollworks.gcs.utility.DummyWriter;
import com.trollworks.gcs.utility.Log;

import java.awt.BasicStroke;
import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.Shape;
import java.awt.Toolkit;
import java.awt.geom.AffineTransform;
import java.awt.geom.Rectangle2D;
import java.awt.image.BufferedImage;
import java.io.IOException;
import java.util.List;

import org.apache.fontbox.util.BoundingBox;
import org.apache.pdfbox.pdmodel.PDDocument;
import org.apache.pdfbox.pdmodel.PDPage;
import org.apache.pdfbox.pdmodel.common.PDRectangle;
import org.apache.pdfbox.pdmodel.font.PDFont;
import org.apache.pdfbox.pdmodel.font.PDType3Font;
import org.apache.pdfbox.pdmodel.graphics.blend.BlendComposite;
import org.apache.pdfbox.pdmodel.graphics.blend.BlendMode;
import org.apache.pdfbox.rendering.PDFRenderer;
import org.apache.pdfbox.text.PDFTextStripper;
import org.apache.pdfbox.text.TextPosition;

public class PdfRenderer extends PDFTextStripper {
    private Graphics2D mGC;
    private String     mTextToHighlight;

    public static BufferedImage create(PDDocument pdf, int pageIndex, float scale, String textToHighlight) {
        try {
            PDFRenderer renderer = new PDFRenderer(pdf);
            scale = scale * Toolkit.getDefaultToolkit().getScreenResolution() / 72.0f;
            BufferedImage img = renderer.renderImage(pageIndex, scale);
            if (textToHighlight != null) {
                Graphics2D gc = img.createGraphics();
                gc.setStroke(new BasicStroke(0.1f));
                gc.scale(scale, scale);
                PdfRenderer processor = new PdfRenderer(gc, textToHighlight);
                processor.setSortByPosition(true);
                processor.setStartPage(pageIndex + 1);
                processor.setEndPage(pageIndex + 1);
                try (DummyWriter writer = new DummyWriter()) {
                    processor.writeText(pdf, writer);
                }
                gc.dispose();
            }
            return img;
        } catch (Exception exception) {
            Log.error(exception);
            return null;
        }
    }

    private PdfRenderer(Graphics2D gc, String textToHighlight) throws IOException {
        mGC = gc;
        mGC.setColor(Color.YELLOW);
        mGC.setComposite(BlendComposite.getInstance(BlendMode.MULTIPLY, 0.3f));
        mTextToHighlight = textToHighlight.toLowerCase();
    }

    @Override
    protected void writeString(String text, List<TextPosition> textPositions) throws IOException {
        text = text.toLowerCase();
        int index = text.indexOf(mTextToHighlight);
        if (index != -1) {
            PDPage          currentPage     = getCurrentPage();
            PDRectangle     pageBoundingBox = currentPage.getBBox();
            AffineTransform flip            = new AffineTransform();
            flip.translate(0, pageBoundingBox.getHeight());
            flip.scale(1, -1);
            PDRectangle mediaBox    = currentPage.getMediaBox();
            float       mediaHeight = mediaBox.getHeight();
            float       mediaWidth  = mediaBox.getWidth();
            int         size        = textPositions.size();
            while (index != -1) {
                int last = index + mTextToHighlight.length() - 1;
                for (int i = index; i <= last; i++) {
                    TextPosition      pos  = textPositions.get(i);
                    PDFont            font = pos.getFont();
                    BoundingBox       bbox = font.getBoundingBox();
                    Rectangle2D.Float rect = new Rectangle2D.Float(0, bbox.getLowerLeftY(), font.getWidth(pos.getCharacterCodes()[0]), bbox.getHeight());
                    AffineTransform   at   = pos.getTextMatrix().createAffineTransform();
                    if (font instanceof PDType3Font) {
                        at.concatenate(font.getFontMatrix().createAffineTransform());
                    } else {
                        at.scale(1 / 1000.0f, 1 / 1000.0f);
                    }
                    Shape           shape     = flip.createTransformedShape(at.createTransformedShape(rect));
                    AffineTransform transform = mGC.getTransform();
                    int             rotation  = currentPage.getRotation();
                    if (rotation != 0) {
                        switch (rotation) {
                        case 90:
                            mGC.translate(mediaHeight, 0);
                            break;
                        case 270:
                            mGC.translate(0, mediaWidth);
                            break;
                        case 180:
                            mGC.translate(mediaWidth, mediaHeight);
                            break;
                        default:
                            break;
                        }
                        mGC.rotate(Math.toRadians(rotation));
                    }
                    mGC.fill(shape);
                    if (rotation != 0) {
                        mGC.setTransform(transform);
                    }
                }
                index = last < size - 1 ? text.indexOf(mTextToHighlight, last + 1) : -1;
            }
        }
    }
}
