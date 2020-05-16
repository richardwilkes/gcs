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

package com.trollworks.gcs.notes;

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.layout.RowDistribution;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Dimension;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTextArea;
import javax.swing.JTextField;
import javax.swing.ScrollPaneConstants;
import javax.swing.SwingConstants;

/** The detailed editor for {@link Note}s. */
public class NoteEditor extends RowEditor<Note> {
    private JTextArea  mEditor;
    private JTextField mReferenceField;

    /**
     * Creates a new {@link Note} editor.
     *
     * @param note The {@link Note} to edit.
     */
    public NoteEditor(Note note) {
        super(note);
        JPanel content   = new JPanel(new ColumnLayout(2, RowDistribution.GIVE_EXCESS_TO_LAST));
        JLabel iconLabel = new JLabel(note.getIcon(true));
        JPanel right     = new JPanel(new ColumnLayout(1, RowDistribution.GIVE_EXCESS_TO_LAST));
        content.add(iconLabel);
        content.add(right);

        mReferenceField = new JTextField(Text.makeFiller(6, 'M'));
        UIUtilities.setToPreferredSizeOnly(mReferenceField);
        mReferenceField.setText(note.getReference());
        mReferenceField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("A reference to the book and page this note applies to (e.g. B22 would refer to \"Basic Set\", page 22)")));
        mReferenceField.setEnabled(mIsEditable);
        JPanel wrapper = new JPanel(new ColumnLayout(4));
        wrapper.add(new LinkedLabel(I18n.Text("Note Content:")));
        wrapper.add(new JPanel());
        wrapper.add(new LinkedLabel(I18n.Text("Page Reference"), mReferenceField));
        wrapper.add(mReferenceField);
        right.add(wrapper);

        mEditor = new JTextArea(note.getDescription());
        mEditor.setLineWrap(true);
        mEditor.setWrapStyleWord(true);
        mEditor.setEnabled(mIsEditable);
        JScrollPane scroller = new JScrollPane(mEditor, ScrollPaneConstants.VERTICAL_SCROLLBAR_ALWAYS, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_ALWAYS);
        scroller.setMinimumSize(new Dimension(400, 300));
        iconLabel.setVerticalAlignment(SwingConstants.TOP);
        iconLabel.setAlignmentY(-1.0f);
        right.add(scroller);

        add(content);
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified = mRow.setReference(mReferenceField.getText());
        modified |= mRow.setDescription(mEditor.getText());
        return modified;
    }

    @Override
    public void finished() {
        // Unused
    }
}
