"use client";

import { Folder, Search, Settings } from "lucide-react";
import { useEffect, useState } from "react";
import { listLists, listTasks } from "@/lib/api";
import type { Task, TaskList, TaskSort, TaskStatus } from "@/lib/types";
import ListManageModal from "./ListManageModal";
import TaskTree from "./TaskTree";

const SORT_OPTIONS: { value: TaskSort | ""; label: string }[] = [
  { value: "", label: "新しい順" },
  { value: "due_date_asc", label: "締切日順" },
  { value: "due_date_desc", label: "締切日逆順" },
  { value: "estimated_hours_asc", label: "想定工数順" },
  { value: "estimated_hours_desc", label: "想定工数逆順" },
];

const SEARCH_DEBOUNCE_MS = 300;

export default function TaskListSection({
  onCreateClick,
  onOpenTask,
  refreshKey,
}: {
  onCreateClick: () => void;
  onOpenTask: (id: number) => void;
  refreshKey: number;
}) {
  const [lists, setLists] = useState<TaskList[]>([]);
  const [listManageOpen, setListManageOpen] = useState(false);
  const [status, setStatus] = useState<TaskStatus | "">("");
  // ""=すべて(グループ表示) / "null"=未所属 / それ以外=リストID(文字列)
  const [listFilter, setListFilter] = useState<string>("");
  const [qInput, setQInput] = useState("");
  const [q, setQ] = useState("");
  const [sort, setSort] = useState<TaskSort | "">("");
  const [page, setPage] = useState(1);

  const [tasks, setTasks] = useState<Task[]>([]);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const isGrouped = listFilter === "";

  // キーワード検索はタイピングのたびに呼ばず、入力が止まってから呼ぶ(デバウンス、docs/assignment/step5.md「8.」)
  useEffect(() => {
    const timer = setTimeout(() => setQ(qInput), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [qInput]);

  // status/リスト/検索/並び替えを変えたら1ページ目に戻す(docs/assignment/step5.md「8.」)
  useEffect(() => {
    setPage(1);
  }, [status, listFilter, q, sort]);

  function loadLists() {
    listLists()
      .then((res) => setLists(res.data))
      .catch(() => {
        // リスト一覧の取得失敗はフィルタの選択肢が減るだけなので、致命的エラーにはしない
      });
  }

  useEffect(() => {
    loadLists();
  }, []);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    const base = {
      parent_id: "null" as const,
      status: status || undefined,
      q: q || undefined,
      sort: sort || undefined,
    };

    const request = isGrouped
      ? listTasks({ ...base, per_page: 100 })
      : listTasks({
          ...base,
          list_id: listFilter === "null" ? "null" : Number(listFilter),
          page,
          per_page: 20,
        });

    request
      .then((res) => {
        if (cancelled) return;
        setTasks(res.data);
        setTotalPages(res.meta.total_pages);
      })
      .catch(() => {
        if (cancelled) return;
        setError("タスク一覧の取得に失敗しました");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [status, listFilter, q, sort, page, isGrouped, refreshKey]);

  return (
    <section className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-zinc-500">タスク一覧</h2>
        <button
          type="button"
          onClick={onCreateClick}
          className="rounded-md bg-sky-700 px-3 py-1.5 text-sm font-semibold text-white hover:bg-sky-800"
        >
          + 新規タスク作成
        </button>
      </div>

      <div className="flex flex-wrap items-center gap-2 text-sm">
        <div className="relative min-w-[160px] flex-1">
          <Search className="pointer-events-none absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-400" />
          <input
            type="search"
            placeholder="キーワードで検索"
            value={qInput}
            onChange={(e) => setQInput(e.target.value)}
            className="w-full rounded-md border border-zinc-300 py-1 pr-2 pl-7"
          />
        </div>
        <select
          value={status}
          onChange={(e) => setStatus(e.target.value as TaskStatus | "")}
          className="rounded-md border border-zinc-300 px-2 py-1"
        >
          <option value="">status: すべて</option>
          <option value="todo">todo</option>
          <option value="in_progress">in_progress</option>
          <option value="done">done</option>
        </select>
        <select
          value={listFilter}
          onChange={(e) => setListFilter(e.target.value)}
          className="rounded-md border border-zinc-300 px-2 py-1"
        >
          <option value="">リスト: すべて</option>
          <option value="null">未所属</option>
          {lists.map((l) => (
            <option key={l.id} value={l.id}>
              {l.name}
            </option>
          ))}
        </select>
        <button
          type="button"
          onClick={() => setListManageOpen(true)}
          className="flex items-center gap-1 text-sky-700 underline"
        >
          <Settings className="h-3.5 w-3.5" />
          リストを管理
        </button>
        <select
          value={sort}
          onChange={(e) => setSort(e.target.value as TaskSort | "")}
          className="rounded-md border border-zinc-300 px-2 py-1"
        >
          {SORT_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              並び替え: {opt.label}
            </option>
          ))}
        </select>
      </div>

      {error && (
        <p className="rounded-md border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </p>
      )}

      {loading && <p className="text-sm text-zinc-400">読み込み中...</p>}

      {!loading && isGrouped && (
        <div className="flex flex-col gap-4">
          {[...lists.map((l) => ({ id: l.id, name: l.name })), { id: null, name: "未所属" }].map(
            (group) => {
              const groupTasks = tasks.filter((t) => t.list_id === group.id);
              return (
                <div key={group.id ?? "unassigned"} className="flex flex-col gap-2">
                  <div className="flex items-center gap-2 text-sm font-semibold text-zinc-600">
                    {group.id !== null && <Folder className="h-3.5 w-3.5" />}
                    <span>{group.name}</span>
                    <span className="rounded-full border border-zinc-300 px-2 py-0.5 font-mono text-xs text-zinc-500">
                      {groupTasks.length}件
                    </span>
                  </div>
                  {groupTasks.length === 0 ? (
                    <p className="text-xs text-zinc-400">該当タスクなし</p>
                  ) : (
                    groupTasks.map((t) => (
                      <TaskTree key={t.id} task={t} mode="dashboard" onOpenTask={onOpenTask} />
                    ))
                  )}
                </div>
              );
            },
          )}
        </div>
      )}

      {!loading && !isGrouped && (
        <div className="flex flex-col gap-2">
          {tasks.length === 0 ? (
            <p className="text-sm text-zinc-400">該当タスクなし</p>
          ) : (
            tasks.map((t) => (
              <TaskTree key={t.id} task={t} mode="dashboard" onOpenTask={onOpenTask} />
            ))
          )}
          {totalPages > 1 && (
            <div className="mt-2 flex justify-center gap-1 font-mono text-xs">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
                className="rounded border border-zinc-300 px-2 py-1 text-zinc-500 disabled:opacity-40"
              >
                &lt;
              </button>
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <button
                  key={p}
                  type="button"
                  onClick={() => setPage(p)}
                  className={`rounded border px-2 py-1 ${
                    p === page
                      ? "border-sky-500 text-sky-700"
                      : "border-zinc-300 text-zinc-500"
                  }`}
                >
                  {p}
                </button>
              ))}
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="rounded border border-zinc-300 px-2 py-1 text-zinc-500 disabled:opacity-40"
              >
                &gt;
              </button>
            </div>
          )}
        </div>
      )}

      <p className="font-mono text-xs text-zinc-400">
        {isGrouped
          ? "「すべて」はlist_idごとにグループ表示(最大100件、ページネーションなし)"
          : "特定のリスト・「未所属」ではページネーション付きのフラット表示"}
      </p>

      {listManageOpen && (
        <ListManageModal
          onClose={() => setListManageOpen(false)}
          onChanged={loadLists}
        />
      )}
    </section>
  );
}
