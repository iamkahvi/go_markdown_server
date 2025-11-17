"use client";
import React from "react";

export function StatusBar({
  clientCount,
  editorState,
}: {
  clientCount: number | null;
  editorState: string;
}) {
  return (
    <div className="info border rounded-md p-4 flex justify-between max-h-20">
      <div>clients: {clientCount}</div>
      <div>editorState: {editorState}</div>
    </div>
  );
}
