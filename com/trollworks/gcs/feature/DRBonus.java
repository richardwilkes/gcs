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

import com.trollworks.gcs.character.Armor;
import com.trollworks.ttk.collections.Enums;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;

/** A DR bonus. */
public class DRBonus extends Bonus {
	/** The XML tag. */
	public static final String	TAG_ROOT		= "dr_bonus";	//$NON-NLS-1$
	private static final String	TAG_LOCATION	= "location";	//$NON-NLS-1$
	private HitLocation			mLocation;

	/** Creates a new DR bonus. */
	public DRBonus() {
		super(1);
		mLocation = HitLocation.TORSO;
	}

	/**
	 * Loads a {@link DRBonus}.
	 * 
	 * @param reader The XML reader to use.
	 */
	public DRBonus(XMLReader reader) throws IOException {
		this();
		load(reader);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public DRBonus(DRBonus other) {
		super(other);
		mLocation = other.mLocation;
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof DRBonus && super.equals(obj)) {
			return mLocation == ((DRBonus) obj).mLocation;
		}
		return false;
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public String getKey() {
		StringBuffer buffer = new StringBuffer();

		buffer.append(Armor.DR_PREFIX);
		buffer.append(mLocation.name());
		return buffer.toString();
	}

	@Override
	public Feature cloneFeature() {
		return new DRBonus(this);
	}

	@Override
	protected void loadSelf(XMLReader reader) throws IOException {
		if (TAG_LOCATION.equals(reader.getName())) {
			setLocation(Enums.extract(reader.readText(), HitLocation.values(), HitLocation.TORSO));
		} else {
			super.loadSelf(reader);
		}
	}

	/**
	 * Saves the bonus.
	 * 
	 * @param out The XML writer to use.
	 */
	@Override
	public void save(XMLWriter out) {
		out.startSimpleTagEOL(TAG_ROOT);
		out.simpleTag(TAG_LOCATION, mLocation.name().toLowerCase());
		saveBase(out);
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return The location protected by the DR. */
	public HitLocation getLocation() {
		return mLocation;
	}

	/** @param location The location. */
	public void setLocation(HitLocation location) {
		mLocation = location;
	}
}
