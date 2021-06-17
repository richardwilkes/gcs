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

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.font.FontRenderContext;
import java.util.ArrayList;
import java.util.List;
import java.util.StringTokenizer;
import javax.swing.SwingConstants;

/** General text drawing utilities. */
public final class TextDrawing {
    private static FontRenderContext DEFAULT_FONT_RENDER_CONTEXT;

    private TextDrawing() {
    }

    /**
     * @param font The font to work on.
     * @return The font metrics for the specified font.
     */
    public static FontMetrics getFontMetrics(Font font) {
        Graphics2D  g2d = GraphicsUtilities.getGraphics();
        FontMetrics fm  = g2d.getFontMetrics(font);
        g2d.dispose();
        return fm;
    }

    /** @return A default {@link FontRenderContext}. */
    public static FontRenderContext getDefaultFontRenderContext() {
        Graphics2D        g2d = GraphicsUtilities.getGraphics();
        FontRenderContext frc = g2d.getFontRenderContext();
        g2d.dispose();
        return frc;
    }

    /**
     * @param font The {@link Font} to measure with.
     * @param text The text to measure. Returns are not treated specially.
     * @return The width, in pixels.
     */
    public static int getSimpleWidth(Font font, String text) {
        if (text == null || text.isEmpty()) {
            return 0;
        }
        Font              fallbackFont = null;
        char[]            chars        = text.toCharArray();
        int               max          = chars.length;
        FontRenderContext frc          = getDefaultFontRenderContext();
        int               width        = 0;
        int               i            = 0;
        while (i < max) {
            int upTo = font.canDisplayUpTo(chars, i, max);
            if (upTo != 0) {
                width += (int) Math.ceil(font.getStringBounds(chars, i, upTo == -1 ? max : upTo, frc).getWidth());
                if (upTo == -1) {
                    break;
                }
            }
            int j = upTo + 1;
            while (j < max && !font.canDisplay(chars[j])) {
                j++;
            }
            if (fallbackFont == null) {
                fallbackFont = new Font(Font.SANS_SERIF, font.getStyle(), font.getSize());
            }
            width += (int) Math.ceil(fallbackFont.getStringBounds(chars, upTo, j, frc).getWidth());
            i = j;
        }
        return width;
    }

    /**
     * @param font The font the text will be in.
     * @param text The text to calculate a size for. May contain returns. The width of the widest
     *             line will be returned.
     * @return The width of the text in the specified font.
     */
    public static int getWidth(Font font, String text) {
        if (text == null || text.isEmpty()) {
            return 0;
        }
        StringTokenizer tokenizer = new StringTokenizer(text, "\n", true);
        boolean         veryFirst = true;
        boolean         first     = true;
        int             width     = 0;
        while (tokenizer.hasMoreTokens()) {
            String token = tokenizer.nextToken();
            if ("\n".equals(token)) {
                if (first && !veryFirst) {
                    first = false;
                    continue;
                }
                token = " ";
            } else {
                first = true;
            }
            veryFirst = false;
            int bWidth = getSimpleWidth(font, token);
            if (width < bWidth) {
                width = bWidth;
            }
        }
        return width;
    }

    /**
     * Draws the text. Embedded return characters may be present.
     *
     * @param gc     The graphics context.
     * @param bounds The bounding rectangle to draw the text within.
     * @param text   The text to draw.
     * @param hAlign The horizontal alignment to use. One of {@link SwingConstants#LEFT}, {@link
     *               SwingConstants#CENTER}, or {@link SwingConstants#RIGHT}.
     * @param vAlign The vertical alignment to use. One of {@link SwingConstants#LEFT}, {@link
     *               SwingConstants#CENTER}, or {@link SwingConstants#RIGHT}.
     * @return The bottom of the drawn text.
     */
    public static int draw(Graphics2D gc, Rectangle bounds, String text, int hAlign, int vAlign) {
        return draw(gc, bounds, text, hAlign, vAlign, null, 0);
    }

