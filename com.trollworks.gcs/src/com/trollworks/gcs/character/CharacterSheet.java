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

import com.lowagie.text.Document;
import com.lowagie.text.pdf.DefaultFontMapper;
import com.lowagie.text.pdf.PdfContentByte;
import com.lowagie.text.pdf.PdfTemplate;
import com.lowagie.text.pdf.PdfWriter;
import com.trollworks.gcs.GCS;
import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.SelfControlRoll;
import com.trollworks.gcs.advantage.SelfControlRollAdjustments;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentColumn;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.feature.ReactionBonus;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.page.Page;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageOwner;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.Selection;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.print.PrintManager;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.OutlineSyncer;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowIterator;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponDisplayRow;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.KeyboardFocusManager;
import java.awt.Rectangle;
import java.awt.Transparency;
import java.awt.dnd.DropTarget;
import java.awt.event.ActionEvent;
import java.awt.print.PageFormat;
import java.awt.print.Paper;
import java.io.OutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Pattern;
import javax.imageio.ImageIO;
import javax.swing.RepaintManager;
import javax.swing.UIManager;
import javax.swing.event.ChangeEvent;
import javax.swing.event.ChangeListener;

/** The character sheet. */
public class CharacterSheet extends CollectedOutlines implements ChangeListener, PageOwner, PrintProxy, Runnable {
    private static final int              GAP                 = 2;
    public static final  String           REACTIONS_KEY       = "reactions";
    public static final  String           MELEE_KEY           = "melee";
    public static final  String           RANGED_KEY          = "ranged";
    public static final  String           ADVANTAGES_KEY      = "advantages";
    public static final  String           SKILLS_KEY          = "skills";
    public static final  String           SPELLS_KEY          = "spells";
    public static final  String           EQUIPMENT_KEY       = "equipment";
    public static final  String           OTHER_EQUIPMENT_KEY = "other_equipment";
    public static final  String           NOTES_KEY           = "notes";
    private static final String[]         ALL_KEYS            = {REACTIONS_KEY, MELEE_KEY, RANGED_KEY, ADVANTAGES_KEY, SKILLS_KEY, SPELLS_KEY, EQUIPMENT_KEY, OTHER_EQUIPMENT_KEY, NOTES_KEY};
    private static final Pattern          SCHEME_PATTERN      = Pattern.compile(".*://");
    private              GURPSCharacter   mCharacter;
    private              int              mLastPage;
    private              WeaponOutline    mMeleeWeaponOutline;
    private              WeaponOutline    mRangedWeaponOutline;
    private              ReactionsOutline mReactionsOutline;
    private              boolean          mRebuildPending;
    private              Set<Outline>     mRootsToSync;
    private              Scale            mSavedScale;
    private              boolean          mOkToPaint          = true;
    private              boolean          mIsPrinting;
    private              boolean          mSyncWeapons;
    private              boolean          mReloadSpellColumns;

    /**
     * Creates a new character sheet display. {@link #rebuild()} must be called prior to the first
     * display of this panel.
     *
     * @param character The character to display the data for.
     */
    public CharacterSheet(GURPSCharacter character) {
        setLayout(new CharacterSheetLayout(this));
        setOpaque(false);
        mCharacter = character;
        mLastPage = -1;
        mRootsToSync = new HashSet<>();
        if (!GraphicsUtilities.inHeadlessPrintMode()) {
            setDropTarget(new DropTarget(this, this));
        }
        mCharacter.addTarget(this, FEATURES_AND_PREREQS_NOTIFICATIONS.toArray(new String[0]));
        mCharacter.addTarget(this, Settings.PREFIX);
        Preferences.getInstance().getNotifier().add(this, Fonts.FONT_NOTIFICATION_KEY);
    }

    /** Call when the sheet is no longer in use. */
    public void dispose() {
        Preferences.getInstance().getNotifier().remove(this);
        mCharacter.resetNotifier();
    }

    @Override
    protected void scaleChanged() {
        markForRebuild();
    }

