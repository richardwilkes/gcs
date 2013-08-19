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

package com.trollworks.toolkit.print;

import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKTitledBorder;
import com.trollworks.toolkit.widget.button.TKRadioButton;
import com.trollworks.toolkit.widget.button.TKRadioButtonGroup;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

import javax.print.PrintService;
import javax.print.attribute.PrintRequestAttributeSet;
import javax.print.attribute.standard.Copies;
import javax.print.attribute.standard.PageRanges;

/** Provides the basic print panel. */
public class TKPrintPanel extends TKPageSetupPanel implements ActionListener {
	private TKTextField		mCopies;
	private TKRadioButton	mPageRangeAll;
	private TKRadioButton	mPageRangeSome;
	private TKTextField		mPageRangeStart;
	private TKTextField		mPageRangeEnd;

	/**
	 * Creates a new print panel.
	 * 
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 */
	public TKPrintPanel(PrintService service, PrintRequestAttributeSet set) {
		super(service, set);
	}

	@Override protected void rebuildSelf(PrintService service, PrintRequestAttributeSet set) {
		TKPanel content = getContentPanel();
		TKPanel panel = new TKPanel(new TKColumnLayout(2));

		panel.setBorder(new TKCompoundBorder(new TKTitledBorder(Msgs.PRINT_JOB_SETTINGS, TKFont.lookup(TKFont.TEXT_FONT_KEY)), new TKEmptyBorder(5)));
		addCopiesField(service, set, panel);
		addPageRanges(service, set, panel);
		content.add(panel);

		super.rebuildSelf(service, set);
	}

	private void addCopiesField(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		if (service.isAttributeCategorySupported(Copies.class)) {
			TKPanel wrapper = new TKPanel(new TKColumnLayout(2, 0, 0));

			mCopies = new TKTextField("999", 0, TKAlignment.RIGHT); //$NON-NLS-1$
			mCopies.setKeyEventFilter(new TKNumberFilter(false, false, false, 3));
			mCopies.setOnlySize(mCopies.getPreferredSize());
			mCopies.setText(Integer.toString(TKPrintUtilities.getCopies(service, set)));
			parent.add(new TKLabel(Msgs.COPIES, TKAlignment.RIGHT));
			wrapper.add(mCopies);
			wrapper.add(new TKPanel());
			parent.add(wrapper);
		}
	}

	private void addPageRanges(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		if (service.isAttributeCategorySupported(PageRanges.class)) {
			TKPanel wrapper = new TKPanel(new TKColumnLayout(6, TKColumnLayout.DEFAULT_H_GAP_SIZE, 0));
			TKRadioButtonGroup group = new TKRadioButtonGroup();
			PageRanges pageRanges = (PageRanges) set.get(PageRanges.class);
			int start = 1;
			int end = 9999;

			if (pageRanges != null) {
				int[][] ranges = pageRanges.getMembers();

				if (ranges.length > 0 && ranges[0].length > 1) {
					start = ranges[0][0];
					end = ranges[0][1];
				} else {
					pageRanges = null;
				}
			}

			mPageRangeAll = new TKRadioButton(Msgs.ALL);
			mPageRangeSome = new TKRadioButton(Msgs.PAGES);
			mPageRangeStart = createPageRangeField(start);
			mPageRangeEnd = createPageRangeField(end);
			group.add(mPageRangeAll);
			group.add(mPageRangeSome);
			group.setSelection(pageRanges != null ? mPageRangeSome : mPageRangeAll);
			adjustPageRanges();
			mPageRangeAll.addActionListener(this);
			mPageRangeSome.addActionListener(this);
			wrapper.add(mPageRangeAll);
			wrapper.add(mPageRangeSome);
			wrapper.add(mPageRangeStart);
			wrapper.add(new TKLabel(Msgs.TO, TKAlignment.CENTER));
			wrapper.add(mPageRangeEnd);
			wrapper.add(new TKPanel());
			parent.add(new TKLabel(Msgs.PAGE_RANGE, TKAlignment.RIGHT));
			parent.add(wrapper);
		}
	}

	private void adjustPageRanges() {
		boolean enabled = !mPageRangeAll.isSelected();

		mPageRangeStart.setEnabled(enabled);
		mPageRangeEnd.setEnabled(enabled);
	}

	private TKTextField createPageRangeField(int value) {
		TKTextField field = new TKTextField("9999", 0, TKAlignment.RIGHT); //$NON-NLS-1$

		field.setKeyEventFilter(new TKNumberFilter(false, false, false, 4));
		field.setOnlySize(field.getPreferredSize());
		field.setText(Integer.toString(value));
		return field;
	}

	@Override public PrintService accept(PrintRequestAttributeSet set) {
		PrintService service = super.accept(set);

		if (mCopies != null) {
			TKPrintUtilities.setCopies(set, TKNumberUtils.getInteger(mCopies.getText(), 1));
		}
		if (mPageRangeAll != null) {
			if (mPageRangeAll.isSelected()) {
				TKPrintUtilities.setPageRanges(set, null);
			} else {
				int start = Math.max(TKNumberUtils.getInteger(mPageRangeStart.getText(), 1), 1);
				int end = Math.max(TKNumberUtils.getInteger(mPageRangeEnd.getText(), 1), 1);

				if (start > end) {
					int tmp = start;

					start = end;
					end = tmp;
				}
				TKPrintUtilities.setPageRanges(set, new PageRanges(start, end));
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
