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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.OldWeapon;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Enums;
import com.trollworks.toolkit.utility.units.WeightValue;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

/** A piece of equipment. */
public class Equipment extends ListRow {
	@Localize("Equipment")
	@Localize(locale = "de", value = "Ausrüstung")
	@Localize(locale = "ru", value = "Снаряжение")
	private static String			DEFAULT_NAME;

	static {
		Localization.initialize();
	}

	private static final int		CURRENT_VERSION				= 3;
	private static final String		NEWLINE						= "\n";											//$NON-NLS-1$
	private static final String		SPACE						= " ";												//$NON-NLS-1$
	private static final String		DEFAULT_LEGALITY_CLASS		= "4";												//$NON-NLS-1$
	private static final String		EMPTY						= "";												//$NON-NLS-1$
	/** The extension for Equipment lists. */
	public static final String		OLD_EQUIPMENT_EXTENSION		= "eqp";											//$NON-NLS-1$
	/** The XML tag used for items. */
	public static final String		TAG_EQUIPMENT				= "equipment";										//$NON-NLS-1$
	/** The XML tag used for containers. */
	public static final String		TAG_EQUIPMENT_CONTAINER		= "equipment_container";							//$NON-NLS-1$
	private static final String		ATTRIBUTE_STATE				= "state";											//$NON-NLS-1$
	private static final String		ATTRIBUTE_EQUIPPED			= "equipped";										//$NON-NLS-1$
	private static final String		TAG_QUANTITY				= "quantity";										//$NON-NLS-1$
	private static final String		TAG_DESCRIPTION				= "description";									//$NON-NLS-1$
	private static final String		TAG_TECH_LEVEL				= "tech_level";									//$NON-NLS-1$
	private static final String		TAG_LEGALITY_CLASS			= "legality_class";								//$NON-NLS-1$
	private static final String		TAG_VALUE					= "value";											//$NON-NLS-1$
	private static final String		TAG_WEIGHT					= "weight";										//$NON-NLS-1$
	private static final String		TAG_REFERENCE				= "reference";										//$NON-NLS-1$
	/** The prefix used in front of all IDs for the equipment. */
	public static final String		PREFIX						= GURPSCharacter.CHARACTER_PREFIX + "equipment.";	//$NON-NLS-1$
	/** The field ID for equipped/carried/not carried changes. */
	public static final String		ID_STATE					= PREFIX + "State";								//$NON-NLS-1$
	/** The field ID for quantity changes. */
	public static final String		ID_QUANTITY					= PREFIX + "Quantity";								//$NON-NLS-1$
	/** The field ID for description changes. */
	public static final String		ID_DESCRIPTION				= PREFIX + "Description";							//$NON-NLS-1$
	/** The field ID for tech level changes. */
	public static final String		ID_TECH_LEVEL				= PREFIX + "TechLevel";							//$NON-NLS-1$
	/** The field ID for legality changes. */
	public static final String		ID_LEGALITY_CLASS			= PREFIX + "LegalityClass";						//$NON-NLS-1$
	/** The field ID for value changes. */
	public static final String		ID_VALUE					= PREFIX + "Value";								//$NON-NLS-1$
	/** The field ID for weight changes. */
	public static final String		ID_WEIGHT					= PREFIX + "Weight";								//$NON-NLS-1$
	/** The field ID for extended value changes */
	public static final String		ID_EXTENDED_VALUE			= PREFIX + "ExtendedValue";						//$NON-NLS-1$
	/** The field ID for extended weight changes */
	public static final String		ID_EXTENDED_WEIGHT			= PREFIX + "ExtendedWeight";						//$NON-NLS-1$
	/** The field ID for page reference changes. */
	public static final String		ID_REFERENCE				= PREFIX + "Reference";							//$NON-NLS-1$
	/** The field ID for when the categories change. */
	public static final String		ID_CATEGORY					= PREFIX + "Category";								//$NON-NLS-1$
	/** The field ID for when the row hierarchy changes. */
	public static final String		ID_LIST_CHANGED				= PREFIX + "ListChanged";							//$NON-NLS-1$
	/** The field ID for when the equipment becomes or stops being a weapon. */
	public static final String		ID_WEAPON_STATUS_CHANGED	= PREFIX + "WeaponStatus";							//$NON-NLS-1$
	private EquipmentState			mState;
	private int						mQuantity;
	private String					mDescription;
	private String					mTechLevel;
	private String					mLegalityClass;
	private double					mValue;
	private WeightValue				mWeight;
	private double					mExtendedValue;
	private WeightValue				mExtendedWeight;
	private String					mReference;
	private ArrayList<WeaponStats>	mWeapons;

