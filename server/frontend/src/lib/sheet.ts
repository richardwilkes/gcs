/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

import { get, writable } from 'svelte/store';
import { apiPrefix } from '$lib/dev.ts';
import { session } from '$lib/session.ts';

export const sheet = writable<Sheet | undefined>();

export async function fetchSheet(path: string): Promise<Sheet | undefined> {
	const rsp = await fetch(apiPrefix(`/sheet/${path}`), {
		method: 'GET',
		headers: { 'X-Session': get(session)?.ID ?? '' },
		cache: 'no-store'
	});
	if (!rsp.ok) {
		return undefined;
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
	SizeModifier: number;
	TechLevel: string;
	Hair: string;
	Eyes: string;
	Skin: string;
	Hand: string;
}

export interface Points {
	Total: number;
	Unspent: number;
	Ancestry: number;
	Attributes: number;
	Advantages: number;
	Disadvantages: number;
	Quirks: number;
	Skills: number;
	Spells: number;
}

export interface Attribute {
	Type: string;
	Key: string;
	Name: string;
	Value: number;
	Points: number;
}

export interface PointPool {
	Type: string;
	Key: string;
	Name: string;
	Value: number;
	Max: number;
	Points: number;
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
	HitPenalty: number;
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
	Move: number[];
	Dodge: number[];
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
	TitleIsImageKey: boolean;
	RightAligned: boolean;
	IsLink: boolean;
	Indentable: boolean;
}

export interface Cell {
	Primary: string;
	Secondary: string;
	Detail: string;
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
}
