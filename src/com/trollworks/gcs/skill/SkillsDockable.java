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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.widgets.LibraryDockable;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

import javax.swing.Icon;

/** A list of skills from a library. */
public class SkillsDockable extends LibraryDockable {
	@Localize("Skills")
	private static String	TITLE;

	static {
		Localization.initialize();
	}

	/** Creates a new {@link SkillsDockable}. */
	public SkillsDockable(LibraryFile file) {
		super(file);
	}

	@Override
	public String getDescriptor() {
		// RAW: Implement
		return null;
	}

	@Override
	public Icon getTitleIcon() {
		return GCSImages.getSkillIcon(false, false);
	}

	@Override
	public String getTitle() {
		return TITLE;
	}

	@Override
	protected ListOutline createOutline() {
		LibraryFile file = getDataFile();
		SkillList list = file.getSkillList();
		list.addTarget(this, Skill.ID_CATEGORY);
		return new SkillOutline(file, list.getModel());
	}

	@Override
	public ListFile getList() {
		return getDataFile().getSkillList();
	}
}
