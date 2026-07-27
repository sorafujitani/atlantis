import { execFile } from "node:child_process";
import { readFile } from "node:fs/promises";
import { homedir } from "node:os";
import { join } from "node:path";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);
const contextPrefix =
  "Brain vault index. Read only the linked notes relevant to the task before acting.\n\n";

async function regenerateIndex() {
  try {
    await execFileAsync("atlantis", ["brain", "index"], {
      encoding: "utf8",
      shell: false,
      timeout: 10_000,
    });
  } catch {
    // The durable index remains readable when the optional maintenance tool is unavailable.
  }
}

async function loadContext() {
  await regenerateIndex();
  const root = process.env.ATLANTIS_BRAIN_DIR || join(homedir(), "brain");
  return contextPrefix + (await readFile(join(root, "index.md"), "utf8"));
}

export const AtlantisBrainPlugin = async () => {
  const context = await loadContext();

  return {
    "experimental.chat.system.transform": async (_input, output) => {
      output.system.push(`## Atlantis brain\n\n${context}`);
    },
    "tool.execute.after": regenerateIndex,
  };
};
