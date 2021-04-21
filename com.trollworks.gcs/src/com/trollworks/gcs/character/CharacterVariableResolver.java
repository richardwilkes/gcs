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
import com.trollworks.gcs.expression.VariableResolver;
import com.trollworks.gcs.pointpool.PointPool;
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
            return "";
        }
        String[] parts = variableName.split("\\.", 2);
        if (parts.length == 2) {
            switch (parts[0] + ".") {
            case Attribute.ID_ATTR_PREFIX:
                Attribute attr = mCharacter.getAttributes().get(parts[1]);
                if (attr != null) {
                    return String.valueOf(attr.getDoubleValue(mCharacter));
                }
                break;
            case PointPool.ID_POOL_PREFIX:
                parts = parts[1].split("\\.", 2);
                PointPool pool = mCharacter.getPointPools().get(parts[0]);
                if (pool != null) {
                    switch (parts.length) {
                    case 1:
                        return String.valueOf(pool.getMaximum(mCharacter));
                    case 2:
                        switch (parts[1]) {
                        case "current":
                            return String.valueOf(pool.getCurrent(mCharacter));
                        case "maximum":
                            return String.valueOf(pool.getMaximum(mCharacter));
                        default:
                            break;
                        }
                        break;
                    default:
                        break;
                    }
                }
                break;
            default:
                break;
            }
        }
        Log.error("unresolved variable: $" + variableName);
        return null;
    }
}
