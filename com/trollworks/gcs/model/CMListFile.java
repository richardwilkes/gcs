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

import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKRow;

import java.io.File;
import java.io.IOException;
import java.util.List;

/** A list of rows. */
public abstract class CMListFile extends CMDataFile {
	private TKOutlineModel	mModel;

	/**
	 * Creates a new, empty row list.
	 * 
	 * @param listChangedID The ID to use for "list changed".
	 */
	public CMListFile(@SuppressWarnings("unused") String listChangedID) {
		super();
		mModel = new TKOutlineModel();
		initialize();
	}

	/**
	 * Creates a new row list from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @param listChangedID The ID to use for "list changed".
	 * @throws IOException if the data cannot be read or the file doesn't contain a valid list.
	 */
	public CMListFile(File file, String listChangedID) throws IOException {
		super(file, listChangedID);
		initialize();
	}

	@Override protected final void loadSelf(TKXMLReader reader, Object param) throws IOException {
		mModel = new TKOutlineModel();
		loadList(reader);
	}

	/**
	 * Called to load the individual rows.
	 * 
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	protected abstract void loadList(TKXMLReader reader) throws IOException;

	/**
	 * Called to create a new list row.
	 * 
	 * @param isContainer Whether or not the row should be a container.
	 * @return The new list row.
	 */
	public abstract CMRow createNewRow(boolean isContainer);

	@Override protected final void saveSelf(TKXMLWriter out) {
		for (TKRow row2 : getTopLevelRows()) {
			CMRow row = (CMRow) row2;

			row.save(out, false);
		}
	}

	/** @return The top-level rows in this list. */
	public List<TKRow> getTopLevelRows() {
		return mModel.getTopLevelRows();
	}

	/** @return The outline model. */
	public TKOutlineModel getModel() {
		return mModel;
	}
}
