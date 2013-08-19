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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.utility.io.LocalizedMessages;

/** The possible states for a piece of equipment. */
public enum EquipmentState {
	/**
	 * The state for a piece of equipment that is being carried and should also have any of its
	 * {@link Feature}s applied. For example, a magic ring that is being worn on a finger.
	 */
	EQUIPPED {
		@Override public String toString() {
			return MSG_EQUIPPED;
		}

		@Override public String toShortString() {
			return "E"; //$NON-NLS-1$
		}
	},
	/**
	 * The state for a piece of equipment that is being carried, but should not have any of its
	 * {@link Feature}s applied. For example, a magic ring that is being stored in a pouch.
	 */
	CARRIED {
		@Override public String toString() {
			return MSG_CARRIED;
		}

		@Override public String toShortString() {
			return "C"; //$NON-NLS-1$
		}
	},
	/** The state of a piece of equipment that is not being carried. */
	NOT_CARRIED {
		@Override public String toString() {
			return MSG_NOT_CARRIED;
		}

		@Override public String toShortString() {
			return "-"; //$NON-NLS-1$
		}
	};

	static String	MSG_EQUIPPED;
	static String	MSG_CARRIED;
	static String	MSG_NOT_CARRIED;

	static {
		LocalizedMessages.initialize(EquipmentState.class);
	}

	/** @return The short form of its description, typically a single character. */
	public abstract String toShortString();
}
