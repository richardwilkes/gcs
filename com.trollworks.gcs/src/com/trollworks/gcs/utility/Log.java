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

import com.trollworks.gcs.GCS;

import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.text.SimpleDateFormat;
import java.util.Date;

/** Provides standardized logging. */
public class Log {
    private static final String           GCS_LOG_FILE = "gcs.log";
    private static final String           SEPARATOR    = " | ";
    private static final SimpleDateFormat FORMAT       = new SimpleDateFormat("yyyy.MM.dd" + SEPARATOR + "HH:mm:ss.SSS");
    private static       PrintStream      OUT;

    static {
        OUT = System.out;
        Path   path     = null;
        String property = Debug.getPropertyOrEnvironmentSetting("GCS_LOG");
        if (property != null && !property.isBlank()) {
            path = Paths.get(property);
        } else if (!GCS.VERSION.isZero()) { // When running a dev version, assume the console is always appropriate, since you're likely running from an IDE
            String home = System.getProperty("user.home", ".");
            switch (Platform.getPlatform()) {
            case MAC:
                path = Paths.get(home, "Library", "Logs", GCS_LOG_FILE);
                break;
            case WINDOWS:
                String localAppData = System.getenv("LOCALAPPDATA");
                path = Paths.get(localAppData != null ? localAppData : home, "logs", GCS_LOG_FILE);
                break;
            default:
                path = Paths.get(home, ".local", "logs", GCS_LOG_FILE);
                break;
            }
        }
        if (path != null) {
            try {
                path = path.normalize().toAbsolutePath();
                Files.createDirectories(path.getParent());
                OUT = new PrintStream(path.toFile(), StandardCharsets.UTF_8);
            } catch (Throwable throwable) {
                error("Unable to redirect log to " + path, throwable);
            }
        }
    }

    /**
     * Logs an error.
     *
     * @param msg The message to log.
     */
    public static final void error(String msg) {
        error(msg, null);
    }

    /**
     * Logs an error.
     *
     * @param throwable The {@link Throwable} to log.
     */
    public static final void error(Throwable throwable) {
        error(null, throwable);
    }

    /**
     * Logs an error.
     *
     * @param msg       The message to log.
     * @param throwable The {@link Throwable} to log.
     */
    public static final void error(String msg, Throwable throwable) {
        post('E', msg, throwable);
    }

    /**
     * Logs a warning.
     *
     * @param msg The message to log.
     */
    public static final void warn(String msg) {
        warn(msg, null);
    }

    /**
     * Logs a warning.
     *
     * @param throwable The {@link Throwable} to log.
     */
    public static final void warn(Throwable throwable) {
        warn(null, throwable);
    }

    /**
     * Logs a warning.
     *
     * @param msg       The message to log.
     * @param throwable The {@link Throwable} to log.
     */
    public static final void warn(String msg, Throwable throwable) {
        post('W', msg, throwable);
    }

    private static void post(char levelCode, String msg, Throwable throwable) {
        StringBuilder buffer = new StringBuilder();
        buffer.append(levelCode);
        buffer.append(SEPARATOR);
        buffer.append(FORMAT.format(new Date()));
        buffer.append(SEPARATOR);
        if (msg != null && !msg.isEmpty()) {
            buffer.append(msg);
        }
        OUT.println(buffer);
        if (throwable != null) {
            throwable.printStackTrace(OUT);
        }
    }
}
