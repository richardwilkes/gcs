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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.skill;

import static com.trollworks.gcs.skill.OrLabel_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.utility.GraphicsUtilities;
import com.trollworks.ttk.utility.UIUtilities;

import java.awt.Component;
import java.awt.Graphics;

import javax.swing.JLabel;
import javax.swing.SwingConstants;

@Localized({
				@LS(key = "OR", msg = "or"),
})
/** A label that displays the "or" message, or nothing if it is the first one. */
public class OrLabel extends JLabel {
	private Component	mOwner;

	/**
	 * Creates a new {@link OrLabel}.
	 * 
	 * @param owner The owning component.
	 */
	public OrLabel(Component owner) {
		super(OR, SwingConstants.RIGHT);
		mOwner = owner;
		UIUtilities.setOnlySize(this, getPreferredSize());
	}

	@Override
	protected void paintComponent(Graphics gc) {
		setText(UIUtilities.getIndexOf(mOwner.getParent(), mOwner) != 0 ? OR : ""); //$NON-NLS-1$
		super.paintComponent(GraphicsUtilities.prepare(gc));
	}
}
