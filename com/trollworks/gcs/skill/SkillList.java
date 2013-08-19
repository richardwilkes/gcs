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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.xml.XMLNodeType;
import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.widgets.outline.OutlineModel;

import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;

/** A list of skills. */
public class SkillList extends ListFile {
	private static final int	CURRENT_VERSION	= 1;
	/** The XML tag for Skill lists. */
	public static final String	TAG_ROOT		= "skill_list"; //$NON-NLS-1$

	/** Creates a new, empty skills list. */
	public SkillList() {
		super(Skill.ID_LIST_CHANGED);
	}

	/**
	 * Creates a new skills list from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @throws IOException if the data cannot be read or the file doesn't contain a valid skills
	 *             list.
	 */
	public SkillList(File file) throws IOException {
		super(file, Skill.ID_LIST_CHANGED);
	}

	@Override public int getXMLTagVersion() {
		return CURRENT_VERSION;
	}

	@Override public String getXMLTagName() {
		return TAG_ROOT;
	}

	@Override public BufferedImage getFileIcon(boolean large) {
		return Images.getSkillIcon(large, false);
	}

	@Override protected void loadList(XMLReader reader) throws IOException {
		OutlineModel model = getModel();
		String marker = reader.getMarker();

		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String name = reader.getName();

				if (Skill.TAG_SKILL.equals(name) || Skill.TAG_SKILL_CONTAINER.equals(name)) {
					model.addRow(new Skill(this, reader), true);
				} else if (Technique.TAG_TECHNIQUE.equals(name)) {
					model.addRow(new Technique(this, reader), true);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}
}
