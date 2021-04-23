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

package com.trollworks.gcs.character;

import com.trollworks.gcs.attribute.Attribute;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.expression.VariableResolver;
import com.trollworks.gcs.utility.Log;

import java.util.HashSet;
import java.util.Set;

public class CharacterVariableResolver implements VariableResolver {
    private GURPSCharacter mCharacter;
    private Set<String>    mExclude;

    public CharacterVariableResolver(GURPSCharacter character) {
        mCharacter = character;
        mExclude = new HashSet<>();
    }

    public void addExclusion(String exclude) {
        mExclude.add(exclude);
    }

    public void removeExclusion(String exclusion) {
        mExclude.remove(exclusion);
    }

    @Override
    public String resolveVariable(String variableName) {
        if (mExclude.contains(variableName)) {
            Log.error("attempt to resolve variable via itself: $" + variableName);
            return "";
        }
        if ("sm".equals(variableName)) {
            return String.valueOf(mCharacter.getProfile().getSizeModifier());
        }
        String[]  parts = variableName.split("\\.", 2);
        Attribute attr  = mCharacter.getAttributes().get(parts[0]);
        if (attr == null) {
            Log.error("no such variable: $" + variableName);
            return "";
        }
        AttributeDef def = attr.getAttrDef(mCharacter);
        if (def == null) {
            Log.error("no such variable definition: $" + variableName);
            return "";
        }
        if (def.isPool() && parts.length > 1) {
            switch (parts[1]) {
            case "current":
                return String.valueOf(attr.getCurrentIntValue(mCharacter));
            case "maximum":
                return String.valueOf(attr.getIntValue(mCharacter));
            default:
                Log.error("no such variable: $" + variableName);
                return "";
            }
        }
        return String.valueOf(attr.getDoubleValue(mCharacter));
    }
}
