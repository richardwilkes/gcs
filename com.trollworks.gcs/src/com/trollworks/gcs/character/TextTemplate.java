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

package com.trollworks.gcs.character;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageColumn;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentColumn;
import com.trollworks.gcs.feature.DRBonus;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillColumn;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellColumn;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowIterator;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.FilteredIterator;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.DateTimeFormatter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.xml.XMLWriter;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponDisplayRow;
import com.trollworks.gcs.weapon.WeaponStats;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Base64;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.function.Function;
import javax.imageio.ImageIO;

/** Provides text template output. */
public class TextTemplate {
    private static final String         UNIDENTIFIED_KEY                      = "Unidentified key: '%s'";
    private static final String         CURRENT                               = "current";
    private static final String         ITEM                                  = "ITEM";
    private static final String         ONE                                   = "1";
    private static final String         UNDERSCORE                            = "_";
    private static final String         PARAGRAPH_START                       = "<p>";
    private static final String         PARAGRAPH_END                         = "</p>";
    private static final String         NEWLINE                               = "\n";
    private static final String         COMMA_SEPARATOR                       = ", ";
    private static final String         KEY_ACCURACY                          = "ACCURACY";
    private static final String         KEY_ADVANTAGE_POINTS                  = "ADVANTAGE_POINTS";
    private static final String         KEY_ADVANTAGES_ALL_LOOP_END           = "ADVANTAGES_ALL_LOOP_END";
    private static final String         KEY_ADVANTAGES_ALL_LOOP_START         = "ADVANTAGES_ALL_LOOP_START";
    private static final String         KEY_ADVANTAGES_LOOP_END               = "ADVANTAGES_LOOP_END";
    private static final String         KEY_ADVANTAGES_LOOP_START             = "ADVANTAGES_LOOP_START";
    private static final String         KEY_ADVANTAGES_ONLY_LOOP_END          = "ADVANTAGES_ONLY_LOOP_END";
    private static final String         KEY_ADVANTAGES_ONLY_LOOP_START        = "ADVANTAGES_ONLY_LOOP_START";
    private static final String         KEY_AGE                               = "AGE";
    private static final String         KEY_AMMO                              = "AMMO";
    private static final String         KEY_ATTACK_MODES_LOOP_END             = "ATTACK_MODES_LOOP_END";
    private static final String         KEY_ATTACK_MODES_LOOP_START           = "ATTACK_MODES_LOOP_START";
    private static final String         KEY_ATTRIBUTE_POINTS                  = "ATTRIBUTE_POINTS";
    private static final String         KEY_BASIC_FP                          = "BASIC_FP";
    private static final String         KEY_BASIC_HP                          = "BASIC_HP";
    private static final String         KEY_BASIC_LIFT                        = "BASIC_LIFT";
    private static final String         KEY_BASIC_MOVE                        = "BASIC_MOVE";
    private static final String         KEY_BASIC_MOVE_POINTS                 = "BASIC_MOVE_POINTS";
    private static final String         KEY_BASIC_SPEED                       = "BASIC_SPEED";
    private static final String         KEY_BASIC_SPEED_POINTS                = "BASIC_SPEED_POINTS";
    private static final String         KEY_BEST_CURRENT_BLOCK                = "BEST_CURRENT_BLOCK";
    private static final String         KEY_BEST_CURRENT_PARRY                = "BEST_CURRENT_PARRY";
    private static final String         KEY_BIRTHDAY                          = "BIRTHDAY";
    private static final String         KEY_BLOCK                             = "BLOCK";
    private static final String         KEY_BULK                              = "BULK";
    private static final String         KEY_CARRIED_STATUS                    = "CARRIED_STATUS";
    private static final String         KEY_CARRIED_VALUE                     = "CARRIED_VALUE";
    private static final String         KEY_CARRIED_WEIGHT                    = "CARRIED_WEIGHT";
    private static final String         KEY_CARRY_ON_BACK                     = "CARRY_ON_BACK";
    private static final String         KEY_CATEGORIES                        = "CATEGORIES";
    private static final String         KEY_CLASS                             = "CLASS";
    private static final String         KEY_COLLEGE                           = "COLLEGE";
    private static final String         KEY_CONTINUE_ID                       = "CONTINUE_ID";
    private static final String         KEY_COST                              = "COST";
    private static final String         KEY_COST_SUMMARY                      = "COST_SUMMARY";
    private static final String         KEY_CREATED_ON                        = "CREATED_ON";
    private static final String         KEY_CULTURAL_FAMILIARITIES_LOOP_END   = "CULTURAL_FAMILIARITIES_LOOP_END";
    private static final String         KEY_CULTURAL_FAMILIARITIES_LOOP_START = "CULTURAL_FAMILIARITIES_LOOP_START";
    private static final String         KEY_CURRENT_DODGE                     = "CURRENT_DODGE";
    private static final String         KEY_CURRENT_MARKER                    = "CURRENT_MARKER";
    private static final String         KEY_CURRENT_MARKER_1                  = "CURRENT_MARKER_1";
    private static final String         KEY_CURRENT_MARKER_BULLET             = "CURRENT_MARKER_BULLET";
    private static final String         KEY_CURRENT_MOVE                      = "CURRENT_MOVE";
    private static final String         KEY_DAMAGE                            = "DAMAGE";
    private static final String         KEY_UNMODIFIED_DAMAGE                 = "UNMODIFIED_DAMAGE";
    private static final String         KEY_DEAD                              = "DEAD";
    private static final String         KEY_DEATH_CHECK_1                     = "DEATH_CHECK_1";
    private static final String         KEY_DEATH_CHECK_2                     = "DEATH_CHECK_2";
    private static final String         KEY_DEATH_CHECK_3                     = "DEATH_CHECK_3";
    private static final String         KEY_DEATH_CHECK_4                     = "DEATH_CHECK_4";
    private static final String         KEY_DESCRIPTION                       = "DESCRIPTION";
    private static final String         KEY_DESCRIPTION_MODIFIER_NOTES        = "DESCRIPTION_MODIFIER_NOTES";
    private static final String         KEY_DESCRIPTION_NOTES                 = "DESCRIPTION_NOTES";
    private static final String         KEY_DESCRIPTION_PRIMARY               = "DESCRIPTION_PRIMARY";
    private static final String         KEY_DESCRIPTION_USER                  = "DESCRIPTION_USER";
    private static final String         KEY_DESCRIPTION_USER_FORMATTED        = "DESCRIPTION_USER_FORMATTED";
    private static final String         KEY_DIFFICULTY                        = "DIFFICULTY";
    private static final String         KEY_DISADVANTAGE_POINTS               = "DISADVANTAGE_POINTS";
    private static final String         KEY_DISADVANTAGES_ALL_LOOP_END        = "DISADVANTAGES_ALL_LOOP_END";
    private static final String         KEY_DISADVANTAGES_ALL_LOOP_START      = "DISADVANTAGES_ALL_LOOP_START";
    private static final String         KEY_DISADVANTAGES_LOOP_END            = "DISADVANTAGES_LOOP_END";
    private static final String         KEY_DISADVANTAGES_LOOP_START          = "DISADVANTAGES_LOOP_START";
    private static final String         KEY_DODGE                             = "DODGE";
    private static final String         KEY_DR                                = "DR";
    private static final String         KEY_DURATION                          = "DURATION";
    private static final String         KEY_DX                                = "DX";
    private static final String         KEY_DX_POINTS                         = "DX_POINTS";
    private static final String         KEY_UNSPENT_POINTS                    = "UNSPENT_POINTS";
    private static final String         KEY_ENCODING_OFF                      = "ENCODING_OFF";
    private static final String         KEY_ENCUMBRANCE_LOOP_END              = "ENCUMBRANCE_LOOP_END";
    private static final String         KEY_ENCUMBRANCE_LOOP_START            = "ENCUMBRANCE_LOOP_START";
    private static final String         KEY_ENHANCED_KEY_PARSING              = "ENHANCED_KEY_PARSING";
    private static final String         KEY_EQUIPMENT                         = "EQUIPMENT";
    private static final String         KEY_EQUIPMENT_FORMATTED               = "EQUIPMENT_FORMATTED";
    private static final String         KEY_EQUIPMENT_LOOP_END                = "EQUIPMENT_LOOP_END";
    private static final String         KEY_EQUIPMENT_LOOP_START              = "EQUIPMENT_LOOP_START";
    private static final String         KEY_EQUIPPED                          = "EQUIPPED";
    private static final String         KEY_EQUIPPED_NUM                      = "EQUIPPED_NUM";
    private static final String         KEY_EXCLUDE_CATEGORIES                = "EXCLUDE_CATEGORIES_";
    private static final String         KEY_EYES                              = "EYES";
    private static final String         KEY_FP                                = "FP";
    private static final String         KEY_FP_COLLAPSE                       = "FP_COLLAPSE";
    private static final String         KEY_FP_POINTS                         = "FP_POINTS";
    private static final String         KEY_FRIGHT_CHECK                      = "FRIGHT_CHECK";
    private static final String         KEY_GENDER                            = "GENDER";
    private static final String         KEY_GENERAL_DR                        = "GENERAL_DR";
    private static final String         KEY_GRID_TEMPLATE                     = "GRID_TEMPLATE";
    private static final String         KEY_HAIR                              = "HAIR";
    private static final String         KEY_HAND                              = "HAND";
    private static final String         KEY_HEARING                           = "HEARING";
    private static final String         KEY_HEIGHT                            = "HEIGHT";
    private static final String         KEY_HIERARCHICAL_MELEE_LOOP_END       = "HIERARCHICAL_MELEE_LOOP_END";
    private static final String         KEY_HIERARCHICAL_MELEE_LOOP_START     = "HIERARCHICAL_MELEE_LOOP_START";
    private static final String         KEY_HIERARCHICAL_RANGED_LOOP_END      = "HIERARCHICAL_RANGED_LOOP_END";
    private static final String         KEY_HIERARCHICAL_RANGED_LOOP_START    = "HIERARCHICAL_RANGED_LOOP_START";
    private static final String         KEY_HIT_LOCATION_LOOP_END             = "HIT_LOCATION_LOOP_END";
    private static final String         KEY_HIT_LOCATION_LOOP_START           = "HIT_LOCATION_LOOP_START";
    private static final String         KEY_HP                                = "HP";
    private static final String         KEY_HP_COLLAPSE                       = "HP_COLLAPSE";
    private static final String         KEY_HP_POINTS                         = "HP_POINTS";
    private static final String         KEY_HT                                = "HT";
    private static final String         KEY_HT_POINTS                         = "HT_POINTS";
    private static final String         KEY_ID                                = "ID";
    private static final String         KEY_IQ                                = "IQ";
    private static final String         KEY_IQ_POINTS                         = "IQ_POINTS";
    private static final String         KEY_LANGUAGES_LOOP_END                = "LANGUAGES_LOOP_END";
    private static final String         KEY_LANGUAGES_LOOP_START              = "LANGUAGES_LOOP_START";
    private static final String         KEY_LEGALITY_CLASS                    = "LEGALITY_CLASS";
    private static final String         KEY_LEVEL                             = "LEVEL";
    private static final String         KEY_LEVEL_ONLY                        = "LEVEL_ONLY";
    private static final String         KEY_LEVEL_NO_MARKER                   = "LEVEL_NO_MARKER";
    private static final String         KEY_LOCATION                          = "LOCATION";
    private static final String         KEY_MANA_CAST                         = "MANA_CAST";
    private static final String         KEY_MANA_MAINTAIN                     = "MANA_MAINTAIN";
    private static final String         KEY_MAX_LOAD                          = "MAX_LOAD";
    private static final String         KEY_MELEE_LOOP_END                    = "MELEE_LOOP_END";
    private static final String         KEY_MELEE_LOOP_START                  = "MELEE_LOOP_START";
    private static final String         KEY_MODIFIED_ON                       = "MODIFIED_ON";
    private static final String         KEY_MODIFIER                          = "MODIFIER";
    private static final String         KEY_MODIFIER_NOTES_FOR                = "MODIFIER_NOTES_FOR_";
    private static final String         KEY_MOVE                              = "MOVE";
    private static final String         KEY_NAME                              = "NAME";
    private static final String         KEY_NOTE                              = "NOTE";
    private static final String         KEY_NOTE_FORMATTED                    = "NOTE_FORMATTED";
    private static final String         KEY_NOTES                             = "NOTES";
    private static final String         KEY_NOTES_LOOP_END                    = "NOTES_LOOP_END";
    private static final String         KEY_NOTES_LOOP_START                  = "NOTES_LOOP_START";
    private static final String         KEY_ONE_HANDED_LIFT                   = "ONE_HANDED_LIFT";
    private static final String         KEY_ONLY_CATEGORIES                   = "ONLY_CATEGORIES_";
    private static final String         KEY_OTHER_EQUIPMENT_LOOP_END          = "OTHER_EQUIPMENT_LOOP_END";
    private static final String         KEY_OTHER_EQUIPMENT_LOOP_START        = "OTHER_EQUIPMENT_LOOP_START";
    private static final String         KEY_OTHER_VALUE                       = "OTHER_EQUIPMENT_VALUE";
    private static final String         KEY_PARRY                             = "PARRY";
    private static final String         KEY_PENALTY                           = "PENALTY";
    private static final String         KEY_PERCEPTION                        = "PERCEPTION";
    private static final String         KEY_PERCEPTION_POINTS                 = "PERCEPTION_POINTS";
    private static final String         KEY_PERKS_LOOP_END                    = "PERKS_LOOP_END";
    private static final String         KEY_PERKS_LOOP_START                  = "PERKS_LOOP_START";
    private static final String         KEY_PLAYER                            = "PLAYER";
    private static final String         KEY_POINTS                            = "POINTS";
    private static final String         KEY_PORTRAIT                          = "PORTRAIT";
    private static final String         KEY_PORTRAIT_EMBEDDED                 = "PORTRAIT_EMBEDDED";
    private static final String         KEY_PREFIX_DEPTH                      = "DEPTHx";
    private static final String         KEY_QTY                               = "QTY";
    private static final String         KEY_QUIRK_POINTS                      = "QUIRK_POINTS";
    private static final String         KEY_QUIRKS_LOOP_END                   = "QUIRKS_LOOP_END";
    private static final String         KEY_QUIRKS_LOOP_START                 = "QUIRKS_LOOP_START";
    private static final String         KEY_RACE_POINTS                       = "RACE_POINTS";
    private static final String         KEY_RANGE                             = "RANGE";
    private static final String         KEY_RANGED_LOOP_END                   = "RANGED_LOOP_END";
    private static final String         KEY_RANGED_LOOP_START                 = "RANGED_LOOP_START";
    private static final String         KEY_REACH                             = "REACH";
    private static final String         KEY_REACTION_LOOP_END                 = "REACTION_LOOP_END";
    private static final String         KEY_REACTION_LOOP_START               = "REACTION_LOOP_START";
    private static final String         KEY_RECOIL                            = "RECOIL";
    private static final String         KEY_REELING                           = "REELING";
    private static final String         KEY_REF                               = "REF";
    private static final String         KEY_RELIGION                          = "RELIGION";
    private static final String         KEY_RESIST                            = "RESIST";
    private static final String         KEY_ROF                               = "ROF";
    private static final String         KEY_ROLL                              = "ROLL";
    private static final String         KEY_RSL                               = "RSL";
    private static final String         KEY_RUNNING_SHOVE                     = "RUNNING_SHOVE";
    private static final String         KEY_SATISFIED                         = "SATISFIED";
    private static final String         KEY_SHIFT_SLIGHTLY                    = "SHIFT_SLIGHTLY";
    private static final String         KEY_SHOTS                             = "SHOTS";
    private static final String         KEY_SHOVE                             = "SHOVE";
    private static final String         KEY_SITUATION                         = "SITUATION";
    private static final String         KEY_SIZE                              = "SIZE";
    private static final String         KEY_SKILL_POINTS                      = "SKILL_POINTS";
    private static final String         KEY_SKILLS_LOOP_END                   = "SKILLS_LOOP_END";
    private static final String         KEY_SKILLS_LOOP_START                 = "SKILLS_LOOP_START";
    private static final String         KEY_SKIN                              = "SKIN";
    private static final String         KEY_SL                                = "SL";
    private static final String         KEY_SPELL_POINTS                      = "SPELL_POINTS";
    private static final String         KEY_SPELLS_LOOP_END                   = "SPELLS_LOOP_END";
    private static final String         KEY_SPELLS_LOOP_START                 = "SPELLS_LOOP_START";
    private static final String         KEY_ST                                = "ST";
    private static final String         KEY_ST_POINTS                         = "ST_POINTS";
    private static final String         KEY_STATE                             = "STATE";
    private static final String         KEY_STYLE_INDENT_WARNING              = "STYLE_INDENT_WARNING";
    private static final String         KEY_SUFFIX_PAREN                      = "_PAREN";
    private static final String         KEY_SUFFIX_BRACKET                    = "_BRACKET";
    private static final String         KEY_SUFFIX_CURLY                      = "_CURLY";
    private static final String         KEY_SWING                             = "SWING";
    private static final String         KEY_TASTE_SMELL                       = "TASTE_SMELL";
    private static final String         KEY_THRUST                            = "THRUST";
    private static final String         KEY_TIME_CAST                         = "TIME_CAST";
    private static final String         KEY_TIRED                             = "TIRED";
    private static final String         KEY_TITLE                             = "TITLE";
    private static final String         KEY_TL                                = "TL";
    private static final String         KEY_TOTAL_POINTS                      = "TOTAL_POINTS";
    private static final String         KEY_TOUCH                             = "TOUCH";
    private static final String         KEY_TWO_HANDED_LIFT                   = "TWO_HANDED_LIFT";
    private static final String         KEY_TYPE                              = "TYPE";
    private static final String         KEY_UNCONSCIOUS                       = "UNCONSCIOUS";
    private static final String         KEY_USAGE                             = "USAGE";
    private static final String         KEY_VISION                            = "VISION";
    private static final String         KEY_WEAPON_STRENGTH                   = "STRENGTH";
    private static final String         KEY_WEAPON_STRENGTH_NUM               = "WEAPON_STRENGTH";
    private static final String         KEY_WEIGHT                            = "WEIGHT";
    private static final String         KEY_WEIGHT_RAW                        = "WEIGHT_RAW";
    private static final String         KEY_WEIGHT_SUMMARY                    = "WEIGHT_SUMMARY";
    private static final String         KEY_WHERE                             = "WHERE";
    private static final String         KEY_WILL                              = "WILL";
    private static final String         KEY_WILL_POINTS                       = "WILL_POINTS";
    private static final String         KEY_AMMO_TYPE                         = "AmmoType:";
    private static final String         KEY_USES_AMMO_TYPE                    = "UsesAmmoType:";
    private static final String         KEY_OPTIONS_CODE                      = "OPTIONS_CODE";
    // TODO: Eliminate these deprecated keys after a suitable waiting period; last added to May 30, 2020
    private static final String         KEY_EARNED_POINTS_DEPRECATED          = "EARNED_POINTS";
    private static final String         KEY_CAMPAIGN_DEPRECATED               = "CAMPAIGN";
    private static final String         KEY_RACE_DEPRECATED                   = "RACE";
    private              CharacterSheet mSheet;
    private              boolean        mEncodeText                           = true;
    private              boolean        mEnhancedKeyParsing;
    private              int            mCurrentId;
    private              int            mStartId;
    private              Set<String>    mOnlyCategories                       = new HashSet<>();
    private              Set<String>    mExcludedCategories                   = new HashSet<>();

