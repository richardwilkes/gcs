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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.awt.event.WindowEvent;

/** Provides the "Close" command. */
public class CloseCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_CLOSE = "Close";

    /** The singleton {@link CloseCommand}. */
    public static final CloseCommand INSTANCE = new CloseCommand();

    private CloseCommand() {
        super(I18n.Text("Close"), CMD_CLOSE, KeyEvent.VK_W);
    }

    @Override
    public void adjust() {
        CloseHandler handler = getTarget(CloseHandler.class);
        setEnabled(handler != null && handler.mayAttemptClose());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        CloseHandler handler = getTarget(CloseHandler.class);
        if (handler != null) {
            if (handler.mayAttemptClose()) {
                handler.attemptClose();
            }
        }
    }

    /**
     * @param window The {@link Window} to close.
     * @return Whether the window was closed or not.
     */
    public static boolean close(Window window) {
        if (window == null) {
            return true;
        }
        window.dispatchEvent(new WindowEvent(window, WindowEvent.WINDOW_CLOSING));
        return !window.isShowing();
    }
}
