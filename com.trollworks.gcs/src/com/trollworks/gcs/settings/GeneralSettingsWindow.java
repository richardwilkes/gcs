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

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.menu.file.ExportToGCalcCommand;
import com.trollworks.gcs.pageref.PDFViewer;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.Button;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.LayoutConstants;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import java.awt.Cursor;
import java.awt.Desktop;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.io.IOException;
import java.net.URI;
import java.nio.file.Path;
import javax.swing.SwingConstants;

public final class GeneralSettingsWindow extends SettingsWindow<GeneralSettings> {
    private static GeneralSettingsWindow INSTANCE;

    private EditorField            mPlayerName;
    private EditorField            mTechLevel;
    private EditorField            mInitialPoints;
    private Checkbox               mAutoFillProfile;
    private PopupMenu<Scales>      mInitialScale;
    private EditorField            mToolTipTimeout;
    private EditorField            mImageResolution;
    private Checkbox               mIncludeUnspentPointsInTotal;
    private EditorField            mGCalcKey;
    private PopupMenu<CalendarRef> mCalendar;
    private PopupMenu<PDFViewer>   mPDFViewer;
    private Label                  mPDFInstall;
    private Label                  mPDFLink;
    private Button                 mResetButton;

    /** Displays the general settings window. */
    public static void display() {
        if (!UIUtilities.inModalState()) {
            GeneralSettingsWindow wnd;
            synchronized (GeneralSettingsWindow.class) {
                if (INSTANCE == null) {
                    INSTANCE = new GeneralSettingsWindow();
                }
                wnd = INSTANCE;
            }
            wnd.setVisible(true);
        }
    }

    private GeneralSettingsWindow() {
        super(I18n.text("General Settings"));
        fill();
    }

    @Override
    protected void preDispose() {
        synchronized (GeneralSettingsWindow.class) {
            INSTANCE = null;
        }
    }

