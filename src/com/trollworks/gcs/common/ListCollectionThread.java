/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.common;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.menu.data.DataMenu;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.utility.Debug;
import com.trollworks.toolkit.utility.Path;

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
		return mLists == null ? new ArrayList<>() : mLists;
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
				// Someone is trying to terminate us... let them.
			}
		}
	}

	private static ArrayList<Object> collectLists(File dir, int depth) {
		ArrayList<Object> list = new ArrayList<>();
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
			assert false : Debug.toString(exception);
		}
		return list;
	}
}
