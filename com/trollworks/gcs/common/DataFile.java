/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.common;

import com.trollworks.ttk.notification.Notifier;
import com.trollworks.ttk.notification.NotifierTarget;
import com.trollworks.ttk.undo.StdUndoManager;
import com.trollworks.ttk.utility.Debug;
import com.trollworks.ttk.utility.Path;
import com.trollworks.ttk.utility.SafeFileUpdater;
import com.trollworks.ttk.utility.UniqueID;
import com.trollworks.ttk.utility.VersionException;
import com.trollworks.ttk.widgets.DataModifiedListener;
import com.trollworks.ttk.xml.XMLNodeType;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.awt.image.BufferedImage;
import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.FileReader;
import java.io.IOException;
import java.util.ArrayList;

import javax.swing.undo.UndoableEdit;

/** A common super class for all data file-based model objects. */
public abstract class DataFile {
	/** The 'unique ID' attribute. */
	public static final String				ATTRIBUTE_UNIQUE_ID		= "unique_id";			//$NON-NLS-1$
	private File							mFile;
	private UniqueID						mUniqueID;
	private Notifier						mNotifier;
	private boolean							mModified;
	private StdUndoManager					mUndoManager;
	private ArrayList<DataModifiedListener>	mDataModifiedListeners	= new ArrayList<>();

	/** Creates a new data file object. */
	protected DataFile() {
		mUniqueID = new UniqueID();
		mNotifier = new Notifier();
	}

	/**
	 * Creates a new data file object from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @throws IOException if the data cannot be read or the file doesn't contain valid information.
	 */
	protected DataFile(File file) throws IOException {
		this(file, new LoadState());
	}

	/**
	 * Creates a new data file object from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @param state The {@link LoadState} to use.
	 * @throws IOException if the data cannot be read or the file doesn't contain valid information.
	 */
	protected DataFile(File file, LoadState state) throws IOException {
		setFile(file);
		mNotifier = new Notifier();
		try (FileReader fileReader = new FileReader(file);
			XMLReader reader = new XMLReader(fileReader)) {
			XMLNodeType type = reader.next();
			boolean found = false;

			while (type != XMLNodeType.END_DOCUMENT) {
				if (type == XMLNodeType.START_TAG) {
					String name = reader.getName();
					if (matchesRootTag(name)) {
						if (!found) {
							found = true;
							load(reader, state);
						} else {
							throw new IOException();
						}
					} else {
						reader.skipTag(name);
					}
					type = reader.getType();
				} else {
					type = reader.next();
				}
			}
		}
		setModified(false);
	}

	/**
	 * @param reader The {@link XMLReader} to load data from.
	 * @param state The {@link LoadState} to use.
	 */
	public void load(XMLReader reader, LoadState state) throws IOException {
		mUniqueID = new UniqueID(reader.getAttribute(ATTRIBUTE_UNIQUE_ID));
		state.mDataFileVersion = reader.getAttributeAsInteger(LoadState.ATTRIBUTE_VERSION, 0);
		if (state.mDataFileVersion > getXMLTagVersion()) {
			throw VersionException.createTooNew();
		}
		loadSelf(reader, state);
	}

	/**
	 * Called to load the data file.
	 * 
	 * @param reader The {@link XMLReader} to load data from.
	 * @param state The {@link LoadState} to use.
	 */
	protected abstract void loadSelf(XMLReader reader, LoadState state) throws IOException;

	/**
	 * Saves the data out to the specified file. Does not affect the result of {@link #getFile()}.
	 * 
	 * @param file The file to write to.
	 * @return <code>true</code> on success.
	 */
	public boolean save(File file) {
		SafeFileUpdater transaction = new SafeFileUpdater();
		boolean success = false;
		transaction.begin();
		try {
			File transactionFile = transaction.getTransactionFile(file);
			try (XMLWriter out = new XMLWriter(new BufferedOutputStream(new FileOutputStream(transactionFile)))) {
				out.writeHeader();
				save(out, true, false);
				success = !out.checkError();
			}
			if (success) {
				success = false;
				transaction.commit();
				setModified(false);
				success = true;
			} else {
				transaction.abort();
			}
		} catch (Exception exception) {
			Debug.diagnoseLoadAndSave(exception);
			transaction.abort();
		}
		return success;
	}

	/**
	 * Saves the root tag.
	 * 
	 * @param out The XML writer to use.
	 * @param includeUniqueID Whether the {@link UniqueID} should be included in the attribute list.
	 * @param onlyIfNotEmpty Whether to write something even if the file contents are empty.
	 */
	public void save(XMLWriter out, boolean includeUniqueID, boolean onlyIfNotEmpty) {
		if (!onlyIfNotEmpty || !isEmpty()) {
			out.startTag(getXMLTagName());
			if (includeUniqueID) {
				out.writeAttribute(ATTRIBUTE_UNIQUE_ID, getUniqueID().toString());
			}
			out.writeAttribute(LoadState.ATTRIBUTE_VERSION, getXMLTagVersion());
			out.finishTagEOL();
			saveSelf(out);
			out.endTagEOL(getXMLTagName(), true);
		}
	}

	/**
	 * Called to save the data file.
	 * 
	 * @param out The XML writer to use.
	 */
	protected abstract void saveSelf(XMLWriter out);

	/** @return Whether the file is empty. By default, returns <code>false</code>. */
	@SuppressWarnings("static-method")
	public boolean isEmpty() {
		return false;
	}

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

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @return The file icon.
	 */
	public abstract BufferedImage getFileIcon(boolean large);

	/**
	 * Sub-classes must call this method prior to returning from their constructors.
	 */
	protected void initialize() {
		setModified(false);
	}

	/** @return The file associated with this data file. */
	public File getFile() {
		return mFile;
	}

	/** @param file The file associated with this data file. */
	public void setFile(File file) {
		if (file != null) {
			file = Path.getFile(Path.getFullPath(file));
		}
		mFile = file;
	}

	/** @return The unique ID for this data file. */
	public final UniqueID getUniqueID() {
		return mUniqueID;
	}

	/** @return <code>true</code> if the data has been modified. */
	public final boolean isModified() {
		return mModified;
	}

	/** @param modified Whether or not the data has been modified. */
	public final void setModified(boolean modified) {
		if (mModified != modified) {
			mModified = modified;
			for (DataModifiedListener listener : mDataModifiedListeners.toArray(new DataModifiedListener[mDataModifiedListeners.size()])) {
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

	/** Resets the underlying {@link Notifier} by removing all targets. */
	public void resetNotifier() {
		mNotifier.reset();
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
	 * @param names The names to register for.
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
	 * Starts the notification process. Should be called before calling
	 * {@link #notify(String,Object)}.
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

	/** Called when {@link #notify(String,Object)} is called. */
	protected void notifyOccured() {
		// Does nothing by default.
	}

	/**
	 * Ends the notification process. Must be called after calling {@link #notify(String,Object)}.
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

	/** @return The {@link StdUndoManager} to use. May be <code>null</code>. */
	public final StdUndoManager getUndoManager() {
		return mUndoManager;
	}

	/** @param mgr The {@link StdUndoManager} to use. */
	public final void setUndoManager(StdUndoManager mgr) {
		mUndoManager = mgr;
	}

	/** @param edit The {@link UndoableEdit} to add. */
	public final void addEdit(UndoableEdit edit) {
		if (mUndoManager != null) {
			mUndoManager.addEdit(edit);
		}
	}
}
