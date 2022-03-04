/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.skill;

import com.trollworks.gcs.attribute.AttributeChoice;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.ui.widget.PopupMenu;

import java.awt.Container;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;

/** Handling of skill defaults based on type. */
public final class SkillDefaultType {
    private SkillDefaultType() {
    }

    public static PopupMenu<AttributeChoice> createPopup(Container parent, DataFile dataFile, String currentType, PopupMenu.SelectionListener<AttributeChoice> listener, boolean editable) {
        List<AttributeChoice> list = new ArrayList<>();
        for (AttributeDef def : AttributeDef.getOrdered(dataFile.getSheetSettings().getAttributes())) {
            list.add(new AttributeChoice(def.getID(), "%s", def.getName()));
        }
        list.add(new AttributeChoice("skill", "%s", "Skill"));
        list.add(new AttributeChoice("parry", "%s", "Parry"));
        list.add(new AttributeChoice("block", "%s", "Block"));
        list.add(new AttributeChoice("dodge", "%s", "Dodge"));
        list.add(new AttributeChoice("10", "%s", "10"));
        AttributeChoice current = null;
        for (AttributeChoice attributeChoice : list) {
            if (attributeChoice.getAttribute().equals(currentType)) {
                current = attributeChoice;
                break;
            }
        }
        if (current == null) {
            list.add(new AttributeChoice(currentType, "%s", currentType));
            current = list.get(list.size() - 1);
        }
        PopupMenu<AttributeChoice> popup = new PopupMenu<>(list, listener);
        popup.setSelectedItem(current, false);
        popup.setEnabled(editable);
        parent.add(popup);
        return popup;
    }

    /** @return Whether the type is based on another skill or not. */
    public static boolean isSkillBased(String type) {
        return "skill".equalsIgnoreCase(type) || "parry".equalsIgnoreCase(type) || "block".equalsIgnoreCase(type);
    }

    /**
     * @param character     The character to work with.
     * @param skillDefault  The default being calculated.
     * @param requirePoints Only look at {@link Skill}s that have points. {@link Technique}s,
     *                      however, still won't need points even if this is {@code true}.
     * @param excludes      Exclude these {@link Skill}s from consideration.
     * @param ruleOf20      {@code true} if the rule of 20 should apply.
     * @return The base skill level for this SkillDefaultType.
     */
    public static int getSkillLevelFast(GURPSCharacter character, SkillDefault skillDefault, boolean requirePoints, Set<String> excludes, boolean ruleOf20) {
        int    best;
        int    level;
        String type = skillDefault.getType();
        switch (type) {
            case "parry":
                best = getBestFast(character, skillDefault, requirePoints, excludes);
                return finalLevel(skillDefault, best == Integer.MIN_VALUE ? best : best / 2 + 3 + character.getParryBonus());
            case "block":
                best = getBestFast(character, skillDefault, requirePoints, excludes);
                return finalLevel(skillDefault, best == Integer.MIN_VALUE ? best : best / 2 + 3 + character.getBlockBonus());
            case "skill":
                return finalLevel(skillDefault, getBestFast(character, skillDefault, requirePoints, excludes));
            case "dodge":
            	level = character.getDodge(character.getEncumbranceLevel(true));
            	return finalLevel(skillDefault, ruleOf20 ? Math.min(level, 20) : level);
            default:
                level = com.trollworks.gcs.skill.Skill.resolveAttribute(character, type);
                return finalLevel(skillDefault, ruleOf20 ? Math.min(level, 20) : level);
        }
    }

    /**
     * @param character     The character to work with.
     * @param skillDefault  The default being calculated.
     * @param requirePoints Only look at {@link Skill}s that have points. {@link Technique}s,
     *                      however, still won't need points even if this is {@code true}.
     * @param excludes      Exclude these {@link Skill}s from consideration.
     * @param ruleOf20      {@code true} if the rule of 20 should apply.
     * @return The base skill level for this SkillDefaultType.
     */
    public static int getSkillLevel(GURPSCharacter character, SkillDefault skillDefault, boolean requirePoints, Set<String> excludes, boolean ruleOf20) {
        int best;
        switch (skillDefault.getType()) {
            case "parry":
                best = getBest(character, skillDefault, requirePoints, excludes);
                return finalLevel(skillDefault, best == Integer.MIN_VALUE ? best : best / 2 + 3 + character.getParryBonus());
            case "block":
                best = getBest(character, skillDefault, requirePoints, excludes);
                return finalLevel(skillDefault, best == Integer.MIN_VALUE ? best : best / 2 + 3 + character.getBlockBonus());
            case "skill":
                return finalLevel(skillDefault, getBest(character, skillDefault, requirePoints, excludes));
            default:
                return getSkillLevelFast(character, skillDefault, requirePoints, excludes, ruleOf20);
        }
    }

    private static int getBest(GURPSCharacter character, SkillDefault skillDefault, boolean requirePoints, Set<String> excludes) {
        int best = Integer.MIN_VALUE;
        for (Skill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), requirePoints, excludes)) {
            if (skill.getLevel() > best) {
                int level = skill.getLevel(excludes);
                if (level > best) {
                    best = level;
                }
            }
        }
        return best;
    }

    private static int getBestFast(GURPSCharacter character, SkillDefault skillDefault, boolean requirePoints, Set<String> excludes) {
        int best = Integer.MIN_VALUE;
        for (Skill skill : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), requirePoints, excludes)) {
            int level = skill.getLevel();
            if (level > best) {
                best = level;
            }
        }
        return best;
    }

    private static int finalLevel(SkillDefault skillDefault, int level) {
        if (level != Integer.MIN_VALUE) {
            level += skillDefault.getModifier();
        }
        return level;
    }
}