	/**
	 * Creates a new equipment.
	 *
	 * @param dataFile The data file to associate it with.
	 * @param isContainer Whether or not this row allows children.
	 */
	public Equipment(DataFile dataFile, boolean isContainer) {
		super(dataFile, isContainer);
		mState = EquipmentState.EQUIPPED;
		mQuantity = 1;
		mDescription = DEFAULT_NAME;
		mTechLevel = EMPTY;
		mLegalityClass = DEFAULT_LEGALITY_CLASS;
		mReference = EMPTY;
		mWeight = new WeightValue(0, SheetPreferences.getWeightUnits());
		mExtendedWeight = new WeightValue(mWeight);
		mWeapons = new ArrayList<>();
	}

	/**
	 * Creates a clone of an existing equipment and associates it with the specified data file.
	 *
	 * @param dataFile The data file to associate it with.
	 * @param equipment The equipment to clone.
	 * @param deep Whether or not to clone the children, grandchildren, etc.
	 */
	public Equipment(DataFile dataFile, Equipment equipment, boolean deep) {
		super(dataFile, equipment);
		boolean forSheet = dataFile instanceof GURPSCharacter;
		mState = forSheet ? equipment.mState : EquipmentState.EQUIPPED;
		mQuantity = forSheet ? equipment.mQuantity : 1;
		mDescription = equipment.mDescription;
		mTechLevel = equipment.mTechLevel;
		mLegalityClass = equipment.mLegalityClass;
		mValue = equipment.mValue;
		mWeight = new WeightValue(equipment.mWeight);
		mExtendedValue = mQuantity * mValue;
		mExtendedWeight = new WeightValue(mWeight);
		mExtendedWeight.setValue(mExtendedWeight.getValue() * mQuantity);
		mReference = equipment.mReference;
		mWeapons = new ArrayList<>(equipment.mWeapons.size());
		for (WeaponStats weapon : equipment.mWeapons) {
			if (weapon instanceof MeleeWeaponStats) {
				mWeapons.add(new MeleeWeaponStats(this, (MeleeWeaponStats) weapon));
			} else if (weapon instanceof RangedWeaponStats) {
				mWeapons.add(new RangedWeaponStats(this, (RangedWeaponStats) weapon));
			}
		}
		if (deep) {
			int count = equipment.getChildCount();

			for (int i = 0; i < count; i++) {
				addChild(new Equipment(dataFile, (Equipment) equipment.getChild(i), true));
			}
		}
	}

	/**
	 * Loads an equipment and associates it with the specified data file.
	 *
	 * @param dataFile The data file to associate it with.
	 * @param reader The XML reader to load from.
	 * @param state The {@link LoadState} to use.
	 */
	public Equipment(DataFile dataFile, XMLReader reader, LoadState state) throws IOException {
		this(dataFile, TAG_EQUIPMENT_CONTAINER.equals(reader.getName()));
		load(reader, state);
	}

