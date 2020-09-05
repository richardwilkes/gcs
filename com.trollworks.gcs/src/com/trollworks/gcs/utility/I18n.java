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
import com.trollworks.gcs.utility.text.Text;

import java.io.BufferedReader;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;

/** Provides localization support via a single directory of translation files. */
public final class I18n {
    private static Map<String, Map<Integer, String>> TRANSLATIONS;

    private I18n() {
    }

    /**
     * NOTE: The name of this class and function MUST be exactly "I18n.Text" (case-sensitive), as
     * the bundler scans the source trees looking for these calls to generate the localization
     * template file. Also, call sites should only provide a simple string as a parameter to this
     * function -- no variables, functions calls, continuation marks to allow breaking it up onto
     * multiple lines, etc.
     *
     * @param str the text to localize.
     * @return the localized version if one exists, or the original text if not.
     */
    public static String Text(String str) {
        return TextWithContext(0, str);
    }

    /**
     * NOTE: The name of this class and function MUST be exactly "I18n.TextWithContext"
     * (case-sensitive), as the bundler scans the source trees looking for these calls to generate
     * the localization template file. Also, call sites should only provide simple strings as
     * parameters to this function -- no variables, functions calls, continuation marks to allow
     * breaking it up onto multiple lines, etc.
     *
     * @param context the context the text is used in. A value of 0 is the default context and only
     *                single digits are permitted (i.e. values 0-9). This context will be used to
     *                select a more specific variant of the value, if present. If not, it will
     *                default back to the default context of 0.
     * @param str     the text to localize.
     * @return the localized version if one exists, or the original text if not.
     */
    public static String TextWithContext(int context, String str) {
        if (context < 0 || context > 9) {
            context = 0;
        }
        Map<Integer, String> m = TRANSLATIONS.get(str);
        if (m != null) {
            String value = m.get(Integer.valueOf(context));
            if (value != null) {
                return value;
            }
            if (context != 0) {
                value = m.get(Integer.valueOf(0));
                if (value != null) {
                    return value;
                }
            }
        }
        return str;
    }

    /** Initialize the localization data. */
    public static void initialize() {
        TRANSLATIONS = new HashMap<>();

        Path   base   = null;
        String envVar = System.getenv("GCS_TRANSLATIONS");
        if (envVar != null && !envVar.isBlank()) {
            try {
                base = Paths.get(envVar).normalize().toAbsolutePath();
                if (!Files.isDirectory(base)) {
                    base = null;
                }
            } catch (Exception ex) {
                base = null;
            }
        }
        if (base == null) {
            if (GCS.VERSION.isZero()) { // Development mode, just use the working dir
                base = Paths.get(".").toAbsolutePath();
            } else {
                base = Paths.get(System.getProperty("java.home"));
                if (Platform.isMacintosh()) {
                    base = base.resolve("../..");
                }
                base = base.resolve("../app/i18n");
            }
            base = base.normalize();
        }

        String filename = Locale.getDefault().toString();
        while (true) {
            Path path = base.resolve(filename + ".i18n");
            if (Files.isRegularFile(path) && Files.isReadable(path)) {
                try (BufferedReader in = Files.newBufferedReader(path)) {
                    int           lineNum          = 0;
                    int           lastKeyLineStart = 0;
                    StringBuilder keyBuilder       = null;
                    StringBuilder valueBuilder     = null;
                    char          last             = 0;
                    int           context          = 0;
                    String        line             = in.readLine();
                    while (line != null) {
                        lineNum++;
                        if (line.startsWith("k:")) {
                            if (last == 'v') {
                                addContextValue(keyBuilder, context, valueBuilder);
                                keyBuilder = null;
                                valueBuilder = null;
                            }
                            if (keyBuilder == null) {
                                keyBuilder = new StringBuilder(Text.unquote(line.substring(2)));
                                lastKeyLineStart = lineNum;
                            } else {
                                keyBuilder.append("\n").append(Text.unquote(line.substring(2)));
                            }
                            last = 'k';
                            context = 0;
                        } else {
                            for (int i = 0; i < 10; i++) {
                                String lookFor = i == 0 ? "v:" : "v" + i + ":";
                                if (line.startsWith(lookFor)) {
                                    if (keyBuilder != null) {
                                        if (last == 'v' && context != i) {
                                            addContextValue(keyBuilder, context, valueBuilder);
                                            valueBuilder = null;
                                        }
                                        if (valueBuilder == null) {
                                            valueBuilder = new StringBuilder(Text.unquote(line.substring(lookFor.length())));
                                        } else {
                                            valueBuilder.append("\n").append(Text.unquote(line.substring(lookFor.length())));
                                        }
                                        last = 'v';
                                        context = i;
                                    } else {
                                        Log.warn("ignoring value with no previous key on line " + lineNum);
                                    }
                                    break;
                                }
                            }
                        }
                        line = in.readLine();
                    }
                    if (keyBuilder != null) {
                        if (valueBuilder != null) {
                            addContextValue(keyBuilder, context, valueBuilder);
                        } else {
                            Log.warn("ignoring key with missing value on line " + lastKeyLineStart);
                        }
                    }
                } catch (IOException ex) {
                    Log.error(ex);
                }
            }
            int last = Math.max(filename.lastIndexOf('.'), filename.lastIndexOf('_'));
            if (last == -1) {
                break;
            }
            filename = filename.substring(0, last);
        }
    }

    private static void addContextValue(StringBuilder keyBuilder, int context, StringBuilder valueBuilder) {
        Map<Integer, String> m      = TRANSLATIONS.computeIfAbsent(keyBuilder.toString(), k -> new HashMap<>());
        Integer              ctxKey = Integer.valueOf(context);
        if (!m.containsKey(ctxKey)) {
            m.put(ctxKey, valueBuilder.toString());
        }
    }
}
