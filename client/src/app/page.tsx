"use client";
import React from "react";
import { MilkdownEditor } from "./editors/milkdown-editor";
import { ReadOnlyMilkdownEditor } from "./editors/read-only-milkdown-editor";
import { StatusBar } from "./components/status-bar";
import { useEditorSync } from "../hooks/useEditorSync";

export default function Home() {
  const { value, isOpen, clientCount, editorState, onChange } = useEditorSync();

  if (!isOpen) return <div>loading</div>;

  return (
    <div className="ml-auto mr-auto max-w-4xl my-4">
      <div>
        <h1 className="text-3xl font-bold underline pt-2 pb-5">
          kahvi's notepad
        </h1>
      </div>
      <div className="w-auto grid gap-4 mx-4">
        <StatusBar clientCount={clientCount} editorState={editorState} />
        <div
          className="editor border rounded-md overflow-auto p-4"
          style={{ height: "36rem" }}
        >
          {editorState === "EDITOR" ? (
            <MilkdownEditor
              initialValue={value}
              onChange={onChange}
              editable={editorState === "EDITOR"}
            />
          ) : (
            <ReadOnlyMilkdownEditor value={value} />
          )}
        </div>
      </div>
    </div>
  );
}
