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

package com.trollworks.gcs.notes;

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.Dimension;
import javax.swing.SwingConstants;

/** The detailed editor for {@link Note}s. */
public class NoteEditor extends RowEditor<Note> {
    private MultiLineTextField mDescriptionField;
    private EditorField        mReferenceField;

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
        Panel wrapper = new Panel(new PrecisionLayout().setMargins(0).setColumns(2));
        outer.add(wrapper, new PrecisionLayoutData().setFillAlignment().setGrabSpace(true));

        mDescriptionField = new MultiLineTextField(mRow.getDescription(), null, null);
        addLabel(wrapper, I18n.text("Description")).setBeginningVerticalAlignment().setTopMargin(2);
        wrapper.add(mDescriptionField, new PrecisionLayoutData().setFillAlignment().setGrabSpace(true));

        mReferenceField = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, mRow.getReference(), PageRefCell.getStdToolTip(I18n.text("note")));
        addLabel(wrapper, I18n.text("Page Reference"));
        wrapper.add(mReferenceField, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified = mRow.setReference(mReferenceField.getText());
        modified |= mRow.setDescription(mDescriptionField.getText());
        return modified;
    }
}
