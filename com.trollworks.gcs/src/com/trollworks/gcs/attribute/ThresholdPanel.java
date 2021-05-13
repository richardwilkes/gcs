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
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.WidgetHelpers;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Container;
import java.awt.event.ItemEvent;
import java.beans.PropertyChangeListener;
import java.util.List;
import javax.swing.JCheckBox;
import javax.swing.JFormattedTextField;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

public class ThresholdPanel extends JPanel implements DocumentListener {
    private List<PoolThreshold> mThresholds;
    private PoolThreshold       mThreshold;
    private Runnable            mAdjustCallback;
    private FontAwesomeButton   mMoveUpButton;
    private FontAwesomeButton   mMoveDownButton;
    private EditorField         mStateField;
    private EditorField         mDivisorField;
    private MultiLineTextField  mExplanationField;

    public ThresholdPanel(List<PoolThreshold> thresholds, PoolThreshold threshold, Runnable adjustCallback) {
        super(new PrecisionLayout().setColumns(3).setMargins(0));
        setOpaque(false);
        mThresholds = thresholds;
        mThreshold = threshold;
        mAdjustCallback = adjustCallback;

        JPanel left = new JPanel(new PrecisionLayout());
        left.setOpaque(false);
        add(left, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        mMoveUpButton = new FontAwesomeButton("\uf35b", AttributeEditor.BUTTON_SIZE, I18n.Text("Move Up"), () -> {
            ThresholdListPanel parent = (ThresholdListPanel) getParent();
            int                index  = UIUtilities.getIndexOf(parent, this);
            if (index > 0) {
                parent.remove(index);
                mThresholds.remove(index);
                parent.add(this, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment(), index - 1);
                mThresholds.add(index - 1, mThreshold);
                adjustCallback.run();
            }
        });
        left.add(mMoveUpButton);
        mMoveDownButton = new FontAwesomeButton("\uf358", AttributeEditor.BUTTON_SIZE, I18n.Text("Move Down"), () -> {
            ThresholdListPanel parent = (ThresholdListPanel) getParent();
            int                index  = UIUtilities.getIndexOf(parent, this);
            if (index != -1 && index < parent.getComponentCount() - 1) {
                parent.remove(index);
                mThresholds.remove(index);
                parent.add(this, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment(), index + 1);
                mThresholds.add(index + 1, mThreshold);
                adjustCallback.run();
            }
        });
        left.add(mMoveDownButton);

        JPanel center = new JPanel(new PrecisionLayout());
        center.setOpaque(false);
        add(center, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));

        JPanel wrapper = new JPanel(new PrecisionLayout().setColumns(8).setMargins(0));
        wrapper.setOpaque(false);
        mStateField = addField(wrapper,
                I18n.Text("State"),
                I18n.Text("A short description of the current threshold state"),
                mThreshold.getState(),
                null,
                FieldFactory.STRING,
                (evt) -> {
                    mThreshold.setState((String) evt.getNewValue());
                    adjustCallback.run();
                });
        addField(wrapper,
                I18n.Text("Multiplier"),
                I18n.Text("A multiplier to be applied first to determine the threshold value"),
                Integer.valueOf(mThreshold.getMultiplier()),
                Integer.valueOf(-999999),
                FieldFactory.INT6,
                (evt) -> {
                    mThreshold.setMultiplier(((Integer) evt.getNewValue()).intValue());
                    adjustCallback.run();
                });
        mDivisorField = addField(wrapper,
                I18n.Text("Divisor"),
                I18n.Text("A divisor to be applied second to determine the threshold value"),
                Integer.valueOf(mThreshold.getDivisor()),
                Integer.valueOf(-999999),
                FieldFactory.INT6,
                (evt) -> {
                    int value = ((Integer) evt.getNewValue()).intValue();
                    if (value == 0) {
                        mDivisorField.setValue(evt.getOldValue());
                    } else {
                        mThreshold.setDivisor(value);
                        adjustCallback.run();
                    }
                });
        addField(wrapper,
                I18n.Text("Addition"),
                I18n.Text("An addition to be applied third to determine the threshold value"),
                Integer.valueOf(mThreshold.getAddition()),
                Integer.valueOf(-999999),
                FieldFactory.INT6,
                (evt) -> {
                    mThreshold.setAddition(((Integer) evt.getNewValue()).intValue());
                    adjustCallback.run();
                });
        center.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        ThresholdOps[] opValues = ThresholdOps.values();
        wrapper = new JPanel(new PrecisionLayout().setColumns(opValues.length - 1).setMargins(0));
        wrapper.setOpaque(false);
        for (ThresholdOps op : opValues) {
            if (op != ThresholdOps.UNKNOWN) {
                addCheckBox(wrapper, op);
            }
        }
        center.add(wrapper, new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END).setGrabHorizontalSpace(true).setMargins(0));

        wrapper = new JPanel(new PrecisionLayout().setColumns(2).setMargins(0));
        wrapper.setOpaque(false);
        mExplanationField = addTextArea(wrapper,
                I18n.Text("Explanation"),
                I18n.Text("An explanation of effects of the current threshold state"),
                mThreshold.getExplanation());
        center.add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setMargins(0));

        JPanel right = new JPanel(new PrecisionLayout());
        right.setOpaque(false);
        add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        FontAwesomeButton remove = new FontAwesomeButton("\uf1f8", AttributeEditor.BUTTON_SIZE, I18n.Text("Remove"), () -> {
            ThresholdListPanel parent = (ThresholdListPanel) getParent();
            int                index  = UIUtilities.getIndexOf(parent, this);
            if (index != -1) {
                parent.remove(this);
                mThresholds.remove(index);
                adjustCallback.run();
            }
        });
        right.add(remove);
    }

    private void addLabel(Container container, String title, String tooltip) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        container.add(label, new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    private EditorField addField(Container container, String title, String tooltip, Object value, Object protoValue, JFormattedTextField.AbstractFormatterFactory formatter, PropertyChangeListener listener) {
        addLabel(container, title, tooltip);
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

    private JCheckBox addCheckBox(Container container, ThresholdOps op) {
        JCheckBox checkbox = new JCheckBox(op.title());
        checkbox.setToolTipText(Text.wrapPlainTextForToolTip(op.toString()));
        checkbox.setSelected(mThreshold.getOps().contains(op));
        checkbox.addItemListener((evt) -> {
            if (evt.getStateChange() == ItemEvent.SELECTED) {
                mThreshold.getOps().add(op);
            } else {
                mThreshold.getOps().remove(op);
            }
            mAdjustCallback.run();
        });
        container.add(checkbox);
        return checkbox;
    }

    public void adjustButtons(boolean isFirst, boolean isLast) {
        mMoveUpButton.setEnabled(!isFirst);
        mMoveDownButton.setEnabled(!isLast);
    }

    private void adjustExplanation() {
        mThreshold.setExplanation(mExplanationField.getText());
        mAdjustCallback.run();
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        adjustExplanation();
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        adjustExplanation();
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        adjustExplanation();
    }

    public void focusStateField() {
        mStateField.requestFocusInWindow();
    }
}
