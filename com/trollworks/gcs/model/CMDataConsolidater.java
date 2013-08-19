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

package com.trollworks.gcs.model;

import com.trollworks.gcs.CSMain;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.advantage.CMAdvantageList;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.equipment.CMEquipmentList;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMSkillList;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.model.spell.CMSpellList;
import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.gcs.ui.common.CSListCollectionThread;
import com.trollworks.gcs.ui.common.CSListOpener;
import com.trollworks.gcs.ui.template.CSTemplateOpener;
import com.trollworks.toolkit.collections.TKFilteredList;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.util.ArrayList;

/**
 * Loads the files in the GCS 'data' directory and writes them into a single file. This class is
 * here for an experiment I'm working on...
 */
public class CMDataConsolidater {
	private static final String	ROOT_TAG			= "gcs_data";		//$NON-NLS-1$
	private static final String	TEMPLATES_ROOT_TAG	= "template_list";	//$NON-NLS-1$

	/**
	 * Loads the files in the GCS 'data' directory and writes them into a single file.
	 * 
	 * @param args The command line arguments. Ignored.
	 */
	public static final void main(String[] args) {
		TKImage.addLocation(CSMain.class.getResource("ui/common/images/")); //$NON-NLS-1$
		TKPreferences.setPreferenceFile("gcs.pref"); //$NON-NLS-1$
		CSFont.initialize();

		ArrayList<CMAdvantage> advantages = new ArrayList<CMAdvantage>();
		ArrayList<CMEquipment> equipment = new ArrayList<CMEquipment>();
		ArrayList<CMSkill> skills = new ArrayList<CMSkill>();
		ArrayList<CMSpell> spells = new ArrayList<CMSpell>();
		ArrayList<CMTemplate> templates = new ArrayList<CMTemplate>();
		ArrayList<File> files = new ArrayList<File>();
		getFiles(CSListCollectionThread.get().getLists(), files);

		for (File file : files) {
			String extension = TKPath.getExtension(file.getName());

			System.out.println("Trying to load: " + file.getAbsolutePath()); //$NON-NLS-1$
			try {
				if (CSListOpener.ADVANTAGE_EXTENSION.equals(extension)) {
					advantages.addAll(new TKFilteredList<CMAdvantage>((new CMAdvantageList(file)).getTopLevelRows(), CMAdvantage.class));
				} else if (CSListOpener.EQUIPMENT_EXTENSION.equals(extension)) {
					equipment.addAll(new TKFilteredList<CMEquipment>((new CMEquipmentList(file)).getTopLevelRows(), CMEquipment.class));
				} else if (CSListOpener.SKILL_EXTENSION.equals(extension)) {
					skills.addAll(new TKFilteredList<CMSkill>((new CMSkillList(file)).getTopLevelRows(), CMSkill.class));
				} else if (CSListOpener.SPELL_EXTENSION.equals(extension)) {
					spells.addAll(new TKFilteredList<CMSpell>((new CMSpellList(file)).getTopLevelRows(), CMSpell.class));
				} else if (CSTemplateOpener.EXTENSION.equals(extension)) {
					templates.add(new CMTemplate(file));
				} else {
					System.out.println("Unknown type: " + extension); //$NON-NLS-1$
				}
			} catch (Exception exception) {
				exception.printStackTrace();
			}
		}

		System.out.println("Saving..."); //$NON-NLS-1$
		try {
			TKXMLWriter out = new TKXMLWriter(new BufferedOutputStream(new FileOutputStream("gcs.gdf"))); //$NON-NLS-1$
			out.writeHeader();
			out.startSimpleTagEOL(ROOT_TAG);

			if (!advantages.isEmpty()) {
				out.startSimpleTagEOL(CMAdvantageList.TAG_ROOT);
				for (CMAdvantage one : advantages) {
					one.save(out, false);
				}
				out.endTagEOL(CMAdvantageList.TAG_ROOT, true);
			}

			if (!skills.isEmpty()) {
				out.startSimpleTagEOL(CMSkillList.TAG_ROOT);
				for (CMSkill one : skills) {
					one.save(out, false);
				}
				out.endTagEOL(CMSkillList.TAG_ROOT, true);
			}

			if (!spells.isEmpty()) {
				out.startSimpleTagEOL(CMSpellList.TAG_ROOT);
				for (CMSpell one : spells) {
					one.save(out, false);
				}
				out.endTagEOL(CMSpellList.TAG_ROOT, true);
			}

			if (!equipment.isEmpty()) {
				out.startSimpleTagEOL(CMEquipmentList.TAG_ROOT);
				for (CMEquipment one : equipment) {
					one.save(out, false);
				}
				out.endTagEOL(CMEquipmentList.TAG_ROOT, true);
			}

			if (!templates.isEmpty()) {
				out.startSimpleTagEOL(TEMPLATES_ROOT_TAG);
				for (CMTemplate one : templates) {
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
