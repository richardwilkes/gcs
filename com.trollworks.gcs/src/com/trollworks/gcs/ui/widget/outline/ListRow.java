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

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.datafile.Updatable;
import com.trollworks.gcs.feature.AttributeBonus;
import com.trollworks.gcs.feature.ConditionalModifier;
import com.trollworks.gcs.feature.ContainedWeightReduction;
import com.trollworks.gcs.feature.CostReduction;
import com.trollworks.gcs.feature.DRBonus;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.feature.ReactionBonus;
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.gcs.feature.SkillPointBonus;
import com.trollworks.gcs.feature.SpellBonus;
import com.trollworks.gcs.feature.SpellPointBonus;
import com.trollworks.gcs.feature.WeaponBonus;
import com.trollworks.gcs.prereq.PrereqList;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.utility.FilteredList;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.VersionException;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Base64;
import java.util.Collection;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeSet;
import java.util.UUID;

/** A common row super-class for the model. */
public abstract class ListRow extends Row implements Updatable {
    private static final String             ATTRIBUTE_OPEN    = "open";
    private static final String             TAG_NOTES         = "notes";
    private static final String             TAG_CATEGORIES    = "categories";
    private static final String             KEY_ID            = "id";
    private static final String             KEY_BASED_ON_ID   = "based_on_id";
    private static final String             KEY_BASED_ON_HASH = "based_on_hash";
    private static final String             KEY_FEATURES      = "features";
    private static final String             KEY_DEFAULTS      = "defaults";
    private static final String             KEY_CHILDREN      = "children";
    private static final String             KEY_PREREQS       = "prereqs";
    /** The data file the row is associated with. */
    protected            DataFile           mDataFile;
    private              UUID               mID;
    private              UUID               mBasedOnID;
    private              String             mBasedOnHash;
    private              List<Feature>      mFeatures;
    private              PrereqList         mPrereqList;
    private              List<SkillDefault> mDefaults;
    private              boolean            mIsSatisfied;
    private              String             mUnsatisfiedReason;
    private              String             mNotes;
    private              TreeSet<String>    mCategories;

    public static void saveList(JsonWriter w, String key, List<?> list, SaveType saveType) throws IOException {
        FilteredList<ListRow> rows = new FilteredList<>(list, ListRow.class, true);
        if (!rows.isEmpty()) {
            w.key(key);
            w.startArray();
            for (ListRow row : rows) {
                row.save(w, saveType);
            }
            w.endArray();
        }
    }

    /**
     * Extracts any "nameable" portions of the buffer and puts their keys into the provided set.
     *
     * @param set    The set to add the nameable keys to.
     * @param buffer The text to check for nameable portions.
     */
    public static void extractNameables(Set<String> set, String buffer) {
        int first = buffer.indexOf('@');
        int last  = buffer.indexOf('@', first + 1);

        while (first != -1 && last != -1) {
            set.add(buffer.substring(first + 1, last));
            first = buffer.indexOf('@', last + 1);
            last = buffer.indexOf('@', first + 1);
        }
    }

    /**
     * Names any "nameable" portions of the data and returns the resulting string.
     *
     * @param map  The map of nameable keys to names.
     * @param data The data to change.
     * @return The revised string.
     */
    public static String nameNameables(Map<String, String> map, String data) {
        int           first  = data.indexOf('@');
        int           last   = data.indexOf('@', first + 1);
        StringBuilder buffer = new StringBuilder();

        while (first != -1 && last != -1) {
            String key         = data.substring(first + 1, last);
            String replacement = map.get(key);

            if (first != 0) {
                buffer.append(data, 0, first);
            }
            if (replacement != null) {
                buffer.append(replacement);
            } else {
                buffer.append('@');
                buffer.append(key);
                buffer.append('@');
            }
            data = last + 1 == data.length() ? "" : data.substring(last + 1);
            first = data.indexOf('@');
            last = data.indexOf('@', first + 1);
        }
        buffer.append(data);
        return buffer.toString();
    }

    public static Set<String> createCategoriesList(String categories) {
        return new TreeSet<>(createList(categories));
    }

    // This is the decompose method that works with the compose method (getCategoriesAsString())
    private static Collection<String> createList(String categories) {
        return Arrays.asList(categories.split(","));
    }

    /**
     * Creates a new data row.
     *
     * @param dataFile    The data file to associate it with.
     * @param isContainer Whether or not this row allows children.
     */
    public ListRow(DataFile dataFile, boolean isContainer) {
        setCanHaveChildren(isContainer);
        setOpen(isContainer);
        mDataFile = dataFile;
        mID = UUID.randomUUID();
        mFeatures = new ArrayList<>();
        mPrereqList = new PrereqList(null, true);
        mDefaults = new ArrayList<>();
        mIsSatisfied = true;
        mNotes = "";
        mCategories = new TreeSet<>();
    }

