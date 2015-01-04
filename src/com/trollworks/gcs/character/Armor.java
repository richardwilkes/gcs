/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.feature.HitLocation;

/** Tracks the current armor levels. */
public class Armor {
	/** The prefix used in front of all IDs for damage resistance. */
	public static final String	DR_PREFIX					= GURPSCharacter.CHARACTER_PREFIX + "dr.";				//$NON-NLS-1$
	/** The skull hit location's DR. */
	public static final String	ID_SKULL_DR					= DR_PREFIX + HitLocation.SKULL.name();
	/** The eyes hit location's DR. */
	public static final String	ID_EYES_DR					= DR_PREFIX + HitLocation.EYES.name();
	/** The face hit location's DR. */
	public static final String	ID_FACE_DR					= DR_PREFIX + HitLocation.FACE.name();
	/** The neck hit location's DR. */
	public static final String	ID_NECK_DR					= DR_PREFIX + HitLocation.NECK.name();
	/** The torso hit location's DR. */
	public static final String	ID_TORSO_DR					= DR_PREFIX + HitLocation.TORSO.name();
	/** The vitals hit location's DR. */
	public static final String	ID_VITALS_DR				= DR_PREFIX + HitLocation.VITALS.name();
	private static final String	ID_FULL_BODY_DR				= DR_PREFIX + HitLocation.FULL_BODY.name();
	private static final String	ID_FULL_BODY_EXCEPT_EYES_DR	= DR_PREFIX + HitLocation.FULL_BODY_EXCEPT_EYES.name();
	/** The groin hit location's DR. */
	public static final String	ID_GROIN_DR					= DR_PREFIX + HitLocation.GROIN.name();
	/** The arm hit location's DR. */
	public static final String	ID_ARM_DR					= DR_PREFIX + HitLocation.ARMS.name();
	/** The hand hit location's DR. */
	public static final String	ID_HAND_DR					= DR_PREFIX + HitLocation.HANDS.name();
	/** The leg hit location's DR. */
	public static final String	ID_LEG_DR					= DR_PREFIX + HitLocation.LEGS.name();
	/** The foot hit location's DR. */
	public static final String	ID_FOOT_DR					= DR_PREFIX + HitLocation.FEET.name();
	private GURPSCharacter		mCharacter;
	private int					mSkullDR;
	private int					mEyesDR;
	private int					mFaceDR;
	private int					mNeckDR;
	private int					mTorsoDR;
	private int					mVitalsDR;
	private int					mGroinDR;
	private int					mArmDR;
	private int					mHandDR;
	private int					mLegDR;
	private int					mFootDR;

	Armor(GURPSCharacter character) {
		mCharacter = character;
		mSkullDR = 2;
	}

	void update() {
		int fullBodyDR = mCharacter.getIntegerBonusFor(ID_FULL_BODY_DR);
		int fullBodyNoEyesDR = mCharacter.getIntegerBonusFor(ID_FULL_BODY_EXCEPT_EYES_DR);
		mCharacter.startNotify();
		setSkullDR(2 + mCharacter.getIntegerBonusFor(ID_SKULL_DR) + fullBodyDR + fullBodyNoEyesDR);
		setEyesDR(mCharacter.getIntegerBonusFor(ID_EYES_DR) + fullBodyDR);
		setFaceDR(mCharacter.getIntegerBonusFor(ID_FACE_DR) + fullBodyDR + fullBodyNoEyesDR);
		setNeckDR(mCharacter.getIntegerBonusFor(ID_NECK_DR) + fullBodyDR + fullBodyNoEyesDR);
		setTorsoDR(mCharacter.getIntegerBonusFor(ID_TORSO_DR) + fullBodyDR + fullBodyNoEyesDR);
		setVitalsDR(mCharacter.getIntegerBonusFor(ID_VITALS_DR) + fullBodyDR + fullBodyNoEyesDR);
		setGroinDR(mCharacter.getIntegerBonusFor(ID_GROIN_DR) + fullBodyDR + fullBodyNoEyesDR);
		setArmDR(mCharacter.getIntegerBonusFor(ID_ARM_DR) + fullBodyDR + fullBodyNoEyesDR);
		setHandDR(mCharacter.getIntegerBonusFor(ID_HAND_DR) + fullBodyDR + fullBodyNoEyesDR);
		setLegDR(mCharacter.getIntegerBonusFor(ID_LEG_DR) + fullBodyDR + fullBodyNoEyesDR);
		setFootDR(mCharacter.getIntegerBonusFor(ID_FOOT_DR) + fullBodyDR + fullBodyNoEyesDR);
		mCharacter.endNotify();
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
			mCharacter.notifySingle(ID_SKULL_DR, new Integer(mSkullDR));
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
			mCharacter.notifySingle(ID_EYES_DR, new Integer(mEyesDR));
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
			mCharacter.notifySingle(ID_FACE_DR, new Integer(mFaceDR));
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
			mCharacter.notifySingle(ID_NECK_DR, new Integer(mNeckDR));
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
			mCharacter.notifySingle(ID_TORSO_DR, new Integer(mTorsoDR));
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
			mCharacter.notifySingle(ID_VITALS_DR, new Integer(mVitalsDR));
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
			mCharacter.notifySingle(ID_GROIN_DR, new Integer(mGroinDR));
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
			mCharacter.notifySingle(ID_ARM_DR, new Integer(mArmDR));
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
			mCharacter.notifySingle(ID_HAND_DR, new Integer(mHandDR));
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
			mCharacter.notifySingle(ID_LEG_DR, new Integer(mLegDR));
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
			mCharacter.notifySingle(ID_FOOT_DR, new Integer(mFootDR));
		}
	}

	/**
	 * @param id The field ID to retrieve the data for.
	 * @return The value of the specified field ID, or <code>null</code> if the field ID is invalid.
	 */
	public Object getValueForID(String id) {
		if (id != null && id.startsWith(DR_PREFIX)) {
			if (ID_SKULL_DR.equals(id)) {
				return new Integer(getSkullDR());
			} else if (ID_EYES_DR.equals(id)) {
				return new Integer(getEyesDR());
			} else if (ID_FACE_DR.equals(id)) {
				return new Integer(getFaceDR());
			} else if (ID_NECK_DR.equals(id)) {
				return new Integer(getNeckDR());
			} else if (ID_TORSO_DR.equals(id)) {
				return new Integer(getTorsoDR());
			} else if (ID_VITALS_DR.equals(id)) {
				return new Integer(getVitalsDR());
			} else if (ID_GROIN_DR.equals(id)) {
				return new Integer(getGroinDR());
			} else if (ID_ARM_DR.equals(id)) {
				return new Integer(getArmDR());
			} else if (ID_HAND_DR.equals(id)) {
				return new Integer(getHandDR());
			} else if (ID_LEG_DR.equals(id)) {
				return new Integer(getLegDR());
			} else if (ID_FOOT_DR.equals(id)) {
				return new Integer(getFootDR());
			}
		}
		return null;
	}

	/**
	 * @param id The field ID to set the value for.
	 * @param value The value to set.
	 */
	public void setValueForID(String id, Object value) {
		if (id != null && id.startsWith(DR_PREFIX)) {
			if (ID_SKULL_DR.equals(id)) {
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
			} else if (ID_HAND_DR.equals(id)) {
				setHandDR(((Integer) value).intValue());
			} else if (ID_LEG_DR.equals(id)) {
				setLegDR(((Integer) value).intValue());
			} else if (ID_FOOT_DR.equals(id)) {
				setFootDR(((Integer) value).intValue());
			}
		}
	}
}
