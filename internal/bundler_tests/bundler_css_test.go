package bundler_tests

import (
	"testing"

	"github.com/evanw/esbuild/internal/compat"
	"github.com/evanw/esbuild/internal/config"
)

var css_suite = suite{
	name: "css",
}

func TestCSSEntryPoint(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				body {
					background: white;
					color: black }
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:          config.ModeBundle,
			AbsOutputFile: "/out.css",
		},
	})
}

func TestCSSAtImportMissing(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				@import "./missing.css";
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:          config.ModeBundle,
			AbsOutputFile: "/out.css",
		},
		expectedScanLog: `entry.css: ERROR: Could not resolve "./missing.css"
`,
	})
}

func TestCSSAtImportExternal(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				@import "./internal.css";
				@import "./external1.css";
				@import "./external2.css";
				@import "./charset1.css";
				@import "./charset2.css";
				@import "./external5.css" screen;
			`,
			"/internal.css": `
				@import "./external5.css" print;
				.before { color: red }
			`,
			"/charset1.css": `
				@charset "UTF-8";
				@import "./external3.css";
				@import "./external4.css";
				@import "./external5.css";
				@import "https://www.example.com/style1.css";
				@import "https://www.example.com/style2.css";
				@import "https://www.example.com/style3.css" print;
				.middle { color: green }
			`,
			"/charset2.css": `
				@charset "UTF-8";
				@import "./external3.css";
				@import "./external5.css" screen;
				@import "https://www.example.com/style1.css";
				@import "https://www.example.com/style3.css";
				.after { color: blue }
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExternalSettings: config.ExternalSettings{
				PostResolve: config.ExternalMatchers{Exact: map[string]bool{
					"/external1.css": true,
					"/external2.css": true,
					"/external3.css": true,
					"/external4.css": true,
					"/external5.css": true,
				}},
			},
		},
	})
}

func TestCSSAtImport(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				@import "./a.css";
				@import "./b.css";
				.entry { color: red }
			`,
			"/a.css": `
				@import "./shared.css";
				.a { color: green }
			`,
			"/b.css": `
				@import "./shared.css";
				.b { color: blue }
			`,
			"/shared.css": `
				.shared { color: black }
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:          config.ModeBundle,
			AbsOutputFile: "/out.css",
		},
	})
}

func TestCSSFromJSMissingImport(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import {missing} from "./a.css"
				console.log(missing)
			`,
			"/a.css": `
				.a { color: red }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
		expectedCompileLog: `entry.js: ERROR: No matching export in "a.css" for import "missing"
`,
	})
}

func TestCSSFromJSMissingStarImport(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import * as ns from "./a.css"
				console.log(ns.missing)
			`,
			"/a.css": `
				.a { color: red }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
		expectedCompileLog: `entry.js: WARNING: Import "missing" will always be undefined because there is no matching export in "a.css"
`,
	})
}

func TestImportGlobalCSSFromJS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import "./a.js"
				import "./b.js"
			`,
			"/a.js": `
				import * as stylesA from "./a.css"
				console.log('a', stylesA.a, stylesA.default.a)
			`,
			"/a.css": `
				.a { color: red }
			`,
			"/b.js": `
				import * as stylesB from "./b.css"
				console.log('b', stylesB.b, stylesB.default.b)
			`,
			"/b.css": `
				.b { color: blue }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
		expectedCompileLog: `a.js: WARNING: Import "a" will always be undefined because there is no matching export in "a.css"
b.js: WARNING: Import "b" will always be undefined because there is no matching export in "b.css"
`,
	})
}

func TestImportLocalCSSFromJS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import "./a.js"
				import "./b.js"
			`,
			"/a.js": `
				import * as stylesA from "./dir1/style.css"
				console.log('file 1', stylesA.button, stylesA.default.a)
			`,
			"/dir1/style.css": `
				.a { color: red }
				.button { display: none }
			`,
			"/b.js": `
				import * as stylesB from "./dir2/style.css"
				console.log('file 2', stylesB.button, stylesB.default.b)
			`,
			"/dir2/style.css": `
				.b { color: blue }
				.button { display: none }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
		},
	})
}

func TestImportLocalCSSFromJSMinifyIdentifiers(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import "./a.js"
				import "./b.js"
			`,
			"/a.js": `
				import * as stylesA from "./dir1/style.css"
				console.log('file 1', stylesA.button, stylesA.default.a)
			`,
			"/dir1/style.css": `
				.a { color: red }
				.button { display: none }
			`,
			"/b.js": `
				import * as stylesB from "./dir2/style.css"
				console.log('file 2', stylesB.button, stylesB.default.b)
			`,
			"/dir2/style.css": `
				.b { color: blue }
				.button { display: none }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
			MinifyIdentifiers: true,
		},
	})
}

func TestImportLocalCSSFromJSMinifyIdentifiersAvoidGlobalNames(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import "./global.css"
				import "./local.module.css"
			`,
			"/global.css": `
				:is(.a, .b, .c, .d, .e, .f, .g, .h, .i, .j, .k, .l, .m, .n, .o, .p, .q, .r, .s, .t, .u, .v, .w, .x, .y, .z),
				:is(.A, .B, .C, .D, .E, .F, .G, .H, .I, .J, .K, .L, .M, .N, .O, .P, .Q, .R, .S, .T, .U, .V, .W, .X, .Y, .Z),
				._ { color: red }
			`,
			"/local.module.css": `
				.rename-this { color: blue }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":         config.LoaderJS,
				".css":        config.LoaderCSS,
				".module.css": config.LoaderLocalCSS,
			},
			MinifyIdentifiers: true,
		},
	})
}

