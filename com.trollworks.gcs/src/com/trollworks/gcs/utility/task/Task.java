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

import com.trollworks.gcs.utility.Log;

import java.util.HashSet;
import java.util.Set;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public class Task implements Runnable {
    private static final ScheduledExecutorService EXECUTOR = Executors.newScheduledThreadPool(Runtime.getRuntime().availableProcessors() + 1);
    private static final Set<Object>              PENDING  = new HashSet<>();
    private              Runnable                 mTask;
    private              Object                   mKey;
    private              long                     mPeriod  = -1;
    private              boolean                  mWasCancelled;
    private              boolean                  mWasExecuted;

    Task(Runnable runnable, Object key) {
        mTask = runnable;
        mKey = key;
    }

    void schedule(long delay, TimeUnit delayUnits) {
        if (mKey != null) {
            synchronized (PENDING) {
                if (!PENDING.add(mKey)) {
                    mWasCancelled = true;
                    return;
                }
            }
        }
        EXECUTOR.schedule(this, delay, delayUnits);
    }

    void schedulePeriodic(long period, TimeUnit periodUnits) {
        mPeriod = TimeUnit.MILLISECONDS.convert(period, periodUnits);
        schedule(period, periodUnits);
    }

    @Override
    public void run() {
        synchronized (this) {
            if (mWasCancelled) {
                return;
            }
            mWasExecuted = true;
        }
        try {
            if (mKey != null) {
                synchronized (PENDING) {
                    PENDING.remove(mKey);
                }
            }
            if (isPeriodic()) {
                long next = System.currentTimeMillis() + mPeriod;
                mTask.run();
                EXECUTOR.schedule(this, Math.max(next - System.currentTimeMillis(), 0), TimeUnit.MILLISECONDS);
            } else {
                mTask.run();
            }
        } catch (Throwable throwable) {
            cancel();
            Log.error(throwable);
        }
    }

    public boolean isPeriodic() {
        return mPeriod != -1;
    }

    /** @return {@code true} if the task was cancelled and will not be executed. */
    public synchronized boolean wasCancelled() {
        return mWasCancelled;
    }

    /**
     * @return {@code true} if the task was successfully cancelled and will not be executed.
     */
    public synchronized boolean cancel() {
        if (mWasExecuted) {
            return false;
        }
        mWasCancelled = true;
        return true;
    }
}