    /**
     * Draws the text. Embedded return characters may be present.
     *
     * @param gc              The graphics context.
     * @param bounds          The bounding rectangle to draw the text within.
     * @param text            The text to draw.
     * @param hAlign          The horizontal alignment to use. One of {@link SwingConstants#LEFT},
     *                        {@link SwingConstants#CENTER}, or {@link SwingConstants#RIGHT}.
     * @param vAlign          The vertical alignment to use. One of {@link SwingConstants#LEFT},
     *                        {@link SwingConstants#CENTER}, or {@link SwingConstants#RIGHT}.
     * @param strikeThruColor If not {@code null}, then a line of this color will be drawn through
     *                        the text.
     * @param strikeThruSize  The line width to use when drawing the strike-thru.
     * @return The bottom of the drawn text.
     */
    public static int draw(Graphics2D gc, Rectangle bounds, String text, int hAlign, int vAlign, Color strikeThruColor, int strikeThruSize) {
        int y = bounds.y;
        if (!text.isEmpty()) {
            List<String>    list       = new ArrayList<>();
            Font            font       = gc.getFont();
            FontMetrics     fm         = gc.getFontMetrics();
            int             ascent     = fm.getAscent();
            int             descent    = fm.getDescent();
            int             fHeight    = ascent + descent;
            StringTokenizer tokenizer  = new StringTokenizer(text, " \n", true);
            StringBuilder   buffer     = new StringBuilder(text.length());
            int             textHeight = 0;
            int             width;
            while (tokenizer.hasMoreTokens()) {
                String token = tokenizer.nextToken();
                if ("\n".equals(token)) {
                    text = buffer.toString();
                    textHeight += fHeight;
                    list.add(text);
                    buffer.setLength(0);
                } else {
                    width = getSimpleWidth(font, buffer + token);
                    if (width > bounds.width && !buffer.isEmpty()) {
                        text = buffer.toString();
                        textHeight += fHeight;
                        list.add(text);
                        buffer.setLength(0);
                        if (" ".equals(token)) {
                            continue;
                        }
                    }
                    buffer.append(token);
                }
            }
            if (!buffer.isEmpty()) {
                text = buffer.toString();
                textHeight += fHeight;
                list.add(text);
            }
            if (vAlign == SwingConstants.CENTER) {
                y = bounds.y + (bounds.height - textHeight) / 2;
            } else if (vAlign == SwingConstants.BOTTOM) {
                y = bounds.y + bounds.height - textHeight;
            }
            for (String piece : list) {
                width = 0;
                int x = bounds.x;
                if (hAlign == SwingConstants.CENTER) {
                    width = getSimpleWidth(font, piece);
                    x += (bounds.width - width) / 2;
                } else if (hAlign == SwingConstants.RIGHT) {
                    width = getSimpleWidth(font, piece);
                    x += bounds.width - (1 + width);
                }
                drawString(gc, piece, x, y + ascent);
                if (strikeThruColor != null) {
                    Color saved = gc.getColor();
                    gc.setColor(strikeThruColor);
                    if (width == 0) {
                        width = getSimpleWidth(font, piece);
                    }
                    gc.fillRect(x, y + (ascent - strikeThruSize) / 2, x + width, strikeThruSize);
                    gc.setColor(saved);
                }
                y += fHeight;
            }
        }
        return y;
    }

    private static void drawString(Graphics2D gc, String text, int x, int y) {
        Font              font         = gc.getFont();
        Font              fallbackFont = null;
        char[]            chars        = text.toCharArray();
        int               max          = chars.length;
        FontRenderContext frc          = getDefaultFontRenderContext();
        int               i            = 0;
        while (i < max) {
            int upTo = font.canDisplayUpTo(chars, i, max);
            if (upTo != 0) {
                gc.drawString(new String(chars, i, (upTo == -1 ? max : upTo) - i), x, y);
                if (upTo == -1) {
                    break;
                }
                x += (int) Math.ceil(font.getStringBounds(chars, i, upTo, frc).getWidth());
            }
            int j = upTo + 1;
            while (j < max && !font.canDisplay(chars[j])) {
                j++;
            }
            if (fallbackFont == null) {
                fallbackFont = new Font(Font.SANS_SERIF, font.getStyle(), font.getSize());
            }
            gc.setFont(fallbackFont);
            gc.drawString(new String(chars, upTo, j - upTo), x, y);
            gc.setFont(font);
            x += (int) Math.ceil(fallbackFont.getStringBounds(chars, upTo, j, frc).getWidth());
            i = j;
        }
    }

    /**
     * Embedded return characters may be present.
     *
     * @param font The font the text will be in.
     * @param text The text to calculate a size for.
     * @return The preferred size of the text in the specified font.
     */
    public static Dimension getPreferredSize(Font font, String text) {
        if (text == null || text.isEmpty()) {
            return new Dimension();
        }
        StringTokenizer tokenizer = new StringTokenizer(text, "\n", true);
        boolean         veryFirst = true;
        boolean         first     = true;
        int             fHeight   = getFontHeight(font);
        int             width     = 0;
        int             height    = 0;
        while (tokenizer.hasMoreTokens()) {
            String token = tokenizer.nextToken();
            if ("\n".equals(token)) {
                height += fHeight;
                if (first && !veryFirst) {
                    first = false;
                    continue;
                }
                token = " ";
            } else {
                first = true;
            }
            veryFirst = false;
            int bWidth = getSimpleWidth(font, token);
            if (width < bWidth) {
                width = bWidth;
            }
        }
        if (!text.endsWith("\n")) {
            height += fHeight;
        }
        return new Dimension(width, height);
    }

    public static int getFontHeight(Font font) {
        FontMetrics fm = getFontMetrics(font);
        return fm.getAscent() + fm.getDescent();
    }

