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

package com.trollworks.gcs.template;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.datafile.DataFileDockable;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.menu.RetargetableFocus;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.RitualMagicSpell;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.scale.Scales;
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
import com.trollworks.gcs.utility.notification.Notifier;
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
import javax.swing.ListCellRenderer;
import javax.swing.undo.StateEdit;

/** A list of advantages and disadvantages from a library. */
public class TemplateDockable extends DataFileDockable implements NotifierTarget, SearchTarget, RetargetableFocus {
    private static TemplateDockable  LAST_ACTIVATED;
    private        TemplateSheet     mTemplate;
    private        JComboBox<Scales> mScaleCombo;
    private        Search            mSearch;

    /** Creates a new {@link TemplateDockable}. */
    public TemplateDockable(Template template) {
        super(template);
        Template dataFile = getDataFile();
        mTemplate = new TemplateSheet(dataFile);
        Toolbar toolbar = new Toolbar();
        mScaleCombo = new JComboBox<>(Scales.values());
        Preferences prefs = Preferences.getInstance();
        mScaleCombo.setSelectedItem(prefs.getInitialUIScale());
        mScaleCombo.addActionListener((event) -> {
            Scales scale = (Scales) mScaleCombo.getSelectedItem();
            if (scale == null) {
                scale = Scales.ACTUAL_SIZE;
            }
            mTemplate.setScale(scale.getScale());
        });
        toolbar.add(mScaleCombo);
        mSearch = new Search(this);
        toolbar.add(mSearch, Toolbar.LAYOUT_FILL);
        add(toolbar, BorderLayout.NORTH);
        JScrollPane scroller = new JScrollPane(mTemplate);
        scroller.setBorder(null);
        scroller.getViewport().setBackground(Color.LIGHT_GRAY);
        add(scroller, BorderLayout.CENTER);
        dataFile.setModified(false);
        StdUndoManager undoManager = getUndoManager();
        undoManager.discardAllEdits();
        dataFile.setUndoManager(undoManager);
        prefs.getNotifier().add(this, Fonts.FONT_NOTIFICATION_KEY, Preferences.KEY_USE_MULTIPLICATIVE_MODIFIERS);
    }

    @Override
    public boolean attemptClose() {
        boolean closed = super.attemptClose();
        if (closed) {
            Notifier notifier = Preferences.getInstance().getNotifier();
            notifier.remove(mTemplate);
            notifier.remove(this);
        }
        return closed;
    }

    @Override
    public Component getRetargetedFocus() {
        return mTemplate;
    }

    /** @return The last activated {@link TemplateDockable}. */
    public static TemplateDockable getLastActivated() {
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
    public Template getDataFile() {
        return (Template) super.getDataFile();
    }

    /** @return The {@link TemplateSheet}. */
    public TemplateSheet getTemplate() {
        return mTemplate;
    }

    @Override
    public PrintProxy getPrintProxy() {
        return null;
    }

    @Override
    protected String getUntitledBaseName() {
        return I18n.Text("Untitled Template");
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        if (Fonts.FONT_NOTIFICATION_KEY.equals(name)) {
            mTemplate.revalidate();
            mTemplate.getAdvantageOutline().updateRowHeights();
            mTemplate.getSkillOutline().updateRowHeights();
            mTemplate.getSpellOutline().updateRowHeights();
            mTemplate.getEquipmentOutline().updateRowHeights();
        } else {
            getDataFile().notifySingle(Advantage.ID_LIST_CHANGED, null);
        }
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
        searchOne(mTemplate.getAdvantageOutline(), filter, list);
        searchOne(mTemplate.getSkillOutline(), filter, list);
        searchOne(mTemplate.getSpellOutline(), filter, list);
        searchOne(mTemplate.getEquipmentOutline(), filter, list);
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

        mTemplate.getAdvantageOutline().getModel().deselect();
        mTemplate.getSkillOutline().getModel().deselect();
        mTemplate.getSpellOutline().getModel().deselect();
        mTemplate.getEquipmentOutline().getModel().deselect();

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
                primary = mTemplate.getAdvantageOutline();
                if (model != primary.getModel()) {
                    primary = mTemplate.getSkillOutline();
                    if (model != primary.getModel()) {
                        primary = mTemplate.getSpellOutline();
                        if (model != primary.getModel()) {
                            primary = mTemplate.getEquipmentOutline();
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
     * Adds rows to the display.
     *
     * @param rows The rows to add.
     */
    public void addRows(List<Row> rows) {
        Map<ListOutline, StateEdit> map         = new HashMap<>();
        Map<Outline, List<Row>>     selMap      = new HashMap<>();
        Map<Outline, List<ListRow>> nameMap     = new HashMap<>();
        ListOutline                 outline     = null;
        String                      addRowsText = I18n.Text("Add Rows");
        for (Row row : rows) {
            if (row instanceof Advantage) {
                outline = mTemplate.getAdvantageOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRowsText));
                }
                row = new Advantage(getDataFile(), (Advantage) row, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Technique) {
                outline = mTemplate.getSkillOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRowsText));
                }
                row = new Technique(getDataFile(), (Technique) row, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Skill) {
                outline = mTemplate.getSkillOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRowsText));
                }
                row = new Skill(getDataFile(), (Skill) row, true, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof RitualMagicSpell) {
                outline = mTemplate.getSpellOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRowsText));
                }
                row = new RitualMagicSpell(getDataFile(), (RitualMagicSpell) row, true, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Spell) {
                outline = mTemplate.getSpellOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRowsText));
                }
                row = new Spell(getDataFile(), (Spell) row, true, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Equipment) {
                outline = mTemplate.getEquipmentOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRowsText));
                }
                row = new Equipment(getDataFile(), (Equipment) row, true);
                addCompleteRow(outline, row, selMap);
            } else if (row instanceof Note) {
                outline = mTemplate.getNoteOutline();
                if (!map.containsKey(outline)) {
                    map.put(outline, new StateEdit(outline.getModel(), addRowsText));
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
}
