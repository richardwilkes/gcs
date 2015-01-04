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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.feature.Bonus;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.modifier.Modifier;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.notification.NotifierTarget;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.Iterator;

/**
 * A thread for doing background updates of the prerequisite status of a character sheet.
 */
public class PrerequisitesThread extends Thread implements NotifierTarget {
	private static HashMap<GURPSCharacter, PrerequisitesThread>	MAP		= new HashMap<>();
	private static int											COUNTER	= 0;
	private CharacterSheet										mSheet;
	private GURPSCharacter										mCharacter;
	private boolean												mNeedUpdate;
	private boolean												mNeedRepaint;
	private boolean												mIsProcessing;

	/**
	 * @param character The character being processed.
	 * @return The thread that does the processing.
	 */
	public static PrerequisitesThread getThread(GURPSCharacter character) {
		synchronized (MAP) {
			return MAP.get(character);
		}
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
		mCharacter.addTarget(this, Profile.ID_TECH_LEVEL, GURPSCharacter.ID_STRENGTH, GURPSCharacter.ID_DEXTERITY, GURPSCharacter.ID_INTELLIGENCE, GURPSCharacter.ID_HEALTH, GURPSCharacter.ID_WILL, GURPSCharacter.ID_PERCEPTION, Spell.ID_NAME, Spell.ID_COLLEGE, Spell.ID_POINTS, Spell.ID_LIST_CHANGED, Skill.ID_NAME, Skill.ID_SPECIALIZATION, Skill.ID_LEVEL, Skill.ID_RELATIVE_LEVEL, Skill.ID_ENCUMBRANCE_PENALTY, Skill.ID_POINTS, Skill.ID_TECH_LEVEL, Skill.ID_LIST_CHANGED, Advantage.ID_NAME, Advantage.ID_LEVELS, Advantage.ID_LIST_CHANGED, Equipment.ID_EXTENDED_WEIGHT, Equipment.ID_STATE, Equipment.ID_QUANTITY, Equipment.ID_LIST_CHANGED);
		Preferences.getInstance().getNotifier().add(this, SheetPreferences.OPTIONAL_IQ_RULES_PREF_KEY, SheetPreferences.OPTIONAL_MODIFIER_RULES_PREF_KEY);
		synchronized (MAP) {
			MAP.put(mCharacter, this);
		}
	}

	@Override
	public void run() {
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
						processRows(mCharacter.getEquipmentIterator());
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
			// Someone is trying to terminate us... let them.
		}
		mNeedUpdate = mIsProcessing = false;
		Preferences.getInstance().getNotifier().remove(this);
		synchronized (MAP) {
			MAP.remove(mCharacter);
		}
	}

	private void processFeatures() throws Exception {
		HashMap<String, ArrayList<Feature>> map = new HashMap<>();
		buildFeatureMap(map, mCharacter.getAdvantagesIterator());
		buildFeatureMap(map, mCharacter.getSkillsIterator());
		buildFeatureMap(map, mCharacter.getSpellsIterator());
		buildFeatureMap(map, mCharacter.getEquipmentIterator());
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
				processFeature(map, row instanceof Advantage ? ((Advantage) row).getLevels() : 0, feature);
			}
			if (row instanceof Advantage) {
				Advantage advantage = (Advantage) row;
				for (Bonus bonus : advantage.getCRAdj().getBonuses(advantage.getCR())) {
					processFeature(map, 0, bonus);
				}
				for (Modifier modifier : advantage.getModifiers()) {
					if (modifier.isEnabled()) {
						for (Feature feature : modifier.getFeatures()) {
							processFeature(map, modifier.getLevels(), feature);
						}
					}
				}
			}
			checkIfUpdated();
		}
	}

	private static void processFeature(HashMap<String, ArrayList<Feature>> map, int levels, Feature feature) {
		String key = feature.getKey().toLowerCase();
		ArrayList<Feature> list = map.get(key);
		if (list == null) {
			list = new ArrayList<>(1);
			map.put(key, list);
		}
		if (feature instanceof Bonus) {
			((Bonus) feature).getAmount().setLevel(levels);
		}
		list.add(feature);
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
			builder.setLength(0);
			boolean satisfied = row.getPrereqs().satisfied(mCharacter, row, builder, "<li>"); //$NON-NLS-1$
			if (satisfied && row instanceof Technique) {
				satisfied = ((Technique) row).satisfied(builder, "<li>"); //$NON-NLS-1$
			}
			if (row.isSatisfied() != satisfied) {
				row.setSatisfied(satisfied);
				mNeedRepaint = true;
			}
			if (!satisfied) {
				builder.insert(0, "<html><head>Reason</head><body><ul>"); //$NON-NLS-1$
				builder.append("</ul></body></html>"); //$NON-NLS-1$
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

	@Override
	public void handleNotification(Object producer, String type, Object data) {
		if (SheetPreferences.OPTIONAL_IQ_RULES_PREF_KEY.equals(type)) {
			mCharacter.updateWillAndPerceptionDueToOptionalIQRuleUseChange();
		} else if (SheetPreferences.OPTIONAL_MODIFIER_RULES_PREF_KEY.equals(type)) {
			mCharacter.notifySingle(Advantage.ID_LIST_CHANGED, null);
		}
		markForUpdate();
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}
}
