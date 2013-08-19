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

package com.trollworks.toolkit.qa;

import com.trollworks.toolkit.undo.TKUndo;
import com.trollworks.toolkit.undo.TKUndoManager;
import com.trollworks.toolkit.undo.TKUndoManagerMonitor;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKDefaultItemRenderer;
import com.trollworks.toolkit.widget.TKItemList;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;

/** Displays the current undo state for a window. */
public class TKUndoMonitorWindow extends TKWindow implements TKUndoManagerMonitor {
	private TKItemList<TKUndo>	mList;
	/** The index of the next add. */
	protected int				mIndexOfNextAdd;

	/**
	 * Creates a new undo monitor for the specified window.
	 * 
	 * @param window The window to monitor.
	 */
	public TKUndoMonitorWindow(TKWindow window) {
		super(MessageFormat.format(Msgs.TITLE_FORMAT, window.getTitle()), null);

		TKUndoManager undoMgr = window.getUndoManager();
		mList = new TKItemList<TKUndo>(undoMgr.getCurrentEdits());
		mIndexOfNextAdd = undoMgr.getIndexOfNextAdd();
		TKScrollPanel scroller = new TKScrollPanel(mList);

		mList.setItemRenderer(new ItemRenderer());
		getContent().add(scroller);
		scroller.setMinimumSize(new Dimension(200, 300));
		undoMgr.addMonitor(this);
	}

	public void undoManagerStackChanged(TKUndoManager mgr) {
		mList.removeAllItems();
		mIndexOfNextAdd = mgr.getIndexOfNextAdd();
		mList.addItems(mgr.getCurrentEdits());
	}

	private class ItemRenderer extends TKDefaultItemRenderer {
		/** Creates a new item renderer. */
		public ItemRenderer() {
			super();
			setOpaque(true);
		}

		@Override public String getStringForItem(Object item, int index) {
			TKUndo edit = (TKUndo) item;

			return edit.getName(mIndexOfNextAdd > index);
		}

		@Override public Color getBackgroundForItem(Object item, int index, boolean selected, boolean active) {
			if (selected) {
				return super.getBackgroundForItem(item, index, selected, active);
			}
			return index % 2 == 0 ? TKColor.PRIMARY_BANDING : TKColor.SECONDARY_BANDING;
		}
	}
}
