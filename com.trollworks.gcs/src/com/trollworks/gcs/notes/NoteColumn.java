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

import com.trollworks.gcs.character.CollectedModels;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.equipment.FontIconCell;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.widget.outline.Cell;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.HeaderCell;
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
            return I18n.text("Notes");
        }

        @Override
        public String getToolTip() {
            return "";
        }

        @Override
        public String getToolTip(Note note) {
            return note.getDescriptionToolTipText();
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
            return note.getDescriptionText();
        }
    },
    /** The page reference. */
    REFERENCE {
        @Override
        public String toString() {
            return FontAwesome.BOOKMARK;
        }

        @Override
        public String getToolTip() {
            return PageRefCell.getStdToolTip(I18n.text("note"));
        }

        @Override
        public String getToolTip(Note note) {
            return PageRefCell.getStdCellToolTip(note.getReference());
        }

        @Override
        public Cell getCell() {
            return new PageRefCell();
        }

        @Override
        public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
            return new FontIconCell(Fonts.FONT_AWESOME_SOLID, sheetOrTemplate);
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

    /**
     * @param note The {@link Note} to get the data from.
     * @return The tooltip for a specific row within the column.
     */
    public String getToolTip(Note note) {
        return null;
    }

    /** @return The {@link Cell} used to display the data. */
    public abstract Cell getCell();

    /** @return The {@link Cell} used to display the header. */
    public HeaderCell getHeaderCell(boolean sheetOrTemplate) {
        return new ListHeaderCell(sheetOrTemplate);
    }

    /**
     * Adds all relevant {@link Column}s to a {@link Outline}.
     *
     * @param outline  The {@link Outline} to use.
     * @param dataFile The {@link DataFile} that data is being displayed for.
     */
    public static void addColumns(Outline outline, DataFile dataFile) {
        GURPSCharacter character       = dataFile instanceof GURPSCharacter ? (GURPSCharacter) dataFile : null;
        boolean        sheetOrTemplate = dataFile instanceof CollectedModels;
        OutlineModel   model           = outline.getModel();
        for (NoteColumn one : values()) {
            Column column = new Column(one.ordinal(), one.toString(character), one.getToolTip(), one.getCell());
            column.setHeaderCell(one.getHeaderCell(sheetOrTemplate));
            model.addColumn(column);
        }
    }
}
