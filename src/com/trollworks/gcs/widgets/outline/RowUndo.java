/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.Localization;

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
	@Localize("{0} Changes")
	@Localize(locale = "de", value = "{0} Änderungen")
	@Localize(locale = "ru", value = "{0} изменений")
	private static String	UNDO_FORMAT;

	static {
		Localization.initialize();
	}

	private DataFile		mDataFile;
	private ListRow			mRow;
	private String			mName;
	private byte[]			mBefore;
	private byte[]			mAfter;

	/**
	 * Creates a new {@link RowUndo}.
	 *
	 * @param row The row being undone.
	 */
	public RowUndo(ListRow row) {
		super();
		mRow = row;
		mDataFile = mRow.getDataFile();
		mName = MessageFormat.format(UNDO_FORMAT, mRow.getLocalizedName());
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

	private static byte[] serialize(ListRow row) {
		try {
			ByteArrayOutputStream baos = new ByteArrayOutputStream();
			GZIPOutputStream gos = new GZIPOutputStream(baos);
			try (XMLWriter writer = new XMLWriter(gos)) {
				row.save(writer, true);
			}
			return baos.toByteArray();
		} catch (Exception exception) {
			exception.printStackTrace(System.err);
		}
		return new byte[0];
	}

	private void deserialize(byte[] buffer) {
		try (XMLReader reader = new XMLReader(new InputStreamReader(new GZIPInputStream(new ByteArrayInputStream(buffer))))) {
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
		} catch (Exception exception) {
			exception.printStackTrace(System.err);
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
