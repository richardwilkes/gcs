/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.notes;

import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.layout.RowDistribution;
import com.trollworks.gcs.ui.widget.outline.RowEditor;

import java.awt.Dimension;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTextArea;
import javax.swing.ScrollPaneConstants;
import javax.swing.SwingConstants;

/** The detailed editor for {@link Note}s. */
public class NoteEditor extends RowEditor<Note> {
    public static JTextArea addContentTo(JPanel parent, String description, RetinaIcon icon, boolean editable) {
        JPanel    content   = new JPanel(new ColumnLayout(2, RowDistribution.GIVE_EXCESS_TO_LAST));
        JLabel    iconLabel = new JLabel(icon);
        JTextArea editor    = new JTextArea(description);
        editor.setLineWrap(true);
        editor.setWrapStyleWord(true);
        editor.setEnabled(editable);
        JScrollPane scroller = new JScrollPane(editor, ScrollPaneConstants.VERTICAL_SCROLLBAR_ALWAYS, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_ALWAYS);
        scroller.setMinimumSize(new Dimension(400, 300));
        iconLabel.setVerticalAlignment(SwingConstants.TOP);
        iconLabel.setAlignmentY(-1.0f);
        content.add(iconLabel);
        content.add(scroller);
        parent.add(content);
        return editor;
    }

    private JTextArea mEditor;

    /**
     * Creates a new {@link Note} editor.
     *
     * @param note The {@link Note} to edit.
     */
    public NoteEditor(Note note) {
        super(note);
        mEditor = addContentTo(this, note.getDescription(), note.getIcon(true), mIsEditable);
    }

    @Override
    public boolean applyChangesSelf() {
        return mRow.setDescription(mEditor.getText());
    }

    @Override
    public void finished() {
        // Unused
    }
}
