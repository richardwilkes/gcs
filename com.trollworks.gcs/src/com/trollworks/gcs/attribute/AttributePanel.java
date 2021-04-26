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

package com.trollworks.gcs.attribute;

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Container;
import java.awt.event.ItemListener;
import java.beans.PropertyChangeListener;
import java.util.Map;
import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JFormattedTextField.AbstractFormatterFactory;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.SwingConstants;

public class AttributePanel extends JPanel {
    private   Map<String, AttributeDef> mAttributes;
    protected AttributeDef              mAttrDef;
    private   EditorField               mIDField;
    private   FontAwesomeButton         mMoveUpButton;
    private   FontAwesomeButton         mMoveDownButton;
    private   FontAwesomeButton         mAddThresholdButton;

    public AttributePanel(Map<String, AttributeDef> attributes, AttributeDef attrDef, Runnable adjustCallback) {
        super(new PrecisionLayout().setColumns(3));
        setOpaque(false);
        mAttributes = attributes;
        mAttrDef = attrDef;

        JPanel left = new JPanel(new PrecisionLayout());
        left.setOpaque(false);
        add(left, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        mMoveUpButton = new FontAwesomeButton("\uf35b", 16, I18n.Text("Move Up"), () -> {
            AttributeListPanel parent = (AttributeListPanel) getParent();
            int                index  = UIUtilities.getIndexOf(parent, this);
            if (index > 0) {
                parent.remove(index);
                parent.add(this, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment(), index - 1);
                parent.renumber();
                adjustCallback.run();
            }
        });
        left.add(mMoveUpButton);
        mMoveDownButton = new FontAwesomeButton("\uf358", 16, I18n.Text("Move Down"), () -> {
            AttributeListPanel parent = (AttributeListPanel) getParent();
            int                index  = UIUtilities.getIndexOf(parent, this);
            if (index != -1 && index < parent.getComponentCount() - 1) {
                parent.remove(index);
                parent.add(this, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment(), index + 1);
                parent.renumber();
                adjustCallback.run();
            }
        });
        left.add(mMoveDownButton);
        mAddThresholdButton = new FontAwesomeButton("\uf055", 16, I18n.Text("Add Threshold"), () -> {
            // TODO: Implement
        });
        left.add(mAddThresholdButton);

        JPanel center = new JPanel(new PrecisionLayout());
        center.setOpaque(false);
        add(center, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));

        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(6).setMargins(0));
        wrapper.setOpaque(false);
        mIDField = addField(wrapper,
                I18n.Text("ID"),
                I18n.Text("A unique ID for the attribute"),
                attrDef.getID(),
                Text.makeFiller(7, 'm'),
                FieldFactory.STRING,
                (evt) -> {
                    String existingID = mAttrDef.getID();
                    String id         = (String) evt.getNewValue();
                    if (!existingID.equals(id)) {
                        if (mAttributes.containsKey(id)) {
                            mIDField.setText(existingID);
                        } else {
                            mAttributes.remove(existingID);
                            mAttrDef.setID(id);
                            mAttributes.put(id, mAttrDef);
                            adjustCallback.run();
                        }
                    }
                });
        addField(wrapper,
                I18n.Text("Name"),
                I18n.Text("The name of this attribute, often an abbreviation"),
                attrDef.getName(),
                Text.makeFiller(8, 'm'),
                FieldFactory.STRING,
                (evt) -> {
                    mAttrDef.setName((String) evt.getNewValue());
                    adjustCallback.run();
                });
        addField(wrapper,
                I18n.Text("Full Name"),
                I18n.Text("The full name of this attribute (may be omitted)"),
                attrDef.getFullName(),
                null,
                FieldFactory.STRING,
                (evt) -> {
                    mAttrDef.setFullName((String) evt.getNewValue());
                    adjustCallback.run();
                });
        center.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        wrapper = new JPanel(new PrecisionLayout().setColumns(7).setMargins(0));
        wrapper.setOpaque(false);
        addAttributeTypeCombo(wrapper,
                attrDef.getType(),
                (evt) -> {
                    mAttrDef.setType((AttributeType) evt.getItem());
                    adjustCallback.run();
                });
        addField(wrapper,
                I18n.Text("Base"),
                I18n.Text("The base value, which may be a number or a formula"),
                attrDef.getAttributeBase(),
                null,
                FieldFactory.STRING,
                (evt) -> {
                    mAttrDef.setAttributeBase((String) evt.getNewValue());
                    adjustCallback.run();
                });
        addField(wrapper,
                I18n.Text("Cost"),
                I18n.Text("The cost per point difference from the base"),
                Integer.valueOf(attrDef.getCostPerPoint()),
                Integer.valueOf(999999),
                FieldFactory.POSINT6,
                (evt) -> {
                    mAttrDef.setCostPerPoint(((Integer) evt.getNewValue()).intValue());
                    adjustCallback.run();
                });
        addField(wrapper,
                I18n.Text("SM Reduction"),
                I18n.Text("The reduction in cost (as a percentage) for each SM greater than 0"),
                Integer.valueOf(attrDef.getCostAdjPercentPerSM()),
                Integer.valueOf(80),
                FieldFactory.PERCENT_REDUCTION,
                (evt) -> {
                    mAttrDef.setCostAdjPercentPerSM(((Integer) evt.getNewValue()).intValue());
                    adjustCallback.run();
                });
        center.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));


        JPanel right = new JPanel(new PrecisionLayout());
        right.setOpaque(false);
        add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        FontAwesomeButton remove = new FontAwesomeButton("\uf1f8", 16, I18n.Text("Remove"), () -> {
            getParent().remove(this);
            mAttributes.remove(mAttrDef.getID());
            adjustCallback.run();
        });
        right.add(remove);
    }

    private void addLabel(Container container, String title, String tooltip) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        container.add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private EditorField addField(Container container, String title, String tooltip, Object value, Object protoValue, AbstractFormatterFactory formatter, PropertyChangeListener listener) {
        addLabel(container, title, tooltip);
        EditorField         field      = new EditorField(formatter, listener, SwingConstants.LEFT, value, protoValue, tooltip);
        PrecisionLayoutData layoutData = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (protoValue == null) {
            layoutData.setGrabHorizontalSpace(true);
        }
        container.add(field, layoutData);
        return field;
    }

    private JCheckBox addCheckBox(Container container, String title, String tooltip, boolean selected, ItemListener listener) {
        JCheckBox checkbox = new JCheckBox(title);
        checkbox.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        checkbox.setSelected(selected);
        checkbox.addItemListener(listener);
        container.add(checkbox);
        return checkbox;
    }

    private JComboBox<AttributeType> addAttributeTypeCombo(Container container, AttributeType value, ItemListener listener) {
        JComboBox<AttributeType> combo = new JComboBox<>(AttributeType.values());
        combo.setSelectedItem(value);
        combo.addItemListener(listener);
        container.add(combo);
        return combo;
    }

    public void focusIDField() {
        mIDField.requestFocusInWindow();
    }

    public void adjustButtons(boolean isFirst, boolean isLast) {
        mMoveUpButton.setEnabled(!isFirst);
        mMoveDownButton.setEnabled(!isLast);
    }
}
