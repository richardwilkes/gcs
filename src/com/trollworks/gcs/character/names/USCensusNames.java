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

package com.trollworks.gcs.character.names;

/** Generates random names from the 1990 U.S. census data. */
public class USCensusNames extends Names {
	/** The one and only global instance of this class. */
	public static final USCensusNames	INSTANCE	= new USCensusNames();
	private static final String[]		FEMALE		= loadNames(USCensusNames.class.getResource("USCensus1990FemaleFirstNames.txt"), "Mary");	//$NON-NLS-1$ //$NON-NLS-2$
	private static final String[]		MALE		= loadNames(USCensusNames.class.getResource("USCensus1990MaleFirstNames.txt"), "Richard");	//$NON-NLS-1$ //$NON-NLS-2$
	private static final String[]		LAST		= loadNames(USCensusNames.class.getResource("USCensus1990LastNames.txt"), "Wilkes");		//$NON-NLS-1$ //$NON-NLS-2$

	private USCensusNames() {
		// Just here to prevent external instantiation
	}

	@Override
	public String getLastName() {
		return LAST[RANDOM.nextInt(LAST.length)];
	}

	@Override
	public String getFemaleFirstName() {
		return FEMALE[RANDOM.nextInt(FEMALE.length)];
	}

	@Override
	public String getMaleFirstName() {
		return MALE[RANDOM.nextInt(MALE.length)];
	}
}
