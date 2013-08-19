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

package com.trollworks.gcs.model.modifier;

import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.modifiers.CSModifierColumnID;
import com.trollworks.gcs.ui.modifiers.CSModifierEditor;
import com.trollworks.toolkit.collections.TKEnumExtractor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.notification.TKNotifier;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.outline.TKColumn;

import java.awt.image.BufferedImage;
import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** Model for trait modifiers */
public class CMModifier extends CMRow implements Comparable<CMModifier> {
	/** The root tag. */
	public static final String		TAG_MODIFIER		= "modifier";							//$NON-NLS-1$
	/** The tag for the name. */
	protected static final String	TAG_NAME			= "name";								//$NON-NLS-1$
	/** The tag for the base cost. */
	public static final String		TAG_COST			= "cost";								//$NON-NLS-1$
	/** The attribute for the cost type. */
	public static final String		ATTRIBUTE_COST_TYPE	= "type";								//$NON-NLS-1$
	/** The tag for the cost per level. */
	public static final String		TAG_LEVELS			= "levels";							//$NON-NLS-1$
	/** The tag for how the cost is affected. */
	public static final String		TAG_AFFECTS			= "affects";							//$NON-NLS-1$
	/** The tag for the page reference. */
	protected static final String	TAG_REFERENCE		= "reference";							//$NON-NLS-1$
	/** The attribute for whether it is enabled. */
	protected static final String	ATTRIBUTE_ENABLED	= "enabled";							//$NON-NLS-1$
	/** The prefix for notifications. */
	public static final String		MODIFIER_PREFIX		= TAG_MODIFIER + TKNotifier.SEPARATOR;
	/** The ID for name change notification. */
	public static final String		ID_NAME				= MODIFIER_PREFIX + TAG_NAME;
	/** The ID for enabled change notification. */
	public static final String		ID_ENABLED			= MODIFIER_PREFIX + ATTRIBUTE_ENABLED;
	/** The ID for cost change notification. */
	public static final String		ID_COST_MODIFIER	= MODIFIER_PREFIX + TAG_COST;
	/** The ID for cost affect change notification. */
	public static final String		ID_AFFECTS			= MODIFIER_PREFIX + TAG_AFFECTS;
	/** The ID for page reference change notification. */
	public static final String		ID_REFERENCE		= MODIFIER_PREFIX + TAG_REFERENCE;
	/** The ID for list changed change notification. */
	public static final String		ID_LIST_CHANGED		= MODIFIER_PREFIX + "ListChanged";		//$NON-NLS-1$
	private String					mName;
	private String					mReference;
	private CMCostType				mCostType;
	private int						mCost;
	private double					mCostMultiplier;
	private int						mLevels;
	private CMAffects				mAffects;
	private boolean					mEnabled;
	private boolean					mReadOnly;

	/**
	 * Creates a new {@link CMModifier}.
	 * 
	 * @param file The {@link CMDataFile} to use.
	 * @param other Another {@link CMModifier} to clone.
	 */
	public CMModifier(CMDataFile file, CMModifier other) {
		super(file, other);
		mName = other.mName;
		mReference = other.mReference;
		mCostType = other.mCostType;
		mCost = other.mCost;
		mCostMultiplier = other.mCostMultiplier;
		mLevels = other.mLevels;
		mAffects = other.mAffects;
		mEnabled = other.mEnabled;
	}

	/**
	 * Creates a new {@link CMModifier}.
	 * 
	 * @param file The {@link CMDataFile} to use.
	 * @param reader The {@link TKXMLReader} to use.
	 * @throws IOException
	 */
	public CMModifier(CMDataFile file, TKXMLReader reader) throws IOException {
		super(file, false);
		load(reader, false);
	}

	/**
	 * Creates a new {@link CMModifier}.
	 * 
	 * @param file The {@link CMDataFile} to use.
	 */
	public CMModifier(CMDataFile file) {
		super(file, false);
		mName = Msgs.DEFAULT_NAME;
		mReference = ""; //$NON-NLS-1$
		mCostType = CMCostType.PERCENTAGE;
		mCost = 0;
		mCostMultiplier = 1.0;
		mLevels = 0;
		mAffects = CMAffects.TOTAL;
		mEnabled = true;
	}

	/** @return The enabled. */
	public boolean isEnabled() {
		return mEnabled;
	}

	/**
	 * @param enabled The value to set for enabled.
	 * @return <code>true</code> if enabled has changed.
	 */
	public boolean setEnabled(boolean enabled) {
		if (mEnabled != enabled) {
			mEnabled = enabled;
			notifySingle(ID_ENABLED);
			return true;
		}
		return false;
	}

