"use client";
import React from "react";
import { MilkdownEditor } from "./editors/milkdown-editor";
import { ReadOnlyMilkdownEditor } from "./editors/read-only-milkdown-editor";
import { StatusBar } from "./components/status-bar";
import { useEditorSync } from "../hooks/useEditorSync";

export default function Home() {
  const { editorValue, onChange, status, info } = useEditorSync();

  if (status === "loading") return <div>loading</div>;

  if (status === "disconnected") return <div>disconnected</div>;

  return (
    <div className="ml-auto mr-auto max-w-xl py-4">
      <div className="container w-auto grid gap-4 mx-4">
        <h1 className="text-3xl font-bold underline pt-2">kahvi's notepad</h1>
        <StatusBar
          clientCount={info.clientCount}
          editorState={info.editorState}
        />
        <div
          className="editor border rounded-md overflow-auto p-4"
          style={{ height: "36rem" }}
        >
          {info.editorState === "EDITOR" ? (
            <MilkdownEditor
              initialValue={editorValue}
              onChange={onChange}
              editable={info.editorState === "EDITOR"}
            />
          ) : (
            <ReadOnlyMilkdownEditor value={editorValue} />
          )}
        </div>
      </div>
    </div>
  );
}
