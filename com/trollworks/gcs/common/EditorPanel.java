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

package com.trollworks.gcs.common;

import com.trollworks.gcs.criteria.DoubleCriteria;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.criteria.NumericCriteria;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.criteria.WeightCriteria;
import com.trollworks.ttk.text.DoubleFormatter;
import com.trollworks.ttk.text.IntegerFormatter;
import com.trollworks.ttk.text.WeightFormatter;
import com.trollworks.ttk.units.WeightValue;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.ActionPanel;
import com.trollworks.ttk.widgets.EditorField;
import com.trollworks.ttk.widgets.IconButton;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.ArrayList;

import javax.swing.JComboBox;
import javax.swing.SwingConstants;
import javax.swing.border.EmptyBorder;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** A generic editor panel. */
public abstract class EditorPanel extends ActionPanel implements ActionListener, PropertyChangeListener {
	private static final int	GAP			= 5;
	private static String		MSG_EXACTLY;
	private static final String	COMPARISON	= "Comparison"; //$NON-NLS-1$

	static {
		LocalizedMessages.initialize(EditorPanel.class);
	}

	/** Creates a new {@link EditorPanel}. */
	protected EditorPanel() {
		this(0);
	}

	/**
	 * Creates a new {@link EditorPanel}.
	 * 
	 * @param indent The amount of indent to apply to the left side, if any.
	 */
	protected EditorPanel(int indent) {
		super();
		setOpaque(false);
		setBorder(new EmptyBorder(GAP, GAP + indent, GAP, GAP));
	}

	/**
	 * @param icon The image to use.
	 * @param command The command to use.
	 * @param tooltip The tooltip to use.
	 * @return The button that was added.
	 */
	protected IconButton addButton(BufferedImage icon, String command, String tooltip) {
		IconButton button = new IconButton(icon, tooltip);
		button.setActionCommand(command);
		button.addActionListener(this);
		add(button);
		return button;
	}

	/**
	 * Creates a new {@link JComboBox}.
	 * 
	 * @param command The command to issue on selection.
	 * @param items The items to add to the {@link JComboBox}.
	 * @param selection The item to select initially.
	 * @return The new {@link JComboBox}.
	 */
	protected JComboBox<Object> addComboBox(String command, Object[] items, Object selection) {
		JComboBox<Object> combo = new JComboBox<>(items);
		combo.setOpaque(false);
		combo.setSelectedItem(selection);
		combo.setActionCommand(command);
		combo.addActionListener(this);
		combo.setMaximumRowCount(items.length);
		UIUtilities.setOnlySize(combo, combo.getPreferredSize());
		add(combo);
		return combo;
	}

	/**
	 * @param compare The current string compare object.
	 * @param extra The extra text to add to the menu item.
	 * @return The {@link JComboBox} that allows a string comparison to be changed.
	 */
	protected JComboBox<Object> addStringCompareCombo(StringCriteria compare, String extra) {
		Object[] values;
		Object selection;
		if (extra == null) {
			values = StringCompareType.values();
			selection = compare.getType();
		} else {
			ArrayList<String> list = new ArrayList<>();
			selection = null;
			for (StringCompareType type : StringCompareType.values()) {
				String title = extra + type;
				list.add(title);
				if (type == compare.getType()) {
					selection = title;
				}
			}
			values = list.toArray();
		}
		JComboBox<Object> combo = addComboBox(COMPARISON, values, selection);
		combo.putClientProperty(StringCriteria.class, compare);
		return combo;
	}

