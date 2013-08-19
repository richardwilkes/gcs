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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.common;

import com.trollworks.gcs.utility.Debug;
import com.trollworks.gcs.utility.StdUndoManager;
import com.trollworks.gcs.utility.UniqueID;
import com.trollworks.gcs.utility.io.SafeFileUpdater;
import com.trollworks.gcs.utility.io.xml.XMLNodeType;
import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.notification.NotifierTarget;

import java.awt.image.BufferedImage;
import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.FileReader;
import java.io.IOException;

import javax.swing.undo.UndoableEdit;

/** A common super class for all data file-based model objects. */
public abstract class DataFile {
	private static final String	ATTRIBUTE_UNIQUE_ID	= "unique_id";	//$NON-NLS-1$
	private static final String	ATTRIBUTE_VERSION	= "version";	//$NON-NLS-1$
	private File				mFile;
	private int					mVersion;
	private UniqueID			mUniqueID;
	private Notifier			mNotifier;
	private boolean				mModified;
	private StdUndoManager		mUndoManager;

	/** Creates a new data file object. */
	protected DataFile() {
		mUniqueID = new UniqueID();
		mVersion = getXMLTagVersion();
		mNotifier = new Notifier();
	}

	/**
	 * Creates a new data file object from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @throws IOException if the data cannot be read or the file doesn't contain valid information.
	 */
	protected DataFile(File file) throws IOException {
		this(file, null);
	}

	/**
	 * Creates a new data file object from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @param param A parameter to pass through to the call to {@link #loadSelf(XMLReader,Object)}.
	 * @throws IOException if the data cannot be read or the file doesn't contain valid information.
	 */
	protected DataFile(File file, Object param) throws IOException {
		mVersion = getXMLTagVersion();
		mFile = file;
		mNotifier = new Notifier();
		XMLReader reader = new XMLReader(new FileReader(file));
		XMLNodeType type = reader.next();
		boolean found = false;

		while (type != XMLNodeType.END_DOCUMENT) {
			if (type == XMLNodeType.START_TAG) {
				String name = reader.getName();

				if (getXMLTagName().equals(name)) {
					if (!found) {
						found = true;
						mUniqueID = new UniqueID(reader.getAttribute(ATTRIBUTE_UNIQUE_ID));
						mVersion = reader.getAttributeAsInteger(ATTRIBUTE_VERSION, 0);
						loadSelf(reader, param);
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
		reader.close();
		setModified(false);
	}

	/** @return The version. */
	public int getVersion() {
		return mVersion;
	}

	/**
	 * Called to load the data file.
	 * 
	 * @param reader The XML reader to load data from.
	 * @param param A parameter passed through from the constructor.
	 * @throws IOException
	 */
	protected abstract void loadSelf(XMLReader reader, Object param) throws IOException;

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
			XMLWriter out = new XMLWriter(new BufferedOutputStream(new FileOutputStream(transaction.getTransactionFile(file))));

			out.writeHeader();
			save(out);
			out.close();
			if (out.checkError()) {
				transaction.abort();
			} else {
				transaction.commit();
				setModified(false);
				success = true;
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
	 */
	public void save(XMLWriter out) {
		out.startTag(getXMLTagName());
		out.writeAttribute(ATTRIBUTE_UNIQUE_ID, getUniqueID().toString());
		out.finishTagEOL();
		saveSelf(out);
		out.endTagEOL(getXMLTagName(), true);
	}

	/**
	 * Called to save the data file.
	 * 
	 * @param out The XML writer to use.
	 */
	protected abstract void saveSelf(XMLWriter out);

	/** @return The most recent version of the XML tag this object knows how to load. */
	public abstract int getXMLTagVersion();

	/** @return The XML root container tag name for this particular file. */
	public abstract String getXMLTagName();

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
		mFile = file;
	}

	/** @return The unique ID for this data file. */
	public UniqueID getUniqueID() {
		return mUniqueID;
	}

	/** @return <code>true</code> if the data has been modified. */
	public boolean isModified() {
		return mModified;
	}

	/** @param modified Whether or not the data has been modified. */
	public void setModified(boolean modified) {
		mModified = modified;
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
