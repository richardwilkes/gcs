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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.library.LibraryDockable;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

/** A list of skills from a library. */
public class SkillsDockable extends LibraryDockable {
	@Localize("Untitled Skills")
	@Localize(locale = "de", value = "Unbenannte Fertigkeiten-Liste")
	@Localize(locale = "ru", value = "Безымянный список умений")
	private static String	UNTITLED;

	static {
		Localization.initialize();
	}

	/** Creates a new {@link SkillsDockable}. */
	public SkillsDockable(SkillList list) {
		super(list);
	}

	@Override
	public SkillList getDataFile() {
		return (SkillList) super.getDataFile();
	}

	@Override
	protected String getUntitledBaseName() {
		return UNTITLED;
	}

	@Override
	protected ListOutline createOutline() {
		SkillList list = getDataFile();
		list.addTarget(this, Skill.ID_CATEGORY);
		return new SkillOutline(list, list.getModel());
	}
}
