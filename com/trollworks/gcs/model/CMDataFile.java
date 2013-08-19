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

package com.trollworks.gcs.model;

import com.trollworks.gcs.ui.preferences.CSGeneralPreferences;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.io.TKSafeFileUpdater;
import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.notification.TKNotifier;
import com.trollworks.toolkit.notification.TKNotifierTarget;
import com.trollworks.toolkit.undo.TKUndo;
import com.trollworks.toolkit.undo.TKUndoManager;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKUniqueID;

import java.awt.image.BufferedImage;
import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.FileReader;
import java.io.IOException;

/** A common super class for all data file-based model objects. */
public abstract class CMDataFile extends TKUndoManager implements TKNotifierTarget {
	private static final String	ATTRIBUTE_UNIQUE_ID	= "unique_id";	//$NON-NLS-1$
	private File				mFile;
	private TKUniqueID			mUniqueID;
	private TKNotifier			mNotifier;
	private boolean				mModified;
	private boolean				mSuspend;

	/** Creates a new data file object. */
	protected CMDataFile() {
		super(CSGeneralPreferences.getMaximumUndoLevels());
		mUniqueID = new TKUniqueID();
		mNotifier = new TKNotifier();
		TKPreferences.getInstance().getNotifier().add(this, CSGeneralPreferences.UNDO_LEVELS_PREF_KEY);
	}

	/**
	 * Creates a new data file object from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @throws IOException if the data cannot be read or the file doesn't contain valid information.
	 */
	protected CMDataFile(File file) throws IOException {
		this(file, null);
	}

	/**
	 * Creates a new data file object from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @param param A parameter to pass through to the call to {@link #loadSelf(TKXMLReader,Object)}.
	 * @throws IOException if the data cannot be read or the file doesn't contain valid information.
	 */
	protected CMDataFile(File file, Object param) throws IOException {
		super(CSGeneralPreferences.getMaximumUndoLevels());
		TKPreferences.getInstance().getNotifier().add(this, CSGeneralPreferences.UNDO_LEVELS_PREF_KEY);

		mFile = file;
		mNotifier = new TKNotifier();
		TKXMLReader reader = new TKXMLReader(new FileReader(file));
		TKXMLNodeType type = reader.next();
		boolean found = false;

		while (type != TKXMLNodeType.END_DOCUMENT) {
			if (type == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (getXMLTagName().equals(name)) {
					if (!found) {
						found = true;
						mUniqueID = new TKUniqueID(reader.getAttribute(ATTRIBUTE_UNIQUE_ID));
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

	/** @param suspend Whether to suspend undo or not. */
	public void suspendUndo(boolean suspend) {
		mSuspend = suspend;
	}

	@Override public synchronized boolean addEdit(TKUndo edit) {
		return mSuspend ? false : super.addEdit(edit);
	}

	/**
	 * Call this method when the data is no longer needed (i.e. the file is closed.
	 */
	public void noLongerNeeded() {
		TKPreferences.getInstance().getNotifier().remove(this);
	}

	/**
	 * Called to load the data file.
	 * 
	 * @param reader The XML reader to load data from.
	 * @param param A parameter passed through from the constructor.
	 * @throws IOException
	 */
	protected abstract void loadSelf(TKXMLReader reader, Object param) throws IOException;

	/**
	 * Saves the data out to the specified file. Does not affect the result of {@link #getFile()}.
	 * 
	 * @param file The file to write to.
	 * @return <code>true</code> on success.
	 */
	public boolean save(File file) {
		TKSafeFileUpdater transaction = new TKSafeFileUpdater();
		boolean success = false;

		transaction.begin();
		try {
			TKXMLWriter out = new TKXMLWriter(new BufferedOutputStream(new FileOutputStream(transaction.getTransactionFile(file))));

			out.writeHeader();
			out.startTag(getXMLTagName());
			out.writeAttribute(ATTRIBUTE_UNIQUE_ID, getUniqueID().toString());
			out.finishTagEOL();
			saveSelf(out);
			out.endTagEOL(getXMLTagName(), true);
			out.close();
			if (out.checkError()) {
				transaction.abort();
			} else {
				transaction.commit();
				setModified(false);
				success = true;
			}
		} catch (Exception exception) {
			if (TKDebug.isKeySet(TKDebug.KEY_DIAGNOSE_LOAD_SAVE)) {
				exception.printStackTrace(System.err);
			}
			transaction.abort();
		}
		return success;
	}

	/**
	 * Called to save the data file.
	 * 
	 * @param out The XML writer to use.
	 */
	protected abstract void saveSelf(TKXMLWriter out);

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
	public TKUniqueID getUniqueID() {
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

	/** Resets the underlying {@link TKNotifier} by removing all targets. */
	public void resetNotifier() {
		mNotifier.reset();
	}

	/**
	 * Resets the underlying {@link TKNotifier} by removing all targets except the specified ones.
	 * 
	 * @param exclude The {@link TKNotifierTarget}(s) to exclude.
	 */
	public void resetNotifier(TKNotifierTarget... exclude) {
		mNotifier.reset(exclude);
	}

	/**
	 * Registers a {@link TKNotifierTarget} with this data file's {@link TKNotifier}.
	 * 
	 * @param target The {@link TKNotifierTarget} to register.
	 * @param names The names to register for.
	 */
	public void addTarget(TKNotifierTarget target, String... names) {
		mNotifier.add(target, names);
	}

	/**
	 * Un-registers a {@link TKNotifierTarget} with this data file's {@link TKNotifier}.
	 * 
	 * @param target The {@link TKNotifierTarget} to un-register.
	 */
	public void removeTarget(TKNotifierTarget target) {
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

	public void handleNotification(Object producer, String type, Object data) {
		if (CSGeneralPreferences.UNDO_LEVELS_PREF_KEY.equals(type)) {
			setLimit(CSGeneralPreferences.getMaximumUndoLevels());
		}
	}
}
