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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import com.trollworks.gcs.app.GCSFonts;

import java.awt.Color;

import javax.swing.JLabel;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/** A header within the page. */
public class PageHeader extends JLabel {
	/**
	 * Creates a new {@link PageHeader}.
	 * 
	 * @param title The title to use.
	 * @param tooltip The tooltip to use.
	 */
	public PageHeader(String title, String tooltip) {
		super(title, SwingConstants.CENTER);
		setFont(UIManager.getFont(GCSFonts.KEY_LABEL));
		setForeground(Color.white);
		setToolTipText(tooltip);
	}
}
