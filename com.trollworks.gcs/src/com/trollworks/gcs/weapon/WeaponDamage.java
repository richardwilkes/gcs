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
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Objects;
import java.util.Set;

/** Holds damage a weapon does, broken down for easier manipulation. */
public class WeaponDamage {
    /** The XML tag used for weapon damage. */
    public static final  String         TAG_ROOT                         = "damage";
    private static final String         ATTR_ST                          = "st";
    private static final String         ATTR_BASE                        = "base";
    private static final String         ATTR_FRAGMENTATION               = "fragmentation";
    private static final String         ATTR_ARMOR_DIVISOR               = "armor_divisor";
    private static final String         ATTR_FRAGMENTATION_ARMOR_DIVISOR = "fragmentation_armor_divisor";
    private static final String         ATTR_FRAGMENTATION_TYPE          = "fragmentation_type";
    private static final String         ATTR_MODIFIER_PER_DIE            = "modifier_per_die";
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

    public WeaponDamage(JsonMap m, WeaponStats owner) {
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
        return toolTip.isEmpty() ? I18n.Text("No additional modifiers") : I18n.Text("Includes modifiers from") + toolTip;
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

                // Determine which skill default was used
                int          best        = Integer.MIN_VALUE;
                SkillDefault bestDefault = null;
                for (SkillDefault skillDefault : mOwner.getDefaults()) {
                    SkillDefaultType type = skillDefault.getType();
                    if (type.isSkillBased()) {
                        int level = type.getSkillLevelFast(character, skillDefault, false, new HashSet<>(), true);
                        if (level > best) {
                            best = level;
                            bestDefault = skillDefault;
                        }
                    }
                }

                if (bestDefault != null) {
                    String name           = bestDefault.getName();
                    String specialization = bestDefault.getSpecialization();
                    bonusSet.addAll(character.getWeaponComparedBonusesFor(Skill.ID_NAME + "*", name, specialization, categories, toolTip));
                    bonusSet.addAll(character.getWeaponComparedBonusesFor(Skill.ID_NAME + "/" + name, name, specialization, categories, toolTip));
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
                if (!buffer.isEmpty()) {
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
            if (!buffer.isEmpty()) {
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