    public TextTemplate(CharacterSheet sheet) {
        mSheet = sheet;
    }

    /**
     * @param exportTo The path to save to.
     * @param template The template to use.
     * @return {@code true} on success.
     */
    public boolean export(Path exportTo, Path template) {
        try {
            char[]        buffer           = new char[1];
            boolean       lookForKeyMarker = true;
            StringBuilder keyBuffer        = new StringBuilder();
            try (BufferedReader in = Files.newBufferedReader(template, StandardCharsets.UTF_8)) {
                try (BufferedWriter out = Files.newBufferedWriter(exportTo, StandardCharsets.UTF_8)) {
                    while (in.read(buffer) != -1) {
                        char ch = buffer[0];
                        if (lookForKeyMarker) {
                            if (ch == '@') {
                                lookForKeyMarker = false;
                                in.mark(1);
                            } else {
                                out.append(ch);
                            }
                        } else {
                            if (ch == '_' || Character.isLetterOrDigit(ch)) {
                                keyBuffer.append(ch);
                                in.mark(1);
                            } else {
                                if (!mEnhancedKeyParsing || ch != '@') {
                                    in.reset();        // Allow KEYs to be surrounded by @KEY@
                                }
                                emitKey(in, out, keyBuffer.toString(), exportTo);
                                keyBuffer.setLength(0);
                                lookForKeyMarker = true;
                            }
                        }
                    }
                    if (keyBuffer.length() != 0) {
                        emitKey(in, out, keyBuffer.toString(), exportTo);
                    }
                }
            }
            return true;
        } catch (Exception exception) {
            return false;
        }
    }