    @Override
    protected Panel createContent() {
        GeneralSettings settings = Settings.getInstance().getGeneralSettings();
        Panel           panel    = new Panel(new PrecisionLayout().setColumns(3).setMargins(LayoutConstants.WINDOW_BORDER_INSET));

        // First row
        mPlayerName = new EditorField(FieldFactory.STRING, (f) -> {
            Settings.getInstance().getGeneralSettings().setDefaultPlayerName(f.getText().trim());
            adjustResetButton();
        }, SwingConstants.LEFT, settings.getDefaultPlayerName(),
                I18n.text("The player name to use when a new character sheet is created"));
        panel.add(new Label(I18n.text("Player")), new PrecisionLayoutData().setEndHorizontalAlignment());
        panel.add(mPlayerName, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        mAutoFillProfile = new Checkbox(I18n.text("Fill in initial description"),
                settings.autoFillProfile(), (b) -> {
            Settings.getInstance().getGeneralSettings().setAutoFillProfile(b.isChecked());
            adjustResetButton();
        });
        mAutoFillProfile.setToolTipText(I18n.text("Automatically fill in new character identity and description information with randomized choices"));
        mAutoFillProfile.setOpaque(false);
        panel.add(mAutoFillProfile, new PrecisionLayoutData().setLeftMargin(10));

        // Second row
        mTechLevel = new EditorField(FieldFactory.STRING, (f) -> {
            Settings.getInstance().getGeneralSettings().setDefaultTechLevel(f.getText().trim());
            adjustResetButton();
        }, SwingConstants.RIGHT, settings.getDefaultTechLevel(), "99+99^", getTechLevelTooltip());
        panel.add(new Label(I18n.text("Tech Level")),
                new PrecisionLayoutData().setEndHorizontalAlignment());
        Wrapper wrapper = new Wrapper(new PrecisionLayout().setMargins(0).setColumns(3));
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        wrapper.add(mTechLevel, new PrecisionLayoutData().setFillHorizontalAlignment());

        mInitialPoints = new EditorField(FieldFactory.POSINT6, (f) -> {
            Settings.getInstance().getGeneralSettings().setInitialPoints(((Integer) f.getValue()).intValue());
            adjustResetButton();
        }, SwingConstants.RIGHT, Integer.valueOf(settings.getInitialPoints()), Integer.valueOf(999999),
                I18n.text("The initial number of character points to start with"));
        wrapper.add(new Label(I18n.text("Initial Points")),
                new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(10));
        wrapper.add(mInitialPoints, new PrecisionLayoutData().setFillHorizontalAlignment());

        mIncludeUnspentPointsInTotal = new Checkbox(I18n.text("Include unspent points in total"),
                settings.includeUnspentPointsInTotal(), (b) -> {
            Settings s = Settings.getInstance();
            s.getGeneralSettings().setIncludeUnspentPointsInTotal(b.isChecked(), s);
            adjustResetButton();
        });
        mIncludeUnspentPointsInTotal.setToolTipText(I18n.text("Include unspent points in the character point total"));
        mIncludeUnspentPointsInTotal.setOpaque(false);
        panel.add(mIncludeUnspentPointsInTotal, new PrecisionLayoutData().setLeftMargin(10));

        // Third row
        mCalendar = new PopupMenu<>(CalendarRef.choices(), (p) -> {
            Settings.getInstance().getGeneralSettings().setCalendarRef(p.getSelectedItem().name());
            adjustResetButton();
        });
        mCalendar.setSelectedItem(CalendarRef.current(), false);
        panel.add(new Label(I18n.text("Calendar")),
                new PrecisionLayoutData().setEndHorizontalAlignment());
        panel.add(mCalendar, new PrecisionLayoutData().setHorizontalSpan(2));

        // Fourth row
        mInitialScale = new PopupMenu<>(Scales.values(), (p) -> {
            Settings.getInstance().getGeneralSettings().setInitialUIScale(p.getSelectedItem());
            adjustResetButton();
        });
        mInitialScale.setSelectedItem(settings.getInitialUIScale(), false);
        panel.add(new Label(I18n.text("Initial Scale")),
                new PrecisionLayoutData().setEndHorizontalAlignment());
        wrapper = new Wrapper(new PrecisionLayout().setMargins(0).setColumns(7));
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().
                setGrabHorizontalSpace(true).setHorizontalSpan(2));
        wrapper.add(mInitialScale);

        mToolTipTimeout = new EditorField(FieldFactory.TOOLTIP_TIMEOUT, (f) -> {
            Settings.getInstance().getGeneralSettings().setToolTipTimeout(((Integer) f.getValue()).intValue());
            adjustResetButton();
        }, SwingConstants.RIGHT, Integer.valueOf(settings.getToolTipTimeout()),
                FieldFactory.getMaxValue(FieldFactory.TOOLTIP_TIMEOUT),
                I18n.text("The number of seconds before tooltips will dismiss themselves"));
        wrapper.add(new Label(I18n.text("Tooltip Timeout")),
                new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(10));
        wrapper.add(mToolTipTimeout, new PrecisionLayoutData().setFillHorizontalAlignment());
        wrapper.add(new Label(I18n.text("seconds")));

        mImageResolution = new EditorField(FieldFactory.OUTPUT_DPI, (f) -> {
            Settings.getInstance().getGeneralSettings().setImageResolution(((Integer) f.getValue()).intValue());
            adjustResetButton();
        }, SwingConstants.RIGHT, Integer.valueOf(settings.getImageResolution()),
                FieldFactory.getMaxValue(FieldFactory.OUTPUT_DPI),
                I18n.text("The resolution, in dots-per-inch, to use when saving sheets as PNG files"));
        wrapper.add(new Label(I18n.text("Image Resolution")),
                new PrecisionLayoutData().setFillHorizontalAlignment().setLeftMargin(10));
        wrapper.add(mImageResolution, new PrecisionLayoutData().setFillHorizontalAlignment());
        wrapper.add(new Label(I18n.text("dpi")));

        // Fifth row
        mPDFViewer = new PopupMenu<>(PDFViewer.valuesForPlatform(), (p) -> {
            PDFViewer pdfViewer = p.getSelectedItem();
            if (pdfViewer != null) {
                Settings.getInstance().getGeneralSettings().setPDFViewer(pdfViewer);
                updatePDFLinks(pdfViewer);
                adjustResetButton();
            }
        });
        PDFViewer pdfViewer = settings.getPDFViewer();
        mPDFViewer.setSelectedItem(pdfViewer, false);
        panel.add(new Label(I18n.text("PDF Viewer")), new PrecisionLayoutData().setEndHorizontalAlignment());
        wrapper = new Wrapper(new PrecisionLayout().setMargins(0).setColumns(3));
        wrapper.add(mPDFViewer);
        mPDFInstall = new Label("");
        wrapper.add(mPDFInstall, new PrecisionLayoutData().setLeftMargin(10));
        mPDFLink = new Label("");
        mPDFLink.setForeground(Colors.ICON_BUTTON_PRESSED);
        mPDFLink.setCursor(Cursor.getPredefinedCursor(Cursor.HAND_CURSOR));
        mPDFLink.addMouseListener(new MouseAdapter() {
            @Override
            public void mouseClicked(MouseEvent event) {
                String text = mPDFLink.getText();
                if (!text.isBlank()) {
                    try {
                        Desktop.getDesktop().browse(new URI(text));
                    } catch (Exception ex) {
                        Log.error(ex);
                    }
                }
            }
        });
        wrapper.add(mPDFLink, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(2));
        updatePDFLinks(pdfViewer);

        // Sixth row
        wrapper = new Wrapper(new PrecisionLayout().setMargins(0).setColumns(3));
        panel.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(3));
        mGCalcKey = new EditorField(FieldFactory.STRING, (f) -> {
            Settings.getInstance().getGeneralSettings().setGCalcKey(f.getText().trim());
            adjustResetButton();
        }, SwingConstants.LEFT, settings.getGCalcKey(), null);
        wrapper.add(new Label(I18n.text("GURPS Calculator Key")), new PrecisionLayoutData().setEndHorizontalAlignment());
        wrapper.add(mGCalcKey, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        wrapper.add(new FontIconButton(FontAwesome.SEARCH,
                I18n.text("Lookup your key on the GURPS Calculator web site"),
                ExportToGCalcCommand::openBrowserToFindKey));

        return panel;
    }

