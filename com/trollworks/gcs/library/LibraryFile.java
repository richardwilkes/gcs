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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.library;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.Path;
import com.trollworks.ttk.widgets.DataModifiedListener;
import com.trollworks.ttk.widgets.WindowUtils;
import com.trollworks.ttk.widgets.outline.Row;
import com.trollworks.ttk.xml.XMLNodeType;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.awt.EventQueue;
import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;
import java.text.MessageFormat;
import java.util.TreeSet;

/** Holds the contents of a library file. */
public class LibraryFile extends DataFile implements DataModifiedListener {
	static String				MSG_WARNING;
	/** The current version. */
	public static final int		CURRENT_VERSION	= 1;
	/** The XML tag for library files. */
	public static final String	TAG_ROOT		= "gcs_library";	//$NON-NLS-1$
	/** The extension for library files. */
	public static final String	EXTENSION		= ".glb";			//$NON-NLS-1$
	private AdvantageList		mAdvantages;
	private SkillList			mSkills;
	private SpellList			mSpells;
	private EquipmentList		mEquipment;
	private boolean				mImported;
	private String				mSuggestedName;

	static {
		LocalizedMessages.initialize(LibraryFile.class);
	}

	/** Creates a new, empty, {@link LibraryFile}. */
	public LibraryFile() {
		super();
		setup();
		initialize();
	}

	/**
	 * Creates a new {@link LibraryFile} from the specified file.
	 * 
	 * @param file The file to load the data from.
	 * @throws IOException if the data cannot be read or the file doesn't contain valid information.
	 */
	public LibraryFile(final File file) throws IOException {
		super(file, new LoadState());
		initialize();
		if (mImported) {
			mSuggestedName = Path.getLeafName(file.getName(), false);
			setFile(null);
			EventQueue.invokeLater(new Runnable() {
				@Override
				public void run() {
					WindowUtils.showWarning(null, MessageFormat.format(MSG_WARNING, file.getName()));
				}
			});
		}
	}

	/** @return Whether this file was imported and not saved yet. */
	public boolean wasImported() {
		return mImported && getFile() == null;
	}

	/** @return The suggested file name to use after an import. */
	public String getSuggestedFileNameFromImport() {
		return mSuggestedName;
	}

	private void setup() {
		mAdvantages = new AdvantageList();
		mAdvantages.addDataModifiedListener(this);
		mSkills = new SkillList();
		mSkills.addDataModifiedListener(this);
		mSpells = new SpellList();
		mSpells.addDataModifiedListener(this);
		mEquipment = new EquipmentList();
		mEquipment.addDataModifiedListener(this);
	}

	@Override
	public BufferedImage getFileIcon(boolean large) {
		return GCSImages.getLibraryIcon(large);
	}

	@Override
	public boolean matchesRootTag(String name) {
		return TAG_ROOT.equals(name) || AdvantageList.TAG_ROOT.equals(name) || SkillList.TAG_ROOT.equals(name) || SpellList.TAG_ROOT.equals(name) || EquipmentList.TAG_ROOT.equals(name);
	}

	@Override
	public String getXMLTagName() {
		return TAG_ROOT;
	}

	@Override
	public int getXMLTagVersion() {
		return CURRENT_VERSION;
	}

	@Override
	protected void loadSelf(XMLReader reader, LoadState state) throws IOException {
		setup();
		String name = reader.getName();
		if (TAG_ROOT.equals(name)) {
			String marker = reader.getMarker();
			do {
				if (reader.next() == XMLNodeType.START_TAG) {
					name = reader.getName();
					if (AdvantageList.TAG_ROOT.equals(name)) {
						mAdvantages.load(reader, state);
					} else if (SkillList.TAG_ROOT.equals(name)) {
						mSkills.load(reader, state);
					} else if (SpellList.TAG_ROOT.equals(name)) {
						mSpells.load(reader, state);
					} else if (EquipmentList.TAG_ROOT.equals(name)) {
						mEquipment.load(reader, state);
					} else {
						reader.skipTag(name);
					}
				}
			} while (reader.withinMarker(marker));
		} else if (AdvantageList.TAG_ROOT.equals(name)) {
			mImported = true;
			mAdvantages.load(reader, state);
		} else if (SkillList.TAG_ROOT.equals(name)) {
			mImported = true;
			mSkills.load(reader, state);
		} else if (SpellList.TAG_ROOT.equals(name)) {
			mImported = true;
			mSpells.load(reader, state);
		} else if (EquipmentList.TAG_ROOT.equals(name)) {
			mImported = true;
			mEquipment.load(reader, state);
		}
	}

	@Override
	protected void saveSelf(XMLWriter out) {
		mAdvantages.save(out, false, true);
		mSkills.save(out, false, true);
		mSpells.save(out, false, true);
		mEquipment.save(out, false, true);
	}

	/** @return The {@link AdvantageList}. */
	public AdvantageList getAdvantageList() {
		return mAdvantages;
	}

	/** @return The {@link SkillList}. */
	public SkillList getSkillList() {
		return mSkills;
	}

	/** @return The {@link SpellList}. */
	public SpellList getSpellList() {
		return mSpells;
	}

	/** @return The {@link EquipmentList}. */
	public EquipmentList getEquipmentList() {
		return mEquipment;
	}

	/**
	 * @param file The {@link ListFile} to process.
	 * @return The set of categories for the specified {@link ListFile}.
	 */
	public TreeSet<String> getCategoriesFor(ListFile file) {
		TreeSet<String> set = new TreeSet<String>();
		for (Row row : file.getTopLevelRows()) {
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

	@Override
	public void dataModificationStateChanged(Object obj, boolean modified) {
		setModified(mAdvantages.isModified() | mSkills.isModified() | mSpells.isModified() | mEquipment.isModified());
	}
}
