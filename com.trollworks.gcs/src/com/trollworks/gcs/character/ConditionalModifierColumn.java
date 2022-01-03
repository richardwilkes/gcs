/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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

public enum ConditionalModifierColumn {
    MODIFIER {
        @Override
        public String toString() {
            return I18n.text("Modifier");
        }

        @Override
        public Object getData(ConditionalModifierRow row) {
            return Integer.valueOf(row.getTotalAmount());
        }

        @Override
        public String getDataAsText(ConditionalModifierRow row) {
            return Numbers.formatWithForcedSign(row.getTotalAmount());
        }

        @Override
        public Cell getCell(boolean forEditor) {
            if (forEditor) {
                return new TextCell(SwingConstants.RIGHT, false);
            }
            return new ListTextCell(SwingConstants.RIGHT, false);
        }
    },
    CONDITION {
        @Override
        public String toString() {
            return I18n.text("Condition");
        }

        @Override
        public Object getData(ConditionalModifierRow row) {
            return getDataAsText(row);
        }

        @Override
        public String getDataAsText(ConditionalModifierRow row) {
            return row.getFrom();
        }

        @Override
        public Cell getCell(boolean forEditor) {
            return new WrappedCell();
        }
    };

    public abstract Object getData(ConditionalModifierRow row);

    public abstract String getDataAsText(ConditionalModifierRow row);

    public abstract Cell getCell(boolean forEditor);

    public String getToolTip(ConditionalModifierRow row) {
        List<Integer> bonuses = row.getAmounts();
        List<String>  sources = row.getSources();
        StringBuilder buffer  = new StringBuilder();
        int           count   = bonuses.size();
        for (int i = 0; i < count; i++) {
            if (i != 0) {
                buffer.append('\n');
            }
            buffer.append(Numbers.formatWithForcedSign(bonuses.get(i).longValue()));
            buffer.append(" ");
            buffer.append(sources.get(i));
        }
        return buffer.toString();

    }

    public static void addColumns(Outline outline, boolean forEditor) {
        OutlineModel model = outline.getModel();
        for (ConditionalModifierColumn one : values()) {
            Column column = new Column(one.ordinal(), one.toString(), null, one.getCell(forEditor));
            if (!forEditor) {
                column.setHeaderCell(new ListHeaderCell(true));
            }
            model.addColumn(column);
        }
    }
}
