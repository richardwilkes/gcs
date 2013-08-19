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

package com.trollworks.gcs.widgets.layout;

import java.awt.Component;
import java.awt.Dimension;

/** A convenience for retrieving minimum/maximum/preferred component sizes. */
public enum LayoutSize {
	/** The preferred size. */
	PREFERRED {
		@Override public Dimension get(Component component) {
			return sanitizeSize(component.getPreferredSize());
		}
	},
	/** The minimum size. */
	MINIMUM {
		@Override public Dimension get(Component component) {
			return sanitizeSize(component.getMinimumSize());
		}
	},
	/** The maximum size. */
	MAXIMUM {
		@Override public Dimension get(Component component) {
			return sanitizeSize(component.getMaximumSize());
		}
	};

	/** The maximum size to allow. */
	public static final int	MAXIMUM_SIZE	= Integer.MAX_VALUE / 512;

	/**
	 * @param component The {@link Component} to return the size for.
	 * @return The size desired by the {@link Component}.
	 */
	public abstract Dimension get(Component component);

	/**
	 * Ensures the size is within reasonable parameters.
	 * 
	 * @param size The size to check.
	 * @return The passed-in {@link Dimension} object, for convenience.
	 */
	public static Dimension sanitizeSize(Dimension size) {
		if (size.width < 0) {
			size.width = 0;
		} else if (size.width > MAXIMUM_SIZE) {
			size.width = MAXIMUM_SIZE;
		}
		if (size.height < 0) {
			size.height = 0;
		} else if (size.height > MAXIMUM_SIZE) {
			size.height = MAXIMUM_SIZE;
		}
		return size;
	}
}
