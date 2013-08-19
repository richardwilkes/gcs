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

package com.trollworks.gcs.ui.equipment;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	EQUIPMENT;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	CARRIED_EQUIPMENT;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	EQUIPMENT_TOOLTIP;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	EQUIPPED;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	TECH_LEVEL;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	TECH_LEVEL_TOOLTIP;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	LEGALITY_CLASS;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	LEGALITY_CLASS_TOOLTIP;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	QUANTITY;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	QUANTITY_TOOLTIP;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	VALUE;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	WEIGHT;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	WEIGHT_TOOLTIP;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	EXT_VALUE;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	EXT_WEIGHT;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	EXT_WEIGHT_TOOLTIP;
	/** Used by {@link CSEquipmentColumnID}. */
	public static String	REFERENCE;

	/** Used by {@link CSEquipmentColumnID} and {@link CSEquipmentEditor}. */
	public static String	REFERENCE_TOOLTIP;
	/** Used by {@link CSEquipmentColumnID} and {@link CSEquipmentEditor}. */
	public static String	EQUIPPED_TOOLTIP;
	/** Used by {@link CSEquipmentColumnID} and {@link CSEquipmentEditor}. */
	public static String	VALUE_TOOLTIP;
	/** Used by {@link CSEquipmentColumnID} and {@link CSEquipmentEditor}. */
	public static String	EXT_VALUE_TOOLTIP;

	/** Used by {@link CSEquipmentEditor}. */
	public static String	NAME;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	NAME_TOOLTIP;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	NAME_CANNOT_BE_EMPTY;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_TECH_LEVEL;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_TECH_LEVEL_TOOLTIP;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_LEGALITY_CLASS;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_LEGALITY_CLASS_TOOLTIP;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_QUANTITY;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_QUANTITY_TOOLTIP;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_EQUIPPED;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_VALUE;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_EXTENDED_VALUE;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_WEIGHT;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_WEIGHT_TOOLTIP;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_EXTENDED_WEIGHT;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_EXTENDED_WEIGHT_TOOLTIP;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	NOTES;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	NOTES_TOOLTIP;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	EDITOR_REFERENCE;
	/** Used by {@link CSEquipmentEditor}. */
	public static String	POUNDS;

	/** Used by {@link CSEquipmentListWindow}. */
	public static String	UNTITLED;

	/** Used by {@link CSEquipmentOutline}. */
	public static String	INCREMENT;
	/** Used by {@link CSEquipmentOutline}. */
	public static String	DECREMENT;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
