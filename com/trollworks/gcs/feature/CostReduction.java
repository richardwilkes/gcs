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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.ttk.collections.Enums;
import com.trollworks.ttk.xml.XMLNodeType;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** Describes a cost reduction. */
public class CostReduction implements Feature {
	/** The possible {@link BonusAttributeType}s that can be affected. */
	public static final BonusAttributeType[]	TYPES			= { BonusAttributeType.ST, BonusAttributeType.DX, BonusAttributeType.IQ, BonusAttributeType.HT };
	/** The XML tag. */
	public static final String					TAG_ROOT		= "cost_reduction";																				//$NON-NLS-1$
	private static final String					TAG_ATTRIBUTE	= "attribute";																						//$NON-NLS-1$
	private static final String					TAG_PERCENTAGE	= "percentage";																					//$NON-NLS-1$
	private BonusAttributeType					mAttribute;
	private int									mPercentage;

	/** Creates a new cost reduction. */
	public CostReduction() {
		mAttribute = BonusAttributeType.ST;
		mPercentage = 40;
	}

	/**
	 * Creates a clone of the specified cost reduction.
	 * 
	 * @param other The bonus to clone.
	 */
	public CostReduction(CostReduction other) {
		mAttribute = other.mAttribute;
		mPercentage = other.mPercentage;
	}

	/**
	 * Loads a {@link CostReduction}.
	 * 
	 * @param reader The XML reader to use.
	 */
	public CostReduction(XMLReader reader) throws IOException {
		this();
		load(reader);
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof CostReduction) {
			CostReduction cr = (CostReduction) obj;
			return mPercentage == cr.mPercentage && mAttribute == cr.mAttribute;
		}
		return false;
	}

	@Override
	public int hashCode() {
		return super.hashCode();
	}

	/** @return The percentage to use. */
	public int getPercentage() {
		return mPercentage;
	}

	/** @param percentage The percentage to use. */
	public void setPercentage(int percentage) {
		mPercentage = percentage;
	}

	/** @return The attribute this cost reduction applies to. */
	public BonusAttributeType getAttribute() {
		return mAttribute;
	}

	/** @param attribute The attribute. */
	public void setAttribute(BonusAttributeType attribute) {
		mAttribute = attribute;
	}

	/**
	 * Loads a cost reduction.
	 * 
	 * @param reader The XML reader to use.
	 */
	protected void load(XMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String name = reader.getName();

				if (TAG_ATTRIBUTE.equals(name)) {
					setAttribute(Enums.extract(reader.readText(), TYPES, BonusAttributeType.ST));
				} else if (TAG_PERCENTAGE.equals(name)) {
					setPercentage(reader.readInteger(0));
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public String getKey() {
		return GURPSCharacter.ATTRIBUTES_PREFIX + mAttribute.name();
	}

	@Override
	public Feature cloneFeature() {
		return new CostReduction(this);
	}

	@Override
	public void save(XMLWriter out) {
		out.startSimpleTagEOL(TAG_ROOT);
		out.simpleTag(TAG_ATTRIBUTE, mAttribute.name().toLowerCase());
		out.simpleTag(TAG_PERCENTAGE, mPercentage);
		out.endTagEOL(TAG_ROOT, true);
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		// Nothing to do.
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		// Nothing to do.
	}
}
