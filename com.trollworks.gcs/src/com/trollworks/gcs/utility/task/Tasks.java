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

package com.trollworks.gcs.utility.task;

import java.util.concurrent.TimeUnit;

public class Tasks {
    /**
     * Execute a {@link Runnable} on a background thread.
     *
     * @param runnable The {@link Runnable} to execute.
     */
    public static void callOnBackgroundThread(Runnable runnable) {
        scheduleOnBackgroundThread(runnable, 0, TimeUnit.MILLISECONDS, null);
    }

    /**
     * Execute a {@link Runnable} on a background thread.
     *
     * @param runnable The {@link Runnable} to execute.
     * @param delay    The number of units to delay before execution begins.
     * @param units    The units the delay parameter has been specified in.
     * @param key      If this is not {@code null}, then the task will only be scheduled to run if
     *                 there isn't one with the same key already scheduled.
     * @return The {@link Task}.
     */
    public static Task scheduleOnBackgroundThread(Runnable runnable, long delay, TimeUnit units, Object key) {
        Task task = new Task(runnable, key);
        task.schedule(delay, units);
        return task;
    }

    /**
     * Repeatedly execute a {@link Runnable} on the UI thread.
     *
     * @param runnable The {@link Runnable} to execute.
     * @param period   The number of units between executions.
     * @param units    The units the delay parameter has been specified in.
     * @param key      If this is not {@code null}, then the task will only be scheduled to run if
     *                 there isn't one with the same key already scheduled.
     * @return The {@link Task}.
     */
    public static final Task scheduleRepeatedlyOnBackgroundThread(Runnable runnable, long period, TimeUnit units, Object key) {
        Task task = new Task(runnable, key);
        task.schedulePeriodic(period, units);
        return task;
    }

    /**
     * Execute a {@link Runnable} on the UI thread.
     *
     * @param runnable The {@link Runnable} to execute.
     * @param delay    The number of units to delay before execution begins.
     * @param units    The units the delay parameter has been specified in.
     * @param key      If this is not {@code null}, then the task will only be scheduled to run if
     *                 there isn't one with the same key already scheduled.
     * @return The {@link Task}.
     */
    public static final Task scheduleOnUIThread(Runnable runnable, long delay, TimeUnit units, Object key) {
        Task task = new UITask(runnable, key);
        task.schedule(delay, units);
        return task;
    }

    /**
     * Repeatedly execute a {@link Runnable} on the UI thread.
     *
     * @param runnable The {@link Runnable} to execute.
     * @param period   The number of units between executions.
     * @param units    The units the delay parameter has been specified in.
     * @param key      If this is not {@code null}, then the task will only be scheduled to run if
     *                 there isn't one with the same key already scheduled.
     * @return The {@link Task}.
     */
    public static final Task scheduleRepeatedlyOnUIThread(Runnable runnable, long period, TimeUnit units, Object key) {
        Task task = new UITask(runnable, key);
        task.schedulePeriodic(period, units);
        return task;
    }
}
