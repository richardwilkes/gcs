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
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.gcs.feature.WeaponBonus;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.FilteredList;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.xml.XMLNodeType;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/** The stats for a weapon. */
public abstract class WeaponStats {
    private static final String             KEY_DEFAULTS = "defaults";
    private static final String             TAG_STRENGTH = "strength";
    private static final String             TAG_USAGE    = "usage";
    /** The prefix used in front of all IDs for weapons. */
    public static final  String             PREFIX       = GURPSCharacter.CHARACTER_PREFIX + "weapon.";
    /** The field ID for damage changes. */
    public static final  String             ID_DAMAGE    = PREFIX + WeaponDamage.TAG_ROOT;
    /** The field ID for strength changes. */
    public static final  String             ID_STRENGTH  = PREFIX + TAG_STRENGTH;
    /** The field ID for usage changes. */
    public static final  String             ID_USAGE     = PREFIX + TAG_USAGE;
    protected            ListRow            mOwner;
    private              WeaponDamage       mDamage;
    private              String             mStrength;
    private              String             mUsage;
    private              List<SkillDefault> mDefaults;

    public static void loadFromJSONArray(ListRow row, JsonArray a, List<WeaponStats> list) throws IOException {
        int       count = a.size();
        for (int i = 0; i < count; i++) {
            JsonMap m1   = a.getMap(i);
            String  type = m1.getString(DataFile.KEY_TYPE);
            switch (type) {
            case MeleeWeaponStats.TAG_ROOT:
                list.add(new MeleeWeaponStats(row, m1));
                break;
            case RangedWeaponStats.TAG_ROOT:
                list.add(new RangedWeaponStats(row, m1));
                break;
            default:
                Log.warn("unknown weapon type: " + type);
                break;
            }
        }
    }

    public static void saveList(JsonWriter w, String key, List<?> list) throws IOException {
        FilteredList<WeaponStats> rows = new FilteredList<>(list, WeaponStats.class, true);
        if (!rows.isEmpty()) {
            w.key(key);
            w.startArray();
            for (WeaponStats row : rows) {
                row.save(w);
            }
            w.endArray();
        }
    }


    /**
     * Creates a new weapon.
     *
     * @param owner The owning piece of equipment or advantage.
     */
    protected WeaponStats(ListRow owner) {
        mOwner = owner;
        mDamage = new WeaponDamage(this);
        mStrength = "";
        mUsage = "";
        mDefaults = new ArrayList<>();
        initialize();
    }

    /**
     * Creates a clone of the specified weapon.
     *
     * @param owner The owning piece of equipment or advantage.
     * @param other The weapon to clone.
     */
    protected WeaponStats(ListRow owner, WeaponStats other) {
        mOwner = owner;
        mDamage = other.mDamage.clone(this);
        mStrength = other.mStrength;
        mUsage = other.mUsage;
        mDefaults = new ArrayList<>();
        for (SkillDefault skillDefault : other.mDefaults) {
            mDefaults.add(new SkillDefault(skillDefault));
        }
    }

    /**
     * Creates a weapon.
     *
     * @param owner The owning piece of equipment or advantage.
     * @param m     The {@link JsonMap} to load from.
     */
    public WeaponStats(ListRow owner, JsonMap m) throws IOException {
        this(owner);
        loadSelf(m);
    }

