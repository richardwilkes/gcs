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

package com.trollworks.gcs.utility.notification;

import com.trollworks.gcs.utility.Log;

import java.util.Arrays;
import java.util.Comparator;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Set;
import java.util.StringTokenizer;

/** Tracks targets of notifications and provides methods for notifying them. */
public class Notifier implements Comparator<NotifierTarget> {
    /** The separator used between parts of a type. */
    public static final String                           SEPARATOR      = ".";
    private             Set<BatchNotifierTarget>         mBatchTargets  = new HashSet<>();
    private             Map<String, Set<NotifierTarget>> mProductionMap = new HashMap<>();
    private             Map<NotifierTarget, Set<String>> mNameMap       = new HashMap<>();
    private             BatchNotifierTarget[]            mCurrentBatch;
    private             int                              mBatchLevel;
    private             boolean                          mEnabled       = true;

    /**
     * Adds all registrations from the specified {@link Notifier} into this one.
     *
     * @param notifier The {@link Notifier} to use.
     */
    public synchronized void addFrom(Notifier notifier) {
        Map<NotifierTarget, Set<String>> map = new HashMap<>(notifier.mNameMap);
        for (Entry<NotifierTarget, Set<String>> entry : map.entrySet()) {
            Set<String> set = entry.getValue();
            add(entry.getKey(), set.toArray(new String[0]));
        }
    }

    /**
     * Registers a {@link NotifierTarget} with this {@link Notifier}.
     *
     * @param target The {@link NotifierTarget} to register.
     * @param names  The names consumed. Names are hierarchical (separated by {@link #SEPARATOR}),
     *               so specifying a name of "foo.bar" will consume not only a produced name of
     *               "foo.bar", but also sub-names, such as "foo.bar.a", but not "foo.bart" or
     *               "foo.bart.a".
     */
    public synchronized void add(NotifierTarget target, String... names) {
        Set<String> normalizedNames = mNameMap.computeIfAbsent(target, k -> new HashSet<>());
        if (target instanceof BatchNotifierTarget) {
            mBatchTargets.add((BatchNotifierTarget) target);
        }
        if (names != null) {
            for (String name : names) {
                name = normalizeName(name);
                if (!name.isEmpty()) {
                    mProductionMap.computeIfAbsent(name, k -> new HashSet<>()).add(target);
                    normalizedNames.add(name);
                }
            }
        }
        if (normalizedNames.isEmpty()) {
            mNameMap.remove(target);
        }
    }

    private static String normalizeName(String name) {
        StringTokenizer tokenizer = new StringTokenizer(name, SEPARATOR);
        StringBuilder   builder   = new StringBuilder();
        while (tokenizer.hasMoreTokens()) {
            if (!builder.isEmpty()) {
                builder.append(SEPARATOR);
            }
            builder.append(tokenizer.nextToken());
        }
        return builder.toString();
    }

