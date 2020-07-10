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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.feature.LeveledAmount;
import com.trollworks.gcs.feature.WeaponBonus;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Objects;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/** Holds damage a weapon does, broken down for easier manipulation. */
public class WeaponDamage {
    /** The XML tag used for weapon damage. */
    public static final  String         TAG_ROOT                         = "damage";
    private static final String         ATTR_TYPE                        = "type";
    private static final String         ATTR_ST                          = "st";
    private static final String         ATTR_BASE                        = "base";
    private static final String         ATTR_FRAGMENTATION               = "fragmentation";
    private static final String         ATTR_ARMOR_DIVISOR               = "armor_divisor";
    private static final String         ATTR_FRAGMENTATION_ARMOR_DIVISOR = "fragmentation_armor_divisor";
    private static final String         ATTR_FRAGMENTATION_TYPE          = "fragmentation_type";
    private static final String         ATTR_MODIFIER_PER_DIE            = "modifier_per_die";
    private static final String         DICE_REGEXP_PIECE                = "\\d+[dD]\\d*(\\s*[+-]\\s*\\d+)?(\\s*[xX]\\s*\\d+)?";
    private static final String         DIVISOR_REGEXP_PIECE             = "\\s*(\\(\\s*(?<divisor>\\d+(\\.\\d+)?|∞)\\s*\\))?";
    private static final String         FRAG_REGEXP_PIECE                = "\\s*(\\[\\s*(?<frag>" + DICE_REGEXP_PIECE + ")\\s*(\\(\\s*(?<fragDivisor>\\d+(\\.\\d+)?|∞)\\s*\\))?\\s*(?<fragType>cr|cut)?\\])";
    private static final String         REMAINDER_REGEXP_PIECE           = "(?<remainder>.*)";
    private static final String         DAMAGE_REGEXP_STR                = "^\\s*\\+?\\s*(?<dice>" + DICE_REGEXP_PIECE + ")" + DIVISOR_REGEXP_PIECE + FRAG_REGEXP_PIECE + "?\\s*" + REMAINDER_REGEXP_PIECE + "$";
    private static final String         DAMAGE_ALT_REGEX_STR             = "^\\s*(?<dice>[+-]?\\s*\\d+)?(:(?<perDie>[-+]\\d+))?" + DIVISOR_REGEXP_PIECE + FRAG_REGEXP_PIECE + "?\\s*" + REMAINDER_REGEXP_PIECE + "$";
    private static final String         TRAILING_FRAG_REGEXP_STR         = "^" + REMAINDER_REGEXP_PIECE + FRAG_REGEXP_PIECE + "$";
    private static final Pattern        DAMAGE_REGEXP                    = Pattern.compile(DAMAGE_REGEXP_STR);
    private static final Pattern        DAMAGE_ALT_REGEXP                = Pattern.compile(DAMAGE_ALT_REGEX_STR);
    private static final Pattern        TRAILING_FRAG_REGEXP             = Pattern.compile(TRAILING_FRAG_REGEXP_STR);
    private              WeaponStats    mOwner;
    private              String         mType;
    private              WeaponSTDamage mST;
    private              Dice           mBase;
    private              double         mArmorDivisor;
    private              Dice           mFragmentation;
    private              double         mFragmentationArmorDivisor;
    private              String         mFragmentationType;
    private              int            mModifierPerDie;

    public WeaponDamage(WeaponStats owner) {
        mType = "";
        mOwner = owner;
        mST = WeaponSTDamage.NONE;
        mArmorDivisor = 1;
    }

    public WeaponDamage(XMLReader reader, WeaponStats owner) throws IOException {
        mOwner = owner;
        mType = reader.getAttribute(ATTR_TYPE, "");
        mST = reader.hasAttribute(ATTR_ST) ? Enums.extract(reader.getAttribute(ATTR_ST), WeaponSTDamage.values(), WeaponSTDamage.NONE) : WeaponSTDamage.NONE;
        if (reader.hasAttribute(ATTR_BASE)) {
            mBase = new Dice(reader.getAttribute(ATTR_BASE));
        }
        mArmorDivisor = reader.getAttributeAsDouble(ATTR_ARMOR_DIVISOR, 1);
        mModifierPerDie = reader.getAttributeAsInteger(ATTR_MODIFIER_PER_DIE, 0);
        if (reader.hasAttribute(ATTR_FRAGMENTATION)) {
            mFragmentation = new Dice(reader.getAttribute(ATTR_FRAGMENTATION));
            mFragmentationType = reader.getAttribute(ATTR_FRAGMENTATION_TYPE, "cut");
            mFragmentationArmorDivisor = reader.getAttributeAsDouble(ATTR_FRAGMENTATION_ARMOR_DIVISOR, 1);
        }
        String text = reader.readText().trim();
        if (!text.isEmpty()) {
            // If we find text here, then we have an old damage value that needs to be converted.
            setValuesFromFreeformDamageString(text);
        }
    }

