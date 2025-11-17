/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 *
 * @author ENDERZOMBI102 <enderzombi102.end@gmail.com> 2024
 * @description Quick and dirty `lint-openapi` config to better conform to the Prof's requests and style.
 */
import vue from 'eslint-plugin-vue';

// noinspection JSUnusedGlobalSymbols
export default [

	{
        ignores: [
            "**/node_modules/",
            "dist/",
            "public/",      // <--- Questo zittisce gli errori di Bootstrap
            ".yarn/",
            ".pnp.*",
            "**/*.min.js",  // Ignora file minificati
            "**/*.bundle.js" // Ignora bundle
        ]
    },
	
	... vue.configs[ "flat/recommended" ],
	{
		rules: {
			'vue/multi-word-component-names': 'off',
			'vue/max-attributes-per-line': 'off',
			'vue/require-default-prop': 'off',
			'vue/singleline-html-element-content-newline': 'off'
		}
	},
];