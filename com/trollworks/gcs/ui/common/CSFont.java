/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.common;

import com.trollworks.toolkit.utility.TKFont;

import java.awt.Font;

/** Common fonts. */
public class CSFont {
	/** The label font. */
	public static final String	KEY_LABEL				= "Label";				//$NON-NLS-1$
	/** The field font. */
	public static final String	KEY_FIELD				= "Field";				//$NON-NLS-1$
	/** The field notes font. */
	public static final String	KEY_FIELD_NOTES			= "FieldNotes";		//$NON-NLS-1$
	/** The technique field font. */
	public static final String	KEY_TECHNIQUE_FIELD		= "TechniqueField";	//$NON-NLS-1$
	/** The primary footer font. */
	public static final String	KEY_PRIMARY_FOOTER		= "PrimaryFooter";		//$NON-NLS-1$
	/** The secondary footer font. */
	public static final String	KEY_SECONDARY_FOOTER	= "SecondaryFooter";	//$NON-NLS-1$
	/** The notes font. */
	public static final String	KEY_NOTES				= "Notes";				//$NON-NLS-1$

	static {
		String name = TKFont.getDefaultFontName();
		TKFont.register(KEY_LABEL, new Font(name, Font.PLAIN, 9));
		TKFont.register(KEY_FIELD, new Font(name, Font.BOLD, 9));
		TKFont.register(KEY_FIELD_NOTES, new Font(name, Font.PLAIN, 8));
		TKFont.register(KEY_TECHNIQUE_FIELD, new Font(name, Font.BOLD + Font.ITALIC, 9));
		TKFont.register(KEY_PRIMARY_FOOTER, new Font(name, Font.BOLD, 8));
		TKFont.register(KEY_SECONDARY_FOOTER, new Font(name, Font.PLAIN, 6));
		TKFont.register(KEY_NOTES, new Font(name, Font.PLAIN, 9));
	}

	/** Must be called prior to any fonts being used. */
	public static final void initialize() {
		TKFont.loadFromPreferences();
	}
}
