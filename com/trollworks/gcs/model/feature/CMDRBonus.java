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
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;

/** A DR bonus. */
public class CMDRBonus extends CMBonus {
	/** The XML tag. */
	public static final String	TAG_ROOT			= "dr_bonus";				//$NON-NLS-1$
	private static final String	TAG_LOCATION		= "location";				//$NON-NLS-1$
	/** The skull hit location. */
	public static final String	SKULL				= "skull";					//$NON-NLS-1$
	/** The eyes hit location. */
	public static final String	EYES				= "eyes";					//$NON-NLS-1$
	/** The face hit location. */
	public static final String	FACE				= "face";					//$NON-NLS-1$
	/** The neck hit location. */
	public static final String	NECK				= "neck";					//$NON-NLS-1$
	/** The torso hit location. */
	public static final String	TORSO				= "torso";					//$NON-NLS-1$
	/** The vitals hit location. */
	public static final String	VITALS				= "vitals";				//$NON-NLS-1$
	/** The groin hit location. */
	public static final String	GROIN				= "groin";					//$NON-NLS-1$
	/** The arm hit location. */
	public static final String	ARMS				= "arms";					//$NON-NLS-1$
	/** The hand hit location. */
	public static final String	HANDS				= "hands";					//$NON-NLS-1$
	/** The leg hit location. */
	public static final String	LEGS				= "legs";					//$NON-NLS-1$
	/** The foot hit location. */
	public static final String	FEET				= "feet";					//$NON-NLS-1$
	/** The full body hit location. */
	public static final String	FULL_BODY			= "full_body";				//$NON-NLS-1$
	/** The full body except eyes hit location. */
	public static final String	FULL_BODY_NO_EYES	= "full_body_except_eyes";	//$NON-NLS-1$
	private String				mLocation;

	/** Creates a new DR bonus. */
	public CMDRBonus() {
		super(1);
		mLocation = TORSO;
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
			return mLocation.equals(((CMDRBonus) obj).mLocation);
		}
		return false;
	}

	public String getXMLTag() {
		return TAG_ROOT;
	}

	public String getKey() {
		StringBuffer buffer = new StringBuffer();

		buffer.append(CMCharacter.DR_PREFIX);
		buffer.append(mLocation);
		return buffer.toString();
	}

	public CMFeature cloneFeature() {
		return new CMDRBonus(this);
	}

	@Override protected void loadSelf(TKXMLReader reader) throws IOException {
		if (TAG_LOCATION.equals(reader.getName())) {
			setLocation(reader.readText());
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
		out.simpleTag(TAG_LOCATION, mLocation);
		saveBase(out);
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return The location protected by the DR. */
	public String getLocation() {
		return mLocation;
	}

	/** @param location The location. */
	public void setLocation(String location) {
		if (SKULL == location || FACE == location || NECK == location || TORSO == location || GROIN == location || ARMS == location || HANDS == location || LEGS == location || FEET == location || VITALS == location || EYES == location || FULL_BODY == location || FULL_BODY_NO_EYES == location) {
			mLocation = location;
		} else if (SKULL.equals(location)) {
			mLocation = SKULL;
		} else if (EYES.equals(location)) {
			mLocation = EYES;
		} else if (FACE.equals(location)) {
			mLocation = FACE;
		} else if (NECK.equals(location)) {
			mLocation = NECK;
		} else if (TORSO.equals(location)) {
			mLocation = TORSO;
		} else if (VITALS.equals(location)) {
			mLocation = VITALS;
		} else if (FULL_BODY.equals(location)) {
			mLocation = FULL_BODY;
		} else if (FULL_BODY_NO_EYES.equals(location)) {
			mLocation = FULL_BODY_NO_EYES;
		} else if (GROIN.equals(location)) {
			mLocation = GROIN;
		} else if (ARMS.equals(location)) {
			mLocation = ARMS;
		} else if (HANDS.equals(location)) {
			mLocation = HANDS;
		} else if (LEGS.equals(location)) {
			mLocation = LEGS;
		} else if (FEET.equals(location)) {
			mLocation = FEET;
		} else {
			mLocation = TORSO;
		}
	}
}