	@Override
	public boolean isEquivalentTo(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof Equipment && super.isEquivalentTo(obj)) {
			Equipment row = (Equipment) obj;
			if (mQuantity == row.mQuantity && mValue == row.mValue && mWeight.equals(row.mWeight) && mState == row.mState && mDescription.equals(row.mDescription) && mTechLevel.equals(row.mTechLevel) && mLegalityClass.equals(row.mLegalityClass) && mReference.equals(row.mReference)) {
				return mWeapons.equals(row.mWeapons);
			}
		}
		return false;
	}

	@Override
	public String getLocalizedName() {
		return DEFAULT_NAME;
	}

	@Override
	public String getListChangedID() {
		return ID_LIST_CHANGED;
	}

	@Override
	public String getXMLTagName() {
		return canHaveChildren() ? TAG_EQUIPMENT_CONTAINER : TAG_EQUIPMENT;
	}

	@Override
	public int getXMLTagVersion() {
		return CURRENT_VERSION;
	}

	@Override
	public String getRowType() {
		return DEFAULT_NAME;
	}

	@Override
	protected void prepareForLoad(LoadState state) {
		super.prepareForLoad(state);
		mState = EquipmentState.EQUIPPED;
		mQuantity = 1;
		mDescription = DEFAULT_NAME;
		mTechLevel = EMPTY;
		mLegalityClass = DEFAULT_LEGALITY_CLASS;
		mReference = EMPTY;
		mValue = 0.0;
		mWeight.setValue(0.0);
		mWeapons = new ArrayList<>();
	}

	@Override
	protected void loadAttributes(XMLReader reader, LoadState state) {
		super.loadAttributes(reader, state);
		if (mDataFile instanceof GURPSCharacter) {
			if (state.mDataItemVersion == 0) {
				if (state.mDefaultCarried) {
					setState(reader.isAttributeSet(ATTRIBUTE_EQUIPPED) ? EquipmentState.EQUIPPED : EquipmentState.NOT_CARRIED);
				} else {
					setState(EquipmentState.NOT_CARRIED);
				}
			} else {
				setState(Enums.extract(reader.getAttribute(ATTRIBUTE_STATE), EquipmentState.values(), EquipmentState.NOT_CARRIED));
			}
		}
	}

	@Override
	protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
		String name = reader.getName();
		if (TAG_DESCRIPTION.equals(name)) {
			mDescription = reader.readText().replace(NEWLINE, SPACE);
		} else if (TAG_TECH_LEVEL.equals(name)) {
			mTechLevel = reader.readText().replace(NEWLINE, SPACE);
		} else if (TAG_LEGALITY_CLASS.equals(name)) {
			mLegalityClass = reader.readText().replace(NEWLINE, SPACE);
		} else if (TAG_VALUE.equals(name)) {
			mValue = reader.readDouble(0.0);
		} else if (TAG_WEIGHT.equals(name)) {
			mWeight = WeightValue.extract(reader.readText(), false);
		} else if (TAG_REFERENCE.equals(name)) {
			mReference = reader.readText().replace(NEWLINE, SPACE);
		} else if (!state.mForUndo && (TAG_EQUIPMENT.equals(name) || TAG_EQUIPMENT_CONTAINER.equals(name))) {
			addChild(new Equipment(mDataFile, reader, state));
		} else if (MeleeWeaponStats.TAG_ROOT.equals(name)) {
			mWeapons.add(new MeleeWeaponStats(this, reader));
		} else if (RangedWeaponStats.TAG_ROOT.equals(name)) {
			mWeapons.add(new RangedWeaponStats(this, reader));
		} else if (OldWeapon.TAG_ROOT.equals(name)) {
			state.mOldWeapons.put(this, new OldWeapon(reader));
		} else if (!canHaveChildren()) {
			if (TAG_QUANTITY.equals(name)) {
				mQuantity = reader.readInteger(1);
			} else {
				super.loadSubElement(reader, state);
			}
		} else {
			super.loadSubElement(reader, state);
		}
	}

	@Override
	protected void finishedLoading(LoadState state) {
		OldWeapon oldWeapon = state.mOldWeapons.remove(this);
		if (oldWeapon != null) {
			mWeapons.addAll(oldWeapon.getWeapons(this));
		}
		// We no longer have defaults... that was solely for the weapons
		setDefaults(new ArrayList<SkillDefault>());
		updateExtendedValue(false);
		updateExtendedWeight(false);
		super.finishedLoading(state);
	}

	@Override
	protected void saveAttributes(XMLWriter out, boolean forUndo) {
		if (mDataFile instanceof GURPSCharacter) {
			out.writeAttribute(ATTRIBUTE_STATE, Enums.toId(mState));
		}
	}

	@Override
	protected void saveSelf(XMLWriter out, boolean forUndo) {
		if (!canHaveChildren()) {
			out.simpleTag(TAG_QUANTITY, mQuantity);
		}
		out.simpleTagNotEmpty(TAG_DESCRIPTION, mDescription);
		out.simpleTagNotEmpty(TAG_TECH_LEVEL, mTechLevel);
		out.simpleTagNotEmpty(TAG_LEGALITY_CLASS, mLegalityClass);
		out.simpleTag(TAG_VALUE, mValue);
		if (mWeight.getNormalizedValue() != 0) {
			out.simpleTag(TAG_WEIGHT, mWeight.toString(false));
		}
		out.simpleTagNotEmpty(TAG_REFERENCE, mReference);
		for (WeaponStats weapon : mWeapons) {
			weapon.save(out);
		}
	}

	@Override
	public void update() {
		updateExtendedValue(true);
		updateExtendedWeight(true);
	}

	/** @return The quantity. */
	public int getQuantity() {
		return mQuantity;
	}

	/**
	 * @param quantity The quantity to set.
	 * @return Whether it was modified.
	 */
	public boolean setQuantity(int quantity) {
		if (quantity != mQuantity) {
			mQuantity = quantity;
			startNotify();
			notify(ID_QUANTITY, this);
			updateContainingWeights(true);
			updateContainingValues(true);
			endNotify();
			return true;
		}
		return false;
	}

	/** @return The description. */
	public String getDescription() {
		return mDescription;
	}

	/**
	 * @param description The description to set.
	 * @return Whether it was modified.
	 */
	public boolean setDescription(String description) {
		if (!mDescription.equals(description)) {
			mDescription = description;
			notifySingle(ID_DESCRIPTION);
			return true;
		}
		return false;
	}

	/** @return The tech level. */
	public String getTechLevel() {
		return mTechLevel;
	}

	/**
	 * @param techLevel The tech level to set.
	 * @return Whether it was modified.
	 */
	public boolean setTechLevel(String techLevel) {
		if (!mTechLevel.equals(techLevel)) {
			mTechLevel = techLevel;
			notifySingle(ID_TECH_LEVEL);
			return true;
		}
		return false;
	}

	/** @return The legality class. */
	public String getLegalityClass() {
		return mLegalityClass;
	}

	/**
	 * @param legalityClass The legality class to set.
	 * @return Whether it was modified.
	 */
	public boolean setLegalityClass(String legalityClass) {
		if (!mLegalityClass.equals(legalityClass)) {
			mLegalityClass = legalityClass;
			notifySingle(ID_LEGALITY_CLASS);
			return true;
		}
		return false;
	}

	/** @return The value. */
	public double getValue() {
		return mValue;
	}

	/**
	 * @param value The value to set.
	 * @return Whether it was modified.
	 */
	public boolean setValue(double value) {
		if (value != mValue) {
			mValue = value;
			startNotify();
			notify(ID_VALUE, this);
			updateContainingValues(true);
			endNotify();
			return true;
		}
		return false;
	}

	/** @return The extended value. */
	public double getExtendedValue() {
		return mExtendedValue;
	}

	/** @return The weight. */
	public WeightValue getWeight() {
		return mWeight;
	}

	/**
	 * @param weight The weight to set.
	 * @return Whether it was modified.
	 */
	public boolean setWeight(WeightValue weight) {
		if (!mWeight.equals(weight)) {
			mWeight = new WeightValue(weight);
			startNotify();
			notify(ID_WEIGHT, this);
			updateContainingWeights(true);
			endNotify();
			return true;
		}
		return false;
	}

	private boolean updateExtendedWeight(boolean okToNotify) {
		WeightValue saved = mExtendedWeight;
		int count = getChildCount();
		mExtendedWeight = new WeightValue(mWeight.getValue() * mQuantity, mWeight.getUnits());
		for (int i = 0; i < count; i++) {
			mExtendedWeight.add(((Equipment) getChild(i)).mExtendedWeight);
		}
		if (!saved.equals(mExtendedWeight)) {
			if (okToNotify) {
				notify(ID_EXTENDED_WEIGHT, this);
			}
			return true;
		}
		return false;
	}

	private void updateContainingWeights(boolean okToNotify) {
		Row parent = this;
		while (parent != null && parent instanceof Equipment) {
			Equipment parentRow = (Equipment) parent;
			if (parentRow.updateExtendedWeight(okToNotify)) {
				parent = parentRow.getParent();
			} else {
				break;
			}
		}
	}

	private boolean updateExtendedValue(boolean okToNotify) {
		double savedValue = mExtendedValue;
		int count = getChildCount();
		mExtendedValue = mQuantity * mValue;
		for (int i = 0; i < count; i++) {
			Equipment child = (Equipment) getChild(i);
			mExtendedValue += child.mExtendedValue;
		}
		if (savedValue != mExtendedValue) {
			if (okToNotify) {
				notify(ID_EXTENDED_VALUE, this);
			}
			return true;
		}
		return false;
	}

	private void updateContainingValues(boolean okToNotify) {
		Row parent = this;
		while (parent != null && parent instanceof Equipment) {
			Equipment parentRow = (Equipment) parent;
			if (parentRow.updateExtendedValue(okToNotify)) {
				parent = parentRow.getParent();
			} else {
				break;
			}
		}
	}

	/** @return The extended weight. */
	public WeightValue getExtendedWeight() {
		return mExtendedWeight;
	}

	/** @return Whether this item is carried. */
	public boolean isCarried() {
		return mState == EquipmentState.CARRIED || mState == EquipmentState.EQUIPPED;
	}

	/** @return Whether this item is equipped. */
	public boolean isEquipped() {
		return mState == EquipmentState.EQUIPPED;
	}

	/** @return The current {@link EquipmentState}. */
	public EquipmentState getState() {
		return mState;
	}

	/**
	 * @param state The new {@link EquipmentState}.
	 * @return Whether it was changed.
	 */
	public boolean setState(EquipmentState state) {
		if (mState != state) {
			mState = state;
			startNotify();
			notify(ID_STATE, this);
			if (canHaveChildren()) {
				for (Row child : getChildren()) {
					((Equipment) child).setState(state);
				}
			}
			endNotify();
			return true;
		}
		return false;
	}

	/** @return The page reference. */
	public String getReference() {
		return mReference;
	}

	/**
	 * @param reference The page reference to set.
	 * @return Whether it was modified.
	 */
	public boolean setReference(String reference) {
		if (!mReference.equals(reference)) {
			mReference = reference;
			notifySingle(ID_REFERENCE);
			return true;
		}
		return false;
	}

	@Override
	public boolean contains(String text, boolean lowerCaseOnly) {
		if (getDescription().toLowerCase().indexOf(text) != -1) {
			return true;
		}
		return super.contains(text, lowerCaseOnly);
	}

	@Override
	public Object getData(Column column) {
		return EquipmentColumn.values()[column.getID()].getData(this);
	}

	@Override
	public String getDataAsText(Column column) {
		return EquipmentColumn.values()[column.getID()].getDataAsText(this);
	}

	@Override
	public String toString() {
		return getDescription();
	}

	/** @return The weapon list. */
	public List<WeaponStats> getWeapons() {
		return Collections.unmodifiableList(mWeapons);
	}

	/**
	 * @param weapons The weapons to set.
	 * @return Whether it was modified.
	 */
	public boolean setWeapons(List<WeaponStats> weapons) {
		if (!mWeapons.equals(weapons)) {
			mWeapons = new ArrayList<>(weapons);
			for (WeaponStats weapon : mWeapons) {
				weapon.setOwner(this);
			}
			notifySingle(ID_WEAPON_STATUS_CHANGED);
			return true;
		}
		return false;
	}

	@Override
	public StdImage getIcon(boolean large) {
		return GCSImages.getEquipmentIcons().getImage(large ? 64 : 16);
	}

	@Override
	public RowEditor<? extends ListRow> createEditor() {
		return new EquipmentEditor(this);
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		extractNameables(set, mDescription);
		for (WeaponStats weapon : mWeapons) {
			for (SkillDefault one : weapon.getDefaults()) {
				one.fillWithNameableKeys(set);
			}
		}
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mDescription = nameNameables(map, mDescription);
		for (WeaponStats weapon : mWeapons) {
			for (SkillDefault one : weapon.getDefaults()) {
				one.applyNameableKeys(map);
			}
		}
	}

	@Override
	protected String getCategoryID() {
		return ID_CATEGORY;
	}
}
