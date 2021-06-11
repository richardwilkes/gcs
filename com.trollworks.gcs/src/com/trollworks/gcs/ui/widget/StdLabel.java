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

import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.Rectangle;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.Objects;
import javax.swing.Icon;
import javax.swing.JComponent;
import javax.swing.SwingConstants;

/** A simple label replacement that is scalable. */
public class StdLabel extends JComponent implements PropertyChangeListener {
    private static final String ERROR_KEY = "error";
    private static final int    GAP       = 4;

    private RetinaIcon mIcon;
    private String     mText;
    private ThemeFont  mThemeFont;
    private int        mHAlign;
    private JComponent mRefersTo;
    private int        mTruncationPolicy;

    /**
     * Create a new label.
     *
     * @param text The text to use.
     */
    public StdLabel(String text) {
        mText = "";
        mHAlign = SwingConstants.LEFT;
        mTruncationPolicy = SwingConstants.CENTER;
        setThemeFont(ThemeFont.LABEL_PRIMARY);
        setForeground(ThemeColor.ON_BACKGROUND);
        setOpaque(false);
        setText(text);
    }

    /**
     * Create a new label.
     *
     * @param text   The text to use.
     * @param hAlign The horizontal alignment to use.
     */
    public StdLabel(String text, int hAlign) {
        this(text);
        mHAlign = hAlign;
    }

    /**
     * Create a new label.
     *
     * @param text     The text to use.
     * @param refersTo The {@link JComponent} to pair with.
     */
    public StdLabel(String text, JComponent refersTo) {
        this(text);
        setRefersTo(refersTo);
    }

    /**
     * Create a new label.
     *
     * @param icon The icon to use.
     * @param text The text to use.
     */
    public StdLabel(RetinaIcon icon, String text) {
        this(text);
        mIcon = icon;
    }

    public final void setThemeFont(ThemeFont font) {
        mThemeFont = font;
    }

    @Override
    public final Font getFont() {
        return mThemeFont.getFont();
    }

    @Override
    public final void setFont(Font font) {
        System.out.println("ERROR: tried to set font rather than theme font");
        new Exception().printStackTrace(System.out);
    }

    /** @return The icon, or {@code null} if there is none. */
    public Icon getIcon() {
        return mIcon;
    }

    /** @param icon The {@link RetinaIcon} to use, or {@code null}. */
    public void setIcon(RetinaIcon icon) {
        if (mIcon != icon) {
            mIcon = icon;
            invalidate();
        }
    }

    /** @return The text. */
    public String getText() {
        return mText;
    }

    /** @param text The text to use. */
    public void setText(String text) {
        if (text == null) {
            text = "";
        }
        if (!mText.equals(text)) {
            mText = text;
            invalidate();
        }
    }

    /** @return The horizontal alignment. */
    public int getHorizontalAlignment() {
        return mHAlign;
    }

    /** @param alignment The horizontal alignment to use. */
    public void setHorizontalAlignment(int alignment) {
        if (mHAlign != alignment) {
            mHAlign = alignment;
            repaint();
        }
    }

    public int getTruncationPolicy() {
        return mTruncationPolicy;
    }

    public void setTruncationPolicy(int truncationPolicy) {
        mTruncationPolicy = truncationPolicy;
    }

    @Override
    protected void paintComponent(Graphics g) {
        super.paintComponent(g);
        if (mRefersTo != null && mRefersTo.getClientProperty(ERROR_KEY) != null) {
            g.setColor(Color.RED); // TODO: Use themed error color
        }
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        Scale     scale  = Scale.get(this);
        Font      font   = scale.scale(getFont());
        g.setFont(font);
        if (mIcon != null) {
            mIcon.paintIcon(this, g, bounds.x, bounds.y + (bounds.height - scale.scale(mIcon.getIconHeight())) / 2);
            int amt = scale.scale(mIcon.getIconWidth()) + scale.scale(GAP);
            bounds.x += amt;
            bounds.width -= amt;
        }
        String text = mText;
        if (mTruncationPolicy == SwingConstants.LEFT || mTruncationPolicy == SwingConstants.CENTER ||
                mTruncationPolicy == SwingConstants.RIGHT) {
            text = TextDrawing.truncateIfNecessary(font, text, bounds.width, mTruncationPolicy);
        }
        TextDrawing.draw(g, bounds, text, mHAlign, SwingConstants.CENTER);
    }

    @Override
    public Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        Insets    insets = getInsets();
        Scale     scale  = Scale.get(this);
        Dimension size   = TextDrawing.getPreferredSize(scale.scale(getFont()), mText);
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        if (mIcon != null) {
            size.width += scale.scale(mIcon.getIconWidth()) + scale.scale(GAP);
            int height = scale.scale(mIcon.getIconHeight());
            if (height > size.height) {
                size.height = height;
            }
        }
        return size;
    }

    /** @return The {@link JComponent} that is being paired with. */
    public JComponent getRefersTo() {
        return mRefersTo;
    }

    /** @param refersTo The {@link JComponent} to pair with. */
    public void setRefersTo(JComponent refersTo) {
        if (mRefersTo != refersTo) {
            if (mRefersTo != null) {
                mRefersTo.removePropertyChangeListener(TOOL_TIP_TEXT_KEY, this);
                mRefersTo.removePropertyChangeListener(ERROR_KEY, this);
            }
            mRefersTo = refersTo;
            if (mRefersTo != null) {
                mRefersTo.addPropertyChangeListener(TOOL_TIP_TEXT_KEY, this);
                mRefersTo.addPropertyChangeListener(ERROR_KEY, this);
            }
            adjustToLink();
        }
    }

    @Override
    public void propertyChange(PropertyChangeEvent evt) {
        adjustToLink();
    }

    private void adjustToLink() {
        String tooltip = null;
        if (mRefersTo != null) {
            tooltip = (String) mRefersTo.getClientProperty(ERROR_KEY);
            if (tooltip == null) {
                tooltip = mRefersTo.getToolTipText();
            }
        }
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        repaint();
    }

    /**
     * Sets/clears the error message that a Label will respond to.
     *
     * @param comp The {@link JComponent} to set the message on.
     * @param msg  The error message or {@code null}.
     */
    public static void setErrorMessage(JComponent comp, String msg) {
        if (!Objects.equals(comp.getClientProperty(ERROR_KEY), msg)) {
            comp.putClientProperty(ERROR_KEY, msg);
        }
    }
}
