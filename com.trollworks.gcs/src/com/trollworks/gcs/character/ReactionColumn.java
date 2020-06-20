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

package com.trollworks.gcs.character;

import com.trollworks.gcs.ui.widget.outline.Cell;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListHeaderCell;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.TextCell;
import com.trollworks.gcs.ui.widget.outline.WrappedCell;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import java.util.List;
import javax.swing.SwingConstants;

public enum ReactionColumn {
    MODIFIER {
        @Override
        public String toString() {
            return I18n.Text("Modifier");
        }

        @Override
        public Object getData(ReactionRow row) {
            return Integer.valueOf(row.getTotalAmount());
        }

        @Override
        public String getDataAsText(ReactionRow row) {
            return Numbers.formatWithForcedSign(row.getTotalAmount());
        }

        @Override
        public Cell getCell(boolean forEditor) {
            if (forEditor) {
                return new TextCell(SwingConstants.RIGHT, false);
            }
            return new ListTextCell(SwingConstants.RIGHT, false);
        }
    }, REACTION {
        @Override
        public String toString() {
            return I18n.Text("Reaction");
        }

        @Override
        public Object getData(ReactionRow row) {
            return getDataAsText(row);
        }

        @Override
        public String getDataAsText(ReactionRow row) {
            return row.getFrom();
        }

        @Override
        public Cell getCell(boolean forEditor) {
            return new WrappedCell();
        }
    };

    public abstract Object getData(ReactionRow row);

    public abstract String getDataAsText(ReactionRow row);

    public abstract Cell getCell(boolean forEditor);

    public String getToolTip(ReactionRow row) {
        List<Integer> bonuses = row.getAmounts();
        List<String>  sources = row.getSources();
        int           count   = bonuses.size();
        if (count == 1) {
            return Numbers.formatWithForcedSign(bonuses.get(0).longValue()) + " " + sources.get(0);
        }
        StringBuilder buffer = new StringBuilder();
        buffer.append("<html><body><table>");
        for (int i = 0; i < count; i++) {
            buffer.append(String.format("<tr><td>&bull;</td><td align='right'>%s</td><td>%s</td></tr>", Numbers.formatWithForcedSign(bonuses.get(i).longValue()), sources.get(i)));
        }
        buffer.append("</table></body></html>");
        return buffer.toString();

    }

    public static void addColumns(Outline outline, boolean forEditor) {
        OutlineModel model = outline.getModel();
        for (ReactionColumn one : values()) {
            Column column = new Column(one.ordinal(), one.toString(), null, one.getCell(forEditor));
            if (!forEditor) {
                column.setHeaderCell(new ListHeaderCell(true));
            }
            model.addColumn(column);
        }
    }
}
