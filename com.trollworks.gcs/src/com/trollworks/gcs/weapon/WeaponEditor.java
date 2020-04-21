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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.skill.Defaults;
import com.trollworks.gcs.ui.Selection;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.RowDistribution;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Panel;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComboBox;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** An abstract editor for weapon statistics. */
public abstract class WeaponEditor extends JPanel implements ActionListener, PropertyChangeListener {
    private ListRow                      mOwner;
    private WeaponOutline                mOutline;
    private IconButton                   mAddButton;
    private IconButton                   mDeleteButton;
    private EditorField                  mUsage;
    private EditorField                  mStrength;
    private JComboBox<WeaponSTDamage>    mDamageSTCombo;
    private EditorField                  mDamageBase;
    private EditorField                  mDamageArmorDivisor;
    private EditorField                  mDamageType;
    private EditorField                  mDamageModPerDie;
    private EditorField                  mFragDamage;
    private EditorField                  mFragArmorDivisor;
    private EditorField                  mFragType;
    private Defaults                     mDefaults;
    private WeaponStats                  mWeapon;
    private Class<? extends WeaponStats> mWeaponClass;
    private boolean                      mRespond;

    /**
     * Creates a new {@link WeaponEditor}.
     *
     * @param owner       The owning row.
     * @param weapons     The weapons to modify.
     * @param weaponClass The {@link Class} of weapons.
     */
    public WeaponEditor(ListRow owner, List<WeaponStats> weapons, Class<? extends WeaponStats> weaponClass) {
        super(new BorderLayout());
        mOwner = owner;
        mWeaponClass = weaponClass;
        mAddButton = new IconButton(Images.ADD, I18n.Text("Add an attack"), () -> addWeapon());
        mDeleteButton = new IconButton(Images.REMOVE, I18n.Text("Remove the selected attacks"), () -> mOutline.deleteSelection());
        mDeleteButton.setEnabled(false);
        Panel top  = new Panel(new BorderLayout());
        Panel left = new Panel(new PrecisionLayout());
        left.add(mAddButton);
        left.add(mDeleteButton);
        top.add(left, BorderLayout.WEST);
        top.add(createOutline(weapons, weaponClass), BorderLayout.CENTER);
        add(top, BorderLayout.NORTH);
        add(createEditorPanel(), BorderLayout.CENTER);
        setName(toString());
    }

    /** @return The owner. */
    public ListRow getOwner() {
        return mOwner;
    }

    /** @return The weapon. */
    protected WeaponStats getWeapon() {
        return mWeapon;
    }

    /** @return The weapons in this editor. */
    public List<WeaponStats> getWeapons() {
        List<WeaponStats> weapons = new ArrayList<>();
        for (Row row : mOutline.getModel().getRows()) {
            weapons.add(((WeaponDisplayRow) row).getWeapon().clone(mOwner));
        }
        return weapons;
    }

    private Component createOutline(List<WeaponStats> weapons, Class<? extends WeaponStats> weaponClass) {
        mOutline = new WeaponOutline();
        OutlineModel model = mOutline.getModel();
        WeaponColumn.addColumns(mOutline, weaponClass, true);
        mOutline.setAllowColumnDrag(false);
        mOutline.setAllowColumnResize(false);
        mOutline.setAllowRowDrag(false);

        for (WeaponStats weapon : weapons) {
            if (mWeaponClass.isInstance(weapon)) {
                model.addRow(new WeaponDisplayRow(weapon.clone(mOwner)));
            }
        }

        mOutline.addActionListener(this);
        Dimension size = mOutline.getMinimumSize();
        if (size.height < 50) {
            size.height = 50;
            mOutline.setMinimumSize(size);
        }

        JScrollPane scroller = new JScrollPane(mOutline);
        scroller.setColumnHeaderView(mOutline.getHeaderPanel());
        return scroller;
    }

