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

package com.trollworks.gcs.ui.editor.defaults;

import com.trollworks.gcs.model.skill.CMSkillDefault;
import com.trollworks.gcs.ui.editor.CSBandedPanel;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;

/** Displays and edits {@link CMSkillDefault} objects. */
public class CSDefaults extends CSBandedPanel implements ActionListener {
	/**
	 * Creates a new skill defaults editor.
	 * 
	 * @param defaults The initial defaults to display.
	 */
	public CSDefaults(List<CMSkillDefault> defaults) {
		super(Msgs.TITLE);
		setDefaults(defaults);
	}

	/** @param defaults The defaults to set. */
	public void setDefaults(List<CMSkillDefault> defaults) {
		removeAll();
		for (CMSkillDefault skillDefault : defaults) {
			add(new CSSkillDefault(new CMSkillDefault(skillDefault)));
		}
		if (getComponentCount() == 0) {
			add(new CSNoDefault());
		}
		revalidate();
		repaint();
	}

	@Override protected void addImpl(Component comp, Object constraints, int index) {
		super.addImpl(comp, constraints, index);
		if (comp instanceof CSSkillDefault) {
			((CSSkillDefault) comp).addActionListener(this);
		}
	}

	/** @return The current set of skill defaults. */
	public List<CMSkillDefault> getDefaults() {
		int count = getComponentCount();
		ArrayList<CMSkillDefault> list = new ArrayList<CMSkillDefault>(count);

		for (int i = 0; i < count; i++) {
			CMSkillDefault skillDefault = ((CSSkillDefault) getComponent(i)).getSkillDefault();

			if (skillDefault != null) {
				list.add(skillDefault);
			}
		}
		return list;
	}

	public void actionPerformed(ActionEvent event) {
		notifyActionListeners();
	}
}
