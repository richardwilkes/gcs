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

import com.trollworks.gcs.feature.HitLocation;

/** Tracks the current armor levels. */
public class Armor {
    /** The prefix used in front of all IDs for damage resistance. */
    public static final  String DR_PREFIX                   = GURPSCharacter.CHARACTER_PREFIX + "dr.";
    /** The skull hit location's DR. */
    public static final  String ID_SKULL_DR                 = DR_PREFIX + HitLocation.SKULL.name();
    /** The eyes hit location's DR. */
    public static final  String ID_EYES_DR                  = DR_PREFIX + HitLocation.EYES.name();
    /** The face hit location's DR. */
    public static final  String ID_FACE_DR                  = DR_PREFIX + HitLocation.FACE.name();
    /** The neck hit location's DR. */
    public static final  String ID_NECK_DR                  = DR_PREFIX + HitLocation.NECK.name();
    /** The torso hit location's DR. */
    public static final  String ID_TORSO_DR                 = DR_PREFIX + HitLocation.TORSO.name();
    /** The vitals hit location's DR. */
    public static final  String ID_VITALS_DR                = DR_PREFIX + HitLocation.VITALS.name();
    private static final String ID_FULL_BODY_DR             = DR_PREFIX + HitLocation.FULL_BODY.name();
    private static final String ID_FULL_BODY_EXCEPT_EYES_DR = DR_PREFIX + HitLocation.FULL_BODY_EXCEPT_EYES.name();
    /** The groin hit location's DR. */
    public static final  String ID_GROIN_DR                 = DR_PREFIX + HitLocation.GROIN.name();
    /** The arm hit location's DR. */
    public static final  String ID_ARM_DR                   = DR_PREFIX + HitLocation.ARMS.name();
    /** The hand hit location's DR. */
    public static final  String ID_HAND_DR                  = DR_PREFIX + HitLocation.HANDS.name();
    /** The leg hit location's DR. */
    public static final  String ID_LEG_DR                   = DR_PREFIX + HitLocation.LEGS.name();
    /** The foot hit location's DR. */
    public static final  String ID_FOOT_DR                  = DR_PREFIX + HitLocation.FEET.name();
    /** The tail hit location's DR. */
    public static final  String ID_TAIL_DR                  = DR_PREFIX + HitLocation.TAIL.name();
    /** The wing hit location's DR. */
    public static final  String ID_WING_DR                  = DR_PREFIX + HitLocation.WINGS.name();
    /** The fin hit location's DR. */
    public static final  String ID_FIN_DR                   = DR_PREFIX + HitLocation.FINS.name();
    /** The brain hit location's DR. */
    public static final  String ID_BRAIN_DR                 = DR_PREFIX + HitLocation.BRAIN.name();

    private GURPSCharacter mCharacter;
    private int            mBrainDR;
    private int            mSkullDR;
    private int            mEyesDR;
    private int            mFaceDR;
    private int            mNeckDR;
    private int            mTorsoDR;
    private int            mVitalsDR;
    private int            mGroinDR;
    private int            mArmDR;
    private int            mWingDR;
    private int            mHandDR;
    private int            mFinDR;
    private int            mLegDR;
    private int            mFootDR;
    private int            mTailDR;

    Armor(GURPSCharacter character) {
        mCharacter = character;
        mSkullDR = 2;
    }

    void update() {
        int extra = mCharacter.getIntegerBonusFor(ID_FULL_BODY_DR);
        mCharacter.startNotify();
        setEyesDR(getBonusDR(ID_EYES_DR) + extra);
        extra += mCharacter.getIntegerBonusFor(ID_FULL_BODY_EXCEPT_EYES_DR);
        setBrainDR(getBonusDR(ID_BRAIN_DR) + extra);
        setSkullDR(getBonusDR(ID_SKULL_DR) + extra);
        setFaceDR(getBonusDR(ID_FACE_DR) + extra);
        setNeckDR(getBonusDR(ID_NECK_DR) + extra);
        int torsoDR = getBonusDR(ID_TORSO_DR);
        setTorsoDR(torsoDR + extra);
        setVitalsDR(getBonusDR(ID_VITALS_DR) + torsoDR + extra);
        setGroinDR(getBonusDR(ID_GROIN_DR) + extra);
        setArmDR(getBonusDR(ID_ARM_DR) + extra);
        setWingDR(getBonusDR(ID_WING_DR) + extra);
        setHandDR(getBonusDR(ID_HAND_DR) + extra);
        setFinDR(getBonusDR(ID_FIN_DR) + extra);
        setLegDR(getBonusDR(ID_LEG_DR) + extra);
        setFootDR(getBonusDR(ID_FOOT_DR) + extra);
        setTailDR(getBonusDR(ID_TAIL_DR) + extra);
        mCharacter.endNotify();
    }

