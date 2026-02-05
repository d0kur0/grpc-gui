import { onMount, onCleanup, createEffect } from "solid-js";
import { render } from "solid-js/web";
import { EditorView, ViewPlugin, WidgetType, ViewUpdate } from "@codemirror/view";
import { Decoration } from "@codemirror/view";
import { EditorState } from "@codemirror/state";
import { javascript } from "@codemirror/lang-javascript";
import { basicSetup } from "codemirror";
import { materialDark } from "@ddietr/codemirror-themes/material-dark";
import { autocompletion, CompletionContext, CompletionResult } from "@codemirror/autocomplete";
import { MessageInfo, FieldInfo } from "../../bindings/grpc-gui/internal/grpcreflect/models";
import stripJsonComments from "strip-json-comments";
import { EnumPopover } from "./EnumPopover";
import { TimestampPopover } from "./TimestampPopover";
import { DurationPopover } from "./DurationPopover";

import "./JsonEditor.css";

export type JsonEditorProps = {
	value: string;
	onChange?: (value: string) => void;
	placeholder?: string;
	readOnly?: boolean;
	schema?: MessageInfo | null;
};

const getFieldsFromSchema = (schema: MessageInfo | null | undefined): Map<string, FieldInfo> => {
	const fields = new Map<string, FieldInfo>();
	if (!schema?.fields) return fields;

	const addFields = (prefix: string, message: MessageInfo | null | undefined) => {
		if (!message?.fields) return;
		
		for (const field of message.fields) {
			const fieldPath = prefix ? `${prefix}.${field.name}` : field.name;
			fields.set(fieldPath, field);
			
			if (field.message) {
				addFields(fieldPath, field.message);
			}
		}
	};

	addFields("", schema);
	return fields;
};

class EnumInfoWidget extends WidgetType {
	private dispose: (() => void) | null = null;

	constructor(
		private enumValues: Array<{ name: string; number: number }>,
		private fieldPath: string
	) {
		super();
	}

	toDOM() {
		const span = document.createElement("span");
		this.dispose = render(() => EnumPopover({ enumValues: this.enumValues, fieldPath: this.fieldPath }), span);
		return span;
	}

	destroy() {
		this.dispose?.();
	}
}

class TimestampInfoWidget extends WidgetType {
	private dispose: (() => void) | null = null;

	constructor(
		private fieldPath: string,
		private view: EditorView,
		private pos: number
	) {
		super();
	}

	toDOM() {
		const span = document.createElement("span");
		this.dispose = render(
			() =>
				TimestampPopover({
					fieldPath: this.fieldPath,
					onSelect: (value) => {
						const currentText = this.view.state.doc.toString();
						const fieldPattern = new RegExp(
							`"${this.fieldPath.split('.').pop()}"\\s*:\\s*"[^"]*"`,
							'g'
						);
						const newText = currentText.replace(
							fieldPattern,
							`"${this.fieldPath.split('.').pop()}": "${value}"`
						);
						this.view.dispatch({
							changes: { from: 0, to: currentText.length, insert: newText },
						});
					},
				}),
			span
		);
		return span;
	}

	destroy() {
		this.dispose?.();
	}
}

class DurationInfoWidget extends WidgetType {
	private dispose: (() => void) | null = null;

	constructor(
		private fieldPath: string,
		private view: EditorView,
		private pos: number
	) {
		super();
	}

	toDOM() {
		const span = document.createElement("span");
		this.dispose = render(
			() =>
				DurationPopover({
					fieldPath: this.fieldPath,
					onSelect: (value) => {
						const currentText = this.view.state.doc.toString();
						const fieldPattern = new RegExp(
							`"${this.fieldPath.split('.').pop()}"\\s*:\\s*"[^"]*"`,
							'g'
						);
						const newText = currentText.replace(
							fieldPattern,
							`"${this.fieldPath.split('.').pop()}": "${value}"`
						);
						this.view.dispatch({
							changes: { from: 0, to: currentText.length, insert: newText },
						});
					},
				}),
			span
		);
		return span;
	}

	destroy() {
		this.dispose?.();
	}
}

