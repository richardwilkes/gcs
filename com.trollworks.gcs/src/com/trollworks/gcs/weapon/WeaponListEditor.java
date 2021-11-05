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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.skill.Defaults;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.Selection;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontIconButton;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.settings.DamageProgression;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Container;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;
import javax.swing.text.Document;

/** An abstract editor for weapon statistics. */
public abstract class WeaponListEditor extends Panel implements ActionListener, EditorField.ChangeListener, DocumentListener {
    private ListRow                      mOwner;
    private WeaponOutline                mOutline;
    private FontIconButton               mAddButton;
    private FontIconButton               mDeleteButton;
    private FontIconButton               mDuplicateButton;
    private Panel                        mEditorPanel;
    private EditorField                  mUsage;
    private MultiLineTextField           mUsageNotes;
    private EditorField                  mStrength;
    private PopupMenu<WeaponSTDamage>    mDamageSTPopup;
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
    private PopupMenu<String>            mPercentBonusPopup;

    /**
     * Creates a new WeaponListEditor.
     *
     * @param owner       The owning row.
     * @param weapons     The weapons to modify.
     * @param weaponClass The {@link Class} of weapons.
     */
    protected WeaponListEditor(ListRow owner, List<WeaponStats> weapons, Class<? extends WeaponStats> weaponClass) {
        super(new PrecisionLayout().setMargins(0));
        mOwner = owner;
        mWeaponClass = weaponClass;
        mAddButton = new FontIconButton(FontAwesome.PLUS_CIRCLE, I18n.text("Add an attack"),
                (b) -> addWeapon());
        mDeleteButton = new FontIconButton(FontAwesome.TRASH,
                I18n.text("Remove the selected attacks"), (b) -> mOutline.deleteSelection());
        mDeleteButton.setEnabled(false);
        mDuplicateButton = new FontIconButton(FontAwesome.CLONE,
                I18n.text("Duplicate the selected attacks"), (b) -> mOutline.duplicateSelection());
        mDuplicateButton.setEnabled(false);
        Panel right = new Panel(new PrecisionLayout().setMargins(5));
        right.add(mAddButton);
        right.add(mDeleteButton);
        right.add(mDuplicateButton);
        Panel top = new Panel(new PrecisionLayout().setMargins(0).setColumns(2).setHorizontalSpacing(1));
        top.add(createOutline(weapons, weaponClass), new PrecisionLayoutData().setFillAlignment().setGrabHorizontalSpace(true));
        top.add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        add(top, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
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
        mOutline = new WeaponOutline(mOwner);
        OutlineModel model = mOutline.getModel();
        WeaponColumn.addColumns(mOutline, weaponClass, true);
        mOutline.setAllowColumnResize(false);
        mOutline.setAllowRowDrag(false);
        for (WeaponStats weapon : weapons) {
            if (mWeaponClass.isInstance(weapon)) {
                model.addRow(new WeaponDisplayRow(weapon.clone(mOwner)));
            }
        }
        mOutline.addActionListener(this);
        Panel panel = new Panel(new BorderLayout());
        panel.add(mOutline.getHeaderPanel(), BorderLayout.NORTH);
        panel.add(mOutline, BorderLayout.CENTER);
        return panel;
    }

    private Panel createEditorPanel() {
        Panel editorPanel = new Panel(new PrecisionLayout().setMargins(5).setColumns(2));

        Panel firstPanel = new Panel(new PrecisionLayout().setMargins(0).setColumns(3));
        mUsage = addField(editorPanel, firstPanel, null, I18n.text("Usage"));
        mStrength = addField(firstPanel, firstPanel, "99**", I18n.text("Minimum Strength"));
        editorPanel.add(firstPanel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        String notes = I18n.text("Notes");
        mUsageNotes = new MultiLineTextField("", notes, this);
        addLabel(editorPanel, notes);
        editorPanel.add(mUsageNotes, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        Panel damagePanel = new Panel(new PrecisionLayout().setMargins(0).setColumns(9));
        mDamageSTPopup = new PopupMenu<>(WeaponSTDamage.values(), (p) -> {
            if (mRespond) {
                mWeapon.getDamage().setWeaponSTDamage(mDamageSTPopup.getSelectedItem());
                adjustOutlineToContent();
            }
        });
        mDamageSTPopup.setSelectedItem(WeaponSTDamage.NONE, false);
        mDamageSTPopup.setToolTipText(I18n.text("Strength Damage Type"));
        editorPanel.add(new Label(I18n.text("Damage")), new PrecisionLayoutData().setFillHorizontalAlignment());
        damagePanel.add(mDamageSTPopup, new PrecisionLayoutData().setFillHorizontalAlignment());
        // Phoenix D3 Only
        String[] percentages = {"+0%","+25%","+50%","+75%","+100%","+125%","+150%"};
        mPercentBonusPopup = new PopupMenu<>(percentages,(p) -> {
            if (mRespond){
                mWeapon.getDamage().setPercentBonus(mPercentBonusPopup.getSelectedItem());
                adjustOutlineToContent();
            }
        }
        );
        if(mOwner.getDataFile().getSheetSettings().getDamageProgression() == DamageProgression.PHOENIX_D3){
            damagePanel.add(mPercentBonusPopup,new PrecisionLayoutData().setFillHorizontalAlignment());
        }
        mPercentBonusPopup.setSelectedItem(mWeapon.getDamage().getPercentBonus(), true);
        mDamageBase = addField(null, damagePanel, "9999999d+99x999", I18n.text("Base Damage"));
        addLabel(damagePanel, "(");
        mDamageArmorDivisor = addField(null, damagePanel, "100", I18n.text("Armor Divisor"));
        addLabel(damagePanel, ")");
        mDamageType = addField(null, damagePanel, null, I18n.text("Type"));
        mDamageModPerDie = addField(null, damagePanel, "+99", I18n.text("Bonus Per Die"));
        addLabel(damagePanel, I18n.text("per die"));
        editorPanel.add(damagePanel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        Panel fragPanel = new Panel(new PrecisionLayout().setMargins(0).setColumns(5));
        addLabel(editorPanel, I18n.text("Fragmentation"));
        mFragDamage = addField(null, fragPanel, "9999999d+99x999", I18n.text("Fragmentation Damage"));
        addLabel(fragPanel, "(");
        mFragArmorDivisor = addField(null, fragPanel, "9999", I18n.text("Armor Divisor"));
        addLabel(fragPanel, ")");
        mFragType = addField(null, fragPanel, null, I18n.text("Type"));
        editorPanel.add(fragPanel, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        createFields(editorPanel);
        createDefaults(editorPanel);
        return editorPanel;
    }

    /**
     * Called to add the necessary fields.
     *
     * @param parent The parent to add the fields to.
     */
    protected abstract void createFields(Container parent);

    private void createDefaults(Container parent) {
        mDefaults = new Defaults(mOwner.getDataFile(), new ArrayList<>());
        mDefaults.removeAll();
        mDefaults.addActionListener(this);
        parent.add(mDefaults, new PrecisionLayoutData().setFillHorizontalAlignment().setHorizontalSpan(2).setGrabHorizontalSpace(true));
    }

    protected void addLabel(Container parent, String label) {
        parent.add(new Label(label), new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    protected EditorField addField(Container labelParent, Container fieldParent, String protoValue, String label) {
        DefaultFormatter formatter = new DefaultFormatter();
        formatter.setOverwriteMode(false);
        EditorField         field = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, "", protoValue, label);
        PrecisionLayoutData ld    = new PrecisionLayoutData().setFillHorizontalAlignment();
        if (protoValue == null) {
            ld.setGrabHorizontalSpace(true);
        }
        if (labelParent != null) {
            addLabel(labelParent, label);
        }
        fieldParent.add(field, ld);
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
            }
        }
    }

    @Override
    public void editorFieldChanged(EditorField field) {
        if (mRespond) {
            if (mUsage == field) {
                mWeapon.setUsage((String) mUsage.getValue());
            } else if (mDamageBase == field) {
                mWeapon.getDamage().setBase(new Dice((String) mDamageBase.getValue()));
            } else if (mDamageArmorDivisor == field) {
                mWeapon.getDamage().setArmorDivisor(extractArmorDivisor(mDamageArmorDivisor));
            } else if (mDamageType == field) {
                mWeapon.getDamage().setType(((String) mDamageType.getValue()).trim());
            } else if (mDamageModPerDie == field) {
                mWeapon.getDamage().setModifierPerDie(Numbers.extractInteger((String) mDamageModPerDie.getValue(), 0, true));
            } else if (mFragDamage == field) {
                WeaponDamage damage = mWeapon.getDamage();
                damage.setFragmentation(new Dice((String) mFragDamage.getValue()), adjustArmorDivisor(damage.getFragmentationArmorDivisor()), damage.getFragmentationType());
            } else if (mFragArmorDivisor == field) {
                WeaponDamage damage = mWeapon.getDamage();
                damage.setFragmentation(damage.getFragmentation(), extractArmorDivisor(mFragArmorDivisor), damage.getFragmentationType());
            } else if (mFragType == field) {
                WeaponDamage damage = mWeapon.getDamage();
                damage.setFragmentation(damage.getFragmentation(), adjustArmorDivisor(damage.getFragmentationArmorDivisor()), ((String) mFragType.getValue()).trim());
            } else if (mStrength == field) {
                mWeapon.setStrength((String) mStrength.getValue());
            } else {
                updateFromField(field);
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
    protected abstract void updateFromField(EditorField field);

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
            } else if (mEditorPanel != null) {
                setWeapon(null);
            }
            mDeleteButton.setEnabled(count > 0);
            mDuplicateButton.setEnabled(count > 0);
        }
    }

    private void setWeapon(WeaponStats weapon) {
        if (weapon != mWeapon) {
            Commitable.sendCommitToFocusOwner();
            mWeapon = weapon;
            if (mEditorPanel != null) {
                remove(mEditorPanel);
                mEditorPanel = null;
            }
            if (mWeapon != null) {
                mEditorPanel = createEditorPanel();
                add(mEditorPanel, new PrecisionLayoutData().setFillAlignment().setGrabSpace(true));
                setWeaponState();
            }
            revalidate();
        }
    }

    private void setWeaponState() {
        mRespond = false;
        updateFields();
        if (mAddButton.isEnabled()) {
            mRespond = true;
        } else {
            UIUtilities.disableControls(mDefaults);
        }
        enableFields(mAddButton.isEnabled());
    }

    /** Called to update the contents of the fields. */
    protected void updateFields() {
        mUsage.setValue(mWeapon.getUsage());
        mUsageNotes.setText(mWeapon.getUsageNotes());
        WeaponDamage damage = mWeapon.getDamage();
        mDamageSTPopup.setSelectedItem(damage.getWeaponSTDamage(), true);
        mPercentBonusPopup.setSelectedItem(damage.getPercentBonus(), true);
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
        mDamageSTPopup.setEnabled(enabled);
        mDamageBase.setEnabled(enabled);
        mDamageArmorDivisor.setEnabled(enabled);
        mDamageType.setEnabled(enabled);
        mDamageModPerDie.setEnabled(enabled);
        mFragDamage.setEnabled(enabled);
        mFragArmorDivisor.setEnabled(enabled);
        mFragType.setEnabled(enabled);
        mStrength.setEnabled(enabled);
        mPercentBonusPopup.setEnabled(enabled);
    }

    private void docChanged(DocumentEvent event) {
        Document doc = event.getDocument();
        if (mUsageNotes.getDocument() == doc) {
            mWeapon.setUsageNotes(mUsageNotes.getText());
        }
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        docChanged(event);
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        docChanged(event);
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        docChanged(event);
    }
}
