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

package com.trollworks.gcs.character;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.datafile.DataFileDockable;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.menu.RetargetableFocus;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.RitualMagicSpell;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.ui.widget.Toolbar;
import com.trollworks.gcs.ui.widget.dock.Dock;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowItemRenderer;
import com.trollworks.gcs.ui.widget.outline.RowIterator;
import com.trollworks.gcs.ui.widget.outline.RowPostProcessor;
import com.trollworks.gcs.ui.widget.search.Search;
import com.trollworks.gcs.ui.widget.search.SearchTarget;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.undo.StdUndoManager;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.EventQueue;
import java.awt.KeyboardFocusManager;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import javax.swing.JComboBox;
import javax.swing.JScrollPane;
import javax.swing.JViewport;
import javax.swing.ListCellRenderer;
import javax.swing.undo.StateEdit;

/** A list of advantages and disadvantages from a library. */
public class SheetDockable extends DataFileDockable implements SearchTarget, RetargetableFocus, NotifierTarget {
    private static SheetDockable               LAST_ACTIVATED;
    private        CharacterSheet              mSheet;
    private        JComboBox<Scales>           mScaleCombo;
    private        Search                      mSearch;
    private        JComboBox<HitLocationTable> mHitLocationTableCombo;

    /** Creates a new {@link SheetDockable}. */
    public SheetDockable(GURPSCharacter character) {
        super(character);
        GURPSCharacter dataFile = getDataFile();
        mSheet = new CharacterSheet(dataFile);
        createToolbar();
        JScrollPane scroller = new JScrollPane(mSheet);
        scroller.setBorder(null);
        JViewport viewport = scroller.getViewport();
        viewport.setBackground(Color.LIGHT_GRAY);
        viewport.addChangeListener(mSheet);
        add(scroller, BorderLayout.CENTER);
        mSheet.rebuild();
        mSheet.getCharacter().processFeaturesAndPrereqs();
        dataFile.setModified(false);
        StdUndoManager undoManager = getUndoManager();
        undoManager.discardAllEdits();
        dataFile.setUndoManager(undoManager);
        dataFile.addTarget(this, Profile.ID_BODY_TYPE);
    }

    private void createToolbar() {
        Toolbar toolbar = new Toolbar();
        mScaleCombo = new JComboBox<>(Scales.values());
        mScaleCombo.setSelectedItem(Preferences.getInstance().getInitialUIScale());
        mScaleCombo.addActionListener((event) -> {
            Scales scale = (Scales) mScaleCombo.getSelectedItem();
            if (scale == null) {
                scale = Scales.ACTUAL_SIZE;
            }
            mSheet.setScale(scale.getScale());
        });
        toolbar.add(mScaleCombo);
        mSearch = new Search(this);
        toolbar.add(mSearch, Toolbar.LAYOUT_FILL);
        mHitLocationTableCombo = new JComboBox<>(HitLocationTable.ALL);
        mHitLocationTableCombo.setSelectedItem(getDataFile().getProfile().getHitLocationTable());
        mHitLocationTableCombo.addActionListener((event) -> getDataFile().getProfile().setHitLocationTable((HitLocationTable) mHitLocationTableCombo.getSelectedItem()));
        toolbar.add(mHitLocationTableCombo);
        toolbar.add(new IconButton(Images.GEAR, I18n.Text("Settings"), () -> SettingsEditor.display(getDataFile())));
        add(toolbar, BorderLayout.NORTH);
    }

    @Override
    public boolean attemptClose() {
        boolean closed = super.attemptClose();
        if (closed) {
            SettingsEditor editor = SettingsEditor.find(getDataFile());
            if (editor != null) {
                editor.attemptClose();
            }
            mSheet.dispose();
        }
        return closed;
    }

    @Override
    public Component getRetargetedFocus() {
        return mSheet;
    }

