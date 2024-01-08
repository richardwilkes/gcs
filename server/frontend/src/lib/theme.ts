import { writable } from 'svelte/store';

/** Holds the data for a theme. */
export type Theme = {
	kind: 'system' | 'light' | 'dark';
	colors: {
		[key: string]: {
			light: string;
			dark: string;
		};
	};
};

/** The current theme. */
export const currentTheme = writable(defaultTheme());

const systemIsDark = window.matchMedia('(prefers-color-scheme: dark)');
systemIsDark.addEventListener('change', () => currentTheme.update((value) => value));

currentTheme.subscribe((current) => {
	const kind = resolvedThemeKind(current);
	let buffer = ':root {\n';
	buffer += `	--color-scheme: ${kind};\n`;
	for (const [key, value] of Object.entries(current.colors)) {
		buffer += `	--color-${key}: ${value[kind]};\n`;
	}
	buffer += '}';
	document.getElementById('color-scheme')?.setAttribute('content', kind);
	const node = document.getElementById('theme');
	if (node) {
		node.innerHTML = buffer;
	}
});

/** Returns the resolved theme kind, either 'light' or 'dark'. */
export function resolvedThemeKind(theme: Theme) {
	return theme.kind === 'system' ? (systemIsDark.matches ? 'dark' : 'light') : theme.kind;
}

/** Returns true if the theme is dark. */
export function themeIsDark(theme: Theme) {
	return theme.kind === 'system' ? systemIsDark.matches : theme.kind === 'dark';
}

/** Creates and returns the default theme. */
export function defaultTheme(): Theme {
	return {
		kind: 'system',
		colors: {
			accent: {
				light: 'Teal',
				dark: '#649999'
			},
			background: {
				light: '#EEEEEE',
				dark: '#303030'
			},
			banding: {
				light: '#EBEBDC',
				dark: '#2A2A2A'
			},
			content: {
				light: '#F8F8F8',
				dark: '#202020'
			},
			control: {
				light: 'GhostWhite',
				dark: '#404040'
			},
			'control-edge': {
				light: '#606060',
				dark: '#606060'
			},
			'control-pressed': {
				light: '#0060A0',
				dark: '#0060A0'
			},
			divider: {
				light: 'Silver',
				dark: '#666666'
			},
			'drop-area': {
				light: '#CC0033',
				dark: 'Red'
			},
			editable: {
				light: 'White',
				dark: '#101010'
			},
			error: {
				light: '#C04040',
				dark: '#732525'
			},
			header: {
				light: '#2B2B2B',
				dark: '#404040'
			},
			hint: {
				light: 'Grey',
				dark: '#404040'
			},
			'icon-button': {
				light: '#606060',
				dark: 'Grey'
			},
			'icon-button-pressed': {
				light: '#0060A0',
				dark: '#0060A0'
			},
			'icon-button-rollover': {
				light: 'Black',
				dark: 'Silver'
			},
			'inactive-selection': {
				light: '#004080',
				dark: '#004080'
			},
			'indirect-selection': {
				light: '#E6F7FF',
				dark: '#002B40'
			},
			'interior-divider': {
				light: '#D8D8D8',
				dark: '#353535'
			},
			link: {
				light: '#739925',
				dark: '#00CC66'
			},
			'link-pressed': {
				light: '#0080FF',
				dark: '#0060A0'
			},
			'link-rollover': {
				light: '#00C000',
				dark: '#00B300'
			},
			marker: {
				light: '#FCF2C4',
				dark: '#003300'
			},
			'on-background': {
				light: 'Black',
				dark: '#DDDDDD'
			},
			'on-banding': {
				light: 'Black',
				dark: '#DDDDDD'
			},
			'on-content': {
				light: 'Black',
				dark: '#DDDDDD'
			},
			'on-control': {
				light: 'Black',
				dark: '#DDDDDD'
			},
			'on-control-pressed': {
				light: 'White',
				dark: 'White'
			},
			'on-editable': {
				light: '#0000A0',
				dark: '#649999'
			},
			'on-error': {
				light: 'White',
				dark: '#DDDDDD'
			},
			'on-header': {
				light: 'White',
				dark: 'Silver'
			},
			'on-inactive-selection': {
				light: '#E4E4E4',
				dark: '#E4E4E4'
			},
			'on-indirect-selection': {
				light: 'Black',
				dark: '#E4E4E4'
			},
			'on-marker': {
				light: 'Black',
				dark: '#DDDDDD'
			},
			'on-overloaded': {
				light: 'White',
				dark: '#DDDDDD'
			},
			'on-page': {
				light: 'Black',
				dark: '#A0A0A0'
			},
			'on-page-standout': {
				light: 'Black',
				dark: '#A0A0A0'
			},
			'on-search-list': {
				light: 'Black',
				dark: '#CCCCCC'
			},
			'on-selection': {
				light: 'White',
				dark: 'White'
			},
			'on-tab-current': {
				light: 'Black',
				dark: '#DDDDDD'
			},
			'on-tab-focused': {
				light: 'Black',
				dark: '#DDDDDD'
			},
			'on-tooltip': {
				light: 'Black',
				dark: 'Black'
			},
			'on-warning': {
				light: 'White',
				dark: '#DDDDDD'
			},
			overloaded: {
				light: '#C04040',
				dark: '#732525'
			},
			page: {
				light: 'White',
				dark: '#101010'
			},
			'page-standout': {
				light: '#DDDDDD',
				dark: '#404040'
			},
			'page-void': {
				light: 'Grey',
				dark: 'Black'
			},
			'pdf-link': {
				light: 'SpringGreen',
				dark: 'SpringGreen'
			},
			'pdf-marker': {
				light: 'Yellow',
				dark: 'Yellow'
			},
			scroll: {
				light: 'rgba(192,192,192,0.5019608)',
				dark: 'rgba(128,128,128,0.5019608)'
			},
			'scroll-edge': {
				light: 'Grey',
				dark: '#A0A0A0'
			},
			'scroll-rollover': {
				light: 'Silver',
				dark: 'Grey'
			},
			'search-list': {
				light: 'LightCyan',
				dark: '#002B2B'
			},
			selection: {
				light: '#0060A0',
				dark: '#0060A0'
			},
			'tab-current': {
				light: '#D3CFC5',
				dark: '#293D00'
			},
			'tab-focused': {
				light: '#E0D4AF',
				dark: '#446600'
			},
			tooltip: {
				light: '#FCFCC4',
				dark: '#FCFCC4'
			},
			'tooltip-marker': {
				light: '#804080',
				dark: '#996499'
			},
			warning: {
				light: '#E08000',
				dark: '#C06000'
			}
		}
	};
}
