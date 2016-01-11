/*
 * Copyright (c) 1998-2016 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.pdfview;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.Log;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.menu.file.CloseHandler;
import com.trollworks.toolkit.ui.widget.EditorField;
import com.trollworks.toolkit.ui.widget.IconButton;
import com.trollworks.toolkit.ui.widget.KeyStrokeDisplay;
import com.trollworks.toolkit.ui.widget.Toolbar;
import com.trollworks.toolkit.ui.widget.dock.DockContainer;
import com.trollworks.toolkit.ui.widget.dock.Dockable;
import com.trollworks.toolkit.utility.FileProxy;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.PrintProxy;
import com.trollworks.toolkit.utility.text.IntegerFormatter;

import java.awt.BorderLayout;
import java.awt.DefaultFocusTraversalPolicy;
import java.awt.Window;
import java.awt.event.KeyEvent;
import java.io.File;

import javax.swing.Icon;
import javax.swing.JLabel;
import javax.swing.JScrollPane;
import javax.swing.KeyStroke;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

import org.apache.pdfbox.io.MemoryUsageSetting;
import org.apache.pdfbox.pdmodel.PDDocument;

/** Provides the ability to view a PDF. */
public class PdfDockable extends Dockable implements FileProxy, CloseHandler {
	@Localize("Previous Page")
	private static String	PREVIOUS_PAGE;
	@Localize("Next Page")
	private static String	NEXT_PAGE;
	@Localize("Scale Document Up")
	private static String	SCALE_DOC_UP;
	@Localize("Scale Document Down")
	private static String	SCALE_DOC_DOWN;
	@Localize("Actual Size")
	private static String	ACTUAL_SIZE;

	static {
		Localization.initialize();
	}

	private File		mFile;
	private PDDocument	mPdf;
	private Toolbar		mToolbar;
	private PdfPanel	mPanel;
	private IconButton	mZoomInButton;
	private IconButton	mZoomOutButton;
	private IconButton	mActualSizeButton;
	private JLabel		mZoomStatus;
	private EditorField	mPageField;
	private JLabel		mPageStatus;
	private IconButton	mPreviousPageButton;
	private IconButton	mNextPageButton;