func TestImportCSSFromJSLocalVsGlobal(t *testing.T) {
	css := `
		.top_level { color: #000 }

		:global(.GLOBAL) { color: #001 }
		:local(.local) { color: #002 }

		div:global(.GLOBAL) { color: #003 }
		div:local(.local) { color: #004 }

		.top_level:global(div) { color: #005 }
		.top_level:local(div) { color: #006 }

		:global(div.GLOBAL) { color: #007 }
		:local(div.local) { color: #008 }

		div:global(span.GLOBAL) { color: #009 }
		div:local(span.local) { color: #00A }

		div:global(#GLOBAL_A.GLOBAL_B.GLOBAL_C):local(.local_a.local_b#local_c) { color: #00B }
		div:global(#GLOBAL_A .GLOBAL_B .GLOBAL_C):local(.local_a .local_b #local_c) { color: #00C }

		.nested {
			:global(&.GLOBAL) { color: #00D }
			:local(&.local) { color: #00E }

			&:global(.GLOBAL) { color: #00F }
			&:local(.local) { color: #010 }
		}

		:global(.GLOBAL_A .GLOBAL_B) { color: #011 }
		:local(.local_a .local_b) { color: #012 }

		div:global(.GLOBAL_A .GLOBAL_B):hover { color: #013 }
		div:local(.local_a .local_b):hover { color: #014 }

		div :global(.GLOBAL_A .GLOBAL_B) span { color: #015 }
		div :local(.local_a .local_b) span { color: #016 }

		div > :global(.GLOBAL_A ~ .GLOBAL_B) + span { color: #017 }
		div > :local(.local_a ~ .local_b) + span { color: #018 }

		div:global(+ .GLOBAL_A):hover { color: #019 }
		div:local(+ .local_a):hover { color: #01A }

		:global.GLOBAL:local.local { color: #01B }
		:global .GLOBAL :local .local { color: #01C }

		:global {
			.GLOBAL {
				before: outer;
				:local {
					before: inner;
					.local {
						color: #01D;
					}
					after: inner;
				}
				after: outer;
			}
		}
	`

	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import normalStyles from "./normal.css"
				import globalStyles from "./LOCAL.global-css"
				import localStyles from "./LOCAL.local-css"

				console.log('should be empty:', normalStyles)
				console.log('fewer local names:', globalStyles)
				console.log('more local names:', localStyles)
			`,
			"/normal.css":       css,
			"/LOCAL.global-css": css,
			"/LOCAL.local-css":  css,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":         config.LoaderJS,
				".css":        config.LoaderCSS,
				".global-css": config.LoaderGlobalCSS,
				".local-css":  config.LoaderLocalCSS,
			},
		},
	})
}

func TestImportCSSFromJSLowerBareLocalAndGlobal(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.css"
				console.log(styles)
			`,
			"/styles.css": `
				.before { color: #000 }
				:local { .button { color: #000 } }
				.after { color: #000 }

				.before { color: #001 }
				:global { .button { color: #001 } }
				.after { color: #001 }

				div { :local { .button { color: #002 } } }
				div { :global { .button { color: #003 } } }

				:local(:global) { color: #004 }
				:global(:local) { color: #005 }

				:local(:global) { .button { color: #006 } }
				:global(:local) { .button { color: #007 } }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
			UnsupportedCSSFeatures: compat.Nesting,
		},
	})
}

func TestImportCSSFromJSLocalAtKeyframes(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.css"
				console.log(styles)
			`,
			"/styles.css": `
				@keyframes local_name { to { color: red } }

				div :global { animation-name: none }
				div :local { animation-name: none }

				div :global { animation-name: global_name }
				div :local { animation-name: local_name }

				div :global { animation-name: global_name1, none, global_name2, Inherit, INITIAL, revert, revert-layer, unset }
				div :local { animation-name: local_name1, none, local_name2, Inherit, INITIAL, revert, revert-layer, unset }

				div :global { animation: 2s infinite global_name }
				div :local { animation: 2s infinite local_name }

				/* Someone wanted to be able to name their animations "none" */
				@keyframes "none" { to { color: red } }
				div :global { animation-name: "none" }
				div :local { animation-name: "none" }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
			UnsupportedCSSFeatures: compat.Nesting,
		},
	})
}

func TestImportCSSFromJSLocalAtCounterStyle(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import list_style_type from "./list_style_type.css"
				import list_style from "./list_style.css"
				console.log(list_style_type, list_style)
			`,
			"/list_style_type.css": `
				@counter-style local { symbols: A B C }

				div :global { list-style-type: GLOBAL }
				div :local { list-style-type: local }

				/* Must not accept invalid type values */
				div :local { list-style-type: INITIAL }
				div :local { list-style-type: decimal }
				div :local { list-style-type: disc }
				div :local { list-style-type: SQUARE }
				div :local { list-style-type: circle }
				div :local { list-style-type: disclosure-OPEN }
				div :local { list-style-type: DISCLOSURE-closed }
			`,

			"/list_style.css": `
				@counter-style local { symbols: A B C }

				div :global { list-style: GLOBAL }
				div :local { list-style: local }

				/* The first one is the type */
				div :local { list-style: local none }
				div :local { list-style: local url(http://) }
				div :local { list-style: local linear-gradient(red, green) }
				div :local { list-style: local inside }
				div :local { list-style: local outside }

				/* The second one is the type */
				div :local { list-style: none local }
				div :local { list-style: url(http://) local }
				div :local { list-style: linear-gradient(red, green) local }
				div :local { list-style: local inside }
				div :local { list-style: local outside }
				div :local { list-style: inside inside }
				div :local { list-style: inside outside }
				div :local { list-style: outside inside }
				div :local { list-style: outside outside }

				/* The type is set to "none" here */
				div :local { list-style: url(http://) none invalid }
				div :local { list-style: linear-gradient(red, green) none invalid }

				/* Must not accept invalid type values */
				div :local { list-style: INITIAL }
				div :local { list-style: decimal }
				div :local { list-style: disc }
				div :local { list-style: SQUARE }
				div :local { list-style: circle }
				div :local { list-style: disclosure-OPEN }
				div :local { list-style: DISCLOSURE-closed }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
			UnsupportedCSSFeatures: compat.Nesting,
		},
	})
}

func TestImportCSSFromJSLocalAtContainer(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.css"
				console.log(styles)
			`,
			"/styles.css": `
				@container not (max-width: 100px) { div { color: red } }
				@container local (max-width: 100px) { div { color: red } }
				@container local not (max-width: 100px) { div { color: red } }
				@container local (max-width: 100px) or (min-height: 100px) { div { color: red } }
				@container local (max-width: 100px) and (min-height: 100px) { div { color: red } }
				@container general_enclosed(max-width: 100px) { div { color: red } }
				@container local general_enclosed(max-width: 100px) { div { color: red } }

				div :global { container-name: NONE initial }
				div :local { container-name: none INITIAL }
				div :global { container-name: GLOBAL1 GLOBAL2 }
				div :local { container-name: local1 local2 }

				div :global { container: none }
				div :local { container: NONE }
				div :global { container: NONE / size }
				div :local { container: none / size }

				div :global { container: GLOBAL1 GLOBAL2 }
				div :local { container: local1 local2 }
				div :global { container: GLOBAL1 GLOBAL2 / size }
				div :local { container: local1 local2 / size }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
			UnsupportedCSSFeatures: compat.Nesting,
		},
	})
}

func TestImportCSSFromJSNthIndexLocal(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.css"
				console.log(styles)
			`,
			"/styles.css": `
				:nth-child(2n of .local) { color: #000 }
				:nth-child(2n of :local(#local), :global(.GLOBAL)) { color: #001 }
				:nth-child(2n of .local1 :global .GLOBAL1, .GLOBAL2 :local .local2) { color: #002 }
				.local1, :nth-child(2n of :global .GLOBAL), .local2 { color: #003 }

				:nth-last-child(2n of .local) { color: #000 }
				:nth-last-child(2n of :local(#local), :global(.GLOBAL)) { color: #001 }
				:nth-last-child(2n of .local1 :global .GLOBAL1, .GLOBAL2 :local .local2) { color: #002 }
				.local1, :nth-last-child(2n of :global .GLOBAL), .local2 { color: #003 }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
			UnsupportedCSSFeatures: compat.Nesting,
		},
	})
}

func TestImportCSSFromJSComposes(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.module.css"
				console.log(styles)
			`,
			"/global.css": `
				.GLOBAL1 {
					color: black;
				}
			`,
			"/styles.module.css": `
				@import "global.css";
				.local0 {
					composes: local1;
					:global {
						composes: GLOBAL1 GLOBAL2;
					}
				}
				.local0 {
					composes: GLOBAL2 GLOBAL3 from global;
					composes: local1 local2;
					background: green;
				}
				.local0 :global {
					composes: GLOBAL4;
				}
				.local3 {
					border: 1px solid black;
					composes: local4;
				}
				.local4 {
					opacity: 0.5;
				}
				.local1 {
					color: red;
					composes: local3;
				}
				.fromOtherFile {
					composes: local0 from "other1.module.css";
					composes: local0 from "other2.module.css";
				}
			`,
			"/other1.module.css": `
				.local0 {
					composes: base1 base2 from "base.module.css";
					color: blue;
				}
			`,
			"/other2.module.css": `
				.local0 {
					composes: base1 base3 from "base.module.css";
					background: purple;
				}
			`,
			"/base.module.css": `
				.base1 {
					cursor: pointer;
				}
				.base2 {
					display: inline;
				}
				.base3 {
					float: left;
				}
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":         config.LoaderJS,
				".css":        config.LoaderCSS,
				".module.css": config.LoaderLocalCSS,
			},
		},
	})
}

func TestImportCSSFromJSComposesFromMissingImport(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.module.css"
				console.log(styles)
			`,
			"/styles.module.css": `
				.foo {
					composes: x from "file.module.css";
					composes: y from "file.module.css";
					composes: z from "file.module.css";
					composes: x from "file.css";
				}
			`,
			"/file.module.css": `
				.x {
					color: red;
				}
				:global(.y) {
					color: blue;
				}
			`,
			"/file.css": `
				.x {
					color: red;
				}
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":         config.LoaderJS,
				".module.css": config.LoaderLocalCSS,
				".css":        config.LoaderCSS,
			},
		},
		expectedCompileLog: `styles.module.css: ERROR: Cannot use global name "y" with "composes"
file.module.css: NOTE: The global name "y" is defined here:
NOTE: Use the ":local" selector to change "y" into a local name.
styles.module.css: ERROR: The name "z" never appears in "file.module.css"
styles.module.css: ERROR: Cannot use global name "x" with "composes"
file.css: NOTE: The global name "x" is defined here:
NOTE: Use the "local-css" loader for "file.css" to enable local names.
`,
	})
}

func TestImportCSSFromJSComposesFromNotCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.css"
				console.log(styles)
			`,
			"/styles.css": `
				.foo {
					composes: bar from "file.txt";
				}
			`,
			"/file.txt": `
				.bar {
					color: red;
				}
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
				".txt": config.LoaderText,
			},
		},
		expectedScanLog: `styles.css: ERROR: Cannot use "composes" with "file.txt"
NOTE: You can only use "composes" with CSS files and "file.txt" is not a CSS file (it was loaded with the "text" loader).
`,
	})
}

func TestImportCSSFromJSComposesCircular(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.css"
				console.log(styles)
			`,
			"/styles.css": `
				.foo {
					composes: bar;
				}
				.bar {
					composes: foo;
				}
				.baz {
					composes: baz;
				}
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
		},
	})
}

func TestImportCSSFromJSComposesFromCircular(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.css"
				console.log(styles)
			`,
			"/styles.css": `
				.foo {
					composes: bar from "other.css";
				}
				.bar {
					composes: bar from "styles.css";
				}
			`,
			"/other.css": `
				.bar {
					composes: foo from "styles.css";
				}
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
		},
	})
}

func TestImportCSSFromJSComposesFromUndefined(t *testing.T) {
	note := "NOTE: The specification of \"composes\" does not define an order when class declarations from separate files are composed together. " +
		"The value of the \"zoom\" property for \"foo\" may change unpredictably as the code is edited. " +
		"Make sure that all definitions of \"zoom\" for \"foo\" are in a single file."
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import styles from "./styles.css"
				console.log(styles)
			`,
			"/styles.css": `
				@import "well-defined.css";
				@import "undefined/case1.css";
				@import "undefined/case2.css";
				@import "undefined/case3.css";
				@import "undefined/case4.css";
				@import "undefined/case5.css";
			`,
			"/well-defined.css": `
				.z1 { composes: z2; zoom: 1; }
				.z2 { zoom: 2; }

				.z4 { zoom: 4; }
				.z3 { composes: z4; zoom: 3; }

				.z5 { composes: foo bar from "file-1.css"; }
			`,
			"/undefined/case1.css": `
				.foo {
					composes: foo from "../file-1.css";
					zoom: 2;
				}
			`,
			"/undefined/case2.css": `
				.foo {
					composes: foo from "../file-1.css";
					composes: foo from "../file-2.css";
				}
			`,
			"/undefined/case3.css": `
				.foo { composes: nested1 nested2; }
				.nested1 { zoom: 3; }
				.nested2 { composes: foo from "../file-2.css"; }
			`,
			"/undefined/case4.css": `
				.foo { composes: nested1 nested2; }
				.nested1 { composes: foo from "../file-1.css"; }
				.nested2 { zoom: 3; }
			`,
			"/undefined/case5.css": `
				.foo { composes: nested1 nested2; }
				.nested1 { composes: foo from "../file-1.css"; }
				.nested2 { composes: foo from "../file-2.css"; }
			`,
			"/file-1.css": `
				.foo { zoom: 1; }
				.bar { zoom: 2; }
			`,
			"/file-2.css": `
				.foo { zoom: 2; }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderLocalCSS,
			},
		},
		expectedCompileLog: `undefined/case1.css: WARNING: The value of "zoom" in the "foo" class is undefined
file-1.css: NOTE: The first definition of "zoom" is here:
undefined/case1.css: NOTE: The second definition of "zoom" is here:
` + note + `
undefined/case2.css: WARNING: The value of "zoom" in the "foo" class is undefined
file-1.css: NOTE: The first definition of "zoom" is here:
file-2.css: NOTE: The second definition of "zoom" is here:
` + note + `
undefined/case3.css: WARNING: The value of "zoom" in the "foo" class is undefined
undefined/case3.css: NOTE: The first definition of "zoom" is here:
file-2.css: NOTE: The second definition of "zoom" is here:
` + note + `
undefined/case4.css: WARNING: The value of "zoom" in the "foo" class is undefined
file-1.css: NOTE: The first definition of "zoom" is here:
undefined/case4.css: NOTE: The second definition of "zoom" is here:
` + note + `
undefined/case5.css: WARNING: The value of "zoom" in the "foo" class is undefined
file-1.css: NOTE: The first definition of "zoom" is here:
file-2.css: NOTE: The second definition of "zoom" is here:
` + note + `
`,
	})
}

func TestImportCSSFromJSWriteToStdout(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import "./entry.css"
			`,
			"/entry.css": `
				.entry { color: red }
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:          config.ModeBundle,
			WriteToStdout: true,
		},
		expectedScanLog: `entry.js: ERROR: Cannot import "entry.css" into a JavaScript file without an output path configured
`,
	})
}

func TestImportJSFromCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				export default 123
			`,
			"/entry.css": `
				@import "./entry.js";
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
		expectedScanLog: `entry.css: ERROR: Cannot import "entry.js" into a CSS file
NOTE: An "@import" rule can only be used to import another CSS file and "entry.js" is not a CSS file (it was loaded with the "js" loader).
`,
	})
}

func TestImportJSONFromCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.json": `
				{}
			`,
			"/entry.css": `
				@import "./entry.json";
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
		expectedScanLog: `entry.css: ERROR: Cannot import "entry.json" into a CSS file
NOTE: An "@import" rule can only be used to import another CSS file and "entry.json" is not a CSS file (it was loaded with the "json" loader).
`,
	})
}

func TestMissingImportURLInCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/src/entry.css": `
				a { background: url(./one.png); }
				b { background: url("./two.png"); }
			`,
		},
		entryPaths: []string{"/src/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
		expectedScanLog: `src/entry.css: ERROR: Could not resolve "./one.png"
src/entry.css: ERROR: Could not resolve "./two.png"
`,
	})
}

func TestExternalImportURLInCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/src/entry.css": `
				div:after {
					content: 'If this is recognized, the path should become "../src/external.png"';
					background: url(./external.png);
				}

				/* These URLs should be external automatically */
				a { background: url(http://example.com/images/image.png) }
				b { background: url(https://example.com/images/image.png) }
				c { background: url(//example.com/images/image.png) }
				d { background: url(data:image/png;base64,iVBORw0KGgo=) }
				path { fill: url(#filter) }
			`,
		},
		entryPaths: []string{"/src/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExternalSettings: config.ExternalSettings{
				PostResolve: config.ExternalMatchers{Exact: map[string]bool{
					"/src/external.png": true,
				}},
			},
		},
	})
}

func TestInvalidImportURLInCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				a {
					background: url(./js.js);
					background: url("./jsx.jsx");
					background: url(./ts.ts);
					background: url('./tsx.tsx');
					background: url(./json.json);
					background: url(./css.css);
				}
			`,
			"/js.js":     `export default 123`,
			"/jsx.jsx":   `export default 123`,
			"/ts.ts":     `export default 123`,
			"/tsx.tsx":   `export default 123`,
			"/json.json": `{ "test": true }`,
			"/css.css":   `a { color: red }`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
		expectedScanLog: `entry.css: ERROR: Cannot use "js.js" as a URL
NOTE: You can't use a "url()" token to reference the file "js.js" because it was loaded with the "js" loader, which doesn't provide a URL to embed in the resulting CSS.
entry.css: ERROR: Cannot use "jsx.jsx" as a URL
NOTE: You can't use a "url()" token to reference the file "jsx.jsx" because it was loaded with the "jsx" loader, which doesn't provide a URL to embed in the resulting CSS.
entry.css: ERROR: Cannot use "ts.ts" as a URL
NOTE: You can't use a "url()" token to reference the file "ts.ts" because it was loaded with the "ts" loader, which doesn't provide a URL to embed in the resulting CSS.
entry.css: ERROR: Cannot use "tsx.tsx" as a URL
NOTE: You can't use a "url()" token to reference the file "tsx.tsx" because it was loaded with the "tsx" loader, which doesn't provide a URL to embed in the resulting CSS.
entry.css: ERROR: Cannot use "json.json" as a URL
NOTE: You can't use a "url()" token to reference the file "json.json" because it was loaded with the "json" loader, which doesn't provide a URL to embed in the resulting CSS.
entry.css: ERROR: Cannot use "css.css" as a URL
NOTE: You can't use a "url()" token to reference a CSS file, and "css.css" is a CSS file (it was loaded with the "css" loader).
`,
	})
}

func TestTextImportURLInCSSText(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				a {
					background: url(./example.txt);
				}
			`,
			"/example.txt": `This is some text.`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
	})
}

func TestDataURLImportURLInCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				a {
					background: url(./example.png);
				}
			`,
			"/example.png": "\x89\x50\x4E\x47\x0D\x0A\x1A\x0A",
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".css": config.LoaderCSS,
				".png": config.LoaderDataURL,
			},
		},
	})
}

func TestBinaryImportURLInCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				a {
					background: url(./example.png);
				}
			`,
			"/example.png": "\x89\x50\x4E\x47\x0D\x0A\x1A\x0A",
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".css": config.LoaderCSS,
				".png": config.LoaderBinary,
			},
		},
	})
}

func TestBase64ImportURLInCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				a {
					background: url(./example.png);
				}
			`,
			"/example.png": "\x89\x50\x4E\x47\x0D\x0A\x1A\x0A",
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".css": config.LoaderCSS,
				".png": config.LoaderBase64,
			},
		},
	})
}

func TestFileImportURLInCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				@import "./one.css";
				@import "./two.css";
			`,
			"/one.css": `
				a { background: url(./example.data) }
			`,
			"/two.css": `
				b { background: url(./example.data) }
			`,
			"/example.data": "This is some data.",
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".css":  config.LoaderCSS,
				".data": config.LoaderFile,
			},
		},
	})
}

func TestIgnoreURLsInAtRulePrelude(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				/* This should not generate a path resolution error */
				@supports (background: url(ignored.png)) {
					a { color: red }
				}
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
	})
}

func TestPackageURLsInCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
			  @import "test.css";

				a { background: url(a/1.png); }
				b { background: url(b/2.png); }
				c { background: url(c/3.png); }
			`,
			"/test.css":             `.css { color: red }`,
			"/a/1.png":              `a-1`,
			"/node_modules/b/2.png": `b-2-node_modules`,
			"/c/3.png":              `c-3`,
			"/node_modules/c/3.png": `c-3-node_modules`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".css": config.LoaderCSS,
				".png": config.LoaderBase64,
			},
		},
	})
}

func TestCSSAtImportExtensionOrderCollision(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			// This should avoid picking ".js" because it's explicitly configured as non-CSS
			"/entry.css": `@import "./test";`,
			"/test.js":   `console.log('js')`,
			"/test.css":  `.css { color: red }`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:           config.ModeBundle,
			AbsOutputFile:  "/out.css",
			ExtensionOrder: []string{".js", ".css"},
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderCSS,
			},
		},
	})
}

