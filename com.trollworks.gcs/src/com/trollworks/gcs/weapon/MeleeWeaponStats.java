/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.util.HashSet;
import java.util.StringTokenizer;

/** The stats for a melee weapon. */
public class MeleeWeaponStats extends WeaponStats {
    /** The root XML tag. */
    public static final  String TAG_ROOT  = "melee_weapon";
    private static final String TAG_REACH = "reach";
    private static final String TAG_PARRY = "parry";
    private static final String TAG_BLOCK = "block";
    /** The field ID for reach changes. */
    public static final  String ID_REACH  = PREFIX + TAG_REACH;
    /** The field ID for parry changes. */
    public static final  String ID_PARRY  = PREFIX + TAG_PARRY;
    /** The field ID for block changes. */
    public static final  String ID_BLOCK  = PREFIX + TAG_BLOCK;
    private              String mReach;
    private              String mParry;
    private              String mBlock;

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
     * @param m     The {@link JsonMap} to load from.
     */
    public MeleeWeaponStats(ListRow owner, JsonMap m) throws IOException {
        super(owner, m);
    }

    /**
     * Creates a {@link MeleeWeaponStats}.
     *
     * @param owner  The owning piece of equipment or advantage.
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
        mReach = "";
        mParry = "";
        mBlock = "";
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
    public String getJSONTypeName() {
        return TAG_ROOT;
    }

    @Override
    protected void loadSelf(JsonMap m) throws IOException {
        super.loadSelf(m);
        mReach = m.getString(TAG_REACH);
        mParry = m.getString(TAG_PARRY);
        mBlock = m.getString(TAG_BLOCK);
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        w.keyValueNot(TAG_REACH, mReach, "");
        w.keyValueNot(TAG_PARRY, mParry, "");
        w.keyValueNot(TAG_BLOCK, mBlock, "");
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
            GURPSCharacter  character  = (GURPSCharacter) df;
            StringTokenizer tokenizer  = new StringTokenizer(input, "\n\r", true);
            StringBuilder   buffer     = new StringBuilder();
            int             skillLevel = Integer.MAX_VALUE;

            while (tokenizer.hasMoreTokens()) {
                String token = tokenizer.nextToken();

                if (!"\n".equals(token) && !"\r".equals(token)) {
                    int max = token.length();
                    int i   = skipSpaces(token, 0);

                    if (i < max) {
                        char    ch       = token.charAt(i);
                        boolean neg      = false;
                        int     modifier = 0;
                        boolean found    = false;

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
                                    SkillDefaultType type  = skillDefault.getType();
                                    int              level = type.getSkillLevelFast(character, skillDefault, false, new HashSet<>());
                                    if (level != Integer.MIN_VALUE && type != baseDefaultType) {
                                        level = level / 2 + 3 + (baseDefaultType == SkillDefaultType.Parry ? character.getParryBonus() : character.getBlockBonus());
                                    }
                                    if (level > best) {
                                        best = level;
                                    }
                                }
                                skillLevel = best == Integer.MIN_VALUE ? 0 : best;
                            }
                            num = Numbers.format(skillLevel + (neg ? -modifier : modifier) + (token.contains("F") ? character.getEncumbranceLevel().getEncumbrancePenalty() : 0));
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
