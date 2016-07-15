/*
 * Copyright (c) 1998-2016 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.units.WeightValue;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** Describes a contained weight reduction. */
public class ContainedWeightReduction implements Feature {
	/** The XML tag. */
	public static final String	TAG_ROOT	= "contained_weight_reduction";	//$NON-NLS-1$
	private Object				mValue;

	/** Creates a new contained weight reduction. */
	public ContainedWeightReduction() {
		mValue = Integer.valueOf(0);
	}

	/**
	 * Creates a clone of the specified contained weight reduction.
	 *
	 * @param other The bonus to clone.
	 */
	public ContainedWeightReduction(ContainedWeightReduction other) {
		if (other.mValue instanceof WeightValue) {
			mValue = new WeightValue((WeightValue) other.mValue);
		} else {
			mValue = other.mValue;
		}
	}

	/**
	 * Loads a {@link ContainedWeightReduction}.
	 *
	 * @param reader The XML reader to use.
	 */
	public ContainedWeightReduction(XMLReader reader) throws IOException {
		this();
		load(reader);
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof ContainedWeightReduction) {
			ContainedWeightReduction cr = (ContainedWeightReduction) obj;
			return mValue.equals(cr.mValue);
		}
		return false;
	}

	@Override
	public int hashCode() {
		return super.hashCode();
	}

	/** @return The weight reduction value object. */
	public Object getValue() {
		return mValue;
	}

	/**
	 * Sets the value of the weight reduction.
	 *
	 * @param value An Integer value or a WeightValue. Other object types will be turned into a
	 *            value of 0.
	 */
	public void setValue(Object value) {
		if (value instanceof Integer) {
			mValue = value;
		} else if (value instanceof WeightValue) {
			mValue = new WeightValue((WeightValue) value);
		} else {
			mValue = Integer.valueOf(0);
		}
	}

	/**
	 * @return <code>true</code> if the reduction is a percentage of the weight, <code>false</code>
	 *         if it is an absolute number.
	 */
	public boolean isPercentage() {
		return mValue instanceof Integer;
	}

	/**
	 * @return The percentage the weight should be reduced by. Will return 0 if
	 *         {@link #isPercentage()} returns false.
	 */
	public int getPercentageReduction() {
		if (isPercentage()) {
			return ((Integer) mValue).intValue();
		}
		return 0;
	}

	/**
	 * Sets a percentage weight reduction to use.
	 *
	 * @param reduction The amount to reduce the contained weight by.
	 */
	public void setPercentageReduction(int reduction) {
		mValue = Integer.valueOf(reduction);
	}

	/**
	 * @return The absolute weight reduction. Will return a weight value of 0 if
	 *         {@link #isPercentage()} returns true.
	 */
	public WeightValue getAbsoluteReduction() {
		if (isPercentage()) {
			return new WeightValue(0, SheetPreferences.getWeightUnits());
		}
		return (WeightValue) mValue;
	}

	/**
	 * Sets an absolute weight reduction to use.
	 *
	 * @param reduction The amount to reduce the contained weight by.
	 */
	public void setAbsoluteWeightReduction(WeightValue reduction) {
		mValue = new WeightValue(reduction);
	}

	/**
	 * Loads a contained weight reduction.
	 *
	 * @param reader The XML reader to use.
	 */
	protected void load(XMLReader reader) throws IOException {
		String value = reader.readText().trim();
		if (value.endsWith("%")) { //$NON-NLS-1$
			mValue = Integer.valueOf(Numbers.extractInteger(value.substring(0, value.length() - 1), 0, false));
		} else {
			mValue = WeightValue.extract(value, false);
		}
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public String getKey() {
		return Equipment.ID_EXTENDED_WEIGHT;
	}

	@Override
	public Feature cloneFeature() {
		return new ContainedWeightReduction(this);
	}

	@Override
	public void save(XMLWriter out) {
		String text = ""; //$NON-NLS-1$
		if (mValue instanceof Integer) {
			int percentage = ((Integer) mValue).intValue();
			if (percentage != 0) {
				text = Integer.toString(percentage) + "%"; //$NON-NLS-1$
			}
		} else if (mValue instanceof WeightValue) {
			WeightValue weight = (WeightValue) mValue;
			if (weight.getValue() != 0) {
				text = weight.toString(false);
			}
		}
		out.simpleTag(TAG_ROOT, text);
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
