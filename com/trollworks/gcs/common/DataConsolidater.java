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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.advantage.AdvantageListWindow;
import com.trollworks.gcs.app.Main;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.equipment.EquipmentListWindow;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.SkillListWindow;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.spell.SpellListWindow;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.gcs.utility.Fonts;
import com.trollworks.gcs.utility.collections.FilteredList;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.widgets.outline.ListRow;

import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.util.ArrayList;

/**
 * Loads the files in the GCS 'data' directory and writes them into a single file. This class is
 * here for an experiment I'm working on...
 */
public class DataConsolidater {
	private static final String	ROOT_TAG			= "gcs_data";		//$NON-NLS-1$
	private static final String	TEMPLATES_ROOT_TAG	= "template_list";	//$NON-NLS-1$

	/**
	 * Loads the files in the GCS 'data' directory and writes them into a single file.
	 * 
	 * @param args The command line arguments. Ignored.
	 */
	public static final void main(String[] args) {
		Images.addLocation(Main.class.getResource("ui/common/images/")); //$NON-NLS-1$
		Preferences.setPreferenceFile("gcs.pref"); //$NON-NLS-1$
		Fonts.loadFromPreferences();

		ArrayList<Advantage> advantages = new ArrayList<Advantage>();
		ArrayList<Equipment> equipment = new ArrayList<Equipment>();
		ArrayList<Skill> skills = new ArrayList<Skill>();
		ArrayList<Spell> spells = new ArrayList<Spell>();
		ArrayList<Template> templates = new ArrayList<Template>();
		ArrayList<File> files = new ArrayList<File>();
		getFiles(ListCollectionThread.get().getLists(), files);

		for (File file : files) {
			String extension = Path.getExtension(file.getName());

			System.out.println("Trying to load: " + file.getAbsolutePath()); //$NON-NLS-1$
			try {
				if (AdvantageListWindow.EXTENSION.equals(extension)) {
					advantages.addAll(new FilteredList<Advantage>((new AdvantageList(file)).getTopLevelRows(), Advantage.class));
				} else if (EquipmentListWindow.EXTENSION.equals(extension)) {
					equipment.addAll(new FilteredList<Equipment>((new EquipmentList(file)).getTopLevelRows(), Equipment.class));
				} else if (SkillListWindow.EXTENSION.equals(extension)) {
					skills.addAll(new FilteredList<Skill>((new SkillList(file)).getTopLevelRows(), Skill.class));
				} else if (SpellListWindow.EXTENSION.equals(extension)) {
					spells.addAll(new FilteredList<Spell>((new SpellList(file)).getTopLevelRows(), Spell.class));
				} else if (TemplateWindow.EXTENSION.equals(extension)) {
					templates.add(new Template(file));
				} else {
					System.out.println("Unknown type: " + extension); //$NON-NLS-1$
				}
			} catch (Exception exception) {
				exception.printStackTrace();
			}
		}

		System.out.println("Saving..."); //$NON-NLS-1$
		try {
			XMLWriter out = new XMLWriter(new BufferedOutputStream(new FileOutputStream("gcs.gdf"))); //$NON-NLS-1$
			out.writeHeader();
			out.startSimpleTagEOL(ROOT_TAG);

			saveList(AdvantageList.TAG_ROOT, advantages, out);
			saveList(SkillList.TAG_ROOT, skills, out);
			saveList(SpellList.TAG_ROOT, spells, out);
			saveList(EquipmentList.TAG_ROOT, equipment, out);

			if (!templates.isEmpty()) {
				out.startSimpleTagEOL(TEMPLATES_ROOT_TAG);
				for (Template one : templates) {
					one.save(out);
				}
				out.endTagEOL(TEMPLATES_ROOT_TAG, true);
			}

			out.endTagEOL(ROOT_TAG, true);
			out.close();
		} catch (Exception exception) {
			exception.printStackTrace();
		}
	}

	private static void saveList(String tag, ArrayList<? extends ListRow> rows, XMLWriter out) {
		if (!rows.isEmpty()) {
			out.startSimpleTagEOL(tag);
			for (ListRow row : rows) {
				row.save(out, false);
			}
			out.endTagEOL(tag, true);
		}
	}

	@SuppressWarnings("unchecked") private static void getFiles(ArrayList<Object> list, ArrayList<File> files) {
		for (Object one : list) {
			if (one instanceof File) {
				files.add((File) one);
			} else if (one instanceof ArrayList) {
				getFiles((ArrayList<Object>) one, files);
			}
		}
	}
}
