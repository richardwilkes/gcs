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
import com.trollworks.gcs.utility.text.Numbers;
import static java.time.temporal.ChronoField.DAY_OF_MONTH;
import static java.time.temporal.ChronoField.HOUR_OF_DAY;
import static java.time.temporal.ChronoField.MILLI_OF_SECOND;
import static java.time.temporal.ChronoField.MINUTE_OF_HOUR;
import static java.time.temporal.ChronoField.MONTH_OF_YEAR;
import static java.time.temporal.ChronoField.SECOND_OF_MINUTE;
import static java.time.temporal.ChronoField.YEAR;

import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.Instant;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeFormatterBuilder;

/** Provides standardized logging. */
public final class Log {
    private static final String            GCS_LOG_ENV      = "GCS_LOG";
    private static final String            GCS_LOG_FILE     = "gcs.log";
    private static final String            SEPARATOR        = " | ";
    private static final DateTimeFormatter TIMESTAMP_FORMAT = new DateTimeFormatterBuilder().parseCaseInsensitive().parseLenient().appendValue(YEAR, 4).appendLiteral('.').appendValue(MONTH_OF_YEAR, 2).appendLiteral('.').appendValue(DAY_OF_MONTH, 2).appendLiteral(SEPARATOR).appendValue(HOUR_OF_DAY, 2).appendLiteral(':').appendValue(MINUTE_OF_HOUR, 2).appendLiteral(':').appendValue(SECOND_OF_MINUTE, 2).appendLiteral('.').appendValue(MILLI_OF_SECOND, 3).toFormatter();
    private static       PrintStream       OUT;

    static {
        OUT = System.out;
        Path   path     = null;
        String property = System.getProperty(GCS_LOG_ENV, System.getenv(GCS_LOG_ENV));
        if (property != null && !property.isBlank()) {
            path = Paths.get(property);
        } else if (!GCS.VERSION.isZero()) { // When running a dev version, assume the console is always appropriate, since you're likely running from an IDE
            String home = System.getProperty("user.home", ".");
            switch (Platform.getPlatform()) {
            case MAC -> path = Paths.get(home, "Library", "Logs", GCS_LOG_FILE);
            case WINDOWS -> {
                String localAppData = System.getenv("LOCALAPPDATA");
                path = Paths.get(localAppData != null ? localAppData : home, "logs", GCS_LOG_FILE);
            }
            default -> path = Paths.get(home, ".local", "logs", GCS_LOG_FILE);
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

    private Log() {
    }

    /**
     * Logs an error.
     *
     * @param msg The message to log.
     */
    public static void error(String msg) {
        error(msg, null);
    }

    /**
     * Logs an error.
     *
     * @param throwable The {@link Throwable} to log.
     */
    public static void error(Throwable throwable) {
        error(null, throwable);
    }

    /**
     * Logs an error.
     *
     * @param msg       The message to log.
     * @param throwable The {@link Throwable} to log.
     */
    public static void error(String msg, Throwable throwable) {
        post('E', msg, throwable);
    }

    /**
     * Logs a warning.
     *
     * @param msg The message to log.
     */
    public static void warn(String msg) {
        warn(msg, null);
    }

    /**
     * Logs a warning.
     *
     * @param throwable The {@link Throwable} to log.
     */
    public static void warn(Throwable throwable) {
        warn(null, throwable);
    }

    /**
     * Logs a warning.
     *
     * @param msg       The message to log.
     * @param throwable The {@link Throwable} to log.
     */
    public static void warn(String msg, Throwable throwable) {
        post('W', msg, throwable);
    }

    private static void post(char levelCode, String msg, Throwable throwable) {
        StringBuilder buffer = new StringBuilder();
        buffer.append(levelCode);
        buffer.append(SEPARATOR);
        buffer.append(Numbers.formatDateTime(TIMESTAMP_FORMAT, Instant.now().toEpochMilli()));
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
