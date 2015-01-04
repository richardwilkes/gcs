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

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.text.Numbers;

import java.io.IOException;
import java.util.HashSet;
import java.util.StringTokenizer;

/** The stats for a melee weapon. */
public class MeleeWeaponStats extends WeaponStats {
	/** The root XML tag. */
	public static final String	TAG_ROOT	= "melee_weapon";		//$NON-NLS-1$
	private static final String	TAG_REACH	= "reach";				//$NON-NLS-1$
	private static final String	TAG_PARRY	= "parry";				//$NON-NLS-1$
	private static final String	TAG_BLOCK	= "block";				//$NON-NLS-1$
	/** The field ID for reach changes. */
	public static final String	ID_REACH	= PREFIX + TAG_REACH;
	/** The field ID for parry changes. */
	public static final String	ID_PARRY	= PREFIX + TAG_PARRY;
	/** The field ID for block changes. */
	public static final String	ID_BLOCK	= PREFIX + TAG_BLOCK;
	private String				mReach;
	private String				mParry;
	private String				mBlock;

	/**
	 * Creates a new {@link MeleeWeaponStats}.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 */
	public MeleeWeaponStats(ListRow owner) {
		super(owner);
	}

	/**
	 * Creates a clone of the specified {@link MeleeWeaponStats}.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 * @param other The {@link MeleeWeaponStats} to clone.
	 */
	public MeleeWeaponStats(ListRow owner, MeleeWeaponStats other) {
		super(owner, other);
		mReach = other.mReach;
		mParry = other.mParry;
		mBlock = other.mBlock;
	}

	/**
	 * Creates a {@link MeleeWeaponStats}.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 * @param reader The reader to load from.
	 */
	public MeleeWeaponStats(ListRow owner, XMLReader reader) throws IOException {
		super(owner, reader);
	}

	@Override
	public WeaponStats clone(ListRow owner) {
		return new MeleeWeaponStats(owner, this);
	}

	@Override
	protected void initialize() {
		mReach = EMPTY;
		mParry = EMPTY;
		mBlock = EMPTY;
	}

	@Override
	protected void loadSelf(XMLReader reader) throws IOException {
		String name = reader.getName();

		if (TAG_REACH.equals(name)) {
			mReach = reader.readText();
		} else if (TAG_PARRY.equals(name)) {
			mParry = reader.readText();
		} else if (TAG_BLOCK.equals(name)) {
			mBlock = reader.readText();
		} else {
			super.loadSelf(reader);
		}
	}

	@Override
	protected String getRootTag() {
		return TAG_ROOT;
	}

	@Override
	protected void saveSelf(XMLWriter out) {
		out.simpleTagNotEmpty(TAG_REACH, mReach);
		out.simpleTagNotEmpty(TAG_PARRY, mParry);
		out.simpleTagNotEmpty(TAG_BLOCK, mBlock);
	}

	/** @return The parry. */
	public String getParry() {
		return mParry;
	}

	/** @return The parry, fully resolved for the user's skills, if possible. */
	public String getResolvedParry() {
		return getResolvedValue(mParry, SkillDefaultType.Parry);
	}

	private String getResolvedValue(String input, SkillDefaultType baseDefaultType) {
		DataFile df = getOwner().getDataFile();
		if (df instanceof GURPSCharacter) {
			GURPSCharacter character = (GURPSCharacter) df;
			StringTokenizer tokenizer = new StringTokenizer(input, "\n\r", true); //$NON-NLS-1$
			StringBuffer buffer = new StringBuffer();
			int skillLevel = Integer.MAX_VALUE;

			while (tokenizer.hasMoreTokens()) {
				String token = tokenizer.nextToken();

				if (!token.equals("\n") && !token.equals("\r")) { //$NON-NLS-1$ //$NON-NLS-2$
					int max = token.length();
					int i = skipSpaces(token, 0);

					if (i < max) {
						char ch = token.charAt(i);
						boolean neg = false;
						int modifier = 0;
						boolean found = false;

						if (ch == '-' || ch == '+') {
							neg = ch == '-';
							if (++i < max) {
								ch = token.charAt(i);
							}
						}
						while (i < max && ch >= '0' && ch <= '9') {
							found = true;
							modifier *= 10;
							modifier += ch - '0';
							if (++i < max) {
								ch = token.charAt(i);
							}
						}

						if (found) {
							String num;

							if (skillLevel == Integer.MAX_VALUE) {
								int best = Integer.MIN_VALUE;
								for (SkillDefault skillDefault : getDefaults()) {
									SkillDefaultType type = skillDefault.getType();
									int level = type.getSkillLevelFast(character, skillDefault, new HashSet<String>());
									if (level != Integer.MIN_VALUE && type != baseDefaultType) {
										level = level / 2 + 3 + (baseDefaultType == SkillDefaultType.Parry ? character.getParryBonus() : character.getBlockBonus());
									}
									if (level > best) {
										best = level;
									}
								}
								skillLevel = best != Integer.MIN_VALUE ? best : 0;
							}
							num = Numbers.format(skillLevel + (neg ? -modifier : modifier));
							if (i < max) {
								buffer.append(num);
								token = token.substring(i);
							} else {
								token = num;
							}
						}
					}
				}
				buffer.append(token);
			}
			return buffer.toString();
		}
		return input;
	}

	/**
	 * Sets the value of parry.
	 * 
	 * @param parry The value to set.
	 */
	public void setParry(String parry) {
		parry = sanitize(parry);
		if (!mParry.equals(parry)) {
			mParry = parry;
			notifySingle(ID_PARRY);
		}
	}

	/** @return The block. */
	public String getBlock() {
		return mBlock;
	}

	/** @return The block, fully resolved for the user's skills, if possible. */
	public String getResolvedBlock() {
		return getResolvedValue(mBlock, SkillDefaultType.Block);
	}

	/**
	 * Sets the value of block.
	 * 
	 * @param block The value to set.
	 */
	public void setBlock(String block) {
		block = sanitize(block);
		if (!mBlock.equals(block)) {
			mBlock = block;
			notifySingle(ID_BLOCK);
		}
	}

	/** @return The reach. */
	public String getReach() {
		return mReach;
	}

	/**
	 * Sets the value of reach.
	 * 
	 * @param reach The value to set.
	 */
	public void setReach(String reach) {
		reach = sanitize(reach);
		if (!mReach.equals(reach)) {
			mReach = reach;
			notifySingle(ID_REACH);
		}
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof MeleeWeaponStats && super.equals(obj)) {
			MeleeWeaponStats mws = (MeleeWeaponStats) obj;
			return mReach.equals(mws.mReach) && mParry.equals(mws.mParry) && mBlock.equals(mws.mBlock);
		}
		return false;
	}
}
