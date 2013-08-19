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

package com.trollworks.toolkit.notification;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.StringTokenizer;

/** Tracks targets of notifications and provides methods for notifying them. */
public class TKNotifier {
	/** The separator used between parts of a type. */
	public static final String								SEPARATOR		= ".";													//$NON-NLS-1$
	private HashSet<TKNotifierTarget>						mTargets		= new HashSet<TKNotifierTarget>();
	private HashSet<TKBatchNotifierTarget>					mBatchTargets	= new HashSet<TKBatchNotifierTarget>();
	private HashMap<String, HashSet<TKNotifierTarget>>		mProductionMap	= new HashMap<String, HashSet<TKNotifierTarget>>();
	private HashMap<TKNotifierTarget, ArrayList<String>>	mNameMap		= new HashMap<TKNotifierTarget, ArrayList<String>>();
	private TKBatchNotifierTarget[]							mCurrentBatch;
	private int												mBatchLevel;

	/**
	 * Registers a {@link TKNotifierTarget} with this {@link TKNotifier}.
	 * 
	 * @param target The {@link TKNotifierTarget} to register.
	 * @param names The names consumed. Names are hierarchical (separated by {@link #SEPARATOR}),
	 *            so specifying a name of "foo.bar" will consume not only a produced name of
	 *            "foo.bar", but also sub-names, such as "foo.bar.a", but not "foo.bart" or
	 *            "foo.bart.a".
	 */
	public synchronized void add(TKNotifierTarget target, String... names) {
		if (!mTargets.contains(target)) {
			ArrayList<String> normalizedNames = new ArrayList<String>();

			mTargets.add(target);
			if (target instanceof TKBatchNotifierTarget) {
				mBatchTargets.add((TKBatchNotifierTarget) target);
			}
			for (String name : names) {
				name = normalizeName(name);
				if (name.length() > 0) {
					HashSet<TKNotifierTarget> set = mProductionMap.get(name);

					if (set == null) {
						set = new HashSet<TKNotifierTarget>();
						mProductionMap.put(name, set);
					}
					set.add(target);
					normalizedNames.add(name);
				}
			}
			mNameMap.put(target, normalizedNames);
		}
	}

	private String normalizeName(String name) {
		StringTokenizer tokenizer = new StringTokenizer(name, SEPARATOR);
		StringBuilder builder = new StringBuilder();

		while (tokenizer.hasMoreTokens()) {
			if (builder.length() > 0) {
				builder.append('.');
			}
			builder.append(tokenizer.nextToken());
		}
		return builder.toString();
	}

	/**
	 * Un-registers a {@link TKNotifierTarget} from this {@link TKNotifier}.
	 * 
	 * @param target The {@link TKNotifierTarget} to un-register.
	 */
	public synchronized void remove(TKNotifierTarget target) {
		if (mTargets.remove(target)) {
			if (target instanceof TKBatchNotifierTarget) {
				mBatchTargets.remove(target);
			}
			for (String name : mNameMap.get(target)) {
				if (name.length() > 0) {
					HashSet<TKNotifierTarget> set = mProductionMap.get(name);

					if (set != null) {
						set.remove(target);
						if (set.isEmpty()) {
							mProductionMap.remove(name);
						}
					}
				}
			}
			mNameMap.remove(target);
		}
	}

	/**
	 * Sends a notification to all interested {@link TKNotifierTarget}s.
	 * 
	 * @param producer The producer issuing the notification.
	 * @param name The notification name.
	 * @param data Extra data specific to this notification.
	 */
	public void notify(Object producer, String name, Object data) {
		StringTokenizer tokenizer = new StringTokenizer(name, SEPARATOR);
		StringBuilder builder = new StringBuilder();

		while (tokenizer.hasMoreTokens()) {
			HashSet<TKNotifierTarget> set;
			String value;

			builder.append(tokenizer.nextToken());
			value = builder.toString();
			builder.append(SEPARATOR);
			set = mProductionMap.get(value);
			if (set != null) {
				for (TKNotifierTarget target : set.toArray(new TKNotifierTarget[set.size()])) {
					target.handleNotification(producer, name, data);
				}
			}
		}
	}

	/**
	 * Informs all {@link TKBatchNotifierTarget}s that a batch of notifications will be starting.
	 * If a previous call to this method was made without a call to {@link #endBatch()}, then the
	 * batch level will be incremented, but no notifications will be made.
	 */
	public synchronized void startBatch() {
		if (++mBatchLevel == 1) {
			if (!mBatchTargets.isEmpty()) {
				mCurrentBatch = mBatchTargets.toArray(new TKBatchNotifierTarget[mBatchTargets.size()]);
				for (TKBatchNotifierTarget target : mCurrentBatch) {
					target.enterBatchMode();
				}
			}
		}
	}

	/** @return The current batch level. */
	public synchronized int getBatchLevel() {
		return mBatchLevel;
	}

	/**
	 * Informs all {@link TKBatchNotifierTarget}s that were present when {@link #startBatch()} was
	 * called that a batch of notifications just finished. If batch level is still greater than zero
	 * after being decremented, then no notifications will be done.
	 */
	public synchronized void endBatch() {
		if (--mBatchLevel < 1) {
			if (mCurrentBatch != null) {
				for (TKBatchNotifierTarget target : mCurrentBatch) {
					target.leaveBatchMode();
				}
				mCurrentBatch = null;
			}
		}
	}

	/** Removes all targets. */
	public void reset() {
		mTargets.clear();
		mBatchTargets.clear();
		mProductionMap.clear();
		mNameMap.clear();
	}

	/**
	 * Removes all targets except the specified ones.
	 * 
	 * @param exclude The {@link TKNotifierTarget}(s) to exclude.
	 */
	public void reset(TKNotifierTarget... exclude) {
		HashMap<TKNotifierTarget, ArrayList<String>> set = new HashMap<TKNotifierTarget, ArrayList<String>>();

		for (TKNotifierTarget target : exclude) {
			if (target != null) {
				ArrayList<String> names = mNameMap.get(target);

				if (names != null && !names.isEmpty()) {
					set.put(target, names);
				}
			}
		}

		reset();

		for (TKNotifierTarget target : set.keySet()) {
			ArrayList<String> names = set.get(target);

			add(target, names.toArray(new String[names.size()]));
		}
	}
}
