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

import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.RenderingHints;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.font.TextAttribute;
import java.awt.geom.Path2D;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import javax.swing.SwingConstants;

public class PopupMenu<T> extends Panel implements MouseListener, KeyListener, FocusListener {
    public static final int                  GAP        = 4;
    public static final String               POPUP_MARK = "\uf0d7";
    private             List<T>              mItems;
    private             int                  mSelection;
    private             SelectionListener<T> mSelectionListener;

    public interface SelectionListener<T> {
        void popupMenuItemSelected(PopupMenu<T> popup);
    }

    public PopupMenu() {
        super(null, false);
        setThemeFont(ThemeFont.BUTTON);
        addMouseListener(this);
        addFocusListener(this);
        addKeyListener(this);
        setFocusable(true);
    }

    public PopupMenu(T[] items, SelectionListener<T> selectionListener) {
        this(Arrays.asList(items), selectionListener);
    }

    public PopupMenu(List<T> items, SelectionListener<T> selectionListener) {
        this();
        mItems = new ArrayList<>(items);
        mSelectionListener = selectionListener;
    }

    public SelectionListener<T> getSelectionListener() {
        return mSelectionListener;
    }

    public void setSelectionListener(SelectionListener<T> listener) {
        mSelectionListener = listener;
    }

    public final int getSelectedIndex() {
        return mSelection;
    }

    public final T getSelectedItem() {
        return mSelection < mItems.size() ? mItems.get(mSelection) : null;
    }

    public void setSelectedIndex(int index, boolean notifyListeners) {
        if (index != mSelection && index >= 0 && index < mItems.size() && mItems.get(index) != null) {
            mSelection = index;
            repaint();
            if (notifyListeners && mSelectionListener != null) {
                mSelectionListener.popupMenuItemSelected(this);
            }
        }
    }

    public final void setSelectedItem(T item, boolean notifyListeners) {
        setSelectedIndex(mItems.indexOf(item), notifyListeners);
    }

    public final int itemCount() {
        return mItems.size();
    }

    public final void addSeparator() {
        addItem(null);
    }

    public final void addSeparatorAt(int index) {
        addItemAt(index, null);
    }

    public void addItem(T item) {
        mItems.add(item);
    }

    public void addItemAt(int index, T item) {
        if (index <= mSelection && mSelection < mItems.size()) {
            mSelection++;
        }
        mItems.add(index, item);
    }

    public final void removeItem(T item) {
        int index = mItems.indexOf(item);
        if (index != -1) {
            removeItemAt(index);
        }
    }

    public void removeItemAt(int index) {
        mItems.remove(index);
        if (index <= mSelection && mSelection > 0) {
            mSelection--;
        }
    }

    public void clear() {
        mItems.clear();
        mSelection = 0;
    }

    @Override
    public Dimension getMinimumSize() {
        if (isMinimumSizeSet()) {
            return super.getMinimumSize();
        }
        Insets    insets   = getInsets();
        Scale     scale    = Scale.get(this);
        Font      font     = scale.scale(getFont());
        Dimension size     = new Dimension();
        Dimension textSize = TextDrawing.getPreferredSize(font, "Mg");
        if (size.width < textSize.width) {
            size.width = textSize.width;
        }
        if (size.height < textSize.height) {
            size.height = textSize.height;
        }
        Font      faFont = new Font(ThemeFont.FONT_AWESOME_SOLID, Font.PLAIN, font.getSize());
        Dimension faSize = TextDrawing.getPreferredSize(faFont, POPUP_MARK);
        size.width += faSize.width;
        if (size.height < faSize.height) {
            size.height = faSize.height;
        }
        size.width += insets.left + insets.right + scale.scale(Button.H_MARGIN) +
                scale.scale(Button.H_MARGIN) + scale.scale(GAP);
        size.height += insets.top + insets.bottom + scale.scale(Button.V_MARGIN) +
                scale.scale(Button.V_MARGIN);
        return size;
    }

