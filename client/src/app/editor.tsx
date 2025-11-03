"use client";
import dynamic from "next/dynamic";
import React from "react";
import type { ContextStore } from "@uiw/react-md-editor";

const MDEditor = dynamic(() => import("@uiw/react-md-editor"), { ssr: false });
type OnChange = (
  value?: string,
  event?: React.ChangeEvent<HTMLTextAreaElement>,
  state?: ContextStore
) => void;

export default function Editor({
  value,
  onChange,
}: {
  value: string;
  onChange: OnChange;
}) {
  const onChangeInternal: OnChange = (val, event, state) => {
    onChange(val || "", event, state);
  };

  return (
    <div className="container">
      <MDEditor value={value} onChange={onChangeInternal as OnChange} />
    </div>
  );
}
