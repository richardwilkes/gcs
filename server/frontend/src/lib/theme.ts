// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

import { get, writable } from 'svelte/store';

const themeKey = 'theme';

/** The kind of theme. */
export enum Theme {
	System = 'system',
	Light = 'light',
	Dark = 'dark'
}

/** The current theme. */
export const currentTheme = writable<Theme>((localStorage.getItem(themeKey) as Theme) || Theme.System);

const systemIsDark = window.matchMedia('(prefers-color-scheme: dark)');
systemIsDark.addEventListener('change',
	() => document.documentElement.setAttribute(themeKey, resolvedThemeKind(get(currentTheme))));

currentTheme.subscribe((current) => {
	localStorage.setItem(themeKey, current);
	document.documentElement.setAttribute(themeKey, resolvedThemeKind(current));
});

/** Returns the resolved ThemeKind, either ThemeKind.Light or ThemeKind.Dark. */
export function resolvedThemeKind(theme: Theme) {
	return theme === Theme.System ? (systemIsDark.matches ? Theme.Dark : Theme.Light) : theme;
}

/** Returns true if the theme is dark. */
export function themeIsDark(theme: Theme) {
	return theme === Theme.System ? systemIsDark.matches : theme === Theme.Dark;
}
