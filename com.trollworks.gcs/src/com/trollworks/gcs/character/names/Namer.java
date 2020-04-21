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

package com.trollworks.gcs.character.names;

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.Alignment;
import com.trollworks.gcs.ui.layout.FlexColumn;
import com.trollworks.gcs.ui.layout.FlexComponent;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.layout.LayoutSize;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JTextField;
import javax.swing.SwingConstants;

/** Asks the user to name items that have been marked to be customized. */
public class Namer extends JPanel {
    private ListRow          mRow;
    private List<JTextField> mFields;

    /**
     * Brings up a modal naming dialog for each row in the list.
     *
     * @param owner The owning component.
     * @param list  The rows to name.
     * @return Whether anything was modified.
     */
    public static boolean name(Component owner, List<ListRow> list) {
        List<ListRow>         rowList  = new ArrayList<>();
        List<HashSet<String>> setList  = new ArrayList<>();
        boolean               modified = false;
        int                   count;

        for (ListRow row : list) {
            HashSet<String> set = new HashSet<>();

            row.fillWithNameableKeys(set);
            if (!set.isEmpty()) {
                rowList.add(row);
                setList.add(set);
            }
        }

        count = rowList.size();
        for (int i = 0; i < count; i++) {
            ListRow  row     = rowList.get(i);
            boolean  hasMore = i != count - 1;
            int      type    = hasMore ? JOptionPane.YES_NO_CANCEL_OPTION : JOptionPane.YES_NO_OPTION;
            String[] options = hasMore ? new String[]{I18n.Text("Apply"), I18n.Text("Cancel"), I18n.Text("Cancel Remaining")} : new String[]{I18n.Text("Apply"), I18n.Text("Cancel")};
            Namer    panel   = new Namer(row, setList.get(i), count - i - 1);
            switch (WindowUtils.showOptionDialog(owner, panel, MessageFormat.format(I18n.Text("Name {0}"), row.getLocalizedName()), true, type, JOptionPane.PLAIN_MESSAGE, row.getIcon(true), options, I18n.Text("Apply"))) {
            case JOptionPane.YES_OPTION:
                panel.applyChanges();
                modified = true;
                break;
            case JOptionPane.NO_OPTION:
                break;
            case JOptionPane.CANCEL_OPTION:
            case JOptionPane.CLOSED_OPTION:
            default:
                return modified;
            }
        }
        return modified;
    }

    private Namer(ListRow row, Set<String> set, int remaining) {
        JLabel label;
        mRow = row;
        mFields = new ArrayList<>();

        FlexColumn column = new FlexColumn();
        if (remaining > 0) {
            label = new JLabel(remaining == 1 ? I18n.Text("1 item remaining to be named.") : MessageFormat.format(I18n.Text("{0} items remaining to be named."), Integer.valueOf(remaining)), SwingConstants.CENTER);
            Dimension size = label.getMaximumSize();
            size.width = LayoutSize.MAXIMUM_SIZE;
            label.setMaximumSize(size);
            add(label);
            column.add(label);
        }
        label = new JLabel(Text.truncateIfNecessary(row.toString(), 80, SwingConstants.RIGHT), SwingConstants.CENTER);
        Dimension size = label.getMaximumSize();
        size.width = LayoutSize.MAXIMUM_SIZE;
        size.height += 4;
        label.setMaximumSize(size);
        size = label.getPreferredSize();
        size.height += 4;
        label.setPreferredSize(size);
        label.setMinimumSize(size);
        label.setBackground(Color.BLACK);
        label.setForeground(Color.WHITE);
        label.setOpaque(true);
        add(label);
        column.add(label);
        column.add(new FlexSpacer(0, 10, false, false));

        int      rowIndex = 0;
        FlexGrid grid     = new FlexGrid();
        grid.setFillHorizontal(true);
        List<String> list = new ArrayList<>(set);
        Collections.sort(list);
        for (String name : list) {
            JTextField field = new JTextField(25);
            field.setName(name);
            size = field.getPreferredSize();
            size.width = LayoutSize.MAXIMUM_SIZE;
            field.setMaximumSize(size);
            mFields.add(field);
            label = new JLabel(name, SwingConstants.RIGHT);
            UIUtilities.setToPreferredSizeOnly(label);
            add(label);
            add(field);
            grid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, Alignment.CENTER), rowIndex, 0);
            grid.add(field, rowIndex++, 1);
        }

        column.add(grid);
        column.add(new FlexSpacer(0, 0, false, true));
        column.apply(this);
    }

    private void applyChanges() {
        Commitable.sendCommitToFocusOwner();
        Map<String, String> map = new HashMap<>();
        for (JTextField field : mFields) {
            map.put(field.getName(), field.getText());
        }
        mRow.applyNameableKeys(map);
    }
}
