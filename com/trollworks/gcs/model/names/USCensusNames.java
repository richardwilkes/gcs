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

package com.trollworks.gcs.model.names;

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

	@Override public String getLastName() {
		return LAST[RANDOM.nextInt(LAST.length)];
	}

	@Override public String getFemaleFirstName() {
		return FEMALE[RANDOM.nextInt(FEMALE.length)];
	}

	@Override public String getMaleFirstName() {
		return MALE[RANDOM.nextInt(MALE.length)];
	}
}