	/** @return The page reference. */
	public String getReference() {
		return mReference;
	}

	/**
	 * @param reference The new page reference.
	 * @return <code>true</code> if page reference has changed.
	 */
	public boolean setReference(String reference) {
		if (!mReference.equals(reference)) {
			mReference = reference;
			notifySingle(ID_REFERENCE);
			return true;
		}
		return false;
	}

	/** @return Whether this {@link CMModifier} has been marked as "read-only". */
	public boolean isReadOnly() {
		return mReadOnly;
	}

	/** @param readOnly Whether this {@link CMModifier} has been marked as "read-only". */
	public void setReadOnly(boolean readOnly) {
		mReadOnly = readOnly;
	}

	@Override public String getModifierNotes() {
		return mReadOnly ? Msgs.READ_ONLY : super.getModifierNotes();
	}

	/** @return An exact clone of this modifier. */
	public CMModifier cloneModifier() {
		return new CMModifier(mDataFile, this);
	}

	/** @return The total cost modifier. */
	public int getCostModifier() {
		return mLevels > 0 ? mCost * mLevels : mCost;
	}

	/** @return The costType. */
	public CMCostType getCostType() {
		return mCostType;
	}

	/**
	 * @param costType The value to set for costType.
	 * @return Whether it was changed.
	 */
	public boolean setCostType(CMCostType costType) {
		if (costType != mCostType) {
			mCostType = costType;
			notifySingle(ID_COST_MODIFIER);
			return true;
		}
		return false;
	}

	/** @return The cost. */
	public int getCost() {
		return mCost;
	}

	/**
	 * @param cost The value to set for cost modifier.
	 * @return Whether it was changed.
	 */
	public boolean setCost(int cost) {
		if (mCost != cost) {
			mCost = cost;
			notifySingle(ID_COST_MODIFIER);
			return true;
		}
		return false;
	}

	/** @return The total cost multiplier. */
	public double getCostMultiplier() {
		return mCostMultiplier;
	}

	/**
	 * @param multiplier The value to set for the cost multiplier.
	 * @return Whether it was changed.
	 */
	public boolean setCostMultiplier(double multiplier) {
		if (mCostMultiplier != multiplier) {
			mCostMultiplier = multiplier;
			notifySingle(ID_COST_MODIFIER);
			return true;
		}
		return false;
	}

	/** @return The levels. */
	public int getLevels() {
		return mLevels;
	}

	/**
	 * @param levels The value to set for cost modifier.
	 * @return Whether it was changed.
	 */
	public boolean setLevels(int levels) {
		if (levels < 0) {
			levels = 0;
		}
		if (mLevels != levels) {
			mLevels = levels;
			notifySingle(ID_COST_MODIFIER);
			return true;
		}
		return false;
	}

	/** @return <code>true</code> if this {@link CMModifier} has levels. */
	public boolean hasLevels() {
		return mCostType == CMCostType.PERCENTAGE && mLevels > 0;
	}

	@Override public boolean contains(String text, boolean lowerCaseOnly) {
		return getName().toLowerCase().indexOf(text) != -1;
	}

	@Override public CSRowEditor<CMModifier> createEditor() {
		return new CSModifierEditor(this);
	}

	@Override public BufferedImage getImage(boolean large) {
		return null;
	}

	@Override public String getListChangedID() {
		return ID_LIST_CHANGED;
	}

	@Override public String getLocalizedName() {
		return Msgs.DEFAULT_NAME;
	}

	@Override public String getRowType() {
		return Msgs.MODIFIER_TYPE;
	}

	@Override public String getXMLTagName() {
		return TAG_MODIFIER;
	}

	@Override protected void loadAttributes(TKXMLReader reader, boolean forUndo) {
		super.loadAttributes(reader, forUndo);
		mEnabled = !reader.hasAttribute(ATTRIBUTE_ENABLED) || reader.isAttributeSet(ATTRIBUTE_ENABLED);
	}

