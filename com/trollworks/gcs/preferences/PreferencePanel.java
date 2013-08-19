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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.layout.Alignment;
import com.trollworks.gcs.widgets.layout.FlexComponent;
import com.trollworks.gcs.widgets.layout.FlexContainer;
import com.trollworks.gcs.widgets.layout.LayoutSize;

import java.awt.Dimension;
import java.awt.Insets;
import java.awt.event.ItemListener;

import javax.swing.ImageIcon;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JSeparator;
import javax.swing.SwingConstants;
import javax.swing.border.EmptyBorder;

/** The abstract base class for all preference panels. */
public abstract class PreferencePanel extends JPanel {
	private String				mTitle;
	private PreferencesWindow	mOwner;

	/**
	 * Creates a new preference panel.
	 * 
	 * @param title The title for this panel.
	 * @param owner The owning {@link PreferencesWindow}.
	 */
	public PreferencePanel(String title, PreferencesWindow owner) {
		super();
		setBorder(new EmptyBorder(5, 5, 5, 5));
		setOpaque(false);
		mTitle = title;
		mOwner = owner;
	}

	/** @return The owner. */
	public PreferencesWindow getOwner() {
		return mOwner;
	}

	/** Resets this panel back to its defaults. */
	public abstract void reset();

	/** @return Whether the panel is currently set to defaults or not. */
	public abstract boolean isSetToDefaults();

	/** Call to adjust the reset button for any changes that have been made. */
	protected void adjustResetButton() {
		mOwner.adjustResetButton();
	}

	@Override public String toString() {
		return mTitle;
	}

	/**
	 * Creates a right-aligned {@link JLabel} suitable for use within the preference panel.
	 * 
	 * @param title The title to use.
	 * @param tooltip The tooltip to use.
	 * @return The newly created {@link JLabel}.
	 */
	protected JLabel createLabel(String title, String tooltip) {
		return createLabel(title, tooltip, null, SwingConstants.RIGHT);
	}

	/**
	 * Creates a {@link JLabel} suitable for use within the preference panel.
	 * 
	 * @param title The title to use.
	 * @param tooltip The tooltip to use.
	 * @param alignment The alignment to use.
	 * @return The newly created {@link JLabel}.
	 */
	protected JLabel createLabel(String title, String tooltip, int alignment) {
		return createLabel(title, tooltip, null, alignment);
	}

	/**
	 * Creates a right-aligned {@link JLabel} suitable for use within the preference panel.
	 * 
	 * @param title The title to use.
	 * @param tooltip The tooltip to use.
	 * @param icon The {@link ImageIcon} to use.
	 * @return The newly created {@link JLabel}.
	 */
	protected JLabel createLabel(String title, String tooltip, ImageIcon icon) {
		return createLabel(title, tooltip, icon, SwingConstants.RIGHT);
	}

	/**
	 * Creates a {@link JLabel} suitable for use within the preference panel.
	 * 
	 * @param title The title to use.
	 * @param tooltip The tooltip to use.
	 * @param icon The {@link ImageIcon} to use.
	 * @param alignment The alignment to use.
	 * @return The newly created {@link JLabel}.
	 */
	protected JLabel createLabel(String title, String tooltip, ImageIcon icon, int alignment) {
		JLabel label = new JLabel(title, alignment);
		label.setOpaque(false);
		label.setToolTipText(tooltip);
		label.setIcon(icon);
		UIUtilities.setOnlySize(label, label.getPreferredSize());
		add(label);
		return label;
	}

	/** @param flexContainer The {@link FlexContainer} to add a separator to. */
	protected void addSeparator(FlexContainer flexContainer) {
		JSeparator sep = new JSeparator();
		sep.setOpaque(false);
		sep.setMaximumSize(new Dimension(LayoutSize.MAXIMUM_SIZE, 1));
		add(sep);
		FlexComponent comp = new FlexComponent(sep, Alignment.CENTER, Alignment.CENTER);
		comp.setInsets(new Insets(5, 0, 5, 0));
		flexContainer.add(comp);
	}

	/**
	 * Creates a {@link JCheckBox} suitable for use within the preference panel.
	 * 
	 * @param title The title to use.
	 * @param tooltip The tooltip to use.
	 * @param checked Whether the initial state should be checked.
	 * @return The newly created {@link JCheckBox}.
	 */
	protected JCheckBox createCheckBox(String title, String tooltip, boolean checked) {
		JCheckBox checkbox = new JCheckBox(title, checked);
		checkbox.setOpaque(false);
		checkbox.setToolTipText(tooltip);
		if (this instanceof ItemListener) {
			checkbox.addItemListener((ItemListener) this);
		}
		add(checkbox);
		return checkbox;
	}

	/**
	 * Creates a {@link JComboBox} suitable for use within the preference panel.
	 * 
	 * @param tooltip The tooltip to use.
	 * @return The newly created {@link JComboBox}.
	 */
	protected JComboBox createCombo(String tooltip) {
		JComboBox combo = new JComboBox();
		combo.setOpaque(false);
		combo.setToolTipText(tooltip);
		add(combo);
		return combo;
	}
}
