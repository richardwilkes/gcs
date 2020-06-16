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
        Font   defaultFont = Fonts.getDefaultFont();
        String family      = defaultFont.getFamily();
        int    size        = defaultFont.getSize();
        Style  body        = addStyle("body", null);
        body.addAttribute(StyleConstants.FontFamily, family);
        body.addAttribute(StyleConstants.FontSize, Integer.valueOf(size));
        Style h2 = addStyle("h2", null);
        h2.addAttribute(StyleConstants.FontFamily, family);
        h2.addAttribute(StyleConstants.FontSize, Integer.valueOf(size + size / 3));
        h2.addAttribute(StyleConstants.Bold, Boolean.TRUE);
        Style h3 = addStyle("h3", null);
        h3.addAttribute(StyleConstants.FontFamily, family);
        h3.addAttribute(StyleConstants.FontSize, Integer.valueOf(size + size / 5));
        h3.addAttribute(StyleConstants.Bold, Boolean.TRUE);
        Style h3Paragraph = addStyle("h3P", null);
        int   spacing     = size * 2 / 3;
        h3Paragraph.addAttribute(StyleConstants.SpaceAbove, Float.valueOf(spacing));
        h3Paragraph.addAttribute(StyleConstants.SpaceBelow, Float.valueOf(spacing));
        Style bullet = addStyle("bullet", null);
        bullet.addAttribute(StyleConstants.FontFamily, family);
        bullet.addAttribute(StyleConstants.FontSize, Integer.valueOf(size));
        Style bulletParagraph = addStyle("bulletP", null);
        int   indent          = TextDrawing.getSimpleWidth(defaultFont, "• ");
        bulletParagraph.addAttribute(StyleConstants.FirstLineIndent, Float.valueOf(-indent));
        bulletParagraph.addAttribute(StyleConstants.LeftIndent, Float.valueOf(indent));
        Iterator<String> iterator = markdown.lines().iterator();
        while (iterator.hasNext()) {
            String line = iterator.next();
            Style  charStyle;
            Style  paraStyle;
            if (line.startsWith("## ")) {
                line = line.substring(3);
                charStyle = h2;
                paraStyle = null;
            } else if (line.startsWith("### ")) {
                line = line.substring(4);
                charStyle = h3;
                paraStyle = h3Paragraph;
            } else if (line.startsWith("- ")) {
                line = "•" + line.substring(1);
                charStyle = bullet;
                paraStyle = bulletParagraph;
            } else {
                charStyle = body;
                paraStyle = null;
            }
            line += "\n";
            try {
                int start = getLength();
                insertString(start, line, charStyle);
                if (paraStyle != null) {
                    setParagraphAttributes(start, getLength() - start, paraStyle, false);
                }
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
}
