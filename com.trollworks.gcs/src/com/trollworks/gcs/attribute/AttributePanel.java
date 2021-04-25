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
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Container;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.beans.PropertyChangeListener;
import javax.swing.JCheckBox;
import javax.swing.JFormattedTextField.AbstractFormatterFactory;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.SwingConstants;

public class AttributePanel extends JPanel {
    private AttributeDef mAttrDef;

    public AttributePanel(AttributeDef attrDef, Runnable adjustCallback) {
        super(new PrecisionLayout().setColumns(3));
        setOpaque(false);
        mAttrDef = attrDef;

        JPanel left = new JPanel(new PrecisionLayout());
        left.setOpaque(false);
        add(left, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        FontAwesomeButton moveUp = new FontAwesomeButton("\uf35b", 16, I18n.Text("Move Up"), () -> System.out.println("move up"));
        left.add(moveUp);
        FontAwesomeButton moveDown = new FontAwesomeButton("\uf358", 16, I18n.Text("Move Down"), () -> System.out.println("move down"));
        left.add(moveDown);

        JPanel center = new JPanel(new PrecisionLayout());
        center.setOpaque(false);
        add(center, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));

        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(6).setMargins(0));
        wrapper.setOpaque(false);
        addField(wrapper,
                I18n.Text("ID"),
                I18n.Text("A unique ID for the attribute"),
                attrDef.getID(),
                Text.makeFiller(7, 'm'),
                FieldFactory.STRING,
                (evt) -> {
                    mAttrDef.setID((String) evt.getNewValue());
                    adjustCallback.run();
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
                I18n.Text("Description"),
                I18n.Text("The full name of this attribute"),
                attrDef.getDescription(),
                null,
                FieldFactory.STRING,
                (evt) -> {
                    mAttrDef.setDescription((String) evt.getNewValue());
                    adjustCallback.run();
                });
        center.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        wrapper = new JPanel(new PrecisionLayout().setColumns(7).setMargins(0));
        wrapper.setOpaque(false);
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
        addCheckBox(wrapper,
                I18n.Text("Decimal"),
                I18n.Text("Checked if this field allows fractional values"),
                attrDef.isDecimal(),
                (evt) -> {
                    mAttrDef.setDecimal(evt.getStateChange() == ItemEvent.SELECTED);
                    adjustCallback.run();
                });
        center.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));


        JPanel right = new JPanel(new PrecisionLayout());
        right.setOpaque(false);
        add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        FontAwesomeButton remove = new FontAwesomeButton("\uf1f8", 16, I18n.Text("Remove"), () -> System.out.println("remove"));
        right.add(remove);
    }

    private EditorField addField(Container container, String title, String tooltip, Object value, Object protoValue, AbstractFormatterFactory formatter, PropertyChangeListener listener) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        tooltip = Text.wrapPlainTextForToolTip(tooltip);
        label.setToolTipText(tooltip);
        container.add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
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
        container.add(checkbox, new PrecisionLayoutData());
        return checkbox;
    }
}