    /** @return The last activated {@link SheetDockable}. */
    public static SheetDockable getLastActivated() {
        if (LAST_ACTIVATED != null) {
            Dock dock = UIUtilities.getAncestorOfType(LAST_ACTIVATED, Dock.class);
            if (dock == null) {
                LAST_ACTIVATED = null;
            }
        }
        return LAST_ACTIVATED;
    }

    @Override
    public void activated() {
        super.activated();
        LAST_ACTIVATED = this;
    }

    @Override
    public GURPSCharacter getDataFile() {
        return (GURPSCharacter) super.getDataFile();
    }

    /** @return The {@link CharacterSheet}. */
    public CharacterSheet getSheet() {
        return mSheet;
    }

    @Override
    protected String getUntitledName() {
        String name = mSheet.getCharacter().getProfile().getName().trim();
        return (name.isBlank()) ? super.getUntitledName() : name;
    }

    @Override
    protected String getUntitledBaseName() {
        return I18n.Text("Untitled Sheet");
    }

    @Override
    public PrintProxy getPrintProxy() {
        return mSheet;
    }

    @Override
    public boolean isJumpToSearchAvailable() {
        return mSearch.isEnabled() && mSearch != KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
    }

    @Override
    public void jumpToSearchField() {
        mSearch.requestFocusInWindow();
    }

    @Override
    public ListCellRenderer<Object> getSearchRenderer() {
        return new RowItemRenderer();
    }

    @Override
    public List<Object> search(String filter) {
        List<Object> list = new ArrayList<>();
        filter = filter.toLowerCase();
        searchOne(mSheet.getAdvantageOutline(), filter, list);
        searchOne(mSheet.getSkillOutline(), filter, list);
        searchOne(mSheet.getSpellOutline(), filter, list);
        searchOne(mSheet.getEquipmentOutline(), filter, list);
        searchOne(mSheet.getNoteOutline(), filter, list);
        return list;
    }

    private static void searchOne(ListOutline outline, String text, List<Object> list) {
        for (ListRow row : new RowIterator<ListRow>(outline.getModel())) {
            if (row.contains(text, true)) {
                list.add(row);
            }
        }
    }

    @Override
    public void searchSelect(List<Object> selection) {
        Map<OutlineModel, List<Row>> map     = new HashMap<>();
        Outline                      primary = null;
        List<Row>                    list;

        mSheet.getAdvantageOutline().getModel().deselect();
        mSheet.getSkillOutline().getModel().deselect();
        mSheet.getSpellOutline().getModel().deselect();
        mSheet.getEquipmentOutline().getModel().deselect();
        mSheet.getNoteOutline().getModel().deselect();

        for (Object obj : selection) {
            Row          row    = (Row) obj;
            Row          parent = row.getParent();
            OutlineModel model  = row.getOwner();

            while (parent != null) {
                parent.setOpen(true);
                model = parent.getOwner();
                parent = parent.getParent();
            }
            list = map.get(model);
            if (list == null) {
                list = new ArrayList<>();
                list.add(row);
                map.put(model, list);
            } else {
                list.add(row);
            }
            if (primary == null) {
                primary = mSheet.getAdvantageOutline();
                if (model != primary.getModel()) {
                    primary = mSheet.getSkillOutline();
                    if (model != primary.getModel()) {
                        primary = mSheet.getSpellOutline();
                        if (model != primary.getModel()) {
                            primary = mSheet.getEquipmentOutline();
                            if (model != primary.getModel()) {
                                primary = mSheet.getNoteOutline();
                                if (model != primary.getModel()) {
                                    primary = null;
                                }
                            }
                        }
                    }
                }
            }
        }

        for (Map.Entry<OutlineModel, List<Row>> entry : map.entrySet()) {
            entry.getKey().select(entry.getValue(), false);
        }

        if (primary != null) {
            Outline outline = primary;
            EventQueue.invokeLater(() -> outline.scrollSelectionIntoView());
            primary.requestFocus();
        }
    }

