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
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.WidgetHelpers;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.ID;

import java.awt.Container;
import java.awt.Rectangle;
import java.beans.PropertyChangeListener;
import java.util.List;
import javax.swing.JFormattedTextField;
import javax.swing.JPanel;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

public class HitLocationPanel extends JPanel implements DocumentListener {
    private HitLocation        mLocation;
    private Runnable           mAdjustCallback;
    private EditorField        mIDField;
    private MultiLineTextField mDescriptionField;
    private FontAwesomeButton  mMoveUpButton;
    private FontAwesomeButton  mMoveDownButton;
    private FontAwesomeButton  mAddSubTableButton;
    private JPanel             mCenter;

    public HitLocationPanel(HitLocation location, Runnable adjustCallback) {
        super(new PrecisionLayout().setColumns(3).setMargins(0));
        setOpaque(false);
        mLocation = location;
        mAdjustCallback = adjustCallback;

        JPanel left = new JPanel(new PrecisionLayout());
        left.setOpaque(false);
        add(left, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        mMoveUpButton = new FontAwesomeButton("\uf35b", I18n.Text("Move Up"), () -> {
            HitLocationTablePanel parent = (HitLocationTablePanel) getParent();
            int                   index  = UIUtilities.getIndexOf(parent, this);
            if (index > 0) {
                parent.remove(index);
                parent.add(this, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment(), index - 1);
                List<HitLocation> locations = mLocation.getOwningTable().getLocations();
                index--; // There is a non-item row before the list in the panel, so compensate for it
                locations.remove(index);
                locations.add(index - 1, mLocation);
                parent.adjustForReordering();
                mAdjustCallback.run();
            }
        });
        left.add(mMoveUpButton);
        mMoveDownButton = new FontAwesomeButton("\uf358", I18n.Text("Move Down"), () -> {
            HitLocationTablePanel parent = (HitLocationTablePanel) getParent();
            int                   index  = UIUtilities.getIndexOf(parent, this);
            if (index != -1 && index < parent.getComponentCount() - 1) {
                parent.remove(index);
                parent.add(this, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment(), index + 1);
                List<HitLocation> locations = mLocation.getOwningTable().getLocations();
                index--; // There is a non-item row before the list in the panel, so compensate for it
                locations.remove(index);
                locations.add(index + 1, mLocation);
                parent.adjustForReordering();
                mAdjustCallback.run();
            }
        });
        left.add(mMoveDownButton);
        mAddSubTableButton = new FontAwesomeButton("\uf055", I18n.Text("Add Hit Location Sub-Table"), this::addSubHitLocations);
        left.add(mAddSubTableButton);

        mCenter = new JPanel(new PrecisionLayout());
        mCenter.setOpaque(false);
        add(mCenter, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));

        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(8).setMargins(0));
        wrapper.setOpaque(false);
        mIDField = addField(wrapper,
                I18n.Text("ID"),
                I18n.Text("An ID for the hit location"),
                mLocation.getID(),
                null,
                FieldFactory.STRING,
                (evt) -> {
                    String existingID = mLocation.getID();
                    String id         = ((String) evt.getNewValue());
                    if (!existingID.equals(id)) {
                        id = ID.sanitize(id, null, false);
                        if (id.isEmpty()) {
                            mIDField.setValue(existingID);
                        } else {
                            mLocation.setID(id);
                            mIDField.setValue(id);
                            mAdjustCallback.run();
                        }
                    }
                });
        addField(wrapper,
                I18n.Text("Slots"),
                I18n.Text("The number of consecutive numbers this hit location fills in the table"),
                Integer.valueOf(mLocation.getSlots()),
                Integer.valueOf(999),
                FieldFactory.POSINT6,
                (evt) -> {
                    mLocation.setSlots(((Integer) evt.getNewValue()).intValue());
                    mAdjustCallback.run();
                });
        addField(wrapper,
                I18n.Text("Hit Penalty"),
                I18n.Text("The skill adjustment to hit this location"),
                Integer.valueOf(mLocation.getHitPenalty()),
                Integer.valueOf(-999),
                FieldFactory.INT6,
                (evt) -> {
                    mLocation.setHitPenalty(((Integer) evt.getNewValue()).intValue());
                    mAdjustCallback.run();
                });
        addField(wrapper,
                I18n.Text("DR Bonus"),
                I18n.Text("The amount of DR this hit location grants due to natural toughness"),
                Integer.valueOf(mLocation.getDRBonus()),
                Integer.valueOf(999),
                FieldFactory.POSINT6,
                (evt) -> {
                    mLocation.setDRBonus(((Integer) evt.getNewValue()).intValue());
                    mAdjustCallback.run();
                });
        mCenter.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        wrapper = new JPanel(new PrecisionLayout().setColumns(4).setMargins(0));
        wrapper.setOpaque(false);
        addField(wrapper,
                I18n.Text("Choice Name"),
                I18n.Text("The name of this hit location as it should appear in choice lists"),
                mLocation.getChoiceName(),
                null,
                FieldFactory.STRING,
                (evt) -> {
                    mLocation.setChoiceName((String) evt.getNewValue());
                    mAdjustCallback.run();
                });
        addField(wrapper,
                I18n.Text("Table Name"),
                I18n.Text("The name of this hit location as it should appear in the hit location table"),
                mLocation.getTableName(),
                null,
                FieldFactory.STRING,
                (evt) -> {
                    mLocation.setTableName((String) evt.getNewValue());
                    mAdjustCallback.run();
                });
        mCenter.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        wrapper = new JPanel(new PrecisionLayout().setColumns(2).setMargins(0));
        wrapper.setOpaque(false);
        mDescriptionField = addTextArea(wrapper,
                I18n.Text("Description"),
                I18n.Text("An description of any special effects for hits to this location"),
                mLocation.getDescription());
        mCenter.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        if (mLocation.getSubTable() != null) {
            HitLocationTablePanel subTable = new HitLocationTablePanel(mLocation.getSubTable(), mAdjustCallback);
            mCenter.add(subTable, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));
        }

        JPanel right = new JPanel(new PrecisionLayout());
        right.setOpaque(false);
        add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        FontAwesomeButton remove = new FontAwesomeButton("\uf1f8", I18n.Text("Remove"), () -> {
            getParent().remove(this);
            mLocation.getOwningTable().removeLocation(mLocation);
            mAdjustCallback.run();
        });
        right.add(remove);
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

    private MultiLineTextField addTextArea(Container container, String title, String tooltip, String text) {
        container.add(WidgetHelpers.createLabel(title, tooltip), new PrecisionLayoutData().setFillHorizontalAlignment().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        MultiLineTextField area = new MultiLineTextField(text, tooltip, this);
        container.add(area, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        return area;
    }

    public void focusIDField() {
        mIDField.requestFocusInWindow();
    }

    public void adjustButtons(boolean isFirst, boolean isLast) {
        mMoveUpButton.setEnabled(!isFirst);
        mMoveDownButton.setEnabled(!isLast);
        mAddSubTableButton.setEnabled(mLocation.getSubTable() == null);
    }

    public void addSubHitLocations() {
        HitLocationTable table = new HitLocationTable("id", I18n.Text("name"), new Dice(3));
        mLocation.setSubTable(table);
        HitLocationTablePanel subTable = new HitLocationTablePanel(table, mAdjustCallback);
        mCenter.add(subTable, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));
        mAdjustCallback.run();
        scrollRectToVisible(new Rectangle(0, getPreferredSize().height - 1, 1, 1));
        subTable.focusFirstField();
        subTable.adjustButtons();
    }

    private void adjustDescription() {
        mLocation.setDescription(mDescriptionField.getText());
        mAdjustCallback.run();
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        adjustDescription();
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        adjustDescription();
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        adjustDescription();
    }
}
