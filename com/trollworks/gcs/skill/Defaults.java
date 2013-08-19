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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.skill;

import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.BandedPanel;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;

/** Displays and edits {@link SkillDefault} objects. */
public class Defaults extends BandedPanel implements ActionListener {
	private static String	MSG_TITLE;

	static {
		LocalizedMessages.initialize(Defaults.class);
	}

	/**
	 * Creates a new skill defaults editor.
	 * 
	 * @param defaults The initial defaults to display.
	 */
	public Defaults(List<SkillDefault> defaults) {
		super(MSG_TITLE);
		setDefaults(defaults);
	}

	/** @param defaults The defaults to set. */
	public void setDefaults(List<SkillDefault> defaults) {
		removeAll();
		for (SkillDefault skillDefault : defaults) {
			add(new SkillDefaultEditor(new SkillDefault(skillDefault)));
		}
		if (getComponentCount() == 0) {
			add(new SkillDefaultEditor());
		}
		revalidate();
		repaint();
	}

	@Override
	protected void addImpl(Component comp, Object constraints, int index) {
		super.addImpl(comp, constraints, index);
		if (comp instanceof SkillDefaultEditor) {
			((SkillDefaultEditor) comp).addActionListener(this);
		}
	}

	/** @return The current set of skill defaults. */
	public List<SkillDefault> getDefaults() {
		int count = getComponentCount();
		ArrayList<SkillDefault> list = new ArrayList<>(count);
		for (int i = 0; i < count; i++) {
			SkillDefault skillDefault = ((SkillDefaultEditor) getComponent(i)).getSkillDefault();
			if (skillDefault != null) {
				list.add(skillDefault);
			}
		}
		return list;
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		notifyActionListeners();
	}
}
