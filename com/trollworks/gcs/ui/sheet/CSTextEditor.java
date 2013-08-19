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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.scroll.TKScrollable;
import com.trollworks.toolkit.window.TKOptionDialog;
import com.trollworks.toolkit.window.TKDialog;

import java.awt.Dimension;
import java.awt.Rectangle;
import java.awt.event.FocusEvent;

/** Provides simplistic text editing. */
public class CSTextEditor extends TKPanel {
	private Editor	mEditor;

	/**
	 * Puts up a modal text editor.
	 * 
	 * @param title The title for the dialog.
	 * @param text The text to edit.
	 * @return The new text, or <code>null</code> if changes were cancelled.
	 */
	public static String edit(String title, String text) {
		CSTextEditor editor = new CSTextEditor(text);
		TKOptionDialog dialog = new TKOptionDialog(title, TKOptionDialog.TYPE_OK_CANCEL);

		dialog.setResizable(true);
		if (dialog.doModal(null, editor, null) == TKDialog.OK) {
			return editor.getText();
		}
		return null;
	}

	private CSTextEditor(String text) {
		super(new TKCompassLayout());
		mEditor = new Editor(text);
		mEditor.setMinimumSize(new Dimension(400, 300));
		TKScrollPanel scroller = new TKScrollPanel(mEditor);
		add(scroller);
	}

	private String getText() {
		return mEditor.getText();
	}

	private class Editor extends TKTextField implements TKScrollable {
		/** @param text The text to edit. */
		public Editor(String text) {
			super();
			setSingleLineOnly(false);
			setText(text);
			setSelection(text.length(), text.length());
			setBorder(new TKCompoundBorder(new TKLineBorder(TKLineBorder.TOP_EDGE | TKLineBorder.LEFT_EDGE), new TKEmptyBorder(4, 4, 2, 4)));
		}

		@Override public void focusGained(FocusEvent event) {
			getDocument().setActive(true);
			scrollIntoView();
			repaintAndResetCursor();
		}

		public int getBlockScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
			if (vertical) {
				return upLeftDirection ? -visibleBounds.height : visibleBounds.height;
			}
			return upLeftDirection ? -visibleBounds.width : visibleBounds.width;
		}

		public Dimension getPreferredViewportSize() {
			return getPreferredSize();
		}

		public int getUnitScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
			Dimension size = TKTextDrawing.getPreferredSize(getFont(), null, "M"); //$NON-NLS-1$

			if (vertical) {
				return upLeftDirection ? -size.height : size.height;
			}
			return upLeftDirection ? -size.width : size.width;
		}

		public boolean shouldTrackViewportHeight() {
			return TKScrollPanel.shouldTrackViewportHeight(this);
		}

		public boolean shouldTrackViewportWidth() {
			return TKScrollPanel.shouldTrackViewportWidth(this);
		}
	}
}
