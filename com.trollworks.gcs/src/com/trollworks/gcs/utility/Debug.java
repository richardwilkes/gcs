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

package com.trollworks.gcs.utility;

import com.trollworks.gcs.utility.text.Numbers;

/** Provides various debugging utilities. */
public final class Debug {
    /** Controls whether we are in 'development mode' or not. */
    public static final boolean DEV_MODE = false;

    /**
     * Retrieves the specified key, looking first in the system properties and falling back to the
     * system environment if it is not set in the system properties.
     *
     * @param key The key to check.
     * @return The value of the specified key, or {@code null} if it has not been defined.
     */
    public static String getPropertyOrEnvironmentSetting(String key) {
        String value = System.getProperty(key);
        if (value == null) {
            value = System.getenv(key);
        }
        return value;
    }

    /**
     * Determines whether the specified key is set, looking first in the system properties and
     * falling back to the system environment if it is not set at all in the system properties.
     *
     * @param key The key to check.
     * @return {@code true} if the key is enabled.
     */
    public static boolean isKeySet(String key) {
        return Numbers.extractBoolean(getPropertyOrEnvironmentSetting(key));
    }

    /**
     * Extracts the class name, message and stack trace from the specified {@link Throwable}. The
     * stack trace will be formatted such that Eclipse's console will make each node into a
     * hyperlink.
     *
     * @param throwable The {@link Throwable} to process.
     * @return The formatted {@link Throwable}.
     */
    public static String toString(Throwable throwable) {
        StringBuilder buffer = new StringBuilder();
        buffer.append(throwable.getClass().getSimpleName());
        buffer.append(": ");
        buffer.append(throwable.getMessage());
        buffer.append(": ");
        stackTrace(throwable, buffer);
        return buffer.toString();
    }

    /**
     * Extracts a stack trace for the calling site. The stack trace will be formatted such that
     * Eclipse's console will make each node into a hyperlink.
     *
     * @return The formatted stack trace.
     */
    public static String stackTrace() {
        return stackTrace(new Exception(), 1, new StringBuilder()).toString();
    }

    /**
     * Extracts a stack trace from the specified {@link Throwable}. The stack trace will be
     * formatted such that Eclipse's console will make each node into a hyperlink.
     *
     * @param throwable The {@link Throwable} to process.
     * @param buffer    The buffer to store the result in.
     * @return The {@link StringBuilder} that was passed in.
     */
    public static StringBuilder stackTrace(Throwable throwable, StringBuilder buffer) {
        return stackTrace(throwable, 0, buffer);
    }

    /**
     * Extracts a stack trace from the specified {@link Throwable}. The stack trace will be
     * formatted such that Eclipse's console will make each node into a hyperlink.
     *
     * @param throwable The {@link Throwable} to process.
     * @param startAt   The point in the stack to start processing.
     * @param buffer    The buffer to store the result in.
     * @return The {@link StringBuilder} that was passed in.
     */
    public static StringBuilder stackTrace(Throwable throwable, int startAt, StringBuilder buffer) {
        StackTraceElement[] stackTrace = throwable.getStackTrace();
        int                 length     = stackTrace.length;
        for (int i = startAt; i < length; i++) {
            if (i > startAt) {
                buffer.append(" < ");
            }
            buffer.append('(');
            buffer.append(stackTrace[i].getFileName());
            buffer.append(':');
            buffer.append(stackTrace[i].getLineNumber());
            buffer.append(')');
        }
        return buffer;
    }
}