    private void emitKey(BufferedReader in, BufferedWriter out, String key, Path base) throws IOException {
        GURPSCharacter gurpsCharacter = mSheet.getCharacter();
        Profile        description    = gurpsCharacter.getProfile();
        switch (key) {
        case KEY_GRID_TEMPLATE:
            out.write(mSheet.getHTMLGridTemplate());
            break;
        case KEY_ENCODING_OFF:
            mEncodeText = false;
            break;
        case KEY_ENHANCED_KEY_PARSING:      // Turn on the ability to enclose a KEY with @.
            mEnhancedKeyParsing = true;     // ex: @KEY@. Useful for when output needs to
            break;                          // be embedded. ex: "<HTML@KEY@TAG>"
        case KEY_PORTRAIT:
            String fileName = PathUtils.enforceExtension(PathUtils.getLeafName(base, false), FileType.PNG.getExtension());
            ImageIO.write(description.getPortrait().getRetina(), "png", base.resolveSibling(fileName).toFile());
            writeEncodedData(out, fileName);
            break;
        case KEY_PORTRAIT_EMBEDDED:
            out.write("data:image/png;base64,");
            ByteArrayOutputStream imgBuffer = new ByteArrayOutputStream();
            OutputStream wrapped = Base64.getEncoder().wrap(imgBuffer);
            ImageIO.write(description.getPortrait().getRetina(), "png", wrapped);
            wrapped.close();
            out.write(imgBuffer.toString(StandardCharsets.UTF_8));
            break;
        case KEY_NAME:
            writeEncodedText(out, description.getName());
            break;
        case KEY_TITLE:
            writeEncodedText(out, description.getTitle());
            break;
        case KEY_RELIGION:
            writeEncodedText(out, description.getReligion());
            break;
        case KEY_PLAYER:
            writeEncodedText(out, description.getPlayerName());
            break;
        case KEY_OPTIONS_CODE:
            writeEncodedText(out, gurpsCharacter.getSettings().optionsCode());
            break;
        case KEY_CREATED_ON:
            writeEncodedText(out, DateTimeFormatter.getFormattedDateTime(gurpsCharacter.getCreatedOn()));
            break;
        case KEY_MODIFIED_ON:
            writeEncodedText(out, DateTimeFormatter.getFormattedDateTime(gurpsCharacter.getModifiedOn()));
            break;
        case KEY_TOTAL_POINTS:
            writeEncodedText(out, Numbers.format(Preferences.getInstance().includeUnspentPointsInTotal() ? gurpsCharacter.getTotalPoints() : gurpsCharacter.getSpentPoints()));
            break;
        case KEY_ATTRIBUTE_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getAttributePoints()));
            break;
        case KEY_ST_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getStrengthPoints()));
            break;
        case KEY_DX_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getDexterityPoints()));
            break;
        case KEY_IQ_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getIntelligencePoints()));
            break;
        case KEY_HT_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getHealthPoints()));
            break;
        case KEY_PERCEPTION_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getPerceptionPoints()));
            break;
        case KEY_WILL_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getWillPoints()));
            break;
        case KEY_FP_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getFatiguePointPoints()));
            break;
        case KEY_HP_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getHitPointPoints()));
            break;
        case KEY_BASIC_SPEED_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getBasicSpeedPoints()));
            break;
        case KEY_BASIC_MOVE_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getBasicMovePoints()));
            break;
        case KEY_ADVANTAGE_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getAdvantagePoints()));
            break;
        case KEY_DISADVANTAGE_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getDisadvantagePoints()));
            break;
        case KEY_QUIRK_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getQuirkPoints()));
            break;
        case KEY_SKILL_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getSkillPoints()));
            break;
        case KEY_SPELL_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getSpellPoints()));
            break;
        case KEY_RACE_POINTS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getRacePoints()));
            break;
        case KEY_UNSPENT_POINTS:
        case KEY_EARNED_POINTS_DEPRECATED:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getUnspentPoints()));
            break;
        case KEY_HEIGHT:
            writeEncodedText(out, description.getHeight().toString());
            break;
        case KEY_HAIR:
            writeEncodedText(out, description.getHair());
            break;
        case KEY_GENDER:
            writeEncodedText(out, description.getGender());
            break;
        case KEY_WEIGHT:
            writeEncodedText(out, EquipmentColumn.getDisplayWeight(gurpsCharacter, description.getWeight()));
            break;
        case KEY_EYES:
            writeEncodedText(out, description.getEyeColor());
            break;
        case KEY_AGE:
            writeEncodedText(out, Numbers.format(description.getAge()));
            break;
        case KEY_SIZE:
            writeEncodedText(out, Numbers.formatWithForcedSign(description.getSizeModifier()));
            break;
        case KEY_SKIN:
            writeEncodedText(out, description.getSkinColor());
            break;
        case KEY_BIRTHDAY:
            writeEncodedText(out, description.getBirthday());
            break;
        case KEY_TL:
            writeEncodedText(out, description.getTechLevel());
            break;
        case KEY_HAND:
            writeEncodedText(out, description.getHandedness());
            break;
        case KEY_ST:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getStrength()));
            break;
        case KEY_DX:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getDexterity()));
            break;
        case KEY_IQ:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getIntelligence()));
            break;
        case KEY_HT:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getHealth()));
            break;
        case KEY_WILL:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getWillAdj()));
            break;
        case KEY_FRIGHT_CHECK:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getFrightCheck()));
            break;
        case KEY_BASIC_SPEED:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getBasicSpeed()));
            break;
        case KEY_BASIC_MOVE:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getBasicMove()));
            break;
        case KEY_PERCEPTION:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getPerAdj()));
            break;
        case KEY_VISION:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getVision()));
            break;
        case KEY_HEARING:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getHearing()));
            break;
        case KEY_TASTE_SMELL:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getTasteAndSmell()));
            break;
        case KEY_TOUCH:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getTouch()));
            break;
        case KEY_THRUST:
            writeEncodedText(out, gurpsCharacter.getThrust().toString());
            break;
        case KEY_SWING:
            writeEncodedText(out, gurpsCharacter.getSwing().toString());
            break;
        case KEY_GENERAL_DR:
            writeEncodedText(out, Numbers.format(((Integer) gurpsCharacter.getValueForID(Armor.ID_TORSO_DR)).intValue()));
            break;
        case KEY_CURRENT_DODGE:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getDodge(gurpsCharacter.getEncumbranceLevel())));
            break;
        case KEY_CURRENT_MOVE:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getMove(gurpsCharacter.getEncumbranceLevel())));
            break;
        case KEY_BEST_CURRENT_PARRY:
            writeBestWeaponDefense(out, (weapon) -> weapon.getResolvedParry());
            break;
        case KEY_BEST_CURRENT_BLOCK:
            writeBestWeaponDefense(out, (weapon) -> weapon.getResolvedBlock());
            break;
        case KEY_FP:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getCurrentFatiguePoints()));
            break;
        case KEY_BASIC_FP:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getFatiguePoints()));
            break;
        case KEY_TIRED:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getTiredFatiguePoints()));
            break;
        case KEY_FP_COLLAPSE:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getUnconsciousChecksFatiguePoints()));
            break;
        case KEY_UNCONSCIOUS:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getUnconsciousFatiguePoints()));
            break;
        case KEY_HP:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getCurrentHitPoints()));
            break;
        case KEY_BASIC_HP:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getHitPointsAdj()));
            break;
        case KEY_REELING:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getReelingHitPoints()));
            break;
        case KEY_HP_COLLAPSE:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getUnconsciousChecksHitPoints()));
            break;
        case KEY_DEATH_CHECK_1:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getDeathCheck1HitPoints()));
            break;
        case KEY_DEATH_CHECK_2:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getDeathCheck2HitPoints()));
            break;
        case KEY_DEATH_CHECK_3:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getDeathCheck3HitPoints()));
            break;
        case KEY_DEATH_CHECK_4:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getDeathCheck4HitPoints()));
            break;
        case KEY_DEAD:
            writeEncodedText(out, Numbers.format(gurpsCharacter.getDeadHitPoints()));
            break;
        case KEY_BASIC_LIFT:
            writeEncodedText(out, gurpsCharacter.getBasicLift().toString());
            break;
        case KEY_ONE_HANDED_LIFT:
            writeEncodedText(out, gurpsCharacter.getOneHandedLift().toString());
            break;
        case KEY_TWO_HANDED_LIFT:
            writeEncodedText(out, gurpsCharacter.getTwoHandedLift().toString());
            break;
        case KEY_SHOVE:
            writeEncodedText(out, gurpsCharacter.getShoveAndKnockOver().toString());
            break;
        case KEY_RUNNING_SHOVE:
            writeEncodedText(out, gurpsCharacter.getRunningShoveAndKnockOver().toString());
            break;
        case KEY_CARRY_ON_BACK:
            writeEncodedText(out, gurpsCharacter.getCarryOnBack().toString());
            break;
        case KEY_SHIFT_SLIGHTLY:
            writeEncodedText(out, gurpsCharacter.getShiftSlightly().toString());
            break;
        case KEY_CARRIED_WEIGHT:
            writeEncodedText(out, EquipmentColumn.getDisplayWeight(gurpsCharacter, gurpsCharacter.getWeightCarried()));
            break;
        case KEY_CARRIED_VALUE:
            writeEncodedText(out, "$" + gurpsCharacter.getWealthCarried().toLocalizedString());
            break;
        case KEY_OTHER_VALUE:
            writeEncodedText(out, "$" + gurpsCharacter.getWealthNotCarried().toLocalizedString());
            break;
        case KEY_NOTES:
            StringBuilder buffer = new StringBuilder();
            for (Note note : gurpsCharacter.getNoteIterator()) {
                if (buffer.length() > 0) {
                    buffer.append("\n\n");
                }
                buffer.append(note.getDescription());
            }
            writeEncodedText(out, buffer.toString());
            break;
        case KEY_CONTINUE_ID:
            mStartId = mCurrentId;
            break;
        case KEY_RACE_DEPRECATED:
        case KEY_CAMPAIGN_DEPRECATED:
            break;
        default:
            if (key.startsWith(KEY_ENCUMBRANCE_LOOP_START)) {
                processEncumbranceLoop(out, extractUpToMarker(in, KEY_ENCUMBRANCE_LOOP_END));
            } else if (key.startsWith(KEY_HIT_LOCATION_LOOP_START)) {
                processHitLocationLoop(out, extractUpToMarker(in, KEY_HIT_LOCATION_LOOP_END));
            } else if (key.startsWith(KEY_ADVANTAGES_LOOP_START)) {
                processAdvantagesLoop(out, extractUpToMarker(in, KEY_ADVANTAGES_LOOP_END), AdvantagesLoopType.ALL);
            } else if (key.startsWith(KEY_ADVANTAGES_ALL_LOOP_START)) {
                processAdvantagesLoop(out, extractUpToMarker(in, KEY_ADVANTAGES_ALL_LOOP_END), AdvantagesLoopType.ADS_ALL);
            } else if (key.startsWith(KEY_ADVANTAGES_ONLY_LOOP_START)) {
                processAdvantagesLoop(out, extractUpToMarker(in, KEY_ADVANTAGES_ONLY_LOOP_END), AdvantagesLoopType.ADS);
            } else if (key.startsWith(KEY_DISADVANTAGES_LOOP_START)) {
                processAdvantagesLoop(out, extractUpToMarker(in, KEY_DISADVANTAGES_LOOP_END), AdvantagesLoopType.DISADS);
            } else if (key.startsWith(KEY_DISADVANTAGES_ALL_LOOP_START)) {
                processAdvantagesLoop(out, extractUpToMarker(in, KEY_DISADVANTAGES_ALL_LOOP_END), AdvantagesLoopType.DISADS_ALL);
            } else if (key.startsWith(KEY_QUIRKS_LOOP_START)) {
                processAdvantagesLoop(out, extractUpToMarker(in, KEY_QUIRKS_LOOP_END), AdvantagesLoopType.QUIRKS);
            } else if (key.startsWith(KEY_PERKS_LOOP_START)) {
                processAdvantagesLoop(out, extractUpToMarker(in, KEY_PERKS_LOOP_END), AdvantagesLoopType.PERKS);
            } else if (key.startsWith(KEY_LANGUAGES_LOOP_START)) {
                processAdvantagesLoop(out, extractUpToMarker(in, KEY_LANGUAGES_LOOP_END), AdvantagesLoopType.LANGUAGES);
            } else if (key.startsWith(KEY_CULTURAL_FAMILIARITIES_LOOP_START)) {
                processAdvantagesLoop(out, extractUpToMarker(in, KEY_CULTURAL_FAMILIARITIES_LOOP_END), AdvantagesLoopType.CULTURAL_FAMILIARITIES);
            } else if (key.startsWith(KEY_SKILLS_LOOP_START)) {
                processSkillsLoop(out, extractUpToMarker(in, KEY_SKILLS_LOOP_END));
            } else if (key.startsWith(KEY_SPELLS_LOOP_START)) {
                processSpellsLoop(out, extractUpToMarker(in, KEY_SPELLS_LOOP_END));
            } else if (key.startsWith(KEY_MELEE_LOOP_START)) {
                processMeleeLoop(out, extractUpToMarker(in, KEY_MELEE_LOOP_END));
            } else if (key.startsWith(KEY_HIERARCHICAL_MELEE_LOOP_START)) {
                processHierarchicalMeleeLoop(out, extractUpToMarker(in, KEY_HIERARCHICAL_MELEE_LOOP_END));
            } else if (key.startsWith(KEY_RANGED_LOOP_START)) {
                processRangedLoop(out, extractUpToMarker(in, KEY_RANGED_LOOP_END));
            } else if (key.startsWith(KEY_HIERARCHICAL_RANGED_LOOP_START)) {
                processHierarchicalRangedLoop(out, extractUpToMarker(in, KEY_HIERARCHICAL_RANGED_LOOP_END));
            } else if (key.startsWith(KEY_EQUIPMENT_LOOP_START)) {
                processEquipmentLoop(out, extractUpToMarker(in, KEY_EQUIPMENT_LOOP_END), true);
            } else if (key.startsWith(KEY_OTHER_EQUIPMENT_LOOP_START)) {
                processEquipmentLoop(out, extractUpToMarker(in, KEY_OTHER_EQUIPMENT_LOOP_END), false);
            } else if (key.startsWith(KEY_NOTES_LOOP_START)) {
                processNotesLoop(out, extractUpToMarker(in, KEY_NOTES_LOOP_END));
            } else if (key.startsWith(KEY_REACTION_LOOP_START)) {
                processReactionLoop(out, extractUpToMarker(in, KEY_REACTION_LOOP_END));
            } else if (key.startsWith(KEY_ONLY_CATEGORIES)) {
                setOnlyCategories(key);
            } else if (key.startsWith(KEY_EXCLUDE_CATEGORIES)) {
                setExcludeCategories(key);
            } else {
                writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
            }
            break;
        }
    }

    private void setOnlyCategories(String key) {
        String[] categories = key.substring(KEY_ONLY_CATEGORIES.length()).split(UNDERSCORE);
        mOnlyCategories.addAll(Arrays.asList(categories));
    }

    private void setExcludeCategories(String key) {
        String[] categories = key.substring(KEY_EXCLUDE_CATEGORIES.length()).split(UNDERSCORE);
        mExcludedCategories.addAll(Arrays.asList(categories));
    }

    private void writeBestWeaponDefense(BufferedWriter out, Function<MeleeWeaponStats, String> resolver) throws IOException {
        String best      = "-";
        int    bestValue = Integer.MIN_VALUE;
        for (WeaponDisplayRow row : new FilteredIterator<>(mSheet.getMeleeWeaponOutline().getModel().getRows(), WeaponDisplayRow.class)) {
            MeleeWeaponStats weapon = (MeleeWeaponStats) row.getWeapon();
            String           result = resolver.apply(weapon).trim();
            if (!result.isEmpty() && !"No".equals(result)) {
                int value = Numbers.extractInteger(result, 0, false);
                if (value > bestValue) {
                    bestValue = value;
                    best = result;
                }
            }
        }
        writeEncodedText(out, best);
    }

    private void writeEncodedData(BufferedWriter out, String text) throws IOException {
        out.write(mEncodeText ? XMLWriter.encodeData(text).replaceAll(" ", "%20") : text);
    }

    private void writeEncodedText(BufferedWriter out, String text) throws IOException {
        out.write(mEncodeText ? XMLWriter.encodeData(text).replaceAll("&#10;", "<br>").replaceAll("\"", "&quot;") : text);
    }

    private String extractUpToMarker(BufferedReader in, String marker) throws IOException {
        char[]        buffer           = new char[1];
        StringBuilder keyBuffer        = new StringBuilder();
        StringBuilder extraction       = new StringBuilder();
        boolean       lookForKeyMarker = true;
        while (in.read(buffer) != -1) {
            char ch = buffer[0];
            if (lookForKeyMarker) {
                if (ch == '@') {
                    lookForKeyMarker = false;
                    in.mark(1);
                } else {
                    extraction.append(ch);
                }
            } else {
                if (ch == '_' || Character.isLetterOrDigit(ch)) {
                    keyBuffer.append(ch);
                    in.mark(1);
                } else {
                    if (!mEnhancedKeyParsing || ch != '@') {
                        in.reset();        // Allow KEYs to be surrounded by @KEY@
                    }
                    String key = keyBuffer.toString();
                    in.reset();
                    if (key.equals(marker)) {
                        return extraction.toString();
                    }
                    extraction.append('@');
                    extraction.append(key);
                    keyBuffer.setLength(0);
                    lookForKeyMarker = true;
                }
            }
        }
        return extraction.toString();
    }

    private void processEncumbranceLoop(BufferedWriter out, String contents) throws IOException {
        GURPSCharacter gurpsCharacter   = mSheet.getCharacter();
        int            length           = contents.length();
        StringBuilder  keyBuffer        = new StringBuilder();
        boolean        lookForKeyMarker = true;
        for (Encumbrance encumbrance : Encumbrance.values()) {
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        switch (key) {
                        case KEY_CURRENT_MARKER:
                            if (encumbrance == gurpsCharacter.getEncumbranceLevel()) {
                                out.write(CURRENT);
                            }
                            break;
                        case KEY_CURRENT_MARKER_1:
                            if (encumbrance == gurpsCharacter.getEncumbranceLevel()) {
                                out.write(ONE);
                            }
                            break;
                        case KEY_CURRENT_MARKER_BULLET:
                            if (encumbrance == gurpsCharacter.getEncumbranceLevel()) {
                                out.write("•");
                            }
                            break;
                        case KEY_LEVEL:
                            writeEncodedText(out, MessageFormat.format(encumbrance == gurpsCharacter.getEncumbranceLevel() ? "• {0} ({1})" : "{0} ({1})", encumbrance, Numbers.format(-encumbrance.getEncumbrancePenalty())));
                            break;
                        case KEY_LEVEL_NO_MARKER:
                            writeEncodedText(out, MessageFormat.format("{0} ({1})", encumbrance, Numbers.format(-encumbrance.getEncumbrancePenalty())));
                            break;
                        case KEY_LEVEL_ONLY:
                            writeEncodedText(out, Numbers.format(-encumbrance.getEncumbrancePenalty()));
                            break;
                        case KEY_MAX_LOAD:
                            writeEncodedText(out, gurpsCharacter.getMaximumCarry(encumbrance).toString());
                            break;
                        case KEY_MOVE:
                            writeEncodedText(out, Numbers.format(gurpsCharacter.getMove(encumbrance)));
                            break;
                        case KEY_DODGE:
                            writeEncodedText(out, Numbers.format(gurpsCharacter.getDodge(encumbrance)));
                            break;
                        default:
                            writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
                            break;
                        }
                    }
                }
            }
        }
    }

    private void processHitLocationLoop(BufferedWriter out, String contents) throws IOException {
        GURPSCharacter gurpsCharacter   = mSheet.getCharacter();
        int            length           = contents.length();
        StringBuilder  keyBuffer        = new StringBuilder();
        boolean        lookForKeyMarker = true;
        mCurrentId = mStartId;
        HitLocationTable table = gurpsCharacter.getProfile().getHitLocationTable();
        for (HitLocationTableEntry entry : table.getEntries()) {
            mCurrentId++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        switch (key) {
                        case KEY_ROLL:
                            writeEncodedText(out, entry.getRoll());
                            break;
                        case KEY_WHERE:
                            writeEncodedText(out, entry.getName());
                            break;
                        case KEY_PENALTY:
                            writeEncodedText(out, Numbers.format(entry.getHitPenalty()));
                            break;
                        case KEY_DR:
                            writeEncodedText(out, Numbers.format(((Integer) gurpsCharacter.getValueForID(entry.getKey())).intValue()));
                            break;
                        case KEY_ID:
                            writeEncodedText(out, Integer.toString(mCurrentId));
                            break;
                        case KEY_EQUIPMENT:     // Show the equipment that is providing the DR bonus
                            writeEncodedText(out, hitLocationEquipment(entry).replace(NEWLINE, COMMA_SEPARATOR));
                            break;
                        case KEY_EQUIPMENT_FORMATTED:
                            String loc = hitLocationEquipment(entry);
                            if (!loc.isEmpty()) {
                                writeEncodedText(out, PARAGRAPH_START + loc.replace(NEWLINE, PARAGRAPH_END + NEWLINE + PARAGRAPH_START) + PARAGRAPH_END);
                            }
                            break;
                        default:
                            writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
                            break;
                        }
                    }
                }
            }
        }
        mStartId = 0;
    }

    /* A kludgy method to relate a hitlocation to the armor that is providing the DR for that hit location. */
    private String hitLocationEquipment(HitLocationTableEntry entry) {
        StringBuilder sb    = new StringBuilder();
        boolean       first = true;
        for (Equipment equipment : mSheet.getCharacter().getEquipmentIterator()) {
            if (equipment.isEquipped()) {
                for (Feature feature : equipment.getFeatures()) {
                    if (feature instanceof DRBonus) {
                        String locationKey = entry.getLocation().getKey();
                        if (locationKey.equals(HitLocation.VITALS.getKey())) {
                            // Assume Vitals uses the same equipment as Torso
                            locationKey = HitLocation.TORSO.getKey();
                        }
                        if (locationKey.endsWith(((DRBonus) feature).getLocation().name())) {
                            // HUGE Kludge. Only way I could equate the 2
                            // different HitLocations. I know that one is derived
                            // from the other, so this check will ALWAYS work.
                            if (!first) {
                                sb.append("\n");
                            }
                            sb.append(equipment.getDescription());
                            first = false;
                        }
                    }
                }
            }
        }
        return sb.toString();
    }

    private void processAdvantagesLoop(BufferedWriter out, String contents, AdvantagesLoopType loopType) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        for (Advantage advantage : mSheet.getCharacter().getAdvantagesIterator(false)) {
            if (loopType.shouldInclude(advantage, mOnlyCategories, mExcludedCategories)) {
                mCurrentId++;
                for (int i = 0; i < length; i++) {
                    char ch = contents.charAt(i);
                    if (lookForKeyMarker) {
                        if (ch == '@') {
                            lookForKeyMarker = false;
                        } else {
                            out.append(ch);
                        }
                    } else {
                        if (ch == '_' || Character.isLetterOrDigit(ch)) {
                            keyBuffer.append(ch);
                        } else {
                            String key = keyBuffer.toString();
                            i--;
                            if (mEnhancedKeyParsing && ch == '@') {
                                i++;        // Allow KEYs to be surrounded with @, e.g. @KEY@
                            }
                            keyBuffer.setLength(0);
                            lookForKeyMarker = true;
                            if (!processStyleIndentWarning(key, out, advantage)) {
                                if (!processDescription(key, out, advantage)) {
                                    switch (key) {
                                    case KEY_POINTS:
                                        writeEncodedText(out, AdvantageColumn.POINTS.getDataAsText(advantage));
                                        break;
                                    case KEY_REF:
                                        writeEncodedText(out, AdvantageColumn.REFERENCE.getDataAsText(advantage));
                                        break;
                                    case KEY_ID:
                                        writeEncodedText(out, Integer.toString(mCurrentId));
                                        break;
                                    case KEY_TYPE:
                                        writeEncodedText(out, advantage.canHaveChildren() ? advantage.getContainerType().name() : ITEM);
                                        break;
                                    case KEY_DESCRIPTION_USER:
                                        writeEncodedText(out, advantage.getUserDesc());
                                        break;
                                    case KEY_DESCRIPTION_USER_FORMATTED:
                                        if (!advantage.getUserDesc().isEmpty()) {
                                            writeEncodedText(out, PARAGRAPH_START + advantage.getUserDesc().replace(NEWLINE, PARAGRAPH_END + NEWLINE + PARAGRAPH_START) + PARAGRAPH_END);
                                        }
                                        break;
                                    default:
                                        /* Allows the access to notes on modifiers.  Currently only used in the 'Language' loop.
                                         * e.g. Advantage:Language, Modifier:Spoken -> Note:Native, Advantage:Language, Modifier:Written -> Note:Accented
                                         */
                                        if (key.startsWith(KEY_MODIFIER_NOTES_FOR)) {
                                            AdvantageModifier m = advantage.getActiveModifierFor(key.substring(KEY_MODIFIER_NOTES_FOR.length()));
                                            if (m != null) {
                                                writeEncodedText(out, m.getNotes());
                                            }
                                        } else {
                                            writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
                                        }
                                        break;
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
        mOnlyCategories.clear();
        mExcludedCategories.clear();
        mStartId = 0;
    }

    private boolean processDescription(String key, BufferedWriter out, ListRow row) throws IOException {
        if (key.equals(KEY_DESCRIPTION)) {
            writeEncodedText(out, row.toString());
            writeNote(out, row.getModifierNotes());
            writeNote(out, row.getNotes());
        } else if (key.equals(KEY_DESCRIPTION_PRIMARY)) {
            writeEncodedText(out, row.toString());
        } else if (key.startsWith(KEY_DESCRIPTION_MODIFIER_NOTES)) {
            writeXMLTextWithOptionalParens(key, out, row.getModifierNotes());
        } else if (key.startsWith(KEY_DESCRIPTION_NOTES)) {
            writeXMLTextWithOptionalParens(key, out, row.getNotes());
        } else {
            return false;
        }
        return true;
    }

    private void writeXMLTextWithOptionalParens(String key, BufferedWriter out, String text) throws IOException {
        if (!text.isEmpty()) {
            String pre  = "";
            String post = "";
            if (key.endsWith(KEY_SUFFIX_PAREN)) {
                pre = " (";
                post = ")";
            } else if (key.endsWith(KEY_SUFFIX_BRACKET)) {
                pre = " [";
                post = "]";
            } else if (key.endsWith(KEY_SUFFIX_CURLY)) {
                pre = " {";
                post = "}";
            }
            out.write(pre);
            writeEncodedText(out, text);
            out.write(post);
        }
    }

    private void writeNote(BufferedWriter out, String notes) throws IOException {
        if (!notes.isEmpty()) {
            out.write("<div class=\"note\">");
            writeEncodedText(out, notes);
            out.write("</div>");
        }
    }

    private void processSkillsLoop(BufferedWriter out, String contents) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        for (Skill skill : mSheet.getCharacter().getSkillsIterator()) {
            mCurrentId++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        if (!processStyleIndentWarning(key, out, skill)) {
                            if (!processDescription(key, out, skill)) {
                                switch (key) {
                                case KEY_SL:
                                    writeEncodedText(out, SkillColumn.LEVEL.getDataAsText(skill));
                                    break;
                                case KEY_RSL:
                                    writeEncodedText(out, SkillColumn.RELATIVE_LEVEL.getDataAsText(skill));
                                    break;
                                case KEY_DIFFICULTY:
                                    writeEncodedText(out, SkillColumn.DIFFICULTY.getDataAsText(skill));
                                    break;
                                case KEY_POINTS:
                                    writeEncodedText(out, SkillColumn.POINTS.getDataAsText(skill));
                                    break;
                                case KEY_REF:
                                    writeEncodedText(out, SkillColumn.REFERENCE.getDataAsText(skill));
                                    break;
                                case KEY_ID:
                                    writeEncodedText(out, Integer.toString(mCurrentId));
                                    break;
                                default:
                                    writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
                                    break;
                                }
                            }
                        }
                    }
                }
            }
        }
        mStartId = 0;
    }

    private static boolean processStyleIndentWarning(String key, BufferedWriter out, ListRow row) throws IOException {
        if (key.equals(KEY_STYLE_INDENT_WARNING)) {
            StringBuilder style = new StringBuilder();
            int           depth = row.getDepth();
            if (depth > 0) {
                style.append(" style=\"padding-left: ");
                style.append(depth * 12);
                style.append("px;");
            }
            if (!row.isSatisfied()) {
                if (style.length() == 0) {
                    style.append(" style=\"");
                }
                style.append(" color: red;");
            }
            if (style.length() > 0) {
                style.append("\" ");
                out.write(style.toString());
            }
        } else if (key.startsWith(KEY_PREFIX_DEPTH)) {
            int amt = Numbers.extractInteger(key.substring(6), 1, false);
            out.write("" + amt * row.getDepth());
        } else if (key.equals(KEY_SATISFIED)) {
            out.write(row.isSatisfied() ? "Y" : "N");
        } else {
            return false;
        }
        return true;
    }

    private void processSpellsLoop(BufferedWriter out, String contents) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        for (Spell spell : mSheet.getCharacter().getSpellsIterator()) {
            mCurrentId++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        if (!processStyleIndentWarning(key, out, spell)) {
                            if (!processDescription(key, out, spell)) {
                                switch (key) {
                                case KEY_CLASS:
                                    writeEncodedText(out, spell.getSpellClass());
                                    break;
                                case KEY_COLLEGE:
                                    writeEncodedText(out, spell.getCollege());
                                    break;
                                case KEY_MANA_CAST:
                                    writeEncodedText(out, spell.getCastingCost());
                                    break;
                                case KEY_MANA_MAINTAIN:
                                    writeEncodedText(out, spell.getMaintenance());
                                    break;
                                case KEY_TIME_CAST:
                                    writeEncodedText(out, spell.getCastingTime());
                                    break;
                                case KEY_DURATION:
                                    writeEncodedText(out, spell.getDuration());
                                    break;
                                case KEY_RESIST:
                                    writeEncodedText(out, spell.getResist());
                                    break;
                                case KEY_SL:
                                    writeEncodedText(out, SpellColumn.LEVEL.getDataAsText(spell));
                                    break;
                                case KEY_RSL:
                                    writeEncodedText(out, SpellColumn.RELATIVE_LEVEL.getDataAsText(spell));
                                    break;
                                case KEY_DIFFICULTY:
                                    writeEncodedText(out, spell.getDifficultyAsText());
                                    break;
                                case KEY_POINTS:
                                    writeEncodedText(out, SpellColumn.POINTS.getDataAsText(spell));
                                    break;
                                case KEY_REF:
                                    writeEncodedText(out, SpellColumn.REFERENCE.getDataAsText(spell));
                                    break;
                                case KEY_ID:
                                    writeEncodedText(out, Integer.toString(mCurrentId));
                                    break;
                                default:
                                    writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
                                    break;
                                }
                            }
                        }
                    }
                }
            }
        }
        mStartId = 0;
    }

    private void processMeleeLoop(BufferedWriter out, String contents) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        for (WeaponDisplayRow row : new FilteredIterator<>(mSheet.getMeleeWeaponOutline().getModel().getRows(), WeaponDisplayRow.class)) {
            mCurrentId++;
            MeleeWeaponStats weapon = (MeleeWeaponStats) row.getWeapon();
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        i = processMeleeWeaponKeys(out, key, mCurrentId, weapon, i, contents, null);
                    }
                }
            }
        }
        mStartId = 0;
    }

    /* Handle keys specific to MeleeWeaponStats.   If "attackModes" is NOT NULL, then we could allow processing of a hierarchical loop  */
    private int processMeleeWeaponKeys(BufferedWriter out, String key, int counter, MeleeWeaponStats weapon, int index, String contents, List<MeleeWeaponStats> attackModes) throws IOException {
        switch (key) {
        case KEY_PARRY:
            writeEncodedText(out, weapon.getResolvedParry());
            return index;
        case KEY_BLOCK:
            writeEncodedText(out, weapon.getResolvedBlock());
            return index;
        case KEY_REACH:
            writeEncodedText(out, weapon.getReach());
            return index;
        default:
            if (attackModes != null && key.startsWith(KEY_ATTACK_MODES_LOOP_START)) {
                int endIndex = contents.indexOf(KEY_ATTACK_MODES_LOOP_END);
                if (endIndex > 0) {
                    String subContents = contents.substring(index + 1, endIndex - 1);
                    processMeleeAttackModes(out, subContents, attackModes);
                    return endIndex + KEY_ATTACK_MODES_LOOP_END.length();
                }
            }
            return processWeaponKeys(out, key, counter, weapon, index);
        }
    }

    /* Handle keys specific to RangedWeaponStats.   If "attackModes" is NOT NULL, then we could allow processing of a hierarchical loop  */
    private int processRangedWeaponKeys(BufferedWriter out, String key, int counter, RangedWeaponStats weapon, int index, String contents, List<RangedWeaponStats> attackModes) throws IOException {
        switch (key) {
        case KEY_BULK:
            writeEncodedText(out, weapon.getBulk());
            return index;
        case KEY_ACCURACY:
            writeEncodedText(out, weapon.getAccuracy());
            return index;
        case KEY_RANGE:
            writeEncodedText(out, weapon.getRange());
            return index;
        case KEY_ROF:
            writeEncodedText(out, weapon.getRateOfFire());
            return index;
        case KEY_SHOTS:
            writeEncodedText(out, weapon.getShots());
            return index;
        case KEY_RECOIL:
            writeEncodedText(out, weapon.getRecoil());
            return index;
        default:
            if (attackModes != null && key.startsWith(KEY_ATTACK_MODES_LOOP_START)) {
                int endIndex = contents.indexOf(KEY_ATTACK_MODES_LOOP_END);
                if (endIndex > 0) {
                    String subContents = contents.substring(index + 1, endIndex - 1);
                    processRangedAttackModes(out, subContents, attackModes);
                    return endIndex + KEY_ATTACK_MODES_LOOP_END.length();
                }
            }
            return processWeaponKeys(out, key, counter, weapon, index);
        }
    }

    /* Break out handling of general weapons information. Anything known by WeaponStats or the equipment.  */
    private int processWeaponKeys(BufferedWriter out, String key, int counter, WeaponStats weapon, int index) throws IOException {
        Equipment equipment = null;
        if (weapon.getOwner() instanceof Equipment) {
            equipment = (Equipment) weapon.getOwner();
        }
        if (!processDescription(key, out, weapon)) {
            switch (key) {
            case KEY_USAGE:
                writeEncodedText(out, weapon.getUsage());
                break;
            case KEY_LEVEL:
                writeEncodedText(out, Numbers.format(weapon.getSkillLevel()));
                break;
            case KEY_DAMAGE:
                writeEncodedText(out, weapon.getDamage().getResolvedDamage());
                break;
            case KEY_UNMODIFIED_DAMAGE:
                writeEncodedText(out, weapon.getDamage().toString());
                break;
            case KEY_WEAPON_STRENGTH:
                writeEncodedText(out, weapon.getStrength());
                break;
            case KEY_WEAPON_STRENGTH_NUM:
                writeEncodedText(out, weapon.getStrength().replaceAll("[^0-9]", ""));
                break;
            case KEY_ID:
                writeEncodedText(out, Integer.toString(counter));
                break;
            case KEY_COST:
                if (equipment != null) {
                    writeEncodedText(out, equipment.getAdjustedValue().toLocalizedString());
                }
                break;
            case KEY_LEGALITY_CLASS:
                if (equipment != null) {
                    writeEncodedText(out, equipment.getLegalityClass());
                }
                break;
            case KEY_TL:
                if (equipment != null) {
                    writeEncodedText(out, equipment.getTechLevel());
                }
                break;
            case KEY_WEIGHT:
                if (equipment != null) {
                    writeEncodedText(out, EquipmentColumn.getDisplayWeight(equipment.getDataFile(), equipment.getAdjustedWeight()));
                }
                break;
            case KEY_AMMO:
                if (equipment != null) {
                    writeEncodedText(out, Numbers.format(ammoFor(equipment)));
                }
                break;
            default:
                writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
                break;
            }
        }
        return index;
    }

    /* Process the weapons in a hierarchical format.   One time for each weapon with a unique name,
     * and then possibly one time for each different "attack mode" that the weapon can support.
     * e.g. Weapon Name: Spear, attack modes "1 Handed" and "2 Handed"
     */
    private void processHierarchicalMeleeLoop(BufferedWriter out, String contents) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        Map<String, ArrayList<MeleeWeaponStats>> weaponsMap = new HashMap<>();
        Map<String, MeleeWeaponStats>            weapons    = new HashMap<>();
        for (WeaponDisplayRow row : new FilteredIterator<>(mSheet.getMeleeWeaponOutline().getModel().getRows(), WeaponDisplayRow.class)) {
            MeleeWeaponStats weapon = (MeleeWeaponStats) row.getWeapon();
            weapons.put(weapon.getDescription(), weapon);
            ArrayList<MeleeWeaponStats> attackModes = weaponsMap.get(weapon.getDescription());
            if (attackModes == null) {
                attackModes = new ArrayList<>();
                weaponsMap.put(weapon.getDescription(), attackModes);
            }
            attackModes.add(weapon);
        }
        for (MeleeWeaponStats weapon : weapons.values()) {
            mCurrentId++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        i = processMeleeWeaponKeys(out, key, mCurrentId, weapon, i, contents, weaponsMap.get(weapon.getDescription()));
                    }
                }
            }
        }
        mStartId = 0;
    }

    /* Process the weapons in a hierarchical format.   One time for each weapon with a unique name,
     * and then possibly one time for each different "attack mode" that the weapon can support.
     * e.g. Weapon Name: Atlatl, attack modes "Shoot Dart" and "Shoot Javelin"
     */
    private void processHierarchicalRangedLoop(BufferedWriter out, String contents) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        Map<String, ArrayList<RangedWeaponStats>> weaponsMap = new HashMap<>();
        Map<String, RangedWeaponStats>            weapons    = new HashMap<>();
        for (WeaponDisplayRow row : new FilteredIterator<>(mSheet.getRangedWeaponOutline().getModel().getRows(), WeaponDisplayRow.class)) {
            RangedWeaponStats weapon = (RangedWeaponStats) row.getWeapon();
            weapons.put(weapon.getDescription(), weapon);
            ArrayList<RangedWeaponStats> attackModes = weaponsMap.get(weapon.getDescription());
            if (attackModes == null) {
                attackModes = new ArrayList<>();
                weaponsMap.put(weapon.getDescription(), attackModes);
            }
            attackModes.add(weapon);
        }
        for (RangedWeaponStats weapon : weapons.values()) {
            mCurrentId++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        i = processRangedWeaponKeys(out, key, mCurrentId, weapon, i, contents, weaponsMap.get(weapon.getDescription()));
                    }
                }
            }
        }
        mStartId = 0;
    }

    /* Loop through all of the attackModes for a particular weapon.   We need to make melee/ranged specific
     * versions of this method because they must call the correct "processXXWeaponKeys" method.
     */
    private void processMeleeAttackModes(BufferedWriter out, String contents, List<MeleeWeaponStats> attackModes) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        int           counter          = 0;
        for (MeleeWeaponStats weapon : attackModes) {
            counter++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        i = processMeleeWeaponKeys(out, key, counter, weapon, i, contents, null);
                    }
                }
            }
        }
    }

    /* Loop through all of the attackModes for a particular weapon.   We need to make melee/ranged specific
     * versions of this method because they must call the correct "processXXWeaponKeys" method.
     */
    private void processRangedAttackModes(BufferedWriter out, String contents, List<RangedWeaponStats> attackModes) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        int           counter          = 0;
        for (RangedWeaponStats weapon : attackModes) {
            counter++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        i = processRangedWeaponKeys(out, key, counter, weapon, i, contents, null);
                    }
                }
            }
        }
    }

    /* report the quantity of ammo used by this weapon, if any */
    private int ammoFor(Equipment weapon) {
        String usesAmmoType = null;
        String ammoType     = null;
        for (String category : weapon.getCategories()) {
            if (category.startsWith(KEY_USES_AMMO_TYPE)) {
                usesAmmoType = category.substring(KEY_USES_AMMO_TYPE.length()).trim();
            }
        }
        if (usesAmmoType == null) {
            return 0;
        }
        for (Equipment equipment : mSheet.getCharacter().getEquipmentIterator()) {
            if (equipment.isEquipped()) {
                for (String category : equipment.getCategories()) {
                    if (category.startsWith(KEY_AMMO_TYPE)) {
                        ammoType = category.substring(KEY_AMMO_TYPE.length()).trim();
                    }
                }
                if (usesAmmoType.equalsIgnoreCase(ammoType)) {
                    return equipment.getQuantity();
                }
            }
        }
        return 0;
    }

    private boolean processDescription(String key, BufferedWriter out, WeaponStats stats) throws IOException {
        if (key.equals(KEY_DESCRIPTION)) {
            writeEncodedText(out, stats.toString());
            writeNote(out, stats.getNotes());
        } else if (key.equals(KEY_DESCRIPTION_PRIMARY)) {
            writeEncodedText(out, stats.toString());
        } else if (key.startsWith(KEY_DESCRIPTION_NOTES)) {
            writeXMLTextWithOptionalParens(key, out, stats.getNotes());
        } else {
            return false;
        }
        return true;
    }

    private void processRangedLoop(BufferedWriter out, String contents) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        for (WeaponDisplayRow row : new FilteredIterator<>(mSheet.getRangedWeaponOutline().getModel().getRows(), WeaponDisplayRow.class)) {
            mCurrentId++;
            RangedWeaponStats weapon = (RangedWeaponStats) row.getWeapon();
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        i = processRangedWeaponKeys(out, key, mCurrentId, weapon, i, contents, null);
                    }
                }
            }
        }
        mStartId = 0;
    }

    private void processEquipmentLoop(BufferedWriter out, String contents, boolean carried) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        // Create child-to-parent maps to determine where items are being stored.
        // Used by KEY_LOCATION
        List<List<Row>>        children      = new ArrayList<>();
        List<Equipment>        parents       = new ArrayList<>();
        List<Equipment>        equipmentList = new ArrayList<>();
        RowIterator<Equipment> iter          = carried ? mSheet.getCharacter().getEquipmentIterator() : mSheet.getCharacter().getOtherEquipmentIterator();
        for (Equipment equipment : iter) {
            if (shouldInclude(equipment)) {   // Allows category filtering
                equipmentList.add(equipment);
                if (equipment.hasChildren()) {
                    children.add(equipment.getChildren());
                    parents.add(equipment);
                }
            }
        }
        for (Equipment equipment : equipmentList) {
            mCurrentId++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        if (!processStyleIndentWarning(key, out, equipment)) {
                            if (!processDescription(key, out, equipment)) {
                                switch (key) {
                                case KEY_STATE:
                                    if (carried) {
                                        out.write(equipment.isEquipped() ? "E" : "C");
                                    } else {
                                        out.write("-");
                                    }
                                    break;
                                case KEY_EQUIPPED:
                                    if (carried && equipment.isEquipped()) {
                                        out.write("✓");
                                    }
                                    break;
                                case KEY_EQUIPPED_NUM:
                                    out.write(carried && equipment.isEquipped() ? '1' : '0');
                                    break;
                                case KEY_CARRIED_STATUS:
                                    if (carried) {
                                        out.write(equipment.isEquipped() ? '2' : '1');
                                    } else {
                                        out.write('0');
                                    }
                                    break;
                                case KEY_QTY:
                                    writeEncodedText(out, Numbers.format(equipment.getQuantity()));
                                    break;
                                case KEY_COST:
                                    writeEncodedText(out, equipment.getAdjustedValue().toLocalizedString());
                                    break;
                                case KEY_WEIGHT:
                                    writeEncodedText(out, EquipmentColumn.getDisplayWeight(equipment.getDataFile(), equipment.getAdjustedWeight()));
                                    break;
                                case KEY_COST_SUMMARY:
                                    writeEncodedText(out, equipment.getExtendedValue().toLocalizedString());
                                    break;
                                case KEY_WEIGHT_SUMMARY:
                                    writeEncodedText(out, EquipmentColumn.getDisplayWeight(equipment.getDataFile(), equipment.getExtendedWeight()));
                                    break;
                                case KEY_WEIGHT_RAW:
                                    writeEncodedText(out, equipment.getAdjustedWeight().getNormalizedValue().toLocalizedString());
                                    break;
                                case KEY_REF:
                                    writeEncodedText(out, equipment.getReference());
                                    break;
                                case KEY_ID:
                                    writeEncodedText(out, Integer.toString(mCurrentId));
                                    break;
                                case KEY_TL:
                                    writeEncodedText(out, equipment.getTechLevel());
                                    break;
                                case KEY_LEGALITY_CLASS:
                                    writeEncodedText(out, equipment.getDisplayLegalityClass());
                                    break;
                                case KEY_CATEGORIES:
                                    writeEncodedText(out, equipment.getCategoriesAsString());
                                    break;
                                case KEY_LOCATION:
                                    for (int j = 0; j < children.size(); j++) {
                                        if (children.get(j).contains(equipment)) {
                                            writeEncodedText(out, parents.get(j).getDescription());
                                        }
                                    }
                                    break;
                                default:
                                    if (key.startsWith(KEY_MODIFIER_NOTES_FOR)) {
                                        EquipmentModifier m = equipment.getActiveModifierFor(key.substring(KEY_MODIFIER_NOTES_FOR.length()));
                                        if (m != null) {
                                            writeEncodedText(out, m.getNotes());
                                        }
                                    } else {
                                        writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
                                    }
                                    break;
                                }
                            }
                        }
                    }
                }
            }
        }
        mOnlyCategories.clear();
        mExcludedCategories.clear();
        mStartId = 0;
    }

    private boolean shouldInclude(Equipment equipment) {
        for (String cat : mOnlyCategories) {
            if (equipment.hasCategory(cat)) {
                return true;
            }
        }
        if (!mOnlyCategories.isEmpty()) {  // If 'only' categories were provided, and none matched,
            // then false
            return false;
        }
        for (String cat : mExcludedCategories) {
            if (equipment.hasCategory(cat)) {
                return false;
            }
        }
        return true;
    }

    private void processNotesLoop(BufferedWriter out, String contents) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        for (Note note : mSheet.getCharacter().getNoteIterator()) {
            mCurrentId++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        if (!processStyleIndentWarning(key, out, note)) {
                            switch (key) {
                            case KEY_NOTE:
                                writeEncodedText(out, note.getDescription());
                                break;
                            case KEY_NOTE_FORMATTED:
                                if (!note.getDescription().isEmpty()) {
                                    writeEncodedText(out, PARAGRAPH_START + note.getDescription().replace(NEWLINE, PARAGRAPH_END + NEWLINE + PARAGRAPH_START) + PARAGRAPH_END);
                                }
                                break;
                            case KEY_ID:
                                writeEncodedText(out, Integer.toString(mCurrentId));
                                break;
                            default:
                                writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
                                break;
                            }
                        }
                    }
                }
            }
        }
        mStartId = 0;
    }

    private void processReactionLoop(BufferedWriter out, String contents) throws IOException {
        int           length           = contents.length();
        StringBuilder keyBuffer        = new StringBuilder();
        boolean       lookForKeyMarker = true;
        mCurrentId = mStartId;
        List<ReactionRow> reactions = mSheet.collectReactions();
        for (ReactionRow reaction : reactions) {
            mCurrentId++;
            for (int i = 0; i < length; i++) {
                char ch = contents.charAt(i);
                if (lookForKeyMarker) {
                    if (ch == '@') {
                        lookForKeyMarker = false;
                    } else {
                        out.append(ch);
                    }
                } else {
                    if (ch == '_' || Character.isLetterOrDigit(ch)) {
                        keyBuffer.append(ch);
                    } else {
                        String key = keyBuffer.toString();
                        i--;
                        if (mEnhancedKeyParsing && ch == '@') {
                            i++;        // Allow KEYs to be surrounded by @KEY@
                        }
                        keyBuffer.setLength(0);
                        lookForKeyMarker = true;
                        switch (key) {
                        case KEY_MODIFIER:
                            writeEncodedText(out, Numbers.formatWithForcedSign(reaction.getTotalAmount()));
                            break;
                        case KEY_SITUATION:
                            writeEncodedText(out, reaction.getFrom());
                            break;
                        case KEY_ID:
                            writeEncodedText(out, Integer.toString(mCurrentId));
                            break;
                        default:
                            writeEncodedText(out, String.format(UNIDENTIFIED_KEY, key));
                            break;
                        }
                    }
                }
            }
        }
        mStartId = 0;
    }

    private enum AdvantagesLoopType {
        ALL {
            @Override
            public boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded) {
                return includeByCategories(advantage, included, excluded);
            }

        }, ADS {
            @Override
            public boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded) {
                return advantage.getAdjustedPoints() > 1 && includeByCategories(advantage, included, excluded);
            }

        }, ADS_ALL {
            @Override
            public boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded) {
                return advantage.getAdjustedPoints() > 0 && includeByCategories(advantage, included, excluded);
            }

        }, DISADS {
            @Override
            public boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded) {
                return advantage.getAdjustedPoints() < -1 && includeByCategories(advantage, included, excluded);
            }

        }, DISADS_ALL {
            @Override
            public boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded) {
                return advantage.getAdjustedPoints() < 0 && includeByCategories(advantage, included, excluded);
            }

        }, PERKS {
            @Override
            public boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded) {
                return advantage.getAdjustedPoints() == 1 && includeByCategories(advantage, included, excluded);
            }

        }, QUIRKS {
            @Override
            public boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded) {
                return advantage.getAdjustedPoints() == -1 && includeByCategories(advantage, included, excluded);
            }

        }, LANGUAGES {
            @Override
            public boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded) {
                return advantage.getCategories().contains("Language") && includeByCategories(advantage, included, excluded);
            }

        }, CULTURAL_FAMILIARITIES {
            @Override
            public boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded) {
                return advantage.getName().startsWith("Cultural Familiarity (") && includeByCategories(advantage, included, excluded);
            }

        };

        public abstract boolean shouldInclude(Advantage advantage, Set<String> included, Set<String> excluded);

        private static boolean includeByCategories(Advantage advantage, Set<String> included, Set<String> excluded) {
            for (String cat : included) {
                if (advantage.hasCategory(cat)) {
                    return true;
                }
            }
            if (!included.isEmpty()) {
                return false;
            }
            for (String cat : excluded) {
                if (advantage.hasCategory(cat)) {
                    return false;
                }
            }
            return true;
        }
    }
}
