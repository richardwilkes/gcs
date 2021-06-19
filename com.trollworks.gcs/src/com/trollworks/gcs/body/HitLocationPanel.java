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
import com.trollworks.gcs.ui.widget.ContentPanel;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.ID;

import java.awt.Container;
import java.awt.Rectangle;
import java.util.List;
import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

public class HitLocationPanel extends ContentPanel implements DocumentListener {
    private HitLocation        mLocation;
    private Runnable           mAdjustCallback;
    private EditorField        mIDField;
    private MultiLineTextField mDescriptionField;
    private FontAwesomeButton  mMoveUpButton;
    private FontAwesomeButton  mMoveDownButton;
    private FontAwesomeButton  mAddSubTableButton;
    private ContentPanel       mCenter;

    public HitLocationPanel(HitLocation location, Runnable adjustCallback) {
        super(new PrecisionLayout().setColumns(3).setMargins(0), false);
        mLocation = location;
        mAdjustCallback = adjustCallback;

        ContentPanel left = new ContentPanel(new PrecisionLayout(), false);
        add(left, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        mMoveUpButton = new FontAwesomeButton("\uf35b", I18n.text("Move Up"), () -> {
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
        mMoveDownButton = new FontAwesomeButton("\uf358", I18n.text("Move Down"), () -> {
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
        mAddSubTableButton = new FontAwesomeButton("\uf055", I18n.text("Add Hit Location Sub-Table"), this::addSubHitLocations);
        left.add(mAddSubTableButton);

        mCenter = new ContentPanel(new PrecisionLayout(), false);
        add(mCenter, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));

        ContentPanel wrapper = new ContentPanel(new PrecisionLayout().setColumns(8).setMargins(0), false);
        mIDField = addField(wrapper,
                I18n.text("ID"),
                I18n.text("An ID for the hit location"),
                mLocation.getID(),
                null,
                FieldFactory.STRING,
                (f) -> {
                    String existingID = mLocation.getID();
                    String id         = ((String) f.getValue());
                    if (!existingID.equals(id)) {
                        id = ID.sanitize(id, null, false);
                        if (id.isEmpty()) {
                            f.setValue(existingID);
                        } else {
                            mLocation.setID(id);
                            f.setValue(id);
                            mAdjustCallback.run();
                        }
                    }
                });
        addField(wrapper,
                I18n.text("Slots"),
                I18n.text("The number of consecutive numbers this hit location fills in the table"),
                Integer.valueOf(mLocation.getSlots()),
                Integer.valueOf(999),
                FieldFactory.POSINT6,
                (f) -> {
                    mLocation.setSlots(((Integer) f.getValue()).intValue());
                    mAdjustCallback.run();
                });
        addField(wrapper,
                I18n.text("Hit Penalty"),
                I18n.text("The skill adjustment to hit this location"),
                Integer.valueOf(mLocation.getHitPenalty()),
                Integer.valueOf(-999),
                FieldFactory.INT6,
                (f) -> {
                    mLocation.setHitPenalty(((Integer) f.getValue()).intValue());
                    mAdjustCallback.run();
                });
        addField(wrapper,
                I18n.text("DR Bonus"),
                I18n.text("The amount of DR this hit location grants due to natural toughness"),
                Integer.valueOf(mLocation.getDRBonus()),
                Integer.valueOf(999),
                FieldFactory.POSINT6,
                (f) -> {
                    mLocation.setDRBonus(((Integer) f.getValue()).intValue());
                    mAdjustCallback.run();
                });
        mCenter.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        wrapper = new ContentPanel(new PrecisionLayout().setColumns(4).setMargins(0), false);
        addField(wrapper,
                I18n.text("Choice Name"),
                I18n.text("The name of this hit location as it should appear in choice lists"),
                mLocation.getChoiceName(),
                null,
                FieldFactory.STRING,
                (f) -> {
                    mLocation.setChoiceName((String) f.getValue());
                    mAdjustCallback.run();
                });
        addField(wrapper,
                I18n.text("Table Name"),
                I18n.text("The name of this hit location as it should appear in the hit location table"),
                mLocation.getTableName(),
                null,
                FieldFactory.STRING,
                (f) -> {
                    mLocation.setTableName((String) f.getValue());
                    mAdjustCallback.run();
                });
        mCenter.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        wrapper = new ContentPanel(new PrecisionLayout().setColumns(2).setMargins(0), false);
        mDescriptionField = addTextArea(wrapper,
                I18n.text("Description"),
                I18n.text("An description of any special effects for hits to this location"),
                mLocation.getDescription());
        mCenter.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        if (mLocation.getSubTable() != null) {
            HitLocationTablePanel subTable = new HitLocationTablePanel(mLocation.getSubTable(), mAdjustCallback);
            mCenter.add(subTable, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));
        }

        ContentPanel right = new ContentPanel(new PrecisionLayout(), false);
        add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        FontAwesomeButton remove = new FontAwesomeButton("\uf1f8", I18n.text("Remove"), () -> {
            getParent().remove(this);
            mLocation.getOwningTable().removeLocation(mLocation);
            mAdjustCallback.run();
        });
        right.add(remove);
    }

    private static EditorField addField(Container container, String title, String tooltip, Object value, Object protoValue, JFormattedTextField.AbstractFormatterFactory formatter, EditorField.ChangeListener listener) {
        EditorField         field      = new EditorField(formatter, listener, SwingConstants.LEFT, value, protoValue, tooltip);
        PrecisionLayoutData layoutData = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (protoValue == null) {
            layoutData.setGrabHorizontalSpace(true);
        }
        container.add(new Label(title, field), new PrecisionLayoutData().setFillHorizontalAlignment());
        container.add(field, layoutData);
        return field;
    }

    private MultiLineTextField addTextArea(Container container, String title, String tooltip, String text) {
        MultiLineTextField area = new MultiLineTextField(text, tooltip, this);
        container.add(new Label(title, area), new PrecisionLayoutData().setFillHorizontalAlignment().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
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
        HitLocationTable table = new HitLocationTable("id", I18n.text("name"), new Dice(3));
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
