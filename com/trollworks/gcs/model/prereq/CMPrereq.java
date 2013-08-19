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

package com.trollworks.gcs.model.prereq;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.util.HashMap;
import java.util.HashSet;

/** The abstract base class prerequisite criteria and prerequisite lists. */
public abstract class CMPrereq {
	/** The owning prerequisite list, if any. */
	protected CMPrereqList	mParent;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	protected CMPrereq(CMPrereqList parent) {
		mParent = parent;
	}

	/** @return The owning prerequisite list, if any. */
	public CMPrereqList getParent() {
		return mParent;
	}

	/** Removes this prerequisite from its parent. */
	public void removeFromParent() {
		if (mParent != null) {
			mParent.remove(this);
		}
	}

	/** @return The XML tag representing this prereq. */
	public abstract String getXMLTag();

	/**
	 * Saves the prerequisite.
	 * 
	 * @param out The XML writer to use.
	 */
	public abstract void save(TKXMLWriter out);

	/**
	 * @param character The character to check.
	 * @param exclude The data to exclude from the check.
	 * @param builder The {@link StringBuilder} to append this prerequisite's satisfied/unsatisfied
	 *            description to. May be <code>null</code>.
	 * @param prefix The prefix to add to each line appended to the builder.
	 * @return Whether or not this prerequisite is satisfied by the specified character.
	 */
	public abstract boolean satisfied(CMCharacter character, CMRow exclude, StringBuilder builder, String prefix);

	/**
	 * Creates a deep clone of the prerequisite.
	 * 
	 * @param parent The new owning prerequisite list, if any.
	 * @return The clone.
	 */
	public abstract CMPrereq clone(CMPrereqList parent);

	/** @param set The nameable keys. */
	public void fillWithNameableKeys(@SuppressWarnings("unused") HashSet<String> set) {
		// Do nothing by default
	}

	/** @param map The map of nameable keys to names to apply. */
	public void applyNameableKeys(@SuppressWarnings("unused") HashMap<String, String> map) {
		// Do nothing by default
	}
}
