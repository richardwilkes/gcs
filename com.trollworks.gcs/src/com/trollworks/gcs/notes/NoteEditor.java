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

package com.trollworks.gcs.notes;

import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Component;
import java.awt.Dimension;
import javax.swing.JPanel;
import javax.swing.JTextField;

/** The detailed editor for {@link Note}s. */
public class NoteEditor extends RowEditor<Note> {
    private MultiLineTextField mDescriptionField;
    private JTextField         mReferenceField;

    /**
     * Creates a new {@link Note} editor.
     *
     * @param note The {@link Note} to edit.
     */
    public NoteEditor(Note note) {
        super(note);
        addContent();
        Component scroller = getComponent(0);
        Dimension size     = scroller.getPreferredSize();
        if (size.width < 600) {
            size.width = 600;
        }
        if (size.height < 400) {
            size.height = 400;
        }
        scroller.setPreferredSize(size);
    }

    @Override
    protected void addContentSelf(ScrollContent outer) {
        JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2));
        outer.add(wrapper, new PrecisionLayoutData().setFillAlignment().setGrabSpace(true));

        mDescriptionField = new MultiLineTextField(mRow.getDescription(), null, null);
        wrapper.add(new LinkedLabel(I18n.Text("Description"), mDescriptionField), new PrecisionLayoutData().setBeginningVerticalAlignment().setFillHorizontalAlignment().setTopMargin(2));
        wrapper.add(mDescriptionField, new PrecisionLayoutData().setFillAlignment().setGrabSpace(true));

        mReferenceField = new JTextField(mRow.getReference());
        mReferenceField.setToolTipText(Text.wrapPlainTextForToolTip(PageRefCell.getStdToolTip(I18n.Text("note"))));
        wrapper.add(new LinkedLabel(I18n.Text("Page Reference"), mReferenceField), new PrecisionLayoutData().setFillHorizontalAlignment());
        wrapper.add(mReferenceField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified = mRow.setReference(mReferenceField.getText());
        modified |= mRow.setDescription(mDescriptionField.getText());
        return modified;
    }
}
