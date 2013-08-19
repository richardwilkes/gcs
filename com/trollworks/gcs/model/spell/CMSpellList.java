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

package com.trollworks.gcs.model.spell;

import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;

import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;

/** A list of spells. */
public class CMSpellList extends CMListFile {
	/** The XML tag for Skill lists. */
	public static final String	TAG_ROOT	= "spell_list"; //$NON-NLS-1$

	/** Creates a new, empty spells list. */
	public CMSpellList() {
		super(CMSpell.ID_LIST_CHANGED);
	}

	/**
	 * Creates a new spells list from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @throws IOException if the data cannot be read or the file doesn't contain a valid spells
	 *             list.
	 */
	public CMSpellList(File file) throws IOException {
		super(file, CMSpell.ID_LIST_CHANGED);
	}

	@Override public String getXMLTagName() {
		return TAG_ROOT;
	}

	@Override public BufferedImage getFileIcon(boolean large) {
		return CSImage.getSpellIcon(large, false);
	}

	@Override protected void loadList(TKXMLReader reader) throws IOException {
		TKOutlineModel model = getModel();
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (CMSpell.TAG_SPELL.equals(name) || CMSpell.TAG_SPELL_CONTAINER.equals(name)) {
					model.addRow(new CMSpell(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	@Override public CMRow createNewRow(boolean isContainer) {
		return new CMSpell(this, isContainer);
	}
}
