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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.utility.text.TextUtility;

import java.io.IOException;
import java.util.ArrayList;

/** Helper class to convert old weapon data into the new weapon data. */
public class OldWeapon {
	/** The root XML tag. */
	public static final String	TAG_ROOT			= "weapon";		//$NON-NLS-1$
	private static final String	TAG_DAMAGE			= "damage";		//$NON-NLS-1$
	private static final String	TAG_STRENGTH		= "strength";		//$NON-NLS-1$
	private static final String	TAG_REACH			= "reach";			//$NON-NLS-1$
	private static final String	TAG_PARRY			= "parry";			//$NON-NLS-1$
	private static final String	TAG_ACCURACY		= "accuracy";		//$NON-NLS-1$
	private static final String	TAG_RANGE			= "range";			//$NON-NLS-1$
	private static final String	TAG_RATE_OF_FIRE	= "rate_of_fire";	//$NON-NLS-1$
	private static final String	TAG_SHOTS			= "shots";			//$NON-NLS-1$
	private static final String	TAG_BULK			= "bulk";			//$NON-NLS-1$
	private static final String	TAG_RECOIL			= "recoil";		//$NON-NLS-1$
	private String[]			mDamage;
	private String[]			mStrength;
	private String[]			mAccuracy;
	private String[]			mRange;
	private String[]			mRateOfFire;
	private String[]			mShots;
	private String[]			mBulk;
	private String[]			mRecoil;
	private String[]			mReach;
	private String[]			mParry;

	/**
	 * Creates a new {@link OldWeapon}.
	 * 
	 * @param reader The XML reader to load from.
	 */
	public OldWeapon(XMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String name = reader.getName();

				if (TAG_DAMAGE.equals(name)) {
					mDamage = extract(reader);
				} else if (TAG_STRENGTH.equals(name)) {
					mStrength = extract(reader);
				} else if (TAG_ACCURACY.equals(name)) {
					mAccuracy = extract(reader);
				} else if (TAG_RANGE.equals(name)) {
					mRange = extract(reader);
				} else if (TAG_RATE_OF_FIRE.equals(name)) {
					mRateOfFire = extract(reader);
				} else if (TAG_SHOTS.equals(name)) {
					mShots = extract(reader);
				} else if (TAG_BULK.equals(name)) {
					mBulk = extract(reader);
				} else if (TAG_RECOIL.equals(name)) {
					mRecoil = extract(reader);
				} else if (TAG_REACH.equals(name)) {
					mReach = extract(reader);
				} else if (TAG_PARRY.equals(name)) {
					mParry = extract(reader);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	private static String[] extract(XMLReader reader) throws IOException {
		return TextUtility.extractLines(reader.readText().trim(), 0).toArray(new String[0]);
	}

	/**
	 * @param owner The owning row.
	 * @return The weapons associated with the old data.
	 */
	public ArrayList<WeaponStats> getWeapons(ListRow owner) {
		ArrayList<WeaponStats> weapons = new ArrayList<>();
		int count = count();

		for (int i = 0; i < count; i++) {
			String reach = get(mReach, i);
			String range = get(mRange, i);

			if (reach.length() > 0) {
				MeleeWeaponStats melee = new MeleeWeaponStats(owner);

				melee.setDamage(get(mDamage, i));
				melee.setStrength(get(mStrength, i));
				melee.setParry(get(mParry, i));
				melee.setReach(reach);
				melee.setDefaults(owner.getDefaults());
				weapons.add(melee);
			}
			if (range.length() > 0) {
				RangedWeaponStats ranged = new RangedWeaponStats(owner);

				ranged.setDamage(get(mDamage, i));
				ranged.setStrength(get(mStrength, i));
				ranged.setAccuracy(get(mAccuracy, i));
				ranged.setBulk(get(mBulk, i));
				ranged.setRange(range);
				ranged.setRateOfFire(get(mRateOfFire, i));
				ranged.setRecoil(get(mRecoil, i));
				ranged.setShots(get(mShots, i));
				ranged.setDefaults(owner.getDefaults());
				weapons.add(ranged);
			}
		}

		return weapons;
	}

	private static String get(String[] data, int index) {
		if (data != null && data.length > index) {
			return data[index];
		}
		return ""; //$NON-NLS-1$
	}

	private int count() {
		int max = 0;
		int count = count(mDamage);

		if (count > max) {
			max = count;
		}
		count = count(mStrength);
		if (count > max) {
			max = count;
		}
		count = count(mAccuracy);
		if (count > max) {
			max = count;
		}
		count = count(mRange);
		if (count > max) {
			max = count;
		}
		count = count(mRateOfFire);
		if (count > max) {
			max = count;
		}
		count = count(mShots);
		if (count > max) {
			max = count;
		}
		count = count(mBulk);
		if (count > max) {
			max = count;
		}
		count = count(mRecoil);
		if (count > max) {
			max = count;
		}
		count = count(mReach);
		if (count > max) {
			max = count;
		}
		count = count(mParry);
		return count > max ? count : max;
	}

	private static int count(String[] which) {
		return which != null ? which.length : 0;
	}
}
