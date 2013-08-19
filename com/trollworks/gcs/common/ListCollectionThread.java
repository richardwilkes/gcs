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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.menu.data.DataMenu;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.ttk.utility.App;
import com.trollworks.ttk.utility.Debug;
import com.trollworks.ttk.utility.Path;

import java.awt.EventQueue;
import java.io.File;
import java.util.ArrayList;
import java.util.Arrays;

/** A thread that periodically updates the set of available list files. */
public class ListCollectionThread extends Thread {
	private static final ListCollectionThread	INSTANCE;
	private File								mListDir;
	private ArrayList<Object>					mLists;
	private boolean								mRunning;

	static {
		INSTANCE = new ListCollectionThread();
		INSTANCE.start();
	}

	/** @return The one and only instance of this thread. */
	public static final ListCollectionThread get() {
		return INSTANCE;
	}

	private ListCollectionThread() {
		super("List Collection"); //$NON-NLS-1$
		setPriority(NORM_PRIORITY);
		setDaemon(true);
		mListDir = new File(App.APP_HOME_DIR, "data").getAbsoluteFile(); //$NON-NLS-1$ 
	}

	/** @return The current list of lists. */
	public ArrayList<Object> getLists() {
		try {
			while (mLists == null) {
				sleep(100);
			}
		} catch (InterruptedException outerIEx) {
			// Someone is tring to terminate us... let them.
		}
		return mLists == null ? new ArrayList<Object>() : mLists;
	}

	@Override
	public void run() {
		if (mRunning) {
			DataMenu.update();
		} else {
			mRunning = true;
			try {
				while (true) {
					ArrayList<Object> lists = collectLists(mListDir, 0);
					if (!lists.equals(mLists)) {
						mLists = lists;
						EventQueue.invokeLater(this);
					}
					sleep(5000);
				}
			} catch (InterruptedException outerIEx) {
				// Someone is tring to terminate us... let them.
			}
		}
	}

	private static ArrayList<Object> collectLists(File dir, int depth) {
		ArrayList<Object> list = new ArrayList<Object>();
		try {
			File[] files = dir.listFiles();
			if (files != null) {
				Arrays.sort(files);
				list.add(dir.getName());
				for (File element : files) {
					if (element.isDirectory()) {
						if (depth < 5) {
							ArrayList<Object> subList = collectLists(element, depth + 1);
							if (!subList.isEmpty()) {
								list.add(subList);
							}
						}
					} else {
						String ext = Path.getExtension(element.getName());
						if (LibraryFile.EXTENSION.equalsIgnoreCase(ext) || Advantage.OLD_ADVANTAGE_EXTENSION.equalsIgnoreCase(ext) || Equipment.OLD_EQUIPMENT_EXTENSION.equalsIgnoreCase(ext) || Skill.OLD_SKILL_EXTENSION.equalsIgnoreCase(ext) || Spell.OLD_SPELL_EXTENSION.equalsIgnoreCase(ext) || TemplateWindow.EXTENSION.equalsIgnoreCase(ext)) {
							list.add(element);
						}
					}
				}
				if (list.size() == 1) {
					list.clear();
				}
			}
		} catch (Exception exception) {
			assert false : Debug.throwableToString(exception);
		}
		return list;
	}
}
