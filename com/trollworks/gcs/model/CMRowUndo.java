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

import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.undo.TKSimpleUndo;
import com.trollworks.toolkit.undo.TKUndoException;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.InputStreamReader;
import java.text.MessageFormat;
import java.util.zip.GZIPInputStream;
import java.util.zip.GZIPOutputStream;

/** An undo for the entire row, with the exception of its children. */
public class CMRowUndo extends TKSimpleUndo {
	private CMDataFile	mDataFile;
	private CMRow		mRow;
	private String		mName;
	private byte[]		mBefore;
	private byte[]		mAfter;

	/**
	 * Creates a new {@link CMRowUndo}.
	 * 
	 * @param row The row being undone.
	 */
	public CMRowUndo(CMRow row) {
		super();
		mRow = row;
		mDataFile = mRow.getDataFile();
		mName = MessageFormat.format(Msgs.UNDO_FORMAT, mRow.getLocalizedName());
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

	private byte[] serialize(CMRow row) {
		try {
			ByteArrayOutputStream baos = new ByteArrayOutputStream();
			GZIPOutputStream gos = new GZIPOutputStream(baos);
			TKXMLWriter writer = new TKXMLWriter(gos);

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
			TKXMLReader reader = new TKXMLReader(new InputStreamReader(new GZIPInputStream(new ByteArrayInputStream(buffer))));
			TKXMLNodeType type = reader.next();

			while (type != TKXMLNodeType.END_DOCUMENT) {
				if (type == TKXMLNodeType.START_TAG) {
					mRow.load(reader, true);
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

	@Override public void apply(boolean forUndo) throws TKUndoException {
		super.apply(forUndo);
		deserialize(forUndo ? mBefore : mAfter);
	}

	/** @return The {@link CMDataFile} this undo works on. */
	public CMDataFile getDataFile() {
		return mDataFile;
	}

	/** @return The row this undo works on. */
	public CMRow getRow() {
		return mRow;
	}

	@Override public String getName() {
		return mName;
	}
}
