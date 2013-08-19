/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.model.modifier;

import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.gcs.ui.modifiers.CSModifierColumnID;
import com.trollworks.gcs.ui.modifiers.CSModifierEditor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.widget.outline.TKColumn;

import java.awt.image.BufferedImage;
import java.io.IOException;

/** Model for trait modifiers */
public class CMModifier extends CMRow {
	/** tag name for root */
	public static final String		TAG_MODIFIER		= "modifier";						//$NON-NLS-1$
	/** tag name for name */
	protected static final String	TAG_NAME			= "name";							//$NON-NLS-1$
	/** id for cost modifier base change notification */
	public static final String		TAG_COST			= "cost";							//$NON-NLS-1$
	/** id for cost modifier base change notification */
	public static final String		TAG_LEVELS			= "levels";						//$NON-NLS-1$
	/** tag name for reference */
	protected static final String	TAG_REFERENCE		= "reference";						//$NON-NLS-1$
	/** prefix for modifier notification id */
	public static final String		MODIFIER_PREFIX		= "modifier.";						//$NON-NLS-1$
	/** id for name change notification */
	public static final String		ID_NAME				= MODIFIER_PREFIX + "Name";		//$NON-NLS-1$
	/** id for reference change notification */
	public static final String		ID_REFERENCE		= MODIFIER_PREFIX + "Reference";	//$NON-NLS-1$
	/** id for cost modifier change notification */
	public static final String		ID_COST_MODIFIER	= MODIFIER_PREFIX + TAG_COST;
	/** id for list changed change notification */
	public static final String		ID_LIST_CHANGED		= MODIFIER_PREFIX + "ListChanged";	//$NON-NLS-1$
	private String					mName;
	private String					mReference;
	private int						mCost;
	private int						mLevels;

	/**
	 * Creates a new {@link CMModifier}.
	 * 
	 * @param file The {@link CMDataFile} to use.
	 * @param other Another {@link CMModifier} to clone.
	 */
	public CMModifier(CMDataFile file, CMModifier other) {
		super(file, other);
		mName = other.mName;
		mReference = other.mReference;
		mCost = other.mCost;
		mLevels = other.mLevels;
	}

	/**
	 * Creates a new {@link CMModifier}.
	 * 
	 * @param file The {@link CMDataFile} to use.
	 * @param reader The {@link TKXMLReader} to use.
	 * @throws IOException
	 */
	public CMModifier(CMDataFile file, TKXMLReader reader) throws IOException {
		super(file, false);
		load(reader, false);
	}

	/**
	 * Creates a new {@link CMModifier}.
	 * 
	 * @param file The {@link CMDataFile} to use.
	 */
	public CMModifier(CMDataFile file) {
		super(file, false);
		mName = Msgs.DEFAULT_NAME;
		mReference = ""; //$NON-NLS-1$
		mCost = 0;
		mLevels = 0;
	}

	/** @return The page reference. */
	public String getReference() {
		return mReference;
	}

	/**
	 * @param reference The new page reference.
	 * @return <code>true</code> if page reference has changed.
	 */
	public boolean setReference(String reference) {
		if (!mReference.equals(reference)) {
			mReference = reference;
			notifySingle(ID_REFERENCE);
			return true;
		}
		return false;
	}

	/** @return An exact clone of this modifier. */
	public CMModifier cloneModifier() {
		return new CMModifier(mDataFile, this);
	}

	/** @return The total cost modifier. */
	public int getCostModifier() {
		return (mLevels > 0) ? mCost * mLevels : mCost;
	}

	/** @return The cost. */
	public int getCost() {
		return mCost;
	}

	/**
	 * @param cost The value to set for cost modifier.
	 * @return Whether it was changed.
	 */
	public boolean setCost(int cost) {
		if (mCost != cost) {
			mCost = cost;
			notifySingle(ID_COST_MODIFIER);
			return true;
		}
		return false;
	}

	/** @return The levels. */
	public int getLevels() {
		return mLevels;
	}

	/**
	 * @param levels The value to set for cost modifier.
	 * @return Whether it was changed.
	 */
	public boolean setLevels(int levels) {
		if (levels < 0) {
			levels = 0;
		}
		if (mLevels != levels) {
			mLevels = levels;
			notifySingle(ID_COST_MODIFIER);
			return true;
		}
		return false;
	}

	/** @return <code>true</code> if this {@link CMModifier} has levels. */
	public boolean hasLevels() {
		return mLevels > 0;
	}

	@Override public boolean contains(String text, boolean lowerCaseOnly) {
		return getName().toLowerCase().indexOf(text) != -1;
	}

	@Override public CSRowEditor<CMModifier> createEditor() {
		return new CSModifierEditor(this);
	}

	@Override public BufferedImage getImage(boolean large) {
		return null;
	}

	@Override public String getListChangedID() {
		return ID_LIST_CHANGED;
	}

	@Override public String getLocalizedName() {
		return Msgs.DEFAULT_NAME;
	}

	@Override public String getRowType() {
		return Msgs.MODIFIER_TYPE;
	}

	@Override public String getXMLTagName() {
		return TAG_MODIFIER;
	}

	@Override protected void loadSubElement(TKXMLReader reader, boolean forUndo) throws IOException {
		String name = reader.getName();

		if (TAG_NAME.equals(name)) {
			mName = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
		} else if (TAG_REFERENCE.equals(name)) {
			mReference = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
		} else if (TAG_COST.equals(name)) {
			mCost = reader.readInteger(0);
		} else if (TAG_LEVELS.equals(name)) {
			mLevels = reader.readInteger(0);
		} else {
			super.loadSubElement(reader, forUndo);
		}
	}

	@Override protected void prepareForLoad(boolean forUndo) {
		super.prepareForLoad(forUndo);
		mName = Msgs.DEFAULT_NAME;
		mCost = 0;
		mLevels = 0;
		mReference = ""; //$NON-NLS-1$
	}

	@Override protected void saveSelf(TKXMLWriter out, boolean forUndo) {
		out.simpleTag(TAG_NAME, mName);
		out.simpleTag(TAG_COST, mCost);
		out.simpleTag(TAG_LEVELS, mLevels);
		out.simpleTagNotEmpty(TAG_REFERENCE, mReference);
	}

	@Override public Object getData(TKColumn column) {
		return CSModifierColumnID.values()[column.getID()].getData(this);
	}

	@Override public String getDataAsText(TKColumn column) {
		return CSModifierColumnID.values()[column.getID()].getDataAsText(this);
	}

	@Override public String toString() {
		StringBuilder builder = new StringBuilder();

		builder.append(getName());
		if (hasLevels()) {
			builder.append(' ');
			builder.append(getLevels());
		}
		return builder.toString();
	}

	/** @return The name. */
	public String getName() {
		return mName;
	}

	/**
	 * @param name The value to set for name.
	 * @return <code>true</code> if name has changed
	 */
	public boolean setName(String name) {
		if (!mName.equals(name)) {
			mName = name;
			notifySingle(ID_NAME);
			return true;
		}
		return false;
	}
}
