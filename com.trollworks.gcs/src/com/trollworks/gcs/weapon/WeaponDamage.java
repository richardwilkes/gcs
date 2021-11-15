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
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.settings.DamageProgression;

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
    private static final String KEY_PERCENT_BONUS               = "percent_bonus";

    private WeaponStats    mOwner;
    private String         mType;
    private WeaponSTDamage mST;
    private Dice           mBase;
    private double         mArmorDivisor;
    private Dice           mFragmentation;
    private double         mFragmentationArmorDivisor;
    private String         mFragmentationType;
    private int            mModifierPerDie;
    private Integer        mPercentBonus;

    public WeaponDamage(WeaponStats owner) {
        mType = "";
        mOwner = owner;
        mST = WeaponSTDamage.NONE;
        mArmorDivisor = 1;
        mPercentBonus = 0;
    }

    public WeaponDamage(JsonMap m, WeaponStats owner) {
        mOwner = owner;
        mType = m.getString(DataFile.TYPE);
        mST = Enums.extract(m.getString(KEY_ST), WeaponSTDamage.values(), WeaponSTDamage.NONE);
        if (m.has(KEY_BASE)) {
            mBase = new Dice(m.getString(KEY_BASE));
        }
        mArmorDivisor = m.getDoubleWithDefault(KEY_ARMOR_DIVISOR, 1);
        mPercentBonus = m.getIntWithDefault(KEY_PERCENT_BONUS, 0);
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
        if (mPercentBonus != 0) {
            Integer percentbonus = mPercentBonus;
            w.keyValue(KEY_PERCENT_BONUS, percentbonus);
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
        other.mPercentBonus = mPercentBonus;
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
            if (mType.equals(other.mType) && mST == other.mST && mArmorDivisor == other.mArmorDivisor && mPercentBonus == other.mPercentBonus && mModifierPerDie == other.mModifierPerDie && Objects.equals(mBase, other.mBase)) {
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

    public Integer getPercentBonus() {
        return mPercentBonus;
    }

    public void setPercentBonus(Integer percentBonus) {
        if (mPercentBonus != percentBonus) {
            mPercentBonus = percentBonus;
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
                double percent = 0.0;
                if (mOwner.mOwner.getDataFile().getSheetSettings().useBaseDamagePercentBonus()) {
                    percent = mPercentBonus / 100.0;
                }
                switch (mST) {
                case SW -> {
                    Dice swing = character.getSwing(st);
                    swing.percentAdd(percent);
                    base = addDice(base, swing);
                }
                case SW_LEVELED -> {
                    Dice swing = character.getSwing(st);
                    swing.percentAdd(percent);
                    if (mOwner.mOwner instanceof Advantage advantage) {
                        if (advantage.isLeveled()) {
                            swing.multiply(advantage.getLevels());
                        }
                    }
                    base = addDice(base, swing);
                }
                case THR -> {
                    Dice thrust = character.getThrust(st);
                    thrust.percentAdd(percent);
                    base = addDice(base, thrust);
                }
                case THR_LEVELED -> {
                    Dice thrust = character.getThrust(st);
                    thrust.percentAdd(percent);
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
                for (WeaponDamageBonus bonus : bonusSet) {
                    LeveledAmount lvlAmt = bonus.getAmount();
                    int           amt    = lvlAmt.getIntegerAmount();
                    if (lvlAmt.isPerLevel()) {
                        if (mOwner.mOwner.getDataFile().getSheetSettings().getDamageProgression() == DamageProgression.PHOENIX_D3 && base.getDieSides() == 3) {
                            base.add(amt * Math.max(Math.floorDiv(base.getDieCount(), 2), 1));
                        } else {
                            base.add(amt * base.getDieCount());
                        }

                    } else {
                        base.add(amt);
                    }
                }
                if (mModifierPerDie != 0) {
                    if (mOwner.mOwner.getDataFile().getSheetSettings().getDamageProgression() == DamageProgression.PHOENIX_D3 && base.getDieSides() == 3) {
                        base.add(mModifierPerDie * Math.max(Math.floorDiv(base.getDieCount(), 2), 1));
                    } else {
                        base.add(mModifierPerDie * base.getDieCount());
                    }
                }
                if (mOwner.mOwner.getDataFile().getSheetSettings().useDamageDiceConversion() && mOwner.mOwner.getDataFile().getSheetSettings().getDamageDiceConversionDie() != base.getDieSides()) {
                    // Division by 0 would be bad, and 1 sided die don't exist, so we guard against that.
                    Integer conversionDie = mOwner.mOwner.getDataFile().getSheetSettings().getDamageDiceConversionDie();
                    if (conversionDie > 1) {
                        // Converts dice to raw averages, then halves them and takes any remainder to get a +1
                        double conversionDieAverage = (conversionDie + 1) / 2.0;
                        double dicevalue            = (base.getDieCount() * (base.getDieSides() + 1)) / 2.0;
                        double newcount             = (dicevalue / conversionDieAverage);
                        int    newmod               = (int) Math.round(dicevalue % conversionDieAverage);
                        int    count                = (int) newcount;
                        if (newcount < 1) {
                            newmod = (int) Math.round(dicevalue - conversionDieAverage);
                            count = 1;
                        }
                        Dice newDice = new Dice(count, conversionDie, newmod + base.getModifier(), base.getMultiplier());//we shouldn't have to touch the multiplier because it'll be the same
                        base = newDice;

                    }
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

    private void extractWeaponDamageBonus(Feature feature, Set<WeaponDamageBonus> set, int dieCount, StringBuilder toolTip) {
        if (feature instanceof WeaponDamageBonus wb) {
            LeveledAmount     amount = wb.getAmount();
            int               level  = amount.getLevel();
            amount.setLevel(dieCount);
            switch (wb.getWeaponSelectionType()) {
            case THIS_WEAPON:
            default:
                if (set.add(wb)) {
                    wb.addToToolTip(toolTip);
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

    private Dice addDice(Dice left, Dice right) {
        // Check if the sides are different, otherwise just add as normal.
        String behavior = mOwner.mOwner.getDataFile().getSheetSettings().getDiceAdditionBehavior();
        // set sides to 6 as default.
        int sides = 6;
        if (mOwner.mOwner.getDataFile().getSheetSettings().useDamageDiceConversion()) {
            sides = mOwner.mOwner.getDataFile().getSheetSettings().getDamageDiceConversionDie();
        }
        if (left.getDieSides() != right.getDieSides() && behavior == "Lower" || behavior == "Higher") {
            //get the larger die
            if (behavior == "Lower") {
                sides = Math.min(left.getDieSides(), right.getDieSides());
            } else if (behavior == "Higher") {
                sides = Math.max(left.getDieSides(), right.getDieSides());
            }
            double dicevalue = (sides + 1) / 2.0;
            // Converts both sides to raw averages, then halves them and takes any remainder to get the mod
            double leftval  = (left.getDieCount() * (left.getDieSides() + 1) / 2.0) * left.getMultiplier();
            double rightval = (right.getDieCount() * (right.getDieSides() + 1) / 2.0) * right.getMultiplier();
            int    dieCount = (int) ((leftval + rightval) / dicevalue);
            int    baseMod  = (int) Math.round((leftval + rightval) % dicevalue);
            return new Dice(dieCount, sides, baseMod + left.getModifier() + right.getModifier(), 1);
        } else {
            //Just Add is the default behavior.
            return new Dice(left.getDieCount() + right.getDieCount(), Math.max(left.getDieSides(), right.getDieSides()), left.getModifier() + right.getModifier(), left.getMultiplier() + right.getMultiplier() - 1);
        }
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