    /** Synchronizes the display with the underlying model. */
    public void rebuild() {
        KeyboardFocusManager focusMgr = KeyboardFocusManager.getCurrentKeyboardFocusManager();
        Component            focus    = focusMgr.getPermanentFocusOwner();
        int                  firstRow = 0;
        String               focusKey = null;
        PageAssembler        pageAssembler;

        if (UIUtilities.getSelfOrAncestorOfType(focus, CharacterSheet.class) == this) {
            if (focus instanceof PageField) {
                focusKey = ((PageField) focus).getConsumedType();
                focus = null;
            } else if (focus instanceof Outline) {
                Outline   outline   = (Outline) focus;
                Selection selection = outline.getModel().getSelection();

                firstRow = outline.getFirstRowToDisplay();
                int selRow = selection.nextSelectedIndex(firstRow);
                if (selRow >= 0) {
                    firstRow = selRow;
                }
                focus = outline.getRealOutline();
            }
            focusMgr.clearFocusOwner();
        } else {
            focus = null;
        }

        // Make sure our primary outlines exist
        createOutlines(mCharacter);
        createMeleeWeaponOutline();
        createRangedWeaponOutline();
        createReactionsOutline();

        // Clear out the old pages
        removeAll();
        List<NotifierTarget> targets = new ArrayList<>();
        targets.add(this);
        SheetDockable sheetDockable = UIUtilities.getAncestorOfType(this, SheetDockable.class);
        if (sheetDockable != null) {
            targets.add(sheetDockable);
        }
        mCharacter.resetNotifier(targets.toArray(new NotifierTarget[0]));

        // Create the first page, which holds stuff that has a fixed vertical size.
        pageAssembler = new PageAssembler(this);
        Wrapper wrapper = new Wrapper(new PrecisionLayout().setColumns(4).setMargins(0).setSpacing(GAP, GAP).setFillAlignment());
        wrapper.add(new PortraitPanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment().setVerticalSpan(2));
        wrapper.add(new IdentityPanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment().setGrabHorizontalSpace(true));
        wrapper.add(new MiscPanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment());
        wrapper.add(new PointsPanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment().setVerticalSpan(2));
        wrapper.add(new DescriptionPanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment().setHorizontalSpan(2));
        pageAssembler.addToContent(wrapper, null, null);

        wrapper = new Wrapper(new PrecisionLayout().setColumns(4).setMargins(0).setSpacing(GAP, GAP).setFillAlignment());
        wrapper.add(new AttributesPanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment());
        Wrapper wrapper2 = new Wrapper(new PrecisionLayout().setMargins(0).setSpacing(GAP, GAP).setFillAlignment());
        wrapper2.add(new FatiguePointsPanel(this), new PrecisionLayoutData().setFillAlignment().setGrabHorizontalSpace(true));
        wrapper2.add(new HitPointsPanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment().setGrabHorizontalSpace(true));
        wrapper.add(wrapper2, new PrecisionLayoutData().setGrabSpace(true).setFillAlignment());
        wrapper.add(new HitLocationPanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment());
        wrapper2 = new Wrapper(new PrecisionLayout().setMargins(0).setSpacing(GAP, GAP).setFillAlignment());
        wrapper2.add(new EncumbrancePanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment());
        wrapper2.add(new LiftPanel(this), new PrecisionLayoutData().setGrabVerticalSpace(true).setFillAlignment());
        wrapper.add(wrapper2, new PrecisionLayoutData().setFillAlignment());
        pageAssembler.addToContent(wrapper, null, null);

        // Add the various outline blocks, based on the layout preference.
        boolean     addedAtLeastOneOutline = false;
        Set<String> remaining              = prepBlockLayoutRemaining();
        for (String line : mCharacter.getSettings().blockLayout()) {
            String[] parts = line.trim().toLowerCase().split(" ");
            if (!parts[0].isEmpty() && remaining.contains(parts[0])) {
                Outline o1 = getOutlineForKey(parts[0]);
                if (o1 != null) {
                    String t1 = getOutlineTitleForKey(parts[0]);
                    remaining.remove(parts[0]);
                    if (parts.length > 1 && remaining.contains(parts[1])) {
                        Outline o2 = getOutlineForKey(parts[1]);
                        if (o2 != null) {
                            String t2 = getOutlineTitleForKey(parts[1]);
                            remaining.remove(parts[1]);
                            if (o1.getModel().getRowCount() > 0 && o2.getModel().getRowCount() > 0) {
                                addedAtLeastOneOutline = true;
                                addOutline(pageAssembler, o1, t1, o2, t2);
                            } else {
                                addedAtLeastOneOutline |= addOutline(pageAssembler, o1, t1);
                                addedAtLeastOneOutline |= addOutline(pageAssembler, o2, t2);
                            }
                            continue;
                        }
                    }
                    addedAtLeastOneOutline |= addOutline(pageAssembler, o1, t1);
                }
            }
        }
        for (String one : ALL_KEYS) {
            if (remaining.contains(one)) {
                Outline outline = getOutlineForKey(one);
                if (outline != null) {
                    addedAtLeastOneOutline |= addOutline(pageAssembler, outline, getOutlineTitleForKey(one));
                }
            }
        }
        if (!addedAtLeastOneOutline) {
            pageAssembler.addToContent(new Wrapper(), null, null);
        }
        pageAssembler.finish();

        // Ensure everything is laid out and register for notification
        validate();
        OutlineSyncer.remove(mReactionsOutline);
        OutlineSyncer.remove(mMeleeWeaponOutline);
        OutlineSyncer.remove(mRangedWeaponOutline);
        OutlineSyncer.remove(getAdvantageOutline());
        OutlineSyncer.remove(getSkillOutline());
        OutlineSyncer.remove(getSpellOutline());
        OutlineSyncer.remove(getEquipmentOutline());
        OutlineSyncer.remove(getOtherEquipmentOutline());
        OutlineSyncer.remove(getNoteOutline());
        mCharacter.addTarget(this, GURPSCharacter.CHARACTER_PREFIX);
        mCharacter.calculateWeightAndWealthCarried(true);
        mCharacter.calculateWealthNotCarried(true);
        if (focusKey != null) {
            restoreFocusToKey(focusKey, this);
        } else if (focus instanceof Outline) {
            ((Outline) focus).getBestOutlineForRowIndex(firstRow).requestFocus();
        } else if (focus != null) {
            focus.requestFocus();
        }
        setSize(getPreferredSize());
        repaint();
    }