	/**
	 * @param compare The current string compare object.
	 * @return The field that allows a string comparison to be changed.
	 */
	protected EditorField addStringCompareField(StringCriteria compare) {
		DefaultFormatter formatter = new DefaultFormatter();
		formatter.setOverwriteMode(false);
		EditorField field = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, compare.getQualifier(), null);
		field.putClientProperty(StringCriteria.class, compare);
		add(field);
		return field;
	}

	/**
	 * @param compare The current integer compare object.
	 * @param extra The extra text to add to the menu item.
	 * @return The {@link JComboBox} that allows a comparison to be changed.
	 */
	protected JComboBox<Object> addNumericCompareCombo(NumericCriteria compare, String extra) {
		Object[] values;
		Object selection;
		if (extra == null) {
			Object[] orig = NumericCompareType.values();
			values = new Object[orig.length];
			System.arraycopy(orig, 0, values, 0, orig.length);
			selection = compare.getType();
			values[NumericCompareType.IS.ordinal()] = MSG_EXACTLY;
			if (selection == NumericCompareType.IS) {
				selection = MSG_EXACTLY;
			}
		} else {
			ArrayList<String> list = new ArrayList<>();
			selection = null;
			for (NumericCompareType type : NumericCompareType.values()) {
				String title = extra + type.getDescription();
				list.add(title);
				if (type == compare.getType()) {
					selection = title;
				}
			}
			values = list.toArray();
		}
		JComboBox<Object> combo = addComboBox(COMPARISON, values, selection);
		combo.putClientProperty(NumericCriteria.class, compare);
		return combo;
	}

	/**
	 * @param compare The current compare object.
	 * @param min The minimum value to allow.
	 * @param max The maximum value to allow.
	 * @param forceSign Whether to force the sign to be visible.
	 * @return The {@link EditorField} that allows an integer comparison to be changed.
	 */
	protected EditorField addNumericCompareField(IntegerCriteria compare, int min, int max, boolean forceSign) {
		EditorField field = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(min, max, forceSign)), this, SwingConstants.LEFT, new Integer(compare.getQualifier()), new Integer(max), null);
		field.putClientProperty(IntegerCriteria.class, compare);
		UIUtilities.setOnlySize(field, field.getPreferredSize());
		add(field);
		return field;
	}

	/**
	 * @param compare The current compare object.
	 * @param min The minimum value to allow.
	 * @param max The maximum value to allow.
	 * @param forceSign Whether to force the sign to be visible.
	 * @return The {@link EditorField} that allows a double comparison to be changed.
	 */
	protected EditorField addNumericCompareField(DoubleCriteria compare, double min, double max, boolean forceSign) {
		EditorField field = new EditorField(new DefaultFormatterFactory(new DoubleFormatter(min, max, forceSign)), this, SwingConstants.LEFT, new Double(compare.getQualifier()), new Double(max), null);
		field.putClientProperty(DoubleCriteria.class, compare);
		UIUtilities.setOnlySize(field, field.getPreferredSize());
		add(field);
		return field;
	}

	/**
	 * @param compare The current compare object.
	 * @return The {@link EditorField} that allows a weight comparison to be changed.
	 */
	protected EditorField addWeightCompareField(WeightCriteria compare) {
		EditorField field = new EditorField(new DefaultFormatterFactory(new WeightFormatter()), this, SwingConstants.LEFT, compare.getQualifier(), null, null);
		field.putClientProperty(WeightCriteria.class, compare);
		add(field);
		return field;
	}

	@Override
	public void propertyChange(PropertyChangeEvent event) {
		if ("value".equals(event.getPropertyName())) { //$NON-NLS-1$
			EditorField field = (EditorField) event.getSource();
			StringCriteria criteria = (StringCriteria) field.getClientProperty(StringCriteria.class);
			if (criteria != null) {
				criteria.setQualifier((String) field.getValue());
				notifyActionListeners();
			} else {
				IntegerCriteria integerCriteria = (IntegerCriteria) field.getClientProperty(IntegerCriteria.class);
				if (integerCriteria != null) {
					integerCriteria.setQualifier(((Integer) field.getValue()).intValue());
					notifyActionListeners();
				} else {
					DoubleCriteria doubleCriteria = (DoubleCriteria) field.getClientProperty(DoubleCriteria.class);
					if (doubleCriteria != null) {
						doubleCriteria.setQualifier(((Double) field.getValue()).doubleValue());
						notifyActionListeners();
					} else {
						WeightCriteria weightCriteria = (WeightCriteria) field.getClientProperty(WeightCriteria.class);
						if (weightCriteria != null) {
							weightCriteria.setQualifier((WeightValue) field.getValue());
							notifyActionListeners();
						}
					}
				}
			}
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (COMPARISON.equals(command)) {
			JComboBox<Object> combo = (JComboBox<Object>) event.getSource();
			int selectedIndex = combo.getSelectedIndex();
			StringCriteria stringCriteria = (StringCriteria) combo.getClientProperty(StringCriteria.class);
			if (stringCriteria != null) {
				stringCriteria.setType(StringCompareType.values()[selectedIndex]);
				notifyActionListeners();
			} else {
				NumericCriteria numericCriteria = (NumericCriteria) combo.getClientProperty(NumericCriteria.class);
				if (numericCriteria != null) {
					numericCriteria.setType(NumericCompareType.values()[selectedIndex]);
					notifyActionListeners();
				}
			}
		}
	}
}