    /**
     * Embedded return characters may be present.
     *
     * @param font The font the text will be in.
     * @param text The text to calculate a size for.
     * @return The preferred height of the text in the specified font.
     */
    public static int getPreferredHeight(Font font, String text) {
        int height = 0;
        int length = text.length();
        if (length > 0) {
            int  fHeight = getFontHeight(font);
            char ch      = 0;
            for (int i = 0; i < length; i++) {
                ch = text.charAt(i);
                if (ch == '\n') {
                    height += fHeight;
                }
            }
            if (ch != '\n') {
                height += fHeight;
            }
        }
        return height;
    }

    /**
     * If the text doesn't fit in the specified width, it will be shortened and an ellipse ("...")
     * will be added. This method does not work properly on text with embedded line endings.
     *
     * @param font             The font to use.
     * @param text             The text to work on.
     * @param width            The maximum pixel width.
     * @param truncationPolicy One of {@link SwingConstants#LEFT}, {@link SwingConstants#CENTER}, or
     *                         {@link SwingConstants#RIGHT}.
     * @return The adjusted text.
     */
    public static String truncateIfNecessary(Font font, String text, int width, int truncationPolicy) {
        if (getSimpleWidth(font, text) > width) {
            StringBuilder buffer = new StringBuilder(text);
            int           max    = buffer.length();
            if (truncationPolicy == SwingConstants.LEFT) {
                buffer.insert(0, '…');
                while (max-- > 0 && getSimpleWidth(font, buffer.toString()) > width) {
                    buffer.deleteCharAt(1);
                }
            } else if (truncationPolicy == SwingConstants.CENTER) {
                int     left     = max / 2;
                int     right    = left + 1;
                boolean leftSide = false;
                buffer.insert(left--, '…');
                while (max-- > 0 && getSimpleWidth(font, buffer.toString()) > width) {
                    if (leftSide) {
                        buffer.deleteCharAt(left--);
                        if (--right < max + 1) {
                            leftSide = false;
                        }
                    } else {
                        buffer.deleteCharAt(right);
                        if (left >= 0) {
                            leftSide = true;
                        }
                    }
                }
            } else if (truncationPolicy == SwingConstants.RIGHT) {
                buffer.append('…');
                while (max-- > 0 && getSimpleWidth(font, buffer.toString()) > width) {
                    buffer.deleteCharAt(max);
                }
            }
            text = buffer.toString();
        }
        return text;
    }

    /**
     * @param font  The font to use.
     * @param text  The text to wrap.
     * @param width The maximum pixel width to allow.
     * @return A new, wrapped version of the text.
     */
    public static String wrapToPixelWidth(Font font, String text, int width) {
        int[]           lineWidth  = {0};
        StringBuilder   buffer     = new StringBuilder(text.length() * 2);
        StringBuilder   lineBuffer = new StringBuilder(text.length());
        StringTokenizer tokenizer  = new StringTokenizer(text + "\n", " \t/\\\n", true);
        boolean         wrapped    = false;
        while (tokenizer.hasMoreTokens()) {
            String token = tokenizer.nextToken();
            if ("\n".equals(token)) {
                if (lineWidth[0] > 0) {
                    buffer.append(lineBuffer);
                }
                buffer.append(token);
                wrapped = false;
                lineBuffer.setLength(0);
                lineWidth[0] = 0;
            } else {
                if (!wrapped || lineWidth[0] != 0 || !" ".equals(token)) {
                    wrapped = processOneTokenForWrapToPixelWidth(token, font, buffer, lineBuffer, width, lineWidth, wrapped);
                }
            }
        }
        if (lineWidth[0] > 0) {
            buffer.append(lineBuffer);
        }
        buffer.setLength(buffer.length() - 1);
        return buffer.toString();
    }

    private static boolean processOneTokenForWrapToPixelWidth(String token, Font font, StringBuilder buffer, StringBuilder lineBuffer, int width, int[] lineWidth, boolean hasBeenWrapped) {
        int tokenWidth = getSimpleWidth(font, token);
        if (lineWidth[0] + tokenWidth <= width) {
            lineBuffer.append(token);
            lineWidth[0] += tokenWidth;
        } else if (lineWidth[0] == 0) {
            // Special-case a line that has not had anything put on it yet
            int count = token.length();
            lineBuffer.append(token.charAt(0));
            for (int i = 1; i < count; i++) {
                lineBuffer.append(token.charAt(i));
                if (getSimpleWidth(font, lineBuffer.toString()) > width) {
                    lineBuffer.deleteCharAt(lineBuffer.length() - 1);
                    buffer.append(lineBuffer);
                    buffer.append("\n");
                    hasBeenWrapped = true;
                    lineBuffer.setLength(0);
                    lineBuffer.append(token.charAt(i));
                }
            }
            lineWidth[0] = getSimpleWidth(font, lineBuffer.toString());
        } else {
            buffer.append(lineBuffer);
            buffer.append("\n");
            hasBeenWrapped = true;
            lineBuffer.setLength(0);
            lineWidth[0] = 0;
            if (!" ".equals(token)) {
                return processOneTokenForWrapToPixelWidth(token, font, buffer, lineBuffer, width, lineWidth, true);
            }
        }
        return hasBeenWrapped;
    }
}
