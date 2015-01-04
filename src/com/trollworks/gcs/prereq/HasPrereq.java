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

package com.trollworks.gcs.prereq;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.Localization;

/** An abstract prerequisite class for whether or not the specific item is present. */
public abstract class HasPrereq extends Prereq {
	@Localize("Has")
	@Localize(locale = "de", value = "Hat")
	@Localize(locale = "ru", value = "Имеет")
	static String					HAS;
	@Localize("Does not have")
	@Localize(locale = "de", value = "Hat nicht")
	@Localize(locale = "ru", value = "Не имеет")
	static String					DOES_NOT_HAVE;

	static {
		Localization.initialize();
	}

	/** The "has" attribute name. */
	protected static final String	ATTRIBUTE_HAS	= "has";	//$NON-NLS-1$
	private boolean					mHas;

	/**
	 * Creates a new prerequisite.
	 *
	 * @param parent The owning prerequisite list, if any.
	 */
	public HasPrereq(PrereqList parent) {
		super(parent);
		mHas = true;
	}

	/**
	 * Creates a copy of the specified prerequisite.
	 *
	 * @param parent The owning prerequisite list, if any.
	 * @param prereq The prerequisite to clone.
	 */
	protected HasPrereq(PrereqList parent, HasPrereq prereq) {
		super(parent);
		mHas = prereq.mHas;
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof HasPrereq) {
			return mHas == ((HasPrereq) obj).mHas;
		}
		return false;
	}

	@Override
	public int hashCode() {
		return super.hashCode();
	}

	/**
	 * Loads the "has" attribute.
	 *
	 * @param reader The XML reader to load from.
	 */
	protected void loadHasAttribute(XMLReader reader) {
		mHas = reader.isAttributeSet(ATTRIBUTE_HAS);
	}

	/**
	 * Writes the "has" attribute to the stream.
	 *
	 * @param out The XML writer to use.
	 */
	protected void saveHasAttribute(XMLWriter out) {
		out.writeAttribute(ATTRIBUTE_HAS, mHas);
	}

	/**
	 * @return <code>true</code> if the specified criteria should exist, <code>false</code> if it
	 *         should not.
	 */
	public boolean has() {
		return mHas;
	}

	/**
	 * @param has <code>true</code> if the specified criteria should exist, <code>false</code> if it
	 *            should not.
	 */
	public void has(boolean has) {
		mHas = has;
	}
}
