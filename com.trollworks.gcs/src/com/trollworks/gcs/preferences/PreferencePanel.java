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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.Alignment;
import com.trollworks.gcs.ui.layout.FlexComponent;
import com.trollworks.gcs.ui.layout.FlexContainer;
import com.trollworks.gcs.ui.layout.LayoutSize;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Dimension;
import java.awt.Insets;
import java.awt.event.ItemListener;
import javax.swing.Icon;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JSeparator;
import javax.swing.SwingConstants;

/** The abstract base class for all preference panels. */
public abstract class PreferencePanel extends JPanel {
    private String            mTitle;
    private PreferencesWindow mOwner;

    /**
     * Creates a new preference panel.
     *
     * @param title The title for this panel.
     * @param owner The owning {@link PreferencesWindow}.
     */
    public PreferencePanel(String title, PreferencesWindow owner) {
        setBorder(new EmptyBorder(5));
        setOpaque(false);
        mTitle = title;
        mOwner = owner;
    }

    /** @return The owner. */
    public PreferencesWindow getOwner() {
        return mOwner;
    }

    /** Resets this panel back to its defaults. */
    public abstract void reset();

    /** @return Whether the panel is currently set to defaults or not. */
    public abstract boolean isSetToDefaults();

    /** Call to adjust the reset button for any changes that have been made. */
    protected void adjustResetButton() {
        mOwner.adjustResetButton();
    }

    @Override
    public String toString() {
        return mTitle;
    }

    /**
     * Creates a right-aligned {@link JLabel} suitable for use within the preference panel.
     *
     * @param title   The title to use.
     * @param tooltip The tooltip to use.
     * @return The newly created {@link JLabel}.
     */
    protected JLabel createLabel(String title, String tooltip) {
        return createLabel(title, tooltip, null, SwingConstants.RIGHT);
    }

    /**
     * Creates a {@link JLabel} suitable for use within the preference panel.
     *
     * @param title     The title to use.
     * @param tooltip   The tooltip to use.
     * @param alignment The alignment to use.
     * @return The newly created {@link JLabel}.
     */
    protected JLabel createLabel(String title, String tooltip, int alignment) {
        return createLabel(title, tooltip, null, alignment);
    }

    /**
     * Creates a right-aligned {@link JLabel} suitable for use within the preference panel.
     *
     * @param title   The title to use.
     * @param tooltip The tooltip to use.
     * @param icon    The {@link Icon} to use.
     * @return The newly created {@link JLabel}.
     */
    protected JLabel createLabel(String title, String tooltip, Icon icon) {
        return createLabel(title, tooltip, icon, SwingConstants.RIGHT);
    }

    /**
     * Creates a {@link JLabel} suitable for use within the preference panel.
     *
     * @param title     The title to use.
     * @param tooltip   The tooltip to use.
     * @param icon      The {@link Icon} to use.
     * @param alignment The alignment to use.
     * @return The newly created {@link JLabel}.
     */
    protected JLabel createLabel(String title, String tooltip, Icon icon, int alignment) {
        JLabel label = new JLabel(title, alignment);
        label.setOpaque(false);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        label.setIcon(icon);
        UIUtilities.setToPreferredSizeOnly(label);
        add(label);
        return label;
    }

    /** @param flexContainer The {@link FlexContainer} to add a separator to. */
    protected void addSeparator(FlexContainer flexContainer) {
        JSeparator sep = new JSeparator();
        sep.setOpaque(false);
        sep.setMaximumSize(new Dimension(LayoutSize.MAXIMUM_SIZE, 1));
        add(sep);
        FlexComponent comp = new FlexComponent(sep, Alignment.CENTER, Alignment.CENTER);
        comp.setInsets(new Insets(5, 0, 5, 0));
        flexContainer.add(comp);
    }

    /**
     * Creates a {@link JCheckBox} suitable for use within the preference panel.
     *
     * @param title   The title to use.
     * @param tooltip The tooltip to use.
     * @param checked Whether the initial state should be checked.
     * @return The newly created {@link JCheckBox}.
     */
    protected JCheckBox createCheckBox(String title, String tooltip, boolean checked) {
        JCheckBox checkbox = new JCheckBox(title, checked);
        checkbox.setOpaque(false);
        checkbox.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        if (this instanceof ItemListener) {
            checkbox.addItemListener((ItemListener) this);
        }
        add(checkbox);
        return checkbox;
    }

    /**
     * Sets up a {@link JComboBox} suitable for use within the preference panel.
     *
     * @param combo   The {@link JComboBox} to prepare.
     * @param tooltip The tooltip to use.
     */
    protected void setupCombo(JComboBox<?> combo, String tooltip) {
        combo.setOpaque(false);
        combo.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        add(combo);
    }
}
