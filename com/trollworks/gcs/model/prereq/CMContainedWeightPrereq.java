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
import com.trollworks.gcs.model.criteria.CMDoubleCriteria;
import com.trollworks.gcs.model.criteria.CMNumericCompareType;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.toolkit.collections.TKEnumExtractor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.utility.units.TKWeightUnits;

import java.io.IOException;
import java.text.MessageFormat;

/** An equipment contained weight prerequisite. */
public class CMContainedWeightPrereq extends CMHasPrereq {
	/** The XML tag for this class. */
	public static final String	TAG_ROOT			= "contained_weight_prereq";	//$NON-NLS-1$
	private static final String	ATTRIBUTE_COMPARE	= "compare";					//$NON-NLS-1$
	private static final String	ATTRIBUTE_UNITS		= "units";						//$NON-NLS-1$
	private CMDoubleCriteria	mWeightCompare;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public CMContainedWeightPrereq(CMPrereqList parent) {
		super(parent);
		mWeightCompare = new CMDoubleCriteria(CMNumericCompareType.AT_MOST, 5.0, true);
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML readerment to load from.
	 * @throws IOException
	 */
	public CMContainedWeightPrereq(CMPrereqList parent, TKXMLReader reader) throws IOException {
		this(parent);
		loadHasAttribute(reader);
		mWeightCompare.setType((CMNumericCompareType) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_COMPARE), CMNumericCompareType.values(), CMNumericCompareType.AT_LEAST));
		mWeightCompare.setQualifier(TKWeightUnits.POUNDS.convert((TKWeightUnits) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_UNITS), TKWeightUnits.values(), TKWeightUnits.POUNDS), reader.readDouble(0)));
	}

	/**
	 * Creates a copy of the specified prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param prereq The prerequisite to clone.
	 */
	protected CMContainedWeightPrereq(CMPrereqList parent, CMContainedWeightPrereq prereq) {
		super(parent, prereq);
		mWeightCompare = new CMDoubleCriteria(prereq.mWeightCompare);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMContainedWeightPrereq && super.equals(obj)) {
			return mWeightCompare.equals(((CMContainedWeightPrereq) obj).mWeightCompare);
		}
		return false;
	}

	@Override public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override public CMPrereq clone(CMPrereqList parent) {
		return new CMContainedWeightPrereq(parent, this);
	}

	@Override public void save(TKXMLWriter out) {
		out.startTag(TAG_ROOT);
		saveHasAttribute(out);
		out.writeAttribute(ATTRIBUTE_COMPARE, mWeightCompare.getType().name().toLowerCase());
		out.writeAttribute(ATTRIBUTE_UNITS, TKWeightUnits.POUNDS.toString());
		out.finishTag();
		out.writeEncodedData(Double.toString(mWeightCompare.getQualifier()));
		out.endTagEOL(TAG_ROOT, false);
	}

	/** @return The weight comparison object. */
	public CMDoubleCriteria getWeightCompare() {
		return mWeightCompare;
	}

	@Override public boolean satisfied(CMCharacter character, CMRow exclude, StringBuilder builder, String prefix) {
		boolean satisfied = false;

		if (exclude instanceof CMEquipment) {
			CMEquipment equipment = (CMEquipment) exclude;

			satisfied = !equipment.canHaveChildren() || mWeightCompare.matches(equipment.getExtendedWeight() - equipment.getWeight());
		}
		if (!has()) {
			satisfied = !satisfied;
		}
		if (!satisfied && builder != null) {
			builder.append(MessageFormat.format(Msgs.CONTAINED_WEIGHT, prefix, has() ? Msgs.HAS : Msgs.DOES_NOT_HAVE, mWeightCompare.toString()));
		}
		return satisfied;
	}
}
