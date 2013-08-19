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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.xml.XMLNodeType;
import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.InputStreamReader;
import java.text.MessageFormat;
import java.util.zip.GZIPInputStream;
import java.util.zip.GZIPOutputStream;

import javax.swing.undo.AbstractUndoableEdit;
import javax.swing.undo.CannotRedoException;
import javax.swing.undo.CannotUndoException;

/** An undo for the entire row, with the exception of its children. */
public class RowUndo extends AbstractUndoableEdit {
	private static String	MSG_UNDO_FORMAT;
	private DataFile		mDataFile;
	private ListRow			mRow;
	private String			mName;
	private byte[]			mBefore;
	private byte[]			mAfter;

	static {
		LocalizedMessages.initialize(RowUndo.class);
	}

	/**
	 * Creates a new {@link RowUndo}.
	 * 
	 * @param row The row being undone.
	 */
	public RowUndo(ListRow row) {
		super();
		mRow = row;
		mDataFile = mRow.getDataFile();
		mName = MessageFormat.format(MSG_UNDO_FORMAT, mRow.getLocalizedName());
		mBefore = serialize(mRow);
	}

	/**
	 * Call to finish capturing the undo state.
	 * 
	 * @return <code>true</code> if there is a difference between the before and after state.
	 */
	public boolean finish() {
		mAfter = serialize(mRow);
		if (mBefore.length != mAfter.length) {
			return true;
		}
		for (int i = 0; i < mBefore.length; i++) {
			if (mBefore[i] != mAfter[i]) {
				return true;
			}
		}
		return false;
	}

	private byte[] serialize(ListRow row) {
		try {
			ByteArrayOutputStream baos = new ByteArrayOutputStream();
			GZIPOutputStream gos = new GZIPOutputStream(baos);
			XMLWriter writer = new XMLWriter(gos);

			row.save(writer, true);
			writer.close();
			return baos.toByteArray();
		} catch (Exception exception) {
			exception.printStackTrace(System.err);
		}
		return new byte[0];
	}

	private void deserialize(byte[] buffer) {
		try {
			XMLReader reader = new XMLReader(new InputStreamReader(new GZIPInputStream(new ByteArrayInputStream(buffer))));
			XMLNodeType type = reader.next();
			LoadState state = new LoadState();
			state.mForUndo = true;
			while (type != XMLNodeType.END_DOCUMENT) {
				if (type == XMLNodeType.START_TAG) {
					mRow.load(reader, state);
					type = reader.getType();
				} else {
					type = reader.next();
				}
			}
			reader.close();
		} catch (Exception exception) {
			exception.printStackTrace(System.err);
		}
	}

	@Override public void undo() throws CannotUndoException {
		super.undo();
		deserialize(mBefore);
	}

	@Override public void redo() throws CannotRedoException {
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

	@Override public String getPresentationName() {
		return mName;
	}
}
