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

package com.trollworks.gcs.utility.io.print;

import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.text.IntegerFormatter;
import com.trollworks.gcs.widgets.EditorField;
import com.trollworks.gcs.widgets.LinkedLabel;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.layout.Alignment;
import com.trollworks.gcs.widgets.layout.FlexComponent;
import com.trollworks.gcs.widgets.layout.FlexGrid;
import com.trollworks.gcs.widgets.layout.FlexRow;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

import javax.print.PrintService;
import javax.print.attribute.PrintRequestAttributeSet;
import javax.print.attribute.standard.Copies;
import javax.print.attribute.standard.PageRanges;
import javax.swing.ButtonGroup;
import javax.swing.JLabel;
import javax.swing.JRadioButton;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

/** Provides the basic print panel. */
public class PrintPanel extends PageSetupPanel implements ActionListener {
	private static String	MSG_COPIES;
	private static String	MSG_PAGE_RANGE;
	private static String	MSG_ALL;
	private static String	MSG_PAGES;
	private static String	MSG_TO;
	private EditorField		mCopies;
	private JRadioButton	mPageRangeAll;
	private JRadioButton	mPageRangeSome;
	private EditorField		mPageRangeStart;
	private EditorField		mPageRangeEnd;

	static {
		LocalizedMessages.initialize(PrintPanel.class);
	}

	/**
	 * Creates a new print panel.
	 * 
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 */
	public PrintPanel(PrintService service, PrintRequestAttributeSet set) {
		super(service, set);
	}

	@Override protected void rebuildSelf(PrintRequestAttributeSet set, FlexGrid grid, int row) {
		row = createCopiesField(set, grid, row);
		row = createPageRangeFields(set, grid, row);
		super.rebuildSelf(set, grid, row);
	}

	private int createCopiesField(PrintRequestAttributeSet set, FlexGrid grid, int row) {
		PrintService service = getService();
		if (service.isAttributeCategorySupported(Copies.class)) {
			mCopies = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(1, 999, false)), null, SwingConstants.RIGHT, new Integer(PrintUtilities.getCopies(service, set)), new Integer(999), null);
			UIUtilities.setOnlySize(mCopies, mCopies.getPreferredSize());
			add(mCopies);
			LinkedLabel label = new LinkedLabel(MSG_COPIES, mCopies);
			add(label);
			grid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, Alignment.CENTER), row, 0);
			grid.add(mCopies, row, 1);
			return row + 1;
		}
		mCopies = null;
		return row;
	}

	private int createPageRangeFields(PrintRequestAttributeSet set, FlexGrid grid, int row) {
		PrintService service = getService();
		if (service.isAttributeCategorySupported(PageRanges.class)) {
			ButtonGroup group = new ButtonGroup();
			int start = 1;
			int end = 9999;
			PageRanges pageRanges = (PageRanges) set.get(PageRanges.class);
			if (pageRanges != null) {
				int[][] ranges = pageRanges.getMembers();
				if (ranges.length > 0 && ranges[0].length > 1) {
					start = ranges[0][0];
					end = ranges[0][1];
				} else {
					pageRanges = null;
				}
			}
			JLabel label = new JLabel(MSG_PAGE_RANGE, SwingConstants.CENTER);
			add(label);
			grid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, Alignment.CENTER), row, 0);
			mPageRangeAll = new JRadioButton(MSG_ALL, pageRanges == null);
			add(mPageRangeAll);
			mPageRangeSome = new JRadioButton(MSG_PAGES, pageRanges != null);
			add(mPageRangeSome);
			mPageRangeStart = createPageRangeField(start);
			mPageRangeEnd = createPageRangeField(end);
			group.add(mPageRangeAll);
			group.add(mPageRangeSome);
			adjustPageRanges();
			mPageRangeAll.addActionListener(this);
			mPageRangeSome.addActionListener(this);
			label = new JLabel(MSG_TO, SwingConstants.CENTER);
			add(label);
			FlexRow fRow = new FlexRow();
			fRow.add(mPageRangeAll);
			fRow.add(mPageRangeSome);
			fRow.add(mPageRangeStart);
			fRow.add(label);
			fRow.add(mPageRangeEnd);
			grid.add(fRow, row, 1);
			return row + 1;
		}
		mPageRangeAll = null;
		mPageRangeSome = null;
		mPageRangeStart = null;
		mPageRangeEnd = null;
		return row;
	}

	private EditorField createPageRangeField(int value) {
		EditorField field = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(1, 9999, false)), null, SwingConstants.RIGHT, new Integer(value), new Integer(9999), null);
		UIUtilities.setOnlySize(field, field.getPreferredSize());
		add(field);
		return field;
	}

	private void adjustPageRanges() {
		boolean enabled = !mPageRangeAll.isSelected();
		mPageRangeStart.setEnabled(enabled);
		mPageRangeEnd.setEnabled(enabled);
	}

	@Override public PrintService accept(PrintRequestAttributeSet set) {
		PrintService service = super.accept(set);
		if (mCopies != null) {
			PrintUtilities.setCopies(set, ((Integer) mCopies.getValue()).intValue());
		}
		if (mPageRangeAll != null) {
			if (mPageRangeAll.isSelected()) {
				PrintUtilities.setPageRanges(set, null);
			} else {
				int start = ((Integer) mPageRangeStart.getValue()).intValue();
				int end = ((Integer) mPageRangeEnd.getValue()).intValue();
				if (start > end) {
					int tmp = start;
					start = end;
					end = tmp;
				}
				PrintUtilities.setPageRanges(set, new PageRanges(start, end));
			}
		}
		return service;
	}

	@Override public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		if (src == mPageRangeAll || src == mPageRangeSome) {
			adjustPageRanges();
		} else {
			super.actionPerformed(event);
		}
	}
}
