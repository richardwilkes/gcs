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

import com.trollworks.gcs.utility.I18n;

import java.util.HashMap;
import java.util.Map;

/** Possible hit locations. */
public class HitLocation {
    public static final Map<String, HitLocation> MAP    = new HashMap<>();
    public static final HitLocation              EYE    = new HitLocation(Armor.ID_EYES_DR, I18n.Text("Eye"), I18n.Text("An attack that misses by 1 hits the torso instead. Only impaling, piercing, and tight-beam burning attacks can target the eye – and only from the front or sides. Injury over HP÷10 blinds the eye. Otherwise, treat as skull, but without the extra DR!"), -9, 0);
    public static final HitLocation              SKULL  = new HitLocation(Armor.ID_SKULL_DR, I18n.Text("Skull"), I18n.Text("An attack that misses by 1 hits the torso instead. Wounding modifier is x4. Knockdown rolls are at -10. Critical hits use the Critical Head Blow Table (B556). Exception: These special effects do not apply to toxic damage."), -7, 2);
    public static final HitLocation              FACE   = new HitLocation(Armor.ID_FACE_DR, I18n.Text("Face"), I18n.Text("An attack that misses by 1 hits the torso instead. Jaw, cheeks, nose, ears, etc. If the target has an open-faced helmet, ignore its DR. Knockdown rolls are at -5. Critical hits use the Critical Head Blow Table (B556). Corrosion damage gets a x1½ wounding modifier, and if it inflicts a major wound, it also blinds one eye (both eyes on damage over full HP). Random attacks from behind hit the skull instead."), -5, 0);
    public static final HitLocation              LEG    = new HitLocation(Armor.ID_LEG_DR, I18n.Text("Leg"), I18n.Text("Reduce the wounding multiplier of large piercing, huge piercing, and impaling damage to x1. Any major wound (loss of over ½ HP from one blow) cripples the limb. Damage beyond that threshold is lost."), -2, 0);
    public static final HitLocation              ARM    = new HitLocation(Armor.ID_ARM_DR, I18n.Text("Arm"), I18n.Text("Reduce the wounding multiplier of large piercing, huge piercing, and impaling damage to x1. Any major wound (loss of over ½ HP from one blow) cripples the limb. Damage beyond that threshold is lost. If holding a shield, double the penalty to hit: -4 for shield arm instead of -2."), -2, 0);
    public static final HitLocation              TORSO  = new HitLocation(Armor.ID_TORSO_DR, I18n.Text("Torso"), "", 0, 0);
    public static final HitLocation              GROIN  = new HitLocation(Armor.ID_GROIN_DR, I18n.Text("Groin"), I18n.Text("An attack that misses by 1 hits the torso instead. Human males and the males of similar species suffer double shock from crushing damage, and get -5 to knockdown rolls. Otherwise, treat as a torso hit."), -3, 0);
    public static final HitLocation              HAND   = new HitLocation(Armor.ID_HAND_DR, I18n.Text("Hand"), I18n.Text("If holding a shield, double the penalty to hit: -8 for shield hand instead of -4. Reduce the wounding multiplier of large piercing, huge piercing, and impaling damage to x1. Any major wound (loss of over ⅓ HP from one blow) cripples the limb. Damage beyond that threshold is lost."), -4, 0);
    public static final HitLocation              FOOT   = new HitLocation(Armor.ID_FOOT_DR, I18n.Text("Foot"), I18n.Text("Reduce the wounding multiplier of large piercing, huge piercing, and impaling damage to x1. Any major wound (loss of over ⅓ HP from one blow) cripples the limb. Damage beyond that threshold is lost."), -4, 0);
    public static final HitLocation              NECK   = new HitLocation(Armor.ID_NECK_DR, I18n.Text("Neck"), I18n.Text("An attack that misses by 1 hits the torso instead. Neck and throat. Increase the wounding multiplier of crushing and corrosion attacks to x1.5, and that of cutting damage to x2. At the GM’s option, anyone killed by a cutting blow to the neck is decapitated!"), -5, 0);
    public static final HitLocation              VITALS = new HitLocation(Armor.ID_VITALS_DR, I18n.Text("Vitals"), I18n.Text("An attack that misses by 1 hits the torso instead. Heart, lungs, kidneys, etc. Increase the wounding modifier for an impaling or any piercing attack to x3. Increase the wounding modifier for a tight-beam burning attack to x2. Other attacks cannot target the vitals."), -3, 0);
    public static final HitLocation              TAIL   = new HitLocation(Armor.ID_TAIL_DR, I18n.Text("Tail"), I18n.Text("If a tail counts as an Extra Arm or a Striker, or is a fish tail, treat it as a limb (arm, leg) for crippling purposes; otherwise, treat it as an extremity (hand, foot). A crippled tail affects balance. For a ground creature, this gives -1 DX. For a swimmer or flyer, this gives -2 DX and halves Move. If the creature has no tail, or a very short one (like a rabbit), treat as torso."), -3, 0);
    public static final HitLocation              WING   = new HitLocation(Armor.ID_WING_DR, I18n.Text("Wing"), I18n.Text("Reduce the wounding multiplier of large piercing, huge piercing, and impaling damage to x1. Any major wound (loss of over ½ HP from one blow) cripples the limb. Damage beyond that threshold is lost. A flyer with a crippled wing cannot fly."), -2, 0);
    public static final HitLocation              FIN    = new HitLocation(Armor.ID_FIN_DR, I18n.Text("Fin"), I18n.Text("Reduce the wounding multiplier of large piercing, huge piercing, and impaling damage to x1. Any major wound (loss of over ⅓ HP from one blow) cripples the limb. Damage beyond that threshold is lost. A crippled fin affects balance: -3 DX."), -4, 0);
    public static final HitLocation              BRAIN  = new HitLocation(Armor.ID_BRAIN_DR, I18n.Text("Brain"), I18n.Text("An attack that misses by 1 hits the torso instead. Wounding modifier is x4. Knockdown rolls are at -10. Critical hits use the Critical Head Blow Table (B556). Exception: These special effects do not apply to toxic damage."), -7, 1);

    private String mKey;
    private String mTitle;
    private String mDescription;
    private int    mHitPenalty;
    private int    mDRBonus;

    private HitLocation(String key, String title, String description, int hitPenalty, int drBonus) {
        mKey = key;
        mTitle = title;
        mDescription = description;
        mHitPenalty = hitPenalty;
        mDRBonus = drBonus;
        MAP.put(key, this);
    }

    public String getKey() {
        return mKey;
    }

    public String getTitle() {
        return mTitle;
    }

    public String getDescription() {
        return mDescription;
    }

    public int getHitPenalty() {
        return mHitPenalty;
    }

    public int getDRBonus() {
        return mDRBonus;
    }
}