    public WeaponDamage(JsonMap m, WeaponStats owner) throws IOException {
        mOwner = owner;
        mType = m.getString(DataFile.KEY_TYPE);
        mST = Enums.extract(m.getString(ATTR_ST), WeaponSTDamage.values(), WeaponSTDamage.NONE);
        if (m.has(ATTR_BASE)) {
            mBase = new Dice(m.getString(ATTR_BASE));
        }
        mArmorDivisor = m.getDoubleWithDefault(ATTR_ARMOR_DIVISOR, 1);
        mModifierPerDie = m.getInt(ATTR_MODIFIER_PER_DIE);
        if (m.has(ATTR_FRAGMENTATION)) {
            mFragmentation = new Dice(m.getString(ATTR_FRAGMENTATION));
            mFragmentationType = m.getString(ATTR_FRAGMENTATION_TYPE);
            mFragmentationArmorDivisor = m.getDoubleWithDefault(ATTR_FRAGMENTATION_ARMOR_DIVISOR, 1);
        }
    }

    /**
     * Saves the weapon damage.
     *
     * @param w The {@link JsonWriter} to use.
     */
    public void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(DataFile.KEY_TYPE, mType);
        if (mST != WeaponSTDamage.NONE) {
            w.keyValue(ATTR_ST, mST.name().toLowerCase());
        }
        if (mBase != null) {
            String base = mBase.toString();
            if (!"0".equals(base)) {
                w.keyValue(ATTR_BASE, base);
            }
        }
        if (mArmorDivisor != 1) {
            w.keyValue(ATTR_ARMOR_DIVISOR, mArmorDivisor);
        }
        if (mFragmentation != null) {
            String frag = mFragmentation.toString();
            if (!"0".equals(frag)) {
                w.keyValue(ATTR_FRAGMENTATION, frag);
                w.keyValueNot(ATTR_FRAGMENTATION_ARMOR_DIVISOR, mFragmentationArmorDivisor, 1);
                w.keyValue(ATTR_FRAGMENTATION_TYPE, mFragmentationType);
            }
        }
        w.keyValueNot(ATTR_MODIFIER_PER_DIE, mModifierPerDie, 0);
        w.endMap();
    }

    public WeaponDamage clone(WeaponStats owner) {
        WeaponDamage other = new WeaponDamage(owner);
        other.mType = mType;
        other.mST = mST;
        if (mBase != null) {
            other.mBase = mBase.clone();
        }
        other.mArmorDivisor = mArmorDivisor;
        other.mModifierPerDie = mModifierPerDie;
        if (mFragmentation != null) {
            other.mFragmentation = mFragmentation.clone();
            other.mFragmentationType = mFragmentationType;
            other.mFragmentationArmorDivisor = mFragmentationArmorDivisor;
        }
        return other;
    }

    @Override
    public boolean equals(Object obj) {
        return equivalent(obj);
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    public boolean equivalent(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof WeaponDamage) {
            WeaponDamage other = (WeaponDamage) obj;
            if (mType.equals(other.mType) && mST == other.mST && mArmorDivisor == other.mArmorDivisor && mModifierPerDie == other.mModifierPerDie && Objects.equals(mBase, other.mBase)) {
                if (mFragmentation == null) {
                    return other.mFragmentation == null;
                }
                return mFragmentation.equals(other.mFragmentation) && mFragmentationType.equals(other.mFragmentationType) && mFragmentationArmorDivisor == other.mFragmentationArmorDivisor;
            }
        }
        return false;
    }

    public void setValuesFromFreeformDamageString(String text) {
        // Fix up some known bad data file input
        text = text.trim();
        switch (text) {
        case "1d (+1d) burn":
            text = "1d burn";
            break;
        case "Sw cut -1":
            text = "sw-1 cut";
            break;
        case "Thr imp +1":
            text = "thr+1 imp";
            break;
        case "Thr +1":
            text = "thr+1";
            break;
        case "th-1 imp":
            text = "thr-1 imp";
            break;
        case "40mm warhead":
            text = "2d [2d] cr ex";
            break;
        case "3d cr (x5)":
            text = "3dx5 cr";
            break;
        }
        String saved = text;

        // Find and remove first occurrence of 'sw' or 'thr'
        mST = WeaponSTDamage.NONE;
        for (WeaponSTDamage one : WeaponSTDamage.values()) {
            if (one != WeaponSTDamage.NONE) {
                int i = text.indexOf(one.toString());
                if (i != -1) {
                    mST = one;
                    String s = text.substring(i + one.toString().length());
                    text = i > 0 && text.charAt(i - 1) == '+' ? text.substring(0, i - 1) + s : text.substring(0, i) + s;
                    break;
                }
            }
        }

        // Match against the input
        boolean hasPerDie = false;
        Matcher matcher   = DAMAGE_REGEXP.matcher(text);
        boolean matches   = matcher.matches();
        if (!matches) {
            matcher = DAMAGE_ALT_REGEXP.matcher(text);
            matches = matcher.matches();
            hasPerDie = true;
        }
        if (matches) {
            String value = matcher.group("dice");
            mBase = value != null ? new Dice(value.replaceAll(" ", "").toLowerCase()) : null;
            value = matcher.group("divisor");
            mArmorDivisor = value != null ? Numbers.extractDouble(value.replaceAll(" ", ""), 1, false) : 1;
            extractFragInfo(matcher);
            if (hasPerDie) {
                value = matcher.group("perDie");
                mModifierPerDie = value != null ? Numbers.extractInteger(value.trim(), 0, false) : 0;
            }
            mType = matcher.group("remainder");
            if (mType != null) {
                matcher = TRAILING_FRAG_REGEXP.matcher(mType);
                if (matcher.matches()) {
                    extractFragInfo(matcher);
                    mType = matcher.group("remainder");
                }
            }
            mType = mType == null ? "" : mType.trim();
        } else {
            // No match, just copy the saved text into type and clear the other fields
            mType = saved;
            mST = WeaponSTDamage.NONE;
            mBase = null;
            mArmorDivisor = 1;
            mFragmentation = null;
            mFragmentationArmorDivisor = 0;
            mFragmentationType = null;
            mModifierPerDie = 0;
        }
    }

    private void extractFragInfo(Matcher matcher) {
        String value = matcher.group("frag");
        if (value != null) {
            mFragmentation = new Dice(value.replaceAll(" ", "").toLowerCase());
            value = matcher.group("fragDivisor");
            mFragmentationArmorDivisor = value != null ? Numbers.extractDouble(value.replaceAll(" ", ""), 1, false) : 1;
            mFragmentationType = matcher.group("fragType");
            if (mFragmentationType == null) {
                mFragmentationType = "cut";
            }
        } else {
            mFragmentation = null;
            mFragmentationArmorDivisor = 0;
            mFragmentationType = null;
        }
    }

    protected void notifySingle() {
        if (mOwner != null) {
            mOwner.notifySingle(WeaponStats.ID_DAMAGE);
        }
    }

    public WeaponSTDamage getWeaponSTDamage() {
        return mST;
    }

    public void setWeaponSTDamage(WeaponSTDamage stDamage) {
        if (mST != stDamage) {
            mST = stDamage;
            notifySingle();
        }
    }

    public Dice getBase() {
        return mBase;
    }

    public void setBase(Dice base) {
        if (!Objects.equals(mBase, base)) {
            mBase = base;
            notifySingle();
        }
    }

    public double getArmorDivisor() {
        return mArmorDivisor;
    }

    public void setArmorDivisor(double armorDivisor) {
        if (mArmorDivisor != armorDivisor) {
            mArmorDivisor = armorDivisor;
            notifySingle();
        }
    }

    public Dice getFragmentation() {
        return mFragmentation;
    }

    public double getFragmentationArmorDivisor() {
        return mFragmentationArmorDivisor;
    }

    public String getFragmentationType() {
        return mFragmentationType;
    }

    public void setFragmentation(Dice fragmentation, double armorDivisor, String type) {
        if (type == null) {
            type = "cut";
        }
        mFragmentation = fragmentation;
        mFragmentationArmorDivisor = armorDivisor;
        mFragmentationType = type;
        notifySingle();
    }

    public int getModifierPerDie() {
        return mModifierPerDie;
    }

    public void setModifierPerDie(int modifierPerDie) {
        if (mModifierPerDie != modifierPerDie) {
            mModifierPerDie = modifierPerDie;
            notifySingle();
        }
    }

    public String getType() {
        return mType;
    }

    public void setType(String type) {
        if (!mType.equals(type)) {
            mType = type;
            notifySingle();
        }
    }

    /** @return The damage, fully resolved for the user's sw or thr, if possible. */
    public String getResolvedDamage() {
        return getResolvedDamage(null);
    }

    public String getDamageToolTip() {
        StringBuilder toolTip = new StringBuilder();
        getResolvedDamage(toolTip);
        return toolTip.length() > 0 ? I18n.Text("Includes modifiers from") + toolTip : I18n.Text("No additional modifiers");
    }

    /** @return The damage, fully resolved for the user's sw or thr, if possible. */
    public String getResolvedDamage(StringBuilder toolTip) {
        if (mOwner.mOwner != null) {
            DataFile df = mOwner.mOwner.getDataFile();
            if (df instanceof GURPSCharacter) {
                GURPSCharacter   character  = (GURPSCharacter) df;
                Set<WeaponBonus> bonusSet   = new HashSet<>();
                Set<String>      categories = mOwner.getCategories();
                int              maxST      = mOwner.getMinStrengthValue() * 3;
                int              st         = character.getStrength() + character.getStrikingStrengthBonus();
                Dice             base       = new Dice(0, 0);
                for (SkillDefault one : mOwner.getDefaults()) {
                    if (one.getType().isSkillBased()) {
                        String name           = one.getName();
                        String specialization = one.getSpecialization();
                        bonusSet.addAll(character.getWeaponComparedBonusesFor(Skill.ID_NAME + "*", name, specialization, categories, toolTip));
                        bonusSet.addAll(character.getWeaponComparedBonusesFor(Skill.ID_NAME + "/" + name, name, specialization, categories, toolTip));
                    }
                }
                String nameQualifier  = mOwner.toString();
                String usageQualifier = mOwner.getUsage();
                bonusSet.addAll(character.getNamedWeaponBonusesFor(WeaponBonus.WEAPON_NAMED_ID_PREFIX + "*", nameQualifier, usageQualifier, categories, toolTip));
                bonusSet.addAll(character.getNamedWeaponBonusesFor(WeaponBonus.WEAPON_NAMED_ID_PREFIX + "/" + nameQualifier, nameQualifier, usageQualifier, categories, toolTip));
                List<WeaponBonus> bonuses = new ArrayList<>(bonusSet);
                for (Feature feature : mOwner.mOwner.getFeatures()) {
                    extractWeaponBonus(feature, bonuses, toolTip);
                }
                if (mOwner.mOwner instanceof Advantage) {
                    for (AdvantageModifier modifier : ((Advantage) mOwner.mOwner).getModifiers()) {
                        if (modifier.isEnabled()) {
                            for (Feature feature : modifier.getFeatures()) {
                                extractWeaponBonus(feature, bonuses, toolTip);
                            }
                        }
                    }
                }
                if (mOwner.mOwner instanceof Equipment) {
                    for (EquipmentModifier modifier : ((Equipment) mOwner.mOwner).getModifiers()) {
                        if (modifier.isEnabled()) {
                            for (Feature feature : modifier.getFeatures()) {
                                extractWeaponBonus(feature, bonuses, toolTip);
                            }
                        }
                    }
                }
                if (maxST > 0 && maxST < st) {
                    st = maxST;
                }
                if (mBase != null) {
                    base = mBase.clone();
                }
                if (mOwner.mOwner instanceof Advantage) {
                    Advantage advantage = (Advantage) mOwner.mOwner;
                    if (advantage.isLeveled()) {
                        base.multiply(advantage.getLevels());
                    }
                }
                switch (mST) {
                case SW:
                    base = addDice(base, character.getSwing(st));
                    break;
                case SW_LEVELED:
                    Dice swing = character.getSwing(st);
                    if (mOwner.mOwner instanceof Advantage) {
                        Advantage advantage = (Advantage) mOwner.mOwner;
                        if (advantage.isLeveled()) {
                            swing.multiply(advantage.getLevels());
                        }
                    }
                    base = addDice(base, swing);
                    break;
                case THR:
                    base = addDice(base, character.getThrust(st));
                    break;
                case THR_LEVELED:
                    Dice thrust = character.getThrust(st);
                    if (mOwner.mOwner instanceof Advantage) {
                        Advantage advantage = (Advantage) mOwner.mOwner;
                        if (advantage.isLeveled()) {
                            thrust.multiply(advantage.getLevels());
                        }
                    }
                    base = addDice(base, thrust);
                    break;
                default:
                    break;
                }
                for (WeaponBonus bonus : bonuses) {
                    LeveledAmount lvlAmt = bonus.getAmount();
                    int           amt    = lvlAmt.getIntegerAmount();
                    if (lvlAmt.isPerLevel()) {
                        base.add(amt * base.getDieCount());
                    } else {
                        base.add(amt);
                    }
                }
                if (mModifierPerDie != 0) {
                    base.add(mModifierPerDie * base.getDieCount());
                }
                boolean       convertModifiersToExtraDice = mOwner.mOwner.getDataFile().useModifyingDicePlusAdds();
                StringBuilder buffer                      = new StringBuilder();
                buffer.append(base.toString(convertModifiersToExtraDice));
                if (mArmorDivisor != 1) {
                    buffer.append("(");
                    buffer.append(Numbers.format(mArmorDivisor));
                    buffer.append(")");
                }
                if (!mType.isBlank()) {
                    buffer.append(" ");
                    buffer.append(mType);
                }
                if (mFragmentation != null) {
                    String frag = mFragmentation.toString(convertModifiersToExtraDice);
                    if (!"0".equals(frag)) {
                        buffer.append(" [");
                        buffer.append(frag);
                        if (mFragmentationArmorDivisor != 1) {
                            buffer.append("(");
                            buffer.append(Numbers.format(mFragmentationArmorDivisor));
                            buffer.append(")");
                        }
                        buffer.append(" ");
                        buffer.append(mFragmentationType);
                        buffer.append("]");
                    }
                }
                return buffer.toString();
            }
        }
        return toString();
    }

    private void extractWeaponBonus(Feature feature, List<WeaponBonus> list, StringBuilder toolTip) {
        if (feature instanceof WeaponBonus) {
            WeaponBonus wb = (WeaponBonus) feature;
            switch (wb.getWeaponSelectionType()) {
            case THIS_WEAPON:
            default:
                list.add(wb);
                wb.addToToolTip(toolTip);
                break;
            case WEAPONS_WITH_NAME:
                if (wb.getNameCriteria().matches(mOwner.toString()) && wb.getSpecializationCriteria().matches(mOwner.getUsage()) && wb.matchesCategories(mOwner.getCategories())) {
                    list.add(wb);
                    wb.addToToolTip(toolTip);
                }
                break;
            case WEAPONS_WITH_REQUIRED_SKILL:
                // Already handled
                break;
            }
        }
    }

    private static Dice addDice(Dice left, Dice right) {
        return new Dice(left.getDieCount() + right.getDieCount(), Math.max(left.getDieSides(), right.getDieSides()), left.getModifier() + right.getModifier(), left.getMultiplier() + right.getMultiplier() - 1);
    }

    @Override
    public String toString() {
        boolean       convertModifiersToExtraDice = mOwner.mOwner.getDataFile().useModifyingDicePlusAdds();
        StringBuilder buffer                      = new StringBuilder();
        if (mST != WeaponSTDamage.NONE) {
            buffer.append(mST);
        }
        if (mBase != null) {
            String base = mBase.toString(convertModifiersToExtraDice);
            if (!"0".equals(base)) {
                if (buffer.length() > 0) {
                    char ch = base.charAt(0);
                    if (ch != '+' && ch != '-') {
                        buffer.append("+");
                    }
                }
                buffer.append(base);
            }
        }
        if (mArmorDivisor != 1) {
            buffer.append("(");
            buffer.append(Numbers.format(mArmorDivisor));
            buffer.append(")");
        }
        if (mModifierPerDie != 0) {
            if (buffer.length() > 0) {
                buffer.append(" ");
            }
            buffer.append("(");
            buffer.append(Numbers.formatWithForcedSign(mModifierPerDie));
            buffer.append(" per die)");
        }
        if (!mType.isBlank()) {
            buffer.append(" ");
            buffer.append(mType);
        }
        if (mFragmentation != null) {
            String frag = mFragmentation.toString(convertModifiersToExtraDice);
            if (!"0".equals(frag)) {
                buffer.append(" [");
                buffer.append(frag);
                if (mFragmentationArmorDivisor != 1) {
                    buffer.append("(");
                    buffer.append(Numbers.format(mFragmentationArmorDivisor));
                    buffer.append(")");
                }
                buffer.append(" ");
                buffer.append(mFragmentationType);
                buffer.append("]");
            }
        }
        return buffer.toString().trim();
    }
}
