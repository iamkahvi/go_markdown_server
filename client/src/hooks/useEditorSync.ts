"use client";
import { useRef, useEffect, useState } from "react";
import diff_match_patch from "diff-match-patch";
import { Diff, Message, MyResponse, PatchObj } from "../app/protocol";
import { OnChange } from "../app/editors/types";

// const SERVER_URL = "ws://100.116.9.20:8000/write";
const SERVER_URL = "ws://localhost:8000/write";

const FIRST_MESSAGE: Message = {
  patches: [],
  patchObjs: [],
};

export function useEditorSync() {
  const diffMatchPatch = new diff_match_patch();
  const ws = useRef<WebSocket | null>(null);
  const [value, setInitialValue] = useState<string | null>(null);
  const syncedValueRef = useRef("");
  const [status, setStatus] = useState<
    "loading" | "disconnected" | "connected"
  >("loading");
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>();
  const [clientCount, setClientCount] = useState<number>(0);
  const editorStateRef = useRef("");

  useEffect(() => {
    ws.current = new WebSocket(SERVER_URL);

    ws.current.onopen = function (evt) {
      setStatus("connected");
      if (ws.current) {
        ws.current.send(JSON.stringify(FIRST_MESSAGE));
      }
      console.log("OPEN");
    };

    ws.current.onclose = function (evt) {
      setStatus("disconnected");
      ws.current = null;
      console.log("CLOSE");
    };

    ws.current.onmessage = function (evt) {
      const message: MyResponse = JSON.parse(evt.data);

      console.log(`RECEIVED ${message.type}: ` + JSON.stringify(message));

      switch (message.type) {
        case "client":
          setClientCount(message.count);
          break;
        case "editor": {
          if (message.status === "ERROR") {
            console.log("ERROR from server");
            setStatus("disconnected");
            window.alert("An error occurred while syncing with the server.");
          }
          break;
        }
        case "state": {
          if (message.state === "EDITOR") {
            editorStateRef.current = "EDITOR";
            setInitialValue(message.initialDoc);
            syncedValueRef.current = message.initialDoc;
          } else {
            editorStateRef.current = "READER";
            setInitialValue(message.initialDoc);
          }
          break;
        }
        case "reader": {
          if (message.status === "OK") {
            setInitialValue(message.doc);
          }
          break;
        }
      }
    };

    ws.current.onerror = function (evt: Event) {
      setStatus("disconnected");
      console.log("ERROR: " + (evt as ErrorEvent).message);
    };

    const wsCurr = ws.current;

    return () => {
      wsCurr.close();
    };
  }, []);

  useEffect(() => {
    document.title = `${editorStateRef.current} - note`;
  }, [editorStateRef.current]);

  const onChange: OnChange = (val) => {
    if (!ws.current) return;

    clearTimeout(timeoutRef.current);

    const messageData: string = val || "";
    if (messageData.length > 536870888) {
      window.alert("uh oh, we're writing too much!");
    }

    timeoutRef.current = setTimeout(() => {
      if (ws.current && editorStateRef.current === "EDITOR") {
        const previousSyncedValue = syncedValueRef.current;
        const patches = diffMatchPatch.diff_main(
          previousSyncedValue,
          messageData
        );
        const patchObjs: PatchObj[] = diffMatchPatch
          .patch_make(previousSyncedValue, messageData)
          .map((patch: any) => {
            return {
              diffs: patch.diffs as Diff[],
              start1: patch.start1,
              start2: patch.start2,
              length1: patch.length1,
              length2: patch.length2,
            };
          });
        const message: Message = { patches, patchObjs };
        syncedValueRef.current = messageData;
        console.log("sending: ", message);
        ws.current.send(JSON.stringify(message));
      }
    }, 50);
  };

  return {
    editorValue: value,
    onChange,
    status,
    info: {
      clientCount,
      editorState: editorStateRef.current,
    },
  };
}
