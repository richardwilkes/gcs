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

package com.trollworks.gcs.character;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.feature.Bonus;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.modifier.Modifier;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.widgets.outline.ListRow;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.Iterator;

/**
 * A thread for doing background updates of the prerequisite status of a character sheet.
 */
public class PrerequisitesThread extends Thread implements NotifierTarget {
	private static String										MSG_BULLET_PREFIX;
	private static HashMap<GURPSCharacter, PrerequisitesThread>	MAP		= new HashMap<GURPSCharacter, PrerequisitesThread>();
	private static int											COUNTER	= 0;
	private CharacterSheet												mSheet;
	private GURPSCharacter											mCharacter;
	private boolean												mNeedUpdate;
	private boolean												mNeedRepaint;
	private boolean												mIsProcessing;

	static {
		LocalizedMessages.initialize(PrerequisitesThread.class);
	}

	/**
	 * @param character The character being processed.
	 * @return The thread that does the processing.
	 */
	public static PrerequisitesThread getThread(GURPSCharacter character) {
		return MAP.get(character);
	}

	/**
	 * Returns only when the prerequisites thread is idle.
	 * 
	 * @param character The character to wait for.
	 * @return The thread that does the processing.
	 */
	public static PrerequisitesThread waitForProcessingToFinish(GURPSCharacter character) {
		PrerequisitesThread thread = getThread(character);

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
	public PrerequisitesThread(CharacterSheet sheet) {
		super("Prerequisites #" + ++COUNTER); //$NON-NLS-1$
		setPriority(NORM_PRIORITY);
		setDaemon(true);
		mSheet = sheet;
		mCharacter = sheet.getCharacter();
		mNeedUpdate = true;
		mCharacter.addTarget(this, GURPSCharacter.ID_STRENGTH, GURPSCharacter.ID_DEXTERITY, GURPSCharacter.ID_INTELLIGENCE, GURPSCharacter.ID_HEALTH, Spell.ID_NAME, Spell.ID_COLLEGE, Spell.ID_POINTS, Spell.ID_LIST_CHANGED, Skill.ID_NAME, Skill.ID_SPECIALIZATION, Skill.ID_LEVEL, Skill.ID_RELATIVE_LEVEL, Skill.ID_ENCUMBRANCE_PENALTY, Skill.ID_POINTS, Skill.ID_TECH_LEVEL, Skill.ID_LIST_CHANGED, Advantage.ID_NAME, Advantage.ID_LEVELS, Advantage.ID_LIST_CHANGED, Equipment.ID_EXTENDED_WEIGHT, Equipment.ID_EQUIPPED, Equipment.ID_QUANTITY, Equipment.ID_LIST_CHANGED);
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
		HashMap<String, ArrayList<Feature>> map = new HashMap<String, ArrayList<Feature>>();

		buildFeatureMap(map, mCharacter.getAdvantagesIterator());
		buildFeatureMap(map, mCharacter.getSkillsIterator());
		buildFeatureMap(map, mCharacter.getSpellsIterator());
		buildFeatureMap(map, mCharacter.getCarriedEquipmentIterator());
		mCharacter.setFeatureMap(map);
	}

	private void buildFeatureMap(HashMap<String, ArrayList<Feature>> map, Iterator<? extends ListRow> iterator) throws Exception {
		while (iterator.hasNext()) {
			ListRow row = iterator.next();

			if (row instanceof Equipment) {
				Equipment equipment = (Equipment) row;

				if (!equipment.isEquipped() || equipment.getQuantity() < 1) {
					// Don't allow unequipped equipment to affect the character
					continue;
				}
			}

			for (Feature feature : row.getFeatures()) {
				String key = feature.getKey().toLowerCase();
				ArrayList<Feature> list = map.get(key);

				if (list == null) {
					list = new ArrayList<Feature>(1);
					map.put(key, list);
				}
				if (row instanceof Advantage && feature instanceof Bonus) {
					((Bonus) feature).getAmount().setLevel(((Advantage) row).getLevels());
				}
				list.add(feature);
			}

			if (row instanceof Advantage) {
				Advantage advantage = (Advantage) row;
				for (Modifier modifier : advantage.getModifiers()) {
					if (modifier.isEnabled()) {
						for (Feature feature : modifier.getFeatures()) {
							String key = feature.getKey().toLowerCase();
							ArrayList<Feature> list = map.get(key);

							if (list == null) {
								list = new ArrayList<Feature>(1);
								map.put(key, list);
							}
							if (feature instanceof Bonus) {
								((Bonus) feature).getAmount().setLevel(modifier.getLevels());
							}
							list.add(feature);
						}
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

	private void processRows(Iterator<? extends ListRow> iterator) throws Exception {
		StringBuilder builder = new StringBuilder();

		while (iterator.hasNext()) {
			ListRow row = iterator.next();
			boolean satisfied;

			builder.setLength(0);

			satisfied = row.getPrereqs().satisfied(mCharacter, row, builder, MSG_BULLET_PREFIX);
			if (satisfied && row instanceof Technique) {
				satisfied = ((Technique) row).satisfied(builder, MSG_BULLET_PREFIX);
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
