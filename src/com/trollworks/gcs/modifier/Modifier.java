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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.notification.Notifier;
import com.trollworks.toolkit.utility.text.Enums;
import com.trollworks.toolkit.utility.text.Numbers;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** Model for trait modifiers */
public class Modifier extends ListRow implements Comparable<Modifier> {
	@Localize("Modifier")
	@Localize(locale = "de", value = "Modifikator")
	@Localize(locale = "ru", value = "Модификатор")
	private static String			DEFAULT_NAME;
	@Localize("Enhancement/Limitation")
	@Localize(locale = "de", value = "Verbesserung / Einschränkung")
	@Localize(locale = "ru", value = "Улучшение/ограничение")
	private static String			MODIFIER_TYPE;
	@Localize("** From container - not modifiable here **")
	@Localize(locale = "de", value = "** Aus dem Container \u2013 hier nicht veränderbar **")
	@Localize(locale = "ru", value = "** Из контейнера – не меняйте здесь **")
	private static String			READ_ONLY;

	static {
		Localization.initialize();
	}

	private static final int		CURRENT_VERSION		= 1;
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
	public static final String		MODIFIER_PREFIX		= TAG_MODIFIER + Notifier.SEPARATOR;
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
	/** The name of the {@link Modifier}. */
	protected String				mName;
	/** The page reference for the {@link Modifier}. */
	protected String				mReference;
	/** The cost type of the {@link Modifier}. */
	protected CostType				mCostType;
	private int						mCost;
	private double					mCostMultiplier;
	private int						mLevels;
	private Affects					mAffects;
	private boolean					mEnabled;
	private boolean					mReadOnly;

