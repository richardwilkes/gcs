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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.common;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.widgets.outline.OutlineModel;
import com.trollworks.ttk.widgets.outline.Row;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;
import java.util.List;

/** A list of rows. */
public abstract class ListFile extends DataFile {
	private OutlineModel	mModel;

	/** Creates a new, empty row list. */
	public ListFile() {
		super();
		mModel = new OutlineModel();
		initialize();
	}

	@Override
	protected final void loadSelf(XMLReader reader, LoadState state) throws IOException {
		mModel = new OutlineModel();
		loadList(reader, state);
	}

	/**
	 * Called to load the individual rows.
	 * 
	 * @param reader The XML reader to load from.
	 * @param state The {@link LoadState} to use.
	 */
	protected abstract void loadList(XMLReader reader, LoadState state) throws IOException;

	@Override
	protected final void saveSelf(XMLWriter out) {
		for (Row one : getTopLevelRows()) {
			((ListRow) one).save(out, false);
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

	@Override
	public boolean isEmpty() {
		return mModel.getRowCount() == 0;
	}
}
