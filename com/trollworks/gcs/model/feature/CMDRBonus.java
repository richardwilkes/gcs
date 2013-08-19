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

package com.trollworks.gcs.model.feature;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.toolkit.collections.TKEnumExtractor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;

/** A DR bonus. */
public class CMDRBonus extends CMBonus {
	/** The XML tag. */
	public static final String	TAG_ROOT		= "dr_bonus";	//$NON-NLS-1$
	private static final String	TAG_LOCATION	= "location";	//$NON-NLS-1$
	private CMHitLocation		mLocation;

	/** Creates a new DR bonus. */
	public CMDRBonus() {
		super(1);
		mLocation = CMHitLocation.TORSO;
	}

	/**
	 * Loads a {@link CMDRBonus}.
	 * 
	 * @param reader The XML reader to use.
	 * @throws IOException
	 */
	public CMDRBonus(TKXMLReader reader) throws IOException {
		this();
		load(reader);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public CMDRBonus(CMDRBonus other) {
		super(other);
		mLocation = other.mLocation;
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMDRBonus && super.equals(obj)) {
			return mLocation == ((CMDRBonus) obj).mLocation;
		}
		return false;
	}

	public String getXMLTag() {
		return TAG_ROOT;
	}

	public String getKey() {
		StringBuffer buffer = new StringBuffer();

		buffer.append(CMCharacter.DR_PREFIX);
		buffer.append(mLocation.name());
		return buffer.toString();
	}

	public CMFeature cloneFeature() {
		return new CMDRBonus(this);
	}

	@Override protected void loadSelf(TKXMLReader reader) throws IOException {
		if (TAG_LOCATION.equals(reader.getName())) {
			setLocation((CMHitLocation) TKEnumExtractor.extract(reader.readText(), CMHitLocation.values(), CMHitLocation.TORSO));
		} else {
			super.loadSelf(reader);
		}
	}

	/**
	 * Saves the bonus.
	 * 
	 * @param out The XML writer to use.
	 */
	public void save(TKXMLWriter out) {
		out.startSimpleTagEOL(TAG_ROOT);
		out.simpleTag(TAG_LOCATION, mLocation.name().toLowerCase());
		saveBase(out);
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return The location protected by the DR. */
	public CMHitLocation getLocation() {
		return mLocation;
	}

	/** @param location The location. */
	public void setLocation(CMHitLocation location) {
		mLocation = location;
	}
}