    public static String getTechLevelTooltip() {
        return I18n.text("""
                TL0: Stone Age (Prehistory and later)
                TL1: Bronze Age (3500 B.C.+)
                TL2: Iron Age (1200 B.C.+)
                TL3: Medieval (600 A.D.+)
                TL4: Age of Sail (1450+)
                TL5: Industrial Revolution (1730+)
                TL6: Mechanized Age (1880+)
                TL7: Nuclear Age (1940+)
                TL8: Digital Age (1980+)
                TL9: Microtech Age (2025+?)
                TL10: Robotic Age (2070+?)
                TL11: Age of Exotic Matter
                TL12: Anything Goes""");
    }

    @Override
    public void establishSizing() {
        setResizable(false);
    }

    @Override
    protected boolean shouldResetBeEnabled() {
        return !Settings.getInstance().getGeneralSettings().equals(new GeneralSettings());
    }

    @Override
    protected GeneralSettings getResetData() {
        return new GeneralSettings();
    }

    @Override
    protected void doResetTo(GeneralSettings data) {
        GeneralSettings settings = Settings.getInstance().getGeneralSettings();
        settings.copyFrom(data);
        mPlayerName.setValue(settings.getDefaultPlayerName());
        mTechLevel.setValue(settings.getDefaultTechLevel());
        mInitialPoints.setValue(Integer.valueOf(settings.getInitialPoints()));
        mAutoFillProfile.setChecked(settings.autoFillProfile());
        mCalendar.setSelectedItem(CalendarRef.get(settings.calendarRef()), true);
        mInitialScale.setSelectedItem(settings.getInitialUIScale(), true);
        mToolTipTimeout.setValue(Integer.valueOf(settings.getToolTipTimeout()));
        mImageResolution.setValue(Integer.valueOf(settings.getImageResolution()));
        mIncludeUnspentPointsInTotal.setChecked(settings.includeUnspentPointsInTotal());
        mGCalcKey.setValue(settings.getGCalcKey());
        PDFViewer pdfViewer = settings.getPDFViewer();
        mPDFViewer.setSelectedItem(pdfViewer, true);
        updatePDFLinks(pdfViewer);
    }

    private static String getInstallFromText() {
        return I18n.text("Install from:");
    }

    private void updatePDFLinks(PDFViewer viewer) {
        String from = viewer.installFrom();
        mPDFInstall.setText(from.isBlank() ? "" : getInstallFromText());
        mPDFInstall.revalidate();
        mPDFInstall.repaint();
        mPDFLink.setText(from);
        mPDFLink.revalidate();
        mPDFLink.repaint();
    }

    @Override
    protected Dirs getDir() {
        return Dirs.SETTINGS;
    }

    @Override
    protected FileType getFileType() {
        return FileType.GENERAL_SETTINGS;
    }

    @Override
    protected GeneralSettings createSettingsFrom(Path path) throws IOException {
        return new GeneralSettings(path);
    }

    @Override
    protected void exportSettingsTo(Path path) throws IOException {
        Settings.getInstance().getGeneralSettings().save(path);
    }
}
