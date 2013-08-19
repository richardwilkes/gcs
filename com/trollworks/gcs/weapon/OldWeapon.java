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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.text.TextUtility;
import com.trollworks.ttk.xml.XMLNodeType;
import com.trollworks.ttk.xml.XMLReader;

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
	 * @throws IOException
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

	private String[] extract(XMLReader reader) throws IOException {
		return TextUtility.extractLines(reader.readText().trim(), 0).toArray(new String[0]);
	}

	/**
	 * @param owner The owning row.
	 * @return The weapons associated with the old data.
	 */
	public ArrayList<WeaponStats> getWeapons(ListRow owner) {
		ArrayList<WeaponStats> weapons = new ArrayList<WeaponStats>();
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

	private String get(String[] data, int index) {
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

	private int count(String[] which) {
		return which != null ? which.length : 0;
	}
}
