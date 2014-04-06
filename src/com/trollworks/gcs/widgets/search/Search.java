/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.widgets.search;

import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.utility.text.Numbers;

import java.awt.Point;
import java.awt.Rectangle;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.util.Collections;
import java.util.List;

import javax.swing.JLabel;
import javax.swing.JLayeredPane;
import javax.swing.JPanel;
import javax.swing.JRootPane;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** A standard search control. */
public class Search extends JPanel implements DocumentListener, KeyListener, FocusListener {
	private static String	MSG_HIT_TOOLTIP;
	private static String	MSG_SEARCH_FIELD_TOOLTIP;
	private SearchTarget	mTarget;
	private JLabel			mHits;
	private JTextField		mFilterField;
	private SearchDropDown	mFloater;
	private String			mFilter;

	/**
	 * Creates the search panel.
	 * 
	 * @param target The search target.
	 */
	public Search(SearchTarget target) {
		mTarget = target;

		mFilterField = new JTextField(20);
		mFilterField.getDocument().addDocumentListener(this);
		mFilterField.addKeyListener(this);
		mFilterField.addFocusListener(this);
		mFilterField.setToolTipText(MSG_SEARCH_FIELD_TOOLTIP);
		mFilterField.putClientProperty("JTextField.variant", "search"); //$NON-NLS-1$ //$NON-NLS-2$
		add(mFilterField);

		mHits = new JLabel();
		mHits.setToolTipText(MSG_HIT_TOOLTIP);
		adjustHits();
		add(mHits);

		FlexRow row = new FlexRow();
		row.add(mFilterField);
		row.add(mHits);
		row.apply(this);
	}

	@Override
	public boolean requestFocusInWindow() {
		return mFilterField.requestFocusInWindow();
	}

	private void searchSelect() {
		if (mFloater != null) {
			List<Object> selection = mFloater.getSelectedValues();
			if (!selection.isEmpty()) {
				mTarget.searchSelect(selection);
				return;
			}
		}
		mTarget.searchSelect(adjustHits());
	}

	/**
	 * Adjust the hits count.
	 * 
	 * @return The current hits.
	 */
	public List<Object> adjustHits() {
		List<Object> hits = mFilter != null ? mTarget.search(mFilter) : Collections.emptyList();
		mHits.setText(Numbers.format(hits.size()));
		if (mFloater != null) {
			mFloater.adjustToHits(hits);
		}
		return hits;
	}

	@Override
	public void changedUpdate(DocumentEvent event) {
		documentChanged();
	}

	@Override
	public void insertUpdate(DocumentEvent event) {
		documentChanged();
	}

	@Override
	public void removeUpdate(DocumentEvent event) {
		documentChanged();
	}

	private void documentChanged() {
		String filterText = mFilterField.getText();
		mFilter = filterText.length() > 0 ? filterText : null;
		adjustHits();
	}

	@Override
	public void keyPressed(KeyEvent event) {
		switch (event.getKeyCode()) {
			case KeyEvent.VK_UP:
			case KeyEvent.VK_DOWN:
				if (mFloater != null) {
					mFloater.handleKeyPressed(event);
				}
				break;
			case KeyEvent.VK_ENTER:
				searchSelect();
				break;
			default:
				break;
		}
	}

	@Override
	public void keyReleased(KeyEvent event) {
		// Not used.
	}

	@Override
	public void keyTyped(KeyEvent event) {
		// Not used.
	}

	@Override
	public void focusGained(FocusEvent event) {
		if (mFloater == null) {
			Point where = new Point(0, mFilterField.getHeight() + 1);
			mFloater = new SearchDropDown(mTarget.getSearchRenderer(), mFilterField, mTarget);
			JLayeredPane layeredPane = getRootPane().getLayeredPane();
			UIUtilities.convertPoint(where, mFilterField, layeredPane);
			layeredPane.add(mFloater, JLayeredPane.POPUP_LAYER);
			mFloater.repaint();
			adjustHits();
		}
	}

	@Override
	public void focusLost(FocusEvent event) {
		removeFloater();
	}

	/** @return The filter edit field. */
	public JTextField getFilterField() {
		return mFilterField;
	}

	@Override
	public void setEnabled(boolean enabled) {
		super.setEnabled(enabled);
		mHits.setEnabled(enabled);
		mFilterField.setEnabled(enabled);
		if (!enabled) {
			removeFloater();
		}
	}

	private void removeFloater() {
		if (mFloater != null) {
			Rectangle bounds = mFloater.getBounds();
			JRootPane rootPane = getRootPane();
			UIUtilities.convertRectangle(bounds, mFloater.getParent(), rootPane);
			mFloater.getParent().remove(mFloater);
			mFloater = null;
			rootPane.repaint(bounds);
		}
	}
}
