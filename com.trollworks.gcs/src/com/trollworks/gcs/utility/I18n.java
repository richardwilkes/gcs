/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility;

import com.trollworks.gcs.app.GCS;
import com.trollworks.gcs.utility.text.Text;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;

/** Provides localization support via a single directory of translation files. */
public class I18n {
    public static final String                           EXTENSION = ".i18n";
    private static      I18n                             GLOBAL;
    private             Map<String, Map<String, String>> mLangMap  = new HashMap<>();
    private             Map<String, List<String>>        mHierMap  = new HashMap<>();
    private             Locale                           mLocale   = Locale.getDefault();
    private             String                           mLang     = mLocale.toString().toLowerCase();

    /** @return the global {@link I18n} object. */
    public static synchronized I18n getGlobal() {
        if (GLOBAL == null) {
            GLOBAL = new I18n(null);
        }
        return GLOBAL;
    }

    /** @param i18n the {@link I18n} to use as the global one. */
    public static synchronized void setGlobal(I18n i18n) {
        GLOBAL = i18n;
    }

    /**
     * Use the global {@link I18n} to localize some text.
     *
     * @param str the text to localize.
     * @return the localized version if one exists, or the original text if not.
     */
    public static String Text(String str) {
        return getGlobal().text(str);
    }

    /**
     * Creates a new I18n from the files at 'dir'.
     *
     * @param dir the directory to scan for localization files. If null, then a directory named
     *            'i18n' off of the AppHome.get() will be used.
     */
    public I18n(Path dir) {
        try {
            if (dir == null) {
                dir = GCS.APP_HOME_PATH.resolve("i18n");
                if (!Files.isDirectory(dir)) {
                    dir = Platform.isMacintosh() ? Paths.get(System.getProperty("java.home")).resolve("../../../app/i18n") : GCS.APP_HOME_PATH.resolve("app/i18n");
                }
            }
            if (Files.isDirectory(dir)) {
                Files.list(dir).forEach(path -> {
                    if (path.toString().toLowerCase().endsWith(EXTENSION)) {
                        try {
                            kvInfo kv = new kvInfo();
                            kv.translations = new HashMap<>();
                            Files.lines(path).forEachOrdered(line -> {
                                kv.line++;
                                if (line.startsWith("k:")) {
                                    if (kv.last == 'v') {
                                        if (kv.translations.containsKey(kv.key)) {
                                            System.err.println("ignoring duplicate key on line " + kv.lastKeyLineStart);
                                        } else {
                                            kv.translations.put(kv.key, kv.value);
                                        }
                                        kv.key = null;
                                        kv.value = null;
                                    }
                                    if (kv.key == null) {
                                        kv.key = Text.unquote(line.substring(2));
                                        kv.lastKeyLineStart = kv.line;
                                    } else {
                                        kv.key += '\n';
                                        kv.key += Text.unquote(line.substring(2));
                                    }
                                    kv.last = 'k';
                                } else if (line.startsWith("v:")) {
                                    if (kv.key != null) {
                                        if (kv.value == null) {
                                            kv.value = Text.unquote(line.substring(2));
                                        } else {
                                            kv.value += '\n';
                                            kv.value += Text.unquote(line.substring(2));
                                        }
                                        kv.last = 'v';
                                    } else {
                                        System.err.println("ignoring value with no previous key on line " + kv.line);
                                    }
                                }
                            });
                            if (kv.key != null) {
                                if (kv.value != null) {
                                    if (kv.translations.containsKey(kv.key)) {
                                        System.err.println("ignoring duplicate key on line " + kv.lastKeyLineStart);
                                    } else {
                                        kv.translations.put(kv.key, kv.value);
                                    }
                                } else {
                                    System.err.println("ignoring key with missing value on line " + kv.lastKeyLineStart);
                                }
                            }
                            String key = path.getFileName().toString().toLowerCase();
                            mLangMap.put(key.substring(0, key.length() - EXTENSION.length()), kv.translations);
                        } catch (IOException ex) {
                            ex.printStackTrace(System.err);
                        }
                    }
                });
            }
        } catch (Exception ex) {
            ex.printStackTrace(System.err);
        }
    }

    /**
     * @param str the text to localize.
     * @return the localized version if one exists, or the original text if not.
     */
    public String text(String str) {
        for (String lang : getHierarchy()) {
            Map<String, String> kvMap = mLangMap.get(lang);
            if (kvMap != null) {
                String value = kvMap.get(str);
                if (value != null) {
                    return value;
                }
            }
        }
        return str;
    }

    /** @return the current locale for this {@link I18n}. */
    public Locale getLocale() {
        synchronized (mHierMap) {
            return mLocale;
        }
    }

    /** @param locale the locale to set for this {@link I18n}. */
    public void setLocale(Locale locale) {
        synchronized (mHierMap) {
            mLocale = locale;
            mLang = locale.toString().toLowerCase();
        }
    }

    private List<String> getHierarchy() {
        synchronized (mHierMap) {
            if (mHierMap.containsKey(mLang)) {
                return mHierMap.get(mLang);
            }
            String       one  = mLang;
            List<String> list = new ArrayList<>();
            while (true) {
                list.add(one);
                int last = Math.max(one.lastIndexOf('.'), one.lastIndexOf('_'));
                if (last == -1) {
                    break;
                }
                one = one.substring(0, last);
            }
            mHierMap.put(mLang, list);
            return list;
        }
    }

    private static class kvInfo {
        int                 line;
        int                 lastKeyLineStart;
        String              key;
        String              value;
        Map<String, String> translations;
        char                last;
    }
}