    private static Set<String> prepBlockLayoutRemaining() {
        Set<String> remaining = new HashSet<>();
        remaining.add(REACTIONS_KEY);
        remaining.add(MELEE_KEY);
        remaining.add(RANGED_KEY);
        remaining.add(ADVANTAGES_KEY);
        remaining.add(SKILLS_KEY);
        remaining.add(SPELLS_KEY);
        remaining.add(EQUIPMENT_KEY);
        remaining.add(OTHER_EQUIPMENT_KEY);
        remaining.add(NOTES_KEY);
        return remaining;
    }

    public String getHTMLGridTemplate() {
        Set<String>   remaining = prepBlockLayoutRemaining();
        StringBuilder buffer    = new StringBuilder();
        for (String line : mCharacter.getSettings().blockLayout()) {
            String[] parts = line.trim().toLowerCase().split(" ");
            if (!parts[0].isEmpty() && remaining.contains(parts[0])) {
                remaining.remove(parts[0]);
                if (parts.length > 1 && remaining.contains(parts[1])) {
                    remaining.remove(parts[1]);
                    appendToGridTemplate(buffer, parts[0], parts[1]);
                    continue;
                }
                appendToGridTemplate(buffer, parts[0], parts[0]);
            }
        }
        for (String one : ALL_KEYS) {
            if (remaining.contains(one)) {
                appendToGridTemplate(buffer, one, one);
            }
        }
        return buffer.toString();
    }

    private static void appendToGridTemplate(StringBuilder buffer, String col1, String col2) {
        buffer.append('"');
        buffer.append(col1);
        buffer.append(' ');
        buffer.append(col2);
        buffer.append('"');
        buffer.append('\n');
    }

    private static String getOutlineTitleForKey(String key) {
        return switch (key) {
            case REACTIONS_KEY -> I18n.Text("Reactions");
            case MELEE_KEY -> I18n.Text("Melee Weapons");
            case RANGED_KEY -> I18n.Text("Ranged Weapons");
            case ADVANTAGES_KEY -> I18n.Text("Advantages, Disadvantages & Quirks");
            case SKILLS_KEY -> I18n.Text("Skills");
            case SPELLS_KEY -> I18n.Text("Spells");
            case EQUIPMENT_KEY -> I18n.Text("Equipment");
            case OTHER_EQUIPMENT_KEY -> I18n.Text("Other Equipment");
            case NOTES_KEY -> I18n.Text("Notes");
            default -> "";
        };
    }

    private Outline getOutlineForKey(String key) {
        return switch (key) {
            case REACTIONS_KEY -> mReactionsOutline;
            case MELEE_KEY -> mMeleeWeaponOutline;
            case RANGED_KEY -> mRangedWeaponOutline;
            case ADVANTAGES_KEY -> getAdvantageOutline();
            case SKILLS_KEY -> getSkillOutline();
            case SPELLS_KEY -> getSpellOutline();
            case EQUIPMENT_KEY -> getEquipmentOutline();
            case OTHER_EQUIPMENT_KEY -> getOtherEquipmentOutline();
            case NOTES_KEY -> getNoteOutline();
            default -> null;
        };
    }

    private boolean restoreFocusToKey(String key, Component panel) {
        if (key != null) {
            if (panel instanceof PageField) {
                if (key.equals(((PageField) panel).getConsumedType())) {
                    panel.requestFocus();
                    return true;
                }
            } else if (panel instanceof Container) {
                Container container = (Container) panel;

                if (container.getComponentCount() > 0) {
                    for (Component child : container.getComponents()) {
                        if (restoreFocusToKey(key, child)) {
                            return true;
                        }
                    }
                }
            }
        }
        return false;
    }

    private boolean addOutline(PageAssembler pageAssembler, Outline outline, String title) {
        if (outline.getModel().getRowCount() > 0) {
            OutlineInfo info     = new OutlineInfo(outline, pageAssembler.getContentWidth());
            boolean     useProxy = false;
            while (pageAssembler.addToContent(new SingleOutlinePanel(getScale(), outline, title, useProxy), info, null)) {
                if (!useProxy) {
                    title = MessageFormat.format(I18n.Text("{0} (continued)"), title);
                    useProxy = true;
                }
            }
            return true;
        }
        return false;
    }

    private void addOutline(PageAssembler pageAssembler, Outline leftOutline, String leftTitle, Outline rightOutline, String rightTitle) {
        int         width     = pageAssembler.getContentWidth() / 2 - 1;
        OutlineInfo infoLeft  = new OutlineInfo(leftOutline, width);
        OutlineInfo infoRight = new OutlineInfo(rightOutline, width);
        boolean     useProxy  = false;
        while (pageAssembler.addToContent(new DoubleOutlinePanel(getScale(), leftOutline, leftTitle, rightOutline, rightTitle, useProxy), infoLeft, infoRight)) {
            if (!useProxy) {
                leftTitle = MessageFormat.format(I18n.Text("{0} (continued)"), leftTitle);
                rightTitle = MessageFormat.format(I18n.Text("{0} (continued)"), rightTitle);
                useProxy = true;
            }
        }
    }

