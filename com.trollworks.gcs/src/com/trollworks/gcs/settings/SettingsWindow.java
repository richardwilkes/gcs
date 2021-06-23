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
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.LayoutConstants;
import com.trollworks.gcs.ui.widget.Menu;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import java.awt.BorderLayout;
import java.awt.Dimension;
import java.awt.event.WindowEvent;
import java.util.HashMap;
import java.util.Map;

public abstract class SettingsWindow extends BaseWindow implements CloseHandler {
    private static final Map<Class<? extends SettingsWindow>, SettingsWindow> INSTANCES = new HashMap<>();

    private FontAwesomeButton mResetButton;
    private FontAwesomeButton mMenuButton;

    /** Displays the color settings window. */
    public static void display(Class<? extends SettingsWindow> key) {
        if (!UIUtilities.inModalState()) {
            SettingsWindow wnd;
            synchronized (INSTANCES) {
                wnd = INSTANCES.get(key);
                if (wnd == null) {
                    try {
                        wnd = key.getConstructor().newInstance();
                    } catch (Exception exception) {
                        Log.error(exception);
                        return;
                    }
                    INSTANCES.put(key, wnd);
                }
            }
            wnd.setVisible(true);
        }
    }

    protected SettingsWindow(String title) {
        super(title);
        addToolBar();
        getContentPane().add(new ScrollPanel(createContent()), BorderLayout.CENTER);
        adjustResetButton();
        establishSizing();
        WindowUtils.packAndCenterWindowOn(this, null);
    }

    private void addToolBar() {
        Panel header = new Panel(new PrecisionLayout().setColumns(2).
                setMargins(LayoutConstants.TOOLBAR_VERTICAL_INSET,
                        LayoutConstants.WINDOW_BORDER_INSET,
                        LayoutConstants.TOOLBAR_VERTICAL_INSET,
                        LayoutConstants.WINDOW_BORDER_INSET).
                setHorizontalSpacing(10).setEndHorizontalAlignment());
        header.setBorder(new LineBorder(Colors.DIVIDER, 0, 0, 1, 0));
        mResetButton = new FontAwesomeButton("\uf011", I18n.text("Reset to Factory Defaults"),
                this::reset);
        header.add(mResetButton);
        mMenuButton = new FontAwesomeButton("\uf0c9", I18n.text("Menu"),
                () -> createActionMenu().presentToUser(mMenuButton, 0, mMenuButton::updateRollOver));
        header.add(mMenuButton);
        getContentPane().add(header, BorderLayout.NORTH);
    }

    protected abstract Panel createContent();

    @Override
    public void establishSizing() {
        pack();
        int width = getSize().width;
        setMinimumSize(new Dimension(width, 200));
        setMaximumSize(new Dimension(width, getPreferredSize().height));
    }

    protected final void adjustResetButton() {
        mResetButton.setEnabled(shouldResetBeEnabled());
    }

    protected abstract boolean shouldResetBeEnabled();

    protected abstract void reset();

    protected abstract Menu createActionMenu();

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
        synchronized (INSTANCES) {
            INSTANCES.remove(getClass());
        }
        super.dispose();
    }
}
