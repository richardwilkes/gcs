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

import com.trollworks.gcs.pageref.PageRef;
import com.trollworks.gcs.pageref.PageRefSettings;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.IntegerFormatter;

import java.awt.Color;
import java.awt.Component;
import java.io.IOException;
import java.nio.file.Path;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;
import javax.swing.text.DefaultFormatterFactory;

/** A window for editing page reference lookup settings. */
public final class PageRefSettingsWindow extends SettingsWindow<PageRefSettings> {
    private static PageRefSettingsWindow INSTANCE;

    private BandedPanel mPanel;

    /** Displays the page reference lookup settings window. */
    public static void display() {
        if (!UIUtilities.inModalState()) {
            PageRefSettingsWindow wnd;
            synchronized (PageRefSettingsWindow.class) {
                if (INSTANCE == null) {
                    INSTANCE = new PageRefSettingsWindow();
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

    private PageRefSettingsWindow() {
        super(I18n.text("Page Reference Settings"));
        fill();
    }

    @Override
    protected void preDispose() {
        synchronized (PageRefSettingsWindow.class) {
            INSTANCE = null;
        }
    }

    @Override
    protected Panel createContent() {
        mPanel = new BandedPanel(true);
        buildPanel();
        return mPanel;
    }

    private void buildPanel() {
        Color background = new Color(255, 255, 224);
        mPanel.removeAll();
        mPanel.setLayout(new PrecisionLayout().setColumns(4).setMargins(0, 10, 0, 10).setVerticalSpacing(0));
        for (PageRef ref : Settings.getInstance().getPDFRefSettings().list()) {
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
            FontIconButton removeButton = new FontIconButton(FontAwesome.TRASH, I18n.text("Remove"), null);
            removeButton.setClickFunction((b) -> {
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
                    Settings.getInstance().getPDFRefSettings().remove(ref);
                    Component[] children = mPanel.getComponents();
                    int         length   = children.length;
                    for (int i = 0; i < length; i++) {
                        if (children[i] == b) {
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
    protected boolean shouldResetBeEnabled() {
        return !Settings.getInstance().getPDFRefSettings().isEmpty();
    }

    @Override
    protected PageRefSettings getResetData() {
        return new PageRefSettings();
    }

    @Override
    protected void doResetTo(PageRefSettings data) {
        Settings.getInstance().getPDFRefSettings().copyFrom(data);
        buildPanel();
    }

    @Override
    protected Dirs getDir() {
        return Dirs.SETTINGS;
    }

    @Override
    protected FileType getFileType() {
        return FileType.PAGE_REF_SETTINGS;
    }

    @Override
    protected PageRefSettings createSettingsFrom(Path path) throws IOException {
        return new PageRefSettings(path);
    }

    @Override
    protected void exportSettingsTo(Path path) throws IOException {
        Settings.getInstance().getPDFRefSettings().save(path);
    }
}
