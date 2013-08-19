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

package com.trollworks.gcs.model.modifier;

import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKRow;

import java.awt.image.BufferedImage;
import java.io.IOException;

/** Data Object to hold several {@link CMModifier} */
public class CMModifierList extends CMListFile {
	/** The XML tag for (dis)advantage lists. */
	public static final String	TAG_ROOT	= "modifier_list";	//$NON-NLS-1$

	/** Creates new {@link CMModifierList}. */
	public CMModifierList() {
		super(CMModifier.ID_LIST_CHANGED);
	}

	/**
	 * Creates a new {@link CMModifierList}.
	 * 
	 * @param modifiers The {@link CMModifierList} to clone.
	 */
	public CMModifierList(CMModifierList modifiers) {
		this();
		for (TKRow Row : modifiers.getModel().getRows()) {
			getModel().getRows().add(Row);
		}
	}

	@Override public CMRow createNewRow(boolean isContainer) {
		return new CMModifier(this);
	}

	@Override protected void loadList(TKXMLReader reader) throws IOException {
		TKOutlineModel model = getModel();
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMModifier.TAG_MODIFIER.equals(name)) {
					model.addRow(new CMModifier(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	@Override public BufferedImage getFileIcon(boolean large) {
		return null;
	}

	@Override public String getXMLTagName() {
		return TAG_ROOT;
	}
}