    /**
     * Un-registers a {@link NotifierTarget} from this {@link Notifier}.
     *
     * @param target The {@link NotifierTarget} to un-register.
     */
    public synchronized void remove(NotifierTarget target) {
        if (mNameMap.containsKey(target)) {
            if (target instanceof BatchNotifierTarget) {
                mBatchTargets.remove(target);
            }
            for (String name : mNameMap.get(target)) {
                if (!name.isEmpty()) {
                    Set<NotifierTarget> set = mProductionMap.get(name);
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
     * @return Whether or not this {@link Notifier} is currently enabled (and can therefore be used
     *         to notify {@link NotifierTarget}s).
     */
    public synchronized boolean isEnabled() {
        return mEnabled;
    }

    /**
     * @param enabled Whether or not this {@link Notifier} is currently enabled (and can therefore
     *                be used to notify {@link NotifierTarget}s).
     */
    public synchronized void setEnabled(boolean enabled) {
        mEnabled = enabled;
    }

    /**
     * Sends a notification to all interested {@link NotifierTarget}s.
     *
     * @param producer The producer issuing the notification.
     * @param name     The notification name.
     */
    public void notify(Object producer, String name) {
        notify(producer, name, null);
    }

    /**
     * Sends a notification to all interested {@link NotifierTarget}s.
     *
     * @param producer The producer issuing the notification.
     * @param name     The notification name.
     * @param data     Extra data specific to this notification.
     */
    public void notify(Object producer, String name, Object data) {
        if (isEnabled()) {
            StringTokenizer tokenizer = new StringTokenizer(name, SEPARATOR);
            StringBuilder   builder   = new StringBuilder();
            while (tokenizer.hasMoreTokens()) {
                builder.append(tokenizer.nextToken());
                String value = builder.toString();
                builder.append(SEPARATOR);
                NotifierTarget[] targets = getTargets(value);
                if (targets != null) {
                    Arrays.sort(targets, this);
                    for (NotifierTarget target : targets) {
                        try {
                            target.handleNotification(producer, name, data);
                        } catch (Throwable throwable) {
                            Log.error(throwable);
                        }
                    }
                }
            }
        }
    }

    private synchronized NotifierTarget[] getTargets(String value) {
        Set<NotifierTarget> set = mProductionMap.get(value);
        if (set != null) {
            int size = set.size();
            if (size > 0) {
                return set.toArray(new NotifierTarget[size]);
            }
        }
        return null;
    }

    /**
     * Informs all {@link BatchNotifierTarget}s that a batch of notifications will be starting. If a
     * previous call to this method was made without a call to {@link #endBatch()}, then the batch
     * level will be incremented, but no notifications will be made.
     */
    public synchronized void startBatch() {
        if (isEnabled()) {
            if (++mBatchLevel == 1) {
                if (!mBatchTargets.isEmpty()) {
                    mCurrentBatch = mBatchTargets.toArray(new BatchNotifierTarget[0]);
                    for (BatchNotifierTarget target : mCurrentBatch) {
                        try {
                            target.enterBatchMode();
                        } catch (Throwable throwable) {
                            Log.error(throwable);
                        }
                    }
                }
            }
        }
    }

    /** @return The current batch level. */
    public synchronized int getBatchLevel() {
        return mBatchLevel;
    }

    /**
     * Informs all {@link BatchNotifierTarget}s that were present when {@link #startBatch()} was
     * called that a batch of notifications just finished. If batch level is still greater than zero
     * after being decremented, then no notifications will be done.
     */
    public synchronized void endBatch() {
        if (isEnabled()) {
            if (--mBatchLevel < 1) {
                if (mCurrentBatch != null) {
                    for (BatchNotifierTarget target : mCurrentBatch) {
                        try {
                            target.leaveBatchMode();
                        } catch (Throwable throwable) {
                            Log.error(throwable);
                        }
                    }
                    mCurrentBatch = null;
                }
            }
        }
    }

    /**
     * Removes all targets except the specified ones.
     *
     * @param exclude The {@link NotifierTarget}(s) to exclude.
     */
    public synchronized void reset(NotifierTarget... exclude) {
        Map<NotifierTarget, Set<String>> set = new HashMap<>();
        if (exclude != null) {
            for (NotifierTarget target : exclude) {
                if (target != null) {
                    Set<String> names = mNameMap.get(target);
                    if (names != null && !names.isEmpty()) {
                        set.put(target, names);
                    }
                }
            }
        }
        mBatchTargets.clear();
        mProductionMap.clear();
        mNameMap.clear();
        for (Entry<NotifierTarget, Set<String>> entry : set.entrySet()) {
            Set<String> names = entry.getValue();
            add(entry.getKey(), names.toArray(new String[0]));
        }
    }

    @Override
    public int compare(NotifierTarget t1, NotifierTarget t2) {
        return Integer.compare(t1.getNotificationPriority(), t2.getNotificationPriority());
    }
}