	public PdfDockable(File pdf, int page) {
		super(new BorderLayout());
		mFile = pdf;
		int pageCount = 9999;
		try {
			mPdf = PDDocument.load(pdf, MemoryUsageSetting.setupMixed(50 * 1024 * 1024));
			pageCount = mPdf.getNumberOfPages();
		} catch (Exception exception) {
			Log.error(exception);
		}
		mToolbar = new Toolbar();

		mZoomInButton = new IconButton(StdImage.get("ZoomIn"), formatWithKey(SCALE_DOC_UP, KeyStroke.getKeyStroke('=')), () -> mPanel.zoomIn()); //$NON-NLS-1$
		mToolbar.add(mZoomInButton);
		mZoomOutButton = new IconButton(StdImage.get("ZoomOut"), formatWithKey(SCALE_DOC_DOWN, KeyStroke.getKeyStroke('-')), () -> mPanel.zoomOut()); //$NON-NLS-1$
		mToolbar.add(mZoomOutButton);
		mActualSizeButton = new IconButton(StdImage.get("ActualSize"), formatWithKey(ACTUAL_SIZE, KeyStroke.getKeyStroke('1')), () -> mPanel.actualSize()); //$NON-NLS-1$
		mToolbar.add(mActualSizeButton);
		mZoomStatus = new JLabel("100%"); //$NON-NLS-1$
		mToolbar.add(mZoomStatus);

		mPageField = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(1, pageCount, false)), event -> {
			if (mPanel != null) {
				int pageIndex = ((Integer) mPageField.getValue()).intValue() - 1;
				int newPageIndex = mPanel.goToPageIndex(pageIndex);
				if (pageIndex != newPageIndex) {
					mPageField.setValue(Integer.valueOf(newPageIndex + 1));
				} else {
					mPanel.requestFocusInWindow();
				}
			}
		} , SwingConstants.RIGHT, Integer.valueOf(page), Integer.valueOf(9999), null);
		mToolbar.add(mPageField, Toolbar.LAYOUT_EXTRA_BEFORE);
		mPageStatus = new JLabel("/ -"); //$NON-NLS-1$
		mToolbar.add(mPageStatus);
		mPreviousPageButton = new IconButton(StdImage.get("PageUp"), formatWithKey(PREVIOUS_PAGE, KeyStroke.getKeyStroke(KeyEvent.VK_UP, 0)), () -> mPanel.previousPage()); //$NON-NLS-1$
		mToolbar.add(mPreviousPageButton);
		mNextPageButton = new IconButton(StdImage.get("PageDown"), formatWithKey(NEXT_PAGE, KeyStroke.getKeyStroke(KeyEvent.VK_DOWN, 0)), () -> mPanel.nextPage()); //$NON-NLS-1$
		mToolbar.add(mNextPageButton);

		add(mToolbar, BorderLayout.NORTH);
		mPanel = new PdfPanel(this, mPdf, page);
		add(new JScrollPane(mPanel), BorderLayout.CENTER);

		setFocusCycleRoot(true);
		setFocusTraversalPolicy(new DefaultFocusTraversalPolicy());
	}

	private static String formatWithKey(String title, KeyStroke key) {
		return title + " [" + KeyStrokeDisplay.getKeyStrokeDisplay(key) + "]"; //$NON-NLS-1$ //$NON-NLS-2$
	}

	public void updateStatus(int page, int pageCount, float scale) {
		mZoomInButton.setEnabled(PdfPanel.SCALES[PdfPanel.SCALES.length - 1] != scale);
		mZoomOutButton.setEnabled(PdfPanel.SCALES[0] != scale);
		mActualSizeButton.setEnabled(scale != 1f);
		boolean revalidate = updateZoomInfo(scale);
		revalidate |= updatePageInfo(page, pageCount);
		if (revalidate) {
			mToolbar.revalidate();
			mToolbar.repaint();
		}
	}

	private boolean updateZoomInfo(float scale) {
		String text = (int) (scale * 100) + "%"; //$NON-NLS-1$
		if (!text.equals(mZoomStatus.getText())) {
			mZoomStatus.setText(text);
			return true;
		}
		return false;
	}

	private boolean updatePageInfo(int page, int pageCount) {
		mPreviousPageButton.setEnabled(page > 0);
		mNextPageButton.setEnabled(page < pageCount - 1);
		mPageField.setValue(Integer.valueOf(page + 1));
		String text = "/ " + (pageCount > 0 ? Integer.valueOf(pageCount) : "-"); //$NON-NLS-1$ //$NON-NLS-2$
		if (!text.equals(mPageStatus.getText())) {
			mPageStatus.setText(text);
			return true;
		}
		return false;
	}

	public void goToPage(int page) {
		mPanel.goToPage(page);
	}

	@Override
	public boolean mayAttemptClose() {
		return true;
	}

	@Override
	public boolean attemptClose() {
		try {
			getDockContainer().close(this);
		} finally {
			if (mPdf != null) {
				try {
					mPdf.close();
				} catch (Exception exception) {
					Log.error(exception);
				}
				mPdf = null;
			}
		}
		return true;
	}

	@Override
	public String getDescriptor() {
		return null;
	}

	@Override
	public Icon getTitleIcon() {
		return null;
	}

	@Override
	public String getTitle() {
		return PathUtils.getLeafName(mFile.getName(), false);
	}

	@Override
	public String getTitleTooltip() {
		return null;
	}

	@Override
	public File getBackingFile() {
		return mFile;
	}

	@Override
	public void toFrontAndFocus() {
		Window window = UIUtilities.getAncestorOfType(this, Window.class);
		if (window != null) {
			window.toFront();
		}
		DockContainer dc = getDockContainer();
		dc.setCurrentDockable(this);
		dc.doLayout();
		dc.acquireFocus();
	}

	@Override
	public PrintProxy getPrintProxy() {
		return null;
	}
}
