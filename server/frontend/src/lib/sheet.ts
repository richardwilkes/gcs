// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

import { get, writable } from 'svelte/store';
import { apiPrefix } from '$lib/dev.ts';
import { session } from '$lib/session.ts';

export const sheet = writable<Sheet | undefined>();

export async function fetchSheet(path: string): Promise<Sheet | undefined> {
	const rsp = await fetch(apiPrefix(`/sheet/${path}`), {
		method: 'GET',
		headers: { 'X-Session': get(session)?.ID ?? '' },
		cache: 'no-store',
	});
	if (!rsp.ok) {
		return undefined;
	}
	return await rsp.json();
}

export async function updateSheetField(
	path: string,
	kind: string,
	key: string,
	data: string
): Promise<Sheet | undefined> {
	const rsp = await fetch(apiPrefix(`/sheet/${path}`), {
		method: 'POST',
		headers: { 'X-Session': get(session)?.ID ?? '' },
		cache: 'no-store',
		body: JSON.stringify({
			Kind: kind,
			Key: key,
			Data: data,
		}),
	});
	if (!rsp.ok) {
		throw undefined;
	}
	return await rsp.json();
}

export async function saveSheet(path: string): Promise<Sheet | undefined> {
	const rsp = await fetch(apiPrefix(`/sheet/${path}`), {
		method: 'PUT',
		headers: { 'X-Session': get(session)?.ID ?? '' },
		cache: 'no-store',
	});
	if (!rsp.ok) {
		throw undefined;
	}
	return await rsp.json();
}

export interface Identity {
	Name: string;
	Title: string;
	Organization: string;
}

export interface Misc {
	Created: string;
	Modified: string;
	Player: string;
}

export interface Description {
	Gender: string;
	Age: string;
	Birthday: string;
	Religion: string;
	Height: string;
	Weight: string;
	SizeModifier: string;
	TechLevel: string;
	Hair: string;
	Eyes: string;
	Skin: string;
	Hand: string;
}

export interface Points {
	Total: string;
	Unspent: string;
	Ancestry: string;
	Attributes: string;
	Advantages: string;
	Disadvantages: string;
	Quirks: string;
	Skills: string;
	Spells: string;
}

export interface Attribute {
	Type: string;
	Key: string;
	Name: string;
	Value: string;
	Points: string;
}

export interface PointPool {
	Type: string;
	Key: string;
	Name: string;
	Value: string;
	Max: string;
	Points: string;
	State: string;
	Detail: string;
}

export interface BasicDamage {
	Thrust: string;
	Swing: string;
}

export interface HitLocation {
	Roll: string;
	Location: string;
	LocationDetail: string;
	HitPenalty: string;
	DR: string;
	DRDetail: string;
	SubLocations: HitLocation[];
}

export interface Body {
	Name: string;
	Locations: HitLocation[];
}

export interface Encumbrance {
	Current: number;
	MaxLoad: string[];
	Move: string[];
	Dodge: string[];
	Overloaded: boolean;
}

export interface LiftingAndMovingThings {
	BasicLift: string;
	OneHandedLift: string;
	TwoHandedLift: string;
	ShoveAndKnockOver: string;
	RunningShoveAndKnockOver: string;
	CarryOnBack: string;
	ShiftSlightly: string;
}

export interface Column {
	Title: string;
	Detail: string;
	TitleIsImageKey: boolean;
	Primary: boolean;
}

export interface Cell {
	Type: string;
	Disabled: boolean;
	Dim: boolean;
	Checked: boolean;
	Alignment: string;
	Primary: string;
	Secondary: string;
	Tooltip: string;
	UnsatisfiedReason: string;
	TemplateInfo: string;
	InlineTag: string;
}

export interface Row {
	ID: string;
	Depth: number;
	Cells: Cell[];
}

export interface Table {
	Columns: Column[];
	Rows: Row[];
}

export interface PageRef {
	Name: string;
	Offset: number;
}

export interface Sheet {
	Identity: Identity;
	Misc: Misc;
	Description: Description;
	Points: Points;
	PrimaryAttributes: Attribute[];
	SecondaryAttributes: Attribute[];
	PointPools: PointPool[];
	BasicDamage: BasicDamage;
	Body: Body;
	Encumbrance: Encumbrance;
	LiftingAndMovingThings: LiftingAndMovingThings;
	Reactions: Table | null;
	ConditionalModifiers: Table | null;
	MeleeWeapons: Table | null;
	RangedWeapons: Table | null;
	Traits: Table | null;
	Skills: Table | null;
	Spells: Table | null;
	CarriedEquipment: Table | null;
	OtherEquipment: Table | null;
	Notes: Table | null;
	Portrait: Table | null;
	PageRefs: { [k: string]: PageRef };
	Modified: boolean;
	ReadOnly: boolean;
}