    private Container createEditorPanel() {
        JPanel wrapper = new JPanel(new ColumnLayout(2));
        wrapper.setBorder(new EmptyBorder(5));
        JPanel editorPanel = new JPanel(new ColumnLayout(1, RowDistribution.GIVE_EXCESS_TO_LAST));
        editorPanel.add(wrapper);

        JPanel firstPanel = new JPanel(new ColumnLayout(3));
        String tooltip    = I18n.Text("Usage");
        mUsage = createTextField(null, tooltip);
        wrapper.add(new LinkedLabel(tooltip, mUsage));
        firstPanel.add(mUsage);
        tooltip = I18n.Text("Minimum Strength");
        mStrength = createTextField("99**", tooltip);
        firstPanel.add(new LinkedLabel(tooltip, mStrength));
        firstPanel.add(mStrength);
        wrapper.add(firstPanel);

        JPanel damagePanel = new JPanel(new ColumnLayout(8));
        mDamageSTCombo = new JComboBox<>(WeaponSTDamage.values());
        mDamageSTCombo.setSelectedItem(WeaponSTDamage.NONE);
        mDamageSTCombo.addActionListener(this);
        mDamageSTCombo.setToolTipText(I18n.Text("Strength Damage Type"));
        UIUtilities.setToPreferredSizeOnly(mDamageSTCombo);
        wrapper.add(new LinkedLabel(I18n.Text("Damage")));
        damagePanel.add(mDamageSTCombo);
        mDamageBase = createTextField("100d+20x200", I18n.Text("Base Damage"));
        damagePanel.add(mDamageBase);
        mDamageArmorDivisor = createTextField("100", I18n.Text("Armor Divisor"));
        damagePanel.add(new LinkedLabel("(", mDamageArmorDivisor));
        damagePanel.add(mDamageArmorDivisor);
        damagePanel.add(new LinkedLabel(")", mDamageArmorDivisor));
        mDamageType = createTextField(null, I18n.Text("Type"));
        damagePanel.add(mDamageType);
        mDamageModPerDie = createTextField("+9", I18n.Text("Bonus Per Die"));
        damagePanel.add(mDamageModPerDie);
        damagePanel.add(new LinkedLabel(I18n.Text("per die"), mDamageModPerDie));
        wrapper.add(damagePanel);

        JPanel fragPanel = new JPanel(new ColumnLayout(5));
        wrapper.add(new LinkedLabel(I18n.Text("Fragmentation")));
        mFragDamage = createTextField("100d+20x200", I18n.Text("Fragmentation Damage"));
        fragPanel.add(mFragDamage);
        mFragArmorDivisor = createTextField("100", I18n.Text("Armor Divisor"));
        fragPanel.add(new LinkedLabel("(", mFragArmorDivisor));
        fragPanel.add(mFragArmorDivisor);
        fragPanel.add(new LinkedLabel(")", mFragArmorDivisor));
        mFragType = createTextField(null, I18n.Text("Type"));
        fragPanel.add(mFragType);
        wrapper.add(fragPanel);

        createFields(wrapper);
        createDefaults(editorPanel);
        setWeaponState(false);
        return editorPanel;
    }

    /**
     * Called to add the necessary fields.
     *
     * @param parent The parent to add the fields to.
     */
    protected abstract void createFields(Container parent);

    private void createDefaults(Container parent) {
        mDefaults = new Defaults(new ArrayList<>());
        mDefaults.removeAll();
        mDefaults.addActionListener(this);
        JScrollPane scrollPanel = new JScrollPane(mDefaults);
        Dimension   size        = mDefaults.getMinimumSize();
        if (size.height < 50) {
            size.height = 50;
            mDefaults.setMinimumSize(size);
        }
        scrollPanel.setPreferredSize(new Dimension(50, 50));
        parent.add(scrollPanel);
    }

