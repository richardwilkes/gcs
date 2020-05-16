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

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.widget.outline.Cell;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListHeaderCell;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.utility.I18n;

import javax.swing.SwingConstants;

/** Definitions for note columns. */
public enum NoteColumn {
    /** The text of the note. */
    TEXT {
        @Override
        public String toString() {
            return I18n.Text("Notes");
        }

        @Override
        public String getToolTip() {
            return "";
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.LEFT, true);
        }

        @Override
        public Object getData(Note note) {
            return getDataAsText(note);
        }

        @Override
        public String getDataAsText(Note note) {
            StringBuilder builder = new StringBuilder();
            String        notes   = note.getNotes();
            builder.append(note.toString());
            if (!notes.isEmpty()) {
                builder.append(" - ");
                builder.append(notes);
            }
            return builder.toString();
        }
    },
    /** The page reference. */
    REFERENCE {
        @Override
        public String toString() {
            return I18n.Text("Ref");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("A reference to the book and page this note applies to (e.g. B22 would refer to \"Basic Set\", page 22)");
        }

        @Override
        public Cell getCell() {
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public Object getData(Note note) {
            return getDataAsText(note);
        }

        @Override
        public String getDataAsText(Note note) {
            return note.getReference();
        }
    };

    /**
     * @param character The {@link GURPSCharacter} this note list is associated with, or {@code
     *                  null}.
     * @return The header title.
     */
    public String toString(GURPSCharacter character) {
        return toString();
    }

    /**
     * @param note The {@link Note} to get the data from.
     * @return An object representing the data for this column.
     */
    public abstract Object getData(Note note);

    /**
     * @param note The {@link Note} to get the data from.
     * @return Text representing the data for this column.
     */
    public abstract String getDataAsText(Note note);

    /** @return The tooltip for the column. */
    public abstract String getToolTip();

    /** @return The {@link Cell} used to display the data. */
    public abstract Cell getCell();

    /**
     * Adds all relevant {@link Column}s to a {@link Outline}.
     *
     * @param outline  The {@link Outline} to use.
     * @param dataFile The {@link DataFile} that data is being displayed for.
     */
    public static void addColumns(Outline outline, DataFile dataFile) {
        GURPSCharacter character       = dataFile instanceof GURPSCharacter ? (GURPSCharacter) dataFile : null;
        boolean        sheetOrTemplate = dataFile instanceof GURPSCharacter || dataFile instanceof Template;
        OutlineModel   model           = outline.getModel();
        for (NoteColumn one : values()) {
            Column column = new Column(one.ordinal(), one.toString(character), one.getToolTip(), one.getCell());
            column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
            model.addColumn(column);
        }
    }
}
