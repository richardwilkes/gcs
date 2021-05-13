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

package com.trollworks.gcs.body;

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.WidgetHelpers;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.ID;
import com.trollworks.gcs.utility.text.DiceFormatter;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Rectangle;
import java.beans.PropertyChangeListener;
import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

public class HitLocationTablePanel extends BandedPanel {
    private HitLocationTable mLocations;
    private Runnable         mAdjustCallback;
    private EditorField      mIDField;

    public HitLocationTablePanel(HitLocationTable locations, Runnable adjustCallback) {
        super("hit-locations");
        setLayout(new PrecisionLayout());
        mLocations = locations;
        mAdjustCallback = adjustCallback;
        fill();
    }

    private EditorField addField(Container container, String title, String tooltip, Object value, Object protoValue, JFormattedTextField.AbstractFormatterFactory formatter, PropertyChangeListener listener) {
        container.add(WidgetHelpers.createLabel(title, tooltip), new PrecisionLayoutData().setFillHorizontalAlignment());
        EditorField         field      = new EditorField(formatter, listener, SwingConstants.LEFT, value, protoValue, tooltip);
        PrecisionLayoutData layoutData = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (protoValue == null) {
            layoutData.setGrabHorizontalSpace(true);
        }
        container.add(field, layoutData);
        return field;
    }

    public HitLocationTable getHitLocations() {
        return mLocations;
    }

    public Runnable getAdjustCallback() {
        return mAdjustCallback;
    }

    @Override
    public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
        return 16;
    }

    @Override
    public Dimension getPreferredScrollableViewportSize() {
        return new Dimension(32, 32); // This needs to be small to allow the scroll pane to work
    }

    @Override
    public boolean getScrollableTracksViewportWidth() {
        return true;
    }

    public void reset(HitLocationTable locations) {
        removeAll();
        mLocations.resetTo(locations);
        fill();
    }

    private void fill() {
        Wrapper wrapper = new Wrapper(new PrecisionLayout().setColumns(7).setMargins(0));
        wrapper.add(new FontAwesomeButton("\uf055", HitLocationEditor.BUTTON_SIZE, I18n.Text("Add Hit Location"), this::addHitLocation));
        mIDField = addField(wrapper,
                I18n.Text("ID"),
                I18n.Text("An ID for the hit location table"),
                mLocations.getID(),
                Text.makeFiller(8, 'm'),
                FieldFactory.STRING,
                (evt) -> {
                    String existingID = mLocations.getID();
                    String id         = ((String) evt.getNewValue());
                    if (!existingID.equals(id)) {
                        id = ID.sanitize(id, null, false);
                        if (id.isEmpty()) {
                            mIDField.setValue(existingID);
                        } else {
                            mLocations.setID(id);
                            mIDField.setValue(id);
                            mAdjustCallback.run();
                        }
                    }
                });
        addField(wrapper,
                I18n.Text("Name"),
                I18n.Text("The name of this hit location table"),
                mLocations.getName(),
                null,
                FieldFactory.STRING,
                (evt) -> {
                    mLocations.setName((String) evt.getNewValue());
                    mAdjustCallback.run();
                });
        addField(wrapper,
                I18n.Text("Roll"),
                I18n.Text("The dice to roll on the table"),
                mLocations.getRoll(),
                new Dice(100, 100, 100),
                new DefaultFormatterFactory(new DiceFormatter(null)),
                (evt) -> {
                    mLocations.setRoll((Dice) evt.getNewValue());
                    mAdjustCallback.run();
                });
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        for (HitLocation location : mLocations.getLocations()) {
            add(new HitLocationPanel(location, mAdjustCallback), new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        }
        adjustButtons();
    }

    public void adjustButtons() {
        Component[] children = getComponents();
        int         count    = children.length;
        for (int i = 1; i < count; i++) {
            ((HitLocationPanel) children[i]).adjustButtons(i == 1, i == count - 1);
        }
    }

    public void adjustForReordering() {
        Component[] children = getComponents();
        int         count    = children.length;
        for (int i = 1; i < count; i++) {
            ((HitLocationPanel) children[i]).adjustButtons(i == 1, i == count - 1);
        }
        mLocations.update();
        repaint();
    }

    public void addHitLocation() {
        HitLocation location = new HitLocation("id", I18n.Text("choice name"), I18n.Text("table name"), 0, 0, 0, I18n.Text("description"));
        mLocations.addLocation(location);
        mLocations.update();
        mAdjustCallback.run();
        add(new HitLocationPanel(location, mAdjustCallback), new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
        scrollRectToVisible(new Rectangle(0, getPreferredSize().height - 1, 1, 1));
        ((HitLocationPanel) getComponent(getComponentCount() - 1)).focusIDField();
        adjustButtons();
    }

    public void focusIDField() {
        mIDField.requestFocusInWindow();
    }
}
