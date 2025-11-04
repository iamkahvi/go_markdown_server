"use client";
import React, { useRef, useEffect, useState } from "react";
import diff_match_patch from "diff-match-patch";
import { ReactMDEditor } from "./editors/react-md-editor";
import { Diff, Message, MyResponse, PatchObj } from "./protocol";
import { OnChange } from "./editors/types";
import { MilkdownEditor } from "./editors/milkdown-editor";

const SERVER_URL = "ws://100.116.9.20:8000/write";

const FIRST_MESSAGE: Message = {
  patches: [],
  patchObjs: [],
};

export default function Home() {
  const diffMatchPatch = new diff_match_patch();
  const ws = useRef<WebSocket | null>(null);
  const [initialValue, setInitialValue] = useState<string | null>(null);
  const syncedValueRef = useRef("");
  const [isOpen, setIsOpen] = useState(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>();
  const [clientCount, setClientCount] = useState<number | null>(null);

  useEffect(() => {
    ws.current = new WebSocket(SERVER_URL);

    ws.current.onopen = function (evt) {
      setIsOpen(true);
      if (ws.current) {
        // it would be nice to enforce the api by altering the send method type
        ws.current.send(JSON.stringify(FIRST_MESSAGE));
      }
      console.log("OPEN");
    };

    ws.current.onclose = function (evt) {
      setIsOpen(false);
      ws.current = null;
      console.log("CLOSE");
    };

    ws.current.onmessage = function (evt) {
      const message: MyResponse = JSON.parse(evt.data);

      console.log("RECEIVED: " + JSON.stringify(message));

      switch (message.type) {
        case "client":
          setClientCount(message.count);
          break;
        case "editor":
          {
            if (message.status === "OK" && message.doc) {
              console.log("OK from server, doc: ", message.doc);
              // set the initial value of the editor
              setInitialValue(message.doc);
              syncedValueRef.current = message.doc;
            }

            if (message.status === "ERROR") {
              console.log("ERROR from server");
              // stop showing editor
              setIsOpen(false);
              window.alert("An error occurred while syncing with the server.");
            }
          }
          break;
      }
    };

    ws.current.onerror = function (evt: Event) {
      setIsOpen(false);
      console.log("ERROR: " + (evt as ErrorEvent).message);
    };

    const wsCurr = ws.current;

    return () => {
      wsCurr.close();
    };
  }, [setIsOpen]);

  const onChange: OnChange = (val) => {
    if (!ws.current) return;

    clearTimeout(timeoutRef.current);

    const messageData: string = val || "";

    if (messageData.length > 536870888) {
      window.alert("uh oh, we're writing too much!");
    }

    timeoutRef.current = setTimeout(() => {
      if (ws.current) {
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

  if (!isOpen) return <div>loading</div>;

  return (
    <div>
      <div>
        <h1 className="text-3xl font-bold underline p-4">
          Notepad: {clientCount}
        </h1>
      </div>
      <main className=" w-auto mx-auto p-4">
        {/* <ReactMDEditor initialValue={initialValue} onChange={onChange} /> */}
        <MilkdownEditor initialValue={initialValue} onChange={onChange} />
      </main>
    </div>
  );
}
