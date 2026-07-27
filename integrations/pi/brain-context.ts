import { readFile } from "node:fs/promises";
import { homedir } from "node:os";
import { join } from "node:path";

import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

const contextPrefix =
  "Brain vault index. Read only the linked notes relevant to the task before acting.\n\n";

export default function brainContext(pi: ExtensionAPI) {
  let context = "";

  async function refreshIndex(): Promise<void> {
    try {
      await pi.exec("atlantis", ["brain", "index"], { timeout: 10_000 });
    } catch {
      // The durable index remains readable when the optional maintenance tool is unavailable.
    }
  }

  async function refresh(): Promise<void> {
    await refreshIndex();
    const root = process.env.ATLANTIS_BRAIN_DIR || join(homedir(), "brain");
    context = contextPrefix + (await readFile(join(root, "index.md"), "utf8"));
  }

  pi.on("session_start", refresh);
  pi.on("agent_settled", refresh);
  pi.on("before_agent_start", async (event) => {
    if (!context) {
      await refresh();
    }
    return { systemPrompt: `${event.systemPrompt}\n\n## Atlantis brain\n\n${context}` };
  });
}
