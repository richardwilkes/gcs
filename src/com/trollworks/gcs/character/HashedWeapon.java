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

package com.trollworks.gcs.character;

import com.trollworks.gcs.weapon.WeaponStats;

import java.util.HashMap;

/**
 * Provides a wrapper around a {@link WeaponStats} suitable for putting into a {@link HashMap} as a
 * key. Note that this will not work correctly if the {@link WeaponStats} object is changed while
 * the {@link HashedWeapon} is within the {@link HashMap}.
 */
class HashedWeapon {
	private WeaponStats	mWeapon;

	/**
	 * Creates a new hashed weapon.
	 * 
	 * @param weapon The weapon.
	 */
	HashedWeapon(WeaponStats weapon) {
		mWeapon = weapon;
	}

	@Override
	public int hashCode() {
		return mWeapon.getDescription().hashCode();
	}

	@Override
	public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof HashedWeapon) {
			HashedWeapon other = (HashedWeapon) obj;
			return mWeapon.equals(other.mWeapon) && mWeapon.getDescription().equals(other.mWeapon.getDescription());
		}
		return false;
	}
}
