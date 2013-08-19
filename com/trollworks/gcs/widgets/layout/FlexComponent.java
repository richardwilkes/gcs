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

package com.trollworks.gcs.widgets.layout;

import java.awt.Component;
import java.awt.Dimension;
import java.awt.Rectangle;

/** A {@link Component} within a {@link FlexLayout}. */
public class FlexComponent extends FlexCell {
	private Component	mComponent;

	/**
	 * Creates a new {@link FlexComponent}.
	 * 
	 * @param component The {@link Component} to wrap.
	 */
	public FlexComponent(Component component) {
		mComponent = component;
	}

	/**
	 * Creates a new {@link FlexComponent}.
	 * 
	 * @param component The {@link Component} to wrap.
	 * @param horizontalAlignment The horizontal {@link Alignment} to use. Pass in <code>null</code>
	 *            to use the default.
	 * @param verticalAlignment The vertical {@link Alignment} to use. Pass in <code>null</code>
	 *            to use the default.
	 */
	public FlexComponent(Component component, Alignment horizontalAlignment, Alignment verticalAlignment) {
		mComponent = component;
		if (horizontalAlignment != null) {
			setHorizontalAlignment(horizontalAlignment);
		}
		if (verticalAlignment != null) {
			setVerticalAlignment(verticalAlignment);
		}
	}

	@Override protected Dimension getSizeSelf(LayoutSize type) {
		return type.get(mComponent);
	}

	@Override protected void layoutSelf(Rectangle bounds) {
		Rectangle compBounds = new Rectangle(bounds);
		Dimension size = LayoutSize.MINIMUM.get(mComponent);
		if (compBounds.width < size.width) {
			compBounds.width = size.width;
		}
		if (compBounds.height < size.height) {
			compBounds.height = size.height;
		}
		size = LayoutSize.MAXIMUM.get(mComponent);
		if (compBounds.width > size.width) {
			compBounds.width = size.width;
		}
		if (compBounds.height > size.height) {
			compBounds.height = size.height;
		}
		mComponent.setBounds(Alignment.position(bounds, compBounds, getHorizontalAlignment(), getVerticalAlignment()));
	}

	@Override public String toString() {
		return mComponent.getClass().getSimpleName();
	}
}