func TestCSSAtImportExtensionOrderCollisionUnsupported(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			// This still shouldn't pick ".js" even though ".sass" isn't ".css"
			"/entry.css": `@import "./test";`,
			"/test.js":   `console.log('js')`,
			"/test.sass": `// some code`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:           config.ModeBundle,
			AbsOutputFile:  "/out.css",
			ExtensionOrder: []string{".js", ".sass"},
			ExtensionToLoader: map[string]config.Loader{
				".js":  config.LoaderJS,
				".css": config.LoaderCSS,
			},
		},
		expectedScanLog: `entry.css: ERROR: No loader is configured for ".sass" files: test.sass
`,
	})
}

func TestCSSAtImportConditionsNoBundle(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `@import "./print.css" print;`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:          config.ModePassThrough,
			AbsOutputFile: "/out.css",
		},
	})
}

func TestCSSAtImportConditionsBundleExternal(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `@import "https://example.com/print.css" print;`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:          config.ModeBundle,
			AbsOutputFile: "/out.css",
		},
	})
}

func TestCSSAtImportConditionsBundleExternalConditionWithURL(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				@import "https://example.com/foo.css" (foo: url("foo.png")) and (bar: url("bar.png"));
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:          config.ModeBundle,
			AbsOutputFile: "/out.css",
		},
	})
}

