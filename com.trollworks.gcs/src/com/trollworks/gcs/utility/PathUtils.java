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

import java.io.File;
import java.nio.file.Path;
import java.util.Arrays;
import java.util.HashSet;
import java.util.Set;

/** Provides standard file path manipulation facilities. */
public final class PathUtils {
    private static final char[]         INVALID_CHARACTERS;
    private static final Set<Character> INVALID_CHARACTER_SET;
    private static final String[]       INVALID_BASENAMES;
    private static final String[]       INVALID_FULLNAMES;

    static {
        if (Platform.isWindows()) {
            INVALID_CHARACTERS = new char[]{'\\', '/', ':', '*', '?', '"', '<', '>', '|'};
            INVALID_BASENAMES = new String[]{"aux", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "con", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9", "nul", "prn"};
            INVALID_FULLNAMES = new String[]{"clock$"};
        } else if (Platform.isMacintosh()) {
            INVALID_CHARACTERS = new char[]{'/', ':', '\0',};
            INVALID_BASENAMES = null;
            INVALID_FULLNAMES = null;
        } else {
            INVALID_CHARACTERS = new char[]{'/', '\0',};
            INVALID_BASENAMES = null;
            INVALID_FULLNAMES = null;
        }
        Set<Character> exclude = new HashSet<>();
        for (char ch : INVALID_CHARACTERS) {
            exclude.add(Character.valueOf(ch));
        }
        INVALID_CHARACTER_SET = exclude;
    }

    private PathUtils() {
    }

    /**
     * Ensures that the passed in string has the specified extension on it.
     *
     * @param name      The name to process.
     * @param extension The desired extension.
     * @return A new string with the specified extension.
     */
    public static String enforceExtension(String name, String extension) {
        return enforceExtension(name, extension, false);
    }

    /**
     * Ensures that the passed in string has an extension on it.
     *
     * @param name              The name to process.
     * @param extension         The desired extension.
     * @param onlyIfNoExtension Pass {@code true} if extensions other than the one passed in are
     *                          acceptable.
     * @return A new string with the specified extension.
     */
    public static String enforceExtension(String name, String extension, boolean onlyIfNoExtension) {
        name = name.replace('\\', '/');
        if (extension.charAt(0) != '.') {
            extension = '.' + extension;
        }
        if (!name.endsWith(extension)) {
            int lastDot = name.lastIndexOf('.');

            if (name.lastIndexOf('/') > lastDot) {
                lastDot = -1;
            }
            if (lastDot == -1) {
                return name + extension;
            }
            if (lastDot == name.length() - 1) {
                return name + extension.substring(1);
            }
            if (!onlyIfNoExtension) {
                return name.substring(0, lastDot) + extension;
            }
        }
        return name;
    }

    /**
     * @param path The path to operate on.
     * @return The extension of the path name, excluding the initial ".".
     */
    public static String getExtension(Path path) {
        return getExtension(path != null ? path.toString() : null);
    }

    /**
     * @param path The path to operate on.
     * @return The extension of the path name, excluding the initial ".".
     */
    public static String getExtension(String path) {
        path = getLeafName(path);
        if (path != null) {
            int dot = path.lastIndexOf('.');
            if (dot != -1 && dot + 1 < path.length()) {
                return path.substring(dot + 1);
            }
        }
        return "";
    }

    /**
     * @param file The file to operate on.
     * @return A full path from a file.
     */
    public static String getFullPath(File file) {
        if (file != null) {
            return normalizeFullPath(file.getAbsolutePath().replace('\\', '/'));
        }
        return null;
    }

    /**
     * @param path The path to process.
     * @return The leaf portion of the path name (everything to the right of the last path
     *         separator).
     */
    public static String getLeafName(String path) {
        return getLeafName(path, true);
    }

    /**
     * @param path             The path to process.
     * @param includeExtension Pass in {@code true} to leave the extension on the name or {@code
     *                         false} to strip it off.
     * @return The leaf portion of the path name (everything to the right of the last path
     *         separator).
     */
    public static String getLeafName(Path path, boolean includeExtension) {
        return getLeafName(path.getFileName().toString(), includeExtension);
    }

    /**
     * @param path             The path to process.
     * @param includeExtension Pass in {@code true} to leave the extension on the name or {@code
     *                         false} to strip it off.
     * @return The leaf portion of the path name (everything to the right of the last path
     *         separator).
     */
    public static String getLeafName(String path, boolean includeExtension) {
        if (path != null) {
            int index;

            path = path.replace('\\', '/');
            index = path.lastIndexOf('/');
            if (index != -1) {
                if (index == path.length() - 1) {
                    return "";
                }
                path = path.substring(index + 1);
            }

            if (!includeExtension) {
                index = path.lastIndexOf('.');
                if (index != -1) {
                    path = path.substring(0, index);
                }
            }
            return path;
        }
        return null;
    }

    /**
     * Normalizes full path names by resolving . and .. path portions.
     *
     * @param path The path to operate on.
     * @return The normalized path.
     */
    public static String normalizeFullPath(String path) {
        if (path != null) {
            int           index;
            StringBuilder buffer;
            char          ch;

            path = path.replace('\\', '/');

            do {
                index = path.indexOf("/./");
                if (index != -1) {
                    buffer = new StringBuilder(path);
                    buffer.delete(index, index + 2);
                    path = buffer.toString();
                }
            } while (index != -1);

            do {
                index = path.indexOf("/../");
                if (index != -1) {
                    int length = 3;

                    buffer = new StringBuilder(path);

                    while (index > 0) {
                        ch = buffer.charAt(--index);
                        length++;
                        if (ch == '/') {
                            break;
                        } else if (ch == ':') {
                            index++;
                            length--;
                            break;
                        }
                    }

                    buffer.delete(index, index + length);
                    path = buffer.toString();
                }
            } while (index != -1);

            if (path.endsWith("/.")) {
                path = path.substring(0, path.length() - 2);
            }

            if (path.endsWith("/..")) {
                index = path.length() - 3;

                while (index > 0) {
                    ch = path.charAt(--index);
                    if (ch == '/' || ch == ':') {
                        break;
                    }
                }

                path = path.substring(0, index);
            }

            if (path.length() > 1 && path.charAt(1) == ':') {
                path = Character.toUpperCase(path.charAt(0)) + ":" + path.substring(2);
            }
        }

        return path;
    }

    /**
     * @param name The name to check. This should be just the name and no path components.
     * @return {@code true} if the name is valid as a file name on your platform.
     */
    public static boolean isNameValidForFile(String name) {
        if (name == null || name.isEmpty() || ".".equals(name) || "..".equals(name)) {
            return false;
        }
        for (char ch : INVALID_CHARACTERS) {
            if (name.indexOf(ch) != -1) {
                return false;
            }
        }
        if (Platform.isWindows()) {
            char lastChar = name.charAt(name.length() - 1);
            if (lastChar == '.' || Character.isWhitespace(lastChar)) {
                return false;
            }
            int    dot      = name.indexOf('.');
            String basename = dot == -1 ? name : name.substring(0, dot);
            if (Arrays.binarySearch(INVALID_BASENAMES, basename.toLowerCase()) >= 0) {
                return false;
            }
            return Arrays.binarySearch(INVALID_FULLNAMES, name.toLowerCase()) < 0;
        }
        return true;
    }

    /**
     * Cleans a file name (without the path) to be suitable for use on the current file system.
     *
     * @param name The file name to clean.
     * @return A viable file name.
     */
    public static String cleanNameForFile(String name) {
        if (isNameValidForFile(name)) {
            return name;
        }
        StringBuilder buffer = new StringBuilder();
        int           max    = name.length();
        for (int i = 0; i < max; i++) {
            char ch = name.charAt(i);
            if (!INVALID_CHARACTER_SET.contains(Character.valueOf(ch))) {
                buffer.append(ch);
            }
        }
        name = buffer.toString();
        if (isNameValidForFile(name)) {
            return name;
        }
        return "untitled";
    }
}