    @Override
    public Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        Insets    insets = getInsets();
        Scale     scale  = Scale.get(this);
        Font      font   = scale.scale(getFont());
        Dimension size   = new Dimension();
        for (T item : mItems) {
            if (item != null) {
                Dimension textSize = TextDrawing.getPreferredSize(font, item.toString());
                if (size.width < textSize.width) {
                    size.width = textSize.width;
                }
                if (size.height < textSize.height) {
                    size.height = textSize.height;
                }
            }
        }
        Font      faFont = new Font(ThemeFont.FONT_AWESOME_SOLID, Font.PLAIN, font.getSize());
        Dimension faSize = TextDrawing.getPreferredSize(faFont, POPUP_MARK);
        size.width += faSize.width;
        if (size.height < faSize.height) {
            size.height = faSize.height;
        }
        size.width += insets.left + insets.right + scale.scale(Button.H_MARGIN) +
                scale.scale(Button.H_MARGIN) + scale.scale(GAP);
        size.height += insets.top + insets.bottom + scale.scale(Button.V_MARGIN) +
                scale.scale(Button.V_MARGIN);
        return size;
    }

    @Override
    public Dimension getMaximumSize() {
        if (isMaximumSizeSet()) {
            return super.getMinimumSize();
        }
        Dimension size = getPreferredSize();
        size.width = 10000;
        return size;
    }

    @Override
    protected void paintComponent(Graphics g) {
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        Color     color;
        Color     onColor;
        if (isEnabled()) {
            color = ThemeColor.BUTTON;
            onColor = ThemeColor.ON_BUTTON;
        } else {
            color = ThemeColor.BUTTON;
            onColor = ThemeColor.ON_DISABLED_BUTTON;
        }

        Path2D.Double path         = new Path2D.Double();
        double        corner       = bounds.height / 3.0;
        double        top          = bounds.y;
        double        topCorner    = bounds.y + corner;
        double        left         = bounds.x;
        double        leftCorner   = bounds.x + corner;
        double        bottom       = bounds.y + bounds.height - 1;
        double        bottomCorner = bottom - corner;
        double        right        = bounds.x + bounds.width - 1;
        double        rightCorner  = right - corner;
        path.moveTo(leftCorner, top);
        path.lineTo(rightCorner, top);
        path.curveTo(rightCorner, top, right, top, right, topCorner);
        path.lineTo(right, bottomCorner);
        path.curveTo(right, bottomCorner, right, bottom, rightCorner, bottom);
        path.lineTo(leftCorner, bottom);
        path.curveTo(leftCorner, bottom, left, bottom, left, bottomCorner);
        path.lineTo(left, topCorner);
        path.curveTo(left, topCorner, left, top, leftCorner, top);
        path.closePath();

        Graphics2D gc = GraphicsUtilities.prepare(g);
        gc.setColor(color);
        gc.fill(path);

        Scale     scale  = Scale.get(this);
        Font      font   = scale.scale(getFont());
        Font      faFont = new Font(ThemeFont.FONT_AWESOME_SOLID, Font.PLAIN, font.getSize());
        Dimension faSize = TextDrawing.getPreferredSize(faFont, POPUP_MARK);
        T         item   = getSelectedItem();
        if (item != null) {
            if (isFocusOwner()) {
                Map<TextAttribute, Integer> attributes = new HashMap<>();
                attributes.put(TextAttribute.UNDERLINE, TextAttribute.UNDERLINE_LOW_ONE_PIXEL);
                font = font.deriveFont(attributes);
            }
            gc.setFont(font);
            gc.setColor(onColor);
            Rectangle textBounds = new Rectangle(bounds.x + scale.scale(Button.H_MARGIN),
                    bounds.y + scale.scale(Button.V_MARGIN),
                    bounds.width - (2 * scale.scale(Button.H_MARGIN) + scale.scale(GAP) + faSize.width),
                    bounds.height - 2 * scale.scale(Button.V_MARGIN));
            TextDrawing.draw(gc, textBounds, TextDrawing.truncateIfNecessary(font, item.toString(),
                    textBounds.width, SwingConstants.CENTER), SwingConstants.LEFT, SwingConstants.CENTER);
        }
        gc.setFont(faFont);
        gc.setColor(onColor);
        bounds.width -= scale.scale(Button.H_MARGIN);
        TextDrawing.draw(gc, bounds, POPUP_MARK, SwingConstants.RIGHT, SwingConstants.CENTER);

        gc.setColor(ThemeColor.BUTTON_BORDER);
        RenderingHints saved = GraphicsUtilities.setMaximumQualityForGraphics(gc);
        gc.draw(path);
        gc.setRenderingHints(saved);
    }

    @Override
    public void focusGained(FocusEvent event) {
        repaint();
    }

    @Override
    public void focusLost(FocusEvent event) {
        repaint();
    }

    @Override
    public void keyTyped(KeyEvent event) {
        if (isEnabled() && event.getKeyChar() == ' ' && event.getModifiersEx() == 0) {
            click();
        }
    }

    @Override
    public void keyPressed(KeyEvent event) {
        // Unused
    }

    @Override
    public void keyReleased(KeyEvent event) {
        // Unused
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        // Unused
    }

    @Override
    public void mousePressed(MouseEvent event) {
        if (isEnabled() && !event.isPopupTrigger() && event.getButton() == 1) {
            click();
        }
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseExited(MouseEvent event) {
        // Unused
    }

    public void click() {
        Menu menu = new Menu();
        int  i    = 0;
        for (T item : mItems) {
            if (item == null) {
                menu.addSeparator();
            } else {
                int index = i;
                menu.addItem(new MenuItem(item.toString(), (mi) -> setSelectedIndex(index, true)));
            }
            i++;
        }
        menu.presentToUser(this, mSelection);
    }
}
