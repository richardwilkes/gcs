/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.pdfview;

import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.ui.widget.KeyStrokeDisplay;
import com.trollworks.gcs.ui.widget.Toolbar;
import com.trollworks.gcs.ui.widget.dock.DockContainer;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.utility.FileProxy;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.text.IntegerFormatter;

import java.awt.BorderLayout;
import java.awt.DefaultFocusTraversalPolicy;
import java.awt.Window;
import java.awt.event.KeyEvent;
import java.nio.file.Path;
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
    private Path        mPath;
    private PDDocument  mPdf;
    private Toolbar     mToolbar;
    private PdfPanel    mPanel;
    private IconButton  mZoomInButton;
    private IconButton  mZoomOutButton;
    private IconButton  mActualSizeButton;
    private JLabel      mZoomStatus;
    private EditorField mPageField;
    private JLabel      mPageStatus;
    private IconButton  mPreviousPageButton;
    private IconButton  mNextPageButton;

    public PdfDockable(PdfRef pdfRef, int page, String highlight) {
        super(new BorderLayout());
        mPath = pdfRef.getPath();
        int pageCount = 9999;
        try {
            mPdf = PDDocument.load(pdfRef.getPath().toFile(), MemoryUsageSetting.setupMixed(50 * 1024 * 1024));
            pageCount = mPdf.getNumberOfPages();
        } catch (Exception exception) {
            Log.error(exception);
        }
        mToolbar = new Toolbar();

        mZoomInButton = new IconButton(Images.ZOOM_IN, formatWithKey(I18n.Text("Scale Document Up"), KeyStroke.getKeyStroke('=')), () -> mPanel.zoomIn());
        mToolbar.add(mZoomInButton);
        mZoomOutButton = new IconButton(Images.ZOOM_OUT, formatWithKey(I18n.Text("Scale Document Down"), KeyStroke.getKeyStroke('-')), () -> mPanel.zoomOut());
        mToolbar.add(mZoomOutButton);
        mActualSizeButton = new IconButton(Images.ACTUAL_SIZE, formatWithKey(I18n.Text("Actual Size"), KeyStroke.getKeyStroke('1')), () -> mPanel.actualSize());
        mToolbar.add(mActualSizeButton);
        mZoomStatus = new JLabel("100%");
        mToolbar.add(mZoomStatus);

        mPageField = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(1, pageCount, false)), event -> {
            if (mPanel != null) {
                int pageIndex    = ((Integer) mPageField.getValue()).intValue() - 1;
                int newPageIndex = mPanel.goToPageIndex(pageIndex, null);
                if (pageIndex == newPageIndex) {
                    mPanel.requestFocus();
                } else {
                    mPageField.setValue(Integer.valueOf(newPageIndex + 1));
                }
            }
        }, SwingConstants.RIGHT, Integer.valueOf(Math.max(page, 1)), Integer.valueOf(9999), null);
        mToolbar.add(mPageField, Toolbar.LAYOUT_EXTRA_BEFORE);
        mPageStatus = new JLabel("/ -");
        mToolbar.add(mPageStatus);
        mPreviousPageButton = new IconButton(Images.PAGE_UP, formatWithKey(I18n.Text("Previous Page"), KeyStroke.getKeyStroke(KeyEvent.VK_UP, 0)), () -> mPanel.previousPage());
        mToolbar.add(mPreviousPageButton);
        mNextPageButton = new IconButton(Images.PAGE_DOWN, formatWithKey(I18n.Text("Next Page"), KeyStroke.getKeyStroke(KeyEvent.VK_DOWN, 0)), () -> mPanel.nextPage());
        mToolbar.add(mNextPageButton);

        add(mToolbar, BorderLayout.NORTH);
        mPanel = new PdfPanel(this, mPdf, pdfRef, page, highlight);
        add(new JScrollPane(mPanel), BorderLayout.CENTER);

        setFocusCycleRoot(true);
        setFocusTraversalPolicy(new DefaultFocusTraversalPolicy());
    }

    private static String formatWithKey(String title, KeyStroke key) {
        return title + " [" + KeyStrokeDisplay.getKeyStrokeDisplay(key) + "]";
    }

    public void updateStatus(int page, int pageCount, float scale) {
        mZoomInButton.setEnabled(PdfPanel.SCALES[PdfPanel.SCALES.length - 1] != scale);
        mZoomOutButton.setEnabled(PdfPanel.SCALES[0] != scale);
        mActualSizeButton.setEnabled(scale != 1.0f);
        boolean revalidate = updateZoomInfo(scale);
        revalidate |= updatePageInfo(page, pageCount);
        if (revalidate) {
            mToolbar.revalidate();
            mToolbar.repaint();
        }
    }

    private boolean updateZoomInfo(float scale) {
        String text = (int) (scale * 100) + "%";
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
        String text = "/ " + (pageCount > 0 ? Integer.valueOf(pageCount) : "-");
        if (!text.equals(mPageStatus.getText())) {
            mPageStatus.setText(text);
            return true;
        }
        return false;
    }

    public void goToPage(PdfRef pdfRef, int page, String highlight) {
        mPanel.goToPage(pdfRef, page, highlight);
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
    public Icon getTitleIcon() {
        return FileType.PDF.getIcon();
    }

    @Override
    public String getTitle() {
        return PathUtils.getLeafName(mPath, false);
    }

    @Override
    public String getTitleTooltip() {
        return null;
    }

    @Override
    public Path getBackingFile() {
        return mPath;
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
