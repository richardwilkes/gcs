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

package com.trollworks.gcs.datafile;

import com.trollworks.gcs.character.DisplayOption;
import com.trollworks.gcs.menu.edit.Undoable;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.widget.DataModifiedListener;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.VersionException;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.undo.StdUndoManager;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.utility.xml.XMLNodeType;
import com.trollworks.gcs.utility.xml.XMLReader;

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
import javax.swing.undo.UndoableEdit;

/** A common super class for all data file-based model objects. */
public abstract class DataFile implements Undoable {
    /** The 'id' attribute. */
    public static final String                     ATTRIBUTE_ID           = "id";
    /** Identifies the type of a JSON object. */
    public static final String                     KEY_TYPE               = "type";
    private             Path                       mPath;
    private             UUID                       mId                    = UUID.randomUUID();
    private             Notifier                   mNotifier              = new Notifier();
    private             boolean                    mModified;
    private             StdUndoManager             mUndoManager           = new StdUndoManager();
    private             List<DataModifiedListener> mDataModifiedListeners = new ArrayList<>();
    private             boolean                    mSortingMarksDirty     = true;

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
                // Load xml format from version 4.18 and earlier
                try (XMLReader reader = new XMLReader(fileReader)) {
                    XMLNodeType type  = reader.next();
                    boolean     found = false;
                    while (type != XMLNodeType.END_DOCUMENT) {
                        if (type == XMLNodeType.START_TAG) {
                            String name = reader.getName();
                            if (matchesRootTag(name)) {
                                if (found) {
                                    throw new IOException();
                                }
                                found = true;
                                load(reader, new LoadState());
                            } else {
                                reader.skipTag(name);
                            }
                            type = reader.getType();
                        } else {
                            type = reader.next();
                        }
                    }
                }
            } else {
                load(Json.asMap(Json.parse(fileReader)), new LoadState());
            }
        }
        mModified = false;
    }

    /**
     * @param reader The {@link XMLReader} to load data from.
     * @param state  The {@link LoadState} to use.
     */
    public void load(XMLReader reader, LoadState state) throws IOException {
        try {
            mId = UUID.fromString(reader.getAttribute(ATTRIBUTE_ID));
        } catch (Exception exception) {
            mId = UUID.randomUUID();
        }
        state.mDataFileVersion = reader.getAttributeAsInteger(LoadState.ATTRIBUTE_VERSION, 0);
        if (state.mDataFileVersion > getXMLTagVersion()) {
            throw VersionException.createTooNew();
        }
        loadSelf(reader, state);
    }

    /**
     * @param m     The {@link JsonMap} to load data from.
     * @param state The {@link LoadState} to use.
     */
    public void load(JsonMap m, LoadState state) throws IOException {
        try {
            mId = UUID.fromString(m.getString(ATTRIBUTE_ID));
        } catch (Exception exception) {
            mId = UUID.randomUUID();
        }
        state.mDataFileVersion = m.getInt(LoadState.ATTRIBUTE_VERSION);
        if (state.mDataFileVersion > getJSONVersion()) {
            throw VersionException.createTooNew();
        }
        loadSelf(m, state);
    }

    /**
     * Called to load the data file.
     *
     * @param reader The {@link XMLReader} to load data from.
     * @param state  The {@link LoadState} to use.
     */
    protected abstract void loadSelf(XMLReader reader, LoadState state) throws IOException;

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
                save(w, true, false);
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
     * @param w               The {@link JsonWriter} to use.
     * @param includeUniqueID Whether the unique should be included in the attribute list.
     * @param onlyIfNotEmpty  Whether to write something even if the file contents are empty.
     */
    public void save(JsonWriter w, boolean includeUniqueID, boolean onlyIfNotEmpty) throws IOException {
        if (!onlyIfNotEmpty || !isEmpty()) {
            w.startMap();
            w.keyValue(KEY_TYPE, getJSONTypeName());
            w.keyValue(LoadState.ATTRIBUTE_VERSION, getJSONVersion());
            if (includeUniqueID) {
                w.keyValue(ATTRIBUTE_ID, mId.toString());
            }
            saveSelf(w);
            w.endMap();
        }
    }

    /**
     * Called to save the data file.
     *
     * @param w The {@link JsonWriter} to use.
     */
    protected abstract void saveSelf(JsonWriter w) throws IOException;

    /** @return Whether the file is empty. By default, returns {@code false}. */
    @SuppressWarnings("static-method")
    public boolean isEmpty() {
        return false;
    }

    /** @return The most recent version of the JSON data this object knows how to load. */
    public abstract int getJSONVersion();

    /** @return The type name to use for this data. */
    public abstract String getJSONTypeName();

    /** @return The most recent version of the XML tag this object knows how to load. */
    public abstract int getXMLTagVersion();

    /** @return The XML root container tag name for this particular file. */
    public abstract String getXMLTagName();

    /**
     * Called to match an XML tag name with the root tag for this data file.
     *
     * @param name The tag name to check.
     * @return Whether it matches the root tag or not.
     */
    public boolean matchesRootTag(String name) {
        return getXMLTagName().equals(name);
    }

    /** @return The {@link FileType}. */
    public abstract FileType getFileType();

    /** @return The icons representing this file. */
    public abstract RetinaIcon getFileIcons();

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
    public UUID getId() {
        return mId;
    }

    /** Replaces the existing ID with a new randomly generated one. */
    public void generateNewId() {
        mId = UUID.randomUUID();
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

    /**
     * Resets the underlying {@link Notifier} by removing all targets except the specified ones.
     *
     * @param exclude The {@link NotifierTarget}(s) to exclude.
     */
    public void resetNotifier(NotifierTarget... exclude) {
        mNotifier.reset(exclude);
    }

    /**
     * Registers a {@link NotifierTarget} with this data file's {@link Notifier}.
     *
     * @param target The {@link NotifierTarget} to register.
     * @param names  The names to register for.
     */
    public void addTarget(NotifierTarget target, String... names) {
        mNotifier.add(target, names);
    }

    /**
     * Un-registers a {@link NotifierTarget} with this data file's {@link Notifier}.
     *
     * @param target The {@link NotifierTarget} to un-register.
     */
    public void removeTarget(NotifierTarget target) {
        mNotifier.remove(target);
    }

    /**
     * Starts the notification process. Should be called before calling {@link #notify(String,
     * Object)}.
     */
    public void startNotify() {
        if (mNotifier.getBatchLevel() == 0) {
            startNotifyAtBatchLevelZero();
        }
        mNotifier.startBatch();
    }

    /**
     * Called when {@link #startNotify()} is called and the current batch level is zero.
     */
    protected void startNotifyAtBatchLevelZero() {
        // Does nothing by default.
    }

    /**
     * Sends a notification to all interested consumers.
     *
     * @param type The notification type.
     * @param data Extra data specific to this notification.
     */
    public void notify(String type, Object data) {
        setModified(true);
        mNotifier.notify(this, type, data);
        notifyOccured();
    }

    /** Called when {@link #notify(String, Object)} is called. */
    protected void notifyOccured() {
        // Does nothing by default.
    }

    /**
     * Ends the notification process. Must be called after calling {@link #notify(String, Object)}.
     */
    public void endNotify() {
        if (mNotifier.getBatchLevel() == 1) {
            endNotifyAtBatchLevelOne();
        }
        mNotifier.endBatch();
    }

    /**
     * Called when {@link #endNotify()} is called and the current batch level is one.
     */
    protected void endNotifyAtBatchLevelOne() {
        // Does nothing by default.
    }

    /**
     * Sends a notification to all interested consumers.
     *
     * @param type The notification type.
     * @param data Extra data specific to this notification.
     */
    public void notifySingle(String type, Object data) {
        startNotify();
        notify(type, data);
        endNotify();
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

    public WeightUnits defaultWeightUnits() {
        return Preferences.getInstance().getDefaultWeightUnits();
    }

    public boolean useSimpleMetricConversions() {
        return Preferences.getInstance().useSimpleMetricConversions();
    }

    public boolean useMultiplicativeModifiers() {
        return Preferences.getInstance().useMultiplicativeModifiers();
    }

    public boolean useModifyingDicePlusAdds() {
        return Preferences.getInstance().useModifyingDicePlusAdds();
    }

    public DisplayOption userDescriptionDisplay() {
        return Preferences.getInstance().getUserDescriptionDisplay();
    }

    public DisplayOption modifiersDisplay() {
        return Preferences.getInstance().getModifiersDisplay();
    }

    public DisplayOption notesDisplay() {
        return Preferences.getInstance().getNotesDisplay();
    }
}
