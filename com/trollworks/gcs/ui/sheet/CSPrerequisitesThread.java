/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.sheet;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.feature.CMBonus;
import com.trollworks.gcs.model.feature.CMFeature;
import com.trollworks.gcs.model.modifier.CMModifier;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMTechnique;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.toolkit.notification.TKNotifierTarget;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.Iterator;

/**
 * A thread for doing background updates of the prerequisite status of a character sheet.
 */
public class CSPrerequisitesThread extends Thread implements TKNotifierTarget {
	private static HashMap<CMCharacter, CSPrerequisitesThread>	MAP		= new HashMap<CMCharacter, CSPrerequisitesThread>();
	private static int											COUNTER	= 0;
	private CSSheet												mSheet;
	private CMCharacter											mCharacter;
	private boolean												mNeedUpdate;
	private boolean												mNeedRepaint;
	private boolean												mIsProcessing;

	/**
	 * @param character The character being processed.
	 * @return The thread that does the processing.
	 */
	public static CSPrerequisitesThread getThread(CMCharacter character) {
		return MAP.get(character);
	}

	/**
	 * Returns only when the prerequisites thread is idle.
	 * 
	 * @param character The character to wait for.
	 * @return The thread that does the processing.
	 */
	public static CSPrerequisitesThread waitForProcessingToFinish(CMCharacter character) {
		CSPrerequisitesThread thread = getThread(character);

		if (thread != null && thread != Thread.currentThread()) {
			boolean checkAgain = true;

			while (checkAgain) {
				synchronized (thread) {
					checkAgain = thread.mIsProcessing || thread.mNeedUpdate;
				}
				try {
					sleep(200);
				} catch (Exception exception) {
					// Ignore...
				}
			}
		}
		return thread;
	}

	/**
	 * Creates a new prerequisites thread.
	 * 
	 * @param sheet The sheet we're attached to.
	 */
	public CSPrerequisitesThread(CSSheet sheet) {
		super("Prerequisites #" + ++COUNTER); //$NON-NLS-1$
		setPriority(NORM_PRIORITY);
		setDaemon(true);
		mSheet = sheet;
		mCharacter = sheet.getCharacter();
		mNeedUpdate = true;
		mCharacter.addTarget(this, CMCharacter.ID_STRENGTH, CMCharacter.ID_DEXTERITY, CMCharacter.ID_INTELLIGENCE, CMCharacter.ID_HEALTH, CMSpell.ID_NAME, CMSpell.ID_COLLEGE, CMSpell.ID_POINTS, CMSpell.ID_LIST_CHANGED, CMSkill.ID_NAME, CMSkill.ID_SPECIALIZATION, CMSkill.ID_LEVEL, CMSkill.ID_RELATIVE_LEVEL, CMSkill.ID_ENCUMBRANCE_PENALTY, CMSkill.ID_POINTS, CMSkill.ID_TECH_LEVEL, CMSkill.ID_LIST_CHANGED, CMAdvantage.ID_NAME, CMAdvantage.ID_LEVELS, CMAdvantage.ID_LIST_CHANGED, CMEquipment.ID_EXTENDED_WEIGHT, CMEquipment.ID_EQUIPPED, CMEquipment.ID_QUANTITY, CMEquipment.ID_LIST_CHANGED);
		MAP.put(mCharacter, this);
	}

	@Override public void run() {
		try {
			while (!mSheet.hasBeenDisposed()) {
				try {
					boolean needUpdate;

					synchronized (this) {
						needUpdate = mNeedUpdate;
						mNeedUpdate = false;
						mIsProcessing = needUpdate;
					}
					if (!needUpdate) {
						sleep(500);
					} else {
						processFeatures();
						processRows(mCharacter.getAdvantagesIterator());
						processRows(mCharacter.getSkillsIterator());
						processRows(mCharacter.getSpellsIterator());
						processRows(mCharacter.getCarriedEquipmentIterator());
						if (mNeedRepaint) {
							mSheet.repaint();
						}
						synchronized (this) {
							mIsProcessing = false;
						}
					}
				} catch (InterruptedException iEx) {
					throw iEx;
				} catch (Exception exception) {
					// Catch everything here so that manipulations to the character
					// sheet that invalidate state don't stop our thread from
					// continuing.
					synchronized (this) {
						mNeedUpdate = true;
					}
					if (mNeedRepaint) {
						mSheet.repaint();
					}
					sleep(200);
				}
			}
		} catch (InterruptedException outerIEx) {
			// Someone is tring to terminate us... let them.
		}
		mNeedUpdate = mIsProcessing = false;
		MAP.remove(mCharacter);
	}

