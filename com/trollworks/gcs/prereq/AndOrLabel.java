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

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.GraphicsUtilities;
import com.trollworks.gcs.widgets.UIUtilities;

import java.awt.Graphics;

import javax.swing.JLabel;
import javax.swing.SwingConstants;

/** A label that displays the "and" or the "or" message, or nothing if it is the first one. */
public class AndOrLabel extends JLabel {
	private static String	MSG_AND;
	private static String	MSG_OR;
	private Prereq			mOwner;

	static {
		LocalizedMessages.initialize(AndOrLabel.class);
	}

	/**
	 * Creates a new {@link AndOrLabel}.
	 * 
	 * @param owner The owning {@link Prereq}.
	 */
	public AndOrLabel(Prereq owner) {
		super(MSG_AND, SwingConstants.RIGHT);
		mOwner = owner;
		UIUtilities.setOnlySize(this, getPreferredSize());
	}

	@Override protected void paintComponent(Graphics gc) {
		PrereqList parent = mOwner.getParent();
		if (parent != null && parent.getChildren().get(0) != mOwner) {
			setText(parent.requiresAll() ? MSG_AND : MSG_OR);
		} else {
			setText(""); //$NON-NLS-1$
		}
		super.paintComponent(GraphicsUtilities.prepare(gc));
	}
}
