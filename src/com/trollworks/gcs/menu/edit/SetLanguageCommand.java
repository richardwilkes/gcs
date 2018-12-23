/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.edit;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.io.File;
import java.io.PrintWriter;
import java.util.Locale;

/** Provides a menu item to set a language for the user interface. */
public class SetLanguageCommand extends Command {
    @Localize("Automatic")
    private static String AUTO;
    @Localize("English")
    private static String ENGLISH;
    @Localize("German")
    private static String GERMAN;
    @Localize("Russian")
    private static String RUSSIAN;
    @Localize("Spanish")
    private static String SPANISH;
    @Localize("You will need to restart GCS for this change to take effect.")
    private static String NOTICE;

    static {
        Localization.initialize();
    }

    private static final String CMD = "SetLanguage:"; //$NON-NLS-1$
    private Locale              mLocale;

    private static String getTitle(Locale locale) {
        if (locale == null) {
            return AUTO;
        }
        switch (locale.getLanguage()) {
        case "en": //$NON-NLS-1$
            return ENGLISH;
        case "de": //$NON-NLS-1$
            return GERMAN;
        case "ru": //$NON-NLS-1$
            return RUSSIAN;
        case "es": //$NON-NLS-1$
            return SPANISH;
        default:
            return locale.getLanguage();
        }
    }

    public SetLanguageCommand(Locale locale) {
        super(getTitle(locale), CMD + getTitle(locale));
        mLocale = locale;
    }

    private boolean matches(Locale locale) {
        return mLocale == null ? locale == null : mLocale.equals(locale);
    }

    @Override
    public void adjust() {
        Locale  locale  = Localization.getLocaleOverride();
        boolean matches = matches(locale);
        setEnabled(!matches);
        setMarked(matches);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Locale locale = Localization.getLocaleOverride();
        if (!matches(locale)) {
            File file = Localization.getLocaleOverrideFile();
            if (mLocale == null) {
                file.delete();
            } else {
                try (PrintWriter out = new PrintWriter(file)) {
                    out.println(mLocale.getLanguage());
                } catch (Exception ex) {
                    // Ignore
                }
            }
            WindowUtils.showWarning(null, NOTICE);
        }
    }
}
