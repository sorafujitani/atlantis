# Architecture

## Invariants

- roleとmodel/CLIを分離する。
- 外部CLIの出力はadapter境界で検証し、内部では型付きresultだけを扱う。
- supervisorだけがrun stateを更新する。
- read taskはbounded parallel、write taskはexclusiveに実行する。
- 完了eventは再実行より優先され、resumeは同じ終状態へ収束する。
- native sessionが使える場合はresumeし、使えない場合はcheckpoint contextで再実行する。
- raw provider reasoningとcredentialはstateへ保存しない。

## Dependency direction

```text
Agent Skill / CLI
    |          |
  Engine     Brain
  /   \        |
State Adapter  Vault files
   \   /
Contracts
```

`orchestration` packageはCLIやprovider固有型を参照しません。`adapter`は外部CLIを共通`ExecutionResult`へ変換します。`engine`はrole、budget、DAG、retry、fallback、resumeを所有します。`brain`はvaultのindex、validation、plan lifecycleだけを所有し、orchestration stateやprovider credentialを参照しません。Agent Skillは両機能の利用手順とplaybookだけを持ちます。

## Advisor

1. Executorを実行する。
2. `needs_advice`なら構造化された`AdviceRequest`を検証する。
3. Advisorをread permissionで実行する。
4. Advisor結果をExecutorのnative sessionへ返す。
5. Advisor回数の上限を超えた場合は停止する。

## Orchestrator

1. Orchestratorが最小DAGを返す。
2. cycle、未知のdependency、role、permissionを検証する。
3. readyなread taskを並列実行する。
4. write taskを1件ずつ実行する。
5. 全worker resultを元のOrchestrator sessionへ返して統合する。

## State

```text
~/.local/state/atlantis/runs/<run-id>/
├── events.jsonl
└── lock
```

snapshotはevent replayで導出します。lockには実行中supervisorのPIDを保存し、stale PIDは次回起動時に回収します。

保存するprovider情報は、正規化済みresult、usage、adapter名、native session IDだけです。
