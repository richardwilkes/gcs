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

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.OutputStreamWriter;
import java.nio.charset.StandardCharsets;
import java.text.MessageFormat;
import java.util.zip.GZIPInputStream;
import java.util.zip.GZIPOutputStream;
import javax.swing.undo.AbstractUndoableEdit;
import javax.swing.undo.CannotRedoException;
import javax.swing.undo.CannotUndoException;

/** An undo for the entire row, with the exception of its children. */
public class RowUndo extends AbstractUndoableEdit {
    private DataFile mDataFile;
    private ListRow  mRow;
    private String   mName;
    private byte[]   mBefore;
    private byte[]   mAfter;

    /**
     * Creates a new {@link RowUndo}.
     *
     * @param row The row being undone.
     */
    public RowUndo(ListRow row) {
        mRow = row;
        mDataFile = mRow.getDataFile();
        mName = MessageFormat.format(I18n.Text("{0} Changes"), mRow.getLocalizedName());
        mBefore = serialize(mRow);
    }

    /**
     * Call to finish capturing the undo state.
     *
     * @return {@code true} if there is a difference between the before and after state.
     */
    public boolean finish() {
        mAfter = serialize(mRow);
        int length = mBefore.length;
        if (length != mAfter.length) {
            return true;
        }
        for (int i = 0; i < length; i++) {
            if (mBefore[i] != mAfter[i]) {
                return true;
            }
        }
        return false;
    }

    private static byte[] serialize(ListRow row) {
        try {
            ByteArrayOutputStream baos = new ByteArrayOutputStream();
            try (GZIPOutputStream gos = new GZIPOutputStream(baos)) {
                try (JsonWriter w = new JsonWriter(new OutputStreamWriter(gos, StandardCharsets.UTF_8), "")) {
                    row.save(w, true);
                }
            }
            return baos.toByteArray();
        } catch (Exception exception) {
            Log.error(exception);
        }
        return new byte[0];
    }

    private void deserialize(byte[] buffer) {
        try (GZIPInputStream s = new GZIPInputStream(new ByteArrayInputStream(buffer))) {
            JsonMap   m     = Json.asMap(Json.parse(s));
            LoadState state = new LoadState();
            state.mDataFileVersion = mDataFile.getJSONVersion();
            state.mForUndo = true;
            mRow.load(m, state);
        } catch (Exception exception) {
            Log.error(exception);
        }
    }

    @Override
    public void undo() throws CannotUndoException {
        super.undo();
        deserialize(mBefore);
    }

    @Override
    public void redo() throws CannotRedoException {
        super.redo();
        deserialize(mAfter);
    }

    /** @return The {@link DataFile} this undo works on. */
    public DataFile getDataFile() {
        return mDataFile;
    }

    /** @return The row this undo works on. */
    public ListRow getRow() {
        return mRow;
    }

    @Override
    public String getPresentationName() {
        return mName;
    }
}
