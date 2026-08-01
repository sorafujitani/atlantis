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

async function loadContext(state, exec) {
  const fingerprint = (await exec(["brain", "context", "--print-fingerprint"])).trim();
  if (fingerprint && fingerprint === state.fingerprint && state.context) {
    return state.context;
  }
  const raw = await exec(["-o", "json", "brain", "context"]);
  const parsed = JSON.parse(raw);
  state.fingerprint = typeof parsed.fingerprint === "string" ? parsed.fingerprint : fingerprint;
  state.context = typeof parsed.context === "string" ? parsed.context : "";
  return state.context;
}

function registerPiExtension(pi) {
  const state = { fingerprint: "", context: "" };
  const exec = async (args) => {
    const result = await pi.exec("atlantis", args, { timeout: 10_000 });
    return result.stdout;
  };

  async function refresh() {
    await loadContext(state, exec);
  }

  pi.on("session_start", refresh);
  pi.on("agent_settled", refresh);
  pi.on("before_agent_start", async (event) => {
    if (!state.context) {
      await refresh();
    }
    return { systemPrompt: `${event.systemPrompt}\n\n## Atlantis brain\n\n${state.context}` };
  });
}

function createOpenCodePlugin() {
  const state = { fingerprint: "", context: "" };
  return {
    "experimental.chat.system.transform": async (_input, output) => {
      const context = await loadContext(state, runAtlantis);
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
