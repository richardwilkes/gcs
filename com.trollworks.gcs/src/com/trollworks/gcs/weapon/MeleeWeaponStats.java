/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;

import java.io.IOException;
import java.util.HashSet;
import java.util.StringTokenizer;

/** The stats for a melee weapon. */
public class MeleeWeaponStats extends WeaponStats {
    public static final  String KEY_ROOT  = "melee_weapon";
    private static final String KEY_REACH = "reach";
    private static final String KEY_PARRY = "parry";
    private static final String KEY_BLOCK = "block";

    private String mReach;
    private String mParry;
    private String mBlock;

    /**
     * Creates a new MeleeWeaponStats.
     *
     * @param owner The owning piece of equipment or advantage.
     */
    public MeleeWeaponStats(ListRow owner) {
        super(owner);
    }

    /**
     * Creates a clone of the specified MeleeWeaponStats.
     *
     * @param owner The owning piece of equipment or advantage.
     * @param other The MeleeWeaponStats to clone.
     */
    public MeleeWeaponStats(ListRow owner, MeleeWeaponStats other) {
        super(owner, other);
        mReach = other.mReach;
        mParry = other.mParry;
        mBlock = other.mBlock;
    }

    /**
     * Creates a MeleeWeaponStats.
     *
     * @param owner The owning piece of equipment or advantage.
     * @param m     The {@link JsonMap} to load from.
     */
    public MeleeWeaponStats(ListRow owner, JsonMap m) throws IOException {
        super(owner, m);
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
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    @Override
    protected void loadSelf(JsonMap m) throws IOException {
        super.loadSelf(m);
        mReach = m.getString(KEY_REACH);
        mParry = m.getString(KEY_PARRY);
        mBlock = m.getString(KEY_BLOCK);
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        w.keyValueNot(KEY_REACH, mReach, "");
        w.keyValueNot(KEY_PARRY, mParry, "");
        w.keyValueNot(KEY_BLOCK, mBlock, "");

        // Emit the calculated values for third parties
        w.key("calc");
        w.startMap();
        w.keyValue("level", Math.max(getSkillLevel(), 0));
        w.keyValue("parry", getResolvedParry(null));
        w.keyValue("block", getResolvedBlock(null));
        w.keyValue("damage", getDamage().getResolvedDamage());
        w.endMap();
    }

    /** @return The parry. */
    public String getParry() {
        return mParry;
    }

    /** @return The parry, fully resolved for the user's skills, if possible. */
    public String getResolvedParryNoToolTip() {
        return getResolvedParry(null);
    }

    /** @return The parry, fully resolved for the user's skills, if possible. */
    public String getResolvedParry(StringBuilder toolTip) {
        return getResolvedValue(mParry, "parry", toolTip);
    }

    private String getResolvedValue(String input, String baseDefaultType, StringBuilder toolTip) {
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
                                StringBuilder primaryToolTip   = toolTip != null ? new StringBuilder() : null;
                                StringBuilder secondaryToolTip = null;
                                int           preAdj           = getSkillLevelBaseAdjustment(character, primaryToolTip);
                                int           postAdj          = getSkillLevelPostAdjustment(character, primaryToolTip);
                                int           adj              = 3 + ("parry".equals(baseDefaultType) ? character.getParryBonus() : character.getBlockBonus());
                                int           best             = Integer.MIN_VALUE;
                                for (SkillDefault skillDefault : getDefaults()) {
                                    int level = SkillDefaultType.getSkillLevelFast(character, skillDefault, false, new HashSet<>(), true);
                                    if (level != Integer.MIN_VALUE) {
                                        level += preAdj;
                                        String type = skillDefault.getType();
                                        if (!baseDefaultType.equals(type)) {
                                            level = level / 2 + adj;
                                        }
                                        level += postAdj;
                                        StringBuilder possibleToolTip = null;
                                        if ("skill".equals(type) && "Karate".equals(skillDefault.getName())) {
                                            if (toolTip != null) {
                                                possibleToolTip = new StringBuilder();
                                            }
                                            level += getEncumbrancePenalty(character, possibleToolTip);
                                        }
                                        if (level > best) {
                                            best = level;
                                            secondaryToolTip = possibleToolTip;
                                        }
                                    }
                                }
                                if (best != Integer.MIN_VALUE && toolTip != null) {
                                    if (!primaryToolTip.isEmpty()) {
                                        if (!toolTip.isEmpty()) {
                                            toolTip.append('\n');
                                        }
                                        toolTip.append(primaryToolTip);
                                    }
                                    if (secondaryToolTip != null && !secondaryToolTip.isEmpty()) {
                                        if (!toolTip.isEmpty()) {
                                            toolTip.append('\n');
                                        }
                                        toolTip.append(secondaryToolTip);
                                    }
                                }
                                skillLevel = Math.max(best, 0);
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
            notifyOfChange();
        }
    }

    /** @return The block. */
    public String getBlock() {
        return mBlock;
    }

    /** @return The block, fully resolved for the user's skills, if possible. */
    public String getResolvedBlockNoToolTip() {
        return getResolvedBlock(null);
    }

    /** @return The block, fully resolved for the user's skills, if possible. */
    public String getResolvedBlock(StringBuilder toolTip) {
        return getResolvedValue(mBlock, "block", toolTip);
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
            notifyOfChange();
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
            notifyOfChange();
        }
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof MeleeWeaponStats && super.equals(obj)) {
            MeleeWeaponStats mws = (MeleeWeaponStats) obj;
            return mReach.equals(mws.mReach) && mParry.equals(mws.mParry) &&
                    mBlock.equals(mws.mBlock);
        }
        return false;
    }

    public String getParryToolTip() {
        StringBuilder toolTip = new StringBuilder();
        if (mOwner.getDataFile() instanceof GURPSCharacter) {
            getResolvedParry(toolTip);
        }
        return toolTip.isEmpty() ? SkillLevel.getNoAdditionalModifiers() :
                SkillLevel.getIncludesModifiersFrom() + toolTip;
    }

    public String getBlockToolTip() {
        StringBuilder toolTip = new StringBuilder();
        if (mOwner.getDataFile() instanceof GURPSCharacter) {
            getResolvedBlock(toolTip);
        }
        return toolTip.isEmpty() ? SkillLevel.getNoAdditionalModifiers() :
                SkillLevel.getIncludesModifiersFrom() + toolTip;
    }
}