	@Override protected void loadSubElement(TKXMLReader reader, boolean forUndo) throws IOException {
		String name = reader.getName();

		if (TAG_NAME.equals(name)) {
			mName = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
		} else if (TAG_REFERENCE.equals(name)) {
			mReference = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
		} else if (TAG_COST.equals(name)) {
			mCostType = (CMCostType) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_COST_TYPE), CMCostType.values(), CMCostType.PERCENTAGE);
			if (mCostType == CMCostType.MULTIPLIER) {
				mCostMultiplier = reader.readDouble(1.0);
			} else {
				mCost = reader.readInteger(0);
			}
		} else if (TAG_LEVELS.equals(name)) {
			mLevels = reader.readInteger(0);
		} else if (TAG_AFFECTS.equals(name)) {
			mAffects = (CMAffects) TKEnumExtractor.extract(reader.readText(), CMAffects.values(), CMAffects.TOTAL);
		} else {
			super.loadSubElement(reader, forUndo);
		}
	}

	@Override protected void prepareForLoad(boolean forUndo) {
		super.prepareForLoad(forUndo);
		mName = Msgs.DEFAULT_NAME;
		mCostType = CMCostType.PERCENTAGE;
		mCost = 0;
		mCostMultiplier = 1.0;
		mLevels = 0;
		mAffects = CMAffects.TOTAL;
		mReference = ""; //$NON-NLS-1$
		mEnabled = true;
	}

	@Override protected void saveAttributes(TKXMLWriter out, boolean forUndo) {
		super.saveAttributes(out, forUndo);
		if (!mEnabled) {
			out.writeAttribute(ATTRIBUTE_ENABLED, mEnabled);
		}
	}

	@Override protected void saveSelf(TKXMLWriter out, boolean forUndo) {
		out.simpleTag(TAG_NAME, mName);
		if (mCostType == CMCostType.MULTIPLIER) {
			out.simpleTagWithAttribute(TAG_COST, mCostMultiplier, ATTRIBUTE_COST_TYPE, mCostType.name().toLowerCase());
		} else {
			out.simpleTagWithAttribute(TAG_COST, mCost, ATTRIBUTE_COST_TYPE, mCostType.name().toLowerCase());
		}
		out.simpleTagNotZero(TAG_LEVELS, mLevels);
		if (mCostType != CMCostType.MULTIPLIER) {
			out.simpleTag(TAG_AFFECTS, mAffects.name().toLowerCase());
		}
		out.simpleTagNotEmpty(TAG_REFERENCE, mReference);
	}

	@Override public Object getData(TKColumn column) {
		return CSModifierColumnID.values()[column.getID()].getData(this);
	}

	@Override public String getDataAsText(TKColumn column) {
		return CSModifierColumnID.values()[column.getID()].getDataAsText(this);
	}

	@Override public String toString() {
		StringBuilder builder = new StringBuilder();

		builder.append(getName());
		if (hasLevels()) {
			builder.append(' ');
			builder.append(getLevels());
		}
		return builder.toString();
	}

	/** @return A full description of this {@link CMModifier}. */
	public String getFullDescription() {
		StringBuilder builder = new StringBuilder();
		String modNote = getNotes();

		builder.append(toString());
		if (modNote.length() > 0) {
			builder.append(" ("); //$NON-NLS-1$
			builder.append(modNote);
			builder.append(')');
		}
		builder.append(", "); //$NON-NLS-1$
		builder.append(getCostDescription());
		return builder.toString();
	}

	/** @return The formatted cost. */
	public String getCostDescription() {
		StringBuilder builder = new StringBuilder();
		CMCostType costType = getCostType();

		switch (costType) {
			case PERCENTAGE:
			case POINTS:
				builder.append(TKNumberUtils.format(getCostModifier(), true));
				if (costType == CMCostType.PERCENTAGE) {
					builder.append('%');
				}
				String desc = mAffects.getShortTitle();
				if (desc.length() > 0) {
					builder.append(' ');
					builder.append(desc);
				}
				break;
			case MULTIPLIER:
				builder.append('x');
				builder.append(TKNumberUtils.format(getCostMultiplier()));
				break;
		}
		return builder.toString();
	}

	/** @return The {@link CMAffects} setting. */
	public CMAffects getAffects() {
		return mAffects;
	}

	/**
	 * @param affects The new {@link CMAffects} setting.
	 * @return <code>true</code> if the setting changed.
	 */
	public boolean setAffects(CMAffects affects) {
		if (affects != mAffects) {
			mAffects = affects;
			notifySingle(ID_AFFECTS);
			return true;
		}
		return false;
	}

	/** @return The name. */
	public String getName() {
		return mName;
	}

	/**
	 * @param name The value to set for name.
	 * @return <code>true</code> if name has changed
	 */
	public boolean setName(String name) {
		if (!mName.equals(name)) {
			mName = name;
			notifySingle(ID_NAME);
			return true;
		}
		return false;
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		if (isEnabled()) {
			super.fillWithNameableKeys(set);
			extractNameables(set, mName);
		}
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		if (isEnabled()) {
			super.applyNameableKeys(map);
			mName = nameNameables(map, mName);
		}
	}

	public int compareTo(CMModifier other) {
		if (this == other) {
			return 0;
		}
		int result = mName.compareTo(other.mName);
		if (result == 0) {
			result = getNotes().compareTo(other.getNotes());
		}
		return result;
	}
}
