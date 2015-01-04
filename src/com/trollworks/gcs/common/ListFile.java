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

package com.trollworks.gcs.common;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;

import java.io.IOException;
import java.util.List;
import java.util.TreeSet;

/** A list of rows. */
public abstract class ListFile extends DataFile {
	private OutlineModel	mModel	= new OutlineModel();

	@Override
	protected final void loadSelf(XMLReader reader, LoadState state) throws IOException {
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

	/** @return The set of categories that exist in this {@link ListFile}. */
	public TreeSet<String> getCategories() {
		TreeSet<String> set = new TreeSet<>();
		for (Row row : getTopLevelRows()) {
			processRowForCategories(row, set);
		}
		return set;
	}

	private void processRowForCategories(Row row, TreeSet<String> set) {
		if (row instanceof ListRow) {
			set.addAll(((ListRow) row).getCategories());
		}
		if (row.hasChildren()) {
			for (Row child : row.getChildren()) {
				processRowForCategories(child, set);
			}
		}
	}
}