    @Override
    protected void createOutlines(CollectedModels models) {
        if (mReloadSpellColumns) {
            mReloadSpellColumns = false;
            SpellOutline spellOutline = getSpellOutline();
            if (spellOutline != null) {
                spellOutline.resetColumns();
            }
        }
        super.createOutlines(models);
    }

    private void createReactionsOutline() {
        if (mReactionsOutline == null) {
            OutlineModel outlineModel;
            String       sortConfig;
            mReactionsOutline = new ReactionsOutline();
            outlineModel = mReactionsOutline.getModel();
            sortConfig = outlineModel.getSortConfig();
            for (ReactionRow row : collectReactions()) {
                outlineModel.addRow(row);
            }
            outlineModel.applySortConfig(sortConfig);
            initOutline(mReactionsOutline);
        }
        resetOutline(mReactionsOutline);
    }

    public List<ReactionRow> collectReactions() {
        Map<String, ReactionRow> reactionMap = new HashMap<>();
        for (Advantage advantage : mCharacter.getAdvantagesIterator(false)) {
            String source = String.format(I18n.Text("from advantage %s"), advantage.getName());
            collectReactionsFromFeatureList(source, advantage.getFeatures(), reactionMap);
            for (AdvantageModifier modifier : advantage.getModifiers()) {
                if (modifier.isEnabled()) {
                    collectReactionsFromFeatureList(source, modifier.getFeatures(), reactionMap);
                }
            }
            SelfControlRoll cr = advantage.getCR();
            if (cr != SelfControlRoll.NONE_REQUIRED) {
                SelfControlRollAdjustments crAdj = advantage.getCRAdj();
                if (crAdj == SelfControlRollAdjustments.REACTION_PENALTY) {
                    int         amt       = SelfControlRollAdjustments.REACTION_PENALTY.getAdjustment(cr);
                    String      situation = String.format("from others when %s is triggered", advantage.getName());
                    ReactionRow existing  = reactionMap.get(situation);
                    if (existing == null) {
                        reactionMap.put(situation, new ReactionRow(amt, situation, source));
                    } else {
                        existing.addAmount(amt, source);
                    }
                }
            }
        }
        for (Equipment equipment : mCharacter.getEquipmentIterator()) {
            if (equipment.getQuantity() > 0 && equipment.isEquipped()) {
                String source = String.format(I18n.Text("from equipment %s"), equipment.getDescription());
                collectReactionsFromFeatureList(source, equipment.getFeatures(), reactionMap);
                for (EquipmentModifier modifier : equipment.getModifiers()) {
                    if (modifier.isEnabled()) {
                        collectReactionsFromFeatureList(source, modifier.getFeatures(), reactionMap);
                    }
                }
            }
        }
        return new ArrayList<>(reactionMap.values());
    }

    private void collectReactionsFromFeatureList(String source, List<Feature> features, Map<String, ReactionRow> reactionMap) {
        for (Feature feature : features) {
            if (feature instanceof ReactionBonus) {
                ReactionBonus bonus     = (ReactionBonus) feature;
                int           amount    = bonus.getAmount().getIntegerAdjustedAmount();
                String        situation = bonus.getSituation();
                ReactionRow   existing  = reactionMap.get(situation);
                if (existing == null) {
                    reactionMap.put(situation, new ReactionRow(amount, situation, source));
                } else {
                    existing.addAmount(amount, source);
                }
            }
        }
    }

    /** @return The outline containing the melee weapons. */
    public WeaponOutline getMeleeWeaponOutline() {
        return mMeleeWeaponOutline;
    }

    private void createMeleeWeaponOutline() {
        if (mMeleeWeaponOutline == null) {
            OutlineModel outlineModel;
            String       sortConfig;
            mMeleeWeaponOutline = new WeaponOutline(MeleeWeaponStats.class);
            outlineModel = mMeleeWeaponOutline.getModel();
            sortConfig = outlineModel.getSortConfig();
            for (WeaponDisplayRow row : collectWeapons(MeleeWeaponStats.class)) {
                outlineModel.addRow(row);
            }
            outlineModel.applySortConfig(sortConfig);
            initOutline(mMeleeWeaponOutline);
        }
        resetOutline(mMeleeWeaponOutline);
    }

    /** @return The outline containing the ranged weapons. */
    public WeaponOutline getRangedWeaponOutline() {
        return mRangedWeaponOutline;
    }

    private void createRangedWeaponOutline() {
        if (mRangedWeaponOutline == null) {
            OutlineModel outlineModel;
            String       sortConfig;

            mRangedWeaponOutline = new WeaponOutline(RangedWeaponStats.class);
            outlineModel = mRangedWeaponOutline.getModel();
            sortConfig = outlineModel.getSortConfig();
            for (WeaponDisplayRow row : collectWeapons(RangedWeaponStats.class)) {
                outlineModel.addRow(row);
            }
            outlineModel.applySortConfig(sortConfig);
            initOutline(mRangedWeaponOutline);
        }
        resetOutline(mRangedWeaponOutline);
    }