const createEnumInfoPlugin = (schema: MessageInfo | null | undefined) => {
	const fields = getFieldsFromSchema(schema);
	
	return ViewPlugin.fromClass(class {
		decorations: any;
		private DecorationClass: any;
		private view: EditorView;
		
		constructor(view: EditorView) {
			this.view = view;
			this.DecorationClass = Decoration as any;
			this.decorations = this.buildDecorations(view);
		}
		
		update(update: ViewUpdate) {
			if (update.docChanged || update.viewportChanged) {
				this.decorations = this.buildDecorations(update.view);
			}
		}
		
		buildDecorations(view: EditorView): any {
			if (!schema) return this.DecorationClass.none;
			
			const decorations: any[] = [];
			const text = view.state.doc.toString();
			
			try {
				const cleanedText = stripJsonComments(text);
				const json = JSON.parse(cleanedText);
				this.findEnumValues(json, "", fields, decorations, view.state);
			} catch {
			}
			
			return this.DecorationClass.set(decorations);
		}
		
		findEnumValues(
			obj: any,
			path: string,
			fields: Map<string, FieldInfo>,
			decorations: any[],
			state: EditorState
		) {
			if (typeof obj !== "object" || obj === null) return;
			
			for (const [key, value] of Object.entries(obj)) {
				const currentPath = path ? `${path}.${key}` : key;
				const field = fields.get(currentPath);
				
				if (field?.isEnum && field.enumValues && typeof value === "string") {
					const escapedValue = value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
					const searchPattern = new RegExp(`"${key.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}"\\s*:\\s*"${escapedValue}"`, 'g');
					const text = state.doc.toString();
					let match;
					
					while ((match = searchPattern.exec(text)) !== null) {
						const matchStart = match.index;
						const fullMatch = match[0];
						const valueStart = matchStart + fullMatch.indexOf(`"${value}"`);
						const valueEnd = valueStart + value.length + 2;
						
						const widget = new EnumInfoWidget(field.enumValues, currentPath);
						
						decorations.push(
							this.DecorationClass.widget({
								widget,
								side: 1,
							}).range(valueEnd)
						);
					}
				}
				
				if (field?.isWellKnown && typeof value === "string") {
					const escapedValue = value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
					const searchPattern = new RegExp(`"${key.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}"\\s*:\\s*"${escapedValue}"`, 'g');
					const text = state.doc.toString();
					let match;
					
					while ((match = searchPattern.exec(text)) !== null) {
						const matchStart = match.index;
						const fullMatch = match[0];
						const valueStart = matchStart + fullMatch.indexOf(`"${value}"`);
						const valueEnd = valueStart + value.length + 2;
						
						let widget;
						if (field.wellKnownType === "timestamp") {
							widget = new TimestampInfoWidget(currentPath, (this as any).view, valueEnd);
						} else if (field.wellKnownType === "duration") {
							widget = new DurationInfoWidget(currentPath, (this as any).view, valueEnd);
						}
						
						if (widget) {
							decorations.push(
								this.DecorationClass.widget({
									widget,
									side: 1,
								}).range(valueEnd)
							);
						}
					}
				}
				
				if (typeof value === "object" && value !== null && !Array.isArray(value)) {
					this.findEnumValues(value, currentPath, fields, decorations, state);
				}
			}
		}
	}, {
		decorations: v => v.decorations,
	});
};

const createJsonAutocompletion = (schema: MessageInfo | null | undefined) => {
	return (context: CompletionContext): CompletionResult | null => {
		if (!schema?.fields) return null;

		const word = context.matchBefore(/\w*/);
		if (!word || (word.from === word.to && !context.explicit)) return null;

		const fields = getFieldsFromSchema(schema);
		const options = [];

		const textBefore = context.state.doc.sliceString(0, context.pos);
		const fieldMatch = textBefore.match(/"(\w+)"\s*:\s*"?[^"]*$/);
		
		if (fieldMatch) {
			const fieldName = fieldMatch[1];
			
			for (const [path, field] of fields) {
				if (path.endsWith(fieldName) && field.isEnum && field.enumValues) {
					for (const enumValue of field.enumValues) {
						options.push({
							label: enumValue.name,
							type: "enum",
							detail: `enum (${enumValue.number})`,
							apply: enumValue.name,
						});
					}
					break;
				}
			}
		} else {
			for (const [path, field] of fields) {
				const displayName = path.split('.').pop() || path;
				
				let typeDetail = field.type;
				if (field.repeated) typeDetail = `${typeDetail}[]`;
				if (field.isEnum) typeDetail += " (enum)";
				
				options.push({
					label: displayName,
					type: "property",
					detail: typeDetail,
					apply: displayName,
				});
			}
		}

		if (options.length === 0) return null;

		return {
			from: word.from,
			options,
		};
	};
};

export const JsonEditor = (props: JsonEditorProps) => {
	let editorRef: HTMLDivElement | undefined;
	let view: EditorView | undefined;

	onMount(() => {
		if (!editorRef) return;

		const backgroundTheme = EditorView.theme({
			"&": {
				height: "100%",
				backgroundColor: "oklch(var(--b2)) !important",
			},
			".cm-gutters": {
				backgroundColor: "oklch(var(--b2)) !important",
			},
			".cm-scroller": {
				fontFamily: "ui-monospace, monospace",
			},
		}, { dark: true });

		const selectionTheme = EditorView.theme({
			"& .cm-selectionLayer .cm-selectionBackground": {
				backgroundColor: "rgba(189, 147, 249, 0.20) !important",
			},
			"&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground": {
				backgroundColor: "rgba(189, 147, 249, 0.20) !important",
			},
			"& .cm-line ::selection": {
				color: "oklch(var(--sc)) !important",
			},
			"& ::selection": {
				color: "oklch(var(--sc)) !important",
			},
		}, { dark: true });

		const extensions = [
			basicSetup,
			javascript(),
			backgroundTheme,
			materialDark,
			selectionTheme,
			EditorView.lineWrapping,
			EditorView.updateListener.of((update) => {
				if (update.docChanged && props.onChange) {
					props.onChange(update.state.doc.toString());
				}
			}),
			EditorState.readOnly.of(props.readOnly || false),
		];

		if (props.schema && !props.readOnly) {
			extensions.push(
				autocompletion({
					override: [createJsonAutocompletion(props.schema)],
					activateOnTyping: true,
				})
			);
		}

		if (props.schema) {
			extensions.push(createEnumInfoPlugin(props.schema));
		}

		const startState = EditorState.create({
			doc: props.value || "",
			extensions,
		});

		view = new EditorView({
			state: startState,
			parent: editorRef,
		});
	});

	createEffect(() => {
		if (!view) return;
		
		const currentValue = view.state.doc.toString();
		if (currentValue !== props.value) {
			view.dispatch({
				changes: {
					from: 0,
					to: currentValue.length,
					insert: props.value,
				},
			});
		}
	});

	onCleanup(() => {
		view?.destroy();
	});

	return <div ref={editorRef} class="w-full h-full rounded-lg overflow-hidden" />;
};
