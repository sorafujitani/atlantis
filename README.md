# model-orchestrator

[![CI](https://github.com/sorafujitani/model-orchestrator/actions/workflows/ci.yml/badge.svg)](https://github.com/sorafujitani/model-orchestrator/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sorafujitani/model-orchestrator)](https://go.dev/)
[![License](https://img.shields.io/github/license/sorafujitani/model-orchestrator)](./LICENSE)

ローカルで認証済みのagent CLIを組み合わせる、単一バイナリのmodel orchestratorです。上位モデルをAdvisorまたはOrchestrator、低コストモデルをExecutorまたはWorkerとして利用できます。

providerのAPI keyは保存しません。Codex CLI、Claude Codeなどの既存認証とsession機能を、それぞれのCLIへ委譲します。

## 動作モデル

```text
      orchestration request
                 |
       CLI and Agent Skill
                 |
       local run supervisor
       /       |          \
  planner   workers      advisor
      |         |           |
  local CLI subprocess adapters
```

対応する実行パターン:

- `single`: 1つのExecutorで実行
- `advisor`: Executorが必要な場合だけ上位Advisorへ問い合わせ、元sessionを再開
- `orchestrator`: 上位OrchestratorがDAGを作成し、Workerへ委譲して結果を統合
- `hybrid`: Orchestrator + Worker委譲に、WorkerからAdvisorへの限定的なescalationを追加

## インストール

### Go install

```bash
go install github.com/sorafujitani/model-orchestrator/cmd/model-orchestrator@latest
```

### ローカルcheckout

```bash
make install
model-orchestrator version
model-orchestrator doctor
```

`make install` は `${INSTALL_DIR:-$HOME/.local/bin}` に配置します。

## 最初の実行

設定ファイルがなくても、次の既定値で動きます。

- premium: Codex CLIの既定モデル
- standard: Claude Codeの既定モデル
- default profile: Hybrid
- state: `$XDG_STATE_HOME/model-orchestrator` または `~/.local/state/model-orchestrator`

```bash
# Claude executorによる単発実行
model-orchestrator run --mode single \
  --permission read \
  "このrepositoryの構造を3行で要約して"

# Advisor
model-orchestrator run --mode advisor \
  --permission read \
  "2案を比較し、重要な判断だけ上位モデルへ相談して"

# Orchestrator
model-orchestrator run --mode orchestrator \
  "独立調査をworkerへ委譲し、結果を統合して"
```

JSON出力:

```bash
model-orchestrator --output json run --mode single "Return READY in output."
```

## 設定

解決順は flag > environment > config file > default です。

```text
$MODEL_ORCHESTRATOR_CONFIG
$XDG_CONFIG_HOME/model-orchestrator/config.toml
~/.config/model-orchestrator/config.toml
```

設定例は [config.example.toml](./config.example.toml) にあります。

```toml
default_profile = "default"

[models.premium]
adapter = "codex"
model = "gpt-5.6"

[models.standard]
adapter = "claude"
model = "sonnet"
fallback = ["premium"]

[profiles.default]
mode = "hybrid"
orchestrator = "premium"
executor = "standard"
advisor = "premium"
worker = "standard"
reviewer = "premium"
max_workers = 4
max_calls = 20
max_advisor_calls = 1
max_retries = 1
max_duration = "30m"
```

モデル名はCLIが受け付ける値をそのまま指定します。空文字なら各CLIの既定モデルを使います。

### custom CLI adapter

shellは起動しません。`command`と引数配列を直接実行し、placeholderだけを置換します。

```toml
[adapters.local-agent]
command = "local-agent"
args = ["run", "--model", "{model}", "--cwd", "{cwd}", "{prompt}"]
inherit_env = ["LOCAL_AGENT_TOKEN"]
```

placeholder:

- `{prompt}`
- `{cwd}`
- `{model}`
- `{session}`

custom adapterは既定でread-only、最終テキストのみとして扱います。

## コマンド

```text
model-orchestrator run <objective>
model-orchestrator status <run-id>
model-orchestrator inspect <run-id>
model-orchestrator resume <run-id>
model-orchestrator cancel <run-id>
model-orchestrator doctor
model-orchestrator eval <suite.json>
```

### 再開

run stateはappend-only JSONLとして保存されます。中断後は完了済みtaskを再実行せず、残りのtaskから再開します。

```bash
model-orchestrator status run_xxx
model-orchestrator resume run_xxx
```

### 評価

同じsuiteを複数profileで比較できます。

```bash
model-orchestrator --config config.example.toml --output json \
  eval evals/smoke.json \
  --profiles codex,claude
```

reportには成功数、実行時間、token、取得可能な場合のcostを含めます。subscription実行とAPI従量料金は同じものとして扱いません。

## Agent Skill

Agent SkillはCLIの薄い利用手順です。orchestration本体や状態管理をskill内へ重複させません。

```bash
make install-skill
```

またはSkills CLIからglobal installできます。

```bash
npx skills add https://github.com/sorafujitani/model-orchestrator \
  --skill model-orchestrator -g -y
```

skillは依頼に応じて`single`、`advisor`、`orchestrator`、`hybrid`を選び、`model-orchestrator --output json run ...`を実行します。

## 安全境界

- API key、OAuth token、CLI credentialを読み取り・コピー・保存しない
- child processへ渡す環境変数はallowlist方式
- read assignmentは各CLIのread-only / plan modeで実行
- write assignmentはexclusive laneで直列実行
- shell command文字列を組み立てず、`exec`へ引数配列を渡す
- state directoryは`0700`、event fileは`0600`
- providerのraw reasoning eventは永続化しない
- Advisor回数、総call数、retry、worker並列数、実行時間に上限を持つ

## 開発

```bash
make test
make race
make vet
make lint
make build
```

設計の詳細は [docs/architecture.md](./docs/architecture.md) を参照してください。

## License

[MIT](./LICENSE)
