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

import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.OutlineModel;
import com.trollworks.gcs.widgets.outline.Row;

import java.io.File;
import java.io.IOException;
import java.util.List;

/** A list of rows. */
public abstract class ListFile extends DataFile {
	private OutlineModel	mModel;

	/**
	 * Creates a new, empty row list.
	 * 
	 * @param listChangedID The ID to use for "list changed".
	 */
	public ListFile(@SuppressWarnings("unused") String listChangedID) {
		super();
		mModel = new OutlineModel();
		initialize();
	}

	/**
	 * Creates a new row list from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @param listChangedID The ID to use for "list changed".
	 * @throws IOException if the data cannot be read or the file doesn't contain a valid list.
	 */
	public ListFile(File file, String listChangedID) throws IOException {
		super(file, listChangedID);
		initialize();
	}

	@Override protected final void loadSelf(XMLReader reader, Object param) throws IOException {
		mModel = new OutlineModel();
		loadList(reader);
	}

	/**
	 * Called to load the individual rows.
	 * 
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	protected abstract void loadList(XMLReader reader) throws IOException;

	@Override protected final void saveSelf(XMLWriter out) {
		for (Row row2 : getTopLevelRows()) {
			ListRow row = (ListRow) row2;

			row.save(out, false);
		}
	}

	/** @return The top-level rows in this list. */
	public List<Row> getTopLevelRows() {
		return mModel.getTopLevelRows();
	}

	/** @return The outline model. */
	public OutlineModel getModel() {
		return mModel;
	}
}
