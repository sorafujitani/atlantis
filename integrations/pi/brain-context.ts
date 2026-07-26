import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

export default function brainContext(pi: ExtensionAPI) {
  let context = "";

  async function refresh(): Promise<void> {
    const indexed = await pi.exec("atlantis", ["brain", "index"], { timeout: 10_000 });
    if (indexed.code !== 0) {
      throw new Error(`brain index regeneration failed: ${indexed.stderr}`);
    }
    const injected = await pi.exec("atlantis", ["brain", "inject"], { timeout: 10_000 });
    if (injected.code !== 0) {
      throw new Error(`brain context loading failed: ${injected.stderr}`);
    }
    context = injected.stdout;
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
