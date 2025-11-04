export type Diff = [number, string];

export type PatchObj = {
  diffs: Diff[];
  start1: number | null;
  start2: number | null;
  length1: number;
  length2: number;
};

export type Patch = [-1 | 0 | 1, string];

export interface Message {
  patches: Patch[];
  patchObjs: PatchObj[];
}

export interface EditorResponse {
  type: "editor";
  status: "OK" | "ERROR";
  doc?: string;
}

export interface ClientResponse {
  type: "client";
  count: number;
}

export type MyResponse = EditorResponse | ClientResponse;
