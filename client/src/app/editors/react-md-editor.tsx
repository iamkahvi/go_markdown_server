"use client";
import dynamic from "next/dynamic";
import React, { use, useEffect } from "react";
import type { ContextStore } from "@uiw/react-md-editor";
import { EditorProps } from "./types";

const MDEditor = dynamic(() => import("@uiw/react-md-editor"), { ssr: false });
type MDEditorOnChange = (
  value?: string,
  event?: React.ChangeEvent<HTMLTextAreaElement>,
  state?: ContextStore
) => void;

export function ReactMDEditor({
  initialValue,
  onChange,
  editable,
}: EditorProps) {
  const [value, setValue] = React.useState<string>();

  useEffect(() => {
    if (initialValue !== null) setValue(initialValue);
  }, [initialValue]);

  const onChangeInternal: MDEditorOnChange = (val, _event, _state) => {
    setValue(val || "");
    onChange(val || "");
  };

  return (
    <div className="container">
      <MDEditor
        value={value}
        onChange={onChangeInternal}
        preview="edit"
        textareaProps={{ disabled: !editable }}
        hideToolbar={true}
      />
    </div>
  );
}
