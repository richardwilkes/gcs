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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.character.Armor;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.text.Enums;

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
		out.simpleTag(TAG_LOCATION, Enums.toId(mLocation));
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
