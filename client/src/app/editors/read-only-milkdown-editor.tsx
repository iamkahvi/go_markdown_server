import {
  defaultValueCtx,
  Editor,
  editorViewOptionsCtx,
  rootCtx,
} from "@milkdown/kit/core";

import {
  Milkdown,
  MilkdownProvider,
  useEditor,
  useInstance,
} from "@milkdown/react";
import { commonmark } from "@milkdown/kit/preset/commonmark";
import { nord } from "@milkdown/theme-nord";

import "@milkdown/theme-nord/style.css";
import { listener } from "@milkdown/kit/plugin/listener";
import { useEffect } from "react";
import { replaceAll } from "@milkdown/kit/utils";

function MilkdownEditorInner({ value }: { value: string | null }) {
  useEditor((root) => {
    const editor = Editor.make()
      .config((ctx) => {
        ctx.set(rootCtx, root);
        ctx.update(editorViewOptionsCtx, (prev) => ({
          ...prev,
          editable: () => false,
        }));
      })
      .config(nord)
      .use(commonmark)
      .use(listener);

    return editor;
  }, []);

  const [_isLoading, getInstance] = useInstance();
  const editor = getInstance();

  useEffect(() => {
    if (value !== null && editor) {
      editor.action(replaceAll(value));
    }
  }, [value, editor]);

  return <Milkdown />;
}

export function ReadOnlyMilkdownEditor({ value }: { value: string | null }) {
  return (
    <MilkdownProvider>
      <MilkdownEditorInner value={value} />
    </MilkdownProvider>
  );
}
