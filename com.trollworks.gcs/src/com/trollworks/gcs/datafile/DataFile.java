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

package com.trollworks.gcs.datafile;

import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.menu.edit.Undoable;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.settings.SheetSettings;
import com.trollworks.gcs.ui.widget.DataModifiedListener;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.VersionException;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.undo.StdUndoManager;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import javax.swing.Icon;
import javax.swing.undo.UndoableEdit;

/** A common super class for all data file-based model objects. */
public abstract class DataFile extends ChangeableData implements Undoable {
    /** The 'id' attribute. */
    public static final String                     ID                     = "id";
    /** The attribute used for versioning. */
    public static final String                     VERSION                = "version";
    /**
     * The data file version used with the current release. Note that this is intentionally the same
     * for all data files that GCS processes.
     */
    public static final int                        CURRENT_VERSION        = 2;
    /** Identifies the type of a JSON object. */
    public static final String                     TYPE                   = "type";
    private             Path                       mPath;
    private             UUID                       mID                    = UUID.randomUUID();
    private             StdUndoManager             mUndoManager           = new StdUndoManager();
    private             List<DataModifiedListener> mDataModifiedListeners = new ArrayList<>();
    private             boolean                    mSortingMarksDirty     = true;
    private             boolean                    mModified;

    @Override
    public void notifyOfChange() {
        setModified(true);
        super.notifyOfChange();
    }

    /** @param path The path to load. */
    public void load(Path path) throws IOException {
        setPath(path);
        try (BufferedReader fileReader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            fileReader.mark(20);
            char[] buffer = new char[5];
            int    n      = fileReader.read(buffer);
            if (n < 0) {
                throw new IOException("Premature EOF");
            }
            fileReader.reset();
            if (n == 5 && buffer[0] == '<' && buffer[1] == '?' && buffer[2] == 'x' && buffer[3] == 'm' && buffer[4] == 'l') {
                throw new IOException("The old xml format from versions prior to GCS v4.20 cannot be read by this version of GCS");
            } else {
                load(Json.asMap(Json.parse(fileReader)), new LoadState());
            }
        }
        mModified = false;
    }

    /**
     * @param m     The {@link JsonMap} to load data from.
     * @param state The {@link LoadState} to use.
     */
    public void load(JsonMap m, LoadState state) throws IOException {
        try {
            mID = UUID.fromString(m.getString(ID));
        } catch (Exception exception) {
            mID = UUID.randomUUID();
        }
        state.mDataFileVersion = m.getInt(VERSION);
        if (state.mDataFileVersion > CURRENT_VERSION) {
            throw VersionException.createTooNew();
        }
        loadSelf(m, state);
    }

    /**
     * Called to load the data file.
     *
     * @param m     The {@link JsonMap} to load data from.
     * @param state The {@link LoadState} to use.
     */
    protected abstract void loadSelf(JsonMap m, LoadState state) throws IOException;

    /**
     * Saves the data out to the specified path. Does not affect the result of {@link #getPath()}.
     *
     * @param path The path to write to.
     * @return {@code true} on success.
     */
    public boolean save(Path path) {
        SafeFileUpdater transaction = new SafeFileUpdater();
        boolean         success     = false;
        transaction.begin();
        try {
            File transactionFile = transaction.getTransactionFile(path.toFile());
            try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(transactionFile, StandardCharsets.UTF_8)), "\t")) {
                save(w, SaveType.NORMAL, false);
            }
            transaction.commit();
            setModified(false);
            success = true;
        } catch (Exception exception) {
            Log.error(exception);
            transaction.abort();
        }
        return success;
    }

    /**
     * Writes the data to the specified {@link JsonWriter}.
     *
     * @param w              The {@link JsonWriter} to use.
     * @param saveType       The type of save being performed.
     * @param onlyIfNotEmpty Whether to write something even if the file contents are empty.
     */
    public void save(JsonWriter w, SaveType saveType, boolean onlyIfNotEmpty) throws IOException {
        if (!onlyIfNotEmpty || !isEmpty()) {
            w.startMap();
            w.keyValue(TYPE, getJSONTypeName());
            w.keyValue(VERSION, CURRENT_VERSION);
            w.keyValue(ID, mID.toString());
            saveSelf(w, saveType);
            w.endMap();
        }
    }

    /**
     * Called to save the data file.
     *
     * @param w        The {@link JsonWriter} to use.
     * @param saveType The type of save being performed.
     */
    protected abstract void saveSelf(JsonWriter w, SaveType saveType) throws IOException;

    /** @return Whether the file is empty. By default, returns {@code false}. */
    public boolean isEmpty() {
        return false;
    }

    /** @return The type name to use for this data. */
    public abstract String getJSONTypeName();

    /** @return The {@link FileType}. */
    public abstract FileType getFileType();

    /** @return The icon representing this file. */
    public final Icon getFileIcon() {
        return getFileType().getIcon();
    }

    /** @return The path associated with this data file. */
    public Path getPath() {
        return mPath;
    }

    /** @param path The path associated with this data file. */
    public void setPath(Path path) {
        if (path != null) {
            path = path.normalize().toAbsolutePath();
        }
        mPath = path;
    }

    /** @return The ID for this data file. */
    public UUID getID() {
        return mID;
    }

    /** Replaces the existing ID with a new randomly generated one. */
    public void generateNewID() {
        mID = UUID.randomUUID();
    }

    /** @return {@code true} if the data has been modified. */
    public final boolean isModified() {
        return mModified;
    }

    /** @param modified Whether or not the data has been modified. */
    public final void setModified(boolean modified) {
        if (mModified != modified) {
            mModified = modified;
            for (DataModifiedListener listener : mDataModifiedListeners.toArray(new DataModifiedListener[0])) {
                listener.dataModificationStateChanged(this, mModified);
            }
        }
    }

    /** @param listener The listener to add. */
    public void addDataModifiedListener(DataModifiedListener listener) {
        mDataModifiedListeners.remove(listener);
        mDataModifiedListeners.add(listener);
    }

    /** @param listener The listener to remove. */
    public void removeDataModifiedListener(DataModifiedListener listener) {
        mDataModifiedListeners.remove(listener);
    }

    /** @return The {@link StdUndoManager} to use. */
    @Override
    public final StdUndoManager getUndoManager() {
        return mUndoManager;
    }

    /** @param mgr The {@link StdUndoManager} to use. */
    public final void setUndoManager(StdUndoManager mgr) {
        mUndoManager = mgr;
    }

    /** @param edit The {@link UndoableEdit} to add. */
    public final void addEdit(UndoableEdit edit) {
        mUndoManager.addEdit(edit);
    }

    /**
     * @return {@code true} if sorting a list should be considered a change that marks the file
     *         dirty.
     */
    public final boolean sortingMarksDirty() {
        return mSortingMarksDirty;
    }

    /**
     * @param markDirty {@code true} if sorting a list should be considered a change that marks the
     *                  file dirty.
     */
    public final void setSortingMarksDirty(boolean markDirty) {
        mSortingMarksDirty = markDirty;
    }

    public SheetSettings getSheetSettings() {
        return Settings.getInstance().getSheetSettings();
    }

    public AttributeDef getAttributeDef(String id) {
        return getSheetSettings().getAttributes().get(id);
    }
}