    /**
     * Creates a weapon.
     *
     * @param owner  The owning piece of equipment or advantage.
     * @param reader The reader to load from.
     */
    public WeaponStats(ListRow owner, XMLReader reader) throws IOException {
        this(owner);
        String marker = reader.getMarker();
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                loadSelf(reader);
            }
        } while (reader.withinMarker(marker));
    }

    /**
     * Creates a clone of this weapon.
     *
     * @param owner The owning piece of equipment or advantage.
     * @return The cloned weapon.
     */
    public abstract WeaponStats clone(ListRow owner);

    /** Called so that sub-classes can initialize themselves. */
    protected abstract void initialize();

    /** @return The type name to use for this data. */
    public abstract String getJSONTypeName();

    /** @return The root XML tag to use when saving. */
    protected abstract String getRootTag();

    /** @param reader The reader to load from. */
    protected void loadSelf(XMLReader reader) throws IOException {
        String name = reader.getName();

        if (WeaponDamage.TAG_ROOT.equals(name)) {
            mDamage = new WeaponDamage(reader, this);
        } else if (TAG_STRENGTH.equals(name)) {
            mStrength = reader.readText();
        } else if (TAG_USAGE.equals(name)) {
            mUsage = reader.readText();
        } else if (SkillDefault.TAG_ROOT.equals(name)) {
            mDefaults.add(new SkillDefault(reader));
        } else {
            reader.skipTag(name);
        }
    }

    /** @param m The {@link JsonMap} to load from. */
    protected void loadSelf(JsonMap m) throws IOException {
        mDamage = new WeaponDamage(m.getMap(WeaponDamage.TAG_ROOT), this);
        mStrength = m.getString(TAG_STRENGTH);
        mUsage = m.getString(TAG_USAGE);
        if (m.has(KEY_DEFAULTS)) {
            JsonArray a     = m.getArray(KEY_DEFAULTS);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                mDefaults.add(new SkillDefault(a.getMap(i), false));
            }
        }
    }

    /**
     * Saves the weapon.
     *
     * @param w The {@link JsonWriter} to use.
     */
    public final void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(DataFile.KEY_TYPE, getJSONTypeName());
        w.key(WeaponDamage.TAG_ROOT);
        mDamage.save(w);
        w.keyValueNot(TAG_STRENGTH, mStrength, "");
        w.keyValueNot(TAG_USAGE, mUsage, "");
        saveSelf(w);
        if (!mDefaults.isEmpty()) {
            w.key(KEY_DEFAULTS);
            w.startArray();
            for (SkillDefault skillDefault : mDefaults) {
                skillDefault.save(w, false);
            }
            w.endArray();
        }
        w.endMap();
    }

    /**
     * Called so that sub-classes can save their own data.
     *
     * @param w The {@link JsonWriter} to use.
     */
    protected abstract void saveSelf(JsonWriter w) throws IOException;

    /** @return The defaults for this weapon. */
    public List<SkillDefault> getDefaults() {
        return Collections.unmodifiableList(mDefaults);
    }

    /**
     * @param defaults The new defaults for this weapon.
     * @return Whether there was a change or not.
     */
    public boolean setDefaults(List<SkillDefault> defaults) {
        if (!mDefaults.equals(defaults)) {
            mDefaults = new ArrayList<>(defaults);
            return true;
        }
        return false;
    }

    /** @param id The ID to use for notification. */
    protected void notifySingle(String id) {
        if (mOwner != null) {
            mOwner.notifySingle(id);
        }
    }

    /** @return A description of the weapon. */
    public String getDescription() {
        if (mOwner instanceof Equipment) {
            return ((Equipment) mOwner).getDescription();
        }
        if (mOwner instanceof Advantage) {
            return ((Advantage) mOwner).getName();
        }
        if (mOwner instanceof Spell) {
            return ((Spell) mOwner).getName();
        }
        if (mOwner instanceof Skill) {
            return ((Skill) mOwner).getName();
        }
        return "";
    }

    @Override
    public String toString() {
        return getDescription();
    }

    /** @return The notes for this weapon. */
    public String getNotes() {
        return mOwner != null ? mOwner.getNotes() : "";
    }

    /** @return The damage. */
    public WeaponDamage getDamage() {
        return mDamage;
    }

    /**
     * @param buffer The string to find the next non-space character within.
     * @param index  The index to start looking.
     * @return The index of the next non-space character.
     */
    @SuppressWarnings("static-method")
    protected int skipSpaces(String buffer, int index) {
        int max = buffer.length();
        while (index < max && buffer.charAt(index) == ' ') {
            index++;
        }
        return index;
    }

    /**
     * Sets the value of damage.
     *
     * @param damage The value to set.
     */
    public void setDamage(WeaponDamage damage) {
        if (damage == null) {
            damage = new WeaponDamage(this);
        }
        if (!mDamage.equivalent(damage)) {
            mDamage = damage.clone(this);
            notifySingle(ID_DAMAGE);
        }
    }

    public String getSkillLevelToolTip() {
        StringBuilder toolTip = new StringBuilder();
        DataFile      df      = mOwner.getDataFile();
        if (df instanceof GURPSCharacter) {
            getSkillLevel((GURPSCharacter) df, toolTip);
        }
        return toolTip.length() > 0 ? I18n.Text("Includes modifiers from") + toolTip : I18n.Text("No additional modifiers");
    }

    /** @return The skill level. */
    public int getSkillLevel() {
        DataFile df = mOwner.getDataFile();
        if (df instanceof GURPSCharacter) {
            return getSkillLevel((GURPSCharacter) df, null);
        }
        return 0;
    }

    private int getSkillLevel(GURPSCharacter character, StringBuilder toolTip) {
        int best = Integer.MIN_VALUE;
        for (SkillDefault skillDefault : getDefaults()) {
            SkillDefaultType type  = skillDefault.getType();
            int              level = type.getSkillLevelFast(character, skillDefault, false, new HashSet<>());
            if (level > best) {
                best = level;
            }
        }
        if (best == Integer.MIN_VALUE) {
            best = 0;
        } else {
            int minST = getMinStrengthValue() - (character.getStrength() + character.getStrikingStrengthBonus());
            if (minST > 0) {
                best -= minST;
            }
            if (this instanceof MeleeWeaponStats) {
                if (((MeleeWeaponStats) this).getParry().contains("F")) {
                    best += character.getEncumbranceLevel().getEncumbrancePenalty();
                }
            }
            if (best < 0) {
                best = 0;
            } else {
                String      nameQualifier  = toString();
                String      usageQualifier = getUsage();
                Set<String> categories     = getCategories();
                for (SkillBonus bonus : character.getNamedWeaponSkillBonusesFor(WeaponBonus.WEAPON_NAMED_ID_PREFIX + "*", nameQualifier, usageQualifier, categories, toolTip)) {
                    best += bonus.getAmount().getIntegerAdjustedAmount();
                }
                for (SkillBonus bonus : character.getNamedWeaponSkillBonusesFor(WeaponBonus.WEAPON_NAMED_ID_PREFIX + "/" + nameQualifier, nameQualifier, usageQualifier, categories, toolTip)) {
                    best += bonus.getAmount().getIntegerAdjustedAmount();
                }
                for (Feature feature : mOwner.getFeatures()) {
                    best += extractSkillBonus(feature, toolTip);
                }
                if (mOwner instanceof Advantage) {
                    for (AdvantageModifier modifier : ((Advantage) mOwner).getModifiers()) {
                        if (modifier.isEnabled()) {
                            for (Feature feature : modifier.getFeatures()) {
                                best += extractSkillBonus(feature, toolTip);
                            }
                        }
                    }
                }
                if (mOwner instanceof Equipment) {
                    for (EquipmentModifier modifier : ((Equipment) mOwner).getModifiers()) {
                        if (modifier.isEnabled()) {
                            for (Feature feature : modifier.getFeatures()) {
                                best += extractSkillBonus(feature, toolTip);
                            }
                        }
                    }
                }
                if (best < 0) {
                    best = 0;
                }
            }
        }
        return best;
    }

    private int extractSkillBonus(Feature feature, StringBuilder toolTip) {
        if (feature instanceof SkillBonus) {
            SkillBonus sb = (SkillBonus) feature;
            switch (sb.getSkillSelectionType()) {
            case THIS_WEAPON:
            default:
                sb.addToToolTip(toolTip);
                return sb.getAmount().getIntegerAdjustedAmount();
            case WEAPONS_WITH_NAME:
                if (sb.getNameCriteria().matches(mOwner.toString()) && sb.getSpecializationCriteria().matches(getUsage()) && sb.matchesCategories(getCategories())) {
                    sb.addToToolTip(toolTip);
                    return sb.getAmount().getIntegerAdjustedAmount();
                }
                break;
            case SKILLS_WITH_NAME:
                // Already handled
                break;
            }
        }
        return 0;
    }

    /** @return The minimum ST to use this weapon, or -1 if there is none. */
    public int getMinStrengthValue() {
        StringBuilder builder = new StringBuilder();
        int           count   = mStrength.length();
        boolean       started = false;
        for (int i = 0; i < count; i++) {
            char ch = mStrength.charAt(i);
            if (Character.isDigit(ch)) {
                builder.append(ch);
                started = true;
            } else if (started) {
                break;
            }
        }
        return started ? Numbers.extractInteger(builder.toString(), -1, false) : -1;
    }

    /** @return The usage. */
    public String getUsage() {
        return mUsage;
    }

    /** @param usage The value to set. */
    public void setUsage(String usage) {
        usage = sanitize(usage);
        if (!mUsage.equals(usage)) {
            mUsage = usage;
            notifySingle(ID_USAGE);
        }
    }

    /** @return The strength. */
    public String getStrength() {
        return mStrength;
    }

    /**
     * Sets the value of strength.
     *
     * @param strength The value to set.
     */
    public void setStrength(String strength) {
        strength = sanitize(strength);
        if (!mStrength.equals(strength)) {
            mStrength = strength;
            notifySingle(ID_STRENGTH);
        }
    }

    /** @return The owner. */
    public ListRow getOwner() {
        return mOwner;
    }

    public Set<String> getCategories() {
        return mOwner.getCategories();
    }

    /**
     * Sets the value of owner.
     *
     * @param owner The value to set.
     */
    public void setOwner(ListRow owner) {
        mOwner = owner;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof WeaponStats) {
            WeaponStats ws = (WeaponStats) obj;
            return mDamage.equivalent(ws.mDamage) && mStrength.equals(ws.mStrength) && mUsage.equals(ws.mUsage) && mDefaults.equals(ws.mDefaults);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    /**
     * @param data The data to sanitize.
     * @return The original data, or "", if the data was {@code null}.
     */
    @SuppressWarnings("static-method")
    protected String sanitize(String data) {
        if (data == null) {
            return "";
        }
        return data;
    }
}
