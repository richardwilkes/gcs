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

import com.trollworks.gcs.utility.Log;

import java.awt.Font;
import java.util.Iterator;
import javax.swing.text.BadLocationException;
import javax.swing.text.DefaultStyledDocument;
import javax.swing.text.Style;
import javax.swing.text.StyleConstants;

public class MarkdownDocument extends DefaultStyledDocument {
    public MarkdownDocument(String markdown) {
        Font             font     = Fonts.getDefaultFont();
        Style            body     = createBodyStyle(font);
        Style[]          headers  = createHeaderStyles(font);
        Style            bullet   = createBulletStyle(font);
        Iterator<String> iterator = markdown.lines().iterator();
        while (iterator.hasNext()) {
            String line  = iterator.next();
            Style  charStyle;
            Style  paraStyle;
            Style  style = null;
            for (int i = 1; i < 5; i++) {
                if (line.startsWith("#".repeat(i) + " ")) {
                    line = line.substring(i + 1);
                    style = headers[i - 1];
                    break;
                }
            }
            if (style == null) {
                if (line.startsWith("- ")) {
                    line = "•" + line.substring(1);
                    style = bullet;
                } else {
                    style = body;
                }
            }
            line += "\n";
            try {
                int start = getLength();
                insertString(start, line, null);
                setParagraphAttributes(start, getLength() - start, style, true);
            } catch (BadLocationException exception) {
                Log.error(exception);
            }
        }
        if (getLength() > 0) {
            try {
                remove(getLength() - 1, 1);
            } catch (BadLocationException exception) {
                Log.error(exception);
            }
        }
    }

    private Style createBodyStyle(Font font) {
        Style style = addStyle("body", null);
        style.addAttribute(StyleConstants.FontFamily, font.getFamily());
        style.addAttribute(StyleConstants.FontSize, Integer.valueOf(font.getSize()));
        return style;
    }

    private Style[] createHeaderStyles(Font font) {
        String  family = font.getFamily();
        int     size   = font.getSize();
        int[]   sizes  = {size * 2, size * 3 / 2, size * 5 / 4, size, size * 7 / 8};
        int     count  = sizes.length;
        Style[] styles = new Style[count];
        for (int i = 0; i < count; i++) {
            Style h = addStyle("h" + (i + 1), null);
            h.addAttribute(StyleConstants.FontFamily, family);
            h.addAttribute(StyleConstants.FontSize, Integer.valueOf(sizes[i]));
            h.addAttribute(StyleConstants.Bold, Boolean.TRUE);
            h.addAttribute(StyleConstants.SpaceBelow, Float.valueOf(sizes[i] / 2.0f));
            styles[i] = h;
        }
        return styles;
    }

    private Style createBulletStyle(Font font) {
        Style style = addStyle("bullet", null);
        style.addAttribute(StyleConstants.FontFamily, font.getFamily());
        style.addAttribute(StyleConstants.FontSize, Integer.valueOf(font.getSize()));
        int indent = TextDrawing.getSimpleWidth(font, "• ");
        style.addAttribute(StyleConstants.FirstLineIndent, Float.valueOf(-indent));
        style.addAttribute(StyleConstants.LeftIndent, Float.valueOf(indent));
        return style;
    }
}
