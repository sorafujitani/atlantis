import { execFile } from "node:child_process";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);

async function runAtlantis(args) {
  const result = await execFileAsync("atlantis", args, {
    encoding: "utf8",
    shell: false,
    timeout: 10_000,
  });
  return result.stdout;
}


function registerPiExtension(pi) {
  let context = "";

  async function refresh() {
    const result = await pi.exec("atlantis", ["brain", "context"], {
      timeout: 10_000,
    });
    context = result.stdout;
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

function createOpenCodePlugin() {
  return {
    "experimental.chat.system.transform": async (_input, output) => {
      const context = await runAtlantis(["brain", "context"]);
      output.system.push(`## Atlantis brain\n\n${context}`);
    },
  };
}

export default function atlantisBrain(host) {
  if (typeof host?.on === "function" && typeof host?.exec === "function") {
    registerPiExtension(host);
    return;
  }
  return createOpenCodePlugin();
}