func TestCSSAtImportConditionsBundle(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `@import "./print.css" print;`,
			"/print.css": `body { color: red }`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:          config.ModeBundle,
			AbsOutputFile: "/out.css",
		},
		expectedScanLog: `entry.css: ERROR: Bundling with conditional "@import" rules is not currently supported
`,
	})
}

// This test mainly just makes sure that this scenario doesn't crash
func TestCSSAndJavaScriptCodeSplittingIssue1064(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/a.js": `
				import shared from './shared.js'
				console.log(shared() + 1)
			`,
			"/b.js": `
				import shared from './shared.js'
				console.log(shared() + 2)
			`,
			"/c.css": `
				@import "./shared.css";
				body { color: red }
			`,
			"/d.css": `
				@import "./shared.css";
				body { color: blue }
			`,
			"/shared.js": `
				export default function() { return 3 }
			`,
			"/shared.css": `
				body { background: black }
			`,
		},
		entryPaths: []string{
			"/a.js",
			"/b.js",
			"/c.css",
			"/d.css",
		},
		options: config.Options{
			Mode:          config.ModeBundle,
			OutputFormat:  config.FormatESModule,
			CodeSplitting: true,
			AbsOutputDir:  "/out",
		},
	})
}

func TestCSSExternalQueryAndHashNoMatchIssue1822(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				a { background: url(foo/bar.png?baz) }
				b { background: url(foo/bar.png#baz) }
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:          config.ModeBundle,
			AbsOutputFile: "/out.css",
			ExternalSettings: config.ExternalSettings{
				PreResolve: config.ExternalMatchers{Patterns: []config.WildcardPattern{
					{Suffix: ".png"},
				}},
			},
		},
		expectedScanLog: `entry.css: ERROR: Could not resolve "foo/bar.png?baz"
NOTE: You can mark the path "foo/bar.png?baz" as external to exclude it from the bundle, which will remove this error.
entry.css: ERROR: Could not resolve "foo/bar.png#baz"
NOTE: You can mark the path "foo/bar.png#baz" as external to exclude it from the bundle, which will remove this error.
`,
	})
}

func TestCSSExternalQueryAndHashMatchIssue1822(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				a { background: url(foo/bar.png?baz) }
				b { background: url(foo/bar.png#baz) }
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:          config.ModeBundle,
			AbsOutputFile: "/out.css",
			ExternalSettings: config.ExternalSettings{
				PreResolve: config.ExternalMatchers{Patterns: []config.WildcardPattern{
					{Suffix: ".png?baz"},
					{Suffix: ".png#baz"},
				}},
			},
		},
	})
}

func TestCSSNestingOldBrowser(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			// These are now the only two cases that warn about ":is" not being supported
			"/two-type-selectors.css":   `a { .c b& { color: red; } }`,
			"/two-parent-selectors.css": `a b { .c & { color: red; } }`,

			// Make sure this only generates one warning (even though it generates ":is" three times)
			"/only-one-warning.css": `.a, .b .c, .d { & > & { color: red; } }`,

			"/nested-@layer.css":          `a { @layer base { color: red; } }`,
			"/nested-@media.css":          `a { @media screen { color: red; } }`,
			"/nested-ampersand-twice.css": `a { &, & { color: red; } }`,
			"/nested-ampersand-first.css": `a { &, b { color: red; } }`,
			"/nested-attribute.css":       `a { [href] { color: red; } }`,
			"/nested-colon.css":           `a { :hover { color: red; } }`,
			"/nested-dot.css":             `a { .cls { color: red; } }`,
			"/nested-greaterthan.css":     `a { > b { color: red; } }`,
			"/nested-hash.css":            `a { #id { color: red; } }`,
			"/nested-plus.css":            `a { + b { color: red; } }`,
			"/nested-tilde.css":           `a { ~ b { color: red; } }`,

			"/toplevel-ampersand-twice.css":  `&, & { color: red; }`,
			"/toplevel-ampersand-first.css":  `&, a { color: red; }`,
			"/toplevel-ampersand-second.css": `a, & { color: red; }`,
			"/toplevel-attribute.css":        `[href] { color: red; }`,
			"/toplevel-colon.css":            `:hover { color: red; }`,
			"/toplevel-dot.css":              `.cls { color: red; }`,
			"/toplevel-greaterthan.css":      `> b { color: red; }`,
			"/toplevel-hash.css":             `#id { color: red; }`,
			"/toplevel-plus.css":             `+ b { color: red; }`,
			"/toplevel-tilde.css":            `~ b { color: red; }`,

			"/media-ampersand-twice.css":  `@media screen { &, & { color: red; } }`,
			"/media-ampersand-first.css":  `@media screen { &, a { color: red; } }`,
			"/media-ampersand-second.css": `@media screen { a, & { color: red; } }`,
			"/media-attribute.css":        `@media screen { [href] { color: red; } }`,
			"/media-colon.css":            `@media screen { :hover { color: red; } }`,
			"/media-dot.css":              `@media screen { .cls { color: red; } }`,
			"/media-greaterthan.css":      `@media screen { > b { color: red; } }`,
			"/media-hash.css":             `@media screen { #id { color: red; } }`,
			"/media-plus.css":             `@media screen { + b { color: red; } }`,
			"/media-tilde.css":            `@media screen { ~ b { color: red; } }`,

			// See: https://github.com/evanw/esbuild/issues/3197
			"/page-no-warning.css": `@page { @top-left { background: red } }`,
		},
		entryPaths: []string{
			"/two-type-selectors.css",
			"/two-parent-selectors.css",

			"/only-one-warning.css",

			"/nested-@layer.css",
			"/nested-@media.css",
			"/nested-ampersand-twice.css",
			"/nested-ampersand-first.css",
			"/nested-attribute.css",
			"/nested-colon.css",
			"/nested-dot.css",
			"/nested-greaterthan.css",
			"/nested-hash.css",
			"/nested-plus.css",
			"/nested-tilde.css",

			"/toplevel-ampersand-twice.css",
			"/toplevel-ampersand-first.css",
			"/toplevel-ampersand-second.css",
			"/toplevel-attribute.css",
			"/toplevel-colon.css",
			"/toplevel-dot.css",
			"/toplevel-greaterthan.css",
			"/toplevel-hash.css",
			"/toplevel-plus.css",
			"/toplevel-tilde.css",

			"/media-ampersand-twice.css",
			"/media-ampersand-first.css",
			"/media-ampersand-second.css",
			"/media-attribute.css",
			"/media-colon.css",
			"/media-dot.css",
			"/media-greaterthan.css",
			"/media-hash.css",
			"/media-plus.css",
			"/media-tilde.css",

			"/page-no-warning.css",
		},
		options: config.Options{
			Mode:                   config.ModeBundle,
			AbsOutputDir:           "/out",
			UnsupportedCSSFeatures: compat.Nesting | compat.IsPseudoClass,
			OriginalTargetEnv:      "chrome10",
		},
		expectedScanLog: `only-one-warning.css: WARNING: Transforming this CSS nesting syntax is not supported in the configured target environment (chrome10)
NOTE: The nesting transform for this case must generate an ":is(...)" but the configured target environment does not support the ":is" pseudo-class.
two-parent-selectors.css: WARNING: Transforming this CSS nesting syntax is not supported in the configured target environment (chrome10)
NOTE: The nesting transform for this case must generate an ":is(...)" but the configured target environment does not support the ":is" pseudo-class.
two-type-selectors.css: WARNING: Transforming this CSS nesting syntax is not supported in the configured target environment (chrome10)
NOTE: The nesting transform for this case must generate an ":is(...)" but the configured target environment does not support the ":is" pseudo-class.
`,
	})
}