    private List<WeaponDisplayRow> collectWeapons(Class<? extends WeaponStats> weaponClass) {
        Map<HashedWeapon, WeaponDisplayRow> weaponMap = new HashMap<>();
        for (Advantage advantage : mCharacter.getAdvantagesIterator(false)) {
            for (WeaponStats weapon : advantage.getWeapons()) {
                if (weaponClass.isInstance(weapon)) {
                    weaponMap.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
                }
            }
        }
        for (Equipment equipment : mCharacter.getEquipmentIterator()) {
            if (equipment.getQuantity() > 0 && equipment.isEquipped()) {
                for (WeaponStats weapon : equipment.getWeapons()) {
                    if (weaponClass.isInstance(weapon)) {
                        weaponMap.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
                    }
                }
            }
        }
        for (Spell spell : mCharacter.getSpellsIterator()) {
            for (WeaponStats weapon : spell.getWeapons()) {
                if (weaponClass.isInstance(weapon)) {
                    weaponMap.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
                }
            }
        }
        for (Skill skill : mCharacter.getSkillsIterator()) {
            for (WeaponStats weapon : skill.getWeapons()) {
                if (weaponClass.isInstance(weapon)) {
                    weaponMap.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
                }
            }
        }
        return new ArrayList<>(weaponMap.values());
    }

    /** @return The number of pages in this character sheet. */
    public int getPageCount() {
        return getComponentCount();
    }

    @Override
    public void paint(Graphics g) {
        if (mOkToPaint) {
            super.paint(g);
        }
    }

    @Override
    public int print(Graphics graphics, PageFormat pageFormat, int pageIndex) {
        if (pageIndex >= getComponentCount()) {
            mLastPage = -1;
            return NO_SUCH_PAGE;
        }

        // We do the following trick to avoid going through the work twice,
        // as we are called twice for each page, the first of which doesn't
        // seem to be used.
        if (mLastPage == pageIndex) {
            Component      comp  = getComponent(pageIndex);
            RepaintManager mgr   = RepaintManager.currentManager(comp);
            boolean        saved = mgr.isDoubleBufferingEnabled();
            mgr.setDoubleBufferingEnabled(false);
            mOkToPaint = true;
            comp.print(graphics);
            mOkToPaint = false;
            mgr.setDoubleBufferingEnabled(saved);
        } else {
            mLastPage = pageIndex;
        }
        return PAGE_EXISTS;
    }

    private static Set<String> MARK_FOR_REBUILD_NOTIFICATIONS        = new HashSet<>();
    private static Set<String> MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS = new HashSet<>();
    private static Set<String> FEATURES_AND_PREREQS_NOTIFICATIONS    = new HashSet<>();

