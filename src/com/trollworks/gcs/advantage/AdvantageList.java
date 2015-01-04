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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.ui.image.StdImageSet;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;

import java.io.IOException;

/** A list of Advantages. */
public class AdvantageList extends ListFile {
	/** The current version. */
	public static final int		CURRENT_VERSION	= 1;
	/** The XML tag for {@link AdvantageList}s. */
	public static final String	TAG_ROOT		= "advantage_list"; //$NON-NLS-1$
	/** The extension for {@link AdvantageList}s. */
	public static final String	EXTENSION		= "adq";			//$NON-NLS-1$

	@Override
	public int getXMLTagVersion() {
		return CURRENT_VERSION;
	}

	@Override
	public String getXMLTagName() {
		return TAG_ROOT;
	}

	@Override
	public String getExtension() {
		return EXTENSION;
	}

	@Override
	public StdImageSet getFileIcons() {
		return GCSImages.getAdvantagesDocumentIcons();
	}

	@Override
	protected void loadList(XMLReader reader, LoadState state) throws IOException {
		OutlineModel model = getModel();
		String marker = reader.getMarker();
		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String name = reader.getName();

				if (Advantage.TAG_ADVANTAGE.equals(name) || Advantage.TAG_ADVANTAGE_CONTAINER.equals(name)) {
					model.addRow(new Advantage(this, reader, state), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}
}