	/**
	 * Creates a new {@link Modifier}.
	 *
	 * @param file The {@link DataFile} to use.
	 * @param other Another {@link Modifier} to clone.
	 */
	public Modifier(DataFile file, Modifier other) {
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
	 * Creates a new {@link Modifier}.
	 *
	 * @param file The {@link DataFile} to use.
	 * @param reader The {@link XMLReader} to use.
	 * @param state The {@link LoadState} to use.
	 */
	public Modifier(DataFile file, XMLReader reader, LoadState state) throws IOException {
		super(file, false);
		load(reader, state);
	}

	/**
	 * Creates a new {@link Modifier}.
	 *
	 * @param file The {@link DataFile} to use.
	 */
	public Modifier(DataFile file) {
		super(file, false);
		mName = DEFAULT_NAME;
		mReference = ""; //$NON-NLS-1$
		mCostType = CostType.PERCENTAGE;
		mCost = 0;
		mCostMultiplier = 1.0;
		mLevels = 0;
		mAffects = Affects.TOTAL;
		mEnabled = true;
	}

	@Override
	public boolean isEquivalentTo(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof Modifier && super.isEquivalentTo(obj)) {
			Modifier row = (Modifier) obj;
			return mEnabled == row.mEnabled && mLevels == row.mLevels && mCost == row.mCost && mCostMultiplier == row.mCostMultiplier && mCostType == row.mCostType && mAffects == row.mAffects && mName.equals(row.mName) && mReference.equals(row.mReference);
		}
		return false;
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

	/** @return Whether this {@link Modifier} has been marked as "read-only". */
	public boolean isReadOnly() {
		return mReadOnly;
	}

	/** @param readOnly Whether this {@link Modifier} has been marked as "read-only". */
	public void setReadOnly(boolean readOnly) {
		mReadOnly = readOnly;
	}

	@Override
	public String getModifierNotes() {
		return mReadOnly ? READ_ONLY : super.getModifierNotes();
	}

	/** @return An exact clone of this modifier. */
	public Modifier cloneModifier() {
		return new Modifier(mDataFile, this);
	}

	/** @return The total cost modifier. */
	public int getCostModifier() {
		return mLevels > 0 ? mCost * mLevels : mCost;
	}

	/** @return The costType. */
	public CostType getCostType() {
		return mCostType;
	}

	/**
	 * @param costType The value to set for costType.
	 * @return Whether it was changed.
	 */
	public boolean setCostType(CostType costType) {
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

	/** @return <code>true</code> if this {@link Modifier} has levels. */
	public boolean hasLevels() {
		return mCostType == CostType.PERCENTAGE && mLevels > 0;
	}

	@Override
	public boolean contains(String text, boolean lowerCaseOnly) {
		if (getName().toLowerCase().indexOf(text) != -1) {
			return true;
		}
		return super.contains(text, lowerCaseOnly);
	}

	@Override
	public RowEditor<Modifier> createEditor() {
		return new ModifierEditor(this);
	}

	@Override
	public StdImage getIcon(boolean large) {
		return null;
	}

	@Override
	public String getListChangedID() {
		return ID_LIST_CHANGED;
	}

	@Override
	public String getLocalizedName() {
		return DEFAULT_NAME;
	}

	@Override
	public String getRowType() {
		return MODIFIER_TYPE;
	}

	@Override
	public String getXMLTagName() {
		return TAG_MODIFIER;
	}

	@Override
	public int getXMLTagVersion() {
		return CURRENT_VERSION;
	}

	@Override
	protected void loadAttributes(XMLReader reader, LoadState state) {
		super.loadAttributes(reader, state);
		mEnabled = !reader.hasAttribute(ATTRIBUTE_ENABLED) || reader.isAttributeSet(ATTRIBUTE_ENABLED);
	}

	@Override
	protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
		String name = reader.getName();
		if (TAG_NAME.equals(name)) {
			mName = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
		} else if (TAG_REFERENCE.equals(name)) {
			mReference = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
		} else if (TAG_COST.equals(name)) {
			mCostType = Enums.extract(reader.getAttribute(ATTRIBUTE_COST_TYPE), CostType.values(), CostType.PERCENTAGE);
			if (mCostType == CostType.MULTIPLIER) {
				mCostMultiplier = reader.readDouble(1.0);
			} else {
				mCost = reader.readInteger(0);
			}
		} else if (TAG_LEVELS.equals(name)) {
			mLevels = reader.readInteger(0);
		} else if (TAG_AFFECTS.equals(name)) {
			mAffects = Enums.extract(reader.readText(), Affects.values(), Affects.TOTAL);
		} else {
			super.loadSubElement(reader, state);
		}
	}

	@Override
	protected void prepareForLoad(LoadState state) {
		super.prepareForLoad(state);
		mName = DEFAULT_NAME;
		mCostType = CostType.PERCENTAGE;
		mCost = 0;
		mCostMultiplier = 1.0;
		mLevels = 0;
		mAffects = Affects.TOTAL;
		mReference = ""; //$NON-NLS-1$
		mEnabled = true;
	}

	@Override
	protected void saveAttributes(XMLWriter out, boolean forUndo) {
		super.saveAttributes(out, forUndo);
		if (!mEnabled) {
			out.writeAttribute(ATTRIBUTE_ENABLED, mEnabled);
		}
	}

	@Override
	protected void saveSelf(XMLWriter out, boolean forUndo) {
		out.simpleTag(TAG_NAME, mName);
		if (mCostType == CostType.MULTIPLIER) {
			out.simpleTagWithAttribute(TAG_COST, mCostMultiplier, ATTRIBUTE_COST_TYPE, Enums.toId(mCostType));
		} else {
			out.simpleTagWithAttribute(TAG_COST, mCost, ATTRIBUTE_COST_TYPE, Enums.toId(mCostType));
		}
		out.simpleTagNotZero(TAG_LEVELS, mLevels);
		if (mCostType != CostType.MULTIPLIER) {
			out.simpleTag(TAG_AFFECTS, Enums.toId(mAffects));
		}
		out.simpleTagNotEmpty(TAG_REFERENCE, mReference);
	}

	@Override
	public Object getData(Column column) {
		return ModifierColumnID.values()[column.getID()].getData(this);
	}

	@Override
	public String getDataAsText(Column column) {
		return ModifierColumnID.values()[column.getID()].getDataAsText(this);
	}

	@Override
	public String toString() {
		StringBuilder builder = new StringBuilder();

		builder.append(getName());
		if (hasLevels()) {
			builder.append(' ');
			builder.append(getLevels());
		}
		return builder.toString();
	}

	/** @return A full description of this {@link Modifier}. */
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
		CostType costType = getCostType();

		switch (costType) {
			case PERCENTAGE:
			case POINTS:
			default:
				builder.append(Numbers.formatWithForcedSign(getCostModifier()));
				if (costType == CostType.PERCENTAGE) {
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
				builder.append(Numbers.format(getCostMultiplier()));
				break;
		}
		return builder.toString();
	}

	/** @return The {@link Affects} setting. */
	public Affects getAffects() {
		return mAffects;
	}

	/**
	 * @param affects The new {@link Affects} setting.
	 * @return <code>true</code> if the setting changed.
	 */
	public boolean setAffects(Affects affects) {
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

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		if (isEnabled()) {
			super.fillWithNameableKeys(set);
			extractNameables(set, mName);
		}
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		if (isEnabled()) {
			super.applyNameableKeys(map);
			mName = nameNameables(map, mName);
		}
	}

	@Override
	public int compareTo(Modifier other) {
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
