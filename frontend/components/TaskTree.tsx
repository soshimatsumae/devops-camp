"use client";

import { ChevronRight } from "lucide-react";
import { useState } from "react";
import { listTasks } from "@/lib/api";
import { STATUS_CHIP_CLASS, STATUS_LABEL } from "@/lib/taskStatus";
import type { Task } from "@/lib/types";

// S-03のタスク一覧・M-03の子タスク一覧で使う、クリックで1階層ずつ子タスクを
// インライン展開していくツリー表示(docs/assignment/step5.md「1.」「4-2.」参照)。
// ネイティブの<details>/<summary>を使い、開閉状態そのものはDOM(open属性)に
// 任せる。Reactはクリックのたびに取得する子タスクのキャッシュだけを持つ。
//
// 同じ階層で複数のタスクを同時に展開できる設計(1つ開くと他が閉じる、という
// 排他動作ではない)ため、兄弟の<details>に共通のname属性は付けていない
// (nameを揃えると「同時に1つしか開けない」動作になってしまうため)。
//
// mode="dashboard": summary内の行クリックで展開/折りたたみ。ただしタイトル
//   部分だけはネストしたボタンにして、そのクリックはpreventDefault/
//   stopPropagationでsummaryのネイティブなトグルを止め、代わりにM-03を開く
// mode="modal": summary(行全体)のクリックで展開/折りたたみのみ(ネイティブの
//   挙動そのまま、タスク詳細を開く機能は無い)
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
  const [children, setChildren] = useState<Task[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);
  // 矢印の回転だけに使う、見た目専用の状態。ネイティブのopen属性自体は
  // DOMに任せたままで、この値はtoggleイベントからミラーしているだけ。
  // (TailwindのgroupOpen:は「同じgroupクラスを持つ祖先のどれか」にマッチするため、
  // 入れ子の<details>では自分自身の開閉と無関係に親の状態を拾ってしまう。
  // そのため祖先探索に頼らず、自分のtoggleイベントだけを見るこの方式にしている)
  const [open, setOpen] = useState(false);

  function handleToggle(e: React.SyntheticEvent<HTMLDetailsElement>) {
    const isOpen = e.currentTarget.open;
    setOpen(isOpen);
    if (!isOpen || children !== null) return;
    setLoading(true);
    setError(false);
    listTasks({ parent_id: task.id })
      .then((res) => setChildren(res.data))
      .catch(() => setError(true))
      .finally(() => setLoading(false));
  }

  return (
    <details style={{ marginLeft: depth * 28 }} onToggle={handleToggle}>
      <summary className="flex cursor-pointer list-none items-center justify-between gap-2 rounded-md px-3 py-2 marker:content-none hover:bg-zinc-50 [&::-webkit-details-marker]:hidden">
        <span className="flex items-center gap-2">
          <ChevronRight
            className={`h-4 w-4 shrink-0 text-sky-700 transition-transform duration-150 ${open ? "rotate-90" : ""}`}
          />
          {mode === "dashboard" ? (
            <button
              type="button"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onOpenTask?.(task.id);
              }}
              className="text-left text-[0.95rem] text-zinc-900 hover:underline"
            >
              {task.title}
            </button>
          ) : (
            <span className="text-[0.95rem] text-zinc-900">{task.title}</span>
          )}
        </span>
        <span className="flex items-center gap-2 font-mono text-xs text-zinc-500">
          <span
            className={`rounded border px-1.5 py-0.5 font-semibold ${STATUS_CHIP_CLASS[task.status]}`}
          >
            {STATUS_LABEL[task.status]}
          </span>
          <span>
            想定:{task.estimated_hours}h 実績:{task.actual_hours ?? "-"}
          </span>
        </span>
      </summary>

      <div className="accordion-panel flex flex-col gap-2 pt-2">
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
        {!loading &&
          !error &&
          children?.map((c) => (
            <TaskTree key={c.id} task={c} depth={depth + 1} mode={mode} onOpenTask={onOpenTask} />
          ))}
      </div>
    </details>
  );
}
