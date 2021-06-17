/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.settings;

import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.pdfview.PDFRef;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.IntegerFormatter;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.event.WindowEvent;
import java.nio.file.Path;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;
import javax.swing.text.DefaultFormatterFactory;

/** A window for editing page reference lookup settings. */
public final class PageRefMappingsWindow extends BaseWindow implements CloseHandler {
    private static PageRefMappingsWindow INSTANCE;
    private        BandedPanel           mPanel;

    /** Displays the page reference lookup settings window. */
    public static void display() {
        if (!UIUtilities.inModalState()) {
            PageRefMappingsWindow wnd;
            synchronized (PageRefMappingsWindow.class) {
                if (INSTANCE == null) {
                    INSTANCE = new PageRefMappingsWindow();
                }
                wnd = INSTANCE;
            }
            wnd.setVisible(true);
        }
    }

    public static void rebuild() {
        if (INSTANCE != null) {
            INSTANCE.buildPanel();
        }
    }

    private PageRefMappingsWindow() {
        super(I18n.text("Page Reference Mappings"));
        setLayout(new BorderLayout());
        mPanel = new BandedPanel(true);
        buildPanel();
        getContentPane().add(new ScrollPanel(mPanel), BorderLayout.CENTER);
        WindowUtils.packAndCenterWindowOn(this, null);
    }

    private void buildPanel() {
        Settings prefs      = Settings.getInstance();
        Color    background = new Color(255, 255, 224);
        mPanel.removeAll();
        mPanel.setLayout(new PrecisionLayout().setColumns(4).setMargins(0, 10, 0, 10).setVerticalSpacing(0));
        for (PDFRef ref : prefs.allPdfRefs(false)) {
            Label idLabel = new Label(ref.getID(), SwingConstants.CENTER);
            idLabel.setBorder(new CompoundBorder(new LineBorder(), new EmptyBorder(1, 4, 1, 4)));
            idLabel.setOpaque(true);
            idLabel.setBackground(background);
            Panel wrapper = new Panel(new PrecisionLayout().setMargins(6, 0, 6, 0), false);
            wrapper.add(idLabel, new PrecisionLayoutData().setFillHorizontalAlignment().setMinimumWidth(50).setVerticalAlignment(PrecisionLayoutAlignment.MIDDLE));
            mPanel.add(wrapper, new PrecisionLayoutData().setFillAlignment());
            EditorField field = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(-9999, 9999, true)),
                    (f) -> ref.setPageToIndexOffset(((Integer) f.getValue()).intValue()),
                    SwingConstants.RIGHT, Integer.valueOf(ref.getPageToIndexOffset()),
                    Integer.valueOf(-9999),
                    I18n.text("If your PDF is opening up to the wrong page when opening page references, enter an offset here to compensate."));
            mPanel.add(field);
            Path  path      = ref.getPath().normalize().toAbsolutePath();
            Label fileLabel = new Label(path.getFileName().toString());
            fileLabel.setToolTipText(path.toString());
            mPanel.add(fileLabel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
            FontAwesomeButton removeButton = new FontAwesomeButton("\uf1f8", I18n.text("Remove"), null);
            removeButton.setClickFunction(() -> {
                Modal dialog = Modal.prepareToShowMessage(this,
                        I18n.text("Confirm Change"),
                        MessageType.QUESTION,
                        String.format(I18n.text("""
                                Are you sure you want to remove this page reference
                                mapping from %s to "%s"?"""), ref.getID(), ref.getPath().getFileName().toString()));
                dialog.addCancelButton();
                dialog.addButton(I18n.text("Remove"), Modal.OK);
                dialog.presentToUser();
                if (dialog.getResult() == Modal.OK) {
                    Settings.getInstance().removePdfRef(ref);
                    Component[] children = mPanel.getComponents();
                    int         length   = children.length;
                    for (int i = 0; i < length; i++) {
                        if (children[i] == removeButton) {
                            int max = ((PrecisionLayout) mPanel.getLayout()).getColumns();
                            for (int j = 0; j < max; j++) {
                                mPanel.remove(i - j);
                            }
                            break;
                        }
                    }
                    mPanel.revalidate();
                    mPanel.repaint();
                }
            });
            removeButton.setBorder(new EmptyBorder(4));
            mPanel.add(removeButton);
        }
        if (mPanel.getComponentCount() == 0) {
            mPanel.setLayout(new PrecisionLayout().setMargins(10));
            mPanel.add(new Label(I18n.text("No page reference mappings have been set."), SwingConstants.CENTER), new PrecisionLayoutData().setFillAlignment().setGrabSpace(true));
        }
        mPanel.revalidate();
        mPanel.repaint();
    }

    @Override
    public boolean mayAttemptClose() {
        return true;
    }

    @Override
    public boolean attemptClose() {
        windowClosing(new WindowEvent(this, WindowEvent.WINDOW_CLOSING));
        return true;
    }

    @Override
    public void dispose() {
        synchronized (PageRefMappingsWindow.class) {
            INSTANCE = null;
        }
        super.dispose();
    }
}
