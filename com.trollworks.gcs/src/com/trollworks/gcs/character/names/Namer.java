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

package com.trollworks.gcs.character.names;

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.Separator;
import com.trollworks.gcs.ui.widget.Button;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Component;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** Asks the user to name items that have been marked to be customized. */
public final class Namer extends Panel implements DocumentListener {
    private ListRow           mRow;
    private List<EditorField> mFields;
    private Button            mApplyButton;

    /**
     * Brings up a modal naming dialog for each row in the list.
     *
     * @param owner The owning component.
     * @param list  The rows to name.
     * @return Whether anything was modified.
     */
    public static boolean name(Component owner, List<ListRow> list) {
        List<ListRow>     rowList  = new ArrayList<>();
        List<Set<String>> setList  = new ArrayList<>();
        boolean           modified = false;
        int               count;
        for (ListRow row : list) {
            Set<String> set = new HashSet<>();
            row.fillWithNameableKeys(set);
            if (!set.isEmpty()) {
                rowList.add(row);
                setList.add(set);
            }
        }
        count = rowList.size();
        for (int i = 0; i < count; i++) {
            ListRow row   = rowList.get(i);
            Namer   panel = new Namer(row, setList.get(i), count - i - 1);
            Modal dialog = Modal.prepareToShowMessage(owner,
                    MessageFormat.format(I18n.text("Name {0}"), row.getLocalizedName()),
                    MessageType.QUESTION, panel);
            if (i != count - 1) {
                dialog.addCancelRemainingButton();
            }
            dialog.addCancelButton();
            panel.mApplyButton = dialog.addApplyButton();
            panel.mApplyButton.setEnabled(false);
            dialog.presentToUser();
            switch (dialog.getResult()) {
            case Modal.OK:
                panel.applyChanges();
                modified = true;
                break;
            case Modal.CANCEL:
                break;
            case Modal.CLOSED:
            default:
                return modified;
            }
        }
        return modified;
    }

    private Namer(ListRow row, Set<String> set, int remaining) {
        super(new PrecisionLayout().setColumns(2));
        mRow = row;
        mFields = new ArrayList<>();
        Label header = new Label(Text.truncateIfNecessary(row.toString(), 80, SwingConstants.CENTER));
        header.setThemeFont(ThemeFont.HEADER);
        add(header, new PrecisionLayoutData().setMiddleHorizontalAlignment().setHorizontalSpan(2));
        add(new Separator(), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(2).setBottomMargin(10));
        List<String> list = new ArrayList<>(set);
        Collections.sort(list);
        for (String name : list) {
            EditorField field = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, "", null);
            field.setName(name);
            field.getDocument().addDocumentListener(this);
            add(new Label(name, field), new PrecisionLayoutData().setFillHorizontalAlignment());
            add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMinimumWidth(200));
            mFields.add(field);
        }
        if (remaining > 0) {
            Label reminder = new Label(remaining == 1 ? I18n.text("1 item remaining to be named.") :
                    MessageFormat.format(I18n.text("{0} items remaining to be named."),
                            Integer.valueOf(remaining)));
            reminder.setThemeFont(ThemeFont.LABEL_SECONDARY);
            add(new Separator(), new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(2).setTopMargin(10));
            add(reminder, new PrecisionLayoutData().setMiddleHorizontalAlignment().setHorizontalSpan(2));
        }
    }

    private void applyChanges() {
        Commitable.sendCommitToFocusOwner();
        Map<String, String> map = new HashMap<>();
        for (EditorField field : mFields) {
            map.put(field.getName(), field.getText());
        }
        mRow.applyNameableKeys(map);
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        changedUpdate(event);
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        changedUpdate(event);
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        boolean enable = true;
        for (EditorField field : mFields) {
            if (field.getText().isBlank()) {
                enable = false;
                break;
            }
        }
        mApplyButton.setEnabled(enable);
    }
}
