/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.toolkit.ui.Fonts;
import com.trollworks.toolkit.utility.I18n;

import java.awt.Font;

/** GCS-specific fonts. */
public class GCSFonts {
    /** The label font. */
    public static final String KEY_LABEL            = "trollworks.v2.label";
    /** The field font. */
    public static final String KEY_FIELD            = "trollworks.v2.field";
    /** The field notes font. */
    public static final String KEY_FIELD_NOTES      = "trollworks.v2.field.notes";
    /** The technique field font. */
    public static final String KEY_TECHNIQUE_FIELD  = "trollworks.v2.field.technique";
    /** The primary footer font. */
    public static final String KEY_PRIMARY_FOOTER   = "trollworks.v2.footer.primary";
    /** The secondary footer font. */
    public static final String KEY_SECONDARY_FOOTER = "trollworks.v2.footer.secondary";
    /** The notes font. */
    public static final String KEY_NOTES            = "trollworks.v2.notes";

    /** Register our fonts. */
    public static final void register() {
        String name = Fonts.getDefaultFontName();
        Fonts.register(KEY_LABEL, I18n.Text("Labels"), new Font(name, Font.PLAIN, 9));
        Fonts.register(KEY_FIELD, I18n.Text("Fields"), new Font(name, Font.BOLD, 9));
        Fonts.register(KEY_FIELD_NOTES, I18n.Text("Field Notes"), new Font(name, Font.PLAIN, 8));
        Fonts.register(KEY_TECHNIQUE_FIELD, I18n.Text("Technique Fields"), new Font(name, Font.BOLD + Font.ITALIC, 9));
        Fonts.register(KEY_PRIMARY_FOOTER, I18n.Text("Primary Footer"), new Font(name, Font.BOLD, 8));
        Fonts.register(KEY_SECONDARY_FOOTER, I18n.Text("Secondary Footer"), new Font(name, Font.PLAIN, 6));
        Fonts.register(KEY_NOTES, I18n.Text("Notes"), new Font(name, Font.PLAIN, 9));
    }
}