    /**
     * Adds rows to the sheet.
     *
     * @param rows The rows to add.
     */
    public void addRows(List<Row> rows) {
        Map<ListOutline, StateEdit> map     = new HashMap<>();
        Map<Outline, List<Row>>     selMap  = new HashMap<>();
        Map<Outline, List<ListRow>> nameMap = new HashMap<>();
        ListOutline                 outline = null;
        String                      addRows = I18n.Text("Add Rows");

        for (Row row : rows) {
            if (row instanceof Advantage) {
                outline = mSheet.getAdvantageOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRows));
                }
                row = new Advantage(getDataFile(), (Advantage) row, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Technique) {
                outline = mSheet.getSkillOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRows));
                }
                row = new Technique(getDataFile(), (Technique) row, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Skill) {
                outline = mSheet.getSkillOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRows));
                }
                row = new Skill(getDataFile(), (Skill) row, true, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof RitualMagicSpell) {
                outline = mSheet.getSpellOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRows));
                }
                row = new RitualMagicSpell(getDataFile(), (RitualMagicSpell) row, true, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Spell) {
                outline = mSheet.getSpellOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRows));
                }
                row = new Spell(getDataFile(), (Spell) row, true, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Equipment) {
                outline = row.getOwner().getProperty(EquipmentList.TAG_OTHER_ROOT) != null ? mSheet.getOtherEquipmentOutline() : mSheet.getEquipmentOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRows));
                }
                row = new Equipment(getDataFile(), (Equipment) row, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Note) {
                outline = mSheet.getNoteOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRows));
                }
                row = new Note(getDataFile(), (Note) row, true);
                addCompleteRow(outline, row, selMap);
            } else {
                row = null;
            }
            //noinspection ConstantConditions
            if (row instanceof ListRow) {
                List<ListRow> process = nameMap.get(outline);
                if (process == null) {
                    process = new ArrayList<>();
                    nameMap.put(outline, process);
                }
                addRowsToBeProcessed(process, (ListRow) row);
            }
        }
        for (Map.Entry<ListOutline, StateEdit> entry : map.entrySet()) {
            ListOutline  anOutline = entry.getKey();
            OutlineModel model     = anOutline.getModel();
            model.select(selMap.get(anOutline), false);
            StateEdit edit = entry.getValue();
            edit.end();
            anOutline.postUndo(edit);
            anOutline.scrollSelectionIntoView();
            anOutline.requestFocus();
        }
        if (!nameMap.isEmpty()) {
            EventQueue.invokeLater(new RowPostProcessor(nameMap));
        }
    }

    private void addRowsToBeProcessed(List<ListRow> list, ListRow row) {
        int count = row.getChildCount();
        list.add(row);
        for (int i = 0; i < count; i++) {
            addRowsToBeProcessed(list, (ListRow) row.getChild(i));
        }
    }

    private void addCompleteRow(Outline outline, Row row, Map<Outline, List<Row>> selMap) {
        List<Row> selection = selMap.get(outline);
        addCompleteRow(outline.getModel(), row);
        outline.contentSizeMayHaveChanged();
        if (selection == null) {
            selection = new ArrayList<>();
            selMap.put(outline, selection);
        }
        selection.add(row);
    }

    private void addCompleteRow(OutlineModel outlineModel, Row row) {
        outlineModel.addRow(row);
        if (row.isOpen() && row.hasChildren()) {
            for (Row child : row.getChildren()) {
                addCompleteRow(outlineModel, child);
            }
        }
    }

    /** Notify background threads of prereq or feature modifications. */
    public void notifyOfPrereqOrFeatureModification() {
        if (mSheet.getCharacter().processFeaturesAndPrereqs()) {
            mSheet.repaint();
        }
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        if (Profile.ID_BODY_TYPE.equals(name)) {
            mHitLocationTableCombo.setSelectedItem(getDataFile().getProfile().getHitLocationTable());
        }
    }
}
