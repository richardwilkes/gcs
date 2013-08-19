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

package com.trollworks.toolkit.widget.outline;

import java.awt.datatransfer.DataFlavor;
import java.awt.datatransfer.Transferable;
import java.awt.datatransfer.UnsupportedFlavorException;

/** Allows rows to be part of drag and drop operations internal to the JVM. */
public class TKRowSelection implements Transferable {
	/** The data flavor for this class. */
	public static final DataFlavor	DATA_FLAVOR	= new DataFlavor(TKRowSelection.class, "Outline Rows"); //$NON-NLS-1$
	private TKOutlineModel			mModel;
	private TKRow[]					mRows;
	private String					mCache;

	/**
	 * Creates a new transferable row object.
	 * 
	 * @param model The owning outline model.
	 * @param rows The rows to transfer.
	 */
	public TKRowSelection(TKOutlineModel model, TKRow[] rows) {
		mModel = model;
		mRows = new TKRow[rows.length];
		System.arraycopy(rows, 0, mRows, 0, rows.length);
	}

	public DataFlavor[] getTransferDataFlavors() {
		return new DataFlavor[] { DATA_FLAVOR, DataFlavor.stringFlavor };
	}

	public boolean isDataFlavorSupported(DataFlavor flavor) {
		return DATA_FLAVOR.equals(flavor) || DataFlavor.stringFlavor.equals(flavor);
	}

	public Object getTransferData(DataFlavor flavor) throws UnsupportedFlavorException {
		if (DATA_FLAVOR.equals(flavor)) {
			return mRows;
		}
		if (DataFlavor.stringFlavor.equals(flavor)) {
			if (mCache == null) {
				StringBuilder buffer = new StringBuilder();

				if (mRows.length > 0) {
					int count = mModel.getColumnCount();

					for (TKRow element : mRows) {
						boolean first = true;

						for (int j = 0; j < count; j++) {
							TKColumn column = mModel.getColumnAtIndex(j);

							if (column.isVisible()) {
								if (first) {
									first = false;
								} else {
									buffer.append('\t');
								}
								buffer.append(element.getDataAsText(column));
							}
						}
						buffer.append('\n');
					}
				}
				mCache = buffer.toString();
			}
			return mCache;
		}
		throw new UnsupportedFlavorException(flavor);
	}
}
