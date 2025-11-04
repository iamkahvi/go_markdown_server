import { defaultValueCtx, Editor, rootCtx } from "@milkdown/kit/core";

import {
  Milkdown,
  MilkdownProvider,
  useEditor,
  useInstance,
} from "@milkdown/react";
import { commonmark } from "@milkdown/kit/preset/commonmark";
import { nord } from "@milkdown/theme-nord";

import "@milkdown/theme-nord/style.css";
import { listener, listenerCtx } from "@milkdown/kit/plugin/listener";
import { EditorProps } from "./types";
import { useEffect } from "react";
import { replaceAll } from "@milkdown/kit/utils";

function MilkdownEditorInner({ initialValue, onChange }: EditorProps) {
  useEditor((root) => {
    const editor = Editor.make()
      .config((ctx) => {
        ctx.set(rootCtx, root);
        ctx.get(listenerCtx).markdownUpdated((ctx, markdown) => {
          // Save content to your backend or storage
          onChange(markdown);
        });
      })
      .config(nord)
      .use(commonmark)
      .use(listener);

    editor.action((ctx) => {});

    return editor;
  }, []);

  const [_isLoading, getInstance] = useInstance();
  const editor = getInstance();

  useEffect(() => {
    if (initialValue !== null && editor) {
      editor.action(replaceAll(initialValue));
    }
  }, [initialValue, editor]);

  return <Milkdown />;
}

export function MilkdownEditor({ initialValue, onChange }: EditorProps) {
  return (
    <MilkdownProvider>
      <MilkdownEditorInner initialValue={initialValue} onChange={onChange} />
    </MilkdownProvider>
  );
}
