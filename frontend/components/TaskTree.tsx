"use client";

import { useState } from "react";
import { listTasks } from "@/lib/api";
import type { Task } from "@/lib/types";
import TaskRow from "./TaskRow";

// S-03のタスク一覧・M-03の子タスク一覧で使う、クリックで1階層ずつ子タスクを
// インライン展開していくツリー表示(docs/assignment/step5.md「1.」「4-2.」参照)。
//
// mode="dashboard": キャレット(▸▾)クリックで展開/折りたたみ、行クリックでonOpenTask(M-03を開く)
// mode="modal": 行クリック・キャレットクリックのどちらでも展開/折りたたみ(M-03内の子タスク一覧用)
//
// 展開済みの子タスクは一度取得したらstateに保持し、折りたたんでも再展開時に取得し直さない
// (docs/assignment/step5.md「8.」の「実装時に決める」としていたキャッシュ方針の決定)。
export default function TaskTree({
  task,
  depth = 0,
  mode,
  onOpenTask,
}: {
  task: Task;
  depth?: number;
  mode: "dashboard" | "modal";
  onOpenTask?: (id: number) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const [children, setChildren] = useState<Task[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);

  function toggle() {
    if (expanded) {
      setExpanded(false);
      return;
    }
    setExpanded(true);
    if (children === null) {
      setLoading(true);
      setError(false);
      listTasks({ parent_id: task.id })
        .then((res) => setChildren(res.data))
        .catch(() => setError(true))
        .finally(() => setLoading(false));
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <TaskRow
        task={task}
        depth={depth}
        caret={expanded ? "expanded" : "collapsed"}
        onCaretClick={toggle}
        onClick={mode === "dashboard" ? () => onOpenTask?.(task.id) : toggle}
      />
      {expanded && (
        <div className="flex flex-col gap-2">
          {loading && (
            <p className="text-xs text-zinc-400" style={{ marginLeft: (depth + 1) * 28 }}>
              読み込み中...
            </p>
          )}
          {error && (
            <p className="text-xs text-red-600" style={{ marginLeft: (depth + 1) * 28 }}>
              取得に失敗しました
            </p>
          )}
          {!loading && !error && children?.length === 0 && (
            <p className="text-xs text-zinc-400" style={{ marginLeft: (depth + 1) * 28 }}>
              子タスクなし
            </p>
          )}
          {!loading &&
            !error &&
            children?.map((c) => (
              <TaskTree key={c.id} task={c} depth={depth + 1} mode={mode} onOpenTask={onOpenTask} />
            ))}
        </div>
      )}
    </div>
  );
}
