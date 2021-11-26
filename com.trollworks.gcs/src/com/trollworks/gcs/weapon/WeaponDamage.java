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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.feature.LeveledAmount;
import com.trollworks.gcs.feature.WeaponDamageBonus;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.settings.DamageProgression;
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
import java.util.HashSet;
import java.util.Objects;
import java.util.Set;

/** Holds damage a weapon does, broken down for easier manipulation. */
public class WeaponDamage {
    public static final  String KEY_ROOT                        = "damage";
    private static final String KEY_ST                          = "st";
    private static final String KEY_BASE                        = "base";
    private static final String KEY_FRAGMENTATION               = "fragmentation";
    private static final String KEY_ARMOR_DIVISOR               = "armor_divisor";
    private static final String KEY_FRAGMENTATION_ARMOR_DIVISOR = "fragmentation_armor_divisor";
    private static final String KEY_FRAGMENTATION_TYPE          = "fragmentation_type";
    private static final String KEY_MODIFIER_PER_DIE            = "modifier_per_die";

    private WeaponStats    mOwner;
    private String         mType;
    private WeaponSTDamage mST;
    private Dice           mBase;
    private double         mArmorDivisor;
    private Dice           mFragmentation;
    private double         mFragmentationArmorDivisor;
    private String         mFragmentationType;
    private int            mModifierPerDie;

    public WeaponDamage(WeaponStats owner) {
        mType = "";
        mOwner = owner;
        mST = WeaponSTDamage.NONE;
        mArmorDivisor = 1;
    }

    public WeaponDamage(JsonMap m, WeaponStats owner) {
        mOwner = owner;
        mType = m.getString(DataFile.TYPE);
        mST = Enums.extract(m.getString(KEY_ST), WeaponSTDamage.values(), WeaponSTDamage.NONE);
        if (m.has(KEY_BASE)) {
            mBase = new Dice(m.getString(KEY_BASE));
        }
        mArmorDivisor = m.getDoubleWithDefault(KEY_ARMOR_DIVISOR, 1);
        mModifierPerDie = m.getInt(KEY_MODIFIER_PER_DIE);
        if (m.has(KEY_FRAGMENTATION)) {
            mFragmentation = new Dice(m.getString(KEY_FRAGMENTATION));
            mFragmentationType = m.getString(KEY_FRAGMENTATION_TYPE);
            mFragmentationArmorDivisor = m.getDoubleWithDefault(KEY_FRAGMENTATION_ARMOR_DIVISOR, 1);
        }
    }

