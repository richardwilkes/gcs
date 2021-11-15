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

package com.trollworks.gcs.datafile;

import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.menu.item.OpenPageReferenceCommand;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.TextCell;
import com.trollworks.gcs.utility.I18n;

import java.awt.Cursor;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.util.List;
import java.util.regex.Pattern;
import javax.swing.SwingConstants;

public class RefCell extends TextCell {
    public static final Pattern SEPARATORS_PATTERN = Pattern.compile("[, ;]");

    public static final String getStdToolTip(String type) {
        return String.format(I18n.text("A reference to the book and page this %s appears on (e.g. B22 would refer to \"Basic Set\", page 22)"), type);
    }

    public static final String getStdCellToolTip(String text) {
        return (SEPARATORS_PATTERN.split(text, 2).length == 1) ? null : text;
    }

    public RefCell() {
        super(SwingConstants.LEFT, false);
    }

    @Override
    protected String getPresentationText(Outline outline, Row row, Column column) {
        String[] parts = SEPARATORS_PATTERN.split(getData(row, column), 2);
        return parts.length == 1 ? parts[0] : parts[0] + "+";
    }

    @Override
    public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
        if (row instanceof HasSourceReference srcRef) {
            List<String>       refs   = OpenPageReferenceCommand.getReferences(srcRef);
            if (!refs.isEmpty()) {
                OpenPageReferenceCommand.openReference(refs.get(0), (srcRef).getReferenceHighlight());
            }
        }
    }

    @Override
    public Cursor getCursor(MouseEvent event, Rectangle bounds, Row row, Column column) {
        if (row instanceof HasSourceReference srcRef) {
            List<String>       refs   = OpenPageReferenceCommand.getReferences(srcRef);
            if (!refs.isEmpty()) {
                return Cursor.getPredefinedCursor(Cursor.HAND_CURSOR);
            }
        }
        return Cursor.getDefaultCursor();
    }
}