    /**
     * Creates a new text field.
     *
     * @param protoValue A prototype value. If not {@code null}, will be used to set the only size
     *                   the field can have.
     * @param tooltip    The tooltip to set on the field.
     * @return The newly created field.
     */
    protected EditorField createTextField(String protoValue, String tooltip) {
        DefaultFormatter formatter = new DefaultFormatter();
        formatter.setOverwriteMode(false);
        EditorField field = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, "", protoValue, tooltip);
        if (protoValue != null) {
            UIUtilities.setToPreferredSizeOnly(field);
        }
        return field;
    }

    @Override
    public final void actionPerformed(ActionEvent event) {
        Object source = event.getSource();
        if (mOutline == source) {
            handleOutline(event.getActionCommand());
        } else if (mRespond) {
            if (mDefaults == source) {
                mWeapon.setDefaults(mDefaults.getDefaults());
                adjustOutlineToContent();
            } else if (mDamageSTCombo == source) {
                mWeapon.getDamage().setWeaponSTDamage((WeaponSTDamage) mDamageSTCombo.getSelectedItem());
                adjustOutlineToContent();
            }
        }
    }

    @Override
    public void propertyChange(PropertyChangeEvent event) {
        if (mRespond) {
            Object source = event.getSource();
            if (mUsage == source) {
                mWeapon.setUsage((String) mUsage.getValue());
            } else if (mDamageBase == source) {
                mWeapon.getDamage().setBase(new Dice((String) mDamageBase.getValue()));
            } else if (mDamageArmorDivisor == source) {
                mWeapon.getDamage().setArmorDivisor(extractArmorDivisor(mDamageArmorDivisor));
            } else if (mDamageType == source) {
                mWeapon.getDamage().setType(((String) mDamageType.getValue()).trim());
            } else if (mDamageModPerDie == source) {
                mWeapon.getDamage().setModifierPerDie(Numbers.extractInteger((String) mDamageModPerDie.getValue(), 0, true));
            } else if (mFragDamage == source) {
                WeaponDamage damage = mWeapon.getDamage();
                damage.setFragmentation(new Dice((String) mFragDamage.getValue()), adjustArmorDivisor(damage.getFragmentationArmorDivisor()), damage.getFragmentationType());
            } else if (mFragArmorDivisor == source) {
                WeaponDamage damage = mWeapon.getDamage();
                damage.setFragmentation(damage.getFragmentation(), extractArmorDivisor(mFragArmorDivisor), damage.getFragmentationType());
            } else if (mFragType == source) {
                WeaponDamage damage = mWeapon.getDamage();
                damage.setFragmentation(damage.getFragmentation(), adjustArmorDivisor(damage.getFragmentationArmorDivisor()), ((String) mFragType.getValue()).trim());
            } else if (mStrength == source) {
                mWeapon.setStrength((String) mStrength.getValue());
            } else {
                updateFromField(source);
                return;
            }
            adjustOutlineToContent();
        }
    }

    private static double extractArmorDivisor(EditorField field) {
        double value = Numbers.extractDouble((String) field.getValue(), 1, true);
        return value <= 0 ? 1 : value;
    }

    private static double adjustArmorDivisor(double divisor) {
        return divisor <= 0 ? 1 : divisor;
    }

    private static String getArmorDivisorForDisplay(double divisor) {
        return divisor == 1 || divisor <= 0 ? "" : Numbers.format(divisor);
    }

    /**
     * Called to update from a field.
     *
     * @param field The field that was altered.
     */
    protected abstract void updateFromField(Object field);

    /** Call to adjust the {@link Outline} to its new content. */
    protected void adjustOutlineToContent() {
        mOutline.sizeColumnsToFit();
        mOutline.repaintSelection();
    }

    private void addWeapon() {
        WeaponDisplayRow weapon = new WeaponDisplayRow(createWeaponStats());
        OutlineModel     model  = mOutline.getModel();
        model.addRow(weapon);
        mOutline.sizeColumnsToFit();
        model.select(weapon, false);
        mOutline.revalidate();
        mOutline.scrollSelectionIntoView();
        mOutline.requestFocus();
    }

    /** @return A newly created {@link WeaponStats}. */
    protected abstract WeaponStats createWeaponStats();

    private void handleOutline(String cmd) {
        if (Outline.CMD_SELECTION_CHANGED.equals(cmd)) {
            OutlineModel model     = mOutline.getModel();
            Selection    selection = model.getSelection();
            int          count     = selection.getCount();
            if (count == 1) {
                setWeapon(((WeaponDisplayRow) model.getRowAtIndex(selection.firstSelectedIndex())).getWeapon());
            } else {
                setWeapon(null);
            }
            mDeleteButton.setEnabled(count > 0);
        }
    }

    private void setWeapon(WeaponStats weapon) {
        if (weapon != mWeapon) {
            Commitable.sendCommitToFocusOwner();
            mWeapon = weapon;
            setWeaponState(mWeapon != null);
        }
    }

    private void setWeaponState(boolean enabled) {
        mRespond = false;
        if (enabled) {
            updateFields();
            if (mAddButton.isEnabled()) {
                mRespond = true;
            } else {
                UIUtilities.disableControls(mDefaults);
            }
        } else {
            blankFields();
        }
        if (!mAddButton.isEnabled()) {
            enabled = false;
        }
        enableFields(enabled);
    }

    /** Called to update the contents of the fields. */
    protected void updateFields() {
        mUsage.setValue(mWeapon.getUsage());
        WeaponDamage damage = mWeapon.getDamage();
        mDamageSTCombo.setSelectedItem(damage.getWeaponSTDamage());
        mDamageBase.setValue(damage.getBase() != null ? damage.getBase().toString() : "");
        mDamageArmorDivisor.setValue(getArmorDivisorForDisplay(damage.getArmorDivisor()));
        mDamageType.setValue(damage.getType());
        mDamageModPerDie.setValue(Numbers.formatWithForcedSign(damage.getModifierPerDie()));
        Dice frag = damage.getFragmentation();
        if (frag != null) {
            mFragDamage.setValue(frag.toString());
            mFragArmorDivisor.setValue(getArmorDivisorForDisplay(damage.getFragmentationArmorDivisor()));
            mFragType.setValue(damage.getFragmentationType());
        } else {
            mFragDamage.setValue("");
            mFragArmorDivisor.setValue("");
            mFragType.setValue("");
        }
        mStrength.setValue(mWeapon.getStrength());
        mDefaults.setDefaults(mWeapon.getDefaults());
    }

    /**
     * Called to enable/disable all fields.
     *
     * @param enabled Whether to enable the fields.
     */
    protected void enableFields(boolean enabled) {
        mUsage.setEnabled(enabled);
        mDamageSTCombo.setEnabled(enabled);
        mDamageBase.setEnabled(enabled);
        mDamageArmorDivisor.setEnabled(enabled);
        mDamageType.setEnabled(enabled);
        mDamageModPerDie.setEnabled(enabled);
        mFragDamage.setEnabled(enabled);
        mFragArmorDivisor.setEnabled(enabled);
        mFragType.setEnabled(enabled);
        mStrength.setEnabled(enabled);
    }

    /** Called to blank all fields. */
    protected void blankFields() {
        mUsage.setValue("");
        mDamageSTCombo.setSelectedItem(WeaponSTDamage.NONE);
        mDamageBase.setValue("");
        mDamageArmorDivisor.setValue("");
        mDamageType.setValue("");
        mDamageModPerDie.setValue("");
        mFragDamage.setValue("");
        mFragArmorDivisor.setValue("");
        mFragType.setValue("");
        mStrength.setValue("");
        mDefaults.removeAll();
        mDefaults.revalidate();
        mDefaults.repaint();
    }
}