    private int getBonusDR(String key) {
        int                                      bonus       = mCharacter.getIntegerBonusFor(key);
        com.trollworks.gcs.character.HitLocation hitLocation = com.trollworks.gcs.character.HitLocation.MAP.get(key);
        if (hitLocation != null) {
            bonus += hitLocation.getDRBonus();
        }
        return bonus;
    }

    /** @return The brain hit location's DR. */
    public int getBrainDR() {
        return mBrainDR;
    }

    /**
     * Sets the brain hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setBrainDR(int dr) {
        if (mBrainDR != dr) {
            mBrainDR = dr;
            mCharacter.notifySingle(ID_BRAIN_DR, Integer.valueOf(mBrainDR));
        }
    }

    /** @return The skull hit location's DR. */
    public int getSkullDR() {
        return mSkullDR;
    }

    /**
     * Sets the skull hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setSkullDR(int dr) {
        if (mSkullDR != dr) {
            mSkullDR = dr;
            mCharacter.notifySingle(ID_SKULL_DR, Integer.valueOf(mSkullDR));
        }
    }

    /** @return The eyes hit location's DR. */
    public int getEyesDR() {
        return mEyesDR;
    }

    /**
     * Sets the eyes hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setEyesDR(int dr) {
        if (mEyesDR != dr) {
            mEyesDR = dr;
            mCharacter.notifySingle(ID_EYES_DR, Integer.valueOf(mEyesDR));
        }
    }

    /** @return The face hit location's DR. */
    public int getFaceDR() {
        return mFaceDR;
    }

    /**
     * Sets the face hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setFaceDR(int dr) {
        if (mFaceDR != dr) {
            mFaceDR = dr;
            mCharacter.notifySingle(ID_FACE_DR, Integer.valueOf(mFaceDR));
        }
    }

    /** @return The neck hit location's DR. */
    public int getNeckDR() {
        return mNeckDR;
    }

    /**
     * Sets the neck hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setNeckDR(int dr) {
        if (mNeckDR != dr) {
            mNeckDR = dr;
            mCharacter.notifySingle(ID_NECK_DR, Integer.valueOf(mNeckDR));
        }
    }

    /** @return The torso hit location's DR. */
    public int getTorsoDR() {
        return mTorsoDR;
    }

    /**
     * Sets the torso hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setTorsoDR(int dr) {
        if (mTorsoDR != dr) {
            mTorsoDR = dr;
            mCharacter.notifySingle(ID_TORSO_DR, Integer.valueOf(mTorsoDR));
        }
    }

    /** @return The vitals hit location's DR. */
    public int getVitalsDR() {
        return mVitalsDR;
    }

    /**
     * Sets the vitals hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setVitalsDR(int dr) {
        if (mVitalsDR != dr) {
            mVitalsDR = dr;
            mCharacter.notifySingle(ID_VITALS_DR, Integer.valueOf(mVitalsDR));
        }
    }

    /** @return The groin hit location's DR. */
    public int getGroinDR() {
        return mGroinDR;
    }

    /**
     * Sets the groin hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setGroinDR(int dr) {
        if (mGroinDR != dr) {
            mGroinDR = dr;
            mCharacter.notifySingle(ID_GROIN_DR, Integer.valueOf(mGroinDR));
        }
    }

    /** @return The arm hit location's DR. */
    public int getArmDR() {
        return mArmDR;
    }

    /**
     * Sets the arm hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setArmDR(int dr) {
        if (mArmDR != dr) {
            mArmDR = dr;
            mCharacter.notifySingle(ID_ARM_DR, Integer.valueOf(mArmDR));
        }
    }

    /** @return The wing hit location's DR. */
    public int getWingDR() {
        return mWingDR;
    }

    /**
     * Sets the wing hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setWingDR(int dr) {
        if (mWingDR != dr) {
            mWingDR = dr;
            mCharacter.notifySingle(ID_WING_DR, Integer.valueOf(mWingDR));
        }
    }

    /** @return The hand hit location's DR. */
    public int getHandDR() {
        return mHandDR;
    }

    /**
     * Sets the hand hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setHandDR(int dr) {
        if (mHandDR != dr) {
            mHandDR = dr;
            mCharacter.notifySingle(ID_HAND_DR, Integer.valueOf(mHandDR));
        }
    }

    /** @return The fin hit location's DR. */
    public int getFinDR() {
        return mFinDR;
    }

    /**
     * Sets the fin hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setFinDR(int dr) {
        if (mFinDR != dr) {
            mFinDR = dr;
            mCharacter.notifySingle(ID_FIN_DR, Integer.valueOf(mFinDR));
        }
    }

    /** @return The leg hit location's DR. */
    public int getLegDR() {
        return mLegDR;
    }

