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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.ttk.xml.XMLWriter;

import java.util.HashMap;
import java.util.HashSet;

/** Describes a feature of an advantage, skill, spell, or piece of equipment. */
public interface Feature {
	/** @return The XML tag representing this feature. */
	String getXMLTag();

	/** @return The feature key used in the feature map. */
	String getKey();

	/** @return An exact clone of this feature. */
	Feature cloneFeature();

	/**
	 * Saves the feature.
	 * 
	 * @param out The XML writer to use.
	 */
	void save(XMLWriter out);

	/** @param set The nameable keys. */
	void fillWithNameableKeys(HashSet<String> set);

	/** @param map The map of nameable keys to names to apply. */
	void applyNameableKeys(HashMap<String, String> map);
}