    static {
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_BLOCK_LAYOUT);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_DEFAULT_WEIGHT_UNITS);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Fonts.FONT_NOTIFICATION_KEY);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Profile.ID_BODY_TYPE);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_SIMPLE_METRIC_CONVERSIONS);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_MODIFYING_DICE_PLUS_ADDS);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_REDUCED_SWING);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_KNOW_YOUR_OWN_STRENGTH);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_THRUST_EQUALS_SWING_MINUS_2);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_SHOW_COLLEGE_IN_SPELLS);

        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Advantage.ID_DISABLED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Advantage.ID_WEAPON_STATUS_CHANGED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Equipment.ID_EQUIPPED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Equipment.ID_QUANTITY);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Equipment.ID_WEAPON_STATUS_CHANGED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Skill.ID_WEAPON_STATUS_CHANGED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Spell.ID_WEAPON_STATUS_CHANGED);

        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Advantage.ID_LEVELS);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Advantage.ID_LIST_CHANGED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Advantage.ID_NAME);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Equipment.ID_EQUIPPED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Equipment.ID_EXTENDED_WEIGHT);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Equipment.ID_LIST_CHANGED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Equipment.ID_QUANTITY);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(EquipmentModifier.ID_COST_ADJ);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(EquipmentModifier.ID_WEIGHT_ADJ);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_DEXTERITY);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_HEALTH);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_INTELLIGENCE);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_PERCEPTION);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_STRENGTH);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_WILL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Profile.ID_TECH_LEVEL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_ENCUMBRANCE_PENALTY);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_LEVEL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_LIST_CHANGED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_NAME);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_POINTS);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_RELATIVE_LEVEL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_SPECIALIZATION);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_TECH_LEVEL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Spell.ID_COLLEGE);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Spell.ID_LIST_CHANGED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Spell.ID_NAME);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Spell.ID_POINTS);
    }

    @Override
    public void handleNotification(Object producer, String type, Object data) {
        if (Settings.ID_SHOW_COLLEGE_IN_SPELLS.equals(type)) {
            mReloadSpellColumns = true;
        }
        if (MARK_FOR_REBUILD_NOTIFICATIONS.contains(type)) {
            markForRebuild();
        } else {
            if (type.startsWith(Advantage.PREFIX)) {
                OutlineSyncer.add(getAdvantageOutline());
                OutlineSyncer.add(mReactionsOutline);
                mSyncWeapons = true;
                markForRebuild();
            } else if (Settings.ID_USER_DESCRIPTION_DISPLAY.equals(type) || Settings.ID_MODIFIERS_DISPLAY.equals(type) || Settings.ID_NOTES_DISPLAY.equals(type)) {
                OutlineSyncer.add(getAdvantageOutline());
                OutlineSyncer.add(getSkillOutline());
                OutlineSyncer.add(getSpellOutline());
                OutlineSyncer.add(getEquipmentOutline());
                OutlineSyncer.add(getOtherEquipmentOutline());
                mSyncWeapons = true;
                markForRebuild();
            } else if (type.startsWith(Skill.PREFIX)) {
                OutlineSyncer.add(getSkillOutline());
            } else if (type.startsWith(Spell.PREFIX)) {
                OutlineSyncer.add(getSpellOutline());
            } else if (type.startsWith(Equipment.PREFIX)) {
                OutlineSyncer.add(getEquipmentOutline());
                OutlineSyncer.add(getOtherEquipmentOutline());
                OutlineSyncer.add(mReactionsOutline);
                mSyncWeapons = true;
                markForRebuild();
            } else if (type.startsWith(Note.PREFIX)) {
                OutlineSyncer.add(getNoteOutline());
            }

            if (MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.contains(type)) {
                mSyncWeapons = true;
                markForRebuild();
            } else if (GURPSCharacter.ID_PARRY_BONUS.equals(type) || Skill.ID_LEVEL.equals(type)) {
                OutlineSyncer.add(mMeleeWeaponOutline);
                OutlineSyncer.add(mRangedWeaponOutline);
            } else if (GURPSCharacter.ID_CARRIED_WEIGHT.equals(type) || GURPSCharacter.ID_CARRIED_WEALTH.equals(type)) {
                Column column = getEquipmentOutline().getModel().getColumnWithID(EquipmentColumn.DESCRIPTION.ordinal());
                column.setName(EquipmentColumn.DESCRIPTION.toString(mCharacter, true));
                if (GURPSCharacter.ID_CARRIED_WEIGHT.equals(type)) {
                    mCharacter.updateSkills();
                }
            } else if (GURPSCharacter.ID_NOT_CARRIED_WEALTH.equals(type)) {
                Column column = getOtherEquipmentOutline().getModel().getColumnWithID(EquipmentColumn.DESCRIPTION.ordinal());
                column.setName(EquipmentColumn.DESCRIPTION.toString(mCharacter, false));
            } else if (Settings.ID_BASE_WILL_AND_PER_ON_10.equals(type)) {
                mCharacter.updateWillAndPerceptionDueToOptionalIQRuleUseChange();
            } else if (Settings.ID_USE_MULTIPLICATIVE_MODIFIERS.equals(type)) {
                mCharacter.notifySingle(Advantage.ID_LIST_CHANGED, null);
            } else if (Settings.ID_USE_KNOW_YOUR_OWN_STRENGTH.equals(type)) {
                mCharacter.notifySingle(type, data);
            } else if (Settings.ID_USE_THRUST_EQUALS_SWING_MINUS_2.equals(type)) {
                mCharacter.notifySingle(type, data);
            } else if (Advantage.ID_LEVELS.equals(type)) {
                markForRebuild();
            } else if (GURPSCharacter.ID_MODIFIED.equals(type) || Settings.ID_USE_TITLE_IN_FOOTER.equals(type)) {
                int count = getComponentCount();
                for (int i = 0; i < count; i++) {
                    Page      page   = (Page) getComponent(i);
                    Rectangle bounds = page.getBounds();
                    Insets    insets = page.getInsets();
                    bounds.y = bounds.y + bounds.height - insets.bottom;
                    bounds.height = insets.bottom;
                    repaint(bounds);
                }
            }
        }
        if (FEATURES_AND_PREREQS_NOTIFICATIONS.contains(type)) {
            if (mCharacter.processFeaturesAndPrereqs()) {
                repaint();
            }
        }
        if (!inBatchMode() && !mRebuildPending) {
            validate();
        }
    }

    @Override
    public void drawPageAdornments(Page page, Graphics gc) {
        Rectangle bounds = page.getBounds();
        Insets    insets = page.getInsets();
        bounds.width -= insets.left + insets.right;
        bounds.height -= insets.top + insets.bottom;
        bounds.x = insets.left;
        bounds.y = insets.top;
        int         pageNumber = 1 + UIUtilities.getIndexOf(this, page);
        String      pageString = MessageFormat.format(I18n.Text("Page {0} of {1}"), Numbers.format(pageNumber), Numbers.format(getPageCount()));
        Scale       scale      = getScale();
        Font        font1      = scale.scale(UIManager.getFont(Fonts.KEY_FOOTER_SECONDARY));
        Font        font2      = scale.scale(UIManager.getFont(Fonts.KEY_FOOTER_PRIMARY));
        FontMetrics fm1        = gc.getFontMetrics(font1);
        FontMetrics fm2        = gc.getFontMetrics(font2);
        int         y          = bounds.y + bounds.height + fm2.getAscent();
        String      modified   = String.format("Modified %s", Numbers.formatDateTime(Numbers.DATE_AT_TIME_FORMAT, mCharacter.getModifiedOn()));
        String      left;
        String      right;

        if ((pageNumber & 1) == 1) {
            left = GCS.COPYRIGHT_FOOTER;
            right = modified;
        } else {
            left = modified;
            right = GCS.COPYRIGHT_FOOTER;
        }

        Font savedFont = gc.getFont();
        gc.setColor(ThemeColor.ON_PAGE);
        gc.setFont(font1);
        gc.drawString(left, bounds.x, y);
        gc.drawString(right, bounds.x + bounds.width - (int) fm1.getStringBounds(right, gc).getWidth(), y);
        gc.setFont(font2);
        Profile profile = mCharacter.getProfile();
        String  center  = mCharacter.getSettings().useTitleInFooter() ? profile.getTitle() : profile.getName();
        gc.drawString(center, bounds.x + (bounds.width - (int) fm2.getStringBounds(center, gc).getWidth()) / 2, y);

        String allRightsReserved = I18n.Text("All rights reserved");
        if ((pageNumber & 1) == 1) {
            left = allRightsReserved;
            right = pageString;
        } else {
            left = pageString;
            right = allRightsReserved;
        }

        y += fm2.getDescent() + fm1.getAscent();

        gc.setFont(font1);
        gc.drawString(left, bounds.x, y);
        gc.drawString(right, bounds.x + bounds.width - (int) fm1.getStringBounds(right, gc).getWidth(), y);
        // Trim off the leading URI scheme and authority path component. (http://, https://, ...)
        String webSite = SCHEME_PATTERN.matcher(GCS.WEB_SITE).replaceAll("");
        gc.drawString(webSite, bounds.x + (bounds.width - (int) fm1.getStringBounds(webSite, gc).getWidth()) / 2, y);

        gc.setFont(savedFont);
    }

    @Override
    public Insets getPageAdornmentsInsets(Page page) {
        FontMetrics fm1 = Fonts.getFontMetrics(UIManager.getFont(Fonts.KEY_FOOTER_SECONDARY));
        FontMetrics fm2 = Fonts.getFontMetrics(UIManager.getFont(Fonts.KEY_FOOTER_PRIMARY));
        return new Insets(0, 0, fm1.getAscent() + fm1.getDescent() + fm2.getAscent() + fm2.getDescent(), 0);
    }

    @Override
    public PrintManager getPageSettings() {
        return mCharacter.getPageSettings();
    }

    /** @return The character being displayed. */
    public GURPSCharacter getCharacter() {
        return mCharacter;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (Outline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(command)) {
            mRootsToSync.add(((Outline) event.getSource()).getRealOutline());
            markForRebuild();
        }
    }

    /** Marks the sheet for a rebuild in the near future. */
    public void markForRebuild() {
        if (!mRebuildPending) {
            mRebuildPending = true;
            EventQueue.invokeLater(this);
        }
    }

    @Override
    public void run() {
        syncRoots();
        rebuild();
        mRebuildPending = false;
    }

    private void syncRoots() {
        if (mRootsToSync.contains(mReactionsOutline) || mRootsToSync.contains(getEquipmentOutline()) || mRootsToSync.contains(getAdvantageOutline())) {
            OutlineModel outlineModel = mReactionsOutline.getModel();
            String       sortConfig   = outlineModel.getSortConfig();
            outlineModel.removeAllRows();
            for (ReactionRow row : collectReactions()) {
                outlineModel.addRow(row);
            }
            outlineModel.applySortConfig(sortConfig);
        }
        if (mSyncWeapons || mRootsToSync.contains(getEquipmentOutline()) || mRootsToSync.contains(getAdvantageOutline()) || mRootsToSync.contains(getSpellOutline()) || mRootsToSync.contains(getSkillOutline())) {
            OutlineModel outlineModel = mMeleeWeaponOutline.getModel();
            String       sortConfig   = outlineModel.getSortConfig();

            outlineModel.removeAllRows();
            for (WeaponDisplayRow row : collectWeapons(MeleeWeaponStats.class)) {
                outlineModel.addRow(row);
            }
            outlineModel.applySortConfig(sortConfig);

            outlineModel = mRangedWeaponOutline.getModel();
            sortConfig = outlineModel.getSortConfig();

            outlineModel.removeAllRows();
            for (WeaponDisplayRow row : collectWeapons(RangedWeaponStats.class)) {
                outlineModel.addRow(row);
            }
            outlineModel.applySortConfig(sortConfig);
        }
        mSyncWeapons = false;
        mRootsToSync.clear();
    }

    @Override
    public void stateChanged(ChangeEvent event) {
        Dimension size = getLayout().preferredLayoutSize(this);
        if (!getSize().equals(size)) {
            invalidate();
            repaint();
            setSize(size);
        }
    }

    private static PageFormat createDefaultPageFormat() {
        Paper      paper  = new Paper();
        PageFormat format = new PageFormat();
        format.setOrientation(PageFormat.PORTRAIT);
        paper.setSize(8.5 * 72.0, 11.0 * 72.0);
        paper.setImageableArea(0.25 * 72.0, 0.25 * 72.0, 8 * 72.0, 10.5 * 72.0);
        format.setPaper(paper);
        return format;
    }

    private Set<Row> expandAllContainers() {
        Set<Row> changed = new HashSet<>();
        expandAllContainers(mCharacter.getAdvantagesIterator(true), changed);
        expandAllContainers(mCharacter.getSkillsIterator(), changed);
        expandAllContainers(mCharacter.getSpellsIterator(), changed);
        expandAllContainers(mCharacter.getEquipmentIterator(), changed);
        expandAllContainers(mCharacter.getOtherEquipmentIterator(), changed);
        if (mRebuildPending) {
            syncRoots();
            rebuild();
        }
        return changed;
    }

    private static void expandAllContainers(RowIterator<? extends Row> iterator, Set<Row> changed) {
        for (Row row : iterator) {
            if (!row.isOpen()) {
                row.setOpen(true);
                changed.add(row);
            }
        }
    }

    private void closeContainers(Set<Row> rows) {
        for (Row row : rows) {
            row.setOpen(false);
        }
        if (mRebuildPending) {
            syncRoots();
            rebuild();
        }
    }

    /**
     * @param path The path to save to.
     * @return {@code true} on success.
     */
    public boolean saveAsPDF(Path path) {
        Set<Row> changed = expandAllContainers();
        try {
            PrintManager settings = mCharacter.getPageSettings();
            PageFormat   format   = settings != null ? settings.createPageFormat() : createDefaultPageFormat();
            float        width    = (float) format.getWidth();
            float        height   = (float) format.getHeight();

            adjustToPageSetupChanges(true);
            setPrinting(true);

            Document pdfDoc = new Document(new com.lowagie.text.Rectangle(width, height));
            try (OutputStream out = Files.newOutputStream(path)) {
                PdfWriter      writer  = PdfWriter.getInstance(pdfDoc, out);
                int            pageNum = 0;
                PdfContentByte cb;

                pdfDoc.open();
                cb = writer.getDirectContent();
                while (true) {
                    PdfTemplate template = cb.createTemplate(width, height);
                    Graphics2D  g2d      = template.createGraphics(width, height, new DefaultFontMapper());

                    if (print(g2d, format, pageNum) == NO_SUCH_PAGE) {
                        g2d.dispose();
                        break;
                    }
                    if (pageNum != 0) {
                        pdfDoc.newPage();
                    }
                    g2d.setClip(0, 0, (int) width, (int) height);
                    print(g2d, format, pageNum++);
                    g2d.dispose();
                    cb.addTemplate(template, 0, 0);
                }
                pdfDoc.close();
            }
            return true;
        } catch (Exception exception) {
            Log.error(exception);
            return false;
        } finally {
            setPrinting(false);
            closeContainers(changed);
        }
    }

    /**
     * @param path         The path to save to.
     * @param createdPaths The paths that were created.
     * @return {@code true} on success.
     */
    public boolean saveAsPNG(Path path, List<Path> createdPaths) {
        Set<Row> changed = expandAllContainers();
        try {
            int          dpi      = Preferences.getInstance().getPNGResolution();
            PrintManager settings = mCharacter.getPageSettings();
            PageFormat   format   = settings != null ? settings.createPageFormat() : createDefaultPageFormat();
            int          width    = (int) (format.getWidth() / 72.0 * dpi);
            int          height   = (int) (format.getHeight() / 72.0 * dpi);
            Img          buffer   = Img.create(width, height, Transparency.OPAQUE);
            int          pageNum  = 0;
            String       name     = PathUtils.getLeafName(path, false);

            path = path.getParent();

            adjustToPageSetupChanges(true);
            setPrinting(true);

            while (true) {
                Graphics2D gc = buffer.getGraphics();
                if (print(gc, format, pageNum) == NO_SUCH_PAGE) {
                    gc.dispose();
                    break;
                }
                gc.setClip(0, 0, width, height);
                gc.setBackground(Color.WHITE);
                gc.clearRect(0, 0, width, height);
                gc.scale(dpi / 72.0, dpi / 72.0);
                print(gc, format, pageNum++);
                gc.dispose();
                Path pngPath = path.resolve(PathUtils.enforceExtension(name + (pageNum > 1 ? " " + pageNum : ""), FileType.PNG.getExtension()));
                ImageIO.write(buffer, "png", pngPath.toFile());
                createdPaths.add(pngPath);
            }
            return true;
        } catch (Exception exception) {
            Log.error(exception);
            return false;
        } finally {
            setPrinting(false);
            closeContainers(changed);
        }
    }

    @Override
    public PrintManager getPrintManager() {
        return mCharacter.getPageSettings();
    }

    @Override
    public String getPrintJobTitle() {
        Dockable dockable = UIUtilities.getAncestorOfType(this, Dockable.class);
        if (dockable != null) {
            return dockable.getTitle();
        }
        Frame frame = UIUtilities.getAncestorOfType(this, Frame.class);
        if (frame != null) {
            return frame.getTitle();
        }
        return mCharacter.getProfile().getName();
    }

    @Override
    public void adjustToPageSetupChanges(boolean willPrint) {
        PrintManager pm = getPrintManager();
        Preferences.getInstance().setDefaultPageSettings(pm);
        if (!mCharacter.getLastPageSettingsAsString().equals(pm.toString())) {
            mCharacter.setModified(true);
        }
        if (willPrint) {
            mSavedScale = getScale();
            setScale(Scales.ACTUAL_SIZE.getScale());
            mOkToPaint = false;
        }
        rebuild();
    }

    @Override
    public boolean isPrinting() {
        return mIsPrinting;
    }

    @Override
    public void setPrinting(boolean printing) {
        mIsPrinting = printing;
        if (!printing) {
            mOkToPaint = true;
            if (mSavedScale != null && mSavedScale.getScale() != getScale().getScale()) {
                setScale(mSavedScale);
                rebuild();
            } else {
                repaint();
            }
        }
    }
}
