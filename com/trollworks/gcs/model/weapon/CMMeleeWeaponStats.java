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

package com.trollworks.gcs.model.weapon;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMSkillDefault;
import com.trollworks.gcs.model.skill.CMSkillDefaultType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.utility.TKNumberUtils;

import java.io.IOException;
import java.util.HashSet;
import java.util.StringTokenizer;

/** The stats for a melee weapon. */
public class CMMeleeWeaponStats extends CMWeaponStats {
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
	 * Creates a new {@link CMMeleeWeaponStats}.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 */
	public CMMeleeWeaponStats(CMRow owner) {
		super(owner);
	}

	/**
	 * Creates a clone of the specified {@link CMMeleeWeaponStats}.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 * @param other The {@link CMMeleeWeaponStats} to clone.
	 */
	public CMMeleeWeaponStats(CMRow owner, CMMeleeWeaponStats other) {
		super(owner, other);
		mReach = other.mReach;
		mParry = other.mParry;
		mBlock = other.mBlock;
	}

	/**
	 * Creates a {@link CMMeleeWeaponStats}.
	 * 
	 * @param owner The owning piece of equipment or advantage.
	 * @param reader The reader to load from.
	 * @throws IOException
	 */
	public CMMeleeWeaponStats(CMRow owner, TKXMLReader reader) throws IOException {
		super(owner, reader);
	}

	@Override protected void initialize() {
		mReach = EMPTY;
		mParry = EMPTY;
		mBlock = EMPTY;
	}

	@Override protected void loadSelf(TKXMLReader reader) throws IOException {
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

	@Override protected String getRootTag() {
		return TAG_ROOT;
	}

	@Override protected void saveSelf(TKXMLWriter out) {
		out.simpleTag(TAG_REACH, mReach);
		out.simpleTag(TAG_PARRY, mParry);
		out.simpleTag(TAG_BLOCK, mBlock);
	}

	/** @return The parry. */
	public String getParry() {
		return mParry;
	}

	/** @return The parry, fully resolved for the user's skills, if possible. */
	public String getResolvedParry() {
		CMDataFile df = getOwner().getDataFile();

		if (df instanceof CMCharacter) {
			CMCharacter character = (CMCharacter) df;
			StringTokenizer tokenizer = new StringTokenizer(mParry, "\n\r", true); //$NON-NLS-1$
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
								skillLevel = getSkillLevel(character);
							}
							num = TKNumberUtils.format(3 + skillLevel / 2 + (neg ? -modifier : modifier) + character.getParryBonus());
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
		return mParry;
	}

	private int getSkillLevel(CMCharacter character) {
		int best = Integer.MIN_VALUE;

		for (CMSkillDefault skillDefault : getDefaults()) {
			CMSkillDefaultType type = skillDefault.getType();
			int level = type.getSkillLevelFast(character, skillDefault, new HashSet<CMSkill>());

			if (level > best) {
				best = level;
			}
		}

		return best != Integer.MIN_VALUE ? best : 0;
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
		CMDataFile df = getOwner().getDataFile();

		if (df instanceof CMCharacter) {
			CMCharacter character = (CMCharacter) df;
			StringTokenizer tokenizer = new StringTokenizer(mBlock, "\n\r", true); //$NON-NLS-1$
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
								skillLevel = getSkillLevel(character);
							}
							num = TKNumberUtils.format(3 + skillLevel / 2 + (neg ? -modifier : modifier) + character.getBlockBonus());
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
		return mBlock;
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

	@Override public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof CMMeleeWeaponStats) {
			CMMeleeWeaponStats other = (CMMeleeWeaponStats) obj;

			return mReach.equals(other.mReach) && mParry.equals(other.mParry) && mBlock.equals(other.mBlock) && super.equals(obj);
		}
		return false;
	}
}
