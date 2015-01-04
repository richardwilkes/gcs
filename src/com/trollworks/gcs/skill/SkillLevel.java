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

/** Provides simple storage for the skill level/relative level pair. */
public class SkillLevel {
	/** The skill level. */
	public int	mLevel;
	/** The relative skill level. */
	public int	mRelativeLevel;

	/**
	 * Creates a new {@link SkillLevel}.
	 * 
	 * @param level The skill level.
	 * @param relativeLevel The relative skill level.
	 */
	public SkillLevel(int level, int relativeLevel) {
		mLevel = level;
		mRelativeLevel = relativeLevel;
	}
}
