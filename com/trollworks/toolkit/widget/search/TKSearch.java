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

package com.trollworks.toolkit.widget.search;

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.text.TKDocument;
import com.trollworks.toolkit.text.TKDocumentListener;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKKeyEventFilter;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.window.TKBaseWindow;

import java.awt.AWTEvent;
import java.awt.Container;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.util.ArrayList;
import java.util.Collection;

/** A standard search control. */
public class TKSearch extends TKPanel implements ActionListener, TKDocumentListener, FocusListener, TKKeyEventFilter {
	/** The commands the tool panel provides. */
	public static final String	CMD_SEARCH	= "search"; //$NON-NLS-1$
	private TKSearchTarget		mTarget;
	private TKLabel				mHits;
	private TKTextField			mFilterField;
	private TKSearchDropDown	mFloater;
	private String				mFilter;

	/**
	 * Creates the search panel.
	 * 
	 * @param target The search target.
	 */
	public TKSearch(TKSearchTarget target) {
		this(target, TKFont.CONTROL_FONT_KEY, TKFont.TEXT_FONT_KEY);
	}

	/**
	 * Creates the search panel.
	 * 
	 * @param target The search target.
	 * @param fieldFont The dynamic font to use for the search field.
	 * @param matchesFont The dynamic font to use for the matches found label.
	 */
	public TKSearch(TKSearchTarget target, String fieldFont, String matchesFont) {
		super(new TKColumnLayout(3));

		mTarget = target;

		TKButton button = new TKButton(TKImage.getQuickSearchIcon());
		button.setBorder(null);
		button.setPressedBorder(null);
		button.setOpaque(false);
		button.setToolTipText(Msgs.SEARCH_BUTTON_TOOLTIP);
		button.setActionCommand(CMD_SEARCH);
		button.addActionListener(this);
		add(button);

		mFilterField = new TKTextField(75);
		mFilterField.setFontKey(fieldFont);
		mFilterField.setKeyEventFilter(this);
		mFilterField.addDocumentListener(this);
		mFilterField.addFocusListener(this);
		mFilterField.setToolTipText(Msgs.SEARCH_FIELD_TOOLTIP);
		add(mFilterField);

		mHits = new TKLabel(matchesFont, TKAlignment.LEFT);
		mHits.setToolTipText(Msgs.HIT_TOOLTIP);
		mHits.enableAWTEvents(AWTEvent.MOUSE_MOTION_EVENT_MASK); // Just to allow tooltips...
		adjustHits();
		add(mHits);
	}

	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (CMD_SEARCH.equals(command)) {
			mTarget.searchSelect(adjustHits());
		}
	}

	/**
	 * Adjust the hits count.
	 * 
	 * @return The current collection of hits.
	 */
	public Collection<Object> adjustHits() {
		Collection<Object> hits = mFilter != null ? mTarget.search(mFilter) : new ArrayList<Object>();

		mHits.setText(TKNumberUtils.format(hits.size()));
		if (mFloater != null) {
			mFloater.adjustToHits(hits);
		}
		return hits;
	}

	public void documentChanged(TKDocument document) {
		String filterText = mFilterField.getText();

		if (filterText.length() > 0) {
			mFilter = filterText;
		} else {
			mFilter = null;
		}
		adjustHits();
	}

	public boolean filterKeyEvent(TKPanel owner, KeyEvent event, boolean isReal) {
		if (isReal && !TKKeystroke.isCommandKeyDown(event)) {
			boolean ignore = false;

			if (mFloater != null) {
				if (mFloater.doKey(event)) {
					ignore = true;
				}
			}

			int id = event.getID();

			if (id == KeyEvent.KEY_TYPED) {
				char ch = event.getKeyChar();

				if (ch == '\r' || ch == '\n') {
					if (ignore) {
						return false;
					}
					mTarget.searchSelect(adjustHits());
				}
			}
			if (ignore) {
				return true;
			}
		}
		return false;
	}

	public void focusGained(FocusEvent event) {
		TKBaseWindow bWindow = getBaseWindow();

		if (bWindow != null) {
			Container window = (Container) bWindow;

			if (mFloater == null) {
				Point where = new Point(0, mFilterField.getHeight() + 1);

				mFloater = new TKSearchDropDown(mTarget.getSearchRenderer(), mFilterField, mTarget);
				convertPoint(where, mFilterField, window);
				mFilterField.addReshapeListener(mFloater);
				window.add(mFloater, 0);
				adjustHits();
			}
		}
	}

	public void focusLost(FocusEvent event) {
		TKBaseWindow bWindow = getBaseWindow();

		if (bWindow != null) {
			Container window = (Container) bWindow;

			if (mFloater != null) {
				Rectangle bounds = mFloater.getBounds();

				mFilterField.removeReshapeListener(mFloater);
				window.repaint(bounds.x, bounds.y, bounds.width, bounds.height);
				window.remove(mFloater);
				mFloater = null;
			}
		}
	}

	/** @return The filter edit field. */
	public TKTextField getFilterField() {
		return mFilterField;
	}

	@Override public void setEnabled(boolean enabled) {
		super.setEnabled(enabled);
		mHits.setEnabled(enabled);
		mFilterField.setEnabled(enabled);
		if (mFloater != null && !enabled) {
			TKPanel floaterParent = (TKPanel) mFloater.getParent();

			mFilterField.removeReshapeListener(mFloater);
			floaterParent.repaint(mFloater.getBounds());
			floaterParent.remove(mFloater);
			mFloater = null;
		}
	}
}