// The mapping of JS entry point to associated CSS bundle isn't necessarily 1:1.
// Here is a case where it isn't. Two JS entry points share the same associated
// CSS bundle. This must be reflected in the metafile by only having the JS
// entry points point to the associated CSS bundle but not the other way around
// (since there isn't one JS entry point to point to). This test mainly exists
// to document this edge case.
func TestMetafileCSSBundleTwoToOne(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/foo/entry.js": `
				import '../common.css'
				console.log('foo')
			`,
			"/bar/entry.js": `
				import '../common.css'
				console.log('bar')
			`,
			"/common.css": `
				body { color: red }
			`,
		},
		entryPaths: []string{
			"/foo/entry.js",
			"/bar/entry.js",
		},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			EntryPathTemplate: []config.PathTemplate{
				// "[ext]/[hash]"
				{Data: "./", Placeholder: config.ExtPlaceholder},
				{Data: "/", Placeholder: config.HashPlaceholder},
			},
			NeedsMetafile: true,
		},
	})
}

func TestDeduplicateRules(t *testing.T) {
	// These are done as bundler tests instead of parser tests because rule
	// deduplication now happens during linking (so that it has effects across files)
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/yes0.css": "a { color: red; color: green; color: red }",
			"/yes1.css": "a { color: red } a { color: green } a { color: red }",
			"/yes2.css": "@media screen { a { color: red } } @media screen { a { color: red } }",
			"/yes3.css": "@media screen { a { color: red } } @media screen { & a { color: red } }",

			"/no0.css": "@media screen { a { color: red } } @media screen { b a& { color: red } }",
			"/no1.css": "@media screen { a { color: red } } @media screen { a[x] { color: red } }",
			"/no2.css": "@media screen { a { color: red } } @media screen { a.x { color: red } }",
			"/no3.css": "@media screen { a { color: red } } @media screen { a#x { color: red } }",
			"/no4.css": "@media screen { a { color: red } } @media screen { a:x { color: red } }",
			"/no5.css": "@media screen { a:x { color: red } } @media screen { a:x(y) { color: red } }",
			"/no6.css": "@media screen { a b { color: red } } @media screen { a + b { color: red } }",

			"/across-files.css":   "@import 'across-files-0.css'; @import 'across-files-1.css'; @import 'across-files-2.css';",
			"/across-files-0.css": "a { color: red; color: red }",
			"/across-files-1.css": "a { color: green }",
			"/across-files-2.css": "a { color: red }",

			"/across-files-url.css":   "@import 'across-files-url-0.css'; @import 'across-files-url-1.css'; @import 'across-files-url-2.css';",
			"/across-files-url-0.css": "@import 'http://example.com/some.css'; @font-face { src: url(http://example.com/some.font); }",
			"/across-files-url-1.css": "@font-face { src: url(http://example.com/some.other.font); }",
			"/across-files-url-2.css": "@font-face { src: url(http://example.com/some.font); }",
		},
		entryPaths: []string{
			"/yes0.css",
			"/yes1.css",
			"/yes2.css",
			"/yes3.css",

			"/no0.css",
			"/no1.css",
			"/no2.css",
			"/no3.css",
			"/no4.css",
			"/no5.css",
			"/no6.css",

			"/across-files.css",
			"/across-files-url.css",
		},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			MinifySyntax: true,
		},
	})
}

func TestDeduplicateRulesGlobalVsLocalNames(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				@import "a.css";
				@import "b.css";
			`,
			"/a.css": `
				a { color: red } /* SHOULD BE REMOVED */
				b { color: green }

				:global(.foo) { color: red } /* SHOULD BE REMOVED */
				:global(.bar) { color: green }

				:local(.foo) { color: red }
				:local(.bar) { color: green }

				div :global { animation-name: anim_global } /* SHOULD BE REMOVED */
				div :local { animation-name: anim_local }
			`,
			"/b.css": `
				a { color: red }
				b { color: blue }

				:global(.foo) { color: red }
				:global(.bar) { color: blue }

				:local(.foo) { color: red }
				:local(.bar) { color: blue }

				div :global { animation-name: anim_global }
				div :local { animation-name: anim_local }
			`,
		},
		entryPaths: []string{"entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			MinifySyntax: true,
			ExtensionToLoader: map[string]config.Loader{
				".css": config.LoaderLocalCSS,
			},
		},
	})
}