	private void processFeatures() throws Exception {
		HashMap<String, ArrayList<CMFeature>> map = new HashMap<String, ArrayList<CMFeature>>();

		buildFeatureMap(map, mCharacter.getAdvantagesIterator());
		buildFeatureMap(map, mCharacter.getSkillsIterator());
		buildFeatureMap(map, mCharacter.getSpellsIterator());
		buildFeatureMap(map, mCharacter.getCarriedEquipmentIterator());
		mCharacter.setFeatureMap(map);
	}

	private void buildFeatureMap(HashMap<String, ArrayList<CMFeature>> map, Iterator<? extends CMRow> iterator) throws Exception {
		while (iterator.hasNext()) {
			CMRow row = iterator.next();

			if (row instanceof CMEquipment) {
				CMEquipment equipment = (CMEquipment) row;

				if (!equipment.isEquipped() || equipment.getQuantity() < 1) {
					// Don't allow unequipped equipment to affect the character
					continue;
				}
			}

			for (CMFeature feature : row.getFeatures()) {
				String key = feature.getKey().toLowerCase();
				ArrayList<CMFeature> list = map.get(key);

				if (list == null) {
					list = new ArrayList<CMFeature>(1);
					map.put(key, list);
				}
				if (row instanceof CMAdvantage && feature instanceof CMBonus) {
					((CMBonus) feature).getAmount().setLevel(((CMAdvantage) row).getLevels());
				}
				list.add(feature);
			}

			if (row instanceof CMAdvantage) {
				CMAdvantage advantage = (CMAdvantage) row;
				for (CMModifier modifier : advantage.getModifiers()) {
					for (CMFeature feature : modifier.getFeatures()) {
						String key = feature.getKey().toLowerCase();
						ArrayList<CMFeature> list = map.get(key);

						if (list == null) {
							list = new ArrayList<CMFeature>(1);
							map.put(key, list);
						}
						if (feature instanceof CMBonus) {
							((CMBonus) feature).getAmount().setLevel(modifier.getLevels());
						}
						list.add(feature);
					}
				}
			}

			checkIfUpdated();
		}
	}

	private void checkIfUpdated() throws Exception {
		boolean needUpdate;

		synchronized (this) {
			needUpdate = mNeedUpdate;
		}
		if (needUpdate || mSheet.hasBeenDisposed()) {
			throw new Exception();
		}
	}

	private void processRows(Iterator<? extends CMRow> iterator) throws Exception {
		StringBuilder builder = new StringBuilder();

		while (iterator.hasNext()) {
			CMRow row = iterator.next();
			boolean satisfied;

			builder.setLength(0);

			satisfied = row.getPrereqs().satisfied(mCharacter, row, builder, Msgs.BULLET_PREFIX);
			if (satisfied && row instanceof CMTechnique) {
				satisfied = ((CMTechnique) row).satisfied(builder, Msgs.BULLET_PREFIX);
			}

			if (row.isSatisfied() != satisfied) {
				row.setSatisfied(satisfied);
				mNeedRepaint = true;
			}
			if (!satisfied) {
				row.setReasonForUnsatisfied(builder.toString());
			}
			checkIfUpdated();
		}
	}

	/** Marks an update request. */
	public void markForUpdate() {
		synchronized (this) {
			mNeedUpdate = true;
		}
	}

	public void handleNotification(Object producer, String type, Object data) {
		markForUpdate();
	}
}