    /**
     * Saves the weapon damage.
     *
     * @param w The {@link JsonWriter} to use.
     */
    public void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(DataFile.TYPE, mType);
        if (mST != WeaponSTDamage.NONE) {
            w.keyValue(KEY_ST, mST.name().toLowerCase());
        }
        if (mBase != null) {
            String base = mBase.toString();
            if (!"0".equals(base)) {
                w.keyValue(KEY_BASE, base);
            }
        }
        if (mArmorDivisor != 1) {
            w.keyValue(KEY_ARMOR_DIVISOR, mArmorDivisor);
        }
        if (mFragmentation != null) {
            String frag = mFragmentation.toString();
            if (!"0".equals(frag)) {
                w.keyValue(KEY_FRAGMENTATION, frag);
                w.keyValueNot(KEY_FRAGMENTATION_ARMOR_DIVISOR, mFragmentationArmorDivisor, 1);
                w.keyValue(KEY_FRAGMENTATION_TYPE, mFragmentationType);
            }
        }
        w.keyValueNot(KEY_MODIFIER_PER_DIE, mModifierPerDie, 0);
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
        if (obj instanceof WeaponDamage other) {
            if (mType.equals(other.mType) && mST == other.mST && mArmorDivisor == other.mArmorDivisor && mModifierPerDie == other.mModifierPerDie && Objects.equals(mBase, other.mBase)) {
                if (mFragmentation == null) {
                    return other.mFragmentation == null;
                }
                return mFragmentation.equals(other.mFragmentation) && mFragmentationType.equals(other.mFragmentationType) && mFragmentationArmorDivisor == other.mFragmentationArmorDivisor;
            }
        }
        return false;
    }

    protected void notifyOfChange() {
        if (mOwner != null) {
            mOwner.notifyOfChange();
        }
    }

    public WeaponSTDamage getWeaponSTDamage() {
        return mST;
    }

    public void setWeaponSTDamage(WeaponSTDamage stDamage) {
        if (mST != stDamage) {
            mST = stDamage;
            notifyOfChange();
        }
    }

    public Dice getBase() {
        return mBase;
    }

    public void setBase(Dice base) {
        if (!Objects.equals(mBase, base)) {
            mBase = base;
            notifyOfChange();
        }
    }

    public double getArmorDivisor() {
        return mArmorDivisor;
    }

    public void setArmorDivisor(double armorDivisor) {
        if (mArmorDivisor != armorDivisor) {
            mArmorDivisor = armorDivisor;
            notifyOfChange();
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
        notifyOfChange();
    }

    public int getModifierPerDie() {
        return mModifierPerDie;
    }

    public void setModifierPerDie(int modifierPerDie) {
        if (mModifierPerDie != modifierPerDie) {
            mModifierPerDie = modifierPerDie;
            notifyOfChange();
        }
    }

    public String getType() {
        return mType;
    }

    public void setType(String type) {
        if (!mType.equals(type)) {
            mType = type;
            notifyOfChange();
        }
    }

    /** @return The damage, fully resolved for the user's sw or thr, if possible. */
    public String getResolvedDamage() {
        return getResolvedDamage(null);
    }

    public String getDamageToolTip() {
        StringBuilder toolTip = new StringBuilder();
        getResolvedDamage(toolTip);
        return toolTip.isEmpty() ? I18n.text("No additional modifiers") : I18n.text("Includes modifiers from") + toolTip;
    }

    /** @return The damage, fully resolved for the user's sw or thr, if possible. */
    public String getResolvedDamage(StringBuilder toolTip) {
        if (mOwner.mOwner != null) {
            DataFile df = mOwner.mOwner.getDataFile();
            if (df instanceof GURPSCharacter character) {
                Set<WeaponDamageBonus> bonusSet   = new HashSet<>();
                Set<String>            categories = mOwner.getCategories();
                int                    maxST      = mOwner.getMinStrengthValue() * 3;
                int                    st         = character.getAttributeIntValue("st") + character.getStrikingStrengthBonus();
                Dice                   base       = new Dice(0, 0);

                if (maxST > 0 && maxST < st) {
                    st = maxST;
                }
                if (mBase != null) {
                    base = mBase.clone();
                }
                if (mOwner.mOwner instanceof Advantage advantage) {
                    if (advantage.isLeveled()) {
                        base.multiply(advantage.getLevels());
                    }
                }
                switch (mST) {
                    case SW -> base = addDice(base, character.getSwing(st));
                    case SW_LEVELED -> {
                        Dice swing = character.getSwing(st);
                        if (mOwner.mOwner instanceof Advantage advantage) {
                            if (advantage.isLeveled()) {
                                swing.multiply(advantage.getLevels());
                            }
                        }
                        base = addDice(base, swing);
                    }
                    case THR -> base = addDice(base, character.getThrust(st));
                    case THR_LEVELED -> {
                        Dice thrust = character.getThrust(st);
                        if (mOwner.mOwner instanceof Advantage advantage) {
                            if (advantage.isLeveled()) {
                                thrust.multiply(advantage.getLevels());
                            }
                        }
                        base = addDice(base, thrust);
                    }
                }
                int dieCount = base.getDieCount();

                // Determine which skill default was used
                int          best        = Integer.MIN_VALUE;
                SkillDefault bestDefault = null;
                for (SkillDefault skillDefault : mOwner.getDefaults()) {
                    if (SkillDefaultType.isSkillBased(skillDefault.getType())) {
                        int level = SkillDefaultType.getSkillLevelFast(character, skillDefault, false, new HashSet<>(), true);
                        if (level > best) {
                            best = level;
                            bestDefault = skillDefault;
                        }
                    }
                }

                if (bestDefault != null) {
                    String name           = bestDefault.getName();
                    String specialization = bestDefault.getSpecialization();
                    bonusSet.addAll(character.getWeaponComparedDamageBonusesFor(Skill.ID_NAME + "*", name, specialization, categories, dieCount, toolTip));
                    bonusSet.addAll(character.getWeaponComparedDamageBonusesFor(Skill.ID_NAME + "/" + name, name, specialization, categories, dieCount, toolTip));
                }
                String nameQualifier  = mOwner.toString();
                String usageQualifier = mOwner.getUsage();
                bonusSet.addAll(character.getNamedWeaponDamageBonusesFor(WeaponDamageBonus.WEAPON_NAMED_ID_PREFIX + "*", nameQualifier, usageQualifier, categories, dieCount, toolTip));
                bonusSet.addAll(character.getNamedWeaponDamageBonusesFor(WeaponDamageBonus.WEAPON_NAMED_ID_PREFIX + "/" + nameQualifier, nameQualifier, usageQualifier, categories, dieCount, toolTip));
                for (Feature feature : mOwner.mOwner.getFeatures()) {
                    extractWeaponDamageBonus(feature, bonusSet, dieCount, toolTip);
                }
                if (mOwner.mOwner instanceof Advantage) {
                    for (AdvantageModifier modifier : ((Advantage) mOwner.mOwner).getModifiers()) {
                        if (modifier.isEnabled()) {
                            for (Feature feature : modifier.getFeatures()) {
                                extractWeaponDamageBonus(feature, bonusSet, dieCount, toolTip);
                            }
                        }
                    }
                }
                if (mOwner.mOwner instanceof Equipment) {
                    for (EquipmentModifier modifier : ((Equipment) mOwner.mOwner).getModifiers()) {
                        if (modifier.isEnabled()) {
                            for (Feature feature : modifier.getFeatures()) {
                                extractWeaponDamageBonus(feature, bonusSet, dieCount, toolTip);
                            }
                        }
                    }
                }
                boolean adjustForPhoenixFlame = mOwner.mOwner.getDataFile().getSheetSettings().getDamageProgression() == DamageProgression.PHOENIX_FLAME_D3 && base.getDieSides() == 3;
                int     percent               = 0;
                for (WeaponDamageBonus bonus : bonusSet) {
                    LeveledAmount lvlAmt = bonus.getAmount();
                    int           amt    = lvlAmt.getIntegerAmount();
                    if (bonus.isPercent()) {
                        percent += amt;
                    } else {
                        if (lvlAmt.isPerLevel()) {
                            amt *= base.getDieCount();
                            if (adjustForPhoenixFlame) {
                                amt /= 2;
                            }
                        }
                        base.add(amt);
                    }
                }
                if (mModifierPerDie != 0) {
                    int amt = mModifierPerDie * base.getDieCount();
                    if (adjustForPhoenixFlame) {
                        amt /= 2;
                    }
                    base.add(amt);
                }
                if (percent != 0) {
                    base = adjustDiceForPercentBonus(base, percent);
                }
                boolean       convertModifiersToExtraDice = mOwner.mOwner.getDataFile().getSheetSettings().useModifyingDicePlusAdds();
                StringBuilder buffer                      = new StringBuilder();
                if (base.getDieCount() != 0 || base.getModifier() != 0) {
                    buffer.append(base.toString(convertModifiersToExtraDice));
                }
                if (mArmorDivisor != 1) {
                    buffer.append("(");
                    buffer.append(Numbers.format(mArmorDivisor));
                    buffer.append(")");
                }
                if (!mType.isBlank()) {
                    if (!buffer.isEmpty()) {
                        buffer.append(" ");
                    }
                    buffer.append(mType);
                }
                if (mFragmentation != null) {
                    String frag = mFragmentation.toString(convertModifiersToExtraDice);
                    if (!"0".equals(frag)) {
                        if (!buffer.isEmpty()) {
                            buffer.append(" ");
                        }
                        buffer.append("[");
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

    private static Dice adjustDiceForPercentBonus(Dice dice, int percent) {
        int    count         = dice.getDieCount();
        int    sides         = dice.getDieSides();
        int    modifier      = dice.getModifier();
        int    multiplier    = dice.getMultiplier();
        double averagePerDie = (sides + 1) / 2.0;
        double average       = averagePerDie * count + modifier;
        modifier = modifier * (100 + percent) / 100;
        if (average < 0) {
            return new Dice(Math.max(count * (100 + percent) / 100, 0), sides, modifier, multiplier);
        }
        average = (average * (100 + percent) / 100.0) - modifier;
        int adjustedDieCount = Math.max((int) Math.floor(average / averagePerDie), 0);
        modifier += (int) Math.round(average - adjustedDieCount * averagePerDie);
        return new Dice(adjustedDieCount, sides, modifier, multiplier);
    }

    private void extractWeaponDamageBonus(Feature feature, Set<WeaponDamageBonus> set, int dieCount, StringBuilder toolTip) {
        if (feature instanceof WeaponDamageBonus wb) {
            LeveledAmount amount = wb.getAmount();
            int           level  = amount.getLevel();
            amount.setLevel(dieCount);
            switch (wb.getWeaponSelectionType()) {
                case THIS_WEAPON:
                default:
                    if (wb.getSpecializationCriteria().matches(mOwner.getUsage())) {
                        if (set.add(wb)) {
                            wb.addToToolTip(toolTip);
                        }
                    }
                    break;
                case WEAPONS_WITH_NAME:
                    if (wb.getNameCriteria().matches(mOwner.toString()) &&
                            wb.getSpecializationCriteria().matches(mOwner.getUsage()) &&
                            wb.matchesCategories(mOwner.getCategories())) {
                        if (set.add(wb)) {
                            wb.addToToolTip(toolTip);
                        }
                    }
                    break;
                case WEAPONS_WITH_REQUIRED_SKILL:
                    // Already handled
                    break;
            }
            amount.setLevel(level);
        }
    }

    private static Dice addDice(Dice left, Dice right) {
        int leftDieCount    = left.getDieCount();
        int rightDieCount   = right.getDieCount();
        int leftMultiplier  = left.getMultiplier();
        int rightMultiplier = right.getMultiplier();
        int leftModifier    = left.getModifier();
        int rightModifier   = right.getModifier();
        int leftDieSides    = left.getDieSides();
        int rightDieSides   = right.getDieSides();
        if (leftDieSides > 1 && rightDieSides > 1 && leftDieSides != rightDieSides) {
            int    sides        = Math.min(leftDieSides, rightDieSides);
            double average      = (sides + 1) / 2.0;
            double averageLeft  = (leftDieCount * (leftDieSides + 1) / 2.0) * leftMultiplier;
            double averageRight = (rightDieCount * (rightDieSides + 1) / 2.0) * rightMultiplier;
            double averageBoth  = averageLeft + averageRight;
            return new Dice((int) (averageBoth / average), sides,
                    ((int) Math.round(averageBoth % average)) + leftModifier + rightModifier, 1);
        }
        return new Dice(leftDieCount + rightDieCount, Math.max(leftDieSides, rightDieSides),
                leftModifier + rightModifier, leftMultiplier + rightMultiplier - 1);
    }

    @Override
    public String toString() {
        boolean       convertModifiersToExtraDice = mOwner.mOwner.getDataFile().getSheetSettings().useModifyingDicePlusAdds();
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