// This test makes sure JS files that import local CSS names using the
// wrong name (e.g. a typo) get a warning so that the problem is noticed.
func TestUndefinedImportWarningCSS(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.js": `
				import * as empty_js from './empty.js'
				import * as empty_esm_js from './empty.esm.js'
				import * as empty_json from './empty.json'
				import * as empty_css from './empty.css'
				import * as empty_global_css from './empty.global-css'
				import * as empty_local_css from './empty.local-css'

				import * as pkg_empty_js from 'pkg/empty.js'
				import * as pkg_empty_esm_js from 'pkg/empty.esm.js'
				import * as pkg_empty_json from 'pkg/empty.json'
				import * as pkg_empty_css from 'pkg/empty.css'
				import * as pkg_empty_global_css from 'pkg/empty.global-css'
				import * as pkg_empty_local_css from 'pkg/empty.local-css'

				import 'pkg'

				console.log(
					empty_js.foo,
					empty_esm_js.foo,
					empty_json.foo,
					empty_css.foo,
					empty_global_css.foo,
					empty_local_css.foo,
				)

				console.log(
					pkg_empty_js.foo,
					pkg_empty_esm_js.foo,
					pkg_empty_json.foo,
					pkg_empty_css.foo,
					pkg_empty_global_css.foo,
					pkg_empty_local_css.foo,
				)
			`,

			"/empty.js":         ``,
			"/empty.esm.js":     `export {}`,
			"/empty.json":       `{}`,
			"/empty.css":        ``,
			"/empty.global-css": ``,
			"/empty.local-css":  ``,

			"/node_modules/pkg/empty.js":         ``,
			"/node_modules/pkg/empty.esm.js":     `export {}`,
			"/node_modules/pkg/empty.json":       `{}`,
			"/node_modules/pkg/empty.css":        ``,
			"/node_modules/pkg/empty.global-css": ``,
			"/node_modules/pkg/empty.local-css":  ``,

			// Files inside of "node_modules" should not generate a warning
			"/node_modules/pkg/index.js": `
				import * as empty_js from './empty.js'
				import * as empty_esm_js from './empty.esm.js'
				import * as empty_json from './empty.json'
				import * as empty_css from './empty.css'
				import * as empty_global_css from './empty.global-css'
				import * as empty_local_css from './empty.local-css'

				console.log(
					empty_js.foo,
					empty_esm_js.foo,
					empty_json.foo,
					empty_css.foo,
					empty_global_css.foo,
					empty_local_css.foo,
				)
			`,
		},
		entryPaths: []string{"/entry.js"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
			ExtensionToLoader: map[string]config.Loader{
				".js":         config.LoaderJS,
				".json":       config.LoaderJSON,
				".css":        config.LoaderCSS,
				".global-css": config.LoaderGlobalCSS,
				".local-css":  config.LoaderLocalCSS,
			},
		},
		expectedCompileLog: `entry.js: WARNING: Import "foo" will always be undefined because the file "empty.js" has no exports
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "empty.esm.js"
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "empty.json"
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "empty.css"
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "empty.global-css"
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "empty.local-css"
entry.js: WARNING: Import "foo" will always be undefined because the file "node_modules/pkg/empty.js" has no exports
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "node_modules/pkg/empty.esm.js"
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "node_modules/pkg/empty.json"
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "node_modules/pkg/empty.css"
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "node_modules/pkg/empty.global-css"
entry.js: WARNING: Import "foo" will always be undefined because there is no matching export in "node_modules/pkg/empty.local-css"
`,
	})
}

func TestCSSMalformedAtImport(t *testing.T) {
	css_suite.expectBundled(t, bundled{
		files: map[string]string{
			"/entry.css": `
				@import "./url-token-eof.css";
				@import "./url-token-whitespace-eof.css";
				@import "./function-token-eof.css";
				@import "./function-token-whitespace-eof.css";
			`,
			"/url-token-eof.css": `@import url(https://example.com/url-token-eof.css`,
			"/url-token-whitespace-eof.css": `
				@import url(https://example.com/url-token-whitespace-eof.css
			`,
			"/function-token-eof.css": `@import url("https://example.com/function-token-eof.css"`,
			"/function-token-whitespace-eof.css": `
				@import url("https://example.com/function-token-whitespace-eof.css"
			`,
		},
		entryPaths: []string{"/entry.css"},
		options: config.Options{
			Mode:         config.ModeBundle,
			AbsOutputDir: "/out",
		},
		expectedScanLog: `function-token-eof.css: WARNING: Expected ")" to go with "("
function-token-eof.css: NOTE: The unbalanced "(" is here:
function-token-whitespace-eof.css: WARNING: Expected ")" to go with "("
function-token-whitespace-eof.css: NOTE: The unbalanced "(" is here:
url-token-eof.css: WARNING: Expected ")" to end URL token
url-token-eof.css: NOTE: The unbalanced "(" is here:
url-token-eof.css: WARNING: Expected ";" but found end of file
url-token-whitespace-eof.css: WARNING: Expected ")" to end URL token
url-token-whitespace-eof.css: NOTE: The unbalanced "(" is here:
url-token-whitespace-eof.css: WARNING: Expected ";" but found end of file
`,
	})
}
