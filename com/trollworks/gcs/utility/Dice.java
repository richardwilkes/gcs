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

package com.trollworks.gcs.utility;

import java.util.Random;

/** Simulates dice. */
public class Dice implements Cloneable {
	/**
	 * If this system property does not exist, then the <code>toString()</code> method will use
	 * the GURPS style of dice display (i.e. 6-sided dice are assumed and therefore the "6" is
	 * omitted in the description).
	 */
	public static final String	DISPLAY_D6						= "com.trollworks.gcs.model.CMDice.DisplayD6";				//$NON-NLS-1$
	/**
	 * If this system property exists, then the optional method on B269 will be used to modify the
	 * actual die count + modifiers.
	 */
	public static final String	USE_OPTIONAL_GURPS_DICE_ADDS	= "com.trollworks.gcs.model.CMDice.OptionalGURPSDiceAdds";	//$NON-NLS-1$
	private static final Random	RANDOM							= new Random();
	private int					mCount;
	private int					mSides;
	private int					mModifier;
	private int					mMultiplier;
	private int					mAltCount;
	private int					mAltModifier;

	/**
	 * Creates a new 1d6 dice object.
	 */
	public Dice() {
		this(1, 6, 0, 1);
	}

	/**
	 * Creates a new d6 dice object.
	 * 
	 * @param count The number of dice.
	 */
	public Dice(int count) {
		this(count, 6, 0, 1);
	}

	/**
	 * Creates a new d6 dice object.
	 * 
	 * @param count The number of dice.
	 * @param modifier The bonus or penalty to the roll.
	 */
	public Dice(int count, int modifier) {
		this(count, 6, modifier, 1);
	}

	/**
	 * Creates a new d6 dice object.
	 * 
	 * @param count The number of dice.
	 * @param modifier The bonus or penalty to the roll.
	 * @param multiplier A multiplier for the roll.
	 */
	public Dice(int count, int modifier, int multiplier) {
		this(count, 6, modifier, multiplier);
	}

	/**
	 * Creates a new dice object.
	 * 
	 * @param count The number of dice.
	 * @param sides The number of sides on each die.
	 * @param modifier The bonus or penalty to the roll.
	 * @param multiplier A multiplier for the roll.
	 */
	public Dice(int count, int sides, int modifier, int multiplier) {
		mCount = count;
		mSides = sides;
		mModifier = modifier;
		mMultiplier = multiplier;
	}

	@Override public Object clone() {
		try {
			return super.clone();
		} catch (CloneNotSupportedException cnse) {
			return null; // Can't happen.
		}
	}

	/**
	 * Adds a modifier to the dice.
	 * 
	 * @param modifier The modifier to add.
	 */
	public void add(int modifier) {
		mModifier += modifier;
	}

	/** @return The number of dice to roll. */
	public int getDieCount() {
		return mCount;
	}

	/** @return The result of rolling the dice. */
	public int roll() {
		int result = 0;

		updateAlt();

		for (int i = 0; i < mAltCount; i++) {
			result += 1 + RANDOM.nextInt(mSides);
		}
		return (result + mAltModifier) * mMultiplier;
	}

	@Override public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof Dice) {
			Dice other = (Dice) obj;

			return mCount == other.mCount && mSides == other.mSides && mModifier == other.mModifier && mMultiplier == other.mMultiplier;
		}
		return false;
	}

	@Override public String toString() {
		StringBuffer buffer = new StringBuffer();

		updateAlt();

		buffer.append(mAltCount);
		buffer.append('d');
		if (mSides != 6 || System.getProperty(DISPLAY_D6) != null) {
			buffer.append(mSides);
		}
		if (mAltModifier > 0) {
			buffer.append('+');
			buffer.append(mAltModifier);
		} else if (mAltModifier < 0) {
			buffer.append(mAltModifier);
		}
		if (mMultiplier != 1) {
			buffer.append('x');
			buffer.append(mMultiplier);
		}
		return buffer.toString();
	}

	private void updateAlt() {
		mAltCount = mCount;
		mAltModifier = mModifier;

		if (System.getProperty(USE_OPTIONAL_GURPS_DICE_ADDS) != null) {
			while (mAltModifier > 3) {
				if (mAltModifier > 6) {
					mAltModifier -= 7;
					mAltCount += 2;
				} else {
					mAltModifier -= 4;
					mAltCount++;
				}
			}
		}
	}
}