    /**
     * Creates a clone of an existing data row and associates it with the specified data file.
     *
     * @param dataFile   The data file to associate it with.
     * @param rowToClone The data row to clone.
     */
    public ListRow(DataFile dataFile, ListRow rowToClone) {
        this(dataFile, rowToClone.canHaveChildren());
        setOpen(rowToClone.isOpen());
        mNotes = rowToClone.mNotes;
        for (Feature feature : rowToClone.mFeatures) {
            mFeatures.add(feature.cloneFeature());
        }
        mPrereqList = new PrereqList(null, rowToClone.getPrereqs());
        mDefaults = new ArrayList<>();
        for (SkillDefault skillDefault : rowToClone.mDefaults) {
            mDefaults.add(new SkillDefault(skillDefault));
        }
        mCategories = new TreeSet<>(rowToClone.mCategories);
        try {
            MessageDigest         digest = MessageDigest.getInstance("SHA3-256");
            ByteArrayOutputStream baos   = new ByteArrayOutputStream();
            try (JsonWriter w = new JsonWriter(new OutputStreamWriter(baos, StandardCharsets.UTF_8), "")) {
                rowToClone.save(w, SaveType.HASH);
            }
            mBasedOnHash = Base64.getEncoder().withoutPadding().encodeToString(digest.digest(baos.toByteArray()));
            mBasedOnID = rowToClone.mID;
        } catch (Exception exception) {
            mBasedOnID = null;
            mBasedOnHash = null;
            Log.warn(exception);
        }
    }

    @Override
    public UUID getID() {
        return mID;
    }