    /**
     * Sets the leg hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setLegDR(int dr) {
        if (mLegDR != dr) {
            mLegDR = dr;
            mCharacter.notifySingle(ID_LEG_DR, Integer.valueOf(mLegDR));
        }
    }

    /** @return The foot hit location's DR. */
    public int getFootDR() {
        return mFootDR;
    }

    /**
     * Sets the foot hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setFootDR(int dr) {
        if (mFootDR != dr) {
            mFootDR = dr;
            mCharacter.notifySingle(ID_FOOT_DR, Integer.valueOf(mFootDR));
        }
    }

    /** @return The tail hit location's DR. */
    public int getTailDR() {
        return mTailDR;
    }

    /**
     * Sets the tail hit location's DR.
     *
     * @param dr The DR amount.
     */
    public void setTailDR(int dr) {
        if (mTailDR != dr) {
            mTailDR = dr;
            mCharacter.notifySingle(ID_TAIL_DR, Integer.valueOf(mTailDR));
        }
    }

    /**
     * @param id The field ID to retrieve the data for.
     * @return The value of the specified field ID, or {@code null} if the field ID is invalid.
     */
    public Object getValueForID(String id) {
        if (id != null && id.startsWith(DR_PREFIX)) {
            if (ID_BRAIN_DR.equals(id)) {
                return Integer.valueOf(getBrainDR());
            } else if (ID_SKULL_DR.equals(id)) {
                return Integer.valueOf(getSkullDR());
            } else if (ID_EYES_DR.equals(id)) {
                return Integer.valueOf(getEyesDR());
            } else if (ID_FACE_DR.equals(id)) {
                return Integer.valueOf(getFaceDR());
            } else if (ID_NECK_DR.equals(id)) {
                return Integer.valueOf(getNeckDR());
            } else if (ID_TORSO_DR.equals(id)) {
                return Integer.valueOf(getTorsoDR());
            } else if (ID_VITALS_DR.equals(id)) {
                return Integer.valueOf(getVitalsDR());
            } else if (ID_GROIN_DR.equals(id)) {
                return Integer.valueOf(getGroinDR());
            } else if (ID_ARM_DR.equals(id)) {
                return Integer.valueOf(getArmDR());
            } else if (ID_WING_DR.equals(id)) {
                return Integer.valueOf(getWingDR());
            } else if (ID_HAND_DR.equals(id)) {
                return Integer.valueOf(getHandDR());
            } else if (ID_FIN_DR.equals(id)) {
                return Integer.valueOf(getFinDR());
            } else if (ID_LEG_DR.equals(id)) {
                return Integer.valueOf(getLegDR());
            } else if (ID_FOOT_DR.equals(id)) {
                return Integer.valueOf(getFootDR());
            } else if (ID_TAIL_DR.equals(id)) {
                return Integer.valueOf(getTailDR());
            }
        }
        return null;
    }

    /**
     * @param id    The field ID to set the value for.
     * @param value The value to set.
     */
    public void setValueForID(String id, Object value) {
        if (id != null && id.startsWith(DR_PREFIX)) {
            if (ID_BRAIN_DR.equals(id)) {
                setBrainDR(((Integer) value).intValue());
            } else if (ID_SKULL_DR.equals(id)) {
                setSkullDR(((Integer) value).intValue());
            } else if (ID_EYES_DR.equals(id)) {
                setEyesDR(((Integer) value).intValue());
            } else if (ID_FACE_DR.equals(id)) {
                setFaceDR(((Integer) value).intValue());
            } else if (ID_NECK_DR.equals(id)) {
                setNeckDR(((Integer) value).intValue());
            } else if (ID_TORSO_DR.equals(id)) {
                setTorsoDR(((Integer) value).intValue());
            } else if (ID_VITALS_DR.equals(id)) {
                setVitalsDR(((Integer) value).intValue());
            } else if (ID_GROIN_DR.equals(id)) {
                setGroinDR(((Integer) value).intValue());
            } else if (ID_ARM_DR.equals(id)) {
                setArmDR(((Integer) value).intValue());
            } else if (ID_WING_DR.equals(id)) {
                setWingDR(((Integer) value).intValue());
            } else if (ID_HAND_DR.equals(id)) {
                setHandDR(((Integer) value).intValue());
            } else if (ID_FIN_DR.equals(id)) {
                setFinDR(((Integer) value).intValue());
            } else if (ID_LEG_DR.equals(id)) {
                setLegDR(((Integer) value).intValue());
            } else if (ID_FOOT_DR.equals(id)) {
                setFootDR(((Integer) value).intValue());
            } else if (ID_TAIL_DR.equals(id)) {
                setTailDR(((Integer) value).intValue());
            }
        }
    }
}
