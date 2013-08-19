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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.DoubleCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.utility.collections.EnumExtractor;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.widgets.outline.ListRow;

import java.io.IOException;
import java.text.MessageFormat;

/** An equipment contained weight prerequisite. */
public class ContainedWeightPrereq extends HasPrereq {
	private static String		MSG_CONTAINED_WEIGHT;
	/** The XML tag for this class. */
	public static final String	TAG_ROOT			= "contained_weight_prereq";	//$NON-NLS-1$
	private static final String	ATTRIBUTE_COMPARE	= "compare";					//$NON-NLS-1$
	private static final String	ATTRIBUTE_UNITS		= "units";						//$NON-NLS-1$
	private DoubleCriteria	mWeightCompare;

	static {
		LocalizedMessages.initialize(ContainedWeightPrereq.class);
	}

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public ContainedWeightPrereq(PrereqList parent) {
		super(parent);
		mWeightCompare = new DoubleCriteria(NumericCompareType.AT_MOST, 5.0, true);
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML readerment to load from.
	 * @throws IOException
	 */
	public ContainedWeightPrereq(PrereqList parent, XMLReader reader) throws IOException {
		this(parent);
		loadHasAttribute(reader);
		mWeightCompare.setType((NumericCompareType) EnumExtractor.extract(reader.getAttribute(ATTRIBUTE_COMPARE), NumericCompareType.values(), NumericCompareType.AT_LEAST));
		mWeightCompare.setQualifier(WeightUnits.POUNDS.convert((WeightUnits) EnumExtractor.extract(reader.getAttribute(ATTRIBUTE_UNITS), WeightUnits.values(), WeightUnits.POUNDS), reader.readDouble(0)));
	}

	/**
	 * Creates a copy of the specified prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param prereq The prerequisite to clone.
	 */
	protected ContainedWeightPrereq(PrereqList parent, ContainedWeightPrereq prereq) {
		super(parent, prereq);
		mWeightCompare = new DoubleCriteria(prereq.mWeightCompare);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof ContainedWeightPrereq && super.equals(obj)) {
			return mWeightCompare.equals(((ContainedWeightPrereq) obj).mWeightCompare);
		}
		return false;
	}

	@Override public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override public Prereq clone(PrereqList parent) {
		return new ContainedWeightPrereq(parent, this);
	}

	@Override public void save(XMLWriter out) {
		out.startTag(TAG_ROOT);
		saveHasAttribute(out);
		out.writeAttribute(ATTRIBUTE_COMPARE, mWeightCompare.getType().name().toLowerCase());
		out.writeAttribute(ATTRIBUTE_UNITS, WeightUnits.POUNDS.toString());
		out.finishTag();
		out.writeEncodedData(Double.toString(mWeightCompare.getQualifier()));
		out.endTagEOL(TAG_ROOT, false);
	}

	/** @return The weight comparison object. */
	public DoubleCriteria getWeightCompare() {
		return mWeightCompare;
	}

	@Override public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
		boolean satisfied = false;

		if (exclude instanceof Equipment) {
			Equipment equipment = (Equipment) exclude;

			satisfied = !equipment.canHaveChildren() || mWeightCompare.matches(equipment.getExtendedWeight() - equipment.getWeight());
		}
		if (!has()) {
			satisfied = !satisfied;
		}
		if (!satisfied && builder != null) {
			builder.append(MessageFormat.format(MSG_CONTAINED_WEIGHT, prefix, has() ? MSG_HAS : MSG_DOES_NOT_HAVE, mWeightCompare.toString()));
		}
		return satisfied;
	}
}