    /**
     * @param obj The other object to compare against.
     * @return Whether or not this {@link ListRow} is equivalent.
     */
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof ListRow) {
            ListRow row = (ListRow) obj;
            if (mNotes.equals(row.mNotes) && mCategories.equals(row.mCategories)) {
                if (mDefaults.equals(row.mDefaults)) {
                    if (mPrereqList.equals(row.mPrereqList)) {
                        if (mFeatures.equals(row.mFeatures)) {
                            int childCount = getChildCount();
                            if (childCount == row.getChildCount()) {
                                for (int i = 0; i < childCount; i++) {
                                    if (!((ListRow) getChild(i)).isEquivalentTo(row.getChild(i))) {
                                        return false;
                                    }
                                }
                                return true;
                            }
                        }
                    }
                }
            }
        }
        return false;
    }

    /** @return Creates a detailed editor for this row. */
    public abstract RowEditor<? extends ListRow> createEditor();

    /** @return The localized name for this row object. */
    public abstract String getLocalizedName();

    @Override
    public boolean addChild(Row row) {
        boolean result = super.addChild(row);
        if (result) {
            notifySingle(getListChangedID());
        }
        return result;
    }

    /** @return The ID for the "list changed" notification. */
    public abstract String getListChangedID();

    /** @return The most recent version of the JSON data this object knows how to load. */
    public abstract int getJSONVersion();

    /** @return The type name to use for this data. */
    public abstract String getJSONTypeName();

    /** @return The type of row. */
    public abstract String getRowType();

    /** @return Whether or not this row's prerequisites are currently satisfied. */
    public boolean isSatisfied() {
        return mIsSatisfied;
    }

    /** @param satisfied Whether or not this row's prerequisites are currently satisfied. */
    public void setSatisfied(boolean satisfied) {
        mIsSatisfied = satisfied;
        if (satisfied) {
            mUnsatisfiedReason = null;
        }
    }

    /** @return The reason {@link #isSatisfied()} is returning {@code false}. */
    public String getReasonForUnsatisfied() {
        return mUnsatisfiedReason;
    }

    /** @param reason The reason {@link #isSatisfied()} is returning {@code false}. */
    public void setReasonForUnsatisfied(String reason) {
        mUnsatisfiedReason = reason;
    }

    /**
     * Loads this row's contents.
     *
     * @param m     The {@link JsonMap} to load data from.
     * @param state The {@link LoadState} to use.
     */
    public final void load(JsonMap m, LoadState state) throws IOException {
        if (m.has(KEY_ID)) {
            try {
                mID = UUID.fromString(m.getString(KEY_ID));
            } catch (Exception exception) {
                mID = UUID.randomUUID();
            }
        }
        if (m.has(KEY_BASED_ON_ID)) {
            try {
                mBasedOnID = UUID.fromString(m.getString(KEY_BASED_ON_ID));
                mBasedOnHash = m.getString(KEY_BASED_ON_HASH);
            } catch (Exception exception) {
                mBasedOnID = null;
                mBasedOnHash = null;
            }
        }
        state.mDataItemVersion = m.getInt(LoadState.ATTRIBUTE_VERSION);
        if (state.mDataItemVersion > getJSONVersion()) {
            throw VersionException.createTooNew();
        }
        boolean isContainer = m.getString(DataFile.KEY_TYPE).endsWith("_container");
        setCanHaveChildren(isContainer);
        setOpen(isContainer);
        prepareForLoad(state);
        loadSelf(m, state);
        if (m.has(KEY_PREREQS)) {
            mPrereqList = new PrereqList(null, mDataFile.defaultWeightUnits(), m.getMap(KEY_PREREQS));
        }
        if (!(this instanceof Technique) && m.has(KEY_DEFAULTS)) {
            JsonArray a     = m.getArray(KEY_DEFAULTS);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                mDefaults.add(new SkillDefault(a.getMap(i), false));
            }
        }
        if (m.has(KEY_FEATURES)) {
            JsonArray a     = m.getArray(KEY_FEATURES);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                JsonMap m1   = a.getMap(i);
                String  type = m1.getString(DataFile.KEY_TYPE);
                switch (type) {
                case AttributeBonus.TAG_ROOT -> mFeatures.add(new AttributeBonus(m1));
                case DRBonus.TAG_ROOT -> mFeatures.add(new DRBonus(m1));
                case ReactionBonus.TAG_ROOT -> mFeatures.add(new ReactionBonus(m1));
                case ConditionalModifier.TAG_ROOT -> mFeatures.add(new ConditionalModifier(m1));
                case SkillBonus.TAG_ROOT -> mFeatures.add(new SkillBonus(m1));
                case SkillPointBonus.TAG_ROOT -> mFeatures.add(new SkillPointBonus(m1));
                case SpellBonus.TAG_ROOT -> mFeatures.add(new SpellBonus(m1));
                case SpellPointBonus.TAG_ROOT -> mFeatures.add(new SpellPointBonus(m1));
                case WeaponBonus.TAG_ROOT -> mFeatures.add(new WeaponBonus(m1));
                case CostReduction.TAG_ROOT -> mFeatures.add(new CostReduction(m1));
                case ContainedWeightReduction.TAG_ROOT -> mFeatures.add(new ContainedWeightReduction(m1));
                default -> Log.warn("unknown feature type: " + type);
                }
            }
        }
        mNotes = m.getString(TAG_NOTES);
        if (m.has(TAG_CATEGORIES)) {
            JsonArray a     = m.getArray(TAG_CATEGORIES);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                mCategories.add(a.getString(i));
            }
        }
        if (canHaveChildren()) {
            setOpen(m.getBoolean(ATTRIBUTE_OPEN));
            if (m.has(KEY_CHILDREN)) {
                JsonArray a     = m.getArray(KEY_CHILDREN);
                int       count = a.size();
                for (int i = 0; i < count; i++) {
                    loadChild(a.getMap(i), state);
                }
            }
        }
        finishedLoading(state);
    }

    protected abstract void loadSelf(JsonMap m, LoadState state) throws IOException;

    protected abstract void loadChild(JsonMap m, LoadState state) throws IOException;

    /**
     * Called to prepare the row for loading.
     *
     * @param state The {@link LoadState} to use.
     */
    protected void prepareForLoad(LoadState state) {
        mNotes = "";
        mFeatures.clear();
        mDefaults.clear();
        mPrereqList = new PrereqList(null, true);
        mCategories.clear();
    }

    /**
     * Called when loading of this row is complete. Does nothing by default.
     *
     * @param state The {@link LoadState} to use.
     */
    protected void finishedLoading(LoadState state) {
        // Nothing to do.
    }

    /**
     * Saves the row.
     *
     * @param w        The {@link JsonWriter} to use.
     * @param saveType The type of save being performed.
     */
    public void save(JsonWriter w, SaveType saveType) throws IOException {
        w.startMap();
        w.keyValue(DataFile.KEY_TYPE, getJSONTypeName());
        w.keyValue(LoadState.ATTRIBUTE_VERSION, getJSONVersion());
        w.keyValue(KEY_ID, mID.toString());
        if (mBasedOnID != null) {
            w.keyValue(KEY_BASED_ON_ID, mBasedOnID.toString());
            w.keyValue(KEY_BASED_ON_HASH, mBasedOnHash);
        }
        saveSelf(w, saveType);
        if (!mPrereqList.isEmpty()) {
            w.key(KEY_PREREQS);
            mPrereqList.save(w);
        }
        if (!(this instanceof Technique) && !mDefaults.isEmpty()) {
            w.key(KEY_DEFAULTS);
            w.startArray();
            for (SkillDefault skillDefault : mDefaults) {
                skillDefault.save(w, false);
            }
            w.endArray();
        }
        if (!mFeatures.isEmpty()) {
            w.key(KEY_FEATURES);
            w.startArray();
            for (Feature feature : mFeatures) {
                feature.save(w);
            }
            w.endArray();
        }
        w.keyValueNot(TAG_NOTES, mNotes, "");
        if (!mCategories.isEmpty()) {
            w.key(TAG_CATEGORIES);
            w.startArray();
            for (String category : mCategories) {
                w.value(category);
            }
            w.endArray();
        }
        if (canHaveChildren()) {
            if (saveType != SaveType.HASH) {
                w.keyValue(ATTRIBUTE_OPEN, isOpen());
            }
            if (saveType != SaveType.UNDO) {
                saveList(w, KEY_CHILDREN, getChildren(), saveType);
            }
        }
        w.endMap();
    }

    /**
     * Saves the row.
     *
     * @param w        The {@link JsonWriter} to use.
     * @param saveType The type of save being performed.
     */
    protected abstract void saveSelf(JsonWriter w, SaveType saveType) throws IOException;

    /**
     * Starts the notification process. Should be called before calling {@link #notify(String,
     * Object)}.
     */
    protected final void startNotify() {
        if (mDataFile != null) {
            mDataFile.startNotify();
        }
    }

    /**
     * Sends a notification to all interested consumers.
     *
     * @param type The notification type.
     * @param data Extra data specific to this notification.
     */
    public void notify(String type, Object data) {
        if (mDataFile != null) {
            mDataFile.notify(type, this);
        }
    }

    /**
     * Sends a notification to all interested consumers.
     *
     * @param type The notification type.
     */
    public final void notifySingle(String type) {
        if (mDataFile != null) {
            mDataFile.notifySingle(type, this);
        }
    }

    /**
     * Ends the notification process. Must be called after calling {@link #notify(String, Object)}.
     */
    public void endNotify() {
        if (mDataFile != null) {
            mDataFile.endNotify();
        }
    }

    /** Called to update any information that relies on children. */
    public void update() {
        // Do nothing by default.
    }

    /** @return The owning data file. */
    public DataFile getDataFile() {
        return mDataFile;
    }

    /** @return The owning template. */
    public Template getTemplate() {
        return mDataFile instanceof Template ? (Template) mDataFile : null;
    }

    /** @return The "secondary" text, the text display below an Advantage. */
    protected String getSecondaryText() {
        StringBuilder builder = new StringBuilder();
        DataFile      df      = getDataFile();
        if (df.modifiersDisplay().inline()) {
            String txt = getModifierNotes();
            if (!txt.isBlank()) {
                builder.append(txt);
            }
        }
        if (df.notesDisplay().inline()) {
            String txt = getNotes();
            if (!txt.isBlank()) {
                if (!builder.isEmpty()) {
                    builder.append('\n');
                }
                builder.append(txt);
            }
        }
        return builder.toString();
    }

    /** @return The owning character. */
    public GURPSCharacter getCharacter() {
        return mDataFile instanceof GURPSCharacter ? (GURPSCharacter) mDataFile : null;
    }

    /** @return The features provided by this data row. */
    public List<Feature> getFeatures() {
        return Collections.unmodifiableList(mFeatures);
    }

    /**
     * @param features The new features of this data row.
     * @return Whether there was a change or not.
     */
    public boolean setFeatures(List<Feature> features) {
        if (!mFeatures.equals(features)) {
            mFeatures = new ArrayList<>(features);
            return true;
        }
        return false;
    }

    /** @return The categories this data row belongs to. */
    public Set<String> getCategories() {
        return Collections.unmodifiableSet(mCategories);
    }

    /** @return The categories this data row belongs to. */
    public String getCategoriesAsString() {
        StringBuilder buffer = new StringBuilder();
        for (String category : mCategories) {
            if (!buffer.isEmpty()) {
                buffer.append(",");
                buffer.append(" ");
            }
            buffer.append(category);
        }
        return buffer.toString();
    }

    /*
     * Does this belong to a category?   Added the ability to check for compound
     * categories like "Money: US", "Concoctions:Potions" using "Money" or "Concoctions".
     */
    public boolean hasCategory(String cat) {
        for (String category : mCategories) {
            int indexOfColon = category.indexOf(':');
            if (indexOfColon > 0) {
                if (category.substring(0, indexOfColon).equalsIgnoreCase(cat)) {
                    return true;
                }
            }
            if (category.equalsIgnoreCase(cat)) {
                return true;
            }
        }
        return false;
    }

    /**
     * @param categories The categories this data row belongs to.
     * @return Whether there was a change or not.
     */
    public boolean setCategories(Collection<String> categories) {
        Set<String> old = mCategories;
        mCategories = new TreeSet<>();
        for (String category : categories) {
            category = category.trim();
            if (!category.isEmpty()) {
                mCategories.add(category);
            }
        }
        if (!old.equals(mCategories)) {
            String id = getCategoryID();
            if (id != null) {
                notifySingle(id);
            }
            return true;
        }
        return false;
    }

    /**
     * @param categories The categories this data row belongs to. Use commas to separate
     *                   categories.
     * @return Whether there was a change or not.
     */
    public final boolean setCategories(String categories) {
        return setCategories(createList(categories));
    }

    /**
     * @param category The category to add.
     * @return Whether there was a change or not.
     */
    public boolean addCategory(String category) {
        category = category.trim();
        if (!category.isEmpty()) {
            if (mCategories.add(category)) {
                String id = getCategoryID();
                if (id != null) {
                    notifySingle(id);
                }
                return true;
            }
        }
        return false;
    }

    /** @return The notification ID to use with categories. */
    @SuppressWarnings("static-method")
    protected String getCategoryID() {
        return null;
    }

    /** @return The prerequisites needed by this data row. */
    public PrereqList getPrereqs() {
        return mPrereqList;
    }

    /**
     * @param prereqs The new prerequisites needed by this data row.
     * @return Whether there was a change or not.
     */
    public boolean setPrereqs(PrereqList prereqs) {
        if (!mPrereqList.equals(prereqs)) {
            mPrereqList = (PrereqList) prereqs.clone(null);
            return true;
        }
        return false;
    }

    /** @return The defaults for this row. */
    public List<SkillDefault> getDefaults() {
        return Collections.unmodifiableList(mDefaults);
    }

    /**
     * @param defaults The new defaults for this row.
     * @return Whether there was a change or not.
     */
    public boolean setDefaults(List<SkillDefault> defaults) {
        if (!mDefaults.equals(defaults)) {
            mDefaults = new ArrayList<>(defaults);
            return true;
        }
        return false;
    }

    @Override
    public final void setData(Column column, Object data) {
        // Not used.
    }

    /**
     * @param text          The text to search for.
     * @param lowerCaseOnly The passed in text is all lowercase.
     * @return {@code true} if this row contains the text.
     */
    @SuppressWarnings("static-method")
    public boolean contains(String text, boolean lowerCaseOnly) {
        return false;
    }

    /**
     * @param marker Whether to return the marker or file image.
     * @return An image representative of this row.
     */
    public abstract RetinaIcon getIcon(boolean marker);

    /** @param set The nameable keys. */
    public void fillWithNameableKeys(Set<String> set) {
        extractNameables(set, mNotes);
        for (SkillDefault def : mDefaults) {
            def.fillWithNameableKeys(set);
        }
        for (Feature feature : mFeatures) {
            feature.fillWithNameableKeys(set);
        }
        mPrereqList.fillWithNameableKeys(set);
    }

    /** @param map The map of nameable keys to names to apply. */
    public void applyNameableKeys(Map<String, String> map) {
        mNotes = nameNameables(map, mNotes);
        for (SkillDefault def : mDefaults) {
            def.applyNameableKeys(map);
        }
        for (Feature feature : mFeatures) {
            feature.applyNameableKeys(map);
        }
        mPrereqList.applyNameableKeys(map);
    }

    /** @return The notes. */
    public String getNotes() {
        return mNotes;
    }

    /** @return The notes due to modifiers. */
    @SuppressWarnings("static-method")
    public String getModifierNotes() {
        return "";
    }

    /**
     * @param notes The notes to set.
     * @return Whether it was changed.
     */
    public boolean setNotes(String notes) {
        if (!mNotes.equals(notes)) {
            mNotes = notes;
            return true;
        }
        return false;
    }

    @Override
    public void getContainedUpdatables(Map<UUID, Updatable> updatables) {
        if (canHaveChildren()) {
            for (Row one : getChildren()) {
                if (one instanceof Updatable) {
                    Updatable u = (Updatable) one;
                    updatables.put(u.getID(), u);
                    u.getContainedUpdatables(updatables);
                }
            }
        }
    }
}
