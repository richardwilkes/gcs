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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;

import java.awt.Container;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import javax.swing.Icon;
import javax.swing.JRootPane;

public class MenuItem extends Label {
    private Command           mCommand;
    private SelectionListener mSelectionListener;
    private boolean           mHighlighted;

    public interface SelectionListener {
        void menuItemSelected(MenuItem item);
    }

    public MenuItem(String title, SelectionListener listener) {
        this(null, title, listener);
    }

    public MenuItem(Icon icon, String title, SelectionListener listener) {
        super(icon, title);
        setOpaque(true);
        setThemeFont(Fonts.BUTTON);
        setBorder(new EmptyBorder(Button.V_MARGIN, Button.H_MARGIN, Button.V_MARGIN, Button.H_MARGIN));
        mSelectionListener = listener;
    }

    public MenuItem(Command cmd) {
        this(cmd.getTitle(), null);
        mCommand = cmd;
    }

    public void adjust() {
        if (mCommand != null) {
            mCommand.adjust();
            String title = mCommand.getTitle();
            if (!getText().equals(title)) {
                setText(title);
            }
            setEnabled(mCommand.isEnabled());
        }
    }

    public void click() {
        if (mCommand != null) {
            mCommand.actionPerformed(new ActionEvent(this, ActionEvent.ACTION_PERFORMED, mCommand.getCommand()));
        } else if (mSelectionListener != null) {
            mSelectionListener.menuItemSelected(this);
        }
    }

    public boolean isHighlighted() {
        return mHighlighted;
    }

    public void setHighlighted(boolean highlighted) {
        if (mHighlighted != highlighted) {
            mHighlighted = highlighted;
            setStdColors();
        }
    }

    @Override
    protected void setStdColors() {
        if (mHighlighted) {
            setBackground(Colors.CONTROL_PRESSED);
            setForeground(Colors.ON_CONTROL_PRESSED);
        } else {
            setBackground(Colors.CONTROL);
            setForeground(Colors.ON_CONTROL);
        }
    }

    @Override
    public void repaint(long tm, int x, int y, int width, int height) {
        // Redirect the repaint to the top level manually, since things don't work correctly
        // otherwise -- for no apparent reason that I can determine.
        JRootPane root = getRootPane();
        if (root != null) {
            Container parent = root.getParent();
            if (parent != null) {
                Rectangle bounds = new Rectangle(x, y, width, height);
                UIUtilities.convertRectangle(bounds, this, parent);
                parent.repaint(tm, bounds.x, bounds.y, bounds.width, bounds.height);
            }
        }
    }
}
