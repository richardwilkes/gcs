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

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantagesDockable;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.WeaponDamage;
import com.trollworks.gcs.weapon.WeaponSTDamage;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.event.ActionEvent;
import java.util.List;

/** Provides the "Add Natural Attacks Advantage" command. */
public class AddNaturalAttacksAdvantageCommand extends Command {
    /** The action command this command will issue. */
    public static final String                            CMD      = "AddNaturalAttacksAdvantage";
    /** The "Add Natural Attacks Advantage" command. */
    public static final AddNaturalAttacksAdvantageCommand INSTANCE = new AddNaturalAttacksAdvantageCommand();
    private             boolean                           mContainer;

    private AddNaturalAttacksAdvantageCommand() {
        super(I18n.Text("Add Natural Attacks Advantage"), CMD);
    }

    @Override
    public void adjust() {
        AdvantagesDockable adq = getTarget(AdvantagesDockable.class);
        if (adq != null) {
            setEnabled(!adq.getOutline().getModel().isLocked());
        } else {
            SheetDockable sheet = getTarget(SheetDockable.class);
            if (sheet != null) {
                setEnabled(true);
            } else {
                setEnabled(getTarget(TemplateDockable.class) != null);
            }
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ListOutline        outline;
        DataFile           dataFile;
        AdvantagesDockable adq = getTarget(AdvantagesDockable.class);
        if (adq != null) {
            dataFile = adq.getDataFile();
            outline = adq.getOutline();
            if (outline.getModel().isLocked()) {
                return;
            }
        } else {
            SheetDockable sheet = getTarget(SheetDockable.class);
            if (sheet != null) {
                dataFile = sheet.getDataFile();
                outline = sheet.getSheet().getAdvantageOutline();
            } else {
                TemplateDockable template = getTarget(TemplateDockable.class);
                if (template != null) {
                    dataFile = template.getDataFile();
                    outline = template.getTemplate().getAdvantageOutline();
                } else {
                    return;
                }
            }
        }
        Advantage advantage = new Advantage(dataFile, false);
        advantage.setName(I18n.Text("Natural Attacks"));
        advantage.setReference("B271");
        advantage.setWeapons(List.of(createBite(advantage), createPunch(advantage), createKick(advantage)));
        outline.addRow(advantage, getTitle(), false);
        outline.getModel().select(advantage, false);
        outline.scrollSelectionIntoView();
    }

    private WeaponStats createBite(Advantage owner) {
        MeleeWeaponStats  bite    = new MeleeWeaponStats(owner);
        WeaponDamage      damage  = new WeaponDamage(bite);
        damage.setType("cr");
        damage.setWeaponSTDamage(WeaponSTDamage.THR);
        damage.setBase(new Dice(0, -1));
        bite.setDamage(damage);
        bite.setUsage("Bite");
        bite.setReach("C");
        bite.setParry("No");
        bite.setBlock("No");
        bite.setDefaults(List.of(new SkillDefault(SkillDefaultType.DX, null, null, 0), new SkillDefault(SkillDefaultType.Skill, "Brawling", null, 0)));
        return bite;
    }

    private WeaponStats createPunch(Advantage owner) {
        MeleeWeaponStats  punch    = new MeleeWeaponStats(owner);
        WeaponDamage      damage  = new WeaponDamage(punch);
        damage.setType("cr");
        damage.setWeaponSTDamage(WeaponSTDamage.THR);
        damage.setBase(new Dice(0, -1));
        punch.setDamage(damage);
        punch.setUsage("Punch");
        punch.setReach("C");
        punch.setParry("0");
        punch.setDefaults(List.of(new SkillDefault(SkillDefaultType.DX, null, null, 0), new SkillDefault(SkillDefaultType.Skill, "Boxing", null, 0), new SkillDefault(SkillDefaultType.Skill, "Brawling", null, 0), new SkillDefault(SkillDefaultType.Skill, "Karate", null, 0)));
        return punch;
    }

    private WeaponStats createKick(Advantage owner) {
        MeleeWeaponStats  kick    = new MeleeWeaponStats(owner);
        WeaponDamage      damage  = new WeaponDamage(kick);
        damage.setType("cr");
        damage.setWeaponSTDamage(WeaponSTDamage.THR);
        kick.setDamage(damage);
        kick.setUsage("Kick");
        kick.setReach("C,1");
        kick.setParry("No");
        kick.setDefaults(List.of(new SkillDefault(SkillDefaultType.DX, null, null, -2), new SkillDefault(SkillDefaultType.Skill, "Brawling", null, -2), new SkillDefault(SkillDefaultType.Skill, "Karate", null, -2)));
        return kick;
    }
}
