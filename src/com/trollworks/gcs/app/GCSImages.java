/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.template.Template;
import com.trollworks.toolkit.ui.RetinaIcon;
import com.trollworks.toolkit.ui.image.ModuleImageLoader;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.image.StdImageSet;

/** Provides standardized image access. */
public class GCSImages {
    static {
        StdImage.addLoader(new ModuleImageLoader(GCSImages.class.getModule(), "/com/trollworks/gcs/app/images"));
    }

    public static final RetinaIcon getGCalcLogo() {
        return new RetinaIcon("gcalc_logo");
    }

    /** @return The exotic type icon. */
    public static final RetinaIcon getExoticTypeIcon() {
        return new RetinaIcon("exotic_type");
    }

    /** @return The mental type icon. */
    public static final RetinaIcon getMentalTypeIcon() {
        return new RetinaIcon("mental_type");
    }

    /** @return The physical type icon. */
    public static final RetinaIcon getPhysicalTypeIcon() {
        return new RetinaIcon("physical_type");
    }

    /** @return The social type icon. */
    public static final RetinaIcon getSocialTypeIcon() {
        return new RetinaIcon("social_type");
    }

    /** @return The supernatural type icon. */
    public static final RetinaIcon getSupernaturalTypeIcon() {
        return new RetinaIcon("supernatural_type");
    }

    /** @return The 'about' image. */
    public static final StdImage getAbout() {
        return StdImage.get("about");
    }

    /** @return The application icons. */
    public static final StdImageSet getAppIcons() {
        return StdImageSet.getOrLoad("app");
    }

    /** @return The character sheet icons. */
    public static final StdImageSet getCharacterSheetIcons() {
        return StdImageSet.getOrLoad(GURPSCharacter.EXTENSION);
    }

    /** @return The character template icons. */
    public static final StdImageSet getTemplateIcons() {
        return StdImageSet.getOrLoad(Template.EXTENSION);
    }

    /** @return The advantages icons. */
    public static final StdImageSet getAdvantagesIcons() {
        return StdImageSet.getOrLoad(AdvantageList.EXTENSION);
    }

    /** @return The skills icons. */
    public static final StdImageSet getSkillsIcons() {
        return StdImageSet.getOrLoad(SkillList.EXTENSION);
    }

    /** @return The spells icons. */
    public static final StdImageSet getSpellsIcons() {
        return StdImageSet.getOrLoad(SpellList.EXTENSION);
    }

    /** @return The equipment icons. */
    public static final StdImageSet getEquipmentIcons() {
        return StdImageSet.getOrLoad(EquipmentList.EXTENSION);
    }

    /** @return The note icons. */
    public static final StdImageSet getNoteIcons() {
        return StdImageSet.getOrLoad(NoteList.EXTENSION);
    }

    /** @return The character sheet icons. */
    public static final StdImageSet getCharacterSheetDocumentIcons() {
        return getDocumentIcons(GURPSCharacter.EXTENSION);
    }

    /** @return The character template icons. */
    public static final StdImageSet getTemplateDocumentIcons() {
        return getDocumentIcons(Template.EXTENSION);
    }

    /** @return The advantages icons. */
    public static final StdImageSet getAdvantagesDocumentIcons() {
        return getDocumentIcons(AdvantageList.EXTENSION);
    }

    /** @return The skills icons. */
    public static final StdImageSet getSkillsDocumentIcons() {
        return getDocumentIcons(SkillList.EXTENSION);
    }

    /** @return The spells icons. */
    public static final StdImageSet getSpellsDocumentIcons() {
        return getDocumentIcons(SpellList.EXTENSION);
    }

    /** @return The equipment icons. */
    public static final StdImageSet getEquipmentDocumentIcons() {
        return getDocumentIcons(EquipmentList.EXTENSION);
    }

    /** @return The note icons. */
    public static final StdImageSet getNoteDocumentIcons() {
        return getDocumentIcons(NoteList.EXTENSION);
    }

    private static StdImageSet getDocumentIcons(String prefix) {
        String      name = prefix + "_doc";
        StdImageSet set  = StdImageSet.get(name);
        if (set == null) {
            set = new StdImageSet(name, StdImageSet.getOrLoad("document"), StdImageSet.getOrLoad(prefix));
        }
        return set;
    }
}
