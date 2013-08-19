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

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.common.DataModifiedListener;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import javax.swing.ImageIcon;
import javax.swing.JLabel;
import javax.swing.JRootPane;

/** A toolbar marker that tracks the modified state. */
public class ModifiedMarker extends JLabel implements DataModifiedListener {
	private static String		MSG_NOT_MODIFIED;
	private static String		MSG_MODIFIED;
	private static ImageIcon	ICON_NOT_MODIFIED	= new ImageIcon(Images.getNotModifiedMarker());
	private static ImageIcon	ICON_MODIFIED		= new ImageIcon(Images.getModifiedMarker());

	static {
		LocalizedMessages.initialize(ModifiedMarker.class);
	}

	/** Creates a new {@link ModifiedMarker}. */
	public ModifiedMarker() {
		super(ICON_NOT_MODIFIED);
		setToolTipText(MSG_NOT_MODIFIED);
	}

	public void dataModificationStateChanged(Object obj, boolean modified) {
		if (modified) {
			setIcon(ICON_MODIFIED);
			setToolTipText(MSG_MODIFIED);
		} else {
			setIcon(ICON_NOT_MODIFIED);
			setToolTipText(MSG_NOT_MODIFIED);
		}
		repaint();
		JRootPane rootPane = getRootPane();
		if (rootPane != null) {
			rootPane.putClientProperty("Window.documentModified", modified ? Boolean.TRUE : Boolean.FALSE); //$NON-NLS-1$
		}
	}
}
